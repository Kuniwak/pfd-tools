package fsm

import (
	"container/heap"
	"fmt"
	"hash/maphash"

	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/sets"
)

const searchProgressLogInterval = 10_000

func SearchBestPlansWithPrefix() SearchWithPrefixFunc {
	return func(e *Env, prefix *Plan) (*sets.Set[*Plan], error) {
		var start State
		var prefixPlan *Plan
		if prefix != nil {
			replayed, replayedState, err := ReplayPrefix(e, prefix)
			if err != nil {
				return nil, fmt.Errorf("fsm.SearchBestPlansWithPrefix: %w", err)
			}
			start = replayedState
			prefixPlan = replayed
		} else {
			start = e.InitialState()
		}

		results, err := searchBestPlansFromState(e, start)
		if err != nil {
			return nil, err
		}

		if prefixPlan != nil {
			merged := sets.NewWithCapacity[*Plan](results.Len())
			for _, plan := range results.Iter() {
				combined := prefixPlan.Clone()
				for _, tr := range plan.Transitions {
					combined.Add(tr)
				}
				merged.Add((*Plan).Compare, combined)
			}
			return merged, nil
		}
		return results, nil
	}
}

func SearchBestPlans() SearchFunc {
	return searchBestPlans
}

func searchBestPlans(e *Env) (*sets.Set[*Plan], error) {
	start := e.InitialState()
	return searchBestPlansFromState(e, start)
}

func searchBestPlansFromState(e *Env, start State) (*sets.Set[*Plan], error) {

	type parentInfo struct {
		parent uint64
		alloc  Allocation
		child  State
	}

	dist := make(map[uint64]execmodel.Time, 1024)

	parents := make(map[uint64][]parentInfo, 1024)

	stateRep := make(map[uint64]State, 1024)

	h := &maphash.Hash{}
	if err := HashStateWithoutTime(start, h); err != nil {
		return nil, fmt.Errorf("fsm.Env.SearchBestPlans: %w", err)
	}
	startKey := h.Sum64()

	dist[startKey] = start.Time
	stateRep[startKey] = start

	pq := &planPQ{}
	heap.Init(pq)
	heap.Push(pq, &pqItem{
		key:        startKey,
		state:      start,
		priorityT:  start.Time,
		negTotCons: 0,
		seq:        0,
	})

	bestTime := execmodel.Time(-1)
	goalKeys := make([]uint64, 0, 8)

	iter := 0
	e.Logger.Debug("fsm.Env.SearchBestPlans: start", "time", float64(start.Time))

	for pq.Len() > 0 {
		it := heap.Pop(pq).(*pqItem)
		k := it.key
		s := it.state

		iter++
		if iter%searchProgressLogInterval == 0 {
			e.Logger.Debug("fsm.Env.SearchBestPlans: progress", "iter", iter, "pqLen", pq.Len(), "time", float64(s.Time), "bestTime", float64(bestTime))
		}

		if d, ok := dist[k]; !ok || d != s.Time {
			continue
		}

		if bestTime >= 0 && s.Time > bestTime {
			break
		}

		isCompleted := e.IsCompleted(s)
		if isCompleted {
			if bestTime < 0 || s.Time < bestTime {
				bestTime = s.Time
				goalKeys = goalKeys[:0]
				goalKeys = append(goalKeys, k)
			} else if s.Time == bestTime {
				goalKeys = append(goalKeys, k)
			}

			continue
		}

		ts := e.Transitions(s)
		if ts.Len() == 0 {
			e.Logger.Warn("fsm.Env.SearchBestPlans: deadlock found", "state", s)
			continue
		}

		for _, tr := range ts.Iter() {
			ns := tr.NextState

			if bestTime >= 0 && ns.Time > bestTime {
				continue
			}
			h.Reset()
			if err := HashStateWithoutTime(ns, h); err != nil {
				return nil, fmt.Errorf("fsm.Env.SearchBestPlans: %w", err)
			}
			nk := h.Sum64()
			newT := ns.Time

			oldT, ok := dist[nk]
			if !ok || newT < oldT {
				dist[nk] = newT
				stateRep[nk] = ns
				parents[nk] = []parentInfo{{
					parent: k,
					alloc:  tr.Allocation,
					child:  ns,
				}}
				heap.Push(pq, &pqItem{
					key:        nk,
					state:      ns,
					priorityT:  newT,
					negTotCons: -int(tr.Allocation.TotalConsumedVolume()),
					seq:        pq.nextSeq(),
				})
			} else if newT == oldT {

				parents[nk] = append(parents[nk], parentInfo{
					parent: k,
					alloc:  tr.Allocation,
					child:  ns,
				})
			}
		}
	}

	e.Logger.Debug("fsm.Env.SearchBestPlans: done", "iter", iter, "goals", len(goalKeys), "bestTime", float64(bestTime))

	if bestTime < 0 || len(goalKeys) == 0 {
		return sets.NewWithCapacity[*Plan](0), nil
	}

	results := make([]*Plan, 0, len(goalKeys))
	memoPaths := make(map[uint64][]*Plan, len(parents))

	var buildAll func(k uint64) []*Plan
	buildAll = func(k uint64) []*Plan {
		if k == startKey {

			return []*Plan{NewEmptyPlan(start)}
		}
		if v, ok := memoPaths[k]; ok {
			return v
		}
		ps := parents[k]
		if len(ps) == 0 {

			return nil
		}
		acc := make([]*Plan, 0, 8)
		for _, p := range ps {
			subPlans := buildAll(p.parent)
			for _, sp := range subPlans {
				np := sp.Clone()
				np.Add(&Trans{
					Allocation: p.alloc,
					NextState:  p.child,
				})
				acc = append(acc, np)
			}
		}
		memoPaths[k] = acc
		return acc
	}

	for _, gk := range goalKeys {
		results = append(results, buildAll(gk)...)
	}

	return sets.New((*Plan).Compare, results...), nil
}

type pqItem struct {
	key        uint64
	state      State
	priorityT  execmodel.Time
	negTotCons int
	seq        int64
	index      int
}
type planPQ struct {
	data []*pqItem
	seqc int64
}

func (q *planPQ) Len() int { return len(q.data) }
func (q *planPQ) Less(i, j int) bool {
	a, b := q.data[i], q.data[j]
	if a.priorityT != b.priorityT {
		return a.priorityT < b.priorityT
	}
	if a.negTotCons != b.negTotCons {
		return a.negTotCons < b.negTotCons
	}
	return a.seq < b.seq
}
func (q *planPQ) Swap(i, j int) {
	q.data[i], q.data[j] = q.data[j], q.data[i]
	q.data[i].index = i
	q.data[j].index = j
}
func (q *planPQ) Push(x any) {
	it := x.(*pqItem)
	it.index = len(q.data)
	q.data = append(q.data, it)
}
func (q *planPQ) Pop() any {
	n := len(q.data)
	it := q.data[n-1]
	q.data[n-1] = nil
	q.data = q.data[:n-1]
	return it
}
func (q *planPQ) nextSeq() int64 {
	q.seqc++
	return q.seqc
}
