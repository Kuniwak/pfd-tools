package pfdcommon

import (
	"github.com/Kuniwak/pfd-tools/pfd"
)

type Target struct {
	PFD                       *pfd.PFD
	AtomicProcessTable        *pfd.AtomicProcessTable
	AtomicDeliverableTable    *pfd.AtomicDeliverableTable
	CompositeProcessTable     *pfd.CompositeProcessTable
	CompositeDeliverableTable *pfd.CompositeDeliverableTable
	Memoized                  Memoized
}

func NewTarget(
	pfd *pfd.PFD,
	atomicProcessTable *pfd.AtomicProcessTable,
	atomicDeliverableTable *pfd.AtomicDeliverableTable,
	compositeProcessTable *pfd.CompositeProcessTable,
	compositeDeliverableTable *pfd.CompositeDeliverableTable,
	memoized Memoized,
) Target {
	return Target{
		PFD:                       pfd,
		AtomicProcessTable:        atomicProcessTable,
		AtomicDeliverableTable:    atomicDeliverableTable,
		CompositeProcessTable:     compositeProcessTable,
		CompositeDeliverableTable: compositeDeliverableTable,
		Memoized:                  memoized,
	}
}
