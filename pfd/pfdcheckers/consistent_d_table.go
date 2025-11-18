package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
	"github.com/Kuniwak/pfd-tools/sets"
)

var ConsistentDTable = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "consistent-d-table",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return t.AtomicDeliverableTable != nil
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		expected := sets.New(pfd.AtomicDeliverableID.Compare)
		for _, d := range t.PFD.Nodes.Iter() {
			if d.Type != pfd.NodeTypeAtomicDeliverable {
				continue
			}
			d := pfd.AtomicDeliverableIDFromNodeID(d.ID, t.Memoized.NodeMap)
			expected.Add(pfd.AtomicDeliverableID.Compare, d)
		}
		actual := sets.New(pfd.AtomicDeliverableID.Compare)
		for _, d := range t.AtomicDeliverableTable.Rows {
			actual.Add(pfd.AtomicDeliverableID.Compare, d.ID)
		}

		missing := expected.Clone()
		missing.Difference(pfd.AtomicDeliverableID.Compare, actual)

		extra := actual.Clone()
		extra.Difference(pfd.AtomicDeliverableID.Compare, expected)

		const missingProblemID = "missing-d-table"
		for _, d := range missing.Iter() {
			ch <- checkers.NewProblem(missingProblemID, checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeDeliverableTable, pfd.NodeID(d)))...)
		}

		const extraProblemID = "extra-d-table"
		for _, d := range extra.Iter() {
			ch <- checkers.NewProblem(extraProblemID, checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeDeliverableTable, pfd.NodeID(d)))...)
		}

		return nil
	},
}
