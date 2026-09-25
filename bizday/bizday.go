package bizday

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/Kuniwak/pfd-tools/sets"
)

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

type BusinessHoursFunc func(day Day) (start Time, end Time)

func NewBusinessHoursFunc(start Time, duration time.Duration) (BusinessHoursFunc, error) {
	end, err := start.Add(duration)
	if err != nil {
		return nil, fmt.Errorf("NewBusinessHoursFunc: %w", err)
	}
	return func(day Day) (Time, Time) {
		return start, end
	}, nil
}

type IsBusinessDayFunc func(day Day) bool

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

type BusinessTimeFunc func(start Day, t float64) time.Time

func PassthroughBusinessTimeFunc(startTime Time, bizTimeDuration time.Duration) (BusinessTimeFunc, error) {
	hoursFunc, err := NewBusinessHoursFunc(startTime, bizTimeDuration)
	if err != nil {
		return nil, fmt.Errorf("PassthroughBusinessTimeFunc: %w", err)
	}
	return NewBusinessTime(hoursFunc, EverydayIsBusinessDayFunc()), nil
}

func NewBusinessTime(hours BusinessHoursFunc, isBiz IsBusinessDayFunc) BusinessTimeFunc {
	cachedAddBusinessDays := CachedAddBusinessDays(AddBusinessDays)
	return func(start Day, t float64) time.Time {
		if t < 0 {
			panic("BusinessTime: negative t is not supported")
		}

		day := AddBusinessDays(start, 0, isBiz)

		open0, close0 := hours(day)
		bizDur := close0.Sub(open0)
		if bizDur <= 0 {
			panic("BusinessTime: invalid business hours (non-positive length)")
		}

		days := int(math.Floor(float64(t)))
		frac := float64(t) - float64(days)

		if days > 0 {
			day = cachedAddBusinessDays(day, days, isBiz)
		}

		openN, _ := hours(day)

		offset := time.Duration(float64(bizDur) * frac)
		return Join(day, openN).Add(offset)
	}
}

type AddBusinessDaysFunc func(day Day, n int, isBiz IsBusinessDayFunc) Day

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
