package fsm

import (
	"fmt"

	"github.com/Kuniwak/pfd-tools/sets"
)

type SearchFunc func(e *Env) (*sets.Set[*Plan], error)

type SearchWithPrefixFunc func(e *Env, prefix *Plan) (*sets.Set[*Plan], error)

func ReplayPrefix(e *Env, prefix *Plan) (*Plan, State, error) {
	return ReplayPrefixWith(e, prefix, e.AvailableAllocationsFunc)
}

func ReplayPrefixWith(e *Env, prefix *Plan, availableAllocationsFunc AvailableAllocationsFunc) (*Plan, State, error) {
	s := e.InitialState()
	replayed := NewEmptyPlan(s)

	for i, tr := range prefix.Transitions {
		trs := e.TransitionsWith(s, availableAllocationsFunc)
		found := false
		for _, candidate := range trs.Iter() {
			if candidate.Allocation.Equals(tr.Allocation) {
				replayed.Add(candidate)
				s = candidate.NextState
				found = true
				break
			}
		}
		if !found {
			return nil, State{}, fmt.Errorf("fsm.ReplayPrefixWith: prefix transition %d not found in available transitions (allocation=%v)", i, tr.Allocation)
		}
	}

	return replayed, s, nil
}
