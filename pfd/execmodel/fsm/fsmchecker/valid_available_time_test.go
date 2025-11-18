package fsmchecker

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/chans"
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmchecker/fsmcommon"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
)

func TestValidAvailableTime(t *testing.T) {
	testCases := map[string]struct {
		AtomicDeliverableTable *pfd.AtomicDeliverableTable
		Expected               []checkers.Problem
	}{
		"ng": {
			AtomicDeliverableTable: &pfd.AtomicDeliverableTable{
				ExtraHeaders: []string{fsmtable.AvailableTimeHeaderEn},
				Rows: []*pfd.AtomicDeliverableRow{
					{ID: "D1", Description: "Deliverable 1", ExtraCells: []string{"-1"}},
					{ID: "D2", Description: "Deliverable 2", ExtraCells: []string{"-1"}}, // not get reported because it is not initial deliverable
				},
			},
			Expected: []checkers.Problem{
				checkers.NewProblem("valid-available-time", checkers.SeverityError, fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicDeliverableTable, fsmcommon.NewAtomicDeliverableID("D1"))),
			},
		},
		"ok": {
			AtomicDeliverableTable: &pfd.AtomicDeliverableTable{
				ExtraHeaders: []string{fsmtable.AvailableTimeHeaderEn},
				Rows: []*pfd.AtomicDeliverableRow{
					{ID: "D1", Description: "Deliverable 1", ExtraCells: []string{"0"}},
					{ID: "D2", Description: "Deliverable 2", ExtraCells: []string{"-1"}}, // not get reported because it is not initial deliverable
				},
			},
			Expected: []checkers.Problem{},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			p, err := pfd.NewSafePFDByUnsafePFD(&pfd.PFD{
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
			if err != nil {
				t.Fatalf("pfd.NewSafePFDByUnsafePFD: %v", err)
			}
			m, err := fsmcommon.NewMemoized(nil, tc.AtomicDeliverableTable, nil, nil)
			if err != nil {
				t.Fatalf("fsmcommon.NewMemoized: %v", err)
			}
			ch := make(chan checkers.Problem)
			go func() {
				defer close(ch)
				tgt := fsmcommon.NewTarget(p, nil, tc.AtomicDeliverableTable, nil, nil, nil, m, slog.New(slogtest.NewTestHandler(t)))
				if err := ValidAvailableTime.Check(tgt, ch); err != nil {
					t.Errorf("ValidAvailableTime.Check: %v", err)
				}
			}()
			got := chans.Slice(ch)
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got))
			}
		})
	}
}
