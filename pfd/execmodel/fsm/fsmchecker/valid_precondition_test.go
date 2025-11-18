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

func TestValidPrecondition(t *testing.T) {
	testCases := map[string]struct {
		PFD                *pfd.PFD
		AtomicProcessTable *pfd.AtomicProcessTable
		Want               []checkers.Problem
	}{
		"ng (syntax error)": {
			PFD: pfd.PresetButterflyLoop,
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.PreconditionColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Process 1", ExtraCells: []string{`#ERROR`}},
					{ID: "P2", Description: "Process 1", ExtraCells: []string{``}},
					{ID: "P3", Description: "Process 1", ExtraCells: []string{``}},
				},
			},
			Want: []checkers.Problem{
				checkers.NewProblem(
					"malformed-precondition",
					checkers.SeverityError,
					fsmcommon.NewLocation(
						fsmcommon.LocationTypeAtomicProcessTable,
						fsmcommon.NewAtomicProcessID("P1"),
					),
				),
			},
		},
		"ng (cyclic executable reference)": {
			PFD: pfd.PresetTripleBranch,
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.PreconditionColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Process 3", ExtraCells: []string{`!\exec(P2)`}},
					{ID: "P2", Description: "Process 3", ExtraCells: []string{`!\exec(P1)`}},
					{ID: "P3", Description: "Process 3", ExtraCells: []string{``}},
				},
			},
			Want: []checkers.Problem{
				checkers.NewProblem(
					"precondition-cyclic-executable-reference",
					checkers.SeverityError,
					fsmcommon.NewLocation(
						fsmcommon.LocationTypeAtomicProcessTable,
						fsmcommon.NewAtomicProcessID("P1"),
						fsmcommon.NewAtomicProcessID("P2"),
					),
				),
			},
		},
		"ng (not feedback source, deliverable)": {
			PFD: pfd.PresetButterflyLoop,
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.PreconditionColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Process 1", ExtraCells: []string{`\complete(D2)`}},
					{ID: "P2", Description: "Process 1", ExtraCells: []string{``}},
					{ID: "P3", Description: "Process 1", ExtraCells: []string{``}},
				},
			},
			Want: []checkers.Problem{
				checkers.NewProblem(
					"precondition-not-feedback-source",
					checkers.SeverityError,
					fsmcommon.NewLocation(
						fsmcommon.LocationTypeAtomicProcessTable,
						fsmcommon.NewAtomicProcessID("P1"),
						fsmcommon.NewAtomicDeliverableID("D2"),
					),
				),
			},
		},
		"ng (not feedback source, process)": {
			PFD: pfd.PresetButterflyLoop,
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.PreconditionColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Process 1", ExtraCells: []string{`\complete(P2)`}},
					{ID: "P2", Description: "Process 1", ExtraCells: []string{``}},
					{ID: "P3", Description: "Process 1", ExtraCells: []string{``}},
				},
			},
			Want: []checkers.Problem{
				checkers.NewProblem(
					"precondition-not-feedback-source",
					checkers.SeverityError,
					fsmcommon.NewLocation(
						fsmcommon.LocationTypeAtomicProcessTable,
						fsmcommon.NewAtomicProcessID("P1"),
						fsmcommon.NewAtomicDeliverableID("P2"),
					),
				),
			},
		},
		"ng (reachable feedback source)": {
			PFD: pfd.PresetButterflyLoop,
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.PreconditionColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Process 1", ExtraCells: []string{`\complete(D4)`}},
					{ID: "P2", Description: "Process 1", ExtraCells: []string{``}},
					{ID: "P3", Description: "Process 1", ExtraCells: []string{``}},
				},
			},
			Want: []checkers.Problem{
				checkers.NewProblem(
					"precondition-reachable-feedback-source",
					checkers.SeverityError,
					fsmcommon.NewLocation(
						fsmcommon.LocationTypeAtomicProcessTable,
						fsmcommon.NewAtomicProcessID("P1"),
						fsmcommon.NewAtomicDeliverableID("D4"),
					),
				),
			},
		},
		"ng (reachable executable target)": {
			PFD: pfd.PresetButterflyLoop,
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.PreconditionColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Process 1", ExtraCells: []string{`!\exec(P2)`}},
					{ID: "P2", Description: "Process 1", ExtraCells: []string{``}},
					{ID: "P3", Description: "Process 1", ExtraCells: []string{``}},
				},
			},
			Want: []checkers.Problem{
				checkers.NewProblem(
					"precondition-reachable-executable-target",
					checkers.SeverityError,
					fsmcommon.NewLocation(
						fsmcommon.LocationTypeAtomicProcessTable,
						fsmcommon.NewAtomicProcessID("P1"),
						fsmcommon.NewAtomicProcessID("P2"),
					),
				),
			},
		},
		"ok": {
			PFD: pfd.PresetButterflyLoop,
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.PreconditionColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Process 1", ExtraCells: []string{``}},
					{ID: "P2", Description: "Process 1", ExtraCells: []string{``}},
					{ID: "P3", Description: "Process 1", ExtraCells: []string{`\complete(D4) && !\exec(P1) && !\exec(P2)`}},
				},
			},
			Want: []checkers.Problem{},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			p, err := pfd.NewSafePFDByUnsafePFD(testCase.PFD)
			if err != nil {
				t.Fatalf("pfd.NewSafePFDByUnsafePFD: %v", err)
			}
			m, err := fsmcommon.NewMemoized(testCase.AtomicProcessTable, nil, nil, nil)
			if err != nil {
				t.Fatalf("fsmcommon.NewMemoized: %v", err)
			}
			ch := make(chan checkers.Problem)
			go func() {
				defer close(ch)
				tgt := fsmcommon.NewTarget(p, testCase.AtomicProcessTable, nil, nil, nil, nil, m, slog.New(slogtest.NewTestHandler(t)))
				if err := ValidPrecondition.Check(tgt, ch); err != nil {
					t.Errorf("ValidPrecondition.Check: %v", err)
				}
			}()
			got := chans.Slice(ch)
			if !reflect.DeepEqual(got, testCase.Want) {
				t.Error(cmp.Diff(testCase.Want, got))
			}
		})
	}
}
