package fsmreporter

import (
	"fmt"

	"github.com/Kuniwak/pfd-tools/pfd"
)

type TimelineKey string

func NewTimelineKey(ap pfd.NodeID, numOfReworks int) TimelineKey {
	return TimelineKey(fmt.Sprintf("%s[%d]", ap, numOfReworks))
}
