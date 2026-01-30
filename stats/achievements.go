package stats

import (
	"time"

	"github.com/tduyng/codeme/core"
)

type AchievementConfig struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Threshold   int    `json:"threshold,omitempty"`
	Icon        string `json:"icon"`
	Hours       []int  `json:"hours,omitempty"`
	MinSession  int    `json:"min_session,omitempty"`
}

var AchievementConfigs = []AchievementConfig{
	{ID: "streak_5", Name: "5-Day Fire", Description: "Code for 5 days in a row", Type: "streak", Threshold: 5, Icon: "🔥"},
	{ID: "streak_30", Name: "30-Day Streak", Description: "Code consistently for 30 days", Type: "streak", Threshold: 30, Icon: "🧨"},
	{ID: "streak_90", Name: "90-Day Inferno", Description: "Maintain a 90-day coding streak", Type: "streak", Threshold: 90, Icon: "💥"},
	{ID: "streak_180", Name: "180-Day Blaze", Description: "Code for 180 consecutive days", Type: "streak", Threshold: 180, Icon: "🌋"},
	{ID: "streak_365", Name: "365-Day Eternal Flame", Description: "Maintain a full year coding streak", Type: "streak", Threshold: 365, Icon: "🌞"},
	{ID: "lines_1000", Name: "1K Line Wave", Description: "Write 1,000 lines of code", Type: "lines", Threshold: 1000, Icon: "🌧️"},
	{ID: "lines_10000", Name: "10K Line Surge", Description: "Write 10,000 lines of code", Type: "lines", Threshold: 10000, Icon: "⚡"},
	{ID: "lines_50000", Name: "50K Line Flood", Description: "Write 50,000 lines of code", Type: "lines", Threshold: 50000, Icon: "⛈️"},
	{ID: "lines_100000", Name: "100K Line Ocean", Description: "Write 100,000 lines of code", Type: "lines", Threshold: 100000, Icon: "🌊"},
	{ID: "hours_50", Name: "50h Spark", Description: "Code for 50 hours total", Type: "hours", Threshold: 180000, Icon: "⚡"},
	{ID: "hours_1000", Name: "1K h Lightning", Description: "Code for 1000 hours total", Type: "hours", Threshold: 3600000, Icon: "🌩️"},
	{ID: "hours_5000", Name: "5K h Thunder", Description: "Code for 5000 hours total", Type: "hours", Threshold: 18000000, Icon: "⛈️"},
	{ID: "hours_10000", Name: "10K h Mastery", Description: "Code for 10000 hours total", Type: "hours", Threshold: 36000000, Icon: "🌀"},
	{ID: "hours_20000", Name: "20K h Grandmaster", Description: "Code for 20000 hours total", Type: "hours", Threshold: 720000000, Icon: "💡"},
	{ID: "polyglot_2", Name: "Bilingual", Description: "Code in 2 different languages", Type: "languages", Threshold: 2, Icon: "🚀"},
	{ID: "polyglot_5", Name: "Polyglot", Description: "Code in 5 different languages", Type: "languages", Threshold: 5, Icon: "🌍"},
	{ID: "polyglot_10", Name: "Polyglot Master", Description: "Code in 10 different languages", Type: "languages", Threshold: 10, Icon: "🧠"},
	{ID: "polyglot_15", Name: "Code Polymath", Description: "Code in 15 different languages", Type: "languages", Threshold: 15, Icon: "🎓"},
	{ID: "early_bird", Name: "Dawn Coder", Description: "Code before 6 AM", Type: "habit", Hours: []int{4, 5}, Icon: "🌅"},
	{ID: "night_owl", Name: "Night Coder", Description: "Code after midnight", Type: "habit", Hours: []int{0, 1, 2}, Icon: "🌌"},
	{ID: "session_2h", Name: "2h Warm Up", Description: "Code for 2+ hours in a single session", Type: "session", MinSession: 7200, Icon: "☕"},
	{ID: "session_4h", Name: "4h Focus", Description: "Code for 4+ hours in a single session", Type: "session", MinSession: 14400, Icon: "🎯"},
	{ID: "session_6h", Name: "6h Flow State", Description: "Code for 6+ hours in a single session", Type: "session", MinSession: 21600, Icon: "🌊"},
	{ID: "session_8h", Name: "8h Deep Work", Description: "Code for 8+ hours in a single session", Type: "session", MinSession: 28800, Icon: "🧠"},
	{ID: "session_10h", Name: "10h Monk Mode", Description: "Code for 10+ hours in a single session", Type: "session", MinSession: 36000, Icon: "🧘‍♂️"},
	{ID: "session_12h", Name: "12h Legendary", Description: "Code for 12+ hours in a single session", Type: "session", MinSession: 43200, Icon: "👑"},
}

func CalculateAchievements(allTime APIPeriodStats, activities []core.Activity, streakInfo StreakInfo) []Achievement {
	var achievements []Achievement

	langCount := 0
	for _, lang := range allTime.Languages {
		if IsCodeLanguage(lang.Name) {
			langCount++
		}
	}

	hourlyActivity := AggregateByHour(activities, time.Local)

	for _, cfg := range AchievementConfigs {
		unlocked := false

		switch cfg.Type {
		case "streak":
			unlocked = streakInfo.Current >= cfg.Threshold || streakInfo.Longest >= cfg.Threshold

		case "lines":
			unlocked = allTime.TotalLines >= cfg.Threshold

		case "hours":
			unlocked = allTime.TotalTime >= float64(cfg.Threshold)

		case "languages":
			unlocked = langCount >= cfg.Threshold

		case "habit":
			for _, h := range cfg.Hours {
				if h < len(hourlyActivity) && hourlyActivity[h].Duration > 0 {
					unlocked = true
					break
				}
			}

		case "session":
			for _, s := range allTime.Sessions {
				if s.Duration >= float64(cfg.MinSession) {
					unlocked = true
					break
				}
			}
		}

		achievements = append(achievements, Achievement{
			ID:          cfg.ID,
			Name:        cfg.Name,
			Description: cfg.Description,
			Icon:        cfg.Icon,
			Unlocked:    unlocked,
		})
	}

	return achievements
}
