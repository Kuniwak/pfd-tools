package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
)

var InField = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "in-field",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "in-field"
		for _, edge := range t.PFD.Edges.Iter() {
			if _, ok := t.Memoized.NodeMap[edge.Source]; !ok {
				ch <- checkers.NewProblem(problemID, checkers.SeverityError, pfdcommon.NewLocation(pfdcommon.LocationTypePFD, edge.Source))
			}
			if _, ok := t.Memoized.NodeMap[edge.Target]; !ok {
				ch <- checkers.NewProblem(problemID, checkers.SeverityError, pfdcommon.NewLocation(pfdcommon.LocationTypePFD, edge.Target))
			}
		}
		return nil
	},
}
