package fsm

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/sets"
)

func TestPlanProcessTime(t *testing.T) {

	elem := AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1}
	plan := NewEmptyPlan(State{Time: 0})
	for _, tr := range []struct {
		Time       execmodel.Time
		Allocation Allocation
	}{
		{Time: 2, Allocation: Allocation{"P1": elem, "P2": elem}},
		{Time: 3, Allocation: Allocation{"P2": elem}},
		{Time: 7, Allocation: Allocation{"P1": elem}},
	} {
		plan.Add(&Trans{Allocation: tr.Allocation, NextState: State{Time: tr.Time}})
	}

	testCases := map[string]struct {
		AtomicProcess pfd.AtomicProcessID
		Expected      execmodel.Time
	}{
		"never allocated": {AtomicProcess: "P3", Expected: 0},
		"allocated once":  {AtomicProcess: "P2", Expected: 3},
		"allocated twice": {AtomicProcess: "P1", Expected: 6},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := plan.ProcessTime(testCase.AtomicProcess)
			if actual != testCase.Expected {
				t.Errorf("got %v, want %v", actual, testCase.Expected)
			}
		})
	}
}
