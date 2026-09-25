package fsmreporter

import (
	"github.com/Kuniwak/pfd-tools/pfd"
)

func FirstExecutions(t TimelineTable) map[pfd.AtomicProcessID]TimelineTableRow {
	res := make(map[pfd.AtomicProcessID]TimelineTableRow)
	for _, row := range t {
		if row.NumOfComplete != 0 {
			continue
		}
		prev, ok := res[row.AtomicProcess]
		if !ok {
			res[row.AtomicProcess] = row
			continue
		}
		prev.StartTime = min(prev.StartTime, row.StartTime)
		prev.EndTime = max(prev.EndTime, row.EndTime)
		res[row.AtomicProcess] = prev
	}
	return res
}
