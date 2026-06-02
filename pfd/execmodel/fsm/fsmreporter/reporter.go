package fsmreporter

import (
	"cmp"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Kuniwak/pfd-tools/bizday"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/sets"
)

type GoogleSpreadsheetTimelineTable []GoogleSpreadsheetTimelineTableRow

func (g GoogleSpreadsheetTimelineTable) Headers() []string {
	return []string{"AtomicProcess", "NumOfComplete", "AllocatedResources", "Description", "StartTime", "EndTime", "Start", "End"}
}

type GoogleSpreadsheetTimelineTableRow struct {
	AtomicProcess      pfd.AtomicProcessID
	AllocatedResources *sets.Set[fsm.ResourceID]
	NumOfReworks       int
	Description        string
	StartTime          time.Time
	EndTime            time.Time
	Start              execmodel.Time
	End                execmodel.Time
}

func (r GoogleSpreadsheetTimelineTableRow) Compare(b GoogleSpreadsheetTimelineTableRow) int {
	c := time.Time.Compare(r.StartTime, b.StartTime)
	if c != 0 {
		return c
	}
	c = time.Time.Compare(r.EndTime, b.EndTime)
	if c != 0 {
		return c
	}
	return cmp.Compare(r.AtomicProcess, b.AtomicProcess)
}

func (r GoogleSpreadsheetTimelineTableRow) Values(sb *strings.Builder) []string {
	sb.Reset()
	for i, resource := range r.AllocatedResources.Iter() {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(string(resource))
	}
	return []string{string(r.AtomicProcess), strconv.Itoa(r.NumOfReworks), sb.String(), r.Description, r.StartTime.Format(time.DateTime), r.EndTime.Format(time.DateTime), strconv.FormatFloat(float64(r.Start), 'f', -1, 64), strconv.FormatFloat(float64(r.End), 'f', -1, 64)}
}

type PlanReporter func(w io.Writer, plan *fsm.Plan, descMap map[pfd.AtomicProcessID]string) error

func NewGoogleSpreadsheetTimelineTSVReporter(startDay bizday.Day, bizTimeFunc bizday.BusinessTimeFunc, logger *slog.Logger) PlanReporter {
	return func(w io.Writer, plan *fsm.Plan, descMap map[pfd.AtomicProcessID]string) error {
		tt := BuildTimelineTable(plan, logger)
		t := BuildGoogleSpreadsheetTimelineTable(tt, startDay, bizTimeFunc, descMap)
		return TimelineTableToGoogleSpreadsheetTimelineTSV(w, t)
	}
}

func BuildGoogleSpreadsheetTimelineTable(tt TimelineTable, startDay bizday.Day, bizTimeFunc bizday.BusinessTimeFunc, descMap map[pfd.AtomicProcessID]string) []GoogleSpreadsheetTimelineTableRow {
	t := make(GoogleSpreadsheetTimelineTable, len(tt))
	for i, row := range tt {
		desc, ok := descMap[row.AtomicProcess]
		if !ok {
			panic(fmt.Sprintf("fsmreporter.BuildGoogleSpreadsheetTimelineTable: missing node: %q", row.AtomicProcess))
		}

		t[i] = GoogleSpreadsheetTimelineTableRow{
			AtomicProcess:      row.AtomicProcess,
			AllocatedResources: row.AllocatedResources,
			NumOfReworks:       row.NumOfComplete,
			Description:        desc,
			StartTime:          bizTimeFunc(startDay, float64(row.StartTime)),
			EndTime:            bizTimeFunc(startDay, float64(row.EndTime)),
			Start:              row.StartTime,
			End:                row.EndTime,
		}
	}
	slices.SortFunc(t, GoogleSpreadsheetTimelineTableRow.Compare)
	return t
}

func TimelineTableToGoogleSpreadsheetTimelineTSV(w io.Writer, t GoogleSpreadsheetTimelineTable) error {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = '\t'
	csvWriter.Write(t.Headers())

	sb := &strings.Builder{}
	for _, row := range t {
		csvWriter.Write(row.Values(sb))
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("fsmreporter.TimelineTableToGoogleSpreadsheetTimelineTSV: %w", err)
	}
	return nil
}

func NewMermaidGanttReporter(startDay bizday.Day, bizTimeFunc bizday.BusinessTimeFunc, logger *slog.Logger) PlanReporter {
	return func(w io.Writer, plan *fsm.Plan, descMap map[pfd.AtomicProcessID]string) error {
		tt := BuildTimelineTable(plan, logger)
		gtt := BuildGoogleSpreadsheetTimelineTable(tt, startDay, bizTimeFunc, descMap)
		return TimelineTableToMermaidGantt(w, gtt)
	}
}

const mermaidTimeFormat = "2006-01-02 15:04"

func TimelineTableToMermaidGantt(w io.Writer, t GoogleSpreadsheetTimelineTable) error {
	if _, err := fmt.Fprint(w, "gantt\n    dateFormat YYYY-MM-DD HH:mm\n"); err != nil {
		return fmt.Errorf("fsmreporter.TimelineTableToMermaidGantt: %w", err)
	}

	var currentAP pfd.AtomicProcessID
	sb := &strings.Builder{}
	for _, row := range t {
		if row.AtomicProcess != currentAP {
			currentAP = row.AtomicProcess
			if _, err := fmt.Fprintf(w, "    section %s %s\n", row.AtomicProcess, row.Description); err != nil {
				return fmt.Errorf("fsmreporter.TimelineTableToMermaidGantt: %w", err)
			}
		}
		sb.Reset()
		for i, resource := range row.AllocatedResources.Iter() {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(string(resource))
		}
		if _, err := fmt.Fprintf(w, "    %s[%d] %s :%s, %s\n", row.AtomicProcess, row.NumOfReworks, sb.String(), row.StartTime.Format(mermaidTimeFormat), row.EndTime.Format(mermaidTimeFormat)); err != nil {
			return fmt.Errorf("fsmreporter.TimelineTableToMermaidGantt: %w", err)
		}
	}

	return nil
}

func NewPlantUMLGanttReporter(startDay bizday.Day, bizTimeFunc bizday.BusinessTimeFunc, logger *slog.Logger) PlanReporter {
	return func(w io.Writer, plan *fsm.Plan, descMap map[pfd.AtomicProcessID]string) error {
		tt := BuildTimelineTable(plan, logger)
		gtt := BuildGoogleSpreadsheetTimelineTable(tt, startDay, bizTimeFunc, descMap)
		return TimelineTableToPlantUMLGantt(w, gtt)
	}
}

const plantUMLDateFormat = "2006-01-02"

func TimelineTableToPlantUMLGantt(w io.Writer, t GoogleSpreadsheetTimelineTable) error {
	if len(t) == 0 {
		if _, err := fmt.Fprint(w, "@startgantt\n@endgantt\n"); err != nil {
			return fmt.Errorf("fsmreporter.TimelineTableToPlantUMLGantt: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintf(w, "@startgantt\nProject starts %s\n", t[0].StartTime.Format(plantUMLDateFormat)); err != nil {
		return fmt.Errorf("fsmreporter.TimelineTableToPlantUMLGantt: %w", err)
	}

	var currentAP pfd.AtomicProcessID
	sb := &strings.Builder{}
	for _, row := range t {
		if row.AtomicProcess != currentAP {
			currentAP = row.AtomicProcess
			if _, err := fmt.Fprintf(w, "-- %s %s --\n", row.AtomicProcess, row.Description); err != nil {
				return fmt.Errorf("fsmreporter.TimelineTableToPlantUMLGantt: %w", err)
			}
		}
		sb.Reset()
		for i, resource := range row.AllocatedResources.Iter() {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(string(resource))
		}
		startDate := row.StartTime.Truncate(24 * time.Hour)
		endDate := row.EndTime.Truncate(24 * time.Hour)
		if !endDate.After(startDate) {
			endDate = startDate.AddDate(0, 0, 1)
		}
		if _, err := fmt.Fprintf(w, "[%s(%d) %s] starts %s and ends %s\n", row.AtomicProcess, row.NumOfReworks, sb.String(), startDate.Format(plantUMLDateFormat), endDate.Format(plantUMLDateFormat)); err != nil {
			return fmt.Errorf("fsmreporter.TimelineTableToPlantUMLGantt: %w", err)
		}
	}

	if _, err := fmt.Fprint(w, "@endgantt\n"); err != nil {
		return fmt.Errorf("fsmreporter.TimelineTableToPlantUMLGantt: %w", err)
	}

	return nil
}

func NewPlanJSONReporter() PlanReporter {
	return func(w io.Writer, plan *fsm.Plan, _ map[pfd.AtomicProcessID]string) error {
		e := json.NewEncoder(w)
		e.SetEscapeHTML(false)
		e.SetIndent("", "  ")
		if err := e.Encode(plan); err != nil {
			return fmt.Errorf("fsmreporter.NewPlanJSONReporter: %w", err)
		}
		return nil
	}
}

func NewTimelineJSONReporter(logger *slog.Logger) PlanReporter {
	return func(w io.Writer, plan *fsm.Plan, _ map[pfd.AtomicProcessID]string) error {
		tt := BuildTimelineTable(plan, logger)
		e := json.NewEncoder(w)
		e.SetEscapeHTML(false)
		e.SetIndent("", "  ")
		if err := e.Encode(tt); err != nil {
			return fmt.Errorf("fsmreporter.NewTimelineJSONReporter: %w", err)
		}
		return nil
	}
}
