package core

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Heartbeat struct {
	ID           int64
	Timestamp    time.Time
	File         string
	Language     string
	Project      string
	Branch       string
	Lines        int // Kept for backward compatibility (lines changed)
	LinesChanged int // New: Actual changed lines
	LinesTotal   int // New: Total file size
	CreatedAt    time.Time
}

type Stats struct {
	// Total stats (all-time)
	TotalTime  int64 `json:"total_time"`
	TotalLines int   `json:"total_lines"`
	TotalFiles int   `json:"total_files"`

	// Today stats
	TodayTime  int64     `json:"today_time"`
	TodayLines int       `json:"today_lines"`
	TodayFiles int       `json:"today_files"`
	Today      DailyStat `json:"today"` // For nvim plugin compatibility

	// Yesterday stats (for comparison)
	YesterdayTime  int64 `json:"yesterday_time"`
	YesterdayLines int   `json:"yesterday_lines"`
	YesterdayFiles int   `json:"yesterday_files"`

	// This week stats (Mon-Sun)
	WeekTime  int64 `json:"week_time"`
	WeekLines int   `json:"week_lines"`
	WeekFiles int   `json:"week_files"`

	// Last week stats (for comparison)
	LastWeekTime  int64 `json:"last_week_time"`
	LastWeekLines int   `json:"last_week_lines"`
	LastWeekFiles int   `json:"last_week_files"`

	// This month stats
	MonthTime  int64 `json:"month_time"`
	MonthLines int   `json:"month_lines"`
	MonthFiles int   `json:"month_files"`

	// Peak activity insights
	MostActiveHour    int    `json:"most_active_hour"`     // 0-23
	MostActiveDay     string `json:"most_active_day"`      // "Monday", "Tuesday", etc.
	MostActiveDayTime int64  `json:"most_active_day_time"` // Time spent on most active day

	// Aggregated data
	Projects       map[string]ProjectStat `json:"projects"`
	Languages      map[string]LangStat    `json:"languages"`
	TopFiles       []FileStat             `json:"top_files"`
	DailyActivity  map[string]DailyStat   `json:"daily_activity"`
	HourlyActivity map[int]int            `json:"hourly_activity"`

	// Weekly heatmap (last 12 weeks, for GitHub-style contribution grid)
	// Keys are dates in "2006-01-02" format, values are activity levels 0-4
	WeeklyHeatmap []HeatmapDay `json:"weekly_heatmap"`

	// Session history (today's sessions)
	Sessions []Session `json:"sessions"`

	// Streaks
	Streak        int `json:"streak"`
	LongestStreak int `json:"longest_streak"`

	// Achievements
	Achievements []Achievement `json:"achievements"`
}

type HeatmapDay struct {
	Date  string `json:"date"`
	Level int    `json:"level"` // 0-4 (none, low, medium, high, max)
	Lines int    `json:"lines"`
	Time  int64  `json:"time"`
}

type Session struct {
	Start    string `json:"start"`    // RFC3339 timestamp
	End      string `json:"end"`      // RFC3339 timestamp
	Duration int64  `json:"duration"` // seconds
	Project  string `json:"project"`
}

type Achievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Unlocked    bool   `json:"unlocked"`
	UnlockedAt  string `json:"unlocked_at,omitempty"` // RFC3339 timestamp
}

type ProjectStat struct {
	Time  int64 `json:"time"`
	Lines int   `json:"lines"`
	Files int   `json:"files"`
}

type LangStat struct {
	Lines int   `json:"lines"`
	Time  int64 `json:"time"`
	Files int   `json:"files"`
}

type FileStat struct {
	Path  string `json:"path"`
	Lines int    `json:"lines"`
	Time  int64  `json:"time"`
}

type DailyStat struct {
	Lines int   `json:"lines"`
	Time  int64 `json:"time"`
	Files int   `json:"files"`
}

func MigrateSchema(db *sql.DB) error {
	// Add new columns if they don't exist (ignore errors if columns already exist)
	db.Exec(`ALTER TABLE heartbeats ADD COLUMN lines_changed INTEGER DEFAULT 0`)
	db.Exec(`ALTER TABLE heartbeats ADD COLUMN lines_total INTEGER DEFAULT 0`)
	return nil
}

func OpenDB() (*sql.DB, error) {
	return OpenDBWithPath("")
}

// OpenDBWithPath opens a database at the specified path, or uses default if empty
func OpenDBWithPath(customPath string) (*sql.DB, error) {
	var dbPath string

	if customPath != "" {
		dbPath = customPath
	} else {
		home, _ := os.UserHomeDir()
		dbPath = filepath.Join(home, ".local", "share", "codeme", "codeme.db")
	}

	os.MkdirAll(filepath.Dir(dbPath), 0755)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Auto-create schema
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS heartbeats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME NOT NULL,
			file TEXT NOT NULL,
			language TEXT,
			project TEXT,
			branch TEXT,
			lines INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_timestamp ON heartbeats(timestamp);
		CREATE INDEX IF NOT EXISTS idx_project ON heartbeats(project);
		CREATE INDEX IF NOT EXISTS idx_language ON heartbeats(language);
	`)

	if err != nil {
		return nil, err
	}

	// Run migration to add new columns
	MigrateSchema(db)

	return db, nil
}

func SaveHeartbeat(db *sql.DB, hb Heartbeat) error {
	// Format timestamp as RFC3339 for proper SQLite DATETIME storage
	timestampStr := hb.Timestamp.Format(time.RFC3339)

	_, err := db.Exec(`
		INSERT INTO heartbeats (timestamp, file, language, project, branch, lines, lines_changed, lines_total)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, timestampStr, hb.File, hb.Language, hb.Project, hb.Branch, hb.Lines, hb.LinesChanged, hb.LinesTotal)
	return err
}
