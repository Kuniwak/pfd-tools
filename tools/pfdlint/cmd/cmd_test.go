package cmd

import (
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
)

func TestMainCommandByArgs(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})

	t.Run("composite process table loaded from config", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/loop/config.json"}, spy.NewProcInout())
		// loop PFD has no composite process but cp table has P0, so lint should report extra-cp-table
		if exitStatus != 1 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 1", exitStatus)
		}
		if !strings.Contains(spy.Stdout.String(), "extra-cp-table") {
			t.Errorf("expected extra-cp-table lint error, got: %s", spy.Stdout.String())
		}
	})

	t.Run("invalid", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/invalid/config.json"}, spy.NewProcInout())
		if exitStatus != 1 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 1", exitStatus)
		}
	})

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
