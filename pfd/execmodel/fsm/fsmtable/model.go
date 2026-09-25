package fsmtable

import (
	"fmt"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
)

func NewResourceAspect(
	mode execmodel.ResourceMode,
	apTable *pfd.AtomicProcessTable,
	rTable *ResourceTable,
) (fsm.ResourceAspect, error) {
	switch mode {
	case execmodel.ResourceModeFinite:
		if rTable == nil {
			return fsm.ResourceAspect{}, fmt.Errorf("fsmtable.NewResourceAspect: resource table is required for the %s resource mode", mode)
		}
		neededResourceSetsFunc, err := NeededResourcesSetFuncByTable(apTable, DefaultNeededResourceSetsColumnSelectFunc)
		if err != nil {
			return fsm.ResourceAspect{}, fmt.Errorf("fsmtable.NewResourceAspect: needed resource sets func: %w", err)
		}
		return fsm.ResourceAspect{
			AvailableResources:     AvailableResources(rTable),
			NeededResourceSetsFunc: neededResourceSetsFunc,
		}, nil
	case execmodel.ResourceModeInfinite:
		return fsm.UnlimitedResourceAspect(), nil
	default:
		return fsm.ResourceAspect{}, fmt.Errorf("fsmtable.NewResourceAspect: unknown resource mode: %q", mode)
	}
}

func NewFeedbackAspect(
	mode execmodel.FeedbackMode,
	p *pfd.ValidPFD,
	apTable *pfd.AtomicProcessTable,
	adTable *pfd.AtomicDeliverableTable,
	initialVolumeFunc fsm.InitialVolumeFunc,
) (fsm.FeedbackAspect, error) {
	switch mode {
	case execmodel.FeedbackModeEnabled:
		reworkVolumeFunc, err := ReworkVolumeFuncByTableFunc(apTable, DefaultReworkVolumeRatioColumnMatchFunc, initialVolumeFunc)
		if err != nil {
			return fsm.FeedbackAspect{}, fmt.Errorf("fsmtable.NewFeedbackAspect: rework volume func: %w", err)
		}
		maxRevisionMap, err := MaxRevisionMapByTableFunc(adTable, DefaultMaxRevisionColumnMatchFunc, p.FeedbackSourceDeliverables())
		if err != nil {
			return fsm.FeedbackAspect{}, fmt.Errorf("fsmtable.NewFeedbackAspect: max revision map: %w", err)
		}
		return fsm.FeedbackAspect{
			ReworkVolumeFunc:          reworkVolumeFunc,
			FeedbackSourceMaxRevision: maxRevisionMap,
		}, nil
	case execmodel.FeedbackModeDisabled:
		if err := fsm.ValidateNoFeedbackEdges(p); err != nil {
			return fsm.FeedbackAspect{}, fmt.Errorf("fsmtable.NewFeedbackAspect: %w", err)
		}
		return fsm.NoFeedbackAspect(initialVolumeFunc), nil
	default:
		return fsm.FeedbackAspect{}, fmt.Errorf("fsmtable.NewFeedbackAspect: unknown feedback mode: %q", mode)
	}
}
