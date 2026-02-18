package stats

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/tduyng/codeme/core"
	"github.com/tduyng/codeme/util"
)

type Calculator struct {
	timezone *time.Location
}

func NewCalculator(timezone *time.Location) *Calculator {
	if timezone == nil {
		timezone = time.UTC
	}
	return &Calculator{timezone: timezone}
}

func (c *Calculator) CalculateAPI(storage core.Storage, opts APIOptions) (*APIStats, error) {
	startTime := time.Now()

	if opts.LoadRecentDays == 0 {
		opts.LoadRecentDays = 365
	}

	cutoff := time.Now().AddDate(0, 0, -opts.LoadRecentDays)
	activities, err := storage.GetActivitiesSince(cutoff)
	if err != nil {
		return nil, err
	}

	totalCount, _ := storage.GetActivityCount()

	activities = c.calculateDurations(activities)

	sessionMgr := NewSessionManager(15*time.Minute, 1*time.Minute)
	sessions := sessionMgr.GroupSessions(activities)
	sessionsByDay := c.indexSessionsByDay(sessions)

	now := time.Now().In(c.timezone)

	todayActivities := c.filterPeriod(activities, now, "today")
	yesterdayActivities := c.filterPeriod(activities, now.AddDate(0, 0, -1), "today")
	thisWeekActivities := c.filterPeriod(activities, now, "week")
	lastWeekActivities := c.filterPeriod(activities, now.AddDate(0, 0, -7), "week")
	thisMonthActivities := c.filterPeriod(activities, now, "month")
	lastMonthActivities := c.filterPeriod(activities, now.AddDate(0, -1, 0), "month")

	lifetimeHours := c.calculateLifetimeHours(activities)
	projectLangs := c.calculateProjectLanguages(activities)

	today := c.buildPeriod("today", todayActivities, sessions, sessionsByDay, lifetimeHours, projectLangs, util.StartOfDay(now, c.timezone), now)
	yesterday := c.buildPeriod("yesterday", yesterdayActivities, sessions, sessionsByDay, lifetimeHours, projectLangs, util.StartOfDay(now.AddDate(0, 0, -1), c.timezone), util.StartOfDay(now, c.timezone))
	thisWeek := c.buildPeriod("this_week", thisWeekActivities, sessions, sessionsByDay, lifetimeHours, projectLangs, util.StartOfWeek(now, c.timezone), now)
	lastWeek := c.buildPeriod("last_week", lastWeekActivities, sessions, sessionsByDay, lifetimeHours, projectLangs, util.StartOfWeek(now.AddDate(0, 0, -7), c.timezone), util.StartOfWeek(now, c.timezone))
	thisMonth := c.buildPeriod("this_month", thisMonthActivities, sessions, sessionsByDay, lifetimeHours, projectLangs, util.StartOfMonth(now, c.timezone), now)
	lastMonth := c.buildPeriod("last_month", lastMonthActivities, sessions, sessionsByDay, lifetimeHours, projectLangs, util.StartOfMonth(now.AddDate(0, -1, 0), c.timezone), util.StartOfMonth(now, c.timezone))
	allTime := c.buildPeriod("all_time", activities, sessions, sessionsByDay, lifetimeHours, projectLangs, c.getEarliestDate(activities), now)

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

	return &APIStats{
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
	}, nil
}

func (c *Calculator) calculateDurations(activities []core.Activity) []core.Activity {
	if len(activities) == 0 {
		return activities
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Timestamp.Before(activities[j].Timestamp)
	})

	const maxGap = 2 * 60.0

	for i := 0; i < len(activities)-1; i++ {
		gap := activities[i+1].Timestamp.Sub(activities[i].Timestamp).Seconds()
		if gap > maxGap {
			activities[i].Duration = maxGap
		} else {
			activities[i].Duration = gap
		}
	}

	if len(activities) > 0 {
		activities[len(activities)-1].Duration = maxGap
	}

	return activities
}

func (c *Calculator) indexSessionsByDay(sessions []core.Session) map[string][]core.Session {
	index := make(map[string][]core.Session)
	for _, s := range sessions {
		day := util.DateString(s.StartTime, c.timezone)
		index[day] = append(index[day], s)
	}
	return index
}

func (c *Calculator) filterPeriod(activities []core.Activity, anchor time.Time, period string) []core.Activity {
	var start, end time.Time

	switch period {
	case "today":
		start = util.StartOfDay(anchor, c.timezone)
		end = start.AddDate(0, 0, 1)
	case "week":
		start = util.StartOfWeek(anchor, c.timezone)
		end = start.AddDate(0, 0, 7)
	case "month":
		start = util.StartOfMonth(anchor, c.timezone)
		end = start.AddDate(0, 1, 0)
	default:
		return activities
	}

	filtered := make([]core.Activity, 0)
	for _, a := range activities {
		if !a.Timestamp.Before(start) && a.Timestamp.Before(end) {
			filtered = append(filtered, a)
		}
	}

	return filtered
}

func (c *Calculator) calculateLifetimeHours(activities []core.Activity) map[string]float64 {
	hours := make(map[string]float64)
	for _, a := range activities {
		hours[a.Language] += a.Duration / 3600
	}
	return hours
}

func (c *Calculator) calculateProjectLanguages(activities []core.Activity) map[string]map[string]float64 {
	projectLangs := make(map[string]map[string]float64)
	for _, a := range activities {
		if projectLangs[a.Project] == nil {
			projectLangs[a.Project] = make(map[string]float64)
		}
		projectLangs[a.Project][a.Language] += a.Duration
	}
	return projectLangs
}

func (c *Calculator) getEarliestDate(activities []core.Activity) time.Time {
	if len(activities) == 0 {
		return time.Now()
	}
	earliest := activities[0].Timestamp
	for _, a := range activities {
		if a.Timestamp.Before(earliest) {
			earliest = a.Timestamp
		}
	}
	return earliest
}

func (c *Calculator) buildPeriod(
	period string,
	activities []core.Activity,
	allSessions []core.Session,
	sessionsByDay map[string][]core.Session,
	lifetimeHours map[string]float64,
	projectLangs map[string]map[string]float64,
	start, end time.Time,
) APIPeriodStats {
	periodSessions := c.filterSessions(allSessions, start, end)

	totalTime := 0.0
	totalLines := 0
	files := util.NewStringSet()

	for _, a := range activities {
		totalTime += a.Duration
		totalLines += a.Lines
		if a.File != "" {
			files.Add(a.File)
		}
	}

	langAgg := AggregateByLanguage(activities)
	projAgg := AggregateByProject(activities)
	editorAgg := AggregateByEditor(activities)
	fileAgg := AggregateByFile(activities)
	hourAgg := AggregateByHour(activities, c.timezone)

	languages := TopLanguages(langAgg, lifetimeHours, totalTime, 10)
	projects := TopProjects(projAgg, projectLangs, totalTime, 10)
	editors := TopEditors(editorAgg, totalTime, 10)
	topFiles := TopFiles(fileAgg, totalTime, 10)

	hourlyActivity := c.buildHourlyActivity(hourAgg, totalTime)
	peakHour := c.findPeakHour(hourlyActivity)

	apiSessions := ConvertSessionsToAPI(periodSessions)
	focusScore := c.calculateFocusScore(periodSessions)

	result := APIPeriodStats{
		Period:         period,
		StartDate:      start,
		EndDate:        end,
		TotalTime:      totalTime,
		TotalLines:     totalLines,
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
		dayAgg := AggregateByDay(activities, c.timezone)

		var bestTimeDay *DayRecord
		var bestLinesDay *DayRecord

		for date, day := range dayAgg {
			sessCount := 0
			if sessions, ok := sessionsByDay[date]; ok {
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

			// Most productive by time
			if bestTimeDay == nil || day.Time > bestTimeDay.Time {
				bestTimeDay = record
			}

			// Highest output by lines
			if bestLinesDay == nil || day.Lines > bestLinesDay.Lines {
				bestLinesDay = record
			}
		}

		result.MostProductiveDay = bestTimeDay
		result.HighestDailyOutput = bestLinesDay

	}

	return result
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

	for date, stat := range dayAgg {
		if stat.Time > maxDayTime {
			maxDayTime = stat.Time
			maxDayDate = date
			maxDayLines = stat.Lines
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

	var maxLines int
	var maxLinesDate string

	for date, stat := range dayAgg {
		if stat.Lines > maxLines {
			maxLines = stat.Lines
			maxLinesDate = date
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

	var earliestHour = 24
	var earliestDate string

	for _, activity := range activities {
		hour := activity.Timestamp.In(c.timezone).Hour()
		if hour < earliestHour {
			earliestHour = hour
			earliestDate = activity.Timestamp.Format("2006-01-02")
		}
	}

	if earliestDate != "" {
		records.EarliestStart = TimeRecord{
			Time: time.Date(0, 1, 1, earliestHour, 0, 0, 0, time.UTC).Format("15:04"),
			Date: earliestDate,
		}
	}

	var latestHour = -1
	var latestDate string

	for _, activity := range activities {
		hour := activity.Timestamp.In(c.timezone).Hour()
		if hour > latestHour {
			latestHour = hour
			latestDate = activity.Timestamp.Format("2006-01-02")
		}
	}

	if latestDate != "" {
		records.LatestEnd = TimeRecord{
			Time: time.Date(0, 1, 1, latestHour, 0, 0, 0, time.UTC).Format("15:04"),
			Date: latestDate,
		}
	}

	languagesByDay := make(map[string]map[string]bool)

	for _, activity := range activities {
		date := activity.Timestamp.Format("2006-01-02")
		if languagesByDay[date] == nil {
			languagesByDay[date] = make(map[string]bool)
		}
		languagesByDay[date][activity.Language] = true
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
