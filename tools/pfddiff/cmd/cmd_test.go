package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
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
	t.Run("png input", func(t *testing.T) {
		xml, err := os.ReadFile("testdata/loop/pfd.drawio")
		if err != nil {
			t.Fatal(err)
		}
		tmpDir := t.TempDir()
		pngPath1 := filepath.Join(tmpDir, "a.drawio.png")
		pngPath2 := filepath.Join(tmpDir, "b.drawio.png")
		if err := os.WriteFile(pngPath1, pngtest.WrapMxfile(t, xml), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(pngPath2, pngtest.WrapMxfile(t, xml), 0644); err != nil {
			t.Fatal(err)
		}
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{
			"-p1", pngPath1, "-cd1", "testdata/loop/comp_deliv.tsv",
			"-p2", pngPath2, "-cd2", "testdata/loop/comp_deliv.tsv",
		}, spy.NewProcInout())

		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
}
