package fsmreporter

import (
	"fmt"

	"github.com/Kuniwak/pfd-tools/pfd"
)

// TimelineKey is the key for rows in the TimelineTable.
type TimelineKey string

// NewTimelineKey generates a TimelineKey from an atomic process and its execution count.
func NewTimelineKey(ap pfd.NodeID, numOfReworks int) TimelineKey {
	return TimelineKey(fmt.Sprintf("%s[%d]", ap, numOfReworks))
}
