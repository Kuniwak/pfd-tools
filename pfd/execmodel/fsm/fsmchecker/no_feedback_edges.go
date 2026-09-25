package fsmchecker

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmchecker/fsmcommon"
)

var NoFeedbackEdges = checkers.AtomicChecker[*fsmcommon.Target]{
	ID: "no-feedback-edges",
	AvailableIfFunc: func(t *fsmcommon.Target) bool {
		return t.Model.Feedback == execmodel.FeedbackModeDisabled
	},
	CheckFunc: func(t *fsmcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "feedback-edge-not-available"
		for _, e := range t.PFD.FeedbackEdges() {
			ch <- checkers.NewProblem(problemID, checkers.SeverityError, fsmcommon.NewLocations(fsmcommon.NewLocation(
				fsmcommon.LocationTypePFD,
				fsmcommon.NewAtomicDeliverableID(e.AtomicDeliverable),
				fsmcommon.NewAtomicProcessID(e.AtomicProcess),
			))...)
		}
		return nil
	},
}
