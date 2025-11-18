package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
)

var NoP2P = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "no-p2p",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "no-p2p"
		for _, edge := range t.PFD.Edges.Iter() {
			src, srcOk := t.Memoized.NodeMap[edge.Source]
			if !srcOk {
				// NOTE: Skip as it will be caught by in-field.
				continue
			}
			target, targetOk := t.Memoized.NodeMap[edge.Target]
			if !targetOk {
				// NOTE: Skip as it will be caught by in-field.
				continue
			}
			if src.Type.IsProcess() && target.Type.IsProcess() {
				ch <- checkers.NewProblem(problemID, checkers.SeverityError, pfdcommon.NewLocation(pfdcommon.LocationTypePFD, edge.Source, edge.Target))
			}
		}
		return nil
	},
}
