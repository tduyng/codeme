package stats

import (
	"fmt"
	"math"
	"time"

	"github.com/tduyng/codeme/core"
	"github.com/tduyng/codeme/util"
)

type Calculator struct {
	timezone *time.Location
	cache    *StatsCache
}

func NewCalculator(timezone *time.Location) *Calculator {
	if timezone == nil {
		timezone = time.UTC
	}
	return &Calculator{
		timezone: timezone,
		cache:    NewStatsCache(30 * time.Second),
	}
}

func (c *Calculator) CalculateAPI(storage core.Storage, opts APIOptions) (*APIStats, error) {
	startTime := time.Now()

	if opts.LoadRecentDays == 0 {
		opts.LoadRecentDays = 365
	}

	if cached, ok := c.cache.Get(opts); ok {
		cached.GeneratedAt = time.Now()
		cached.Meta.QueryTimeMs = float64(time.Since(startTime).Milliseconds())
		return cached, nil
	}

	now := time.Now().In(c.timezone)

	todayStart := util.StartOfDay(now, c.timezone)
	yesterdayStart := util.StartOfDay(now.AddDate(0, 0, -1), c.timezone)
	thisWeekStart := util.StartOfWeek(now, c.timezone)
	lastWeekStart := util.StartOfWeek(now.AddDate(0, 0, -7), c.timezone)
	thisMonthStart := util.StartOfMonth(now, c.timezone)
	lastMonthStart := util.StartOfMonth(now.AddDate(0, -1, 0), c.timezone)

	todaySummary, _ := storage.GetPeriodSummary(todayStart, now)
	yesterdaySummary, _ := storage.GetPeriodSummary(yesterdayStart, todayStart)
	thisWeekSummary, _ := storage.GetPeriodSummary(thisWeekStart, now)
	lastWeekSummary, _ := storage.GetPeriodSummary(lastWeekStart, thisWeekStart)
	thisMonthSummary, _ := storage.GetPeriodSummary(thisMonthStart, now)
	lastMonthSummary, _ := storage.GetPeriodSummary(lastMonthStart, thisMonthStart)
	allTimeSummary, _ := storage.GetPeriodSummary(time.Time{}, now)

	todayLangs, _ := storage.GetLanguageSummary(todayStart, now)
	yesterdayLangs, _ := storage.GetLanguageSummary(yesterdayStart, todayStart)
	thisWeekLangs, _ := storage.GetLanguageSummary(thisWeekStart, now)
	lastWeekLangs, _ := storage.GetLanguageSummary(lastWeekStart, thisWeekStart)
	thisMonthLangs, _ := storage.GetLanguageSummary(thisMonthStart, now)
	lastMonthLangs, _ := storage.GetLanguageSummary(lastMonthStart, thisMonthStart)
	allTimeLangs, _ := storage.GetLanguageSummary(time.Time{}, now)

	todayProjs, _ := storage.GetProjectSummary(todayStart, now)
	yesterdayProjs, _ := storage.GetProjectSummary(yesterdayStart, todayStart)
	thisWeekProjs, _ := storage.GetProjectSummary(thisWeekStart, now)
	lastWeekProjs, _ := storage.GetProjectSummary(lastWeekStart, thisWeekStart)
	thisMonthProjs, _ := storage.GetProjectSummary(thisMonthStart, now)
	lastMonthProjs, _ := storage.GetProjectSummary(lastMonthStart, thisMonthStart)
	allTimeProjs, _ := storage.GetProjectSummary(time.Time{}, now)

	todayEditors, _ := storage.GetEditorSummary(todayStart, now)
	yesterdayEditors, _ := storage.GetEditorSummary(yesterdayStart, todayStart)
	thisWeekEditors, _ := storage.GetEditorSummary(thisWeekStart, now)
	lastWeekEditors, _ := storage.GetEditorSummary(lastWeekStart, thisWeekStart)
	thisMonthEditors, _ := storage.GetEditorSummary(thisMonthStart, now)
	lastMonthEditors, _ := storage.GetEditorSummary(lastMonthStart, thisMonthStart)
	allTimeEditors, _ := storage.GetEditorSummary(time.Time{}, now)

	lifetimeHours := make(map[string]float64)
	for _, lr := range allTimeLangs {
		lifetimeHours[lr.Language] = lr.TotalTime / 3600
	}

	projectLangs := make(map[string]map[string]float64)
	for _, pr := range allTimeProjs {
		pl := make(map[string]float64)
		for _, lr := range allTimeLangs {
			pl[lr.Language] = lr.TotalTime / 3600
		}
		projectLangs[pr.Project] = pl
	}

	cutoff := time.Now().AddDate(0, 0, -opts.LoadRecentDays)
	activities, err := storage.GetActivitiesSince(cutoff)
	if err != nil {
		return nil, err
	}

	totalCount, _ := storage.GetActivityCount()

	sessionMgr := NewSessionManager(0, 0)
	activities, sessions := sessionMgr.GroupAndCalculate(activities)
	sessionsByDay := c.indexSessionsByDay(sessions)

	today := c.buildPeriodFromSummary("today", todaySummary, todayLangs, todayProjs, todayEditors, sessions, sessionsByDay, lifetimeHours, projectLangs, todayStart, now, activities)
	yesterday := c.buildPeriodFromSummary("yesterday", yesterdaySummary, yesterdayLangs, yesterdayProjs, yesterdayEditors, sessions, sessionsByDay, lifetimeHours, projectLangs, yesterdayStart, todayStart, activities)
	thisWeek := c.buildPeriodFromSummary("this_week", thisWeekSummary, thisWeekLangs, thisWeekProjs, thisWeekEditors, sessions, sessionsByDay, lifetimeHours, projectLangs, thisWeekStart, now, activities)
	lastWeek := c.buildPeriodFromSummary("last_week", lastWeekSummary, lastWeekLangs, lastWeekProjs, lastWeekEditors, sessions, sessionsByDay, lifetimeHours, projectLangs, lastWeekStart, thisWeekStart, activities)
	thisMonth := c.buildPeriodFromSummary("this_month", thisMonthSummary, thisMonthLangs, thisMonthProjs, thisMonthEditors, sessions, sessionsByDay, lifetimeHours, projectLangs, thisMonthStart, now, activities)
	lastMonth := c.buildPeriodFromSummary("last_month", lastMonthSummary, lastMonthLangs, lastMonthProjs, lastMonthEditors, sessions, sessionsByDay, lifetimeHours, projectLangs, lastMonthStart, thisMonthStart, activities)
	allTime := c.buildPeriodFromSummary("all_time", allTimeSummary, allTimeLangs, allTimeProjs, allTimeEditors, sessions, sessionsByDay, lifetimeHours, projectLangs, time.Time{}, now, activities)

	streakCalc := NewStreakCalculator(c.timezone)
	streakInfo := streakCalc.Calculate(activities)

	achievements := CalculateAchievements(allTime, activities, streakInfo)

	dayAgg := AggregateByDay(activities, c.timezone)
	for date, day := range dayAgg {
		if sessionsForDay, ok := sessionsByDay[date]; ok {
			day.SessionCount = len(sessionsForDay)
			dayAgg[date] = day
		}
	}

	dailyActivity := make(map[string]DailyStat)
	for date, agg := range dayAgg {
		dailyActivity[date] = DailyStat{
			Date:         agg.Date,
			Time:         int64(agg.Time),
			Lines:        agg.Lines,
			Files:        agg.Files.Len(),
			SessionCount: agg.SessionCount,
		}
	}

	heatmap := c.generateWeeklyHeatmap(dailyActivity, 12)
	records := c.calculateRecords(activities, dayAgg, sessions, sessionsByDay)

	queryTime := time.Since(startTime).Seconds() * 1000

	result := &APIStats{
		Today:         today,
		Yesterday:     yesterday,
		ThisWeek:      thisWeek,
		LastWeek:      lastWeek,
		ThisMonth:     thisMonth,
		LastMonth:     lastMonth,
		AllTime:       allTime,
		StreakInfo:    streakInfo,
		Achievements:  achievements,
		Records:       records,
		DailyActivity: dailyActivity,
		WeeklyHeatmap: heatmap,
		GeneratedAt:   time.Now(),
		Meta: APIMeta{
			LoadedActivities: len(activities),
			TotalActivities:  totalCount,
			QueryTimeMs:      queryTime,
			DataWindow:       fmt.Sprintf("last_%d_days", opts.LoadRecentDays),
		},
	}

	c.cache.Set(result)

	return result, nil
}

func (c *Calculator) buildPeriodFromSummary(
	period string,
	summary core.PeriodSummary,
	langRows []core.LanguageRow,
	projRows []core.ProjectRow,
	editorRows []core.EditorRow,
	allSessions []core.Session,
	sessionsByDay map[string][]core.Session,
	lifetimeHours map[string]float64,
	projectLangs map[string]map[string]float64,
	start, end time.Time,
	activities []core.Activity,
) APIPeriodStats {
	periodSessions := c.filterSessions(allSessions, start, end)

	var periodActivities []core.Activity
	if !start.IsZero() {
		for _, a := range activities {
			if !a.Timestamp.Before(start) && a.Timestamp.Before(end) {
				periodActivities = append(periodActivities, a)
			}
		}
	} else {
		periodActivities = activities
	}

	periodSessionsByDay := make(map[string][]core.Session)
	for date, sess := range sessionsByDay {
		if !end.IsZero() && date < end.Format("2006-01-02") {
			if start.IsZero() || date >= start.Format("2006-01-02") {
				periodSessionsByDay[date] = sess
			}
		}
	}

	files := util.NewStringSet()
	for _, a := range periodActivities {
		if a.File != "" {
			files.Add(a.File)
		}
	}

	languages := c.convertLanguageRows(langRows, lifetimeHours, summary.TotalTime)
	projects := c.convertProjectRows(projRows, projectLangs, summary.TotalTime)
	editors := c.convertEditorRows(editorRows, summary.TotalTime)

	hourAgg := AggregateByHour(periodActivities, c.timezone)
	hourlyActivity := c.buildHourlyActivity(hourAgg, summary.TotalTime)
	peakHour := c.findPeakHour(hourlyActivity)

	fileAgg := AggregateByFile(periodActivities)
	topFiles := TopFiles(fileAgg, summary.TotalTime, 10)

	apiSessions := ConvertSessionsToAPI(periodSessions)
	focusScore := c.calculateFocusScore(periodSessions)

	result := APIPeriodStats{
		Period:         period,
		StartDate:      start,
		EndDate:        end,
		TotalTime:      summary.TotalTime,
		TotalLines:     summary.TotalLines,
		TotalFiles:     files.Len(),
		Languages:      languages,
		Projects:       projects,
		Editors:        editors,
		Files:          topFiles,
		HourlyActivity: hourlyActivity,
		PeakHour:       peakHour,
		Sessions:       apiSessions,
		SessionCount:   len(periodSessions),
		FocusScore:     focusScore,
	}

	if period == "today" {
		result.DailyGoals = c.calculateDailyGoals(result, 4*3600, 500)
	} else {
		dayAgg := AggregateByDay(periodActivities, c.timezone)

		var bestTimeDay *DayRecord
		var bestLinesDay *DayRecord

		for date, day := range dayAgg {
			sessCount := 0
			if sessions, ok := periodSessionsByDay[date]; ok {
				sessCount = len(sessions)
			}

			record := &DayRecord{
				Date:         date,
				Time:         day.Time,
				Lines:        day.Lines,
				SessionCount: sessCount,
				Weekday:      util.ParseWeekday(date, c.timezone),
				Languages:    day.Languages.ToSortedSlice(),
				Projects:     day.Projects.ToSortedSlice(),
			}

			if bestTimeDay == nil || day.Time > bestTimeDay.Time {
				bestTimeDay = record
			}

			if bestLinesDay == nil || day.Lines > bestLinesDay.Lines {
				bestLinesDay = record
			}
		}

		result.MostProductiveDay = bestTimeDay
		result.HighestDailyOutput = bestLinesDay

	}

	return result
}

func (c *Calculator) convertLanguageRows(rows []core.LanguageRow, lifetimeHours map[string]float64, total float64) []APILanguageStats {
	result := make([]APILanguageStats, 0, len(rows))
	for _, r := range rows {
		hours := lifetimeHours[r.Language]
		proficiency := CalculateProficiency(hours)

		pct := 0.0
		if total > 0 {
			pct = (r.TotalTime / total) * 100
		}

		result = append(result, APILanguageStats{
			Name:         r.Language,
			Time:         r.TotalTime,
			Lines:        r.TotalLines,
			PercentTotal: pct,
			Proficiency:  proficiency,
			HoursTotal:   r.TotalTime / 3600,
			IsCode:       IsCodeLanguage(r.Language),
		})
	}
	return result
}

func (c *Calculator) convertProjectRows(rows []core.ProjectRow, projectLangs map[string]map[string]float64, total float64) []APIProjectStats {
	result := make([]APIProjectStats, 0, len(rows))
	for _, r := range rows {
		mainLang := r.MainLanguage

		// If mainLang is not a code language, find the first code language from projectLangs
		if mainLang == "" || !IsCodeLanguage(mainLang) {
			if langs, ok := projectLangs[r.Project]; ok {
				for lang := range langs {
					if IsCodeLanguage(lang) {
						mainLang = lang
						break
					}
				}
			}
		}

		if mainLang == "" || !IsCodeLanguage(mainLang) {
			mainLang = "Mixed"
		}

		pct := 0.0
		if total > 0 {
			pct = (r.TotalTime / total) * 100
		}

		result = append(result, APIProjectStats{
			Name:         r.Project,
			Time:         r.TotalTime,
			Lines:        r.TotalLines,
			PercentTotal: pct,
			MainLanguage: mainLang,
		})
	}
	return result
}

func (c *Calculator) convertEditorRows(rows []core.EditorRow, total float64) []APIEditorStats {
	result := make([]APIEditorStats, 0, len(rows))
	for _, r := range rows {
		pct := 0.0
		if total > 0 {
			pct = (r.TotalTime / total) * 100
		}

		result = append(result, APIEditorStats{
			Name:         r.Editor,
			Time:         r.TotalTime,
			PercentTotal: pct,
		})
	}
	return result
}

func (c *Calculator) indexSessionsByDay(sessions []core.Session) map[string][]core.Session {
	index := make(map[string][]core.Session)
	for _, s := range sessions {
		day := util.DateString(s.StartTime, c.timezone)
		index[day] = append(index[day], s)
	}
	return index
}

func (c *Calculator) filterSessions(sessions []core.Session, start, end time.Time) []core.Session {
	filtered := make([]core.Session, 0)
	for _, s := range sessions {
		if !s.StartTime.Before(start) && s.StartTime.Before(end) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func (c *Calculator) buildHourlyActivity(hourAgg [24]HourAgg, total float64) []HourlyActivity {
	result := make([]HourlyActivity, 24)
	maxDuration := 0.0

	for i := range 24 {
		duration := hourAgg[i].Duration
		percentage := 0.0
		if total > 0 {
			percentage = (duration / total) * 100
		}
		result[i] = HourlyActivity{
			Hour:       i,
			Duration:   duration,
			Percentage: percentage,
		}
		if duration > maxDuration {
			maxDuration = duration
		}
	}

	threshold := maxDuration * 0.9
	for i := range result {
		result[i].IsPeak = result[i].Duration >= threshold && maxDuration > 0
	}

	return result
}

func (c *Calculator) findPeakHour(hourly []HourlyActivity) int {
	maxHour := 0
	maxDuration := 0.0
	for _, h := range hourly {
		if h.Duration > maxDuration {
			maxDuration = h.Duration
			maxHour = h.Hour
		}
	}
	return maxHour
}

func (c *Calculator) calculateFocusScore(sessions []core.Session) int {
	if len(sessions) == 0 {
		return 0
	}

	totalDuration := 0.0
	sessionCount := len(sessions)

	for _, session := range sessions {
		totalDuration += session.Duration
	}

	avgSession := totalDuration / float64(sessionCount)

	var baseScore int
	if avgSession >= 7200 {
		baseScore = 90
	} else if avgSession >= 5400 {
		baseScore = 80
	} else if avgSession >= 3600 {
		baseScore = 70
	} else if avgSession >= 2700 {
		baseScore = 60
	} else if avgSession >= 1800 {
		baseScore = 50
	} else if avgSession >= 900 {
		baseScore = 40
	} else {
		baseScore = 30
	}

	var sessionBonus int
	if sessionCount == 1 {
		if avgSession >= 5400 {
			sessionBonus = 10
		} else {
			sessionBonus = -5
		}
	} else if sessionCount >= 2 && sessionCount <= 5 {
		sessionBonus = 10
	} else if sessionCount > 8 {
		sessionBonus = -10
	}

	var variance float64
	for _, session := range sessions {
		diff := session.Duration - avgSession
		variance += diff * diff
	}
	variance = variance / float64(sessionCount)
	stdDev := math.Sqrt(variance)

	var consistencyBonus int
	if stdDev < avgSession*0.3 {
		consistencyBonus = 10
	} else if stdDev < avgSession*0.5 {
		consistencyBonus = 5
	} else {
		consistencyBonus = -5
	}

	breakPenalty := 0
	if sessionCount > 1 {
		longBreaks := 0
		shortBreaks := 0

		for i := 0; i < len(sessions)-1; i++ {
			breakDuration := sessions[i].BreakAfter
			if breakDuration > 7200 {
				longBreaks++
			} else if breakDuration < 900 && breakDuration > 0 {
				shortBreaks++
			}
		}

		if longBreaks > sessionCount/2 {
			breakPenalty = -10
		}
		if shortBreaks > sessionCount/2 {
			breakPenalty = -5
		}
	}

	finalScore := baseScore + sessionBonus + consistencyBonus + breakPenalty
	return int(math.Max(0, math.Min(100, float64(finalScore))))
}

func (c *Calculator) calculateDailyGoals(period APIPeriodStats, timeGoal float64, linesGoal int) DailyGoals {
	timeProgress := 0.0
	if timeGoal > 0 {
		timeProgress = (period.TotalTime / timeGoal) * 100
		if timeProgress > 100 {
			timeProgress = 100
		}
	}

	linesProgress := 0.0
	if linesGoal > 0 {
		linesProgress = (float64(period.TotalLines) / float64(linesGoal)) * 100
		if linesProgress > 100 {
			linesProgress = 100
		}
	}

	onTrack := timeProgress >= 50 || linesProgress >= 50

	return DailyGoals{
		TimeGoal:      timeGoal,
		LinesGoal:     linesGoal,
		TimeProgress:  timeProgress,
		LinesProgress: linesProgress,
		OnTrack:       onTrack,
	}
}

func (c *Calculator) generateWeeklyHeatmap(daily map[string]DailyStat, weeks int) []HeatmapDay {
	maxDuration := 0.0
	var firstActivityDate time.Time
	hasActivity := false

	for dateStr, ds := range daily {
		duration := float64(ds.Time)
		if duration > maxDuration {
			maxDuration = duration
		}

		// Track first activity date
		if ds.Time > 0 || ds.Lines > 0 {
			if d, err := time.Parse("2006-01-02", dateStr); err == nil {
				if !hasActivity || d.Before(firstActivityDate) {
					firstActivityDate = d
					hasActivity = true
				}
			}
		}
	}

	now := time.Now()

	// Find the START of the current week (Monday)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	daysFromMonday := weekday - 1 // Monday = day 1, so subtract 1

	currentWeekMonday := now.AddDate(0, 0, -daysFromMonday)
	currentWeekMonday = time.Date(
		currentWeekMonday.Year(),
		currentWeekMonday.Month(),
		currentWeekMonday.Day(),
		0, 0, 0, 0,
		currentWeekMonday.Location(),
	)

	// Calculate start date using smart range logic
	maxWeeksBack := currentWeekMonday.AddDate(0, 0, -(weeks-1)*7)

	var startDate time.Time
	if hasActivity {
		// Find Monday of the week containing first activity
		firstActivityWeekday := int(firstActivityDate.Weekday())
		if firstActivityWeekday == 0 {
			firstActivityWeekday = 7
		}
		firstActivityMonday := firstActivityDate.AddDate(0, 0, -(firstActivityWeekday - 1))
		firstActivityMonday = time.Date(
			firstActivityMonday.Year(),
			firstActivityMonday.Month(),
			firstActivityMonday.Day(),
			0, 0, 0, 0,
			firstActivityMonday.Location(),
		)

		// Use the more recent of: first activity Monday or max weeks back
		if firstActivityMonday.After(maxWeeksBack) {
			startDate = firstActivityMonday
		} else {
			startDate = maxWeeksBack
		}
	} else {
		// No activity, use max weeks back
		startDate = maxWeeksBack
	}

	// Calculate total days from start to end of current week
	totalDays := int(currentWeekMonday.Sub(startDate).Hours()/24) + 7

	heatmap := make([]HeatmapDay, totalDays)

	for i := range totalDays {
		d := startDate.AddDate(0, 0, i)
		date := d.Format("2006-01-02")
		ds := daily[date]

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

func (c *Calculator) calculateRecords(
	activities []core.Activity,
	dayAgg map[string]*DayAgg,
	sessions []core.Session,
	sessionsByDay map[string][]core.Session,
) Records {
	records := Records{}

	if len(activities) == 0 {
		return records
	}

	var maxDayTime float64
	var maxDayDate string
	var maxDayLines int
	var maxLines int
	var maxLinesDate string

	for date, stat := range dayAgg {
		if stat.Time > maxDayTime {
			maxDayTime = stat.Time
			maxDayDate = date
			maxDayLines = stat.Lines
		}
		if stat.Lines > maxLines {
			maxLines = stat.Lines
			maxLinesDate = date
		}
	}

	if maxDayDate != "" {
		records.MostProductiveDay = DayRecord{
			Date:         maxDayDate,
			Time:         maxDayTime,
			Lines:        maxDayLines,
			SessionCount: len(sessionsByDay[maxDayDate]),
			Weekday:      util.ParseWeekday(maxDayDate, c.timezone),
		}
	}

	if maxLinesDate != "" {
		records.HighestDailyOutput = DayRecord{
			Date:         maxLinesDate,
			Lines:        maxLines,
			Time:         dayAgg[maxLinesDate].Time,
			SessionCount: len(sessionsByDay[maxLinesDate]),
			Weekday:      util.ParseWeekday(maxLinesDate, c.timezone),
		}
	}

	var longestSession core.Session
	for _, session := range sessions {
		if session.Duration > longestSession.Duration {
			longestSession = session
		}
	}

	if longestSession.Duration > 0 {
		records.LongestSession = SessionRecord{
			Date:     longestSession.StartTime.Format("2006-01-02"),
			Start:    longestSession.StartTime.Format(time.RFC3339),
			End:      longestSession.EndTime.Format(time.RFC3339),
			Duration: longestSession.Duration,
		}
	}

	var earliestHour = 24
	var earliestDate string
	var latestHour = -1
	var latestDate string
	languagesByDay := make(map[string]map[string]bool)

	for _, activity := range activities {
		hour := activity.Timestamp.In(c.timezone).Hour()
		date := activity.Timestamp.Format("2006-01-02")

		if hour < earliestHour {
			earliestHour = hour
			earliestDate = date
		}
		if hour > latestHour {
			latestHour = hour
			latestDate = date
		}

		if languagesByDay[date] == nil {
			languagesByDay[date] = make(map[string]bool)
		}
		languagesByDay[date][activity.Language] = true
	}

	if earliestDate != "" {
		records.EarliestStart = TimeRecord{
			Time: time.Date(0, 1, 1, earliestHour, 0, 0, 0, time.UTC).Format("15:04"),
			Date: earliestDate,
		}
	}

	if latestDate != "" {
		records.LatestEnd = TimeRecord{
			Time: time.Date(0, 1, 1, latestHour, 0, 0, 0, time.UTC).Format("15:04"),
			Date: latestDate,
		}
	}

	var maxLangCount int
	var maxLangDate string

	for date, langs := range languagesByDay {
		if len(langs) > maxLangCount {
			maxLangCount = len(langs)
			maxLangDate = date
		}
	}

	if maxLangDate != "" {
		langList := make([]string, 0, len(languagesByDay[maxLangDate]))
		for lang := range languagesByDay[maxLangDate] {
			langList = append(langList, lang)
		}

		records.MostLanguagesDay = LanguagesDay{
			Date:      maxLangDate,
			Languages: langList,
			Count:     maxLangCount,
		}
	}

	return records
}
