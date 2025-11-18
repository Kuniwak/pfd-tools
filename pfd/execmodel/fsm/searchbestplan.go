package fsm

import (
	"container/heap"
	"fmt"
	"hash/maphash"

	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/sets"
)

func SearchBestPlans() SearchFunc {
	return searchBestPlans
}

// searchBestPlans returns execution plans with the shortest completion time (all of them if there are ties).
// Optimality: Guarantees minimum completion time using Dijkstra (Uniform-Cost Search).
func searchBestPlans(e *Env) (*sets.Set[*Plan], error) {
	start := e.InitialState()

	type parentInfo struct {
		parent uint64
		alloc  Allocation
		child  State
	}

	// dist[k] = the shortest time to reach that "state content (excluding time)"
	dist := make(map[uint64]execmodel.Time, 1024)
	// parents[k] = parent transitions that can reach k in the shortest time dist[k] (multiple)
	parents := make(map[uint64][]parentInfo, 1024)
	// keep representative State for storage (used for restoration and transition generation)
	stateRep := make(map[uint64]State, 1024)

	h := &maphash.Hash{}
	if err := HashStateWithoutTime(start, h); err != nil {
		return nil, fmt.Errorf("fsm.Env.SearchBestPlans: %w", err)
	}
	startKey := h.Sum64()

	dist[startKey] = start.Time
	stateRep[startKey] = start

	// Priority queue: Time ascending, for same time prioritize "nodes that came from transitions with larger total consumption"
	pq := &planPQ{}
	heap.Init(pq)
	heap.Push(pq, &pqItem{
		key:        startKey,
		state:      start,
		priorityT:  start.Time,
		negTotCons: 0, // 0 because it's the start
		seq:        0,
	})

	bestTime := execmodel.Time(-1) // undetermined
	goalKeys := make([]uint64, 0, 8)

	for pq.Len() > 0 {
		it := heap.Pop(pq).(*pqItem)
		k := it.key
		s := it.state

		// Discard stale entries
		if d, ok := dist[k]; !ok || d != s.Time {
			continue
		}
		// If the best completion time is already determined and this is slower, cut off
		if bestTime >= 0 && s.Time > bestTime {
			break
		}

		// Completion check
		isCompleted := e.IsCompleted(s)
		if isCompleted {
			if bestTime < 0 || s.Time < bestTime {
				bestTime = s.Time
				goalKeys = goalKeys[:0]
				goalKeys = append(goalKeys, k)
			} else if s.Time == bestTime {
				goalKeys = append(goalKeys, k)
			}
			// Do not expand from completed state (DAG property: time is monotonically increasing)
			continue
		}

		ts := e.Transitions(s)
		if ts.Len() == 0 {
			e.Logger.Warn("fsm.Env.SearchBestPlans: deadlock found", "state", s)
			continue
		}

		// Normal expansion
		for _, tr := range ts.Iter() {
			ns := tr.NextState
			// If the best time is determined, expansion beyond it is unnecessary
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
				// Also save tied parents (create shortest DAG)
				parents[nk] = append(parents[nk], parentInfo{
					parent: k,
					alloc:  tr.Allocation,
					child:  ns,
				})
			}
		}
	}

	// Empty if there is no shortest goal
	if bestTime < 0 || len(goalKeys) == 0 {
		return sets.NewWithCapacity[*Plan](0), nil
	}

	// Full restoration from goal to start (backtrack on DAG)
	results := make([]*Plan, 0, len(goalKeys))
	memoPaths := make(map[uint64][]*Plan, len(parents)) // Memoization (suppress expansion of duplicate nodes)

	var buildAll func(k uint64) []*Plan
	buildAll = func(k uint64) []*Plan {
		if k == startKey {
			// Empty plan (transitions will be added here)
			return []*Plan{NewEmptyPlan(start)}
		}
		if v, ok := memoPaths[k]; ok {
			return v
		}
		ps := parents[k]
		if len(ps) == 0 {
			// No parent other than start = unreachable (normally doesn't happen)
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
	priorityT  execmodel.Time // ascending order
	negTotCons int            // prioritize larger total consumption at the same time (stored as negative)
	seq        int64          // for stabilization
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
