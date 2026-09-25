package cmd

import (
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
)

func TestHelpWritesToInoutStderr(t *testing.T) {
	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs([]string{"-h"}, spy.NewProcInout())
	if exitStatus != 0 {
		t.Errorf("exitStatus = %d, want 0", exitStatus)
	}
	if got := spy.Stderr.String(); !strings.Contains(got, "Usage") {
		t.Errorf("stderr should contain the help text, got %q", got)
	}
}

func TestMainCommandByArgs(t *testing.T) {
	testCases := map[string][]string{
		"atomic process":               {"-p", "testdata/loop/pfd.drawio", "-ap", "testdata/loop/atomic_proc.tsv", "P1"},
		"atomic deliverable":           {"-p", "testdata/loop/pfd.drawio", "-ad", "testdata/loop/deliv.tsv", "D1.1"},
		"composite process":            {"-cp", "testdata/loop/comp_proc.tsv", "P0"},
		"composite process via config": {"-f", "testdata/loop/config.json", "P0"},
		"composite deliverable":        {"-cd", "testdata/loop/comp_deliv.tsv", "D1"},

		"pfd node without composite deliverable table": {"-p", "testdata/loop/pfd.drawio", "P1"},
		"reachable":          {"-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-ap", "testdata/loop/atomic_proc.tsv", "-reachable", "P1"},
		"backward reachable": {"-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-ap", "testdata/loop/atomic_proc.tsv", "-backward-reachable", "P1"},
		"backward reachable feedback destination": {"-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-ap", "testdata/loop/atomic_proc.tsv", "-backward-reachable-fb", "P1"},
		"milestone table":                         {"-m", "testdata/loop/milestone.tsv", "M1"},
		"group table":                             {"-g", "testdata/loop/group.tsv", "G1"},
	}
	for name, args := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs(args, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
	}

	t.Run("png input", func(t *testing.T) {
		pngPath := pngtest.WriteTempPNG(t, "testdata/loop/pfd.drawio")
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-p", pngPath, "-ap", "testdata/loop/atomic_proc.tsv", "P1"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
}
