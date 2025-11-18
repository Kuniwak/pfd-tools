package fsmreporter

import (
	"cmp"
	"fmt"
	"log/slog"
	"slices"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/sets"
)

type TimelineTable []TimelineTableRow

type TimelineTableRow struct {
	AtomicProcess      pfd.AtomicProcessID       `json:"atomic_process"`
	AllocatedResources *sets.Set[fsm.ResourceID] `json:"allocated_resources"`
	NumOfComplete      int                       `json:"num_of_complete"`
	StartTime          execmodel.Time            `json:"start_time"`
	EndTime            execmodel.Time            `json:"end_time"`
}

func (a TimelineTableRow) Compare(b TimelineTableRow) int {
	var c int
	c = cmp.Compare(a.StartTime, b.StartTime)
	if c != 0 {
		return c
	}
	c = cmp.Compare(a.EndTime, b.EndTime)
	if c != 0 {
		return c
	}
	c = sets.Compare(fsm.ResourceID.Compare)(a.AllocatedResources, b.AllocatedResources)
	if c != 0 {
		return c
	}
	c = cmp.Compare(a.AtomicProcess, b.AtomicProcess)
	if c != 0 {
		return c
	}
	return cmp.Compare(a.NumOfComplete, b.NumOfComplete)
}

// BuildTimelineTable segments the Plan and generates a TimelineTable.
func BuildTimelineTable(plan *fsm.Plan, logger *slog.Logger) TimelineTable {
	tt := make(TimelineTable, 0, plan.Len())
	if plan.Len() == 0 {
		return tt
	}

	aps := sets.New(pfd.AtomicProcessID.Compare)
	for ap := range plan.InitialState.NumOfCompleteMap {
		aps.Add(pfd.AtomicProcessID.Compare, ap)
	}

	states := plan.States()

	for _, ap := range aps.Iter() {
		initNumOfComplete, ok := plan.InitialState.NumOfCompleteMap[ap]
		if !ok {
			panic(fmt.Sprintf("fsmreporter.BuildTimelineTable: missing num of complete: %q", ap))
		}

		if initNumOfComplete > 0 {
			logger.Warn("fsmreporter.BuildTimelineTable: IDAP found", "atomic_process", ap, "init_num_of_complete", initNumOfComplete)
		}

		for i := 0; i < len(states)-1; i++ {
			// NOTE: prevState --(allocation)--> nextState
			//               |                           |
			//              time                       time
			nextState := states[i+1]
			prevState := states[i]
			allocation := plan.Transitions[i].Allocation
			initNumOfComplete, ok := prevState.NumOfCompleteMap[ap]
			if !ok {
				panic(fmt.Sprintf("fsmreporter.BuildTimelineTable: missing num of complete: %q", ap))
			}

			// NOTE: Find the nearest future time among the following:
			// 1. Allocation time
			// 2. Time immediately after completion count changes
			// 3. Time after the last time

			// NOTE: Case 1
			if elem, ok := allocation[ap]; ok {
				// NOTE: Find the time immediately after the completion count changes
				startTime := prevState.Time
				var endTime execmodel.Time
				found := false
				for j := i + 1; j < len(states); j++ {
					futureNumOfComplete, ok := states[j].NumOfCompleteMap[ap]
					if !ok {
						panic(fmt.Sprintf("fsmreporter.BuildTimelineTable: missing num of complete: %q", ap))
					}

					if initNumOfComplete == futureNumOfComplete {
						continue
					}

					endTime = states[j].Time
					found = true
					i = j - 1
					break
				}
				if !found {
					logger.Warn("fsmreporter.BuildTimelineTable: end time not found", "atomic_process", ap, "start_time", startTime)
				}
				tt = append(tt, TimelineTableRow{
					AtomicProcess:      ap,
					AllocatedResources: elem.Resources,
					NumOfComplete:      initNumOfComplete,
					StartTime:          startTime,
					EndTime:            endTime,
				})
				continue
			}

			// NOTE: Case 2
			if nextState.NumOfCompleteMap[ap] != initNumOfComplete {
				logger.Warn("fsmreporter.BuildTimelineTable: IDAP found", "atomic_process", ap, "prev_num_of_complete", initNumOfComplete)
				continue
			}

			// NOTE: Case 3: Exit without adding to timeline
		}
	}

	slices.SortFunc(tt, TimelineTableRow.Compare)
	return tt
}
