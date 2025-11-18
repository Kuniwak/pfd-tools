package fsmcommon

import (
	"log/slog"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
)

type Target struct {
	PFD                    *pfd.ValidPFD
	AtomicProcessTable     *pfd.AtomicProcessTable
	AtomicDeliverableTable *pfd.AtomicDeliverableTable
	ResourceTable          *fsmtable.ResourceTable
	MilestoneTable         *fsmtable.MilestoneTable
	GroupTable             *fsmtable.GroupTable
	Memoized               *Memoized
	Logger                 *slog.Logger
}

func NewTarget(
	pfd *pfd.ValidPFD,
	atomicProcessTable *pfd.AtomicProcessTable,
	atomicDeliverableTable *pfd.AtomicDeliverableTable,
	resourceTable *fsmtable.ResourceTable,
	milestoneTable *fsmtable.MilestoneTable,
	groupTable *fsmtable.GroupTable,
	memoized *Memoized,
	logger *slog.Logger,
) *Target {
	return &Target{
		PFD:                    pfd,
		AtomicProcessTable:     atomicProcessTable,
		AtomicDeliverableTable: atomicDeliverableTable,
		ResourceTable:          resourceTable,
		MilestoneTable:         milestoneTable,
		GroupTable:             groupTable,
		Memoized:               memoized,
		Logger:                 logger,
	}
}
