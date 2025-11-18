package fsmmasterschedule

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/graph"
	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/google/go-cmp/cmp"
)

func TestNewMilestoneTimelineFromPlan(t *testing.T) {
	testCases := map[string]struct {
		Plan           *fsm.Plan
		MilestoneMap   map[pfd.AtomicProcessID]Milestone
		GroupsMap      map[pfd.AtomicProcessID]*sets.Set[Group]
		MilestoneGraph map[Group]*graph.Graph
		Expected       *Timeline
	}{
		"empty": {
			Plan: &fsm.Plan{
				InitialState: fsm.State{
					Time: 0,
					NumOfCompleteMap: map[pfd.AtomicProcessID]int{
						"P1": 0,
					},
				},
				Transitions: []*fsm.Trans{},
			},
			MilestoneMap: map[pfd.AtomicProcessID]Milestone{
				"P1": "M1",
			},
			GroupsMap: map[pfd.AtomicProcessID]*sets.Set[Group]{
				"P1": sets.New(Group.Compare, "G1"),
			},
			MilestoneGraph: map[Group]*graph.Graph{
				"G1": {
					Nodes: sets.New(graph.Node.Compare, "M1"),
					Edges: sets.New(pairs.Compare(graph.Node.Compare, graph.Node.Compare)),
				},
			},
			Expected: &Timeline{
				"G1": NewMilestoneTimeline(sets.New(Milestone.Compare, "M1")),
			},
		},
		"begin at 0 and end at last and neither": {
			Plan: &fsm.Plan{
				InitialState: fsm.State{
					Time: 0,
					NumOfCompleteMap: map[pfd.AtomicProcessID]int{
						"P1": 0,
						"P2": 0,
						"P3": 0,
					},
				},
				Transitions: []*fsm.Trans{
					{
						Allocation: fsm.Allocation{
							"P1": {Resources: sets.New(fsm.ResourceID.Compare, "R1")},
						},
						NextState: fsm.State{
							Time: 1,
							NumOfCompleteMap: map[pfd.AtomicProcessID]int{
								"P1": 1,
								"P2": 0,
								"P3": 0,
							},
						},
					},
					{
						Allocation: fsm.Allocation{
							"P2": {Resources: sets.New(fsm.ResourceID.Compare, "R1")},
						},
						NextState: fsm.State{
							Time: 2,
							NumOfCompleteMap: map[pfd.AtomicProcessID]int{
								"P1": 1,
								"P2": 1,
								"P3": 0,
							},
						},
					},
					{
						Allocation: fsm.Allocation{
							"P3": {Resources: sets.New(fsm.ResourceID.Compare, "R1")},
						},
						NextState: fsm.State{
							Time: 3,
							NumOfCompleteMap: map[pfd.AtomicProcessID]int{
								"P1": 1,
								"P2": 1,
								"P3": 1,
							},
						},
					},
				},
			},
			MilestoneMap: map[pfd.AtomicProcessID]Milestone{
				"P1": "M1",
				"P2": "M2",
				"P3": "M3",
			},
			GroupsMap: map[pfd.AtomicProcessID]*sets.Set[Group]{
				"P1": sets.New(Group.Compare, "G1"),
				"P2": sets.New(Group.Compare, "G1"),
				"P3": sets.New(Group.Compare, "G1"),
			},
			MilestoneGraph: map[Group]*graph.Graph{
				"G1": {
					Nodes: sets.New(graph.Node.Compare, "M1", "M2", "M3"),
					Edges: sets.New(
						pairs.Compare(graph.Node.Compare, graph.Node.Compare),
						pairs.New[graph.Node, graph.Node]("M1", "M2"),
						pairs.New[graph.Node, graph.Node]("M2", "M3"),
					),
				},
			},
			Expected: &Timeline{
				"G1": {
					"M1": {
						StartTime: 0,
						EndTime:   1,
					},
					"M2": {
						StartTime: 1,
						EndTime:   2,
					},
					"M3": {
						StartTime: 2,
						EndTime:   3,
					},
				},
			},
		},
		"continuous": {
			Plan: &fsm.Plan{
				InitialState: fsm.State{
					Time: 0,
					NumOfCompleteMap: map[pfd.AtomicProcessID]int{
						"P1": 0,
					},
				},
				Transitions: []*fsm.Trans{
					{
						Allocation: fsm.Allocation{
							"P1": {Resources: sets.New(fsm.ResourceID.Compare, "R1")},
						},
						NextState: fsm.State{
							Time: 1,
							NumOfCompleteMap: map[pfd.AtomicProcessID]int{
								"P1": 1,
							},
						},
					},
					{
						Allocation: fsm.Allocation{},
						NextState: fsm.State{
							Time: 2,
							NumOfCompleteMap: map[pfd.AtomicProcessID]int{
								"P1": 1,
							},
						},
					},
					{
						Allocation: fsm.Allocation{
							"P1": {Resources: sets.New(fsm.ResourceID.Compare, "R1")},
						},
						NextState: fsm.State{
							Time: 3,
							NumOfCompleteMap: map[pfd.AtomicProcessID]int{
								"P1": 2,
							},
						},
					},
				},
			},
			MilestoneMap: map[pfd.AtomicProcessID]Milestone{
				"P1": "M1",
			},
			GroupsMap: map[pfd.AtomicProcessID]*sets.Set[Group]{
				"P1": sets.New(Group.Compare, "G1"),
			},
			MilestoneGraph: map[Group]*graph.Graph{
				"G1": {
					Nodes: sets.New(graph.Node.Compare, "M1"),
					Edges: sets.New(
						pairs.Compare(graph.Node.Compare, graph.Node.Compare),
					),
				},
			},
			Expected: &Timeline{
				"G1": &MilestoneTimeline{
					"M1": &TimelineItem{
						StartTime: 0,
						EndTime:   3,
					},
				},
			},
		},
		"overlap": {
			Plan: &fsm.Plan{
				InitialState: fsm.State{
					Time: 0,
					NumOfCompleteMap: map[pfd.AtomicProcessID]int{
						"P1": 0,
						"P2": 0,
					},
				},
				Transitions: []*fsm.Trans{
					{
						Allocation: fsm.Allocation{
							"P1": {Resources: sets.New(fsm.ResourceID.Compare, "R1")},
						},
						NextState: fsm.State{
							Time: 1,
							NumOfCompleteMap: map[pfd.AtomicProcessID]int{
								"P1": 0,
								"P2": 0,
							},
						},
					},
					{
						Allocation: fsm.Allocation{
							"P1": {Resources: sets.New(fsm.ResourceID.Compare, "R1")},
							"P2": {Resources: sets.New(fsm.ResourceID.Compare, "R2")},
						},
						NextState: fsm.State{
							Time: 2,
							NumOfCompleteMap: map[pfd.AtomicProcessID]int{
								"P1": 0,
								"P2": 1,
							},
						},
					},
					{
						Allocation: fsm.Allocation{
							"P1": {Resources: sets.New(fsm.ResourceID.Compare, "R1")},
						},
						NextState: fsm.State{
							Time: 3,
							NumOfCompleteMap: map[pfd.AtomicProcessID]int{
								"P1": 1,
								"P2": 1,
							},
						},
					},
				},
			},
			MilestoneMap: map[pfd.AtomicProcessID]Milestone{
				"P1": "M1",
				"P2": "M2",
			},
			GroupsMap: map[pfd.AtomicProcessID]*sets.Set[Group]{
				"P1": sets.New(Group.Compare, "G1"),
				"P2": sets.New(Group.Compare, "G1"),
			},
			MilestoneGraph: map[Group]*graph.Graph{
				"G1": {
					Nodes: sets.New(graph.Node.Compare, "M1", "M2"),
					Edges: sets.New(
						pairs.Compare(graph.Node.Compare, graph.Node.Compare),
						pairs.New[graph.Node, graph.Node]("M1", "M2"),
					),
				},
			},
			Expected: &Timeline{
				"G1": &MilestoneTimeline{
					"M1": &TimelineItem{
						StartTime: 0,
						EndTime:   1,
					},
					"M2": &TimelineItem{
						StartTime: 1,
						EndTime:   2,
					},
				},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := NewTimelineFromPlan(
				tc.Plan,
				collectAtomicProcesses(tc.MilestoneMap, tc.GroupsMap),
				collectGroups(tc.GroupsMap),
				tc.GroupsMap,
				tc.MilestoneMap,
				tc.MilestoneGraph,
			)
			if err != nil {
				t.Fatalf("NewTimelineFromPlan: %v", err)
			}

			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got))
			}
		})
	}
}

func collectAtomicProcesses(mm map[pfd.AtomicProcessID]Milestone, gsm map[pfd.AtomicProcessID]*sets.Set[Group]) *sets.Set[pfd.AtomicProcessID] {
	s1 := sets.New(pfd.AtomicProcessID.Compare)
	for ap := range mm {
		s1.Add(pfd.AtomicProcessID.Compare, ap)
	}

	s2 := sets.New(pfd.AtomicProcessID.Compare)
	for ap := range gsm {
		s2.Add(pfd.AtomicProcessID.Compare, ap)
	}

	if !sets.IsEqual(pfd.AtomicProcessID.Compare, s1, s2) {
		panic(fmt.Sprintf("collectAtomicProcesses: atomic processes are not equal: %v, %v", s1, s2))
	}

	return s1
}

func collectGroups(gsm map[pfd.AtomicProcessID]*sets.Set[Group]) *sets.Set[Group] {
	s := sets.New(Group.Compare)
	for _, gs := range gsm {
		s.Union(Group.Compare, gs)
	}
	return s
}
