package fsm

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
)

func newSmallPoorestEnv(t *testing.T) *Env {
	t.Helper()
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
	return NewEnv(
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
}

func TestReplayPrefixWithGreedyMatchesEnumeration(t *testing.T) {
	target := pfd.AtomicProcessID("P3")

	assertReplayEquivalent := func(t *testing.T, e *Env) {
		t.Helper()
		basePlans, err := SearchPoorest(1)(e)
		if err != nil {
			t.Fatal(err)
		}
		basePlan, ok := basePlans.At(0)
		if !ok {
			t.Fatal("SearchPoorest returned no plan")
		}
		prefix := PrefixUntilProcessStart(basePlan, target)
		if len(prefix.Transitions) == 0 {
			t.Fatalf("prefix is empty; test needs a non-empty prefix (target=%s scheduled too early)", target)
		}

		enumPlan, enumState, err := ReplayPrefix(e, prefix)
		if err != nil {
			t.Fatalf("ReplayPrefix (enumeration): %v", err)
		}
		greedy := NewGreedyAvailableAllocationsFunc(e.NeededResourceSetsFunc)
		greedyPlan, greedyState, err := ReplayPrefixWith(e, prefix, greedy)
		if err != nil {
			t.Fatalf("ReplayPrefixWith (greedy): %v", err)
		}

		if (*Plan).Compare(enumPlan, greedyPlan) != 0 {
			t.Errorf("greedy replay plan differs from enumeration replay plan")
		}
		if enumState.Compare(greedyState) != 0 {
			t.Errorf("greedy replay state differs from enumeration replay state")
		}
	}

	t.Run("base env", func(t *testing.T) {
		assertReplayEquivalent(t, newSmallPoorestEnv(t))
	})

	t.Run("eLonger clone with inflated target volume", func(t *testing.T) {
		base := newSmallPoorestEnv(t)
		eLonger := base.Clone()
		eLonger.InitialVolumeFunc = func(ap pfd.AtomicProcessID) Volume {
			if ap == target {
				return base.InitialVolumeFunc(ap) + Volume(10)
			}
			return base.InitialVolumeFunc(ap)
		}
		assertReplayEquivalent(t, eLonger)
	})
}

func newWidePoorestEnv(t *testing.T, numProcesses, numResources int) *Env {
	t.Helper()
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
	return NewEnv(
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
}

func TestSearchPoorestWithPrefixCompletesOnWideBranching(t *testing.T) {
	const numProcesses = 15
	const numResources = 10
	env := newWidePoorestEnv(t, numProcesses, numResources)

	basePlans, err := SearchPoorest(1)(env)
	if err != nil {
		t.Fatal(err)
	}
	basePlan, ok := basePlans.At(0)
	if !ok || len(basePlan.Transitions) < 2 {
		t.Fatalf("expected a multi-step base plan, got %d transitions", len(basePlan.Transitions))
	}

	lastAlloc := basePlan.Transitions[len(basePlan.Transitions)-1].Allocation
	var target pfd.AtomicProcessID
	for ap := range lastAlloc {
		target = ap
		break
	}
	prefix := PrefixUntilProcessStart(basePlan, target)
	if len(prefix.Transitions) == 0 {
		t.Fatalf("prefix is empty; test needs a non-empty prefix containing the wide first step")
	}

	done := make(chan error, 1)
	go func() {
		_, err := SearchPoorestWithPrefix(1)(env, prefix)
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("SearchPoorestWithPrefix returned error: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("SearchPoorestWithPrefix did not complete on wide branching within 10s (prefix replay enumerated instead of using the greedy func)")
	}
}
