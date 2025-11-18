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

func TestNoD2D(t *testing.T) {
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
		"ng_no_fbs": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{Source: "D1", Target: "D2"},
				),
			},
			Expected: []checkers.Problem{checkers.NewProblem("no-d2d", checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypePFD, "D1", "D2"))...)},
		},
		"ok_no_fbs": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "P1", Type: pfd.NodeTypeAtomicProcess},
					&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{Source: "D1", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D2"},
				),
			},
			Expected: []checkers.Problem{},
		},
		"ng_fbs": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P1", Type: pfd.NodeTypeAtomicProcess},
					&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{Source: "D1", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D2"},
					&pfd.Edge{Source: "D2", Target: "D1", IsFeedback: true},
				),
			},
			Expected: []checkers.Problem{checkers.NewProblem("no-d2d", checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypePFD, "D2", "D1"))...)},
		},
		"ok_fbs": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P1", Type: pfd.NodeTypeAtomicProcess},
					&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{Source: "D1", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D2"},
					&pfd.Edge{Source: "D1", Target: "P1", IsFeedback: true},
				),
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
				if err := NoD2D.Check(tgt, ch); err != nil {
					t.Errorf("NoD2D.Check: %v", err)
				}
			}()
			got := chans.Slice(ch)
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Errorf("got %v, expected %v", got, tc.Expected)
			}
		})
	}
}
