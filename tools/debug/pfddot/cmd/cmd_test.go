package cmd

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
)

func TestCmd(t *testing.T) {
	testCases := map[string]struct {
		Args     []string
		WantExit int
	}{
		"with composite deliverable table": {
			Args:     []string{"-p", "testdata/simple/pfd.drawio", "-cd", "testdata/simple/comp_deliv.tsv"},
			WantExit: 0,
		},
		"without composite deliverable table": {
			Args:     []string{"-p", "testdata/simple/pfd.drawio"},
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
