package pfdfmt

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
)

type Format string

const (
	FormatDrawio    Format = "drawio"
	FormatDrawioPNG Format = "drawio-png"
	FormatJSON      Format = "json"
	FormatUnknown   Format = "unknown"
)

const jsonPrefix = "{"

var drawioHead = regexp.MustCompile(`^(?:<\?xml[^>]*\?>\s*)?<mxfile[ \t\r\n>]`)

var pngSignature = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

const maxPrefixLen = 256

func Detect(r io.Reader) (Format, io.Reader, error) {
	buf, r2, err := peek(r, maxPrefixLen)
	if err != nil {
		return FormatUnknown, r, fmt.Errorf("pfdfmt.Detect: %w", err)
	}

	switch {
	case bytes.HasPrefix(buf, pngSignature):
		return FormatDrawioPNG, r2, nil
	case drawioHead.Match(buf):
		return FormatDrawio, r2, nil
	case bytes.HasPrefix(buf, []byte(jsonPrefix)):
		return FormatJSON, r2, nil
	default:
		return FormatUnknown, r2, nil
	}
}

func peek(r io.Reader, n int) ([]byte, io.Reader, error) {
	buf := make([]byte, n)
	m, err := io.ReadFull(r, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, r, err
	}
	buf = buf[:m]
	return buf, io.MultiReader(bytes.NewReader(buf), r), nil
}
