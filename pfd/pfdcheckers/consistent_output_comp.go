package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
	"github.com/Kuniwak/pfd-tools/sets"
)

var ConsistentOutputComp = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "consistent-output-comp",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "consistent-output-comp"
		for comp, ps := range t.PFD.ProcessComposition {
			actualOutputs := t.Memoized.EdgeMap[comp]

			expectedOutputs := sets.New(pfd.NodeID.Compare)
			for _, p1 := range ps.Iter() {
				outputs1 := t.PFD.OutputsIncludingFeedback(p1)
				expectedOutputs.Union(pfd.NodeID.Compare, outputs1)
			}

			if !actualOutputs.IsSubsetOf(pfd.NodeID.Compare, expectedOutputs) {
				diff := actualOutputs.Clone()
				diff.Difference(pfd.NodeID.Compare, expectedOutputs)
				ch <- checkers.NewProblem(problemID, checkers.SeverityError, pfdcommon.NewLocation(pfdcommon.LocationTypePFD, append(diff.Slice(), comp)...))
			}
		}
		return nil
	},
}
