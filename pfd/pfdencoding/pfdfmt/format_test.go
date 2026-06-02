package pfdfmt

import (
	"bytes"
	"encoding/binary"
	"io"
	"net/url"
	"testing"
)

func TestDetect(t *testing.T) {
	pngWithMxfile := buildPNGWithMxfile(t, "<mxfile></mxfile>")

	testCases := map[string]struct {
		Input          []byte
		ExpectedFormat Format
	}{
		"drawio without xml decl": {
			Input:          []byte(drawioPrefixWithoutXMLDecl + `host="abcdefghijklmnopqrstuvwxyz"></mxfile>`),
			ExpectedFormat: FormatDrawio,
		},
		"drawio with xml decl": {
			Input:          []byte(drawioPrefixWithXMLDecl + `host="abcdefghijklmnopqrstuvwxyz"></mxfile>`),
			ExpectedFormat: FormatDrawio,
		},
		"json": {
			Input:          []byte(`{"key": "abcdefghijklmnopqrstuvwxyz0123456789"}`),
			ExpectedFormat: FormatJSON,
		},
		"drawio-png": {
			Input:          pngWithMxfile,
			ExpectedFormat: FormatDrawioPNG,
		},
		"unknown": {
			Input:          []byte("plain text content that does not match any prefix at all"),
			ExpectedFormat: FormatUnknown,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			format, r, err := Detect(bytes.NewReader(testCase.Input))
			if err != nil {
				t.Fatal(err)
			}
			if format != testCase.ExpectedFormat {
				t.Errorf("expected format %q, got %q", testCase.ExpectedFormat, format)
			}
			actual, err := io.ReadAll(r)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(actual, testCase.Input) {
				t.Errorf("returned reader did not stream original bytes back")
			}
		})
	}
}

func TestDetect_TruncatedPNGSignature(t *testing.T) {
	// Reading at least 8 bytes should still succeed with FormatUnknown when
	// non-PNG content is shorter than maxPrefixLen but at least 8 bytes long.
	input := []byte("short")
	if _, _, err := Detect(bytes.NewReader(input)); err == nil {
		t.Errorf("expected error for too-short input, got nil")
	}
}

// buildPNGWithMxfile builds a minimal PNG byte stream with a tEXt chunk
// containing keyword "mxfile" and the given XML URL-encoded as the value.
func buildPNGWithMxfile(t *testing.T, xml string) []byte {
	t.Helper()
	signature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	buf := bytes.NewBuffer(nil)
	buf.Write(signature)
	writeChunk(buf, "IHDR", make([]byte, 13))
	textData := []byte("mxfile")
	textData = append(textData, 0x00)
	textData = append(textData, []byte(url.QueryEscape(xml))...)
	writeChunk(buf, "tEXt", textData)
	writeChunk(buf, "IEND", nil)
	return buf.Bytes()
}

func writeChunk(w *bytes.Buffer, typ string, data []byte) {
	if len(typ) != 4 {
		panic("chunk type must be 4 bytes")
	}
	var lenBytes [4]byte
	binary.BigEndian.PutUint32(lenBytes[:], uint32(len(data)))
	w.Write(lenBytes[:])
	w.WriteString(typ)
	w.Write(data)
	w.Write([]byte{0x00, 0x00, 0x00, 0x00})
}

