package fsmchecker

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmchecker/fsmcommon"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmmasterschedule"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/sets"
)

var ConsistentMTable = checkers.AtomicChecker[*fsmcommon.Target]{
	ID: "consistent-m-table",
	AvailableIfFunc: func(t *fsmcommon.Target) bool {
		return t.MilestoneTable != nil && t.Memoized.HasMilestoneMap && t.Memoized.HasMilestoneEdgesMap
	},
	CheckFunc: func(t *fsmcommon.Target, ch chan<- checkers.Problem) error {
		const malformedMilestoneProblemID = "malformed-m-table"
		const malformedMilestoneSuccessorsProblemID = "malformed-m-table-successors"

		actual := sets.New(fsmmasterschedule.Milestone.Compare)
		for _, row := range t.MilestoneTable.Rows {
			actual.Add(fsmmasterschedule.Milestone.Compare, row.MilestoneID)
		}

		expected1 := sets.New(fsmmasterschedule.Milestone.Compare)
		for _, milestoneIDText := range t.Memoized.MilestoneMap {
			milestoneID, err := fsmtable.ParseMilestone(milestoneIDText)
			if err != nil {
				ch <- checkers.NewProblem(
					malformedMilestoneProblemID,
					checkers.SeverityError,
					fsmcommon.NewLocation(fsmcommon.LocationTypeMilestoneTable, fsmcommon.NewMilestoneID(milestoneID)),
				)
				continue
			}
			expected1.Add(fsmmasterschedule.Milestone.Compare, milestoneID)
		}

		expected2 := sets.New(fsmmasterschedule.Milestone.Compare)
		for m, successorsText := range t.Memoized.MilestoneEdgesMap {
			successors, err := fsmtable.ParseSuccessors(successorsText)
			if err != nil {
				ch <- checkers.NewProblem(
					malformedMilestoneSuccessorsProblemID,
					checkers.SeverityError,
					fsmcommon.NewLocation(fsmcommon.LocationTypeMilestoneTable, fsmcommon.NewMilestoneID(m)),
				)
				continue
			}
			for _, successor := range successors.Iter() {
				expected2.Add(fsmmasterschedule.Milestone.Compare, successor)
			}
		}

		missing1 := expected1.Clone()
		missing1.Difference(fsmmasterschedule.Milestone.Compare, actual)

		extra1 := actual.Clone()
		extra1.Difference(fsmmasterschedule.Milestone.Compare, expected1)

		missing2 := expected2.Clone()
		missing2.Difference(fsmmasterschedule.Milestone.Compare, actual)

		const missingMilestoneProblemID = "missing-m-table"
		for _, milestone := range missing1.Iter() {
			ch <- checkers.NewProblem(
				missingMilestoneProblemID,
				checkers.SeverityError,
				fsmcommon.NewLocation(
					fsmcommon.LocationTypeMilestoneTable,
					fsmcommon.NewMilestoneID(milestone),
				),
			)
		}

		for _, milestone := range missing2.Iter() {
			ch <- checkers.NewProblem(
				missingMilestoneProblemID,
				checkers.SeverityError,
				fsmcommon.NewLocation(
					fsmcommon.LocationTypeMilestoneTable,
					fsmcommon.NewMilestoneID(milestone),
				),
			)
		}

		const extraMilestoneProblemID = "extra-m-table"
		for _, milestone := range extra1.Iter() {
			ch <- checkers.NewProblem(
				extraMilestoneProblemID,
				checkers.SeverityWarning,
				fsmcommon.NewLocation(
					fsmcommon.LocationTypeMilestoneTable,
					fsmcommon.NewMilestoneID(milestone),
				),
			)
		}

		return nil
	},
}
