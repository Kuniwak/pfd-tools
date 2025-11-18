package cmd

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
)

func TestCmd(t *testing.T) {
	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs([]string{"-p", "testdata/simple/pfd.drawio", "-cd", "testdata/simple/comp_deliv.tsv"}, spy.NewProcInout())

	if exitStatus != 0 {
		t.Log(spy.Stderr.String())
		t.Log(spy.Stdout.String())
		t.Errorf("exitStatus = %d, want 0", exitStatus)
	}
}
