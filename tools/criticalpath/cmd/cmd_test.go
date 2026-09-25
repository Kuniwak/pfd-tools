package cmd

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/locale"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable/encoding/pfdtsv"
)

func TestHeaders(t *testing.T) {
	cases := map[string]struct {
		locale locale.Locale
		want   []string
	}{
		"ja": {locale.LocaleJa, []string{"ID", "最大弾性値（全余裕）", "最小弾性値"}},
		"en": {locale.LocaleEn, []string{"ID", "Maximum Elasticity (Total Float)", "Minimum Elasticity"}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := Headers(c.locale)
			if !slices.Equal(got, c.want) {
				t.Errorf("Headers(%q) = %v, want %v", c.locale, got, c.want)
			}
		})
	}
}

func TestCmd_Locale(t *testing.T) {
	cases := map[string]struct {
		locale string
		want   string
	}{
		"ja": {"ja", "ID\t最大弾性値（全余裕）\t最小弾性値"},
		"en": {"en", "ID\tMaximum Elasticity (Total Float)\tMinimum Elasticity"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-poor", "-f", "testdata/simple/config.json", "-locale", c.locale}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Fatalf("exitStatus = %d, want 0", exitStatus)
			}
			got, _, _ := strings.Cut(spy.Stdout.String(), "\n")
			if got != c.want {
				t.Errorf("header line = %q, want %q", got, c.want)
			}
		})
	}
}

func TestCmd(t *testing.T) {
	testCases := map[string]struct {
		Args     []string
		WantExit int
	}{
		"config": {
			Args:     []string{"-poor", "-f", "testdata/simple/config.json"},
			WantExit: 0,
		},

		"without composite deliverable table": {
			Args: []string{
				"-poor",
				"-p", "testdata/simple/pfd.drawio",
				"-ap", "testdata/simple/atomic_proc.tsv",
				"-ad", "testdata/simple/deliv.tsv",
				"-r", "testdata/simple/resource.tsv",
				"-res", "finite",
				"-fb", "enabled",
			},
			WantExit: 0,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs(tc.Args, spy.NewProcInout())
			if exitStatus != tc.WantExit {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want %d", exitStatus, tc.WantExit)
			}
		})
	}
}

func TestCmd_PNGInput(t *testing.T) {
	pngPath := pngtest.WriteTempPNG(t, "testdata/simple/pfd.drawio")
	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs([]string{"-poor", "-f", "testdata/simple/config.json", "-p", pngPath}, spy.NewProcInout())
	if exitStatus != 0 {
		t.Log(spy.Stderr.String())
		t.Log(spy.Stdout.String())
		t.Errorf("exitStatus = %d, want 0", exitStatus)
	}
}

func TestCmd_Profile(t *testing.T) {
	dir := t.TempDir()
	cpuPath := filepath.Join(dir, "cpu.pprof")
	memPath := filepath.Join(dir, "mem.pprof")
	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs([]string{"-poor", "-f", "testdata/simple/config.json", "-cpuprofile", cpuPath, "-memprofile", memPath}, spy.NewProcInout())
	if exitStatus != 0 {
		t.Log(spy.Stderr.String())
		t.Fatalf("exitStatus = %d, want 0", exitStatus)
	}
	for _, p := range []string{cpuPath, memPath} {
		info, err := os.Stat(p)
		if err != nil {
			t.Fatalf("stat %q: %v", p, err)
		}
		if info.Size() == 0 {
			t.Fatalf("profile %q is empty", p)
		}
	}
}

func TestCmd_ExecModelLarge(t *testing.T) {
	searchOptions := []string{"-poor", "-poorest"}
	testCases := map[string]struct {
		Config               string
		AtomicProcessTable   string
		ExpectedProcessCount int
	}{
		"finite resources with feedback edges": {
			Config:               "testdata/large/project.json",
			AtomicProcessTable:   "testdata/large/ap.tsv",
			ExpectedProcessCount: 22,
		},
		"infinite resources with feedback edges": {
			Config:               "testdata/large/project_infinite.json",
			AtomicProcessTable:   "testdata/large/ap_infinite.tsv",
			ExpectedProcessCount: 22,
		},
		"finite resources without feedback edges": {
			Config:               "testdata/large_nofb/project_finite.json",
			AtomicProcessTable:   "testdata/large_nofb/ap_finite.tsv",
			ExpectedProcessCount: 22,
		},
		"infinite resources without feedback edges": {
			Config:               "testdata/large_nofb/project_infinite.json",
			AtomicProcessTable:   "testdata/large_nofb/ap_infinite.tsv",
			ExpectedProcessCount: 22,
		},
	}

	for name, tc := range testCases {
		for _, searchOption := range searchOptions {
			t.Run(name+"/"+searchOption, func(t *testing.T) {
				spy := cli.SpyProcInout()
				args := []string{searchOption, "-random-seed", "0", "-f", tc.Config}
				if exitStatus := MainCommandByArgs(args, spy.NewProcInout()); exitStatus != 0 {
					t.Log(spy.Stderr.String())
					t.Fatalf("exitStatus = %d, want 0", exitStatus)
				}

				r := csv.NewReader(strings.NewReader(spy.Stdout.String()))
				r.Comma = '\t'
				r.FieldsPerRecord = -1
				rows, err := r.ReadAll()
				if err != nil {
					t.Fatalf("read the output as tsv: %v", err)
				}
				if len(rows) == 0 {
					t.Fatal("expected at least the header row, but got none")
				}
				if want := Headers(locale.LocaleJa); !slices.Equal(rows[0], want) {
					t.Errorf("header row = %v, want %v", rows[0], want)
				}

				actual := make([]string, 0, len(rows)-1)
				for _, row := range rows[1:] {
					actual = append(actual, row[0])
				}
				slices.Sort(actual)

				expected := sortedAtomicProcessIDs(t, tc.AtomicProcessTable)

				if len(expected) != tc.ExpectedProcessCount {
					t.Fatalf("expected %d atomic processes in %q, but got %d", tc.ExpectedProcessCount, tc.AtomicProcessTable, len(expected))
				}
				if !slices.Equal(actual, expected) {
					t.Errorf("expected a row for every atomic process\ngot:      %v\nexpected: %v", actual, expected)
				}
			})
		}
	}
}

func sortedAtomicProcessIDs(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %q: %v", path, err)
	}
	defer f.Close()

	table, err := pfdtsv.ParseAtomicProcessTable(f)
	if err != nil {
		t.Fatalf("parse %q: %v", path, err)
	}

	res := make([]string, 0, len(table.Rows))
	for _, id := range table.IDs() {
		res = append(res, string(id))
	}
	slices.Sort(res)
	return res
}
