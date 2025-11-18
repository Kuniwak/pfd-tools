package fsmchecker

import (
	"slices"

	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmchecker/fsmcommon"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/sets"
)

var ConsistentResourceTable = checkers.AtomicChecker[*fsmcommon.Target]{
	ID: "consistent-r-table",
	AvailableIfFunc: func(t *fsmcommon.Target) bool {
		return t.Memoized.HasAllResources && t.Memoized.HasNeededResourceSetsMap
	},
	CheckFunc: func(t *fsmcommon.Target, ch chan<- checkers.Problem) error {
		expected := sets.New(fsm.ResourceID.Compare)
		for _, entryText := range t.Memoized.NeededResourceSetsMap {
			entries, err := fsmtable.ParseNeededResourceSetEntry(entryText)
			if err != nil {
				// NOTE: Skip because it will be caught by valid-resources-set.
				continue
			}
			for _, entry := range entries.Iter() {
				for _, res := range entry.Resources.Iter() {
					expected.Add(fsm.ResourceID.Compare, res)
				}
			}
		}

		actual := t.Memoized.AllResources

		missing := expected.Clone()
		missing.Difference(fsm.ResourceID.Compare, actual)
		missingSorted := missing.Slice()
		slices.SortFunc(missingSorted, fsm.ResourceID.Compare)

		extra := actual.Clone()
		extra.Difference(fsm.ResourceID.Compare, expected)
		extraSorted := extra.Slice()
		slices.SortFunc(extraSorted, fsm.ResourceID.Compare)

		const missingProblemID = "missing-r-table"
		for _, r := range missing.Iter() {
			ch <- checkers.NewProblem(missingProblemID, checkers.SeverityError, fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicProcessTable, fsmcommon.NewResourceID(r)))
		}

		const extraProblemID = "extra-r-table"
		for _, r := range extra.Iter() {
			ch <- checkers.NewProblem(extraProblemID, checkers.SeverityWarning, fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicProcessTable, fsmcommon.NewResourceID(r)))
		}
		return nil
	},
}
