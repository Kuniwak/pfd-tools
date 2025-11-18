package fsm

import (
	"fmt"
	"strings"

	"github.com/Kuniwak/pfd-tools/sets"
)

// SearchFastest returns an approximate solution using a greedy algorithm.
func SearchFastest() SearchFunc {
	return func(e *Env) (*sets.Set[*Plan], error) {
		s := e.InitialState()
		p := NewEmptyPlan(s)
		depth := 0
		for !e.IsCompleted(s) && depth < 1_000_000 {
			depth++

			trs := e.Transitions(s)
			if trs.Len() == 0 {
				sb := &strings.Builder{}
				_ = s.Write(sb)
				sb.WriteString("\n")
				m := e.AllocatabilityInfoMap(s)
				_ = m.Write(sb)
				return nil, fmt.Errorf("fsm.SearchFastest: deadlock found at depth %d:\n%s", depth, sb.String())
			}

			best := Allocation{}

			for _, tr := range trs.Iter() {
				if tr.Allocation.TotalConsumedVolume() > best.TotalConsumedVolume() {
					best = tr.Allocation
				}
			}

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
