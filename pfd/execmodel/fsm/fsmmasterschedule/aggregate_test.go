package fsmmasterschedule

import (
	"log/slog"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Kuniwak/pfd-tools/bizday"
	"github.com/Kuniwak/pfd-tools/masterschedule"
	"github.com/Kuniwak/pfd-tools/mastertsv"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmreporter"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
)

func TestNewMasterScheduleFromPlan(t *testing.T) {
	base := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	startDay := bizday.NewDayByTime(base)
	bizTimeFunc := func(start bizday.Day, tt float64) time.Time {
		return base.Add(time.Duration(tt * float64(time.Hour)))
	}

	plan := &fsm.Plan{
		InitialState: fsm.State{
			Time:             0,
			NumOfCompleteMap: map[pfd.AtomicProcessID]int{"P1": 0},
		},
		Transitions: []*fsm.Trans{
			{
				Allocation: fsm.Allocation{"P1": {Resources: sets.New(fsm.ResourceID.Compare, "R1")}},
				NextState: fsm.State{
					Time:             1,
					NumOfCompleteMap: map[pfd.AtomicProcessID]int{"P1": 1},
				},
			},
			{
				Allocation: fsm.Allocation{"P1": {Resources: sets.New(fsm.ResourceID.Compare, "R1")}},
				NextState: fsm.State{
					Time:             3,
					NumOfCompleteMap: map[pfd.AtomicProcessID]int{"P1": 2},
				},
			},
		},
	}

	testCases := map[string]struct {
		Table    *mastertsv.Table
		Expected *masterschedule.MasterSchedule
		WantErr  string
	}{
		"初回実行区間だけをバーにする": {
			Table: &mastertsv.Table{
				Classifications: []*mastertsv.Classification{
					{ID: "P1", Rows: sets.New(mastertsv.Row.Compare, mastertsv.Row("G1")), Bar: "M1"},
				},
			},
			Expected: &masterschedule.MasterSchedule{
				masterschedule.NewItem(base, base.Add(2*time.Hour), "G1", "グループ1", "M1", ""),
			},
		},
		"実行計画にない ID はエラー": {
			Table: &mastertsv.Table{
				Classifications: []*mastertsv.Classification{
					{ID: "P9", Rows: sets.New(mastertsv.Row.Compare, mastertsv.Row("G1")), Bar: "M1"},
				},
			},
			WantErr: "P9",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := NewMasterScheduleFromPlan(
				plan,
				tc.Table,
				2.0,
				bizTimeFunc,
				startDay,
				map[string]string{"G1": "グループ1"},
				map[string]string{},
				slog.New(slogtest.NewTestHandler(t)),
			)

			if tc.WantErr != "" {
				if err == nil {
					t.Fatalf("want error containing %q, got nil (result: %v)", tc.WantErr, got)
				}
				if !strings.Contains(err.Error(), tc.WantErr) {
					t.Fatalf("want error containing %q, got: %s", tc.WantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewMasterScheduleFromPlan: %s", err)
			}
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got))
			}
		})
	}
}

func TestFirstExecutionTimeline(t *testing.T) {
	testCases := map[string]struct {
		Table      *mastertsv.Table
		FirstExecs map[pfd.AtomicProcessID]fsmreporter.TimelineTableRow
		Expected   RowBarTimeline
	}{
		"1 プロセスは初回実行区間がそのままバーになる": {
			Table: &mastertsv.Table{
				Classifications: []*mastertsv.Classification{
					{ID: "P1", Rows: sets.New(mastertsv.Row.Compare, mastertsv.Row("G1")), Bar: "M1"},
				},
			},
			FirstExecs: map[pfd.AtomicProcessID]fsmreporter.TimelineTableRow{
				"P1": {AtomicProcess: "P1", StartTime: 1, EndTime: 2},
			},
			Expected: RowBarTimeline{
				"G1": {"M1": {StartTime: 1, EndTime: 2}},
			},
		},
		"同じ行とバーの複数プロセスは min 開始と max 終了に畳む": {
			Table: &mastertsv.Table{
				Classifications: []*mastertsv.Classification{
					{ID: "P1", Rows: sets.New(mastertsv.Row.Compare, mastertsv.Row("G1")), Bar: "M1"},
					{ID: "P2", Rows: sets.New(mastertsv.Row.Compare, mastertsv.Row("G1")), Bar: "M1"},
				},
			},
			FirstExecs: map[pfd.AtomicProcessID]fsmreporter.TimelineTableRow{
				"P1": {AtomicProcess: "P1", StartTime: 0, EndTime: 3},
				"P2": {AtomicProcess: "P2", StartTime: 1, EndTime: 5},
			},
			Expected: RowBarTimeline{
				"G1": {"M1": {StartTime: 0, EndTime: 5}},
			},
		},
		"行が違えばバーは独立する": {

			Table: &mastertsv.Table{
				Classifications: []*mastertsv.Classification{
					{ID: "P1", Rows: sets.New(mastertsv.Row.Compare, mastertsv.Row("G1")), Bar: "M1"},
					{ID: "P2", Rows: sets.New(mastertsv.Row.Compare, mastertsv.Row("G2")), Bar: "M1"},
				},
			},
			FirstExecs: map[pfd.AtomicProcessID]fsmreporter.TimelineTableRow{
				"P1": {AtomicProcess: "P1", StartTime: 0, EndTime: 1},
				"P2": {AtomicProcess: "P2", StartTime: 0, EndTime: 3},
			},
			Expected: RowBarTimeline{
				"G1": {"M1": {StartTime: 0, EndTime: 1}},
				"G2": {"M1": {StartTime: 0, EndTime: 3}},
			},
		},
		"多値の行は各行に寄与する": {
			Table: &mastertsv.Table{
				Classifications: []*mastertsv.Classification{
					{ID: "P1", Rows: sets.New(mastertsv.Row.Compare, mastertsv.Row("G1"), mastertsv.Row("G2")), Bar: "M1"},
				},
			},
			FirstExecs: map[pfd.AtomicProcessID]fsmreporter.TimelineTableRow{
				"P1": {AtomicProcess: "P1", StartTime: 1, EndTime: 2},
			},
			Expected: RowBarTimeline{
				"G1": {"M1": {StartTime: 1, EndTime: 2}},
				"G2": {"M1": {StartTime: 1, EndTime: 2}},
			},
		},
		"実行されなかったプロセスはバーに寄与しない": {
			Table: &mastertsv.Table{
				Classifications: []*mastertsv.Classification{
					{ID: "P1", Rows: sets.New(mastertsv.Row.Compare, mastertsv.Row("G1")), Bar: "M1"},
					{ID: "P2", Rows: sets.New(mastertsv.Row.Compare, mastertsv.Row("G1")), Bar: "M2"},
				},
			},
			FirstExecs: map[pfd.AtomicProcessID]fsmreporter.TimelineTableRow{
				"P1": {AtomicProcess: "P1", StartTime: 1, EndTime: 2},
			},
			Expected: RowBarTimeline{
				"G1": {"M1": {StartTime: 1, EndTime: 2}},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := FirstExecutionTimeline(tc.Table, tc.FirstExecs)

			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got))
			}
		})
	}
}

func TestScaleRowBarTimeline(t *testing.T) {
	testCases := map[string]struct {
		Timeline   RowBarTimeline
		Multiplier float64
		Expected   RowBarTimeline
	}{
		"乗数 1.5 で開始と終了が伸びる": {
			Timeline:   RowBarTimeline{"G1": {"M1": {StartTime: 2, EndTime: 4}}},
			Multiplier: 1.5,
			Expected:   RowBarTimeline{"G1": {"M1": {StartTime: 3, EndTime: 6}}},
		},
		"乗数 1.0 は恒等": {
			Timeline:   RowBarTimeline{"G1": {"M1": {StartTime: 2, EndTime: 4}}},
			Multiplier: 1.0,
			Expected:   RowBarTimeline{"G1": {"M1": {StartTime: 2, EndTime: 4}}},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := ScaleRowBarTimeline(tc.Timeline, tc.Multiplier)

			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got))
			}
		})
	}
}

func TestNewMasterScheduleByRowBarTimeline(t *testing.T) {
	base := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	startDay := bizday.NewDayByTime(base)
	bizTimeFunc := func(start bizday.Day, tt float64) time.Time {
		return base.Add(time.Duration(tt * float64(time.Hour)))
	}

	testCases := map[string]struct {
		Timeline RowBarTimeline
		RowDescs map[string]string
		BarDescs map[string]string
		Expected *masterschedule.MasterSchedule
	}{
		"営業時間変換と説明の付与": {
			Timeline: RowBarTimeline{"G1": {"M1": {StartTime: 0, EndTime: 2}}},
			RowDescs: map[string]string{"G1": "グループ1"},
			BarDescs: map[string]string{"M1": "マイルストーン1"},
			Expected: &masterschedule.MasterSchedule{
				masterschedule.NewItem(base, base.Add(2*time.Hour), "G1", "グループ1", "M1", "マイルストーン1"),
			},
		},
		"説明がなければ空欄": {
			Timeline: RowBarTimeline{"G1": {"M1": {StartTime: 0, EndTime: 2}}},
			RowDescs: map[string]string{},
			BarDescs: map[string]string{},
			Expected: &masterschedule.MasterSchedule{
				masterschedule.NewItem(base, base.Add(2*time.Hour), "G1", "", "M1", ""),
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := NewMasterScheduleByRowBarTimeline(tc.Timeline, bizTimeFunc, startDay, tc.RowDescs, tc.BarDescs)

			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got))
			}
		})
	}
}
