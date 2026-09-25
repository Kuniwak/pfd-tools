package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
)

var MalformedID = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "malformed-id",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "malformed-id"
		for _, node := range t.PFD.Nodes.Iter() {
			if !node.HasID() {
				continue
			}
			if !pfd.IsWellFormedNodeID(node.ID) {
				ch <- checkers.NewProblem(problemID, checkers.SeverityWarning, pfdcommon.NewLocation(pfdcommon.LocationTypePFD, node.ID))
			}
		}
		return nil
	},
}
