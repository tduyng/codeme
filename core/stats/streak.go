package stats

import "time"

// CalculateStreak computes current consecutive days streak
func CalculateStreak(daily map[string]DailyStat) int {
	if len(daily) == 0 {
		return 0
	}

	today := time.Now().Format("2006-01-02")
	streak := 0

	for i := 0; i < 365; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if ds, ok := daily[date]; ok && ds.Lines > 0 {
			streak++
		} else if date != today {
			break
		}
	}

	return streak
}

// CalculateLongestStreak finds the longest streak in last 365 days
func CalculateLongestStreak(daily map[string]DailyStat) int {
	if len(daily) == 0 {
		return 0
	}

	longest := 0
	current := 0

	for i := 364; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if ds, ok := daily[date]; ok && ds.Lines > 0 {
			current++
			if current > longest {
				longest = current
			}
		} else {
			current = 0
		}
	}

	return longest
}

// CalculateStreakInfo provides comprehensive streak information
func CalculateStreakInfo(daily map[string]DailyStat) StreakInfo {
	if len(daily) == 0 {
		return StreakInfo{}
	}

	today := time.Now().Format("2006-01-02")
	current := 0
	longest := 0
	var currentStart, bestStart, bestEnd string

	// Calculate current streak
	for i := 0; i < 365; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if ds, ok := daily[date]; ok && ds.Lines > 0 {
			if current == 0 {
				currentStart = date
			}
			current++
		} else if date != today {
			break
		}
	}

	// Calculate longest streak
	tempCurrent := 0
	var tempStart string

	for i := 364; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if ds, ok := daily[date]; ok && ds.Lines > 0 {
			if tempCurrent == 0 {
				tempStart = date
			}
			tempCurrent++
			if tempCurrent > longest {
				longest = tempCurrent
				bestStart = tempStart
				bestEnd = date
			}
		} else {
			tempCurrent = 0
		}
	}

	// Calculate weekly pattern for current week
	weeklyPattern := make([]bool, 7)
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	for i := 0; i < 7; i++ {
		dayOffset := -(weekday - 1 - i)
		date := now.AddDate(0, 0, dayOffset).Format("2006-01-02")
		if ds, ok := daily[date]; ok && ds.Lines > 0 {
			weeklyPattern[i] = true
		}
	}

	return StreakInfo{
		Current:       current,
		Longest:       longest,
		StartDate:     currentStart,
		BestStartDate: bestStart,
		BestEndDate:   bestEnd,
		WeeklyPattern: weeklyPattern,
	}
}

// CalculateBestStreak finds the best historical streak with metadata
func CalculateBestStreak(daily map[string]DailyStat) RecordStreak {
	if len(daily) == 0 {
		return RecordStreak{}
	}

	longest := 0
	current := 0
	var bestStart, bestEnd string
	var tempStart string
	totalTime := int64(0)
	totalLines := 0

	for i := 364; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if ds, ok := daily[date]; ok && ds.Lines > 0 {
			if current == 0 {
				tempStart = date
			}
			current++
			totalTime += ds.Time
			totalLines += ds.Lines

			if current > longest {
				longest = current
				bestStart = tempStart
				bestEnd = date
			}
		} else {
			current = 0
			totalTime = 0
			totalLines = 0
		}
	}

	endReason := "ended"
	if longest == CalculateStreak(daily) {
		endReason = "active"
	}

	avgTime := int64(0)
	avgLines := 0
	if longest > 0 {
		avgTime = totalTime / int64(longest)
		avgLines = totalLines / longest
	}

	return RecordStreak{
		StartDate:     bestStart,
		EndDate:       bestEnd,
		DayCount:      longest,
		DailyAvgTime:  avgTime,
		DailyAvgLines: avgLines,
		EndReason:     endReason,
		TotalTime:     totalTime,
		TotalLines:    totalLines,
	}
}
