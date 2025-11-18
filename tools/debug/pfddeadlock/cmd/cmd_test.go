package cmd

import (
	"os"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
)

func TestCmd(t *testing.T) {
	tempDir, err := os.MkdirTemp("/tmp", "pfddeadlock-test.XXXXXX")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-o", tempDir}, spy.NewProcInout())

	if exitStatus != 0 {
		t.Log(spy.Stderr.String())
		t.Log(spy.Stdout.String())
		t.Errorf("exitStatus = %d, want 0", exitStatus)
	}
}
