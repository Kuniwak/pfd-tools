package pfddrawiopng

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"io"
	"net/url"
	"testing"
)

var pngSignature = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

type chunk struct {
	Type string
	Data []byte
}

func buildPNG(chunks []chunk) []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(pngSignature)
	for _, c := range chunks {
		var lenBytes [4]byte
		binary.BigEndian.PutUint32(lenBytes[:], uint32(len(c.Data)))
		buf.Write(lenBytes[:])
		buf.WriteString(c.Type)
		buf.Write(c.Data)
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00})
	}
	return buf.Bytes()
}

func tEXtPayload(keyword, value string) []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(keyword)
	buf.WriteByte(0x00)
	buf.WriteString(value)
	return buf.Bytes()
}

func TestExtractMxfile(t *testing.T) {
	encodedMxfile := url.QueryEscape("<mxfile></mxfile>")

	testCases := map[string]struct {
		Input         []byte
		ExpectedXML   string
		ExpectedError bool
	}{
		"minimal mxfile tEXt": {
			Input: buildPNG([]chunk{
				{Type: "IHDR", Data: make([]byte, 13)},
				{Type: "tEXt", Data: tEXtPayload("mxfile", encodedMxfile)},
				{Type: "IEND", Data: nil},
			}),
			ExpectedXML: "<mxfile></mxfile>",
		},
		"non-mxfile tEXt is skipped": {
			Input: buildPNG([]chunk{
				{Type: "IHDR", Data: make([]byte, 13)},
				{Type: "tEXt", Data: tEXtPayload("Software", "draw.io")},
				{Type: "tEXt", Data: tEXtPayload("mxfile", encodedMxfile)},
				{Type: "IEND", Data: nil},
			}),
			ExpectedXML: "<mxfile></mxfile>",
		},
		"missing mxfile chunk": {
			Input: buildPNG([]chunk{
				{Type: "IHDR", Data: make([]byte, 13)},
				{Type: "tEXt", Data: tEXtPayload("Software", "draw.io")},
				{Type: "IEND", Data: nil},
			}),
			ExpectedError: true,
		},
		"not a PNG": {
			Input:         []byte("<mxfile></mxfile>"),
			ExpectedError: true,
		},
		"truncated before IEND": {
			Input: func() []byte {
				full := buildPNG([]chunk{
					{Type: "IHDR", Data: make([]byte, 13)},
					{Type: "tEXt", Data: tEXtPayload("Software", "draw.io")},
				})
				return full
			}(),
			ExpectedError: true,
		},
		"invalid url-encoded value": {
			Input: buildPNG([]chunk{
				{Type: "IHDR", Data: make([]byte, 13)},
				{Type: "tEXt", Data: tEXtPayload("mxfile", "%ZZ")},
				{Type: "IEND", Data: nil},
			}),
			ExpectedError: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			r, err := ExtractMxfile(bytes.NewReader(testCase.Input))
			if testCase.ExpectedError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			actual, err := io.ReadAll(r)
			if err != nil {
				t.Fatal(err)
			}
			if string(actual) != testCase.ExpectedXML {
				t.Errorf("expected %q, got %q", testCase.ExpectedXML, string(actual))
			}
		})
	}
}

func chunkSlice(t *testing.T, png []byte) []struct {
	Type string
	Data []byte
	CRC  []byte
} {
	t.Helper()
	if !bytes.HasPrefix(png, pngSignature) {
		t.Fatalf("not a PNG")
	}
	var out []struct {
		Type string
		Data []byte
		CRC  []byte
	}
	i := len(pngSignature)
	for i < len(png) {
		if i+8 > len(png) {
			t.Fatalf("truncated chunk header")
		}
		length := binary.BigEndian.Uint32(png[i : i+4])
		typ := string(png[i+4 : i+8])
		end := i + 8 + int(length) + 4
		if end > len(png) {
			t.Fatalf("truncated chunk body")
		}
		out = append(out, struct {
			Type string
			Data []byte
			CRC  []byte
		}{
			Type: typ,
			Data: png[i+8 : i+8+int(length)],
			CRC:  png[i+8+int(length) : end],
		})
		i = end
		if typ == "IEND" {
			break
		}
	}
	return out
}

func TestReplaceMxfile(t *testing.T) {
	idatPayload := []byte("DATA-bytes-1234567890abcdef")
	newXML := []byte("<mxfile><a/></mxfile>")

	testCases := map[string]struct {
		Input         []byte
		NewXML        []byte
		ExpectedError bool

		Verify func(t *testing.T, originalChunks, outputChunks []struct {
			Type string
			Data []byte
			CRC  []byte
		}, output []byte)
	}{
		"replace existing mxfile tEXt": {
			Input: buildPNG([]chunk{
				{Type: "IHDR", Data: make([]byte, 13)},
				{Type: "tEXt", Data: tEXtPayload("mxfile", url.QueryEscape("<mxfile/>"))},
				{Type: "IEND", Data: nil},
			}),
			NewXML: newXML,
			Verify: func(t *testing.T, orig, out []struct {
				Type string
				Data []byte
				CRC  []byte
			}, output []byte) {
				if len(out) != len(orig) {
					t.Errorf("expected %d chunks, got %d", len(orig), len(out))
				}
				if out[0].Type != "IHDR" || !bytes.Equal(out[0].Data, orig[0].Data) || !bytes.Equal(out[0].CRC, orig[0].CRC) {
					t.Errorf("IHDR not preserved byte-for-byte")
				}
				if out[1].Type != "tEXt" {
					t.Errorf("expected tEXt at position 1, got %q", out[1].Type)
				}
				keyword, value, ok := splitTEXt(out[1].Data)
				if !ok || keyword != "mxfile" {
					t.Errorf("expected mxfile keyword, got %q ok=%v", keyword, ok)
				}
				decoded, err := url.PathUnescape(value)
				if err != nil {
					t.Fatal(err)
				}
				if decoded != string(newXML) {
					t.Errorf("expected %q, got %q", newXML, decoded)
				}
				if out[2].Type != "IEND" {
					t.Errorf("expected IEND at end")
				}
			},
		},
		"encodes spaces as %20 not + (drawio decodeURIComponent compatible)": {
			Input: buildPNG([]chunk{
				{Type: "IHDR", Data: make([]byte, 13)},
				{Type: "tEXt", Data: tEXtPayload("mxfile", url.QueryEscape("<mxfile/>"))},
				{Type: "IEND", Data: nil},
			}),
			NewXML: []byte("<mxfile>\n  <diagram name=\"Page 1\"/>\n</mxfile>"),
			Verify: func(t *testing.T, orig, out []struct {
				Type string
				Data []byte
				CRC  []byte
			}, output []byte) {
				var mxfileChunk *struct {
					Type string
					Data []byte
					CRC  []byte
				}
				for i := range out {
					if out[i].Type == "tEXt" {
						keyword, _, ok := splitTEXt(out[i].Data)
						if ok && keyword == "mxfile" {
							mxfileChunk = &out[i]
							break
						}
					}
				}
				if mxfileChunk == nil {
					t.Fatal("mxfile tEXt chunk missing in output")
				}
				_, value, _ := splitTEXt(mxfileChunk.Data)
				if bytes.ContainsRune([]byte(value), '+') {
					t.Errorf("encoded value must not contain '+' (drawio decodeURIComponent does not decode '+' to space): %q", value)
				}
				if !bytes.Contains([]byte(value), []byte("%20")) {
					t.Errorf("encoded value must contain %%20 for spaces: %q", value)
				}
				decoded, err := url.PathUnescape(value)
				if err != nil {
					t.Fatal(err)
				}
				expected := "<mxfile>\n  <diagram name=\"Page 1\"/>\n</mxfile>"
				if decoded != expected {
					t.Errorf("expected %q, got %q", expected, decoded)
				}
			},
		},
		"insert mxfile tEXt when absent": {
			Input: buildPNG([]chunk{
				{Type: "IHDR", Data: make([]byte, 13)},
				{Type: "tEXt", Data: tEXtPayload("Software", "draw.io")},
				{Type: "IEND", Data: nil},
			}),
			NewXML: newXML,
			Verify: func(t *testing.T, orig, out []struct {
				Type string
				Data []byte
				CRC  []byte
			}, output []byte) {
				if len(out) != len(orig)+1 {
					t.Errorf("expected %d chunks, got %d", len(orig)+1, len(out))
				}
				if out[0].Type != "IHDR" {
					t.Errorf("first chunk should be IHDR")
				}
				if out[1].Type != "tEXt" {
					t.Errorf("Software tEXt should be preserved at position 1")
				}
				if !bytes.Equal(out[1].Data, orig[1].Data) || !bytes.Equal(out[1].CRC, orig[1].CRC) {
					t.Errorf("Software tEXt mutated")
				}
				if out[2].Type != "tEXt" {
					t.Errorf("expected new mxfile tEXt at position 2, got %q", out[2].Type)
				}
				keyword, _, ok := splitTEXt(out[2].Data)
				if !ok || keyword != "mxfile" {
					t.Errorf("expected mxfile keyword in inserted chunk, got %q", keyword)
				}
				if out[3].Type != "IEND" {
					t.Errorf("IEND should be last chunk")
				}
			},
		},
		"preserves IDAT bytes byte-for-byte": {
			Input: buildPNG([]chunk{
				{Type: "IHDR", Data: make([]byte, 13)},
				{Type: "IDAT", Data: idatPayload},
				{Type: "tEXt", Data: tEXtPayload("mxfile", url.QueryEscape("<mxfile/>"))},
				{Type: "IEND", Data: nil},
			}),
			NewXML: newXML,
			Verify: func(t *testing.T, orig, out []struct {
				Type string
				Data []byte
				CRC  []byte
			}, output []byte) {
				var origIDAT, outIDAT *struct {
					Type string
					Data []byte
					CRC  []byte
				}
				for i := range orig {
					if orig[i].Type == "IDAT" {
						origIDAT = &orig[i]
					}
				}
				for i := range out {
					if out[i].Type == "IDAT" {
						outIDAT = &out[i]
					}
				}
				if origIDAT == nil || outIDAT == nil {
					t.Fatal("IDAT chunk missing")
				}
				if !bytes.Equal(origIDAT.Data, outIDAT.Data) {
					t.Errorf("IDAT data changed")
				}
				if !bytes.Equal(origIDAT.CRC, outIDAT.CRC) {
					t.Errorf("IDAT CRC changed")
				}
			},
		},
		"recomputes CRC for new tEXt chunk": {
			Input: buildPNG([]chunk{
				{Type: "IHDR", Data: make([]byte, 13)},
				{Type: "tEXt", Data: tEXtPayload("mxfile", url.QueryEscape("<mxfile/>"))},
				{Type: "IEND", Data: nil},
			}),
			NewXML: newXML,
			Verify: func(t *testing.T, orig, out []struct {
				Type string
				Data []byte
				CRC  []byte
			}, output []byte) {
				var mxfileChunk *struct {
					Type string
					Data []byte
					CRC  []byte
				}
				for i := range out {
					if out[i].Type == "tEXt" {
						keyword, _, ok := splitTEXt(out[i].Data)
						if ok && keyword == "mxfile" {
							mxfileChunk = &out[i]
							break
						}
					}
				}
				if mxfileChunk == nil {
					t.Fatal("mxfile tEXt chunk missing in output")
				}
				expected := crc32.ChecksumIEEE(append([]byte(mxfileChunk.Type), mxfileChunk.Data...))
				actual := binary.BigEndian.Uint32(mxfileChunk.CRC)
				if expected != actual {
					t.Errorf("expected CRC %x, got %x", expected, actual)
				}
			},
		},
		"not a PNG": {
			Input:         []byte("hello world"),
			NewXML:        newXML,
			ExpectedError: true,
		},
		"missing IEND": {
			Input: buildPNG([]chunk{
				{Type: "IHDR", Data: make([]byte, 13)},
				{Type: "tEXt", Data: tEXtPayload("mxfile", url.QueryEscape("<mxfile/>"))},
			}),
			NewXML:        newXML,
			ExpectedError: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			out, err := ReplaceMxfile(testCase.Input, testCase.NewXML)
			if testCase.ExpectedError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.HasPrefix(out, pngSignature) {
				t.Errorf("output does not start with PNG signature")
			}
			if testCase.Verify != nil {
				testCase.Verify(t, chunkSlice(t, testCase.Input), chunkSlice(t, out), out)
			}
		})
	}
}
