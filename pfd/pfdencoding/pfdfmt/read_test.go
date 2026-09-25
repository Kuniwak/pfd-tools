package pfdfmt

import (
	"io"
	"log/slog"
	"testing"

	"github.com/Kuniwak/pfd-tools/xmldom"
)

func countMxfileElements(nodes []*xmldom.Node) int {
	count := 0
	for _, root := range nodes {
		root.Traverse(func(n *xmldom.Node) {
			if n.Kind == xmldom.ElementNode && n.Start.Name.Local == "mxfile" {
				count++
			}
		}, nil)
	}
	return count
}

func TestReadDrawioNodes(t *testing.T) {
	const xmlWithAttrs = `<mxfile host="test"><diagram id="a" name="P0"></diagram></mxfile>`
	const xmlBare = `<mxfile><diagram id="a" name="P0"></diagram></mxfile>`
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	testCases := map[string]struct {
		Input           []byte
		ExpectError     bool
		ExpectedMxfiles int
	}{
		"drawio in, with attrs": {
			Input:           []byte(xmlWithAttrs),
			ExpectedMxfiles: 1,
		},

		"drawio in, no attrs": {
			Input:           []byte(xmlBare),
			ExpectedMxfiles: 1,
		},
		"png in": {
			Input:           buildPNGWithMxfile(t, xmlWithAttrs),
			ExpectedMxfiles: 1,
		},
		"unsupported format": {
			Input:       []byte("not a drawio"),
			ExpectError: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			nodes, err := ReadDrawioNodes(testCase.Input, logger)
			if testCase.ExpectError {
				if err == nil {
					t.Error("want error for unsupported format, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := countMxfileElements(nodes); got != testCase.ExpectedMxfiles {
				t.Errorf("mxfile elements = %d, want %d", got, testCase.ExpectedMxfiles)
			}
		})
	}
}
