package pfddrawio

import (
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
)

var exampleFile = []Diagram{
	{
		ID:   "PrNXtJMoFKdakpcB9KSm",
		Name: "P0",
		Cells: []Cell{
			NewRoot("0"),
			NewLayer("1", "PFD"),
			NewEdge("4", "1", "2", "3", StyleMap{"edgeStyle": "none", "html": "1"}),
			NewVertex("2", "1", "D4: Specification", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
			NewEdge("6", "1", "3", "5", StyleMap{"edgeStyle": "none", "html": "1"}),
			NewVertex("3", "1", "P1: Implement", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "strokeColor": "default", "strokeWidth": "2"}),
			NewEdge("9", "1", "5", "8", StyleMap{"edgeStyle": "none", "html": "1"}),
			NewVertex("5", "1", "D1: Implementation", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
			NewEdge("11", "1", "8", "10", StyleMap{"edgeStyle": "none", "html": "1"}),
			NewVertex("8", "1", "P2: Review", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1"}),
			NewEdge("12", "1", "10", "3", StyleMap{"edgeStyle": "orthogonalEdgeStyle", "html": "1", "entryX": "0.5", "entryY": "0", "entryDx": "0", "entryDy": "0", "dashed": "1"}),
			NewVertex("10", "1", "D2: Review\ncomments", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
			NewEdge("17", "1", "13", "16", StyleMap{"edgeStyle": "none", "html": "1"}),
			NewVertex("13", "1", "P3: Verify", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1"}),
			NewEdge("14", "1", "5", "13", StyleMap{"edgeStyle": "orthogonalEdgeStyle", "html": "1"}),
			NewEdge("15", "1", "2", "13", StyleMap{"edgeStyle": "orthogonalEdgeStyle", "html": "1"}),
			NewEdge("18", "1", "16", "3", StyleMap{"edgeStyle": "orthogonalEdgeStyle", "html": "1", "jumpStyle": "gap", "dashed": "1"}),
			NewVertex("16", "1", "D3: Verification result", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
			NewLayer("20", "Comment"),
			NewVertex("21", "20", "Comments", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
		},
	},
	{
		ID:   "g7jUJpcNx9cCV950jQFk",
		Name: "P1",
		Cells: []Cell{
			NewRoot("0"),
			NewLayer("1", ""),
			NewEdge("h_2N5Sa6hUkllRCv-iia-3", "1", "hO8RJ9AdBYUCDEpOiIOb-1", "h_2N5Sa6hUkllRCv-iia-1", StyleMap{"edgeStyle": "none", "html": "1"}),
			NewVertex("hO8RJ9AdBYUCDEpOiIOb-1", "1", "D4: Specification", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
			NewVertex("hO8RJ9AdBYUCDEpOiIOb-2", "1", "D1: Implementation", StyleMap{"rounded": "0", "whiteSpace": "wrap", "html": "1"}),
			NewEdge("h_2N5Sa6hUkllRCv-iia-2", "1", "h_2N5Sa6hUkllRCv-iia-1", "hO8RJ9AdBYUCDEpOiIOb-2", StyleMap{"edgeStyle": "none", "html": "1"}),
			NewVertex("h_2N5Sa6hUkllRCv-iia-1", "1", "P4: Implement", StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "strokeColor": "default", "strokeWidth": "1"}),
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
