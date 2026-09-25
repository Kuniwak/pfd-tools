package fsm

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"math"
	"slices"
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/sets"
)

type Env struct {
	PFD *pfd.ValidPFD

	AvailableResources *sets.Set[ResourceID]

	AvailableAllocationsFunc AvailableAllocationsFunc

	InitialVolumeFunc InitialVolumeFunc

	ReworkVolumeFunc ReworkVolumeFunc

	FeedbackSourceMaxRevision map[pfd.AtomicDeliverableID]int

	PreconditionMap map[pfd.AtomicProcessID]*Precondition

	NeededResourceSetsFunc NeededResourceSetsFunc

	DeliverableAvailableTimeFunc DeliverableAvailableTimeFunc

	Memoized *Memoized

	Logger *slog.Logger
}

func NewEnv(
	pfd *pfd.ValidPFD,
	availableResources *sets.Set[ResourceID],
	availableAllocationsFunc AvailableAllocationsFunc,
	initialVolumeFunc InitialVolumeFunc,
	reworkVolumeFunc ReworkVolumeFunc,
	feedbackSourceMaxRevision map[pfd.AtomicDeliverableID]int,
	preconditionMap map[pfd.AtomicProcessID]*Precondition,
	neededResourceSetsFunc NeededResourceSetsFunc,
	deliverableAvailableTimeFunc DeliverableAvailableTimeFunc,
	logger *slog.Logger,
) *Env {
	return &Env{
		PFD:                          pfd,
		AvailableResources:           availableResources,
		AvailableAllocationsFunc:     availableAllocationsFunc,
		InitialVolumeFunc:            initialVolumeFunc,
		ReworkVolumeFunc:             reworkVolumeFunc,
		FeedbackSourceMaxRevision:    feedbackSourceMaxRevision,
		PreconditionMap:              preconditionMap,
		NeededResourceSetsFunc:       neededResourceSetsFunc,
		DeliverableAvailableTimeFunc: deliverableAvailableTimeFunc,
		Memoized:                     NewMemoized(),
		Logger:                       logger,
	}
}

func (e *Env) Clone() *Env {
	return NewEnv(
		e.PFD.Clone(),
		e.AvailableResources.Clone(),
		e.AvailableAllocationsFunc,
		e.InitialVolumeFunc,
		e.ReworkVolumeFunc,
		maps.Clone(e.FeedbackSourceMaxRevision),
		maps.Clone(e.PreconditionMap),
		e.NeededResourceSetsFunc,
		e.DeliverableAvailableTimeFunc,
		e.Logger,
	)
}

func (e *Env) FreeResources(state State) *sets.Set[ResourceID] {
	avail := e.AvailableResources.Clone()
	allocations := state.AllocationShouldContinue
	for _, alloc := range allocations {
		avail.Difference(ResourceID.Compare, alloc.Resources)
	}
	return avail
}

func (e *Env) AllocatabilityInfo(
	ap pfd.AtomicProcessID,
	remainedVolumeMap map[pfd.AtomicProcessID]Volume,
	revisionMap map[pfd.AtomicDeliverableID]int,
	allocationShouldContinue Allocation,
	updatedDeliverablesNotHandled map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID],
) *AllocatabilityInfo {
	if _, ok := allocationShouldContinue[ap]; ok {

		return &AllocatabilityInfo{Allocatability: AllocatabilityOKContinuable}
	}

	insufficientInputs := sets.NewWithCapacity[pfd.AtomicDeliverableID](e.PFD.AtomicDeliverables.Len())
	for _, d := range e.PFD.InputDeliverablesExceptFeedback(ap).Iter() {
		revision, ok := revisionMap[d]
		if !ok {
			panic(fmt.Sprintf("fsm.Env.Allocatability: missing deliverable in revisionMap: %q", d))
		}

		if revision == 0 {
			insufficientInputs.Add(pfd.AtomicDeliverableID.Compare, d)
		}
	}
	if insufficientInputs.Len() > 0 {

		return &AllocatabilityInfo{
			Allocatability:     AllocatabilityNGInsufficientInputs,
			InsufficientInputs: insufficientInputs,
		}
	}

	ds, ok := updatedDeliverablesNotHandled[ap]
	if !ok {
		panic(fmt.Sprintf("fsm.Env.Allocatability: missing updated deliverables: %q", ap))
	}
	if ds.Len() == 0 {
		return &AllocatabilityInfo{
			Allocatability:         AllocatabilityNGNoDeliverableUpdates,
			DeliverablesNotUpdated: e.PFD.InputDeliverablesIncludingFeedback(ap),
		}
	}

	precondition, ok := e.PreconditionMap[ap]
	if !ok {
		panic(fmt.Sprintf("fsm.Env.Allocatability: missing precondition: %q", ap))
	}
	e.Memoized.StringBuilder.Reset()
	r := precondition.Eval(e, remainedVolumeMap, revisionMap, allocationShouldContinue, updatedDeliverablesNotHandled)
	e.Memoized.StringBuilder.Reset()
	r.Write(e.Memoized.StringBuilder)
	if !r.Result {
		return &AllocatabilityInfo{
			Allocatability:     AllocatabilityNGPreconditionNotMet,
			PreconditionNotMet: e.Memoized.StringBuilder.String(),
		}
	}

	return &AllocatabilityInfo{Allocatability: AllocatabilityOKStartable}
}

func (e *Env) Allocatability(
	ap pfd.AtomicProcessID,
	remainedVolumeMap map[pfd.AtomicProcessID]Volume,
	revisionMap map[pfd.AtomicDeliverableID]int,
	allocationShouldContinue Allocation,
	updatedDeliverablesNotHandled map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID],
) Allocatability {
	info := e.AllocatabilityInfo(ap, remainedVolumeMap, revisionMap, allocationShouldContinue, updatedDeliverablesNotHandled)
	return info.Allocatability
}

func (e *Env) AllocatabilityInfoMap(state State) AllocatabilityInfoMap {
	res := make(AllocatabilityInfoMap, e.PFD.AtomicProcesses.Len())
	for _, ap := range e.PFD.AtomicProcesses.Iter() {
		res[ap] = e.AllocatabilityInfo(
			ap,
			state.RemainedVolumeMap,
			state.RevisionMap,
			state.AllocationShouldContinue,
			state.UpdatedDeliverablesNotHandled,
		)
	}
	return res
}

func (e *Env) NewlyAllocatables(state State) *sets.Set[pfd.AtomicProcessID] {
	res := sets.NewWithCapacity[pfd.AtomicProcessID](e.PFD.AtomicProcesses.Len())
	for _, ap := range e.PFD.AtomicProcesses.Iter() {
		a := e.Allocatability(ap, state.RemainedVolumeMap, state.RevisionMap, state.AllocationShouldContinue, state.UpdatedDeliverablesNotHandled)
		switch a {
		case AllocatabilityOKStartable:
			res.Add(pfd.AtomicProcessID.Compare, ap)
		case AllocatabilityOKContinuable, AllocatabilityNGInsufficientInputs, AllocatabilityNGPreconditionNotMet, AllocatabilityNGNoDeliverableUpdates:

		default:
			panic(fmt.Sprintf("fsm.Env.NewlyAllocatables: unknown allocatability: %s", a))
		}
	}
	return res
}

func (e *Env) MinimumCompletedTime(currentTime execmodel.Time, remainedVolumeMap map[pfd.AtomicProcessID]Volume, allocation Allocation) (execmodel.Time, bool) {
	minTime := execmodel.Time(math.MaxFloat64)
	for ap, alloc := range allocation {
		remainedVolume, ok := remainedVolumeMap[ap]
		if !ok {
			panic(fmt.Sprintf("fsm.Env.MinimumCompletedTime: missing remained volume: %q", ap))
		}

		restTime := execmodel.Time(0)
		if !remainedVolume.IsZero() {
			restTime = execmodel.Time(float64(remainedVolume) / float64(alloc.ConsumedVolume))
		}
		if restTime < minTime {
			minTime = restTime
		}
	}
	return currentTime + minTime, minTime != execmodel.Time(math.MaxFloat64)
}

func (e *Env) NewRevisionMap(
	newlyAvailableInitialDeliverables *sets.Set[pfd.AtomicDeliverableID],
	pastRevisionMap map[pfd.AtomicDeliverableID]int,
	completedAtomicProcesses *sets.Set[pfd.AtomicProcessID],
) (map[pfd.AtomicDeliverableID]int, *sets.Set[pfd.AtomicDeliverableID]) {
	newRevisionMap := maps.Clone(pastRevisionMap)
	revisionsUpdated := newlyAvailableInitialDeliverables.Clone()

	for _, ap := range completedAtomicProcesses.Iter() {
		revisionsUpdated.Union(pfd.AtomicDeliverableID.Compare, e.PFD.OutputDeliverables(ap))
	}

	for _, d := range revisionsUpdated.Iter() {
		newRevisionMap[d]++
	}

	return newRevisionMap, revisionsUpdated
}

func (e *Env) NewlyAvailableInitialDeliverables(t execmodel.Time) *sets.Set[pfd.AtomicDeliverableID] {
	ids := sets.NewWithCapacity[pfd.AtomicDeliverableID](e.PFD.InitialDeliverables().Len())
	for _, id := range e.PFD.InitialDeliverables().Iter() {
		if t != e.DeliverableAvailableTimeFunc(id) {
			continue
		}
		ids.Add(pfd.AtomicDeliverableID.Compare, id)
	}
	return ids
}

func (e *Env) NewRemainedVolumeMap(remainedVolumeMap map[pfd.AtomicProcessID]Volume, allocation Allocation, timeDelta execmodel.Time) map[pfd.AtomicProcessID]Volume {
	newRemainedVolumeMap := maps.Clone(remainedVolumeMap)

	for ap, elem := range allocation {
		remainedVolume, ok := remainedVolumeMap[ap]
		if !ok {
			panic(fmt.Sprintf("fsm.Env.NewRemainedVolumeMap: missing remained volume: %q", ap))
		}

		newRemainedVolume := max(remainedVolume-Volume(float64(elem.ConsumedVolume)*float64(timeDelta)), 0)
		if newRemainedVolume.IsZero() {
			newRemainedVolume = Volume(0)
		}
		newRemainedVolumeMap[ap] = newRemainedVolume
	}

	return newRemainedVolumeMap
}

func (e *Env) UpdateNumberOfReworksMap(
	numberOfReworksMap map[pfd.AtomicProcessID]int,
	completedAtomicProcesses *sets.Set[pfd.AtomicProcessID],
) map[pfd.AtomicProcessID]int {
	newNumOfReworksMap := maps.Clone(numberOfReworksMap)
	for _, ap := range completedAtomicProcesses.Iter() {
		newNumOfReworksMap[ap]++
	}
	return newNumOfReworksMap
}

func (e *Env) CollectCompletedAtomicProcesses(
	allocation Allocation,
	newRemainedVolumeMap map[pfd.AtomicProcessID]Volume,
	res *sets.Set[pfd.AtomicProcessID],
) {
	for ap := range allocation {
		remainedVolume, ok := newRemainedVolumeMap[ap]
		if !ok {
			panic(fmt.Sprintf("fsm.Env.CollectCompletedAtomicProcesses: missing remained volume: %q", ap))
		}
		if remainedVolume.IsZero() {
			res.Add(pfd.AtomicProcessID.Compare, ap)
		}
	}
}

func (e *Env) UpdatedDeliverables(completedAtomicProcesses *sets.Set[pfd.AtomicProcessID]) *sets.Set[pfd.AtomicDeliverableID] {
	updatedDeliverables := sets.NewWithCapacity[pfd.AtomicDeliverableID](e.PFD.AtomicDeliverables.Len())

	for _, ap := range completedAtomicProcesses.Iter() {
		updatedDeliverables.Union(pfd.AtomicDeliverableID.Compare, e.PFD.OutputDeliverables(ap))
	}

	return updatedDeliverables
}

func (e *Env) NewUpdateDeliverablesNotHandled(
	newlyAvailableInitialDeliverables *sets.Set[pfd.AtomicDeliverableID],
	pastUpdatedDeliverablesNotHandled map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID],
	newRevisionMap map[pfd.AtomicDeliverableID]int,
	executedAtomicProcesses *sets.Set[pfd.AtomicProcessID],
	completedAtomicProcesses *sets.Set[pfd.AtomicProcessID],
) map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID] {
	newUpdatedDeliverablesNotHandled := make(map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID], e.PFD.AtomicProcesses.Len())
	for _, ap := range e.PFD.AtomicProcesses.Iter() {
		if executedAtomicProcesses.Contains(pfd.AtomicProcessID.Compare, ap) {

			newUpdatedDeliverablesNotHandled[ap] = sets.New(pfd.AtomicDeliverableID.Compare)
		} else {

			newUpdatedDeliverablesNotHandled[ap] = pastUpdatedDeliverablesNotHandled[ap].Clone()
		}
	}

	for _, d := range newlyAvailableInitialDeliverables.Iter() {
		for _, ap := range e.PFD.EitherFeedbackOrNotDestinationAtomicProcesses(d).Iter() {
			newUpdatedDeliverablesNotHandled[ap].Add(pfd.AtomicDeliverableID.Compare, d)
		}
	}

	for _, ap := range completedAtomicProcesses.Iter() {

		for _, d := range e.PFD.OutputDeliverables(ap).Iter() {
			curRevision, ok := newRevisionMap[d]
			if !ok {
				panic(fmt.Sprintf("fsm.Env.NewUpdateDeliverablesNotHandled: missing revision: %q", d))
			}

			maxRevision, ok := e.FeedbackSourceMaxRevision[d]
			if ok && curRevision >= maxRevision {

				continue
			}

			for _, ap2 := range e.PFD.EitherFeedbackOrNotDestinationAtomicProcesses(d).Iter() {
				newUpdatedDeliverablesNotHandled[ap2].Add(pfd.AtomicDeliverableID.Compare, d)
			}
		}
	}
	return newUpdatedDeliverablesNotHandled
}

func (e *Env) InitialState() State {
	t := execmodel.Time(0)

	numOfReworksMap := make(map[pfd.AtomicProcessID]int, e.PFD.AtomicProcesses.Len())
	for _, ap := range e.PFD.AtomicProcesses.Iter() {
		numOfReworksMap[ap] = 0
	}

	emptyAllocation := Allocation{}

	newRevisionMap := make(map[pfd.AtomicDeliverableID]int, e.PFD.AtomicDeliverables.Len())
	for _, d := range e.PFD.AtomicDeliverables.Iter() {
		newRevisionMap[d] = 0
	}
	newlyAvailableInitialDeliverables := e.NewlyAvailableInitialDeliverables(t)
	for _, d := range newlyAvailableInitialDeliverables.Iter() {
		newRevisionMap[d] = 1
	}

	updatedDeliverablesNotHandled := make(map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID], e.PFD.AtomicProcesses.Len())
	for _, ap := range e.PFD.AtomicProcesses.Iter() {
		updatedDeliverablesNotHandled[ap] = sets.New(pfd.AtomicDeliverableID.Compare)
	}
	for _, d := range newlyAvailableInitialDeliverables.Iter() {
		for _, ap := range e.PFD.EitherFeedbackOrNotDestinationAtomicProcesses(d).Iter() {
			updatedDeliverablesNotHandled[ap].Add(pfd.AtomicDeliverableID.Compare, d)
		}
	}

	newRemainedVolumeMap := make(map[pfd.AtomicProcessID]Volume, e.PFD.AtomicProcesses.Len())
	for _, ap := range e.PFD.AtomicProcesses.Iter() {
		newRemainedVolumeMap[ap] = e.InitialVolumeFunc(ap)
	}

	completedAtomicProcesses := sets.NewWithCapacity[pfd.AtomicProcessID](e.PFD.AtomicProcesses.Len())
	e.CollectCompletedAtomicProcesses(emptyAllocation, newRemainedVolumeMap, completedAtomicProcesses)

	newNumOfReworksMap := e.UpdateNumberOfReworksMap(numOfReworksMap, completedAtomicProcesses)

	allocationShouldContinue := emptyAllocation

	return NewState(t, newRevisionMap, newRemainedVolumeMap, newNumOfReworksMap, allocationShouldContinue, updatedDeliverablesNotHandled)
}

func (e *Env) nextTime(state State, allocation Allocation) (execmodel.Time, error) {
	minCompletedTime, hasMinCompletedTime := e.MinimumCompletedTime(state.Time, state.RemainedVolumeMap, allocation)
	minNotGeneratedDeliverableAvailableTime, hasMinNotGeneratedDeliverableAvailableTime := MinimumNotGeneratedDeliverableAvailableTime(e.PFD.InitialDeliverables(), state.Time, e.DeliverableAvailableTimeFunc)
	if hasMinNotGeneratedDeliverableAvailableTime {
		if hasMinCompletedTime {
			return min(minCompletedTime, minNotGeneratedDeliverableAvailableTime), nil
		}
		return minNotGeneratedDeliverableAvailableTime, nil
	} else {
		if hasMinCompletedTime {
			return minCompletedTime, nil
		} else {
			sb := &strings.Builder{}
			sb.WriteString("fsm.Env.NextState: no minimum completed time and no minimum not generated deliverable available time. Deadlock!\n")
			_ = state.Write(sb)
			sb.WriteString("\n")
			_ = allocation.Write(sb)
			return 0, errors.New(sb.String())
		}
	}
}

func (e *Env) NextState(state State, allocation Allocation) (State, bool) {
	nextTime, err := e.nextTime(state, allocation)
	if err != nil {
		return State{}, false
	}

	timeDelta := nextTime - state.Time
	executedAtomicProcesses := sets.NewWithCapacity[pfd.AtomicProcessID](len(allocation))
	for ap := range allocation {
		executedAtomicProcesses.Add(pfd.AtomicProcessID.Compare, ap)
	}

	pastRemainedVolumeMap := state.RemainedVolumeMap
	pastRevisionMap := state.RevisionMap
	pastNumOfReworksMap := state.NumOfCompleteMap
	pastUpdatedDeliverablesNotHandled := state.UpdatedDeliverablesNotHandled

	newRevisionMap := maps.Clone(pastRevisionMap)
	newlyAvailableInitialDeliverables := e.NewlyAvailableInitialDeliverables(nextTime)
	for _, d := range newlyAvailableInitialDeliverables.Iter() {
		newRevisionMap[d] = 1
	}

	remainedVolumeMapNotRecovered := e.NewRemainedVolumeMap(pastRemainedVolumeMap, allocation, timeDelta)

	completedAtomicProcesses := sets.NewWithCapacity[pfd.AtomicProcessID](e.PFD.AtomicProcesses.Len())
	e.CollectCompletedAtomicProcesses(allocation, remainedVolumeMapNotRecovered, completedAtomicProcesses)

	updatedDeliverables := e.UpdatedDeliverables(completedAtomicProcesses)
	for _, d := range updatedDeliverables.Iter() {
		newRevisionMap[d]++
	}

	newUpdatedDeliverablesNotHandled := e.NewUpdateDeliverablesNotHandled(
		newlyAvailableInitialDeliverables,
		pastUpdatedDeliverablesNotHandled,
		newRevisionMap,
		executedAtomicProcesses,
		completedAtomicProcesses,
	)

	newNumOfReworksMap := e.UpdateNumberOfReworksMap(pastNumOfReworksMap, completedAtomicProcesses)

	allocationShouldContinue := make(Allocation, len(allocation))
	for ap, elem := range allocation {
		remained, ok := remainedVolumeMapNotRecovered[ap]
		if !ok {
			panic(fmt.Sprintf("fsm.NextState: missing remained volume: %q", ap))
		}

		if remained.IsZero() {
			continue
		}
		allocationShouldContinue[ap] = elem
	}

	recoveredVolumeMap := maps.Clone(remainedVolumeMapNotRecovered)
	for _, ap := range completedAtomicProcesses.Iter() {
		recoveredVolumeMap[ap] = e.ReworkVolumeFunc(ap, newNumOfReworksMap[ap])
	}

	return NewState(
		nextTime,
		newRevisionMap,
		recoveredVolumeMap,
		newNumOfReworksMap,
		allocationShouldContinue,
		newUpdatedDeliverablesNotHandled,
	), true
}

type Trans struct {
	Allocation Allocation `json:"allocation"`
	NextState  State      `json:"next_state"`
}

func CompareTrans(a, b *Trans) int {
	c := Allocation.Compare(a.Allocation, b.Allocation)
	if c != 0 {
		return c
	}
	return a.NextState.Compare(b.NextState)
}

func (e *Env) Transitions(state State) *sets.Set[*Trans] {
	return e.TransitionsWith(state, e.AvailableAllocationsFunc)
}

func (e *Env) TransitionsWith(state State, availableAllocationsFunc AvailableAllocationsFunc) *sets.Set[*Trans] {
	if e.IsCompleted(state) {
		return sets.NewWithCapacity[*Trans](0)
	}

	newlyAllocatables := e.NewlyAllocatables(state)
	allocations := availableAllocationsFunc(state, newlyAllocatables)
	if allocations.Len() == 0 {

		_, err := e.nextTime(state, state.AllocationShouldContinue)
		if err != nil {
			sb := &strings.Builder{}
			_ = state.Write(sb)
			m := e.AllocatabilityInfoMap(state)
			_ = m.Write(sb)
			ks := slices.Collect(maps.Keys(e.PreconditionMap))
			slices.SortFunc(ks, pfd.AtomicProcessID.Compare)
			for _, ap := range ks {
				fmt.Fprintf(sb, "precondition[%q]: ", ap)
				precondition := e.PreconditionMap[ap]
				precondition.Eval(e, state.RemainedVolumeMap, state.RevisionMap, state.AllocationShouldContinue, state.UpdatedDeliverablesNotHandled).Write(sb)
			}
			e.Logger.Warn("fsm.Env.Transitions: no progress state found", "state", sb.String())
			return sets.NewWithCapacity[*Trans](0)
		}
		allocations = sets.New(CompareAllocationByTotalConsumedVolume, state.AllocationShouldContinue)
	}

	transitions := sets.NewWithCapacity[*Trans](allocations.Len())
	for _, alloc := range allocations.Iter() {
		nextState, ok := e.NextState(state, alloc)
		if !ok {
			sb := &strings.Builder{}
			sb.WriteString("fsm.Env.Transitions: no next state found:\n")
			_ = state.Write(sb)
			panic(sb.String())
		}
		transitions.Add(CompareTrans, &Trans{Allocation: alloc, NextState: nextState})
	}
	return transitions
}

func (e *Env) IsCompleted(state State) bool {

	for _, d := range e.PFD.InitialDeliverables().Iter() {
		if state.Time < e.DeliverableAvailableTimeFunc(d) {

			return false
		}
	}
	for _, ap := range e.PFD.AtomicProcesses.Iter() {
		a := e.Allocatability(ap, state.RemainedVolumeMap, state.RevisionMap, state.AllocationShouldContinue, state.UpdatedDeliverablesNotHandled)
		switch a {
		case AllocatabilityOKContinuable, AllocatabilityOKStartable:

			return false
		case AllocatabilityNGNoDeliverableUpdates:

		case AllocatabilityNGInsufficientInputs, AllocatabilityNGPreconditionNotMet:

			return false
		default:
			panic(fmt.Sprintf("fsm.Env.IsCompleted: unknown allocatability: %s", a))
		}
	}
	return true
}
