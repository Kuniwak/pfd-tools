package fsm

import (
	"fmt"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
)

func FakeResource(n int) ResourceID {
	return ResourceID(fmt.Sprintf("R%d", n))
}

func FakeAvailableResources(n int) *sets.Set[ResourceID] {
	s := sets.New(ResourceID.Compare)
	for i := 1; i <= n; i++ {
		s.Add(ResourceID.Compare, FakeResource(i))
	}
	return s
}

func FakeNeededResourceSets(rs *sets.Set[ResourceID]) *sets.Set[AllocationElement] {
	ps := sets.PowerSet(rs, ResourceID.Compare)
	s := sets.NewWithCapacity[AllocationElement](ps.Len())
	for i, subset := range ps.Iter() {
		if subset.Len() == 0 {
			continue
		}
		// NOTE: Mix both divisible and indivisible resources.
		if i%2 == 0 {
			// NOTE: Divisible.
			s.Add(AllocationElement.Compare, AllocationElement{Resources: subset, ConsumedVolume: Volume(subset.Len())})
		} else {
			// NOTE: Indivisible.
			s.Add(AllocationElement.Compare, AllocationElement{Resources: subset, ConsumedVolume: 1})
		}
	}
	return s
}

func FakeNeededResourceSetsFunc(rs *sets.Set[ResourceID]) NeededResourceSetsFunc {
	return func(ap pfd.AtomicProcessID) *sets.Set[AllocationElement] {
		return FakeNeededResourceSets(rs)
	}
}
