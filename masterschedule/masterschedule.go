package masterschedule

import (
	"encoding/csv"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"
)

type MasterSchedule []*Item

type Item struct {
	Start                time.Time
	End                  time.Time
	Group                string
	GroupDescription     string
	Milestone            string
	MilestoneDescription string
}

func NewMasterSchedule() *MasterSchedule {
	return &MasterSchedule{}
}

func (m *MasterSchedule) Add(item *Item) {
	*m = append(*m, item)
}

func NewItem(start time.Time, end time.Time, group string, groupDescription string, milestone string, milestoneDescription string) *Item {
	return &Item{Start: start, End: end, Group: group, GroupDescription: groupDescription, Milestone: milestone, MilestoneDescription: milestoneDescription}
}

func WriteGoogleSpreadsheetTSV(w io.Writer, ms *MasterSchedule) error {
	tsvWriter := csv.NewWriter(w)
	tsvWriter.Comma = '\t'
	if err := tsvWriter.Write([]string{"Group", "GroupDescription", "Milestone", "MilestoneDescription", "Start", "End"}); err != nil {
		return fmt.Errorf("masterschedule.WriteGoogleSpreadsheetTSV: %w", err)
	}
	rows := make([][]string, 0, len(*ms))
	for _, item := range *ms {
		rows = append(rows, []string{item.Group, item.GroupDescription, item.Milestone, item.MilestoneDescription, item.Start.Format(time.DateTime), item.End.Format(time.DateTime)})
	}
	slices.SortFunc(rows, func(a, b []string) int {
		return slices.CompareFunc(a, b, strings.Compare)
	})

	if err := tsvWriter.WriteAll(rows); err != nil {
		return fmt.Errorf("masterschedule.WriteGoogleSpreadsheetTSV: %w", err)
	}
	tsvWriter.Flush()
	if err := tsvWriter.Error(); err != nil {
		return fmt.Errorf("masterschedule.WriteGoogleSpreadsheetTSV: %w", err)
	}
	return nil
}
