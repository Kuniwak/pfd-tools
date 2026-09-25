package fsm

import (
	"encoding/json"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
)

func TestInitialState(t *testing.T) {
	maxLoopCount := 3
	initVolume := Volume(2)
	logger := slog.New(slogtest.NewTestHandler(t))

	p := newSafePFDByUnsafePFD(&pfd.PFD{
		Nodes: sets.New(
			(*pfd.Node).Compare,
			&pfd.Node{ID: "P1", Type: pfd.NodeTypeAtomicProcess},
			&pfd.Node{ID: "P2", Type: pfd.NodeTypeAtomicProcess},
			&pfd.Node{ID: "P3", Type: pfd.NodeTypeAtomicProcess},
			&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "D3", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "D4", Type: pfd.NodeTypeAtomicDeliverable},
		),
		Edges: sets.New(
			(*pfd.Edge).Compare,
			&pfd.Edge{Source: "D1", Target: "P1"},
			&pfd.Edge{Source: "D1", Target: "P2"},
			&pfd.Edge{Source: "P1", Target: "D2"},
			&pfd.Edge{Source: "P2", Target: "D3"},
			&pfd.Edge{Source: "D2", Target: "P3"},
			&pfd.Edge{Source: "P3", Target: "D4"},
		),
	})
	availableTimeFunc := AlwaysAvailableTimeFunc()
	neededResourceSetsFunc := NeededResourceSetsFuncByMap(map[pfd.AtomicProcessID]*sets.Set[AllocationElement]{
		"P1": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R1", "R2"), ConsumedVolume: 2},
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: 1},
		),
		"P2": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R3"), ConsumedVolume: 2},
		),
		"P3": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R3"), ConsumedVolume: 1},
		),
	})
	env := NewEnv(
		p,
		sets.New(ResourceID.Compare, "R1", "R2", "R3"),
		NewAvailableAllocationsFunc(neededResourceSetsFunc),
		ConstInitialVolumeFunc(initVolume),
		ExponentialReworkVolumeFunc(0.5, ConstInitialVolumeFunc(initVolume)),
		ConstMaxRevisionMap(maxLoopCount, p.FeedbackSourceDeliverables()),
		NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
		neededResourceSetsFunc,
		availableTimeFunc,
		logger,
	)

	state := env.InitialState()
	expected := State{
		Time: 0,
		RemainedVolumeMap: map[pfd.AtomicProcessID]Volume{
			"P1": 2,
			"P2": 2,
			"P3": 2,
		},
		RevisionMap: map[pfd.AtomicDeliverableID]int{
			"D1": 1,
			"D2": 0,
			"D3": 0,
			"D4": 0,
		},
		NumOfCompleteMap: map[pfd.AtomicProcessID]int{
			"P1": 0,
			"P2": 0,
			"P3": 0,
		},
		AllocationShouldContinue: Allocation{},
		UpdatedDeliverablesNotHandled: map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]{
			"P1": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
			"P2": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
			"P3": sets.New(pfd.AtomicDeliverableID.Compare),
		},
	}

	if !reflect.DeepEqual(state, expected) {
		t.Error(cmp.Diff(expected, state))
	}
}

func TestAllocatability(t *testing.T) {
	t.Run("first work", func(t *testing.T) {
		maxLoopCount := 3
		initVolume := Volume(2)
		logger := slog.New(slogtest.NewTestHandler(t))
		p := newSafePFDByUnsafePFD(&pfd.PFD{

			Nodes: sets.New(
				(*pfd.Node).Compare,
				&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
				&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
				&pfd.Node{ID: "P1", Type: pfd.NodeTypeAtomicProcess},
			),
			Edges: sets.New(
				(*pfd.Edge).Compare,
				&pfd.Edge{Source: "D1", Target: "P1"},
				&pfd.Edge{Source: "P1", Target: "D2"},
				&pfd.Edge{Source: "D2", Target: "P1", IsFeedback: true},
			),
		})
		availableTimeFunc := AlwaysAvailableTimeFunc()
		neededResourceSetsFunc := AnyNeededResourceSetsFunc()
		env := NewEnv(
			p,
			sets.New(ResourceID.Compare, "R1"),
			NewAvailableAllocationsFunc(neededResourceSetsFunc),
			ConstInitialVolumeFunc(initVolume),
			ExponentialReworkVolumeFunc(0.5, ConstInitialVolumeFunc(initVolume)),
			ConstMaxRevisionMap(maxLoopCount, p.FeedbackSourceDeliverables()),
			NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
			neededResourceSetsFunc,
			availableTimeFunc,
			logger,
		)

		state := State{
			Time: 123,
			RemainedVolumeMap: map[pfd.AtomicProcessID]Volume{
				"P1": initVolume,
			},
			RevisionMap: map[pfd.AtomicDeliverableID]int{
				"D1": 1,
				"D2": 0,
			},
			NumOfCompleteMap: map[pfd.AtomicProcessID]int{
				"P1": 0,
			},
			AllocationShouldContinue: Allocation{},
			UpdatedDeliverablesNotHandled: map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]{
				"P1": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
			},
		}

		got := env.AllocatabilityInfoMap(state)
		expected := AllocatabilityInfoMap{
			"P1": NewAllocatabilityOKStartable(),
		}

		if !reflect.DeepEqual(got, expected) {
			t.Error(cmp.Diff(expected, got))
		}
	})

	t.Run("continue", func(t *testing.T) {

		maxLoopCount := 0
		initVolume := Volume(2)
		logger := slog.New(slogtest.NewTestHandler(t))
		p := newSafePFDByUnsafePFD(&pfd.PFD{

			Nodes: sets.New(
				(*pfd.Node).Compare,
				&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
				&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
				&pfd.Node{ID: "P1", Type: pfd.NodeTypeAtomicProcess},
			),
			Edges: sets.New(
				(*pfd.Edge).Compare,
				&pfd.Edge{Source: "D1", Target: "P1"},
				&pfd.Edge{Source: "P1", Target: "D2"},
			),
		})
		p1Consumed := Volume(1)
		availableTimeFunc := AlwaysAvailableTimeFunc()
		neededResourceSetsFunc := AnyNeededResourceSetsFunc()
		env := NewEnv(
			p,
			sets.New(ResourceID.Compare, "R1"),
			NewAvailableAllocationsFunc(neededResourceSetsFunc),
			ConstInitialVolumeFunc(initVolume),
			ExponentialReworkVolumeFunc(0.5, ConstInitialVolumeFunc(initVolume)),
			ConstMaxRevisionMap(maxLoopCount, p.FeedbackSourceDeliverables()),
			NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
			neededResourceSetsFunc,
			availableTimeFunc,
			logger,
		)
		state := State{
			Time: 123,
			RemainedVolumeMap: map[pfd.AtomicProcessID]Volume{
				"P1": initVolume - p1Consumed,
			},
			RevisionMap: map[pfd.AtomicDeliverableID]int{
				"D1": 1,
				"D2": 0,
			},
			NumOfCompleteMap: map[pfd.AtomicProcessID]int{
				"P1": 0,
			},
			AllocationShouldContinue: Allocation{
				"P1": {
					Resources:      sets.New(ResourceID.Compare, "R1"),
					ConsumedVolume: p1Consumed,
				},
			},
			UpdatedDeliverablesNotHandled: map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]{
				"P1": sets.New(pfd.AtomicDeliverableID.Compare),
			},
		}

		got := env.AllocatabilityInfoMap(state)
		expected := AllocatabilityInfoMap{
			"P1": NewAllocatabilityOKContinuable(),
		}

		if !reflect.DeepEqual(got, expected) {
			t.Error(cmp.Diff(expected, got))
		}
	})

	t.Run("rework", func(t *testing.T) {
		maxLoopCount := 3
		initVolume := Volume(2)
		logger := slog.New(slogtest.NewTestHandler(t))

		p := newSafePFDByUnsafePFD(&pfd.PFD{
			Nodes: sets.New(
				(*pfd.Node).Compare,
				&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
				&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
				&pfd.Node{ID: "P1", Type: pfd.NodeTypeAtomicProcess},
			),
			Edges: sets.New(
				(*pfd.Edge).Compare,
				&pfd.Edge{Source: "D1", Target: "P1"},
				&pfd.Edge{Source: "D2", Target: "P1", IsFeedback: true},
				&pfd.Edge{Source: "P1", Target: "D2"},
			),
		})
		availableTimeFunc := AlwaysAvailableTimeFunc()
		neededResourceSetsFunc := AnyNeededResourceSetsFunc()
		env := NewEnv(
			p,
			sets.New(ResourceID.Compare, "R1"),
			NewAvailableAllocationsFunc(neededResourceSetsFunc),
			ConstInitialVolumeFunc(initVolume),
			ExponentialReworkVolumeFunc(0.5, ConstInitialVolumeFunc(initVolume)),
			ConstMaxRevisionMap(maxLoopCount, p.FeedbackSourceDeliverables()),
			NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
			neededResourceSetsFunc,
			availableTimeFunc,
			logger,
		)

		state := State{
			Time: 123,
			RemainedVolumeMap: map[pfd.AtomicProcessID]Volume{
				"P1": 1,
			},
			RevisionMap: map[pfd.AtomicDeliverableID]int{
				"D1": 1,
				"D2": 1,
			},
			NumOfCompleteMap: map[pfd.AtomicProcessID]int{
				"P1": 1,
			},
			AllocationShouldContinue: Allocation{},
			UpdatedDeliverablesNotHandled: map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]{
				"P1": sets.New(pfd.AtomicDeliverableID.Compare, "D2"),
			},
		}

		got := env.AllocatabilityInfoMap(state)
		expected := AllocatabilityInfoMap{
			"P1": NewAllocatabilityOKStartable(),
		}

		if !reflect.DeepEqual(got, expected) {
			t.Error(cmp.Diff(expected, got))
		}
	})

	t.Run("insufficient inputs (initial deliverable)", func(t *testing.T) {
		maxLoopCount := 3
		initVolume := Volume(2)
		logger := slog.New(slogtest.NewTestHandler(t))

		p := newSafePFDByUnsafePFD(&pfd.PFD{
			Nodes: sets.New(
				(*pfd.Node).Compare,
				&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
				&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
				&pfd.Node{ID: "P1", Type: pfd.NodeTypeAtomicProcess},
			),
			Edges: sets.New(
				(*pfd.Edge).Compare,
				&pfd.Edge{Source: "D1", Target: "P1"},
				&pfd.Edge{Source: "P1", Target: "D2"},
			),
		})
		availableTime := execmodel.Time(1)
		availableTimeFunc := ConstDeliverableAvailableTimeFunc(availableTime)
		neededResourceSetsFunc := AnyNeededResourceSetsFunc()
		env := NewEnv(
			p,
			sets.New(ResourceID.Compare, "R1"),
			NewAvailableAllocationsFunc(neededResourceSetsFunc),
			ConstInitialVolumeFunc(initVolume),
			ExponentialReworkVolumeFunc(0.5, ConstInitialVolumeFunc(initVolume)),
			ConstMaxRevisionMap(maxLoopCount, p.FeedbackSourceDeliverables()),
			NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
			neededResourceSetsFunc,
			availableTimeFunc,
			logger,
		)

		state := State{
			Time: availableTime - 1,
			RemainedVolumeMap: map[pfd.AtomicProcessID]Volume{
				"P1": 1,
			},
			RevisionMap: map[pfd.AtomicDeliverableID]int{
				"D1": 0,
				"D2": 0,
			},
			NumOfCompleteMap: map[pfd.AtomicProcessID]int{
				"P1": maxLoopCount - 1,
			},
			AllocationShouldContinue:      Allocation{},
			UpdatedDeliverablesNotHandled: map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]{},
		}

		got := env.AllocatabilityInfoMap(state)
		expected := AllocatabilityInfoMap{
			"P1": NewAllocatabilityNGInsufficientInputs(sets.New(pfd.AtomicDeliverableID.Compare, "D1")),
		}

		if !reflect.DeepEqual(got, expected) {
			t.Error(cmp.Diff(expected, got))
		}
	})

	t.Run("insufficient inputs (intermediate deliverable)", func(t *testing.T) {
		maxLoopCount := 3
		logger := slog.New(slogtest.NewTestHandler(t))
		initVolume := Volume(2)

		p := newSafePFDByUnsafePFD(&pfd.PFD{
			Nodes: sets.New(
				(*pfd.Node).Compare,
				&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
				&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
				&pfd.Node{ID: "D3", Type: pfd.NodeTypeAtomicDeliverable},
				&pfd.Node{ID: "P1", Type: pfd.NodeTypeAtomicProcess},
				&pfd.Node{ID: "P2", Type: pfd.NodeTypeAtomicProcess},
			),
			Edges: sets.New(
				(*pfd.Edge).Compare,
				&pfd.Edge{Source: "D1", Target: "P1"},
				&pfd.Edge{Source: "D2", Target: "P2"},
				&pfd.Edge{Source: "P1", Target: "D2"},
				&pfd.Edge{Source: "P2", Target: "D3"},
			),
		})
		availableTimeFunc := AlwaysAvailableTimeFunc()
		neededResourceSetsFunc := AnyNeededResourceSetsFunc()
		env := NewEnv(
			p,
			sets.New(ResourceID.Compare, "R1"),
			NewAvailableAllocationsFunc(neededResourceSetsFunc),
			ConstInitialVolumeFunc(initVolume),
			ExponentialReworkVolumeFunc(0.5, ConstInitialVolumeFunc(initVolume)),
			ConstMaxRevisionMap(maxLoopCount, p.FeedbackSourceDeliverables()),
			NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
			neededResourceSetsFunc,
			availableTimeFunc,
			logger,
		)

		state := State{
			Time: 123,
			RemainedVolumeMap: map[pfd.AtomicProcessID]Volume{
				"P1": initVolume,
				"P2": initVolume,
			},
			RevisionMap: map[pfd.AtomicDeliverableID]int{
				"D1": 1,
				"D2": 0,
				"D3": 0,
			},
			NumOfCompleteMap: map[pfd.AtomicProcessID]int{
				"P1": 0,
				"P2": 0,
			},
			AllocationShouldContinue: Allocation{},
			UpdatedDeliverablesNotHandled: map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]{
				"P1": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
				"P2": sets.New(pfd.AtomicDeliverableID.Compare),
			},
		}

		got := env.AllocatabilityInfoMap(state)
		expected := AllocatabilityInfoMap{
			"P1": NewAllocatabilityOKStartable(),
			"P2": NewAllocatabilityNGInsufficientInputs(sets.New(pfd.AtomicDeliverableID.Compare, "D2")),
		}

		if !reflect.DeepEqual(got, expected) {
			t.Error(cmp.Diff(expected, got))
		}
	})

	t.Run("precondition not met", func(t *testing.T) {
		initVolume := Volume(2)
		maxLoopCount := 3
		logger := slog.New(slogtest.NewTestHandler(t))

		p := newSafePFDByUnsafePFD(pfd.PresetButterflyLoop)
		availableTimeFunc := AlwaysAvailableTimeFunc()
		reworkVolumeFunc := ExponentialReworkVolumeFunc(0.5, ConstInitialVolumeFunc(initVolume))
		neededResourceSetsFunc := AnyNeededResourceSetsFunc()
		env := NewEnv(
			p,
			sets.New(ResourceID.Compare, "R1"),
			NewAvailableAllocationsFunc(neededResourceSetsFunc),
			ConstInitialVolumeFunc(initVolume),
			reworkVolumeFunc,
			ConstMaxRevisionMap(maxLoopCount, p.FeedbackSourceDeliverables()),
			NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{
				"P3": NewFeedbackSourceCompletedPrecondition(pfd.AtomicDeliverableID("D4")),
			}),
			neededResourceSetsFunc,
			availableTimeFunc,
			logger,
		)
		state := State{
			Time: 123,
			RemainedVolumeMap: map[pfd.AtomicProcessID]Volume{
				"P1": reworkVolumeFunc("P1", 1),
				"P2": 0,
				"P3": initVolume,
			},
			RevisionMap: map[pfd.AtomicDeliverableID]int{
				"D1": 1,
				"D2": 1,
				"D3": 1,
				"D4": 1,
				"D5": 0,
			},
			NumOfCompleteMap: map[pfd.AtomicProcessID]int{
				"P1": 1,
				"P2": 1,
				"P3": 0,
			},
			AllocationShouldContinue: Allocation{},
			UpdatedDeliverablesNotHandled: map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]{
				"P1": sets.New(pfd.AtomicDeliverableID.Compare, "D4"),
				"P2": sets.New(pfd.AtomicDeliverableID.Compare),
				"P3": sets.New(pfd.AtomicDeliverableID.Compare, "D3"),
			},
		}

		got := env.AllocatabilityInfoMap(state)
		expected := AllocatabilityInfoMap{
			"P1": NewAllocatabilityOKStartable(),
			"P2": NewAllocatabilityNGNoDeliverableUpdates(sets.New(pfd.AtomicDeliverableID.Compare, "D2")),
			"P3": NewAllocatabilityNGPreconditionNotMet(`{
  "type": "EXEC_BETWEEN",
  "result": false,
  "exec_between_target": "D4",
  "exec_between_begin": {
    "type": "MAX_REVISION",
    "value": 3,
    "max_revision_deliverable": "D4"
  },
  "exec_between_end": {
    "type": "INFINITY",
    "value": 0
  },
  "revision": 1
}
`),
		}
		if !reflect.DeepEqual(got, expected) {
			t.Error(cmp.Diff(expected, got))
		}
	})

	t.Run("updated all deliverables handled", func(t *testing.T) {
		initVolume := Volume(2)
		maxLoopCount := 3
		logger := slog.New(slogtest.NewTestHandler(t))
		p := newSafePFDByUnsafePFD(pfd.PresetNestedLoop)
		availableTimeFunc := AlwaysAvailableTimeFunc()
		reworkVolumeFunc := ExponentialReworkVolumeFunc(0.5, ConstInitialVolumeFunc(initVolume))
		neededResourceSetsFunc := AnyNeededResourceSetsFunc()
		env := NewEnv(
			p,
			sets.New(ResourceID.Compare, "R1"),
			NewAvailableAllocationsFunc(neededResourceSetsFunc),
			ConstInitialVolumeFunc(initVolume),
			reworkVolumeFunc,
			ConstMaxRevisionMap(maxLoopCount, p.FeedbackSourceDeliverables()),
			NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
			neededResourceSetsFunc,
			availableTimeFunc,
			logger,
		)
		state := State{
			Time: 123,
			RemainedVolumeMap: map[pfd.AtomicProcessID]Volume{
				"P1": reworkVolumeFunc("P1", 1),
				"P2": reworkVolumeFunc("P2", 1),
				"P3": reworkVolumeFunc("P3", 1),
			},
			RevisionMap: map[pfd.AtomicDeliverableID]int{
				"D1": 1,
				"D2": 1,
				"D3": 1,
				"D4": 1,
			},
			NumOfCompleteMap: map[pfd.AtomicProcessID]int{
				"P1": 1,
				"P2": 1,
				"P3": 1,
			},
			AllocationShouldContinue: Allocation{},
			UpdatedDeliverablesNotHandled: map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]{
				"P1": sets.New(pfd.AtomicDeliverableID.Compare, "D4"),
				"P2": sets.New(pfd.AtomicDeliverableID.Compare),
				"P3": sets.New(pfd.AtomicDeliverableID.Compare),
			},
		}
		got := env.AllocatabilityInfoMap(state)
		expected := AllocatabilityInfoMap{
			"P1": NewAllocatabilityOKStartable(),
			"P2": NewAllocatabilityNGNoDeliverableUpdates(sets.New(pfd.AtomicDeliverableID.Compare, "D2", "D3")),
			"P3": NewAllocatabilityNGNoDeliverableUpdates(sets.New(pfd.AtomicDeliverableID.Compare, "D3")),
		}
		if !reflect.DeepEqual(got, expected) {
			t.Error(cmp.Diff(expected, got))
		}
	})
}

func TestNextState(t *testing.T) {
	t.Run("triple branch", func(t *testing.T) {
		initVolume := Volume(2)
		maxLoopCount := 3
		logger := slog.New(slogtest.NewTestHandler(t))

		p := newSafePFDByUnsafePFD(&pfd.PFD{
			Nodes: sets.New(
				(*pfd.Node).Compare,
				&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
				&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
				&pfd.Node{ID: "D3", Type: pfd.NodeTypeAtomicDeliverable},
				&pfd.Node{ID: "D4", Type: pfd.NodeTypeAtomicDeliverable},
				&pfd.Node{ID: "P1", Type: pfd.NodeTypeAtomicProcess},
				&pfd.Node{ID: "P2", Type: pfd.NodeTypeAtomicProcess},
				&pfd.Node{ID: "P3", Type: pfd.NodeTypeAtomicProcess},
			),
			Edges: sets.New(
				(*pfd.Edge).Compare,
				&pfd.Edge{Source: "D1", Target: "P1"},
				&pfd.Edge{Source: "D1", Target: "P2"},
				&pfd.Edge{Source: "D1", Target: "P3"},
				&pfd.Edge{Source: "P1", Target: "D2"},
				&pfd.Edge{Source: "P2", Target: "D3"},
				&pfd.Edge{Source: "P3", Target: "D4"},
			),
		})
		initVolumeFunc := ConstInitialVolumeFunc(initVolume)
		reworkVolumeFunc := ExponentialReworkVolumeFunc(0.5, initVolumeFunc)
		remainedVolumeP2 := Volume(1)
		availableTimeFunc := AlwaysAvailableTimeFunc()
		neededResourceSetsFunc := NeededResourceSetsFuncByMap(map[pfd.AtomicProcessID]*sets.Set[AllocationElement]{
			"P1": sets.New(
				AllocationElement.Compare,
				AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: initVolume},
				AllocationElement{Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: 0.001},
			),
			"P2": sets.New(
				AllocationElement.Compare,
				AllocationElement{Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: initVolume - remainedVolumeP2},
				AllocationElement{Resources: sets.New(ResourceID.Compare, "R3"), ConsumedVolume: 0.001},
			),
			"P3": sets.New(
				AllocationElement.Compare,
				AllocationElement{Resources: sets.New(ResourceID.Compare, "R1", "R2", "R3"), ConsumedVolume: 3},
			),
		})
		env := NewEnv(
			p,
			sets.New(ResourceID.Compare, "R1", "R2", "R3"),
			NewAvailableAllocationsFunc(neededResourceSetsFunc),
			initVolumeFunc,
			reworkVolumeFunc,
			ConstMaxRevisionMap(maxLoopCount, p.FeedbackSourceDeliverables()),
			NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
			neededResourceSetsFunc,
			availableTimeFunc,
			logger,
		)

		state := State{
			Time: 123,
			RemainedVolumeMap: map[pfd.AtomicProcessID]Volume{
				"P1": initVolume,
				"P2": initVolume,
				"P3": initVolume,
			},
			RevisionMap: map[pfd.AtomicDeliverableID]int{
				"D1": 1,
				"D2": 0,
				"D3": 0,
				"D4": 0,
			},
			NumOfCompleteMap: map[pfd.AtomicProcessID]int{
				"P1": 0,
				"P2": 0,
				"P3": 0,
			},
			UpdatedDeliverablesNotHandled: map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]{
				"P1": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
				"P2": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
				"P3": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
			},
		}

		alloc := Allocation{
			"P1": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: initVolume},
			"P2": {Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: initVolume - remainedVolumeP2},
		}

		got, ok := env.NextState(state, alloc)
		if !ok {
			t.Fatal("next state is not found")
		}
		expected := State{
			Time: 124,
			RemainedVolumeMap: map[pfd.AtomicProcessID]Volume{
				"P1": reworkVolumeFunc("P1", 1),
				"P2": remainedVolumeP2,
				"P3": initVolume,
			},
			RevisionMap: map[pfd.AtomicDeliverableID]int{
				"D1": 1,
				"D2": 1,
				"D3": 0,
				"D4": 0,
			},
			NumOfCompleteMap: map[pfd.AtomicProcessID]int{
				"P1": 1,
				"P2": 0,
				"P3": 0,
			},
			AllocationShouldContinue: Allocation{
				"P2": {Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: initVolume - remainedVolumeP2},
			},
			UpdatedDeliverablesNotHandled: map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]{
				"P1": sets.New(pfd.AtomicDeliverableID.Compare),
				"P2": sets.New(pfd.AtomicDeliverableID.Compare),
				"P3": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.Error(cmp.Diff(expected, got))
		}
	})

	t.Run("recover volume", func(t *testing.T) {
		maxLoopCount := 3
		reworkVolumeRatio := 0.5
		initVolume := Volume(2)
		initVolumeFunc := ConstInitialVolumeFunc(initVolume)
		logger := slog.New(slogtest.NewTestHandler(t))

		p := newSafePFDByUnsafePFD(pfd.PresetSmallestLoop)
		availableTimeFunc := AlwaysAvailableTimeFunc()
		reworkVolumeFunc := ExponentialReworkVolumeFunc(reworkVolumeRatio, initVolumeFunc)
		neededResourceSetsFunc := NeededResourceSetsFuncByMap(map[pfd.AtomicProcessID]*sets.Set[AllocationElement]{
			"P1": sets.New(
				AllocationElement.Compare,
				AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
			),
		})
		env := NewEnv(
			p,
			sets.New(ResourceID.Compare, "R1"),
			NewAvailableAllocationsFunc(neededResourceSetsFunc),
			initVolumeFunc,
			reworkVolumeFunc,
			ConstMaxRevisionMap(maxLoopCount, p.FeedbackSourceDeliverables()),
			NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
			neededResourceSetsFunc,
			availableTimeFunc,
			logger,
		)
		state := env.InitialState()
		got, ok := env.NextState(state, Allocation{"P1": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1}})
		if !ok {
			t.Fatal("next state is not found")
		}
		expected := State{
			Time: 2,
			RemainedVolumeMap: map[pfd.AtomicProcessID]Volume{
				"P1": reworkVolumeFunc("P1", 1),
			},
			NumOfCompleteMap: map[pfd.AtomicProcessID]int{
				"P1": 1,
			},
			RevisionMap: map[pfd.AtomicDeliverableID]int{
				"D1": 1,
				"D2": 1,
			},
			AllocationShouldContinue: Allocation{},
			UpdatedDeliverablesNotHandled: map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]{
				"P1": sets.New(pfd.AtomicDeliverableID.Compare, "D2"),
			},
		}
		if !reflect.DeepEqual(got, expected) {
			t.Error(cmp.Diff(expected, got))
		}
	})

}

func AllocationSetString(a *sets.Set[Allocation]) string {
	s := strings.Builder{}
	e := json.NewEncoder(&s)
	e.SetIndent("", "  ")
	e.SetEscapeHTML(false)
	_ = e.Encode(a.Slice())
	return s.String()
}

func newSafePFDByUnsafePFD(p *pfd.PFD) *pfd.ValidPFD {
	return pfd.MustNewSafePFDByUnsafePFD(p)
}
