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

func TestMalformedID(t *testing.T) {
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
		"ok": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable, Description: "リスク一覧"},
					&pfd.Node{ID: "P1", Type: pfd.NodeTypeAtomicProcess, Description: "検討"},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{Source: "P1", Target: "D1"},
				),
			},
			Expected: []checkers.Problem{},
		},
		"ok_dotted": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D3.1", Type: pfd.NodeTypeAtomicDeliverable, Description: "Bug tickets"},
				),
				Edges: sets.New((*pfd.Edge).Compare),
			},
			Expected: []checkers.Problem{},
		},
		"ng_colon_title": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "検討", Type: pfd.NodeTypeAtomicProcess, Description: "リスク一覧"},
				),
				Edges: sets.New((*pfd.Edge).Compare),
			},
			Expected: []checkers.Problem{checkers.NewProblem("malformed-id", checkers.SeverityWarning, pfdcommon.NewLocation(pfdcommon.LocationTypePFD, "検討"))},
		},
		"ng_p_prefix": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "P設計", Type: pfd.NodeTypeAtomicProcess, Description: "手順"},
				),
				Edges: sets.New((*pfd.Edge).Compare),
			},
			Expected: []checkers.Problem{checkers.NewProblem("malformed-id", checkers.SeverityWarning, pfdcommon.NewLocation(pfdcommon.LocationTypePFD, "P設計"))},
		},
		"unnumbered_skipped": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "(検討: リスク一覧)", Type: pfd.NodeTypeAtomicProcess},
				),
				Edges: sets.New((*pfd.Edge).Compare),
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
				if err := MalformedID.Check(tgt, ch); err != nil {
					t.Errorf("MalformedID.Check: %v", err)
				}
			}()
			got := chans.Slice(ch)
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Errorf("got %v, expected %v", got, tc.Expected)
			}
		})
	}
}
