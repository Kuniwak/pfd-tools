package masterschedule

import (
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/Kuniwak/pfd-tools/emphasis"
	"github.com/Kuniwak/pfd-tools/ganttaxis"
)

type MasterSchedule []*Item

type Item struct {
	Start          time.Time
	End            time.Time
	Row            string
	RowDescription string
	Bar            string
	BarDescription string
}

type Writer func(w io.Writer, ms *MasterSchedule) error

func ParseOutputFormat(raw string, em *emphasis.Set, logger *slog.Logger) (Writer, error) {
	withEmphasisWarning := func(write func(w io.Writer, ms *MasterSchedule, em *emphasis.Set) error) Writer {
		return func(w io.Writer, ms *MasterSchedule) error {
			em.WarnUnknownIDs(logger, BarIDs(ms))
			return write(w, ms, em)
		}
	}

	switch raw {
	case "", "google-spreadsheet-tsv":
		return withEmphasisWarning(WriteGoogleSpreadsheetTSV), nil
	case "mermaid":
		return withEmphasisWarning(WriteMermaidGantt), nil
	case "plantuml":
		return withEmphasisWarning(WritePlantUMLGantt), nil
	case "plan-json", "timeline-json":
		return nil, fmt.Errorf("masterschedule.ParseOutputFormat: output format %q is not available for the master schedule (available: google-spreadsheet-tsv, mermaid, plantuml)", raw)
	default:
		return nil, fmt.Errorf("masterschedule.ParseOutputFormat: invalid output format: %q", raw)
	}
}

func NewMasterSchedule() *MasterSchedule {
	return &MasterSchedule{}
}

func (m *MasterSchedule) Add(item *Item) {
	*m = append(*m, item)
}

func NewItem(start time.Time, end time.Time, row string, rowDescription string, bar string, barDescription string) *Item {
	return &Item{Start: start, End: end, Row: row, RowDescription: rowDescription, Bar: bar, BarDescription: barDescription}
}

func BarIDs(ms *MasterSchedule) []string {
	ids := make([]string, 0, len(*ms))
	for _, item := range *ms {
		ids = append(ids, item.Bar)
	}
	return ids
}

func WriteMermaidGantt(w io.Writer, ms *MasterSchedule, em *emphasis.Set) error {
	minStart, maxEnd := ganttaxis.Span(*ms, func(item *Item) time.Time { return item.Start }, func(item *Item) time.Time { return item.End })
	if _, err := fmt.Fprint(w, ganttaxis.Header(minStart, maxEnd)); err != nil {
		return fmt.Errorf("masterschedule.WriteMermaidGantt: %w", err)
	}

	sorted := SortedItems(ms)

	var currentRow string
	for _, item := range sorted {
		rowKey := item.Row + " " + item.RowDescription
		if rowKey != currentRow {
			currentRow = rowKey
			if _, err := fmt.Fprintf(w, "    section %s %s\n", item.Row, item.RowDescription); err != nil {
				return fmt.Errorf("masterschedule.WriteMermaidGantt: %w", err)
			}
		}
		const mermaidTimeFormat = "2006-01-02 15:04"
		var tags string
		if em.Contains(item.Bar) {
			tags = emphasis.MermaidTag + ", "
		}
		if _, err := fmt.Fprintf(w, "    %s %s :%s%s, %s\n", item.Bar, item.BarDescription, tags, item.Start.Format(mermaidTimeFormat), item.End.Format(mermaidTimeFormat)); err != nil {
			return fmt.Errorf("masterschedule.WriteMermaidGantt: %w", err)
		}
	}

	return nil
}

const plantUMLDateFormat = "2006-01-02"

func WritePlantUMLGantt(w io.Writer, ms *MasterSchedule, em *emphasis.Set) error {
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

	var currentRow string
	for _, item := range sorted {
		rowKey := item.Row + " " + item.RowDescription
		if rowKey != currentRow {
			currentRow = rowKey
			if _, err := fmt.Fprintf(w, "-- %s %s --\n", item.Row, item.RowDescription); err != nil {
				return fmt.Errorf("masterschedule.WritePlantUMLGantt: %w", err)
			}
		}
		startDate := item.Start.Truncate(24 * time.Hour)
		endDate := item.End.Truncate(24 * time.Hour)
		if !endDate.After(startDate) {
			endDate = startDate.AddDate(0, 0, 1)
		}
		taskName := item.Bar + " " + item.BarDescription
		if _, err := fmt.Fprintf(w, "[%s] starts %s and ends %s\n", taskName, startDate.Format(plantUMLDateFormat), endDate.Format(plantUMLDateFormat)); err != nil {
			return fmt.Errorf("masterschedule.WritePlantUMLGantt: %w", err)
		}
		if em.Contains(item.Bar) {
			if _, err := fmt.Fprintf(w, "[%s] is colored in %s\n", taskName, emphasis.PlantUMLColor); err != nil {
				return fmt.Errorf("masterschedule.WritePlantUMLGantt: %w", err)
			}
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
		if c := strings.Compare(a.Row, b.Row); c != 0 {
			return c
		}
		if c := strings.Compare(a.RowDescription, b.RowDescription); c != 0 {
			return c
		}
		if c := strings.Compare(a.Bar, b.Bar); c != 0 {
			return c
		}
		if c := strings.Compare(a.BarDescription, b.BarDescription); c != 0 {
			return c
		}
		if c := strings.Compare(a.Start.Format(time.DateTime), b.Start.Format(time.DateTime)); c != 0 {
			return c
		}
		return strings.Compare(a.End.Format(time.DateTime), b.End.Format(time.DateTime))
	})
	return sorted
}

func WriteGoogleSpreadsheetTSV(w io.Writer, ms *MasterSchedule, em *emphasis.Set) error {
	tsvWriter := csv.NewWriter(w)
	tsvWriter.Comma = '\t'
	headers := []string{"MasterRow", "MasterRowDescription", "MasterBar", "MasterBarDescription", "Start", "End"}
	if em.Enabled() {
		headers = append(headers, emphasis.ColumnHeader)
	}
	if err := tsvWriter.Write(headers); err != nil {
		return fmt.Errorf("masterschedule.WriteGoogleSpreadsheetTSV: %w", err)
	}
	rows := make([][]string, 0, len(*ms))
	for _, item := range *ms {
		row := []string{item.Row, item.RowDescription, item.Bar, item.BarDescription, item.Start.Format(time.DateTime), item.End.Format(time.DateTime)}
		if em.Enabled() {
			if em.Contains(item.Bar) {
				row = append(row, emphasis.TrueValue)
			} else {
				row = append(row, "")
			}
		}
		rows = append(rows, row)
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
