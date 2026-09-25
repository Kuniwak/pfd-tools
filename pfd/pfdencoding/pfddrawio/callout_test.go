package pfddrawio_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"testing"

	"github.com/Kuniwak/pfd-tools/geom"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawio"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/Kuniwak/pfd-tools/sugar"
	"github.com/Kuniwak/pfd-tools/xmldom"
	"github.com/google/go-cmp/cmp"
)

type box struct {
	Value  string
	Style  string
	Parent string
	Rect   geom.Rect
}

func collectEstimateBoxes(nodes []*xmldom.Node) []box {
	var boxes []box
	for _, root := range nodes {
		root.Traverse(func(n *xmldom.Node) {
			if n.Kind != xmldom.ElementNode || n.Start.Name.Local != "mxCell" {
				return
			}
			if v, _ := n.GetAttr("vertex", ""); v != "1" {
				return
			}
			value, _ := n.GetAttr("value", "")
			styleStr, _ := n.GetAttr("style", "")
			style, err := pfddrawio.ParseStyle(styleStr)
			if err != nil || !pfddrawio.IsEstimateBox(style, value) {
				return
			}
			parent, _ := n.GetAttr("parent", "")
			rect, _ := pfddrawio.GeometryRect(n)
			boxes = append(boxes, box{Value: value, Style: styleStr, Parent: parent, Rect: rect})
		}, nil)
	}
	sort.Slice(boxes, func(i, j int) bool {
		if boxes[i].Rect.Y != boxes[j].Rect.Y {
			return boxes[i].Rect.Y < boxes[j].Rect.Y
		}
		return boxes[i].Rect.X < boxes[j].Rect.X
	})
	return boxes
}

func callout(t *testing.T, xmlBytes []byte) []*xmldom.Node {
	t.Helper()
	logger := slog.New(slogtest.NewTestHandler(t))
	return sugar.Must(pfddrawio.Callout(bytes.NewReader(xmlBytes), logger))
}

const compositeMultiPageXML = `<mxfile host="test">
    <diagram id="pg0" name="P0">
        <mxGraphModel>
            <root>
                <mxCell id="0"/>
                <mxCell id="1" parent="0"/>
                <mxCell id="2" value="D1: 入力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="40" y="40" width="120" height="80" as="geometry"/></mxCell>
                <mxCell id="3" value="P1: 複合" style="ellipse;whiteSpace=wrap;html=1;strokeWidth=2;" parent="1" vertex="1"><mxGeometry x="200" y="40" width="120" height="80" as="geometry"/></mxCell>
                <mxCell id="4" style="edgeStyle=none;html=1;" parent="1" source="2" target="3" edge="1"><mxGeometry relative="1" as="geometry"/></mxCell>
            </root>
        </mxGraphModel>
    </diagram>
    <diagram id="pg1" name="P1">
        <mxGraphModel>
            <root>
                <mxCell id="0"/>
                <mxCell id="1" parent="0"/>
                <mxCell id="12" value="D2: 中間" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="40" y="40" width="120" height="80" as="geometry"/></mxCell>
                <mxCell id="13" value="P2: 原子" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="200" y="40" width="120" height="80" as="geometry"/></mxCell>
                <mxCell id="14" style="edgeStyle=none;html=1;" parent="1" source="12" target="13" edge="1"><mxGeometry relative="1" as="geometry"/></mxCell>
            </root>
        </mxGraphModel>
    </diagram>
</mxfile>`

func TestCalloutAddsBoxes(t *testing.T) {
	testCases := map[string]struct {
		Input []byte
		Want  []box
	}{

		"simple: one box below the only atomic process": {
			Input: sugar.Must(os.ReadFile("../../../testdata/simple/pfd.drawio")),
			Want: []box{
				{Value: pfddrawio.EstimateBoxPlaceholder, Style: pfddrawio.EstimateBoxStyle, Parent: "1", Rect: geom.Rect{X: 485, Y: 320, Width: 110, Height: 30}},
			},
		},

		"composite/multi-page: box only below the atomic process on the detail page": {
			Input: []byte(compositeMultiPageXML),
			Want: []box{
				{Value: pfddrawio.EstimateBoxPlaceholder, Style: pfddrawio.EstimateBoxStyle, Parent: "1", Rect: geom.Rect{X: 205, Y: 120, Width: 110, Height: 30}},
			},
		},

		"comment layer memo does not suppress placeholders": {
			Input: twoProcessChainXML(
				`<mxCell id="90" value="コメント" parent="0"/>` +
					textBoxCellXML("91", "90", "楽観すぎ？悲観すぎ？あとで確認", 205, 150),
			),
			Want: []box{
				{Value: pfddrawio.EstimateBoxPlaceholder, Style: pfddrawio.EstimateBoxStyle, Parent: "1", Rect: geom.Rect{X: 205, Y: 120, Width: 110, Height: 30}},
				{Value: pfddrawio.EstimateBoxPlaceholder, Style: pfddrawio.EstimateBoxStyle, Parent: "1", Rect: geom.Rect{X: 525, Y: 120, Width: 110, Height: 30}},
				{Value: "楽観すぎ？悲観すぎ？あとで確認", Style: pfddrawio.EstimateBoxStyle, Parent: "90", Rect: geom.Rect{X: 205, Y: 150, Width: 110, Height: 30}},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := collectEstimateBoxes(callout(t, tc.Input))
			if diff := cmp.Diff(tc.Want, got); diff != "" {
				t.Errorf("estimate boxes mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestEstimateBoxes(t *testing.T) {
	src := sugar.Must(os.ReadFile("../../../testdata/callout/pfd.drawio"))
	nodes := sugar.Must(xmldom.ParseXML(bytes.NewReader(src)))
	got := pfddrawio.EstimateBoxes(nodes)
	want := []pfddrawio.EstimateBox{
		{DiagramID: "wRU_aafd9vpDkhm-03GV", CellID: "18", Value: "楽観: 2d 悲観: 3d", Rect: geom.Rect{X: 485, Y: 320, Width: 110, Height: 30}},
		{DiagramID: "wRU_aafd9vpDkhm-03GV", CellID: "21", Value: "楽観: 1d 悲観: 5d", Rect: geom.Rect{X: 805, Y: 320, Width: 110, Height: 30}},
		{DiagramID: "wRU_aafd9vpDkhm-03GV", CellID: "20", Value: "楽観: 1d 悲観: 2d", Rect: geom.Rect{X: 485, Y: 440, Width: 110, Height: 30}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("EstimateBoxes mismatch (-want +got):\n%s", diff)
	}
}

func TestCalloutIdempotent(t *testing.T) {
	testCases := map[string][]byte{
		"simple":               sugar.Must(os.ReadFile("../../../testdata/simple/pfd.drawio")),
		"composite/multi-page": []byte(compositeMultiPageXML),
	}

	for name, input := range testCases {
		t.Run(name, func(t *testing.T) {
			once := sugar.Must(xmldom.Marshal(callout(t, input)))
			twice := sugar.Must(xmldom.Marshal(callout(t, once)))

			gotOnce := pfddrawio.EstimateBoxes(callout(t, input))
			gotTwice := pfddrawio.EstimateBoxes(callout(t, once))
			if len(gotOnce) != len(gotTwice) {
				t.Errorf("box count changed on re-run: once=%d twice=%d", len(gotOnce), len(gotTwice))
			}
			if diff := cmp.Diff(once, twice); diff != "" {
				t.Errorf("output changed on re-run (-once +twice):\n%s", diff)
			}
		})
	}
}

func TestCalloutPreservesExistingValues(t *testing.T) {
	input := sugar.Must(os.ReadFile("../../../testdata/callout/pfd.drawio"))

	before := pfddrawio.EstimateBoxes(sugar.Must(xmldom.ParseXML(bytes.NewReader(input))))
	after := pfddrawio.EstimateBoxes(callout(t, input))

	if diff := cmp.Diff(before, after); diff != "" {
		t.Errorf("existing estimate boxes changed (-before +after):\n%s", diff)
	}
	if len(after) != 3 {
		t.Errorf("estimate boxes = %d, want 3 (unchanged)", len(after))
	}
}

func TestCalloutDoesNotAffectModel(t *testing.T) {
	testCases := map[string][]byte{
		"simple":               sugar.Must(os.ReadFile("../../../testdata/simple/pfd.drawio")),
		"callout":              sugar.Must(os.ReadFile("../../../testdata/callout/pfd.drawio")),
		"composite/multi-page": []byte(compositeMultiPageXML),
	}

	for name, input := range testCases {
		t.Run(name, func(t *testing.T) {
			before := modelSummary(t, input)
			after := modelSummary(t, sugar.Must(xmldom.Marshal(callout(t, input))))
			if diff := cmp.Diff(before, after); diff != "" {
				t.Errorf("PFD model changed (-before +after):\n%s", diff)
			}
		})
	}
}

func modelSummary(t *testing.T, xmlBytes []byte) []string {
	t.Helper()
	logger := slog.New(slogtest.NewTestHandler(t))
	p, _, err := pfddrawio.ParseExceptCompositeDeliverables("", bytes.NewReader(xmlBytes), logger)
	if err != nil {
		t.Fatalf("ParseExceptCompositeDeliverables: %v", err)
	}
	var lines []string
	for _, n := range p.Nodes.Slice() {
		lines = append(lines, fmt.Sprintf("node %s|%s|%s", n.ID, n.Type, n.Description))
	}
	for _, e := range p.Edges.Slice() {
		lines = append(lines, fmt.Sprintf("edge %s->%s feedback=%t", e.Source, e.Target, e.IsFeedback))
	}
	sort.Strings(lines)
	return lines
}
