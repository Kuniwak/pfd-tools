package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
)

var ImplicitAtomicComposite = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "implicit-atomic-composite",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "implicit-atomic-composite"

		for _, id := range t.PFD.ImplicitAtomicProcesses.Iter() {
			ch <- checkers.NewProblem(problemID, checkers.SeverityWarning, pfdcommon.NewLocation(pfdcommon.LocationTypePFD, id))
		}
		return nil
	},
}
