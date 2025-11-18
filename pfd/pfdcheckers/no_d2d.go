package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
)

var NoD2D = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "no-d2d",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "no-d2d"
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
			if src.Type.IsDeliverable() && target.Type.IsDeliverable() {
				ch <- checkers.NewProblem(problemID, checkers.SeverityError, pfdcommon.NewLocation(pfdcommon.LocationTypePFD, edge.Source, edge.Target))
			}
		}
		return nil
	},
}
