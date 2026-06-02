package pfdfmt

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Format string

const (
	FormatDrawio    Format = "drawio"
	FormatDrawioPNG Format = "drawio-png"
	FormatJSON      Format = "json"
	FormatUnknown   Format = "unknown"
)

const (
	drawioPrefixWithoutXMLDecl = "<mxfile "
	drawioPrefixWithXMLDecl    = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<mxfile "
	jsonPrefix                 = "{"
)

var pngSignature = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

var allPrefixes = []string{
	drawioPrefixWithoutXMLDecl,
	drawioPrefixWithXMLDecl,
	jsonPrefix,
}

func Detect(r io.Reader) (Format, io.Reader, error) {
	maxPrefixLen := len(pngSignature)
	for _, prefix := range allPrefixes {
		if len(prefix) > maxPrefixLen {
			maxPrefixLen = len(prefix)
		}
	}

	buf, r2, err := peek(r, maxPrefixLen)
	if err != nil {
		return FormatUnknown, r, fmt.Errorf("pfdfmt.Detect: %w", err)
	}

	if bytes.HasPrefix(buf, pngSignature) {
		return FormatDrawioPNG, r2, nil
	}

	s := string(buf)
	if strings.HasPrefix(s, drawioPrefixWithoutXMLDecl) || strings.HasPrefix(s, drawioPrefixWithXMLDecl) {
		return FormatDrawio, r2, nil
	}
	if strings.HasPrefix(s, jsonPrefix) {
		return FormatJSON, r2, nil
	}
	return FormatUnknown, r2, nil
}

// peek reads exactly n bytes from r and returns them along with a reader
// that re-streams the consumed bytes followed by the rest of r.
func peek(r io.Reader, n int) ([]byte, io.Reader, error) {
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, r, err
	}
	return buf, io.MultiReader(bytes.NewReader(buf), r), nil
}
