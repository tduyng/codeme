// stats/types.go
package stats

import "time"

type APIStats struct {
	Today         APIPeriodStats       `json:"today"`
	Yesterday     APIPeriodStats       `json:"yesterday"`
	ThisWeek      APIPeriodStats       `json:"this_week"`
	LastWeek      APIPeriodStats       `json:"last_week"`
	ThisMonth     APIPeriodStats       `json:"this_month"`
	LastMonth     APIPeriodStats       `json:"last_month"`
	AllTime       APIPeriodStats       `json:"all_time"`
	StreakInfo    StreakInfo           `json:"streak_info"`
	Achievements  []Achievement        `json:"achievements"`
	Records       Records              `json:"records"`
	DailyActivity map[string]DailyStat `json:"daily_activity"`
	WeeklyHeatmap []HeatmapDay         `json:"weekly_heatmap"`
	GeneratedAt   time.Time            `json:"generated_at"`
	Meta          APIMeta              `json:"_meta"`
}

type APIPeriodStats struct {
	Period             string             `json:"period"`
	StartDate          time.Time          `json:"start_date"`
	EndDate            time.Time          `json:"end_date"`
	TotalTime          float64            `json:"total_time"`
	TotalLines         int                `json:"total_lines"`
	TotalFiles         int                `json:"total_files"`
	Languages          []APILanguageStats `json:"languages"`
	Projects           []APIProjectStats  `json:"projects"`
	Editors            []APIEditorStats   `json:"editors"`
	Files              []APIFileStats     `json:"top_files"`
	HourlyActivity     []HourlyActivity   `json:"hourly_activity"`
	PeakHour           int                `json:"peak_hour"`
	Sessions           []APISession       `json:"sessions"`
	SessionCount       int                `json:"session_count"`
	FocusScore         int                `json:"focus_score"`
	DailyGoals         DailyGoals         `json:"daily_goals"`
	MostProductiveDay  *DayRecord         `json:"most_productive_day,omitempty"`
	HighestDailyOutput *DayRecord         `json:"highest_daily_output"`
	PercentTotal       float64            `json:"percent_total,omitempty"`
}

type APILanguageStats struct {
	Name         string  `json:"name"`
	Time         float64 `json:"time"`
	Lines        int     `json:"lines"`
	Files        int     `json:"files"`
	PercentTotal float64 `json:"percent_total"`
	Proficiency  string  `json:"proficiency"`
	HoursTotal   float64 `json:"hours_total"`
}

type APIProjectStats struct {
	Name         string  `json:"name"`
	Time         float64 `json:"time"`
	Lines        int     `json:"lines"`
	Files        int     `json:"files"`
	PercentTotal float64 `json:"percent_total"`
	MainLanguage string  `json:"main_lang"`
}

type APIEditorStats struct {
	Name         string  `json:"name"`
	Time         float64 `json:"time"`
	PercentTotal float64 `json:"percent_total"`
}

type APIFileStats struct {
	Name         string    `json:"name"`
	Time         float64   `json:"time"`
	Lines        int       `json:"lines"`
	PercentTotal float64   `json:"percent_total"`
	LastEdited   time.Time `json:"last_edited"`
}

type APISession struct {
	ID         string    `json:"id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Duration   float64   `json:"duration"`
	Projects   []string  `json:"projects"`
	Languages  []string  `json:"languages,omitempty"`
	IsActive   bool      `json:"is_active"`
	BreakAfter float64   `json:"break_after,omitempty"`
}

type APIMeta struct {
	LoadedActivities int     `json:"loaded_activities"`
	TotalActivities  int     `json:"total_activities"`
	QueryTimeMs      float64 `json:"query_time_ms"`
	DataWindow       string  `json:"data_window"`
}

type StreakInfo struct {
	Current      int       `json:"current"`
	Longest      int       `json:"longest"`
	LastActivity time.Time `json:"last_activity"`
	IsActive     bool      `json:"is_active"`
}

type Achievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Unlocked    bool   `json:"unlocked"`
}

type Records struct {
	MostProductiveDay  DayRecord     `json:"most_productive_day"`
	LongestSession     SessionRecord `json:"longest_session"`
	HighestDailyOutput DayRecord     `json:"highest_daily_output"`
	BestStreak         StreakRecord  `json:"best_streak"`
	EarliestStart      TimeRecord    `json:"earliest_start"`
	LatestEnd          TimeRecord    `json:"latest_end"`
	MostLanguagesDay   LanguagesDay  `json:"most_languages_day"`
}

type DayRecord struct {
	Date         string   `json:"date"`
	Time         float64  `json:"time"`
	Lines        int      `json:"lines"`
	SessionCount int      `json:"session_count"`
	Weekday      string   `json:"weekday,omitempty"`
	Languages    []string `json:"languages,omitempty"`
	Projects     []string `json:"projects,omitempty"`
}

type SessionRecord struct {
	Date     string  `json:"date"`
	Start    string  `json:"start"`
	End      string  `json:"end"`
	Duration float64 `json:"duration"`
	Project  string  `json:"project"`
	Language string  `json:"language"`
}

type StreakRecord struct {
	DayCount  int     `json:"day_count"`
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	TotalTime float64 `json:"total_time"`
}

type TimeRecord struct {
	Time string `json:"time"`
	Date string `json:"date"`
}

type LanguagesDay struct {
	Date      string   `json:"date"`
	Languages []string `json:"languages"`
	Count     int      `json:"count"`
}

type DailyGoals struct {
	TimeGoal      float64 `json:"time_goal"`
	LinesGoal     int     `json:"lines_goal"`
	TimeProgress  float64 `json:"time_progress"`
	LinesProgress float64 `json:"lines_progress"`
	OnTrack       bool    `json:"on_track"`
}

type DailyStat struct {
	Date         string `json:"date"`
	Time         int64  `json:"time"`
	Lines        int    `json:"lines"`
	Files        int    `json:"files"`
	SessionCount int    `json:"session_count,omitempty"`
}

type HeatmapDay struct {
	Date  string `json:"date"`
	Level int    `json:"level"`
	Lines int    `json:"lines"`
	Time  int64  `json:"time"`
}

type HourlyActivity struct {
	Hour       int     `json:"hour"`
	Duration   float64 `json:"duration"`
	IsPeak     bool    `json:"is_peak"`
	Percentage float64 `json:"percentage"`
}

type APIOptions struct {
	LoadRecentDays int
}
