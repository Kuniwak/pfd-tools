package cmd

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawio"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
)

const (
	unlinkPath   = "testdata/unlink/pfd.drawio"
	simplePath   = "testdata/simple/pfd.drawio"
	selfLoopPath = "testdata/selfloop/pfd.drawio"
)

func edgeLinks(t *testing.T, xmlBytes []byte) map[pfddrawio.CellID][2]pfddrawio.CellID {
	t.Helper()
	diagrams, err := pfddrawio.ParseDiagrams(bytes.NewReader(xmlBytes), slog.New(slogtest.NewTestHandler(t)))
	if err != nil {
		t.Fatalf("ParseDiagrams: %v", err)
	}
	links := make(map[pfddrawio.CellID][2]pfddrawio.CellID)
	for _, d := range diagrams {
		for _, c := range d.Cells {
			if c.IsEdge {
				links[c.ID] = [2]pfddrawio.CellID{c.Source, c.Target}
			}
		}
	}
	return links
}

func TestMainCommandByArgs(t *testing.T) {
	spy := cli.SpyProcInout()
	exit := MainCommandByArgs([]string{unlinkPath}, spy.NewProcInout())
	if exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
	}

	got := edgeLinks(t, spy.Stdout.Bytes())

	wantBytes, err := os.ReadFile(simplePath)
	if err != nil {
		t.Fatal(err)
	}
	want := edgeLinks(t, wantBytes)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("edge links mismatch vs testdata/simple (-want +got):\n%s", diff)
	}
}

func TestMainCommandByArgs_Inplace(t *testing.T) {
	src, err := os.ReadFile(unlinkPath)
	if err != nil {
		t.Fatal(err)
	}
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

	out, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatal(err)
	}
	wantBytes, err := os.ReadFile(simplePath)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(edgeLinks(t, wantBytes), edgeLinks(t, out)); diff != "" {
		t.Errorf("in-place result mismatch vs testdata/simple (-want +got):\n%s", diff)
	}
}

func TestMainCommandByArgs_PNGInput(t *testing.T) {
	xmlBytes, err := os.ReadFile(unlinkPath)
	if err != nil {
		t.Fatal(err)
	}
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

	wantBytes, err := os.ReadFile(simplePath)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(edgeLinks(t, wantBytes), edgeLinks(t, gotXML)); diff != "" {
		t.Errorf("PNG-embedded result mismatch vs testdata/simple (-want +got):\n%s", diff)
	}
}

func TestMainCommandByArgs_SelfLoop(t *testing.T) {
	t.Run("default keeps the edge unconnected (no self-loop is created)", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exit := MainCommandByArgs([]string{selfLoopPath}, spy.NewProcInout())
		if exit != 0 {
			t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
		}
		out := spy.Stdout.Bytes()
		for id, link := range edgeLinks(t, out) {
			if link[0] != "" && link[0] == link[1] {
				t.Errorf("edge %s is a self-loop %v; pfdfix must not create one", id, link)
			}
		}
		if !bytes.Contains(out, []byte(`id="4"`)) {
			t.Errorf("self-loop edge 4 should be kept (unconnected), but it is missing from output")
		}
	})

	t.Run("-delete-unhit removes the self-loop edge", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exit := MainCommandByArgs([]string{"-delete-unhit", selfLoopPath}, spy.NewProcInout())
		if exit != 0 {
			t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
		}
		if bytes.Contains(spy.Stdout.Bytes(), []byte(`id="4"`)) {
			t.Errorf("self-loop edge 4 should be deleted, but it is present in output")
		}
	})
}

func TestMainCommandByArgs_DeleteUnhitWithShrink(t *testing.T) {
	spy := cli.SpyProcInout()
	exit := MainCommandByArgs([]string{"-delete-unhit", "-hitbox-expand", "0.9", unlinkPath}, spy.NewProcInout())
	if exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
	}
	if got := edgeLinks(t, spy.Stdout.Bytes()); len(got) != 0 {
		t.Errorf("expected all edges deleted, but got %d edges: %v", len(got), got)
	}
}
