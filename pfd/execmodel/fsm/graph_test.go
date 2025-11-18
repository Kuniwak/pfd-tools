package fsm

import (
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/google/go-cmp/cmp"
)

func TestFindDeadlocks(t *testing.T) {
	testCases := map[string]struct {
		Graph       *StateTransitionGraph
		IsCompleted func(State) bool
		Expected    *sets.Set[Deadlock]
	}{
		"smallest no deadlock": {
			Graph: &StateTransitionGraph{
				InitialState: 0,
				Nodes: map[StateID]State{
					0: {Time: 0},
				},
				Edges: map[StateID]map[StateID]*sets.Set[Allocation]{},
			},
			IsCompleted: func(s State) bool {
				return true
			},
			Expected: sets.New(Deadlock.Compare),
		},
		"smallest deadlock": {
			Graph: &StateTransitionGraph{
				InitialState: 0,
				Nodes: map[StateID]State{
					0: {Time: 0},
				},
				Edges: map[StateID]map[StateID]*sets.Set[Allocation]{},
			},
			IsCompleted: func(s State) bool {
				return false
			},
			Expected: sets.New(Deadlock.Compare,
				Deadlock{
					StateID: 0,
					Path:    []*pairs.Pair[StateID, Allocation]{},
				},
			),
		},
		"no deadlock": {
			Graph: &StateTransitionGraph{
				InitialState: 0,
				Nodes: map[StateID]State{
					0: {Time: 0},
					1: {Time: 1},
					2: {Time: 2},
				},
				Edges: map[StateID]map[StateID]*sets.Set[Allocation]{
					0: {
						1: sets.New(Allocation.Compare, Allocation{}),
					},
					1: {
						2: sets.New(Allocation.Compare, Allocation{}),
					},
				},
			},
			IsCompleted: func(s State) bool {
				return s.Time == 2
			},
			Expected: sets.New(Deadlock.Compare),
		},
		"deadlocks": {
			//    (P1)-> 1 -(P2)
			//    /             \
			//   /               V
			// 0                  3
			// | \               ^
			// |  \             /
			// |  (P3)-> 2 -(P4)
			// |
			// +--(P5)-> 4 -(P6)-> 5
			Graph: &StateTransitionGraph{
				InitialState: 0,
				Nodes: map[StateID]State{
					0: {Time: 0},
					1: {Time: 1},
					2: {Time: 2},
					3: {Time: 3},
					4: {Time: 4},
					5: {Time: 5},
				},
				Edges: map[StateID]map[StateID]*sets.Set[Allocation]{
					0: {
						1: sets.New(Allocation.Compare, Allocation{"P1": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1}}),
						2: sets.New(Allocation.Compare, Allocation{"P3": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1}}),
						4: sets.New(Allocation.Compare, Allocation{"P5": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1}}),
					},
					1: {
						3: sets.New(Allocation.Compare, Allocation{"P2": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1}}),
					},
					2: {
						3: sets.New(Allocation.Compare, Allocation{"P4": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1}}),
					},
					4: {
						5: sets.New(Allocation.Compare, Allocation{"P6": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1}}),
					},
				},
			},
			IsCompleted: func(s State) bool {
				return false
			},
			Expected: sets.New(Deadlock.Compare,
				Deadlock{
					StateID: 3,
					Path: []*pairs.Pair[StateID, Allocation]{
						pairs.New(StateID(0), Allocation{
							"P1": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
						}),
						pairs.New(StateID(1), Allocation{
							"P2": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
						}),
					},
				},
				Deadlock{
					StateID: 3,
					Path: []*pairs.Pair[StateID, Allocation]{
						pairs.New(StateID(0), Allocation{
							"P3": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
						}),
						pairs.New(StateID(2), Allocation{
							"P4": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
						}),
					},
				},
				Deadlock{
					StateID: 5,
					Path: []*pairs.Pair[StateID, Allocation]{
						pairs.New(StateID(0), Allocation{
							"P5": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
						}),
						pairs.New(StateID(4), Allocation{
							"P6": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
						}),
					},
				},
			),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ch := make(chan Deadlock)
			go func() {
				defer close(ch)
				testCase.Graph.FindDeadlocks(testCase.IsCompleted, ch)
			}()

			got := sets.New(Deadlock.Compare)
			for deadlock := range ch {
				got.Add(Deadlock.Compare, deadlock)
			}

			if !reflect.DeepEqual(got, testCase.Expected) {
				t.Error(cmp.Diff(testCase.Expected, got))
			}
		})
	}
}
