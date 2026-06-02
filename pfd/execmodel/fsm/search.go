package fsm

import (
	"fmt"

	"github.com/Kuniwak/pfd-tools/sets"
)

type SearchFunc func(e *Env) (*sets.Set[*Plan], error)

// SearchWithPrefixFunc is a search function with a plan prefix.
// If prefix is nil, the search starts from InitialState().
// If prefix is non-nil, the prefix's transitions are replayed with validation from InitialState(),
// and the search continues from the state after the replay completes.
type SearchWithPrefixFunc func(e *Env, prefix *Plan) (*sets.Set[*Plan], error)

// ReplayPrefix replays the prefix's transitions with validation from InitialState() in the modified environment e.
// At each step, it computes e.Transitions(currentState) and validates that the prefix's allocation is included in the options.
// If it is not included, it returns an error (detecting an inconsistency between the plan and the environment).
// On success, it returns a Plan containing the replayed transitions and the state after the replay.
func ReplayPrefix(e *Env, prefix *Plan) (*Plan, State, error) {
	s := e.InitialState()
	replayed := NewEmptyPlan(s)

	for i, tr := range prefix.Transitions {
		trs := e.Transitions(s)
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
			return nil, State{}, fmt.Errorf("fsm.ReplayPrefix: prefix transition %d not found in available transitions (allocation=%v)", i, tr.Allocation)
		}
	}

	return replayed, s, nil
}
