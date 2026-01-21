// core/stats/types.go
package stats

import "time"

// Stats contains all period statistics with comparisons
type Stats struct {
	Today         PeriodStats          `json:"today"`
	Yesterday     PeriodStats          `json:"yesterday"`
	ThisWeek      PeriodStats          `json:"this_week"`
	LastWeek      PeriodStats          `json:"last_week"`
	ThisMonth     PeriodStats          `json:"this_month"`
	LastMonth     PeriodStats          `json:"last_month"`
	AllTime       PeriodStats          `json:"all_time"`
	DayOverDay    ComparisonResult     `json:"day_over_day"`
	StreakInfo    StreakInfo           `json:"streak_info"`
	Achievements  []Achievement        `json:"achievements"`
	Records       Records              `json:"records"`
	DailyActivity map[string]DailyStat `json:"daily_activity"` // Historical data
	WeeklyHeatmap []HeatmapDay         `json:"weekly_heatmap"` // Historical data
	GeneratedAt   time.Time            `json:"generated_at"`
}

// PeriodStats represents statistics for a specific time period
type PeriodStats struct {
	Period            string           `json:"period"`
	StartDate         time.Time        `json:"start_date"`
	EndDate           time.Time        `json:"end_date"`
	TotalTime         float64          `json:"total_time"`
	TotalLines        int              `json:"total_lines"`
	TotalFiles        int              `json:"total_files"`
	Languages         []LanguageStats  `json:"languages"`
	Projects          []ProjectStats   `json:"projects"`
	Editors           []EditorStats    `json:"editors"`
	Files             []FileStats      `json:"top_files"`
	Hourly            []HourlyActivity `json:"hourly_activity"`
	Daily             []DailyPattern   `json:"daily_patterns"`
	PeakHour          int              `json:"peak_hour"`
	Sessions          []Session        `json:"sessions"`
	SessionCount      int              `json:"session_count"`
	FocusScore        int              `json:"focus_score"`
	DailyGoals        DailyGoals       `json:"daily_goals,omitempty"`
	MostProductiveDay *DayRecord       `json:"most_productive_day,omitempty"`
}

// ProjectStats represents detailed project statistics
type ProjectStats struct {
	Name         string    `json:"name"`
	Time         float64   `json:"time"` // seconds
	Lines        int       `json:"lines"`
	Files        int       `json:"files"`
	MainLanguage string    `json:"main_lang"`
	Growth       string    `json:"growth"` // "↑", "↓", "→"
	LastActive   time.Time `json:"last_active"`
}

// LanguageStats represents detailed language statistics
type LanguageStats struct {
	Name        string  `json:"name"`
	Time        float64 `json:"time"` // seconds
	Lines       int     `json:"lines"`
	Files       int     `json:"files"`
	Proficiency string  `json:"proficiency"` // Calculated from total hours
	HoursTotal  float64 `json:"hours_total"` // Lifetime hours
	Trending    bool    `json:"trending"`    // Active in last 7 days
}

// EditorStats represents editor usage statistics
type EditorStats struct {
	Name string  `json:"name"`
	Time float64 `json:"time"`
}

// FileStats represents file editing statistics
type FileStats struct {
	Name       string    `json:"name"` // file path
	Time       float64   `json:"time"`
	Lines      int       `json:"lines"`
	LastEdited time.Time `json:"last_edited"`
}

// ComparisonResult shows changes between two periods
type ComparisonResult struct {
	PreviousPeriod string `json:"previous_period"`
	CurrentPeriod  string `json:"current_period"`
	Trend          string `json:"trend"` // increasing | decreasing | stable
}

// LanguageStat represents statistics for a language
type LanguageStat struct {
	Time  int64 `json:"time"`
	Lines int   `json:"lines"`
}

// ProjectStat represents statistics for a project
type ProjectStat struct {
	Time  int64 `json:"time"`
	Files int   `json:"files"`
	Lines int   `json:"lines"`
}

// FileStat represents statistics for a file
type FileStat struct {
	Path string `json:"path"`
	Time int64  `json:"time"`
}

// HourlyActivity tracks activity by hour with peak identification
type HourlyActivity struct {
	Hour       int     `json:"hour"` // 0-23
	Duration   float64 `json:"duration"`
	IsPeak     bool    `json:"is_peak"`
	Percentage float64 `json:"percentage"`
}

// DailyPattern tracks activity by day of week
type DailyPattern struct {
	DayOfWeek  string  `json:"day_of_week"`
	Duration   float64 `json:"duration"`
	IsWeekend  bool    `json:"is_weekend"`
	Percentage float64 `json:"percentage"`
}

// Activity represents a single coding activity event
type Activity struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Duration  float64   `json:"duration"` // seconds (calculated, not stored)
	Lines     int       `json:"lines"`    // lines changed
	Language  string    `json:"language"`
	Project   string    `json:"project"`
	Editor    string    `json:"editor"`
	File      string    `json:"file"`
	Branch    string    `json:"branch,omitempty"`
	IsWrite   bool      `json:"is_write"`
}

// Session represents a continuous coding session
type Session struct {
	ID         string     `json:"id"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    time.Time  `json:"end_time"`
	Duration   float64    `json:"duration"`
	Activities []Activity `json:"activities"`
	Projects   []string   `json:"projects"`
	Languages  []string   `json:"languages"`
	IsActive   bool       `json:"is_active"`
	BreakAfter float64    `json:"break_after,omitempty"` // Duration until next session
}

// StreakInfo tracks coding streaks
type StreakInfo struct {
	Current      int       `json:"current"`
	Longest      int       `json:"longest"`
	LastActivity time.Time `json:"last_activity"`
	IsActive     bool      `json:"is_active"`
}

// Records tracks all-time personal bests
type Records struct {
	MostProductiveDay  DayRecord     `json:"most_productive_day"`
	LongestSession     SessionRecord `json:"longest_session"`
	HighestDailyOutput DayRecord     `json:"highest_daily_output"`
	BestStreak         StreakRecord  `json:"best_streak"`
	EarliestStart      TimeRecord    `json:"earliest_start"`
	LatestEnd          TimeRecord    `json:"latest_end"`
	MostLanguagesDay   LanguagesDay  `json:"most_languages_day"`
}

// DayRecord represents a record day
type DayRecord struct {
	Date         string   `json:"date"`
	Time         float64  `json:"time"`
	Lines        int      `json:"lines"`
	SessionCount int      `json:"session_count"`
	Weekday      string   `json:"weekday,omitempty"`
	Languages    []string `json:"languages,omitempty"`
	Projects     []string `json:"projects,omitempty"`
}

// SessionRecord represents a record session
type SessionRecord struct {
	Date     string  `json:"date"`
	Start    string  `json:"start"`
	End      string  `json:"end"`
	Duration float64 `json:"duration"`
	Project  string  `json:"project"`
	Language string  `json:"language"`
}

// StreakRecord represents a record streak
type StreakRecord struct {
	DayCount  int     `json:"day_count"`
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	TotalTime float64 `json:"total_time"`
}

// TimeRecord represents a time record (earliest/latest)
type TimeRecord struct {
	Time string `json:"time"`
	Date string `json:"date"`
}

// LanguagesDay represents most polyglot day
type LanguagesDay struct {
	Date      string   `json:"date"`
	Languages []string `json:"languages"`
	Count     int      `json:"count"`
}

// DailyGoals tracks progress toward daily goals
type DailyGoals struct {
	TimeGoal      float64 `json:"time_goal"`      // Target seconds per day
	LinesGoal     int     `json:"lines_goal"`     // Target lines per day
	TimeProgress  float64 `json:"time_progress"`  // Current progress (%)
	LinesProgress float64 `json:"lines_progress"` // Current progress (%)
	OnTrack       bool    `json:"on_track"`       // Meeting goals?
}

// DailyStat represents statistics for a single day
type DailyStat struct {
	Date         string `json:"date"`
	Time         int64  `json:"time"` // seconds
	Lines        int    `json:"lines"`
	Files        int    `json:"files"`
	SessionCount int    `json:"session_count,omitempty"`
}

// HeatmapDay represents a single day in the heatmap
type HeatmapDay struct {
	Date  string `json:"date"`
	Level int    `json:"level"` // 0-4 activity level, -1 for future
	Lines int    `json:"lines"`
	Time  int64  `json:"time"`
}

// Achievement represents a milestone or achievement
type Achievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Unlocked    bool   `json:"unlocked"`
}
