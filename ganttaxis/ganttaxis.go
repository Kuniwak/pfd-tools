package ganttaxis

import "time"

func Header(start, end time.Time) string {
	header := "gantt\n    dateFormat YYYY-MM-DD HH:mm\n"
	if end.After(start.AddDate(0, 3, 0)) {
		header += "    axisFormat %Y-%m\n    tickInterval 1month\n"
	}
	return header
}

func Span[T any](items []T, start, end func(T) time.Time) (time.Time, time.Time) {
	var minStart, maxEnd time.Time
	for _, item := range items {
		s := start(item)
		e := end(item)
		if minStart.IsZero() || s.Before(minStart) {
			minStart = s
		}
		if e.After(maxEnd) {
			maxEnd = e
		}
	}
	return minStart, maxEnd
}
