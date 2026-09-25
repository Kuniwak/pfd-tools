package fsm

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
)

func TestPlanAtomicProcessIDs(t *testing.T) {
	testCases := map[string]struct {
		Plan     *Plan
		Expected *sets.Set[pfd.AtomicProcessID]
	}{
		"空の計画は空集合": {
			Plan: &Plan{
				InitialState: State{NumOfCompleteMap: map[pfd.AtomicProcessID]int{}},
			},
			Expected: sets.New(pfd.AtomicProcessID.Compare),
		},
		"初期状態の全原子プロセスを返す": {
			Plan: &Plan{
				InitialState: State{NumOfCompleteMap: map[pfd.AtomicProcessID]int{"P1": 0, "P2": 0}},
			},
			Expected: sets.New(pfd.AtomicProcessID.Compare, "P1", "P2"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := tc.Plan.AtomicProcessIDs()

			if !sets.IsEqual(pfd.AtomicProcessID.Compare, got, tc.Expected) {
				t.Errorf("AtomicProcessIDs() = %v, want %v", got, tc.Expected)
			}
		})
	}
}
