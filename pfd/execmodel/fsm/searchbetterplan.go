package fsm

import (
	"container/heap"
	"fmt"
	"hash/maphash"
	"math"
	"math/rand"
	"slices"

	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/sets"
)

type Quality struct {
	NodeBudget int

	TopKPerState int

	Weight float64

	MaxResults int

	RandomSeed int64

	Restarts int
}

func SearchBetterPlansWithPrefix(q Quality) SearchWithPrefixFunc {
	return func(e *Env, prefix *Plan) (*sets.Set[*Plan], error) {
		var start State
		var prefixPlan *Plan
		if prefix != nil {
			replayed, replayedState, err := ReplayPrefix(e, prefix)
			if err != nil {
				return nil, fmt.Errorf("fsm.SearchBetterPlansWithPrefix: %w", err)
			}
			start = replayedState
			prefixPlan = replayed
		} else {
			start = e.InitialState()
		}

		results, err := searchBetterPlansFromState(e, q, start)
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

func SearchBetterPlans(q Quality) SearchFunc {
	return func(e *Env) (*sets.Set[*Plan], error) {
		return searchBetterPlans(e, q)
	}
}

func searchBetterPlans(e *Env, q Quality) (*sets.Set[*Plan], error) {
	start := e.InitialState()
	return searchBetterPlansFromState(e, q, start)
}

func searchBetterPlansFromState(e *Env, q Quality, start State) (*sets.Set[*Plan], error) {
	normalizeQuality(&q)

	results := make([]*Plan, 0, max(1, q.MaxResults))
	for trial := 0; trial < max(1, q.Restarts+1); trial++ {
		e.Logger.Debug("fsm.Env.SearchBetterPlans: trial", "trial", trial, "results", len(results))
		seed := q.RandomSeed
		if q.RandomSeed != 0 {
			seed = q.RandomSeed + int64(trial)*1315423911
		}
		plans := e.searchBetterPlansOnceFromState(q, seed, start)
		results = append(results, plans...)
		if len(results) >= q.MaxResults {
			break
		}
	}

	return sets.New((*Plan).Compare, results...), nil
}

func normalizeQuality(q *Quality) {
	if q.NodeBudget <= 0 {
		q.NodeBudget = 10_000
	}
	if q.TopKPerState < 0 {
		q.TopKPerState = 0
	}
	if q.Weight < 1.0 {
		q.Weight = 1.0
	}
	if q.MaxResults <= 0 {
		q.MaxResults = 3
	}
}

type parentInfo struct {
	parent uint64
	alloc  Allocation
	child  State
}

func (e *Env) searchBetterPlansOnce(q Quality, seed int64) []*Plan {
	return e.searchBetterPlansOnceFromState(q, seed, e.InitialState())
}

func (e *Env) searchBetterPlansOnceFromState(q Quality, seed int64, start State) []*Plan {
	rng := rand.New(rand.NewSource(seed))
	h := &maphash.Hash{}
	if err := HashStateWithoutTime(start, h); err != nil {
		e.Logger.Warn(fmt.Sprintf("fsm.Env.SearchBetterPlans: %v", err))
		return nil
	}
	startKey := h.Sum64()

	bestG := map[uint64]execmodel.Time{startKey: start.Time}
	parents := make(map[uint64]parentInfo, 1024)
	stateRep := map[uint64]State{startKey: start}

	pq := &waPQ{}
	heap.Init(pq)
	heap.Push(pq, &waItem{
		key: startKey, state: start,
		g:   start.Time,
		f:   float64(start.Time) + q.Weight*float64(e.heuristicLB(start)),
		seq: pq.nextSeq(),
	})

	found := make([]*Plan, 0, max(1, q.MaxResults))
	expansions := 0

	for pq.Len() > 0 && expansions < q.NodeBudget && len(found) < q.MaxResults {
		item := heap.Pop(pq).(*waItem)
		k, s, g := item.key, item.state, item.g

		if bg, ok := bestG[k]; !ok || bg != g {
			continue
		}
		expansions++
		if expansions%searchProgressLogInterval == 0 {
			e.Logger.Debug("fsm.Env.SearchBetterPlans: progress", "expansions", expansions, "budget", q.NodeBudget, "pqLen", pq.Len(), "found", len(found), "time", float64(s.Time))
		}

		trs := e.transitionsSortedForHeuristic(s, rng)
		if q.TopKPerState > 0 && len(trs) > q.TopKPerState {
			trs = trs[:q.TopKPerState]
		}

		if len(trs) == 0 && !e.IsCompleted(s) {
			e.Logger.Warn("fsm.Env.SearchBetterPlans: deadlock found", "state", s)
			continue
		}

		for _, tr := range trs {
			ns := tr.NextState
			h.Reset()
			if err := HashStateWithoutTime(ns, h); err != nil {
				e.Logger.Warn(fmt.Sprintf("fsm.Env.SearchBetterPlans: %v", err))
				continue
			}
			nk := h.Sum64()
			newG := ns.Time

			if old, ok := bestG[nk]; ok && newG >= old {
				continue
			}
			bestG[nk] = newG
			stateRep[nk] = ns
			parents[nk] = parentInfo{
				parent: k,
				alloc:  tr.Allocation,
				child:  ns,
			}

			if e.IsCompleted(ns) {
				if plan, ok := buildPlan(startKey, nk, parents, start); ok {
					found = append(found, plan)
					if len(found) >= q.MaxResults {
						break
					}
				}

				continue
			}

			fv := float64(newG) + q.Weight*float64(e.heuristicLB(ns))
			heap.Push(pq, &waItem{
				key: nk, state: ns,
				g: newG, f: fv,
				seq: pq.nextSeq(),
			})
		}
	}

	e.Logger.Debug("fsm.Env.SearchBetterPlans: done", "expansions", expansions, "found", len(found))
	return found
}

func (e *Env) heuristicLB(s State) execmodel.Time {

	var maxAvail execmodel.Time
	for _, d := range e.PFD.InitialDeliverables().Iter() {
		t := e.DeliverableAvailableTimeFunc(d)
		if t > maxAvail {
			maxAvail = t
		}
	}
	wait := execmodel.Time(0)
	if s.Time < maxAvail {
		wait = maxAvail - s.Time
	}

	var total Volume
	for _, v := range s.RemainedVolumeMap {
		total += v
	}

	newly := e.NewlyAllocatables(s)
	allocs := e.AvailableAllocationsFunc(s, newly)
	maxTV := Volume(0)
	for _, a := range allocs.Iter() {
		if tv := a.TotalConsumedVolume(); tv > maxTV {
			maxTV = tv
		}
	}

	work := execmodel.Time(0)
	if maxTV > 0 && total > 0 {
		work = execmodel.Time(math.Ceil(float64(total) / float64(maxTV)))
	}

	return wait + work
}

func (e *Env) transitionsSortedForHeuristic(s State, rng *rand.Rand) []*Trans {
	set := e.Transitions(s)
	trs := make([]*Trans, 0, set.Len())
	for _, tr := range set.Iter() {
		trs = append(trs, tr)
	}
	if len(trs) <= 1 {
		return trs
	}

	if rng != nil && rng.Int63() != 0 {
		rng.Shuffle(len(trs), func(i, j int) { trs[i], trs[j] = trs[j], trs[i] })
	}

	slices.SortFunc(trs, func(a, b *Trans) int {

		if ta, tb := a.Allocation.TotalConsumedVolume(), b.Allocation.TotalConsumedVolume(); ta != tb {
			if ta > tb {
				return -1
			}
			return 1
		}

		if a.NextState.Time != b.NextState.Time {
			if a.NextState.Time < b.NextState.Time {
				return -1
			}
			return 1
		}

		ha := hashTrans(a)
		hb := hashTrans(b)
		if ha < hb {
			return -1
		} else if ha > hb {
			return 1
		}
		return 0
	})
	return trs
}

func hashTrans(t *Trans) uint64 {
	var h maphash.Hash
	_, _ = h.WriteString(fmt.Sprintf("%v|%v", t.Allocation, t.NextState.Time))
	return h.Sum64()
}

func buildPlan(startKey, goalKey uint64, parents map[uint64]parentInfo, initialState State) (*Plan, bool) {
	if startKey == goalKey {
		return NewEmptyPlan(initialState), true
	}
	path := make([]*Trans, 0, 32)
	for k := goalKey; k != startKey; {
		p, ok := parents[k]
		if !ok {
			return nil, false
		}
		path = append(path, &Trans{
			Allocation: p.alloc,
			NextState:  p.child,
		})
		k = p.parent
	}

	slices.Reverse(path)
	pl := make([]*Trans, len(path))
	copy(pl, path)
	return &Plan{
		InitialState: initialState,
		Transitions:  pl,
	}, true
}

type waItem struct {
	key   uint64
	state State
	g     execmodel.Time
	f     float64
	seq   int64
	index int
}
type waPQ struct {
	data []*waItem
	seqc int64
}

func (q *waPQ) Len() int { return len(q.data) }
func (q *waPQ) Less(i, j int) bool {
	a, b := q.data[i], q.data[j]
	if a.f != b.f {
		return a.f < b.f
	}
	if a.g != b.g {
		return a.g < b.g
	}
	return a.seq < b.seq
}
func (q *waPQ) Swap(i, j int) {
	q.data[i], q.data[j] = q.data[j], q.data[i]
	q.data[i].index = i
	q.data[j].index = j
}
func (q *waPQ) Push(x any) {
	it := x.(*waItem)
	it.index = len(q.data)
	q.data = append(q.data, it)
}
func (q *waPQ) Pop() any {
	n := len(q.data)
	it := q.data[n-1]
	q.data[n-1] = nil
	q.data = q.data[:n-1]
	return it
}
func (q *waPQ) nextSeq() int64 {
	q.seqc++
	return q.seqc
}
