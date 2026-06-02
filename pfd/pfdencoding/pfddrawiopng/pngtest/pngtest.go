// Package pngtest provides helpers for tests that need to construct
// minimal drawio-style PNG byte streams with an embedded mxfile tEXt
// chunk. The output is not a renderable PNG image — it satisfies the
// drawio-PNG container format only well enough for pfd-tools to parse
// out the embedded mxfile XML.
package pngtest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng"
)

// WrapMxfile builds a minimal drawio-style PNG byte stream that
// embeds the given mxfile XML as a URL-encoded tEXt chunk. The
// container consists of only the PNG signature, a zero-filled IHDR,
// the mxfile tEXt chunk, and an IEND marker.
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

// WriteTempPNG reads the mxfile XML at xmlPath, wraps it via
// WrapMxfile, and writes the resulting drawio-PNG to t.TempDir()
// under the basename of xmlPath suffixed with ".png". The absolute
// path of the written PNG is returned. Use this from tests that
// only need one PNG per test; for multiple distinct PNGs from the
// same XML in a single test, call WrapMxfile and os.WriteFile
// directly.
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
