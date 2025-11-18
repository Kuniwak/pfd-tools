package pfdfmt

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Format string

const (
	FormatDrawio  Format = "drawio"
	FormatJSON    Format = "json"
	FormatUnknown Format = "unknown"
)

const (
	drawioPrefixWithoutXMLDecl = "<mxfile "
	drawioPrefixWithXMLDecl    = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<mxfile "
	jsonPrefix                 = "{"
)

var allPrefixes = []string{
	drawioPrefixWithoutXMLDecl,
	drawioPrefixWithXMLDecl,
	jsonPrefix,
}

func Detect(r io.Reader) (Format, io.Reader, error) {
	maxPrefixLen := 0
	for _, prefix := range allPrefixes {
		if len(prefix) > maxPrefixLen {
			maxPrefixLen = len(prefix)
		}
	}
	buf := make([]byte, maxPrefixLen)

	_, err := io.ReadAtLeast(r, buf, maxPrefixLen)
	if err != nil {
		return FormatUnknown, r, fmt.Errorf("pfdfmt.Detect: %w", err)
	}

	r2 := io.MultiReader(bytes.NewReader(buf), r)
	s := string(buf)
	if strings.HasPrefix(s, drawioPrefixWithoutXMLDecl) || strings.HasPrefix(s, drawioPrefixWithXMLDecl) {
		return FormatDrawio, r2, nil
	}
	if strings.HasPrefix(s, jsonPrefix) {
		return FormatJSON, r2, nil
	}
	return FormatUnknown, r2, nil
}
