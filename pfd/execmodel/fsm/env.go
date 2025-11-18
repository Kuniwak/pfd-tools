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

// Env is the environment for the finite resource single deliverable execution model.
type Env struct {
	// PFD is the PFD.
	PFD *pfd.ValidPFD

	// AvailableResources is the set of available resources.
	AvailableResources *sets.Set[ResourceID]

	// AvailableAllocationsFunc is a function that enumerates and returns possible resource allocations in the given state.
	AvailableAllocationsFunc AvailableAllocationsFunc

	// InitialVolumeFunc is a function that provides the initial work volume for each atomic process.
	InitialVolumeFunc InitialVolumeFunc

	// ReworkVolumeFunc, when given the number of rework iterations for each atomic process, returns the work volume
	// that is recovered when feedback edge deliverables are created or recreated.
	ReworkVolumeFunc ReworkVolumeFunc

	// FeedbackSourceMaxRevision returns the maximum revision for each feedback source deliverable.
	FeedbackSourceMaxRevision map[pfd.AtomicDeliverableID]int

	// PreconditionMap returns whether execution conditions are satisfied for each atomic process.
	PreconditionMap map[pfd.AtomicProcessID]*Precondition

	// NeededResourceSetsFunc is a function that provides the set of required resources for each atomic process.
	NeededResourceSetsFunc NeededResourceSetsFunc

	// DeliverableAvailableTimeFunc is a function that provides the available time for each deliverable.
	DeliverableAvailableTimeFunc DeliverableAvailableTimeFunc

	Memoized *Memoized

	// Logger is the logger.
	Logger *slog.Logger
}

// NewEnv returns a new environment.
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

// FreeResources returns the free resources in the given state.
// This is the available resources at the current time minus the allocated resources of atomic processes that are continuing execution.
func (e *Env) FreeResources(state State) *sets.Set[ResourceID] {
	avail := e.AvailableResources.Clone()
	allocations := state.AllocationShouldContinue
	for _, alloc := range allocations {
		avail.Difference(ResourceID.Compare, alloc.Resources)
	}
	return avail
}

// AllocatabilityInfo returns whether resources can be allocated to an atomic process if resources can be occupied.
// An atomic process is allocatable if it satisfies any of the following conditions:
//
// - Continuing execution
// - All input deliverables of the atomic process have been generated, at least one input deliverable has been updated but not processed, remaining work volume is not 0, and start conditions are satisfied
func (e *Env) AllocatabilityInfo(
	ap pfd.AtomicProcessID,
	remainedVolumeMap map[pfd.AtomicProcessID]Volume,
	revisionMap map[pfd.AtomicDeliverableID]int,
	allocationShouldContinue Allocation,
	updatedDeliverablesNotHandled map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID],
) *AllocatabilityInfo {
	if _, ok := allocationShouldContinue[ap]; ok {
		// NOTE: Atomic processes continuing execution are executable.
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
		// NOTE: Atomic processes with ungenerated input deliverables are not executable.
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

// NewlyAllocatables returns the set of atomic processes that are newly allocatable in the given state.
func (e *Env) NewlyAllocatables(state State) *sets.Set[pfd.AtomicProcessID] {
	res := sets.NewWithCapacity[pfd.AtomicProcessID](e.PFD.AtomicProcesses.Len())
	for _, ap := range e.PFD.AtomicProcesses.Iter() {
		a := e.Allocatability(ap, state.RemainedVolumeMap, state.RevisionMap, state.AllocationShouldContinue, state.UpdatedDeliverablesNotHandled)
		switch a {
		case AllocatabilityOKStartable:
			res.Add(pfd.AtomicProcessID.Compare, ap)
		case AllocatabilityOKContinuable, AllocatabilityNGInsufficientInputs, AllocatabilityNGPreconditionNotMet, AllocatabilityNGNoDeliverableUpdates:
			// Do nothing.
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
		if remainedVolume.IsZero() {
			panic(fmt.Sprintf("fsm.Env.MinimumCompletedTime: remained volume is zero: %q", ap))
		}

		restTime := execmodel.Time(float64(remainedVolume) / float64(alloc.ConsumedVolume))
		if restTime < minTime {
			minTime = restTime
		}
	}
	return currentTime + minTime, minTime != execmodel.Time(math.MaxFloat64)
}

// NewRevisionMap returns a new RevisionMap.
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

// NewlyAvailableInitialDeliverables returns the set of initial deliverables that became newly available at the current time.
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

// NewRemainedVolumeMap returns a new dictionary of remaining work time after reducing the remaining work time of atomic processes by the given allocation.
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

// UpdateNumberOfReworksMap increments the execution completion count of completed atomic processes by one.
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

// CollectCompletedAtomicProcesses finds completed atomic processes.
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

// UpdatedDeliverables returns the set of deliverables that were updated from the set of atomic processes completed in this transition.
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
			// NOTE: Input deliverables have been processed.
			newUpdatedDeliverablesNotHandled[ap] = sets.New(pfd.AtomicDeliverableID.Compare)
		} else {
			// NOTE: Input deliverables remain unprocessed.
			newUpdatedDeliverablesNotHandled[ap] = pastUpdatedDeliverablesNotHandled[ap].Clone()
		}
	}

	for _, d := range newlyAvailableInitialDeliverables.Iter() {
		for _, ap := range e.PFD.EitherFeedbackOrNotDestinationAtomicProcesses(d).Iter() {
			newUpdatedDeliverablesNotHandled[ap].Add(pfd.AtomicDeliverableID.Compare, d)
		}
	}

	for _, ap := range completedAtomicProcesses.Iter() {
		// NOTE: Among output deliverables, only mark as unprocessed those that are not feedback deliverables or feedback deliverables that have not reached the maximum revision.
		for _, d := range e.PFD.OutputDeliverables(ap).Iter() {
			curRevision, ok := newRevisionMap[d]
			if !ok {
				panic(fmt.Sprintf("fsm.Env.NewUpdateDeliverablesNotHandled: missing revision: %q", d))
			}

			maxRevision, ok := e.FeedbackSourceMaxRevision[d]
			if ok && curRevision >= maxRevision {
				// NOTE: Feedback deliverables that have reached the maximum revision are not marked as unprocessed.
				continue
			}

			for _, ap2 := range e.PFD.EitherFeedbackOrNotDestinationAtomicProcesses(d).Iter() {
				newUpdatedDeliverablesNotHandled[ap2].Add(pfd.AtomicDeliverableID.Compare, d)
			}
		}
	}
	return newUpdatedDeliverablesNotHandled
}

// InitialState returns the initial state.
func (e *Env) InitialState() State {
	t := execmodel.Time(0)

	numOfReworksMap := make(map[pfd.AtomicProcessID]int, e.PFD.AtomicProcesses.Len())
	for _, ap := range e.PFD.AtomicProcesses.Iter() {
		numOfReworksMap[ap] = 0
	}

	// NOTE: There is no resource allocation in the initial state.
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

// NextState returns the next state from the given state and allocation.
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

// Trans is a transition in the FSM.
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

// Transitions returns the set of transitions from the given state.
func (e *Env) Transitions(state State) *sets.Set[*Trans] {
	if e.IsCompleted(state) {
		return sets.NewWithCapacity[*Trans](0)
	}

	newlyAllocatables := e.NewlyAllocatables(state)
	allocations := e.AvailableAllocationsFunc(state, newlyAllocatables)
	if allocations.Len() == 0 {
		// NOTE: If not in a completed state but no allocations exist, we need to wait for the completion of continuing processes or until the available time of initial deliverables.
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

// IsCompleted returns whether the given state is completed.
func (e *Env) IsCompleted(state State) bool {
	// TODO(PROOF_NEEDED): From the initial state, any transition will eventually reach a completed state (otherwise the execution plan search will not stop).
	for _, d := range e.PFD.InitialDeliverables().Iter() {
		if state.Time < e.DeliverableAvailableTimeFunc(d) {
			// NOTE: Not completed if the current time is not greater than or equal to the maximum available time of the deliverable.
			return false
		}
	}
	for _, ap := range e.PFD.AtomicProcesses.Iter() {
		a := e.Allocatability(ap, state.RemainedVolumeMap, state.RevisionMap, state.AllocationShouldContinue, state.UpdatedDeliverablesNotHandled)
		switch a {
		case AllocatabilityOKContinuable, AllocatabilityOKStartable:
			// NOTE: Not completed because there are allocatable atomic processes.
			return false
		case AllocatabilityNGNoDeliverableUpdates:
			// Do nothing.
		case AllocatabilityNGInsufficientInputs, AllocatabilityNGPreconditionNotMet:
			// NOTE: Not considered as execution completed state because there are processes that have not been executed or have not completed execution.
			return false
		default:
			panic(fmt.Sprintf("fsm.Env.IsCompleted: unknown allocatability: %s", a))
		}
	}
	return true
}
