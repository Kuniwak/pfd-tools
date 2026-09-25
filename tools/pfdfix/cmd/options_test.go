package cmd

import "testing"

func TestParseHitboxExpand(t *testing.T) {
	testCases := map[string]struct {
		Input     string
		WantX     float64
		WantY     float64
		WantError bool
	}{
		"single value":     {"1.0", 1.0, 1.0, false},
		"single shrink":    {"0.9", 0.9, 0.9, false},
		"pair":             {"1.1:0.9", 1.1, 0.9, false},
		"zero":             {"0", 0, 0, false},
		"zero pair":        {"0:0", 0, 0, false},
		"whitespace":       {" 1.1 : 0.9 ", 1.1, 0.9, false},
		"negative single":  {"-0.1", 0, 0, true},
		"negative pair":    {"1.0:-0.1", 0, 0, true},
		"non numeric":      {"abc", 0, 0, true},
		"non numeric pair": {"a:b", 0, 0, true},
		"too many parts":   {"1:2:3", 0, 0, true},
		"empty":            {"", 0, 0, true},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			x, y, err := ParseHitboxExpand(tc.Input)
			if tc.WantError {
				if err == nil {
					t.Fatalf("ParseHitboxExpand(%q) = (%v, %v, nil), want error", tc.Input, x, y)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseHitboxExpand(%q) unexpected error: %v", tc.Input, err)
			}
			if x != tc.WantX || y != tc.WantY {
				t.Errorf("ParseHitboxExpand(%q) = (%v, %v), want (%v, %v)", tc.Input, x, y, tc.WantX, tc.WantY)
			}
		})
	}
}
