package core

import "time"

type Activity struct {
	ID        string
	Timestamp time.Time
	Duration  float64
	Lines     int
	Language  string
	Project   string
	Editor    string
	File      string
	Branch    string
	IsWrite   bool
}

type Session struct {
	ID         string
	StartTime  time.Time
	EndTime    time.Time
	Duration   float64
	Projects   []string
	Languages  []string
	IsActive   bool
	BreakAfter float64
}

type DailySummary struct {
	Date          string
	TotalTime     float64
	TotalLines    int
	ActivityCount int
}

type PeriodSummary struct {
	TotalTime     float64
	TotalLines    int
	ActivityCount int
}

type LanguageRow struct {
	Language   string
	TotalTime  float64
	TotalLines int
}

type ProjectRow struct {
	Project      string
	TotalTime    float64
	TotalLines   int
	MainLanguage string
}

type EditorRow struct {
	Editor     string
	TotalTime  float64
	TotalLines int
}

type Storage interface {
	SaveActivity(Activity) error
	GetActivitiesSince(time.Time) ([]Activity, error)
	GetActivityCount() (int, error)
	GetDailySummaries() (map[string]DailySummary, error)
	GetPeriodSummary(from, to time.Time) (PeriodSummary, error)
	GetLanguageSummary(from, to time.Time) ([]LanguageRow, error)
	GetProjectSummary(from, to time.Time) ([]ProjectRow, error)
	GetEditorSummary(from, to time.Time) ([]EditorRow, error)
	Optimize() error
	RebuildSummaries() error
	Close() error
}
