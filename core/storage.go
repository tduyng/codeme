// core/storage.go
package core

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tduyng/codeme/core/stats"
)

// Storage defines the interface for persisting activity data
type Storage interface {
	SaveActivity(activity stats.Activity) error
	GetActivities(start, end time.Time) ([]stats.Activity, error)
	GetAllActivities() ([]stats.Activity, error)
	GetActivitiesSince(since time.Time) ([]stats.Activity, error) // Load recent data only
	GetActivityCount() (int, error)                               // For monitoring
	Optimize() error                                              // Maintenance
	Close() error
}

// SQLiteStorage implements Storage using SQLite
type SQLiteStorage struct {
	db *sql.DB
}

// OpenDB opens or creates the default SQLite database
func OpenDB() (*sql.DB, error) {
	dbPath, err := GetDefaultDBPath()
	if err != nil {
		return nil, err
	}

	storage, err := NewSQLiteStorage(dbPath)
	if err != nil {
		return nil, err
	}

	return storage.db, nil
}

// NewSQLiteStorage creates a new SQLite storage with optimizations
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	// Ensure directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 🚀 OPTIMIZATION: Configure SQLite for performance
	pragmas := []string{
		"PRAGMA journal_mode = WAL",    // Better concurrency, faster writes
		"PRAGMA synchronous = NORMAL",  // Balance safety/speed
		"PRAGMA cache_size = -64000",   // 64MB cache
		"PRAGMA temp_store = MEMORY",   // Use memory for temp operations
		"PRAGMA mmap_size = 268435456", // 256MB memory-mapped I/O
		"PRAGMA busy_timeout = 5000",   // Wait 5s on lock
		"PRAGMA foreign_keys = ON",     // Enforce constraints
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set pragma: %w", err)
		}
	}

	// Create schema
	if err := createSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	// Set connection pool limits (SQLite works best with minimal connections)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	return &SQLiteStorage{db: db}, nil
}

// SaveActivity persists a single activity to the database
func (s *SQLiteStorage) SaveActivity(activity stats.Activity) error {
	query := `
	INSERT INTO activities 
	(id, timestamp, lines, language, project, editor, file, branch, is_write)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(
		query,
		activity.ID,
		activity.Timestamp.Unix(),
		activity.Lines,
		activity.Language,
		activity.Project,
		activity.Editor,
		activity.File,
		activity.Branch,
		boolToInt(activity.IsWrite),
	)

	if err != nil {
		return fmt.Errorf("failed to insert activity: %w", err)
	}

	return nil
}

// GetActivities retrieves activities for a date range
func (s *SQLiteStorage) GetActivities(startTime, endTime time.Time) ([]stats.Activity, error) {
	query := `
	SELECT id, timestamp, lines, language, project, editor, file, branch, is_write
	FROM activities
	WHERE timestamp >= ? AND timestamp < ?
	ORDER BY timestamp ASC
	`

	rows, err := s.db.Query(query, startTime.Unix(), endTime.Unix())
	if err != nil {
		return nil, fmt.Errorf("failed to query activities: %w", err)
	}
	defer rows.Close()

	return scanActivities(rows)
}

// GetAllActivities retrieves all activities from the database
// ⚠️ WARNING: This can be slow with large datasets. Use GetActivitiesSince() instead.
func (s *SQLiteStorage) GetAllActivities() ([]stats.Activity, error) {
	query := `
	SELECT id, timestamp, lines, language, project, editor, file, branch, is_write
	FROM activities
	ORDER BY timestamp ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all activities: %w", err)
	}
	defer rows.Close()

	return scanActivities(rows)
}

// 🚀 NEW: GetActivitiesSince loads only recent activities (FAST)
// This is the preferred method for most operations
func (s *SQLiteStorage) GetActivitiesSince(since time.Time) ([]stats.Activity, error) {
	query := `
	SELECT id, timestamp, lines, language, project, editor, file, branch, is_write
	FROM activities
	WHERE timestamp >= ?
	ORDER BY timestamp ASC
	`

	rows, err := s.db.Query(query, since.Unix())
	if err != nil {
		return nil, fmt.Errorf("failed to query activities since %v: %w", since, err)
	}
	defer rows.Close()

	return scanActivities(rows)
}

// 🚀 NEW: GetActivityCount returns total number of activities (for monitoring)
func (s *SQLiteStorage) GetActivityCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM activities").Scan(&count)
	return count, err
}

// scanActivities is a helper function to scan activity rows
func scanActivities(rows *sql.Rows) ([]stats.Activity, error) {
	var activities []stats.Activity

	for rows.Next() {
		var a stats.Activity
		var timestamp int64
		var branch sql.NullString
		var isWriteInt int

		err := rows.Scan(
			&a.ID,
			&timestamp,
			&a.Lines,
			&a.Language,
			&a.Project,
			&a.Editor,
			&a.File,
			&branch,
			&isWriteInt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}

		a.Timestamp = time.Unix(timestamp, 0)
		if branch.Valid {
			a.Branch = branch.String
		}
		a.IsWrite = isWriteInt == 1

		activities = append(activities, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return activities, nil
}

// 🚀 NEW: Optimize performs database maintenance (run monthly)
func (s *SQLiteStorage) Optimize() error {
	// Rebuild indexes and reclaim space
	if _, err := s.db.Exec("VACUUM"); err != nil {
		return fmt.Errorf("failed to vacuum: %w", err)
	}

	// Analyze query patterns for better optimization
	if _, err := s.db.Exec("ANALYZE"); err != nil {
		return fmt.Errorf("failed to analyze: %w", err)
	}

	return nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// GetDB returns the underlying database connection
func (s *SQLiteStorage) GetDB() *sql.DB {
	return s.db
}

// createSchema initializes the database schema with optimizations
func createSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS activities (
		id TEXT PRIMARY KEY,
		timestamp INTEGER NOT NULL,
		lines INTEGER DEFAULT 0,
		language TEXT NOT NULL,
		project TEXT NOT NULL,
		editor TEXT DEFAULT 'unknown',
		file TEXT DEFAULT '',
		branch TEXT,
		is_write INTEGER DEFAULT 1,
		created_at INTEGER DEFAULT (strftime('%s', 'now'))
	) WITHOUT ROWID;

	-- Composite indexes for common query patterns
	CREATE INDEX IF NOT EXISTS idx_timestamp_project ON activities(timestamp, project);
	CREATE INDEX IF NOT EXISTS idx_timestamp_language ON activities(timestamp, language);
	CREATE INDEX IF NOT EXISTS idx_timestamp_editor ON activities(timestamp, editor);
	`

	statements := strings.Split(schema, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		_, err := db.Exec(stmt)
		if err != nil {
			return fmt.Errorf("failed to execute schema statement: %w", err)
		}
	}

	return nil
}

// GetDefaultDBPath returns the default database path
func GetDefaultDBPath() (string, error) {
	// Try XDG_DATA_HOME first (Linux standard)
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		dataHome = filepath.Join(home, ".local", "share")
	}

	dbDir := filepath.Join(dataHome, "codeme")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	return filepath.Join(dbDir, "codeme.db"), nil
}

// boolToInt converts a boolean to an integer for SQLite storage
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
