package fsmreporter

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
)

func TestBuildTimelineTable(t *testing.T) {
	testCases := map[string]struct {
		Plan     *fsm.Plan
		Expected TimelineTable
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
			Expected: TimelineTable{},
		},
		"not continuous": {
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
						Allocation: fsm.Allocation{
							"P1": {Resources: sets.New(fsm.ResourceID.Compare, "R1")},
						},
						NextState: fsm.State{
							Time: 1.5,
							NumOfCompleteMap: map[pfd.AtomicProcessID]int{
								"P1": 2,
							},
						},
					},
				},
			},
			Expected: TimelineTable{
				{
					AtomicProcess:      "P1",
					NumOfComplete:      0,
					AllocatedResources: sets.New(fsm.ResourceID.Compare, "R1"),
					StartTime:          0,
					EndTime:            1,
				},
				{
					AtomicProcess:      "P1",
					NumOfComplete:      1,
					AllocatedResources: sets.New(fsm.ResourceID.Compare, "R1"),
					StartTime:          1,
					EndTime:            1.5,
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
								"P1": 0,
							},
						},
					},
					{
						Allocation: fsm.Allocation{
							"P1": {Resources: sets.New(fsm.ResourceID.Compare, "R1")},
						},
						NextState: fsm.State{
							Time: 1.5,
							NumOfCompleteMap: map[pfd.AtomicProcessID]int{
								"P1": 1,
							},
						},
					},
				},
			},
			Expected: TimelineTable{
				{
					AtomicProcess:      "P1",
					NumOfComplete:      0,
					AllocatedResources: sets.New(fsm.ResourceID.Compare, "R1"),
					StartTime:          0,
					EndTime:            1.5,
				},
			},
		},
		"double": {
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
								"P1": 1,
								"P2": 0,
							},
						},
					},
					{
						Allocation: fsm.Allocation{
							"P2": {Resources: sets.New(fsm.ResourceID.Compare, "R2")},
						},
						NextState: fsm.State{
							Time: 1.5,
							NumOfCompleteMap: map[pfd.AtomicProcessID]int{
								"P1": 1,
								"P2": 1,
							},
						},
					},
				},
			},
			Expected: TimelineTable{
				{
					AtomicProcess:      "P1",
					NumOfComplete:      0,
					AllocatedResources: sets.New(fsm.ResourceID.Compare, "R1"),
					StartTime:          0,
					EndTime:            1,
				},
				{
					AtomicProcess:      "P2",
					NumOfComplete:      0,
					AllocatedResources: sets.New(fsm.ResourceID.Compare, "R2"),
					StartTime:          1,
					EndTime:            1.5,
				},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			got := BuildTimelineTable(tc.Plan, logger)
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got))
			}
		})
	}
}
