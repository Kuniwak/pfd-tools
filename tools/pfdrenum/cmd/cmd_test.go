package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
)

func TestMainCommandByArgs(t *testing.T) {
	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs([]string{"testdata/simple/pfd.drawio"}, spy.NewProcInout())
	if exitStatus != 0 {
		t.Log(spy.Stderr.String())
		t.Log(spy.Stdout.String())
		t.Errorf("exitStatus = %d, want 0", exitStatus)
	}
}

func TestMainCommandByArgs_PNGInput(t *testing.T) {
	withoutID, err := os.ReadFile("../../../pfd/pfdencoding/pfddrawio/testdata/sequential_without_id.drawio")
	if err != nil {
		t.Fatal(err)
	}
	withID, err := os.ReadFile("../../../pfd/pfdencoding/pfddrawio/testdata/sequential_with_id.drawio")
	if err != nil {
		t.Fatal(err)
	}

	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "in.png")
	if err := os.WriteFile(pngPath, pngtest.WrapMxfile(t, withoutID), 0644); err != nil {
		t.Fatal(err)
	}

	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs([]string{pngPath}, spy.NewProcInout())
	if exitStatus != 0 {
		t.Log(spy.Stderr.String())
		t.Fatalf("exitStatus = %d, want 0", exitStatus)
	}

	out := spy.Stdout.Bytes()
	if !bytes.HasPrefix(out, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		t.Fatalf("output is not a PNG (stdout starts with %x)", out[:min(len(out), 16)])
	}

	xmlReader, err := pfddrawiopng.ExtractMxfile(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	xml, err := io.ReadAll(xmlReader)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(xml), strings.TrimSpace(extractRoot(string(withID)))) {
		t.Errorf("renumbered XML does not match expected sequential_with_id.drawio\n--- got ---\n%s\n--- want (excerpt) ---\n%s",
			string(xml), strings.TrimSpace(extractRoot(string(withID))))
	}
}

// extractRoot returns a small, structurally invariant snippet of the input
// XML (the contents of the first <root>...</root>) for loose comparison.
func extractRoot(xml string) string {
	start := strings.Index(xml, "<root>")
	end := strings.Index(xml, "</root>")
	if start < 0 || end < 0 {
		return xml
	}
	return xml[start : end+len("</root>")]
}
