package pfd

import (
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/google/go-cmp/cmp"
)

func TestFlattenProcessComposition(t *testing.T) {
	testCases := map[string]struct {
		NodeMap  map[NodeID]*Node
		Direct   map[NodeID]*sets.Set[NodeID]
		Expected map[NodeID]*sets.Set[NodeID]
		WantErr  bool
	}{

		"flat": {
			NodeMap: map[NodeID]*Node{
				"P1": {ID: "P1", Type: NodeTypeCompositeProcess},
				"P2": {ID: "P2", Type: NodeTypeAtomicProcess},
				"P3": {ID: "P3", Type: NodeTypeAtomicProcess},
			},
			Direct: map[NodeID]*sets.Set[NodeID]{
				"P1": sets.New(NodeID.Compare, "P2", "P3"),
			},
			Expected: map[NodeID]*sets.Set[NodeID]{
				"P1": sets.New(NodeID.Compare, "P2", "P3"),
			},
		},

		"one-level": {
			NodeMap: map[NodeID]*Node{
				"P1": {ID: "P1", Type: NodeTypeCompositeProcess},
				"P2": {ID: "P2", Type: NodeTypeCompositeProcess},
				"P3": {ID: "P3", Type: NodeTypeAtomicProcess},
				"P4": {ID: "P4", Type: NodeTypeAtomicProcess},
				"P5": {ID: "P5", Type: NodeTypeAtomicProcess},
			},
			Direct: map[NodeID]*sets.Set[NodeID]{
				"P1": sets.New(NodeID.Compare, "P2", "P3"),
				"P2": sets.New(NodeID.Compare, "P4", "P5"),
			},
			Expected: map[NodeID]*sets.Set[NodeID]{
				"P1": sets.New(NodeID.Compare, "P3", "P4", "P5"),
				"P2": sets.New(NodeID.Compare, "P4", "P5"),
			},
		},

		"multi-level": {
			NodeMap: map[NodeID]*Node{
				"P1": {ID: "P1", Type: NodeTypeCompositeProcess},
				"P2": {ID: "P2", Type: NodeTypeCompositeProcess},
				"P3": {ID: "P3", Type: NodeTypeCompositeProcess},
				"P4": {ID: "P4", Type: NodeTypeAtomicProcess},
			},
			Direct: map[NodeID]*sets.Set[NodeID]{
				"P1": sets.New(NodeID.Compare, "P2"),
				"P2": sets.New(NodeID.Compare, "P3"),
				"P3": sets.New(NodeID.Compare, "P4"),
			},
			Expected: map[NodeID]*sets.Set[NodeID]{
				"P1": sets.New(NodeID.Compare, "P4"),
				"P2": sets.New(NodeID.Compare, "P4"),
				"P3": sets.New(NodeID.Compare, "P4"),
			},
		},

		"cycle": {
			NodeMap: map[NodeID]*Node{
				"P1": {ID: "P1", Type: NodeTypeCompositeProcess},
				"P2": {ID: "P2", Type: NodeTypeCompositeProcess},
			},
			Direct: map[NodeID]*sets.Set[NodeID]{
				"P1": sets.New(NodeID.Compare, "P2"),
				"P2": sets.New(NodeID.Compare, "P1"),
			},
			WantErr: true,
		},

		"self": {
			NodeMap: map[NodeID]*Node{
				"P1": {ID: "P1", Type: NodeTypeCompositeProcess},
			},
			Direct: map[NodeID]*sets.Set[NodeID]{
				"P1": sets.New(NodeID.Compare, "P1"),
			},
			WantErr: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual, err := FlattenProcessComposition(testCase.Direct, testCase.NodeMap)
			if testCase.WantErr {
				if err == nil {
					t.Fatalf("expected an error, but got nil (actual=%v)", actual)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(actual, testCase.Expected) {
				t.Error(cmp.Diff(testCase.Expected, actual))
			}
		})
	}
}

func TestFlattenDeliverableComposition(t *testing.T) {
	testCases := map[string]struct {
		Direct   map[NodeID]*sets.Set[NodeID]
		Nodes    *sets.Set[*Node]
		Expected map[NodeID]*sets.Set[NodeID]
		WantErr  bool
	}{

		"flat": {
			Direct: map[NodeID]*sets.Set[NodeID]{
				"D3": sets.New(NodeID.Compare, "D1", "D2"),
			},
			Expected: map[NodeID]*sets.Set[NodeID]{
				"D3": sets.New(NodeID.Compare, "D1", "D2"),
			},
		},

		"one-level": {
			Direct: map[NodeID]*sets.Set[NodeID]{
				"D3": sets.New(NodeID.Compare, "D1", "D2"),
				"D6": sets.New(NodeID.Compare, "D3", "D5"),
			},
			Expected: map[NodeID]*sets.Set[NodeID]{
				"D3": sets.New(NodeID.Compare, "D1", "D2"),
				"D6": sets.New(NodeID.Compare, "D1", "D2", "D5"),
			},
		},

		"multi-level": {
			Direct: map[NodeID]*sets.Set[NodeID]{
				"D3": sets.New(NodeID.Compare, "D1", "D2"),
				"D6": sets.New(NodeID.Compare, "D3"),
				"D9": sets.New(NodeID.Compare, "D6"),
			},
			Expected: map[NodeID]*sets.Set[NodeID]{
				"D3": sets.New(NodeID.Compare, "D1", "D2"),
				"D6": sets.New(NodeID.Compare, "D1", "D2"),
				"D9": sets.New(NodeID.Compare, "D1", "D2"),
			},
		},
		"empty": {
			Direct:   map[NodeID]*sets.Set[NodeID]{},
			Expected: map[NodeID]*sets.Set[NodeID]{},
		},

		"empty members": {
			Direct: map[NodeID]*sets.Set[NodeID]{
				"D3": sets.New(NodeID.Compare),
			},
			Expected: map[NodeID]*sets.Set[NodeID]{
				"D3": sets.New(NodeID.Compare),
			},
		},

		"member drawn as an atomic deliverable stays atomic": {
			Direct: map[NodeID]*sets.Set[NodeID]{
				"D5": sets.New(NodeID.Compare, "D1"),
				"D6": sets.New(NodeID.Compare, "D3", "D5"),
			},
			Nodes: sets.New(
				(*Node).Compare,
				&Node{ID: "D5", Type: NodeTypeAtomicDeliverable},
				&Node{ID: "D6", Type: NodeTypeCompositeDeliverable},
			),
			Expected: map[NodeID]*sets.Set[NodeID]{
				"D5": sets.New(NodeID.Compare, "D1"),
				"D6": sets.New(NodeID.Compare, "D3", "D5"),
			},
		},
		"self cycle": {
			Direct: map[NodeID]*sets.Set[NodeID]{
				"D3": sets.New(NodeID.Compare, "D3"),
			},
			WantErr: true,
		},
		"mutual cycle": {
			Direct: map[NodeID]*sets.Set[NodeID]{
				"D3": sets.New(NodeID.Compare, "D6"),
				"D6": sets.New(NodeID.Compare, "D3"),
			},
			WantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := FlattenDeliverableComposition(tc.Direct, tc.Nodes)
			if tc.WantErr {
				if err == nil {
					t.Fatal("expected an error for the cyclic composition, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got))
			}
		})
	}
}

func TestDrawnDeliverableComposition(t *testing.T) {
	testCases := map[string]struct {
		Composition map[NodeID]*sets.Set[NodeID]
		Nodes       *sets.Set[*Node]
		Expected    map[NodeID]*sets.Set[NodeID]
	}{
		"drawn as a composite deliverable": {
			Composition: map[NodeID]*sets.Set[NodeID]{
				"D3": sets.New(NodeID.Compare, "D1", "D2"),
			},
			Nodes: sets.New(
				(*Node).Compare,
				&Node{ID: "D3", Type: NodeTypeCompositeDeliverable},
			),
			Expected: map[NodeID]*sets.Set[NodeID]{
				"D3": sets.New(NodeID.Compare, "D1", "D2"),
			},
		},

		"member only": {
			Composition: map[NodeID]*sets.Set[NodeID]{
				"D3": sets.New(NodeID.Compare, "D1", "D2"),
				"D6": sets.New(NodeID.Compare, "D1", "D2", "D5"),
			},
			Nodes: sets.New(
				(*Node).Compare,
				&Node{ID: "D6", Type: NodeTypeCompositeDeliverable},
			),
			Expected: map[NodeID]*sets.Set[NodeID]{
				"D6": sets.New(NodeID.Compare, "D1", "D2", "D5"),
			},
		},

		"drawn as an atomic deliverable": {
			Composition: map[NodeID]*sets.Set[NodeID]{
				"D9": sets.New(NodeID.Compare, "D1"),
			},
			Nodes: sets.New(
				(*Node).Compare,
				&Node{ID: "D9", Type: NodeTypeAtomicDeliverable},
			),
			Expected: map[NodeID]*sets.Set[NodeID]{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := DrawnDeliverableComposition(tc.Composition, tc.Nodes)
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got))
			}
		})
	}
}
