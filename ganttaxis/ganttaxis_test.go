package ganttaxis_test

import (
	"testing"
	"time"

	"github.com/Kuniwak/pfd-tools/ganttaxis"
)

func TestHeader(t *testing.T) {
	tests := []struct {
		name  string
		start time.Time
		end   time.Time
		want  string
	}{
		{
			name:  "3ヶ月以内なら axisFormat を付けない",
			start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
			want:  "gantt\n    dateFormat YYYY-MM-DD HH:mm\n",
		},
		{
			name:  "ちょうど3ヶ月なら axisFormat を付けない",
			start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			want:  "gantt\n    dateFormat YYYY-MM-DD HH:mm\n",
		},
		{
			name:  "3ヶ月を超えたら axisFormat と tickInterval を付ける",
			start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 4, 2, 0, 0, 0, 0, time.UTC),
			want:  "gantt\n    dateFormat YYYY-MM-DD HH:mm\n    axisFormat %Y-%m\n    tickInterval 1month\n",
		},
		{
			name:  "start と end が同じ（ゼロ幅）なら axisFormat を付けない",
			start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			want:  "gantt\n    dateFormat YYYY-MM-DD HH:mm\n",
		},
		{
			name:  "end が start より前（逆転データ）なら axisFormat を付けない",
			start: time.Date(2024, 4, 2, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			want:  "gantt\n    dateFormat YYYY-MM-DD HH:mm\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ganttaxis.Header(tc.start, tc.end)
			if got != tc.want {
				t.Errorf("Header(%v, %v) = %q, want %q", tc.start, tc.end, got, tc.want)
			}
		})
	}
}

func TestSpan(t *testing.T) {
	type interval struct {
		start time.Time
		end   time.Time
	}

	tests := []struct {
		name      string
		intervals []interval
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name:      "空なら両方ゼロ値",
			intervals: nil,
			wantStart: time.Time{},
			wantEnd:   time.Time{},
		},
		{
			name: "単一区間ならその区間そのまま",
			intervals: []interval{
				{start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), end: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)},
			},
			wantStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "複数区間なら最小開始・最大終了",
			intervals: []interval{
				{start: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), end: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)},
				{start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), end: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)},
				{start: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), end: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)},
			},
			wantStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotStart, gotEnd := ganttaxis.Span(tc.intervals, func(i interval) time.Time { return i.start }, func(i interval) time.Time { return i.end })
			if !gotStart.Equal(tc.wantStart) || !gotEnd.Equal(tc.wantEnd) {
				t.Errorf("Span(%v) = (%v, %v), want (%v, %v)", tc.intervals, gotStart, gotEnd, tc.wantStart, tc.wantEnd)
			}
		})
	}
}
