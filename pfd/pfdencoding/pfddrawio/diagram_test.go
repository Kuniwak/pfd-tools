package pfddrawio

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
)

func TestCellIsConnector(t *testing.T) {
	connectorStyle := StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "aspect": "fixed"}
	processStyle := StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1"}
	rectangleStyle := StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}

	testCases := map[string]struct {
		Cell     Cell
		Expected bool
	}{
		"connector": {
			Cell:     NewVertex("1", "0", "", connectorStyle),
			Expected: true,
		},
		"process ellipse without aspect": {
			Cell:     NewVertex("1", "0", "P1: プロセス", processStyle),
			Expected: false,
		},
		"empty-value ellipse without aspect": {
			Cell:     NewVertex("1", "0", "", processStyle),
			Expected: false,
		},
		"connector-shaped but non-empty value": {
			Cell:     NewVertex("1", "0", "X", connectorStyle),
			Expected: false,
		},
		"rectangle with empty value": {
			Cell:     NewVertex("1", "0", "", rectangleStyle),
			Expected: false,
		},
		"edge": {
			Cell:     NewEdge("1", "0", "2", "3", connectorStyle),
			Expected: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			if actual := testCase.Cell.IsConnector(); actual != testCase.Expected {
				t.Errorf("expected: %v, actual: %v", testCase.Expected, actual)
			}
		})
	}
}

func TestStyleMapIsAtomicStrokeWidth(t *testing.T) {
	testCases := map[string]struct {
		Style    StyleMap
		Expected bool
	}{
		"borderless (strokeWidth=0)": {
			Style:    StyleMap{"strokeWidth": "0"},
			Expected: true,
		},
		"thin line (strokeWidth=1)": {
			Style:    StyleMap{"strokeWidth": "1"},
			Expected: true,
		},
		"thick line (strokeWidth=2)": {
			Style:    StyleMap{"strokeWidth": "2"},
			Expected: false,
		},
		"attribute missing (defaults to 1)": {
			Style:    StyleMap{"ellipse": ""},
			Expected: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			if actual := testCase.Style.IsAtomicStrokeWidth(); actual != testCase.Expected {
				t.Errorf("expected: %v, actual: %v", testCase.Expected, actual)
			}
		})
	}
}

var exampleFile = []Diagram{
	{
		ID:   "PrNXtJMoFKdakpcB9KSm",
		Name: "P0",
		Cells: []Cell{
			NewRoot("0"),
			NewLayer("1", "PFD"),
			NewEdge("4", "1", "2", "3", StyleMap{"edgeStyle": "none", "html": "1"}),
			NewVertex("2", "1", "D4: 仕様", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
			NewEdge("6", "1", "3", "5", StyleMap{"edgeStyle": "none", "html": "1"}),
			NewVertex("3", "1", "P1: 実装する", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "strokeColor": "default", "strokeWidth": "2"}),
			NewEdge("9", "1", "5", "8", StyleMap{"edgeStyle": "none", "html": "1"}),
			NewVertex("5", "1", "D1: 実装", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
			NewEdge("11", "1", "8", "10", StyleMap{"edgeStyle": "none", "html": "1"}),
			NewVertex("8", "1", "P2: レビューする", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1"}),
			NewEdge("12", "1", "10", "3", StyleMap{"edgeStyle": "orthogonalEdgeStyle", "html": "1", "entryX": "0.5", "entryY": "0", "entryDx": "0", "entryDy": "0", "dashed": "1"}),
			NewVertex("10", "1", "D2: レビューコメント", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
			NewEdge("17", "1", "13", "16", StyleMap{"edgeStyle": "none", "html": "1"}),
			NewVertex("13", "1", "P3: 検証する", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1"}),
			NewEdge("14", "1", "5", "13", StyleMap{"edgeStyle": "orthogonalEdgeStyle", "html": "1"}),
			NewEdge("15", "1", "2", "13", StyleMap{"edgeStyle": "orthogonalEdgeStyle", "html": "1"}),
			NewEdge("18", "1", "16", "3", StyleMap{"edgeStyle": "orthogonalEdgeStyle", "html": "1", "jumpStyle": "gap", "dashed": "1"}),
			NewVertex("16", "1", "D3: 検証結果", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
			NewLayer("20", "Comment"),
			NewVertex("21", "20", "コメント", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
		},
	},
	{
		ID:   "g7jUJpcNx9cCV950jQFk",
		Name: "P1",
		Cells: []Cell{
			NewRoot("0"),
			NewLayer("1", ""),
			NewEdge("h_2N5Sa6hUkllRCv-iia-3", "1", "hO8RJ9AdBYUCDEpOiIOb-1", "h_2N5Sa6hUkllRCv-iia-1", StyleMap{"edgeStyle": "none", "html": "1"}),
			NewVertex("hO8RJ9AdBYUCDEpOiIOb-1", "1", "D4: 仕様", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
			NewVertex("hO8RJ9AdBYUCDEpOiIOb-2", "1", "D1: 実装", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
			NewEdge("h_2N5Sa6hUkllRCv-iia-2", "1", "h_2N5Sa6hUkllRCv-iia-1", "hO8RJ9AdBYUCDEpOiIOb-2", StyleMap{"edgeStyle": "none", "html": "1"}),
			NewVertex("h_2N5Sa6hUkllRCv-iia-1", "1", "P4: 実装する", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "strokeColor": "default", "strokeWidth": "1"}),
		},
	},
}

var exampleSourceMap = &SourceMap{
	NodeIDMap: map[pfd.NodeID]*sets.Set[DrawIOLocation]{
		"D1": sets.New(
			DrawIOLocation.Compare,
			DrawIOLocation{DiagramID: "PrNXtJMoFKdakpcB9KSm", CellID: "5"},
			DrawIOLocation{DiagramID: "g7jUJpcNx9cCV950jQFk", CellID: "hO8RJ9AdBYUCDEpOiIOb-2"},
		),
		"D2": sets.New(
			DrawIOLocation.Compare,
			DrawIOLocation{DiagramID: "PrNXtJMoFKdakpcB9KSm", CellID: "10"},
		),
		"D3": sets.New(
			DrawIOLocation.Compare,
			DrawIOLocation{DiagramID: "PrNXtJMoFKdakpcB9KSm", CellID: "16"},
		),
		"D4": sets.New(
			DrawIOLocation.Compare,
			DrawIOLocation{DiagramID: "PrNXtJMoFKdakpcB9KSm", CellID: "2"},
			DrawIOLocation{DiagramID: "g7jUJpcNx9cCV950jQFk", CellID: "hO8RJ9AdBYUCDEpOiIOb-1"},
		),
		"P1": sets.New(
			DrawIOLocation.Compare,
			DrawIOLocation{DiagramID: "PrNXtJMoFKdakpcB9KSm", CellID: "3"},
		),
		"P2": sets.New(
			DrawIOLocation.Compare,
			DrawIOLocation{DiagramID: "PrNXtJMoFKdakpcB9KSm", CellID: "8"},
		),
		"P3": sets.New(
			DrawIOLocation.Compare,
			DrawIOLocation{DiagramID: "PrNXtJMoFKdakpcB9KSm", CellID: "13"},
		),
		"P4": sets.New(
			DrawIOLocation.Compare,
			DrawIOLocation{DiagramID: "g7jUJpcNx9cCV950jQFk", CellID: "h_2N5Sa6hUkllRCv-iia-1"},
		),
	},
	EdgeIDMap: map[pfd.NodeID]map[pfd.NodeID]*sets.Set[DrawIOLocation]{
		"D1": {
			"P2": sets.New(
				DrawIOLocation.Compare,
				DrawIOLocation{DiagramID: "PrNXtJMoFKdakpcB9KSm", CellID: "9"},
			),
			"P3": sets.New(
				DrawIOLocation.Compare,
				DrawIOLocation{DiagramID: "PrNXtJMoFKdakpcB9KSm", CellID: "14"},
			),
		},
		"D2": {
			"P1": sets.New(
				DrawIOLocation.Compare,
				DrawIOLocation{DiagramID: "PrNXtJMoFKdakpcB9KSm", CellID: "12"},
			),
		},
		"D3": {
			"P1": sets.New(
				DrawIOLocation.Compare,
				DrawIOLocation{DiagramID: "PrNXtJMoFKdakpcB9KSm", CellID: "18"},
			),
		},
		"D4": {
			"P1": sets.New(
				DrawIOLocation.Compare,
				DrawIOLocation{DiagramID: "PrNXtJMoFKdakpcB9KSm", CellID: "4"},
			),
			"P3": sets.New(
				DrawIOLocation.Compare,
				DrawIOLocation{DiagramID: "PrNXtJMoFKdakpcB9KSm", CellID: "15"},
			),
			"P4": sets.New(
				DrawIOLocation.Compare,
				DrawIOLocation{DiagramID: "g7jUJpcNx9cCV950jQFk", CellID: "h_2N5Sa6hUkllRCv-iia-3"},
			),
		},
		"P1": {"D1": sets.New(
			DrawIOLocation.Compare,
			DrawIOLocation{DiagramID: "PrNXtJMoFKdakpcB9KSm", CellID: "6"},
		)},
		"P2": {"D2": sets.New(
			DrawIOLocation.Compare,
			DrawIOLocation{DiagramID: "PrNXtJMoFKdakpcB9KSm", CellID: "11"},
		)},
		"P3": {"D3": sets.New(
			DrawIOLocation.Compare,
			DrawIOLocation{DiagramID: "PrNXtJMoFKdakpcB9KSm", CellID: "17"},
		)},
		"P4": {
			"D1": sets.New(
				DrawIOLocation.Compare,
				DrawIOLocation{DiagramID: "g7jUJpcNx9cCV950jQFk", CellID: "h_2N5Sa6hUkllRCv-iia-2"},
			),
		},
	},
}

func TestIsContextDiagramName(t *testing.T) {
	testCases := map[string]struct {
		Name     string
		Expected bool
	}{
		"context diagram ID":                    {Name: "P0", Expected: true},
		"context diagram ID with description":   {Name: "P0: コンテキストダイアグラム", Expected: true},
		"default page name en":                  {Name: "Page-1", Expected: true},
		"default page name en with description": {Name: "Page-1: 全体図", Expected: true},
		"default page name ja":                  {Name: "ページ-1", Expected: true},
		"default page name ja with description": {Name: "ページ-1: 全体図", Expected: true},
		"detail page":                           {Name: "P1", Expected: false},
		"detail page with description":          {Name: "P1: 実装する", Expected: false},
		"unnumbered detail page":                {Name: "実装する", Expected: false},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := IsContextDiagramName(testCase.Name)
			if actual != testCase.Expected {
				t.Errorf("IsContextDiagramName(%q) = %v, want %v", testCase.Name, actual, testCase.Expected)
			}
		})
	}
}

func TestCellIDCompare(t *testing.T) {
	testCases := map[string]struct {
		A    CellID
		B    CellID
		Want int
	}{
		"equal":               {A: "3", B: "3", Want: 0},
		"shorter is smaller":  {A: "9", B: "10", Want: -1},
		"longer is larger":    {A: "10", B: "9", Want: 1},
		"same length lexical": {A: "2", B: "3", Want: -1},
		"non-numeric lexical": {A: "aa", B: "ab", Want: -1},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := tc.A.Compare(tc.B)
			if sign(got) != tc.Want {
				t.Errorf("CellID(%q).Compare(%q) = %d (sign %d), want sign %d", tc.A, tc.B, got, sign(got), tc.Want)
			}
		})
	}
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}
