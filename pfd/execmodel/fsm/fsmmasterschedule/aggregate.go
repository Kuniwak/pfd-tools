package fsmmasterschedule

import (
	"fmt"
	"log/slog"
	"maps"
	"slices"

	"github.com/Kuniwak/pfd-tools/bizday"
	"github.com/Kuniwak/pfd-tools/masterschedule"
	"github.com/Kuniwak/pfd-tools/mastertsv"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmreporter"
)

func NewMasterScheduleFromPlan(
	plan *fsm.Plan,
	t *mastertsv.Table,
	bufferMultiplier float64,
	bizTimeFunc bizday.BusinessTimeFunc,
	startDay bizday.Day,
	rowDescriptions map[string]string,
	barDescriptions map[string]string,
	logger *slog.Logger,
) (*masterschedule.MasterSchedule, error) {
	if err := mastertsv.RequireKnownIDs(t, plan.AtomicProcessIDs()); err != nil {
		return nil, fmt.Errorf("fsmmasterschedule.NewMasterScheduleFromPlan: %w", err)
	}

	firstExecs := fsmreporter.FirstExecutions(fsmreporter.BuildTimelineTable(plan, logger))
	timeline := ScaleRowBarTimeline(FirstExecutionTimeline(t, firstExecs), bufferMultiplier)
	return NewMasterScheduleByRowBarTimeline(timeline, bizTimeFunc, startDay, rowDescriptions, barDescriptions), nil
}

type RowBarTimeline map[mastertsv.Row]map[mastertsv.Bar]*TimelineItem

func FirstExecutionTimeline(t *mastertsv.Table, firstExecs map[pfd.AtomicProcessID]fsmreporter.TimelineTableRow) RowBarTimeline {
	res := make(RowBarTimeline)
	for _, c := range t.Classifications {
		exec, ok := firstExecs[c.ID]
		if !ok {
			continue
		}
		for _, row := range c.Rows.Iter() {
			bars, ok := res[row]
			if !ok {
				bars = make(map[mastertsv.Bar]*TimelineItem)
				res[row] = bars
			}
			item, ok := bars[c.Bar]
			if !ok {
				bars[c.Bar] = NewTimelineItem(exec.StartTime, exec.EndTime)
				continue
			}
			item.StartTime = min(item.StartTime, exec.StartTime)
			item.EndTime = max(item.EndTime, exec.EndTime)
		}
	}
	return res
}

func ScaleRowBarTimeline(t RowBarTimeline, multiplier float64) RowBarTimeline {
	res := make(RowBarTimeline, len(t))
	for row, bars := range t {
		newBars := make(map[mastertsv.Bar]*TimelineItem, len(bars))
		for bar, item := range bars {
			newBars[bar] = NewTimelineItem(item.StartTime*execmodel.Time(multiplier), item.EndTime*execmodel.Time(multiplier))
		}
		res[row] = newBars
	}
	return res
}

func NewMasterScheduleByRowBarTimeline(
	t RowBarTimeline,
	bizTimeFunc bizday.BusinessTimeFunc,
	startDay bizday.Day,
	rowDescriptions map[string]string,
	barDescriptions map[string]string,
) *masterschedule.MasterSchedule {
	ms := masterschedule.NewMasterSchedule()
	for _, row := range slices.SortedFunc(maps.Keys(t), mastertsv.Row.Compare) {
		bars := t[row]
		for _, bar := range slices.SortedFunc(maps.Keys(bars), mastertsv.Bar.Compare) {
			item := bars[bar]
			ms.Add(masterschedule.NewItem(
				bizTimeFunc(startDay, float64(item.StartTime)),
				bizTimeFunc(startDay, float64(item.EndTime)),
				string(row),
				rowDescriptions[string(row)],
				string(bar),
				barDescriptions[string(bar)],
			))
		}
	}
	return ms
}
