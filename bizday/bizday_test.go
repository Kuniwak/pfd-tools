package bizday

import (
	"testing"
	"time"

	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/google/go-cmp/cmp"
)

func TestAddBusinessDays(t *testing.T) {
	tests := map[string]struct {
		StartDay     Day
		DurationUnit int
		Expected     Day
		BusinessDays []Day
	}{
		"only business days": {
			StartDay:     NewDay(2021, 1, 1, time.UTC),
			DurationUnit: 1,
			Expected:     NewDay(2021, 1, 2, time.UTC),
			BusinessDays: []Day{
				NewDay(2021, 1, 1, time.UTC),
				NewDay(2021, 1, 2, time.UTC),
			},
		},
		"contains non-business day": {
			StartDay:     NewDay(2021, 1, 1, time.UTC),
			DurationUnit: 1,
			Expected:     NewDay(2021, 1, 3, time.UTC),
			BusinessDays: []Day{
				NewDay(2021, 1, 1, time.UTC),
				NewDay(2021, 1, 3, time.UTC),
			},
		},
		"contains non-business day at first": {
			StartDay:     NewDay(2021, 1, 1, time.UTC),
			DurationUnit: 1,
			Expected:     NewDay(2021, 1, 3, time.UTC),
			BusinessDays: []Day{
				NewDay(2021, 1, 2, time.UTC),
				NewDay(2021, 1, 3, time.UTC),
			},
		},
		"long duration": {
			StartDay:     NewDay(2021, 1, 1, time.UTC),
			DurationUnit: 5,
			Expected:     NewDay(2021, 1, 6, time.UTC),
			BusinessDays: []Day{
				NewDay(2021, 1, 1, time.UTC),
				NewDay(2021, 1, 2, time.UTC),
				NewDay(2021, 1, 3, time.UTC),
				NewDay(2021, 1, 4, time.UTC),
				NewDay(2021, 1, 5, time.UTC),
				NewDay(2021, 1, 6, time.UTC),
			},
		},
	}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			actual := AddBusinessDays(testCase.StartDay, testCase.DurationUnit, NewIsBusinessDayFuncByMap(sets.New(Day.Compare, testCase.BusinessDays...)))
			if !actual.Equal(testCase.Expected) {
				t.Error(cmp.Diff(testCase.Expected, actual))
			}
		})
	}
}

func TestNewBusinessTime(t *testing.T) {
	tests := map[string]struct {
		StartDay              Day
		Delta                 float64
		BusinessHoursStart    Time
		BusinessHoursDuration time.Duration
		BusinessDays          []Day
		Expected              time.Time
	}{
		"0 (business day)": {
			StartDay:              NewDay(2021, 1, 1, time.UTC),
			Delta:                 0,
			BusinessHoursStart:    NewTime(8, 0, 0, 0, time.UTC),
			BusinessHoursDuration: 8 * time.Hour,
			BusinessDays: []Day{
				NewDay(2021, 1, 1, time.UTC),
			},
			Expected: time.Date(2021, 1, 1, 8, 0, 0, 0, time.UTC),
		},
		"0 (not business day)": {
			StartDay:              NewDay(2021, 1, 1, time.UTC),
			Delta:                 0,
			BusinessHoursStart:    NewTime(8, 0, 0, 0, time.UTC),
			BusinessHoursDuration: 8 * time.Hour,
			BusinessDays: []Day{
				NewDay(2021, 1, 2, time.UTC),
			},
			Expected: time.Date(2021, 1, 2, 8, 0, 0, 0, time.UTC),
		},
		"1 (only business day)": {
			StartDay:              NewDay(2021, 1, 1, time.UTC),
			Delta:                 2,
			BusinessHoursStart:    NewTime(8, 0, 0, 0, time.UTC),
			BusinessHoursDuration: 8 * time.Hour,
			BusinessDays: []Day{
				NewDay(2021, 1, 1, time.UTC),
				NewDay(2021, 1, 2, time.UTC),
				NewDay(2021, 1, 3, time.UTC),
			},
			Expected: time.Date(2021, 1, 3, 8, 0, 0, 0, time.UTC),
		},
		"contains non-business day": {
			StartDay:              NewDay(2021, 1, 1, time.UTC),
			Delta:                 2,
			BusinessHoursStart:    NewTime(8, 0, 0, 0, time.UTC),
			BusinessHoursDuration: 8 * time.Hour,
			BusinessDays: []Day{
				NewDay(2021, 1, 1, time.UTC),
				NewDay(2021, 1, 3, time.UTC),
				NewDay(2021, 1, 4, time.UTC),
			},
			Expected: time.Date(2021, 1, 4, 8, 0, 0, 0, time.UTC),
		},
		"floating point delta": {
			StartDay:              NewDay(2021, 1, 1, time.UTC),
			Delta:                 1.5,
			BusinessHoursStart:    NewTime(8, 0, 0, 0, time.UTC),
			BusinessHoursDuration: 8 * time.Hour,
			BusinessDays: []Day{
				NewDay(2021, 1, 1, time.UTC),
				NewDay(2021, 1, 3, time.UTC),
			},
			Expected: time.Date(2021, 1, 3, 12, 0, 0, 0, time.UTC),
		},
	}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			hoursFunc, err := NewBusinessHoursFunc(testCase.BusinessHoursStart, testCase.BusinessHoursDuration)
			if err != nil {
				t.Fatal(err)
			}
			actual := NewBusinessTime(
				hoursFunc,
				NewIsBusinessDayFuncByMap(sets.New(Day.Compare, testCase.BusinessDays...)),
			)(testCase.StartDay, testCase.Delta)
			if !actual.Equal(testCase.Expected) {
				t.Error(cmp.Diff(testCase.Expected, actual))
			}
		})
	}
}

func TestPassthroughBusinessTimeFunc(t *testing.T) {
	tests := map[string]struct {
		StartDay Day
		Delta    float64
		Expected time.Time
	}{
		"zero": {
			StartDay: NewDay(2021, 1, 1, time.UTC),
			Delta:    0,
			Expected: time.Date(2021, 1, 1, 8, 0, 0, 0, time.UTC),
		},
		"integer": {
			StartDay: NewDay(2021, 1, 1, time.UTC),
			Delta:    1,
			Expected: time.Date(2021, 1, 2, 8, 0, 0, 0, time.UTC),
		},
		"floating point": {
			StartDay: NewDay(2021, 1, 1, time.UTC),
			Delta:    0.5,
			Expected: time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC),
		},
	}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			btFunc, err := PassthroughBusinessTimeFunc(NewTime(8, 0, 0, 0, time.UTC), 8*time.Hour)
			if err != nil {
				t.Fatal(err)
			}
			actual := btFunc(testCase.StartDay, testCase.Delta)
			if !actual.Equal(testCase.Expected) {
				t.Error(cmp.Diff(testCase.Expected, actual))
			}
		})
	}
}

func TestNewIsBusinessDayFunc(t *testing.T) {
	tests := map[string]struct {
		Day                       Day
		Weekdays                  []time.Weekday
		AdditionalNotBusinessDays []Day
		Expected                  bool
	}{
		"not weekday": {
			Day:                       NewDay(2021, 1, 1, time.UTC),
			Weekdays:                  []time.Weekday{},
			AdditionalNotBusinessDays: []Day{},
			Expected:                  false,
		},
		"weekday": {
			Day:                       NewDay(2021, 1, 1, time.UTC),
			Weekdays:                  []time.Weekday{time.Friday},
			AdditionalNotBusinessDays: []Day{},
			Expected:                  true,
		},
		"additional not business day": {
			Day:                       NewDay(2021, 1, 1, time.UTC),
			Weekdays:                  []time.Weekday{time.Friday},
			AdditionalNotBusinessDays: []Day{NewDay(2021, 1, 1, time.UTC)},
			Expected:                  false,
		},
	}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			isBizDay := NewIsBusinessDayFunc(testCase.Weekdays, sets.New(Day.Compare, testCase.AdditionalNotBusinessDays...))
			actual := isBizDay(testCase.Day)

			if actual != testCase.Expected {
				t.Error(cmp.Diff(testCase.Expected, actual))
			}
		})
	}
}
