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

	query := `SELECT timestamp, file, language, project, lines FROM heartbeats ORDER BY timestamp`
	if todayOnly {
		today := time.Now().Format("2006-01-02")
		query = `SELECT timestamp, file, language, project, lines FROM heartbeats 
				 WHERE DATE(timestamp) = '` + today + `' ORDER BY timestamp`
	}

	rows, err := db.Query(query)
	if err != nil {
		return stats, err
	}
	defer rows.Close()

	var lastTime time.Time
	var totalSessionTime int64
	var todaySessionTime int64
	files := make(map[string]bool)
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

		// Session detection (15-minute gap)
		var sessionDelta int64
		if !lastTime.IsZero() && timestamp.Sub(lastTime).Minutes() <= 15 {
			sessionDelta = int64(timestamp.Sub(lastTime).Seconds())
			totalSessionTime += sessionDelta

			date := timestamp.Format("2006-01-02")
			if date == today {
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
		if date == today {
			stats.TodayLines += lines
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
		stats.Today = DailyStat{Lines: stats.TodayLines, Time: stats.TodayTime, Files: 0}
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
