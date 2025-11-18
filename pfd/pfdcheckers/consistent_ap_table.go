package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
	"github.com/Kuniwak/pfd-tools/sets"
)

var ConsistentAPTable = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "consistent-ap-table",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return t.AtomicProcessTable != nil
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		expected := sets.New(pfd.AtomicProcessID.Compare)
		for _, node := range t.PFD.Nodes.Iter() {
			if node.Type != pfd.NodeTypeAtomicProcess {
				continue
			}
			ap := pfd.AtomicProcessIDFromNodeID(node.ID, t.Memoized.NodeMap)
			expected.Add(pfd.AtomicProcessID.Compare, ap)
		}
		actual := sets.New(pfd.AtomicProcessID.Compare)
		for _, ap := range t.AtomicProcessTable.Rows {
			actual.Add(pfd.AtomicProcessID.Compare, ap.ID)
		}

		missing := expected.Clone()
		missing.Difference(pfd.AtomicProcessID.Compare, actual)
		extra := actual.Clone()
		extra.Difference(pfd.AtomicProcessID.Compare, expected)

		const missingProblemID = "missing-ap-table"
		for _, ap := range missing.Iter() {
			ch <- checkers.NewProblem(missingProblemID, checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeAtomicProcessTable, pfd.NodeID(ap)))...)
		}
		const extraProblemID = "extra-ap-table"
		for _, ap := range extra.Iter() {
			ch <- checkers.NewProblem(extraProblemID, checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeAtomicProcessTable, pfd.NodeID(ap)))...)
		}
		return nil
	},
}
