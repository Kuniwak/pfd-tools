package cmd

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
)

func TestMainCommandByArgs(t *testing.T) {
	t.Run("atomic process", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-p", "testdata/loop/pfd.drawio", "-ap", "testdata/loop/atomic_proc.tsv", "P1"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
	t.Run("atomic deliverable", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-p", "testdata/loop/pfd.drawio", "-ad", "testdata/loop/deliv.tsv", "D1.1"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
	t.Run("composite process", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-cp", "testdata/loop/comp_proc.tsv", "P0"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
	t.Run("composite deliverable", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-cd", "testdata/loop/comp_deliv.tsv", "D1"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
	t.Run("reachable", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-ap", "testdata/loop/atomic_proc.tsv", "-reachable", "P1"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
	t.Run("backward reachable", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-ap", "testdata/loop/atomic_proc.tsv", "-backward-reachable", "P1"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
	t.Run("backward reachable feedback destination", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-ap", "testdata/loop/atomic_proc.tsv", "-backward-reachable-fb", "P1"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
	t.Run("milestone table", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-m", "testdata/loop/milestone.tsv", "M1"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
	t.Run("group table", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-g", "testdata/loop/group.tsv", "G1"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
}
