package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable/encoding/pfdtsv"
	"github.com/Kuniwak/pfd-tools/sugar"
	"github.com/Kuniwak/pfd-tools/tools"
	"github.com/google/go-cmp/cmp"
)

const simplePath = "testdata/simple/pfd.drawio"

const brokenDrawio = `<mxfile host="x">
    <diagram id="d1" name="P1">
        <mxGraphModel><root><mxCell id="0"/><mxCell id="1" parent="0"/></root></mxGraphModel>
    </diagram>
</mxfile>
`

func TestMainCommandByArgs(t *testing.T) {
	testCases := map[string]struct {
		Args     []string
		Stdin    []string
		WantExit int
	}{
		"renumbers a file argument": {
			Args:     []string{simplePath},
			WantExit: 0,
		},
		"renumbers stdin": {
			Stdin:    []string{string(sugar.Must(os.ReadFile(simplePath)))},
			WantExit: 0,
		},
		"too many arguments": {
			Args:     []string{simplePath, simplePath},
			WantExit: 1,
		},
		"inplace with stdin": {
			Args:     []string{"-inplace"},
			WantExit: 1,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout(testCase.Stdin...)
			exit := MainCommandByArgs(testCase.Args, spy.NewProcInout())
			if exit != testCase.WantExit {
				t.Fatalf("exit = %d, want %d\nstderr: %s", exit, testCase.WantExit, spy.Stderr.String())
			}
		})
	}
}

func TestMainCommandByArgs_Inplace(t *testing.T) {

	src := sugar.Must(os.ReadFile("../../../pfd/pfdencoding/pfddrawio/testdata/sequential_without_id.drawio"))
	tmpPath := filepath.Join(t.TempDir(), "pfd.drawio")
	if err := os.WriteFile(tmpPath, src, 0644); err != nil {
		t.Fatal(err)
	}

	spy := cli.SpyProcInout()
	exit := MainCommandByArgs([]string{"-inplace", tmpPath}, spy.NewProcInout())
	if exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
	}
	if spy.Stdout.Len() != 0 {
		t.Errorf("stdout should be empty in -inplace mode, got %q", spy.Stdout.String())
	}

	stdoutSpy := cli.SpyProcInout()
	if exit := MainCommandByArgs([]string{tmpPath}, stdoutSpy.NewProcInout()); exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", exit, stdoutSpy.Stderr.String())
	}

	got := sugar.Must(os.ReadFile(tmpPath))
	if string(got) != stdoutSpy.Stdout.String() {
		t.Errorf("in-place result differs from the stdout result:\n%s", cmp.Diff(stdoutSpy.Stdout.String(), string(got)))
	}
	if string(got) == string(src) {
		t.Error("the target file is unchanged: renumbering did not write back")
	}
}

func TestMainCommandByArgs_InplaceKeepsFileOnError(t *testing.T) {
	tmpPath := filepath.Join(t.TempDir(), "pfd.drawio")
	if err := os.WriteFile(tmpPath, []byte(brokenDrawio), 0644); err != nil {
		t.Fatal(err)
	}

	spy := cli.SpyProcInout()
	exit := MainCommandByArgs([]string{"-inplace", tmpPath}, spy.NewProcInout())
	if exit != 1 {
		t.Fatalf("exit = %d, want 1\nstderr: %s", exit, spy.Stderr.String())
	}

	got := sugar.Must(os.ReadFile(tmpPath))
	if string(got) != brokenDrawio {
		t.Errorf("the target file must not change on error:\n%s", cmp.Diff(brokenDrawio, string(got)))
	}
}

func TestMainCommandByArgs_PNGInput(t *testing.T) {
	testCases := map[string]struct {
		SourcePath string
		Args       func(pngPath string) []string
		WantPath   string
	}{
		"deciding and applying a plan returns a PNG": {
			SourcePath: "../../../pfd/pfdencoding/pfddrawio/testdata/sequential_without_id.drawio",
			Args:       func(pngPath string) []string { return []string{pngPath} },
			WantPath:   "../../../pfd/pfdencoding/pfddrawio/testdata/sequential_with_id.drawio",
		},
		"applying an existing plan returns a PNG": {
			SourcePath: "testdata/split/detail.drawio",
			Args: func(pngPath string) []string {
				return []string{"-renum-plan", "testdata/split/renum.tsv", pngPath}
			},
			WantPath: "testdata/split/detail_applied.drawio",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			pngPath := pngtest.WriteTempPNG(t, tc.SourcePath)

			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs(tc.Args(pngPath), spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Fatalf("exitStatus = %d, want 0", exitStatus)
			}

			want, err := os.ReadFile(tc.WantPath)
			if err != nil {
				t.Fatal(err)
			}

			got := pngtest.MxfileInPNG(t, spy.Stdout.Bytes())
			if !strings.Contains(got, strings.TrimSpace(extractRoot(string(want)))) {
				t.Errorf("renumbered XML does not match %s\n--- got ---\n%s\n--- want (excerpt) ---\n%s",
					tc.WantPath, got, strings.TrimSpace(extractRoot(string(want))))
			}
		})
	}
}

func extractRoot(xml string) string {
	start := strings.Index(xml, "<root>")
	end := strings.Index(xml, "</root>")
	if start < 0 || end < 0 {
		return xml
	}
	return xml[start : end+len("</root>")]
}

func TestMainCommandByArgs_OutputTSV(t *testing.T) {
	testCases := map[string]struct {
		InputPath string
		AsPNG     bool
		WantPath  string
		Want      string
	}{
		"merged pfd yields a renumber plan": {
			InputPath: "testdata/split/merged.drawio",
			WantPath:  "testdata/split/renum.tsv",
		},
		"merged pfd as a PNG yields the same plan": {
			InputPath: "testdata/split/merged.drawio",
			AsPNG:     true,
			WantPath:  "testdata/split/renum.tsv",
		},
		"fully numbered pfd yields a header-only plan": {
			InputPath: "testdata/split/merged_applied.drawio",
			Want:      "Key\tID\n",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			inputPath := tc.InputPath
			if tc.AsPNG {
				inputPath = pngtest.WriteTempPNG(t, tc.InputPath)
			}

			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-out-format", "tsv", inputPath}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Fatalf("exitStatus = %d, want 0", exitStatus)
			}

			want := tc.Want
			if tc.WantPath != "" {
				bs, err := os.ReadFile(tc.WantPath)
				if err != nil {
					t.Fatal(err)
				}
				want = string(bs)
			}
			if got := spy.Stdout.String(); got != want {
				t.Errorf("got:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

func TestMainCommandByArgs_SameLabelProcessAndDeliverable(t *testing.T) {
	const samelabelPath = "testdata/samelabel/pfd.drawio"

	t.Run("the plan numbers them separately", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-out-format", "tsv", samelabelPath}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Fatalf("exitStatus = %d, want 0", exitStatus)
		}

		want := "Key\tID\n(hogeの実装)\tP1\n[hogeの実装]\tD1\n[要件]\tD2\n"
		if got := spy.Stdout.String(); got != want {
			t.Errorf("the plan differs:\n%s", cmp.Diff(want, got))
		}
	})

	t.Run("renumbering writes both IDs", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{samelabelPath}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Fatalf("exitStatus = %d, want 0", exitStatus)
		}

		got := spy.Stdout.String()
		for _, want := range []string{`value="P1: hogeの実装"`, `value="D1: hogeの実装"`, `value="D2: 要件"`} {
			if !strings.Contains(got, want) {
				t.Errorf("the renumbered XML does not contain %s:\n%s", want, got)
			}
		}
	})
}

func TestMainCommandByArgs_DecidePlan(t *testing.T) {
	testCases := map[string]struct {
		Base  string
		MinID string
		Want  string
	}{
		"no base and no floor numbers from the input": {
			Want: "Key\tID\n(実装する)\tP3\n(設計する)\tP2\n[設計書]\tD3\n",
		},

		"the base plan raises the numbering floor": {
			Base: "Key\tID\n(テストを書く)\tP9\n",
			Want: "Key\tID\n(テストを書く)\tP9\n(実装する)\tP11\n(設計する)\tP10\n[設計書]\tD3\n",
		},
		"a key in the base plan keeps its ID and consumes no new number": {
			Base: "Key\tID\n(設計する)\tP7\n",
			Want: "Key\tID\n(実装する)\tP8\n(設計する)\tP7\n[設計書]\tD3\n",
		},

		"extending with its own output is idempotent": {
			Base: "Key\tID\n(実装する)\tP3\n(設計する)\tP2\n[設計書]\tD3\n",
			Want: "Key\tID\n(実装する)\tP3\n(設計する)\tP2\n[設計書]\tD3\n",
		},
		"an empty base plan yields the same plan as no base": {
			Base: "Key\tID\n",
			Want: "Key\tID\n(実装する)\tP3\n(設計する)\tP2\n[設計書]\tD3\n",
		},

		"a key absent from the input is carried over": {
			Base: "Key\tID\n(他の作業)\tP20\n",
			Want: "Key\tID\n(他の作業)\tP20\n(実装する)\tP22\n(設計する)\tP21\n[設計書]\tD3\n",
		},
		"the floor raises both counters": {
			MinID: "P10\tD20",
			Want:  "Key\tID\n(実装する)\tP12\n(設計する)\tP11\n[設計書]\tD21\n",
		},
		"a floor lower than the input is ignored": {
			MinID: "P0\tD0",
			Want:  "Key\tID\n(実装する)\tP3\n(設計する)\tP2\n[設計書]\tD3\n",
		},
		"only the given kind is raised": {
			MinID: "D20",
			Want:  "Key\tID\n(実装する)\tP3\n(設計する)\tP2\n[設計書]\tD21\n",
		},
		"IDs can be separated by a comma or a space": {
			MinID: "P10, D20",
			Want:  "Key\tID\n(実装する)\tP12\n(設計する)\tP11\n[設計書]\tD21\n",
		},

		"a multi-line floor takes the max of each kind": {
			MinID: "P10\tD0\nP0\tD20\n",
			Want:  "Key\tID\n(実装する)\tP12\n(設計する)\tP11\n[設計書]\tD21\n",
		},

		"the base plan and the floor are both honored": {
			Base:  "Key\tID\n(他の作業)\tP20\n",
			MinID: "P10\tD20",
			Want:  "Key\tID\n(他の作業)\tP20\n(実装する)\tP22\n(設計する)\tP21\n[設計書]\tD21\n",
		},
		"the floor wins over a lower base plan": {
			Base:  "Key\tID\n(他の作業)\tP5\n",
			MinID: "P30",
			Want:  "Key\tID\n(他の作業)\tP5\n(実装する)\tP32\n(設計する)\tP31\n[設計書]\tD3\n",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			args := []string{"-out-format", "tsv"}
			if tc.Base != "" {
				planPath := filepath.Join(t.TempDir(), "renum.tsv")
				if err := os.WriteFile(planPath, []byte(tc.Base), 0644); err != nil {
					t.Fatal(err)
				}
				args = append(args, "-renum-plan", planPath)
			}
			if tc.MinID != "" {
				args = append(args, "-min-id", tc.MinID)
			}
			args = append(args, "testdata/split/merged.drawio")

			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs(args, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Fatalf("exitStatus = %d, want 0", exitStatus)
			}

			if got := spy.Stdout.String(); got != tc.Want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tc.Want)
			}
		})
	}
}

func TestMainCommandByArgs_ExtendRenumberPlanConflict(t *testing.T) {

	input := string(sugar.Must(os.ReadFile("testdata/variants/parent.drawio"))) +
		strings.Replace(
			string(sugar.Must(os.ReadFile("testdata/variants/detail_a.drawio"))),
			`value="実装する"`, `value="P2: 実装する"`, 1)

	planPath := filepath.Join(t.TempDir(), "renum.tsv")
	if err := os.WriteFile(planPath, []byte("Key\tID\n(設計する)\tP2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	spy := cli.SpyProcInout(input)
	exitStatus := MainCommandByArgs([]string{"-out-format", "tsv", "-renum-plan", planPath}, spy.NewProcInout())
	if exitStatus == 0 {
		t.Errorf("exitStatus = 0, want non-zero (stdout: %q)", spy.Stdout.String())
	}
}

func TestMainCommandByArgs_ExtendRenumberPlanTypeMismatch(t *testing.T) {
	testCases := map[string]struct {
		Base string
	}{
		"a deliverable ID for a process": {
			Base: "Key\tID\n(設計する)\tD9\n",
		},
		"a process ID for a deliverable": {
			Base: "Key\tID\n[設計書]\tP9\n",
		},
		"a deliverable ID for a process absent from the input": {
			Base: "Key\tID\n(他の作業)\tD9\n",
		},

		"a legacy key without a kind": {
			Base: "Key\tID\n設計する\tP2\n",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			planPath := filepath.Join(t.TempDir(), "renum.tsv")
			if err := os.WriteFile(planPath, []byte(tc.Base), 0644); err != nil {
				t.Fatal(err)
			}

			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-out-format", "tsv", "-renum-plan", planPath, "testdata/split/merged.drawio"}, spy.NewProcInout())
			if exitStatus == 0 {
				t.Errorf("exitStatus = 0, want non-zero (stdout: %q)", spy.Stdout.String())
			}
		})
	}
}

func TestMainCommandByArgs_ExtendRenumberPlanUncheckedKey(t *testing.T) {
	planPath := filepath.Join(t.TempDir(), "renum.tsv")

	base := "Key\tID\n(他の作業)\tP1\n"
	if err := os.WriteFile(planPath, []byte(base), 0644); err != nil {
		t.Fatal(err)
	}

	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs([]string{"-out-format", "tsv", "-renum-plan", planPath, "testdata/split/merged.drawio"}, spy.NewProcInout())
	if exitStatus != 0 {
		t.Log(spy.Stderr.String())
		t.Fatalf("exitStatus = %d, want 0", exitStatus)
	}

	want := "Key\tID\n(他の作業)\tP1\n(実装する)\tP3\n(設計する)\tP2\n[設計書]\tD3\n"
	if got := spy.Stdout.String(); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestMainCommandByArgs_VariantsUniqueNumbering(t *testing.T) {
	parent := string(sugar.Must(os.ReadFile("testdata/variants/parent.drawio")))
	variants := []string{
		parent + string(sugar.Must(os.ReadFile("testdata/variants/detail_a.drawio"))),
		parent + string(sugar.Must(os.ReadFile("testdata/variants/detail_b.drawio"))),
	}

	planPath := filepath.Join(t.TempDir(), "renum.tsv")
	if err := os.WriteFile(planPath, []byte("Key\tID\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ledger := ""
	for _, variant := range variants {
		spy := cli.SpyProcInout(variant)
		exitStatus := MainCommandByArgs([]string{"-out-format", "tsv", "-renum-plan", planPath}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Fatalf("exitStatus = %d, want 0", exitStatus)
		}

		ledger = spy.Stdout.String()
		if err := os.WriteFile(planPath, []byte(ledger), 0644); err != nil {
			t.Fatal(err)
		}
	}

	want := "Key\tID\n(実装する)\tP3\n(設計する)\tP2\n(調査する)\tP4\n[設計書]\tD3\n[調査結果]\tD4\n"
	if ledger != want {
		t.Errorf("the ledger differs:\n%s", cmp.Diff(want, ledger))
	}

	plan, err := pfdtsv.ParseRenumberPlan(strings.NewReader(ledger))
	if err != nil {
		t.Fatalf("ParseRenumberPlan: %v\nledger:\n%s", err, ledger)
	}
	keyByID := map[pfd.NodeID]pfd.NodeID{}
	for key, node := range plan {
		if dup, ok := keyByID[node.ID]; ok {
			t.Errorf("ID %q is assigned to both %q and %q\nledger:\n%s", node.ID, dup, key, ledger)
		}
		keyByID[node.ID] = key
	}

	for _, variant := range variants {
		spy := cli.SpyProcInout(variant)
		exitStatus := MainCommandByArgs([]string{"-out-format", "tsv", "-renum-plan", planPath}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Fatalf("exitStatus = %d, want 0", exitStatus)
		}
		if got := spy.Stdout.String(); got != ledger {
			t.Errorf("the ledger is not stable:\n%s", cmp.Diff(ledger, got))
		}
	}
}

func TestMainCommandByArgs_MergedVariantsAreRejected(t *testing.T) {
	merged := string(sugar.Must(os.ReadFile("testdata/variants/parent.drawio"))) +
		string(sugar.Must(os.ReadFile("testdata/variants/detail_a.drawio"))) +
		string(sugar.Must(os.ReadFile("testdata/variants/detail_b.drawio")))

	spy := cli.SpyProcInout(merged)
	exitStatus := MainCommandByArgs([]string{"-out-format", "tsv"}, spy.NewProcInout())
	if exitStatus == 0 {
		t.Errorf("exitStatus = 0, want non-zero (stdout: %q)", spy.Stdout.String())
	}
	if want := "のページが複数あります"; !strings.Contains(spy.Stderr.String(), want) {
		t.Errorf("stderr does not mention the duplicated detail page name (%q):\n%s", want, spy.Stderr.String())
	}
}

func TestMainCommandByArgs_OutputMaxID(t *testing.T) {
	testCases := map[string]struct {
		InputPath string
		AsPNG     bool
		Stdin     []string
		Want      string
	}{
		"a partially numbered pfd yields the max of each kind": {
			InputPath: "testdata/split/merged.drawio",
			Want:      "P1\tD2\n",
		},
		"a fully numbered pfd yields the max after renumbering": {
			InputPath: "testdata/split/merged_applied.drawio",
			Want:      "P3\tD3\n",
		},
		"an unnumbered pfd yields P0 and D0": {
			InputPath: "../../../pfd/pfdencoding/pfddrawio/testdata/sequential_without_id.drawio",
			Want:      "P0\tD0\n",
		},

		"duplicated detail page names are fine": {
			Stdin: []string{
				string(sugar.Must(os.ReadFile("testdata/split/merged.drawio"))) +
					string(sugar.Must(os.ReadFile("testdata/split/detail_applied.drawio"))),
			},
			Want: "P3\tD3\n",
		},

		"a detail page alone is fine": {
			InputPath: "testdata/split/detail_applied.drawio",
			Want:      "P3\tD3\n",
		},
		"a PNG yields the same max as the drawio": {
			InputPath: "testdata/split/merged.drawio",
			AsPNG:     true,
			Want:      "P1\tD2\n",
		},

		"an ID on the comment layer raises the max": {
			Stdin: []string{strings.Replace(
				string(sugar.Must(os.ReadFile("testdata/split/merged.drawio"))),
				`<mxCell id="1" value="PFD" parent="0"></mxCell>`,
				`<mxCell id="1" value="PFD" parent="0"></mxCell>`+
					`<mxCell id="cl" value="コメント" parent="0"></mxCell>`+
					`<mxCell id="c1" value="P99: 旧案" style="ellipse;whiteSpace=wrap;html=1;" parent="cl" vertex="1"><mxGeometry x="0" y="0" width="1" height="1" as="geometry"></mxGeometry></mxCell>`,
				1)},
			Want: "P99\tD2\n",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			args := []string{"-out-format", "maxid"}
			if tc.InputPath != "" {
				inputPath := tc.InputPath
				if tc.AsPNG {
					inputPath = pngtest.WriteTempPNG(t, tc.InputPath)
				}
				args = append(args, inputPath)
			}

			spy := cli.SpyProcInout(tc.Stdin...)
			exitStatus := MainCommandByArgs(args, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Fatalf("exitStatus = %d, want 0", exitStatus)
			}

			if got := spy.Stdout.String(); got != tc.Want {
				t.Errorf("got %q, want %q", got, tc.Want)
			}
		})
	}
}

func TestMainCommandByArgs_ApplyRenumberPlan(t *testing.T) {
	testCases := map[string]struct {
		InputPath string
		WantPath  string
	}{

		"detail page alone can be renumbered by a plan from the merged pfd": {
			InputPath: "testdata/split/detail.drawio",
			WantPath:  "testdata/split/detail_applied.drawio",
		},
		"applying a plan twice is idempotent": {
			InputPath: "testdata/split/detail_applied.drawio",
			WantPath:  "testdata/split/detail_applied.drawio",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-renum-plan", "testdata/split/renum.tsv", tc.InputPath}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Fatalf("exitStatus = %d, want 0", exitStatus)
			}

			want, err := os.ReadFile(tc.WantPath)
			if err != nil {
				t.Fatal(err)
			}
			if got := spy.Stdout.String(); got != string(want) {
				t.Errorf("got:\n%s\nwant:\n%s", got, string(want))
			}
		})
	}
}

func TestMainCommandByArgs_InvalidOptions(t *testing.T) {
	testCases := map[string]struct {
		Args []string
	}{
		"-o tsv with -inplace": {
			Args: []string{"-out-format", "tsv", "-inplace", "testdata/split/merged.drawio"},
		},
		"-o tsv with -renum-plan and -inplace": {
			Args: []string{"-out-format", "tsv", "-renum-plan", "testdata/split/renum.tsv", "-inplace", "testdata/split/merged.drawio"},
		},
		"missing renumber plan file with -o tsv": {
			Args: []string{"-out-format", "tsv", "-renum-plan", "testdata/split/nonexistent.tsv", "testdata/split/merged.drawio"},
		},
		"-min-id with an unparsable ID": {
			Args: []string{"-out-format", "tsv", "-min-id", "X1", "testdata/split/merged.drawio"},
		},
		"-min-id with -renum-plan": {
			Args: []string{"-min-id", "P12", "-renum-plan", "testdata/split/renum.tsv", "testdata/split/detail.drawio"},
		},
		"unknown output format": {
			Args: []string{"-out-format", "json", "testdata/split/merged.drawio"},
		},
		"missing renumber plan file": {
			Args: []string{"-renum-plan", "testdata/split/nonexistent.tsv", "testdata/split/merged.drawio"},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs(tc.Args, spy.NewProcInout())
			if exitStatus == 0 {
				t.Errorf("exitStatus = 0, want non-zero (stdout: %q)", spy.Stdout.String())
			}
		})
	}
}

func TestMainCommandByArgs_InplaceKeepsTheFileOnBrokenPlan(t *testing.T) {
	original, err := os.ReadFile("testdata/split/detail.drawio")
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	drawioPath := filepath.Join(dir, "detail.drawio")
	if err := os.WriteFile(drawioPath, original, 0644); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(dir, "broken.tsv")
	if err := os.WriteFile(planPath, []byte("設計する\tP2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs([]string{"-renum-plan", planPath, "-inplace", drawioPath}, spy.NewProcInout())
	if exitStatus == 0 {
		t.Fatal("exitStatus = 0, want non-zero")
	}

	got, err := os.ReadFile(drawioPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, got) {
		t.Errorf("the drawio has been damaged: %d bytes, want %d bytes", len(got), len(original))
	}
}

func TestMainCommandByArgs_InplaceApplyRenumberPlan(t *testing.T) {
	original, err := os.ReadFile("testdata/split/detail.drawio")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/split/detail_applied.drawio")
	if err != nil {
		t.Fatal(err)
	}

	drawioPath := filepath.Join(t.TempDir(), "detail.drawio")
	if err := os.WriteFile(drawioPath, original, 0644); err != nil {
		t.Fatal(err)
	}

	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs([]string{"-renum-plan", "testdata/split/renum.tsv", "-inplace", drawioPath}, spy.NewProcInout())
	if exitStatus != 0 {
		t.Log(spy.Stderr.String())
		t.Fatalf("exitStatus = %d, want 0", exitStatus)
	}

	got, err := os.ReadFile(drawioPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("got:\n%s\nwant:\n%s", string(got), string(want))
	}
}

func TestMainCommandByOptions_EmitPlanWithInplace(t *testing.T) {
	tmpPath := filepath.Join(t.TempDir(), "pfd.drawio")
	src := sugar.Must(os.ReadFile(simplePath))
	if err := os.WriteFile(tmpPath, src, 0644); err != nil {
		t.Fatal(err)
	}

	spy := cli.SpyProcInout()
	err := MainCommandByOptions(&Options{
		CommonOptions: &tools.CommonOptions{},
		Reader:        bytes.NewReader(src),
		Mode:          ModeEmitPlan,
		Inplace:       true,
		InputFilePath: tmpPath,
	}, spy.NewProcInout())
	if err == nil {
		t.Fatal("err = nil, want an error")
	}

	got := sugar.Must(os.ReadFile(tmpPath))
	if string(got) != string(src) {
		t.Errorf("the target file must not change:\n%s", cmp.Diff(string(src), string(got)))
	}
}
