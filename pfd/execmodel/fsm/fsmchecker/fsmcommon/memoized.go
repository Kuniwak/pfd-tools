package fsmcommon

import (
	"fmt"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmmasterschedule"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/sets"
)

type NeededResourceSetEntry struct {
	Resources      sets.Set[fsm.ResourceID]
	ConsumedVolume string
}

type Memoized struct {
	InitialVolumeMap    map[pfd.AtomicProcessID]string
	HasInitialVolumeMap bool

	MaxRevisionMap    map[pfd.AtomicDeliverableID]string
	HasMaxRevisionMap bool

	NeededResourceSetsMap    map[pfd.AtomicProcessID]string
	HasNeededResourceSetsMap bool

	AllResources    *sets.Set[fsm.ResourceID]
	HasAllResources bool

	AvailableTimeMap    map[pfd.AtomicDeliverableID]string
	HasAvailableTimeMap bool

	PreconditionMap    map[pfd.AtomicProcessID]string
	HasPreconditionMap bool

	GroupMap    map[pfd.AtomicProcessID]string
	HasGroupMap bool

	MilestoneMap    map[pfd.AtomicProcessID]string
	HasMilestoneMap bool

	MilestoneEdgesMap    map[fsmmasterschedule.Milestone]string
	HasMilestoneEdgesMap bool
}

func NewMemoized(
	apTable *pfd.AtomicProcessTable,
	adTable *pfd.AtomicDeliverableTable,
	rTable *fsmtable.ResourceTable,
	mt *fsmtable.MilestoneTable,
) (*Memoized, error) {
	var err error
	var hasInitialVolumeMap bool
	var hasMaxRevisionMap bool
	var hasNeededResourceSetsMap bool
	var hasAllResources bool
	var hasAvailableTimeMap bool
	var hasPreconditionMap bool
	var hasGroupMap bool
	var hasMilestoneMap bool
	var hasMilestoneEdgesMap bool

	var initialVolumeMap map[pfd.AtomicProcessID]string
	var maxRevisionMap map[pfd.AtomicDeliverableID]string
	var neededResourceSetsMap map[pfd.AtomicProcessID]string
	var preconditionMap map[pfd.AtomicProcessID]string
	var groupMap map[pfd.AtomicProcessID]string
	var milestoneMap map[pfd.AtomicProcessID]string
	var milestoneEdgesMap map[fsmmasterschedule.Milestone]string
	if apTable != nil {
		if fsmtable.DefaultInitialVolumeColumnMatchFunc(apTable.ExtraHeaders) >= 0 {
			initialVolumeMap, err = fsmtable.RawInitialVolumeMap(apTable, fsmtable.DefaultInitialVolumeColumnMatchFunc)
			if err != nil {
				return nil, fmt.Errorf("fsmcommon.NewMemoized: %w", err)
			}
			hasInitialVolumeMap = true

		}

		if fsmtable.DefaultNeededResourceSetsColumnSelectFunc(apTable.ExtraHeaders) >= 0 {
			neededResourceSetsMap, err = fsmtable.RawNeededResourceSetsMap(apTable, fsmtable.DefaultNeededResourceSetsColumnSelectFunc)
			if err != nil {
				return nil, fmt.Errorf("fsmcommon.NewMemoized: %w", err)
			}
			hasNeededResourceSetsMap = true
		}

		if fsmtable.DefaultPreconditionColumnMatchFunc(apTable.ExtraHeaders) >= 0 {
			preconditionMap, err = fsmtable.RawPreconditionMap(apTable, fsmtable.DefaultPreconditionColumnMatchFunc)
			if err != nil {
				return nil, fmt.Errorf("fsmcommon.NewMemoized: %w", err)
			}
			hasPreconditionMap = true
		}

		if fsmtable.DefaultGroupColumnMatchFunc(apTable.ExtraHeaders) >= 0 {
			groupMap, err = fsmtable.RawGroupsMap(apTable, fsmtable.DefaultGroupColumnMatchFunc)
			if err != nil {
				return nil, fmt.Errorf("fsmcommon.NewMemoized: %w", err)
			}
			hasGroupMap = true
		}

		if fsmtable.DefaultMilestoneColumnMatchFunc(apTable.ExtraHeaders) >= 0 {
			milestoneMap, err = fsmtable.RawMilestoneMap(apTable, fsmtable.DefaultMilestoneColumnMatchFunc)
			if err != nil {
				return nil, fmt.Errorf("fsmcommon.NewMemoized: %w", err)
			}
			hasMilestoneMap = true
		}
	}

	var resources *sets.Set[fsm.ResourceID]
	if rTable != nil {
		resources = fsmtable.AvailableResources(rTable)
		hasAllResources = true
	}

	var availableTimeMap map[pfd.AtomicDeliverableID]string
	if adTable != nil {
		if fsmtable.DefaultAvailableTimeColumnMatchFunc(adTable.ExtraHeaders) >= 0 {
			availableTimeMap, err = fsmtable.RawAvailableTimeMap(adTable, fsmtable.DefaultAvailableTimeColumnMatchFunc)
			if err != nil {
				return nil, fmt.Errorf("fsmcommon.NewMemoized: %w", err)
			}
			hasAvailableTimeMap = true
		}

		if fsmtable.DefaultMaxRevisionColumnMatchFunc(adTable.ExtraHeaders) >= 0 {
			maxRevisionMap, err = fsmtable.RawMaxRevisionMap(adTable, fsmtable.DefaultMaxRevisionColumnMatchFunc)
			if err != nil {
				return nil, fmt.Errorf("fsmcommon.NewMemoized: %w", err)
			}
			hasMaxRevisionMap = true
		}
	}

	if mt != nil {
		milestoneEdgesMap, err = fsmtable.RawMilestoneEdgesMap(mt)
		if err != nil {
			return nil, fmt.Errorf("fsmcommon.NewMemoized: %w", err)
		}
		hasMilestoneEdgesMap = true
	}

	return &Memoized{
		InitialVolumeMap:    initialVolumeMap,
		HasInitialVolumeMap: hasInitialVolumeMap,

		MaxRevisionMap:    maxRevisionMap,
		HasMaxRevisionMap: hasMaxRevisionMap,

		NeededResourceSetsMap:    neededResourceSetsMap,
		HasNeededResourceSetsMap: hasNeededResourceSetsMap,

		AllResources:    resources,
		HasAllResources: hasAllResources,

		AvailableTimeMap:    availableTimeMap,
		HasAvailableTimeMap: hasAvailableTimeMap,

		PreconditionMap:    preconditionMap,
		HasPreconditionMap: hasPreconditionMap,

		GroupMap:    groupMap,
		HasGroupMap: hasGroupMap,

		MilestoneMap:    milestoneMap,
		HasMilestoneMap: hasMilestoneMap,

		MilestoneEdgesMap:    milestoneEdgesMap,
		HasMilestoneEdgesMap: hasMilestoneEdgesMap,
	}, nil
}
