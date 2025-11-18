package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
)

var AcyclicExceptFB = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "acyclic-except-fb",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "acyclic-except-fb"

		cycles := t.Memoized.GraphExceptFB.Cycles()
		if cycles.Len() > 0 {
			for _, cycle := range cycles.Iter() {
				ids := make([]pfd.NodeID, len(cycle))
				for i, id := range cycle {
					ids[i] = pfd.NodeID(id)
				}
				ch <- checkers.NewProblem(problemID, checkers.SeverityError, pfdcommon.NewLocation(pfdcommon.LocationTypePFD, ids...))
			}
		}
		return nil
	},
}
