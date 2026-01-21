// core/stats/calculator.go
package stats

import (
	"math"
	"sort"
	"time"

	"github.com/tduyng/codeme/util"
)

// Calculator computes comprehensive statistics from activities
type Calculator struct {
	timezone *time.Location
}

func NewCalculator(timezone *time.Location) *Calculator {
	if timezone == nil {
		timezone = time.UTC
	}
	return &Calculator{timezone: timezone}
}

// Calculate generates complete aggregated statistics
func (c *Calculator) Calculate(activities []Activity) (*Stats, error) {
	now := time.Now().In(c.timezone)

	// calculate duration for activity
	activities = c.calculateDurationsFromGaps(activities)
	sessionManager := NewSessionManager(15*time.Minute, 1*time.Minute)
	allSessions := sessionManager.GroupActivitiesIntoSessions(activities)
	allSessionsByDay := make(map[string][]Session)
	for _, s := range allSessions {
		day := s.StartTime.In(c.timezone).Format("2006-01-02")
		allSessionsByDay[day] = append(allSessionsByDay[day], s)
	}

	stats := &Stats{
		GeneratedAt: now,
	}

	// Filter activities by period
	todayActivities := c.filterByPeriod(activities, now, "today")
	yesterdayActivities := c.filterByPeriod(activities, now.AddDate(0, 0, -1), "today")
	thisWeekActivities := c.filterByPeriod(activities, now, "week")
	lastWeekActivities := c.filterByPeriod(activities, now.AddDate(0, 0, -7), "week")
	thisMonthActivities := c.filterByPeriod(activities, now, "month")
	lastMonthActivities := c.filterByPeriod(activities, now.AddDate(0, -1, 0), "month")

	// Calculate period stats (now with per-period metrics)
	stats.Today = c.calculatePeriod(todayActivities, "today", c.startOfDay(now), now, activities, sessionManager)
	stats.Yesterday = c.calculatePeriod(yesterdayActivities, "yesterday",
		c.startOfDay(now.AddDate(0, 0, -1)), c.startOfDay(now), activities, sessionManager)
	stats.ThisWeek = c.calculatePeriod(thisWeekActivities, "this_week",
		c.startOfWeek(now), now, activities, sessionManager)
	stats.LastWeek = c.calculatePeriod(lastWeekActivities, "last_week",
		c.startOfWeek(now.AddDate(0, 0, -7)), c.startOfWeek(now), activities, sessionManager)
	stats.ThisMonth = c.calculatePeriod(thisMonthActivities, "this_month",
		c.startOfMonth(now), now, activities, sessionManager)
	stats.LastMonth = c.calculatePeriod(lastMonthActivities, "last_month",
		c.startOfMonth(now.AddDate(0, -1, 0)), c.startOfMonth(now), activities, sessionManager)
	stats.AllTime = c.calculatePeriod(activities, "all_time",
		c.getEarliestDate(activities), now, activities, sessionManager)

	stats.DayOverDay = c.compareStats(stats.Yesterday, stats.Today)

	aggregator := NewAggregator(c.timezone)
	aggregator.CalculateGlobalMetrics(stats, activities, allSessions, allSessionsByDay)

	return stats, nil
}

func (c *Calculator) calculateDurationsFromGaps(activities []Activity) []Activity {
	if len(activities) == 0 {
		return activities
	}

	// Sort by timestamp to ensure correct ordering
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Timestamp.Before(activities[j].Timestamp)
	})

	const maxGap = 2 * 60.0 // 2 minutes in seconds

	// Calculate duration for each activity based on gap to next activity
	for i := 0; i < len(activities)-1; i++ {
		gap := activities[i+1].Timestamp.Sub(activities[i].Timestamp).Seconds()

		// Cap duration at 2 minutes (assume continuous work)
		if gap > maxGap {
			activities[i].Duration = maxGap
		} else {
			activities[i].Duration = gap
		}
	}

	// Last activity: assume 2 minutes of work after final heartbeat
	if len(activities) > 0 {
		activities[len(activities)-1].Duration = maxGap
	}

	return activities
}

// calculatePeriod computes statistics for a specific period
func (c *Calculator) calculatePeriod(
	activities []Activity,
	period string,
	start, end time.Time,
	allActivities []Activity,
	sessionManager *SessionManager,
) PeriodStats {
	periodSessions := sessionManager.GroupActivitiesIntoSessions(activities)

	sessionsByDay := make(map[string][]Session)
	for _, s := range periodSessions {
		day := s.StartTime.In(c.timezone).Format("2006-01-02")
		sessionsByDay[day] = append(sessionsByDay[day], s)
	}

	stats := PeriodStats{
		Period:       period,
		StartDate:    start,
		EndDate:      end,
		TotalTime:    c.calculateTotalTime(activities),
		TotalLines:   c.calculateTotalLines(activities),
		TotalFiles:   c.calculateTotalFiles(activities),
		Languages:    c.calculateLanguageStats(activities, allActivities),
		Projects:     c.calculateProjectStats(activities, allActivities),
		Editors:      c.calculateEditorStats(activities),
		Files:        c.calculateFileStats(activities, 10),
		Hourly:       c.calculateHourlyActivity(activities),
		Daily:        c.calculateDailyPattern(activities),
		PeakHour:     c.findPeakHour(c.calculateHourlyActivity(activities)),
		Sessions:     periodSessions,
		SessionCount: len(periodSessions),
		FocusScore:   CalculateFocusScore(periodSessions),
	}

	if period == "today" {
		stats.DailyGoals = c.calculateDailyGoals(stats, 4*3600, 500)
	} else {
		daily := make(map[string]*DayRecord)
		bestDay := (*DayRecord)(nil)

		langSets := make(map[string]util.StringSet)
		projSets := make(map[string]util.StringSet)

		for _, a := range activities {
			date := a.Timestamp.In(c.timezone).Format("2006-01-02")

			day, ok := daily[date]
			if !ok {
				day = &DayRecord{
					Date:         date,
					Weekday:      util.ParseWeekday(date, c.timezone),
					SessionCount: len(sessionsByDay[date]),
				}
				daily[date] = day

				// Init sets
				langSets[date] = util.NewStringSet()
				projSets[date] = util.NewStringSet()
			}

			day.Time += a.Duration
			day.Lines += a.Lines

			if IsValidLanguage(a.Language) {
				langSets[date].Add(a.Language)
			}
			if a.Project != "" {
				projSets[date].Add(a.Project)
			}

			// Track best day
			if bestDay == nil || day.Time > bestDay.Time {
				bestDay = day
			}
		}

		// Assign sets to slices in DayRecord
		for date, day := range daily {
			day.Languages = langSets[date].ToSortedSlice()
			day.Projects = projSets[date].ToSortedSlice()
		}

		stats.MostProductiveDay = bestDay

	}

	return stats
}

// calculateLanguageStats creates rich language stats
func (c *Calculator) calculateLanguageStats(
	periodActivities []Activity,
	allActivities []Activity,
) []LanguageStats {
	periodData := make(map[string]*LanguageStats)
	total := 0.0

	for _, a := range periodActivities {
		if !IsValidLanguage(a.Language) {
			continue
		}

		lang := NormalizeLanguage(a.Language)
		if periodData[lang] == nil {
			periodData[lang] = &LanguageStats{Name: lang}
		}
		periodData[lang].Time += a.Duration
		periodData[lang].Lines += a.Lines
		periodData[lang].Files++
		total += a.Duration

	}

	// Calculate lifetime hours for proficiency
	lifetimeHours := make(map[string]float64)
	for _, a := range allActivities {
		lifetimeHours[a.Language] += a.Duration / 3600
	}

	// Check trending (active in last 7 days)
	sevenDaysAgo := time.Now().In(c.timezone).AddDate(0, 0, -7)
	recentLangs := make(map[string]bool)
	for _, a := range allActivities {
		if a.Timestamp.After(sevenDaysAgo) {
			recentLangs[a.Language] = true
		}
	}

	stats := make([]LanguageStats, 0, len(periodData))
	for name, data := range periodData {
		hours := lifetimeHours[name]
		proficiency := CalculateProficiency(hours)

		stats = append(stats, LanguageStats{
			Name:        name,
			Time:        data.Time,
			Lines:       data.Lines,
			Files:       data.Files,
			Proficiency: proficiency,
			HoursTotal:  hours,
			Trending:    recentLangs[name],
		})
	}

	// Sort by time descending
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Time > stats[j].Time
	})

	return stats
}

// calculateProjectStats creates rich project stats
func (c *Calculator) calculateProjectStats(
	periodActivities []Activity,
	allActivities []Activity,
) []ProjectStats {
	// Aggregate period data
	periodData := make(map[string]*ProjectStats)
	total := 0.0

	for _, a := range periodActivities {
		if periodData[a.Project] == nil {
			periodData[a.Project] = &ProjectStats{
				Name:       a.Project,
				LastActive: a.Timestamp,
			}
		}
		periodData[a.Project].Time += a.Duration
		periodData[a.Project].Lines += a.Lines
		periodData[a.Project].Files++
		total += a.Duration

		if a.Timestamp.After(periodData[a.Project].LastActive) {
			periodData[a.Project].LastActive = a.Timestamp
		}
	}

	// Calculate main language per project (from all activities)
	projectLangs := make(map[string]map[string]float64)
	for _, a := range allActivities {
		if projectLangs[a.Project] == nil {
			projectLangs[a.Project] = make(map[string]float64)
		}
		projectLangs[a.Project][a.Language] += a.Duration
	}

	// Build stats
	stats := make([]ProjectStats, 0, len(periodData))
	for name, data := range periodData {
		// Find main language
		mainLang := "Mixed"
		maxTime := 0.0
		for lang, time := range projectLangs[name] {
			if time > maxTime {
				maxTime = time
				mainLang = lang
			}
		}

		// Default growth indicator (will be updated in comparison)
		growth := "→"

		stats = append(stats, ProjectStats{
			Name:         name,
			Time:         data.Time,
			Lines:        data.Lines,
			Files:        data.Files,
			MainLanguage: mainLang,
			Growth:       growth,
			LastActive:   data.LastActive,
		})
	}

	// Sort by time descending
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Time > stats[j].Time
	})

	return stats
}

// calculateEditorStats creates editor stats
func (c *Calculator) calculateEditorStats(activities []Activity) []EditorStats {
	counts := make(map[string]*EditorStats)
	total := 0.0

	for _, a := range activities {
		if counts[a.Editor] == nil {
			counts[a.Editor] = &EditorStats{Name: a.Editor}
		}
		counts[a.Editor].Time += a.Duration
		total += a.Duration
	}

	stats := make([]EditorStats, 0, len(counts))
	for name, data := range counts {
		stats = append(stats, EditorStats{
			Name: name,
			Time: data.Time,
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Time > stats[j].Time
	})

	return stats
}

// calculateFileStats creates file stats
func (c *Calculator) calculateFileStats(activities []Activity, limit int) []FileStats {
	counts := make(map[string]*FileStats)
	total := 0.0

	for _, a := range activities {
		if a.File == "" {
			continue
		}
		if counts[a.File] == nil {
			counts[a.File] = &FileStats{
				Name:       a.File,
				LastEdited: a.Timestamp,
			}
		}
		counts[a.File].Time += a.Duration
		counts[a.File].Lines += a.Lines
		total += a.Duration

		if a.Timestamp.After(counts[a.File].LastEdited) {
			counts[a.File].LastEdited = a.Timestamp
		}
	}

	stats := make([]FileStats, 0, len(counts))
	for name, data := range counts {
		stats = append(stats, FileStats{
			Name:       name,
			Time:       data.Time,
			Lines:      data.Lines,
			LastEdited: data.LastEdited,
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Time > stats[j].Time
	})

	if len(stats) > limit {
		stats = stats[:limit]
	}

	return stats
}

// compareStats generates comparison between two periods
func (c *Calculator) compareStats(prev, curr PeriodStats) ComparisonResult {
	var direction string

	switch {
	case prev.TotalTime == 0 && curr.TotalTime == 0:
		direction = "stable"

	case prev.TotalTime == 0 && curr.TotalTime > 0:
		direction = "increasing" // new activity

	case prev.TotalTime > 0 && curr.TotalTime == 0:
		direction = "decreasing" // stopped activity

	default:
		change := curr.TotalTime - prev.TotalTime
		changeRate := (change / prev.TotalTime) * 100

		if math.Abs(changeRate) <= 5 {
			direction = "stable"
		} else if change > 0 {
			direction = "increasing"
		} else {
			direction = "decreasing"
		}
	}

	return ComparisonResult{
		PreviousPeriod: prev.Period,
		CurrentPeriod:  curr.Period,
		Trend:          direction,
	}
}

func (c *Calculator) calculateHourlyActivity(activities []Activity) []HourlyActivity {
	hourly := make(map[int]float64)
	total := 0.0

	for _, a := range activities {
		hour := a.Timestamp.In(c.timezone).Hour()
		hourly[hour] += a.Duration
		total += a.Duration
	}

	result := make([]HourlyActivity, 24)
	maxDuration := 0.0

	for i := range 24 {
		duration := hourly[i]
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

func (c *Calculator) calculateDailyPattern(activities []Activity) []DailyPattern {
	daily := make(map[time.Weekday]float64)
	total := 0.0

	for _, a := range activities {
		day := a.Timestamp.In(c.timezone).Weekday()
		daily[day] += a.Duration
		total += a.Duration
	}

	result := make([]DailyPattern, 7)
	days := []time.Weekday{
		time.Sunday, time.Monday, time.Tuesday, time.Wednesday,
		time.Thursday, time.Friday, time.Saturday,
	}

	for i, day := range days {
		duration := daily[day]
		percentage := 0.0
		if total > 0 {
			percentage = (duration / total) * 100
		}
		result[i] = DailyPattern{
			DayOfWeek:  day.String(),
			Duration:   duration,
			IsWeekend:  day == time.Saturday || day == time.Sunday,
			Percentage: percentage,
		}
	}

	return result
}

func (c *Calculator) calculateDailyGoals(period PeriodStats, timeGoal float64, linesGoal int) DailyGoals {
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

func (c *Calculator) filterByPeriod(activities []Activity, anchor time.Time, period string) []Activity {
	var start, end time.Time

	switch period {
	case "today":
		start = c.startOfDay(anchor)
		end = start.AddDate(0, 0, 1)
	case "week":
		start = c.startOfWeek(anchor)
		end = start.AddDate(0, 0, 7)
	case "month":
		start = c.startOfMonth(anchor)
		end = start.AddDate(0, 1, 0)
	default:
		return activities
	}

	filtered := make([]Activity, 0)
	for _, a := range activities {
		if !a.Timestamp.Before(start) && a.Timestamp.Before(end) {
			filtered = append(filtered, a)
		}
	}

	return filtered
}

func (c *Calculator) startOfDay(t time.Time) time.Time {
	y, m, d := t.In(c.timezone).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, c.timezone)
}

func (c *Calculator) startOfWeek(t time.Time) time.Time {
	start := c.startOfDay(t)
	weekday := int(start.Weekday())
	return start.AddDate(0, 0, -weekday)
}

func (c *Calculator) startOfMonth(t time.Time) time.Time {
	y, m, _ := t.In(c.timezone).Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, c.timezone)
}

func (c *Calculator) calculateTotalTime(activities []Activity) float64 {
	total := 0.0
	for _, a := range activities {
		total += a.Duration
	}
	return total
}

func (c *Calculator) calculateTotalLines(activities []Activity) int {
	total := 0
	for _, a := range activities {
		total += a.Lines
	}
	return total
}

func (c *Calculator) calculateTotalFiles(activities []Activity) int {
	files := util.NewStringSet()

	for _, a := range activities {
		if a.File != "" {
			files.Add(a.File)
		}
	}

	return files.Len()
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

func (c *Calculator) getEarliestDate(activities []Activity) time.Time {
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
