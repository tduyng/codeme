package stats

import (
	"sort"
	"time"
)

// StreakCalculator calculates coding streaks
type StreakCalculator struct {
	timezone *time.Location
}

// NewStreakCalculator creates a new streak calculator
func NewStreakCalculator(timezone *time.Location) *StreakCalculator {
	if timezone == nil {
		timezone = time.UTC
	}
	return &StreakCalculator{timezone: timezone}
}

// CalculateStreak computes current and longest streaks
func (sc *StreakCalculator) CalculateStreak(activities []Activity) StreakInfo {
	if len(activities) == 0 {
		return StreakInfo{
			Current:  0,
			Longest:  0,
			IsActive: false,
		}
	}

	// Group activities by date
	dayMap := make(map[string]bool)
	var lastActivity time.Time

	for _, a := range activities {
		day := sc.startOfDay(a.Timestamp).Format("2006-01-02")
		dayMap[day] = true

		if a.Timestamp.After(lastActivity) {
			lastActivity = a.Timestamp
		}
	}

	// Get sorted list of days
	days := make([]string, 0, len(dayMap))
	for day := range dayMap {
		days = append(days, day)
	}
	sort.Strings(days)

	// Calculate streaks by checking consecutive days backwards from today
	today := sc.startOfDay(time.Now().In(sc.timezone)).Format("2006-01-02")
	yesterday := sc.startOfDay(time.Now().In(sc.timezone).AddDate(0, 0, -1)).Format("2006-01-02")

	current := 0
	longest := 0
	streak := 0

	// Check backwards from today
	checkDay := time.Now().In(sc.timezone)
	for i := range 365 { // Check last year
		dayStr := sc.startOfDay(checkDay).Format("2006-01-02")

		if dayMap[dayStr] {
			streak++
			if i <= 1 { // Today or yesterday
				current = streak
			}
			if streak > longest {
				longest = streak
			}
		} else {
			if i <= 1 { // Break in current streak
				current = 0
			}
			if streak > longest {
				longest = streak
			}
			streak = 0
		}

		checkDay = checkDay.AddDate(0, 0, -1)
	}

	// Check if streak is active (coded today or yesterday)
	isActive := dayMap[today] || (dayMap[yesterday] && !dayMap[today])

	return StreakInfo{
		Current:      current,
		Longest:      longest,
		LastActivity: lastActivity,
		IsActive:     isActive,
	}
}

func (sc *StreakCalculator) startOfDay(t time.Time) time.Time {
	y, m, d := t.In(sc.timezone).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, sc.timezone)
}
