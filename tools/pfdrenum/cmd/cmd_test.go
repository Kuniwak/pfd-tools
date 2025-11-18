package cmd

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
)

func TestMainCommandByArgs(t *testing.T) {
	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs([]string{"testdata/simple/pfd.drawio"}, spy.NewProcInout())
	if exitStatus != 0 {
		t.Log(spy.Stderr.String())
		t.Log(spy.Stdout.String())
		t.Errorf("exitStatus = %d, want 0", exitStatus)
	}
}
