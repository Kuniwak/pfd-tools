package fsm

import (
	"fmt"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/sets"
)

// DeliverableAvailableTimeFunc returns the available time of initial deliverables. Behavior is undefined when given an ID of an element that is not an initial deliverable.
type DeliverableAvailableTimeFunc func(pfd.AtomicDeliverableID) execmodel.Time

func ConstDeliverableAvailableTimeFunc(t execmodel.Time) DeliverableAvailableTimeFunc {
	return func(pfd.AtomicDeliverableID) execmodel.Time {
		return t
	}
}

func AlwaysAvailableTimeFunc() DeliverableAvailableTimeFunc {
	return ConstDeliverableAvailableTimeFunc(0)
}

func AvailableTimeFuncByMap(m map[pfd.AtomicDeliverableID]execmodel.Time) DeliverableAvailableTimeFunc {
	return func(d pfd.AtomicDeliverableID) execmodel.Time {
		t, ok := m[d]
		if !ok {
			panic(fmt.Sprintf("fsm.AvailableTimeFuncByMap: missing %q in map", d))
		}
		return t
	}
}

func ConstMaxRevisionMap(maxRevision int, fbSourceDeliverables *sets.Set[pfd.AtomicDeliverableID]) map[pfd.AtomicDeliverableID]int {
	m := make(map[pfd.AtomicDeliverableID]int, fbSourceDeliverables.Len())
	for _, d := range fbSourceDeliverables.Iter() {
		m[d] = maxRevision
	}
	return m
}
