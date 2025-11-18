package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
	"github.com/Kuniwak/pfd-tools/sets"
)

var SingleSrc = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "single-src",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "single-src"

		for _, node := range t.PFD.Nodes.Iter() {
			if node.Type != pfd.NodeTypeAtomicDeliverable {
				continue
			}
			ss := sets.New(pfd.NodeID.Compare)
			for _, edge := range t.PFD.Edges.Iter() {
				src, srcOk := t.Memoized.NodeMap[edge.Source]
				if !srcOk {
					// NOTE: Skip as it will be caught by in-field.
					continue
				}
				if src.Type == pfd.NodeTypeCompositeProcess {
					// NOTE: Skip composite processes as they duplicate outputs of their contained atomic processes.
					continue
				}

				if edge.Target == node.ID {
					ss.Add(pfd.NodeID.Compare, edge.Source)
				}
			}
			if ss.Len() > 1 {
				ch <- checkers.NewProblem(problemID, checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypePFD, node.ID), pfdcommon.NewLocation(pfdcommon.LocationTypePFD, ss.Slice()...))...)
			}
		}
		return nil
	},
}
