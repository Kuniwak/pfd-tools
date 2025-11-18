package cmd

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
)

func TestMainCommandByArgs(t *testing.T) {
	t.Run("-poor", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-poor"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
	t.Run("-better", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-better", "-quality", "small"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
	t.Run("-best", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-best"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
}
