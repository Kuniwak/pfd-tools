package fsm

import (
	"cmp"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"slices"
	"strconv"

	"github.com/Kuniwak/pfd-tools/cmp2"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
)

// Allocatability represents whether an atomic process can be executed.
type Allocatability string

const (
	// AllocatabilityOKContinuable means executable. Execution will continue existing work.
	AllocatabilityOKContinuable Allocatability = "OK_CONTINUE"
	// AllocatabilityOKStartable means executable. Execution will start new work.
	AllocatabilityOKStartable Allocatability = "OK_START"
	// AllocatabilityNGInsufficientInputs means not allocatable. Some input deliverables have not been generated yet.
	AllocatabilityNGInsufficientInputs Allocatability = "NG_INSUFFICIENT_INPUTS"
	// AllocatabilityNGPreconditionNotMet means not allocatable. Preconditions are not satisfied.
	AllocatabilityNGPreconditionNotMet Allocatability = "NG_PRECONDITION_NOT_MET"
	// AllocatabilityNGNoDeliverableUpdates means not allocatable. There are no deliverable updates.
	AllocatabilityNGNoDeliverableUpdates Allocatability = "NG_NO_DELIVERABLE_UPDATES"
)

// IsOK returns whether execution is possible.
func (e Allocatability) IsOK() bool {
	return e == AllocatabilityOKContinuable || e == AllocatabilityOKStartable
}

// Allocation is a dictionary from atomic processes to AllocationElements.
// Atomic processes without resource allocation are not included.
type Allocation map[pfd.AtomicProcessID]AllocationElement

// Compare compares Allocations.
func (a Allocation) Compare(b Allocation) int {
	return cmp2.CompareMap(a, b, pfd.AtomicProcessID.Compare, AllocationElement.Compare)
}

func (a Allocation) Equals(b Allocation) bool {
	return a.Compare(b) == 0
}

// CompareAllocationByTotalConsumedVolume compares Allocations.
// The one with higher total consumed work volume is considered smaller.
func CompareAllocationByTotalConsumedVolume(a Allocation, b Allocation) int {
	t1 := a.TotalConsumedVolume()
	t2 := b.TotalConsumedVolume()
	if t1 != t2 {
		return int(t2) - int(t1)
	}
	return cmp2.CompareMap(a, b, pfd.AtomicProcessID.Compare, AllocationElement.Compare)
}

func (a Allocation) Clone() Allocation {
	return maps.Clone(a)
}

// TotalConsumedVolume returns the total consumed work volume for the given Allocation.
func (a Allocation) TotalConsumedVolume() Volume {
	total := Volume(0)
	for _, element := range a {
		total += element.ConsumedVolume
	}
	return total
}

func (a Allocation) Write(w io.Writer) error {
	ks := slices.Collect(maps.Keys(a))
	if len(ks) == 0 {
		if _, err := io.WriteString(w, "(empty allocation)"); err != nil {
			return fmt.Errorf("fsm.Allocation.Write: %w", err)
		}
		return nil
	}

	slices.SortFunc(ks, pfd.AtomicProcessID.Compare)
	for _, k := range ks {
		element := a[k]
		if _, err := io.WriteString(w, string(k)); err != nil {
			return fmt.Errorf("fsm.Allocation.Write: %w", err)
		}
		if _, err := io.WriteString(w, " -> "); err != nil {
			return fmt.Errorf("fsm.Allocation.Write: %w", err)
		}
		for i, resource := range element.Resources.Iter() {
			if i > 0 {
				if _, err := io.WriteString(w, ", "); err != nil {
					return fmt.Errorf("fsm.Allocation.Write: %w", err)
				}
			}
			if _, err := io.WriteString(w, string(resource)); err != nil {
				return fmt.Errorf("fsm.Allocation.Write: %w", err)
			}
		}
		if _, err := io.WriteString(w, ", "); err != nil {
			return fmt.Errorf("fsm.Allocation.Write: %w", err)
		}
		if _, err := io.WriteString(w, strconv.Itoa(int(element.ConsumedVolume))); err != nil {
			return fmt.Errorf("fsm.Allocation.Write: %w", err)
		}
		if _, err := io.WriteString(w, ";\n"); err != nil {
			return fmt.Errorf("fsm.Allocation.Write: %w", err)
		}
	}
	return nil
}

// AllocationElement is a pair from an atomic process to the resources to allocate and the reduced work volume per unit time elapsed due to this allocation.
// Resources is never empty. ConsumedVolume is greater than 0.
type AllocationElement struct {
	Resources      *sets.Set[ResourceID] `json:"resources"`
	ConsumedVolume Volume                `json:"consumed_volume"`
}

func (a AllocationElement) Compare(b AllocationElement) int {
	c := sets.Compare(ResourceID.Compare)(a.Resources, b.Resources)
	if c != 0 {
		return c
	}
	return int(a.ConsumedVolume) - int(b.ConsumedVolume)
}

type AllocatabilityInfo struct {
	Allocatability         Allocatability                     `json:"allocatability"`
	InsufficientInputs     *sets.Set[pfd.AtomicDeliverableID] `json:"insufficient_inputs,omitempty"`
	PreconditionNotMet     string                             `json:"precondition_not_met,omitempty"`
	DeliverablesNotUpdated *sets.Set[pfd.AtomicDeliverableID] `json:"deliverables_not_updated,omitempty"`
}

func NewAllocatabilityOKContinuable() *AllocatabilityInfo {
	return &AllocatabilityInfo{Allocatability: AllocatabilityOKContinuable}
}

func NewAllocatabilityOKStartable() *AllocatabilityInfo {
	return &AllocatabilityInfo{Allocatability: AllocatabilityOKStartable}
}

func NewAllocatabilityNGInsufficientInputs(insufficientInputs *sets.Set[pfd.AtomicDeliverableID]) *AllocatabilityInfo {
	return &AllocatabilityInfo{Allocatability: AllocatabilityNGInsufficientInputs, InsufficientInputs: insufficientInputs}
}

func NewAllocatabilityNGPreconditionNotMet(preconditionNotMet string) *AllocatabilityInfo {
	return &AllocatabilityInfo{Allocatability: AllocatabilityNGPreconditionNotMet, PreconditionNotMet: preconditionNotMet}
}

func NewAllocatabilityNGNoDeliverableUpdates(deliverablesNotUpdated *sets.Set[pfd.AtomicDeliverableID]) *AllocatabilityInfo {
	return &AllocatabilityInfo{Allocatability: AllocatabilityNGNoDeliverableUpdates, DeliverablesNotUpdated: deliverablesNotUpdated.Clone()}
}

func (a *AllocatabilityInfo) Write(w io.Writer) error {
	switch a.Allocatability {
	case AllocatabilityOKContinuable:
		if _, err := io.WriteString(w, "OK_CONTINUE"); err != nil {
			return fmt.Errorf("fsm.AllocatabilityInfo.Write: %w", err)
		}
	case AllocatabilityOKStartable:
		if _, err := io.WriteString(w, "OK_START"); err != nil {
			return fmt.Errorf("fsm.AllocatabilityInfo.Write: %w", err)
		}
	case AllocatabilityNGInsufficientInputs:
		if _, err := io.WriteString(w, "NG_INSUFFICIENT_INPUTS["); err != nil {
			return fmt.Errorf("fsm.AllocatabilityInfo.Write: %w", err)
		}
		for i, input := range a.InsufficientInputs.Iter() {
			if i > 0 {
				if _, err := io.WriteString(w, ", "); err != nil {
					return fmt.Errorf("fsm.AllocatabilityInfo.Write: %w", err)
				}
			}
			if _, err := io.WriteString(w, string(input)); err != nil {
				return fmt.Errorf("fsm.AllocatabilityInfo.Write: %w", err)
			}
		}
		if _, err := io.WriteString(w, "]"); err != nil {
			return fmt.Errorf("fsm.AllocatabilityInfo.Write: %w", err)
		}
	case AllocatabilityNGPreconditionNotMet:
		if _, err := io.WriteString(w, "NG_PRECONDITION_NOT_MET"); err != nil {
			return fmt.Errorf("fsm.AllocatabilityInfo.Write: %w", err)
		}
	case AllocatabilityNGNoDeliverableUpdates:
		if _, err := io.WriteString(w, "NG_NO_DELIVERABLE_UPDATES["); err != nil {
			return fmt.Errorf("fsm.AllocatabilityInfo.Write: %w", err)
		}
		for i, input := range a.DeliverablesNotUpdated.Iter() {
			if i > 0 {
				if _, err := io.WriteString(w, ", "); err != nil {
					return fmt.Errorf("fsm.AllocatabilityInfo.Write: %w", err)
				}
			}

			if _, err := io.WriteString(w, string(input)); err != nil {
				return fmt.Errorf("fsm.AllocatabilityInfo.Write: %w", err)
			}
		}
		if _, err := io.WriteString(w, "]"); err != nil {
			return fmt.Errorf("fsm.AllocatabilityInfo.Write: %w", err)
		}
	default:
		panic(fmt.Sprintf("fsm.AllocatabilityInfo.Write: unknown allocatability: %s", a.Allocatability))
	}
	return nil
}

type AvailableAllocationsFunc func(state State, newlyAllocatables *sets.Set[pfd.AtomicProcessID]) *sets.Set[Allocation]

func NewThresholdAvailableAllocationsFunc(threshold int, neededResourceSetsFunc NeededResourceSetsFunc, logger *slog.Logger) AvailableAllocationsFunc {
	all := NewAvailableAllocationsFunc(neededResourceSetsFunc)
	maximal := NewMaximalAvailableAllocationsFunc(neededResourceSetsFunc)
	return func(state State, newlyAllocatables *sets.Set[pfd.AtomicProcessID]) *sets.Set[Allocation] {
		if threshold > 0 && newlyAllocatables.Len() > threshold {
			logger.Debug("using maximal available allocations", "threshold", threshold, "newlyAllocatables", newlyAllocatables.Len())
			return maximal(state, newlyAllocatables)
		}
		return all(state, newlyAllocatables)
	}
}

// AvailableAllocations enumerates and returns possible resource allocations in the given state.
func NewAvailableAllocationsFunc(neededResourceSetsFunc NeededResourceSetsFunc) AvailableAllocationsFunc {
	return func(state State, newlyAllocatables *sets.Set[pfd.AtomicProcessID]) *sets.Set[Allocation] {
		// Assign unique IDs to each m[ap][i]
		type key struct {
			AtomicProcess pfd.AtomicProcessID
			Index         int
		}
		idOf := make(map[key]int)
		var allSets []AllocationElement
		for _, ap := range newlyAllocatables.Iter() {
			for i, s := range neededResourceSetsFunc(ap).Iter() {
				idOf[key{AtomicProcess: ap, Index: i}] = len(allSets)
				allSets = append(allSets, s)
			}
		}

		// Pre-compute whether they are disjoint
		n := len(allSets)
		disjoint := make([][]bool, n)
		for i := range n {
			disjoint[i] = make([]bool, n)
			for j := range n {
				if i == j {
					continue
				}
				disjoint[i][j] = allSets[i].Resources.IsDisjointWith(ResourceID.Compare, allSets[j].Resources)
			}
		}

		res := sets.NewWithCapacity[Allocation](0)
		cur := make(Allocation)
		var chosenIDs []int

		var dfs func(int)
		dfs = func(i int) {
			if i == newlyAllocatables.Len() {
				if len(cur) > 0 || len(state.AllocationShouldContinue) > 0 {
					cp := make(Allocation, len(cur)+len(state.AllocationShouldContinue))
					maps.Copy(cp, cur)
					maps.Copy(cp, state.AllocationShouldContinue)
					res.Add(CompareAllocationByTotalConsumedVolume, cp)
				}
				return
			}
			p, ok := newlyAllocatables.At(i)
			if !ok {
				panic(fmt.Sprintf("fsm.Env.NewlyAllocatables: index out of bounds: %d on (%v)", i, newlyAllocatables.Slice()))
			}
			rs := neededResourceSetsFunc(p)

			// Case where no element is added
			dfs(i + 1)

			// Case where an element is added
			for idx, nr := range rs.Iter() {
				id := idOf[key{p, idx}]
				ok := true
				for _, cid := range chosenIDs {
					if !disjoint[id][cid] {
						ok = false
						break
					}
				}
				if ok {
					cur[p] = AllocationElement{Resources: nr.Resources, ConsumedVolume: nr.ConsumedVolume}
					chosenIDs = append(chosenIDs, id)
					dfs(i + 1)
					chosenIDs = chosenIDs[:len(chosenIDs)-1]
					delete(cur, p)
				}
			}
		}

		dfs(0)
		return res
	}
}

func NewMaximalAvailableAllocationsFunc(neededResourceSetsFunc NeededResourceSetsFunc) AvailableAllocationsFunc {
	return func(state State, newlyAllocatables *sets.Set[pfd.AtomicProcessID]) *sets.Set[Allocation] {
		type allocOption struct {
			idx            int
			ap             pfd.AtomicProcessID
			res            *sets.Set[ResourceID]
			consumedVolume Volume
		}

		// Allocation candidates for each AP (NeededResourceSets that are contained within avail)
		options := make([]allocOption, 0, 32)
		for _, ap := range newlyAllocatables.Iter() {
			for _, entry := range neededResourceSetsFunc(ap).Iter() {
				if entry.ConsumedVolume <= 0 {
					panic(fmt.Sprintf("fsm.NewMaximalAvailableAllocationsFunc: consumed volume is zero: %v", entry))
				}
				options = append(options, allocOption{
					idx:            len(options),
					ap:             ap,
					res:            entry.Resources.Clone(),
					consumedVolume: entry.ConsumedVolume,
				})
			}
		}

		if len(options) == 0 {
			// Nothing can be allocated due to resource constraints
			return nil
		}

		// 4) Create conflict graph (same AP or shared resource sets)
		//    conflict[i] is the set of vertices that conflict with i
		conflict := make([]*sets.Set[int], len(options))
		for i := range options {
			conflict[i] = sets.New(cmp.Compare, i)
		}
		for i := 0; i < len(options); i++ {
			for j := i + 1; j < len(options); j++ {
				sameAP := options[i].ap == options[j].ap
				shareRes := !options[i].res.IsDisjointWith(ResourceID.Compare, options[j].res)
				if sameAP || shareRes {
					conflict[i].Add(cmp.Compare, j)
					conflict[j].Add(cmp.Compare, i)
				}
			}
		}

		// 5) Bron–Kerbosch (maximum clique of complement graph ≒ maximum independent set of conflict graph)
		// R: current clique (mutually non-conflicting set)
		// P: set of vertices that can be added next
		// X: set of vertices already considered in this branch
		all := sets.New[int](cmp.Compare)
		for i := range options {
			all.Add(cmp.Compare, i)
		}
		P := all.Clone()
		X := sets.New[int](cmp.Compare)
		R := sets.New[int](cmp.Compare)

		results := sets.NewWithCapacity[Allocation](0)

		var rbk func(R, P, X *sets.Set[int])
		rbk = func(R, P, X *sets.Set[int]) {
			if P.Len() == 0 && X.Len() == 0 {
				// Maximal (cannot add any more)
				alloc := maps.Clone(state.AllocationShouldContinue)
				for _, i := range R.Iter() {
					opt := options[i]
					alloc[opt.ap] = AllocationElement{
						Resources:      opt.res.Clone(),
						ConsumedVolume: opt.consumedVolume,
					}
				}
				results.Add(CompareAllocationByTotalConsumedVolume, alloc)
				return
			}

			// Pivot selection: choose u from P∪X, and loop target is P \ N̄(u) = P ∩ (conflict[u] ∪ {u})
			// N̄(u) is the adjacency set of the complement graph (= non-conflicting vertices). Calculated using conflict.
			union := P.Clone()
			union.Union(cmp.Compare, X)
			var pivot int
			maxRemain := -1
			for _, u := range union.Iter() {
				// Size of P ∩ N̄(u) = Size after removing u and its conflict set from P
				cand := P.Clone()
				cand.Remove(cmp.Compare, u)
				cand.Difference(cmp.Compare, conflict[u])
				if cand.Len() > maxRemain {
					maxRemain = cand.Len()
					pivot = u
				}
			}

			// Loop target: P \ N̄(u) = P ∩ (conflict[u] ∪ {u})
			loopSet := P.Clone()
			keep := sets.New[int](cmp.Compare)
			keep.Union(cmp.Compare, conflict[pivot])
			keep.Add(cmp.Compare, pivot)
			loopSet.Intersection(cmp.Compare, keep)

			// Bron–Kerbosch: for v in P \ N̄(u)
			//  1) R'=R∪{v}
			//  2) P'=P ∩ N̄(v) = P \ ({v} ∪ conflict[v])
			//  3) X'=X ∩ N̄(v) = X \ ({v} ∪ conflict[v])
			for _, v := range loopSet.Iter() {
				Rp := R.Clone()
				Rp.Add(cmp.Compare, v)

				Pp := P.Clone()
				Pp.Remove(cmp.Compare, v)
				Pp.Difference(cmp.Compare, conflict[v])

				Xp := X.Clone()
				Xp.Remove(cmp.Compare, v)
				Xp.Difference(cmp.Compare, conflict[v])

				rbk(Rp, Pp, Xp)

				// Post-process: remove v from P and move to X
				P.Remove(cmp.Compare, v)
				X.Add(cmp.Compare, v)
			}
		}

		rbk(R, P, X)

		return results
	}
}
