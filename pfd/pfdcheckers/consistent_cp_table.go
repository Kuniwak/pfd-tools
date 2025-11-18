package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
	"github.com/Kuniwak/pfd-tools/sets"
)

var ConsistentCPTable = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "consistent-cp-table",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return t.CompositeProcessTable != nil
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		expected := sets.New(pfd.CompositeProcessID.Compare)
		for _, cp := range t.PFD.Nodes.Iter() {
			if cp.Type != pfd.NodeTypeCompositeProcess {
				continue
			}
			cp := pfd.CompositeProcessIDFromNodeID(cp.ID, t.Memoized.NodeMap)
			expected.Add(pfd.CompositeProcessID.Compare, cp)
		}

		actual := sets.New(pfd.CompositeProcessID.Compare)
		for _, cp := range t.CompositeProcessTable.Rows {
			actual.Add(pfd.CompositeProcessID.Compare, cp.ID)
		}

		missing := expected.Clone()
		missing.Difference(pfd.CompositeProcessID.Compare, actual)

		extra := actual.Clone()
		extra.Difference(pfd.CompositeProcessID.Compare, expected)

		const missingProblemID = "missing-cp-table"
		for _, cp := range missing.Iter() {
			ch <- checkers.NewProblem(missingProblemID, checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeCompositeProcessTable, pfd.NodeID(cp)))...)
		}

		const extraProblemID = "extra-cp-table"
		for _, cp := range extra.Iter() {
			ch <- checkers.NewProblem(extraProblemID, checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeCompositeProcessTable, pfd.NodeID(cp)))...)
		}

		return nil
	},
}
