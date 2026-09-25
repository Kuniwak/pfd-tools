package cmd

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawio"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
)

const newPagePath = "testdata/newpage/pfd.drawio"

func pageNodeIDs(t *testing.T, xmlBytes []byte, pageName string) []pfd.NodeID {
	t.Helper()
	diagrams, err := pfddrawio.ParseDiagrams(bytes.NewReader(xmlBytes), slog.New(slogtest.NewTestHandler(t)))
	if err != nil {
		t.Fatalf("ParseDiagrams: %v", err)
	}
	ids := make([]pfd.NodeID, 0)
	for _, d := range diagrams {
		if pfddrawio.PageNameElementID(d.Name) != pfddrawio.PageNameElementID(pageName) {
			continue
		}
		for _, c := range d.Cells {
			if !c.IsVertex {
				continue
			}
			id, _, err := pfddrawio.ParseVertexValue(c.Value)
			if err != nil {
				continue
			}
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].Compare(ids[j]) < 0 })
	return ids
}

func TestMainCommandByArgs(t *testing.T) {
	spy := cli.SpyProcInout()
	exit := MainCommandByArgs([]string{newPagePath}, spy.NewProcInout())
	if exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
	}

	got := pageNodeIDs(t, spy.Stdout.Bytes(), "P1")
	want := []pfd.NodeID{"D1", "D2"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("page P1 node IDs mismatch (-want +got):\n%s", diff)
	}
}

func TestMainCommandByArgs_Inplace(t *testing.T) {
	src, err := os.ReadFile(newPagePath)
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
	got := pageNodeIDs(t, out, "P1")
	want := []pfd.NodeID{"D1", "D2"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("in-place page P1 node IDs mismatch (-want +got):\n%s", diff)
	}
}

func TestMainCommandByArgs_PNGInput(t *testing.T) {
	xmlBytes, err := os.ReadFile(newPagePath)
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

	got := pageNodeIDs(t, gotXML, "P1")
	want := []pfd.NodeID{"D1", "D2"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("PNG-embedded page P1 node IDs mismatch (-want +got):\n%s", diff)
	}
}
