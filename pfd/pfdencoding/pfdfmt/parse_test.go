package pfdfmt

import (
	"bytes"
	"log/slog"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
)

func TestParse_PNG(t *testing.T) {
	xml, err := os.ReadFile("../pfddrawio/testdata/example.drawio")
	if err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slogtest.NewTestHandler(t))
	parseOpts := &ParseOptions{CompositeDeliverableTable: &pfd.CompositeDeliverableTable{}}

	expected, err := Parse("png", bytes.NewReader(xml), parseOpts, logger)
	if err != nil {
		t.Fatal(err)
	}

	pngBytes := buildPNGWithMxfile(t, string(xml))

	actual, err := Parse("png", bytes.NewReader(pngBytes), parseOpts, logger)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Error(cmp.Diff(expected, actual))
	}
}

func TestParse_PNGWithUrlEncodedXML(t *testing.T) {
	xml := `<mxfile host="x"><diagram id="d1" name="P0"><mxGraphModel><root><mxCell id="0"/><mxCell id="1" parent="0"/></root></mxGraphModel></diagram></mxfile>`
	pngBytes := buildPNGWithMxfile(t, xml)

	logger := slog.New(slogtest.NewTestHandler(t))
	parseOpts := &ParseOptions{CompositeDeliverableTable: &pfd.CompositeDeliverableTable{}}

	if _, err := Parse("png", bytes.NewReader(pngBytes), parseOpts, logger); err != nil {
		t.Fatal(err)
	}

	// Sanity-check the helper produces a properly URL-encoded value.
	if !bytes.Contains(pngBytes, []byte(url.QueryEscape(xml))) {
		t.Errorf("PNG bytes do not contain url-escaped XML")
	}
}

// TestParse_PNGWithMxfileShapes covers various mxfile XML shapes that real
// drawio outputs may embed in the PNG tEXt chunk. The format detection of the
// extracted XML must not be sensitive to whitespace, attributes, or XML
// declaration variations because the surrounding PNG envelope already
// identifies the content as drawio.
func TestParse_PNGWithMxfileShapes(t *testing.T) {
	body := `<diagram id="d1" name="P0"><mxGraphModel><root><mxCell id="0"/><mxCell id="1" parent="0"/></root></mxGraphModel></diagram>`

	testCases := map[string]string{
		"no attributes":                  `<mxfile>` + body + `</mxfile>`,
		"no attributes with newline":     "<mxfile>\n  " + body + "\n</mxfile>",
		"xml decl with newline":          "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<mxfile>" + body + `</mxfile>`,
		"xml decl without newline":       `<?xml version="1.0" encoding="UTF-8"?><mxfile>` + body + `</mxfile>`,
	}

	logger := slog.New(slogtest.NewTestHandler(t))
	parseOpts := &ParseOptions{CompositeDeliverableTable: &pfd.CompositeDeliverableTable{}}

	for name, xml := range testCases {
		t.Run(name, func(t *testing.T) {
			pngBytes := buildPNGWithMxfile(t, xml)
			if _, err := Parse("png", bytes.NewReader(pngBytes), parseOpts, logger); err != nil {
				t.Fatalf("Parse failed for %s: %v", name, err)
			}
		})
	}
}
