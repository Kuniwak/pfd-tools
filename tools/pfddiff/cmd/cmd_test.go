package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
)

func TestMainCommandByArgs(t *testing.T) {
	testCases := map[string]struct {
		Args     []string
		WantExit int
	}{
		"same": {
			Args: []string{
				"-p1", "testdata/loop/pfd.drawio", "-cd1", "testdata/loop/comp_deliv.tsv",
				"-p2", "testdata/loop/pfd.drawio", "-cd2", "testdata/loop/comp_deliv.tsv",
			},
			WantExit: 0,
		},

		"same without composite deliverable tables": {
			Args: []string{
				"-p1", "testdata/loop/pfd.drawio",
				"-p2", "testdata/loop/pfd.drawio",
			},
			WantExit: 0,
		},
		"different": {
			Args: []string{
				"-p1", "testdata/loop/pfd.drawio", "-cd1", "testdata/loop/comp_deliv.tsv",
				"-p2", "testdata/longloop/pfd.drawio", "-cd2", "testdata/longloop/comp_deliv.tsv",
			},
			WantExit: 1,
		},
		"prompt": {
			Args: []string{
				"-p1", "testdata/loop/pfd.drawio", "-cd1", "testdata/loop/comp_deliv.tsv",
				"-p2", "testdata/longloop/pfd.drawio", "-cd2", "testdata/longloop/comp_deliv.tsv",
				"-prompt",
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

	t.Run("absolute paths", func(t *testing.T) {
		pfdAbs, err := filepath.Abs("testdata/loop/pfd.drawio")
		if err != nil {
			t.Fatal(err)
		}
		cdAbs, err := filepath.Abs("testdata/loop/comp_deliv.tsv")
		if err != nil {
			t.Fatal(err)
		}
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{
			"-p1", pfdAbs, "-cd1", cdAbs,
			"-p2", pfdAbs, "-cd2", cdAbs,
		}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})

	t.Run("png input", func(t *testing.T) {
		xml, err := os.ReadFile("testdata/loop/pfd.drawio")
		if err != nil {
			t.Fatal(err)
		}
		tmpDir := t.TempDir()
		pngPath1 := filepath.Join(tmpDir, "a.drawio.png")
		pngPath2 := filepath.Join(tmpDir, "b.drawio.png")
		if err := os.WriteFile(pngPath1, pngtest.WrapMxfile(t, xml), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(pngPath2, pngtest.WrapMxfile(t, xml), 0644); err != nil {
			t.Fatal(err)
		}
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{
			"-p1", pngPath1, "-cd1", "testdata/loop/comp_deliv.tsv",
			"-p2", pngPath2, "-cd2", "testdata/loop/comp_deliv.tsv",
		}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
}

func TestMainCommandByArgs_Fragment(t *testing.T) {

	fragment := func(processLabel string) string {
		return `<mxfile host="x"><diagram id="d1" name="P1: 設計する"><mxGraphModel><root>` +
			`<mxCell id="0"/><mxCell id="1" parent="0"/>` +
			`<mxCell id="2" value="D1: 要求" style="rounded=0;whiteSpace=wrap;html=1;strokeWidth=1;" parent="1" vertex="1"><mxGeometry as="geometry"/></mxCell>` +
			`<mxCell id="3" value="` + processLabel + `" style="ellipse;whiteSpace=wrap;html=1;strokeWidth=1;" parent="1" vertex="1"><mxGeometry as="geometry"/></mxCell>` +
			`<mxCell id="4" style="edgeStyle=none;html=1;" parent="1" source="2" target="3" edge="1"><mxGeometry as="geometry"/></mxCell>` +
			`</root></mxGraphModel></diagram></mxfile>`
	}

	testCases := map[string]struct {
		Fragment2  string
		Fragment   bool
		WantExit   int
		WantStderr string
	}{

		"a fragment is rejected without -fragment": {
			Fragment2:  "P2: 方式を決める",
			Fragment:   false,
			WantExit:   1,
			WantStderr: "ページ名",
		},
		"same fragments with -fragment": {
			Fragment2: "P2: 方式を決める",
			Fragment:  true,
			WantExit:  0,
		},
		"different fragments with -fragment": {
			Fragment2: "P2: 方式を選ぶ",
			Fragment:  true,
			WantExit:  1,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path1 := filepath.Join(dir, "a.drawio")
			path2 := filepath.Join(dir, "b.drawio")
			if err := os.WriteFile(path1, []byte(fragment("P2: 方式を決める")), 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path2, []byte(fragment(tc.Fragment2)), 0644); err != nil {
				t.Fatal(err)
			}

			args := []string{"-p1", path1, "-p2", path2}
			if tc.Fragment {
				args = append(args, "-fragment")
			}

			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs(args, spy.NewProcInout())
			if exitStatus != tc.WantExit {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want %d", exitStatus, tc.WantExit)
			}
			if tc.WantStderr != "" && !strings.Contains(spy.Stderr.String(), tc.WantStderr) {
				t.Errorf("stderr does not contain %q:\n%s", tc.WantStderr, spy.Stderr.String())
			}
		})
	}
}
