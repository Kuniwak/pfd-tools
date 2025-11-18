package cmd

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
)

func TestMainCommandByArgs(t *testing.T) {
	t.Run("same", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{
			"-p1", "testdata/loop/pfd.drawio", "-cd1", "testdata/loop/comp_deliv.tsv",
			"-p2", "testdata/loop/pfd.drawio", "-cd2", "testdata/loop/comp_deliv.tsv",
		}, spy.NewProcInout())

		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
	t.Run("different", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{
			"-p1", "testdata/loop/pfd.drawio", "-cd1", "testdata/loop/comp_deliv.tsv",
			"-p2", "testdata/longloop/pfd.drawio", "-cd2", "testdata/longloop/comp_deliv.tsv",
		}, spy.NewProcInout())

		if exitStatus != 1 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
	t.Run("prompt", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{
			"-p1", "testdata/loop/pfd.drawio", "-cd1", "testdata/loop/comp_deliv.tsv",
			"-p2", "testdata/longloop/pfd.drawio", "-cd2", "testdata/longloop/comp_deliv.tsv",
			"-prompt",
		}, spy.NewProcInout())

		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
}
