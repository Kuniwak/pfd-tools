package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
)

var NoP2DFB = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "no-p2d-fb",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "no-p2d-fb"
		for _, edge := range t.PFD.Edges.Iter() {
			if !edge.IsFeedback {
				continue
			}
			src, ok := t.Memoized.NodeMap[edge.Source]
			if !ok {
				// NOTE: Skip as it will be caught by in-field.
				continue
			}
			target, ok := t.Memoized.NodeMap[edge.Target]
			if !ok {
				// NOTE: Skip as it will be caught by in-field.
				continue
			}
			if src.Type.IsDeliverable() && target.Type.IsProcess() {
				continue
			}
			ch <- checkers.NewProblem(problemID, checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypePFD, edge.Source, edge.Target))...)
		}
		return nil
	},
}
