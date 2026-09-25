package masterschedule

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/Kuniwak/pfd-tools/emphasis"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestWriteMermaidGantt(t *testing.T) {
	cases := []struct {
		name string
		ms   *MasterSchedule
		em   *emphasis.Set
		want string
	}{
		{
			name: "EmptySchedule",
			ms:   NewMasterSchedule(),
			want: "gantt\n    dateFormat YYYY-MM-DD HH:mm\n",
		},
		{
			name: "SingleItem",
			ms: func() *MasterSchedule {
				ms := NewMasterSchedule()
				ms.Add(NewItem(
					time.Date(2026, 3, 26, 10, 0, 0, 0, time.Local),
					time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
					"G1", "グループ", "M1", "マイルストーン1",
				))
				return ms
			}(),
			want: "gantt\n    dateFormat YYYY-MM-DD HH:mm\n    section G1 グループ\n    M1 マイルストーン1 :2026-03-26 10:00, 2026-04-02 12:15\n",
		},
		{
			name: "MultipleItemsSameGroup",
			ms: func() *MasterSchedule {
				ms := NewMasterSchedule()
				ms.Add(NewItem(
					time.Date(2026, 3, 26, 10, 0, 0, 0, time.Local),
					time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
					"G1", "グループ", "M1", "マイルストーン1",
				))
				ms.Add(NewItem(
					time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
					time.Date(2026, 4, 7, 12, 15, 0, 0, time.Local),
					"G1", "グループ", "M2", "マイルストーン2",
				))
				return ms
			}(),
			want: "gantt\n    dateFormat YYYY-MM-DD HH:mm\n    section G1 グループ\n    M1 マイルストーン1 :2026-03-26 10:00, 2026-04-02 12:15\n    M2 マイルストーン2 :2026-04-02 12:15, 2026-04-07 12:15\n",
		},
		{
			name: "MultipleGroups",
			ms: func() *MasterSchedule {
				ms := NewMasterSchedule()
				ms.Add(NewItem(
					time.Date(2026, 3, 26, 10, 0, 0, 0, time.Local),
					time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
					"G1", "グループ1", "M1", "マイルストーン1",
				))
				ms.Add(NewItem(
					time.Date(2026, 4, 1, 10, 0, 0, 0, time.Local),
					time.Date(2026, 4, 5, 12, 0, 0, 0, time.Local),
					"G2", "グループ2", "M3", "マイルストーン3",
				))
				return ms
			}(),
			want: "gantt\n    dateFormat YYYY-MM-DD HH:mm\n    section G1 グループ1\n    M1 マイルストーン1 :2026-03-26 10:00, 2026-04-02 12:15\n    section G2 グループ2\n    M3 マイルストーン3 :2026-04-01 10:00, 2026-04-05 12:00\n",
		},
		{
			name: "SortOrder",
			ms: func() *MasterSchedule {
				ms := NewMasterSchedule()

				ms.Add(NewItem(
					time.Date(2026, 4, 1, 10, 0, 0, 0, time.Local),
					time.Date(2026, 4, 5, 12, 0, 0, 0, time.Local),
					"G2", "グループ2", "M3", "マイルストーン3",
				))
				ms.Add(NewItem(
					time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
					time.Date(2026, 4, 7, 12, 15, 0, 0, time.Local),
					"G1", "グループ1", "M2", "マイルストーン2",
				))
				ms.Add(NewItem(
					time.Date(2026, 3, 26, 10, 0, 0, 0, time.Local),
					time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
					"G1", "グループ1", "M1", "マイルストーン1",
				))
				return ms
			}(),
			want: "gantt\n    dateFormat YYYY-MM-DD HH:mm\n    section G1 グループ1\n    M1 マイルストーン1 :2026-03-26 10:00, 2026-04-02 12:15\n    M2 マイルストーン2 :2026-04-02 12:15, 2026-04-07 12:15\n    section G2 グループ2\n    M3 マイルストーン3 :2026-04-01 10:00, 2026-04-05 12:00\n",
		},
		{
			name: "EmphasizedMilestone",
			ms: func() *MasterSchedule {
				ms := NewMasterSchedule()
				ms.Add(NewItem(
					time.Date(2026, 3, 26, 10, 0, 0, 0, time.Local),
					time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
					"G1", "グループ", "M1", "マイルストーン1",
				))
				ms.Add(NewItem(
					time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
					time.Date(2026, 4, 7, 12, 15, 0, 0, time.Local),
					"G1", "グループ", "M2", "マイルストーン2",
				))
				return ms
			}(),
			em:   emphasis.MustParse("ID\nM1\n"),
			want: "gantt\n    dateFormat YYYY-MM-DD HH:mm\n    section G1 グループ\n    M1 マイルストーン1 :crit, 2026-03-26 10:00, 2026-04-02 12:15\n    M2 マイルストーン2 :2026-04-02 12:15, 2026-04-07 12:15\n",
		},
		{
			name: "LongSchedule",
			ms: func() *MasterSchedule {
				ms := NewMasterSchedule()
				ms.Add(NewItem(
					time.Date(2026, 1, 1, 10, 0, 0, 0, time.Local),
					time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
					"G1", "グループ", "M1", "マイルストーン1",
				))
				return ms
			}(),
			want: "gantt\n    dateFormat YYYY-MM-DD HH:mm\n    axisFormat %Y-%m\n    tickInterval 1month\n    section G1 グループ\n    M1 マイルストーン1 :2026-01-01 10:00, 2026-04-02 12:15\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := WriteMermaidGantt(&buf, tc.ms, tc.em); err != nil {
				t.Fatalf("WriteMermaidGantt returned error: %v", err)
			}
			got := buf.String()
			if got != tc.want {
				t.Errorf("WriteMermaidGantt output mismatch\ngot:\n%s\nwant:\n%s", got, tc.want)
			}
		})
	}
}

func TestParseOutputFormat(t *testing.T) {
	items := &MasterSchedule{}
	items.Add(NewItem(
		time.Date(2025, 11, 18, 10, 0, 0, 0, time.Local),
		time.Date(2025, 11, 21, 14, 30, 0, 0, time.Local),
		"G1", "グループ", "M1", "マイルストーン1",
	))

	testCases := map[string]struct {
		Raw            string
		WantErr        bool
		WantOutPrefix  string
		WantErrMention string
	}{
		"an empty format is the default google-spreadsheet-tsv": {
			Raw:           "",
			WantOutPrefix: "MasterRow\t",
		},
		"google-spreadsheet-tsv": {
			Raw:           "google-spreadsheet-tsv",
			WantOutPrefix: "MasterRow\t",
		},
		"mermaid": {
			Raw:           "mermaid",
			WantOutPrefix: "gantt\n",
		},
		"plantuml": {
			Raw:           "plantuml",
			WantOutPrefix: "@startgantt\n",
		},
		"plan-json is not available for the master schedule": {
			Raw:            "plan-json",
			WantErr:        true,
			WantErrMention: "plan-json",
		},
		"timeline-json is not available for the master schedule": {
			Raw:            "timeline-json",
			WantErr:        true,
			WantErrMention: "timeline-json",
		},
		"an unknown format is invalid": {
			Raw:            "xml",
			WantErr:        true,
			WantErrMention: "invalid output format",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			writer, err := ParseOutputFormat(tc.Raw, nil, testLogger())
			if tc.WantErr {
				if err == nil {
					t.Fatalf("ParseOutputFormat(%q) = nil error, want an error", tc.Raw)
				}
				if !strings.Contains(err.Error(), tc.WantErrMention) {
					t.Errorf("expected the error to mention %q, but got: %v", tc.WantErrMention, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseOutputFormat(%q): %v", tc.Raw, err)
			}
			var buf bytes.Buffer
			if err := writer(&buf, items); err != nil {
				t.Fatalf("writer: %v", err)
			}
			if !strings.HasPrefix(buf.String(), tc.WantOutPrefix) {
				t.Errorf("output = %q, want it to start with %q", buf.String(), tc.WantOutPrefix)
			}
		})
	}
}

func TestWritePlantUMLGantt_Emphasis(t *testing.T) {
	ms := NewMasterSchedule()
	ms.Add(NewItem(
		time.Date(2026, 3, 26, 10, 0, 0, 0, time.Local),
		time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
		"G1", "グループ", "M1", "マイルストーン1",
	))
	ms.Add(NewItem(
		time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
		time.Date(2026, 4, 7, 12, 15, 0, 0, time.Local),
		"G1", "グループ", "M2", "マイルストーン2",
	))

	cases := map[string]struct {
		Emphasis *emphasis.Set
		Want     string
	}{
		"強調なし": {
			Emphasis: nil,
			Want: `@startgantt
Project starts 2026-03-26
-- G1 グループ --
[M1 マイルストーン1] starts 2026-03-26 and ends 2026-04-02
[M2 マイルストーン2] starts 2026-04-02 and ends 2026-04-07
@endgantt
`,
		},
		"M1 だけ強調": {
			Emphasis: emphasis.MustParse("ID\nM1\n"),
			Want: `@startgantt
Project starts 2026-03-26
-- G1 グループ --
[M1 マイルストーン1] starts 2026-03-26 and ends 2026-04-02
[M1 マイルストーン1] is colored in Salmon
[M2 マイルストーン2] starts 2026-04-02 and ends 2026-04-07
@endgantt
`,
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := WritePlantUMLGantt(&buf, ms, c.Emphasis); err != nil {
				t.Fatalf("WritePlantUMLGantt = %v, want no error", err)
			}
			if buf.String() != c.Want {
				t.Errorf("WritePlantUMLGantt =\n%s\nwant\n%s", buf.String(), c.Want)
			}
		})
	}
}

func TestWriteGoogleSpreadsheetTSV_Emphasis(t *testing.T) {
	ms := NewMasterSchedule()
	ms.Add(NewItem(
		time.Date(2026, 3, 26, 10, 0, 0, 0, time.Local),
		time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
		"G1", "グループ", "M1", "マイルストーン1",
	))
	ms.Add(NewItem(
		time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
		time.Date(2026, 4, 7, 12, 15, 0, 0, time.Local),
		"G1", "グループ", "M2", "マイルストーン2",
	))

	cases := map[string]struct {
		Emphasis *emphasis.Set
		Want     string
	}{
		"強調なしなら Emphasis 列は増えない": {
			Emphasis: nil,
			Want: "MasterRow\tMasterRowDescription\tMasterBar\tMasterBarDescription\tStart\tEnd\n" +
				"G1\tグループ\tM1\tマイルストーン1\t2026-03-26 10:00:00\t2026-04-02 12:15:00\n" +
				"G1\tグループ\tM2\tマイルストーン2\t2026-04-02 12:15:00\t2026-04-07 12:15:00\n",
		},
		"M1 だけ強調": {
			Emphasis: emphasis.MustParse("ID\nM1\n"),
			Want: "MasterRow\tMasterRowDescription\tMasterBar\tMasterBarDescription\tStart\tEnd\tEmphasis\n" +
				"G1\tグループ\tM1\tマイルストーン1\t2026-03-26 10:00:00\t2026-04-02 12:15:00\tTRUE\n" +
				"G1\tグループ\tM2\tマイルストーン2\t2026-04-02 12:15:00\t2026-04-07 12:15:00\t\n",
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := WriteGoogleSpreadsheetTSV(&buf, ms, c.Emphasis); err != nil {
				t.Fatalf("WriteGoogleSpreadsheetTSV = %v, want no error", err)
			}
			if buf.String() != c.Want {
				t.Errorf("WriteGoogleSpreadsheetTSV =\n%s\nwant\n%s", buf.String(), c.Want)
			}
		})
	}
}

func TestParseOutputFormat_UnknownEmphasisID(t *testing.T) {
	ms := NewMasterSchedule()
	ms.Add(NewItem(
		time.Date(2026, 3, 26, 10, 0, 0, 0, time.Local),
		time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
		"G1", "グループ", "M1", "マイルストーン1",
	))

	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelWarn}))
	writer, err := ParseOutputFormat("mermaid", emphasis.MustParse("ID\nM1\nM99\n"), logger)
	if err != nil {
		t.Fatalf("ParseOutputFormat = %v, want no error", err)
	}
	var buf bytes.Buffer
	if err := writer(&buf, ms); err != nil {
		t.Fatalf("writer = %v, want no error", err)
	}
	if !strings.Contains(logs.String(), "M99") {
		t.Errorf("log = %q, want to contain %q", logs.String(), "M99")
	}
	if strings.Contains(buf.String(), "M99") {
		t.Errorf("output = %q, want not to contain %q", buf.String(), "M99")
	}
}
