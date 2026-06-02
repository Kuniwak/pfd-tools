package masterschedule

import (
	"bytes"
	"testing"
	"time"
)

func TestWriteMermaidGantt(t *testing.T) {
	cases := []struct {
		name string
		ms   *MasterSchedule
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
					"G1", "Group", "M1", "Milestone 1",
				))
				return ms
			}(),
			want: "gantt\n    dateFormat YYYY-MM-DD HH:mm\n    section G1 Group\n    M1 Milestone 1 :2026-03-26 10:00, 2026-04-02 12:15\n",
		},
		{
			name: "MultipleItemsSameGroup",
			ms: func() *MasterSchedule {
				ms := NewMasterSchedule()
				ms.Add(NewItem(
					time.Date(2026, 3, 26, 10, 0, 0, 0, time.Local),
					time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
					"G1", "Group", "M1", "Milestone 1",
				))
				ms.Add(NewItem(
					time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
					time.Date(2026, 4, 7, 12, 15, 0, 0, time.Local),
					"G1", "Group", "M2", "Milestone 2",
				))
				return ms
			}(),
			want: "gantt\n    dateFormat YYYY-MM-DD HH:mm\n    section G1 Group\n    M1 Milestone 1 :2026-03-26 10:00, 2026-04-02 12:15\n    M2 Milestone 2 :2026-04-02 12:15, 2026-04-07 12:15\n",
		},
		{
			name: "MultipleGroups",
			ms: func() *MasterSchedule {
				ms := NewMasterSchedule()
				ms.Add(NewItem(
					time.Date(2026, 3, 26, 10, 0, 0, 0, time.Local),
					time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
					"G1", "Group 1", "M1", "Milestone 1",
				))
				ms.Add(NewItem(
					time.Date(2026, 4, 1, 10, 0, 0, 0, time.Local),
					time.Date(2026, 4, 5, 12, 0, 0, 0, time.Local),
					"G2", "Group 2", "M3", "Milestone 3",
				))
				return ms
			}(),
			want: "gantt\n    dateFormat YYYY-MM-DD HH:mm\n    section G1 Group 1\n    M1 Milestone 1 :2026-03-26 10:00, 2026-04-02 12:15\n    section G2 Group 2\n    M3 Milestone 3 :2026-04-01 10:00, 2026-04-05 12:00\n",
		},
		{
			name: "SortOrder",
			ms: func() *MasterSchedule {
				ms := NewMasterSchedule()
				// Add in reverse order
				ms.Add(NewItem(
					time.Date(2026, 4, 1, 10, 0, 0, 0, time.Local),
					time.Date(2026, 4, 5, 12, 0, 0, 0, time.Local),
					"G2", "Group 2", "M3", "Milestone 3",
				))
				ms.Add(NewItem(
					time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
					time.Date(2026, 4, 7, 12, 15, 0, 0, time.Local),
					"G1", "Group 1", "M2", "Milestone 2",
				))
				ms.Add(NewItem(
					time.Date(2026, 3, 26, 10, 0, 0, 0, time.Local),
					time.Date(2026, 4, 2, 12, 15, 0, 0, time.Local),
					"G1", "Group 1", "M1", "Milestone 1",
				))
				return ms
			}(),
			want: "gantt\n    dateFormat YYYY-MM-DD HH:mm\n    section G1 Group 1\n    M1 Milestone 1 :2026-03-26 10:00, 2026-04-02 12:15\n    M2 Milestone 2 :2026-04-02 12:15, 2026-04-07 12:15\n    section G2 Group 2\n    M3 Milestone 3 :2026-04-01 10:00, 2026-04-05 12:00\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := WriteMermaidGantt(&buf, tc.ms); err != nil {
				t.Fatalf("WriteMermaidGantt returned error: %v", err)
			}
			got := buf.String()
			if got != tc.want {
				t.Errorf("WriteMermaidGantt output mismatch\ngot:\n%s\nwant:\n%s", got, tc.want)
			}
		})
	}
}
