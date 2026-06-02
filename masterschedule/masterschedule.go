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

func WriteMermaidGantt(w io.Writer, ms *MasterSchedule) error {
	if _, err := fmt.Fprint(w, "gantt\n    dateFormat YYYY-MM-DD HH:mm\n"); err != nil {
		return fmt.Errorf("masterschedule.WriteMermaidGantt: %w", err)
	}

	sorted := SortedItems(ms)

	var currentGroup string
	for _, item := range sorted {
		groupKey := item.Group + " " + item.GroupDescription
		if groupKey != currentGroup {
			currentGroup = groupKey
			if _, err := fmt.Fprintf(w, "    section %s %s\n", item.Group, item.GroupDescription); err != nil {
				return fmt.Errorf("masterschedule.WriteMermaidGantt: %w", err)
			}
		}
		const mermaidTimeFormat = "2006-01-02 15:04"
		if _, err := fmt.Fprintf(w, "    %s %s :%s, %s\n", item.Milestone, item.MilestoneDescription, item.Start.Format(mermaidTimeFormat), item.End.Format(mermaidTimeFormat)); err != nil {
			return fmt.Errorf("masterschedule.WriteMermaidGantt: %w", err)
		}
	}

	return nil
}

const plantUMLDateFormat = "2006-01-02"

func WritePlantUMLGantt(w io.Writer, ms *MasterSchedule) error {
	sorted := SortedItems(ms)

	if len(sorted) == 0 {
		if _, err := fmt.Fprint(w, "@startgantt\n@endgantt\n"); err != nil {
			return fmt.Errorf("masterschedule.WritePlantUMLGantt: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintf(w, "@startgantt\nProject starts %s\n", sorted[0].Start.Format(plantUMLDateFormat)); err != nil {
		return fmt.Errorf("masterschedule.WritePlantUMLGantt: %w", err)
	}

	var currentGroup string
	for _, item := range sorted {
		groupKey := item.Group + " " + item.GroupDescription
		if groupKey != currentGroup {
			currentGroup = groupKey
			if _, err := fmt.Fprintf(w, "-- %s %s --\n", item.Group, item.GroupDescription); err != nil {
				return fmt.Errorf("masterschedule.WritePlantUMLGantt: %w", err)
			}
		}
		startDate := item.Start.Truncate(24 * time.Hour)
		endDate := item.End.Truncate(24 * time.Hour)
		if !endDate.After(startDate) {
			endDate = startDate.AddDate(0, 0, 1)
		}
		if _, err := fmt.Fprintf(w, "[%s %s] starts %s and ends %s\n", item.Milestone, item.MilestoneDescription, startDate.Format(plantUMLDateFormat), endDate.Format(plantUMLDateFormat)); err != nil {
			return fmt.Errorf("masterschedule.WritePlantUMLGantt: %w", err)
		}
	}

	if _, err := fmt.Fprint(w, "@endgantt\n"); err != nil {
		return fmt.Errorf("masterschedule.WritePlantUMLGantt: %w", err)
	}

	return nil
}

func SortedItems(ms *MasterSchedule) []*Item {
	sorted := make([]*Item, len(*ms))
	copy(sorted, *ms)
	slices.SortFunc(sorted, func(a, b *Item) int {
		if c := strings.Compare(a.Group, b.Group); c != 0 {
			return c
		}
		if c := strings.Compare(a.GroupDescription, b.GroupDescription); c != 0 {
			return c
		}
		if c := strings.Compare(a.Milestone, b.Milestone); c != 0 {
			return c
		}
		if c := strings.Compare(a.MilestoneDescription, b.MilestoneDescription); c != 0 {
			return c
		}
		if c := strings.Compare(a.Start.Format(time.DateTime), b.Start.Format(time.DateTime)); c != 0 {
			return c
		}
		return strings.Compare(a.End.Format(time.DateTime), b.End.Format(time.DateTime))
	})
	return sorted
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
