package fsm

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/cmp2"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
	"pgregory.net/rapid"
)

func TestSearchBestPlans(t *testing.T) {
	// [D1] -> (P1) -> [D2] -> (P3) -> [D4]
	//      \
	//       +-> (P2) -> [D3]
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
			&pfd.Edge{Source: "D2", Target: "P3"},
			&pfd.Edge{Source: "P1", Target: "D2"},
			&pfd.Edge{Source: "P2", Target: "D3"},
			&pfd.Edge{Source: "P3", Target: "D4"},
		),
	})
	availableTimeFunc := AlwaysAvailableTimeFunc()
	initVolume := Volume(1)
	initVolumeFunc := ConstInitialVolumeFunc(initVolume)
	logger := slog.New(slogtest.NewTestHandler(t))
	neededResourceSetsFunc := NeededResourceSetsFuncByMap(map[pfd.AtomicProcessID]*sets.Set[AllocationElement]{
		"P1": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: initVolume},
		),
		"P2": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: initVolume},
		),
		"P3": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R3"), ConsumedVolume: initVolume},
		),
	})
	env := NewEnv(
		p,
		sets.New(ResourceID.Compare, "R1", "R2", "R3"),
		NewAvailableAllocationsFunc(neededResourceSetsFunc),
		initVolumeFunc,
		ExponentialReworkVolumeFunc(0.5, initVolumeFunc),
		ConstMaxRevisionMap(3, p.FeedbackSourceDeliverables()),
		NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
		neededResourceSetsFunc,
		availableTimeFunc,
		logger,
	)

	plans, err := searchBestPlans(env)
	if err != nil {
		t.Fatal(err)
	}
	expected := sets.New(cmp2.CompareSlice[[]Allocation](Allocation.Compare),
		[]Allocation{
			{
				"P1": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: initVolume},
			},
			{
				"P2": {Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: initVolume},
				"P3": {Resources: sets.New(ResourceID.Compare, "R3"), ConsumedVolume: initVolume},
			},
		},
		[]Allocation{
			{
				"P1": {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: initVolume},
				"P2": {Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: initVolume},
			},
			{
				"P3": {Resources: sets.New(ResourceID.Compare, "R3"), ConsumedVolume: initVolume},
			},
		},
	)

	if !reflect.DeepEqual(pickOnlyAllocation(plans), expected) {
		t.Error(cmp.Diff(expected, pickOnlyAllocation(plans)))
	}
}

func TestSearchBestPlansPreset(t *testing.T) {
	for name, up := range pfd.PresetsAll {
		p := newSafePFDByUnsafePFD(up)
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			initVolumeFunc := ConstInitialVolumeFunc(Volume(2))
			availableResources := FakeAvailableResources(2)
			neededResourceSetsFunc := FakeNeededResourceSetsFunc(availableResources)
			env := NewEnv(
				p,
				availableResources,
				NewAvailableAllocationsFunc(neededResourceSetsFunc),
				initVolumeFunc,
				FakeReworkVolumeFunc(initVolumeFunc),
				ConstMaxRevisionMap(2, p.FeedbackSourceDeliverables()),
				NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
				neededResourceSetsFunc,
				FakeDeliverableAvailableTimeFunc(p, logger),
				logger,
			)

			_, err := searchBestPlans(env)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func FuzzSearchBestPlans(t *testing.F) {
	t.Fuzz(rapid.MakeFuzz(func(t *rapid.T) {
		up := pfd.AnyValidPFD(t, 100)
		p := newSafePFDByUnsafePFD(up)
		logger := slog.New(slogtest.NewRapidHandler(t))
		initVolumeFunc := ConstInitialVolumeFunc(Volume(2))
		availableResources := FakeAvailableResources(3)
		neededResourceSetsFunc := FakeNeededResourceSetsFunc(availableResources)
		env := NewEnv(
			p,
			availableResources,
			NewAvailableAllocationsFunc(neededResourceSetsFunc),
			initVolumeFunc,
			FakeReworkVolumeFunc(initVolumeFunc),
			ConstMaxRevisionMap(3, p.FeedbackSourceDeliverables()),
			NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
			neededResourceSetsFunc,
			FakeDeliverableAvailableTimeFunc(p, logger),
			logger,
		)

		_, err := searchBestPlans(env)
		if err != nil {
			t.Fatal(err)
		}
	}))
}

func pickOnlyAllocation(plans *sets.Set[*Plan]) *sets.Set[[]Allocation] {
	acc := sets.New(cmp2.CompareSlice[[]Allocation](Allocation.Compare))
	for _, plan := range plans.Iter() {
		allocations := make([]Allocation, 0, len(plan.Transitions))
		for _, tr := range plan.Transitions {
			allocations = append(allocations, tr.Allocation)
		}
		acc.Add(cmp2.CompareSlice[[]Allocation](Allocation.Compare), allocations)
	}
	return acc
}
