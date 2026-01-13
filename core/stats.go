package core

import (
	"database/sql"
	"time"
)

func CalculateStats(db *sql.DB, todayOnly bool) (Stats, error) {
	stats := Stats{
		Projects:       make(map[string]ProjectStat),
		Languages:      make(map[string]LangStat),
		DailyActivity:  make(map[string]DailyStat),
		HourlyActivity: make(map[int]int),
		TopFiles:       []FileStat{},
	}

	// When todayOnly is true, we still need total stats for display
	// So we first get total stats, then filter today's data
	query := `SELECT timestamp, file, language, project, lines FROM heartbeats ORDER BY timestamp`

	rows, err := db.Query(query)
	if err != nil {
		return stats, err
	}
	defer rows.Close()

	var lastTime time.Time
	var totalSessionTime int64
	var todaySessionTime int64
	files := make(map[string]bool)
	todayFiles := make(map[string]bool)
	fileStats := make(map[string]*FileStat)
	projectFiles := make(map[string]map[string]bool)
	langFiles := make(map[string]map[string]bool)
	today := time.Now().Format("2006-01-02")

	for rows.Next() {
		var timestamp time.Time
		var file, language, project string
		var lines int

		rows.Scan(&timestamp, &file, &language, &project, &lines)

		files[file] = true
		isToday := timestamp.Format("2006-01-02") == today

		// Session detection (15-minute gap)
		var sessionDelta int64
		if !lastTime.IsZero() && timestamp.Sub(lastTime).Minutes() <= 15 {
			sessionDelta = int64(timestamp.Sub(lastTime).Seconds())
			totalSessionTime += sessionDelta

			if isToday {
				todaySessionTime += sessionDelta
			}
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
		date := timestamp.Format("2006-01-02")
		ds := stats.DailyActivity[date]
		ds.Lines += lines
		ds.Time += sessionDelta
		ds.Files++
		stats.DailyActivity[date] = ds

		// Hourly activity
		hour := timestamp.Hour()
		stats.HourlyActivity[hour]++

		stats.TotalLines += lines
		if isToday {
			stats.TodayLines += lines
			todayFiles[file] = true
		}
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

	// Get top 10 files
	for _, fs := range fileStats {
		stats.TopFiles = append(stats.TopFiles, *fs)
	}
	if len(stats.TopFiles) > 10 {
		stats.TopFiles = stats.TopFiles[:10]
	}

	stats.TotalTime = totalSessionTime
	stats.TodayTime = todaySessionTime
	stats.TotalFiles = len(files)
	stats.Streak = calculateStreak(stats.DailyActivity)
	stats.LongestStreak = calculateLongestStreak(stats.DailyActivity)

	// Populate Today field for nvim plugin compatibility
	if todayStat, ok := stats.DailyActivity[today]; ok {
		stats.Today = todayStat
	} else {
		stats.Today = DailyStat{Lines: stats.TodayLines, Time: stats.TodayTime, Files: len(todayFiles)}
	}

	// If todayOnly is requested, filter the results to show only today's data
	if todayOnly {
		// Keep only today's projects, languages, and files
		todayProjects := make(map[string]ProjectStat)
		todayLanguages := make(map[string]LangStat)
		todayTopFiles := []FileStat{}

		// Filter projects, languages, and files to only today's data
		// We need to recalculate from today's heartbeats only
		// Use substr() to handle both RFC3339 and Go default time formats
		todayQuery := `SELECT timestamp, file, language, project, lines FROM heartbeats 
				 WHERE substr(timestamp, 1, 10) = '` + today + `' ORDER BY timestamp`

		todayRows, err := db.Query(todayQuery)
		if err != nil {
			return stats, err
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

		// Get top files (today only)
		for _, fs := range todayFileStats {
			todayTopFiles = append(todayTopFiles, *fs)
		}
		if len(todayTopFiles) > 10 {
			todayTopFiles = todayTopFiles[:10]
		}

		// Replace stats with today-only data
		stats.Projects = todayProjects
		stats.Languages = todayLanguages
		stats.TopFiles = todayTopFiles
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
	}

	return stats, nil
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
