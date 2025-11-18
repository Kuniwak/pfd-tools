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
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
)

func TestConsistentGTable(t *testing.T) {
	tests := map[string]struct {
		AtomicProcessTable *pfd.AtomicProcessTable
		GroupTable         *fsmtable.GroupTable
		Want               []checkers.Problem
	}{
		"ng (missing)": {
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.GroupColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Atomic Process 1", ExtraCells: []string{"G1,G2"}},
				},
			},
			GroupTable: &fsmtable.GroupTable{
				ExtraHeaders: []string{},
				Rows: []*fsmtable.GroupTableRow{
					{ID: "G1", Description: "Group 1", ExtraCells: []string{}},
				},
			},
			Want: []checkers.Problem{
				checkers.NewProblem("missing-g-table", checkers.SeverityError, fsmcommon.NewLocation(fsmcommon.LocationTypeGroupTable, fsmcommon.NewGroupID("G2"))),
			},
		},
		"ng (extra)": {
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.GroupColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Atomic Process 1", ExtraCells: []string{"G1"}},
				},
			},
			GroupTable: &fsmtable.GroupTable{
				ExtraHeaders: []string{},
				Rows: []*fsmtable.GroupTableRow{
					{ID: "G1", Description: "Group 1", ExtraCells: []string{}},
					{ID: "G2", Description: "Group 2", ExtraCells: []string{}},
				},
			},
			Want: []checkers.Problem{
				checkers.NewProblem("extra-g-table", checkers.SeverityWarning, fsmcommon.NewLocation(fsmcommon.LocationTypeGroupTable, fsmcommon.NewGroupID("G2"))),
			},
		},
		"ok (not empty)": {
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.GroupColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Atomic Process 1", ExtraCells: []string{"G1,G2"}},
				},
			},
			GroupTable: &fsmtable.GroupTable{
				ExtraHeaders: []string{},
				Rows: []*fsmtable.GroupTableRow{
					{ID: "G1", Description: "Group 1", ExtraCells: []string{}},
					{ID: "G2", Description: "Group 2", ExtraCells: []string{}},
				},
			},
			Want: []checkers.Problem{},
		},
		"ok (empty)": {
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.GroupColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Atomic Process 1", ExtraCells: []string{""}},
				},
			},
			GroupTable: &fsmtable.GroupTable{
				ExtraHeaders: []string{},
				Rows:         []*fsmtable.GroupTableRow{},
			},
			Want: []checkers.Problem{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			m, err := fsmcommon.NewMemoized(tt.AtomicProcessTable, nil, nil, nil)
			if err != nil {
				t.Fatalf("fsmcommon.NewMemoized: %v", err)
			}

			tgt := fsmcommon.NewTarget(nil, tt.AtomicProcessTable, nil, nil, nil, tt.GroupTable, m, logger)
			if !ConsistentGTable.AvailableIfFunc(tgt) {
				t.Fatalf("ConsistentGTable.AvailableIfFunc: %v", ConsistentGTable.AvailableIfFunc(tgt))
			}

			ch := make(chan checkers.Problem)
			go func() {
				defer close(ch)

				if err := ConsistentGTable.Check(tgt, ch); err != nil {
					t.Errorf("ConsistentGTable.Check: %v", err)
				}
			}()
			got := chans.Slice(ch)
			if !reflect.DeepEqual(got, tt.Want) {
				t.Error(cmp.Diff(tt.Want, got))
			}
		})
	}
}
