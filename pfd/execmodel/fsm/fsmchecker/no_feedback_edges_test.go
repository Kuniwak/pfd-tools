package fsmchecker

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/chans"
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmchecker/fsmcommon"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
)

func TestNoFeedbackEdges(t *testing.T) {

	withoutFeedback := pfd.NewSafePFD(
		map[pfd.AtomicProcessID]string{"P1": "P1"},
		map[pfd.AtomicDeliverableID]string{"D1": "D1", "D2": "D2"},
		map[pfd.AtomicProcessID]*pfd.RelationTriple{
			"P1": {
				Inputs:         sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
				FeedbackInputs: sets.New(pfd.AtomicDeliverableID.Compare),
				Outputs:        sets.New(pfd.AtomicDeliverableID.Compare, "D2"),
			},
		},
		map[pfd.CompositeProcessID]*pairs.Pair[string, *sets.Set[pfd.AtomicProcessID]]{},
		map[pfd.CompositeDeliverableID]*pairs.Pair[string, *sets.Set[pfd.AtomicDeliverableID]]{},
	)

	withFeedback := pfd.NewSafePFD(
		map[pfd.AtomicProcessID]string{"P1": "P1"},
		map[pfd.AtomicDeliverableID]string{"D1": "D1", "D2": "D2"},
		map[pfd.AtomicProcessID]*pfd.RelationTriple{
			"P1": {
				Inputs:         sets.New(pfd.AtomicDeliverableID.Compare, "D1"),
				FeedbackInputs: sets.New(pfd.AtomicDeliverableID.Compare, "D2"),
				Outputs:        sets.New(pfd.AtomicDeliverableID.Compare, "D2"),
			},
		},
		map[pfd.CompositeProcessID]*pairs.Pair[string, *sets.Set[pfd.AtomicProcessID]]{},
		map[pfd.CompositeDeliverableID]*pairs.Pair[string, *sets.Set[pfd.AtomicDeliverableID]]{},
	)

	testCases := map[string]struct {
		PFD      *pfd.ValidPFD
		Model    execmodel.Model
		Expected []checkers.Problem
	}{
		"ok (no feedback edge, feedback disabled)": {
			PFD:      withoutFeedback,
			Model:    execmodel.Model{Resource: execmodel.ResourceModeInfinite, Feedback: execmodel.FeedbackModeDisabled},
			Expected: []checkers.Problem{},
		},
		"ok (feedback edge, feedback enabled)": {
			PFD:      withFeedback,
			Model:    execmodel.Model{Resource: execmodel.ResourceModeInfinite, Feedback: execmodel.FeedbackModeEnabled},
			Expected: []checkers.Problem{},
		},
		"ng (feedback edge, feedback disabled)": {
			PFD:   withFeedback,
			Model: execmodel.Model{Resource: execmodel.ResourceModeInfinite, Feedback: execmodel.FeedbackModeDisabled},
			Expected: []checkers.Problem{
				checkers.NewProblem("feedback-edge-not-available", checkers.SeverityError, fsmcommon.NewLocation(
					fsmcommon.LocationTypePFD,
					fsmcommon.NewAtomicDeliverableID("D2"),
					fsmcommon.NewAtomicProcessID("P1"),
				)),
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			m, err := fsmcommon.NewMemoized(nil, nil, nil, nil)
			if err != nil {
				t.Fatalf("fsmcommon.NewMemoized: %v", err)
			}
			ch := make(chan checkers.Problem)
			go func() {
				defer close(ch)
				tgt := &fsmcommon.Target{PFD: tc.PFD, Model: tc.Model, Memoized: m, Logger: slog.New(slogtest.NewTestHandler(t))}
				if err := NoFeedbackEdges.Check(tgt, ch); err != nil {
					t.Errorf("NoFeedbackEdges.Check: %v", err)
				}
			}()
			got := chans.Slice(ch)
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got))
			}
		})
	}
}
