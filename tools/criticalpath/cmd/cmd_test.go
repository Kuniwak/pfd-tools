package cmd

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
)

func TestCmd(t *testing.T) {
	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs([]string{"-poor", "-f", "testdata/simple/config.json"}, spy.NewProcInout())

	if exitStatus != 0 {
		t.Log(spy.Stderr.String())
		t.Log(spy.Stdout.String())
		t.Errorf("exitStatus = %d, want 0", exitStatus)
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
