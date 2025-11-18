package cmd

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
)

func TestMainCommandByArgs(t *testing.T) {
	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs([]string{"-p", "testdata/loop/plan.json", "-ap", "testdata/loop/atomic_proc.tsv", "-m", "testdata/loop/milestone.tsv", "-g", "testdata/loop/group.tsv", "-b", "1.5"}, spy.NewProcInout())
	if exitStatus != 0 {
		t.Log(spy.Stderr.String())
		t.Log(spy.Stdout.String())
		t.Errorf("exitStatus = %d, want 0", exitStatus)
	}
}
