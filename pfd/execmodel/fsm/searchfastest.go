package fsm

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Kuniwak/pfd-tools/sets"
)

func PickFastestAllocation(trs *sets.Set[*Trans], rng *rand.Rand) Allocation {
	best := Allocation{}
	for _, tr := range trs.Iter() {
		if tr.Allocation.TotalConsumedVolume() > best.TotalConsumedVolume() {
			best = tr.Allocation
		}
	}

	maxVol := best.TotalConsumedVolume()
	if maxVol == 0 {

		return best
	}

	tied := make([]Allocation, 0)
	for _, tr := range trs.Iter() {
		if tr.Allocation.TotalConsumedVolume() == maxVol {
			tied = append(tied, tr.Allocation)
		}
	}
	return tied[rng.Intn(len(tied))]
}

func SearchFastestWithPrefix(seed int64) SearchWithPrefixFunc {
	return func(e *Env, prefix *Plan) (*sets.Set[*Plan], error) {
		rng := rand.New(rand.NewSource(seed))
		var s State
		var p *Plan
		if prefix != nil {
			replayed, replayedState, err := ReplayPrefix(e, prefix)
			if err != nil {
				return nil, fmt.Errorf("fsm.SearchFastestWithPrefix: %w", err)
			}
			s = replayedState
			p = replayed
		} else {
			s = e.InitialState()
			p = NewEmptyPlan(s)
		}
		depth := 0
		for !e.IsCompleted(s) && depth < 1_000_000 {
			depth++
			e.Logger.Debug("fsm.SearchFastest: step", "depth", depth, "time", float64(s.Time))
			stepStart := time.Now()

			trs := e.Transitions(s)
			e.Logger.Debug("fsm.SearchFastest: transitions", "depth", depth, "count", trs.Len(), "elapsedMs", time.Since(stepStart).Milliseconds())
			if trs.Len() == 0 {
				sb := &strings.Builder{}
				_ = s.Write(sb)
				sb.WriteString("\n")
				m := e.AllocatabilityInfoMap(s)
				_ = m.Write(sb)
				return nil, fmt.Errorf("fsm.SearchFastestWithPrefix: deadlock found at depth %d:\n%s", depth, sb.String())
			}

			best := PickFastestAllocation(trs, rng)

			ns, ok := e.NextState(s, best)
			if !ok {
				continue
			}
			p.Add(&Trans{
				Allocation: best,
				NextState:  ns,
			})
			s = ns
		}
		return sets.New((*Plan).Compare, p), nil
	}
}

func runFastestLoop(e *Env, rng *rand.Rand, s State, p *Plan, name string, availableAllocationsFunc AvailableAllocationsFunc) (*sets.Set[*Plan], error) {
	depth := 0
	for !e.IsCompleted(s) && depth < 1_000_000 {
		depth++
		e.Logger.Debug(name+": step", "depth", depth, "time", float64(s.Time))
		stepStart := time.Now()

		trs := e.TransitionsWith(s, availableAllocationsFunc)
		e.Logger.Debug(name+": transitions", "depth", depth, "count", trs.Len(), "elapsedMs", time.Since(stepStart).Milliseconds())
		if trs.Len() == 0 {
			sb := &strings.Builder{}
			_ = s.Write(sb)
			sb.WriteString("\n")
			m := e.AllocatabilityInfoMap(s)
			_ = m.Write(sb)
			return nil, fmt.Errorf("%s: deadlock found at depth %d:\n%s", name, depth, sb.String())
		}

		best := PickFastestAllocation(trs, rng)

		ns, ok := e.NextState(s, best)
		if !ok {
			continue
		}
		p.Add(&Trans{
			Allocation: best,
			NextState:  ns,
		})
		s = ns
	}
	return sets.New((*Plan).Compare, p), nil
}

func SearchFastest(seed int64) SearchFunc {
	return func(e *Env) (*sets.Set[*Plan], error) {
		rng := rand.New(rand.NewSource(seed))
		s := e.InitialState()
		p := NewEmptyPlan(s)
		return runFastestLoop(e, rng, s, p, "fsm.SearchFastest", e.AvailableAllocationsFunc)
	}
}

func SearchPoorest(seed int64) SearchFunc {
	return func(e *Env) (*sets.Set[*Plan], error) {
		rng := rand.New(rand.NewSource(seed))
		s := e.InitialState()
		p := NewEmptyPlan(s)
		greedy := NewGreedyAvailableAllocationsFunc(e.NeededResourceSetsFunc)
		return runFastestLoop(e, rng, s, p, "fsm.SearchPoorest", greedy)
	}
}

func SearchPoorestWithPrefix(seed int64) SearchWithPrefixFunc {
	return func(e *Env, prefix *Plan) (*sets.Set[*Plan], error) {
		rng := rand.New(rand.NewSource(seed))
		greedy := NewGreedyAvailableAllocationsFunc(e.NeededResourceSetsFunc)
		var s State
		var p *Plan
		if prefix != nil {
			replayed, replayedState, err := ReplayPrefixWith(e, prefix, greedy)
			if err != nil {
				return nil, fmt.Errorf("fsm.SearchPoorestWithPrefix: %w", err)
			}
			s = replayedState
			p = replayed
		} else {
			s = e.InitialState()
			p = NewEmptyPlan(s)
		}
		return runFastestLoop(e, rng, s, p, "fsm.SearchPoorestWithPrefix", greedy)
	}
}
