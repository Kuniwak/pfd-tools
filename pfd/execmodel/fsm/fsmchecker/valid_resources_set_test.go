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

func TestValidResourcesSet(t *testing.T) {
	testCases := map[string]struct {
		AtomicProcessTable *pfd.AtomicProcessTable
		Expected           []checkers.Problem
	}{
		"ng (cannot parse)": {
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.InitialVolumeColumnHeaderEn, fsmtable.NeededResourceSetsColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Atomic Process 1", ExtraCells: []string{"1", "R1"}},
				},
			},
			Expected: []checkers.Problem{
				checkers.NewProblem("malformed-resources-set-notation", checkers.SeverityError, fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicProcessTable, fsmcommon.NewAtomicProcessID("P1"))),
			},
		},
		"ng (initial volume is not zero and empty resource set)": {
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.InitialVolumeColumnHeaderEn, fsmtable.NeededResourceSetsColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Atomic Process 1", ExtraCells: []string{"1", ""}},
				},
			},
			Expected: []checkers.Problem{
				checkers.NewProblem("empty-resources-set", checkers.SeverityError, fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicProcessTable, fsmcommon.NewAtomicProcessID("P1"))),
			},
		},
		"ng (consumed volume is zero)": {
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.InitialVolumeColumnHeaderEn, fsmtable.NeededResourceSetsColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Atomic Process 1", ExtraCells: []string{"1", "R1:0"}},
				},
			},
			Expected: []checkers.Problem{
				checkers.NewProblem("zero-consumed-volume", checkers.SeverityError, fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicProcessTable, fsmcommon.NewAtomicProcessID("P1"))),
			},
		},
		"ok": {
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.InitialVolumeColumnHeaderEn, fsmtable.NeededResourceSetsColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Atomic Process 1", ExtraCells: []string{"1", "R1:1;R2:1"}},
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
			m, err := fsmcommon.NewMemoized(tc.AtomicProcessTable, nil, nil, nil)
			if err != nil {
				t.Fatalf("fsmcommon.NewMemoized: %v", err)
			}
			ch := make(chan checkers.Problem)
			go func() {
				defer close(ch)
				tgt := fsmcommon.NewTarget(p, tc.AtomicProcessTable, nil, nil, nil, nil, m, slog.New(slogtest.NewTestHandler(t)))
				if err := ValidResourcesSet.Check(tgt, ch); err != nil {
					t.Errorf("ValidResourcesSet.Check: %v", err)
				}
			}()
			got := chans.Slice(ch)
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got))
			}
		})
	}
}
