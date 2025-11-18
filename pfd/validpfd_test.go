package pfd

import (
	"log/slog"
	"reflect"
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

func TestCollectPaths(t *testing.T) {
	testCases := map[string]struct {
		PFD      *PFD
		Src      AtomicProcessID
		Dst      AtomicProcessID
		Expected *sets.Set[[]AtomicProcessID]
	}{
		"example": {
			PFD: &PFD{
				//            - - -
				//           V     \
				// [D1] -> (P1) -> [D5]
				//           \
				//            +->  [D2] ->  (P2) -> [D3] -> (P3) -> [D4]
				//                    \                      ^
				//                     +->  (P4) -> [D6] ---+
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
			//                                    - - - - - - - - - - - - - - - - - - - - -
			//               - - -              /                    - - -                  \
			//             V       \           V                   V       \                 \
			// [D1] ---> (P1) ---> [D2] ---> (P2) ---> [D3] ---> (P3) ---> [D4] --> (P4) --> [D5]
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
