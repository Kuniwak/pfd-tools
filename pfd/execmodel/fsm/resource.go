package fsm

import (
	"fmt"
	"hash/maphash"
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
)

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

func NewResourcesByMap(m map[ResourceID]int) *sets.Set[ResourceID] {
	s := sets.NewWithCapacity[ResourceID](len(m))
	for id, count := range m {
		for i := 1; i <= count; i++ {
			s.Add(ResourceID.Compare, ResourceID(fmt.Sprintf("%s%d", id, i)))
		}
	}
	return s
}

type NeededResourceSetsFunc func(ap pfd.AtomicProcessID) *sets.Set[AllocationElement]

func AnyNeededResourceSetsFunc() NeededResourceSetsFunc {
	return func(ap pfd.AtomicProcessID) *sets.Set[AllocationElement] {
		panic(fmt.Sprintf("fsm.AnyNeededResourceSetsFunc: does not affect: %q", ap))
	}
}

func UnlimitedNeededResourceSetsFunc() NeededResourceSetsFunc {
	return func(pfd.AtomicProcessID) *sets.Set[AllocationElement] {
		return sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare), ConsumedVolume: Volume(1)},
		)
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
