package fsm

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
)

func TestPrecondition_Eval(t *testing.T) {
	branch, err := pfd.NewSafePFDByUnsafePFD(pfd.PresetTripleBranch)
	if err != nil {
		t.Fatalf("pfd.NewSafePFDByUnsafePFD: %v", err)
	}

	neededResourceSetsFunc := NeededResourceSetsFuncByMap(map[pfd.AtomicProcessID]*sets.Set[AllocationElement]{
		"P1": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1}),
		"P2": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1}),
		"P3": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1}),
	})

	testCases := map[string]struct {
		Env   *Env
		State State
		Want  map[pfd.AtomicProcessID]*PreconditionEvalResult
	}{
		"and condition (true)": {
			Env: NewEnv(
				// [D1] -> (P1) -> [D2]
				//  \ \
				//   \ +-> (P2) -> [D3]
				//    \
				//     +-> (P3) -> [D4]
				branch,
				sets.New(ResourceID.Compare, "R1"),
				NewAvailableAllocationsFunc(neededResourceSetsFunc),
				ConstInitialVolumeFunc(1),
				ExponentialReworkVolumeFunc(0.5, ConstInitialVolumeFunc(1)),
				ConstMaxRevisionMap(3, branch.FeedbackSourceDeliverables()),
				map[pfd.AtomicProcessID]*Precondition{
					"P1": NewTruePrecondition(),
					"P2": NewTruePrecondition(),
					"P3": NewAndPrecondition(
						NewNotPrecondition(NewExecutablePrecondition("P1")),
						NewNotPrecondition(NewExecutablePrecondition("P2")),
					),
				},
				neededResourceSetsFunc,
				ConstDeliverableAvailableTimeFunc(0),
				slog.New(slogtest.NewTestHandler(t)),
			),
			State: NewState(
				0,
				map[pfd.AtomicDeliverableID]int{
					"D1": 1, // Initial deliverable
					"D2": 1, // Not created yet
					"D3": 1, // Not created yet
					"D4": 0, // Not created yet
				},
				map[pfd.AtomicProcessID]Volume{
					"P1": 0.5, // Any greater than or equal to MinimumVolume
					"P2": 0.5, // Any greater than or equal to MinimumVolume
					"P3": 1,   // Any less than or equal to initial volume
				},
				map[pfd.AtomicProcessID]int{
					"P1": 1, // Any
					"P2": 1, // Any
					"P3": 0, // Any
				},
				Allocation{},
				map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]{
					"P1": sets.New(pfd.AtomicDeliverableID.Compare),
					"P2": sets.New(pfd.AtomicDeliverableID.Compare),
					"P3": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
				},
			),
			Want: map[pfd.AtomicProcessID]*PreconditionEvalResult{
				"P1": {
					Type:   PreconditionTypeTrue,
					Result: true,
				},
				"P2": {
					Type:   PreconditionTypeTrue,
					Result: true,
				},
				"P3": {
					Type:   PreconditionTypeAnd,
					Result: true,
					And: []*PreconditionEvalResult{
						{
							Type:   PreconditionTypeNot,
							Result: true,
							Not: &PreconditionEvalResult{
								Type:   PreconditionTypeExecutable,
								Result: false,
								Executable: &AllocatabilityInfo{
									Allocatability:         AllocatabilityNGNoDeliverableUpdates,
									DeliverablesNotUpdated: sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
								},
							},
						},
						{
							Type:   PreconditionTypeNot,
							Result: true,
							Not: &PreconditionEvalResult{
								Type:   PreconditionTypeExecutable,
								Result: false,
								Executable: &AllocatabilityInfo{
									Allocatability:         AllocatabilityNGNoDeliverableUpdates,
									DeliverablesNotUpdated: sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
								},
							},
						},
					},
				},
			},
		},
		"and condition (false)": {
			Env: NewEnv(
				// [D1] -> (P1) -> [D2]
				//  \ \
				//   \ +-> (P2) -> [D3]
				//    \
				//     +-> (P3) -> [D4]
				branch,
				sets.New(ResourceID.Compare, "R1"),
				NewAvailableAllocationsFunc(neededResourceSetsFunc),
				ConstInitialVolumeFunc(1),
				ExponentialReworkVolumeFunc(0.5, ConstInitialVolumeFunc(1)),
				ConstMaxRevisionMap(3, branch.FeedbackSourceDeliverables()),
				map[pfd.AtomicProcessID]*Precondition{
					"P1": NewTruePrecondition(),
					"P2": NewTruePrecondition(),
					"P3": NewAndPrecondition(
						NewNotPrecondition(NewExecutablePrecondition("P1")),
						NewNotPrecondition(NewExecutablePrecondition("P2")),
					),
				},
				neededResourceSetsFunc,
				ConstDeliverableAvailableTimeFunc(0),
				slog.New(slogtest.NewTestHandler(t)),
			),
			State: NewState(
				0,
				map[pfd.AtomicDeliverableID]int{
					"D1": 1, // Initial deliverable
					"D2": 0, // Any
					"D3": 0, // Any
					"D4": 0, // Any
				},
				map[pfd.AtomicProcessID]Volume{
					"P1": 1, // Any greater than or equal to MinimumVolume
					"P2": 1, // Any greater than or equal to MinimumVolume
					"P3": 1, // Any less than or equal to initial volume
				},
				map[pfd.AtomicProcessID]int{
					"P1": 0, // Any
					"P2": 0, // Any
					"P3": 0, // Any
				},
				Allocation{},
				map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]{
					"P1": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
					"P2": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
					"P3": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
				},
			),
			Want: map[pfd.AtomicProcessID]*PreconditionEvalResult{
				"P1": {
					Type:   PreconditionTypeTrue,
					Result: true,
				},
				"P2": {
					Type:   PreconditionTypeTrue,
					Result: true,
				},
				"P3": {
					Type:   PreconditionTypeAnd,
					Result: false,
					And: []*PreconditionEvalResult{
						{
							Type:   PreconditionTypeNot,
							Result: false,
							Not: &PreconditionEvalResult{
								Type:   PreconditionTypeExecutable,
								Result: true,
								Executable: &AllocatabilityInfo{
									Allocatability: AllocatabilityOKStartable,
								},
							},
						},
						{
							Type:   PreconditionTypeNot,
							Result: false,
							Not: &PreconditionEvalResult{
								Type:   PreconditionTypeExecutable,
								Result: true,
								Executable: &AllocatabilityInfo{
									Allocatability: AllocatabilityOKStartable,
								},
							},
						},
					},
				},
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			m := make(map[pfd.AtomicProcessID]*PreconditionEvalResult)
			for ap, precondition := range testCase.Env.PreconditionMap {
				m[ap] = precondition.Eval(testCase.Env, testCase.State.RemainedVolumeMap, testCase.State.RevisionMap, testCase.State.AllocationShouldContinue, testCase.State.UpdatedDeliverablesNotHandled)
			}
			if !reflect.DeepEqual(m, testCase.Want) {
				t.Error(cmp.Diff(testCase.Want, m))
			}
		})
	}
}

func TestPrecondition_Compile(t *testing.T) {
	testCases := map[string]struct {
		PFD   *pfd.ValidPFD
		Input *Precondition
		Want  *Precondition
	}{
		"/regression_complete": {
			//                                    - - - - - - - - - - - - - - - - - - - - -
			//               - - -              /                    - - -                  \
			//             V       \           V                   V       \                 \
			// [D1] ---> (P1) ---> [D2] ---> (P2) ---> [D3] ---> (P3) ---> [D4] --> (P4) --> [D5]
			PFD: pfd.NewSafePFD(
				map[pfd.AtomicProcessID]string{
					"P1": "P1",
					"P2": "P2",
					"P3": "P3",
					"P4": "P4",
				},
				map[pfd.AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
					"D3": "D3",
					"D4": "D4",
					"D5": "D5",
				},
				map[pfd.AtomicProcessID]*pfd.RelationTriple{
					"P1": {
						Inputs:         sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
						FeedbackInputs: sets.New(pfd.AtomicDeliverableID.Compare, "D2"),
						Outputs:        sets.New(pfd.AtomicDeliverableID.Compare, "D2"),
					},
					"P2": {
						Inputs:         sets.New(pfd.AtomicDeliverableID.Compare, "D2"),
						FeedbackInputs: sets.New(pfd.AtomicDeliverableID.Compare, "D5"),
						Outputs:        sets.New(pfd.AtomicDeliverableID.Compare, "D3"),
					},
					"P3": {
						Inputs:         sets.New(pfd.AtomicDeliverableID.Compare, "D3"),
						FeedbackInputs: sets.New(pfd.AtomicDeliverableID.Compare, "D4"),
						Outputs:        sets.New(pfd.AtomicDeliverableID.Compare, "D4"),
					},
					"P4": {
						Inputs:         sets.New(pfd.AtomicDeliverableID.Compare, "D4"),
						FeedbackInputs: sets.New(pfd.AtomicDeliverableID.Compare),
						Outputs:        sets.New(pfd.AtomicDeliverableID.Compare, "D5"),
					},
				},
				map[pfd.CompositeProcessID]*pairs.Pair[string, *sets.Set[pfd.AtomicProcessID]]{},
				map[pfd.CompositeDeliverableID]*pairs.Pair[string, *sets.Set[pfd.AtomicDeliverableID]]{},
			),
			Input: NewAllBackwardReachableFeedbackSourcesCompleted("P3"),
			Want: NewAndPrecondition(
				NewFeedbackSourceCompletedPrecondition("D2"),
				NewNotPrecondition(NewExecutablePrecondition("P1")),
				NewNotPrecondition(NewExecutablePrecondition("P2")),
			),
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			got := testCase.Input.Compile(testCase.PFD, slog.New(slogtest.NewTestHandler(t)))
			if !reflect.DeepEqual(got, testCase.Want) {
				t.Error(cmp.Diff(testCase.Want, got))
			}
		})
	}
}
