package stats

import "time"

// GetDailySummary creates daily summary from activities
func GetDailySummary(activities []Activity) map[string]DailyStat {
	daily := make(map[string]DailyStat)

	for _, a := range activities {
		date := a.Timestamp.Format("2006-01-02")
		stat := daily[date]
		stat.Date = date
		stat.Time += int64(a.Duration)
		stat.Lines += a.Lines
		stat.Files++
		daily[date] = stat
	}

	return daily
}

// GenerateWeeklyHeatmap creates a heatmap for the last N weeks
func GenerateWeeklyHeatmap(activities []Activity, weeks int) []HeatmapDay {
	// Create daily summary
	daily := GetDailySummary(activities)

	// Find max activity for level calculation
	maxDuration := 0.0
	for _, ds := range daily {
		duration := float64(ds.Time)
		if duration > maxDuration {
			maxDuration = duration
		}
	}

	// Calculate date range aligned to weeks (Monday-Sunday)
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}

	// End on this week's Sunday
	daysUntilSunday := 7 - weekday
	endDate := now.AddDate(0, 0, daysUntilSunday)

	// Start from weeks before that, on Monday
	totalDays := weeks * 7
	startDate := endDate.AddDate(0, 0, -(totalDays - 1))

	heatmap := make([]HeatmapDay, totalDays)

	// Generate heatmap starting from Monday of first week
	for i := range totalDays {
		d := startDate.AddDate(0, 0, i)
		date := d.Format("2006-01-02")
		ds := daily[date]

		// Calculate level (0-4)
		level := 0
		duration := float64(ds.Time)
		if maxDuration > 0 && duration > 0 {
			ratio := duration / maxDuration
			if ratio > 0.75 {
				level = 4
			} else if ratio > 0.5 {
				level = 3
			} else if ratio > 0.25 {
				level = 2
			} else {
				level = 1
			}
		}

		// Mark future days as level -1
		isFuture := d.After(now)

		heatmap[i] = HeatmapDay{
			Date:  date,
			Level: level,
			Lines: ds.Lines,
			Time:  ds.Time,
		}

		if isFuture {
			heatmap[i].Level = -1
		}
	}

	return heatmap
}
