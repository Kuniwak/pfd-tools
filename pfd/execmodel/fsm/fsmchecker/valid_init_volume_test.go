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
)

func TestValidInitVolume(t *testing.T) {
	testCases := map[string]struct {
		AtomicProcessTable *pfd.AtomicProcessTable
		Expected           []checkers.Problem
	}{
		"ng": {
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.InitialVolumeColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Atomic Process 1", ExtraCells: []string{"-1"}},
				},
			},
			Expected: []checkers.Problem{
				checkers.NewProblem("valid-init-volume", checkers.SeverityError, fsmcommon.NewLocations(fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicProcessTable, fsmcommon.NewAtomicProcessID("P1")))...),
			},
		},
		"ok": {
			AtomicProcessTable: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{fsmtable.InitialVolumeColumnHeaderEn},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "Atomic Process 1", ExtraCells: []string{"0"}},
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
				if err := ValidInitVolume.Check(tgt, ch); err != nil {
					t.Errorf("ValidInitVolume.Check: %v", err)
				}
			}()
			got := chans.Slice(ch)
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Errorf("got %v, expected %v", got, tc.Expected)
			}
		})
	}
}
