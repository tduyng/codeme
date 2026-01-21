// core/stats/aggregator.go - FIXED VERSION
package stats

import (
	"time"

	"github.com/tduyng/codeme/util"
)

// Aggregator calculates ONLY truly global metrics
type Aggregator struct {
	timezone *time.Location
}

func NewAggregator(timezone *time.Location) *Aggregator {
	if timezone == nil {
		timezone = time.UTC
	}
	return &Aggregator{timezone: timezone}
}

// CalculateGlobalMetrics calculates ONLY metrics that span all time periods
func (a *Aggregator) CalculateGlobalMetrics(stats *Stats, activities []Activity, allSessions []Session, allSessionsByDay map[string][]Session) {
	streakCalc := NewStreakCalculator(a.timezone)

	stats.StreakInfo = streakCalc.CalculateStreak(activities)
	stats.Achievements = CalculateAchievements(stats.AllTime, activities, stats.StreakInfo)
	stats.DailyActivity = a.calculateDailyActivity(activities, allSessionsByDay)
	stats.WeeklyHeatmap = GenerateWeeklyHeatmap(activities, 12)
	stats.Records = a.calculateRecords(activities, stats.DailyActivity, stats.StreakInfo, allSessions, allSessionsByDay)
}

// calculateDailyActivity creates historical map of date -> DailyStat
func (a *Aggregator) calculateDailyActivity(
	activities []Activity,
	sessionsByDay map[string][]Session,
) map[string]DailyStat {
	dailyMap := make(map[string]DailyStat)

	for _, activity := range activities {
		date := activity.Timestamp.In(a.timezone).Format("2006-01-02")
		stat := dailyMap[date]

		stat.Date = date
		stat.Time += int64(activity.Duration)
		stat.Lines += activity.Lines
		stat.Files++

		dailyMap[date] = stat
	}

	// 2️⃣ Recalculate session counts per day
	for date, stat := range dailyMap {
		sessions := sessionsByDay[date]

		dayStart := util.DateFrom(date, a.timezone)
		dayEnd := dayStart.AddDate(0, 0, 1)

		stat.SessionCount = 0
		for _, s := range sessions {
			if s.StartTime.Before(dayEnd) && s.EndTime.After(dayStart) {
				stat.SessionCount++
			}
		}

		dailyMap[date] = stat
	}

	return dailyMap
}

// calculateRecords finds all-time personal bests
func (a *Aggregator) calculateRecords(
	activities []Activity,
	dailyActivity map[string]DailyStat,
	streakInfo StreakInfo,
	sessions []Session,
	sessionsByDay map[string][]Session,
) Records {
	records := Records{}

	if len(activities) == 0 {
		return records
	}

	// Find most productive day
	var maxDayTime float64
	var maxDayDate string
	var maxDayLines int

	for date, stat := range dailyActivity {
		if float64(stat.Time) > maxDayTime {
			maxDayTime = float64(stat.Time)
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
			Weekday:      util.ParseWeekday(maxDayDate, a.timezone),
		}
	}

	var longestSession Session
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

	// Find highest daily output (most lines)
	var maxLines int
	var maxLinesDate string

	for date, stat := range dailyActivity {
		if stat.Lines > maxLines {
			maxLines = stat.Lines
			maxLinesDate = date
		}
	}

	if maxLinesDate != "" {
		records.HighestDailyOutput = DayRecord{
			Date:         maxLinesDate,
			Lines:        maxLines,
			Time:         float64(dailyActivity[maxLinesDate].Time),
			SessionCount: len(sessionsByDay[maxLinesDate]),
			Weekday:      util.ParseWeekday(maxLinesDate, a.timezone),
		}
	}

	// Best streak record
	if streakInfo.Longest > 0 {
		end := streakInfo.LastActivity.In(a.timezone)
		start := end.AddDate(0, 0, -(streakInfo.Longest - 1))
		totalTime := 0.0
		current := start
		for !current.After(end) {
			day := current.Format("2006-01-02")
			for _, s := range sessionsByDay[day] {
				totalTime += s.Duration
			}
			current = current.AddDate(0, 0, 1)
		}

		records.BestStreak = StreakRecord{
			DayCount:  streakInfo.Longest,
			StartDate: start.Format("2006-01-02"),
			EndDate:   end.Format("2006-01-02"),
			TotalTime: totalTime,
		}
	}

	// Find earliest start time
	var earliestHour = 24
	var earliestDate string

	for _, activity := range activities {
		hour := activity.Timestamp.In(a.timezone).Hour()
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

	// Find latest end time
	var latestHour = -1
	var latestDate string

	for _, activity := range activities {
		hour := activity.Timestamp.In(a.timezone).Hour()
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

	// Find most languages in one day
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
