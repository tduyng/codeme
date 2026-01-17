package stats

// Stats represents comprehensive coding statistics
type Stats struct {
	// Time statistics
	TotalTime     int64 `json:"total_time"`
	TodayTime     int64 `json:"today_time"`
	YesterdayTime int64 `json:"yesterday_time"`
	WeekTime      int64 `json:"week_time"`
	LastWeekTime  int64 `json:"last_week_time"`
	MonthTime     int64 `json:"month_time"`

	// Line statistics
	TotalLines     int `json:"total_lines"`
	TodayLines     int `json:"today_lines"`
	YesterdayLines int `json:"yesterday_lines"`
	WeekLines      int `json:"week_lines"`
	LastWeekLines  int `json:"last_week_lines"`
	MonthLines     int `json:"month_lines"`

	// File statistics
	TotalFiles     int `json:"total_files"`
	TodayFiles     int `json:"today_files"`
	YesterdayFiles int `json:"yesterday_files"`
	WeekFiles      int `json:"week_files"`
	LastWeekFiles  int `json:"last_week_files"`
	MonthFiles     int `json:"month_files"`

	// Streaks
	Streak        int        `json:"streak"`
	LongestStreak int        `json:"longest_streak"`
	StreakInfo    StreakInfo `json:"streak_info"`

	// Aggregated data
	Projects             map[string]ProjectStat `json:"projects"`
	Languages            map[string]LangStat    `json:"languages"`
	ProgrammingLanguages map[string]LangStat    `json:"programming_languages"`
	DailyActivity        map[string]DailyStat   `json:"daily_activity"`
	HourlyActivity       map[int]int            `json:"hourly_activity"`
	TopFiles             []FileStat             `json:"top_files"`
	WeeklyHeatmap        []HeatmapDay           `json:"weekly_heatmap"`
	Sessions             []Session              `json:"sessions"`
	Today                DailyStat              `json:"today"`

	// Patterns and insights
	MostActiveHour    int     `json:"most_active_hour"`
	MostActiveDay     string  `json:"most_active_day"`
	MostActiveDayTime int64   `json:"most_active_day_time"`
	WeekdayPattern    []int64 `json:"weekday_pattern"`
	PeakHours         []int   `json:"peak_hours"`

	// Performance metrics
	AvgSessionLength  int64  `json:"avg_session_length"`
	FocusScore        int    `json:"focus_score"`
	ProductivityTrend string `json:"productivity_trend"`

	// Goals and achievements
	DailyGoals   DailyGoals    `json:"daily_goals"`
	Achievements []Achievement `json:"achievements"`
	Records      Records       `json:"records"`
}

type ProjectStat struct {
	Lines      int              `json:"lines"`
	Time       int64            `json:"time"`
	Files      int              `json:"files"`
	Languages  map[string]int64 `json:"languages"`
	MainLang   string           `json:"main_lang"`
	LastActive string           `json:"last_active"`
	Growth     string           `json:"growth"`
}

type LangStat struct {
	Lines       int    `json:"lines"`
	Time        int64  `json:"time"`
	Files       int    `json:"files"`
	LastUsed    string `json:"last_used"`
	HoursTotal  int    `json:"hours_total"`
	Proficiency string `json:"proficiency"`
	Growth      string `json:"growth"`
	Trending    bool   `json:"trending"`
}

type DailyStat struct {
	Time         int64     `json:"time"`
	Lines        int       `json:"lines"`
	Files        int       `json:"files"`
	Sessions     []Session `json:"sessions"`
	SessionCount int       `json:"session_count"`
}

type DailyAggregation struct {
	Time         int64
	Lines        int
	Files        map[string]bool
	SessionCount int
	Languages    map[string]bool
	Projects     map[string]bool
}

type DateRange struct {
	Start string
	End   string
}

type FileStat struct {
	Path  string `json:"path"`
	Lines int    `json:"lines"`
	Time  int64  `json:"time"`
}

type HeatmapDay struct {
	Date  string `json:"date"`
	Level int    `json:"level"`
	Lines int    `json:"lines"`
	Time  int64  `json:"time"`
}

type Session struct {
	Start      string   `json:"start"`
	End        string   `json:"end"`
	Duration   int64    `json:"duration"`
	Project    string   `json:"project"`
	Languages  []string `json:"languages"`
	Files      []string `json:"files"`
	LinesTotal int      `json:"lines_total"`
	FilesCount int      `json:"files_count"`
	BreakAfter int64    `json:"break_after"`
}

type StreakInfo struct {
	Current       int    `json:"current"`
	Longest       int    `json:"longest"`
	StartDate     string `json:"start_date"`
	BestStartDate string `json:"best_start_date"`
	BestEndDate   string `json:"best_end_date"`
	WeeklyPattern []bool `json:"weekly_pattern"`
}

type DailyGoals struct {
	TimeGoal     int64 `json:"time_goal"`
	LinesGoal    int   `json:"lines_goal"`
	FilesGoal    int   `json:"files_goal"`
	SessionsGoal int   `json:"sessions_goal"`
}

type Achievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Unlocked    bool   `json:"unlocked"`
	UnlockedAt  string `json:"unlocked_at"`
}

type AchievementConfig struct {
	ID          string
	Name        string
	Description string
	Type        string
	Threshold   int
	Icon        string
	Hours       []int
	MinSession  int
}

type Records struct {
	MostProductiveDay  RecordDay     `json:"most_productive_day"`
	LongestSession     RecordSession `json:"longest_session"`
	HighestDailyOutput RecordDay     `json:"highest_daily_output"`
	BestStreak         RecordStreak  `json:"best_streak"`
	EarliestStart      RecordTime    `json:"earliest_start"`
	LatestEnd          RecordTime    `json:"latest_end"`
	BiggestFileEdit    RecordFile    `json:"biggest_file_edit"`
}

type RecordDay struct {
	Date         string   `json:"date"`
	Time         int64    `json:"time"`
	Lines        int      `json:"lines"`
	Files        int      `json:"files"`
	SessionCount int      `json:"session_count"`
	Languages    []string `json:"languages"`
	Projects     []string `json:"projects"`
	Weekday      string   `json:"weekday"`
}

type RecordSession struct {
	Start     string   `json:"start"`
	End       string   `json:"end"`
	Duration  int64    `json:"duration"`
	Project   string   `json:"project"`
	Languages []string `json:"languages"`
	Files     []string `json:"files"`
	Lines     int      `json:"lines"`
	Breaks    int      `json:"breaks"`
}

type RecordStreak struct {
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	DayCount      int    `json:"day_count"`
	DailyAvgTime  int64  `json:"daily_avg_time"`
	DailyAvgLines int    `json:"daily_avg_lines"`
	EndReason     string `json:"end_reason"`
	TotalTime     int64  `json:"total_time"`
	TotalLines    int    `json:"total_lines"`
}

type RecordTime struct {
	Time     string `json:"time"`
	Date     string `json:"date"`
	Duration int64  `json:"duration"`
	Project  string `json:"project"`
}

type RecordFile struct {
	FilePath string `json:"file_path"`
	Lines    int    `json:"lines"`
	Date     string `json:"date"`
	Language string `json:"language"`
	Project  string `json:"project"`
}
