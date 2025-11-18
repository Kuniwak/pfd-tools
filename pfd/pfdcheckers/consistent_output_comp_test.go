package pfdcheckers

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/chans"
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
)

func TestConsistentOutputComp(t *testing.T) {
	testCases := map[string]struct {
		PFD      *pfd.PFD
		Expected []checkers.Problem
	}{
		"empty": {
			PFD: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare),
				Edges: sets.New((*pfd.Edge).Compare),
			},
			Expected: []checkers.Problem{},
		},
		"ok-missing": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P1", Type: pfd.NodeTypeAtomicProcess},
					&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P2", Type: pfd.NodeTypeAtomicProcess},
					&pfd.Node{ID: "D3", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "D4", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P3", Type: pfd.NodeTypeCompositeProcess},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{Source: "D1", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D2"},
					&pfd.Edge{Source: "D2", Target: "P2"},
					&pfd.Edge{Source: "P2", Target: "D3"},
					&pfd.Edge{Source: "P2", Target: "D4"},
					&pfd.Edge{Source: "D1", Target: "P3"},
					&pfd.Edge{Source: "P3", Target: "D3"},
				),
				ProcessComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{
					"P3": sets.New(pfd.NodeID.Compare, "P1", "P2"),
				},
			},
			Expected: []checkers.Problem{},
		},
		"ng-extra": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P1", Type: pfd.NodeTypeAtomicProcess},
					&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P2", Type: pfd.NodeTypeAtomicProcess},
					&pfd.Node{ID: "D3", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P3", Type: pfd.NodeTypeAtomicProcess},
					&pfd.Node{ID: "D4", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P4", Type: pfd.NodeTypeCompositeProcess},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{Source: "D1", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D2"},
					&pfd.Edge{Source: "D2", Target: "P2"},
					&pfd.Edge{Source: "P2", Target: "D3"},
					&pfd.Edge{Source: "P3", Target: "D4"},
					&pfd.Edge{Source: "D1", Target: "P4"},
					&pfd.Edge{Source: "P4", Target: "D3"},
					&pfd.Edge{Source: "P4", Target: "D4"},
				),
				ProcessComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{
					"P4": sets.New(pfd.NodeID.Compare, "P1", "P2"),
				},
			},
			Expected: []checkers.Problem{checkers.NewProblem("consistent-output-comp", checkers.SeverityError, pfdcommon.NewLocation(pfdcommon.LocationTypePFD, "D4", "P4"))},
		},
		"ok-extra": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P1", Type: pfd.NodeTypeAtomicProcess},
					&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P2", Type: pfd.NodeTypeAtomicProcess},
					&pfd.Node{ID: "D3", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P3", Type: pfd.NodeTypeAtomicProcess},
					&pfd.Node{ID: "D4", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P4", Type: pfd.NodeTypeCompositeProcess},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{Source: "D1", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D2"},
					&pfd.Edge{Source: "D2", Target: "P2"},
					&pfd.Edge{Source: "P2", Target: "D3"},
					&pfd.Edge{Source: "P3", Target: "D4"},
					&pfd.Edge{Source: "D1", Target: "P4"},
					&pfd.Edge{Source: "P4", Target: "D3"},
				),
				ProcessComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{
					"P4": sets.New(pfd.NodeID.Compare, "P1", "P2"),
				},
			},
			Expected: []checkers.Problem{},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			m := pfdcommon.NewMemoized(tc.PFD, slog.New(slogtest.NewTestHandler(t)))
			ch := make(chan checkers.Problem)
			go func() {
				defer close(ch)
				tgt := pfdcommon.NewTarget(tc.PFD, nil, nil, nil, nil, m)
				if err := ConsistentOutputComp.Check(tgt, ch); err != nil {
					t.Errorf("ConsistentOutputComp.Check: %v", err)
				}
			}()
			got := chans.Slice(ch)
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Errorf("got %v, expected %v", got, tc.Expected)
			}
		})
	}
}
