package fsm

import (
	"cmp"
	"fmt"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/sets"
)

type CriticalPathInfo map[pfd.AtomicProcessID]*CriticalPathInfoItem

// There may be times when there is a maximum value m such that increasing the required time of the target atomic process
// does not increase the overall required time. This m is called the maximum elasticity value.
// The maximum elasticity value corresponds to the Total float in CPM.
// There may be times when there is a minimum value n such that decreasing the required time of the target atomic process
// does not decrease the overall required time. This n is called the minimum elasticity value.
type CriticalPathInfoItem struct {
	// Maximum elasticity value.
	MaximumElasticity execmodel.Time `json:"maximum_elasticity,omitempty"`

	// HasMinimumElasticity indicates whether a minimum elasticity value exists.
	HasMinimumElasticity bool `json:"has_minimum_elasticity"`

	// Minimum elasticity value.
	MinimumElasticity execmodel.Time `json:"minimum_elasticity,omitempty"`
}

type CriticalPathInfoFunc func(e *Env) (CriticalPathInfo, error)

func NewCriticalPathInfoFunc(searchFunc SearchFunc) CriticalPathInfoFunc {
	return func(e *Env) (CriticalPathInfo, error) {
		info := make(CriticalPathInfo, e.PFD.AtomicProcesses.Len())
		numOfAP := e.PFD.AtomicProcesses.Len()

		for i, ap := range e.PFD.AtomicProcesses.Iter() {
			e.Logger.Info("processing atomic process", "atomic_process", ap, "progress", fmt.Sprintf("%d/%d", i+1, numOfAP))
			item, err := newCriticalPathInfoItem(ap, e, searchFunc)
			if err != nil {
				return nil, fmt.Errorf("fsm.NewCriticalPathInfoFunc: %w", err)
			}
			info[ap] = item
		}
		return info, nil
	}
}

func newCriticalPathInfoItem(ap pfd.AtomicProcessID, e *Env, searchFunc SearchFunc) (*CriticalPathInfoItem, error) {
	basePlans, err := searchFunc(e)
	if err != nil {
		return nil, fmt.Errorf("fsm.NewCriticalPathInfoFunc: %w", err)
	}
	basePlan, ok := basePlans.At(0)
	if !ok {
		return nil, fmt.Errorf("fsm.NewCriticalPathInfoFunc: no base plan found")
	}
	leadtime := basePlan.Leadtime()

	maxElasticity, err := newMaximumElasticity(ap, leadtime, e, searchFunc)
	if err != nil {
		return nil, fmt.Errorf("fsm.NewCriticalPathInfoFunc: %w", err)
	}

	minElasticity, hasMinimumElasticity, err := newMinimumElasticity(ap, leadtime, e, searchFunc)
	if err != nil {
		return nil, fmt.Errorf("fsm.NewCriticalPathInfoFunc: %w", err)
	}

	return &CriticalPathInfoItem{
		MaximumElasticity:    maxElasticity,
		HasMinimumElasticity: hasMinimumElasticity,
		MinimumElasticity:    minElasticity,
	}, nil
}

// There may be times when there is a maximum value m such that increasing the required time of the target atomic process
// does not increase the overall required time. This m is called the maximum elasticity value.
// The maximum elasticity value corresponds to the Total float in CPM.
func newMaximumElasticity(ap pfd.AtomicProcessID, leadtime execmodel.Time, e *Env, searchFunc SearchFunc) (execmodel.Time, error) {
	extra := leadtime

	eLonger := e.Clone()
	eLonger.InitialVolumeFunc = func(ap2 pfd.AtomicProcessID) Volume {
		if ap2 == ap {
			return e.InitialVolumeFunc(ap2) + Volume(extra)
		}
		return e.InitialVolumeFunc(ap2)
	}
	eLonger.NeededResourceSetsFunc = func(ap2 pfd.AtomicProcessID) *sets.Set[AllocationElement] {
		if ap2 == ap {
			rs1 := e.NeededResourceSetsFunc(ap2)
			rs2 := sets.NewWithCapacity[AllocationElement](rs1.Len())
			for _, rs := range rs1.Iter() {
				// FIXME: Fixed to 1 to extend by extra amount, but if ConsumedVolume is greater than 1, the optimal allocation might change.
				if rs.ConsumedVolume != 1 {
					panic(fmt.Sprintf("fsm.newMaximumElasticity: ConsumedVolume is not 1 is not supported: %v", rs))
				}
				rs2.Add(AllocationElement.Compare, AllocationElement{Resources: rs.Resources, ConsumedVolume: 1})
			}
			return rs2
		}
		return e.NeededResourceSetsFunc(ap2)
	}

	longerPlans, err := searchFunc(eLonger)
	if err != nil {
		return 0, fmt.Errorf("fsm.newMaximumElasticity: %w", err)
	}
	longerPlan, ok := longerPlans.At(0)
	if !ok {
		return 0, fmt.Errorf("fsm.newMaximumElasticity: no longer plan found")
	}
	longerLeadtime := longerPlan.Leadtime()

	maximumElasticity := extra - (longerLeadtime - leadtime)
	return maximumElasticity, nil
}

// There may be times when there is a minimum value n such that decreasing the required time of the target atomic process
// does not decrease the overall required time. This n is called the minimum elasticity value.
// There is no term in CPM that corresponds to the minimum elasticity value.
func newMinimumElasticity(ap pfd.AtomicProcessID, leadtime execmodel.Time, e *Env, searchFunc SearchFunc) (execmodel.Time, bool, error) {
	eShorter := e.Clone()
	eShorter.InitialVolumeFunc = func(ap2 pfd.AtomicProcessID) Volume {
		if ap2 == ap {
			return MinimumVolume
		}
		return e.InitialVolumeFunc(ap2)
	}

	shorterPlans, err := searchFunc(eShorter)
	if err != nil {
		return 0, false, fmt.Errorf("fsm.NewCriticalPathInfoFunc: %w", err)
	}
	shorterPlan, ok := shorterPlans.At(0)
	if !ok {
		return 0, false, fmt.Errorf("fsm.NewCriticalPathInfoFunc: no shorter plan found")
	}
	shorterLeadtime := shorterPlan.Leadtime()
	return leadtime - shorterLeadtime, shorterLeadtime != leadtime, nil
}

type Elasticity int

func (e Elasticity) Days() (int, bool) {
	i := int(e)
	if i == 0 {
		return 0, false
	}
	return i, true
}

func (e Elasticity) Compare(e2 Elasticity) int {
	d1, ok1 := e.Days()
	d2, ok2 := e2.Days()
	if !ok1 {
		if ok2 {
			return 1
		}
		return 0
	}
	if !ok2 {
		return -1
	}
	return cmp.Compare(d1, d2)
}
