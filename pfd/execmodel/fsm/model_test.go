package fsm

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
)

func TestUnlimitedNeededResourceSetsFunc(t *testing.T) {
	f := UnlimitedNeededResourceSetsFunc()

	testCases := map[string]struct {
		Input pfd.AtomicProcessID
	}{
		"any atomic process": {Input: "P1"},
		"another one":        {Input: "P2"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := f(tc.Input)
			if actual.Len() != 1 {
				t.Fatalf("expected exactly 1 allocation element, but got %d", actual.Len())
			}
			elem := actual.Slice()[0]
			if elem.Resources.Len() != 0 {
				t.Errorf("expected no resources, but got %v", elem.Resources.Slice())
			}
			if elem.ConsumedVolume != Volume(1) {
				t.Errorf("expected consumed volume 1, but got %v", elem.ConsumedVolume)
			}
		})
	}
}

func TestNoReworkVolumeFunc(t *testing.T) {
	initialVolumeFunc := InitialVolumeByMap(map[pfd.AtomicProcessID]Volume{"P1": 3, "P2": 7})
	f := NoReworkVolumeFunc(initialVolumeFunc)

	testCases := map[string]struct {
		AtomicProcess pfd.AtomicProcessID
		NumOfRework   int
		Expected      Volume
	}{
		"first rework keeps the initial volume":  {AtomicProcess: "P1", NumOfRework: 1, Expected: 3},
		"second rework keeps the initial volume": {AtomicProcess: "P1", NumOfRework: 2, Expected: 3},
		"another atomic process":                 {AtomicProcess: "P2", NumOfRework: 1, Expected: 7},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if actual := f(tc.AtomicProcess, tc.NumOfRework); actual != tc.Expected {
				t.Errorf("expected %v, but got %v", tc.Expected, actual)
			}
		})
	}
}

func TestLeadtimeByResourceMode(t *testing.T) {

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
	initVolume := Volume(2)
	initVolumeFunc := ConstInitialVolumeFunc(initVolume)
	sharedResourceSetsFunc := NeededResourceSetsFuncByMap(map[pfd.AtomicProcessID]*sets.Set[AllocationElement]{
		"P1": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: Volume(1)},
		),
		"P2": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: Volume(1)},
		),
	})

	testCases := map[string]struct {
		AvailableResources     *sets.Set[ResourceID]
		NeededResourceSetsFunc NeededResourceSetsFunc
		Expected               execmodel.Time
	}{
		"finite resources serialize the two processes": {
			AvailableResources:     sets.New(ResourceID.Compare, "R1"),
			NeededResourceSetsFunc: sharedResourceSetsFunc,
			Expected:               4,
		},
		"infinite resources run the two processes in parallel": {
			AvailableResources:     sets.New(ResourceID.Compare),
			NeededResourceSetsFunc: UnlimitedNeededResourceSetsFunc(),
			Expected:               2,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			env := NewEnv(
				p,
				tc.AvailableResources,
				NewAvailableAllocationsFunc(tc.NeededResourceSetsFunc),
				initVolumeFunc,
				NoReworkVolumeFunc(initVolumeFunc),
				map[pfd.AtomicDeliverableID]int{},
				NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
				tc.NeededResourceSetsFunc,
				AlwaysAvailableTimeFunc(),
				logger,
			)

			plans, err := SearchBestPlans()(env)
			if err != nil {
				t.Fatal(err)
			}
			if plans.Len() == 0 {
				t.Fatal("expected at least 1 plan, but got none")
			}
			for _, plan := range plans.Iter() {
				if actual := plan.Leadtime(); actual != tc.Expected {
					t.Errorf("expected leadtime %v, but got %v", tc.Expected, actual)
				}
			}
		})
	}
}

func TestValidateNoFeedbackEdges(t *testing.T) {

	withoutFeedback := pfd.MustNewSafePFDByUnsafePFD(pfd.PresetSmallest)

	withFeedback := pfd.MustNewSafePFDByUnsafePFD(pfd.PresetSmallestLoop)

	testCases := map[string]struct {
		Input            *pfd.ValidPFD
		HasError         bool
		ExpectedMentions []string
	}{
		"pfd without feedback edges is valid": {Input: withoutFeedback},
		"pfd with a feedback edge is invalid": {
			Input:            withFeedback,
			HasError:         true,
			ExpectedMentions: []string{"D2", "P1"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := ValidateNoFeedbackEdges(tc.Input)
			if !tc.HasError {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected an error, but got none")
			}
			for _, mention := range tc.ExpectedMentions {
				if !strings.Contains(err.Error(), mention) {
					t.Errorf("expected the error to mention %q, but got: %v", mention, err)
				}
			}
		})
	}
}
