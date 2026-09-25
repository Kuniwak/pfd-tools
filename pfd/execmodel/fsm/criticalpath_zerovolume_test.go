package fsm

import (
	"log/slog"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
)

func TestCriticalPathInfoWithZeroVolumeProcess(t *testing.T) {
	logger := slog.New(slogtest.NewTestHandler(t))

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
	initVolumeFunc := func(ap pfd.AtomicProcessID) Volume {
		if ap == "P1" {
			return 0
		}
		return 2
	}
	neededResourceSetsFunc := NeededResourceSetsFuncByMap(map[pfd.AtomicProcessID]*sets.Set[AllocationElement]{
		"P1": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1}),
		"P2": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1}),
	})
	env := NewEnv(
		p,
		sets.New(ResourceID.Compare, "R1"),
		NewAvailableAllocationsFunc(neededResourceSetsFunc),
		initVolumeFunc,
		ExponentialReworkVolumeFunc(0, initVolumeFunc),
		ConstMaxRevisionMap(1, p.FeedbackSourceDeliverables()),
		NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
		neededResourceSetsFunc,
		AlwaysAvailableTimeFunc(),
		logger,
	)

	searchFuncs := map[string]SearchWithPrefixFunc{
		"poorest": SearchPoorestWithPrefix(1),
		"poor":    SearchFastestWithPrefix(1),
		"best":    SearchBestPlansWithPrefix(),
	}
	for name, searchFunc := range searchFuncs {
		t.Run(name, func(t *testing.T) {
			info, err := NewCriticalPathInfoFunc(searchFunc)(env)
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := info["P1"]; !ok {
				t.Error("no critical path info for zero-volume process P1")
			}
		})
	}
}
