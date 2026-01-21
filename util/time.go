package util

import (
	"time"
)

// ParseWeekday returns weekday name from date string
func ParseWeekday(dateStr string, loc *time.Location) string {
	if loc == nil {
		loc = time.UTC
	}
	t, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return ""
	}
	return t.Weekday().String()
}

// GetWeekBounds returns start and end of current week
func GetWeekBounds(t time.Time) (start, end time.Time) {
	weekday := int(t.Weekday())
	start = t.AddDate(0, 0, -weekday)
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	end = start.AddDate(0, 0, 7)
	return start, end
}

// GetMonthBounds returns start and end of current month
func GetMonthBounds(t time.Time) (start, end time.Time) {
	start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	end = start.AddDate(0, 1, 0)
	return start, end
}

func DateFrom(dateStr string, loc *time.Location) time.Time {
	t, _ := time.ParseInLocation("2006-01-02", dateStr, loc)
	return t
}
