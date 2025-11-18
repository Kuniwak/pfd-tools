package fsmchecker

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmchecker/fsmcommon"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
)

var ValidResourcesSet = checkers.AtomicChecker[*fsmcommon.Target]{
	ID: "valid-resources-set",
	AvailableIfFunc: func(t *fsmcommon.Target) bool {
		return t.Memoized.HasNeededResourceSetsMap
	},
	CheckFunc: func(t *fsmcommon.Target, ch chan<- checkers.Problem) error {
		for ap, neededResourceSetsText := range t.Memoized.NeededResourceSetsMap {
			const problemIDMalformedResourceSetNotation = "malformed-resources-set-notation"
			neededResourceSets, err := fsmtable.ParseNeededResourceSetEntry(neededResourceSetsText)
			if err != nil {
				ch <- checkers.NewProblem(problemIDMalformedResourceSetNotation, checkers.SeverityError, fsmcommon.NewLocations(fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicProcessTable, fsmcommon.NewAtomicProcessID(ap)))...)
				return nil
			}

			const problemIDEmptyResourceSet = "empty-resources-set"
			if neededResourceSets.Len() == 0 {
				ch <- checkers.NewProblem(problemIDEmptyResourceSet, checkers.SeverityError, fsmcommon.NewLocations(fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicProcessTable, fsmcommon.NewAtomicProcessID(ap)))...)
				continue
			}

			const problemIDZeroConsumedVolume = "zero-consumed-volume"
			for _, neededResourceSet := range neededResourceSets.Iter() {
				if neededResourceSet.ConsumedVolume.IsZero() {
					ch <- checkers.NewProblem(problemIDZeroConsumedVolume, checkers.SeverityError, fsmcommon.NewLocations(fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicProcessTable, fsmcommon.NewAtomicProcessID(ap)))...)
				}
			}
		}
		return nil
	},
}
