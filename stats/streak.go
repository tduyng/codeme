package stats

import (
	"time"

	"github.com/tduyng/codeme/core"
	"github.com/tduyng/codeme/util"
)

type StreakCalculator struct {
	timezone *time.Location
}

func NewStreakCalculator(timezone *time.Location) *StreakCalculator {
	if timezone == nil {
		timezone = time.UTC
	}
	return &StreakCalculator{timezone: timezone}
}

func (sc *StreakCalculator) Calculate(activities []core.Activity) StreakInfo {
	if len(activities) == 0 {
		return StreakInfo{
			Current:  0,
			Longest:  0,
			IsActive: false,
		}
	}

	dayMap := make(map[string]bool)
	var lastActivity time.Time
	var latestDay string

	for _, a := range activities {
		day := util.DateString(a.Timestamp, sc.timezone)
		dayMap[day] = true

		if a.Timestamp.After(lastActivity) {
			lastActivity = a.Timestamp
			latestDay = day
		}
	}

	current := 0
	checkDay, _ := time.ParseInLocation("2006-01-02", latestDay, sc.timezone)

	for {
		dayStr := util.DateString(checkDay, sc.timezone)
		if !dayMap[dayStr] {
			break
		}
		current++
		checkDay = checkDay.AddDate(0, 0, -1)
	}

	longest := 0
	streak := 0

	checkDay, _ = time.ParseInLocation("2006-01-02", latestDay, sc.timezone)

	for range 365 {
		dayStr := util.DateString(checkDay, sc.timezone)

		if dayMap[dayStr] {
			streak++
			if streak > longest {
				longest = streak
			}
		} else {
			streak = 0
		}

		checkDay = checkDay.AddDate(0, 0, -1)
	}

	today := util.DateString(time.Now().In(sc.timezone), sc.timezone)
	yesterday := util.DateString(time.Now().In(sc.timezone).AddDate(0, 0, -1), sc.timezone)

	isActive := dayMap[today] || dayMap[yesterday]

	return StreakInfo{
		Current:      current,
		Longest:      longest,
		LastActivity: lastActivity,
		IsActive:     isActive,
	}
}
