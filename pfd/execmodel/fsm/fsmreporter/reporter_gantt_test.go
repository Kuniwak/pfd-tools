package fsmreporter

import (
	"cmp"
	"strings"
	"testing"
	"time"

	"github.com/Kuniwak/pfd-tools/emphasis"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/sets"
)

func testTimelineTable() GoogleSpreadsheetTimelineTable {
	return GoogleSpreadsheetTimelineTable{
		{
			AtomicProcess:      "P1",
			AllocatedResources: sets.New(cmp.Compare[fsm.ResourceID], "R1"),
			NumOfReworks:       0,
			Description:        "Process One",
			StartTime:          time.Date(2021, 1, 1, 8, 0, 0, 0, time.UTC),
			EndTime:            time.Date(2021, 1, 2, 8, 0, 0, 0, time.UTC),
			Start:              0,
			End:                1,
		},
		{
			AtomicProcess:      "P2",
			AllocatedResources: sets.New(cmp.Compare[fsm.ResourceID], "R1"),
			NumOfReworks:       0,
			Description:        "Process Two",
			StartTime:          time.Date(2021, 1, 2, 8, 0, 0, 0, time.UTC),
			EndTime:            time.Date(2021, 1, 3, 8, 0, 0, 0, time.UTC),
			Start:              1,
			End:                2,
		},
	}
}

func TestTimelineTableToGoogleSpreadsheetTimelineTSV(t *testing.T) {
	cases := map[string]struct {
		Emphasis *emphasis.Set
		Expected string
	}{
		"強調なしなら Emphasis 列は増えない": {
			Emphasis: nil,
			Expected: `AtomicProcess	NumOfComplete	AllocatedResources	Description	StartTime	EndTime	Start	End
P1	0	R1	Process One	2021-01-01 08:00:00	2021-01-02 08:00:00	0	1
P2	0	R1	Process Two	2021-01-02 08:00:00	2021-01-03 08:00:00	1	2
`,
		},
		"P1 だけ強調": {
			Emphasis: emphasis.MustParse("ID\nP1\n"),
			Expected: `AtomicProcess	NumOfComplete	AllocatedResources	Description	StartTime	EndTime	Start	End	Emphasis
P1	0	R1	Process One	2021-01-01 08:00:00	2021-01-02 08:00:00	0	1	TRUE
P2	0	R1	Process Two	2021-01-02 08:00:00	2021-01-03 08:00:00	1	2	
`,
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			sb := &strings.Builder{}
			if err := TimelineTableToGoogleSpreadsheetTimelineTSV(sb, testTimelineTable(), c.Emphasis); err != nil {
				t.Fatalf("TimelineTableToGoogleSpreadsheetTimelineTSV = %v, want no error", err)
			}
			if sb.String() != c.Expected {
				t.Errorf("TimelineTableToGoogleSpreadsheetTimelineTSV =\n%s\nwant\n%s", sb.String(), c.Expected)
			}
		})
	}
}

func TestTimelineTableToPlantUMLGantt(t *testing.T) {
	cases := map[string]struct {
		Emphasis *emphasis.Set
		Expected string
	}{
		"強調なし": {
			Emphasis: nil,
			Expected: `@startgantt
Project starts 2021-01-01
-- P1 Process One --
[P1(0) R1] starts 2021-01-01 and ends 2021-01-02
-- P2 Process Two --
[P2(0) R1] starts 2021-01-02 and ends 2021-01-03
@endgantt
`,
		},
		"P1 だけ強調": {
			Emphasis: emphasis.MustParse("ID\nP1\n"),
			Expected: `@startgantt
Project starts 2021-01-01
-- P1 Process One --
[P1(0) R1] starts 2021-01-01 and ends 2021-01-02
[P1(0) R1] is colored in Salmon
-- P2 Process Two --
[P2(0) R1] starts 2021-01-02 and ends 2021-01-03
@endgantt
`,
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			sb := &strings.Builder{}
			if err := TimelineTableToPlantUMLGantt(sb, testTimelineTable(), c.Emphasis); err != nil {
				t.Fatalf("TimelineTableToPlantUMLGantt = %v, want no error", err)
			}
			if sb.String() != c.Expected {
				t.Errorf("TimelineTableToPlantUMLGantt =\n%s\nwant\n%s", sb.String(), c.Expected)
			}
		})
	}
}

func TestTimelineTableToMermaidGantt(t *testing.T) {
	longSpanTable := GoogleSpreadsheetTimelineTable{
		{
			AtomicProcess:      "P1",
			AllocatedResources: sets.New(cmp.Compare[fsm.ResourceID], "R1"),
			Description:        "Process One",
			StartTime:          time.Date(2021, 1, 1, 8, 0, 0, 0, time.UTC),
			EndTime:            time.Date(2021, 4, 2, 8, 0, 0, 0, time.UTC),
			Start:              0,
			End:                1,
		},
	}

	cases := map[string]struct {
		Table    GoogleSpreadsheetTimelineTable
		Emphasis *emphasis.Set
		Expected string
	}{
		"強調なし": {
			Table:    testTimelineTable(),
			Emphasis: nil,
			Expected: `gantt
    dateFormat YYYY-MM-DD HH:mm
    section P1 Process One
    P1[0] R1 :2021-01-01 08:00, 2021-01-02 08:00
    section P2 Process Two
    P2[0] R1 :2021-01-02 08:00, 2021-01-03 08:00
`,
		},
		"P1 だけ強調": {
			Table:    testTimelineTable(),
			Emphasis: emphasis.MustParse("ID\nP1\n"),
			Expected: `gantt
    dateFormat YYYY-MM-DD HH:mm
    section P1 Process One
    P1[0] R1 :crit, 2021-01-01 08:00, 2021-01-02 08:00
    section P2 Process Two
    P2[0] R1 :2021-01-02 08:00, 2021-01-03 08:00
`,
		},
		"空の強調 ID 表ならどれも強調しない": {
			Table:    testTimelineTable(),
			Emphasis: emphasis.MustParse("ID\n"),
			Expected: `gantt
    dateFormat YYYY-MM-DD HH:mm
    section P1 Process One
    P1[0] R1 :2021-01-01 08:00, 2021-01-02 08:00
    section P2 Process Two
    P2[0] R1 :2021-01-02 08:00, 2021-01-03 08:00
`,
		},
		"3ヶ月を超える期間なら axisFormat と tickInterval を付ける": {
			Table:    longSpanTable,
			Emphasis: nil,
			Expected: `gantt
    dateFormat YYYY-MM-DD HH:mm
    axisFormat %Y-%m
    tickInterval 1month
    section P1 Process One
    P1[0] R1 :2021-01-01 08:00, 2021-04-02 08:00
`,
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			sb := &strings.Builder{}
			if err := TimelineTableToMermaidGantt(sb, c.Table, c.Emphasis); err != nil {
				t.Fatalf("TimelineTableToMermaidGantt = %v, want no error", err)
			}
			if sb.String() != c.Expected {
				t.Errorf("TimelineTableToMermaidGantt =\n%s\nwant\n%s", sb.String(), c.Expected)
			}
		})
	}
}
