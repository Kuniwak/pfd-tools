package pfddrawiopng

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"net/url"
	"strings"
)

// signature is the 8-byte PNG file signature.
var signature = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

const (
	chunkTypeTEXt = "tEXt"
	chunkTypeIEND = "IEND"
	keywordMxfile = "mxfile"
)

// ExtractMxfile reads a PNG byte stream from r, finds the first tEXt chunk
// whose keyword is "mxfile", URL-decodes its value, and returns a reader
// over the resulting XML bytes. Returns an error if r is not a PNG, ends
// before IEND, or contains no mxfile tEXt chunk.
func ExtractMxfile(r io.Reader) (io.Reader, error) {
	sig := make([]byte, len(signature))
	if _, err := io.ReadFull(r, sig); err != nil {
		return nil, fmt.Errorf("pfddrawiopng.ExtractMxfile: %w", err)
	}
	if !bytes.Equal(sig, signature) {
		return nil, errors.New("pfddrawiopng.ExtractMxfile: not a PNG")
	}

	for {
		typ, data, err := readChunk(r)
		if err != nil {
			return nil, fmt.Errorf("pfddrawiopng.ExtractMxfile: %w", err)
		}
		if typ == chunkTypeTEXt {
			keyword, value, ok := splitTEXt(data)
			if ok && keyword == keywordMxfile {
				decoded, err := url.QueryUnescape(value)
				if err != nil {
					return nil, fmt.Errorf("pfddrawiopng.ExtractMxfile: %w", err)
				}
				return bytes.NewReader([]byte(decoded)), nil
			}
		}
		if typ == chunkTypeIEND {
			return nil, errors.New("pfddrawiopng.ExtractMxfile: no mxfile tEXt chunk found")
		}
	}
}

// readChunk reads one PNG chunk: 4-byte length, 4-byte type, data, 4-byte CRC.
// CRC bytes are read and discarded (not validated).
func readChunk(r io.Reader) (string, []byte, error) {
	header := make([]byte, 8)
	if _, err := io.ReadFull(r, header); err != nil {
		return "", nil, err
	}
	length := binary.BigEndian.Uint32(header[0:4])
	typ := string(header[4:8])
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return "", nil, err
	}
	crc := make([]byte, 4)
	if _, err := io.ReadFull(r, crc); err != nil {
		return "", nil, err
	}
	return typ, data, nil
}

// splitTEXt splits a tEXt chunk data payload into keyword and value at the
// first 0x00 byte. Returns ok=false if no separator is present.
func splitTEXt(data []byte) (string, string, bool) {
	i := bytes.IndexByte(data, 0x00)
	if i < 0 {
		return "", "", false
	}
	return string(data[:i]), string(data[i+1:]), true
}

// ReplaceMxfile returns a copy of orig (a PNG byte stream) with its first
// tEXt chunk whose keyword is "mxfile" replaced by a new tEXt chunk
// containing the URL-encoded form of newXML. All other chunks are preserved
// byte-for-byte (including their original CRC). If no mxfile tEXt chunk
// exists, one is inserted immediately before IEND. CRC32 is recomputed for
// the new chunk only. Returns an error if orig is not a valid PNG.
func ReplaceMxfile(orig []byte, newXML []byte) ([]byte, error) {
	if !bytes.HasPrefix(orig, signature) {
		return nil, errors.New("pfddrawiopng.ReplaceMxfile: not a PNG")
	}

	r := bytes.NewReader(orig[len(signature):])
	out := bytes.NewBuffer(nil)
	out.Write(signature)

	newChunk := buildMxfileChunk(newXML)
	replaced := false
	sawIEND := false

	for r.Len() > 0 {
		header := make([]byte, 8)
		if _, err := io.ReadFull(r, header); err != nil {
			return nil, fmt.Errorf("pfddrawiopng.ReplaceMxfile: %w", err)
		}
		length := binary.BigEndian.Uint32(header[0:4])
		typ := string(header[4:8])
		body := make([]byte, int(length)+4)
		if _, err := io.ReadFull(r, body); err != nil {
			return nil, fmt.Errorf("pfddrawiopng.ReplaceMxfile: %w", err)
		}
		data := body[:length]

		if typ == chunkTypeIEND {
			if !replaced {
				out.Write(newChunk)
				replaced = true
			}
			out.Write(header)
			out.Write(body)
			sawIEND = true
			break
		}

		if !replaced && typ == chunkTypeTEXt {
			if keyword, _, ok := splitTEXt(data); ok && keyword == keywordMxfile {
				out.Write(newChunk)
				replaced = true
				continue
			}
		}

		out.Write(header)
		out.Write(body)
	}

	if !sawIEND {
		return nil, errors.New("pfddrawiopng.ReplaceMxfile: missing IEND chunk")
	}
	return out.Bytes(), nil
}

// buildMxfileChunk builds a complete tEXt chunk (length + type + data + crc)
// for a "mxfile" keyword whose value is the percent-encoded form of xml,
// compatible with JavaScript's encodeURIComponent (spaces as %20, not '+').
// drawio decodes this value with decodeURIComponent, which does not translate
// '+' back to space.
func buildMxfileChunk(xml []byte) []byte {
	encoded := strings.ReplaceAll(url.QueryEscape(string(xml)), "+", "%20")
	data := make([]byte, 0, len(keywordMxfile)+1+len(encoded))
	data = append(data, keywordMxfile...)
	data = append(data, 0x00)
	data = append(data, encoded...)

	buf := bytes.NewBuffer(nil)
	var lenBytes [4]byte
	binary.BigEndian.PutUint32(lenBytes[:], uint32(len(data)))
	buf.Write(lenBytes[:])
	buf.WriteString(chunkTypeTEXt)
	buf.Write(data)

	crcInput := append([]byte(chunkTypeTEXt), data...)
	var crcBytes [4]byte
	binary.BigEndian.PutUint32(crcBytes[:], crc32.ChecksumIEEE(crcInput))
	buf.Write(crcBytes[:])
	return buf.Bytes()
}
