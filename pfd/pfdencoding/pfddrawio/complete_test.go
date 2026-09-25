package pfddrawio_test

import (
	"bytes"
	"log/slog"
	"sort"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawio"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/Kuniwak/pfd-tools/xmldom"
	"github.com/google/go-cmp/cmp"
)

type vtx struct {
	ID pfd.NodeID
	X  float64
	Y  float64
}

func completeWriteVertices(t *testing.T, input string, pageName string) []vtx {
	t.Helper()
	logger := slog.New(slogtest.NewTestHandler(t))
	nodes, err := pfddrawio.CompleteCompositePages(bytes.NewReader([]byte(input)), logger)
	if err != nil {
		t.Fatalf("CompleteCompositePages: %v", err)
	}
	out := writeXMLNodes(t, nodes)
	return pageVertices(t, out, pageName)
}

func writeXMLNodes(t *testing.T, nodes []*xmldom.Node) []byte {
	t.Helper()
	buf := bytes.NewBuffer(nil)
	for _, n := range nodes {
		if err := n.Write(buf); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}
	return buf.Bytes()
}

func pageVertices(t *testing.T, xmlBytes []byte, pageName string) []vtx {
	t.Helper()
	roots, err := xmldom.ParseXML(bytes.NewReader(xmlBytes))
	if err != nil {
		t.Fatalf("ParseXML: %v", err)
	}

	var diagram *xmldom.Node
	for _, root := range roots {
		root.Traverse(func(n *xmldom.Node) {
			if n.Kind != xmldom.ElementNode || n.Start.Name.Local != "diagram" {
				return
			}
			name, _ := n.GetAttr("name", "")
			if pfddrawio.PageNameElementID(name) == pfddrawio.PageNameElementID(pageName) {
				diagram = n
			}
		}, nil)
	}
	if diagram == nil {
		return nil
	}

	layerMap := pfddrawio.NewLayerMapFromNodes([]*xmldom.Node{diagram})
	parents := pfddrawio.NewParentMapFromNodes([]*xmldom.Node{diagram})

	out := make([]vtx, 0)
	diagram.Traverse(func(n *xmldom.Node) {
		if n.Kind != xmldom.ElementNode || n.Start.Name.Local != "mxCell" {
			return
		}
		if v, _ := n.GetAttr("vertex", ""); v == "" {
			return
		}
		idAttr, _ := n.GetAttr("id", "")
		if layerMap.IsCommentDescendant(parents, pfddrawio.CellID(idAttr)) {
			return
		}
		id, ok := pfddrawio.VertexNodeID(n)
		if !ok {
			return
		}
		geo := n.FirstChildElement("mxGeometry")
		if geo == nil {
			return
		}
		x, _ := geo.FloatAttr("x")
		y, _ := geo.FloatAttr("y")
		out = append(out, vtx{ID: id, X: x, Y: y})
	}, nil)

	sort.Slice(out, func(i, j int) bool { return out[i].ID.Compare(out[j].ID) < 0 })
	return out
}

const inputNoPage = `<mxfile host="test">
    <diagram id="p0" name="P0">
        <mxGraphModel dx="800" dy="600" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
            <root>
                <mxCell id="0"/>
                <mxCell id="1" value="PFD" parent="0"/>
                <mxCell id="d1" value="D1: 入力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="40" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="p1" value="P1: 複合" style="ellipse;whiteSpace=wrap;html=1;strokeWidth=2;" parent="1" vertex="1">
                    <mxGeometry x="200" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="d2" value="D2: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="360" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="e1" style="edgeStyle=none;html=1;" parent="1" source="d1" target="p1" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
                <mxCell id="e2" style="edgeStyle=none;html=1;" parent="1" source="p1" target="d2" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
            </root>
        </mxGraphModel>
    </diagram>
</mxfile>
`

const inputUnnumbered = `<mxfile host="test">
    <diagram id="p0" name="P0">
        <mxGraphModel dx="800" dy="600" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
            <root>
                <mxCell id="0"/>
                <mxCell id="1" value="PFD" parent="0"/>
                <mxCell id="d1" value="要件" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="40" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="p1" value="実装する" style="ellipse;whiteSpace=wrap;html=1;strokeWidth=2;" parent="1" vertex="1">
                    <mxGeometry x="200" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="d2" value="実装する" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="360" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="e1" style="edgeStyle=none;html=1;" parent="1" source="d1" target="p1" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
                <mxCell id="e2" style="edgeStyle=none;html=1;" parent="1" source="p1" target="d2" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
            </root>
        </mxGraphModel>
    </diagram>
</mxfile>
`

func TestCompleteCompositePages_Scaffold(t *testing.T) {
	testCases := map[string]struct {
		input string
		page  string
		want  []vtx
	}{
		"new page: input left, output right": {
			input: inputNoPage,
			page:  "P1",
			want:  []vtx{{ID: "D1", X: 40, Y: 40}, {ID: "D2", X: 360, Y: 40}},
		},

		"an unnumbered composite process gets a page named with the label": {
			input: inputUnnumbered,
			page:  "実装する",
			want:  []vtx{{ID: "[要件]", X: 40, Y: 40}, {ID: "[実装する]", X: 360, Y: 40}},
		},
		"HTML label is stripped before deriving the ID": {
			input: inputHTMLLabel,
			page:  "P1",
			want:  []vtx{{ID: "D1", X: 40, Y: 40}, {ID: "D2", X: 360, Y: 40}},
		},
		"feedback boundary deliverable appears on its drawn side": {
			input: inputFeedback,
			page:  "P1",
			want:  []vtx{{ID: "D1", X: 40, Y: 40}, {ID: "D2", X: 360, Y: 40}, {ID: "D3", X: 360, Y: 160}},
		},
		"missing input placed at leftmost x above topmost y": {
			input: inputMissingInput,
			page:  "P1",
			want:  []vtx{{ID: "D1", X: 200, Y: 80}, {ID: "D2", X: 360, Y: 200}, {ID: "P2", X: 200, Y: 200}},
		},
		"missing output placed at rightmost x above topmost y": {
			input: inputMissingOutput,
			page:  "P1",
			want:  []vtx{{ID: "D1", X: 40, Y: 200}, {ID: "D2", X: 200, Y: 80}, {ID: "P2", X: 200, Y: 200}},
		},
		"comment box ignored for bounds": {
			input: inputCommentBounds,
			page:  "P1",
			want:  []vtx{{ID: "D1", X: 200, Y: 80}, {ID: "D2", X: 360, Y: 200}, {ID: "P2", X: 200, Y: 200}},
		},
		"empty existing page falls back to columns": {
			input: inputEmptyPage,
			page:  "P1",
			want:  []vtx{{ID: "D1", X: 40, Y: 40}, {ID: "D2", X: 360, Y: 40}},
		},
		"deliverable that is both input and output is placed once on the input side": {
			input: inputInOutOverlap,
			page:  "P1",
			want:  []vtx{{ID: "D1", X: 40, Y: 40}, {ID: "D2", X: 40, Y: 160}},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := completeWriteVertices(t, tc.input, tc.page)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("page %s vertices mismatch (-want +got):\n%s", tc.page, diff)
			}
		})
	}
}

const inputMissingInput = `<mxfile host="test">
    <diagram id="p0" name="P0">
        <mxGraphModel dx="800" dy="600" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
            <root>
                <mxCell id="0"/>
                <mxCell id="1" value="PFD" parent="0"/>
                <mxCell id="d1" value="D1: 入力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="40" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="p1" value="P1: 複合" style="ellipse;whiteSpace=wrap;html=1;strokeWidth=2;" parent="1" vertex="1">
                    <mxGeometry x="200" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="d2" value="D2: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="360" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="e1" style="edgeStyle=none;html=1;" parent="1" source="d1" target="p1" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
                <mxCell id="e2" style="edgeStyle=none;html=1;" parent="1" source="p1" target="d2" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
            </root>
        </mxGraphModel>
    </diagram>
    <diagram id="pp1" name="P1">
        <mxGraphModel dx="800" dy="600" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
            <root>
                <mxCell id="0"/>
                <mxCell id="1" value="PFD" parent="0"/>
                <mxCell id="pp1-p2" value="P2: 工程" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="200" y="200" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="pp1-d2" value="D2: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="360" y="200" width="120" height="80" as="geometry"/>
                </mxCell>
            </root>
        </mxGraphModel>
    </diagram>
</mxfile>
`

const inputDescribedPageName = `<mxfile host="test">
    <diagram id="p0" name="P0">
        <mxGraphModel dx="800" dy="600" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
            <root>
                <mxCell id="0"/>
                <mxCell id="1" value="PFD" parent="0"/>
                <mxCell id="d1" value="D1: 入力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="40" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="p1" value="P1: 複合" style="ellipse;whiteSpace=wrap;html=1;strokeWidth=2;" parent="1" vertex="1">
                    <mxGeometry x="200" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="d2" value="D2: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="360" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="e1" style="edgeStyle=none;html=1;" parent="1" source="d1" target="p1" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
                <mxCell id="e2" style="edgeStyle=none;html=1;" parent="1" source="p1" target="d2" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
            </root>
        </mxGraphModel>
    </diagram>
    <diagram id="pp1" name="P1: 複合">
        <mxGraphModel dx="800" dy="600" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
            <root>
                <mxCell id="0"/>
                <mxCell id="1" value="PFD" parent="0"/>
                <mxCell id="pp1-p2" value="P2: 工程" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="200" y="200" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="pp1-d2" value="D2: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="360" y="200" width="120" height="80" as="geometry"/>
                </mxCell>
            </root>
        </mxGraphModel>
    </diagram>
</mxfile>
`

var inputMismatchedPageName = strings.Replace(inputDescribedPageName, `name="P1: 複合"`, `name="P1: ちがう説明"`, 1)

const inputMissingOutput = `<mxfile host="test">
    <diagram id="p0" name="P0">
        <mxGraphModel dx="800" dy="600" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
            <root>
                <mxCell id="0"/>
                <mxCell id="1" value="PFD" parent="0"/>
                <mxCell id="d1" value="D1: 入力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="40" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="p1" value="P1: 複合" style="ellipse;whiteSpace=wrap;html=1;strokeWidth=2;" parent="1" vertex="1">
                    <mxGeometry x="200" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="d2" value="D2: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="360" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="e1" style="edgeStyle=none;html=1;" parent="1" source="d1" target="p1" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
                <mxCell id="e2" style="edgeStyle=none;html=1;" parent="1" source="p1" target="d2" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
            </root>
        </mxGraphModel>
    </diagram>
    <diagram id="pp1" name="P1">
        <mxGraphModel dx="800" dy="600" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
            <root>
                <mxCell id="0"/>
                <mxCell id="1" value="PFD" parent="0"/>
                <mxCell id="pp1-p2" value="P2: 工程" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="200" y="200" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="pp1-d1" value="D1: 入力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="40" y="200" width="120" height="80" as="geometry"/>
                </mxCell>
            </root>
        </mxGraphModel>
    </diagram>
</mxfile>
`

func TestCompleteCompositePages_PageAttrs(t *testing.T) {
	testCases := map[string]struct {
		input string
		attr  string
		want  []string
	}{
		"a new page is named after the composite process label": {
			input: inputNoPage,
			attr:  "name",
			want:  []string{"P0", "P1: 複合"},
		},

		"an unnumbered composite process gets a page named with the label": {
			input: inputUnnumbered,
			attr:  "name",
			want:  []string{"P0", "実装する"},
		},
		"an existing page named with the bare ID is kept as is": {
			input: inputMissingInput,
			attr:  "name",
			want:  []string{"P0", "P1"},
		},
		"an existing page named with a description is not duplicated": {
			input: inputDescribedPageName,
			attr:  "name",
			want:  []string{"P0", "P1: 複合"},
		},

		"an existing page keeps its name even if the description differs": {
			input: inputMismatchedPageName,
			attr:  "name",
			want:  []string{"P0", "P1: ちがう説明"},
		},
		"a new page id has no description": {
			input: inputNoPage,
			attr:  "id",
			want:  []string{"p0", "cpcomp-P1"},
		},
		"a new page id of an unnumbered composite process is its label": {
			input: inputUnnumbered,
			attr:  "id",
			want:  []string{"p0", "cpcomp-実装する"},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			nodes, err := pfddrawio.CompleteCompositePages(bytes.NewReader([]byte(tc.input)), logger)
			if err != nil {
				t.Fatalf("CompleteCompositePages: %v", err)
			}
			got := pageAttrs(t, writeXMLNodes(t, nodes), tc.attr)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("page %s mismatch (-want +got):\n%s", tc.attr, diff)
			}
		})
	}
}

func pageAttrs(t *testing.T, xmlBytes []byte, attr string) []string {
	t.Helper()
	roots, err := xmldom.ParseXML(bytes.NewReader(xmlBytes))
	if err != nil {
		t.Fatalf("ParseXML: %v", err)
	}

	values := make([]string, 0)
	for _, root := range roots {
		root.Traverse(func(n *xmldom.Node) {
			if n.Kind != xmldom.ElementNode || n.Start.Name.Local != "diagram" {
				return
			}
			value, _ := n.GetAttr(attr, "")
			values = append(values, value)
		}, nil)
	}
	return values
}

func TestCompleteCompositePages_Idempotent(t *testing.T) {
	testCases := map[string]string{
		"new page path":   inputNoPage,
		"completion path": inputMissingInput,
	}
	for name, input := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))

			nodes1, err := pfddrawio.CompleteCompositePages(bytes.NewReader([]byte(input)), logger)
			if err != nil {
				t.Fatalf("first run: %v", err)
			}
			out1 := writeXMLNodes(t, nodes1)

			nodes2, err := pfddrawio.CompleteCompositePages(bytes.NewReader(out1), logger)
			if err != nil {
				t.Fatalf("second run: %v", err)
			}
			out2 := writeXMLNodes(t, nodes2)

			if diff := cmp.Diff(string(out1), string(out2)); diff != "" {
				t.Errorf("second run changed the output (-first +second):\n%s", diff)
			}
		})
	}
}

const inputHTMLLabel = `<mxfile host="test">
    <diagram id="p0" name="P0">
        <mxGraphModel dx="800" dy="600" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
            <root>
                <mxCell id="0"/>
                <mxCell id="1" value="PFD" parent="0"/>
                <mxCell id="d1" value="&lt;b&gt;D1&lt;/b&gt;: 入力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="40" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="p1" value="P1: 複合" style="ellipse;whiteSpace=wrap;html=1;strokeWidth=2;" parent="1" vertex="1">
                    <mxGeometry x="200" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="d2" value="D2: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="360" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="e1" style="edgeStyle=none;html=1;" parent="1" source="d1" target="p1" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
                <mxCell id="e2" style="edgeStyle=none;html=1;" parent="1" source="p1" target="d2" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
            </root>
        </mxGraphModel>
    </diagram>
</mxfile>
`

const inputFeedback = `<mxfile host="test">
    <diagram id="p0" name="P0">
        <mxGraphModel dx="800" dy="600" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
            <root>
                <mxCell id="0"/>
                <mxCell id="1" value="PFD" parent="0"/>
                <mxCell id="d1" value="D1: 入力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="40" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="p1" value="P1: 複合" style="ellipse;whiteSpace=wrap;html=1;strokeWidth=2;" parent="1" vertex="1">
                    <mxGeometry x="200" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="d2" value="D2: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="360" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="d3" value="D3: 差戻し" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="360" y="200" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="e1" style="edgeStyle=none;html=1;" parent="1" source="d1" target="p1" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
                <mxCell id="e2" style="edgeStyle=none;html=1;" parent="1" source="p1" target="d2" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
                <mxCell id="e3" style="edgeStyle=none;html=1;dashed=1;" parent="1" source="p1" target="d3" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
            </root>
        </mxGraphModel>
    </diagram>
</mxfile>
`

const inputCommentBounds = `<mxfile host="test">
    <diagram id="p0" name="P0">
        <mxGraphModel dx="800" dy="600" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
            <root>
                <mxCell id="0"/>
                <mxCell id="1" value="PFD" parent="0"/>
                <mxCell id="d1" value="D1: 入力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="40" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="p1" value="P1: 複合" style="ellipse;whiteSpace=wrap;html=1;strokeWidth=2;" parent="1" vertex="1">
                    <mxGeometry x="200" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="d2" value="D2: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="360" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="e1" style="edgeStyle=none;html=1;" parent="1" source="d1" target="p1" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
                <mxCell id="e2" style="edgeStyle=none;html=1;" parent="1" source="p1" target="d2" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
            </root>
        </mxGraphModel>
    </diagram>
    <diagram id="pp1" name="P1">
        <mxGraphModel dx="800" dy="600" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
            <root>
                <mxCell id="0"/>
                <mxCell id="1" value="PFD" parent="0"/>
                <mxCell id="cmt" value="コメント" parent="0"/>
                <mxCell id="cmtbox" value="メモ" style="rounded=0;whiteSpace=wrap;html=1;" parent="cmt" vertex="1">
                    <mxGeometry x="0" y="0" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="pp1-p2" value="P2: 工程" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="200" y="200" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="pp1-d2" value="D2: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="360" y="200" width="120" height="80" as="geometry"/>
                </mxCell>
            </root>
        </mxGraphModel>
    </diagram>
</mxfile>
`

const inputMultiComposite = `<mxfile host="test">
    <diagram id="p0" name="P0">
        <mxGraphModel dx="800" dy="600" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
            <root>
                <mxCell id="0"/>
                <mxCell id="1" value="PFD" parent="0"/>
                <mxCell id="d1" value="D1: 入力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="40" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="p1" value="P1: 複合1" style="ellipse;whiteSpace=wrap;html=1;strokeWidth=2;" parent="1" vertex="1">
                    <mxGeometry x="200" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="d2" value="D2: 中間" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="360" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="p3" value="P3: 複合2" style="ellipse;whiteSpace=wrap;html=1;strokeWidth=2;" parent="1" vertex="1">
                    <mxGeometry x="520" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="d3" value="D3: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="680" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="e1" style="edgeStyle=none;html=1;" parent="1" source="d1" target="p1" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
                <mxCell id="e2" style="edgeStyle=none;html=1;" parent="1" source="p1" target="d2" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
                <mxCell id="e3" style="edgeStyle=none;html=1;" parent="1" source="d2" target="p3" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
                <mxCell id="e4" style="edgeStyle=none;html=1;" parent="1" source="p3" target="d3" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
            </root>
        </mxGraphModel>
    </diagram>
</mxfile>
`

func TestCompleteCompositePages_MultipleComposites(t *testing.T) {
	logger := slog.New(slogtest.NewTestHandler(t))
	nodes, err := pfddrawio.CompleteCompositePages(bytes.NewReader([]byte(inputMultiComposite)), logger)
	if err != nil {
		t.Fatalf("CompleteCompositePages: %v", err)
	}
	out := writeXMLNodes(t, nodes)

	gotP1 := pageVertices(t, out, "P1")
	wantP1 := []vtx{
		{ID: "D1", X: 40, Y: 40},
		{ID: "D2", X: 360, Y: 40},
	}
	if diff := cmp.Diff(wantP1, gotP1); diff != "" {
		t.Errorf("page P1 vertices mismatch (-want +got):\n%s", diff)
	}

	gotP3 := pageVertices(t, out, "P3")
	wantP3 := []vtx{
		{ID: "D2", X: 40, Y: 40},
		{ID: "D3", X: 360, Y: 40},
	}
	if diff := cmp.Diff(wantP3, gotP3); diff != "" {
		t.Errorf("page P3 vertices mismatch (-want +got):\n%s", diff)
	}
}

const inputEmptyPage = `<mxfile host="test">
    <diagram id="p0" name="P0">
        <mxGraphModel dx="800" dy="600" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
            <root>
                <mxCell id="0"/>
                <mxCell id="1" value="PFD" parent="0"/>
                <mxCell id="d1" value="D1: 入力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="40" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="p1" value="P1: 複合" style="ellipse;whiteSpace=wrap;html=1;strokeWidth=2;" parent="1" vertex="1">
                    <mxGeometry x="200" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="d2" value="D2: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="360" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="e1" style="edgeStyle=none;html=1;" parent="1" source="d1" target="p1" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
                <mxCell id="e2" style="edgeStyle=none;html=1;" parent="1" source="p1" target="d2" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
            </root>
        </mxGraphModel>
    </diagram>
    <diagram id="pp1" name="P1">
        <mxGraphModel dx="800" dy="600" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
            <root>
                <mxCell id="0"/>
                <mxCell id="1" value="PFD" parent="0"/>
            </root>
        </mxGraphModel>
    </diagram>
</mxfile>
`

const inputInOutOverlap = `<mxfile host="test">
    <diagram id="p0" name="P0">
        <mxGraphModel dx="800" dy="600" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
            <root>
                <mxCell id="0"/>
                <mxCell id="1" value="PFD" parent="0"/>
                <mxCell id="d1" value="D1: 入力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="40" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="p1" value="P1: 複合" style="ellipse;whiteSpace=wrap;html=1;strokeWidth=2;" parent="1" vertex="1">
                    <mxGeometry x="200" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="d2" value="D2: 循環" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
                    <mxGeometry x="360" y="40" width="120" height="80" as="geometry"/>
                </mxCell>
                <mxCell id="e1" style="edgeStyle=none;html=1;" parent="1" source="d1" target="p1" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
                <mxCell id="e2" style="edgeStyle=none;html=1;" parent="1" source="p1" target="d2" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
                <mxCell id="e3" style="edgeStyle=none;html=1;dashed=1;" parent="1" source="d2" target="p1" edge="1">
                    <mxGeometry relative="1" as="geometry"/>
                </mxCell>
            </root>
        </mxGraphModel>
    </diagram>
</mxfile>
`
