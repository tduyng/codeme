package stats

import "time"

// GenerateWeeklyHeatmap creates a heatmap for the last N weeks
func GenerateWeeklyHeatmap(daily map[string]DailyStat, weeks int) []HeatmapDay {
	// Find max activity for level calculation
	var maxLines int
	for _, ds := range daily {
		if ds.Lines > maxLines {
			maxLines = ds.Lines
		}
	}

	// Calculate start date aligned to Monday
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
	for i := 0; i < totalDays; i++ {
		d := startDate.AddDate(0, 0, i)
		date := d.Format("2006-01-02")
		ds := daily[date]

		// Calculate level (0-4)
		level := 0
		if maxLines > 0 && ds.Lines > 0 {
			ratio := float64(ds.Lines) / float64(maxLines)
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
