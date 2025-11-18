package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
)

var DisjOrProperSubsetComp = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "disj-or-psubset-comp",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "disj-or-psubset-comp"
		for comp1, ps1 := range t.PFD.ProcessComposition {
			for comp2, ps2 := range t.PFD.ProcessComposition {
				// NOTE: comp1 and comp2 are different, and there's no need to check twice.
				if comp1 <= comp2 {
					continue
				}
				if !ps1.IsDisjointWith(pfd.NodeID.Compare, ps2) && !ps1.IsProperSubsetOf(pfd.NodeID.Compare, ps2) && !ps2.IsProperSubsetOf(pfd.NodeID.Compare, ps1) {
					ch <- checkers.NewProblem(problemID, checkers.SeverityError, pfdcommon.NewLocation(pfdcommon.LocationTypePFD, comp1), pfdcommon.NewLocation(pfdcommon.LocationTypePFD, comp2))
				}
			}
		}
		return nil
	},
}
