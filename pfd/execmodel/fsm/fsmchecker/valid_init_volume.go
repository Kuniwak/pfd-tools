package fsmchecker

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmchecker/fsmcommon"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
)

var ValidInitVolume = checkers.AtomicChecker[*fsmcommon.Target]{
	ID: "valid-init-volume",
	AvailableIfFunc: func(t *fsmcommon.Target) bool {
		return t.Memoized.HasInitialVolumeMap
	},
	CheckFunc: func(t *fsmcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "valid-init-volume"
		for ap, initVolumeText := range t.Memoized.InitialVolumeMap {
			if _, err := fsmtable.ValidateInitialVolume(initVolumeText); err != nil {
				ch <- checkers.NewProblem(problemID, checkers.SeverityError, fsmcommon.NewLocations(fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicProcessTable, fsmcommon.NewAtomicProcessID(ap)))...)
			}
		}
		return nil
	},
}
