package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
)

var ConsistentDesc = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "consistent-desc",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "consistent-desc"
		descMap := make(map[pfd.NodeID]string)
		for _, node := range t.PFD.Nodes.Iter() {
			if _, ok := descMap[node.ID]; ok {
				if descMap[node.ID] != node.Description {
					ch <- checkers.NewProblem(problemID, checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypePFD, node.ID))...)
				}
			} else {
				descMap[node.ID] = node.Description
			}
		}
		return nil
	},
}
