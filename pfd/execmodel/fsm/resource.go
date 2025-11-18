package fsm

import (
	"fmt"
	"hash/maphash"
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/sets"
)

// ResourceID is the ID of a resource.
type ResourceID string

func (a ResourceID) Compare(b ResourceID) int {
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return strings.Compare(string(a), string(b))
}

func HashResourceID(a ResourceID, h *maphash.Hash) error {
	if _, err := h.WriteString(string(a)); err != nil {
		return fmt.Errorf("fsm.HashResourceID: %w", err)
	}
	return nil
}

// NewResourcesByMap returns a set of elements that distinguish individual resources from resource quantities.
func NewResourcesByMap(m map[ResourceID]int) *sets.Set[ResourceID] {
	s := sets.NewWithCapacity[ResourceID](len(m))
	for id, count := range m {
		for i := 1; i <= count; i++ {
			s.Add(ResourceID.Compare, ResourceID(fmt.Sprintf("%s%d", id, i)))
		}
	}
	return s
}

// AvailableResourcesFunc returns the set of available resources at a given time. Behavior is undefined for negative time values.
type AvailableResourcesFunc func(execmodel.Time) sets.Set[ResourceID]

func ConstAvailableResourcesFunc(s sets.Set[ResourceID]) AvailableResourcesFunc {
	return func(execmodel.Time) sets.Set[ResourceID] {
		return s
	}
}

// NeededResourceSetsFunc returns the required resources and the consumed work volume per unit time when those resources are allocated, given an atomic process.
// Behavior is undefined when given an ID of an element that is not an atomic process.
type NeededResourceSetsFunc func(ap pfd.AtomicProcessID) *sets.Set[AllocationElement]

func AnyNeededResourceSetsFunc() NeededResourceSetsFunc {
	return func(ap pfd.AtomicProcessID) *sets.Set[AllocationElement] {
		panic(fmt.Sprintf("fsm.AnyNeededResourceSetsFunc: does not affect: %q", ap))
	}
}

func NeededResourceSetsFuncByMap(m map[pfd.AtomicProcessID]*sets.Set[AllocationElement]) NeededResourceSetsFunc {
	return func(ap pfd.AtomicProcessID) *sets.Set[AllocationElement] {
		entries, ok := m[ap]
		if !ok {
			panic(fmt.Sprintf("fsm.NeededResourceSetsFuncByMap: missing needed resources set entry: %q", ap))
		}
		return entries
	}
}
