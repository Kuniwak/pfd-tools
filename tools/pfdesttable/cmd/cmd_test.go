package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
	"github.com/Kuniwak/pfd-tools/sugar"
	"github.com/google/go-cmp/cmp"
)

const (
	simplePath  = "testdata/simple/pfd.drawio"
	calloutPath = "testdata/callout/pfd.drawio"
)

const wantCalloutTSV = "ID\tDescription\t楽観作業量\t悲観作業量\n" +
	"P1\tプロセス1\t2\t3\n" +
	"P2\tプロセス2\t1\t2\n" +
	"P3\tプロセス3\t1\t5\n"

func TestMainCommandByArgs(t *testing.T) {
	testCases := map[string]struct {
		Args       []string
		Stdin      []string
		WantExit   int
		WantStdout string
	}{
		"drawio file": {
			Args:       []string{calloutPath},
			WantExit:   0,
			WantStdout: wantCalloutTSV,
		},
		"stdin": {
			Stdin:      []string{string(sugar.Must(os.ReadFile(calloutPath)))},
			WantExit:   0,
			WantStdout: wantCalloutTSV,
		},
		"too many arguments": {
			Args:     []string{calloutPath, simplePath},
			WantExit: 1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout(tc.Stdin...)
			exit := MainCommandByArgs(tc.Args, spy.NewProcInout())
			if exit != tc.WantExit {
				t.Fatalf("exit = %d, want %d\nstderr: %s", exit, tc.WantExit, spy.Stderr.String())
			}
			if tc.WantExit != 0 {
				return
			}
			if diff := cmp.Diff(tc.WantStdout, spy.Stdout.String()); diff != "" {
				t.Errorf("stdout mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMainCommandByArgs_WarnsProcessWithoutEstimateBox(t *testing.T) {
	spy := cli.SpyProcInout()
	if exit := MainCommandByArgs([]string{simplePath}, spy.NewProcInout()); exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
	}

	want := "ID\tDescription\nP1\tプロセス\n"
	if diff := cmp.Diff(want, spy.Stdout.String()); diff != "" {
		t.Errorf("stdout mismatch (-want +got):\n%s", diff)
	}
	if !strings.Contains(spy.Stderr.String(), "no estimate box") {
		t.Errorf("stderr should warn about the process without an estimate box, got: %s", spy.Stderr.String())
	}
}

func TestMainCommandByArgs_PNGInput(t *testing.T) {
	xmlBytes := sugar.Must(os.ReadFile(calloutPath))
	pngPath := filepath.Join(t.TempDir(), "in.png")
	if err := os.WriteFile(pngPath, pngtest.WrapMxfile(t, xmlBytes), 0644); err != nil {
		t.Fatal(err)
	}

	spy := cli.SpyProcInout()
	if exit := MainCommandByArgs([]string{pngPath}, spy.NewProcInout()); exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
	}
	if diff := cmp.Diff(wantCalloutTSV, spy.Stdout.String()); diff != "" {
		t.Errorf("stdout mismatch (-want +got):\n%s", diff)
	}
}
