package stats

import (
	"database/sql"
	"sort"
	"time"
)

func CalculateStats(db *sql.DB, todayOnly bool) (Stats, error) {
	stats := Stats{
		Projects:             make(map[string]ProjectStat),
		Languages:            make(map[string]LangStat),
		ProgrammingLanguages: make(map[string]LangStat),
		DailyActivity:        make(map[string]DailyStat),
		HourlyActivity:       make(map[int]int),
		TopFiles:             []FileStat{},
		WeeklyHeatmap:        []HeatmapDay{},
		Sessions:             []Session{},
		Achievements:         []Achievement{},
		WeekdayPattern:       make([]int64, 8),
		DailyGoals: DailyGoals{
			TimeGoal:     14400,
			LinesGoal:    500,
			FilesGoal:    5,
			SessionsGoal: 3,
		},
	}

	// Date calculations
	now := time.Now()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	// Week boundaries (Monday=1, Sunday=7)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := now.AddDate(0, 0, -(weekday - 1)).Format("2006-01-02")
	lastWeekStart := now.AddDate(0, 0, -(weekday + 6)).Format("2006-01-02")
	lastWeekEnd := now.AddDate(0, 0, -weekday).Format("2006-01-02")
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")

	var query string
	var rows *sql.Rows
	var err error

	if todayOnly {
		query = `SELECT timestamp, file, language, project, lines 
		         FROM heartbeats 
		         WHERE DATE(timestamp) = ? 
		         ORDER BY timestamp`
		rows, err = db.Query(query, today)
	} else {
		query = `SELECT timestamp, file, language, project, lines 
		         FROM heartbeats 
		         ORDER BY timestamp`
		rows, err = db.Query(query)
	}

	if err != nil {
		return stats, err
	}
	defer rows.Close()

	// Tracking variables for time periods
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

	projectLanguages := make(map[string]map[string]int64)
	projectLastActive := make(map[string]time.Time)
	langLastUsed := make(map[string]time.Time)

	// Session tracking
	var sessionStart time.Time
	var sessionDate string
	var sessionProject string
	var currentSessionTime int64
	var currentSessionLines int
	var currentSessionFiles map[string]bool
	var currentSessionLanguages map[string]bool
	var lastTime time.Time

	// Records tracking
	var longestSession Session
	var mostProductiveDay RecordDay
	var highestDailyOutput RecordDay
	var mostLanguagesDay RecordDay
	var earliestStart time.Time
	var latestEnd time.Time
	var biggestFileEdit RecordFile

	// Day-of-week activity
	dayOfWeekTime := make(map[time.Weekday]int64)

	// Daily activity aggregation
	dailyAggregation := make(map[string]DailyAggregation)

	saveSession := func() {
		if sessionStart.IsZero() || currentSessionTime <= 60 {
			return
		}

		session := Session{
			Start:      sessionStart.Format(time.RFC3339),
			End:        lastTime.Format(time.RFC3339),
			Duration:   currentSessionTime,
			Project:    sessionProject,
			LinesTotal: currentSessionLines,
			FilesCount: len(currentSessionFiles),
		}

		for lang := range currentSessionLanguages {
			session.Languages = append(session.Languages, lang)
		}
		for file := range currentSessionFiles {
			session.Files = append(session.Files, file)
		}

		stats.Sessions = append(stats.Sessions, session)

		dailyAgg := dailyAggregation[sessionDate]
		dailyAgg.SessionCount++
		dailyAggregation[sessionDate] = dailyAgg

		if currentSessionTime > longestSession.Duration {
			longestSession = session
		}
	}

	// Process heartbeats
	for rows.Next() {
		var timestamp time.Time
		var file, language, project string
		var lines int

		if err := rows.Scan(&timestamp, &file, &language, &project, &lines); err != nil {
			continue
		}

		files[file] = true
		date := timestamp.Format("2006-01-02")
		isToday := date == today
		isYesterday := date == yesterday
		isThisWeek := date >= weekStart
		isLastWeek := date >= lastWeekStart && date <= lastWeekEnd
		isThisMonth := date >= monthStart

		// Initialize daily aggregation if needed
		if _, exists := dailyAggregation[date]; !exists {
			dailyAggregation[date] = DailyAggregation{
				Files:     make(map[string]bool),
				Languages: make(map[string]bool),
				Projects:  make(map[string]bool),
			}
		}

		var sessionDelta int64
		if !lastTime.IsZero() && timestamp.Sub(lastTime).Minutes() <= SessionGapMinutes {
			// Continue existing session
			sessionDelta = int64(timestamp.Sub(lastTime).Seconds())
			totalSessionTime += sessionDelta

			currentSessionTime += sessionDelta
			currentSessionLines += lines
			if currentSessionFiles == nil {
				currentSessionFiles = make(map[string]bool)
				currentSessionLanguages = make(map[string]bool)
			}
			currentSessionFiles[file] = true
			currentSessionLanguages[language] = true

			if isToday {
				todaySessionTime += sessionDelta
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

			wd := int(timestamp.Weekday())
			if wd == 0 {
				wd = 7
			}
			stats.WeekdayPattern[wd] += sessionDelta
			dayOfWeekTime[timestamp.Weekday()] += sessionDelta

		} else {
			saveSession()

			// Start new session
			sessionStart = timestamp
			sessionDate = date
			sessionProject = project
			currentSessionTime = 0
			currentSessionLines = lines
			currentSessionFiles = make(map[string]bool)
			currentSessionLanguages = make(map[string]bool)
			currentSessionFiles[file] = true
			currentSessionLanguages[language] = true
			sessionDelta = 0 // ✅ CRITICAL FIX: Reset sessionDelta for new sessions
		}

		lastTime = timestamp

		if projectFiles[project] == nil {
			projectFiles[project] = make(map[string]bool)
			projectLanguages[project] = make(map[string]int64)
		}
		projectFiles[project][file] = true

		if sessionDelta > 0 {
			projectLanguages[project][language] += sessionDelta
		}

		// Update last active times
		if projectLastActive[project].Before(timestamp) {
			projectLastActive[project] = timestamp
		}
		if langLastUsed[language].Before(timestamp) {
			langLastUsed[language] = timestamp
		}

		// Track files per language
		if langFiles[language] == nil {
			langFiles[language] = make(map[string]bool)
		}
		langFiles[language][file] = true

		ps := stats.Projects[project]
		ps.Lines += lines
		if sessionDelta > 0 {
			ps.Time += sessionDelta
		}
		stats.Projects[project] = ps

		ls := stats.Languages[language]
		ls.Lines += lines
		if sessionDelta > 0 {
			ls.Time += sessionDelta
		}
		stats.Languages[language] = ls

		if fileStats[file] == nil {
			fileStats[file] = &FileStat{Path: file}
		}
		fileStats[file].Lines += lines
		if sessionDelta > 0 {
			fileStats[file].Time += sessionDelta
		}

		// Check for biggest file edit record
		if lines > biggestFileEdit.Lines {
			biggestFileEdit = RecordFile{
				FilePath: file,
				Lines:    lines,
				Date:     date,
				Language: language,
				Project:  project,
			}
		}

		// Daily activity aggregation
		dailyAgg := dailyAggregation[date]
		if sessionDelta > 0 {
			dailyAgg.Time += sessionDelta
		}
		dailyAgg.Lines += lines
		dailyAgg.Files[file] = true
		dailyAgg.Languages[language] = true
		dailyAgg.Projects[project] = true
		dailyAggregation[date] = dailyAgg

		// Hourly activity
		hour := timestamp.Hour()
		stats.HourlyActivity[hour]++

		if earliestStart.IsZero() || timestamp.Before(earliestStart) {
			earliestStart = timestamp
		}
		if latestEnd.IsZero() || timestamp.After(latestEnd) {
			latestEnd = timestamp
		}

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

	// Save final session
	saveSession()

	// Calculate break times between sessions
	stats.Sessions = CalculateSessionBreaks(stats.Sessions)

	for date, dayData := range dailyAggregation {
		var daySessions []Session
		for _, session := range stats.Sessions {
			sessionStart, err := time.Parse(time.RFC3339, session.Start)
			if err != nil {
				continue
			}
			sessionDate := sessionStart.Format("2006-01-02")
			if sessionDate == date {
				daySessions = append(daySessions, session)
			}
		}

		dailyStat := DailyStat{
			Time:         dayData.Time,
			Lines:        dayData.Lines,
			Files:        len(dayData.Files),
			SessionCount: dayData.SessionCount,
			Sessions:     daySessions, // ✅ NEW: Store actual sessions
		}
		stats.DailyActivity[date] = dailyStat

		// Check for records
		if dayData.Time > mostProductiveDay.Time {
			mostProductiveDay = RecordDay{
				Date:         date,
				Time:         dayData.Time,
				Lines:        dayData.Lines,
				Files:        len(dayData.Files),
				SessionCount: dayData.SessionCount,
				Languages:    MapKeysToSlice(dayData.Languages),
				Projects:     MapKeysToSlice(dayData.Projects),
				Weekday:      ParseWeekday(date),
			}
		}

		if dayData.Lines > highestDailyOutput.Lines {
			highestDailyOutput = RecordDay{
				Date:         date,
				Time:         dayData.Time,
				Lines:        dayData.Lines,
				Files:        len(dayData.Files),
				SessionCount: dayData.SessionCount,
				Languages:    MapKeysToSlice(dayData.Languages),
				Projects:     MapKeysToSlice(dayData.Projects),
				Weekday:      ParseWeekday(date),
			}
		}

		if len(dayData.Languages) > len(mostLanguagesDay.Languages) {
			mostLanguagesDay = RecordDay{
				Date:         date,
				Time:         dayData.Time,
				Lines:        dayData.Lines,
				Files:        len(dayData.Files),
				SessionCount: dayData.SessionCount,
				Languages:    MapKeysToSlice(dayData.Languages),
				Projects:     MapKeysToSlice(dayData.Projects),
				Weekday:      ParseWeekday(date),
			}
		}
	}

	// Enhanced project statistics
	for project, ps := range stats.Projects {
		ps.Files = len(projectFiles[project])
		ps.Languages = projectLanguages[project]
		ps.LastActive = projectLastActive[project].Format(time.RFC3339)

		var maxTime int64
		for lang, langTime := range projectLanguages[project] {
			if langTime > maxTime {
				maxTime = langTime
				ps.MainLang = lang
			}
		}

		ps.Growth = CalculateGrowth(project, stats.DailyActivity, weekStart, lastWeekStart, lastWeekEnd)
		stats.Projects[project] = ps
	}

	// Enhanced language statistics
	for lang, ls := range stats.Languages {
		ls.Files = len(langFiles[lang])
		ls.LastUsed = langLastUsed[lang].Format(time.RFC3339)
		ls.HoursTotal = int(ls.Time / 3600)
		ls.Proficiency = CalculateProficiency(ls.HoursTotal)
		ls.Growth = CalculateLanguageGrowth(lang, stats.DailyActivity, weekStart, lastWeekStart, lastWeekEnd)
		ls.Trending = IsTrending(ls.Growth)
		stats.Languages[lang] = ls
	}

	stats.TopFiles = getSortedTopFiles(fileStats, 10)

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

	// Calculate derived stats
	stats.Streak = CalculateStreak(stats.DailyActivity)
	stats.LongestStreak = CalculateLongestStreak(stats.DailyActivity)
	stats.StreakInfo = CalculateStreakInfo(stats.DailyActivity)
	stats.MostActiveHour = CalculateMostActiveHour(stats.HourlyActivity)
	stats.PeakHours = CalculatePeakHours(stats.HourlyActivity, 3)
	stats.AvgSessionLength = CalculateAvgSessionLength(stats.Sessions)
	stats.FocusScore = CalculateFocusScore(stats.Sessions)
	stats.ProductivityTrend = CalculateProductivityTrend(stats.DailyActivity, weekStart, lastWeekStart, lastWeekEnd)
	stats.WeeklyHeatmap = GenerateWeeklyHeatmap(stats.DailyActivity, 12)
	stats.Achievements = CalculateAchievements(stats)

	// Set Today data with session count and list
	if todayData, exists := stats.DailyActivity[today]; exists {
		stats.Today = todayData
	} else {
		stats.Today = DailyStat{
			Time:         0,
			Lines:        0,
			Files:        0,
			SessionCount: 0,
			Sessions:     []Session{},
		}
	}

	stats.MostActiveDay, stats.MostActiveDayTime = getMostActiveDayOfWeek(dayOfWeekTime)

	// Populate Records structure
	stats.Records = Records{
		MostProductiveDay:  mostProductiveDay,
		HighestDailyOutput: highestDailyOutput,
		BiggestFileEdit:    biggestFileEdit,
		LongestSession: RecordSession{
			Start:     longestSession.Start,
			End:       longestSession.End,
			Duration:  longestSession.Duration,
			Project:   longestSession.Project,
			Languages: longestSession.Languages,
			Files:     longestSession.Files,
			Lines:     longestSession.LinesTotal,
		},
		BestStreak: RecordStreak{
			StartDate:     stats.StreakInfo.BestStartDate,
			EndDate:       stats.StreakInfo.BestEndDate,
			DayCount:      stats.StreakInfo.Longest,
			DailyAvgTime:  0,
			DailyAvgLines: 0,
			EndReason:     "Historical",
		},
		EarliestStart: RecordTime{
			Time:    earliestStart.Format("15:04"),
			Date:    earliestStart.Format("2006-01-02"),
			Project: "",
		},
		LatestEnd: RecordTime{
			Time:    latestEnd.Format("15:04"),
			Date:    latestEnd.Format("2006-01-02"),
			Project: "",
		},
	}

	// Filter programming languages
	for lang, stat := range stats.Languages {
		if IsCodeLanguage(lang) {
			stats.ProgrammingLanguages[lang] = stat
		}
	}

	return stats, nil
}

func getSortedTopFiles(fileStats map[string]*FileStat, limit int) []FileStat {
	var files []FileStat
	for _, fs := range fileStats {
		files = append(files, *fs)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Time > files[j].Time
	})

	if len(files) > limit {
		files = files[:limit]
	}

	return files
}

func getMostActiveDayOfWeek(dayOfWeekTime map[time.Weekday]int64) (string, int64) {
	type weekdayTime struct {
		day  time.Weekday
		time int64
	}

	var sorted []weekdayTime
	for wd, t := range dayOfWeekTime {
		sorted = append(sorted, weekdayTime{wd, t})
	}

	if len(sorted) == 0 {
		return "", 0
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].time > sorted[j].time
	})

	return sorted[0].day.String(), sorted[0].time
}
