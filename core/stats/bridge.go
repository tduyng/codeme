// core/stats/bridge.go
package stats

import (
	"database/sql"
	"fmt"
	"time"
)

// APIStats is the simplified response for API consumers (Neovim, CLI with --json)
type APIStats struct {
	Today         APIPeriodStats       `json:"today"`
	Yesterday     APIPeriodStats       `json:"yesterday"`
	ThisWeek      APIPeriodStats       `json:"this_week"`
	LastWeek      APIPeriodStats       `json:"last_week"`
	AllTime       APIPeriodStats       `json:"all_time"`
	DayOverDay    APIComparisonResult  `json:"day_over_day"`
	StreakInfo    StreakInfo           `json:"streak_info"`
	Achievements  []Achievement        `json:"achievements"`
	Records       Records              `json:"records"`
	DailyActivity map[string]DailyStat `json:"daily_activity"`
	WeeklyHeatmap []HeatmapDay         `json:"weekly_heatmap"`
	GeneratedAt   time.Time            `json:"generated_at"`
	Meta          APIMeta              `json:"_meta"` // NEW: Performance metadata
}

// APIMeta contains performance and diagnostic information
type APIMeta struct {
	LoadedActivities int     `json:"loaded_activities"`
	TotalActivities  int     `json:"total_activities"`
	QueryTimeMs      float64 `json:"query_time_ms"`
	DataWindow       string  `json:"data_window"` // e.g., "last_365_days"
}

// APIPeriodStats is simplified version of PeriodStats for API responses
type APIPeriodStats struct {
	Period            string             `json:"period"`
	StartDate         time.Time          `json:"start_date"`
	EndDate           time.Time          `json:"end_date"`
	TotalTime         float64            `json:"total_time"`
	TotalLines        int                `json:"total_lines"`
	TotalFiles        int                `json:"total_files"`
	Languages         []APILanguageStats `json:"languages"`
	Projects          []APIProjectStats  `json:"projects"`
	Editors           []APIEditorStats   `json:"editors"`
	Files             []APIFileStats     `json:"top_files"`
	HourlyActivity    []HourlyActivity   `json:"hourly_activity"`
	PeakHour          int                `json:"peak_hour"`
	Sessions          []APISession       `json:"sessions"`
	SessionCount      int                `json:"session_count"`
	FocusScore        int                `json:"focus_score"`
	DailyGoals        DailyGoals         `json:"daily_goals,omitempty"`
	MostProductiveDay *DayRecord         `json:"most_productive_day,omitempty"`
}

// APILanguageStats - simplified language stats
type APILanguageStats struct {
	Name         string  `json:"name"`
	Time         float64 `json:"time"`
	Lines        int     `json:"lines"`
	Files        int     `json:"files"`
	PercentTotal float64 `json:"percent_total"`
	Proficiency  string  `json:"proficiency"`
	HoursTotal   float64 `json:"hours_total"`
	Trending     bool    `json:"trending"`
}

// APIProjectStats - simplified project stats
type APIProjectStats struct {
	Name         string    `json:"name"`
	Time         float64   `json:"time"`
	Lines        int       `json:"lines"`
	Files        int       `json:"files"`
	PercentTotal float64   `json:"percent_total"`
	MainLanguage string    `json:"main_lang"`
	Growth       string    `json:"growth"`
	LastActive   time.Time `json:"last_active"`
}

// APIEditorStats - simplified editor stats
type APIEditorStats struct {
	Name         string  `json:"name"`
	Time         float64 `json:"time"`
	PercentTotal float64 `json:"percent_total"`
}

// APIFileStats - simplified file stats
type APIFileStats struct {
	Name         string    `json:"name"`
	Time         float64   `json:"time"`
	Lines        int       `json:"lines"`
	PercentTotal float64   `json:"percent_total"`
	LastEdited   time.Time `json:"last_edited"`
}

// APISession - simplified session without full activities array
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

type APIComparisonResult struct {
	Trend string `json:"trend"`
}

const LOOKBACK_DAYS = 180

// GetAPIStats with smart data loading
func GetAPIStats(db *sql.DB) (*APIStats, error) {
	return GetAPIStatsWithOptions(db, APIStatsOptions{
		LoadRecentDays: LOOKBACK_DAYS, // Default: only load last x days
	})
}

// APIStatsOptions controls how much data to load
type APIStatsOptions struct {
	LoadRecentDays int  // Load only last N days (0 = all time)
	IncludeAllTime bool // Include all-time stats
}

// GetAPIStatsWithOptions allows custom data loading
func GetAPIStatsWithOptions(db *sql.DB, opts APIStatsOptions) (*APIStats, error) {
	startTime := time.Now()

	if opts.LoadRecentDays == 0 {
		opts.LoadRecentDays = LOOKBACK_DAYS
	}

	now := time.Now()
	cutoffDate := now.AddDate(0, 0, -opts.LoadRecentDays)

	// Only load recent activities
	activities, err := loadActivitiesSince(db, cutoffDate)
	if err != nil {
		return nil, err
	}

	// Get total count for metadata
	totalCount, _ := getTotalActivityCount(db)

	calc := NewCalculator(time.Local)
	fullStats, err := calc.Calculate(activities)
	if err != nil {
		return nil, err
	}

	apiStats := convertToAPIStats(fullStats)

	// Add performance metadata
	queryTime := time.Since(startTime).Seconds() * 1000 // Convert to ms
	apiStats.Meta = APIMeta{
		LoadedActivities: len(activities),
		TotalActivities:  totalCount,
		QueryTimeMs:      queryTime,
		DataWindow:       fmt.Sprintf("last_%d_days", opts.LoadRecentDays),
	}

	return apiStats, nil
}

// Load only activities since a specific date
func loadActivitiesSince(db *sql.DB, since time.Time) ([]Activity, error) {
	query := `
		SELECT id, timestamp, lines, language, project,
		       editor, file, COALESCE(branch, ''), is_write
		FROM activities
		WHERE timestamp >= ?
		ORDER BY timestamp ASC
	`

	rows, err := db.Query(query, since.Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activities := []Activity{}
	for rows.Next() {
		var a Activity
		var timestampUnix int64
		var isWriteInt int

		err := rows.Scan(
			&a.ID,
			&timestampUnix,
			&a.Lines,
			&a.Language,
			&a.Project,
			&a.Editor,
			&a.File,
			&a.Branch,
			&isWriteInt,
		)
		if err != nil {
			return nil, err
		}

		a.Timestamp = time.Unix(timestampUnix, 0)
		a.IsWrite = isWriteInt == 1

		activities = append(activities, a)
	}

	return activities, rows.Err()
}

// Get total activity count (for metadata)
func getTotalActivityCount(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM activities").Scan(&count)
	return count, err
}

// convertToAPIStats converts full Stats to simplified APIStats
func convertToAPIStats(s *Stats) *APIStats {
	return &APIStats{
		Today:         convertPeriodToAPI(s.Today),
		Yesterday:     convertPeriodToAPILight(s.Yesterday),
		ThisWeek:      convertPeriodToAPILight(s.ThisWeek),
		LastWeek:      convertPeriodToAPILight(s.LastWeek),
		AllTime:       convertPeriodToAPILight(s.AllTime),
		DayOverDay:    convertComparisonResultToAPI(s.DayOverDay),
		StreakInfo:    s.StreakInfo,
		Achievements:  s.Achievements,
		Records:       s.Records,
		DailyActivity: s.DailyActivity,
		WeeklyHeatmap: s.WeeklyHeatmap,
		GeneratedAt:   s.GeneratedAt,
	}
}

// convertPeriodToAPI converts PeriodStats to APIPeriodStats
func convertPeriodToAPI(p PeriodStats) APIPeriodStats {
	return APIPeriodStats{
		Period:            p.Period,
		StartDate:         p.StartDate,
		EndDate:           p.EndDate,
		TotalTime:         p.TotalTime,
		TotalLines:        p.TotalLines,
		TotalFiles:        p.TotalFiles,
		Languages:         convertLanguagesToAPI(p.Languages),
		Projects:          convertProjectsToAPI(p.Projects),
		Editors:           convertEditorsToAPI(p.Editors),
		Files:             convertFilesToAPI(p.Files),
		HourlyActivity:    p.Hourly,
		PeakHour:          p.PeakHour,
		Sessions:          convertSessionsToAPI(p.Sessions),
		SessionCount:      len(p.Sessions),
		FocusScore:        p.FocusScore,
		DailyGoals:        p.DailyGoals,
		MostProductiveDay: p.MostProductiveDay,
	}
}

func convertPeriodToAPILight(p PeriodStats) APIPeriodStats {
	return APIPeriodStats{
		Period:            p.Period,
		StartDate:         p.StartDate,
		EndDate:           p.EndDate,
		TotalTime:         p.TotalTime,
		TotalLines:        p.TotalLines,
		TotalFiles:        p.TotalFiles,
		Languages:         convertLanguagesToAPI(p.Languages),
		Projects:          convertProjectsToAPI(p.Projects),
		Editors:           convertEditorsToAPI(p.Editors),
		Files:             convertFilesToAPI(p.Files),
		SessionCount:      len(p.Sessions),
		HourlyActivity:    p.Hourly,
		PeakHour:          p.PeakHour,
		FocusScore:        p.FocusScore,
		DailyGoals:        p.DailyGoals,
		MostProductiveDay: p.MostProductiveDay,
	}
}

// convertLanguagesToAPI simplifies language stats
func convertLanguagesToAPI(langs []LanguageStats) []APILanguageStats {
	result := make([]APILanguageStats, len(langs))
	for i, l := range langs {
		result[i] = APILanguageStats{
			Name:        l.Name,
			Time:        l.Time,
			Lines:       l.Lines,
			Files:       l.Files,
			Proficiency: l.Proficiency,
			HoursTotal:  l.HoursTotal,
			Trending:    l.Trending,
		}
	}
	return result
}

// convertProjectsToAPI simplifies project stats
func convertProjectsToAPI(projs []ProjectStats) []APIProjectStats {
	result := make([]APIProjectStats, len(projs))
	for i, p := range projs {
		result[i] = APIProjectStats{
			Name:         p.Name,
			Time:         p.Time,
			Lines:        p.Lines,
			Files:        p.Files,
			MainLanguage: p.MainLanguage,
			Growth:       p.Growth,
			LastActive:   p.LastActive,
		}
	}
	return result
}

// convertEditorsToAPI simplifies editor stats
func convertEditorsToAPI(eds []EditorStats) []APIEditorStats {
	result := make([]APIEditorStats, len(eds))
	for i, e := range eds {
		result[i] = APIEditorStats{
			Name: e.Name,
			Time: e.Time,
		}
	}
	return result
}

// convertFilesToAPI simplifies file stats
func convertFilesToAPI(files []FileStats) []APIFileStats {
	result := make([]APIFileStats, len(files))
	for i, f := range files {
		result[i] = APIFileStats{
			Name:       f.Name,
			Time:       f.Time,
			Lines:      f.Lines,
			LastEdited: f.LastEdited,
		}
	}
	return result
}

func convertSessionsToAPI(sessions []Session) []APISession {
	result := make([]APISession, len(sessions))

	for i, s := range sessions {
		result[i] = APISession{
			ID:         s.ID,
			StartTime:  s.StartTime,
			EndTime:    s.EndTime,
			Duration:   s.Duration,
			Projects:   append([]string(nil), s.Projects...),
			Languages:  append([]string(nil), s.Languages...),
			IsActive:   s.IsActive,
			BreakAfter: s.BreakAfter,
		}
	}

	return result
}

func convertComparisonResultToAPI(comp ComparisonResult) APIComparisonResult {
	return APIComparisonResult{
		Trend: comp.Trend,
	}
}
