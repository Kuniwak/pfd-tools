package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
)

var NoDesc = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "no-desc",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "no-desc"
		for _, node := range t.PFD.Nodes.Iter() {
			if node.Description == "" {
				ch <- checkers.NewProblem(problemID, checkers.SeverityWarning, pfdcommon.NewLocation(pfdcommon.LocationTypePFD, node.ID))
			}
		}
		return nil
	},
}
