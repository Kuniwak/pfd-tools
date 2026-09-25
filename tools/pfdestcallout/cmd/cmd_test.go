package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawio"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
	"github.com/Kuniwak/pfd-tools/sugar"
	"github.com/Kuniwak/pfd-tools/xmldom"
)

const (
	simplePath  = "testdata/simple/pfd.drawio"
	calloutPath = "testdata/callout/pfd.drawio"
)

func countEstimateBoxes(xmlBytes []byte) int {
	nodes := sugar.Must(xmldom.ParseXML(bytes.NewReader(xmlBytes)))
	return len(pfddrawio.EstimateBoxes(nodes))
}

func TestMainCommandByArgs(t *testing.T) {
	testCases := map[string]struct {
		Args      []string
		Stdin     []string
		WantExit  int
		WantBoxes int
	}{

		"adds a box via file argument": {
			Args:      []string{simplePath},
			WantExit:  0,
			WantBoxes: 1,
		},
		"adds a box via stdin": {
			Stdin:     []string{string(sugar.Must(os.ReadFile(simplePath)))},
			WantExit:  0,
			WantBoxes: 1,
		},

		"already estimated": {
			Args:      []string{calloutPath},
			WantExit:  0,
			WantBoxes: 3,
		},
		"too many arguments": {
			Args:     []string{simplePath, calloutPath},
			WantExit: 1,
		},
		"inplace with stdin": {
			Args:     []string{"-inplace"},
			WantExit: 1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout(tc.Stdin...)
			exit := MainCommandByArgs(tc.Args, spy.NewProcInout())
			if exit != tc.WantExit {
				t.Fatalf("exit = %d, want %d\nstderr: %s\nstdout: %s", exit, tc.WantExit, spy.Stderr.String(), spy.Stdout.String())
			}
			if tc.WantExit != 0 {
				return
			}
			if got := countEstimateBoxes(spy.Stdout.Bytes()); got != tc.WantBoxes {
				t.Errorf("estimate boxes = %d, want %d", got, tc.WantBoxes)
			}
		})
	}
}

func TestMainCommandByArgs_Inplace(t *testing.T) {
	src := sugar.Must(os.ReadFile(simplePath))
	tmpPath := filepath.Join(t.TempDir(), "pfd.drawio")
	if err := os.WriteFile(tmpPath, src, 0644); err != nil {
		t.Fatal(err)
	}

	spy := cli.SpyProcInout()
	exit := MainCommandByArgs([]string{"-inplace", tmpPath}, spy.NewProcInout())
	if exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
	}
	if spy.Stdout.Len() != 0 {
		t.Errorf("stdout should be empty in -inplace mode, got %q", spy.Stdout.String())
	}

	afterOnce := sugar.Must(os.ReadFile(tmpPath))
	if got := countEstimateBoxes(afterOnce); got != 1 {
		t.Errorf("estimate boxes after 1st run = %d, want 1", got)
	}

	spy2 := cli.SpyProcInout()
	if exit := MainCommandByArgs([]string{"-inplace", tmpPath}, spy2.NewProcInout()); exit != 0 {
		t.Fatalf("2nd run exit = %d, want 0\nstderr: %s", exit, spy2.Stderr.String())
	}
	afterTwice := sugar.Must(os.ReadFile(tmpPath))
	if !bytes.Equal(afterOnce, afterTwice) {
		t.Errorf("in-place output changed on re-run (not idempotent)")
	}
}

func TestMainCommandByArgs_PNGInput(t *testing.T) {
	xmlBytes := sugar.Must(os.ReadFile(simplePath))
	pngPath := filepath.Join(t.TempDir(), "in.png")
	if err := os.WriteFile(pngPath, pngtest.WrapMxfile(t, xmlBytes), 0644); err != nil {
		t.Fatal(err)
	}

	spy := cli.SpyProcInout()
	exit := MainCommandByArgs([]string{pngPath}, spy.NewProcInout())
	if exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
	}

	gotXML := []byte(pngtest.MxfileInPNG(t, spy.Stdout.Bytes()))
	if got := countEstimateBoxes(gotXML); got != 1 {
		t.Errorf("estimate boxes in PNG-embedded XML = %d, want 1", got)
	}
}
