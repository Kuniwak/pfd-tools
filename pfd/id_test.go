package pfd

import "testing"

func TestIsWellFormedNodeID(t *testing.T) {
	testCases := map[string]struct {
		ID       NodeID
		Expected bool
	}{
		"process":                     {ID: "P1", Expected: true},
		"deliverable":                 {ID: "D12", Expected: true},
		"dotted":                      {ID: "D3.1", Expected: true},
		"dotted multi level":          {ID: "D3.1.2", Expected: true},
		"colon title remainder":       {ID: "検討", Expected: false},
		"prefix without number":       {ID: "P設計", Expected: false},
		"prefix with space and digit": {ID: "Phase 1", Expected: false},
		"trailing dot":                {ID: "D3.", Expected: false},
		"dot without number":          {ID: "D.1", Expected: false},
		"empty":                       {ID: "", Expected: false},
		"unnumbered process":          {ID: "(検討)", Expected: false},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if got := IsWellFormedNodeID(tc.ID); got != tc.Expected {
				t.Errorf("IsWellFormedNodeID(%q) = %v, expected %v", tc.ID, got, tc.Expected)
			}
		})
	}
}
