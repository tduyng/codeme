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

func OpenReadOnlyStorage(dbPath string) (*SQLiteStorage, error) {
	storage, err := NewSQLiteStorage(dbPath)
	if err != nil {
		return nil, err
	}

	storage.db.Exec("PRAGMA query_only = ON")

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

func (s *SQLiteStorage) estimateDuration(tx *sql.Tx, date string, now time.Time) (float64, error) {
	const maxGap = 120.0

	var lastTS int64
	err := tx.QueryRow(`
		SELECT COALESCE(MAX(timestamp), 0) FROM activities
		WHERE date(timestamp, 'unixepoch', 'localtime') = ?
		  AND timestamp < ?
	`, date, now.Unix()).Scan(&lastTS)
	if err != nil {
		return maxGap, err
	}

	if lastTS == 0 {
		return maxGap, nil
	}

	gap := now.Unix() - lastTS
	if gap > maxGap {
		return maxGap, nil
	}
	return float64(gap), nil
}

func (s *SQLiteStorage) SaveActivity(activity Activity) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO activities 
		(id, timestamp, lines, language, project, editor, file, branch, is_write)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
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

	date := activity.Timestamp.Format("2006-01-02")
	duration, err := s.estimateDuration(tx, date, activity.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to estimate duration: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO daily_summary (date, total_time, total_lines, activity_count, first_activity, last_activity)
		VALUES (?, ?, ?, 1, ?, ?)
		ON CONFLICT(date) DO UPDATE SET
			total_time = total_time + excluded.total_time,
			total_lines = total_lines + excluded.total_lines,
			activity_count = activity_count + 1,
			first_activity = CASE WHEN excluded.first_activity < daily_summary.first_activity THEN excluded.first_activity ELSE daily_summary.first_activity END,
			last_activity = CASE WHEN excluded.last_activity > daily_summary.last_activity THEN excluded.last_activity ELSE daily_summary.last_activity END,
			updated_at = strftime('%s', 'now')
	`, date, duration, activity.Lines, activity.Timestamp.Unix(), activity.Timestamp.Unix())
	if err != nil {
		return fmt.Errorf("failed to update daily summary: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO daily_language_summary (date, language, total_time, total_lines, file_count)
		VALUES (?, ?, ?, ?, 1)
		ON CONFLICT(date, language) DO UPDATE SET
			total_time = total_time + excluded.total_time,
			total_lines = total_lines + excluded.total_lines,
			file_count = file_count + 1
	`, date, activity.Language, duration, activity.Lines)
	if err != nil {
		return fmt.Errorf("failed to update language summary: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO daily_project_summary (date, project, total_time, total_lines, main_language, file_count)
		VALUES (?, ?, ?, ?, ?, 1)
		ON CONFLICT(date, project) DO UPDATE SET
			total_time = total_time + excluded.total_time,
			total_lines = total_lines + excluded.total_lines,
			main_language = CASE 
				WHEN (SELECT total_time FROM daily_language_summary WHERE date = excluded.date AND language = excluded.main_language) > 
				      (SELECT COALESCE(MAX(total_time), 0) FROM daily_language_summary WHERE date = excluded.date AND language = excluded.main_language)
				THEN excluded.main_language
				ELSE daily_project_summary.main_language
			END,
			file_count = file_count + 1
	`, date, activity.Project, duration, activity.Lines, activity.Language)
	if err != nil {
		return fmt.Errorf("failed to update project summary: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO daily_editor_summary (date, editor, total_time, total_lines)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(date, editor) DO UPDATE SET
			total_time = total_time + excluded.total_time,
			total_lines = total_lines + excluded.total_lines
	`, date, activity.Editor, duration, activity.Lines)
	if err != nil {
		return fmt.Errorf("failed to update editor summary: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
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

func (s *SQLiteStorage) GetPeriodSummary(from, to time.Time) (PeriodSummary, error) {
	var ps PeriodSummary
	err := s.db.QueryRow(`
		SELECT COALESCE(SUM(total_time), 0), COALESCE(SUM(total_lines), 0),
		       COALESCE(SUM(activity_count), 0)
		FROM daily_summary 
		WHERE date >= ? AND date <= ?
	`, from.Format("2006-01-02"), to.Format("2006-01-02")).Scan(&ps.TotalTime, &ps.TotalLines, &ps.ActivityCount)
	return ps, err
}

func (s *SQLiteStorage) GetLanguageSummary(from, to time.Time) ([]LanguageRow, error) {
	rows, err := s.db.Query(`
		SELECT language, SUM(total_time), SUM(total_lines)
		FROM daily_language_summary 
		WHERE date >= ? AND date <= ?
		GROUP BY language ORDER BY SUM(total_time) DESC
	`, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []LanguageRow
	for rows.Next() {
		var lr LanguageRow
		if err := rows.Scan(&lr.Language, &lr.TotalTime, &lr.TotalLines); err != nil {
			return nil, err
		}
		results = append(results, lr)
	}
	return results, nil
}

func (s *SQLiteStorage) GetProjectSummary(from, to time.Time) ([]ProjectRow, error) {
	rows, err := s.db.Query(`
		SELECT project, SUM(total_time), SUM(total_lines), main_language
		FROM daily_project_summary 
		WHERE date >= ? AND date <= ?
		GROUP BY project ORDER BY SUM(total_time) DESC
	`, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ProjectRow
	for rows.Next() {
		var pr ProjectRow
		if err := rows.Scan(&pr.Project, &pr.TotalTime, &pr.TotalLines, &pr.MainLanguage); err != nil {
			return nil, err
		}
		results = append(results, pr)
	}
	return results, nil
}

func (s *SQLiteStorage) GetEditorSummary(from, to time.Time) ([]EditorRow, error) {
	rows, err := s.db.Query(`
		SELECT editor, SUM(total_time), SUM(total_lines)
		FROM daily_editor_summary 
		WHERE date >= ? AND date <= ?
		GROUP BY editor ORDER BY SUM(total_time) DESC
	`, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []EditorRow
	for rows.Next() {
		var er EditorRow
		if err := rows.Scan(&er.Editor, &er.TotalTime, &er.TotalLines); err != nil {
			return nil, err
		}
		results = append(results, er)
	}
	return results, nil
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

func (s *SQLiteStorage) RebuildSummaries() error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
		SELECT id, timestamp, lines, language, project, editor, file
		FROM activities
		ORDER BY timestamp ASC
	`)
	if err != nil {
		return fmt.Errorf("failed to load activities: %w", err)
	}

	type rawActivity struct {
		id        string
		timestamp int64
		lines     int
		language  string
		project   string
		editor    string
		file      string
	}

	var activitiesList []rawActivity
	for rows.Next() {
		var a rawActivity
		if err := rows.Scan(&a.id, &a.timestamp, &a.lines, &a.language, &a.project, &a.editor, &a.file); err != nil {
			rows.Close()
			return err
		}
		activitiesList = append(activitiesList, a)
	}
	rows.Close()

	if len(activitiesList) == 0 {
		return nil
	}

	const maxGap = 120.0

	type dailyAgg struct {
		totalTime     float64
		totalLines    int
		activityCount int
		firstActivity int64
		lastActivity  int64
	}
	type langAgg struct {
		totalTime  float64
		totalLines int
		fileCount  int
	}
	type projAgg struct {
		totalTime    float64
		totalLines   int
		mainLanguage string
		fileCount    int
	}
	type editorAgg struct {
		totalTime  float64
		totalLines int
	}

	dailySummary := make(map[string]dailyAgg)
	langSummary := make(map[string]langAgg)
	projSummary := make(map[string]projAgg)
	editorSummary := make(map[string]editorAgg)

	var prevTS int64
	for i, a := range activitiesList {
		ts := a.timestamp
		date := time.Unix(ts, 0).In(time.Local).Format("2006-01-02")

		gap := maxGap
		if i > 0 {
			gap = float64(ts - prevTS)
			if gap > maxGap {
				gap = maxGap
			}
		}

		ds := dailySummary[date]
		ds.totalTime += gap
		ds.totalLines += a.lines
		ds.activityCount++
		if ds.firstActivity == 0 || ts < ds.firstActivity {
			ds.firstActivity = ts
		}
		if ts > ds.lastActivity {
			ds.lastActivity = ts
		}
		dailySummary[date] = ds

		langKey := date + "|" + a.language
		ls := langSummary[langKey]
		ls.totalTime += gap
		ls.totalLines += a.lines
		ls.fileCount++
		langSummary[langKey] = ls

		projKey := date + "|" + a.project
		ps := projSummary[projKey]
		ps.totalTime += gap
		ps.totalLines += a.lines
		if a.language != "" {
			ps.mainLanguage = a.language
		}
		ps.fileCount++
		projSummary[projKey] = ps

		editorKey := date + "|" + a.editor
		es := editorSummary[editorKey]
		es.totalTime += gap
		es.totalLines += a.lines
		editorSummary[editorKey] = es

		prevTS = ts
	}

	for date, ds := range dailySummary {
		_, err := tx.Exec(`
			INSERT OR REPLACE INTO daily_summary
				(date, total_time, total_lines, activity_count, first_activity, last_activity)
			VALUES (?, ?, ?, ?, ?, ?)
		`, date, ds.totalTime, ds.totalLines, ds.activityCount, ds.firstActivity, ds.lastActivity)
		if err != nil {
			return fmt.Errorf("failed to rebuild daily summary: %w", err)
		}
	}

	for key, ls := range langSummary {
		parts := splitKey(key)
		date := parts[0]
		lang := parts[1]
		_, err := tx.Exec(`
			INSERT OR REPLACE INTO daily_language_summary
				(date, language, total_time, total_lines, file_count)
			VALUES (?, ?, ?, ?, ?)
		`, date, lang, ls.totalTime, ls.totalLines, ls.fileCount)
		if err != nil {
			return fmt.Errorf("failed to rebuild language summary: %w", err)
		}
	}

	for key, ps := range projSummary {
		parts := splitKey(key)
		date := parts[0]
		proj := parts[1]
		_, err := tx.Exec(`
			INSERT OR REPLACE INTO daily_project_summary
				(date, project, total_time, total_lines, main_language, file_count)
			VALUES (?, ?, ?, ?, ?, ?)
		`, date, proj, ps.totalTime, ps.totalLines, ps.mainLanguage, ps.fileCount)
		if err != nil {
			return fmt.Errorf("failed to rebuild project summary: %w", err)
		}
	}

	for key, es := range editorSummary {
		parts := splitKey(key)
		date := parts[0]
		editor := parts[1]
		_, err := tx.Exec(`
			INSERT OR REPLACE INTO daily_editor_summary
				(date, editor, total_time, total_lines)
			VALUES (?, ?, ?, ?)
		`, date, editor, es.totalTime, es.totalLines)
		if err != nil {
			return fmt.Errorf("failed to rebuild editor summary: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

func splitKey(key string) []string {
	for i := 0; i < len(key); i++ {
		if key[i] == '|' {
			return []string{key[:i], key[i+1:]}
		}
	}
	return []string{key, ""}
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

	CREATE TABLE IF NOT EXISTS daily_summary (
		date TEXT PRIMARY KEY,
		total_time REAL DEFAULT 0,
		total_lines INTEGER DEFAULT 0,
		activity_count INTEGER DEFAULT 0,
		session_count INTEGER DEFAULT 0,
		longest_session REAL DEFAULT 0,
		total_session_time REAL DEFAULT 0,
		first_activity INTEGER,
		last_activity INTEGER,
		created_at INTEGER DEFAULT (strftime('%s', 'now')),
		updated_at INTEGER DEFAULT (strftime('%s', 'now'))
	);

	CREATE TABLE IF NOT EXISTS daily_language_summary (
		date TEXT NOT NULL,
		language TEXT NOT NULL,
		total_time REAL DEFAULT 0,
		total_lines INTEGER DEFAULT 0,
		file_count INTEGER DEFAULT 0,
		PRIMARY KEY (date, language)
	);

	CREATE TABLE IF NOT EXISTS daily_project_summary (
		date TEXT NOT NULL,
		project TEXT NOT NULL,
		total_time REAL DEFAULT 0,
		total_lines INTEGER DEFAULT 0,
		main_language TEXT,
		file_count INTEGER DEFAULT 0,
		PRIMARY KEY (date, project)
	);

	CREATE TABLE IF NOT EXISTS daily_editor_summary (
		date TEXT NOT NULL,
		editor TEXT NOT NULL,
		total_time REAL DEFAULT 0,
		total_lines INTEGER DEFAULT 0,
		PRIMARY KEY (date, editor)
	);
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
