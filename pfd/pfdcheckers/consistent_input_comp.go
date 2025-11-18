package pfdcheckers

import (
	"fmt"

	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
	"github.com/Kuniwak/pfd-tools/sets"
)

var ConsistentInputComp = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "consistent-input-comp",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "consistent-input-comp"
		for comp, ps := range t.PFD.ProcessComposition {
			actualInputs, ok := t.Memoized.ReversedEdgeMap[comp]
			if !ok {
				panic(fmt.Sprintf("consistent-input-comp: missing reversed edge map for %s", comp))
			}

			outputs := sets.New(pfd.NodeID.Compare)
			all := sets.New(pfd.NodeID.Compare)
			for _, p1 := range ps.Iter() {
				inputs1 := t.PFD.InputsIncludingFeedback(p1)
				outputs1 := t.PFD.OutputsIncludingFeedback(p1)
				outputs.Union(pfd.NodeID.Compare, outputs1)
				all.Union(pfd.NodeID.Compare, inputs1)
				all.Union(pfd.NodeID.Compare, outputs1)
			}

			// NOTE: allcheckers is no longer used, so we modify it in-place.
			all.Difference(pfd.NodeID.Compare, outputs)
			expectedInputs := all

			if !sets.IsEqual(pfd.NodeID.Compare, actualInputs, expectedInputs) {
				base := actualInputs.Clone()
				base.Union(pfd.NodeID.Compare, expectedInputs)

				inter := actualInputs.Clone()
				inter.Intersection(pfd.NodeID.Compare, expectedInputs)

				base.Difference(pfd.NodeID.Compare, inter)

				diff := append(base.Slice(), comp)
				ch <- checkers.NewProblem(problemID, checkers.SeverityError, pfdcommon.NewLocation(pfdcommon.LocationTypePFD, diff...))
			}
		}
		return nil
	},
}
