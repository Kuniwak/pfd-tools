package pfddrawio_test

import (
	"bytes"
	"errors"
	"io/fs"
	"log/slog"
	"math"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/geom"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawio"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/Kuniwak/pfd-tools/xmldom"
	"github.com/google/go-cmp/cmp"
)

func testLogger(t *testing.T) *slog.Logger {
	return slog.New(slogtest.NewTestHandler(t))
}

func sortAndReadRects(t *testing.T, doc string, opts pfddrawio.SortOptions) map[pfddrawio.CellID]geom.Rect {
	t.Helper()
	nodes, err := pfddrawio.Sort(strings.NewReader(doc), opts, testLogger(t))
	if err != nil {
		t.Fatalf("Sort: %v", err)
	}
	layers := pfddrawio.NewLayerMapFromNodes(nodes)
	parents := pfddrawio.NewParentMapFromNodes(nodes)
	doms := pfddrawio.CollectDiagramDOMs(nodes)
	dom, ok := doms["d0"]
	if !ok {
		t.Fatalf("diagram d0 not found after Sort")
	}
	vertices, _ := pfddrawio.CollectPageLayout(dom, layers, parents)
	got := make(map[pfddrawio.CellID]geom.Rect, len(vertices))
	for _, v := range vertices {
		got[v.ID] = v.Rect
	}
	return got
}

func drawioPage(cells string) string {
	return `<mxfile><diagram id="d0" name="P0"><mxGraphModel><root>` +
		`<mxCell id="0"/><mxCell id="1" parent="0"/>` +
		cells +
		`</root></mxGraphModel></diagram></mxfile>`
}

func rectCell(id, value string, x, y float64) string {
	return `<mxCell id="` + id + `" value="` + value + `" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">` +
		mxGeom(x, y, 120, 80) + `</mxCell>`
}

func ellipseCell(id, value string, x, y float64) string {
	return `<mxCell id="` + id + `" value="` + value + `" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1">` +
		mxGeom(x, y, 120, 80) + `</mxCell>`
}

func rectCellSized(id, value string, x, y, h float64) string {
	return `<mxCell id="` + id + `" value="` + value + `" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">` +
		mxGeom(x, y, 120, h) + `</mxCell>`
}

func ellipseCellSized(id, value string, x, y, h float64) string {
	return `<mxCell id="` + id + `" value="` + value + `" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1">` +
		mxGeom(x, y, 120, h) + `</mxCell>`
}

func edgeCell(id, source, target string, dashed bool) string {
	style := "edgeStyle=none;html=1;"
	if dashed {
		style = "edgeStyle=none;html=1;dashed=1;"
	}
	return `<mxCell id="` + id + `" value="" style="` + style + `" parent="1" source="` + source + `" target="` + target + `" edge="1"><mxGeometry relative="1" as="geometry"/></mxCell>`
}

func estimateBoxCell(id string, px, py float64) string {
	return `<mxCell id="` + id + `" value="楽観: d 悲観: d" style="text;html=1;align=center;verticalAlign=middle;whiteSpace=wrap;rounded=0;" parent="1" vertex="1">` +
		mxGeom(px+5, py+80, 110, 30) + `</mxCell>`
}

func mxGeom(x, y, w, h float64) string {
	return `<mxGeometry x="` + ftoa(x) + `" y="` + ftoa(y) + `" width="` + ftoa(w) + `" height="` + ftoa(h) + `" as="geometry"/>`
}

func ftoa(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func rect(x, y float64) geom.Rect {
	return geom.Rect{X: x, Y: y, Width: 120, Height: 80}
}

func collectFirstPage(t *testing.T, doc string) ([]pfddrawio.LayoutVertex, []pfddrawio.LayoutEdge) {
	t.Helper()
	nodes, err := xmldom.ParseXML(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("ParseXML: %v", err)
	}
	layers := pfddrawio.NewLayerMapFromNodes(nodes)
	parents := pfddrawio.NewParentMapFromNodes(nodes)
	doms := pfddrawio.CollectDiagramDOMs(nodes)
	dom, ok := doms["d0"]
	if !ok {
		t.Fatalf("diagram d0 not found")
	}
	return pfddrawio.CollectPageLayout(dom, layers, parents)
}

func TestCollectPageLayout(t *testing.T) {
	testCases := map[string]struct {
		Doc              string
		ExpectedVertices []pfddrawio.LayoutVertex
		ExpectedEdges    []pfddrawio.LayoutEdge
	}{
		"serial D->P->D": {
			Doc: drawioPage(
				rectCell("2", "D2: 入力", 160, 240) +
					ellipseCell("3", "P1: 加工", 320, 240) +
					rectCell("5", "D1: 出力", 480, 240) +
					edgeCell("4", "2", "3", false) +
					edgeCell("6", "3", "5", false),
			),
			ExpectedVertices: []pfddrawio.LayoutVertex{
				{ID: "2", NodeID: "D2", Kind: pfddrawio.VertexDeliverable, Rect: rect(160, 240)},
				{ID: "3", NodeID: "P1", Kind: pfddrawio.VertexProcess, Rect: rect(320, 240)},
				{ID: "5", NodeID: "D1", Kind: pfddrawio.VertexDeliverable, Rect: rect(480, 240)},
			},
			ExpectedEdges: []pfddrawio.LayoutEdge{
				{ID: "4", Source: "2", Target: "3", IsFeedback: false},
				{ID: "6", Source: "3", Target: "5", IsFeedback: false},
			},
		},
		"feedback edge is separated by dashed style": {
			Doc: drawioPage(
				rectCell("2", "D1: 種", 160, 240) +
					ellipseCell("3", "P1: 検査", 320, 240) +
					edgeCell("4", "2", "3", false) +
					edgeCell("7", "3", "2", true),
			),
			ExpectedVertices: []pfddrawio.LayoutVertex{
				{ID: "2", NodeID: "D1", Kind: pfddrawio.VertexDeliverable, Rect: rect(160, 240)},
				{ID: "3", NodeID: "P1", Kind: pfddrawio.VertexProcess, Rect: rect(320, 240)},
			},
			ExpectedEdges: []pfddrawio.LayoutEdge{
				{ID: "4", Source: "2", Target: "3", IsFeedback: false},
				{ID: "7", Source: "3", Target: "2", IsFeedback: true},
			},
		},
		"edge to non-layout cell is dropped": {
			Doc: drawioPage(
				rectCell("2", "D1: 種", 160, 240) +
					ellipseCell("3", "P1: 検査", 320, 240) +
					edgeCell("4", "2", "3", false) +
					edgeCell("9", "3", "999", false),
			),
			ExpectedVertices: []pfddrawio.LayoutVertex{
				{ID: "2", NodeID: "D1", Kind: pfddrawio.VertexDeliverable, Rect: rect(160, 240)},
				{ID: "3", NodeID: "P1", Kind: pfddrawio.VertexProcess, Rect: rect(320, 240)},
			},
			ExpectedEdges: []pfddrawio.LayoutEdge{
				{ID: "4", Source: "2", Target: "3", IsFeedback: false},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			vertices, edges := collectFirstPage(t, tc.Doc)
			if diff := cmp.Diff(tc.ExpectedVertices, vertices); diff != "" {
				t.Errorf("vertices mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.ExpectedEdges, edges); diff != "" {
				t.Errorf("edges mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func del(id pfddrawio.CellID) pfddrawio.LayoutVertex {
	return pfddrawio.LayoutVertex{ID: id, Kind: pfddrawio.VertexDeliverable}
}

func proc(id pfddrawio.CellID) pfddrawio.LayoutVertex {
	return pfddrawio.LayoutVertex{ID: id, Kind: pfddrawio.VertexProcess}
}

func delNamed(id pfddrawio.CellID, nodeID pfd.NodeID) pfddrawio.LayoutVertex {
	return pfddrawio.LayoutVertex{ID: id, NodeID: nodeID, Kind: pfddrawio.VertexDeliverable}
}

func procNamed(id pfddrawio.CellID, nodeID pfd.NodeID) pfddrawio.LayoutVertex {
	return pfddrawio.LayoutVertex{ID: id, NodeID: nodeID, Kind: pfddrawio.VertexProcess}
}

func conn(id pfddrawio.CellID) pfddrawio.LayoutVertex {
	return pfddrawio.LayoutVertex{ID: id, Kind: pfddrawio.VertexConnector}
}

func fwd(id, src, tgt pfddrawio.CellID) pfddrawio.LayoutEdge {
	return pfddrawio.LayoutEdge{ID: id, Source: src, Target: tgt}
}

func back(id, src, tgt pfddrawio.CellID) pfddrawio.LayoutEdge {
	return pfddrawio.LayoutEdge{ID: id, Source: src, Target: tgt, IsFeedback: true}
}

func cut(id, src, tgt pfddrawio.CellID) pfddrawio.LayoutEdge {
	return pfddrawio.LayoutEdge{ID: id, Source: src, Target: tgt, IsCut: true}
}

func TestSortPositions(t *testing.T) {
	opts := pfddrawio.DefaultSortOptions()

	testCases := map[string]struct {
		Doc      string
		Expected map[pfddrawio.CellID]geom.Rect
	}{
		"serial is spread left-to-right by rank": {

			Doc: drawioPage(
				rectCell("2", "D2: 入力", 500, 90) +
					ellipseCell("3", "P1: 加工", 100, 400) +
					rectCell("5", "D1: 出力", 900, 250) +
					edgeCell("4", "2", "3", false) +
					edgeCell("6", "3", "5", false),
			),

			Expected: map[pfddrawio.CellID]geom.Rect{
				"2": {X: 100, Y: 90, Width: 120, Height: 80},
				"3": {X: 300, Y: 90, Width: 120, Height: 80},
				"5": {X: 500, Y: 90, Width: 120, Height: 80},
			},
		},
		"same-rank nodes stack; branch source snaps to the nearer row": {

			Doc: drawioPage(
				rectCell("2", "D1: 種", 160, 240) +
					ellipseCell("3", "P1: 甲", 320, 400) +
					ellipseCell("4", "P2: 乙", 320, 100) +
					edgeCell("5", "2", "3", false) +
					edgeCell("6", "2", "4", false),
			),

			Expected: map[pfddrawio.CellID]geom.Rect{
				"2": {X: 160, Y: 220, Width: 120, Height: 80},
				"4": {X: 360, Y: 100, Width: 120, Height: 80},
				"3": {X: 360, Y: 220, Width: 120, Height: 80},
			},
		},
		"a chain into a merge stays exactly horizontal": {

			Doc: drawioPage(
				rectCell("2", "D1: 甲", 100, 100) +
					ellipseCell("3", "P1: 一", 300, 100) +
					rectCell("6", "D2: 乙", 100, 300) +
					ellipseCell("4", "P2: 二", 300, 300) +
					rectCell("7", "D3: 丙", 100, 500) +
					ellipseCell("8", "P3: 三", 300, 500) +
					rectCell("5", "D4: 丁", 500, 200) +
					edgeCell("9", "2", "3", false) +
					edgeCell("10", "3", "5", false) +
					edgeCell("11", "6", "4", false) +
					edgeCell("12", "4", "5", false) +
					edgeCell("13", "7", "8", false) +
					edgeCell("14", "8", "5", false),
			),

			Expected: map[pfddrawio.CellID]geom.Rect{
				"2": {X: 100, Y: 100, Width: 120, Height: 80},
				"3": {X: 300, Y: 100, Width: 120, Height: 80},
				"6": {X: 100, Y: 220, Width: 120, Height: 80},
				"4": {X: 300, Y: 220, Width: 120, Height: 80},
				"7": {X: 100, Y: 340, Width: 120, Height: 80},
				"8": {X: 300, Y: 340, Width: 120, Height: 80},
				"5": {X: 500, Y: 220, Width: 120, Height: 80},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := sortAndReadRects(t, tc.Doc, opts)
			if diff := cmp.Diff(tc.Expected, got); diff != "" {
				t.Errorf("positions mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestComputePositionsKeepsOriginY(t *testing.T) {

	const originY = 100
	sized := func(id pfddrawio.CellID, kind pfddrawio.VertexKind, y, h float64) pfddrawio.LayoutVertex {
		return pfddrawio.LayoutVertex{ID: id, Kind: kind, Rect: geom.Rect{Y: y, Width: 120, Height: h}}
	}

	testCases := map[string]struct {
		Vertices []pfddrawio.LayoutVertex
		Edges    []pfddrawio.LayoutEdge
	}{
		"tallest vertex is on the top row": {
			Vertices: []pfddrawio.LayoutVertex{
				sized("d0", pfddrawio.VertexDeliverable, 0, 80),
				sized("p0", pfddrawio.VertexProcess, 0, 80),
				sized("d1", pfddrawio.VertexDeliverable, 200, 40),
				sized("p1", pfddrawio.VertexProcess, 200, 40),
			},
			Edges: []pfddrawio.LayoutEdge{fwd("e0", "d0", "p0"), fwd("e1", "d1", "p1")},
		},
		"tallest vertex is on a lower row": {
			Vertices: []pfddrawio.LayoutVertex{
				sized("d0", pfddrawio.VertexDeliverable, 0, 40),
				sized("p0", pfddrawio.VertexProcess, 0, 40),
				sized("d1", pfddrawio.VertexDeliverable, 200, 40),
				sized("p1", pfddrawio.VertexProcess, 200, 80),
			},
			Edges: []pfddrawio.LayoutEdge{fwd("e0", "d0", "p0"), fwd("e1", "d1", "p1")},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ranks, err := pfddrawio.AssignRanks(tc.Vertices, tc.Edges)
			if err != nil {
				t.Fatalf("AssignRanks: %v", err)
			}
			positions, _ := pfddrawio.ComputePositions(tc.Vertices, tc.Edges, ranks, 0, originY, pfddrawio.DefaultSortOptions())
			minY := math.Inf(1)
			for _, r := range positions {
				if r.Y < minY {
					minY = r.Y
				}
			}
			if minY != originY {
				t.Errorf("top of the laid out page must stay at originY: want %v, got %v", float64(originY), minY)
			}
		})
	}
}

func delAt(id pfddrawio.CellID, y float64) pfddrawio.LayoutVertex {
	return pfddrawio.LayoutVertex{ID: id, Kind: pfddrawio.VertexDeliverable, Rect: geom.Rect{Y: y, Width: 120, Height: 80}}
}

func procAt(id pfddrawio.CellID, y float64) pfddrawio.LayoutVertex {
	return pfddrawio.LayoutVertex{ID: id, Kind: pfddrawio.VertexProcess, Rect: geom.Rect{Y: y, Width: 120, Height: 80}}
}

func TestEstimateRows(t *testing.T) {

	testCases := map[string]struct {
		Vertices []pfddrawio.LayoutVertex
		Edges    []pfddrawio.LayoutEdge
		Expected map[pfddrawio.CellID]int
	}{
		"parallel chains keep their own rows": {
			Vertices: []pfddrawio.LayoutVertex{
				delAt("d1", 0), procAt("p1", 0),
				delAt("d2", 1000), procAt("p2", 1000),
			},
			Edges: []pfddrawio.LayoutEdge{
				fwd("e1", "d1", "p1"), fwd("e2", "d2", "p2"),
			},
			Expected: map[pfddrawio.CellID]int{"d1": 0, "p1": 0, "d2": 1, "p2": 1},
		},
		"branch producer sits at the center row of its consumers": {
			Vertices: []pfddrawio.LayoutVertex{
				delAt("d1", 0),
				procAt("p1", 0), procAt("p2", 1000), procAt("p3", 2000),
			},
			Edges: []pfddrawio.LayoutEdge{
				fwd("e1", "d1", "p1"), fwd("e2", "d1", "p2"), fwd("e3", "d1", "p3"),
			},
			Expected: map[pfddrawio.CellID]int{"d1": 1, "p1": 0, "p2": 1, "p3": 2},
		},
	}

	opts := pfddrawio.DefaultSortOptions()
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ranks, err := pfddrawio.AssignRanks(tc.Vertices, tc.Edges)
			if err != nil {
				t.Fatalf("AssignRanks: %v", err)
			}
			got := pfddrawio.EstimateRows(tc.Vertices, tc.Edges, ranks, opts)
			if diff := cmp.Diff(tc.Expected, got); diff != "" {
				t.Errorf("rows mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestEstimateRowsIgnoresVertexHeight(t *testing.T) {

	sized := func(id pfddrawio.CellID, kind pfddrawio.VertexKind, y, h float64) pfddrawio.LayoutVertex {
		return pfddrawio.LayoutVertex{ID: id, Kind: kind, Rect: geom.Rect{Y: y, Width: 120, Height: h}}
	}
	rowsFor := func(t *testing.T, d0Height float64) map[pfddrawio.CellID]int {
		t.Helper()
		vertices := []pfddrawio.LayoutVertex{
			sized("d0", pfddrawio.VertexDeliverable, 0, d0Height),
			sized("d1", pfddrawio.VertexDeliverable, 200, 800),
			sized("d2", pfddrawio.VertexDeliverable, 400, 800),
			sized("p0", pfddrawio.VertexProcess, 0, 800),
			sized("p1", pfddrawio.VertexProcess, 400, 800),
			sized("d3", pfddrawio.VertexDeliverable, 0, 800),
		}
		edges := []pfddrawio.LayoutEdge{
			fwd("e0", "d0", "p0"), fwd("e1", "d1", "p0"), fwd("e2", "d2", "p1"),
			fwd("e3", "d0", "p1"), fwd("e4", "p0", "d3"), fwd("e5", "p1", "d3"),
		}
		ranks, err := pfddrawio.AssignRanks(vertices, edges)
		if err != nil {
			t.Fatalf("AssignRanks: %v", err)
		}
		return pfddrawio.EstimateRows(vertices, edges, ranks, pfddrawio.DefaultSortOptions())
	}

	expected := rowsFor(t, 800)

	testCases := map[string]struct{ D0Height float64 }{
		"short box":  {D0Height: 40},
		"medium box": {D0Height: 400},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if diff := cmp.Diff(expected, rowsFor(t, tc.D0Height)); diff != "" {
				t.Errorf("rows must not depend on box height (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGreedyBands(t *testing.T) {

	testCases := map[string]struct {
		Points   []int
		Span     int
		Expected []int
	}{
		"empty":                              {Points: nil, Span: 5, Expected: nil},
		"single point is band 0":             {Points: []int{7}, Span: 5, Expected: []int{0}},
		"gap at the threshold stays in band": {Points: []int{0, 5}, Span: 5, Expected: []int{0, 0}},
		"gap over the threshold splits":      {Points: []int{0, 6}, Span: 5, Expected: []int{0, 1}},

		"distance is measured from the band top": {Points: []int{0, 3, 6, 9, 12}, Span: 5, Expected: []int{0, 0, 1, 1, 2}},
		"equal points share a band":              {Points: []int{4, 4, 4}, Span: 0, Expected: []int{0, 0, 0}},
		"span 0 splits any gap":                  {Points: []int{0, 1, 1, 2}, Span: 0, Expected: []int{0, 1, 1, 2}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := pfddrawio.GreedyBands(tc.Points, tc.Span)
			if diff := cmp.Diff(tc.Expected, got); diff != "" {
				t.Errorf("bands mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPlanDuplications(t *testing.T) {

	vertices := []pfddrawio.LayoutVertex{
		del("1"), proc("2"), del("3"), proc("4"), del("5"),
	}
	edges := []pfddrawio.LayoutEdge{
		fwd("e1", "1", "2"), fwd("e2", "2", "3"),
		fwd("e3", "3", "4"), fwd("e4", "4", "5"),
		fwd("far", "1", "4"),
	}
	ranks, err := pfddrawio.AssignRanks(vertices, edges)
	if err != nil {
		t.Fatalf("AssignRanks: %v", err)
	}

	testCases := map[string]struct {
		DupRankSpan int
		WantDup     bool
	}{
		"span over threshold duplicates and rewires": {DupRankSpan: 1, WantDup: true},
		"raising threshold disables duplication":     {DupRankSpan: 3, WantDup: false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			dups, rewires := pfddrawio.PlanDuplications(vertices, edges, pfddrawio.DupCriteria{Ranks: ranks, RankSpan: tc.DupRankSpan})
			if !tc.WantDup {
				if len(dups) != 0 || len(rewires) != 0 {
					t.Errorf("threshold %d must not duplicate span=3, got dups=%+v rewires=%+v", tc.DupRankSpan, dups, rewires)
				}
				return
			}
			if len(dups) != 1 {
				t.Fatalf("want 1 duplicate, got %d: %+v", len(dups), dups)
			}
			if dups[0].SourceID != "1" {
				t.Errorf("want duplicate of D1, got source %q", dups[0].SourceID)
			}
			if dups[0].Rank != 2 {
				t.Errorf("want duplicate at rank 2 (adjacent to P2), got %d", dups[0].Rank)
			}
			if len(rewires) != 1 || rewires[0].EdgeID != "far" || rewires[0].Endpoint != pfddrawio.EndpointSource {
				t.Errorf("want far edge source rewired, got %+v", rewires)
			}
			if rewires[0].NewCell != dups[0].TempID {
				t.Errorf("rewire should point to the duplicate temp id")
			}
		})
	}
}

func TestPlanDuplicationsFeedback(t *testing.T) {

	chain := []pfddrawio.LayoutVertex{
		del("1"), proc("2"), del("3"), proc("4"), del("5"), proc("6"), del("7"),
	}
	chainEdges := []pfddrawio.LayoutEdge{
		fwd("e1", "1", "2"), fwd("e2", "2", "3"), fwd("e3", "3", "4"),
		fwd("e4", "4", "5"), fwd("e5", "5", "6"), fwd("e6", "6", "7"),
	}
	ranks := map[pfddrawio.CellID]int{"1": 0, "2": 1, "3": 2, "4": 3, "5": 4, "6": 5, "7": 6}

	testCases := map[string]struct {
		Vertices    []pfddrawio.LayoutVertex
		Edges       []pfddrawio.LayoutEdge
		Ranks       map[pfddrawio.CellID]int
		Rows        map[pfddrawio.CellID]int
		RankSpan    int
		RowSpan     int
		WantDups    []pfddrawio.DuplicatePlan
		WantRewires int
	}{
		"a far feedback edge duplicates right of its process": {

			Vertices: chain,
			Edges:    append(append([]pfddrawio.LayoutEdge{}, chainEdges...), back("fb", "7", "2")),
			Ranks:    ranks,
			RankSpan: 1,
			WantDups: []pfddrawio.DuplicatePlan{{SourceID: "7", Rank: 2}},

			WantRewires: 1,
		},
		"a feedback edge within the threshold is not duplicated": {

			Vertices:    chain,
			Edges:       append(append([]pfddrawio.LayoutEdge{}, chainEdges...), back("fb", "3", "2")),
			Ranks:       ranks,
			RankSpan:    1,
			WantDups:    nil,
			WantRewires: 0,
		},
		"a deliverable with only feedback edges keeps its nearest edge": {

			Vertices: append(append([]pfddrawio.LayoutVertex{}, chain...), del("9")),
			Edges: append(append([]pfddrawio.LayoutEdge{}, chainEdges...),
				back("fb1", "9", "2"), back("fb2", "9", "4")),
			Ranks:       map[pfddrawio.CellID]int{"1": 0, "2": 1, "3": 2, "4": 3, "5": 4, "6": 5, "7": 6, "9": 4},
			RankSpan:    1,
			WantDups:    []pfddrawio.DuplicatePlan{{SourceID: "9", Rank: 2}},
			WantRewires: 1,
		},
		"a vertically far feedback edge duplicates too": {

			Vertices:    []pfddrawio.LayoutVertex{del("9"), proc("8"), proc("10")},
			Edges:       []pfddrawio.LayoutEdge{back("fb", "9", "8"), fwd("keep", "9", "10")},
			Ranks:       map[pfddrawio.CellID]int{"9": 2, "8": 1, "10": 3},
			Rows:        map[pfddrawio.CellID]int{"9": 6, "8": 0, "10": 6},
			RankSpan:    1,
			RowSpan:     5,
			WantDups:    []pfddrawio.DuplicatePlan{{SourceID: "9", Rank: 2}},
			WantRewires: 1,
		},
		"the original box keeps a solid edge rather than a feedback edge": {

			Vertices: []pfddrawio.LayoutVertex{del("9"), proc("8"), proc("10")},

			Edges:       []pfddrawio.LayoutEdge{back("fb", "9", "8"), fwd("s", "9", "10")},
			Ranks:       map[pfddrawio.CellID]int{"9": 0, "8": 3, "10": 3},
			RankSpan:    1,
			WantDups:    []pfddrawio.DuplicatePlan{{SourceID: "9", Rank: 4}},
			WantRewires: 1,
		},
		"a feedback edge and a solid edge get separate duplicates": {

			Vertices: append(append([]pfddrawio.LayoutVertex{}, chain...), del("9")),
			Edges: append(append([]pfddrawio.LayoutEdge{}, chainEdges...),
				fwd("s", "9", "6"), back("fb", "9", "4"), fwd("keep", "9", "2")),
			Ranks:    map[pfddrawio.CellID]int{"1": 0, "2": 1, "3": 2, "4": 3, "5": 4, "6": 5, "7": 6, "9": 0},
			RankSpan: 1,
			WantDups: []pfddrawio.DuplicatePlan{
				{SourceID: "9", Rank: 4},
				{SourceID: "9", Rank: 4},
			},
			WantRewires: 2,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			crit := pfddrawio.DupCriteria{
				Ranks: tc.Ranks, Rows: tc.Rows, RankSpan: tc.RankSpan, RowSpan: tc.RowSpan,
			}
			dups, rewires := pfddrawio.PlanDuplications(tc.Vertices, tc.Edges, crit)

			var got []pfddrawio.DuplicatePlan
			for _, d := range dups {
				got = append(got, pfddrawio.DuplicatePlan{SourceID: d.SourceID, Rank: d.Rank})
			}
			if diff := cmp.Diff(tc.WantDups, got); diff != "" {
				t.Errorf("duplicates mismatch (-want +got):\n%s", diff)
			}
			if len(rewires) != tc.WantRewires {
				t.Errorf("want %d rewires, got %d: %+v", tc.WantRewires, len(rewires), rewires)
			}
		})
	}
}

func TestPlanDuplicationsKeepsOriginalConnected(t *testing.T) {

	testCases := map[string]struct {
		Vertices    []pfddrawio.LayoutVertex
		Edges       []pfddrawio.LayoutEdge
		DupRankSpan int
		WantDups    int
		WantRewires int
	}{

		"source with only a far edge is not duplicated": {
			Vertices: []pfddrawio.LayoutVertex{
				del("1"), proc("2"), del("3"), proc("4"), del("5"), del("s"),
			},
			Edges: []pfddrawio.LayoutEdge{
				fwd("e1", "1", "2"), fwd("e2", "2", "3"),
				fwd("e3", "3", "4"), fwd("e4", "4", "5"),
				fwd("far", "s", "4"),
			},
			DupRankSpan: 1,
			WantDups:    0,
			WantRewires: 0,
		},

		"deliverable with a near edge still duplicates its far edge": {
			Vertices: []pfddrawio.LayoutVertex{
				del("1"), proc("2"), del("3"), proc("4"), del("5"),
			},
			Edges: []pfddrawio.LayoutEdge{
				fwd("e1", "1", "2"), fwd("e2", "2", "3"),
				fwd("e3", "3", "4"), fwd("e4", "4", "5"),
				fwd("far", "1", "4"),
			},
			DupRankSpan: 1,
			WantDups:    1,
			WantRewires: 1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ranks, err := pfddrawio.AssignRanks(tc.Vertices, tc.Edges)
			if err != nil {
				t.Fatalf("AssignRanks: %v", err)
			}
			dups, rewires := pfddrawio.PlanDuplications(tc.Vertices, tc.Edges, pfddrawio.DupCriteria{Ranks: ranks, RankSpan: tc.DupRankSpan})
			if len(dups) != tc.WantDups {
				t.Errorf("want %d duplicates, got %d: %+v", tc.WantDups, len(dups), dups)
			}
			if len(rewires) != tc.WantRewires {
				t.Errorf("want %d rewires, got %d: %+v", tc.WantRewires, len(rewires), rewires)
			}
		})
	}
}

func TestPlanDuplicationsKeepsSolidChain(t *testing.T) {
	testCases := map[string]struct {
		Vertices    []pfddrawio.LayoutVertex
		Edges       []pfddrawio.LayoutEdge
		Ranks       map[pfddrawio.CellID]int
		Rows        map[pfddrawio.CellID]int
		WantDups    int
		WantRewires []pfddrawio.CellID
	}{
		"a produced deliverable duplicates its only far consumption edge": {

			Vertices: []pfddrawio.LayoutVertex{proc("p0"), del("d"), proc("pfar")},
			Edges:    []pfddrawio.LayoutEdge{fwd("prod", "p0", "d"), fwd("cons", "d", "pfar")},
			Ranks:    map[pfddrawio.CellID]int{"p0": 1, "d": 2, "pfar": 6},
			WantDups: 1, WantRewires: []pfddrawio.CellID{"cons"},
		},
		"a produced deliverable keeps its only vertically far consumption edge": {

			Vertices: []pfddrawio.LayoutVertex{proc("p0"), del("d"), proc("pfar")},
			Edges:    []pfddrawio.LayoutEdge{fwd("prod", "p0", "d"), fwd("cons", "d", "pfar")},
			Ranks:    map[pfddrawio.CellID]int{"p0": 1, "d": 2, "pfar": 3},
			Rows:     map[pfddrawio.CellID]int{"p0": 0, "d": 0, "pfar": 8},
			WantDups: 0, WantRewires: nil,
		},
		"a produced deliverable duplicates both far consumers in other columns": {

			Vertices: []pfddrawio.LayoutVertex{proc("p0"), del("d"), proc("pa"), proc("pb")},
			Edges: []pfddrawio.LayoutEdge{
				fwd("prod", "p0", "d"), fwd("fara", "d", "pa"), fwd("farb", "d", "pb"),
			},
			Ranks:    map[pfddrawio.CellID]int{"p0": 1, "d": 2, "pa": 5, "pb": 7},
			WantDups: 2, WantRewires: []pfddrawio.CellID{"fara", "farb"},
		},
		"a produced deliverable keeps the same-column consumer and duplicates the far one": {

			Vertices: []pfddrawio.LayoutVertex{proc("p0"), del("d"), proc("pv"), proc("pfar")},
			Edges: []pfddrawio.LayoutEdge{
				fwd("prod", "p0", "d"), fwd("vert", "d", "pv"), fwd("far", "d", "pfar"),
			},
			Ranks:    map[pfddrawio.CellID]int{"p0": 1, "d": 2, "pv": 3, "pfar": 6},
			Rows:     map[pfddrawio.CellID]int{"p0": 0, "d": 0, "pv": 8, "pfar": 0},
			WantDups: 1, WantRewires: []pfddrawio.CellID{"far"},
		},
		"a produced deliverable with a near consumption edge still duplicates far edges": {

			Vertices: []pfddrawio.LayoutVertex{proc("p0"), del("d"), proc("pn"), proc("pfar")},
			Edges: []pfddrawio.LayoutEdge{
				fwd("prod", "p0", "d"), fwd("near", "d", "pn"), fwd("far", "d", "pfar"),
			},
			Ranks:    map[pfddrawio.CellID]int{"p0": 1, "d": 2, "pn": 3, "pfar": 6},
			WantDups: 1, WantRewires: []pfddrawio.CellID{"far"},
		},
		"a consumed deliverable keeps its only far production edge": {

			Vertices: []pfddrawio.LayoutVertex{proc("pp"), del("d"), proc("pn")},
			Edges:    []pfddrawio.LayoutEdge{fwd("prod", "pp", "d"), fwd("cons", "d", "pn")},
			Ranks:    map[pfddrawio.CellID]int{"pp": 1, "d": 2, "pn": 3},
			Rows:     map[pfddrawio.CellID]int{"pp": 8, "d": 0, "pn": 0},
			WantDups: 0, WantRewires: nil,
		},
		"a consumed deliverable duplicates its only cross-column production edge": {

			Vertices: []pfddrawio.LayoutVertex{proc("pp"), del("d"), proc("pn")},
			Edges:    []pfddrawio.LayoutEdge{fwd("prod", "pp", "d"), fwd("cons", "d", "pn")},
			Ranks:    map[pfddrawio.CellID]int{"pp": 1, "d": 5, "pn": 6},
			WantDups: 1, WantRewires: []pfddrawio.CellID{"prod"},
		},
		"a source deliverable is not affected": {

			Vertices: []pfddrawio.LayoutVertex{del("d"), proc("pn"), proc("pfar")},
			Edges:    []pfddrawio.LayoutEdge{fwd("near", "d", "pn"), fwd("far", "d", "pfar")},
			Ranks:    map[pfddrawio.CellID]int{"d": 0, "pn": 1, "pfar": 4},
			WantDups: 1, WantRewires: []pfddrawio.CellID{"far"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			crit := pfddrawio.DupCriteria{Ranks: tc.Ranks, Rows: tc.Rows, RankSpan: 1, RowSpan: 5}
			dups, rewires := pfddrawio.PlanDuplications(tc.Vertices, tc.Edges, crit)
			if len(dups) != tc.WantDups {
				t.Errorf("want %d duplicates, got %d: %+v", tc.WantDups, len(dups), dups)
			}
			var gotEdges []pfddrawio.CellID
			for _, r := range rewires {
				gotEdges = append(gotEdges, r.EdgeID)
			}
			if diff := cmp.Diff(tc.WantRewires, gotEdges); diff != "" {
				t.Errorf("rewired edges mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPlanDuplicationsVerticallyFar(t *testing.T) {

	vertices := []pfddrawio.LayoutVertex{
		del("d1"), proc("p1"), proc("p2"), proc("p3"),
	}
	edges := []pfddrawio.LayoutEdge{
		fwd("near", "d1", "p1"),
		fwd("vfar2", "d1", "p2"),
		fwd("vfar3", "d1", "p3"),
	}
	ranks := map[pfddrawio.CellID]int{"d1": 0, "p1": 1, "p2": 1, "p3": 1}

	testCases := map[string]struct {
		Rows        map[pfddrawio.CellID]int
		WantDups    int
		WantRewires int
	}{
		"row span at the threshold does not duplicate": {
			Rows:     map[pfddrawio.CellID]int{"d1": 0, "p1": 0, "p2": 5, "p3": 5},
			WantDups: 0, WantRewires: 0,
		},
		"row span over the threshold duplicates": {
			Rows:     map[pfddrawio.CellID]int{"d1": 0, "p1": 0, "p2": 6, "p3": 0},
			WantDups: 1, WantRewires: 1,
		},
		"consumers within a row band share one duplicate": {
			Rows:     map[pfddrawio.CellID]int{"d1": 0, "p1": 0, "p2": 6, "p3": 7},
			WantDups: 1, WantRewires: 2,
		},
		"consumers far apart get their own duplicates": {
			Rows:     map[pfddrawio.CellID]int{"d1": 0, "p1": 0, "p2": 6, "p3": 20},
			WantDups: 2, WantRewires: 2,
		},

		"unknown row skips the vertical criterion": {
			Rows:     map[pfddrawio.CellID]int{"p1": 0, "p2": 6, "p3": 20},
			WantDups: 0, WantRewires: 0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			crit := pfddrawio.DupCriteria{Ranks: ranks, Rows: tc.Rows, RankSpan: 1, RowSpan: 5}
			dups, rewires := pfddrawio.PlanDuplications(vertices, edges, crit)
			if len(dups) != tc.WantDups {
				t.Fatalf("want %d duplicates, got %d: %+v", tc.WantDups, len(dups), dups)
			}
			if len(rewires) != tc.WantRewires {
				t.Fatalf("want %d rewires, got %d: %+v", tc.WantRewires, len(rewires), rewires)
			}
			for _, d := range dups {

				if d.SourceID != "d1" || d.Rank != 0 {
					t.Errorf("want a duplicate of d1 at rank 0, got %+v", d)
				}
			}
			for _, r := range rewires {
				if r.Endpoint != pfddrawio.EndpointSource {
					t.Errorf("want the deliverable side rewired, got %+v", r)
				}
			}
			if tc.WantRewires == 2 {
				shared := rewires[0].NewCell == rewires[1].NewCell
				if shared != (tc.WantDups == 1) {
					t.Errorf("rewires sharing a duplicate = %v, want %v: %+v", shared, tc.WantDups == 1, rewires)
				}
			}
		})
	}
}

func pageVertexValues(t *testing.T, nodes []*xmldom.Node) map[pfddrawio.CellID]string {
	t.Helper()
	doms := pfddrawio.CollectDiagramDOMs(nodes)
	dom, ok := doms["d0"]
	if !ok {
		t.Fatalf("diagram d0 not found")
	}
	values := make(map[pfddrawio.CellID]string)
	for id, cell := range dom.CellsByID {
		if v, _ := cell.GetAttr("vertex", ""); v != "1" {
			continue
		}
		value, _ := cell.GetAttr("value", "")
		values[id] = value
	}
	return values
}

func pageVertexNodeIDs(t *testing.T, nodes []*xmldom.Node) map[pfddrawio.CellID]pfd.NodeID {
	t.Helper()
	doms := pfddrawio.CollectDiagramDOMs(nodes)
	dom, ok := doms["d0"]
	if !ok {
		t.Fatalf("diagram d0 not found")
	}
	nodeIDs := make(map[pfddrawio.CellID]pfd.NodeID)
	for id, cell := range dom.CellsByID {
		if v, _ := cell.GetAttr("vertex", ""); v != "1" {
			continue
		}
		if nodeID, ok := pfddrawio.VertexNodeID(cell); ok {
			nodeIDs[id] = nodeID
		}
	}
	return nodeIDs
}

func edgeEndpoints(t *testing.T, nodes []*xmldom.Node, edgeID pfddrawio.CellID) (string, string) {
	t.Helper()
	doms := pfddrawio.CollectDiagramDOMs(nodes)
	dom := doms["d0"]
	cell, ok := dom.CellsByID[edgeID]
	if !ok {
		t.Fatalf("edge %q not found", edgeID)
	}
	src, _ := cell.GetAttr("source", "")
	tgt, _ := cell.GetAttr("target", "")
	return src, tgt
}

func TestSortDuplicatesFarDeliverable(t *testing.T) {

	doc := drawioPage(
		rectCell("2", "D1: 素材", 0, 0) +
			ellipseCell("3", "P1: 加工", 0, 100) +
			rectCell("5", "D2: 中間", 0, 200) +
			ellipseCell("7", "P2: 仕上", 0, 300) +
			rectCell("9", "D3: 完成", 0, 400) +
			edgeCell("4", "2", "3", false) +
			edgeCell("6", "3", "5", false) +
			edgeCell("8", "5", "7", false) +
			edgeCell("10", "7", "9", false) +
			edgeCell("11", "2", "7", false),
	)

	testCases := map[string]struct {
		DupRankSpan int
		WantD1Count int
		WantRewired bool
	}{
		"default threshold adds one duplicate and rewires far edge": {DupRankSpan: pfddrawio.DefaultSortOptions().DupRankSpan, WantD1Count: 2, WantRewired: true},
		"high threshold keeps a single D1 and original edge":        {DupRankSpan: 10, WantD1Count: 1, WantRewired: false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			opts := pfddrawio.DefaultSortOptions()
			opts.DupRankSpan = tc.DupRankSpan
			nodes, err := pfddrawio.Sort(strings.NewReader(doc), opts, testLogger(t))
			if err != nil {
				t.Fatalf("Sort: %v", err)
			}
			nodeIDs := pageVertexNodeIDs(t, nodes)

			countD1 := 0
			for _, nodeID := range nodeIDs {
				if nodeID == "D1" {
					countD1++
				}
			}
			if countD1 != tc.WantD1Count {
				t.Errorf("want %d vertices labeled D1, got %d", tc.WantD1Count, countD1)
			}

			src, _ := edgeEndpoints(t, nodes, "11")
			if tc.WantRewired {
				if src == "2" || src == "" {
					t.Errorf("far edge source should be rewired to the duplicate, got %q", src)
				}
				if nodeIDs[pfddrawio.CellID(src)] != "D1" {
					t.Errorf("rewired source should be a D1 duplicate, got node ID %q", nodeIDs[pfddrawio.CellID(src)])
				}
			} else if src != "2" {
				t.Errorf("far edge source should stay original D1(2), got %q", src)
			}
		})
	}
}

func assertRowSpanExceeds(t *testing.T, doc string, a, b pfddrawio.CellID, span int) {
	t.Helper()
	vertices, edges := collectFirstPage(t, doc)
	ranks, err := pfddrawio.AssignRanks(vertices, edges)
	if err != nil {
		t.Fatalf("AssignRanks: %v", err)
	}
	rows := pfddrawio.EstimateRows(vertices, edges, ranks, pfddrawio.DefaultSortOptions())
	got := rows[b] - rows[a]
	if got < 0 {
		got = -got
	}
	if got <= span {
		t.Fatalf("fixture must place %s and %s more than %d rows apart, got %d (rows %s=%d %s=%d)",
			a, b, span, got, a, rows[a], b, rows[b])
	}
}

func verticallyFarDoc() string {
	cells := rectCell("d1", "D1: 素材", 0, 0)
	for i := 1; i <= 3; i++ {
		p := "t" + strconv.Itoa(i)
		cells += ellipseCell(p, "T"+strconv.Itoa(i)+": 上部", 200, float64(i*100)) +
			edgeCell("et"+strconv.Itoa(i), "d1", p, false)
	}
	for i := 1; i <= 20; i++ {
		b := "b" + strconv.Itoa(i)
		cells += rectCell(b, "B"+strconv.Itoa(i)+": 部品", 0, float64(1000+i*100)) +
			edgeCell("eb"+strconv.Itoa(i), b, "pb", false)
	}
	return drawioPage(cells +
		ellipseCell("pb", "PB: 集約", 200, 3000) +
		edgeCell("vfar", "d1", "pb", false))
}

func TestSortDuplicatesVerticallyFarDeliverable(t *testing.T) {

	doc := verticallyFarDoc()

	assertRowSpanExceeds(t, doc, "d1", "pb", pfddrawio.DefaultSortOptions().DupRowSpan)

	testCases := map[string]struct {
		DupRowSpan  int
		WantD1Count int
		WantRewired bool
	}{
		"default threshold duplicates the vertically far deliverable": {DupRowSpan: pfddrawio.DefaultSortOptions().DupRowSpan, WantD1Count: 2, WantRewired: true},
		"high threshold keeps a single D1 and the long edge":          {DupRowSpan: 100, WantD1Count: 1, WantRewired: false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			opts := pfddrawio.DefaultSortOptions()
			opts.DupRowSpan = tc.DupRowSpan
			nodes, err := pfddrawio.Sort(strings.NewReader(doc), opts, testLogger(t))
			if err != nil {
				t.Fatalf("Sort: %v", err)
			}
			nodeIDs := pageVertexNodeIDs(t, nodes)

			countD1 := 0
			for _, nodeID := range nodeIDs {
				if nodeID == "D1" {
					countD1++
				}
			}
			if countD1 != tc.WantD1Count {
				t.Errorf("want %d vertices labeled D1, got %d", tc.WantD1Count, countD1)
			}

			src, _ := edgeEndpoints(t, nodes, "vfar")
			if tc.WantRewired {
				if src == "d1" || src == "" {
					t.Errorf("vertically far edge source should be rewired to the duplicate, got %q", src)
				}
				if nodeIDs[pfddrawio.CellID(src)] != "D1" {
					t.Errorf("rewired source should be a D1 duplicate, got node ID %q", nodeIDs[pfddrawio.CellID(src)])
				}
			} else if src != "d1" {
				t.Errorf("vertically far edge source should stay original D1, got %q", src)
			}
		})
	}
}

func farSingleConsumerDoc(gap int) string {
	cells := rectCell("d0", "D0: 素材", 0, 0) +
		ellipseCell("p1", "P1: 加工", 200, 0) +
		rectCell("d1", "D1: 中間物", 400, 0) +
		edgeCell("e1", "d0", "p1", false) +
		edgeCell("prod", "p1", "d1", false) +
		rectCell("da0", "DA0: 別系統の素材", 0, 1000)
	prev := "da0"
	x := 200.0
	for i := 1; i < gap; i++ {
		p := "pa" + strconv.Itoa(i)
		d := "da" + strconv.Itoa(i)
		cells += ellipseCell(p, "PA"+strconv.Itoa(i)+": 別系統"+strconv.Itoa(i), x, 1000) +
			rectCell(d, "DA"+strconv.Itoa(i)+": 別系統中間物"+strconv.Itoa(i), x+200, 1000) +
			edgeCell("ep"+strconv.Itoa(i), prev, p, false) +
			edgeCell("ed"+strconv.Itoa(i), p, d, false)
		prev = d
		x += 400
	}
	return drawioPage(cells +
		ellipseCell("pfar", "PFAR: 組み上げ", x, 1000) +
		edgeCell("efar", prev, "pfar", false) +
		edgeCell("far", "d1", "pfar", false))
}

func TestSortDuplicatesFarSingleConsumer(t *testing.T) {
	testCases := map[string]struct {
		Gap         int
		DupRankSpan int
		WantDup     bool
	}{
		"a consumer three columns away duplicates": {
			Gap: 3, DupRankSpan: 1, WantDup: true,
		},
		"a consumer in the next column does not duplicate": {

			Gap: 1, DupRankSpan: 1, WantDup: false,
		},
		"a raised threshold still duplicates a consumer beyond it": {

			Gap: 3, DupRankSpan: 2, WantDup: true,
		},
		"a raised threshold suppresses a consumer within it": {
			Gap: 3, DupRankSpan: 3, WantDup: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			doc := farSingleConsumerDoc(tc.Gap)
			opts := pfddrawio.DefaultSortOptions()
			opts.DupRankSpan = tc.DupRankSpan

			assertRankSpanIs(t, doc, "d1", "pfar", tc.Gap)

			nodes, err := pfddrawio.Sort(strings.NewReader(doc), opts, testLogger(t))
			if err != nil {
				t.Fatalf("Sort: %v", err)
			}
			nodeIDs := pageVertexNodeIDs(t, nodes)
			rects := sortedRects(t, nodes)
			src, _ := edgeEndpoints(t, nodes, "far")

			if !tc.WantDup {
				if src != "d1" {
					t.Errorf("the far edge should stay on the original box, got source %q", src)
				}
				return
			}

			if src == "d1" {
				t.Fatalf("the far edge should be rewired to a duplicate, but it still starts at the original box")
			}
			if nodeIDs[pfddrawio.CellID(src)] != "D1" {
				t.Errorf("the rewired source should be a D1 duplicate, got node ID %q", nodeIDs[pfddrawio.CellID(src)])
			}

			colPitch := rects[pfddrawio.CellID(src)].Width + opts.HGap
			if got, want := rects[pfddrawio.CellID(src)].X, rects["pfar"].X-colPitch; got != want {
				t.Errorf("the duplicate should sit one column left of its consumer: x=%v, want %v", got, want)
			}
			if rects[pfddrawio.CellID(src)].X == rects["d1"].X {
				t.Errorf("the duplicate should not share the column of the original box: x=%v", rects["d1"].X)
			}

			if _, prodTarget := edgeEndpoints(t, nodes, "prod"); prodTarget != "d1" {
				t.Errorf("the production edge should stay on the original box, got target %q", prodTarget)
			}
		})
	}
}

func assertRankSpanIs(t *testing.T, doc string, a, b pfddrawio.CellID, span int) {
	t.Helper()
	vertices, edges := collectFirstPage(t, doc)
	ranks, err := pfddrawio.AssignDisplayRanks(vertices, edges)
	if err != nil {
		t.Fatalf("AssignDisplayRanks: %v", err)
	}
	if got := ranks[b] - ranks[a]; got != span {
		t.Fatalf("fixture must place %s and %s %d ranks apart, got %d (ranks %s=%d %s=%d)",
			a, b, span, got, a, ranks[a], b, ranks[b])
	}
}

func TestSortDuplicatePlacedNearConsumer(t *testing.T) {
	doc := verticallyFarDoc()

	opts := pfddrawio.DefaultSortOptions()
	nodes, err := pfddrawio.Sort(strings.NewReader(doc), opts, testLogger(t))
	if err != nil {
		t.Fatalf("Sort: %v", err)
	}

	nodeIDs := pageVertexNodeIDs(t, nodes)
	src, _ := edgeEndpoints(t, nodes, "vfar")
	if src == "d1" || nodeIDs[pfddrawio.CellID(src)] != "D1" {
		t.Fatalf("premise broken: vfar source should be a D1 duplicate, got %q", src)
	}

	rects := sortedRects(t, nodes)
	center := func(id pfddrawio.CellID) float64 { return rects[id].Y + rects[id].Height/2 }
	rowPitch := opts.VGap + rects["d1"].Height
	rowSpan := math.Abs(center(pfddrawio.CellID(src))-center("pb")) / rowPitch
	if rowSpan > float64(opts.DupRowSpan) {
		t.Errorf("duplicate should sit within %d rows of its consumer PB, got %.1f rows apart (dup y=%v, pb y=%v)",
			opts.DupRowSpan, rowSpan, rects[pfddrawio.CellID(src)].Y, rects["pb"].Y)
	}
}

func sortedRects(t *testing.T, nodes []*xmldom.Node) map[pfddrawio.CellID]geom.Rect {
	t.Helper()
	layers := pfddrawio.NewLayerMapFromNodes(nodes)
	parents := pfddrawio.NewParentMapFromNodes(nodes)
	doms := pfddrawio.CollectDiagramDOMs(nodes)
	dom, ok := doms["d0"]
	if !ok {
		t.Fatalf("diagram d0 not found")
	}
	vertices, _ := pfddrawio.CollectPageLayout(dom, layers, parents)
	rects := make(map[pfddrawio.CellID]geom.Rect, len(vertices))
	for _, v := range vertices {
		rects[v.ID] = v.Rect
	}
	return rects
}

func TestSortSourceSingleConsumerNotFloating(t *testing.T) {

	doc := drawioPage(
		rectCell("src", "D1: source", 0, 0) +
			rectCell("d2", "D2: seed", 0, 100) +
			ellipseCell("p1", "P1: a", 0, 200) +
			rectCell("d3", "D3: x", 0, 300) +
			ellipseCell("p2", "P2: b", 0, 400) +
			rectCell("d4", "D4: y", 0, 500) +
			ellipseCell("p3", "P3: c", 0, 600) +
			rectCell("d5", "D5: z", 0, 700) +
			edgeCell("e2", "d2", "p1", false) +
			edgeCell("e3", "p1", "d3", false) +
			edgeCell("e4", "d3", "p2", false) +
			edgeCell("e5", "p2", "d4", false) +
			edgeCell("e6", "d4", "p3", false) +
			edgeCell("e7", "p3", "d5", false) +
			edgeCell("esrc", "src", "p3", false),
	)

	opts := pfddrawio.DefaultSortOptions()
	nodes, err := pfddrawio.Sort(strings.NewReader(doc), opts, testLogger(t))
	if err != nil {
		t.Fatalf("Sort: %v", err)
	}
	values := pageVertexValues(t, nodes)

	countD1 := 0
	for _, v := range values {
		if v == "D1: source" {
			countD1++
		}
	}
	if countD1 != 1 {
		t.Errorf("source D1 must not be duplicated (would leave a floating box), want 1 vertex, got %d", countD1)
	}

	src, _ := edgeEndpoints(t, nodes, "esrc")
	if src != "src" {
		t.Errorf("D1 の唯一の辺は元の箱に残すべき, far edge source should stay original %q, got %q", "src", src)
	}
}

func TestSortEstimateBoxFollowsProcess(t *testing.T) {

	doc := drawioPage(
		rectCell("2", "D1: 入力", 0, 0) +
			ellipseCell("3", "P1: 加工", 0, 300) +
			estimateBoxCell("99", 0, 300) +
			edgeCell("4", "2", "3", false),
	)
	nodes, err := pfddrawio.Sort(strings.NewReader(doc), pfddrawio.DefaultSortOptions(), testLogger(t))
	if err != nil {
		t.Fatalf("Sort: %v", err)
	}

	procRect := rectOf(t, nodes, "d0", "3")
	boxRect := rectOf(t, nodes, "d0", "99")

	wantProc := geom.Rect{X: 200, Y: 0, Width: 120, Height: 80}
	if procRect != wantProc {
		t.Errorf("process rect = %+v, want %+v", procRect, wantProc)
	}
	wantBox := pfddrawio.EstimateBoxRect(procRect)
	if boxRect != wantBox {
		t.Errorf("estimate box rect = %+v, want %+v (below moved process)", boxRect, wantBox)
	}
	if !pfddrawio.IsBelow(procRect, boxRect) {
		t.Errorf("estimate box must stay below the process after sort")
	}
}

func diagramPage(id, name, cells string) string {
	return `<diagram id="` + id + `" name="` + name + `"><mxGraphModel><root>` +
		`<mxCell id="0"/><mxCell id="1" parent="0"/>` + cells +
		`</root></mxGraphModel></diagram>`
}

func rectOf(t *testing.T, nodes []*xmldom.Node, diagramID pfddrawio.DiagramID, cellID pfddrawio.CellID) geom.Rect {
	t.Helper()
	dom, ok := pfddrawio.CollectDiagramDOMs(nodes)[diagramID]
	if !ok {
		t.Fatalf("diagram %q not found", diagramID)
	}
	r, ok := pfddrawio.GeometryRect(dom.CellsByID[cellID])
	if !ok {
		t.Fatalf("geometry of %q not found", cellID)
	}
	return r
}

func TestSortOnlyPage(t *testing.T) {

	messy := func(dID, pID string) string {
		return rectCell(dID, "D1: 入力", 500, 90) +
			ellipseCell(pID, "P1: 加工", 100, 400) +
			edgeCell("e"+dID, dID, pID, false)
	}
	doc := `<mxfile>` +
		diagramPage("dA", "P0", messy("2", "3")) +
		diagramPage("dB", "P3", messy("12", "13")) +
		`</mxfile>`

	nodes, err := pfddrawio.Sort(strings.NewReader(doc), pfddrawio.SortOptions{OnlyPages: []string{"P3"}}, testLogger(t))
	if err != nil {
		t.Fatalf("Sort: %v", err)
	}

	if got := rectOf(t, nodes, "dA", "2"); got != rect(500, 90) {
		t.Errorf("unselected page A deliverable moved: got %+v, want original %+v", got, rect(500, 90))
	}
	if got := rectOf(t, nodes, "dA", "3"); got != rect(100, 400) {
		t.Errorf("unselected page A process moved: got %+v, want original %+v", got, rect(100, 400))
	}

	dB := rectOf(t, nodes, "dB", "12")
	pB := rectOf(t, nodes, "dB", "13")
	if !(dB.X < pB.X) {
		t.Errorf("selected page B not sorted left-to-right: D.X=%v P.X=%v", dB.X, pB.X)
	}
	if dB == rect(500, 90) {
		t.Errorf("selected page B deliverable should have moved from original")
	}
}

func TestShouldSortPage(t *testing.T) {
	testCases := map[string]struct {
		Name      string
		OnlyPages []string
		Want      bool
	}{
		"empty selection sorts every page":      {Name: "P3", OnlyPages: nil, Want: true},
		"exact page name matches":               {Name: "P3", OnlyPages: []string{"P1", "P3"}, Want: true},
		"non-listed page is skipped":            {Name: "P9", OnlyPages: []string{"P1", "P3"}, Want: false},
		"P0 matches locale context page name":   {Name: "ページ-1", OnlyPages: []string{"P0"}, Want: true},
		"context request does not match detail": {Name: "P3", OnlyPages: []string{"P0"}, Want: false},
		"ID matches described page name":        {Name: "P3: 実装する", OnlyPages: []string{"P3"}, Want: true},
		"described request matches bare ID":     {Name: "P3", OnlyPages: []string{"P3: 実装する"}, Want: true},
		"described request ignores description": {Name: "P3: 実装する", OnlyPages: []string{"P3: 設計する"}, Want: true},
		"other ID does not match":               {Name: "P3: 実装する", OnlyPages: []string{"P4"}, Want: false},
		"P0 matches described context page":     {Name: "P0: コンテキストダイアグラム", OnlyPages: []string{"P0"}, Want: true},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if got := pfddrawio.ShouldSortPage(tc.Name, tc.OnlyPages); got != tc.Want {
				t.Errorf("ShouldSortPage(%q, %v) = %v, want %v", tc.Name, tc.OnlyPages, got, tc.Want)
			}
		})
	}
}

func TestUnmatchedOnlyPages(t *testing.T) {

	withContext := []string{"ページ-1", "P3", "P7"}
	noContext := []string{"P3", "P7"}

	testCases := map[string]struct {
		OnlyPages []string
		Pages     []string
		Want      []string
	}{
		"all listed pages exist":           {OnlyPages: []string{"P3", "P7"}, Pages: withContext, Want: nil},
		"missing detail page is reported":  {OnlyPages: []string{"P3", "P999"}, Pages: withContext, Want: []string{"P999"}},
		"P0 matches existing context page": {OnlyPages: []string{"P0"}, Pages: withContext, Want: nil},
		"P0 reported when no context page": {OnlyPages: []string{"P0"}, Pages: noContext, Want: []string{"P0"}},
		"empty selection reports nothing":  {OnlyPages: nil, Pages: withContext, Want: nil},
		"ID matches described page name":   {OnlyPages: []string{"P3"}, Pages: []string{"P3: 実装する"}, Want: nil},
		"missing ID is reported":           {OnlyPages: []string{"P4"}, Pages: []string{"P3: 実装する"}, Want: []string{"P4"}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := pfddrawio.UnmatchedOnlyPages(tc.OnlyPages, tc.Pages)
			if diff := cmp.Diff(tc.Want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestIsotonicRegressionL2(t *testing.T) {
	testCases := map[string]struct {
		In       []float64
		Expected []float64
	}{
		"empty":                  {In: nil, Expected: nil},
		"already non-decreasing": {In: []float64{1, 2, 3}, Expected: []float64{1, 2, 3}},
		"equal values kept":      {In: []float64{5, 5, 5}, Expected: []float64{5, 5, 5}},
		"single violation pools": {In: []float64{50, -50}, Expected: []float64{0, 0}},
		"chained merges pool all three": {

			In: []float64{3, 2, 1}, Expected: []float64{2, 2, 2},
		},
		"partial pool leaves head": {

			In: []float64{1, 5, 2}, Expected: []float64{1, 3.5, 3.5},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := pfddrawio.IsotonicRegressionL2(tc.In)
			if diff := cmp.Diff(tc.Expected, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAlignYCoordinatesDiamond(t *testing.T) {
	const topCenterY, rowPitch = 50.0, 100.0
	orders := map[pfddrawio.CellID]int{"D": 0, "P1": 0, "P2": 1, "E": 0}
	ranks := map[pfddrawio.CellID]int{"D": 0, "P1": 1, "P2": 1, "E": 2}
	edges := []pfddrawio.LayoutEdge{
		fwd("a", "D", "P1"), fwd("b", "D", "P2"),
		fwd("c", "P1", "E"), fwd("d", "P2", "E"),
	}
	cy := pfddrawio.AlignYCoordinates(orders, ranks, edges, topCenterY, rowPitch)

	if cy["P2"]-cy["P1"] != rowPitch {
		t.Errorf("P1,P2 must keep rowPitch apart and in order: P1=%v P2=%v", cy["P1"], cy["P2"])
	}

	const tol = 1.0
	mid := (cy["P1"] + cy["P2"]) / 2
	if math.Abs(cy["D"]-mid) > tol {
		t.Errorf("branch source D should sit at the center of its targets: D=%v mid=%v", cy["D"], mid)
	}
	if math.Abs(cy["E"]-mid) > tol {
		t.Errorf("merge target E should sit at the center of its sources: E=%v mid=%v", cy["E"], mid)
	}
}

func TestAlignYCoordinates(t *testing.T) {

	const topCenterY, rowPitch = 50.0, 100.0

	testCases := map[string]struct {
		Orders   map[pfddrawio.CellID]int
		Ranks    map[pfddrawio.CellID]int
		Edges    []pfddrawio.LayoutEdge
		Expected map[pfddrawio.CellID]float64
	}{
		"serial chain becomes horizontal": {

			Orders:   map[pfddrawio.CellID]int{"A": 0, "B": 0, "C": 0},
			Ranks:    map[pfddrawio.CellID]int{"A": 0, "B": 1, "C": 2},
			Edges:    []pfddrawio.LayoutEdge{fwd("e1", "A", "B"), fwd("e2", "B", "C")},
			Expected: map[pfddrawio.CellID]float64{"A": 50, "B": 50, "C": 50},
		},
		"branch source is pulled to the center of its targets": {

			Orders: map[pfddrawio.CellID]int{"D": 0, "P1": 0, "P2": 1},
			Ranks:  map[pfddrawio.CellID]int{"D": 0, "P1": 1, "P2": 1},
			Edges:  []pfddrawio.LayoutEdge{fwd("e1", "D", "P1"), fwd("e2", "D", "P2")},

			Expected: map[pfddrawio.CellID]float64{"D": 100, "P1": 50, "P2": 150},
		},
		"feedback edge is ignored": {

			Orders:   map[pfddrawio.CellID]int{"A": 0, "B": 0},
			Ranks:    map[pfddrawio.CellID]int{"A": 0, "B": 1},
			Edges:    []pfddrawio.LayoutEdge{fwd("e1", "A", "B"), back("e2", "B", "A")},
			Expected: map[pfddrawio.CellID]float64{"A": 50, "B": 50},
		},
		"isolated node stays at its initial slot": {
			Orders: map[pfddrawio.CellID]int{"X": 0, "Y": 1},
			Ranks:  map[pfddrawio.CellID]int{"X": 0, "Y": 0},
			Edges:  nil,

			Expected: map[pfddrawio.CellID]float64{"X": 50, "Y": 150},
		},
		"independent components align separately without mixing": {

			Orders:   map[pfddrawio.CellID]int{"A": 0, "C": 1, "B": 0, "D": 1},
			Ranks:    map[pfddrawio.CellID]int{"A": 0, "C": 0, "B": 1, "D": 1},
			Edges:    []pfddrawio.LayoutEdge{fwd("e1", "A", "B"), fwd("e2", "C", "D")},
			Expected: map[pfddrawio.CellID]float64{"A": 50, "B": 50, "C": 150, "D": 150},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := pfddrawio.AlignYCoordinates(tc.Orders, tc.Ranks, tc.Edges, topCenterY, rowPitch)
			if diff := cmp.Diff(tc.Expected, got); diff != "" {
				t.Errorf("center Y mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSortLayoutInvariantsOnTestdata(t *testing.T) {
	const eps = 1e-9
	opts := pfddrawio.DefaultSortOptions()

	entries, err := os.ReadDir("../../../testdata")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	minHorizontalEdges := map[string]int{
		"branch": 5, "butterflyloop": 5, "callout": 5, "connector": 8, "crossloop": 6,
		"dupsevered": 2, "large": 26, "large_nofb": 31, "longloop": 4, "loop": 4, "nestedloop": 6, "nestedloopalt": 6, "simple": 2,
	}

	checked := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := "../../../testdata/" + entry.Name() + "/pfd.drawio"
		src, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			t.Fatalf("ReadFile %s: %v", path, err)
		}
		checked++
		t.Run(entry.Name(), func(t *testing.T) {

			once := assertSortedLayoutInvariants(t, src, opts, minHorizontalEdges[entry.Name()])
			assertSortedLayoutInvariants(t, once, opts, minHorizontalEdges[entry.Name()])
		})
	}
	if checked == 0 {
		t.Fatal("no fixture was checked")
	}
}

func assertSortedLayoutInvariants(t *testing.T, src []byte, opts pfddrawio.SortOptions, minHorizontal int) []byte {
	t.Helper()
	const eps = 1e-9

	originY := make(map[pfddrawio.DiagramID]float64)
	for pageID, dom := range collectDiagramLayouts(t, src) {
		minY := math.Inf(1)
		for _, v := range dom {
			minY = math.Min(minY, v.Rect.Y)
		}
		originY[pageID] = minY
	}

	nodes, err := pfddrawio.Sort(bytes.NewReader(src), opts, testLogger(t))
	if err != nil {
		t.Fatalf("Sort: %v", err)
	}

	if got := danglingRefs(t, nodes); len(got) > 0 {
		t.Errorf("dangling references remain: %v", got)
	}

	layers := pfddrawio.NewLayerMapFromNodes(nodes)
	parents := pfddrawio.NewParentMapFromNodes(nodes)
	horizontal := 0
	for pageID, dom := range pfddrawio.CollectDiagramDOMs(nodes) {
		vertices, edges := pfddrawio.CollectPageLayout(dom, layers, parents)
		if len(vertices) == 0 {
			continue
		}
		rowPitch := pfddrawio.RowPitch(vertices, opts)

		centerY := make(map[pfddrawio.CellID]float64, len(vertices))
		rects := make(map[pfddrawio.CellID]geom.Rect, len(vertices))
		base := math.Inf(1)
		for _, v := range vertices {
			centerY[v.ID] = v.Rect.Y + v.Rect.Height/2
			rects[v.ID] = v.Rect
			base = math.Min(base, centerY[v.ID])
		}
		for _, v := range vertices {
			rows := (centerY[v.ID] - base) / rowPitch
			if math.Abs(rows-math.Round(rows)) > eps {
				t.Errorf("page %s: vertex %s is off the row grid: rows from the topmost=%v (rowPitch=%v)",
					pageID, v.ID, rows, rowPitch)
			}
		}

		top := math.Inf(1)
		for _, v := range vertices {
			top = math.Min(top, v.Rect.Y)
		}
		if math.Abs(top-originY[pageID]) > eps {
			t.Errorf("page %s: the top of the page must stay at the origin: top=%v originY=%v",
				pageID, top, originY[pageID])
		}

		if severed := pfddrawio.SameColumnSeveredDuplicateGroups(vertices, edges); len(severed) > 0 {
			t.Errorf("page %s: production chain is severed for %v", pageID, severed)
		}

		centersByColumn := make(map[float64][]float64)
		for _, v := range vertices {
			centersByColumn[v.Rect.X] = append(centersByColumn[v.Rect.X], centerY[v.ID])
		}
		for x, centers := range centersByColumn {
			sort.Float64s(centers)
			for i := 1; i < len(centers); i++ {
				if gap := centers[i] - centers[i-1]; gap < rowPitch-eps {
					t.Errorf("page %s column x=%v: vertices are closer than the row pitch: gap=%v rowPitch=%v",
						pageID, x, gap, rowPitch)
				}
			}
		}

		for _, e := range edges {
			if e.IsFeedback {

				if rects[e.Source].X <= rects[e.Target].X {
					t.Errorf("page %s: feedback edge %s does not point leftward: source %s x=%v, target %s x=%v",
						pageID, e.ID, e.Source, rects[e.Source].X, e.Target, rects[e.Target].X)
				}

				if span := rects[e.Source].X - rects[e.Target].X; span >= 2*(rects[e.Target].Width+opts.HGap) {
					t.Errorf("page %s: feedback edge %s spans more than one column: dx=%v",
						pageID, e.ID, span)
				}
				continue
			}

			if rects[e.Source].X >= rects[e.Target].X {
				t.Errorf("page %s: solid edge %s points leftward: source %s x=%v, target %s x=%v",
					pageID, e.ID, e.Source, rects[e.Source].X, e.Target, rects[e.Target].X)
			}

			if span := rects[e.Target].X - rects[e.Source].X; span >= 2*(rects[e.Source].Width+opts.HGap) {
				t.Errorf("page %s: solid edge %s spans more than one column: dx=%v",
					pageID, e.ID, span)
			}
			if centerY[e.Source] == centerY[e.Target] {
				horizontal++
			}
		}
	}
	if horizontal < minHorizontal {
		t.Errorf("horizontal solid edges regressed: got %d, want >= %d", horizontal, minHorizontal)
	}
	out, err := xmldom.Marshal(nodes)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	return out
}

func collectDiagramLayouts(t *testing.T, src []byte) map[pfddrawio.DiagramID][]pfddrawio.LayoutVertex {
	t.Helper()
	nodes, err := xmldom.ParseXML(bytes.NewReader(src))
	if err != nil {
		t.Fatalf("ParseXML: %v", err)
	}
	layers := pfddrawio.NewLayerMapFromNodes(nodes)
	parents := pfddrawio.NewParentMapFromNodes(nodes)
	byPage := make(map[pfddrawio.DiagramID][]pfddrawio.LayoutVertex)
	for pageID, dom := range pfddrawio.CollectDiagramDOMs(nodes) {
		vertices, _ := pfddrawio.CollectPageLayout(dom, layers, parents)
		byPage[pageID] = vertices
	}
	return byPage
}

func TestAssignRows(t *testing.T) {

	const topCenterY, rowPitch = 50.0, 100.0

	testCases := map[string]struct {
		Orders   map[pfddrawio.CellID]int
		Ranks    map[pfddrawio.CellID]int
		CenterY  map[pfddrawio.CellID]float64
		Expected map[pfddrawio.CellID]int
	}{
		"empty": {
			Orders: nil, Ranks: nil, CenterY: nil,
			Expected: map[pfddrawio.CellID]int{},
		},
		"single vertex sits at row 0": {
			Orders:   map[pfddrawio.CellID]int{"A": 0},
			Ranks:    map[pfddrawio.CellID]int{"A": 0},
			CenterY:  map[pfddrawio.CellID]float64{"A": 50},
			Expected: map[pfddrawio.CellID]int{"A": 0},
		},
		"off-grid centers snap to the nearest row": {

			Orders:   map[pfddrawio.CellID]int{"A": 0, "B": 1},
			Ranks:    map[pfddrawio.CellID]int{"A": 0, "B": 0},
			CenterY:  map[pfddrawio.CellID]float64{"A": 70, "B": 170},
			Expected: map[pfddrawio.CellID]int{"A": 0, "B": 1},
		},
		"exactly half a row rounds down": {

			Orders:   map[pfddrawio.CellID]int{"X": 0, "A": 0},
			Ranks:    map[pfddrawio.CellID]int{"X": 0, "A": 1},
			CenterY:  map[pfddrawio.CellID]float64{"X": 50, "A": 100},
			Expected: map[pfddrawio.CellID]int{"X": 0, "A": 1},
		},
		"exactly half a row rounds down on the negative side too": {

			Orders:   map[pfddrawio.CellID]int{"X": 0, "A": 0},
			Ranks:    map[pfddrawio.CellID]int{"X": 0, "A": 1},
			CenterY:  map[pfddrawio.CellID]float64{"X": 50, "A": 0},
			Expected: map[pfddrawio.CellID]int{"X": 0, "A": 0},
		},
		"same-rank collision is pushed down": {

			Orders:   map[pfddrawio.CellID]int{"A": 0, "B": 1},
			Ranks:    map[pfddrawio.CellID]int{"A": 0, "B": 0},
			CenterY:  map[pfddrawio.CellID]float64{"A": 60, "B": 60},
			Expected: map[pfddrawio.CellID]int{"A": 0, "B": 1},
		},
		"push down cascades": {

			Orders:   map[pfddrawio.CellID]int{"A": 0, "B": 1, "C": 2},
			Ranks:    map[pfddrawio.CellID]int{"A": 0, "B": 0, "C": 0},
			CenterY:  map[pfddrawio.CellID]float64{"A": 50, "B": 50, "C": 50},
			Expected: map[pfddrawio.CellID]int{"A": 0, "B": 1, "C": 2},
		},
		"inverted centers are corrected to keep the order": {

			Orders:   map[pfddrawio.CellID]int{"A": 0, "B": 1},
			Ranks:    map[pfddrawio.CellID]int{"A": 0, "B": 0},
			CenterY:  map[pfddrawio.CellID]float64{"A": 250, "B": 50},
			Expected: map[pfddrawio.CellID]int{"A": 0, "B": 1},
		},
		"already separated rows are kept as is": {
			Orders:   map[pfddrawio.CellID]int{"A": 0, "B": 1},
			Ranks:    map[pfddrawio.CellID]int{"A": 0, "B": 0},
			CenterY:  map[pfddrawio.CellID]float64{"A": 50, "B": 350},
			Expected: map[pfddrawio.CellID]int{"A": 0, "B": 3},
		},
		"whole diagram is shifted so that the minimum row is 0": {

			Orders:   map[pfddrawio.CellID]int{"A": 0, "B": 0},
			Ranks:    map[pfddrawio.CellID]int{"A": 0, "B": 1},
			CenterY:  map[pfddrawio.CellID]float64{"A": -250, "B": -150},
			Expected: map[pfddrawio.CellID]int{"A": 0, "B": 1},
		},
		"ranks share one grid and one shift": {

			Orders:   map[pfddrawio.CellID]int{"A": 0, "B": 1, "C": 0},
			Ranks:    map[pfddrawio.CellID]int{"A": 0, "B": 0, "C": 1},
			CenterY:  map[pfddrawio.CellID]float64{"A": 150, "B": 250, "C": 150},
			Expected: map[pfddrawio.CellID]int{"A": 0, "B": 1, "C": 0},
		},
		"push down of one rank does not drag another": {

			Orders:   map[pfddrawio.CellID]int{"A": 0, "B": 1, "C": 0},
			Ranks:    map[pfddrawio.CellID]int{"A": 0, "B": 0, "C": 1},
			CenterY:  map[pfddrawio.CellID]float64{"A": 50, "B": 50, "C": 50},
			Expected: map[pfddrawio.CellID]int{"A": 0, "B": 1, "C": 0},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := pfddrawio.AssignRows(tc.Orders, tc.Ranks, tc.CenterY, topCenterY, rowPitch)
			if diff := cmp.Diff(tc.Expected, got); diff != "" {
				t.Errorf("rows mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOrderWithinRanks(t *testing.T) {
	testCases := map[string]struct {
		Vertices []pfddrawio.LayoutVertex
		Edges    []pfddrawio.LayoutEdge
		Ranks    map[pfddrawio.CellID]int
		Expected map[pfddrawio.CellID]int
	}{
		"crossings are reduced by barycenter": {

			Vertices: []pfddrawio.LayoutVertex{
				delAt("1", 0), delAt("2", 100),
				procAt("3", 0), procAt("4", 100),
			},
			Edges: []pfddrawio.LayoutEdge{
				fwd("e1", "1", "4"),
				fwd("e2", "2", "3"),
			},
			Ranks: map[pfddrawio.CellID]int{"1": 0, "2": 0, "3": 1, "4": 1},

			Expected: map[pfddrawio.CellID]int{"1": 0, "2": 1, "4": 0, "3": 1},
		},
		"a vertex with no solid neighbor follows its feedback partner": {

			Vertices: []pfddrawio.LayoutVertex{
				procAt("1", 0), procAt("2", 200),
				delAt("3", 0), delAt("4", 200), delAt("5", 400), delAt("6", 500),
			},
			Edges: []pfddrawio.LayoutEdge{
				fwd("e1", "1", "3"), fwd("e2", "2", "4"),
				back("fb1", "5", "2"), back("fb2", "6", "1"),
			},
			Ranks: map[pfddrawio.CellID]int{"1": 0, "2": 0, "3": 1, "4": 1, "5": 1, "6": 1},

			Expected: map[pfddrawio.CellID]int{"1": 0, "2": 1, "3": 0, "6": 1, "4": 2, "5": 3},
		},
		"a terminal deliverable is not pulled by its feedback edge": {

			Vertices: []pfddrawio.LayoutVertex{
				procAt("1", 0), procAt("2", 200), procAt("3", 400),
				delAt("4", 0), delAt("5", 200), delAt("6", 400),
				procAt("7", 400),
			},
			Edges: []pfddrawio.LayoutEdge{
				fwd("e1", "1", "4"), fwd("e2", "2", "5"), fwd("e3", "3", "6"),
				fwd("e4", "6", "7"),
				back("fb1", "4", "2"), back("fb2", "5", "1"),
			},
			Ranks: map[pfddrawio.CellID]int{"1": 0, "2": 0, "3": 0, "4": 1, "5": 1, "6": 1, "7": 2},

			Expected: map[pfddrawio.CellID]int{
				"1": 0, "3": 1, "2": 2, "4": 0, "6": 1, "5": 2, "7": 0,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			orders := pfddrawio.OrderWithinRanks(tc.Vertices, tc.Edges, tc.Ranks)
			if diff := cmp.Diff(tc.Expected, orders); diff != "" {
				t.Errorf("orders mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAssignRanks(t *testing.T) {
	testCases := map[string]struct {
		Vertices []pfddrawio.LayoutVertex
		Edges    []pfddrawio.LayoutEdge
		Expected map[pfddrawio.CellID]int
		WantErr  bool
	}{
		"serial D->P->D": {
			Vertices: []pfddrawio.LayoutVertex{del("2"), proc("3"), del("5")},
			Edges:    []pfddrawio.LayoutEdge{fwd("4", "2", "3"), fwd("6", "3", "5")},
			Expected: map[pfddrawio.CellID]int{"2": 0, "3": 1, "5": 2},
		},
		"branch merges at longest path": {
			Vertices: []pfddrawio.LayoutVertex{del("1"), proc("2"), proc("3"), del("4")},
			Edges: []pfddrawio.LayoutEdge{
				fwd("a", "1", "2"), fwd("b", "1", "3"),
				fwd("c", "2", "4"), fwd("d", "3", "4"),
			},
			Expected: map[pfddrawio.CellID]int{"1": 0, "2": 1, "3": 1, "4": 2},
		},
		"feedback edge does not affect ranks": {
			Vertices: []pfddrawio.LayoutVertex{del("1"), proc("2"), del("3")},
			Edges: []pfddrawio.LayoutEdge{
				fwd("a", "1", "2"), fwd("b", "2", "3"),
				back("c", "3", "1"),
			},
			Expected: map[pfddrawio.CellID]int{"1": 0, "2": 1, "3": 2},
		},
		"cycle among non-feedback edges is an error": {
			Vertices: []pfddrawio.LayoutVertex{del("1"), proc("2")},
			Edges:    []pfddrawio.LayoutEdge{fwd("a", "1", "2"), fwd("b", "2", "1")},
			WantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ranks, err := pfddrawio.AssignRanks(tc.Vertices, tc.Edges)
			if tc.WantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.Expected, ranks); diff != "" {
				t.Errorf("ranks mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func delNamedInColumn(id pfddrawio.CellID, nodeID pfd.NodeID, x float64) pfddrawio.LayoutVertex {
	return pfddrawio.LayoutVertex{
		ID: id, NodeID: nodeID, Kind: pfddrawio.VertexDeliverable,
		Rect: geom.Rect{X: x, Width: 120, Height: 80},
	}
}

func TestSameColumnSeveredDuplicateGroups(t *testing.T) {

	testCases := map[string]struct {
		Vertices []pfddrawio.LayoutVertex
		Edges    []pfddrawio.LayoutEdge
		Expected []pfd.NodeID
	}{
		"severed within one column is reported": {

			Vertices: []pfddrawio.LayoutVertex{
				proc("p1"), proc("p2"),
				delNamedInColumn("a", "D1", 200), delNamedInColumn("b", "D1", 200),
			},
			Edges:    []pfddrawio.LayoutEdge{fwd("prod", "p1", "a"), fwd("cons", "b", "p2")},
			Expected: []pfd.NodeID{"D1"},
		},
		"severed across columns is not reported": {

			Vertices: []pfddrawio.LayoutVertex{
				proc("p1"), proc("p2"),
				delNamedInColumn("a", "D1", 200), delNamedInColumn("b", "D1", 1000),
			},
			Edges:    []pfddrawio.LayoutEdge{fwd("prod", "p1", "a"), fwd("cons", "b", "p2")},
			Expected: nil,
		},
		"a duplicate in the same column that keeps the chain is not reported": {

			Vertices: []pfddrawio.LayoutVertex{
				proc("p1"), proc("p2"), proc("p3"),
				delNamedInColumn("a", "D1", 200), delNamedInColumn("b", "D1", 200),
			},
			Edges: []pfddrawio.LayoutEdge{
				fwd("prod", "p1", "a"), fwd("cons", "a", "p2"), fwd("cons2", "b", "p3"),
			},
			Expected: nil,
		},
		"a feedback-only duplicate is not a severance": {

			Vertices: []pfddrawio.LayoutVertex{
				proc("p1"), proc("p2"),
				delNamedInColumn("a", "D1", 200), delNamedInColumn("b", "D1", 200),
			},
			Edges:    []pfddrawio.LayoutEdge{fwd("prod", "p1", "a"), back("fb", "b", "p2")},
			Expected: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := pfddrawio.SameColumnSeveredDuplicateGroups(tc.Vertices, tc.Edges)
			if diff := cmp.Diff(tc.Expected, got); diff != "" {
				t.Errorf("severed groups mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestContractDuplicates(t *testing.T) {

	testCases := map[string]struct {
		Vertices []pfddrawio.LayoutVertex
		Edges    []pfddrawio.LayoutEdge
		Expected pfddrawio.Representatives
	}{
		"same NodeID deliverables share one representative": {
			Vertices: []pfddrawio.LayoutVertex{delNamed("3", "D2"), delNamed("9", "D2")},
			Expected: pfddrawio.Representatives{"3": "3", "9": "3"},
		},
		"the smallest CellID becomes the representative": {
			Vertices: []pfddrawio.LayoutVertex{delNamed("9", "D2"), delNamed("12", "D2"), delNamed("3", "D2")},
			Expected: pfddrawio.Representatives{"3": "3", "9": "3", "12": "3"},
		},
		"distinct NodeIDs are not contracted": {
			Vertices: []pfddrawio.LayoutVertex{delNamed("3", "D1"), delNamed("9", "D2")},
			Expected: pfddrawio.Representatives{"3": "3", "9": "9"},
		},
		"empty NodeID is not contracted": {
			Vertices: []pfddrawio.LayoutVertex{del("3"), del("9")},
			Expected: pfddrawio.Representatives{"3": "3", "9": "9"},
		},
		"processes with the same NodeID are not contracted": {
			Vertices: []pfddrawio.LayoutVertex{procNamed("3", "P1"), procNamed("9", "P1")},
			Expected: pfddrawio.Representatives{"3": "3", "9": "9"},
		},
		"connectors are not contracted": {
			Vertices: []pfddrawio.LayoutVertex{conn("3"), conn("9")},
			Expected: pfddrawio.Representatives{"3": "3", "9": "9"},
		},
		"the cell with a production edge becomes the representative": {
			Vertices: []pfddrawio.LayoutVertex{procNamed("p1", "P1"), delNamed("9", "D2"), delNamed("zz", "D2")},
			Edges:    []pfddrawio.LayoutEdge{fwd("e1", "p1", "zz")},
			Expected: pfddrawio.Representatives{"p1": "p1", "9": "zz", "zz": "zz"},
		},
		"a feedback in-edge does not make a representative": {
			Vertices: []pfddrawio.LayoutVertex{procNamed("p1", "P1"), delNamed("9", "D2"), delNamed("zz", "D2")},
			Edges:    []pfddrawio.LayoutEdge{back("e1", "zz", "p1")},
			Expected: pfddrawio.Representatives{"p1": "p1", "9": "9", "zz": "9"},
		},
		"among produced cells the smallest CellID wins": {
			Vertices: []pfddrawio.LayoutVertex{procNamed("p1", "P1"), delNamed("9", "D2"), delNamed("zz", "D2")},
			Edges:    []pfddrawio.LayoutEdge{fwd("e1", "p1", "zz"), fwd("e2", "p1", "9")},
			Expected: pfddrawio.Representatives{"p1": "p1", "9": "9", "zz": "9"},
		},

		"a derived duplicate ID loses to its source without production edges": {
			Vertices: []pfddrawio.LayoutVertex{delNamed("d1", "D1"), delNamed("d1-dup1", "D1")},
			Expected: pfddrawio.Representatives{"d1": "d1", "d1-dup1": "d1"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if diff := cmp.Diff(tc.Expected, pfddrawio.ContractDuplicates(tc.Vertices, tc.Edges)); diff != "" {
				t.Errorf("representatives mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExtraDuplicateCells(t *testing.T) {
	testCases := map[string]struct {
		Vertices []pfddrawio.LayoutVertex
		Edges    []pfddrawio.LayoutEdge
		Expected map[pfddrawio.CellID][]pfddrawio.CellID
	}{
		"no duplicate has no extra": {
			Vertices: []pfddrawio.LayoutVertex{delNamed("3", "D1"), delNamed("9", "D2")},
			Expected: map[pfddrawio.CellID][]pfddrawio.CellID{},
		},
		"every cell but the representative is an extra": {
			Vertices: []pfddrawio.LayoutVertex{delNamed("9", "D2"), delNamed("3", "D2"), delNamed("12", "D2")},
			Expected: map[pfddrawio.CellID][]pfddrawio.CellID{"3": {"9", "12"}},
		},

		"extras are sorted by CellID": {
			Vertices: []pfddrawio.LayoutVertex{delNamed("d", "D2"), delNamed("b", "D2"), delNamed("a", "D2"), delNamed("c", "D2")},
			Expected: map[pfddrawio.CellID][]pfddrawio.CellID{"a": {"b", "c", "d"}},
		},
		"processes with the same NodeID are not extras": {
			Vertices: []pfddrawio.LayoutVertex{procNamed("3", "P1"), procNamed("9", "P1")},
			Expected: map[pfddrawio.CellID][]pfddrawio.CellID{},
		},
		"connectors are not extras": {
			Vertices: []pfddrawio.LayoutVertex{conn("3"), conn("9")},
			Expected: map[pfddrawio.CellID][]pfddrawio.CellID{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			rep := pfddrawio.ContractDuplicates(tc.Vertices, tc.Edges)
			if diff := cmp.Diff(tc.Expected, pfddrawio.ExtraDuplicateCells(tc.Vertices, rep)); diff != "" {
				t.Errorf("extras mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPrefersAsRepresentative(t *testing.T) {

	produced := map[pfddrawio.CellID]bool{"withProducer": true}

	testCases := map[string]struct {
		Candidate pfddrawio.CellID
		Current   pfddrawio.CellID
		Expected  bool
	}{
		"a cell with a production edge wins":               {Candidate: "withProducer", Current: "a", Expected: true},
		"a cell without a production edge loses":           {Candidate: "a", Current: "withProducer", Expected: false},
		"without production edges the smaller CellID wins": {Candidate: "a", Current: "b", Expected: true},
		"without production edges the larger CellID loses": {Candidate: "b", Current: "a", Expected: false},
		"a derived duplicate ID never beats its source":    {Candidate: "d1-dup1", Current: "d1", Expected: false},
		"a source always beats its derived duplicate ID":   {Candidate: "d1", Current: "d1-dup1", Expected: true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if got := pfddrawio.PrefersAsRepresentative(tc.Candidate, tc.Current, produced); got != tc.Expected {
				t.Errorf("want %v, got %v", tc.Expected, got)
			}
		})
	}
}

func TestDuplicateCellID(t *testing.T) {

	testCases := map[string]struct {
		Source   pfddrawio.CellID
		N        int
		Expected pfddrawio.CellID
	}{
		"the first duplicate":           {Source: "d1", N: 0, Expected: "d1-dup1"},
		"the second duplicate":          {Source: "d1", N: 1, Expected: "d1-dup2"},
		"a numeric source cell ID":      {Source: "42", N: 0, Expected: "42-dup1"},
		"a drawio style source cell ID": {Source: "xY3k-12", N: 0, Expected: "xY3k-12-dup1"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := pfddrawio.DuplicateCellID(tc.Source, tc.N)
			if got != tc.Expected {
				t.Errorf("want %q, got %q", tc.Expected, got)
			}
			if got.Compare(tc.Source) <= 0 {
				t.Errorf("a duplicate cell ID must sort after its source: %q vs %q", got, tc.Source)
			}
		})
	}
}

func TestNextDuplicateCellID(t *testing.T) {

	taken := func(ids ...pfddrawio.CellID) map[pfddrawio.CellID]struct{} {
		m := make(map[pfddrawio.CellID]struct{}, len(ids))
		for _, id := range ids {
			m[id] = struct{}{}
		}
		return m
	}

	testCases := map[string]struct {
		Source   pfddrawio.CellID
		Taken    map[pfddrawio.CellID]struct{}
		Expected pfddrawio.CellID
	}{
		"the first duplicate of a source": {
			Source: "d1", Taken: taken("d1"), Expected: "d1-dup1",
		},
		"after an existing duplicate": {
			Source: "d1", Taken: taken("d1", "d1-dup1"), Expected: "d1-dup2",
		},

		"a gap is not filled": {
			Source: "d1", Taken: taken("d1", "d1-dup2"), Expected: "d1-dup3",
		},
		"another source's duplicates do not matter": {
			Source: "d1", Taken: taken("d2-dup1", "d2-dup2"), Expected: "d1-dup1",
		},

		"an id taken by a wrapper element is skipped": {
			Source: "d1", Taken: taken("d1", "d1-dup1", "d1-dup2"), Expected: "d1-dup3",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := pfddrawio.NextDuplicateCellID(tc.Source, tc.Taken)
			if got != tc.Expected {
				t.Errorf("want %q, got %q", tc.Expected, got)
			}
			if _, used := tc.Taken[got]; used {
				t.Errorf("the returned id must be unused: %q", got)
			}
		})
	}
}

func TestCanonicalizeRewires(t *testing.T) {

	rep := pfddrawio.Representatives{"d1": "d1", "d1b": "d1", "p1": "p1", "p2": "p2"}

	testCases := map[string]struct {
		Edges    []pfddrawio.LayoutEdge
		Expected []pfddrawio.RewirePlan
	}{
		"an edge on the representative is left alone": {
			Edges:    []pfddrawio.LayoutEdge{fwd("e1", "d1", "p1")},
			Expected: nil,
		},
		"a source on an extra cell is moved to the representative": {
			Edges: []pfddrawio.LayoutEdge{fwd("e1", "d1b", "p1")},
			Expected: []pfddrawio.RewirePlan{
				{EdgeID: "e1", Endpoint: pfddrawio.EndpointSource, NewCell: "d1"},
			},
		},
		"a target on an extra cell is moved to the representative": {
			Edges: []pfddrawio.LayoutEdge{fwd("e1", "p1", "d1b")},
			Expected: []pfddrawio.RewirePlan{
				{EdgeID: "e1", Endpoint: pfddrawio.EndpointTarget, NewCell: "d1"},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := pfddrawio.CanonicalizeRewires(tc.Edges, rep)
			if diff := cmp.Diff(tc.Expected, got); diff != "" {
				t.Errorf("rewires mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestApplyRewires(t *testing.T) {

	src := func(edgeID, cell pfddrawio.CellID) pfddrawio.RewirePlan {
		return pfddrawio.RewirePlan{EdgeID: edgeID, Endpoint: pfddrawio.EndpointSource, NewCell: cell}
	}
	tgt := func(edgeID, cell pfddrawio.CellID) pfddrawio.RewirePlan {
		return pfddrawio.RewirePlan{EdgeID: edgeID, Endpoint: pfddrawio.EndpointTarget, NewCell: cell}
	}

	testCases := map[string]struct {
		Edges    []pfddrawio.LayoutEdge
		Rewires  []pfddrawio.RewirePlan
		Expected []pfddrawio.LayoutEdge
	}{
		"no rewire leaves the edges as they are": {
			Edges:    []pfddrawio.LayoutEdge{fwd("e1", "a", "b")},
			Rewires:  nil,
			Expected: []pfddrawio.LayoutEdge{fwd("e1", "a", "b")},
		},
		"both endpoints of one edge are rewired": {
			Edges:    []pfddrawio.LayoutEdge{fwd("e1", "a", "b")},
			Rewires:  []pfddrawio.RewirePlan{src("e1", "a2"), tgt("e1", "b2")},
			Expected: []pfddrawio.LayoutEdge{fwd("e1", "a2", "b2")},
		},

		"the later rewire of the same endpoint wins": {
			Edges:    []pfddrawio.LayoutEdge{fwd("e1", "a", "b")},
			Rewires:  []pfddrawio.RewirePlan{src("e1", "rep"), src("e1", "dup")},
			Expected: []pfddrawio.LayoutEdge{fwd("e1", "dup", "b")},
		},
		"a rewire for an unknown edge is ignored": {
			Edges:    []pfddrawio.LayoutEdge{fwd("e1", "a", "b")},
			Rewires:  []pfddrawio.RewirePlan{src("nope", "x")},
			Expected: []pfddrawio.LayoutEdge{fwd("e1", "a", "b")},
		},
		"feedback flags are preserved": {
			Edges:    []pfddrawio.LayoutEdge{back("e1", "a", "b")},
			Rewires:  []pfddrawio.RewirePlan{src("e1", "a2")},
			Expected: []pfddrawio.LayoutEdge{back("e1", "a2", "b")},
		},

		"the cut flag is cleared once the edge is rewired": {
			Edges:    []pfddrawio.LayoutEdge{cut("e1", "a", "b")},
			Rewires:  []pfddrawio.RewirePlan{src("e1", "dup")},
			Expected: []pfddrawio.LayoutEdge{fwd("e1", "dup", "b")},
		},
		"an unrewired cut edge keeps its flag": {
			Edges:    []pfddrawio.LayoutEdge{cut("e1", "a", "b")},
			Rewires:  nil,
			Expected: []pfddrawio.LayoutEdge{cut("e1", "a", "b")},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			original := make([]pfddrawio.LayoutEdge, len(tc.Edges))
			copy(original, tc.Edges)

			got := pfddrawio.ApplyRewires(tc.Edges, tc.Rewires)
			if diff := cmp.Diff(tc.Expected, got); diff != "" {
				t.Errorf("edges mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(original, tc.Edges); diff != "" {
				t.Errorf("the input edges must not be modified (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAssignDuplicateCells(t *testing.T) {

	dup := func(tempID, sourceID pfddrawio.CellID) pfddrawio.DuplicatePlan {
		return pfddrawio.DuplicatePlan{TempID: tempID, SourceID: sourceID}
	}
	reused := func(tempID, sourceID, reuseID pfddrawio.CellID) pfddrawio.DuplicatePlan {
		return pfddrawio.DuplicatePlan{TempID: tempID, SourceID: sourceID, ReuseID: reuseID}
	}

	testCases := map[string]struct {
		Dups             []pfddrawio.DuplicatePlan
		Extras           map[pfddrawio.CellID][]pfddrawio.CellID
		ExpectedDups     []pfddrawio.DuplicatePlan
		ExpectedRemovals map[pfddrawio.CellID]pfddrawio.CellID
	}{
		"without stock every duplicate is a new cell": {
			Dups:             []pfddrawio.DuplicatePlan{dup("dup#0", "d1")},
			Extras:           nil,
			ExpectedDups:     []pfddrawio.DuplicatePlan{dup("dup#0", "d1")},
			ExpectedRemovals: map[pfddrawio.CellID]pfddrawio.CellID{},
		},
		"stock is consumed in order": {
			Dups:   []pfddrawio.DuplicatePlan{dup("dup#0", "d1"), dup("dup#1", "d1")},
			Extras: map[pfddrawio.CellID][]pfddrawio.CellID{"d1": {"a", "b"}},
			ExpectedDups: []pfddrawio.DuplicatePlan{
				reused("dup#0", "d1", "a"), reused("dup#1", "d1", "b"),
			},
			ExpectedRemovals: map[pfddrawio.CellID]pfddrawio.CellID{},
		},
		"a shortage falls back to new cells": {
			Dups:   []pfddrawio.DuplicatePlan{dup("dup#0", "d1"), dup("dup#1", "d1")},
			Extras: map[pfddrawio.CellID][]pfddrawio.CellID{"d1": {"a"}},
			ExpectedDups: []pfddrawio.DuplicatePlan{
				reused("dup#0", "d1", "a"), dup("dup#1", "d1"),
			},
			ExpectedRemovals: map[pfddrawio.CellID]pfddrawio.CellID{},
		},
		"leftover stock is removed": {
			Dups:             []pfddrawio.DuplicatePlan{dup("dup#0", "d1")},
			Extras:           map[pfddrawio.CellID][]pfddrawio.CellID{"d1": {"a", "b", "c"}},
			ExpectedDups:     []pfddrawio.DuplicatePlan{reused("dup#0", "d1", "a")},
			ExpectedRemovals: map[pfddrawio.CellID]pfddrawio.CellID{"b": "d1", "c": "d1"},
		},
		"stock of a deliverable that needs no duplicate is removed": {
			Dups:             nil,
			Extras:           map[pfddrawio.CellID][]pfddrawio.CellID{"d1": {"a"}},
			ExpectedDups:     []pfddrawio.DuplicatePlan{},
			ExpectedRemovals: map[pfddrawio.CellID]pfddrawio.CellID{"a": "d1"},
		},
		"stock is not shared across deliverables": {
			Dups:   []pfddrawio.DuplicatePlan{dup("dup#0", "d1"), dup("dup#1", "d2")},
			Extras: map[pfddrawio.CellID][]pfddrawio.CellID{"d1": {"a", "b"}},
			ExpectedDups: []pfddrawio.DuplicatePlan{
				reused("dup#0", "d1", "a"), dup("dup#1", "d2"),
			},
			ExpectedRemovals: map[pfddrawio.CellID]pfddrawio.CellID{"b": "d1"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			dups, removals := pfddrawio.AssignDuplicateCells(tc.Dups, tc.Extras)
			if diff := cmp.Diff(tc.ExpectedDups, dups); diff != "" {
				t.Errorf("duplicates mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.ExpectedRemovals, removals); diff != "" {
				t.Errorf("removals mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestContractLayout(t *testing.T) {
	testCases := map[string]struct {
		Vertices         []pfddrawio.LayoutVertex
		Edges            []pfddrawio.LayoutEdge
		ExpectedVertices []pfddrawio.LayoutVertex
		ExpectedEdges    []pfddrawio.LayoutEdge
	}{
		"edges are remapped to representatives": {
			Vertices: []pfddrawio.LayoutVertex{delNamed("3", "D2"), delNamed("9", "D2"), procNamed("10", "P2")},
			Edges:    []pfddrawio.LayoutEdge{fwd("a", "9", "10")},

			ExpectedVertices: []pfddrawio.LayoutVertex{delNamed("3", "D2"), procNamed("10", "P2")},
			ExpectedEdges:    []pfddrawio.LayoutEdge{fwd("a", "3", "10")},
		},
		"an edge inside one group is dropped": {
			Vertices: []pfddrawio.LayoutVertex{delNamed("3", "D2"), delNamed("9", "D2")},

			Edges:            []pfddrawio.LayoutEdge{fwd("a", "3", "9")},
			ExpectedVertices: []pfddrawio.LayoutVertex{delNamed("9", "D2")},
			ExpectedEdges:    nil,
		},
		"feedback edges are kept and remapped": {
			Vertices:         []pfddrawio.LayoutVertex{delNamed("3", "D2"), delNamed("9", "D2"), procNamed("10", "P2")},
			Edges:            []pfddrawio.LayoutEdge{back("a", "9", "10")},
			ExpectedVertices: []pfddrawio.LayoutVertex{delNamed("3", "D2"), procNamed("10", "P2")},
			ExpectedEdges:    []pfddrawio.LayoutEdge{back("a", "3", "10")},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			rep := pfddrawio.ContractDuplicates(tc.Vertices, tc.Edges)
			vertices, edges := pfddrawio.ContractLayout(tc.Vertices, tc.Edges, rep)
			if diff := cmp.Diff(tc.ExpectedVertices, vertices); diff != "" {
				t.Errorf("vertices mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.ExpectedEdges, edges); diff != "" {
				t.Errorf("edges mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAssignDisplayRanks(t *testing.T) {
	testCases := map[string]struct {
		Vertices []pfddrawio.LayoutVertex
		Edges    []pfddrawio.LayoutEdge
		Expected map[pfddrawio.CellID]int
		WantErr  bool
	}{
		"cells without duplicates keep the per-cell ranks": {
			Vertices: []pfddrawio.LayoutVertex{delNamed("2", "D1"), procNamed("3", "P1"), delNamed("5", "D2")},
			Edges:    []pfddrawio.LayoutEdge{fwd("4", "2", "3"), fwd("6", "3", "5")},
			Expected: map[pfddrawio.CellID]int{"2": 0, "3": 1, "5": 2},
		},
		"a duplicate takes the rank of its group instead of rank 0": {

			Vertices: []pfddrawio.LayoutVertex{
				delNamed("1", "D1"), procNamed("2", "P1"), delNamed("3", "D2"),
				delNamed("9", "D2"), procNamed("10", "P2"), delNamed("11", "D3"),
			},
			Edges: []pfddrawio.LayoutEdge{
				fwd("a", "1", "2"), fwd("b", "2", "3"),
				fwd("c", "9", "10"), fwd("d", "10", "11"),
			},
			Expected: map[pfddrawio.CellID]int{"1": 0, "2": 1, "3": 2, "9": 2, "10": 3, "11": 4},
		},
		"a cycle created only by contraction is an error": {

			Vertices: []pfddrawio.LayoutVertex{
				delNamed("1", "D1"), procNamed("2", "P1"), delNamed("3", "D2"),
				delNamed("4", "D2"), procNamed("5", "P2"), delNamed("6", "D1"),
			},
			Edges: []pfddrawio.LayoutEdge{
				fwd("a", "1", "2"), fwd("b", "2", "3"),
				fwd("c", "4", "5"), fwd("d", "5", "6"),
			},
			WantErr: true,
		},
		"a cut edge does not constrain the ranks": {

			Vertices: []pfddrawio.LayoutVertex{
				delNamed("1", "D1"), procNamed("2", "P10"), delNamed("3", "D2"),
				procNamed("4", "P20"), delNamed("5", "D3"),
			},
			Edges: []pfddrawio.LayoutEdge{
				fwd("a", "1", "2"), fwd("b", "2", "3"), fwd("c", "3", "4"),
				fwd("d", "4", "5"), cut("e", "5", "2"),
			},
			Expected: map[pfddrawio.CellID]int{"1": 0, "2": 1, "3": 2, "4": 3, "5": 4},
		},
		"a cycle among non-feedback edges is an error": {
			Vertices: []pfddrawio.LayoutVertex{delNamed("1", "D1"), procNamed("2", "P1")},
			Edges:    []pfddrawio.LayoutEdge{fwd("a", "1", "2"), fwd("b", "2", "1")},
			WantErr:  true,
		},
		"a solid edge between two cells of one group is an error": {

			Vertices: []pfddrawio.LayoutVertex{
				delNamed("1", "D1"), delNamed("2", "D1"), procNamed("6", "P2"),
			},
			Edges:   []pfddrawio.LayoutEdge{fwd("x", "1", "2"), fwd("y", "1", "6")},
			WantErr: true,
		},
		"a feedback edge from a deliverable to a process is fine": {

			Vertices: []pfddrawio.LayoutVertex{
				delNamed("1", "D1"), procNamed("2", "P1"), delNamed("3", "D2"),
			},
			Edges:    []pfddrawio.LayoutEdge{fwd("a", "1", "2"), fwd("b", "2", "3"), back("c", "3", "2")},
			Expected: map[pfddrawio.CellID]int{"1": 0, "2": 1, "3": 2},
		},
		"a feedback edge starting at a process is an error": {

			Vertices: []pfddrawio.LayoutVertex{
				delNamed("1", "D1"), procNamed("2", "P1"), delNamed("3", "D2"),
			},
			Edges:   []pfddrawio.LayoutEdge{fwd("a", "1", "2"), fwd("b", "2", "3"), back("c", "2", "1")},
			WantErr: true,
		},
		"a feedback edge between two deliverables is an error": {

			Vertices: []pfddrawio.LayoutVertex{
				delNamed("1", "D1"), procNamed("2", "P1"), delNamed("3", "D1"),
			},
			Edges:   []pfddrawio.LayoutEdge{fwd("a", "1", "2"), back("b", "1", "3")},
			WantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ranks, err := pfddrawio.AssignDisplayRanks(tc.Vertices, tc.Edges)
			if tc.WantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.Expected, ranks); diff != "" {
				t.Errorf("ranks mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSpreadSourceRanks(t *testing.T) {

	chain := []pfddrawio.LayoutVertex{
		delNamed("1", "D1"), procNamed("2", "P1"), delNamed("3", "D2"), procNamed("4", "P2"),
		delNamed("5", "D3"), procNamed("6", "P3"), delNamed("7", "D4"),
	}
	chainEdges := []pfddrawio.LayoutEdge{
		fwd("a", "1", "2"), fwd("b", "2", "3"), fwd("c", "3", "4"),
		fwd("d", "4", "5"), fwd("e", "5", "6"), fwd("f", "6", "7"),
	}
	withChain := func(extra ...pfddrawio.LayoutVertex) []pfddrawio.LayoutVertex {
		return append(append([]pfddrawio.LayoutVertex{}, chain...), extra...)
	}
	withChainEdges := func(extra ...pfddrawio.LayoutEdge) []pfddrawio.LayoutEdge {
		return append(append([]pfddrawio.LayoutEdge{}, chainEdges...), extra...)
	}

	withChainRanks := func(extra map[pfddrawio.CellID]int) map[pfddrawio.CellID]int {
		merged := map[pfddrawio.CellID]int{"1": 0, "2": 1, "3": 2, "4": 3, "5": 4, "6": 5, "7": 6}
		for id, r := range extra {
			merged[id] = r
		}
		return merged
	}

	testCases := map[string]struct {
		Vertices []pfddrawio.LayoutVertex
		Edges    []pfddrawio.LayoutEdge
		Logical  map[pfddrawio.CellID]int
		Expected map[pfddrawio.CellID]int
	}{
		"a duplicate without a producer moves next to its consumer": {

			Vertices: withChain(delNamed("9", "D2")),
			Edges:    withChainEdges(fwd("g", "9", "6")),
			Logical:  withChainRanks(map[pfddrawio.CellID]int{"9": 2}),
			Expected: withChainRanks(map[pfddrawio.CellID]int{"9": 4}),
		},
		"a cell with a producer stays at the group rank": {

			Vertices: withChain(),
			Edges:    withChainEdges(fwd("g", "3", "6")),
			Logical:  withChainRanks(nil),
			Expected: withChainRanks(nil),
		},
		"a cell with no solid edge stays at the group rank": {
			Vertices: withChain(delNamed("9", "D2")),
			Edges:    withChainEdges(),
			Logical:  withChainRanks(map[pfddrawio.CellID]int{"9": 2}),
			Expected: withChainRanks(map[pfddrawio.CellID]int{"9": 2}),
		},
		"a feedback-only out-edge moves the cell right of its process": {

			Vertices: withChain(delNamed("9", "D2")),
			Edges:    withChainEdges(back("g", "9", "6")),
			Logical:  withChainRanks(map[pfddrawio.CellID]int{"9": 2}),
			Expected: withChainRanks(map[pfddrawio.CellID]int{"9": 6}),
		},
		"the farthest feedback target decides the rank": {

			Vertices: withChain(delNamed("9", "D2")),
			Edges:    withChainEdges(back("g", "9", "4"), back("h", "9", "6")),
			Logical:  withChainRanks(map[pfddrawio.CellID]int{"9": 2}),
			Expected: withChainRanks(map[pfddrawio.CellID]int{"9": 6}),
		},
		"a solid out-edge wins over a feedback out-edge": {

			Vertices: withChain(delNamed("9", "D9")),
			Edges:    withChainEdges(fwd("g", "9", "4"), back("h", "9", "6")),
			Logical:  withChainRanks(map[pfddrawio.CellID]int{"9": 0}),
			Expected: withChainRanks(map[pfddrawio.CellID]int{"9": 2}),
		},
		"a feedback target left of the group rank pulls the cell left": {

			Vertices: withChain(delNamed("9", "D3")),
			Edges:    withChainEdges(back("g", "9", "2")),
			Logical:  withChainRanks(map[pfddrawio.CellID]int{"9": 4}),
			Expected: withChainRanks(map[pfddrawio.CellID]int{"9": 2}),
		},
		"a feedback out-edge inside one group does not move the cell": {
			Vertices: []pfddrawio.LayoutVertex{delNamed("1", "D1"), delNamed("2", "D1")},
			Edges:    []pfddrawio.LayoutEdge{back("a", "1", "2")},
			Logical:  map[pfddrawio.CellID]int{"1": 0, "2": 0},
			Expected: map[pfddrawio.CellID]int{"1": 0, "2": 0},
		},
		"a cell with a producer stays even with a feedback out-edge": {

			Vertices: withChain(),
			Edges:    withChainEdges(back("g", "3", "6")),
			Logical:  withChainRanks(nil),
			Expected: withChainRanks(nil),
		},
		"the nearest consumer decides the rank": {

			Vertices: withChain(delNamed("9", "D9")),
			Edges:    withChainEdges(fwd("g", "9", "4"), fwd("h", "9", "6")),
			Logical:  withChainRanks(map[pfddrawio.CellID]int{"9": 0}),
			Expected: withChainRanks(map[pfddrawio.CellID]int{"9": 2}),
		},
		"a lone source moves next to its far consumer": {

			Vertices: withChain(delNamed("9", "D9")),
			Edges:    withChainEdges(fwd("g", "9", "6")),
			Logical:  withChainRanks(map[pfddrawio.CellID]int{"9": 0}),
			Expected: withChainRanks(map[pfddrawio.CellID]int{"9": 4}),
		},
		"a consumer through a connector decides the rank": {

			Vertices: withChain(conn("8"), delNamed("9", "D9")),
			Edges:    withChainEdges(fwd("g", "5", "8"), fwd("h", "8", "6"), fwd("i", "9", "8")),
			Logical:  withChainRanks(map[pfddrawio.CellID]int{"6": 6, "7": 7, "8": 5, "9": 0}),
			Expected: withChainRanks(map[pfddrawio.CellID]int{"6": 6, "7": 7, "8": 5, "9": 4}),
		},
		"an out-edge inside one group does not move the cell": {

			Vertices: []pfddrawio.LayoutVertex{delNamed("1", "D1"), delNamed("2", "D1")},
			Edges:    []pfddrawio.LayoutEdge{fwd("a", "1", "2")},
			Logical:  map[pfddrawio.CellID]int{"1": 0, "2": 0},
			Expected: map[pfddrawio.CellID]int{"1": 0, "2": 0},
		},
		"the rank never goes below the given rank": {

			Vertices: []pfddrawio.LayoutVertex{delNamed("9", "D9"), procNamed("10", "P1")},
			Edges:    []pfddrawio.LayoutEdge{fwd("a", "9", "10")},
			Logical:  map[pfddrawio.CellID]int{"9": 3, "10": 0},
			Expected: map[pfddrawio.CellID]int{"9": 3, "10": 0},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			rep := pfddrawio.ContractDuplicates(tc.Vertices, tc.Edges)
			got := pfddrawio.SpreadSourceRanks(tc.Logical, tc.Vertices, tc.Edges, rep)
			if diff := cmp.Diff(tc.Expected, got); diff != "" {
				t.Errorf("ranks mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSortDuplicateDeliverableColumn(t *testing.T) {

	chain := rectCell("2", "D1: 種", 0, 0) +
		ellipseCell("3", "P1: 加工", 0, 200) +
		rectCell("4", "D2: 中間", 0, 400) +
		ellipseCell("5", "P2: 検査", 0, 600) +
		rectCell("6", "D3: 完成", 0, 800) +
		ellipseCell("7", "P3: 報告", 0, 1000) +
		rectCell("8", "D4: 報告書", 0, 1200) +
		edgeCell("20", "2", "3", false) +
		edgeCell("21", "3", "4", false) +
		edgeCell("22", "4", "5", false) +
		edgeCell("23", "5", "6", false) +
		edgeCell("24", "6", "7", false) +
		edgeCell("25", "7", "8", false)

	testCases := map[string]struct {
		Doc      string
		Cell     pfddrawio.CellID
		Expected float64
	}{
		"a duplicated deliverable sits next to its consumer": {

			Doc:      drawioPage(chain + rectCell("9", "D2: 中間", 0, 1400) + edgeCell("26", "9", "7", false)),
			Cell:     "9",
			Expected: 800,
		},
		"the original deliverable stays at its own rank": {
			Doc:      drawioPage(chain + rectCell("9", "D2: 中間", 0, 1400) + edgeCell("26", "9", "7", false)),
			Cell:     "4",
			Expected: 400,
		},
		"a lone source sits next to its far consumer": {

			Doc:      drawioPage(chain + rectCell("9", "D9: 外部", 0, 1400) + edgeCell("26", "9", "7", false)),
			Cell:     "9",
			Expected: 800,
		},
		"a deliverable used only by a feedback edge sits right of its process": {

			Doc:      drawioPage(chain + rectCell("9", "D9: 差戻し", 0, 1400) + edgeCell("26", "9", "7", true)),
			Cell:     "9",
			Expected: 1200,
		},
		"a source consumed nearby stays at the leftmost column": {
			Doc:      drawioPage(chain + rectCell("9", "D9: 外部", 0, 1400) + edgeCell("26", "9", "3", false)),
			Cell:     "9",
			Expected: 0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := sortAndReadRects(t, tc.Doc, pfddrawio.DefaultSortOptions())
			if got[tc.Cell].X != tc.Expected {
				t.Errorf("cell %s is in the wrong column: x=%v, want %v", tc.Cell, got[tc.Cell].X, tc.Expected)
			}
		})
	}
}

func cellIDsByPage(t *testing.T, doc []byte) map[pfddrawio.DiagramID][]pfddrawio.CellID {
	t.Helper()
	byPage := make(map[pfddrawio.DiagramID][]pfddrawio.CellID)
	for pageID, rects := range rectsByPage(t, doc) {
		ids := make([]pfddrawio.CellID, 0, len(rects))
		for id := range rects {
			ids = append(ids, id)
		}
		slices.SortFunc(ids, pfddrawio.CellID.Compare)
		byPage[pageID] = ids
	}
	return byPage
}

func TestSortIsIdempotentOnTestdata(t *testing.T) {
	opts := pfddrawio.DefaultSortOptions()

	entries, err := os.ReadDir("../../../testdata")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	checked := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := "../../../testdata/" + entry.Name() + "/pfd.drawio"
		src, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			t.Fatalf("ReadFile %s: %v", path, err)
		}
		checked++
		t.Run(entry.Name(), func(t *testing.T) {
			once := sortToBytes(t, src, opts)
			twice := sortToBytes(t, once, opts)
			thrice := sortToBytes(t, twice, opts)
			if diff := cmp.Diff(cellIDsByPage(t, once), cellIDsByPage(t, twice)); diff != "" {
				t.Errorf("sorting twice changed the set of cells (-once +twice):\n%s", diff)
			}
			if diff := cmp.Diff(rectsByPage(t, twice), rectsByPage(t, thrice)); diff != "" {
				t.Errorf("sorting a settled figure moved vertices (-twice +thrice):\n%s", diff)
			}
			if !bytes.Equal(twice, thrice) {
				t.Errorf("sorting a settled figure changed the document")
			}
		})
	}
	if checked == 0 {
		t.Fatal("no fixture was checked")
	}
}

func TestSortIsIdempotentWithMixedHeights(t *testing.T) {

	opts := pfddrawio.DefaultSortOptions()

	testCases := map[string]struct{ Doc string }{
		"boxes of one height": {
			Doc: drawioPage(
				rectCell("2", "D1: 甲", 0, 0) +
					ellipseCell("3", "P1: 一", 200, 0) +
					rectCell("4", "D2: 乙", 0, 200) +
					ellipseCell("5", "P2: 二", 200, 200) +
					edgeCell("6", "2", "3", false) +
					edgeCell("7", "4", "5", false),
			),
		},

		"the tallest process sits on a lower row": {
			Doc: drawioPage(
				rectCellSized("2", "D1: 甲", 0, 0, 40) +
					ellipseCellSized("3", "P1: 一", 200, 0, 40) +
					rectCellSized("4", "D2: 乙", 0, 200, 40) +
					ellipseCellSized("5", "P2: 二", 200, 200, 80) +
					edgeCell("6", "2", "3", false) +
					edgeCell("7", "4", "5", false),
			),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			once := sortToBytes(t, []byte(tc.Doc), opts)
			twice := sortToBytes(t, once, opts)
			if diff := cmp.Diff(rectsByPage(t, once), rectsByPage(t, twice)); diff != "" {
				t.Errorf("sorting twice moved vertices (-once +twice):\n%s", diff)
			}
		})
	}
}

func sortAndCollect(t *testing.T, doc string, opts pfddrawio.SortOptions) ([]pfddrawio.LayoutVertex, []pfddrawio.LayoutEdge) {
	t.Helper()
	nodes, err := pfddrawio.Sort(strings.NewReader(doc), opts, testLogger(t))
	if err != nil {
		t.Fatalf("Sort: %v", err)
	}
	layers := pfddrawio.NewLayerMapFromNodes(nodes)
	parents := pfddrawio.NewParentMapFromNodes(nodes)
	doms := pfddrawio.CollectDiagramDOMs(nodes)
	dom, ok := doms["d0"]
	if !ok {
		t.Fatalf("diagram d0 not found after Sort")
	}
	return pfddrawio.CollectPageLayout(dom, layers, parents)
}

func severedDuplicateDoc() string {
	return drawioPage(
		rectCell("d1", "D1: 入力", 0, 0) +
			ellipseCell("p1", "P1: 前工程", 200, 0) +

			rectCell("d2a", "D2: 中間", 400, 0) +
			rectCell("d2b", "D2: 中間", 400, 600) +
			ellipseCell("p2", "P2: 後工程", 600, 600) +
			rectCell("d3", "D3: 出力", 800, 600) +
			edgeCell("e1", "d1", "p1", false) +
			edgeCell("e2", "p1", "d2a", false) +
			edgeCell("e3", "d2b", "p2", false) +
			edgeCell("e4", "p2", "d3", false),
	)
}

func danglingRefs(t *testing.T, nodes []*xmldom.Node) []string {
	t.Helper()
	var dangling []string
	for pageID, dom := range pfddrawio.CollectDiagramDOMs(nodes) {
		for _, ref := range pfddrawio.DanglingRefs(dom.Root) {
			dangling = append(dangling, string(pageID)+":"+ref)
		}
	}
	sort.Strings(dangling)
	return dangling
}

func TestSortLeavesNoDanglingReference(t *testing.T) {

	testCases := map[string]struct{ Doc string }{

		"a removed duplicate has a child cell": {
			Doc: drawioPage(
				rectCell("d1", "D1: 入力", 0, 0) +
					ellipseCell("p1", "P1: 前工程", 200, 0) +
					rectCell("d2a", "D2: 中間", 400, 0) +
					rectCell("d2b", "D2: 中間", 400, 600) +
					`<mxCell id="note" value="注記" style="text;html=1;" parent="d2b" vertex="1">` + mxGeom(410, 610, 100, 20) + `</mxCell>` +
					ellipseCell("p2", "P2: 後工程", 600, 600) +
					rectCell("d3", "D3: 出力", 800, 600) +
					edgeCell("e1", "d1", "p1", false) +
					edgeCell("e2", "p1", "d2a", false) +
					edgeCell("e3", "d2b", "p2", false) +
					edgeCell("e4", "p2", "d3", false),
			),
		},

		"a removed duplicate has a child cell wrapped in an object": {
			Doc: drawioPage(
				rectCell("d1", "D1: 入力", 0, 0) +
					ellipseCell("p1", "P1: 前工程", 200, 0) +
					rectCell("d2a", "D2: 中間", 400, 0) +
					rectCell("d2b", "D2: 中間", 400, 600) +
					`<object id="note" label="注記" link="https://example.com/"><mxCell style="text;html=1;" parent="d2b" vertex="1">` + mxGeom(410, 610, 100, 20) + `</mxCell></object>` +
					ellipseCell("p2", "P2: 後工程", 600, 600) +
					rectCell("d3", "D3: 出力", 800, 600) +
					edgeCell("e1", "d1", "p1", false) +
					edgeCell("e2", "p1", "d2a", false) +
					edgeCell("e3", "d2b", "p2", false) +
					edgeCell("e4", "p2", "d3", false),
			),
		},

		"a removed duplicate is pointed at by a non-layout edge": {
			Doc: drawioPage(
				rectCell("d1", "D1: 入力", 0, 0) +
					ellipseCell("p1", "P1: 前工程", 200, 0) +
					rectCell("d2a", "D2: 中間", 400, 0) +
					rectCell("d2b", "D2: 中間", 400, 600) +
					`<mxCell id="memo" value="補足" style="text;html=1;" parent="1" vertex="1">` + mxGeom(400, 900, 100, 20) + `</mxCell>` +
					`<mxCell id="ememo" value="" style="edgeStyle=none;html=1;" parent="1" source="d2b" target="memo" edge="1"><mxGeometry relative="1" as="geometry"/></mxCell>` +
					ellipseCell("p2", "P2: 後工程", 600, 600) +
					rectCell("d3", "D3: 出力", 800, 600) +
					edgeCell("e1", "d1", "p1", false) +
					edgeCell("e2", "p1", "d2a", false) +
					edgeCell("e3", "d2b", "p2", false) +
					edgeCell("e4", "p2", "d3", false),
			),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			nodes, err := pfddrawio.Sort(strings.NewReader(tc.Doc), pfddrawio.DefaultSortOptions(), testLogger(t))
			if err != nil {
				t.Fatalf("Sort: %v", err)
			}
			if got := danglingRefs(t, nodes); len(got) > 0 {
				t.Errorf("dangling references remain: %v", got)
			}
		})
	}
}

func TestSortKeepsPageTopWithDuplicateCells(t *testing.T) {

	opts := pfddrawio.DefaultSortOptions()

	parts := rectCell("d1", "D1: 素材", 0, 0) +
		rectCell("d1x1", "D1: 素材", 0, -1000) +
		ellipseCell("t1", "T1: 上部", 200, 100) +
		edgeCell("et1", "d1", "t1", false)
	for i := 1; i <= 20; i++ {
		b := "b" + strconv.Itoa(i)
		parts += rectCell(b, "B"+strconv.Itoa(i)+": 部品", 0, float64(1000+i*100)) +
			edgeCell("eb"+strconv.Itoa(i), b, "pb", false)
	}
	doc := drawioPage(parts +
		ellipseCell("pb", "PB: 集約", 200, 3000) +
		edgeCell("vfar", "d1x1", "pb", false))

	testCases := map[string]struct {
		Doc            string
		ExpectedTopY   float64
		ExpectedRounds int
	}{
		"a duplicate cell above every other cell": {Doc: doc, ExpectedTopY: -1000, ExpectedRounds: 3},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			cur := []byte(tc.Doc)
			for run := 1; run <= tc.ExpectedRounds; run++ {
				cur = sortToBytes(t, cur, opts)
				top := math.Inf(1)
				for _, vertices := range collectDiagramLayouts(t, cur) {
					for _, v := range vertices {
						top = math.Min(top, v.Rect.Y)
					}
				}
				if top != tc.ExpectedTopY {
					t.Fatalf("run %d: page top must stay at %v, got %v", run, tc.ExpectedTopY, top)
				}
			}
		})
	}
}

func TestSortIsIdempotentAfterCreatingADuplicate(t *testing.T) {

	opts := pfddrawio.DefaultSortOptions()

	testCases := map[string]struct{ Doc string }{

		"a duplicate of a deliverable without a producer": {Doc: verticallyFarDoc()},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			once := sortToBytes(t, []byte(tc.Doc), opts)
			twice := sortToBytes(t, once, opts)
			if diff := cmp.Diff(rectsByPage(t, once), rectsByPage(t, twice)); diff != "" {
				t.Errorf("sorting twice moved vertices (-once +twice):\n%s", diff)
			}
			if !bytes.Equal(once, twice) {
				t.Errorf("sorting twice changed the document")
			}
		})
	}
}

func TestSortSyncsReusedDuplicateAppearance(t *testing.T) {

	testCases := map[string]struct {
		Doc string
	}{
		"a stale label and a different fill color are synced": {
			Doc: verticallyFarDupDocWith(1, "D1: 素材（古い説明）", "rounded=0;whiteSpace=wrap;html=1;fillColor=#f8cecc;"),
		},
		"an already synced duplicate stays as it is": {
			Doc: verticallyFarDupDoc(1),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			nodes, err := pfddrawio.Sort(strings.NewReader(tc.Doc), pfddrawio.DefaultSortOptions(), testLogger(t))
			if err != nil {
				t.Fatalf("Sort: %v", err)
			}
			dom, ok := pfddrawio.CollectDiagramDOMs(nodes)["d0"]
			if !ok {
				t.Fatalf("diagram d0 not found after Sort")
			}

			cell, ok := dom.CellsByID["d1x1"]
			if !ok {
				t.Fatalf("premise broken: the existing duplicate cell must be reused")
			}

			wantByAttr := map[string]string{
				"value": "D1＊: 素材",
				"style": "rounded=0;whiteSpace=wrap;html=1;",
			}
			for attr, want := range wantByAttr {
				got, _ := cell.GetAttr(attr, "")
				if got != want {
					t.Errorf("reused duplicate %s mismatch: want %q, got %q", attr, want, got)
				}
			}

			src := dom.CellsByID["d1"]
			if got, _ := src.GetAttr("value", ""); got != "D1: 素材" {
				t.Errorf("複製元は無印のはずだが value = %q", got)
			}
		})
	}
}

func TestSortReusesExistingDuplicateCell(t *testing.T) {

	opts := pfddrawio.DefaultSortOptions()
	cellsOf := func(t *testing.T, src []byte, nodeID pfd.NodeID) []pfddrawio.CellID {
		t.Helper()
		var ids []pfddrawio.CellID
		for _, vertices := range collectDiagramLayouts(t, src) {
			for _, v := range vertices {
				if v.NodeID == nodeID {
					ids = append(ids, v.ID)
				}
			}
		}
		slices.SortFunc(ids, pfddrawio.CellID.Compare)
		return ids
	}

	testCases := map[string]struct {
		Doc string

		WantCells int
	}{
		"a duplicate created by the first sort is reused by the second": {
			Doc: verticallyFarDoc(), WantCells: 2,
		},
		"a duplicate that already exists in the document is reused": {
			Doc: verticallyFarDupDoc(1), WantCells: 2,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			once := sortToBytes(t, []byte(tc.Doc), opts)
			twice := sortToBytes(t, once, opts)

			first := cellsOf(t, once, "D1")
			if len(first) != tc.WantCells {
				t.Fatalf("premise broken: want %d cells labeled D1 after the first sort, got %v", tc.WantCells, first)
			}
			if diff := cmp.Diff(first, cellsOf(t, twice, "D1")); diff != "" {
				t.Errorf("the duplicate cell must be reused, not recreated (-once +twice):\n%s", diff)
			}
		})
	}
}

func verticallyFarDupDoc(extras int) string {
	return verticallyFarDupDocWith(extras, "D1: 素材", "rounded=0;whiteSpace=wrap;html=1;")
}

func verticallyFarDupDocWith(extras int, copyValue, copyStyle string) string {
	cells := rectCell("d1", "D1: 素材", 0, 0)
	for i := 1; i <= 3; i++ {
		p := "t" + strconv.Itoa(i)
		cells += ellipseCell(p, "T"+strconv.Itoa(i)+": 上部", 200, float64(i*100)) +
			edgeCell("et"+strconv.Itoa(i), "d1", p, false)
	}
	for i := 1; i <= 20; i++ {
		b := "b" + strconv.Itoa(i)
		cells += rectCell(b, "B"+strconv.Itoa(i)+": 部品", 0, float64(1000+i*100)) +
			edgeCell("eb"+strconv.Itoa(i), b, "pb", false)
	}
	for i := 1; i <= extras; i++ {
		cells += `<mxCell id="d1x` + strconv.Itoa(i) + `" value="` + copyValue + `" style="` + copyStyle + `" parent="1" vertex="1">` +
			mxGeom(0, float64(2900+i*100), 120, 80) + `</mxCell>`
	}
	return drawioPage(cells +
		ellipseCell("pb", "PB: 集約", 200, 3000) +
		edgeCell("vfar", "d1x1", "pb", false))
}

func TestPlanLayoutReplansExistingDuplicates(t *testing.T) {

	testCases := map[string]struct {
		Doc              string
		ExpectedRemovals map[pfddrawio.CellID]pfddrawio.CellID
		ExpectedRewires  []pfddrawio.RewirePlan
		ExpectedReuseIDs []pfddrawio.CellID
	}{

		"a duplicate that is no longer needed is removed": {
			Doc:              severedDuplicateDoc(),
			ExpectedRemovals: map[pfddrawio.CellID]pfddrawio.CellID{"d2b": "d2a"},
			ExpectedRewires: []pfddrawio.RewirePlan{
				{EdgeID: "e3", Endpoint: pfddrawio.EndpointSource, NewCell: "d2a"},
			},
			ExpectedReuseIDs: nil,
		},

		"a duplicate that is still needed reuses the existing cell": {
			Doc:              verticallyFarDupDoc(1),
			ExpectedRemovals: map[pfddrawio.CellID]pfddrawio.CellID{},
			ExpectedRewires: []pfddrawio.RewirePlan{
				{EdgeID: "vfar", Endpoint: pfddrawio.EndpointSource, NewCell: "d1"},
				{EdgeID: "vfar", Endpoint: pfddrawio.EndpointSource, NewCell: "dup#0"},
			},
			ExpectedReuseIDs: []pfddrawio.CellID{"d1x1"},
		},

		"leftover stock is removed while one cell is reused": {
			Doc:              verticallyFarDupDoc(2),
			ExpectedRemovals: map[pfddrawio.CellID]pfddrawio.CellID{"d1x2": "d1"},
			ExpectedRewires: []pfddrawio.RewirePlan{
				{EdgeID: "vfar", Endpoint: pfddrawio.EndpointSource, NewCell: "d1"},
				{EdgeID: "vfar", Endpoint: pfddrawio.EndpointSource, NewCell: "dup#0"},
			},
			ExpectedReuseIDs: []pfddrawio.CellID{"d1x1"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			vertices, edges := collectFirstPage(t, tc.Doc)
			plan, err := pfddrawio.PlanLayout(vertices, edges, pfddrawio.DefaultSortOptions())
			if err != nil {
				t.Fatalf("PlanLayout: %v", err)
			}

			if diff := cmp.Diff(tc.ExpectedRemovals, plan.Removals); diff != "" {
				t.Errorf("removals mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.ExpectedRewires, plan.Rewires); diff != "" {
				t.Errorf("rewires mismatch (-want +got):\n%s", diff)
			}

			var reuseIDs []pfddrawio.CellID
			for _, d := range plan.Duplicates {
				reuseIDs = append(reuseIDs, d.ReuseID)
			}
			if diff := cmp.Diff(tc.ExpectedReuseIDs, reuseIDs); diff != "" {
				t.Errorf("reuse ids mismatch (-want +got):\n%s", diff)
			}

			for id := range tc.ExpectedRemovals {
				if _, ok := plan.Positions[id]; ok {
					t.Errorf("a removed cell must not get a position: %s", id)
				}
			}
		})
	}
}

func TestSortRemergesSeveredDuplicates(t *testing.T) {

	opts := pfddrawio.DefaultSortOptions()

	testCases := map[string]struct {
		Doc string

		SeveredBefore bool

		WantCells     []pfddrawio.CellID
		WantGoneCells []pfddrawio.CellID
	}{
		"a severed duplicate pair is merged back into one box": {
			Doc:           severedDuplicateDoc(),
			SeveredBefore: true,
			WantCells:     []pfddrawio.CellID{"d2a"},
			WantGoneCells: []pfddrawio.CellID{"d2b"},
		},

		"leftover duplicates are removed while the needed one stays": {
			Doc:           verticallyFarDupDoc(2),
			SeveredBefore: false,
			WantCells:     []pfddrawio.CellID{"d1", "d1x1"},
			WantGoneCells: []pfddrawio.CellID{"d1x2"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			beforeVertices, beforeEdges := collectFirstPage(t, tc.Doc)
			if severed := pfddrawio.SeveredDuplicateGroups(beforeVertices, beforeEdges); (len(severed) > 0) != tc.SeveredBefore {
				t.Fatalf("premise broken: severed before sorting = %v, want %v", severed, tc.SeveredBefore)
			}

			vertices, edges := sortAndCollect(t, tc.Doc, opts)
			if got := pfddrawio.SeveredDuplicateGroups(vertices, edges); len(got) > 0 {
				t.Errorf("production chain is still severed for %v", got)
			}

			alive := make(map[pfddrawio.CellID]bool, len(vertices))
			for _, v := range vertices {
				alive[v.ID] = true
			}
			for _, id := range tc.WantCells {
				if !alive[id] {
					t.Errorf("cell %s must be kept", id)
				}
			}
			for _, id := range tc.WantGoneCells {
				if alive[id] {
					t.Errorf("cell %s must be removed", id)
				}
			}
		})
	}
}

func sortToBytes(t *testing.T, src []byte, opts pfddrawio.SortOptions) []byte {
	t.Helper()
	nodes, err := pfddrawio.Sort(bytes.NewReader(src), opts, testLogger(t))
	if err != nil {
		t.Fatalf("Sort: %v", err)
	}
	out, err := xmldom.Marshal(nodes)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	return out
}

func rectsByPage(t *testing.T, src []byte) map[pfddrawio.DiagramID]map[pfddrawio.CellID]geom.Rect {
	t.Helper()
	byPage := make(map[pfddrawio.DiagramID]map[pfddrawio.CellID]geom.Rect)
	for pageID, vertices := range collectDiagramLayouts(t, src) {
		rects := make(map[pfddrawio.CellID]geom.Rect, len(vertices))
		for _, v := range vertices {
			rects[v.ID] = v.Rect
		}
		byPage[pageID] = rects
	}
	return byPage
}

func TestRemoveEdgeWaypoints(t *testing.T) {
	testCases := map[string]struct {
		cell string
		want string
	}{
		"waypoints を除く": {
			cell: `<mxCell id="e" edge="1">` + "\n" + `  <mxGeometry relative="1" as="geometry">` + "\n" +
				`    <Array as="points">` + "\n" + `      <mxPoint x="10" y="20"></mxPoint>` + "\n" + `    </Array>` + "\n" +
				`  </mxGeometry>` + "\n" + `</mxCell>`,
			want: `<mxCell id="e" edge="1">` + "\n" + `  <mxGeometry relative="1" as="geometry"></mxGeometry>` + "\n" + `</mxCell>`,
		},
		"sourcePoint/targetPoint は残す": {
			cell: `<mxCell id="e" edge="1">` + "\n" + `  <mxGeometry relative="1" as="geometry">` + "\n" +
				`    <mxPoint x="1" y="2" as="sourcePoint"></mxPoint>` + "\n" +
				`    <Array as="points">` + "\n" + `      <mxPoint x="10" y="20"></mxPoint>` + "\n" + `    </Array>` + "\n" +
				`  </mxGeometry>` + "\n" + `</mxCell>`,
			want: `<mxCell id="e" edge="1">` + "\n" + `  <mxGeometry relative="1" as="geometry">` + "\n" +
				`    <mxPoint x="1" y="2" as="sourcePoint"></mxPoint>` + "\n" + `  </mxGeometry>` + "\n" + `</mxCell>`,
		},
		"waypoints が無ければ変えない": {
			cell: `<mxCell id="e" edge="1">` + "\n" + `  <mxGeometry relative="1" as="geometry"></mxGeometry>` + "\n" + `</mxCell>`,
			want: `<mxCell id="e" edge="1">` + "\n" + `  <mxGeometry relative="1" as="geometry"></mxGeometry>` + "\n" + `</mxCell>`,
		},
		"mxGeometry が無ければ変えない": {
			cell: `<mxCell id="e" edge="1"></mxCell>`,
			want: `<mxCell id="e" edge="1"></mxCell>`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			roots, err := xmldom.ParseXML(strings.NewReader(tc.cell))
			if err != nil {
				t.Fatal(err)
			}

			pfddrawio.RemoveEdgeWaypoints(roots[0])

			buf := &strings.Builder{}
			if err := roots[0].Write(buf); err != nil {
				t.Fatal(err)
			}
			if buf.String() != tc.want {
				t.Errorf("after RemoveEdgeWaypoints =\n%s\nwant\n%s", buf.String(), tc.want)
			}
		})
	}
}

func styledEdgeCell(id, source, target, style string, withWaypoints bool) string {
	geo := `<mxGeometry relative="1" as="geometry"/>`
	if withWaypoints {
		geo = `<mxGeometry relative="1" as="geometry"><Array as="points"><mxPoint x="10" y="20"/></Array></mxGeometry>`
	}
	return `<mxCell id="` + id + `" value="" style="` + style + `" parent="1" source="` + source +
		`" target="` + target + `" edge="1">` + geo + `</mxCell>`
}

func edgeCellOf(t *testing.T, nodes []*xmldom.Node, diagramID pfddrawio.DiagramID, cellID pfddrawio.CellID) *xmldom.Node {
	t.Helper()
	dom, ok := pfddrawio.CollectDiagramDOMs(nodes)[diagramID]
	if !ok {
		t.Fatalf("diagram %q not found", diagramID)
	}
	cell, ok := dom.CellsByID[cellID]
	if !ok {
		t.Fatalf("cell %q not found", cellID)
	}
	return cell
}

func TestSortEdgeStyle(t *testing.T) {
	const messyStyle = "edgeStyle=orthogonalEdgeStyle;shape=connector;rounded=1;jumpStyle=gap;html=1;entryX=0;entryY=0.5;entryDx=0;entryDy=0;exitX=1;exitY=0.5;exitDx=0;exitDy=0;strokeColor=default;endArrow=classic;"
	const straightStyle = "edgeStyle=none;shape=connector;rounded=1;jumpStyle=gap;html=1;strokeColor=default;endArrow=classic;"

	type want struct {
		style     string
		waypoints bool
	}

	testCases := map[string]struct {
		doc string

		onlyPages []string
		want      map[pfddrawio.DiagramID]map[pfddrawio.CellID]want
	}{
		"整列対象のエッジは直線化して waypoints を消す": {
			doc: drawioPage(rectCell("2", "D1: 入力", 500, 90) + ellipseCell("3", "P1: 加工", 100, 400) +
				styledEdgeCell("4", "2", "3", messyStyle, true)),
			want: map[pfddrawio.DiagramID]map[pfddrawio.CellID]want{
				"d0": {"4": {style: straightStyle}},
			},
		},
		"破線（手戻り）も直線化する": {
			doc: drawioPage(rectCell("2", "D1: 入力", 500, 90) + ellipseCell("3", "P1: 加工", 100, 400) +
				rectCell("5", "D2: 出力", 900, 90) + edgeCell("4", "2", "3", false) +
				edgeCell("6", "3", "5", false) +
				styledEdgeCell("7", "5", "3", "edgeStyle=orthogonalEdgeStyle;html=1;dashed=1;", true)),
			want: map[pfddrawio.DiagramID]map[pfddrawio.CellID]want{
				"d0": {"7": {style: "edgeStyle=none;html=1;dashed=1;jumpStyle=gap;"}},
			},
		},
		"端点がテキストボックスのエッジは触らない": {
			doc: drawioPage(rectCell("2", "D1: 入力", 500, 90) + ellipseCell("3", "P1: 加工", 100, 400) +
				edgeCell("4", "2", "3", false) +
				`<mxCell id="8" value="注記" style="text;html=1;" parent="1" vertex="1">` + mxGeom(0, 700, 100, 30) + `</mxCell>` +
				styledEdgeCell("9", "8", "3", messyStyle, true)),
			want: map[pfddrawio.DiagramID]map[pfddrawio.CellID]want{
				"d0": {"9": {style: messyStyle, waypoints: true}},
			},
		},
		"コメントレイヤー上のエッジは触らない": {
			doc: drawioPage(rectCell("2", "D1: 入力", 500, 90) + ellipseCell("3", "P1: 加工", 100, 400) +
				edgeCell("4", "2", "3", false) +
				`<mxCell id="cl" value="コメント" parent="0"/>` +
				`<mxCell id="c1" value="D9: 注" style="rounded=0;whiteSpace=wrap;html=1;" parent="cl" vertex="1">` + mxGeom(0, 900, 120, 80) + `</mxCell>` +
				`<mxCell id="c2" value="P9: 注" style="ellipse;whiteSpace=wrap;html=1;" parent="cl" vertex="1">` + mxGeom(200, 900, 120, 80) + `</mxCell>` +
				styledEdgeCell("c3", "c1", "c2", messyStyle, true)),
			want: map[pfddrawio.DiagramID]map[pfddrawio.CellID]want{
				"d0": {"c3": {style: messyStyle, waypoints: true}},
			},
		},
		"-only-page で整列しなかったページのエッジは触らない": {
			doc: `<mxfile>` +
				diagramPage("dA", "P0", rectCell("2", "D1: 入力", 500, 90)+ellipseCell("3", "P1: 加工", 100, 400)+
					styledEdgeCell("4", "2", "3", "edgeStyle=orthogonalEdgeStyle;html=1;entryX=0;entryY=0.5;", true)) +
				diagramPage("dB", "P3", rectCell("12", "D1: 入力", 500, 90)+ellipseCell("13", "P1: 加工", 100, 400)+
					styledEdgeCell("14", "12", "13", "edgeStyle=orthogonalEdgeStyle;html=1;entryX=0;entryY=0.5;", true)) +
				`</mxfile>`,
			onlyPages: []string{"P0"},
			want: map[pfddrawio.DiagramID]map[pfddrawio.CellID]want{
				"dA": {"4": {style: "edgeStyle=none;html=1;jumpStyle=gap;"}},
				"dB": {"14": {style: "edgeStyle=orthogonalEdgeStyle;html=1;entryX=0;entryY=0.5;", waypoints: true}},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			opts := pfddrawio.DefaultSortOptions()
			opts.OnlyPages = tc.onlyPages
			nodes, err := pfddrawio.Sort(strings.NewReader(tc.doc), opts, testLogger(t))
			if err != nil {
				t.Fatalf("Sort: %v", err)
			}

			for diagramID, wantByCell := range tc.want {
				for cellID, w := range wantByCell {
					cell := edgeCellOf(t, nodes, diagramID, cellID)
					gotStyle, _ := cell.GetAttr("style", "")
					if gotStyle != w.style {
						t.Errorf("page %s: style of edge %q =\n %q\nwant\n %q", diagramID, cellID, gotStyle, w.style)
					}

					geo := cell.FirstChildElement("mxGeometry")
					gotWaypoints := geo != nil && geo.FirstChildElement("Array") != nil
					if gotWaypoints != w.waypoints {
						t.Errorf("page %s: edge %q has waypoints = %v, want %v", diagramID, cellID, gotWaypoints, w.waypoints)
					}
				}
			}
		})
	}
}

func assertStraightEdges(t *testing.T, nodes []*xmldom.Node) {
	t.Helper()
	layers := pfddrawio.NewLayerMapFromNodes(nodes)
	parents := pfddrawio.NewParentMapFromNodes(nodes)
	fixed := []string{"entryX", "entryY", "entryDx", "entryDy", "exitX", "exitY", "exitDx", "exitDy"}

	for pageID, dom := range pfddrawio.CollectDiagramDOMs(nodes) {
		_, edges := pfddrawio.CollectPageLayout(dom, layers, parents)
		for _, e := range edges {
			cell, ok := dom.CellsByID[e.ID]
			if !ok {
				t.Fatalf("page %s: edge %s not found", pageID, e.ID)
			}
			styleStr, _ := cell.GetAttr("style", "")
			style, err := pfddrawio.ParseStyle(styleStr)
			if err != nil {
				t.Fatalf("page %s: edge %s: ParseStyle: %v", pageID, e.ID, err)
			}
			if got := style.Get("edgeStyle"); got != "none" {
				t.Errorf("page %s: edge %s: edgeStyle = %q, want none (style=%q)", pageID, e.ID, got, styleStr)
			}
			if got := style.Get("jumpStyle"); got != "gap" {
				t.Errorf("page %s: edge %s: jumpStyle = %q, want gap (style=%q)", pageID, e.ID, got, styleStr)
			}
			for _, key := range fixed {
				if _, ok := style[key]; ok {
					t.Errorf("page %s: edge %s: %s must be removed (style=%q)", pageID, e.ID, key, styleStr)
				}
			}
			if geo := cell.FirstChildElement("mxGeometry"); geo != nil && geo.FirstChildElement("Array") != nil {
				t.Errorf("page %s: edge %s: waypoints must be removed", pageID, e.ID)
			}
		}
	}
}

func TestSortStraightensEdgesOnTestdata(t *testing.T) {
	opts := pfddrawio.DefaultSortOptions()

	entries, err := os.ReadDir("../../../testdata")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	checked := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := "../../../testdata/" + entry.Name() + "/pfd.drawio"
		src, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			t.Fatalf("ReadFile %s: %v", path, err)
		}
		checked++
		t.Run(entry.Name(), func(t *testing.T) {
			nodes, err := pfddrawio.Sort(bytes.NewReader(src), opts, testLogger(t))
			if err != nil {
				t.Fatalf("Sort: %v", err)
			}
			assertStraightEdges(t, nodes)

			out, err := xmldom.Marshal(nodes)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			renodes, err := pfddrawio.Sort(bytes.NewReader(out), opts, testLogger(t))
			if err != nil {
				t.Fatalf("re-Sort: %v", err)
			}
			assertStraightEdges(t, renodes)
		})
	}
	if checked == 0 {
		t.Fatal("no testdata figure was checked")
	}
}

func TestSortMarksDuplicates(t *testing.T) {

	xmlText := `<mxfile>` + diagramPage("g0", "P0",
		ellipseCell("p0", "P0x: 作る", 0, 0)+
			rectCell("d1", "D1: 素材", 200, 0)+
			ellipseCell("p1", "P1: 使う1", 400, 0)+
			rectCell("d2", "D2: 中間1", 600, 0)+
			ellipseCell("p2", "P2: 使う2", 800, 0)+
			rectCell("d3", "D3: 中間2", 1000, 0)+
			ellipseCell("p3", "P3: 使う3", 1200, 0)+
			edgeCell("e0", "p0", "d1", false)+
			edgeCell("e1", "d1", "p1", false)+
			edgeCell("e2", "p1", "d2", false)+
			edgeCell("e3", "d2", "p2", false)+
			edgeCell("e4", "p2", "d3", false)+
			edgeCell("e5", "d3", "p3", false)+
			edgeCell("e6", "d1", "p3", false)) + `</mxfile>`

	logger := slog.New(slogtest.NewTestHandler(t))
	nodes, err := pfddrawio.Sort(strings.NewReader(xmlText), pfddrawio.DefaultSortOptions(), logger)
	if err != nil {
		t.Fatalf("Sort: %v", err)
	}

	dom, ok := pfddrawio.CollectDiagramDOMs(nodes)["g0"]
	if !ok {
		t.Fatal("diagram g0 not found")
	}

	unmarked, marked := 0, 0
	for id, cell := range dom.CellsByID {
		if nodeID, ok := pfddrawio.VertexNodeID(cell); !ok || nodeID != "D1" {
			continue
		}
		value, _ := cell.GetAttr("value", "")
		if strings.Contains(value, pfddrawio.DuplicateMark) {
			marked++
			continue
		}
		unmarked++
		if id != "d1" {
			t.Errorf("正本は d1 のはずだが %q が無印だった", id)
		}
	}
	if unmarked != 1 {
		t.Errorf("無印のセルは 1 つのはずだが %d 個あった", unmarked)
	}
	if marked == 0 {
		t.Error("複製セルにマークが付いていない")
	}
}

func TestParsePosRestriction(t *testing.T) {
	testCases := map[string]struct {
		Input   string
		Want    pfddrawio.PosRestriction
		WantErr bool
	}{
		"lock":         {"lock", pfddrawio.PosLock, false},
		"free-v":       {"free-v", pfddrawio.PosFreeV, false},
		"free-h":       {"free-h", pfddrawio.PosFreeH, false},
		"free":         {"free", pfddrawio.PosFree, false},
		"empty":        {"", 0, true},
		"unknown":      {"diagonal", 0, true},
		"case matters": {"Lock", 0, true},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := pfddrawio.ParsePosRestriction(tc.Input)
			if tc.WantErr {
				if err == nil {
					t.Errorf("ParsePosRestriction(%q) = %v, nil; want error", tc.Input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("ParsePosRestriction(%q): %v", tc.Input, err)
			}
			if got != tc.Want {
				t.Errorf("ParsePosRestriction(%q) = %v, want %v", tc.Input, got, tc.Want)
			}
		})
	}
}

func TestSelectCells(t *testing.T) {

	doc := drawioPage(
		rectCell("d1a", "D1", 0, 0) +
			rectCell("d1b", "D1", 0, 200) +
			ellipseCell("p1", "P1", 300, 0) +
			rectCell("d2", "D2", 600, 0),
	)
	vertices, _ := collectFirstPage(t, doc)

	testCases := map[string]struct {
		OnlyNodes []pfd.NodeID
		Want      []pfddrawio.CellID
	}{
		"duplicate id selects all its cells": {[]pfd.NodeID{"D1"}, []pfddrawio.CellID{"d1a", "d1b"}},
		"process id":                         {[]pfd.NodeID{"P1"}, []pfddrawio.CellID{"p1"}},
		"multiple ids":                       {[]pfd.NodeID{"P1", "D2"}, []pfddrawio.CellID{"d2", "p1"}},
		"unknown id selects nothing":         {[]pfd.NodeID{"D9"}, nil},
		"empty selects nothing":              {nil, nil},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := pfddrawio.SelectCells(vertices, tc.OnlyNodes)
			var ids []pfddrawio.CellID
			for id := range got {
				ids = append(ids, id)
			}
			slices.SortFunc(ids, pfddrawio.CellID.Compare)
			if !slices.Equal(tc.Want, ids) {
				t.Errorf("SelectCells = %v, want %v", ids, tc.Want)
			}
		})
	}
}

func TestSelectedTopologicalOrder(t *testing.T) {
	testCases := map[string]struct {
		Doc      string
		Selected []pfddrawio.CellID
		Want     []pfddrawio.CellID
	}{
		"respects chain order over alphabetical CellID order": {
			Doc: drawioPage(
				rectCell("z1", "D1", 0, 0) +
					ellipseCell("a1", "P1", 300, 0) +
					rectCell("m1", "D2", 600, 0) +
					rectCell("b1", "D3", 0, 800) +
					rectCell("c1", "D4", 300, 800) +
					edgeCell("e1", "z1", "a1", false) +
					edgeCell("e2", "a1", "m1", false),
			),
			Selected: []pfddrawio.CellID{"z1", "a1", "m1", "b1", "c1"},

			Want: []pfddrawio.CellID{"b1", "c1", "z1", "a1", "m1"},
		},
		"respects dependency through an unselected relay": {
			Doc: drawioPage(
				rectCell("d1", "D1", 0, 0) +
					ellipseCell("p1", "P1", 300, 0) +
					rectCell("d2", "D2", 600, 0) +
					edgeCell("e1", "d1", "p1", false) +
					edgeCell("e2", "p1", "d2", false),
			),
			Selected: []pfddrawio.CellID{"d2", "d1"},
			Want:     []pfddrawio.CellID{"d1", "d2"},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			vertices, edges := collectFirstPage(t, tc.Doc)
			selected := make(map[pfddrawio.CellID]struct{}, len(tc.Selected))
			for _, id := range tc.Selected {
				selected[id] = struct{}{}
			}
			got, err := pfddrawio.SelectedTopologicalOrder(vertices, edges, selected)
			if err != nil {
				t.Fatalf("SelectedTopologicalOrder: %v", err)
			}
			if !slices.Equal(got, tc.Want) {
				t.Errorf("SelectedTopologicalOrder = %v, want %v", got, tc.Want)
			}
		})
	}
}

func TestSelectedTopologicalOrderCycleError(t *testing.T) {

	vertices := []pfddrawio.LayoutVertex{
		{ID: "a", Kind: pfddrawio.VertexDeliverable},
		{ID: "b", Kind: pfddrawio.VertexProcess},
	}
	edges := []pfddrawio.LayoutEdge{
		{ID: "e1", Source: "a", Target: "b"},
		{ID: "e2", Source: "b", Target: "a"},
	}
	selected := map[pfddrawio.CellID]struct{}{"a": {}, "b": {}}

	if _, err := pfddrawio.SelectedTopologicalOrder(vertices, edges, selected); err == nil {
		t.Error("循環がある入力はエラーになるはず")
	}
}

func TestRelaxOnlyNodeTargetsDiamond(t *testing.T) {

	edges := []pfddrawio.LayoutEdge{
		{ID: "eAB", Source: "A", Target: "B"},
		{ID: "eAC", Source: "A", Target: "C"},
		{ID: "eBD", Source: "B", Target: "D"},
		{ID: "eCD", Source: "C", Target: "D"},
	}
	frozen := map[pfddrawio.CellID]geom.Rect{
		"A": {X: 1000, Y: 0, Width: 120, Height: 80},
		"B": {X: 0, Y: 0, Width: 120, Height: 80},
		"C": {X: 500, Y: 0, Width: 120, Height: 80},
		"D": {X: 200, Y: 0, Width: 120, Height: 80},
	}
	ids := []pfddrawio.CellID{"A", "B", "C", "D"}

	got := pfddrawio.RelaxOnlyNodeTargets(ids, frozen, edges, 80)

	want := map[pfddrawio.CellID]geom.Rect{
		"A": {X: -200, Y: 0, Width: 120, Height: 80},
		"B": {X: 0, Y: 0, Width: 120, Height: 80},
		"C": {X: 0, Y: 0, Width: 120, Height: 80},
		"D": {X: 200, Y: 0, Width: 120, Height: 80},
	}
	for _, id := range ids {
		if got[id] != want[id] {
			t.Errorf("RelaxOnlyNodeTargets[%s] = %v, want %v", id, got[id], want[id])
		}
	}
	for _, e := range edges {
		pred, succ := got[e.Source], got[e.Target]
		if pred.X+pred.Width+80 > succ.X {
			t.Errorf("辺 %s が逆流している: %s.X=%v (右端=%v) > %s.X=%v", e.ID, e.Source, pred.X, pred.X+pred.Width+80, e.Target, succ.X)
		}
	}
}

func TestRelaxOnlyNodeTargetsIndependentChains(t *testing.T) {

	edges := []pfddrawio.LayoutEdge{
		{ID: "ex", Source: "x1", Target: "x2"},
		{ID: "ey", Source: "y1", Target: "y2"},
	}
	frozen := map[pfddrawio.CellID]geom.Rect{
		"x1": {X: 300, Y: 0, Width: 120, Height: 80},
		"x2": {X: 0, Y: 0, Width: 120, Height: 80},
		"y1": {X: 300, Y: 0, Width: 120, Height: 80},
		"y2": {X: 0, Y: 0, Width: 120, Height: 80},
	}
	ids := []pfddrawio.CellID{"x1", "y1", "x2", "y2"}

	got := pfddrawio.RelaxOnlyNodeTargets(ids, frozen, edges, 80)

	want := map[pfddrawio.CellID]geom.Rect{
		"x1": {X: -200, Y: 0, Width: 120, Height: 80},
		"x2": {X: 0, Y: 0, Width: 120, Height: 80},
		"y1": {X: -200, Y: 0, Width: 120, Height: 80},
		"y2": {X: 0, Y: 0, Width: 120, Height: 80},
	}
	for _, id := range ids {
		if got[id] != want[id] {
			t.Errorf("RelaxOnlyNodeTargets[%s] = %v, want %v", id, got[id], want[id])
		}
	}

	if got["x1"] != got["y1"] || got["x2"] != got["y2"] {
		t.Errorf("無関係なチェーン同士が干渉している: x=%v/%v, y=%v/%v", got["x1"], got["x2"], got["y1"], got["y2"])
	}
}

func TestRelaxOnlyNodeTargets(t *testing.T) {

	edges := []pfddrawio.LayoutEdge{
		{ID: "e1", Source: "d1", Target: "p1"},
		{ID: "e2", Source: "p1", Target: "d2"},
	}
	frozen := map[pfddrawio.CellID]geom.Rect{
		"d1": {X: 600, Y: 0, Width: 120, Height: 80},
		"p1": {X: 300, Y: 0, Width: 120, Height: 80},
		"d2": {X: 0, Y: 0, Width: 120, Height: 80},
	}
	ids := []pfddrawio.CellID{"d1", "p1", "d2"}

	got := pfddrawio.RelaxOnlyNodeTargets(ids, frozen, edges, 80)

	want := map[pfddrawio.CellID]geom.Rect{
		"d1": {X: 100, Y: 0, Width: 120, Height: 80},
		"p1": {X: 300, Y: 0, Width: 120, Height: 80},
		"d2": {X: 500, Y: 0, Width: 120, Height: 80},
	}
	for _, id := range ids {
		if got[id] != want[id] {
			t.Errorf("RelaxOnlyNodeTargets[%s] = %v, want %v", id, got[id], want[id])
		}
	}

	for _, e := range edges {
		pred, succ := got[e.Source], got[e.Target]
		if pred.X+pred.Width+80 > succ.X {
			t.Errorf("辺 %s が逆流している: %s.X=%v (右端=%v) > %s.X=%v", e.ID, e.Source, pred.X, pred.X+pred.Width+80, e.Target, succ.X)
		}
	}
}

func TestSortOnlyNodeLockPlacesSelectedRelativeToNeighbors(t *testing.T) {

	doc := drawioPage(
		rectCell("d1", "D1", 0, 0) +
			ellipseCell("p1", "P1", 50, 400) +
			rectCell("d2", "D2", 600, 200) +
			edgeCell("e1", "d1", "p1", false) +
			edgeCell("e2", "p1", "d2", false),
	)
	opts := pfddrawio.SortOptions{HGap: 80, VGap: 40, OnlyNodes: []pfd.NodeID{"P1"}, PosRestriction: pfddrawio.PosLock}
	got := sortAndReadRects(t, doc, opts)

	if got["d1"] != rect(0, 0) {
		t.Errorf("d1 は不動のはず: got %v", got["d1"])
	}
	if got["d2"] != rect(600, 200) {
		t.Errorf("d2 は不動のはず: got %v", got["d2"])
	}

	if got["p1"] != rect(200, 100) {
		t.Errorf("p1 want (200,100): got %v", got["p1"])
	}
}

func TestSortOnlyNodeLockShiftsSelectedOffImmovable(t *testing.T) {

	doc := drawioPage(
		rectCell("d1", "D1", 0, 0) +
			ellipseCell("p1", "P1", 50, 400) +
			rectCell("d2", "D2", 600, 200) +
			rectCell("d3", "D3", 200, 100) +
			edgeCell("e1", "d1", "p1", false) +
			edgeCell("e2", "p1", "d2", false),
	)
	opts := pfddrawio.SortOptions{HGap: 80, VGap: 40, OnlyNodes: []pfd.NodeID{"P1"}, PosRestriction: pfddrawio.PosLock}
	got := sortAndReadRects(t, doc, opts)

	if got["d1"] != rect(0, 0) || got["d2"] != rect(600, 200) || got["d3"] != rect(200, 100) {
		t.Errorf("非選択セルは不動のはず: d1=%v d2=%v d3=%v", got["d1"], got["d2"], got["d3"])
	}
	if got["p1"] != rect(200, 220) {
		t.Errorf("p1 は D3 を避けて (200,220) のはず: got %v", got["p1"])
	}
}

func TestSortOnlyNodeFreePushesNonSelected(t *testing.T) {

	testCases := map[string]struct {
		Mode             pfddrawio.PosRestriction
		D3X, D3Y         float64
		WantD3X, WantD3Y float64
	}{

		"free-v identical": {pfddrawio.PosFreeV, 200, 100, 200, 180},
		"free-h identical": {pfddrawio.PosFreeH, 200, 100, 320, 100},

		"free picks vertical":   {pfddrawio.PosFree, 210, 170, 210, 180},
		"free picks horizontal": {pfddrawio.PosFree, 280, 110, 320, 110},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			doc := drawioPage(
				rectCell("d1", "D1", 0, 0) +
					ellipseCell("p1", "P1", 50, 400) +
					rectCell("d2", "D2", 600, 200) +
					rectCell("d3", "D3", tc.D3X, tc.D3Y) +
					edgeCell("e1", "d1", "p1", false) +
					edgeCell("e2", "p1", "d2", false),
			)
			opts := pfddrawio.SortOptions{HGap: 80, VGap: 40, OnlyNodes: []pfd.NodeID{"P1"}, PosRestriction: tc.Mode}
			got := sortAndReadRects(t, doc, opts)

			if got["p1"] != rect(200, 100) {
				t.Errorf("p1 は目標 (200,100) に留まるはず: got %v", got["p1"])
			}
			if got["d1"] != rect(0, 0) || got["d2"] != rect(600, 200) {
				t.Errorf("d1/d2 は不動のはず: d1=%v d2=%v", got["d1"], got["d2"])
			}
			if got["d3"] != rect(tc.WantD3X, tc.WantD3Y) {
				t.Errorf("d3 want (%v,%v): got %v", tc.WantD3X, tc.WantD3Y, got["d3"])
			}
		})
	}
}

func TestSortOnlyNodeResolvesReversedChainAmongSelected(t *testing.T) {

	doc := drawioPage(
		rectCell("d1", "D1", 600, 0) +
			ellipseCell("p1", "P1", 300, 0) +
			rectCell("d2", "D2", 0, 0) +
			edgeCell("e1", "d1", "p1", false) +
			edgeCell("e2", "p1", "d2", false),
	)
	opts := pfddrawio.SortOptions{
		HGap: 80, VGap: 40,
		OnlyNodes:      []pfd.NodeID{"D1", "P1", "D2"},
		PosRestriction: pfddrawio.PosLock,
	}
	got := sortAndReadRects(t, doc, opts)

	want := map[pfddrawio.CellID]geom.Rect{
		"d1": rect(100, 0),
		"p1": rect(300, 0),
		"d2": rect(500, 0),
	}
	for id, w := range want {
		if got[id] != w {
			t.Errorf("%s = %v, want %v", id, got[id], w)
		}
	}
	if got["d1"].X+got["d1"].Width+opts.HGap > got["p1"].X {
		t.Errorf("辺 e1 (d1->p1) が逆流している: d1=%v p1=%v", got["d1"], got["p1"])
	}
	if got["p1"].X+got["p1"].Width+opts.HGap > got["d2"].X {
		t.Errorf("辺 e2 (p1->d2) が逆流している: p1=%v d2=%v", got["p1"], got["d2"])
	}
}

func TestUnmatchedOnlyNodes(t *testing.T) {
	testCases := map[string]struct {
		OnlyNodes []pfd.NodeID
		NodeIDs   []pfd.NodeID
		Want      []pfd.NodeID
	}{
		"all matched":       {[]pfd.NodeID{"P1", "D2"}, []pfd.NodeID{"D1", "P1", "D2"}, nil},
		"one unmatched":     {[]pfd.NodeID{"P1", "D9"}, []pfd.NodeID{"D1", "P1", "D2"}, []pfd.NodeID{"D9"}},
		"keeps input order": {[]pfd.NodeID{"D9", "P1", "X3"}, []pfd.NodeID{"P1"}, []pfd.NodeID{"D9", "X3"}},
		"empty only":        {nil, []pfd.NodeID{"P1"}, nil},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := pfddrawio.UnmatchedOnlyNodes(tc.OnlyNodes, tc.NodeIDs)
			if !slices.Equal(tc.Want, got) {
				t.Errorf("UnmatchedOnlyNodes(%v, %v) = %v, want %v", tc.OnlyNodes, tc.NodeIDs, got, tc.Want)
			}
		})
	}
}

func edgeStyleByID(t *testing.T, nodes []*xmldom.Node) map[pfddrawio.CellID]string {
	t.Helper()
	styles := make(map[pfddrawio.CellID]string)
	for _, dom := range pfddrawio.CollectDiagramDOMs(nodes) {
		for id, cell := range dom.CellsByID {
			if e, _ := cell.GetAttr("edge", ""); e != "1" {
				continue
			}
			s, _ := cell.GetAttr("style", "")
			styles[id] = s
		}
	}
	return styles
}

func TestSortOnlyNodeLeavesUnrelatedEdgesUntouched(t *testing.T) {

	doc := drawioPage(
		rectCell("d1", "D1", 0, 0) +
			ellipseCell("p1", "P1", 50, 400) +
			rectCell("d2", "D2", 600, 200) +
			rectCell("d4", "D4", 0, 800) +
			ellipseCell("p5", "P5", 300, 800) +
			edgeCell("e1", "d1", "p1", false) +
			edgeCell("e2", "p1", "d2", false) +
			`<mxCell id="eU" value="" style="edgeStyle=orthogonalEdgeStyle;html=1;" parent="1" source="d4" target="p5" edge="1"><mxGeometry relative="1" as="geometry"/></mxCell>`,
	)
	opts := pfddrawio.SortOptions{HGap: 80, VGap: 40, OnlyNodes: []pfd.NodeID{"P1"}, PosRestriction: pfddrawio.PosLock}
	nodes, err := pfddrawio.Sort(strings.NewReader(doc), opts, testLogger(t))
	if err != nil {
		t.Fatalf("Sort: %v", err)
	}
	styles := edgeStyleByID(t, nodes)
	if !strings.Contains(styles["eU"], "orthogonalEdgeStyle") {
		t.Errorf("無関係エッジ eU は直交のまま残るはず: %q", styles["eU"])
	}

	if strings.Contains(styles["e1"], "orthogonalEdgeStyle") {
		t.Errorf("選択に接続する e1 は直線化されるはず: %q", styles["e1"])
	}
}

func TestAverageCenterY(t *testing.T) {
	rects := map[pfddrawio.CellID]geom.Rect{
		"a": {X: 0, Y: 0, Width: 120, Height: 80},
		"b": {X: 0, Y: 200, Width: 120, Height: 80},
	}
	testCases := map[string]struct {
		IDs    []pfddrawio.CellID
		WantY  float64
		WantOK bool
	}{
		"single":          {[]pfddrawio.CellID{"a"}, 40, true},
		"average of two":  {[]pfddrawio.CellID{"a", "b"}, 140, true},
		"ignores missing": {[]pfddrawio.CellID{"a", "z"}, 40, true},
		"none present":    {[]pfddrawio.CellID{"z"}, 0, false},
		"empty":           {nil, 0, false},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			gotY, gotOK := pfddrawio.AverageCenterY(tc.IDs, rects)
			if gotOK != tc.WantOK || gotY != tc.WantY {
				t.Errorf("AverageCenterY(%v) = (%v, %v), want (%v, %v)", tc.IDs, gotY, gotOK, tc.WantY, tc.WantOK)
			}
		})
	}
}

func mutuallyDependentCompositesDoc() string {
	return `<mxfile>` + diagramPage("d0", "P0",
		rectCell("a1", "D1: 初期成果物", 40, 160)+
			ellipseCell("a2", "P10: 甲", 260, 160)+
			rectCell("a3", "D2: 甲が作り乙が使う", 460, 160)+
			ellipseCell("a4", "P20: 乙", 680, 160)+
			rectCell("a5", "D3: 乙が作り甲が使う", 460, 300)+
			rectCell("a6", "D4: 最終成果物", 260, 40)+
			edgeCell("e1", "a1", "a2", false)+
			edgeCell("e2", "a2", "a3", false)+
			edgeCell("e3", "a3", "a4", false)+
			edgeCell("e4", "a4", "a5", false)+
			edgeCell("e5", "a5", "a2", false)+
			edgeCell("e6", "a2", "a6", false)) + `</mxfile>`
}

func cellsWithNodeID(t *testing.T, nodes []*xmldom.Node, diagramID pfddrawio.DiagramID, nodeID pfd.NodeID) []pfddrawio.CellID {
	t.Helper()
	dom, ok := pfddrawio.CollectDiagramDOMs(nodes)[diagramID]
	if !ok {
		t.Fatalf("diagram %q not found", diagramID)
	}
	var ids []pfddrawio.CellID
	for id, cell := range dom.CellsByID {
		if got, ok := pfddrawio.VertexNodeID(cell); ok && got == nodeID {
			ids = append(ids, id)
		}
	}
	slices.SortFunc(ids, pfddrawio.CellID.Compare)
	return ids
}

func leftwardSolidEdges(t *testing.T, nodes []*xmldom.Node, diagramID pfddrawio.DiagramID) []pfddrawio.CellID {
	t.Helper()
	dom, ok := pfddrawio.CollectDiagramDOMs(nodes)[diagramID]
	if !ok {
		t.Fatalf("diagram %q not found", diagramID)
	}
	layers := pfddrawio.NewLayerMapFromNodes(nodes)
	parents := pfddrawio.NewParentMapFromNodes(nodes)
	vertices, edges := pfddrawio.CollectPageLayout(dom, layers, parents)
	rects := make(map[pfddrawio.CellID]geom.Rect, len(vertices))
	for _, v := range vertices {
		rects[v.ID] = v.Rect
	}
	var leftward []pfddrawio.CellID
	for _, e := range edges {
		if e.IsFeedback {
			continue
		}
		if rects[e.Source].X >= rects[e.Target].X {
			leftward = append(leftward, e.ID)
		}
	}
	return leftward
}

func TestSortBreakCycles(t *testing.T) {
	testCases := map[string]struct {
		BreakCycles bool
		WantErr     bool

		WantD3Cells int
	}{
		"without -break-cycles a mutual dependency is an error": {
			BreakCycles: false,
			WantErr:     true,
		},
		"with -break-cycles the cut edge is shown as a duplicate": {
			BreakCycles: true,
			WantD3Cells: 2,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			opts := pfddrawio.DefaultSortOptions()
			opts.BreakCycles = tc.BreakCycles
			nodes, err := pfddrawio.Sort(strings.NewReader(mutuallyDependentCompositesDoc()), opts, testLogger(t))
			if tc.WantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Sort: %v", err)
			}
			if got := cellsWithNodeID(t, nodes, "d0", "D3"); len(got) != tc.WantD3Cells {
				t.Errorf("D3 のセル数が %d 個のはずだが %d 個だった: %v", tc.WantD3Cells, len(got), got)
			}
			if got := leftwardSolidEdges(t, nodes, "d0"); len(got) > 0 {
				t.Errorf("左向きに残った実線がある: %v", got)
			}
		})
	}
}

func delNamedAt(id pfddrawio.CellID, nodeID pfd.NodeID, x float64) pfddrawio.LayoutVertex {
	return pfddrawio.LayoutVertex{
		ID: id, NodeID: nodeID, Kind: pfddrawio.VertexDeliverable,
		Rect: geom.Rect{X: x, Width: 120, Height: 80},
	}
}

func procNamedAt(id pfddrawio.CellID, nodeID pfd.NodeID, x float64) pfddrawio.LayoutVertex {
	return pfddrawio.LayoutVertex{
		ID: id, NodeID: nodeID, Kind: pfddrawio.VertexProcess,
		Rect: geom.Rect{X: x, Width: 120, Height: 80},
	}
}

func TestMarkCycleCuts(t *testing.T) {
	testCases := map[string]struct {
		Vertices []pfddrawio.LayoutVertex
		Edges    []pfddrawio.LayoutEdge

		WantCut []pfddrawio.CellID
	}{
		"an acyclic page is left untouched": {
			Vertices: []pfddrawio.LayoutVertex{
				delNamedAt("1", "D1", 0), procNamedAt("2", "P1", 200), delNamedAt("3", "D2", 400),
			},
			Edges: []pfddrawio.LayoutEdge{fwd("a", "1", "2"), fwd("b", "2", "3")},
		},
		"the deliverable edge drawn right-to-left is cut": {

			Vertices: []pfddrawio.LayoutVertex{
				delNamedAt("1", "D1", 40), procNamedAt("2", "P10", 260), delNamedAt("3", "D2", 460),
				procNamedAt("4", "P20", 680), delNamedAt("5", "D3", 460),
			},
			Edges: []pfddrawio.LayoutEdge{
				fwd("a", "1", "2"), fwd("b", "2", "3"), fwd("c", "3", "4"),
				fwd("d", "4", "5"), fwd("e", "5", "2"),
			},
			WantCut: []pfddrawio.CellID{"e"},
		},
		"two cycles sharing an edge are broken by one cut each": {

			Vertices: []pfddrawio.LayoutVertex{
				delNamedAt("1", "D1", 40), procNamedAt("2", "P1", 260), delNamedAt("3", "D2", 460),
				procNamedAt("4", "P2", 680), delNamedAt("5", "D3", 460), delNamedAt("6", "D4", 460),
			},
			Edges: []pfddrawio.LayoutEdge{
				fwd("a", "1", "2"), fwd("b", "2", "3"), fwd("c", "3", "4"),
				fwd("d", "4", "5"), fwd("e", "5", "2"), fwd("f", "4", "6"), fwd("g", "6", "2"),
			},
			WantCut: []pfddrawio.CellID{"e", "g"},
		},
		"a candidate no longer on a cycle is kept": {

			Vertices: []pfddrawio.LayoutVertex{
				procNamedAt("1", "P1", 260), delNamedAt("2", "D2", 460),
				procNamedAt("3", "P2", 680), delNamedAt("4", "D3", 880),
			},
			Edges: []pfddrawio.LayoutEdge{
				fwd("a", "1", "2"), fwd("b", "2", "3"), fwd("c", "3", "4"), fwd("d", "4", "1"),
			},
			WantCut: []pfddrawio.CellID{"d"},
		},
		"a cut that another cut already resolves is not kept": {

			Vertices: []pfddrawio.LayoutVertex{
				delNamedAt("1", "D1", 40), procNamedAt("2", "P1", 300), delNamedAt("3", "D2", 560),
				procNamedAt("4", "P2", 820), delNamedAt("5", "D3", 100),
				delNamedAt("6", "D5", 1080), procNamedAt("7", "P3", 1340),
			},
			Edges: []pfddrawio.LayoutEdge{
				fwd("a", "1", "2"), fwd("b", "2", "3"), fwd("c", "3", "4"),
				fwd("d", "4", "5"), fwd("e", "5", "2"),
				fwd("f", "4", "6"), fwd("g", "6", "7"), fwd("h", "7", "3"),
			},
			WantCut: []pfddrawio.CellID{"c"},
		},
		"a feedback edge is never a candidate": {

			Vertices: []pfddrawio.LayoutVertex{
				delNamedAt("1", "D1", 40), procNamedAt("2", "P1", 260), delNamedAt("3", "D2", 460),
			},
			Edges: []pfddrawio.LayoutEdge{fwd("a", "1", "2"), fwd("b", "2", "3"), back("c", "3", "2")},
		},
		"an edge whose endpoint is not a known vertex is not a candidate": {

			Vertices: []pfddrawio.LayoutVertex{procNamedAt("1", "P1", 40)},
			Edges:    []pfddrawio.LayoutEdge{fwd("a", "1", "x"), fwd("b", "x", "1")},
		},
		"a cycle with no deliverable-to-process edge is left to the ranking error": {

			Vertices: []pfddrawio.LayoutVertex{procNamedAt("1", "P1", 40), procNamedAt("2", "P2", 260)},
			Edges:    []pfddrawio.LayoutEdge{fwd("a", "1", "2"), fwd("b", "2", "1")},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			marked := pfddrawio.MarkCycleCuts(tc.Vertices, tc.Edges)

			var gotCut []pfddrawio.CellID
			for _, e := range marked {
				if e.IsCut {
					gotCut = append(gotCut, e.ID)
				}
			}
			slices.SortFunc(gotCut, pfddrawio.CellID.Compare)
			if diff := cmp.Diff(tc.WantCut, gotCut); diff != "" {
				t.Errorf("cut edges mismatch (-want +got):\n%s", diff)
			}

			for _, e := range tc.Edges {
				if e.IsCut {
					t.Errorf("入力の辺 %s が変更されている", e.ID)
				}
			}
		})
	}
}

func TestPlanDuplicationsCutEdge(t *testing.T) {

	vertices := []pfddrawio.LayoutVertex{
		delNamed("1", "D1"), procNamed("2", "P10"), delNamed("3", "D2"),
		procNamed("4", "P20"), delNamed("5", "D3"), procNamed("9", "P90"),
	}
	chain := []pfddrawio.LayoutEdge{
		fwd("a", "1", "2"), fwd("b", "2", "3"), fwd("c", "3", "4"), fwd("d", "4", "5"),
	}
	withChain := func(extra ...pfddrawio.LayoutEdge) []pfddrawio.LayoutEdge {
		return append(append([]pfddrawio.LayoutEdge{}, chain...), extra...)
	}
	ranks := map[pfddrawio.CellID]int{"1": 0, "2": 1, "3": 2, "4": 3, "5": 4}

	testCases := map[string]struct {
		Edges []pfddrawio.LayoutEdge

		RankSpan int
		WantDup  bool
	}{
		"a cut edge is duplicated even above the threshold": {
			Edges: withChain(cut("e", "5", "2")), RankSpan: 99, WantDup: true,
		},
		"a solid edge of the same span is not duplicated above the threshold": {
			Edges: withChain(fwd("e", "5", "2")), RankSpan: 99, WantDup: false,
		},

		"a cut edge with an unranked endpoint is not duplicated": {
			Edges: append(withChain(cut("e", "5", "9")), fwd("f", "9", "5")), RankSpan: 99, WantDup: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			dups, rewires := pfddrawio.PlanDuplications(vertices, tc.Edges,
				pfddrawio.DupCriteria{Ranks: ranks, RankSpan: tc.RankSpan})
			if !tc.WantDup {
				if len(dups) != 0 || len(rewires) != 0 {
					t.Errorf("複製しないはずだが dups=%+v rewires=%+v", dups, rewires)
				}
				return
			}
			if len(dups) != 1 || dups[0].SourceID != "5" {
				t.Fatalf("D3 の複製 1 個を期待したが %+v", dups)
			}

			if dups[0].Rank != 0 {
				t.Errorf("複製は rank 0 のはずだが %d", dups[0].Rank)
			}
			if len(rewires) != 1 || rewires[0].EdgeID != "e" || rewires[0].Endpoint != pfddrawio.EndpointSource {
				t.Errorf("切った辺の始端が複製へ付け替わるはずだが %+v", rewires)
			}
		})
	}
}

func twoCyclesSharingAnEdgeDoc() string {
	return `<mxfile>` + diagramPage("d0", "P0",
		rectCell("d1", "D1: 初期", 40, 200)+
			ellipseCell("p10", "P10: 甲", 300, 200)+
			rectCell("d2", "D2: 甲が作り乙が使う", 560, 200)+
			ellipseCell("p20", "P20: 乙", 820, 200)+
			rectCell("d3", "D3: 乙が作り甲が使う", 100, 400)+
			rectCell("d5", "D5: 乙が作り丙が使う", 1080, 400)+
			ellipseCell("p30", "P30: 丙", 1340, 400)+
			edgeCell("e1", "d1", "p10", false)+
			edgeCell("e2", "p10", "d2", false)+
			edgeCell("eA", "d2", "p20", false)+
			edgeCell("e4", "p20", "d3", false)+
			edgeCell("eB", "d3", "p10", false)+
			edgeCell("e6", "p20", "d5", false)+
			edgeCell("eC", "d5", "p30", false)+
			edgeCell("e8", "p30", "d2", false)) + `</mxfile>`
}

func TestSortIsIdempotentWithBreakCycles(t *testing.T) {

	testCases := map[string]struct{ Doc string }{
		"one cycle":                  {Doc: mutuallyDependentCompositesDoc()},
		"two cycles sharing an edge": {Doc: twoCyclesSharingAnEdgeDoc()},
	}

	opts := pfddrawio.DefaultSortOptions()
	opts.BreakCycles = true

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			once := sortToBytes(t, []byte(tc.Doc), opts)
			twice := sortToBytes(t, once, opts)
			thrice := sortToBytes(t, twice, opts)

			if diff := cmp.Diff(cellIDsByPage(t, once), cellIDsByPage(t, twice)); diff != "" {
				t.Errorf("2 回目の整列でセルの集合が変わった (-once +twice):\n%s", diff)
			}
			if diff := cmp.Diff(rectsByPage(t, twice), rectsByPage(t, thrice)); diff != "" {
				t.Errorf("落ち着いた図の整列で頂点が動いた (-twice +thrice):\n%s", diff)
			}
			if !bytes.Equal(twice, thrice) {
				t.Error("落ち着いた図の整列でドキュメントが変わった")
			}
		})
	}
}

func TestSortOnlyNodeBreakCycles(t *testing.T) {

	testCases := map[string]struct {
		BreakCycles bool
		WantErr     bool
	}{
		"without -break-cycles the topological order fails":     {BreakCycles: false, WantErr: true},
		"with -break-cycles the selected node is placed anyway": {BreakCycles: true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			opts := pfddrawio.DefaultSortOptions()
			opts.OnlyNodes = []pfd.NodeID{"D3"}
			opts.BreakCycles = tc.BreakCycles
			nodes, err := pfddrawio.Sort(strings.NewReader(mutuallyDependentCompositesDoc()), opts, testLogger(t))
			if tc.WantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Sort: %v", err)
			}

			if got := cellsWithNodeID(t, nodes, "d0", "D3"); len(got) != 1 {
				t.Errorf("D3 のセルは 1 個のままのはずだが %v", got)
			}
		})
	}
}

func TestSortCycleErrorMessage(t *testing.T) {

	duplicateCycleDoc := drawioPage(
		rectCell("1", "D1: 甲", 0, 0) +
			ellipseCell("2", "P1: 一", 200, 0) +
			rectCell("3", "D2: 乙", 400, 0) +
			rectCell("4", "D2: 乙", 0, 200) +
			ellipseCell("5", "P2: 二", 200, 200) +
			rectCell("6", "D1: 甲", 400, 200) +
			edgeCell("a", "1", "2", false) +
			edgeCell("b", "2", "3", false) +
			edgeCell("c", "4", "5", false) +
			edgeCell("d", "5", "6", false))

	partlyCuttableDoc := drawioPage(
		rectCell("1", "D1: 甲", 0, 0) +
			ellipseCell("2", "P1: 一", 200, 0) +
			rectCell("3", "D2: 乙", 400, 0) +
			ellipseCell("4", "P2: 二", 200, 200) +
			ellipseCell("5", "P3: 三", 400, 200) +
			edgeCell("a", "1", "2", false) +
			edgeCell("b", "2", "3", false) +
			edgeCell("c", "3", "2", false) +
			edgeCell("d", "4", "5", false) +
			edgeCell("e", "5", "4", false))

	testCases := map[string]struct {
		Doc string

		OnlyNodes []pfd.NodeID

		BreakCycles bool

		WantContains []string

		WantNotContains []string
	}{
		"a cycle without any duplicate cell does not blame duplicates": {
			Doc:             mutuallyDependentCompositesDoc(),
			WantContains:    []string{"-break-cycles", "D3"},
			WantNotContains: []string{"複製表示のセル"},
		},
		"a cycle created by duplicate cells still points at them": {
			Doc:          duplicateCycleDoc,
			WantContains: []string{"-break-cycles", "複製表示"},
		},
		"the -only-node path suggests -break-cycles too": {
			Doc:          mutuallyDependentCompositesDoc(),
			OnlyNodes:    []pfd.NodeID{"D3"},
			WantContains: []string{"-break-cycles"},
		},

		"with -break-cycles already on it does not suggest the same option": {
			Doc:             partlyCuttableDoc,
			BreakCycles:     true,
			WantContains:    []string{"切りましたが", "D2→P1"},
			WantNotContains: []string{"-break-cycles を付けると"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			opts := pfddrawio.DefaultSortOptions()
			opts.OnlyNodes = tc.OnlyNodes
			opts.BreakCycles = tc.BreakCycles
			_, err := pfddrawio.Sort(strings.NewReader(tc.Doc), opts, testLogger(t))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			for _, want := range tc.WantContains {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("エラーメッセージに %q が含まれていない: %v", want, err)
				}
			}
			for _, notWant := range tc.WantNotContains {
				if strings.Contains(err.Error(), notWant) {
					t.Errorf("エラーメッセージに %q が含まれてはいけない: %v", notWant, err)
				}
			}
		})
	}
}

func logicalEdges(t *testing.T, doc []byte) []string {
	t.Helper()
	diagrams, err := pfddrawio.ParseDiagrams(bytes.NewReader(doc), testLogger(t))
	if err != nil {
		t.Fatalf("ParseDiagrams: %v", err)
	}
	p, _, err := pfddrawio.NormalizeDiagrams("Example", diagrams, testLogger(t))
	if err != nil {
		t.Fatalf("NormalizeDiagrams: %v", err)
	}
	var edges []string
	for _, e := range p.Edges.Iter() {
		edges = append(edges, string(e.Source)+"->"+string(e.Target))
	}
	slices.Sort(edges)
	return edges
}

func TestSortBreakCyclesKeepsMeaning(t *testing.T) {

	testCases := map[string]struct{ Doc string }{
		"one cycle":                  {Doc: mutuallyDependentCompositesDoc()},
		"two cycles sharing an edge": {Doc: twoCyclesSharingAnEdgeDoc()},
	}

	opts := pfddrawio.DefaultSortOptions()
	opts.BreakCycles = true

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			src := []byte(tc.Doc)
			sorted := sortToBytes(t, src, opts)
			if diff := cmp.Diff(logicalEdges(t, src), logicalEdges(t, sorted)); diff != "" {
				t.Errorf("整列が論理 PFD の辺を変えた (-before +after):\n%s", diff)
			}
		})
	}
}

func TestPlanDuplicationsKeepsOriginalConnectedWithCutEdge(t *testing.T) {

	vertices := []pfddrawio.LayoutVertex{
		delNamed("1", "D1"), procNamed("2", "P10"), delNamed("3", "D2"),
		procNamed("4", "P20"), delNamed("5", "D3"),
	}
	edges := []pfddrawio.LayoutEdge{
		fwd("a", "1", "2"), fwd("b", "2", "3"), fwd("c", "3", "4"),
		fwd("d", "4", "5"), cut("e", "5", "2"),
	}
	ranks := map[pfddrawio.CellID]int{"1": 0, "2": 1, "3": 2, "4": 3, "5": 4}

	_, rewires := pfddrawio.PlanDuplications(vertices, edges,
		pfddrawio.DupCriteria{Ranks: ranks, RankSpan: 0})

	rewired := make(map[pfddrawio.CellID]bool, len(rewires))
	for _, r := range rewires {
		rewired[r.EdgeID] = true
	}
	if !rewired["e"] {
		t.Error("切断辺 e は必ず複製へ付け替えられなければならない（元の箱に残ると循環が解けない）")
	}
	if rewired["d"] {
		t.Error("元の箱に残す 1 本は生産辺 d でなければならない")
	}
}
