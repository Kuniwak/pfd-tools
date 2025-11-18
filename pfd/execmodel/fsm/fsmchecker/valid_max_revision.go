package fsmchecker

import (
	"fmt"

	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmchecker/fsmcommon"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
)

var ValidMaxRevision = checkers.AtomicChecker[*fsmcommon.Target]{
	ID: "valid-max-revision",
	AvailableIfFunc: func(t *fsmcommon.Target) bool {
		return t.Memoized.HasMaxRevisionMap
	},
	CheckFunc: func(t *fsmcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "malformed-max-revision"
		for _, d := range t.PFD.AtomicDeliverables.Iter() {
			maxRevisionText, ok := t.Memoized.MaxRevisionMap[d]
			if !ok {
				panic(fmt.Sprintf("pfd.ValidPFD.AtomicDeliverableDescriptionMap: missing description for deliverable: %q", d))
			}
			isFeedbackSource := t.PFD.FeedbackSourceDeliverables().Contains(pfd.AtomicDeliverableID.Compare, d)

			if _, err := fsmtable.ValidateMaxRevision(maxRevisionText, isFeedbackSource); err != nil {
				ch <- checkers.NewProblem(problemID, checkers.SeverityError, fsmcommon.NewLocations(fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicDeliverableTable, fsmcommon.NewAtomicDeliverableID(d)))...)
				continue
			}
		}
		return nil
	},
}
