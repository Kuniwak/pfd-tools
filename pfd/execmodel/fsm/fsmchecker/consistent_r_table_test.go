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

func TestConsistentResourceTable(t *testing.T) {
	testCases := map[string]struct {
		AtomicProcessTable *pfd.AtomicProcessTable
		ResourceTable      *fsmtable.ResourceTable
		Expected           []checkers.Problem
	}{
		"ng": {
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.InitialVolumeColumnHeaderEn, fsmtable.NeededResourceSetsColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Atomic Process 1", ExtraCells: []string{"0", "R1,R2:1"}},
				},
			},
			ResourceTable: &fsmtable.ResourceTable{
				ExtraHeaders: []string{},
				Rows: []*fsmtable.ResourceTableRow{
					{ID: "R2", Description: "Resource 2", ExtraCells: []string{}},
					{ID: "R3", Description: "Resource 3", ExtraCells: []string{}},
				},
			},
			Expected: []checkers.Problem{
				checkers.NewProblem("missing-r-table", checkers.SeverityError, fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicProcessTable, fsmcommon.NewResourceID("R1"))),
				checkers.NewProblem("extra-r-table", checkers.SeverityWarning, fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicProcessTable, fsmcommon.NewResourceID("R3"))),
			},
		},
		"ok": {
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.InitialVolumeColumnHeaderEn, fsmtable.NeededResourceSetsColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Atomic Process 1", ExtraCells: []string{"0", "R1,R2:1;R3:1"}},
				},
			},
			ResourceTable: &fsmtable.ResourceTable{
				ExtraHeaders: []string{},
				Rows: []*fsmtable.ResourceTableRow{
					{ID: "R1", Description: "Resource 1", ExtraCells: []string{}},
					{ID: "R2", Description: "Resource 2", ExtraCells: []string{}},
					{ID: "R3", Description: "Resource 3", ExtraCells: []string{}},
				},
			},
			Expected: []checkers.Problem{},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			p, err := pfd.NewSafePFDByUnsafePFD(&pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare,
					&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
					&pfd.Node{ID: "P1", Type: pfd.NodeTypeAtomicProcess},
				),
				Edges: sets.New((*pfd.Edge).Compare,
					&pfd.Edge{Source: "D1", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D2"},
				),
			})
			if err != nil {
				t.Fatalf("fsmchecker.TestConsistentResourceTable: %v", err)
			}
			dt := &pfd.AtomicDeliverableTable{
				ExtraHeaders: []string{},
				Rows: []*pfd.AtomicDeliverableRow{
					{ID: "D1", Description: "Deliverable 1", ExtraCells: []string{}},
				},
			}
			m, err := fsmcommon.NewMemoized(tc.AtomicProcessTable, dt, tc.ResourceTable, nil)
			if err != nil {
				t.Fatalf("fsmcommon.NewMemoized: %v", err)
			}

			ch := make(chan checkers.Problem)
			go func() {
				defer close(ch)
				tgt := fsmcommon.NewTarget(p, tc.AtomicProcessTable, nil, tc.ResourceTable, nil, nil, m, logger)
				if err := ConsistentResourceTable.Check(tgt, ch); err != nil {
					t.Errorf("ConsistentResourceTable.Check: %v", err)
				}
			}()
			got := chans.Slice(ch)
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got))
			}
		})
	}
}
