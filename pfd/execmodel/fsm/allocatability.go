package fsm

import (
	"fmt"
	"io"
	"maps"
	"slices"

	"github.com/Kuniwak/pfd-tools/pfd"
)

type AllocatabilityInfoMap map[pfd.AtomicProcessID]*AllocatabilityInfo

func (m AllocatabilityInfoMap) Write(w io.Writer) error {
	keys := slices.Collect(maps.Keys(m))
	slices.SortFunc(keys, pfd.AtomicProcessID.Compare)
	for _, key := range keys {
		if _, err := io.WriteString(w, string(key)); err != nil {
			return fmt.Errorf("fsm.AllocatabilityInfoMap.Write: %v", err)
		}
		if _, err := io.WriteString(w, ": "); err != nil {
			return fmt.Errorf("fsm.AllocatabilityInfoMap.Write: %v", err)
		}
		if err := m[key].Write(w); err != nil {
			return fmt.Errorf("fsm.AllocatabilityInfoMap.Write: %v", err)
		}
		if _, err := io.WriteString(w, "\n"); err != nil {
			return fmt.Errorf("fsm.AllocatabilityInfoMap.Write: %v", err)
		}
	}
	return nil
}
