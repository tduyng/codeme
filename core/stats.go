package core

import (
	"database/sql"
	"sort"
	"time"
)

// CalculateStats computes comprehensive coding statistics from heartbeat data.
// When todayOnly is true, it filters results to show only today's data while
// still maintaining comparison data for yesterday/last week.
func CalculateStats(db *sql.DB, todayOnly bool) (Stats, error) {
	stats := Stats{
		Projects:       make(map[string]ProjectStat),
		Languages:      make(map[string]LangStat),
		DailyActivity:  make(map[string]DailyStat),
		HourlyActivity: make(map[int]int),
		TopFiles:       []FileStat{},
		WeeklyHeatmap:  []HeatmapDay{},
		Sessions:       []Session{},
		Achievements:   []Achievement{},
	}

	// Date calculations
	now := time.Now()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	// Week boundaries (Monday to Sunday)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is 7, not 0
	}
	weekStart := now.AddDate(0, 0, -(weekday - 1)).Format("2006-01-02")
	lastWeekStart := now.AddDate(0, 0, -(weekday + 6)).Format("2006-01-02")
	lastWeekEnd := now.AddDate(0, 0, -weekday).Format("2006-01-02")

	// Month boundary
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")

	query := `SELECT timestamp, file, language, project, lines FROM heartbeats ORDER BY timestamp`

	rows, err := db.Query(query)
	if err != nil {
		return stats, err
	}
	defer rows.Close()

	var lastTime time.Time
	var totalSessionTime int64
	var todaySessionTime int64
	var yesterdaySessionTime int64
	var weekSessionTime int64
	var lastWeekSessionTime int64
	var monthSessionTime int64

	files := make(map[string]bool)
	todayFiles := make(map[string]bool)
	yesterdayFiles := make(map[string]bool)
	weekFiles := make(map[string]bool)
	lastWeekFiles := make(map[string]bool)
	monthFiles := make(map[string]bool)

	fileStats := make(map[string]*FileStat)
	projectFiles := make(map[string]map[string]bool)
	langFiles := make(map[string]map[string]bool)

	// Track day-of-week activity for "most active day" calculation
	dayOfWeekTime := make(map[time.Weekday]int64)

	// Session tracking for today
	var sessionStart time.Time
	var sessionProject string
	var currentSessionTime int64

	for rows.Next() {
		var timestamp time.Time
		var file, language, project string
		var lines int

		rows.Scan(&timestamp, &file, &language, &project, &lines)

		files[file] = true
		date := timestamp.Format("2006-01-02")
		isToday := date == today
		isYesterday := date == yesterday
		isThisWeek := date >= weekStart
		isLastWeek := date >= lastWeekStart && date <= lastWeekEnd
		isThisMonth := date >= monthStart

		// Session detection (15-minute gap)
		var sessionDelta int64
		if !lastTime.IsZero() && timestamp.Sub(lastTime).Minutes() <= 15 {
			sessionDelta = int64(timestamp.Sub(lastTime).Seconds())
			totalSessionTime += sessionDelta

			if isToday {
				todaySessionTime += sessionDelta
				currentSessionTime += sessionDelta
			}
			if isYesterday {
				yesterdaySessionTime += sessionDelta
			}
			if isThisWeek {
				weekSessionTime += sessionDelta
			}
			if isLastWeek {
				lastWeekSessionTime += sessionDelta
			}
			if isThisMonth {
				monthSessionTime += sessionDelta
			}

			// Track day-of-week activity
			dayOfWeekTime[timestamp.Weekday()] += sessionDelta
		} else if isToday {
			// New session started
			if !sessionStart.IsZero() && currentSessionTime > 60 {
				// Save previous session (only if > 1 minute)
				stats.Sessions = append(stats.Sessions, Session{
					Start:    sessionStart.Format(time.RFC3339),
					End:      lastTime.Format(time.RFC3339),
					Duration: currentSessionTime,
					Project:  sessionProject,
				})
			}
			sessionStart = timestamp
			sessionProject = project
			currentSessionTime = 0
		}
		lastTime = timestamp

		// Track files per project
		if projectFiles[project] == nil {
			projectFiles[project] = make(map[string]bool)
		}
		projectFiles[project][file] = true

		// Track files per language
		if langFiles[language] == nil {
			langFiles[language] = make(map[string]bool)
		}
		langFiles[language][file] = true

		// Aggregate by project
		ps := stats.Projects[project]
		ps.Lines += lines
		ps.Time += sessionDelta
		stats.Projects[project] = ps

		// Aggregate by language
		ls := stats.Languages[language]
		ls.Lines += lines
		ls.Time += sessionDelta
		stats.Languages[language] = ls

		// Track top files
		if fileStats[file] == nil {
			fileStats[file] = &FileStat{Path: file}
		}
		fileStats[file].Lines += lines
		fileStats[file].Time += sessionDelta

		// Daily activity
		ds := stats.DailyActivity[date]
		ds.Lines += lines
		ds.Time += sessionDelta
		ds.Files++
		stats.DailyActivity[date] = ds

		// Hourly activity
		hour := timestamp.Hour()
		stats.HourlyActivity[hour]++

		// Aggregate totals
		stats.TotalLines += lines
		if isToday {
			stats.TodayLines += lines
			todayFiles[file] = true
		}
		if isYesterday {
			stats.YesterdayLines += lines
			yesterdayFiles[file] = true
		}
		if isThisWeek {
			stats.WeekLines += lines
			weekFiles[file] = true
		}
		if isLastWeek {
			stats.LastWeekLines += lines
			lastWeekFiles[file] = true
		}
		if isThisMonth {
			stats.MonthLines += lines
			monthFiles[file] = true
		}
	}

	// Save final session if still active
	if !sessionStart.IsZero() && currentSessionTime > 60 {
		stats.Sessions = append(stats.Sessions, Session{
			Start:    sessionStart.Format(time.RFC3339),
			End:      lastTime.Format(time.RFC3339),
			Duration: currentSessionTime,
			Project:  sessionProject,
		})
	}

	// Add file counts to projects and languages
	for project, ps := range stats.Projects {
		ps.Files = len(projectFiles[project])
		stats.Projects[project] = ps
	}
	for lang, ls := range stats.Languages {
		ls.Files = len(langFiles[lang])
		stats.Languages[lang] = ls
	}

	// Get top 10 files sorted by time
	for _, fs := range fileStats {
		stats.TopFiles = append(stats.TopFiles, *fs)
	}
	sort.Slice(stats.TopFiles, func(i, j int) bool {
		return stats.TopFiles[i].Time > stats.TopFiles[j].Time
	})
	if len(stats.TopFiles) > 10 {
		stats.TopFiles = stats.TopFiles[:10]
	}

	// Set time stats
	stats.TotalTime = totalSessionTime
	stats.TodayTime = todaySessionTime
	stats.TodayFiles = len(todayFiles)
	stats.YesterdayTime = yesterdaySessionTime
	stats.YesterdayFiles = len(yesterdayFiles)
	stats.WeekTime = weekSessionTime
	stats.WeekFiles = len(weekFiles)
	stats.LastWeekTime = lastWeekSessionTime
	stats.LastWeekFiles = len(lastWeekFiles)
	stats.MonthTime = monthSessionTime
	stats.MonthFiles = len(monthFiles)
	stats.TotalFiles = len(files)

	// Calculate streaks
	stats.Streak = calculateStreak(stats.DailyActivity)
	stats.LongestStreak = calculateLongestStreak(stats.DailyActivity)

	// Calculate peak hours
	stats.MostActiveHour = calculateMostActiveHour(stats.HourlyActivity)

	// Calculate most active day of week
	stats.MostActiveDay, stats.MostActiveDayTime = calculateMostActiveDay(dayOfWeekTime)

	// Generate weekly heatmap (last 12 weeks, aligned to Monday-Sunday)
	stats.WeeklyHeatmap = generateWeeklyHeatmap(stats.DailyActivity, 12)

	// Generate achievements
	stats.Achievements = calculateAchievements(stats)

	// Populate Today field for nvim plugin compatibility
	if todayStat, ok := stats.DailyActivity[today]; ok {
		stats.Today = todayStat
	} else {
		stats.Today = DailyStat{Lines: stats.TodayLines, Time: stats.TodayTime, Files: len(todayFiles)}
	}

	// If todayOnly is requested, filter the results to show only today's data
	if todayOnly {
		stats = filterTodayOnly(db, stats, today)
	}

	return stats, nil
}

// filterTodayOnly recalculates stats for today only while preserving comparison data
func filterTodayOnly(db *sql.DB, stats Stats, today string) Stats {
	todayProjects := make(map[string]ProjectStat)
	todayLanguages := make(map[string]LangStat)
	todayTopFiles := []FileStat{}
	todayHourlyActivity := make(map[int]int)

	todayQuery := `SELECT timestamp, file, language, project, lines FROM heartbeats 
			 WHERE substr(timestamp, 1, 10) = '` + today + `' ORDER BY timestamp`

	todayRows, err := db.Query(todayQuery)
	if err != nil {
		return stats
	}
	defer todayRows.Close()

	var todayLastTime time.Time
	var todayOnlySessionTime int64
	todayFileStats := make(map[string]*FileStat)
	todayProjectFiles := make(map[string]map[string]bool)
	todayLangFiles := make(map[string]map[string]bool)
	todayOnlyFiles := make(map[string]bool)
	todayTotalLines := 0

	for todayRows.Next() {
		var timestamp time.Time
		var file, language, project string
		var lines int

		todayRows.Scan(&timestamp, &file, &language, &project, &lines)

		todayOnlyFiles[file] = true

		// Session detection for today only
		var sessionDelta int64
		if !todayLastTime.IsZero() && timestamp.Sub(todayLastTime).Minutes() <= 15 {
			sessionDelta = int64(timestamp.Sub(todayLastTime).Seconds())
			todayOnlySessionTime += sessionDelta
		}
		todayLastTime = timestamp

		// Track files per project (today only)
		if todayProjectFiles[project] == nil {
			todayProjectFiles[project] = make(map[string]bool)
		}
		todayProjectFiles[project][file] = true

		// Track files per language (today only)
		if todayLangFiles[language] == nil {
			todayLangFiles[language] = make(map[string]bool)
		}
		todayLangFiles[language][file] = true

		// Aggregate by project (today only)
		ps := todayProjects[project]
		ps.Lines += lines
		ps.Time += sessionDelta
		todayProjects[project] = ps

		// Aggregate by language (today only)
		ls := todayLanguages[language]
		ls.Lines += lines
		ls.Time += sessionDelta
		todayLanguages[language] = ls

		// Track top files (today only)
		if todayFileStats[file] == nil {
			todayFileStats[file] = &FileStat{Path: file}
		}
		todayFileStats[file].Lines += lines
		todayFileStats[file].Time += sessionDelta

		// Hourly activity (today only)
		hour := timestamp.Hour()
		todayHourlyActivity[hour]++

		todayTotalLines += lines
	}

	// Add file counts (today only)
	for project, ps := range todayProjects {
		ps.Files = len(todayProjectFiles[project])
		todayProjects[project] = ps
	}
	for lang, ls := range todayLanguages {
		ls.Files = len(todayLangFiles[lang])
		todayLanguages[lang] = ls
	}

	// Get top files (today only) sorted by time
	for _, fs := range todayFileStats {
		todayTopFiles = append(todayTopFiles, *fs)
	}
	sort.Slice(todayTopFiles, func(i, j int) bool {
		return todayTopFiles[i].Time > todayTopFiles[j].Time
	})
	if len(todayTopFiles) > 10 {
		todayTopFiles = todayTopFiles[:10]
	}

	// Replace stats with today-only data (preserve comparison stats)
	stats.Projects = todayProjects
	stats.Languages = todayLanguages
	stats.TopFiles = todayTopFiles
	stats.HourlyActivity = todayHourlyActivity
	stats.TotalTime = todayOnlySessionTime
	stats.TotalLines = todayTotalLines
	stats.TotalFiles = len(todayOnlyFiles)
	stats.TodayTime = todayOnlySessionTime
	stats.TodayLines = todayTotalLines

	// For today-only, streak should be 1 if there's activity, 0 if not
	if todayTotalLines > 0 {
		stats.Streak = 1
	} else {
		stats.Streak = 0
	}

	return stats
}

func calculateStreak(daily map[string]DailyStat) int {
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

func calculateLongestStreak(daily map[string]DailyStat) int {
	if len(daily) == 0 {
		return 0
	}

	longest := 0
	current := 0

	// Check last 365 days
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

func calculateMostActiveHour(hourly map[int]int) int {
	maxHour := 0
	maxCount := 0

	for hour, count := range hourly {
		if count > maxCount {
			maxCount = count
			maxHour = hour
		}
	}

	return maxHour
}

func calculateMostActiveDay(dayTime map[time.Weekday]int64) (string, int64) {
	dayNames := map[time.Weekday]string{
		time.Sunday:    "Sunday",
		time.Monday:    "Monday",
		time.Tuesday:   "Tuesday",
		time.Wednesday: "Wednesday",
		time.Thursday:  "Thursday",
		time.Friday:    "Friday",
		time.Saturday:  "Saturday",
	}

	var maxDay time.Weekday
	var maxTime int64

	for day, t := range dayTime {
		if t > maxTime {
			maxTime = t
			maxDay = day
		}
	}

	return dayNames[maxDay], maxTime
}

func generateWeeklyHeatmap(daily map[string]DailyStat, weeks int) []HeatmapDay {
	// Find max activity for level calculation
	var maxLines int
	for _, ds := range daily {
		if ds.Lines > maxLines {
			maxLines = ds.Lines
		}
	}

	// Calculate start date aligned to Monday
	// We want exactly `weeks` complete weeks, ending on the current week's Sunday
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}

	// End on this week's Sunday (days until Sunday from today)
	daysUntilSunday := 7 - weekday
	endDate := now.AddDate(0, 0, daysUntilSunday)

	// Start from `weeks` weeks before that, on Monday
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

		// Mark future days as level -1 so UI can skip/dim them
		isFuture := d.After(now)

		heatmap[i] = HeatmapDay{
			Date:  date,
			Level: level,
			Lines: ds.Lines,
			Time:  ds.Time,
		}

		if isFuture {
			heatmap[i].Level = -1 // Future day marker
		}
	}

	return heatmap
}

func calculateAchievements(stats Stats) []Achievement {
	achievements := []Achievement{
		{
			ID:          "first_heartbeat",
			Name:        "First Steps",
			Description: "Track your first coding activity",
			Icon:        "🎯",
			Unlocked:    stats.TotalLines > 0,
		},
		{
			ID:          "streak_3",
			Name:        "Getting Started",
			Description: "Maintain a 3-day coding streak",
			Icon:        "🔥",
			Unlocked:    stats.Streak >= 3 || stats.LongestStreak >= 3,
		},
		{
			ID:          "streak_7",
			Name:        "Weekly Warrior",
			Description: "Maintain a 7-day coding streak",
			Icon:        "⚡",
			Unlocked:    stats.Streak >= 7 || stats.LongestStreak >= 7,
		},
		{
			ID:          "streak_30",
			Name:        "Monthly Master",
			Description: "Maintain a 30-day coding streak",
			Icon:        "👑",
			Unlocked:    stats.Streak >= 30 || stats.LongestStreak >= 30,
		},
		{
			ID:          "lines_1000",
			Name:        "Code Machine",
			Description: "Write 1,000 lines of code",
			Icon:        "💻",
			Unlocked:    stats.TotalLines >= 1000,
		},
		{
			ID:          "lines_10000",
			Name:        "Prolific Programmer",
			Description: "Write 10,000 lines of code",
			Icon:        "🚀",
			Unlocked:    stats.TotalLines >= 10000,
		},
		{
			ID:          "hours_10",
			Name:        "Dedicated Developer",
			Description: "Code for 10 hours total",
			Icon:        "⏰",
			Unlocked:    stats.TotalTime >= 36000, // 10 hours in seconds
		},
		{
			ID:          "hours_100",
			Name:        "Century Coder",
			Description: "Code for 100 hours total",
			Icon:        "🏆",
			Unlocked:    stats.TotalTime >= 360000, // 100 hours in seconds
		},
		{
			ID:          "polyglot_5",
			Name:        "Polyglot",
			Description: "Code in 5 different languages",
			Icon:        "🌍",
			Unlocked:    len(stats.Languages) >= 5,
		},
		{
			ID:          "early_bird",
			Name:        "Early Bird",
			Description: "Code before 7 AM",
			Icon:        "🌅",
			Unlocked:    stats.HourlyActivity[5] > 0 || stats.HourlyActivity[6] > 0,
		},
		{
			ID:          "night_owl",
			Name:        "Night Owl",
			Description: "Code after midnight",
			Icon:        "🦉",
			Unlocked:    stats.HourlyActivity[0] > 0 || stats.HourlyActivity[1] > 0 || stats.HourlyActivity[2] > 0,
		},
	}

	return achievements
}
