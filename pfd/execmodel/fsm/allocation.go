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

type Allocatability string

const (
	AllocatabilityOKContinuable Allocatability = "OK_CONTINUE"

	AllocatabilityOKStartable Allocatability = "OK_START"

	AllocatabilityNGInsufficientInputs Allocatability = "NG_INSUFFICIENT_INPUTS"

	AllocatabilityNGPreconditionNotMet Allocatability = "NG_PRECONDITION_NOT_MET"

	AllocatabilityNGNoDeliverableUpdates Allocatability = "NG_NO_DELIVERABLE_UPDATES"
)

func (e Allocatability) IsOK() bool {
	return e == AllocatabilityOKContinuable || e == AllocatabilityOKStartable
}

type Allocation map[pfd.AtomicProcessID]AllocationElement

func (a Allocation) Compare(b Allocation) int {
	return cmp2.CompareMap(a, b, pfd.AtomicProcessID.Compare, AllocationElement.Compare)
}

func (a Allocation) Equals(b Allocation) bool {
	return a.Compare(b) == 0
}

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

type AllocationElement struct {
	Resources      *sets.Set[ResourceID] `json:"resources"`
	ConsumedVolume Volume                `json:"consumed_volume"`
}

func MaxConsumedVolume(elements *sets.Set[AllocationElement]) Volume {
	if elements.Len() == 0 {
		return Volume(1)
	}
	maxVolume := Volume(0)
	for _, element := range elements.Iter() {
		maxVolume = max(maxVolume, element.ConsumedVolume)
	}
	return maxVolume
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

func NewAvailableAllocationsFunc(neededResourceSetsFunc NeededResourceSetsFunc) AvailableAllocationsFunc {
	return func(state State, newlyAllocatables *sets.Set[pfd.AtomicProcessID]) *sets.Set[Allocation] {

		continuingResources := sets.New[ResourceID](ResourceID.Compare)
		for _, elem := range state.AllocationShouldContinue {
			continuingResources.Union(ResourceID.Compare, elem.Resources)
		}

		type key struct {
			AtomicProcess pfd.AtomicProcessID
			Index         int
		}
		idOf := make(map[key]int)
		var allSets []AllocationElement
		for _, ap := range newlyAllocatables.Iter() {
			for i, s := range neededResourceSetsFunc(ap).Iter() {

				if !s.Resources.IsDisjointWith(ResourceID.Compare, continuingResources) {
					continue
				}
				idOf[key{AtomicProcess: ap, Index: i}] = len(allSets)
				allSets = append(allSets, s)
			}
		}

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

			dfs(i + 1)

			for idx, nr := range rs.Iter() {
				id, filtered := idOf[key{p, idx}]
				if !filtered {

					continue
				}
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

		continuingResources := sets.New[ResourceID](ResourceID.Compare)
		for _, elem := range state.AllocationShouldContinue {
			continuingResources.Union(ResourceID.Compare, elem.Resources)
		}

		type allocOption struct {
			idx            int
			ap             pfd.AtomicProcessID
			res            *sets.Set[ResourceID]
			consumedVolume Volume
		}

		options := make([]allocOption, 0, 32)
		for _, ap := range newlyAllocatables.Iter() {
			for _, entry := range neededResourceSetsFunc(ap).Iter() {
				if entry.ConsumedVolume <= 0 {
					panic(fmt.Sprintf("fsm.NewMaximalAvailableAllocationsFunc: consumed volume is zero: %v", entry))
				}

				if !entry.Resources.IsDisjointWith(ResourceID.Compare, continuingResources) {
					continue
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

			if len(state.AllocationShouldContinue) > 0 {

				return sets.New(CompareAllocationByTotalConsumedVolume, maps.Clone(state.AllocationShouldContinue))
			}
			return nil
		}

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

			union := P.Clone()
			union.Union(cmp.Compare, X)
			var pivot int
			maxRemain := -1
			for _, u := range union.Iter() {

				cand := P.Clone()
				cand.Remove(cmp.Compare, u)
				cand.Difference(cmp.Compare, conflict[u])
				if cand.Len() > maxRemain {
					maxRemain = cand.Len()
					pivot = u
				}
			}

			loopSet := P.Clone()
			keep := sets.New[int](cmp.Compare)
			keep.Union(cmp.Compare, conflict[pivot])
			keep.Add(cmp.Compare, pivot)
			loopSet.Intersection(cmp.Compare, keep)

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

				P.Remove(cmp.Compare, v)
				X.Add(cmp.Compare, v)
			}
		}

		rbk(R, P, X)

		return results
	}
}

func NewGreedyAvailableAllocationsFunc(neededResourceSetsFunc NeededResourceSetsFunc) AvailableAllocationsFunc {
	resCompare := sets.Compare(ResourceID.Compare)
	return func(state State, newlyAllocatables *sets.Set[pfd.AtomicProcessID]) *sets.Set[Allocation] {

		continuingResources := sets.New[ResourceID](ResourceID.Compare)
		for _, elem := range state.AllocationShouldContinue {
			continuingResources.Union(ResourceID.Compare, elem.Resources)
		}

		type option struct {
			ap  pfd.AtomicProcessID
			res *sets.Set[ResourceID]
			vol Volume
		}
		var options []option
		for _, ap := range newlyAllocatables.Iter() {
			for _, nr := range neededResourceSetsFunc(ap).Iter() {

				if !nr.Resources.IsDisjointWith(ResourceID.Compare, continuingResources) {
					continue
				}
				options = append(options, option{ap: ap, res: nr.Resources, vol: nr.ConsumedVolume})
			}
		}

		slices.SortFunc(options, func(a, b option) int {
			if c := cmp.Compare(b.vol, a.vol); c != 0 {
				return c
			}
			if c := pfd.AtomicProcessID.Compare(a.ap, b.ap); c != 0 {
				return c
			}
			return resCompare(a.res, b.res)
		})

		alloc := make(Allocation, len(state.AllocationShouldContinue)+len(options))
		maps.Copy(alloc, state.AllocationShouldContinue)
		used := continuingResources.Clone()
		newCount := 0
		for _, opt := range options {
			if _, ok := alloc[opt.ap]; ok {

				continue
			}
			if !opt.res.IsDisjointWith(ResourceID.Compare, used) {
				continue
			}
			alloc[opt.ap] = AllocationElement{Resources: opt.res.Clone(), ConsumedVolume: opt.vol}
			used.Union(ResourceID.Compare, opt.res)
			newCount++
		}

		if newCount == 0 && len(state.AllocationShouldContinue) == 0 {
			return sets.NewWithCapacity[Allocation](0)
		}
		return sets.New(CompareAllocationByTotalConsumedVolume, alloc)
	}
}
