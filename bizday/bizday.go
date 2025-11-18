package bizday

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/Kuniwak/pfd-tools/sets"
)

// Day represents a date (year, month, day). Does not hold time information.
type Day struct {
	t time.Time
}

func NewDay(year int, month time.Month, day int, loc *time.Location) Day {
	return Day{t: time.Date(year, month, day, 0, 0, 0, 0, loc)}
}

func NewDayByTime(t time.Time) Day {
	return NewDay(t.Year(), t.Month(), t.Day(), t.Location())
}

func (d Day) String() string {
	return d.t.Format("2006-01-02")
}

func (d Day) Year() int {
	return d.t.Year()
}
func (d Day) Month() time.Month {
	return d.t.Month()
}
func (d Day) Day() int {
	return d.t.Day()
}
func (d Day) Weekday() time.Weekday {
	return d.t.Weekday()
}
func (d Day) Sub(d2 Day) time.Duration {
	return d.t.Sub(d2.t)
}
func (d Day) Add(d2 time.Duration) Day {
	return Day{t: d.t.Add(d2)}
}
func (d Day) AddDate(year int, month int, day int) Day {
	return Day{t: d.t.AddDate(year, month, day)}
}
func (d Day) Equal(d2 Day) bool {
	return d.t.Equal(d2.t)
}
func (d Day) Compare(d2 Day) int {
	return d.t.Compare(d2.t)
}

// Time represents time from 00:00 to 23:59:59.999999999.
// Does not hold date information.
type Time struct {
	t time.Time
}

func NewTime(hour int, minute int, second int, nanosecond int, loc *time.Location) Time {
	return Time{t: time.Date(1970, 1, 1, hour, minute, second, nanosecond, loc)}
}

func NewTimeByTime(t time.Time) Time {
	return NewTime(t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func (t Time) String() string {
	return t.t.Format("15:04:05.999999999")
}

func (t Time) Hour() int {
	return t.t.Hour()
}

func (t Time) Minute() int {
	return t.t.Minute()
}

func (t Time) Second() int {
	return t.t.Second()
}

func (t Time) Nanosecond() int {
	return t.t.Nanosecond()
}

func (t Time) Add(d time.Duration) (Time, error) {
	x := t.t.Add(d)
	if x.Year() != t.t.Year() || x.Month() != t.t.Month() || x.Day() != t.t.Day() {
		return Time{}, fmt.Errorf("Time.Add: overflow: x=%v, d=%v", x.Format(time.DateTime), d.String())
	}
	return Time{t: x}, nil
}

func (t Time) Sub(t2 Time) time.Duration {
	return t.t.Sub(t2.t)
}

func (t Time) Equal(t2 Time) bool {
	return t.t.Equal(t2.t)
}

func (t Time) Compare(t2 Time) int {
	return t.t.Compare(t2.t)
}

func Join(day Day, t Time) time.Time {
	if day.t.Location() != t.t.Location() {
		panic("Join: day and t have different locations")
	}
	return time.Date(day.t.Year(), day.t.Month(), day.t.Day(), t.t.Hour(), t.t.Minute(), t.t.Second(), t.t.Nanosecond(), day.t.Location())
}

// BusinessHoursFunc day is passed as 00:00:00.000 of that day, start and end are both times of that day, satisfy start < end, and the difference between end and start is constant for any business day.
type BusinessHoursFunc func(day Day) (start Time, end Time)

// NewBusinessHoursFunc returns a function that returns the business hours for that day when given start and duration.
func NewBusinessHoursFunc(start Time, duration time.Duration) (BusinessHoursFunc, error) {
	end, err := start.Add(duration)
	if err != nil {
		return nil, fmt.Errorf("NewBusinessHoursFunc: %w", err)
	}
	return func(day Day) (Time, Time) {
		return start, end
	}, nil
}

// IsBusinessDayFunc receives day (00:00:00.000) and returns true if that day is a business day, otherwise returns false.
type IsBusinessDayFunc func(day Day) bool

// NewIsBusinessDayFunc returns a function that returns true if the day is a business day, otherwise returns false, when given weekdays and additionalNotBusinessDays.
func NewIsBusinessDayFunc(weekdays []time.Weekday, additionalNotBusinessDays *sets.Set[Day]) IsBusinessDayFunc {
	return func(day Day) bool {
		if !slices.Contains(weekdays, day.Weekday()) {
			return false
		}
		return !additionalNotBusinessDays.Contains(Day.Compare, day)
	}
}

func NewIsBusinessDayFuncByMap(s *sets.Set[Day]) IsBusinessDayFunc {
	s2 := sets.NewWithCapacity[Day](s.Len())
	for _, day := range s.Iter() {
		s2.Add(Day.Compare, day)
	}

	return func(day Day) bool {
		return s2.Contains(Day.Compare, day)
	}
}

func EverydayIsBusinessDayFunc() IsBusinessDayFunc {
	return func(Day) bool {
		return true
	}
}

// BusinessTimeFunc is a function that returns the time after business time t >= 0 has elapsed from start. t = 1.0 means 1 business day worth of business hours.
type BusinessTimeFunc func(start Day, t float64) time.Time

// PassthroughBusinessTimeFunc is a function that returns the time after bizTimeDuration has elapsed from start. t = 1.0 means 1 business day worth of bizTimeDuration.
func PassthroughBusinessTimeFunc(startTime Time, bizTimeDuration time.Duration) (BusinessTimeFunc, error) {
	hoursFunc, err := NewBusinessHoursFunc(startTime, bizTimeDuration)
	if err != nil {
		return nil, fmt.Errorf("PassthroughBusinessTimeFunc: %w", err)
	}
	return NewBusinessTime(hoursFunc, EverydayIsBusinessDayFunc()), nil
}

// NewBusinessTime returns BusinessTimeFunc from functions that provide start time, business day determination, and business hours.
func NewBusinessTime(hours BusinessHoursFunc, isBiz IsBusinessDayFunc) BusinessTimeFunc {
	cachedAddBusinessDays := CachedAddBusinessDays(AddBusinessDays)
	return func(start Day, t float64) time.Time {
		if t < 0 {
			panic("BusinessTime: negative t is not supported")
		}

		day := AddBusinessDays(start, 0, isBiz)

		// Working hours for 1 business day in this calendar (assumed constant)
		open0, close0 := hours(day)
		bizDur := close0.Sub(open0)
		if bizDur <= 0 {
			panic("BusinessTime: invalid business hours (non-positive length)")
		}

		// Integer part = full business days, decimal part = progress within the day
		days := int(math.Floor(float64(t)))
		frac := float64(t) - float64(days) // 0 <= frac < 1

		// Skip days worth of business days first
		if days > 0 {
			day = cachedAddBusinessDays(day, days, isBiz)
		}

		// opening time of that business day + frac * bizDur
		openN, _ := hours(day)

		// Duration × decimal is truncated to prevent rounding overflow (up to just before reaching end)
		// Example: 9h * 0.5 = 4h30m
		offset := time.Duration(float64(bizDur) * frac)
		return Join(day, openN).Add(offset)
	}
}

type AddBusinessDaysFunc func(day Day, n int, isBiz IsBusinessDayFunc) Day

// AddBusinessDays advances n business days from day (00:00:00 of business day).
// isBiz determines whether each day from the next day onward is a business day.
func AddBusinessDays(day Day, n int, isBiz IsBusinessDayFunc) Day {
	base := day
	for !isBiz(base) {
		base = base.AddDate(0, 0, 1)
	}

	for added := 0; added < n; {
		base = base.AddDate(0, 0, 1)
		if isBiz(base) {
			added++
		}
	}
	return base
}

func CachedAddBusinessDays(f AddBusinessDaysFunc) AddBusinessDaysFunc {
	cache := make(map[Day]map[int]Day)
	return func(day Day, n int, isBiz IsBusinessDayFunc) Day {
		if m, ok := cache[day]; ok {
			if t, ok := m[n]; ok {
				return t
			}
			r := f(day, n, isBiz)
			m[n] = r
			return r
		}
		r := f(day, n, isBiz)
		m := map[int]Day{n: r}
		cache[day] = m
		return r
	}
}
