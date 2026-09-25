package fsm

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
)

func TestMinimumCompletedTimeZeroVolume(t *testing.T) {
	e := &Env{}

	tests := map[string]struct {
		currentTime  execmodel.Time
		remained     map[pfd.AtomicProcessID]Volume
		allocation   Allocation
		wantTime     execmodel.Time
		wantHasValue bool
	}{
		"zero volume completes instantly": {
			currentTime:  3,
			remained:     map[pfd.AtomicProcessID]Volume{"P1": 0},
			allocation:   Allocation{"P1": {ConsumedVolume: 1}},
			wantTime:     3,
			wantHasValue: true,
		},
		"zero and positive volume: min is the instant one": {
			currentTime:  3,
			remained:     map[pfd.AtomicProcessID]Volume{"P1": 0, "P2": 4},
			allocation:   Allocation{"P1": {ConsumedVolume: 1}, "P2": {ConsumedVolume: 2}},
			wantTime:     3,
			wantHasValue: true,
		},
		"positive volume only (regression)": {
			currentTime:  3,
			remained:     map[pfd.AtomicProcessID]Volume{"P2": 4},
			allocation:   Allocation{"P2": {ConsumedVolume: 2}},
			wantTime:     5,
			wantHasValue: true,
		},
		"empty allocation has no completion time": {
			currentTime:  3,
			remained:     map[pfd.AtomicProcessID]Volume{},
			allocation:   Allocation{},
			wantTime:     3,
			wantHasValue: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, hasValue := e.MinimumCompletedTime(tt.currentTime, tt.remained, tt.allocation)
			if hasValue != tt.wantHasValue {
				t.Fatalf("hasValue = %v, want %v", hasValue, tt.wantHasValue)
			}
			if hasValue && got != tt.wantTime {
				t.Errorf("time = %v, want %v", got, tt.wantTime)
			}
		})
	}
}
