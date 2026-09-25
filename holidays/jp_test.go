package holidays_test

import (
	"testing"
	"time"

	"github.com/Kuniwak/pfd-tools/holidays"
)

func TestJPHolidaysContains(t *testing.T) {
	jp, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}

	testCases := map[string]struct {
		Date     string
		Expected bool
	}{
		"元日 2000":     {Date: "2000-01-01", Expected: true},
		"元日 2027":     {Date: "2027-01-01", Expected: true},
		"振替休日 2027":   {Date: "2027-03-22", Expected: true},
		"勤労感謝の日 2027": {Date: "2027-11-23", Expected: true},
		"平日 2027":     {Date: "2027-11-24", Expected: false},
		"2028 は範囲外":   {Date: "2028-01-01", Expected: false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			d, err := time.ParseInLocation("2006-01-02", tc.Date, jp)
			if err != nil {
				t.Fatal(err)
			}

			if actual := holidays.JPHolidays().Contains(time.Time.Compare, d); actual != tc.Expected {
				t.Errorf("want %v, got %v", tc.Expected, actual)
			}
		})
	}
}
