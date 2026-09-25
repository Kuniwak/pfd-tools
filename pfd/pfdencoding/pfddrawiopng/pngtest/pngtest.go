package pngtest

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng"
)

func WrapMxfile(t *testing.T, xml []byte) []byte {
	t.Helper()
	empty := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0, 0, 0, 13, 'I', 'H', 'D', 'R',
	}
	empty = append(empty, make([]byte, 13)...)
	empty = append(empty, 0, 0, 0, 0)
	empty = append(empty, 0, 0, 0, 0, 'I', 'E', 'N', 'D', 0, 0, 0, 0)

	out, err := pfddrawiopng.ReplaceMxfile(empty, xml)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func MxfileInPNG(t *testing.T, png []byte) string {
	t.Helper()

	signature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if !bytes.HasPrefix(png, signature) {
		t.Fatalf("pngtest.MxfileInPNG: not a PNG (it starts with %x)", png[:min(len(png), len(signature))])
	}

	xmlReader, err := pfddrawiopng.ExtractMxfile(bytes.NewReader(png))
	if err != nil {
		t.Fatal(err)
	}
	xml, err := io.ReadAll(xmlReader)
	if err != nil {
		t.Fatal(err)
	}
	return string(xml)
}

func WriteTempPNG(t *testing.T, xmlPath string) string {
	t.Helper()
	xml, err := os.ReadFile(xmlPath)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), filepath.Base(xmlPath)+".png")
	if err := os.WriteFile(out, WrapMxfile(t, xml), 0644); err != nil {
		t.Fatal(err)
	}
	return out
}
