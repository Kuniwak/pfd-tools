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

// Quality is a parameter that summarizes search scale and heuristic weights (speed vs. accuracy trade-off).
type Quality struct {
	// Upper limit on node expansion. Roughly corresponds to the upper limit of computation time (higher = more accuracy, more time).
	NodeBudget int

	// Upper limit on transitions (allocations) considered in each state. Higher values make the search broader.
	// Unlimited if 0 or negative.
	TopKPerState int

	// Weighted A* weight (>=1). Larger values bias more strongly toward "likely to finish early" (fast but rough).
	Weight float64

	// Upper limit on the number of plans to return. Specify when you want multiple candidates.
	MaxResults int

	// Randomization for tie-breaking and TopK selection. Deterministic if left at 0.
	RandomSeed int64

	// Number of random restarts (for diversity improvement). No restarts if 0.
	Restarts int
}

func SearchBetterPlans(q Quality) SearchFunc {
	return func(e *Env) (*sets.Set[*Plan], error) {
		return searchBetterPlans(e, q)
	}
}

// searchBetterPlans quickly finds "better" execution plans.
//   - Does not guarantee optimality.
//   - Can control search scale with Quality.
//   - To get multiple candidates, increase MaxResults.
//   - If no goal is found within the search budget, return at least one plan
//     as a fallback using greedy Gantt generation.
func searchBetterPlans(e *Env, q Quality) (*sets.Set[*Plan], error) {
	normalizeQuality(&q)

	// Restarts for diversity improvement (optional)
	results := make([]*Plan, 0, max(1, q.MaxResults))
	for trial := 0; trial < max(1, q.Restarts+1); trial++ {
		seed := q.RandomSeed
		if q.RandomSeed != 0 {
			seed = q.RandomSeed + int64(trial)*1315423911
		}
		plans := e.searchBetterPlansOnce(q, seed)
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

// One iteration of Weighted A*.
func (e *Env) searchBetterPlansOnce(q Quality, seed int64) []*Plan {
	rng := rand.New(rand.NewSource(seed))

	start := e.InitialState()
	h := &maphash.Hash{}
	if err := HashStateWithoutTime(start, h); err != nil {
		e.Logger.Warn(fmt.Sprintf("fsm.Env.SearchBetterPlans: %v", err))
		return nil
	}
	startKey := h.Sum64()

	// Record best g-values (time) to prune inferior solutions
	bestG := map[uint64]execmodel.Time{startKey: start.Time}
	parents := make(map[uint64]parentInfo, 1024)
	stateRep := map[uint64]State{startKey: start}

	// Open: f(s) minimum heap
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

		// Discard if stale
		if bg, ok := bestG[k]; !ok || bg != g {
			continue
		}
		expansions++

		// Expand
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
				continue // Existing one is better or equivalent
			}
			bestG[nk] = newG
			stateRep[nk] = ns
			parents[nk] = parentInfo{
				parent: k,
				alloc:  tr.Allocation,
				child:  ns,
			}

			// Restore goal (completed state) as soon as found
			if e.IsCompleted(ns) {
				if plan, ok := buildPlan(startKey, nk, parents, start); ok {
					found = append(found, plan)
					if len(found) >= q.MaxResults {
						break
					}
				}
				// Continue searching for alternative solutions
				continue
			}

			// Regular node
			fv := float64(newG) + q.Weight*float64(e.heuristicLB(ns))
			heap.Push(pq, &waItem{
				key: nk, state: ns,
				g: newG, f: fv,
				seq: pq.nextSeq(),
			})
		}
	}

	return found
}

// ===== Heuristic (lower bound target: remaining time approximation) ==========================

// heuristicLB approximates "the minimum time probably needed from the current state to completion".
//   - The wait time until the maximum available time of initial deliverables (for those not yet arrived) is always required.
//   - Add remaining work volume / maximum total throughput achievable at the current time (optimistic value).
//     ※ This is a "lenient" estimate that ignores resource competition and dependencies, so it doesn't guarantee optimality but is effective as a search guideline.
func (e *Env) heuristicLB(s State) execmodel.Time {
	// 1) Wait until the "maximum available time of initial deliverables that haven't arrived yet"
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

	// 2) Remaining work volume and "maximum total throughput achievable now"
	var total Volume
	for _, v := range s.RemainedVolumeMap {
		total += v
	}

	// Among allocations determined from NewlyAllocatables, the one with maximum instantaneous total throughput
	newly := e.NewlyAllocatables(s)
	allocs := e.AvailableAllocationsFunc(s, newly)
	maxTV := Volume(0)
	for _, a := range allocs.Iter() {
		if tv := a.TotalConsumedVolume(); tv > maxTV {
			maxTV = tv
		}
	}
	// If nothing can be allocated, we need to wait at least until "the time when something happens"
	// wait is included in 1). Here we treat the work lower bound as 0.
	work := execmodel.Time(0)
	if maxTV > 0 && total > 0 {
		work = execmodel.Time(math.Ceil(float64(total) / float64(maxTV)))
	}

	return wait + work
}

// transitionsSortedForHeuristic arranges transitions in the order of heuristically "likely to advance".
// Here we prioritize "high total consumed work volume (instantaneous throughput)" and "early next time".
// Ties are lightly shuffled with randomization for stabilization.
func (e *Env) transitionsSortedForHeuristic(s State, rng *rand.Rand) []*Trans {
	set := e.Transitions(s)
	trs := make([]*Trans, 0, set.Len())
	for _, tr := range set.Iter() {
		trs = append(trs, tr)
	}
	if len(trs) <= 1 {
		return trs
	}

	// Light pre-shuffle (deterministic when random seed is 0)
	if rng != nil && rng.Int63() != 0 {
		rng.Shuffle(len(trs), func(i, j int) { trs[i], trs[j] = trs[j], trs[i] })
	}

	slices.SortFunc(trs, func(a, b *Trans) int {
		// 1) Instantaneous total throughput (descending)
		if ta, tb := a.Allocation.TotalConsumedVolume(), b.Allocation.TotalConsumedVolume(); ta != tb {
			if ta > tb {
				return -1
			}
			return 1
		}
		// 2) Next time (ascending)
		if a.NextState.Time != b.NextState.Time {
			if a.NextState.Time < b.NextState.Time {
				return -1
			}
			return 1
		}
		// 3) Hash for stabilization (ascending)
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

// ====== Plan restoration/fallback ============================================

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
	// Reverse order to forward order
	slices.Reverse(path)
	pl := make([]*Trans, len(path))
	copy(pl, path)
	return &Plan{
		InitialState: initialState,
		Transitions:  pl,
	}, true
}

// ====== PQ for WA* ==========================================================

type waItem struct {
	key   uint64
	state State
	g     execmodel.Time // Real time
	f     float64        // Priority (smaller is better)
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
