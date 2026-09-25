package execmodel

import (
	"testing"
)

func TestTimeApproximateEqual(t *testing.T) {
	testCases := map[string]struct {
		A        Time
		B        Time
		Expected bool
	}{
		"same value":                {A: 1, B: 1, Expected: true},
		"difference below minimum":  {A: 1, B: 1 + MinimumTime/2, Expected: true},
		"difference above minimum":  {A: 1, B: 1 + MinimumTime*2, Expected: false},
		"negative difference below": {A: 1, B: 1 - MinimumTime/2, Expected: true},
		"negative difference above": {A: 1, B: 1 - MinimumTime*2, Expected: false},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := testCase.A.ApproximateEqual(testCase.B)
			if actual != testCase.Expected {
				t.Errorf("got %v, want %v", actual, testCase.Expected)
			}
		})
	}
}
