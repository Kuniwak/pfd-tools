package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
	"github.com/Kuniwak/pfd-tools/sets"
)

var ConsistentCDTable = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "consistent-cd-table",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		expected := sets.New(pfd.CompositeDeliverableID.Compare)
		for _, cd := range t.PFD.Nodes.Iter() {
			if cd.Type != pfd.NodeTypeCompositeDeliverable {
				continue
			}
			cd := pfd.CompositeDeliverableIDFromNodeID(cd.ID, t.Memoized.NodeMap)
			expected.Add(pfd.CompositeDeliverableID.Compare, cd)
		}

		actual := sets.New(pfd.CompositeDeliverableID.Compare)
		if t.CompositeDeliverableTable != nil {
			for _, cd := range t.CompositeDeliverableTable.Rows {
				actual.Add(pfd.CompositeDeliverableID.Compare, cd.ID)
			}
		}

		missing := expected.Clone()
		missing.Difference(pfd.CompositeDeliverableID.Compare, actual)

		nested := sets.New(pfd.CompositeDeliverableID.Compare)
		if t.CompositeDeliverableTable != nil {
			nested = t.CompositeDeliverableTable.NestedIDs(expected, t.PFD.Nodes)
		}

		extra := actual.Clone()
		extra.Difference(pfd.CompositeDeliverableID.Compare, expected)
		extra.Difference(pfd.CompositeDeliverableID.Compare, nested)

		const missingProblemID = "missing-cd-table"
		for _, cd := range missing.Iter() {
			ch <- checkers.NewProblem(missingProblemID, checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeCompositeDeliverableTable, pfd.NodeID(cd)))...)
		}

		const extraProblemID = "extra-cd-table"
		for _, cd := range extra.Iter() {
			ch <- checkers.NewProblem(extraProblemID, checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeCompositeDeliverableTable, pfd.NodeID(cd)))...)
		}

		return nil
	},
}
