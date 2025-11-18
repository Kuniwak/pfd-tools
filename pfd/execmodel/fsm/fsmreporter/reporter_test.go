package fsmreporter

import (
	"reflect"
	"testing"
	"time"

	"github.com/Kuniwak/pfd-tools/bizday"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/google/go-cmp/cmp"
)

func TestBuildGoogleSpreadsheetTimelineTable(t *testing.T) {
	tests := map[string]struct {
		TimelineTable TimelineTable
		StartDay      bizday.Day
		StartTime     bizday.Time
		Duration      time.Duration
		BizTimeFunc   bizday.BusinessTimeFunc
		Expected      []GoogleSpreadsheetTimelineTableRow
	}{
		"empty": {
			TimelineTable: TimelineTable{},
			StartDay:      bizday.NewDay(2021, 1, 1, time.UTC),
			StartTime:     bizday.NewTime(8, 0, 0, 0, time.UTC),
			Duration:      8 * time.Hour,
			Expected:      []GoogleSpreadsheetTimelineTableRow{},
		},
		"single": {
			TimelineTable: TimelineTable{
				{
					AtomicProcess: "P1",
					NumOfComplete: 0,
					StartTime:     0,
					EndTime:       1,
				},
			},
			StartDay:  bizday.NewDay(2021, 1, 1, time.UTC),
			StartTime: bizday.NewTime(8, 0, 0, 0, time.UTC),
			Duration:  8 * time.Hour,
			Expected: []GoogleSpreadsheetTimelineTableRow{
				{
					AtomicProcess: "P1",
					Description:   "Process One",
					NumOfReworks:  0,
					StartTime:     time.Date(2021, 1, 1, 8, 0, 0, 0, time.UTC),
					EndTime:       time.Date(2021, 1, 2, 8, 0, 0, 0, time.UTC),
					Start:         0,
					End:           1,
				},
			},
		},
		"several": {
			TimelineTable: TimelineTable{
				{
					AtomicProcess: "P1",
					NumOfComplete: 0,
					StartTime:     0,
					EndTime:       1,
				},
				{
					AtomicProcess: "P1",
					NumOfComplete: 0,
					StartTime:     1,
					EndTime:       2,
				},
			},
			StartDay:  bizday.NewDay(2021, 1, 1, time.UTC),
			StartTime: bizday.NewTime(8, 0, 0, 0, time.UTC),
			Duration:  8 * time.Hour,
			Expected: []GoogleSpreadsheetTimelineTableRow{
				{
					AtomicProcess: "P1",
					Description:   "Process One",
					NumOfReworks:  0,
					StartTime:     time.Date(2021, 1, 1, 8, 0, 0, 0, time.UTC),
					EndTime:       time.Date(2021, 1, 2, 8, 0, 0, 0, time.UTC),
					Start:         0,
					End:           1,
				},
				{
					AtomicProcess: "P1",
					Description:   "Process One",
					NumOfReworks:  0,
					StartTime:     time.Date(2021, 1, 2, 8, 0, 0, 0, time.UTC),
					EndTime:       time.Date(2021, 1, 3, 8, 0, 0, 0, time.UTC),
					Start:         1,
					End:           2,
				},
			},
		},
		"floating point": {
			TimelineTable: TimelineTable{
				{
					AtomicProcess: "P1",
					NumOfComplete: 0,
					StartTime:     0,
					EndTime:       1.5,
				},
			},
			StartDay:  bizday.NewDay(2021, 1, 1, time.UTC),
			StartTime: bizday.NewTime(8, 0, 0, 0, time.UTC),
			Duration:  8 * time.Hour,
			Expected: []GoogleSpreadsheetTimelineTableRow{
				{
					AtomicProcess: "P1",
					Description:   "Process One",
					NumOfReworks:  0,
					StartTime:     time.Date(2021, 1, 1, 8, 0, 0, 0, time.UTC),
					EndTime:       time.Date(2021, 1, 2, 12, 0, 0, 0, time.UTC),
					Start:         0,
					End:           1.5,
				},
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			btFunc, err := bizday.PassthroughBusinessTimeFunc(tc.StartTime, tc.Duration)
			if err != nil {
				t.Fatal(err)
			}
			descMap := map[pfd.AtomicProcessID]string{
				"P1": "Process One",
			}
			got := BuildGoogleSpreadsheetTimelineTable(tc.TimelineTable, tc.StartDay, btFunc, descMap)
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got))
			}
		})
	}
}
