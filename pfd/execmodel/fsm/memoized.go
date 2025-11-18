package fsm

import (
	"math"
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/sets"
)

// Memoized is a collection of memoized calculations that can be memoized in execution model computations.
type Memoized struct {
	StringBuilder *strings.Builder
}

func (m *Memoized) Clone() *Memoized {
	sb := strings.Builder{}
	sb.WriteString(m.StringBuilder.String())
	return &Memoized{
		StringBuilder: &sb,
	}
}

// NewMemoized returns a new Memoized.
func NewMemoized() *Memoized {
	return &Memoized{
		StringBuilder: &strings.Builder{},
	}
}

func MinimumNotGeneratedDeliverableAvailableTime(ids *sets.Set[pfd.AtomicDeliverableID], currentTime execmodel.Time, deliverableAvailableTimeFunc DeliverableAvailableTimeFunc) (execmodel.Time, bool) {
	minTime := execmodel.Time(math.MaxFloat64)
	for _, d := range ids.Iter() {
		if currentTime < deliverableAvailableTimeFunc(d) {
			minTime = min(minTime, deliverableAvailableTimeFunc(d))
		}
	}
	return minTime, minTime != execmodel.Time(math.MaxFloat64)
}
