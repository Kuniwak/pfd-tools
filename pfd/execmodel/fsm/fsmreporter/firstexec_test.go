package fsmreporter

import (
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/google/go-cmp/cmp"
)

func TestFirstExecutions(t *testing.T) {
	r1 := sets.New(fsm.ResourceID.Compare, "R1")

	testCases := map[string]struct {
		Table    TimelineTable
		Expected map[pfd.AtomicProcessID]TimelineTableRow
	}{
		"空のタイムラインは空": {
			Table:    TimelineTable{},
			Expected: map[pfd.AtomicProcessID]TimelineTableRow{},
		},
		"実行が1回なら初回はそれ": {
			Table: TimelineTable{
				{AtomicProcess: "P1", AllocatedResources: r1, NumOfComplete: 0, StartTime: 0, EndTime: 1},
			},
			Expected: map[pfd.AtomicProcessID]TimelineTableRow{
				"P1": {AtomicProcess: "P1", AllocatedResources: r1, NumOfComplete: 0, StartTime: 0, EndTime: 1},
			},
		},
		"手戻りの2周目は初回に含めない": {
			Table: TimelineTable{
				{AtomicProcess: "P1", AllocatedResources: r1, NumOfComplete: 0, StartTime: 0, EndTime: 1},
				{AtomicProcess: "P1", AllocatedResources: r1, NumOfComplete: 1, StartTime: 2, EndTime: 3},
			},
			Expected: map[pfd.AtomicProcessID]TimelineTableRow{
				"P1": {AtomicProcess: "P1", AllocatedResources: r1, NumOfComplete: 0, StartTime: 0, EndTime: 1},
			},
		},
		"複数プロセスはそれぞれの初回を拾う": {
			Table: TimelineTable{
				{AtomicProcess: "P1", AllocatedResources: r1, NumOfComplete: 0, StartTime: 0, EndTime: 1},
				{AtomicProcess: "P2", AllocatedResources: r1, NumOfComplete: 0, StartTime: 1, EndTime: 3},
				{AtomicProcess: "P1", AllocatedResources: r1, NumOfComplete: 1, StartTime: 3, EndTime: 4},
			},
			Expected: map[pfd.AtomicProcessID]TimelineTableRow{
				"P1": {AtomicProcess: "P1", AllocatedResources: r1, NumOfComplete: 0, StartTime: 0, EndTime: 1},
				"P2": {AtomicProcess: "P2", AllocatedResources: r1, NumOfComplete: 0, StartTime: 1, EndTime: 3},
			},
		},
		"初回が複数行に割れていたら min 開始と max 終了に畳む": {

			Table: TimelineTable{
				{AtomicProcess: "P1", AllocatedResources: r1, NumOfComplete: 0, StartTime: 0, EndTime: 0},
				{AtomicProcess: "P1", AllocatedResources: r1, NumOfComplete: 0, StartTime: 2, EndTime: 3},
			},
			Expected: map[pfd.AtomicProcessID]TimelineTableRow{
				"P1": {AtomicProcess: "P1", AllocatedResources: r1, NumOfComplete: 0, StartTime: 0, EndTime: 3},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := FirstExecutions(tc.Table)

			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got, cmp.Comparer(func(a, b *sets.Set[fsm.ResourceID]) bool {
					return sets.IsEqual(fsm.ResourceID.Compare, a, b)
				})))
			}
		})
	}
}
