package core

import (
	"testing"
	"time"
)

func TestEndToEnd(t *testing.T) {
	// Create test database
	db, err := OpenDB()
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Clear any existing data
	db.Exec("DELETE FROM heartbeats")

	// Simulate a coding session
	session := []struct {
		file         string
		language     string
		linesChanged int
		linesTotal   int
		delay        time.Duration
	}{
		{"/project/main.go", "Go", 50, 50, 0},
		{"/project/utils.go", "Go", 30, 30, 5 * time.Minute},
		{"/project/config.lua", "Lua", 20, 20, 10 * time.Minute},
		{"/project/README.md", "Markdown", 15, 15, 30 * time.Minute}, // New session (>15min gap)
		{"/project/main.go", "Go", 10, 60, 32 * time.Minute},         // Same session as README
	}

	// Track all files with proper timing
	baseTime := time.Now()
	for _, item := range session {
		hb := Heartbeat{
			Timestamp:    baseTime.Add(item.delay),
			File:         item.file,
			Language:     item.language,
			Project:      "project",
			Lines:        item.linesChanged,
			LinesChanged: item.linesChanged,
			LinesTotal:   item.linesTotal,
		}
		if err := SaveHeartbeat(db, hb); err != nil {
			t.Fatalf("Failed to save heartbeat: %v", err)
		}
	}

	// Calculate stats
	stats, err := CalculateStats(db, false)
	if err != nil {
		t.Fatalf("CalculateStats() error = %v", err)
	}

	// Verify total lines (sum of lines_changed)
	expectedTotalLines := 50 + 30 + 20 + 15 + 10 // 125
	if stats.TotalLines != expectedTotalLines {
		t.Errorf("TotalLines = %d, want %d", stats.TotalLines, expectedTotalLines)
	}

	// Verify languages
	expectedLanguages := map[string]int{
		"Go":       90, // 50 + 30 + 10
		"Lua":      20,
		"Markdown": 15,
	}

	for lang, expectedLines := range expectedLanguages {
		if langStat, ok := stats.Languages[lang]; !ok {
			t.Errorf("Language %s not found in stats", lang)
		} else if langStat.Lines != expectedLines {
			t.Errorf("Language %s lines = %d, want %d", lang, langStat.Lines, expectedLines)
		}
	}

	// Verify project stats
	if projectStat, ok := stats.Projects["project"]; !ok {
		t.Error("Project 'project' not found in stats")
	} else {
		if projectStat.Lines != expectedTotalLines {
			t.Errorf("Project lines = %d, want %d", projectStat.Lines, expectedTotalLines)
		}
	}

	// Verify session time calculation
	// Heartbeats: 0min, 5min, 10min, 30min, 32min
	// Session 1: 0->5min (300s) + 5->10min (300s) = 600s (gap 10->30 is 20min >15min, ends session)
	// Session 2: 30->32min (120s) = 120s
	// Total: 720s
	expectedTime := int64(720)
	tolerance := int64(120) // Allow 2 minute tolerance

	if stats.TotalTime < expectedTime-tolerance || stats.TotalTime > expectedTime+tolerance {
		t.Logf("Session breakdown: actual=%d, expected=%d", stats.TotalTime, expectedTime)
		t.Errorf("TotalTime = %d, want ~%d (±%d)", stats.TotalTime, expectedTime, tolerance)
	}

	// Verify streak (should be 1 since all activity is today)
	if stats.Streak != 1 {
		t.Errorf("Streak = %d, want 1", stats.Streak)
	}

	// Verify today's stats
	if stats.TodayLines != expectedTotalLines {
		t.Errorf("TodayLines = %d, want %d", stats.TodayLines, expectedTotalLines)
	}

	// Clean up
	db.Exec("DELETE FROM heartbeats")
}

func TestMultipleDayStats(t *testing.T) {
	db, err := OpenDB()
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Clear any existing data
	db.Exec("DELETE FROM heartbeats")

	// Simulate activity over 3 consecutive days
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())

	multiDayData := []struct {
		day   time.Time
		file  string
		lang  string
		lines int
	}{
		{today.AddDate(0, 0, -2), "/test/day1.go", "Go", 100},
		{today.AddDate(0, 0, -1), "/test/day2.go", "Go", 75},
		{today, "/test/day3.go", "Go", 50},
	}

	for _, data := range multiDayData {
		hb := Heartbeat{
			Timestamp:    data.day,
			File:         data.file,
			Language:     data.lang,
			Project:      "multiday",
			Lines:        data.lines,
			LinesChanged: data.lines,
			LinesTotal:   data.lines,
		}
		if err := SaveHeartbeat(db, hb); err != nil {
			t.Fatalf("Failed to save heartbeat: %v", err)
		}
	}

	// Calculate stats
	stats, err := CalculateStats(db, false)
	if err != nil {
		t.Fatalf("CalculateStats() error = %v", err)
	}

	// Verify total lines across all days
	if stats.TotalLines != 225 { // 100 + 75 + 50
		t.Errorf("TotalLines = %d, want 225", stats.TotalLines)
	}

	// Verify today's lines
	if stats.TodayLines != 50 {
		t.Errorf("TodayLines = %d, want 50", stats.TodayLines)
	}

	// Verify streak (3 consecutive days)
	if stats.Streak != 3 {
		t.Errorf("Streak = %d, want 3", stats.Streak)
	}

	// Verify daily activity map
	if len(stats.DailyActivity) != 3 {
		t.Errorf("DailyActivity count = %d, want 3", len(stats.DailyActivity))
	}

	// Clean up
	db.Exec("DELETE FROM heartbeats")
}
