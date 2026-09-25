package pfddrawio

import (
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slograw"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/Kuniwak/pfd-tools/xmldom"
	"github.com/google/go-cmp/cmp"
)

func TestIsCommentLayer(t *testing.T) {
	testCases := map[string]struct {
		Cells    []Cell
		Query    CellID
		Expected bool
	}{
		"English Comment":         {Cells: []Cell{NewLayer("l", "Comment")}, Query: "l", Expected: true},
		"lowercase comment":       {Cells: []Cell{NewLayer("l", "comment")}, Query: "l", Expected: true},
		"English prefix Comments": {Cells: []Cell{NewLayer("l", "Comments")}, Query: "l", Expected: true},
		"Japanese コメント":           {Cells: []Cell{NewLayer("l", "コメント")}, Query: "l", Expected: true},
		"Japanese prefix コメント2":   {Cells: []Cell{NewLayer("l", "コメント2")}, Query: "l", Expected: true},
		"non-comment layer PFD":   {Cells: []Cell{NewLayer("l", "PFD")}, Query: "l", Expected: false},
		"empty layer name":        {Cells: []Cell{NewLayer("l", "")}, Query: "l", Expected: false},
		"unknown layer id":        {Cells: []Cell{NewLayer("l", "Comment")}, Query: "other", Expected: false},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			layerMap := NewLayerMap(testCase.Cells)
			if actual := layerMap.IsCommentLayer(testCase.Query); actual != testCase.Expected {
				t.Errorf("expected: %v, actual: %v", testCase.Expected, actual)
			}
		})
	}
}

func TestIsCommentDescendant(t *testing.T) {
	rect := StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}
	testCases := map[string]struct {
		Cells    []Cell
		Query    CellID
		Expected bool
	}{
		"direct child of comment layer": {
			Cells: []Cell{
				NewRoot("0"),
				NewLayer("c", "Comment"),
				NewVertex("v", "c", "X: コメント", rect),
			},
			Query:    "v",
			Expected: true,
		},
		"grandchild via group inside comment layer": {
			Cells: []Cell{
				NewRoot("0"),
				NewLayer("c", "Comment"),
				NewVertex("g", "c", "", rect),
				NewVertex("v", "g", "X: コメント", rect),
			},
			Query:    "v",
			Expected: true,
		},
		"deeply nested inside comment layer": {
			Cells: []Cell{
				NewRoot("0"),
				NewLayer("c", "Comment"),
				NewVertex("g1", "c", "", rect),
				NewVertex("g2", "g1", "", rect),
				NewVertex("v", "g2", "X: コメント", rect),
			},
			Query:    "v",
			Expected: true,
		},
		"direct child of non-comment layer": {
			Cells: []Cell{
				NewRoot("0"),
				NewLayer("l", "PFD"),
				NewVertex("v", "l", "P1: プロセス", rect),
			},
			Query:    "v",
			Expected: false,
		},
		"grandchild inside non-comment layer": {
			Cells: []Cell{
				NewRoot("0"),
				NewLayer("l", "PFD"),
				NewVertex("g", "l", "", rect),
				NewVertex("v", "g", "P1: プロセス", rect),
			},
			Query:    "v",
			Expected: false,
		},
		"parent id not found": {
			Cells: []Cell{
				NewRoot("0"),
				NewLayer("c", "Comment"),
				NewVertex("v", "nonexistent", "X: コメント", rect),
			},
			Query:    "v",
			Expected: false,
		},
		"unknown query id": {
			Cells: []Cell{
				NewRoot("0"),
				NewLayer("c", "Comment"),
			},
			Query:    "missing",
			Expected: false,
		},
		"cycle is broken without hanging": {
			Cells: []Cell{
				NewVertex("a", "b", "X: コメント", rect),
				NewVertex("b", "a", "Y: コメント", rect),
			},
			Query:    "a",
			Expected: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			layerMap := NewLayerMap(testCase.Cells)
			parents := NewParentMap(testCase.Cells)
			if actual := layerMap.IsCommentDescendant(parents, testCase.Query); actual != testCase.Expected {
				t.Errorf("expected: %v, actual: %v", testCase.Expected, actual)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	testCases := map[string]struct {
		File              []Diagram
		ExpectedPFD       *pfd.PFD
		ExpectedSourceMap *SourceMap
	}{
		"example": {
			File: exampleFile,
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{
						ID:          "D1",
						Description: "実装",
						Type:        pfd.NodeTypeAtomicDeliverable,
					},
					&pfd.Node{
						ID:          "D1",
						Description: "実装",
						Type:        pfd.NodeTypeAtomicDeliverable,
					},
					&pfd.Node{
						ID:          "D2",
						Description: "レビューコメント",
						Type:        pfd.NodeTypeAtomicDeliverable,
					},
					&pfd.Node{
						ID:          "D3",
						Description: "検証結果",
						Type:        pfd.NodeTypeAtomicDeliverable,
					},
					&pfd.Node{
						ID:          "D4",
						Description: "仕様",
						Type:        pfd.NodeTypeAtomicDeliverable,
					},
					&pfd.Node{
						ID:          "D4",
						Description: "仕様",
						Type:        pfd.NodeTypeAtomicDeliverable,
					},
					&pfd.Node{
						ID:          "P1",
						Description: "実装する",
						Type:        pfd.NodeTypeCompositeProcess,
					},
					&pfd.Node{
						ID:          "P2",
						Description: "レビューする",
						Type:        pfd.NodeTypeAtomicProcess,
					},
					&pfd.Node{
						ID:          "P3",
						Description: "検証する",
						Type:        pfd.NodeTypeAtomicProcess,
					},
					&pfd.Node{
						ID:          "P4",
						Description: "実装する",
						Type:        pfd.NodeTypeAtomicProcess,
					},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{
						Source: "D1",
						Target: "P2",
					},
					&pfd.Edge{
						Source: "D1",
						Target: "P3",
					},
					&pfd.Edge{
						Source:     "D2",
						Target:     "P1",
						IsFeedback: true,
					},
					&pfd.Edge{
						Source:     "D3",
						Target:     "P1",
						IsFeedback: true,
					},
					&pfd.Edge{
						Source: "D4",
						Target: "P1",
					},
					&pfd.Edge{
						Source: "D4",
						Target: "P3",
					},
					&pfd.Edge{
						Source: "D4",
						Target: "P4",
					},
					&pfd.Edge{
						Source: "P1",
						Target: "D1",
					},
					&pfd.Edge{
						Source: "P2",
						Target: "D2",
					},
					&pfd.Edge{
						Source: "P3",
						Target: "D3",
					},
					&pfd.Edge{
						Source: "P4",
						Target: "D1",
					},
				),
				ProcessComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{
					"P1": sets.New(pfd.NodeID.Compare, "P4"),
				},
				DeliverableComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
			},
			ExpectedSourceMap: exampleSourceMap,
		},
		"default_page_name_en": {
			File: []Diagram{
				{
					ID:   "d1",
					Name: "Page-1",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewVertex("2", "1", "P1: 実装する", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1"}),
					},
				},
			},
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "P1", Description: "実装する", Type: pfd.NodeTypeAtomicProcess},
				),
				Edges:                  sets.New((*pfd.Edge).Compare),
				ProcessComposition:     map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
				DeliverableComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
			},
			ExpectedSourceMap: &SourceMap{
				NodeIDMap: map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"P1": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "2"}),
				},
				EdgeIDMap: map[pfd.NodeID]map[pfd.NodeID]*sets.Set[DrawIOLocation]{},
			},
		},
		"default_page_name_ja": {
			File: []Diagram{
				{
					ID:   "d1",
					Name: "ページ-1",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewVertex("2", "1", "P1: 実装する", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1"}),
					},
				},
			},
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "P1", Description: "実装する", Type: pfd.NodeTypeAtomicProcess},
				),
				Edges:                  sets.New((*pfd.Edge).Compare),
				ProcessComposition:     map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
				DeliverableComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
			},
			ExpectedSourceMap: &SourceMap{
				NodeIDMap: map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"P1": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "2"}),
				},
				EdgeIDMap: map[pfd.NodeID]map[pfd.NodeID]*sets.Set[DrawIOLocation]{},
			},
		},
		"described_detail_page_name": {

			File: []Diagram{
				{
					ID:   "d0",
					Name: "P0",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewVertex("2", "1", "P1: 実装する", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "strokeWidth": "2"}),
					},
				},
				{
					ID:   "d1",
					Name: "P1: 実装する",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewVertex("3", "1", "P2: 実装作業", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1"}),
					},
				},
			},
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "P1", Description: "実装する", Type: pfd.NodeTypeCompositeProcess},
					&pfd.Node{ID: "P2", Description: "実装作業", Type: pfd.NodeTypeAtomicProcess},
				),
				Edges: sets.New((*pfd.Edge).Compare),
				ProcessComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{
					"P1": sets.New(pfd.NodeID.Compare, "P2"),
				},
				DeliverableComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
			},
			ExpectedSourceMap: &SourceMap{
				NodeIDMap: map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"P1": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d0", CellID: "2"}),
					"P2": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "3"}),
				},
				EdgeIDMap: map[pfd.NodeID]map[pfd.NodeID]*sets.Set[DrawIOLocation]{},
			},
		},
		"borderless_atomic_process": {

			File: []Diagram{
				{
					ID:   "d1",
					Name: "P0",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewVertex("2", "1", "P1: 実装する", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "strokeWidth": "0"}),
					},
				},
			},
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "P1", Description: "実装する", Type: pfd.NodeTypeAtomicProcess},
				),
				Edges:                  sets.New((*pfd.Edge).Compare),
				ProcessComposition:     map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
				DeliverableComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
			},
			ExpectedSourceMap: &SourceMap{
				NodeIDMap: map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"P1": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "2"}),
				},
				EdgeIDMap: map[pfd.NodeID]map[pfd.NodeID]*sets.Set[DrawIOLocation]{},
			},
		},
		"borderless_atomic_deliverable": {

			File: []Diagram{
				{
					ID:   "d1",
					Name: "P0",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewVertex("2", "1", "D1: 初期成果物", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1", "strokeWidth": "0"}),
					},
				},
			},
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D1", Description: "初期成果物", Type: pfd.NodeTypeAtomicDeliverable},
				),
				Edges:                  sets.New((*pfd.Edge).Compare),
				ProcessComposition:     map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
				DeliverableComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
			},
			ExpectedSourceMap: &SourceMap{
				NodeIDMap: map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"D1": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "2"}),
				},
				EdgeIDMap: map[pfd.NodeID]map[pfd.NodeID]*sets.Set[DrawIOLocation]{},
			},
		},
		"composite process without detail page": {

			File: []Diagram{
				{
					ID:   "d1",
					Name: "P0",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewVertex("2", "1", "D1: 入力", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
						NewVertex("3", "1", "P1: 実装する", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "strokeWidth": "2"}),
						NewVertex("4", "1", "D2: 出力", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
						NewEdge("5", "1", "2", "3", StyleMap{"edgeStyle": "none", "html": "1"}),
						NewEdge("6", "1", "3", "4", StyleMap{"edgeStyle": "none", "html": "1"}),
					},
				},
			},
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D1", Description: "入力", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "D2", Description: "出力", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P1", Description: "実装する", Type: pfd.NodeTypeAtomicProcess},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{Source: "D1", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D2"},
				),
				ProcessComposition:      map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
				DeliverableComposition:  map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
				ImplicitAtomicProcesses: sets.New(pfd.NodeID.Compare, "P1"),
			},
			ExpectedSourceMap: &SourceMap{
				NodeIDMap: map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"D1": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "2"}),
					"P1": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "3"}),
					"D2": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "4"}),
				},
				EdgeIDMap: map[pfd.NodeID]map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"D1": {"P1": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "5"})},
					"P1": {"D2": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "6"})},
				},
			},
		},

		"an unnumbered process and deliverable with the same label": {
			File: []Diagram{
				{
					ID:   "d1",
					Name: "P0",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewVertex("2", "1", "要件", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
						NewVertex("3", "1", "実装する", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1"}),
						NewVertex("4", "1", "実装する", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
						NewEdge("5", "1", "2", "3", StyleMap{"edgeStyle": "none", "html": "1"}),
						NewEdge("6", "1", "3", "4", StyleMap{"edgeStyle": "none", "html": "1"}),
					},
				},
			},
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "[要件]", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "(実装する)", Type: pfd.NodeTypeAtomicProcess},
					&pfd.Node{ID: "[実装する]", Type: pfd.NodeTypeAtomicDeliverable},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{Source: "[要件]", Target: "(実装する)"},
					&pfd.Edge{Source: "(実装する)", Target: "[実装する]"},
				),
				ProcessComposition:     map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
				DeliverableComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
			},
			ExpectedSourceMap: &SourceMap{
				NodeIDMap: map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"[要件]":   sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "2"}),
					"(実装する)": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "3"}),
					"[実装する]": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "4"}),
				},
				EdgeIDMap: map[pfd.NodeID]map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"[要件]":   {"(実装する)": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "5"})},
					"(実装する)": {"[実装する]": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "6"})},
				},
			},
		},

		"an unnumbered composite process and its detail page": {
			File: []Diagram{
				{
					ID:   "d1",
					Name: "P0",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewVertex("2", "1", "設計する", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "strokeWidth": "2"}),
					},
				},
				{
					ID:   "d2",
					Name: "設計する",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewVertex("2", "1", "実装する", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1"}),
					},
				},
			},
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "(設計する)", Type: pfd.NodeTypeCompositeProcess},
					&pfd.Node{ID: "(実装する)", Type: pfd.NodeTypeAtomicProcess},
				),
				Edges: sets.New((*pfd.Edge).Compare),
				ProcessComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{
					"(設計する)": sets.New(pfd.NodeID.Compare, "(実装する)"),
				},
				DeliverableComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
			},
			ExpectedSourceMap: &SourceMap{
				NodeIDMap: map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"(設計する)": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "2"}),
					"(実装する)": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d2", CellID: "2"}),
				},
				EdgeIDMap: map[pfd.NodeID]map[pfd.NodeID]*sets.Set[DrawIOLocation]{},
			},
		},
		"an unnumbered thick ellipse without a detail page stays atomic": {
			File: []Diagram{
				{
					ID:   "d1",
					Name: "P0",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewVertex("2", "1", "設計する", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "strokeWidth": "2"}),
					},
				},
			},
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "(設計する)", Type: pfd.NodeTypeAtomicProcess},
				),
				Edges:                   sets.New((*pfd.Edge).Compare),
				ProcessComposition:      map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
				DeliverableComposition:  map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
				ImplicitAtomicProcesses: sets.New(pfd.NodeID.Compare, "(設計する)"),
			},
			ExpectedSourceMap: &SourceMap{
				NodeIDMap: map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"(設計する)": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "2"}),
				},
				EdgeIDMap: map[pfd.NodeID]map[pfd.NodeID]*sets.Set[DrawIOLocation]{},
			},
		},

		"a vertex labelled with a bare assigned ID is not wrapped": {
			File: []Diagram{
				{
					ID:   "d1",
					Name: "P0",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewVertex("2", "1", "P1", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1"}),
						NewVertex("3", "1", "D1", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
					},
				},
			},
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "P1", Type: pfd.NodeTypeAtomicProcess},
					&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
				),
				Edges:                  sets.New((*pfd.Edge).Compare),
				ProcessComposition:     map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
				DeliverableComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
			},
			ExpectedSourceMap: &SourceMap{
				NodeIDMap: map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"P1": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "2"}),
					"D1": sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: "d1", CellID: "3"}),
				},
				EdgeIDMap: map[pfd.NodeID]map[pfd.NodeID]*sets.Set[DrawIOLocation]{},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			actual, srcMap, err := NormalizeDiagrams("Example", testCase.File, logger)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, testCase.ExpectedPFD) {
				t.Error(cmp.Diff(testCase.ExpectedPFD, actual))
			}
			if !reflect.DeepEqual(srcMap, testCase.ExpectedSourceMap) {
				t.Error(cmp.Diff(testCase.ExpectedSourceMap, srcMap))
			}
		})
	}
}

func TestNormalizeConnector(t *testing.T) {
	rect := StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}
	proc := StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1"}
	compositeProc := StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "strokeWidth": "2"}
	conn := StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "aspect": "fixed"}
	edge := StyleMap{"edgeStyle": "none", "html": "1"}

	loc := func(cellID CellID) DrawIOLocation {
		return DrawIOLocation{DiagramID: "diag0", CellID: cellID}
	}

	testCases := map[string]struct {
		File              []Diagram
		ExpectedPFD       *pfd.PFD
		ExpectedSourceMap *SourceMap
	}{

		"cartesian product diamond": {
			File: []Diagram{
				{
					ID:   "diag0",
					Name: "P0",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewVertex("2", "1", "D1: 初期成果物", rect),
						NewVertex("3", "1", "P1: プロセス1", proc),
						NewVertex("5", "1", "D2: 中間成果物A", rect),
						NewVertex("7", "1", "P4: プロセス4", proc),
						NewVertex("10", "1", "D0: 最終成果物", rect),
						NewVertex("12", "1", "P2: プロセス2", proc),
						NewVertex("14", "1", "D3: 中間成果物B", rect),
						NewVertex("16", "1", "", conn),
						NewVertex("24", "1", "", conn),
						NewVertex("28", "1", "P3: プロセス3", proc),
						NewVertex("30", "1", "D4: 中間成果物C", rect),
						NewEdge("20", "1", "2", "16", edge),
						NewEdge("6", "1", "3", "5", edge),
						NewEdge("25", "1", "5", "24", edge),
						NewEdge("11", "1", "7", "10", edge),
						NewEdge("15", "1", "12", "14", edge),
						NewEdge("27", "1", "14", "24", edge),
						NewEdge("19", "1", "16", "12", edge),
						NewEdge("21", "1", "16", "3", edge),
						NewEdge("29", "1", "16", "28", edge),
						NewEdge("26", "1", "24", "7", edge),
						NewEdge("31", "1", "28", "30", edge),
						NewEdge("32", "1", "30", "24", edge),
					},
				},
			},
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D0", Description: "最終成果物", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "D1", Description: "初期成果物", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "D2", Description: "中間成果物A", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "D3", Description: "中間成果物B", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "D4", Description: "中間成果物C", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P1", Description: "プロセス1", Type: pfd.NodeTypeAtomicProcess},
					&pfd.Node{ID: "P2", Description: "プロセス2", Type: pfd.NodeTypeAtomicProcess},
					&pfd.Node{ID: "P3", Description: "プロセス3", Type: pfd.NodeTypeAtomicProcess},
					&pfd.Node{ID: "P4", Description: "プロセス4", Type: pfd.NodeTypeAtomicProcess},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{Source: "D1", Target: "P1"},
					&pfd.Edge{Source: "D1", Target: "P2"},
					&pfd.Edge{Source: "D1", Target: "P3"},
					&pfd.Edge{Source: "P1", Target: "D2"},
					&pfd.Edge{Source: "P2", Target: "D3"},
					&pfd.Edge{Source: "P3", Target: "D4"},
					&pfd.Edge{Source: "D2", Target: "P4"},
					&pfd.Edge{Source: "D3", Target: "P4"},
					&pfd.Edge{Source: "D4", Target: "P4"},
					&pfd.Edge{Source: "P4", Target: "D0"},
				),
				ProcessComposition:     map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
				DeliverableComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
			},
			ExpectedSourceMap: &SourceMap{
				NodeIDMap: map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"D0": sets.New(DrawIOLocation.Compare, loc("10")),
					"D1": sets.New(DrawIOLocation.Compare, loc("2")),
					"D2": sets.New(DrawIOLocation.Compare, loc("5")),
					"D3": sets.New(DrawIOLocation.Compare, loc("14")),
					"D4": sets.New(DrawIOLocation.Compare, loc("30")),
					"P1": sets.New(DrawIOLocation.Compare, loc("3")),
					"P2": sets.New(DrawIOLocation.Compare, loc("12")),
					"P3": sets.New(DrawIOLocation.Compare, loc("28")),
					"P4": sets.New(DrawIOLocation.Compare, loc("7")),
				},
				EdgeIDMap: map[pfd.NodeID]map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"D1": {
						"P1": sets.New(DrawIOLocation.Compare, loc("20"), loc("21")),
						"P2": sets.New(DrawIOLocation.Compare, loc("20"), loc("19")),
						"P3": sets.New(DrawIOLocation.Compare, loc("20"), loc("29")),
					},
					"P1": {"D2": sets.New(DrawIOLocation.Compare, loc("6"))},
					"P2": {"D3": sets.New(DrawIOLocation.Compare, loc("15"))},
					"P3": {"D4": sets.New(DrawIOLocation.Compare, loc("31"))},
					"D2": {"P4": sets.New(DrawIOLocation.Compare, loc("25"), loc("26"))},
					"D3": {"P4": sets.New(DrawIOLocation.Compare, loc("27"), loc("26"))},
					"D4": {"P4": sets.New(DrawIOLocation.Compare, loc("32"), loc("26"))},
					"P4": {"D0": sets.New(DrawIOLocation.Compare, loc("11"))},
				},
			},
		},

		"dedup and location union": {
			File: []Diagram{
				{
					ID:   "diag0",
					Name: "P0",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewVertex("a", "1", "D1: 成果物", rect),
						NewVertex("b", "1", "P1: プロセス", proc),
						NewVertex("c1", "1", "", conn),
						NewVertex("c2", "1", "", conn),
						NewEdge("ac1", "1", "a", "c1", edge),
						NewEdge("c1b", "1", "c1", "b", edge),
						NewEdge("ac2", "1", "a", "c2", edge),
						NewEdge("c2b", "1", "c2", "b", edge),
					},
				},
			},
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D1", Description: "成果物", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P1", Description: "プロセス", Type: pfd.NodeTypeAtomicProcess},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{Source: "D1", Target: "P1"},
				),
				ProcessComposition:     map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
				DeliverableComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
			},
			ExpectedSourceMap: &SourceMap{
				NodeIDMap: map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"D1": sets.New(DrawIOLocation.Compare, loc("a")),
					"P1": sets.New(DrawIOLocation.Compare, loc("b")),
				},
				EdgeIDMap: map[pfd.NodeID]map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"D1": {
						"P1": sets.New(DrawIOLocation.Compare, loc("ac1"), loc("c1b"), loc("ac2"), loc("c2b")),
					},
				},
			},
		},

		"chained connectors expand transitively": {
			File: []Diagram{
				{
					ID:   "diag0",
					Name: "P0",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewVertex("2", "1", "D1: 初期成果物", rect),
						NewVertex("3", "1", "P1: プロセス", proc),
						NewVertex("c1", "1", "", conn),
						NewVertex("c2", "1", "", conn),
						NewEdge("in", "1", "2", "c1", edge),
						NewEdge("mid", "1", "c1", "c2", edge),
						NewEdge("out", "1", "c2", "3", edge),
					},
				},
			},
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D1", Description: "初期成果物", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P1", Description: "プロセス", Type: pfd.NodeTypeAtomicProcess},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{Source: "D1", Target: "P1"},
				),
				ProcessComposition:     map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
				DeliverableComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
			},
			ExpectedSourceMap: &SourceMap{
				NodeIDMap: map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"D1": sets.New(DrawIOLocation.Compare, loc("2")),
					"P1": sets.New(DrawIOLocation.Compare, loc("3")),
				},
				EdgeIDMap: map[pfd.NodeID]map[pfd.NodeID]*sets.Set[DrawIOLocation]{
					"D1": {"P1": sets.New(DrawIOLocation.Compare, loc("in"), loc("mid"), loc("out"))},
				},
			},
		},

		"connector on comment layer is ignored": {
			File: []Diagram{
				{
					ID:   "diag0",
					Name: "P0",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewLayer("99", "Comment"),
						NewVertex("2", "1", "D1: 初期成果物", rect),
						NewVertex("3", "1", "P1: プロセス", proc),
						NewEdge("4", "1", "2", "3", edge),
						NewVertex("c", "99", "", conn),
						NewVertex("x", "99", "X: 無視", proc),
						NewVertex("y", "99", "Y: 無視", proc),
						NewEdge("ex", "99", "x", "c", edge),
						NewEdge("ey", "99", "c", "y", edge),
					},
				},
			},
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D1", Description: "初期成果物", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P1", Description: "プロセス", Type: pfd.NodeTypeAtomicProcess},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{Source: "D1", Target: "P1"},
				),
				ProcessComposition:     map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
				DeliverableComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
			},
		},

		"grouped comment descendants are ignored": {
			File: []Diagram{
				{
					ID:   "diag0",
					Name: "P0",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewLayer("99", "Comment"),
						NewVertex("2", "1", "D1: 初期成果物", rect),
						NewVertex("3", "1", "P1: プロセス", proc),
						NewEdge("4", "1", "2", "3", edge),
						NewVertex("g", "99", "", StyleMap{"group": "1"}),
						NewVertex("gx", "g", "P9: 埋もれプロセス", proc),
						NewVertex("gy", "g", "D9: 埋もれ成果物", rect),
						NewEdge("ge", "g", "gy", "gx", edge),
					},
				},
			},
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D1", Description: "初期成果物", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P1", Description: "プロセス", Type: pfd.NodeTypeAtomicProcess},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{Source: "D1", Target: "P1"},
				),
				ProcessComposition:     map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
				DeliverableComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
			},
		},

		"connector on detail page does not leak empty id": {
			File: []Diagram{
				{
					ID:    "diag0",
					Name:  "P0",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("2", "1", "P1: 実装する", compositeProc)},
				},
				{
					ID:   "diag1",
					Name: "P1",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewVertex("10", "1", "P2: 作業A", proc),
						NewVertex("11", "1", "P3: 作業B", proc),
						NewVertex("12", "1", "", conn),
						NewEdge("20", "1", "10", "12", edge),
						NewEdge("21", "1", "12", "11", edge),
					},
				},
			},
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "P1", Description: "実装する", Type: pfd.NodeTypeCompositeProcess},
					&pfd.Node{ID: "P2", Description: "作業A", Type: pfd.NodeTypeAtomicProcess},
					&pfd.Node{ID: "P3", Description: "作業B", Type: pfd.NodeTypeAtomicProcess},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{Source: "P2", Target: "P3"},
				),
				ProcessComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{
					"P1": sets.New(pfd.NodeID.Compare, "P2", "P3"),
				},
				DeliverableComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
			},
		},

		"dangling connector drops its leg": {
			File: []Diagram{
				{
					ID:   "diag0",
					Name: "P0",
					Cells: []Cell{
						NewRoot("0"),
						NewLayer("1", ""),
						NewVertex("2", "1", "D1: 初期成果物", rect),
						NewVertex("c", "1", "", conn),
						NewEdge("4", "1", "2", "c", edge),
					},
				},
			},
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D1", Description: "初期成果物", Type: pfd.NodeTypeAtomicDeliverable},
				),
				Edges:                  sets.New((*pfd.Edge).Compare),
				ProcessComposition:     map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
				DeliverableComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			actualPFD, actualSourceMap, err := NormalizeDiagrams("Example", testCase.File, logger)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actualPFD, testCase.ExpectedPFD) {
				t.Error(cmp.Diff(testCase.ExpectedPFD, actualPFD))
			}
			if testCase.ExpectedSourceMap != nil && !reflect.DeepEqual(actualSourceMap, testCase.ExpectedSourceMap) {
				t.Error(cmp.Diff(testCase.ExpectedSourceMap, actualSourceMap))
			}
		})
	}
}

func TestNormalizeDiagramsError(t *testing.T) {
	compositeProcess := StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "strokeWidth": "2"}
	atomicProcess := StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "strokeWidth": "1"}
	atomicDeliverable := StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1", "strokeWidth": "1"}
	compositeDeliverable := StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1", "strokeWidth": "2"}

	testCases := map[string]struct {
		File []Diagram
	}{
		"uniq-context-page: duplicated context diagram": {
			File: []Diagram{
				{

					ID:    "d0",
					Name:  "P0",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("2", "1", "P1: 実装する", atomicProcess)},
				},
				{
					ID:    "d1",
					Name:  "P0: コンテキストダイアグラム",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("3", "1", "P2: 設計する", atomicProcess)},
				},
			},
		},
		"uniq-page-name: duplicated detail page with a different description": {
			File: []Diagram{
				{
					ID:    "d0",
					Name:  "P0",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("2", "1", "P1: 実装する", compositeProcess)},
				},
				{

					ID:    "d1",
					Name:  "P1: 実装する",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("3", "1", "P2: 実装作業", atomicProcess)},
				},
				{
					ID:    "d2",
					Name:  "P1: べつの説明",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("4", "1", "P3: べつの作業", atomicProcess)},
				},
			},
		},
		"page-name-comp: leftover default page": {
			File: []Diagram{
				{
					ID:    "d0",
					Name:  "P0",
					Cells: []Cell{NewRoot("0"), NewLayer("1", "")},
				},
				{

					ID:    "d1",
					Name:  "ページ-2",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("2", "1", "P1: 実装作業", atomicProcess)},
				},
			},
		},
		"page-name-comp: undefined composite process": {
			File: []Diagram{
				{
					ID:    "d0",
					Name:  "P0",
					Cells: []Cell{NewRoot("0"), NewLayer("1", "")},
				},
				{

					ID:    "d1",
					Name:  "P99",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("2", "1", "P1: 実装作業", atomicProcess)},
				},
			},
		},
		"page-name-comp: page name is atomic process": {
			File: []Diagram{
				{

					ID:    "d0",
					Name:  "P0",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("2", "1", "P1: 実装する", atomicProcess)},
				},
				{
					ID:    "d1",
					Name:  "P1",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("3", "1", "P2: 実装作業", atomicProcess)},
				},
			},
		},
		"nesting cycle": {

			File: []Diagram{
				{
					ID:    "d0",
					Name:  "P0",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("2", "1", "P1: 開発する", compositeProcess)},
				},
				{
					ID:    "d1",
					Name:  "P1",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("3", "1", "P2: 実装する", compositeProcess)},
				},
				{
					ID:    "d2",
					Name:  "P2",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("4", "1", "P1: 開発する", compositeProcess)},
				},
			},
		},

		"uniq-page-name: duplicated detail page name": {
			File: []Diagram{
				{
					ID:    "d0",
					Name:  "P0",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("2", "1", "P1: 開発する", compositeProcess)},
				},
				{
					ID:    "d1",
					Name:  "P1",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("3", "1", "P2: 実装する", atomicProcess)},
				},
				{
					ID:    "d2",
					Name:  "P1",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("4", "1", "P3: 別の実装をする", atomicProcess)},
				},
			},
		},

		"uniq-page-id: duplicated diagram ID": {
			File: []Diagram{
				{
					ID:    "d0",
					Name:  "P0",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("2", "1", "P1: 開発する", compositeProcess)},
				},
				{
					ID:    "d0",
					Name:  "P1",
					Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("3", "1", "P2: 実装する", atomicProcess)},
				},
			},
		},

		"same element ID as a rectangle and an ellipse": {
			File: []Diagram{
				{
					ID:   "d0",
					Name: "P0",
					Cells: []Cell{
						NewRoot("0"), NewLayer("1", ""),
						NewVertex("2", "1", "D1: しかく", atomicDeliverable),
						NewVertex("3", "1", "D1: だえん", atomicProcess),
					},
				},
			},
		},

		"same element ID and description with different shapes": {
			File: []Diagram{
				{
					ID:   "d0",
					Name: "P0",
					Cells: []Cell{
						NewRoot("0"), NewLayer("1", ""),
						NewVertex("2", "1", "D1: おなじ", atomicDeliverable),
						NewVertex("3", "1", "D1: おなじ", atomicProcess),
					},
				},
			},
		},

		"same deliverable ID drawn thin on one page and thick on another": {
			File: []Diagram{
				{
					ID:   "d0",
					Name: "P0",
					Cells: []Cell{
						NewRoot("0"), NewLayer("1", ""),
						NewVertex("2", "1", "P1: 開発する", compositeProcess),
						NewVertex("3", "1", "D1: 境界成果物", atomicDeliverable),
					},
				},
				{
					ID:   "d1",
					Name: "P1",
					Cells: []Cell{
						NewRoot("0"), NewLayer("1", ""),
						NewVertex("4", "1", "P2: 実装する", atomicProcess),
						NewVertex("5", "1", "D1: 境界成果物", compositeDeliverable),
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			_, _, err := NormalizeDiagrams("Example", testCase.File, logger)
			if err == nil {
				t.Fatal("expected an error for the invalid diagrams, got nil")
			}
		})
	}
}

func TestNormalizeDiagramsPageNameDescriptionWarning(t *testing.T) {
	compositeProcess := StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "strokeWidth": "2"}
	atomicProcess := StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "strokeWidth": "1"}

	file := func(compositeLabel, pageName string) []Diagram {
		return []Diagram{
			{
				ID:    "d0",
				Name:  "P0",
				Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("2", "1", compositeLabel, compositeProcess)},
			},
			{
				ID:    "d1",
				Name:  pageName,
				Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("3", "1", "P2: 実装作業", atomicProcess)},
			},
		}
	}

	testCases := map[string]struct {
		File            []Diagram
		ExpectedWarning string
	}{
		"matched": {
			File:            file("P1: 実装する", "P1: 実装する"),
			ExpectedWarning: "",
		},
		"mismatched": {
			File:            file("P1: 実装する", "P1: 設計する"),
			ExpectedWarning: "WARN: pfddrawio.NormalizeDiagramsWithOptions: ページ名の説明が複合プロセスのラベルと一致しません: page=P1: 設計する, id=P1, pageDesc=設計する, nodeDesc=実装する\n",
		},
		"page name without description": {
			File:            file("P1: 実装する", "P1"),
			ExpectedWarning: "",
		},

		"composite process label without description": {
			File:            file("P1", "P1: 実装する"),
			ExpectedWarning: "",
		},
		"unnumbered": {
			File:            file("実装する", "実装する"),
			ExpectedWarning: "",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			sb := &strings.Builder{}
			logger := slog.New(slograw.NewHandler(sb, slog.LevelWarn))
			if _, _, err := NormalizeDiagrams("Example", testCase.File, logger); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(testCase.ExpectedWarning, sb.String()); diff != "" {
				t.Errorf("warnings mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNormalizeNestedComposite(t *testing.T) {
	compositeProcess := StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "strokeWidth": "2"}
	atomicProcess := StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "strokeWidth": "1"}

	file := []Diagram{
		{
			ID:    "d0",
			Name:  "P0",
			Cells: []Cell{NewRoot("0"), NewLayer("1", ""), NewVertex("c1", "1", "P1: 開発する", compositeProcess)},
		},
		{
			ID:   "d1",
			Name: "P1",
			Cells: []Cell{
				NewRoot("0"), NewLayer("1", ""),
				NewVertex("c2", "1", "P2: 実装する", compositeProcess),
				NewVertex("a5", "1", "P5: リリースする", atomicProcess),
			},
		},
		{
			ID:   "d2",
			Name: "P2",
			Cells: []Cell{
				NewRoot("0"), NewLayer("1", ""),
				NewVertex("a3", "1", "P3: 設計する", atomicProcess),
				NewVertex("a4", "1", "P4: コーディングする", atomicProcess),
			},
		},
	}

	logger := slog.New(slogtest.NewTestHandler(t))
	actual, _, err := NormalizeDiagrams("Example", file, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := map[pfd.NodeID]*sets.Set[pfd.NodeID]{
		"P1": sets.New(pfd.NodeID.Compare, "P3", "P4", "P5"),
		"P2": sets.New(pfd.NodeID.Compare, "P3", "P4"),
	}
	if !reflect.DeepEqual(actual.ProcessComposition, expected) {
		t.Errorf("ProcessComposition mismatch:\n%s", cmp.Diff(expected, actual.ProcessComposition))
	}

	if _, err := pfd.NewSafePFDByUnsafePFD(actual); err != nil {
		t.Errorf("NewSafePFDByUnsafePFD failed for a nested composite PFD: %v", err)
	}
}

const layerMapFromNodesXML = `<mxfile host="test">
  <diagram id="d1" name="P0">
    <mxGraphModel>
      <root>
        <mxCell id="0"/>
        <mxCell id="1" value="背景" parent="0"/>
        <mxCell id="cmt1" value="Comment" parent="0"/>
        <mxCell id="cmt2" value="comment (draft)" parent="0"/>
        <mxCell id="cmt3" value="COMMENT" parent="0"/>
        <mxCell id="2" value="D1" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
          <mxGeometry x="0" y="0" width="120" height="80" as="geometry"/>
        </mxCell>
        <mxCell id="3" value="メモ" style="text;html=1;" parent="cmt1" vertex="1">
          <mxGeometry x="0" y="200" width="120" height="80" as="geometry"/>
        </mxCell>
      </root>
    </mxGraphModel>
  </diagram>
</mxfile>`

func TestNewLayerMapFromNodes(t *testing.T) {
	nodes, err := xmldom.ParseXML(strings.NewReader(layerMapFromNodesXML))
	if err != nil {
		t.Fatal(err)
	}
	layerMap := NewLayerMapFromNodes(nodes)

	testCases := map[string]struct {
		ID       CellID
		Expected bool
	}{
		"default layer is not a comment layer":  {ID: "1", Expected: false},
		"layer named Comment":                   {ID: "cmt1", Expected: true},
		"layer name with comment prefix":        {ID: "cmt2", Expected: true},
		"layer name in uppercase COMMENT":       {ID: "cmt3", Expected: true},
		"non-layer cell id is not a comment":    {ID: "2", Expected: false},
		"unknown id is not a comment":           {ID: "nope", Expected: false},
		"root cell id 0 is not a comment layer": {ID: "0", Expected: false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if got := layerMap.IsCommentLayer(tc.ID); got != tc.Expected {
				t.Errorf("IsCommentLayer(%q) = %v, want %v", tc.ID, got, tc.Expected)
			}
		})
	}
}

func TestNormalizeDiagramsAllowDetachedDetailPage(t *testing.T) {
	atomicProcess := StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "strokeWidth": "1"}
	atomicDeliverable := StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1", "strokeWidth": "1"}
	edge := StyleMap{"edgeStyle": "none", "html": "1"}

	detached := []Diagram{
		{
			ID:   "d0",
			Name: "P1: 設計する",
			Cells: []Cell{
				NewRoot("0"), NewLayer("1", ""),
				NewVertex("2", "1", "D1: 要求", atomicDeliverable),
				NewVertex("3", "1", "P2: 方式を決める", atomicProcess),
				NewEdge("4", "1", "2", "3", edge),
			},
		},
	}

	misnamed := []Diagram{
		{
			ID:   "d0",
			Name: "P2",
			Cells: []Cell{
				NewRoot("0"), NewLayer("1", ""),
				NewVertex("2", "1", "D1: 要求", atomicDeliverable),
				NewVertex("3", "1", "P2: 方式を決める", atomicProcess),
				NewEdge("4", "1", "2", "3", edge),
			},
		},
	}

	testCases := map[string]struct {
		File            []Diagram
		Options         NormalizeOptions
		WantErr         bool
		ExpectedWarning string
	}{
		"page-name-comp: a detached detail page is rejected by default": {
			File:    detached,
			Options: NormalizeOptions{},
			WantErr: true,
		},
		"page-name-comp: a detached detail page is allowed with AllowDetachedDetailPage": {
			File:            detached,
			Options:         NormalizeOptions{AllowDetachedDetailPage: true},
			WantErr:         false,
			ExpectedWarning: "WARN: pfddrawio.NormalizeDiagramsWithOptions: ページ名が指す要素が PFD 中にありません（断片として読んでいるため続行します）: page=P1: 設計する\n",
		},

		"page-name-comp: a page named after an atomic process is rejected even with AllowDetachedDetailPage": {
			File:    misnamed,
			Options: NormalizeOptions{AllowDetachedDetailPage: true},
			WantErr: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			sb := &strings.Builder{}
			logger := slog.New(slograw.NewHandler(sb, slog.LevelWarn))
			p, _, err := NormalizeDiagramsWithOptions("", tc.File, tc.Options, logger)
			if tc.WantErr {
				if err == nil {
					t.Fatal("err = nil, want an error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.ExpectedWarning, sb.String()); diff != "" {
				t.Errorf("warnings mismatch (-want +got):\n%s", diff)
			}

			wantNodes := sets.New(
				(*pfd.Node).Compare,
				&pfd.Node{ID: "D1", Description: "要求", Type: pfd.NodeTypeAtomicDeliverable},
				&pfd.Node{ID: "P2", Description: "方式を決める", Type: pfd.NodeTypeAtomicProcess},
			)
			if !reflect.DeepEqual(wantNodes, p.Nodes) {
				t.Errorf("nodes mismatch (-want +got):\n%s", cmp.Diff(wantNodes, p.Nodes))
			}

			wantEdges := sets.New(
				(*pfd.Edge).Compare,
				&pfd.Edge{Source: "D1", Target: "P2"},
			)
			if !reflect.DeepEqual(wantEdges, p.Edges) {
				t.Errorf("edges mismatch (-want +got):\n%s", cmp.Diff(wantEdges, p.Edges))
			}
		})
	}
}
