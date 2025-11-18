package cmd

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
)

func TestMainCommandByArgs(t *testing.T) {
	t.Run("atomic process", func(t *testing.T) {
		t.Run("no -existing", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
		t.Run("-existing", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", "testdata/loop/atomic_proc.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
	})
	t.Run("atomic deliverable", func(t *testing.T) {
		t.Run("no -existing", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ad", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
		t.Run("-existing", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ad", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", "testdata/loop/deliv.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
	})
	t.Run("composite process", func(t *testing.T) {
		t.Run("no -existing", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "cp", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
		t.Run("-existing", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "cp", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", "testdata/loop/comp_proc.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
	})
	t.Run("composite deliverable", func(t *testing.T) {
		t.Run("no -existing", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "cd", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
		t.Run("-existing", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "cd", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", "testdata/loop/comp_deliv.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
	})
}
