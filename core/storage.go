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
