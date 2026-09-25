package fsm

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/sets"
)

func TestMaxConsumedVolume(t *testing.T) {
	testCases := map[string]struct {
		Elements *sets.Set[AllocationElement]
		Expected Volume
	}{
		"empty": {
			Elements: sets.New(AllocationElement.Compare),
			Expected: 1,
		},
		"single element of volume 1": {
			Elements: sets.New(
				AllocationElement.Compare,
				AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
			),
			Expected: 1,
		},
		"the largest of many": {
			Elements: sets.New(
				AllocationElement.Compare,
				AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1},
				AllocationElement{Resources: sets.New(ResourceID.Compare, "R1", "R2"), ConsumedVolume: 2},
			),
			Expected: 2,
		},
		"fractional volume less than 1": {
			Elements: sets.New(
				AllocationElement.Compare,
				AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 0.5},
			),
			Expected: 0.5,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := MaxConsumedVolume(testCase.Elements)
			if actual != testCase.Expected {
				t.Errorf("got %v, want %v", actual, testCase.Expected)
			}
		})
	}
}
