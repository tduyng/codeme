package stats

// AchievementConfigs defines all available achievements
var AchievementConfigs = []AchievementConfig{
	// streaks
	{ID: "streak_5", Name: "5-Day Fire", Description: "Code for 5 days in a row", Type: "streak", Threshold: 5, Icon: "🔥"},
	{ID: "streak_30", Name: "30-Day Streak", Description: "Code consistently for 30 days", Type: "streak", Threshold: 30, Icon: "🧨"},
	{ID: "streak_90", Name: "90-Day Inferno", Description: "Maintain a 90-day coding streak", Type: "streak", Threshold: 90, Icon: "💥"},
	{ID: "streak_180", Name: "180-Day Blaze", Description: "Code for 180 consecutive days", Type: "streak", Threshold: 180, Icon: "🌋"},
	{ID: "streak_365", Name: "365-Day Eternal Flame", Description: "Maintain a full year coding streak", Type: "streak", Threshold: 365, Icon: "🕯️"},

	// lines
	{ID: "lines_1000", Name: "1K Line Wave", Description: "Write 1,000 lines of code", Type: "lines", Threshold: 1000, Icon: "🌊"},
	{ID: "lines_10000", Name: "10K Line Surge", Description: "Write 10,000 lines of code", Type: "lines", Threshold: 10000, Icon: "💦"},
	{ID: "lines_50000", Name: "50K Line Flood", Description: "Write 50,000 lines of code", Type: "lines", Threshold: 50000, Icon: "🌧️"},
	{ID: "lines_100000", Name: "100K Line Ocean", Description: "Write 100,000 lines of code", Type: "lines", Threshold: 100000, Icon: "🏝️"},

	// hours
	{ID: "hours_50", Name: "50h Spark", Description: "Code for 50 hours total", Type: "hours", Threshold: 180000, Icon: "⚡"},
	{ID: "hours_1000", Name: "1K h Lightning", Description: "Code for 1000 hours total", Type: "hours", Threshold: 3600000, Icon: "🌩️"},
	{ID: "hours_5000", Name: "5K h Thunder", Description: "Code for 5000 hours total", Type: "hours", Threshold: 18000000, Icon: "⛈️"},
	{ID: "hours_10000", Name: "10K h Storm", Description: "Code for 10000 hours total", Type: "hours", Threshold: 36000000, Icon: "🌀"},
	{ID: "hours_100000", Name: "100K h Powerhouse", Description: "Code for 100000 hours total", Type: "hours", Threshold: 360000000, Icon: "💡"},

	// Languages
	{ID: "polyglot_2", Name: "Bilingual", Description: "Code in 5 different languages", Type: "languages", Threshold: 2, Icon: "🚀"},
	{ID: "polyglot_5", Name: "Polyglot", Description: "Code in 5 different languages", Type: "languages", Threshold: 5, Icon: "🌍"},
	{ID: "polyglot_10", Name: "Polyglot Master", Description: "Code in 10 different languages", Type: "languages", Threshold: 10, Icon: "🧠"},
	{ID: "polyglot_15", Name: "Code Polymath", Description: "Code in 15 different languages", Type: "languages", Threshold: 15, Icon: "🎓"},

	// Habits
	{ID: "early_bird", Name: "Dawn Coder", Description: "Code before 6 AM", Type: "habit", Hours: []int{4, 5}, Icon: "🌅"},
	{ID: "night_owl", Name: "Night Coder", Description: "Code after midnight", Type: "habit", Hours: []int{0, 1, 2}, Icon: "🌌"},

	// Sessions
	{ID: "session_3h", Name: "3h Focus", Description: "Code for 3+ hours in a single session", Type: "session", MinSession: 10800, Icon: "👁️"},
	{ID: "session_5h", Name: "5h Zone", Description: "Code for 5+ hours in a single session", Type: "session", MinSession: 18000, Icon: "🎯"},
	{ID: "session_8h", Name: "8h Deep Zone", Description: "Code for 8+ hours in a single session", Type: "session", MinSession: 28800, Icon: "🧠"},
}

// CalculateAchievements determines which achievements are unlocked
func CalculateAchievements(stats Stats) []Achievement {
	var achievements []Achievement

	for _, cfg := range AchievementConfigs {
		unlocked := false

		switch cfg.Type {
		case "streak":
			unlocked = stats.Streak >= cfg.Threshold || stats.LongestStreak >= cfg.Threshold
		case "lines":
			unlocked = stats.TotalLines >= cfg.Threshold
		case "hours":
			unlocked = stats.TotalTime >= int64(cfg.Threshold)
		case "languages":
			unlocked = len(stats.ProgrammingLanguages) >= cfg.Threshold
		case "habit":
			for _, h := range cfg.Hours {
				if stats.HourlyActivity[h] > 0 {
					unlocked = true
					break
				}
			}
		case "session":
			for _, s := range stats.Sessions {
				if s.Duration >= int64(cfg.MinSession) {
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
