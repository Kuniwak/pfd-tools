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

func TestParse_NilCompositeDeliverableTableIsError(t *testing.T) {

	const xml = `<mxfile host="x"><diagram id="d1" name="P0"><mxGraphModel><root><mxCell id="0"/><mxCell id="1" parent="0"/></root></mxGraphModel></diagram></mxfile>`

	testCases := map[string]*ParseOptions{
		"nil options":                     nil,
		"options with nil cd table":       {},
		"options with explicit nil field": {CompositeDeliverableTable: nil},
	}

	for name, parseOpts := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))

			if _, err := Parse("", bytes.NewReader([]byte(xml)), parseOpts, logger); err == nil {
				t.Fatal("expected an error for a nil composite deliverable table, got nil")
			}
		})
	}
}

func TestParse_EmptyCompositeDeliverableTable(t *testing.T) {

	const xml = `<mxfile host="x"><diagram id="d1" name="P0"><mxGraphModel><root><mxCell id="0"/><mxCell id="1" parent="0"/></root></mxGraphModel></diagram></mxfile>`

	logger := slog.New(slogtest.NewTestHandler(t))
	parseOpts := &ParseOptions{CompositeDeliverableTable: &pfd.CompositeDeliverableTable{}}

	p, err := Parse("", bytes.NewReader([]byte(xml)), parseOpts, logger)
	if err != nil {
		t.Fatalf("Parse returned error for empty composite deliverable table: %v", err)
	}
	if p.DeliverableComposition == nil {
		t.Errorf("DeliverableComposition should be a non-nil empty map")
	}
	if len(p.DeliverableComposition) != 0 {
		t.Errorf("DeliverableComposition should be empty, got %d entries", len(p.DeliverableComposition))
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

	if !bytes.Contains(pngBytes, []byte(url.QueryEscape(xml))) {
		t.Errorf("PNG bytes do not contain url-escaped XML")
	}
}

func TestParse_PNGWithMxfileShapes(t *testing.T) {
	body := `<diagram id="d1" name="P0"><mxGraphModel><root><mxCell id="0"/><mxCell id="1" parent="0"/></root></mxGraphModel></diagram>`

	testCases := map[string]string{
		"no attributes":              `<mxfile>` + body + `</mxfile>`,
		"no attributes with newline": "<mxfile>\n  " + body + "\n</mxfile>",
		"xml decl with newline":      "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<mxfile>" + body + `</mxfile>`,
		"xml decl without newline":   `<?xml version="1.0" encoding="UTF-8"?><mxfile>` + body + `</mxfile>`,
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

func TestParse_AllowDetachedDetailPage(t *testing.T) {

	const fragment = `<mxfile host="x"><diagram id="d1" name="P1: 設計する"><mxGraphModel><root>` +
		`<mxCell id="0"/><mxCell id="1" parent="0"/>` +
		`<mxCell id="2" value="D1: 要求" style="rounded=0;whiteSpace=wrap;html=1;strokeWidth=1;" parent="1" vertex="1"><mxGeometry as="geometry"/></mxCell>` +
		`<mxCell id="3" value="P2: 方式を決める" style="ellipse;whiteSpace=wrap;html=1;strokeWidth=1;" parent="1" vertex="1"><mxGeometry as="geometry"/></mxCell>` +
		`<mxCell id="4" style="edgeStyle=none;html=1;" parent="1" source="2" target="3" edge="1"><mxGeometry as="geometry"/></mxCell>` +
		`</root></mxGraphModel></diagram></mxfile>`

	testCases := map[string]struct {
		AllowDetachedDetailPage bool
		AsPNG                   bool
		WantErr                 bool
	}{
		"a detached detail page is rejected by default":              {AllowDetachedDetailPage: false, WantErr: true},
		"a detached detail page is allowed with the option":          {AllowDetachedDetailPage: true, WantErr: false},
		"a detached detail page in a PNG is rejected by default":     {AllowDetachedDetailPage: false, AsPNG: true, WantErr: true},
		"a detached detail page in a PNG is allowed with the option": {AllowDetachedDetailPage: true, AsPNG: true, WantErr: false},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			parseOpts := &ParseOptions{
				CompositeDeliverableTable: &pfd.CompositeDeliverableTable{},
				AllowDetachedDetailPage:   tc.AllowDetachedDetailPage,
			}

			input := []byte(fragment)
			if tc.AsPNG {
				input = buildPNGWithMxfile(t, fragment)
			}

			p, err := Parse("", bytes.NewReader(input), parseOpts, logger)
			if tc.WantErr {
				if err == nil {
					t.Fatal("err = nil, want an error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if got := p.Nodes.Len(); got != 2 {
				t.Errorf("p.Nodes.Len() = %d, want 2", got)
			}
			if got := p.Edges.Len(); got != 1 {
				t.Errorf("p.Edges.Len() = %d, want 1", got)
			}
		})
	}
}
