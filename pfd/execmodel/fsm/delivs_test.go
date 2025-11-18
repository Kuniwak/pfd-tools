package fsm

import (
	"fmt"
	"log/slog"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
)

func FakeDeliverableAvailableTimeFunc(p *pfd.ValidPFD, logger *slog.Logger) DeliverableAvailableTimeFunc {
	ids := p.InitialDeliverables()

	return func(id pfd.AtomicDeliverableID) execmodel.Time {
		if !ids.Contains(pfd.AtomicDeliverableID.Compare, id) {
			return execmodel.Time(-1)
		}

		idx := ids.IndexOf(pfd.AtomicDeliverableID.Compare, id)
		if idx < 0 {
			panic(fmt.Sprintf("fsm.FakeDeliverableAvailableTimeFunc: missing deliverable: %q", id))
		}

		return execmodel.Time(idx)
	}
}
