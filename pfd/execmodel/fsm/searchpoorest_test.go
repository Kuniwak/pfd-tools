package fsm

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
)

func TestSearchPoorest(t *testing.T) {
	logger := slog.New(slogtest.NewTestHandler(t))
	initVolume := Volume(1)
	initVolumeFunc := ConstInitialVolumeFunc(initVolume)

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
	neededResourceSetsFunc := NeededResourceSetsFuncByMap(map[pfd.AtomicProcessID]*sets.Set[AllocationElement]{
		"P1": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: initVolume}),
		"P2": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: initVolume}),
		"P3": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R3"), ConsumedVolume: initVolume}),
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
		AlwaysAvailableTimeFunc(),
		logger,
	)

	plans, err := SearchPoorest(1)(env)
	if err != nil {
		t.Fatal(err)
	}
	plan, ok := plans.At(0)
	if !ok {
		t.Fatal("SearchPoorest returned no plan")
	}
	if len(plan.Transitions) == 0 {
		t.Fatal("SearchPoorest returned an empty plan")
	}

	plans2, err := SearchPoorest(1)(env)
	if err != nil {
		t.Fatal(err)
	}
	plan2, _ := plans2.At(0)
	if (*Plan).Compare(plan, plan2) != 0 {
		t.Error("SearchPoorest is not deterministic across runs")
	}
}

func TestSearchPoorestCompletesOnWideBranching(t *testing.T) {
	const numProcesses = 15
	const numResources = 10

	logger := slog.New(slogtest.NewTestHandler(t))
	initVolume := Volume(1)
	initVolumeFunc := ConstInitialVolumeFunc(initVolume)

	nodes := []*pfd.Node{{ID: "D0", Type: pfd.NodeTypeAtomicDeliverable}}
	edges := []*pfd.Edge{}
	needed := map[pfd.AtomicProcessID]*sets.Set[AllocationElement]{}
	resources := sets.New[ResourceID](ResourceID.Compare)
	for r := 1; r <= numResources; r++ {
		resources.Add(ResourceID.Compare, ResourceID(fmt.Sprintf("R%d", r)))
	}
	for i := 1; i <= numProcesses; i++ {
		pid := pfd.NodeID(fmt.Sprintf("P%d", i))
		did := pfd.NodeID(fmt.Sprintf("D%d", i))
		nodes = append(nodes,
			&pfd.Node{ID: pid, Type: pfd.NodeTypeAtomicProcess},
			&pfd.Node{ID: did, Type: pfd.NodeTypeAtomicDeliverable},
		)
		edges = append(edges,
			&pfd.Edge{Source: "D0", Target: pid},
			&pfd.Edge{Source: pid, Target: did},
		)
		alts := sets.New[AllocationElement](AllocationElement.Compare)
		for r := 1; r <= numResources; r++ {
			alts.Add(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, ResourceID(fmt.Sprintf("R%d", r))), ConsumedVolume: initVolume})
		}
		needed[pfd.AtomicProcessID(pid)] = alts
	}
	p := newSafePFDByUnsafePFD(&pfd.PFD{
		Nodes: sets.New((*pfd.Node).Compare, nodes...),
		Edges: sets.New((*pfd.Edge).Compare, edges...),
	})
	neededResourceSetsFunc := NeededResourceSetsFuncByMap(needed)
	env := NewEnv(
		p,
		resources,
		NewAvailableAllocationsFunc(neededResourceSetsFunc),
		initVolumeFunc,
		ExponentialReworkVolumeFunc(0.5, initVolumeFunc),
		ConstMaxRevisionMap(3, p.FeedbackSourceDeliverables()),
		NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
		neededResourceSetsFunc,
		AlwaysAvailableTimeFunc(),
		logger,
	)

	plans, err := SearchPoorest(1)(env)
	if err != nil {
		t.Fatal(err)
	}
	plan, ok := plans.At(0)
	if !ok || len(plan.Transitions) == 0 {
		t.Fatal("SearchPoorest did not produce a plan on wide branching")
	}
}
