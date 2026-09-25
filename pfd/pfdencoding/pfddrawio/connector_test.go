package pfddrawio

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSpliceConnectors(t *testing.T) {
	connectorStyle := StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "aspect": "fixed"}
	processStyle := StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1"}
	edgeStyle := StyleMap{"edgeStyle": "none", "html": "1"}
	dashedStyle := StyleMap{"edgeStyle": "none", "html": "1", "dashed": "1"}

	testCases := map[string]struct {
		Cells    []Cell
		Expected []EffectiveEdge
	}{
		"direct edges pass through unchanged": {
			Cells: []Cell{
				NewVertex("a", "1", "A", processStyle),
				NewVertex("b", "1", "B", processStyle),
				NewEdge("e1", "1", "a", "b", edgeStyle),
			},
			Expected: []EffectiveEdge{
				{Source: "a", Target: "b", IsFeedback: false, Origins: []CellID{"e1"}},
			},
		},

		"duplicate direct edges between the same pair are both kept as origins": {
			Cells: []Cell{
				NewVertex("a", "1", "A", processStyle),
				NewVertex("b", "1", "B", processStyle),
				NewEdge("e1", "1", "a", "b", edgeStyle),
				NewEdge("e2", "1", "a", "b", edgeStyle),
			},
			Expected: []EffectiveEdge{
				{Source: "a", Target: "b", IsFeedback: false, Origins: []CellID{"e1"}},
				{Source: "a", Target: "b", IsFeedback: false, Origins: []CellID{"e2"}},
			},
		},
		"cartesian product a,b -> conn -> c,d": {
			Cells: []Cell{
				NewVertex("a", "1", "A", processStyle),
				NewVertex("b", "1", "B", processStyle),
				NewVertex("conn", "1", "", connectorStyle),
				NewVertex("c", "1", "C", processStyle),
				NewVertex("d", "1", "D", processStyle),
				NewEdge("e1", "1", "a", "conn", edgeStyle),
				NewEdge("e2", "1", "b", "conn", edgeStyle),
				NewEdge("e3", "1", "conn", "c", edgeStyle),
				NewEdge("e4", "1", "conn", "d", edgeStyle),
			},
			Expected: []EffectiveEdge{
				{Source: "a", Target: "c", IsFeedback: false, Origins: []CellID{"e1", "e3"}},
				{Source: "a", Target: "d", IsFeedback: false, Origins: []CellID{"e1", "e4"}},
				{Source: "b", Target: "c", IsFeedback: false, Origins: []CellID{"e2", "e3"}},
				{Source: "b", Target: "d", IsFeedback: false, Origins: []CellID{"e2", "e4"}},
			},
		},
		"direct dashed edge is feedback": {
			Cells: []Cell{
				NewEdge("e1", "1", "a", "b", dashedStyle),
			},
			Expected: []EffectiveEdge{
				{Source: "a", Target: "b", IsFeedback: true, Origins: []CellID{"e1"}},
			},
		},
		"feedback propagates via OR when a leg is dashed": {
			Cells: []Cell{
				NewVertex("conn", "1", "", connectorStyle),
				NewEdge("in", "1", "a", "conn", dashedStyle),
				NewEdge("out", "1", "conn", "c", edgeStyle),
			},
			Expected: []EffectiveEdge{
				{Source: "a", Target: "c", IsFeedback: true, Origins: []CellID{"in", "out"}},
			},
		},
		"chained connectors are transitively expanded": {

			Cells: []Cell{
				NewVertex("conn1", "1", "", connectorStyle),
				NewVertex("conn2", "1", "", connectorStyle),
				NewEdge("in", "1", "a", "conn1", edgeStyle),
				NewEdge("mid", "1", "conn1", "conn2", edgeStyle),
				NewEdge("out", "1", "conn2", "b", edgeStyle),
			},
			Expected: []EffectiveEdge{
				{Source: "a", Target: "b", IsFeedback: false, Origins: []CellID{"in", "mid", "out"}},
			},
		},
		"feedback propagates across a transitive chain": {

			Cells: []Cell{
				NewVertex("conn1", "1", "", connectorStyle),
				NewVertex("conn2", "1", "", connectorStyle),
				NewEdge("in", "1", "a", "conn1", edgeStyle),
				NewEdge("mid", "1", "conn1", "conn2", dashedStyle),
				NewEdge("out", "1", "conn2", "b", edgeStyle),
			},
			Expected: []EffectiveEdge{
				{Source: "a", Target: "b", IsFeedback: true, Origins: []CellID{"in", "mid", "out"}},
			},
		},
		"connector cycle is broken without hanging": {

			Cells: []Cell{
				NewVertex("conn1", "1", "", connectorStyle),
				NewVertex("conn2", "1", "", connectorStyle),
				NewEdge("in", "1", "a", "conn1", edgeStyle),
				NewEdge("mid", "1", "conn1", "conn2", edgeStyle),
				NewEdge("back", "1", "conn2", "conn1", edgeStyle),
				NewEdge("toC", "1", "conn2", "c", edgeStyle),
			},
			Expected: []EffectiveEdge{
				{Source: "a", Target: "c", IsFeedback: false, Origins: []CellID{"in", "mid", "toC"}},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := SpliceConnectors(testCase.Cells)
			if !reflect.DeepEqual(actual, testCase.Expected) {
				t.Error(cmp.Diff(testCase.Expected, actual))
			}
		})
	}
}

func TestDanglingConnectors(t *testing.T) {
	connectorStyle := StyleMap{"ellipse": "", "whiteSpace": "wrap", "html": "1", "aspect": "fixed"}
	edgeStyle := StyleMap{"edgeStyle": "none", "html": "1"}

	testCases := map[string]struct {
		Cells    []Cell
		Expected []CellID
	}{
		"balanced connector is not dangling": {
			Cells: []Cell{
				NewVertex("conn", "1", "", connectorStyle),
				NewEdge("in", "1", "a", "conn", edgeStyle),
				NewEdge("out", "1", "conn", "b", edgeStyle),
			},
			Expected: nil,
		},
		"connector with only an incoming edge is dangling": {
			Cells: []Cell{
				NewVertex("conn", "1", "", connectorStyle),
				NewEdge("in", "1", "a", "conn", edgeStyle),
			},
			Expected: []CellID{"conn"},
		},
		"connector with only an outgoing edge is dangling": {
			Cells: []Cell{
				NewVertex("conn", "1", "", connectorStyle),
				NewEdge("out", "1", "conn", "b", edgeStyle),
			},
			Expected: []CellID{"conn"},
		},
		"isolated connector is not dangling": {
			Cells: []Cell{
				NewVertex("conn", "1", "", connectorStyle),
			},
			Expected: nil,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := DanglingConnectors(testCase.Cells)
			if !reflect.DeepEqual(actual, testCase.Expected) {
				t.Error(cmp.Diff(testCase.Expected, actual))
			}
		})
	}
}
