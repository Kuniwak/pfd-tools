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

func TestConsistentMTable(t *testing.T) {
	tests := map[string]struct {
		AtomicProcessTable *pfd.AtomicProcessTable
		MilestoneTable     *fsmtable.MilestoneTable
		Want               []checkers.Problem
	}{
		"ng (missing successor)": {
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.MilestoneColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Atomic Process 1", ExtraCells: []string{"M1"}},
					{ID: "P2", Description: "Atomic Process 2", ExtraCells: []string{"M2"}},
				},
			},
			MilestoneTable: &fsmtable.MilestoneTable{
				ExtraHeaders: []string{},
				Rows: []*fsmtable.MilestoneTableRow{
					{MilestoneID: "M1", Description: "Milestone 1", Successors: "M2,M3", ExtraCells: []string{}},
					{MilestoneID: "M2", Description: "Milestone 2", Successors: "", ExtraCells: []string{}},
				},
			},
			Want: []checkers.Problem{
				checkers.NewProblem(
					"missing-m-table",
					checkers.SeverityError,
					fsmcommon.NewLocation(
						fsmcommon.LocationTypeMilestoneTable,
						fsmcommon.NewMilestoneID("M3"),
					),
				),
			},
		},
		"ng (missing)": {
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.MilestoneColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Atomic Process 1", ExtraCells: []string{"M1"}},
				},
			},
			MilestoneTable: &fsmtable.MilestoneTable{
				ExtraHeaders: []string{},
				Rows:         []*fsmtable.MilestoneTableRow{},
			},
			Want: []checkers.Problem{
				checkers.NewProblem("missing-m-table", checkers.SeverityError, fsmcommon.NewLocation(fsmcommon.LocationTypeMilestoneTable, fsmcommon.NewMilestoneID("M1"))),
			},
		},
		"ng (extra)": {
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.MilestoneColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Atomic Process 1", ExtraCells: []string{"M1"}},
				},
			},
			MilestoneTable: &fsmtable.MilestoneTable{
				ExtraHeaders: []string{},
				Rows: []*fsmtable.MilestoneTableRow{
					{MilestoneID: "M1", Description: "Milestone 1", Successors: "M2", ExtraCells: []string{}},
					{MilestoneID: "M2", Description: "Milestone 2", Successors: "", ExtraCells: []string{}},
				},
			},
			Want: []checkers.Problem{
				checkers.NewProblem("extra-m-table", checkers.SeverityWarning, fsmcommon.NewLocation(fsmcommon.LocationTypeMilestoneTable, fsmcommon.NewMilestoneID("M2"))),
			},
		},
		"ok": {
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.MilestoneColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Atomic Process 1", ExtraCells: []string{"M1"}},
				},
			},
			MilestoneTable: &fsmtable.MilestoneTable{
				ExtraHeaders: []string{},
				Rows: []*fsmtable.MilestoneTableRow{
					{MilestoneID: "M1", Description: "Milestone 1", Successors: "", ExtraCells: []string{}},
				},
			},
			Want: []checkers.Problem{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			m, err := fsmcommon.NewMemoized(tt.AtomicProcessTable, nil, nil, tt.MilestoneTable)
			if err != nil {
				t.Fatalf("fsmcommon.NewMemoized: %v", err)
			}
			tgt := fsmcommon.NewTarget(nil, tt.AtomicProcessTable, nil, nil, tt.MilestoneTable, nil, m, logger)
			if !ConsistentMTable.AvailableIfFunc(tgt) {
				t.Fatalf("ConsistentMTable.AvailableIfFunc: %v", ConsistentMTable.AvailableIfFunc(tgt))
			}

			ch := make(chan checkers.Problem)
			go func() {
				defer close(ch)
				if err := ConsistentMTable.Check(tgt, ch); err != nil {
					t.Errorf("ConsistentMTable.Check: %v", err)
				}
			}()
			got := chans.Slice(ch)
			if !reflect.DeepEqual(got, tt.Want) {
				t.Error(cmp.Diff(tt.Want, got))
			}
		})
	}
}
