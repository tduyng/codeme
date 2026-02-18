// core/storage.go
package core

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteStorage struct {
	db            *sql.DB
	saveStmt      *sql.Stmt
	getRecentStmt *sql.Stmt
	countStmt     *sql.Stmt
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA cache_size = -64000",
		"PRAGMA temp_store = MEMORY",
		"PRAGMA mmap_size = 268435456",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA foreign_keys = ON",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set pragma: %w", err)
		}
	}

	if err := createSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	storage := &SQLiteStorage{db: db}
	if err := storage.prepareStatements(); err != nil {
		db.Close()
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	return storage, nil
}

func (s *SQLiteStorage) prepareStatements() error {
	var err error

	s.saveStmt, err = s.db.Prepare(`
		INSERT INTO activities 
		(id, timestamp, lines, language, project, editor, file, branch, is_write)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare save statement: %w", err)
	}

	s.getRecentStmt, err = s.db.Prepare(`
		SELECT id, timestamp, lines, language, project, editor, file, 
		       branch, is_write
		FROM activities
		WHERE timestamp >= ?
		ORDER BY timestamp ASC
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare get recent statement: %w", err)
	}

	s.countStmt, err = s.db.Prepare("SELECT COUNT(*) FROM activities")
	if err != nil {
		return fmt.Errorf("failed to prepare count statement: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) SaveActivity(activity Activity) error {
	_, err := s.saveStmt.Exec(
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

func (s *SQLiteStorage) GetActivitiesSince(since time.Time) ([]Activity, error) {
	rows, err := s.getRecentStmt.Query(since.Unix())
	if err != nil {
		return nil, fmt.Errorf("failed to query activities: %w", err)
	}
	defer rows.Close()

	return scanActivities(rows, 1000)
}

func (s *SQLiteStorage) GetActivityCount() (int, error) {
	var count int
	err := s.countStmt.QueryRow().Scan(&count)
	return count, err
}

func (s *SQLiteStorage) Optimize() error {
	if _, err := s.db.Exec("VACUUM"); err != nil {
		return fmt.Errorf("failed to vacuum: %w", err)
	}

	if _, err := s.db.Exec("ANALYZE"); err != nil {
		return fmt.Errorf("failed to analyze: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) Close() error {
	if s.saveStmt != nil {
		s.saveStmt.Close()
	}
	if s.getRecentStmt != nil {
		s.getRecentStmt.Close()
	}
	if s.countStmt != nil {
		s.countStmt.Close()
	}
	return s.db.Close()
}

func (s *SQLiteStorage) GetDB() *sql.DB {
	return s.db
}

func scanActivities(rows *sql.Rows, capacityHint int) ([]Activity, error) {
	if capacityHint < 100 {
		capacityHint = 100
	}
	activities := make([]Activity, 0, capacityHint)

	for rows.Next() {
		var a Activity
		var timestamp int64
		var isWriteInt int
		var branch sql.NullString

		err := rows.Scan(
			&a.ID, &timestamp, &a.Lines, &a.Language,
			&a.Project, &a.Editor, &a.File, &branch,
			&isWriteInt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}

		a.Timestamp = time.Unix(timestamp, 0)
		a.IsWrite = isWriteInt == 1
		a.Branch = branch.String

		activities = append(activities, a)
	}

	return activities, rows.Err()
}

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

	CREATE INDEX IF NOT EXISTS idx_timestamp_project ON activities(timestamp, project);
	CREATE INDEX IF NOT EXISTS idx_timestamp_language ON activities(timestamp, language);
	CREATE INDEX IF NOT EXISTS idx_timestamp_editor ON activities(timestamp, editor);
	CREATE INDEX IF NOT EXISTS idx_stats_covering ON activities(timestamp, id, lines, language, project, editor, file, is_write);
	`

	_, err := db.Exec(schema)
	return err
}

func GetDefaultDBPath() (string, error) {
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

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
