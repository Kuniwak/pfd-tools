package pfdcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
)

var Finite = checkers.AtomicChecker[pfdcommon.Target]{
	ID: "finite",
	AvailableIfFunc: func(t pfdcommon.Target) bool {
		return true
	},
	CheckFunc: func(t pfdcommon.Target, ch chan<- checkers.Problem) error {
		// NOTE: Infinite graphs cannot be expanded in memory, so there's no need to consider them.
		return nil
	},
}
