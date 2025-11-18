package fsmchecker

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/chans"
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmchecker/fsmcommon"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
)

func TestValidMaxRevision(t *testing.T) {
	testCases := map[string]struct {
		PFD                    *pfd.ValidPFD
		AtomicDeliverableTable *pfd.AtomicDeliverableTable
		FeedbackSources        *sets.Set[pfd.AtomicDeliverableID]
		Expected               []checkers.Problem
	}{
		"ng (not feedback source)": {
			PFD: pfd.NewSafePFD(
				map[pfd.AtomicProcessID]string{
					"P1": "P1",
				},
				map[pfd.AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
				},
				map[pfd.AtomicProcessID]*pfd.RelationTriple{
					"P1": {
						Inputs:         sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
						FeedbackInputs: sets.New(pfd.AtomicDeliverableID.Compare),
						Outputs:        sets.New(pfd.AtomicDeliverableID.Compare, "D2"),
					},
				},
				map[pfd.CompositeProcessID]*pairs.Pair[string, *sets.Set[pfd.AtomicProcessID]]{},
				map[pfd.CompositeDeliverableID]*pairs.Pair[string, *sets.Set[pfd.AtomicDeliverableID]]{},
			),
			AtomicDeliverableTable: &pfd.AtomicDeliverableTable{
				ExtraHeaders: []string{fsmtable.MaxRevisionHeaderEn},
				Rows: []*pfd.AtomicDeliverableRow{
					{ID: "D1", Description: "Deliverable 1", ExtraCells: []string{"1"}},
					{ID: "D2", Description: "Deliverable 2", ExtraCells: []string{""}},
				},
			},
			Expected: []checkers.Problem{
				checkers.NewProblem("malformed-max-revision", checkers.SeverityError, fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicDeliverableTable, fsmcommon.NewAtomicDeliverableID("D1"))),
			},
		},
		"ng (feedback source, empty)": {
			PFD: pfd.NewSafePFD(
				map[pfd.AtomicProcessID]string{
					"P1": "P1",
				},
				map[pfd.AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
				},
				map[pfd.AtomicProcessID]*pfd.RelationTriple{
					"P1": {
						Inputs:         sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
						FeedbackInputs: sets.New(pfd.AtomicDeliverableID.Compare, "D2"),
						Outputs:        sets.New(pfd.AtomicDeliverableID.Compare, "D2"),
					},
				},
				map[pfd.CompositeProcessID]*pairs.Pair[string, *sets.Set[pfd.AtomicProcessID]]{},
				map[pfd.CompositeDeliverableID]*pairs.Pair[string, *sets.Set[pfd.AtomicDeliverableID]]{},
			),
			AtomicDeliverableTable: &pfd.AtomicDeliverableTable{
				ExtraHeaders: []string{fsmtable.MaxRevisionHeaderEn},
				Rows: []*pfd.AtomicDeliverableRow{
					{ID: "D1", Description: "Deliverable 1", ExtraCells: []string{""}},
					{ID: "D2", Description: "Deliverable 2", ExtraCells: []string{""}},
				},
			},
			Expected: []checkers.Problem{
				checkers.NewProblem("malformed-max-revision", checkers.SeverityError, fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicDeliverableTable, fsmcommon.NewAtomicDeliverableID("D2"))),
			},
		},
		"ng (feedback source, negative)": {
			PFD: pfd.NewSafePFD(
				map[pfd.AtomicProcessID]string{
					"P1": "P1",
				},
				map[pfd.AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
				},
				map[pfd.AtomicProcessID]*pfd.RelationTriple{
					"P1": {
						Inputs:         sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
						FeedbackInputs: sets.New(pfd.AtomicDeliverableID.Compare, "D2"),
						Outputs:        sets.New(pfd.AtomicDeliverableID.Compare, "D2"),
					},
				},
				map[pfd.CompositeProcessID]*pairs.Pair[string, *sets.Set[pfd.AtomicProcessID]]{},
				map[pfd.CompositeDeliverableID]*pairs.Pair[string, *sets.Set[pfd.AtomicDeliverableID]]{},
			),
			AtomicDeliverableTable: &pfd.AtomicDeliverableTable{
				ExtraHeaders: []string{fsmtable.MaxRevisionHeaderEn},
				Rows: []*pfd.AtomicDeliverableRow{
					{ID: "D1", Description: "Deliverable 1", ExtraCells: []string{""}},
					{ID: "D2", Description: "Deliverable 2", ExtraCells: []string{"-1"}},
				},
			},
			Expected: []checkers.Problem{
				checkers.NewProblem("malformed-max-revision", checkers.SeverityError, fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicDeliverableTable, fsmcommon.NewAtomicDeliverableID("D2"))),
			},
		},
		"ok (feedback source)": {
			PFD: pfd.NewSafePFD(
				map[pfd.AtomicProcessID]string{
					"P1": "P1",
				},
				map[pfd.AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
				},
				map[pfd.AtomicProcessID]*pfd.RelationTriple{
					"P1": {
						Inputs:         sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
						FeedbackInputs: sets.New(pfd.AtomicDeliverableID.Compare, "D2"),
						Outputs:        sets.New(pfd.AtomicDeliverableID.Compare, "D2"),
					},
				},
				map[pfd.CompositeProcessID]*pairs.Pair[string, *sets.Set[pfd.AtomicProcessID]]{},
				map[pfd.CompositeDeliverableID]*pairs.Pair[string, *sets.Set[pfd.AtomicDeliverableID]]{},
			),
			AtomicDeliverableTable: &pfd.AtomicDeliverableTable{
				ExtraHeaders: []string{fsmtable.MaxRevisionHeaderEn},
				Rows: []*pfd.AtomicDeliverableRow{
					{ID: "D1", Description: "Deliverable 1", ExtraCells: []string{""}},
					{ID: "D2", Description: "Deliverable 2", ExtraCells: []string{"1"}},
				},
			},
			Expected: []checkers.Problem{},
		},
		"ok (not feedback source)": {
			PFD: pfd.NewSafePFD(
				map[pfd.AtomicProcessID]string{
					"P1": "P1",
				},
				map[pfd.AtomicDeliverableID]string{
					"D1": "D1",
					"D2": "D2",
				},
				map[pfd.AtomicProcessID]*pfd.RelationTriple{
					"P1": {
						Inputs:         sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
						FeedbackInputs: sets.New(pfd.AtomicDeliverableID.Compare),
						Outputs:        sets.New(pfd.AtomicDeliverableID.Compare, "D2"),
					},
				},
				map[pfd.CompositeProcessID]*pairs.Pair[string, *sets.Set[pfd.AtomicProcessID]]{},
				map[pfd.CompositeDeliverableID]*pairs.Pair[string, *sets.Set[pfd.AtomicDeliverableID]]{},
			),
			AtomicDeliverableTable: &pfd.AtomicDeliverableTable{
				ExtraHeaders: []string{fsmtable.MaxRevisionHeaderEn},
				Rows: []*pfd.AtomicDeliverableRow{
					{ID: "D1", Description: "Deliverable 1", ExtraCells: []string{""}},
					{ID: "D2", Description: "Deliverable 2", ExtraCells: []string{""}},
				},
			},
			Expected: []checkers.Problem{},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			m, err := fsmcommon.NewMemoized(nil, tc.AtomicDeliverableTable, nil, nil)
			if err != nil {
				t.Fatalf("fsmcommon.NewMemoized: %v", err)
			}
			ch := make(chan checkers.Problem)
			go func() {
				defer close(ch)
				tgt := fsmcommon.NewTarget(tc.PFD, nil, tc.AtomicDeliverableTable, nil, nil, nil, m, slog.New(slogtest.NewTestHandler(t)))
				if err := ValidMaxRevision.Check(tgt, ch); err != nil {
					t.Errorf("ValidMaxRevision.Check: %v", err)
				}
			}()
			got := chans.Slice(ch)
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got))
			}
		})
	}
}
