package pfd

import (
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"slices"
	"testing"

	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/google/go-cmp/cmp"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func testNode(id, desc string, t NodeType) *Node {
	return &Node{ID: NodeID(id), Description: desc, Type: t}
}

func TestNewNodeMap(t *testing.T) {
	testCases := map[string]struct {
		Nodes    []*Node
		Expected map[NodeID]*Node
	}{
		"different IDs are all kept": {
			Nodes: []*Node{
				testNode("D1", "仕様", NodeTypeAtomicDeliverable),
				testNode("P1", "実装", NodeTypeAtomicProcess),
			},
			Expected: map[NodeID]*Node{
				"D1": testNode("D1", "仕様", NodeTypeAtomicDeliverable),
				"P1": testNode("P1", "実装", NodeTypeAtomicProcess),
			},
		},

		"boundary deliverable appearing on two pages": {
			Nodes: []*Node{
				testNode("D1", "仕様", NodeTypeAtomicDeliverable),
				testNode("D1", "仕様", NodeTypeAtomicDeliverable),
			},
			Expected: map[NodeID]*Node{
				"D1": testNode("D1", "仕様", NodeTypeAtomicDeliverable),
			},
		},

		"same ID with different descriptions keeps the first": {
			Nodes: []*Node{
				testNode("D1", "実装", NodeTypeAtomicDeliverable),
				testNode("D1", "仕様", NodeTypeAtomicDeliverable),
			},
			Expected: map[NodeID]*Node{
				"D1": testNode("D1", "仕様", NodeTypeAtomicDeliverable),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := NewNodeMap(sets.New((*Node).Compare, tc.Nodes...), testLogger())
			if !reflect.DeepEqual(actual, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, actual))
			}
		})
	}
}

func TestProducingAtomicProcess(t *testing.T) {
	testCases := map[string]struct {
		nodes     []*Node
		edges     []*Edge
		target    NodeID
		exclude   []NodeID
		wantID    NodeID
		wantFound bool
	}{
		"produced by one atomic process via normal edge": {
			nodes: []*Node{
				testNode("P1", "実装", NodeTypeAtomicProcess),
				testNode("D1", "実装コード", NodeTypeAtomicDeliverable),
			},
			edges:     []*Edge{{Source: "P1", Target: "D1"}},
			target:    "D1",
			wantID:    "P1",
			wantFound: true,
		},
		"produced via feedback edge only": {
			nodes: []*Node{
				testNode("P1", "レビュー", NodeTypeAtomicProcess),
				testNode("D1", "レビュー指摘", NodeTypeAtomicDeliverable),
			},
			edges:     []*Edge{{Source: "P1", Target: "D1", IsFeedback: true}},
			target:    "D1",
			wantID:    "P1",
			wantFound: true,
		},
		"no producing edge is initial deliverable": {
			nodes: []*Node{
				testNode("P1", "実装", NodeTypeAtomicProcess),
				testNode("D1", "設計書", NodeTypeAtomicDeliverable),
			},
			edges:     []*Edge{{Source: "D1", Target: "P1"}},
			target:    "D1",
			wantFound: false,
		},
		"context diagram P0 is excluded": {
			nodes: []*Node{
				testNode("P0", "全体", NodeTypeAtomicProcess),
				testNode("D1", "設計書", NodeTypeAtomicDeliverable),
			},
			edges:     []*Edge{{Source: "P0", Target: "D1"}},
			target:    "D1",
			wantFound: false,
		},
		"excluded id is skipped (feedback self-output)": {
			nodes: []*Node{
				testNode("P1", "実装", NodeTypeAtomicProcess),
				testNode("D1", "実装コード", NodeTypeAtomicDeliverable),
			},
			edges:     []*Edge{{Source: "P1", Target: "D1", IsFeedback: true}},
			target:    "D1",
			exclude:   []NodeID{"P1"},
			wantFound: false,
		},
		"composite process source only is not a producer": {
			nodes: []*Node{
				testNode("P1", "複合", NodeTypeCompositeProcess),
				testNode("D1", "設計書", NodeTypeAtomicDeliverable),
			},
			edges:     []*Edge{{Source: "P1", Target: "D1"}},
			target:    "D1",
			wantFound: false,
		},
		"defensive: two atomic producers returns deterministic min": {
			nodes: []*Node{
				testNode("P2", "実装B", NodeTypeAtomicProcess),
				testNode("P10", "実装C", NodeTypeAtomicProcess),
				testNode("D1", "実装コード", NodeTypeAtomicDeliverable),
			},
			edges: []*Edge{
				{Source: "P10", Target: "D1"},
				{Source: "P2", Target: "D1"},
			},
			target:    "D1",
			wantID:    "P2",
			wantFound: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			p := &PFD{
				Nodes: sets.New((*Node).Compare, tc.nodes...),
				Edges: sets.New((*Edge).Compare, tc.edges...),
			}
			nodeMap := NewNodeMap(p.Nodes, testLogger())
			gotID, gotFound := p.ProducingAtomicProcess(tc.target, nodeMap, tc.exclude...)
			if gotFound != tc.wantFound {
				t.Fatalf("ProducingAtomicProcess() found = %v, want %v", gotFound, tc.wantFound)
			}
			if gotFound && gotID != tc.wantID {
				t.Errorf("ProducingAtomicProcess() id = %q, want %q", gotID, tc.wantID)
			}
		})
	}
}

func TestExpandCompositeDeliverables(t *testing.T) {
	testCases := map[string]struct {
		DeliverableComposition map[NodeID]*sets.Set[NodeID]
		Given                  *sets.Set[NodeID]
		Expected               *sets.Set[NodeID]
	}{
		"empty": {
			DeliverableComposition: map[NodeID]*sets.Set[NodeID]{},
			Given:                  sets.New(NodeID.Compare),
			Expected:               sets.New(NodeID.Compare),
		},
		"atomic only": {
			DeliverableComposition: map[NodeID]*sets.Set[NodeID]{
				"D3": sets.New(NodeID.Compare, "D1", "D2"),
			},
			Given:    sets.New(NodeID.Compare, "D1", "D4"),
			Expected: sets.New(NodeID.Compare, "D1", "D4"),
		},
		"composite is replaced by its members": {
			DeliverableComposition: map[NodeID]*sets.Set[NodeID]{
				"D3": sets.New(NodeID.Compare, "D1", "D2"),
			},
			Given:    sets.New(NodeID.Compare, "D3", "D4"),
			Expected: sets.New(NodeID.Compare, "D1", "D2", "D4"),
		},

		"idempotent": {
			DeliverableComposition: map[NodeID]*sets.Set[NodeID]{
				"D3": sets.New(NodeID.Compare, "D1", "D2"),
			},
			Given:    sets.New(NodeID.Compare, "D1", "D2", "D3"),
			Expected: sets.New(NodeID.Compare, "D1", "D2"),
		},
		"composition is unknown": {
			DeliverableComposition: nil,
			Given:                  sets.New(NodeID.Compare, "D3"),
			Expected:               sets.New(NodeID.Compare, "D3"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			p := &PFD{DeliverableComposition: tc.DeliverableComposition}
			got := p.ExpandCompositeDeliverables(tc.Given)
			if !sets.IsEqual(NodeID.Compare, got, tc.Expected) {
				t.Errorf("got %v, expected %v", got.Slice(), tc.Expected.Slice())
			}
		})
	}
}

func TestGraphExceptFeedbackCompositeDeliverable(t *testing.T) {
	newPFD := func(deliverableComposition map[NodeID]*sets.Set[NodeID]) *PFD {
		return &PFD{
			Nodes: sets.New(
				(*Node).Compare,
				testNode("D1", "D1", NodeTypeAtomicDeliverable),
				testNode("D2", "D2", NodeTypeAtomicDeliverable),
				testNode("D3", "D3", NodeTypeCompositeDeliverable),
				testNode("P1", "P1", NodeTypeAtomicProcess),
			),
			Edges: sets.New(
				(*Edge).Compare,
				&Edge{Source: "D1", Target: "P1"},
				&Edge{Source: "P1", Target: "D3"},
				&Edge{Source: "P1", Target: "D2"},
			),
			DeliverableComposition: deliverableComposition,
		}
	}

	testCases := map[string]struct {
		PFD           *PFD
		ExpectedNodes []string
		ExpectedEdges []string
	}{
		"known composition": {
			PFD: newPFD(map[NodeID]*sets.Set[NodeID]{
				"D3": sets.New(NodeID.Compare, "D2"),
			}),
			ExpectedNodes: []string{"D1", "D2", "P1"},
			ExpectedEdges: []string{"D1->P1", "P1->D2"},
		},
		"unknown composition": {
			PFD:           newPFD(map[NodeID]*sets.Set[NodeID]{}),
			ExpectedNodes: []string{"D1", "D2", "D3", "P1"},
			ExpectedEdges: []string{"D1->P1", "P1->D2", "P1->D3"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := testLogger()
			g := tc.PFD.GraphExceptFeedback(NewNodeMap(tc.PFD.Nodes, logger), logger)

			got := make([]string, 0, g.Nodes.Len())
			for _, n := range g.Nodes.Iter() {
				got = append(got, string(n))
			}
			if !reflect.DeepEqual(got, tc.ExpectedNodes) {
				t.Errorf("nodes = %v, want %v", got, tc.ExpectedNodes)
			}

			gotEdges := make([]string, 0, g.Edges.Len())
			for _, e := range g.Edges.Iter() {
				gotEdges = append(gotEdges, fmt.Sprintf("%s->%s", e.First, e.Second))
			}
			slices.Sort(gotEdges)
			if !reflect.DeepEqual(gotEdges, tc.ExpectedEdges) {
				t.Errorf("edges = %v, want %v", gotEdges, tc.ExpectedEdges)
			}
		})
	}
}
