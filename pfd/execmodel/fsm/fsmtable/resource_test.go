package fsmtable

import (
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/google/go-cmp/cmp"
)

func TestParseAllocationElementOK(t *testing.T) {
	testCases := map[string]struct {
		Input    string
		Expected *sets.Set[fsm.AllocationElement]
	}{
		"empty": {
			Input:    "",
			Expected: sets.New(fsm.AllocationElement.Compare),
		},
		"single-single": {
			Input: "R1:1",
			Expected: sets.New(fsm.AllocationElement.Compare,
				fsm.AllocationElement{Resources: sets.New(fsm.ResourceID.Compare, "R1"), ConsumedVolume: 1},
			),
		},
		"single-several": {
			Input: "R1:1;R2:2",
			Expected: sets.New(fsm.AllocationElement.Compare,
				fsm.AllocationElement{Resources: sets.New(fsm.ResourceID.Compare, "R1"), ConsumedVolume: 1},
				fsm.AllocationElement{Resources: sets.New(fsm.ResourceID.Compare, "R2"), ConsumedVolume: 2},
			),
		},
		"several-single": {
			Input: "R1,R2:1",
			Expected: sets.New(fsm.AllocationElement.Compare,
				fsm.AllocationElement{Resources: sets.New(fsm.ResourceID.Compare, "R1", "R2"), ConsumedVolume: 1},
			),
		},
		"trailing-semicolon": {
			Input: "R1:1;",
			Expected: sets.New(fsm.AllocationElement.Compare,
				fsm.AllocationElement{Resources: sets.New(fsm.ResourceID.Compare, "R1"), ConsumedVolume: 1},
			),
		},
		"trailing-comma": {
			Input: "R1,:1",
			Expected: sets.New(fsm.AllocationElement.Compare,
				fsm.AllocationElement{Resources: sets.New(fsm.ResourceID.Compare, "R1"), ConsumedVolume: 1},
			),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := ParseNeededResourceSetEntry(tc.Input)
			if err != nil {
				t.Fatalf("ParseAllocationElement: %v", err)
			}
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got))
			}
		})
	}
}

func TestParseAllocationElementNG(t *testing.T) {
	testCases := map[string]struct {
		Input string
	}{
		"missing-consumed-volume": {
			Input: "R1:",
		},
		"invalid-consumed-volume": {
			Input: "R1:a",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := ParseNeededResourceSetEntry(tc.Input)
			if err == nil {
				t.Fatalf("ParseAllocationElement: expected error, got nil")
			}
		})
	}
}
