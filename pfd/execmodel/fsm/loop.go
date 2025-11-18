package fsm

import (
	"github.com/Kuniwak/pfd-tools/pfd"
)

// LoopEndFunc is the loop termination condition for atomic processes. Returns true if the loop should terminate, false otherwise.
type LoopEndFunc func(ap pfd.AtomicProcessID, src pfd.AtomicDeliverableID, numOfLoops int) bool

// ConstLoopEndFunc returns true if the given maximum number of rework iterations is exceeded, false otherwise.
func ConstLoopEndFunc(maxNumOfLoops int) LoopEndFunc {
	return func(_ pfd.AtomicProcessID, _ pfd.AtomicDeliverableID, numOfLoops int) bool {
		return numOfLoops >= maxNumOfLoops
	}
}
