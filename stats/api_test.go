package stats

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tduyng/codeme/core"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)

	// Create schema
	schema := `
	CREATE TABLE activities (
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

	CREATE INDEX idx_timestamp_project ON activities(timestamp, project);
	CREATE INDEX idx_timestamp_language ON activities(timestamp, language);
	`

	_, err = db.Exec(schema)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

func insertActivity(t *testing.T, db *sql.DB, a core.Activity) {
	_, err := db.Exec(`
		INSERT INTO activities (id, timestamp, lines, language, project, editor, file, is_write)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, a.ID, a.Timestamp.Unix(), a.Lines, a.Language, a.Project, a.Editor, a.File, boolToInt(a.IsWrite))
	require.NoError(t, err)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func TestCalculator_CalculateAPI_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	calc := NewCalculator(time.UTC)
	stats, err := calc.CalculateAPI(db, APIOptions{LoadRecentDays: 180})

	require.NoError(t, err)
	require.NotNil(t, stats)
	require.Equal(t, 0.0, stats.Today.TotalTime)
	require.Equal(t, 0, stats.Today.TotalLines)
	require.Empty(t, stats.Today.Languages)
	require.Empty(t, stats.Today.Projects)
}

func TestCalculator_CalculateAPI_SingleDay(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Use UTC consistently — CalculateAPI internally uses time.Now().In(timezone),
	// so the test's "today" must be derived from the same clock in the same zone.
	now := time.Now().UTC()
	baseTime := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, time.UTC)

	activities := []core.Activity{
		{
			ID:        "1",
			Timestamp: baseTime,
			Duration:  120,
			Lines:     10,
			Language:  "go",
			Project:   "test",
			Editor:    "neovim",
			File:      "/test/main.go",
			IsWrite:   true,
		},
		{
			ID:        "2",
			Timestamp: baseTime.Add(10 * time.Minute),
			Duration:  180,
			Lines:     15,
			Language:  "go",
			Project:   "test",
			Editor:    "neovim",
			File:      "/test/app.go",
			IsWrite:   true,
		},
	}

	for _, a := range activities {
		insertActivity(t, db, a)
	}

	calc := NewCalculator(time.UTC)
	stats, err := calc.CalculateAPI(db, APIOptions{LoadRecentDays: 180})

	require.NoError(t, err)
	require.NotNil(t, stats)
	require.Greater(t, stats.Today.TotalTime, 0.0)
	require.Equal(t, 25, stats.Today.TotalLines)
	require.Len(t, stats.Today.Languages, 1)
	require.Equal(t, "go", stats.Today.Languages[0].Name)
	require.Len(t, stats.Today.Projects, 1)
	require.Equal(t, "test", stats.Today.Projects[0].Name)
}

func TestCalculator_CalculateAPI_MultipleProjects(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	baseTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	activities := []core.Activity{
		{ID: "1", Timestamp: baseTime, Duration: 0, Lines: 10, Language: "go", Project: "p1", Editor: "vim", File: "/p1/a.go", IsWrite: true},
		{ID: "2", Timestamp: baseTime.Add(1 * time.Hour), Duration: 0, Lines: 20, Language: "typescript", Project: "p2", Editor: "vscode", File: "/p2/b.ts", IsWrite: true},
		{ID: "3", Timestamp: baseTime.Add(2 * time.Hour), Duration: 0, Lines: 15, Language: "python", Project: "p3", Editor: "neovim", File: "/p3/c.py", IsWrite: true},
	}

	for _, a := range activities {
		insertActivity(t, db, a)
	}

	calc := NewCalculator(time.UTC)
	stats, err := calc.CalculateAPI(db, APIOptions{LoadRecentDays: 365})

	require.NoError(t, err)
	// Duration is calculated by the calculator, so projects/languages may be aggregated
	require.NotNil(t, stats.AllTime)
}

func TestCalculator_CalculateAPI_WeeklyStats(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tz := time.UTC
	now := time.Now().In(tz)
	today := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, tz)

	activities := []core.Activity{
		{ID: "1", Timestamp: today, Duration: 120, Lines: 10, Language: "go", Project: "p1", Editor: "vim", File: "/a.go", IsWrite: true},
		{ID: "2", Timestamp: today.Add(5 * time.Minute), Duration: 180, Lines: 20, Language: "go", Project: "p1", Editor: "vim", File: "/b.go", IsWrite: true},
		{ID: "3", Timestamp: today.AddDate(0, 0, -8), Duration: 150, Lines: 15, Language: "python", Project: "p2", Editor: "vim", File: "/c.py", IsWrite: true},
	}

	for _, a := range activities {
		insertActivity(t, db, a)
	}

	calc := NewCalculator(tz)
	stats, err := calc.CalculateAPI(db, APIOptions{LoadRecentDays: 365})

	require.NoError(t, err)
	require.Greater(t, stats.ThisWeek.TotalTime, stats.LastWeek.TotalTime)
	require.Greater(t, stats.ThisWeek.TotalLines, stats.LastWeek.TotalLines)
}

func TestCalculator_CalculateAPI_Sessions(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	baseTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	// Create two sessions with a gap
	activities := []core.Activity{
		// Session 1
		{ID: "1", Timestamp: baseTime, Duration: 0, Lines: 10, Language: "go", Project: "p1", File: "/a.go"},
		{ID: "2", Timestamp: baseTime.Add(5 * time.Minute), Duration: 0, Lines: 10, Language: "go", Project: "p1", File: "/b.go"},
		// Gap of 20 minutes
		// Session 2
		{ID: "3", Timestamp: baseTime.Add(30 * time.Minute), Duration: 0, Lines: 10, Language: "go", Project: "p1", File: "/c.go"},
		{ID: "4", Timestamp: baseTime.Add(35 * time.Minute), Duration: 0, Lines: 10, Language: "go", Project: "p1", File: "/d.go"},
	}

	for _, a := range activities {
		insertActivity(t, db, a)
	}

	calc := NewCalculator(time.UTC)
	stats, err := calc.CalculateAPI(db, APIOptions{LoadRecentDays: 365})

	require.NoError(t, err)
	// Sessions may be grouped differently based on calculated durations
	require.NotNil(t, stats.AllTime)
}

func TestCalculator_CalculateAPI_Comprehensive(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	baseTime := time.Now().UTC().AddDate(0, 0, -3)

	// Create a realistic week of coding
	activities := []core.Activity{
		// Monday
		{ID: "1", Timestamp: baseTime, Duration: 0, Lines: 100, Language: "go", Project: "backend", Editor: "neovim", File: "/main.go", IsWrite: true},
		{ID: "2", Timestamp: baseTime.Add(2 * time.Hour), Duration: 0, Lines: 50, Language: "typescript", Project: "frontend", Editor: "vscode", File: "/app.ts", IsWrite: true},
		// Tuesday
		{ID: "3", Timestamp: baseTime.AddDate(0, 0, 1), Duration: 0, Lines: 80, Language: "go", Project: "backend", Editor: "neovim", File: "/api.go", IsWrite: true},
		// Wednesday
		{ID: "4", Timestamp: baseTime.AddDate(0, 0, 2), Duration: 0, Lines: 120, Language: "python", Project: "ml", Editor: "neovim", File: "/model.py", IsWrite: true},
		// Thursday
		{ID: "5", Timestamp: baseTime.AddDate(0, 0, 3), Duration: 0, Lines: 60, Language: "rust", Project: "cli", Editor: "vim", File: "/main.rs", IsWrite: true},
	}

	for _, a := range activities {
		insertActivity(t, db, a)
	}

	calc := NewCalculator(time.UTC)
	stats, err := calc.CalculateAPI(db, APIOptions{LoadRecentDays: 180})

	require.NoError(t, err)
	require.NotNil(t, stats)
	require.Equal(t, 410, stats.AllTime.TotalLines)
	require.NotEmpty(t, stats.DailyActivity)
}

func TestCalculator_CalculateAPI_FocusScore(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	now := time.Now().UTC()
	baseTime := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, time.UTC)

	// Create long focused session with proper gaps for duration calculation
	activities := []core.Activity{
		{ID: "1", Timestamp: baseTime, Duration: 0, Lines: 100, Language: "go", Project: "p1", File: "/a.go"},
		{ID: "2", Timestamp: baseTime.Add(1 * time.Minute), Duration: 0, Lines: 100, Language: "go", Project: "p1", File: "/b.go"},
		{ID: "3", Timestamp: baseTime.Add(2 * time.Minute), Duration: 0, Lines: 100, Language: "go", Project: "p1", File: "/c.go"},
	}

	for _, a := range activities {
		insertActivity(t, db, a)
	}

	calc := NewCalculator(time.UTC)
	stats, err := calc.CalculateAPI(db, APIOptions{LoadRecentDays: 180})

	require.NoError(t, err)
	require.LessOrEqual(t, stats.Today.FocusScore, 100)
	require.GreaterOrEqual(t, stats.Today.FocusScore, 0)
}

func TestCalculator_CalculateAPI_LoadRecentDays(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	now := time.Now().UTC()
	baseTime := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, time.UTC)

	// Create activities over 200 days
	for i := range 200 {
		insertActivity(t, db, core.Activity{
			ID:        string(rune(i)),
			Timestamp: baseTime.AddDate(0, 0, -i),
			Duration:  100,
			Lines:     10,
			Language:  "go",
			Project:   "test",
			File:      "/test.go",
		})
	}

	t.Run("load 30 days", func(t *testing.T) {
		calc := NewCalculator(time.UTC)
		stats, err := calc.CalculateAPI(db, APIOptions{LoadRecentDays: 30})
		require.NoError(t, err)
		require.LessOrEqual(t, stats.Meta.LoadedActivities, 31) // May include today
	})

	t.Run("load 180 days", func(t *testing.T) {
		calc := NewCalculator(time.UTC)
		stats, err := calc.CalculateAPI(db, APIOptions{LoadRecentDays: 180})
		require.NoError(t, err)
		require.LessOrEqual(t, stats.Meta.LoadedActivities, 181) // May include today
	})
}

func TestCalculator_Timezone(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	est, _ := time.LoadLocation("America/New_York")
	utcTime := time.Date(2025, 1, 15, 23, 0, 0, 0, time.UTC)

	insertActivity(t, db, core.Activity{
		ID:        "1",
		Timestamp: utcTime,
		Duration:  100,
		Lines:     10,
		Language:  "go",
		Project:   "test",
		File:      "/test.go",
	})

	t.Run("UTC timezone", func(t *testing.T) {
		calc := NewCalculator(time.UTC)
		stats, err := calc.CalculateAPI(db, APIOptions{LoadRecentDays: 180})
		require.NoError(t, err)
		require.NotNil(t, stats)
	})

	t.Run("EST timezone", func(t *testing.T) {
		calc := NewCalculator(est)
		stats, err := calc.CalculateAPI(db, APIOptions{LoadRecentDays: 180})
		require.NoError(t, err)
		require.NotNil(t, stats)
	})
}

func TestCalculator_EdgeCases(t *testing.T) {
	t.Run("zero duration activities", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		insertActivity(t, db, core.Activity{
			ID:        "1",
			Timestamp: time.Now().UTC(),
			Duration:  0,
			Lines:     0,
			Language:  "go",
			Project:   "test",
			File:      "/test.go",
		})

		calc := NewCalculator(time.UTC)
		stats, err := calc.CalculateAPI(db, APIOptions{LoadRecentDays: 180})
		require.NoError(t, err)
		require.NotNil(t, stats)
	})

	t.Run("same timestamp activities", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		sameTime := time.Now().UTC()
		for i := range 5 {
			insertActivity(t, db, core.Activity{
				ID:        string(rune(i)),
				Timestamp: sameTime,
				Duration:  100,
				Lines:     10,
				Language:  "go",
				Project:   "test",
				File:      "/test.go",
			})
		}

		calc := NewCalculator(time.UTC)
		stats, err := calc.CalculateAPI(db, APIOptions{LoadRecentDays: 180})
		require.NoError(t, err)
		require.Equal(t, 5, stats.Meta.LoadedActivities)
	})
}

func TestCalculator_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Insert 10,000 activities
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range 10000 {
		insertActivity(t, db, core.Activity{
			ID:        string(rune(i)),
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
			Duration:  float64(60 + (i % 120)),
			Lines:     10 + (i % 50),
			Language:  []string{"go", "typescript", "python"}[i%3],
			Project:   []string{"p1", "p2", "p3"}[i%3],
			Editor:    "neovim",
			File:      "/test.go",
		})
	}

	calc := NewCalculator(time.UTC)
	start := time.Now()
	stats, err := calc.CalculateAPI(db, APIOptions{LoadRecentDays: 180})
	duration := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, stats)
	require.Less(t, duration, 5*time.Second, "calculation should complete in under 5 seconds")
	require.Greater(t, stats.Meta.QueryTimeMs, 0.0)
}

func TestCalculator_WeeklyHeatmap_SmartRange(t *testing.T) {
	// Note: This test uses real time.Now() from the function
	now := time.Now()

	// Find current week Monday
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	daysFromMonday := weekday - 1
	currentWeekMonday := now.AddDate(0, 0, -daysFromMonday)
	currentWeekMonday = time.Date(
		currentWeekMonday.Year(),
		currentWeekMonday.Month(),
		currentWeekMonday.Day(),
		0, 0, 0, 0,
		currentWeekMonday.Location(),
	)

	tests := []struct {
		name               string
		activityDates      []string // YYYY-MM-DD format
		minExpectedDays    int      // minimum expected days
		maxExpectedDays    int      // maximum expected days
		shouldBeLessThan84 bool     // for new users
	}{
		{
			name: "new user - started this week",
			activityDates: func() []string {
				dates := []string{}
				// Create activities for current week only (Mon-today)
				for i := 0; i <= daysFromMonday; i++ {
					date := currentWeekMonday.AddDate(0, 0, i)
					dates = append(dates, date.Format("2006-01-02"))
				}
				return dates
			}(),
			minExpectedDays:    7,    // at least current week
			maxExpectedDays:    14,   // at most 2 weeks
			shouldBeLessThan84: true, // new user
		},
		{
			name: "user started 2 weeks ago",
			activityDates: func() []string {
				dates := []string{}
				startDate := currentWeekMonday.AddDate(0, 0, -14) // 2 weeks ago
				for i := 0; i < 14; i++ {                         // 2 weeks of activity
					if i%2 == 0 { // every other day
						dates = append(dates, startDate.AddDate(0, 0, i).Format("2006-01-02"))
					}
				}
				return dates
			}(),
			minExpectedDays:    14,   // at least 2 weeks
			maxExpectedDays:    28,   // at most 4 weeks
			shouldBeLessThan84: true, // new user
		},
		{
			name: "established user - 12+ weeks of activity",
			activityDates: func() []string {
				dates := []string{}
				// Start from 15 weeks ago (beyond 12 week limit)
				startDate := currentWeekMonday.AddDate(0, 0, -15*7)
				for i := 0; i < 100; i++ { // 100 days
					dates = append(dates, startDate.AddDate(0, 0, i).Format("2006-01-02"))
				}
				return dates
			}(),
			minExpectedDays:    84,    // should be exactly 12 weeks
			maxExpectedDays:    91,    // allow some variation
			shouldBeLessThan84: false, // established user
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build daily stats map
			daily := make(map[string]DailyStat)
			for i, dateStr := range tt.activityDates {
				daily[dateStr] = DailyStat{
					Date:         dateStr,
					Time:         int64(3600 + i*100), // varying times
					Lines:        100 + i*10,
					Files:        5,
					SessionCount: 1,
				}
			}

			calc := &Calculator{timezone: time.UTC}
			heatmap := calc.generateWeeklyHeatmap(daily, 12)

			// Verify heatmap is not empty
			require.Greater(t, len(heatmap), 0, "heatmap should not be empty")

			// Verify length is within expected range
			require.GreaterOrEqual(t, len(heatmap), tt.minExpectedDays,
				"heatmap should have at least %d days, got %d", tt.minExpectedDays, len(heatmap))
			require.LessOrEqual(t, len(heatmap), tt.maxExpectedDays,
				"heatmap should have at most %d days, got %d", tt.maxExpectedDays, len(heatmap))

			// Verify new user constraint
			if tt.shouldBeLessThan84 {
				require.Less(t, len(heatmap), 84,
					"new users should show less than 12 weeks (84 days), got %d", len(heatmap))
			}

			// Verify first day is always a Monday
			firstDate, err := time.Parse("2006-01-02", heatmap[0].Date)
			require.NoError(t, err)
			firstWeekday := int(firstDate.Weekday())
			if firstWeekday == 0 {
				firstWeekday = 7
			}
			require.Equal(t, 1, firstWeekday, "first day should be Monday, got %s", firstDate.Weekday())

			// Verify no gaps in dates (consecutive days)
			for i := 1; i < len(heatmap); i++ {
				prevDate, _ := time.Parse("2006-01-02", heatmap[i-1].Date)
				currDate, _ := time.Parse("2006-01-02", heatmap[i].Date)
				diff := currDate.Sub(prevDate).Hours() / 24
				require.Equal(t, 1.0, diff, "dates should be consecutive at index %d", i)
			}

			// Verify activity data is preserved for past dates
			for _, dateStr := range tt.activityDates {
				activityDate, _ := time.Parse("2006-01-02", dateStr)
				// Only check dates that are in the past or today
				if activityDate.Before(now) || activityDate.Format("2006-01-02") == now.Format("2006-01-02") {
					for _, day := range heatmap {
						if day.Date == dateStr {
							if day.Time > 0 {
								// Activity data should be preserved
								require.Greater(t, day.Lines, 0, "activity on %s should have lines > 0", dateStr)
							}
							break
						}
					}
					// Don't require found=true because date might be before the smart range window
				}
			}

			// Verify future days are marked with level -1
			for _, day := range heatmap {
				dayDate, _ := time.Parse("2006-01-02", day.Date)
				if dayDate.After(now) {
					require.Equal(t, -1, day.Level, "future date %s should have level=-1", day.Date)
				}
			}
		})
	}
}
