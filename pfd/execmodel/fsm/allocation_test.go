package fsm

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
)

func TestAvailableAllocations(t *testing.T) {
	initVolume := /* any */ Volume(2)
	maxLoopCount := /* any */ 3
	logger := slog.New(slogtest.NewTestHandler(t))
	// [D1]----> (P1) -> [D2]
	//    \
	//     +---> (P2) -> [D3]
	//      \
	//       +-> (P3) -> [D4]
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
	availableTimeFunc := AlwaysAvailableTimeFunc()
	neededResourceSetsFunc := NeededResourceSetsFuncByMap(map[pfd.AtomicProcessID]*sets.Set[AllocationElement]{
		"P1": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: 2},
		),
		"P2": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: 3},
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R3"), ConsumedVolume: 4},
		),
		"P3": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R1", "R2", "R3"), ConsumedVolume: 5},
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

	state := State{
		Time: 0, // Any
		RemainedVolumeMap: map[pfd.AtomicProcessID]Volume{
			"P1": initVolume,
			"P2": initVolume,
			"P3": initVolume,
		},
		RevisionMap: map[pfd.AtomicDeliverableID]int{
			"D1": 1, // Initial deliverable
			"D2": 0, // Not created yet
			"D3": 0, // Not created yet
			"D4": 0, // Not created yet
		},
		NumOfCompleteMap: map[pfd.AtomicProcessID]int{
			"P1": 0, // Any
			"P2": 0, // Any
			"P3": 0, // Any
		},
		AllocationShouldContinue: Allocation{},
		UpdatedDeliverablesNotHandled: map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]{
			"P1": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
			"P2": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
			"P3": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
		},
	}

	got := NewAvailableAllocationsFunc(neededResourceSetsFunc)(state, env.NewlyAllocatables(state))

	expected := sets.New(
		CompareAllocationByTotalConsumedVolume,
		// NOTE: Empty allocations are not allowed in order to ensure the finiteness of the FSM state graph search. Allowing empty allocations would create infinite transitions that keep doing empty allocations forever.
		// Allocation{},
		Allocation{
			"P1": {Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: 2},
			"P2": {Resources: sets.New(ResourceID.Compare, "R3"), ConsumedVolume: 4},
		},
		Allocation{
			"P1": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
			"P2": {Resources: sets.New(ResourceID.Compare, "R3"), ConsumedVolume: 4},
		},
		Allocation{
			"P3": {Resources: sets.New(ResourceID.Compare, "R1", "R2", "R3"), ConsumedVolume: 5},
		},
		Allocation{
			"P2": {Resources: sets.New(ResourceID.Compare, "R3"), ConsumedVolume: 4},
		},
		Allocation{
			"P1": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
			"P2": {Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: 3},
		},
		Allocation{
			"P2": {Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: 3},
		},
		Allocation{
			"P1": {Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: 2},
		},
		Allocation{
			"P1": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
		},
	)

	if !reflect.DeepEqual(got, expected) {
		t.Log(AllocationSetString(expected))
		t.Log(AllocationSetString(got))
		t.Error(cmp.Diff(expected, got))
	}
}

func TestAvailableAllocationsWithContinuingAllocation(t *testing.T) {
	initVolume := /* any */ Volume(2)
	maxLoopCount := /* any */ 3
	logger := slog.New(slogtest.NewTestHandler(t))
	// [D1]----> (P1) -> [D2]
	//    \
	//     +---> (P2) -> [D3]
	//
	// P1 needs R1, P2 needs R1.
	// P1 is already continuing with R1, so P2 should NOT be allocated R1.
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
			&pfd.Edge{Source: "D1", Target: "P2"},
			&pfd.Edge{Source: "P1", Target: "D2"},
			&pfd.Edge{Source: "P2", Target: "D3"},
		),
	})
	availableTimeFunc := AlwaysAvailableTimeFunc()
	neededResourceSetsFunc := NeededResourceSetsFuncByMap(map[pfd.AtomicProcessID]*sets.Set[AllocationElement]{
		"P1": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
		),
		"P2": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
		),
	})
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
		Time: 1,
		RemainedVolumeMap: map[pfd.AtomicProcessID]Volume{
			"P1": 1, // P1 is still running
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
		// P1 is continuing with R1
		AllocationShouldContinue: Allocation{
			"P1": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
		},
		UpdatedDeliverablesNotHandled: map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]{
			"P1": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
			"P2": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
		},
	}

	// P2 is newly allocatable (P1 is continuing, not newly allocatable)
	newlyAllocatables := env.NewlyAllocatables(state)
	got := NewAvailableAllocationsFunc(neededResourceSetsFunc)(state, newlyAllocatables)

	// P2 needs R1 but P1 is already continuing with R1, so P2 cannot be allocated.
	// The only valid allocation is just the continuing P1.
	expected := sets.New(
		CompareAllocationByTotalConsumedVolume,
		Allocation{
			"P1": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
		},
	)

	if !reflect.DeepEqual(got, expected) {
		t.Log(AllocationSetString(expected))
		t.Log(AllocationSetString(got))
		t.Error(cmp.Diff(expected, got))
	}
}

func TestMaximalAvailableAllocationsWithContinuingAllocation(t *testing.T) {
	initVolume := /* any */ Volume(2)
	maxLoopCount := /* any */ 3
	logger := slog.New(slogtest.NewTestHandler(t))
	// [D1]----> (P1) -> [D2]
	//    \
	//     +---> (P2) -> [D3]
	//
	// P1 needs R1, P2 needs R1.
	// P1 is already continuing with R1, so P2 should NOT be allocated R1.
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
			&pfd.Edge{Source: "D1", Target: "P2"},
			&pfd.Edge{Source: "P1", Target: "D2"},
			&pfd.Edge{Source: "P2", Target: "D3"},
		),
	})
	availableTimeFunc := AlwaysAvailableTimeFunc()
	neededResourceSetsFunc := NeededResourceSetsFuncByMap(map[pfd.AtomicProcessID]*sets.Set[AllocationElement]{
		"P1": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
		),
		"P2": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
		),
	})
	env := NewEnv(
		p,
		sets.New(ResourceID.Compare, "R1"),
		NewMaximalAvailableAllocationsFunc(neededResourceSetsFunc),
		ConstInitialVolumeFunc(initVolume),
		ExponentialReworkVolumeFunc(0.5, ConstInitialVolumeFunc(initVolume)),
		ConstMaxRevisionMap(maxLoopCount, p.FeedbackSourceDeliverables()),
		NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
		neededResourceSetsFunc,
		availableTimeFunc,
		logger,
	)

	state := State{
		Time: 1,
		RemainedVolumeMap: map[pfd.AtomicProcessID]Volume{
			"P1": 1,
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
		AllocationShouldContinue: Allocation{
			"P1": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
		},
		UpdatedDeliverablesNotHandled: map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]{
			"P1": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
			"P2": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
		},
	}

	newlyAllocatables := env.NewlyAllocatables(state)
	got := NewMaximalAvailableAllocationsFunc(neededResourceSetsFunc)(state, newlyAllocatables)

	// P2 needs R1 but P1 is already continuing with R1, so P2 cannot be allocated.
	// The only valid allocation is just the continuing P1.
	expected := sets.New(
		CompareAllocationByTotalConsumedVolume,
		Allocation{
			"P1": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
		},
	)

	if !reflect.DeepEqual(got, expected) {
		t.Log(AllocationSetString(expected))
		t.Log(AllocationSetString(got))
		t.Error(cmp.Diff(expected, got))
	}
}

func TestMaximalAvailableAllocations(t *testing.T) {
	initVolume := /* any */ Volume(2)
	maxLoopCount := /* any */ 3
	logger := slog.New(slogtest.NewTestHandler(t))
	// [D1]----> (P1) -> [D2]
	//    \
	//     +---> (P2) -> [D3]
	//      \
	//       +-> (P3) -> [D4]
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
	availableTimeFunc := AlwaysAvailableTimeFunc()
	neededResourceSetsFunc := NeededResourceSetsFuncByMap(map[pfd.AtomicProcessID]*sets.Set[AllocationElement]{
		"P1": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: 2},
		),
		"P2": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: 3},
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R3"), ConsumedVolume: 4},
		),
		"P3": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R1", "R2", "R3"), ConsumedVolume: 5},
		),
	})
	env := NewEnv(
		p,
		sets.New(ResourceID.Compare, "R1", "R2", "R3"),
		NewMaximalAvailableAllocationsFunc(neededResourceSetsFunc),
		ConstInitialVolumeFunc(initVolume),
		ExponentialReworkVolumeFunc(0.5, ConstInitialVolumeFunc(initVolume)),
		ConstMaxRevisionMap(maxLoopCount, p.FeedbackSourceDeliverables()),
		NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
		neededResourceSetsFunc,
		availableTimeFunc,
		logger,
	)

	state := State{
		Time: 0, // Any
		RemainedVolumeMap: map[pfd.AtomicProcessID]Volume{
			"P1": initVolume,
			"P2": initVolume,
			"P3": initVolume,
		},
		RevisionMap: map[pfd.AtomicDeliverableID]int{
			"D1": 1, // Initial deliverable
			"D2": 0, // Not created yet
			"D3": 0, // Not created yet
			"D4": 0, // Not created yet
		},
		NumOfCompleteMap: map[pfd.AtomicProcessID]int{
			"P1": 0, // Any
			"P2": 0, // Any
			"P3": 0, // Any
		},
		AllocationShouldContinue: Allocation{},
		UpdatedDeliverablesNotHandled: map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]{
			"P1": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
			"P2": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
			"P3": sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
		},
	}

	got := NewMaximalAvailableAllocationsFunc(neededResourceSetsFunc)(state, env.NewlyAllocatables(state))

	expected := sets.New(
		CompareAllocationByTotalConsumedVolume,
		// NOTE: Empty allocations are not allowed in order to ensure the finiteness of the FSM state graph search. Allowing empty allocations would create infinite transitions that keep doing empty allocations forever.
		// Allocation{},
		Allocation{
			"P1": {Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: 2},
			"P2": {Resources: sets.New(ResourceID.Compare, "R3"), ConsumedVolume: 4},
		},
		Allocation{
			"P1": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
			"P2": {Resources: sets.New(ResourceID.Compare, "R3"), ConsumedVolume: 4},
		},
		Allocation{
			"P3": {Resources: sets.New(ResourceID.Compare, "R1", "R2", "R3"), ConsumedVolume: 5},
		},
		Allocation{
			"P1": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
			"P2": {Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: 3},
		},
	)

	if !reflect.DeepEqual(got, expected) {
		t.Log(AllocationSetString(expected))
		t.Log(AllocationSetString(got))
		t.Error(cmp.Diff(expected, got))
	}
}
