package fsm

import (
	"cmp"
	"fmt"
	"time"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
)

type CriticalPathInfo map[pfd.AtomicProcessID]*CriticalPathInfoItem

type CriticalPathInfoItem struct {
	MaximumElasticity execmodel.Time `json:"maximum_elasticity,omitempty"`

	HasMinimumElasticity bool `json:"has_minimum_elasticity"`

	MinimumElasticity execmodel.Time `json:"minimum_elasticity,omitempty"`
}

type CriticalPathInfoFunc func(e *Env) (CriticalPathInfo, error)

func NewCriticalPathInfoFunc(searchFunc SearchWithPrefixFunc) CriticalPathInfoFunc {
	return func(e *Env) (CriticalPathInfo, error) {
		info := make(CriticalPathInfo, e.PFD.AtomicProcesses.Len())
		numOfAP := e.PFD.AtomicProcesses.Len()

		for i, ap := range e.PFD.AtomicProcesses.Iter() {
			e.Logger.Info("processing atomic process", "atomic_process", ap, "progress", fmt.Sprintf("%d/%d", i+1, numOfAP))
			procStart := time.Now()
			item, err := newCriticalPathInfoItem(ap, e, searchFunc)
			if err != nil {
				return nil, fmt.Errorf("fsm.NewCriticalPathInfoFunc: %w", err)
			}
			info[ap] = item
			e.Logger.Debug("fsm.CriticalPath: process done", "atomicProcess", ap, "progress", fmt.Sprintf("%d/%d", i+1, numOfAP), "elapsedMs", time.Since(procStart).Milliseconds())
		}
		return info, nil
	}
}

func newCriticalPathInfoItem(ap pfd.AtomicProcessID, e *Env, searchFunc SearchWithPrefixFunc) (*CriticalPathInfoItem, error) {
	baseStart := time.Now()
	basePlans, err := searchFunc(e, nil)
	if err != nil {
		return nil, fmt.Errorf("fsm.NewCriticalPathInfoFunc: %w", err)
	}
	basePlan, ok := basePlans.At(0)
	if !ok {
		return nil, fmt.Errorf("fsm.NewCriticalPathInfoFunc: no base plan found")
	}
	leadtime := basePlan.Leadtime()
	e.Logger.Debug("fsm.CriticalPath: base plan", "atomicProcess", ap, "leadtime", float64(leadtime), "elapsedMs", time.Since(baseStart).Milliseconds())

	maxStart := time.Now()
	maxElasticity, err := newMaximumElasticity(ap, leadtime, basePlan, e, searchFunc)
	if err != nil {
		return nil, fmt.Errorf("fsm.NewCriticalPathInfoFunc: %w", err)
	}
	e.Logger.Debug("fsm.CriticalPath: max elasticity", "atomicProcess", ap, "value", float64(maxElasticity), "elapsedMs", time.Since(maxStart).Milliseconds())

	minStart := time.Now()
	minElasticity, hasMinimumElasticity, err := newMinimumElasticity(ap, leadtime, e, searchFunc)
	if err != nil {
		return nil, fmt.Errorf("fsm.NewCriticalPathInfoFunc: %w", err)
	}
	e.Logger.Debug("fsm.CriticalPath: min elasticity", "atomicProcess", ap, "value", float64(minElasticity), "hasMinimumElasticity", hasMinimumElasticity, "elapsedMs", time.Since(minStart).Milliseconds())

	return &CriticalPathInfoItem{
		MaximumElasticity:    maxElasticity,
		HasMinimumElasticity: hasMinimumElasticity,
		MinimumElasticity:    minElasticity,
	}, nil
}

func newMaximumElasticity(ap pfd.AtomicProcessID, leadtime execmodel.Time, basePlan *Plan, e *Env, searchFunc SearchWithPrefixFunc) (execmodel.Time, error) {
	prefix := PrefixUntilProcessStart(basePlan, ap)
	baseProcessTime := basePlan.ProcessTime(ap)

	lo := Volume(0)
	elasticity := execmodel.Time(0)
	hi := Volume(leadtime) * MaxConsumedVolume(e.NeededResourceSetsFunc(ap))
	elasticityHi := execmodel.Time(0)
	hasHi := false
	linear := true

	extraVolume := hi
	for i := 0; i < 100; i++ {
		iterStart := time.Now()
		extra := extraVolume
		eLonger := e.Clone()
		eLonger.InitialVolumeFunc = func(ap2 pfd.AtomicProcessID) Volume {
			if ap2 == ap {
				return e.InitialVolumeFunc(ap2) + extra
			}
			return e.InitialVolumeFunc(ap2)
		}

		longerPlans, err := searchFunc(eLonger, prefix)
		if err != nil {
			return 0, fmt.Errorf("fsm.newMaximumElasticity: %w", err)
		}
		longerPlan, ok := longerPlans.At(0)
		if !ok {
			return 0, fmt.Errorf("fsm.newMaximumElasticity: no longer plan found")
		}

		extended := longerPlan.ProcessTime(ap) - baseProcessTime
		extension := longerPlan.Leadtime() - leadtime
		e.Logger.Debug("fsm.newMaximumElasticity: refine", "atomicProcess", ap, "iter", i, "extraVolume", float64(extraVolume), "extended", float64(extended), "extension", float64(extension), "lo", float64(lo), "hi", float64(hi), "hasHi", hasHi, "linear", linear, "elapsedMs", time.Since(iterStart).Milliseconds())

		if extension < 0 || extension.ApproximateEqual(0) {

			lo = max(lo, extraVolume)
			elasticity = max(elasticity, extended)

			if !hasHi {
				return elasticity, nil
			}

			if linear {
				return elasticity, nil
			}
		} else {
			hi, elasticityHi, hasHi = extraVolume, extended, true

			if lo == 0 && extension.ApproximateEqual(extended) {
				return 0, nil
			}
		}

		if elasticity < elasticityHi && elasticityHi-elasticity <= execmodel.MinimumTime {
			return elasticity, nil
		}
		if hi-lo <= MinimumVolume {
			return elasticity, nil
		}

		if linear && extension > 0 && extended > 0 {
			next := extraVolume * Volume(1-extension/extended)
			if lo < next && next < hi {
				extraVolume = next
				continue
			}
		}
		linear = false
		extraVolume = lo + (hi-lo)/2
	}

	e.Logger.Warn("fsm.newMaximumElasticity: not converged", "atomicProcess", ap, "elasticity", float64(elasticity), "lo", float64(lo), "hi", float64(hi))
	return elasticity, nil
}

func newMinimumElasticity(ap pfd.AtomicProcessID, leadtime execmodel.Time, e *Env, searchFunc SearchWithPrefixFunc) (execmodel.Time, bool, error) {
	eShorter := e.Clone()
	eShorter.InitialVolumeFunc = func(ap2 pfd.AtomicProcessID) Volume {
		if ap2 == ap {
			return MinimumVolume
		}
		return e.InitialVolumeFunc(ap2)
	}

	shorterPlans, err := searchFunc(eShorter, nil)
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
