package cmd

import (
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
)

func TestMainCommandByArgs(t *testing.T) {
	testCases := map[string]struct {
		Args               []string
		WantExit           int
		WantStdoutContains string
		WantStderrContains string
	}{
		"valid": {
			Args:     []string{"-f", "testdata/simple/config.json"},
			WantExit: 0,
		},

		"large: finite resources with feedback edges": {
			Args:     []string{"-f", "testdata/large/project.json"},
			WantExit: 0,
		},
		"large: infinite resources with feedback edges": {
			Args:     []string{"-f", "testdata/large/project_infinite.json"},
			WantExit: 0,
		},
		"large: finite resources without feedback edges": {
			Args:     []string{"-f", "testdata/large_nofb/project_finite.json"},
			WantExit: 0,
		},
		"large: infinite resources without feedback edges": {
			Args:     []string{"-f", "testdata/large_nofb/project_infinite.json"},
			WantExit: 0,
		},
		"composite process table loaded from config": {

			Args:               []string{"-f", "testdata/loop/config.json"},
			WantExit:           1,
			WantStdoutContains: "extra-cp-table",
		},
		"composite process without detail page": {

			Args:               []string{"-f", "testdata/implicit_atomic_composite/config.json"},
			WantExit:           1,
			WantStdoutContains: "implicit-atomic-composite",
		},
		"without composite deliverable table": {

			Args: []string{
				"-p", "testdata/simple/pfd.drawio",
				"-ap", "testdata/simple/atomic_proc.tsv",
				"-ad", "testdata/simple/deliv.tsv",
				"-r", "testdata/simple/resource.tsv",
				"-res", "finite",
				"-fb", "enabled",
			},
			WantExit: 0,
		},
		"composite process outputs a composite deliverable": {

			Args:     []string{"-f", "testdata/composite_deliverable_output/config.json"},
			WantExit: 0,
		},
		"composite process takes a composite deliverable as input": {

			Args:     []string{"-f", "testdata/compdeliv_input/config.json"},
			WantExit: 0,
		},
		"composite deliverable used only as a member of another one": {

			Args:     []string{"-f", "testdata/compdeliv_member_only/config.json"},
			WantExit: 0,
		},
		"cyclic composite deliverable nesting": {

			Args:               []string{"-f", "testdata/compdeliv_cycle/config.json"},
			WantExit:           1,
			WantStdoutContains: "acyclic-cd-comp\t複合成果物の入れ子に循環があります。\tCOMPOSITE_DELIVERABLE_TABLE[D3, D6]",
		},
		"nested composite deliverable": {

			Args:     []string{"-f", "testdata/compdeliv_nested/config.json"},
			WantExit: 0,
		},
		"atomic process outputs a composite deliverable": {

			Args:     []string{"-f", "testdata/compdeliv/config.json"},
			WantExit: 0,
		},
		"composite deliverable node without composite deliverable table entry": {

			Args:               []string{"-p", "testdata/loop/pfd.drawio"},
			WantExit:           1,
			WantStdoutContains: "missing-cd-table",
		},
		"invalid": {
			Args:     []string{"-f", "testdata/invalid/config.json"},
			WantExit: 1,
		},
		"deliverable missing from d-table": {

			Args:               []string{"-f", "testdata/missing_d_table/config.json"},
			WantExit:           1,
			WantStdoutContains: "missing-d-table",
		},
		"invalid page name": {

			Args:               []string{"-f", "testdata/invalid_page_name/config.json"},
			WantExit:           1,
			WantStderrContains: "ページ2",
		},
		"same element ID drawn as a box and an ellipse": {

			Args:               []string{"-f", "testdata/duplicate_node_type/config.json"},
			WantExit:           1,
			WantStderrContains: "両方で使われています",
		},
		"compressed drawio": {

			Args:               []string{"-f", "testdata/compressed/config.json"},
			WantExit:           1,
			WantStderrContains: "圧縮",
		},
		"no args": {

			Args:               []string{},
			WantExit:           1,
			WantStderrContains: "PFD",
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
			if tc.WantStdoutContains != "" && !strings.Contains(spy.Stdout.String(), tc.WantStdoutContains) {
				t.Errorf("expected stdout to contain %q, got: %s", tc.WantStdoutContains, spy.Stdout.String())
			}
			if tc.WantStderrContains != "" && !strings.Contains(spy.Stderr.String(), tc.WantStderrContains) {
				t.Errorf("expected stderr to contain %q, got: %s", tc.WantStderrContains, spy.Stderr.String())
			}

		})
	}

	t.Run("png input", func(t *testing.T) {
		pngPath := pngtest.WriteTempPNG(t, "testdata/simple/pfd.drawio")
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-p", pngPath}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
}

func TestParseOptions_NoPFD(t *testing.T) {

	spy := cli.SpyProcInout()
	opts, err := ParseOptions([]string{}, spy.NewProcInout())
	if err == nil {
		t.Fatalf("expected error for missing PFD, got nil (opts = %+v)", opts)
	}
	if opts != nil {
		t.Errorf("expected nil options on error, got %+v", opts)
	}
}
