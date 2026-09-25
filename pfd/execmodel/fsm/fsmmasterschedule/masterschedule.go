package fsmmasterschedule

import (
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
)

type Group string

func (g Group) Compare(h Group) int {
	return strings.Compare(string(g), string(h))
}

type Milestone string

func (m Milestone) Compare(h Milestone) int {
	return strings.Compare(string(m), string(h))
}

type TimelineItem struct {
	StartTime execmodel.Time
	EndTime   execmodel.Time
}

func NewTimelineItem(startTime, endTime execmodel.Time) *TimelineItem {
	return &TimelineItem{StartTime: startTime, EndTime: endTime}
}
