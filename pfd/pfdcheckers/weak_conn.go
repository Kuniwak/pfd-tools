package pfdcheckers

import (
	"slices"

	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
)

var WeakConn = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "weak-conn",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		const problemID = "weak-conn"
		wcs := t.Memoized.GraphIncludingFB.WeaklyConnectedComponents()
		if wcs.Len() > 1 {
			pfdWcs := make([]pfdcommon.Location, wcs.Len())
			for i, wc := range wcs.Iter() {
				loc := make([]pfd.NodeID, 0, wc.Len())
				for _, id := range wc.Iter() {
					loc = append(loc, pfd.NodeID(id))
					slices.SortFunc(loc, pfd.NodeID.Compare)
				}
				pfdWcs[i] = pfdcommon.NewLocation(pfdcommon.LocationTypePFD, loc...)
			}
			slices.SortFunc(pfdWcs, pfdcommon.CompareLocation)
			ch <- checkers.NewProblem(problemID, checkers.SeverityWarning, pfdcommon.NewLocations(pfdWcs...)...)
		}
		return nil
	},
}
