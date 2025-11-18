package fsmchecker

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmchecker/fsmcommon"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
)

var ValidAvailableTime = checkers.AtomicChecker[*fsmcommon.Target]{
	ID: "valid-available-time",
	AvailableIfFunc: func(t *fsmcommon.Target) bool {
		return t.Memoized.HasAvailableTimeMap
	},
	CheckFunc: func(t *fsmcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "valid-available-time"
		for _, d := range t.PFD.InitialDeliverables().Iter() {
			availableTimeText, ok := t.Memoized.AvailableTimeMap[d]
			if !ok {
				// NOTE: Skip because it will be reported by consistent-d-table.
				continue
			}
			if _, err := fsmtable.ValidateAvailableTime(availableTimeText); err != nil {
				ch <- checkers.NewProblem(problemID, checkers.SeverityError, fsmcommon.NewLocations(fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicDeliverableTable, fsmcommon.NewAtomicDeliverableID(d)))...)
			}
		}
		return nil
	},
}
