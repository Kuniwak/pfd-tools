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
		"drawio without xml decl, with attrs": {
			Input:          []byte(`<mxfile host="abcdefghijklmnopqrstuvwxyz"></mxfile>`),
			ExpectedFormat: FormatDrawio,
		},
		"drawio without xml decl, no attrs": {
			Input:          []byte(`<mxfile><diagram id="a" name="Page-1"></diagram></mxfile>`),
			ExpectedFormat: FormatDrawio,
		},
		"drawio with newline terminator": {
			Input:          []byte("<mxfile\nhost=\"abcdefghijklmnopqrstuvwxyz\"></mxfile>"),
			ExpectedFormat: FormatDrawio,
		},
		"drawio with tab terminator": {
			Input:          []byte("<mxfile\thost=\"abcdefghijklmnopqrstuvwxyz\"></mxfile>"),
			ExpectedFormat: FormatDrawio,
		},
		"drawio with xml decl, with attrs": {
			Input:          []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<mxfile host=\"abcdefghijklmnopqrstuvwxyz\"></mxfile>"),
			ExpectedFormat: FormatDrawio,
		},
		"drawio with xml decl, no attrs": {
			Input:          []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<mxfile></mxfile>"),
			ExpectedFormat: FormatDrawio,
		},
		"drawio short bare mxfile shorter than maxPrefixLen": {
			Input:          []byte(`<mxfile></mxfile>`),
			ExpectedFormat: FormatDrawio,
		},
		"drawio with xml decl, crlf before mxfile": {
			Input:          []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\r\n<mxfile host=\"x\"></mxfile>"),
			ExpectedFormat: FormatDrawio,
		},
		"drawio with longer than standard xml decl": {
			Input:          []byte("<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>\n<mxfile></mxfile>"),
			ExpectedFormat: FormatDrawio,
		},
		"mxfile-prefixed element name is not drawio": {
			Input:          []byte(`<mxfileFoo attr="abcdefghijklmnopqrstuvwxyz"></mxfileFoo>`),
			ExpectedFormat: FormatUnknown,
		},
		"xml decl without mxfile is not drawio": {
			Input:          []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<svg>abcdefghijklmnop</svg>"),
			ExpectedFormat: FormatUnknown,
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

		"unknown, shorter than maxPrefixLen": {
			Input:          []byte("short"),
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
