package fsm

import (
	"cmp"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/maphash"
	"io"
	"maps"
	"math"
	"slices"

	"github.com/Kuniwak/pfd-tools/cmp2"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/sets"
)

// State is the state of the FSM.
type State struct {
	// Time is the current time.
	Time execmodel.Time `json:"time"`

	// RemainedVolumeMap is the remaining work volume at the current time.
	// IDs other than atomic processes are not included.
	RemainedVolumeMap map[pfd.AtomicProcessID]Volume `json:"remained_volume"`

	// RevisionMap is the revision of deliverables at the current time. A revision of 0 means not yet generated.
	// IDs other than deliverables are not included.
	RevisionMap map[pfd.AtomicDeliverableID]int `json:"revision"`

	// NumOfCompleteMap is the number of times an atomic process has completed execution at the current time.
	// IDs other than atomic processes are not included.
	NumOfCompleteMap map[pfd.AtomicProcessID]int `json:"num_of_complete"`

	// AllocationShouldContinue is the allocation from time t-1 to time t.
	AllocationShouldContinue Allocation `json:"allocation_should_continue"`

	// UpdatedDeliverablesNotHandled is the set of atomic processes that have unhandled deliverables updated at the current time.
	UpdatedDeliverablesNotHandled map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID] `json:"updated_deliverables_not_handled"`
}

// NewState returns a new State.
func NewState(
	time execmodel.Time,
	revisionMap map[pfd.AtomicDeliverableID]int,
	remainedVolumeMap map[pfd.AtomicProcessID]Volume,
	numOfReworksMap map[pfd.AtomicProcessID]int,
	prevAllocation Allocation,
	updatedDeliverablesNotHandled map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID],
) State {
	return State{
		Time:                          time,
		RevisionMap:                   revisionMap,
		RemainedVolumeMap:             remainedVolumeMap,
		NumOfCompleteMap:              numOfReworksMap,
		AllocationShouldContinue:      prevAllocation,
		UpdatedDeliverablesNotHandled: updatedDeliverablesNotHandled,
	}
}

func (s State) Compare(b State) int {
	c := cmp.Compare(s.Time, b.Time)
	if c != 0 {
		return c
	}
	c = cmp2.CompareMap(s.RevisionMap, b.RevisionMap, pfd.AtomicDeliverableID.Compare, cmp.Compare)
	if c != 0 {
		return c
	}
	c = cmp2.CompareMap(s.RemainedVolumeMap, b.RemainedVolumeMap, pfd.AtomicProcessID.Compare, cmp.Compare)
	if c != 0 {
		return c
	}
	c = cmp2.CompareMap(s.NumOfCompleteMap, b.NumOfCompleteMap, pfd.AtomicProcessID.Compare, cmp.Compare)
	if c != 0 {
		return c
	}
	return s.AllocationShouldContinue.Compare(b.AllocationShouldContinue)
}

func (s State) Write(w io.Writer) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	if err := e.Encode(s); err != nil {
		return fmt.Errorf("State.Write: %w", err)
	}
	return nil
}

type HashFunc[V any] func(v V, h *maphash.Hash) error

func HashInt(i int, h *maphash.Hash) error {
	bs := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(bs, int64(i))
	if _, err := h.Write(bs); err != nil {
		return fmt.Errorf("fsm.HashInt: %w", err)
	}
	return nil
}

func HashVolume(v Volume, h *maphash.Hash) error {
	bs := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(bs[:], math.Float64bits(float64(v)))
	if _, err := h.Write(bs[:]); err != nil {
		return fmt.Errorf("fsm.HashVolume: %w", err)
	}
	return nil
}

func HashTime(t execmodel.Time, h *maphash.Hash) error {
	bs := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(bs[:], math.Float64bits(float64(t)))
	if _, err := h.Write(bs[:]); err != nil {
		return fmt.Errorf("fsm.HashTime: %w", err)
	}
	return nil
}

func HashState(s State, h *maphash.Hash) error {
	if err := HashTime(s.Time, h); err != nil {
		return fmt.Errorf("fsm.HashState: %w", err)
	}
	if err := HashStateWithoutTime(s, h); err != nil {
		return fmt.Errorf("fsm.HashState: %w", err)
	}
	return nil
}

func HashStateWithoutTime(s State, h *maphash.Hash) error {
	if err := HashMap(pfd.AtomicDeliverableID.Compare, HashInt)(s.RevisionMap, h); err != nil {
		return fmt.Errorf("fsm.HashStateWithoutTime: %w", err)
	}
	if err := HashMap(pfd.AtomicProcessID.Compare, HashVolume)(s.RemainedVolumeMap, h); err != nil {
		return fmt.Errorf("fsm.HashStateWithoutTime: %w", err)
	}
	if err := HashMap(pfd.AtomicProcessID.Compare, HashInt)(s.NumOfCompleteMap, h); err != nil {
		return fmt.Errorf("fsm.HashStateWithoutTime: %w", err)
	}
	return nil
}

func HashSet[V any](hashFunc HashFunc[V]) HashFunc[*sets.Set[V]] {
	return func(s *sets.Set[V], h *maphash.Hash) error {
		for _, v := range s.Iter() {
			if err := hashFunc(v, h); err != nil {
				return fmt.Errorf("fsm.HashSet: %w", err)
			}
		}
		return nil
	}
}

func HashMap[K comparable, V any](compareFunc func(a, b K) int, hashFunc HashFunc[V]) HashFunc[map[K]V] {
	return func(m map[K]V, h *maphash.Hash) error {
		ks := slices.Collect(maps.Keys(m))
		slices.SortFunc(ks, compareFunc)
		for _, k := range ks {
			if err := hashFunc(m[k], h); err != nil {
				return fmt.Errorf("fsm.HashMap: %w", err)
			}
		}
		return nil
	}
}
