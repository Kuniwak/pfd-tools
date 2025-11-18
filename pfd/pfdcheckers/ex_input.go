package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
)

var ExInput = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "ex-input",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "ex-input"

		state := make(map[pfd.NodeID]bool)
		for _, node := range t.PFD.Nodes.Iter() {
			if node.Type.IsProcess() {
				state[node.ID] = false
			}
		}
		for _, edge := range t.PFD.Edges.Iter() {
			state[edge.Target] = true
		}
		for _, node := range t.PFD.Nodes.Iter() {
			if node.Type.IsProcess() && !state[node.ID] {
				ch <- checkers.NewProblem(problemID, checkers.SeverityError, pfdcommon.NewLocation(pfdcommon.LocationTypePFD, node.ID))
			}
		}
		return nil
	},
}
