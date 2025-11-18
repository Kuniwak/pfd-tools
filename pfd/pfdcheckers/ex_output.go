package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
)

var ExOutput = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "ex-output",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "ex-output"

		state := make(map[pfd.NodeID]bool)
		for _, node := range t.PFD.Nodes.Iter() {
			if node.Type.IsProcess() {
				state[node.ID] = false
			}
		}
		for _, edge := range t.PFD.Edges.Iter() {
			state[edge.Source] = true
		}
		for _, node := range t.PFD.Nodes.Iter() {
			if (node.Type.IsProcess()) && !state[node.ID] {
				ch <- checkers.NewProblem(problemID, checkers.SeverityError, pfdcommon.NewLocation(pfdcommon.LocationTypePFD, node.ID))
			}
		}
		return nil
	},
}
