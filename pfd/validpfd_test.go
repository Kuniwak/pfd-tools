package pfd

import (
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cmp2"
	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
)

func TestSafePFD(t *testing.T) {
	testCases := map[string]struct {
		PFD                    *PFD
		ProcessComposition     map[CompositeProcessID]*sets.Set[AtomicProcessID]
		DeliverableComposition map[CompositeDeliverableID]*sets.Set[AtomicDeliverableID]
		Expected               *ValidPFD
	}{
		"smallest": {
			PFD:                    PresetSmallest,
			ProcessComposition:     map[CompositeProcessID]*sets.Set[AtomicProcessID]{},
			DeliverableComposition: map[CompositeDeliverableID]*sets.Set[AtomicDeliverableID]{},
			Expected: NewSafePFD(
				map[AtomicProcessID]string{
					"P1": "P1",
				},
				map[AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
				},
				map[AtomicProcessID]*RelationTriple{
					"P1": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D1",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D2"),
					},
				},
				map[CompositeProcessID]*pairs.Pair[string, *sets.Set[AtomicProcessID]]{},
				map[CompositeDeliverableID]*pairs.Pair[string, *sets.Set[AtomicDeliverableID]]{},
			),
		},
		"sequential": {
			PFD:                    PresetSequential,
			ProcessComposition:     map[CompositeProcessID]*sets.Set[AtomicProcessID]{},
			DeliverableComposition: map[CompositeDeliverableID]*sets.Set[AtomicDeliverableID]{},
			Expected: NewSafePFD(
				map[AtomicProcessID]string{
					"P1": "P1",
					"P2": "P2",
				},
				map[AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
					"D3": "D3",
				},
				map[AtomicProcessID]*RelationTriple{
					"P1": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D1",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D2"),
					},
					"P2": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D2",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D3"),
					},
				},
				map[CompositeProcessID]*pairs.Pair[string, *sets.Set[AtomicProcessID]]{},
				map[CompositeDeliverableID]*pairs.Pair[string, *sets.Set[AtomicDeliverableID]]{},
			),
		},
		"counterclockwise_rotated_y_shape": {
			PFD:                    PresetCounterclockwiseRotatedYShape,
			ProcessComposition:     map[CompositeProcessID]*sets.Set[AtomicProcessID]{},
			DeliverableComposition: map[CompositeDeliverableID]*sets.Set[AtomicDeliverableID]{},
			Expected: NewSafePFD(
				map[AtomicProcessID]string{
					"P1": "P1",
				},
				map[AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
					"D3": "D3",
				},
				map[AtomicProcessID]*RelationTriple{
					"P1": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D1",
							"D2",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D3"),
					},
				},
				map[CompositeProcessID]*pairs.Pair[string, *sets.Set[AtomicProcessID]]{},
				map[CompositeDeliverableID]*pairs.Pair[string, *sets.Set[AtomicDeliverableID]]{},
			),
		},
		"clockwise_rotated_y_shape": {
			PFD:                    PresetClockwiseRotatedYShape,
			ProcessComposition:     map[CompositeProcessID]*sets.Set[AtomicProcessID]{},
			DeliverableComposition: map[CompositeDeliverableID]*sets.Set[AtomicDeliverableID]{},
			Expected: NewSafePFD(
				map[AtomicProcessID]string{
					"P1": "P1",
				},
				map[AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
					"D3": "D3",
				},
				map[AtomicProcessID]*RelationTriple{
					"P1": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D1",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D2", "D3"),
					},
				},
				map[CompositeProcessID]*pairs.Pair[string, *sets.Set[AtomicProcessID]]{},
				map[CompositeDeliverableID]*pairs.Pair[string, *sets.Set[AtomicDeliverableID]]{},
			),
		},
		"bigger_counterclockwise_rotated_y_shape": {
			PFD:                    PresetBiggerCounterclockwiseRotatedYShape,
			ProcessComposition:     map[CompositeProcessID]*sets.Set[AtomicProcessID]{},
			DeliverableComposition: map[CompositeDeliverableID]*sets.Set[AtomicDeliverableID]{},
			Expected: NewSafePFD(
				map[AtomicProcessID]string{
					"P1": "P1",
					"P2": "P2",
					"P3": "P3",
				},
				map[AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
					"D3": "D3",
					"D4": "D4",
					"D5": "D5",
				},
				map[AtomicProcessID]*RelationTriple{
					"P1": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D1",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D3"),
					},
					"P2": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D2",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D4"),
					},
					"P3": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D3",
							"D4",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D5"),
					},
				},
				map[CompositeProcessID]*pairs.Pair[string, *sets.Set[AtomicProcessID]]{},
				map[CompositeDeliverableID]*pairs.Pair[string, *sets.Set[AtomicDeliverableID]]{},
			),
		},
		"bigger_clockwise_rotated_y_shape": {
			PFD:                    PresetBiggerClockwiseRotatedYShape,
			ProcessComposition:     map[CompositeProcessID]*sets.Set[AtomicProcessID]{},
			DeliverableComposition: map[CompositeDeliverableID]*sets.Set[AtomicDeliverableID]{},
			Expected: NewSafePFD(
				map[AtomicProcessID]string{
					"P1": "P1",
					"P2": "P2",
					"P3": "P3",
				},
				map[AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
					"D3": "D3",
					"D4": "D4",
				},
				map[AtomicProcessID]*RelationTriple{
					"P1": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D1",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D2"),
					},
					"P2": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D2",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D3"),
					},
					"P3": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D2",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D4"),
					},
				},
				map[CompositeProcessID]*pairs.Pair[string, *sets.Set[AtomicProcessID]]{},
				map[CompositeDeliverableID]*pairs.Pair[string, *sets.Set[AtomicDeliverableID]]{},
			),
		},
		"smallest_loop": {
			PFD:                    PresetSmallestLoop,
			ProcessComposition:     map[CompositeProcessID]*sets.Set[AtomicProcessID]{},
			DeliverableComposition: map[CompositeDeliverableID]*sets.Set[AtomicDeliverableID]{},
			Expected: NewSafePFD(
				map[AtomicProcessID]string{
					"P1": "P1",
				},
				map[AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
				},
				map[AtomicProcessID]*RelationTriple{
					"P1": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D1",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare, "D2"),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D2"),
					},
				},
				map[CompositeProcessID]*pairs.Pair[string, *sets.Set[AtomicProcessID]]{},
				map[CompositeDeliverableID]*pairs.Pair[string, *sets.Set[AtomicDeliverableID]]{},
			),
		},
		"nested_loop": {
			PFD:                    PresetNestedLoop,
			ProcessComposition:     map[CompositeProcessID]*sets.Set[AtomicProcessID]{},
			DeliverableComposition: map[CompositeDeliverableID]*sets.Set[AtomicDeliverableID]{},
			Expected: NewSafePFD(
				map[AtomicProcessID]string{
					"P1": "P1",
					"P2": "P2",
					"P3": "P3",
				},
				map[AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
					"D3": "D3",
					"D4": "D4",
				},
				map[AtomicProcessID]*RelationTriple{
					"P1": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D1",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare, "D4"),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D2"),
					},
					"P2": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D2",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare, "D3"),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D3"),
					},
					"P3": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D3",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D4"),
					},
				},
				map[CompositeProcessID]*pairs.Pair[string, *sets.Set[AtomicProcessID]]{},
				map[CompositeDeliverableID]*pairs.Pair[string, *sets.Set[AtomicDeliverableID]]{},
			),
		},
		"butterfly_loop": {
			PFD:                    PresetButterflyLoop,
			ProcessComposition:     map[CompositeProcessID]*sets.Set[AtomicProcessID]{},
			DeliverableComposition: map[CompositeDeliverableID]*sets.Set[AtomicDeliverableID]{},
			Expected: NewSafePFD(
				map[AtomicProcessID]string{
					"P1": "P1",
					"P2": "P2",
					"P3": "P3",
				},
				map[AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
					"D3": "D3",
					"D4": "D4",
					"D5": "D5",
				},
				map[AtomicProcessID]*RelationTriple{
					"P1": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D1",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare, "D4", "D5"),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D2", "D3"),
					},
					"P2": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D2",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D4"),
					},
					"P3": {
						Inputs: sets.New(
							AtomicDeliverableID.Compare,
							"D3",
						),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D5"),
					},
				},
				map[CompositeProcessID]*pairs.Pair[string, *sets.Set[AtomicProcessID]]{},
				map[CompositeDeliverableID]*pairs.Pair[string, *sets.Set[AtomicDeliverableID]]{},
			),
		},
		"cross_loop": {
			PFD:                    PresetCrossLoop,
			ProcessComposition:     map[CompositeProcessID]*sets.Set[AtomicProcessID]{},
			DeliverableComposition: map[CompositeDeliverableID]*sets.Set[AtomicDeliverableID]{},
			Expected: NewSafePFD(
				map[AtomicProcessID]string{
					"P1": "P1",
					"P2": "P2",
					"P3": "P3",
				},
				map[AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
					"D3": "D3",
					"D4": "D4",
				},
				map[AtomicProcessID]*RelationTriple{
					"P1": {
						Inputs:         sets.New(AtomicDeliverableID.Compare, "D1"),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare, "D3"),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D2"),
					},
					"P2": {
						Inputs:         sets.New(AtomicDeliverableID.Compare, "D2"),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare, "D4"),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D3"),
					},
					"P3": {
						Inputs:         sets.New(AtomicDeliverableID.Compare, "D3"),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D4"),
					},
				},
				map[CompositeProcessID]*pairs.Pair[string, *sets.Set[AtomicProcessID]]{},
				map[CompositeDeliverableID]*pairs.Pair[string, *sets.Set[AtomicDeliverableID]]{},
			),
		},
		"wait_loop_end": {
			PFD:                    PresetWaitLoopEnd,
			ProcessComposition:     map[CompositeProcessID]*sets.Set[AtomicProcessID]{},
			DeliverableComposition: map[CompositeDeliverableID]*sets.Set[AtomicDeliverableID]{},
			Expected: NewSafePFD(
				map[AtomicProcessID]string{
					"P1": "P1",
					"P2": "P2",
				},
				map[AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
					"D3": "D3",
				},
				map[AtomicProcessID]*RelationTriple{
					"P1": {
						Inputs:         sets.New(AtomicDeliverableID.Compare, "D1"),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare, "D2"),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D2"),
					},
					"P2": {
						Inputs:         sets.New(AtomicDeliverableID.Compare, "D2"),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D3"),
					},
				},
				map[CompositeProcessID]*pairs.Pair[string, *sets.Set[AtomicProcessID]]{},
				map[CompositeDeliverableID]*pairs.Pair[string, *sets.Set[AtomicDeliverableID]]{},
			),
		},
		"composition": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "D0", Description: "D0", Type: NodeTypeCompositeDeliverable},
					&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "D2", Description: "D2", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "D3", Description: "D3", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "P0", Description: "P0", Type: NodeTypeCompositeProcess},
					&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
					&Node{ID: "P2", Description: "P2", Type: NodeTypeAtomicProcess},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "D1", Target: "P1"},
					&Edge{Source: "P1", Target: "D2"},
					&Edge{Source: "D2", Target: "P2"},
					&Edge{Source: "P2", Target: "D3"},
					&Edge{Source: "P0", Target: "D1"},
					&Edge{Source: "P0", Target: "D3"},
					&Edge{Source: "D0", Target: "P1", IsFeedback: true},
					&Edge{Source: "D0", Target: "P0", IsFeedback: true},
				),
				ProcessComposition: map[NodeID]*sets.Set[NodeID]{
					"P0": sets.New(NodeID.Compare, "P1", "P2"),
				},
				DeliverableComposition: map[NodeID]*sets.Set[NodeID]{
					"D0": sets.New(NodeID.Compare, "D2", "D3"),
				},
			},
			Expected: NewSafePFD(
				map[AtomicProcessID]string{
					"P1": "P1",
					"P2": "P2",
				},
				map[AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
					"D3": "D3",
				},
				map[AtomicProcessID]*RelationTriple{
					"P1": {
						Inputs:         sets.New(AtomicDeliverableID.Compare, "D1"),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare, "D2", "D3"),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D2"),
					},
					"P2": {
						Inputs:         sets.New(AtomicDeliverableID.Compare, "D2"),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D3"),
					},
				},
				map[CompositeProcessID]*pairs.Pair[string, *sets.Set[AtomicProcessID]]{
					"P0": pairs.New("P0", sets.New(AtomicProcessID.Compare, "P1", "P2")),
				},
				map[CompositeDeliverableID]*pairs.Pair[string, *sets.Set[AtomicDeliverableID]]{
					"D0": pairs.New("D0", sets.New(AtomicDeliverableID.Compare, "D2", "D3")),
				},
			),
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual, err := NewSafePFDByUnsafePFD(testCase.PFD)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if !reflect.DeepEqual(actual, testCase.Expected) {
				t.Error(cmp.Diff(testCase.Expected, actual))
			}
		})
	}
}

func TestSafePFDCompositeDeliverableExpansion(t *testing.T) {
	nodes := []*Node{
		{ID: "D0", Description: "D0", Type: NodeTypeAtomicDeliverable},
		{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
		{ID: "D2", Description: "D2", Type: NodeTypeAtomicDeliverable},
		{ID: "D3", Description: "D3", Type: NodeTypeCompositeDeliverable},
		{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
	}

	testCases := map[string]struct {
		Edges           []*Edge
		ExpectedInputs  []AtomicDeliverableID
		ExpectedOutputs []AtomicDeliverableID
	}{
		"output side: atomic process -> composite deliverable": {
			Edges: []*Edge{
				{Source: "D0", Target: "P1"},
				{Source: "P1", Target: "D3"},
			},
			ExpectedInputs:  []AtomicDeliverableID{"D0"},
			ExpectedOutputs: []AtomicDeliverableID{"D1", "D2"},
		},
		"input side: composite deliverable -> atomic process": {
			Edges: []*Edge{
				{Source: "D3", Target: "P1"},
				{Source: "P1", Target: "D0"},
			},
			ExpectedInputs:  []AtomicDeliverableID{"D1", "D2"},
			ExpectedOutputs: []AtomicDeliverableID{"D0"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			p := &PFD{
				Nodes:              sets.New((*Node).Compare, nodes...),
				Edges:              sets.New((*Edge).Compare, tc.Edges...),
				ProcessComposition: map[NodeID]*sets.Set[NodeID]{},
				DeliverableComposition: map[NodeID]*sets.Set[NodeID]{
					"D3": sets.New(NodeID.Compare, "D1", "D2"),
				},
			}

			v, err := NewSafePFDByUnsafePFD(p)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(v.Relations["P1"].Inputs.Slice(), tc.ExpectedInputs) {
				t.Errorf("inputs = %v, want %v", v.Relations["P1"].Inputs.Slice(), tc.ExpectedInputs)
			}
			if !reflect.DeepEqual(v.Relations["P1"].Outputs.Slice(), tc.ExpectedOutputs) {
				t.Errorf("outputs = %v, want %v", v.Relations["P1"].Outputs.Slice(), tc.ExpectedOutputs)
			}
			for _, d := range tc.ExpectedOutputs {
				if ap := v.Memoized.SourceAtomicProcess[d]; ap != "P1" {
					t.Errorf("SourceAtomicProcess[%q] = %q, want %q", d, ap, "P1")
				}
			}
		})
	}
}

func TestSafePFDError(t *testing.T) {
	testCases := map[string]struct {
		PFD      *PFD
		Contains string
	}{

		"composite deliverable without a composite deliverable table entry": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "D0", Description: "D0", Type: NodeTypeCompositeDeliverable},
					&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "D0", Target: "P1"},
					&Edge{Source: "P1", Target: "D1"},
				),
				ProcessComposition:     map[NodeID]*sets.Set[NodeID]{},
				DeliverableComposition: map[NodeID]*sets.Set[NodeID]{},
			},
			Contains: "composite deliverable",
		},

		"feedback edge from an atomic process to an atomic deliverable": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "P1", Target: "D1", IsFeedback: true},
				),
				ProcessComposition:     map[NodeID]*sets.Set[NodeID]{},
				DeliverableComposition: map[NodeID]*sets.Set[NodeID]{},
			},
			Contains: `atomic process: "P1" is connected to deliverable: "D1" with feedback edge`,
		},

		"feedback edge from an atomic process to a composite deliverable": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "D3", Description: "D3", Type: NodeTypeCompositeDeliverable},
					&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "P1", Target: "D3", IsFeedback: true},
				),
				ProcessComposition: map[NodeID]*sets.Set[NodeID]{},
				DeliverableComposition: map[NodeID]*sets.Set[NodeID]{
					"D3": sets.New(NodeID.Compare, "D1"),
				},
			},
			Contains: `atomic process: "P1" is connected to deliverable: "D3" with feedback edge`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := NewSafePFDByUnsafePFD(tc.PFD)
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			if !strings.Contains(err.Error(), tc.Contains) {
				t.Errorf("error should contain %q, got: %v", tc.Contains, err)
			}
		})
	}
}

func TestCollectPaths(t *testing.T) {
	testCases := map[string]struct {
		PFD      *PFD
		Src      AtomicProcessID
		Dst      AtomicProcessID
		Expected *sets.Set[[]AtomicProcessID]
	}{
		"example": {
			PFD: &PFD{

				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
					&Node{ID: "P2", Description: "P2", Type: NodeTypeAtomicProcess},
					&Node{ID: "P3", Description: "P3", Type: NodeTypeAtomicProcess},
					&Node{ID: "P4", Description: "P4", Type: NodeTypeAtomicProcess},
					&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "D2", Description: "D2", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "D3", Description: "D3", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "D4", Description: "D4", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "D5", Description: "D5", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "D6", Description: "D6", Type: NodeTypeAtomicDeliverable},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "D1", Target: "P1"},
					&Edge{Source: "D2", Target: "P2"},
					&Edge{Source: "D2", Target: "P4"},
					&Edge{Source: "D3", Target: "P3"},
					&Edge{Source: "P1", Target: "D2"},
					&Edge{Source: "P1", Target: "D5"},
					&Edge{Source: "D5", Target: "P1", IsFeedback: true},
					&Edge{Source: "P2", Target: "D3"},
					&Edge{Source: "P3", Target: "D4"},
					&Edge{Source: "P4", Target: "D6"},
					&Edge{Source: "D6", Target: "P3"},
				),
			},
			Src: "P1",
			Dst: "P3",
			Expected: sets.New(
				cmp2.CompareSlice[[]AtomicProcessID](AtomicProcessID.Compare),
				[]AtomicProcessID{"P1", "P2", "P3"},
				[]AtomicProcessID{"P1", "P4", "P3"},
			),
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := sets.New(cmp2.CompareSlice[[]AtomicProcessID](AtomicProcessID.Compare))
			p, err := NewSafePFDByUnsafePFD(testCase.PFD)
			if err != nil {
				t.Fatalf("error: %v", err)
			}

			p.CollectPaths(testCase.Src, testCase.Dst, actual, slog.New(slogtest.NewTestHandler(t)))
			if !reflect.DeepEqual(actual, testCase.Expected) {
				t.Error(cmp.Diff(testCase.Expected, actual))
			}
		})
	}
}

func TestCollectReachableDeliverables(t *testing.T) {
	testCases := map[string]struct {
		PFD      *ValidPFD
		Src      AtomicProcessID
		Expected *sets.Set[AtomicDeliverableID]
	}{
		"example": {

			PFD: NewSafePFD(
				map[AtomicProcessID]string{
					"P1": "P1",
					"P2": "P2",
					"P3": "P3",
					"P4": "P4",
				},
				map[AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
					"D3": "D3",
					"D4": "D4",
					"D5": "D5",
				},
				map[AtomicProcessID]*RelationTriple{
					"P1": {
						Inputs:         sets.New(AtomicDeliverableID.Compare, "D1"),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare, "D2"),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D2"),
					},
					"P2": {
						Inputs:         sets.New(AtomicDeliverableID.Compare, "D2"),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare, "D5"),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D3"),
					},
					"P3": {
						Inputs:         sets.New(AtomicDeliverableID.Compare, "D3"),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare, "D4"),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D4"),
					},
					"P4": {
						Inputs:         sets.New(AtomicDeliverableID.Compare, "D4"),
						FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
						Outputs:        sets.New(AtomicDeliverableID.Compare, "D5"),
					},
				},
				map[CompositeProcessID]*pairs.Pair[string, *sets.Set[AtomicProcessID]]{},
				map[CompositeDeliverableID]*pairs.Pair[string, *sets.Set[AtomicDeliverableID]]{},
			),
			Src: "P3",
			Expected: sets.New(
				AtomicDeliverableID.Compare,
				"D4",
				"D5",
			),
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := sets.New(AtomicDeliverableID.Compare)
			testCase.PFD.CollectReachableDeliverablesExceptFeedback(testCase.Src, actual, slog.New(slogtest.NewTestHandler(t)))
			if !reflect.DeepEqual(actual, testCase.Expected) {
				t.Error(cmp.Diff(testCase.Expected, actual))
			}
		})
	}
}

func TestFeedbackEdges(t *testing.T) {
	testCases := map[string]struct {
		PFD      *PFD
		Expected []FeedbackEdge
	}{
		"no feedback edge": {

			PFD:      PresetSmallest,
			Expected: []FeedbackEdge{},
		},
		"a single feedback edge": {

			PFD:      PresetSmallestLoop,
			Expected: []FeedbackEdge{{AtomicDeliverable: "D2", AtomicProcess: "P1"}},
		},
		"feedback edges are sorted by deliverable then atomic process": {

			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "D1", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "D2", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "D3", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "P1", Type: NodeTypeAtomicProcess},
					&Node{ID: "P2", Type: NodeTypeAtomicProcess},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "D1", Target: "P1"},
					&Edge{Source: "P1", Target: "D2"},
					&Edge{Source: "D2", Target: "P2"},
					&Edge{Source: "P2", Target: "D3"},
					&Edge{Source: "D3", Target: "P1", IsFeedback: true},
					&Edge{Source: "D2", Target: "P1", IsFeedback: true},
				),
			},
			Expected: []FeedbackEdge{
				{AtomicDeliverable: "D2", AtomicProcess: "P1"},
				{AtomicDeliverable: "D3", AtomicProcess: "P1"},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := MustNewSafePFDByUnsafePFD(tc.PFD).FeedbackEdges()
			if !reflect.DeepEqual(actual, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, actual))
			}
		})
	}
}
