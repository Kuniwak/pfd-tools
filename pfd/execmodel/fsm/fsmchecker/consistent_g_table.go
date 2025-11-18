package fsmchecker

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmchecker/fsmcommon"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmmasterschedule"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/sets"
)

var ConsistentGTable = checkers.AtomicChecker[*fsmcommon.Target]{
	ID: "consistent-g-table",
	AvailableIfFunc: func(t *fsmcommon.Target) bool {
		return t.GroupTable != nil && t.Memoized.HasGroupMap
	},
	CheckFunc: func(t *fsmcommon.Target, ch chan<- checkers.Problem) error {
		const malformedGroupProblemID = "malformed-g-table"

		actual := sets.New(fsmmasterschedule.Group.Compare)
		for _, group := range t.GroupTable.Rows {
			actual.Add(fsmmasterschedule.Group.Compare, group.ID)
		}

		expected := sets.New(fsmmasterschedule.Group.Compare)
		for ap, groupIDsText := range t.Memoized.GroupMap {
			groups, err := fsmtable.ParseGroups(groupIDsText)
			if err != nil {
				ch <- checkers.NewProblem(
					malformedGroupProblemID,
					checkers.SeverityError,
					fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicProcessTable, fsmcommon.NewAtomicProcessID(ap)),
				)
				continue
			}
			expected.Union(fsmmasterschedule.Group.Compare, groups)
		}

		missing := expected.Clone()
		missing.Difference(fsmmasterschedule.Group.Compare, actual)

		extra := actual.Clone()
		extra.Difference(fsmmasterschedule.Group.Compare, expected)

		const missingGroupProblemID = "missing-g-table"
		const extraGroupProblemID = "extra-g-table"
		for _, group := range missing.Iter() {
			ch <- checkers.NewProblem(
				missingGroupProblemID,
				checkers.SeverityError,
				fsmcommon.NewLocation(
					fsmcommon.LocationTypeGroupTable,
					fsmcommon.NewGroupID(fsmmasterschedule.Group(group)),
				),
			)
		}

		for _, group := range extra.Iter() {
			ch <- checkers.NewProblem(
				extraGroupProblemID,
				checkers.SeverityWarning,
				fsmcommon.NewLocation(
					fsmcommon.LocationTypeGroupTable,
					fsmcommon.NewGroupID(fsmmasterschedule.Group(group)),
				),
			)
		}
		return nil
	},
}
