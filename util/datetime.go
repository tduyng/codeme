// util/datetime.go
package util

import "time"

func StartOfDay(t time.Time, tz *time.Location) time.Time {
	if tz == nil {
		tz = time.UTC
	}
	y, m, d := t.In(tz).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, tz)
}

func StartOfWeek(t time.Time, tz *time.Location) time.Time {
	if tz == nil {
		tz = time.UTC
	}

	start := StartOfDay(t, tz)
	wd := int(start.Weekday())

	if wd == 0 {
		wd = 7 // Sunday -> 7
	}

	return start.AddDate(0, 0, -(wd - 1))
}

func StartOfMonth(t time.Time, tz *time.Location) time.Time {
	if tz == nil {
		tz = time.UTC
	}
	y, m, _ := t.In(tz).Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, tz)
}

func ParseWeekday(dateStr string, tz *time.Location) string {
	if tz == nil {
		tz = time.UTC
	}
	t, err := time.ParseInLocation("2006-01-02", dateStr, tz)
	if err != nil {
		return ""
	}
	return t.Weekday().String()
}

func DateString(t time.Time, tz *time.Location) string {
	if tz == nil {
		tz = time.UTC
	}
	return t.In(tz).Format("2006-01-02")
}

func DateFrom(dateStr string, tz *time.Location) time.Time {
	if tz == nil {
		tz = time.UTC
	}
	t, _ := time.ParseInLocation("2006-01-02", dateStr, tz)
	return t
}

func DaysBetween(start, end time.Time) int {
	duration := end.Sub(start)
	return int(duration.Hours() / 24)
}

func IsToday(t time.Time, tz *time.Location) bool {
	if tz == nil {
		tz = time.UTC
	}
	now := time.Now().In(tz)
	return DateString(t, tz) == DateString(now, tz)
}

func GetWeekBounds(t time.Time, tz *time.Location) (start, end time.Time) {
	start = StartOfWeek(t, tz)
	end = start.AddDate(0, 0, 7)
	return
}

func GetMonthBounds(t time.Time, tz *time.Location) (start, end time.Time) {
	start = StartOfMonth(t, tz)
	end = start.AddDate(0, 1, 0)
	return
}

func DefaultLocation(tz *time.Location) *time.Location {
	if tz != nil {
		return tz
	}
	return time.UTC // or time.Local
}
