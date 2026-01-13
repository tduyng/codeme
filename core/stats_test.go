package core

import (
	"testing"
	"time"
)

func TestCalculateStreak(t *testing.T) {
	tests := []struct {
		name     string
		daily    map[string]DailyStat
		expected int
	}{
		{
			name:     "No activity",
			daily:    map[string]DailyStat{},
			expected: 0,
		},
		{
			name: "Current streak of 3 days",
			daily: map[string]DailyStat{
				time.Now().Format("2006-01-02"):                   {Lines: 100, Time: 3600},
				time.Now().AddDate(0, 0, -1).Format("2006-01-02"): {Lines: 50, Time: 1800},
				time.Now().AddDate(0, 0, -2).Format("2006-01-02"): {Lines: 75, Time: 2400},
				time.Now().AddDate(0, 0, -4).Format("2006-01-02"): {Lines: 25, Time: 600},
			},
			expected: 3, // Today + 2 previous consecutive days
		},
		{
			name: "Broken streak",
			daily: map[string]DailyStat{
				time.Now().AddDate(0, 0, -2).Format("2006-01-02"): {Lines: 100, Time: 3600},
				time.Now().AddDate(0, 0, -3).Format("2006-01-02"): {Lines: 50, Time: 1800},
			},
			expected: 0, // No activity today or yesterday, streak is broken
		},
		{
			name: "Today only",
			daily: map[string]DailyStat{
				time.Now().Format("2006-01-02"): {Lines: 50, Time: 1800},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateStreak(tt.daily)
			if result != tt.expected {
				t.Errorf("calculateStreak() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestCalculateLongestStreak(t *testing.T) {
	tests := []struct {
		name     string
		daily    map[string]DailyStat
		expected int
	}{
		{
			name:     "No activity",
			daily:    map[string]DailyStat{},
			expected: 0,
		},
		{
			name: "Longest streak is 5 days",
			daily: map[string]DailyStat{
				// Current streak of 2
				time.Now().Format("2006-01-02"):                   {Lines: 10, Time: 600},
				time.Now().AddDate(0, 0, -1).Format("2006-01-02"): {Lines: 20, Time: 1200},
				// Gap
				// Previous streak of 5 (longest)
				time.Now().AddDate(0, 0, -5).Format("2006-01-02"): {Lines: 30, Time: 1800},
				time.Now().AddDate(0, 0, -6).Format("2006-01-02"): {Lines: 40, Time: 2400},
				time.Now().AddDate(0, 0, -7).Format("2006-01-02"): {Lines: 50, Time: 3000},
				time.Now().AddDate(0, 0, -8).Format("2006-01-02"): {Lines: 60, Time: 3600},
				time.Now().AddDate(0, 0, -9).Format("2006-01-02"): {Lines: 70, Time: 4200},
			},
			expected: 5,
		},
		{
			name: "Single day",
			daily: map[string]DailyStat{
				time.Now().Format("2006-01-02"): {Lines: 100, Time: 7200},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateLongestStreak(tt.daily)
			if result != tt.expected {
				t.Errorf("calculateLongestStreak() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestCalculateStats(t *testing.T) {
	// Create test database
	db := setupTestDB(t)
	defer db.Close()

	// Insert test data
	now := time.Now()
	testData := []Heartbeat{
		{
			Timestamp:    now,
			File:         "/test/file1.go",
			Language:     "Go",
			Project:      "testproject",
			Lines:        50,
			LinesChanged: 50,
			LinesTotal:   100,
		},
		{
			Timestamp:    now.Add(5 * time.Minute), // Same session (within 15 min)
			File:         "/test/file2.go",
			Language:     "Go",
			Project:      "testproject",
			Lines:        30,
			LinesChanged: 30,
			LinesTotal:   80,
		},
		{
			Timestamp:    now.Add(20 * time.Minute), // New session (>15 min gap)
			File:         "/test/file3.lua",
			Language:     "Lua",
			Project:      "anotherproject",
			Lines:        20,
			LinesChanged: 20,
			LinesTotal:   60,
		},
	}

	for _, hb := range testData {
		if err := SaveHeartbeat(db, hb); err != nil {
			t.Fatalf("Failed to save heartbeat: %v", err)
		}
	}

	// Calculate stats
	stats, err := CalculateStats(db, false)
	if err != nil {
		t.Fatalf("CalculateStats() error = %v", err)
	}

	// Verify total lines
	expectedLines := 100 // 50 + 30 + 20
	if stats.TotalLines != expectedLines {
		t.Errorf("TotalLines = %d, want %d", stats.TotalLines, expectedLines)
	}

	// Verify languages
	if len(stats.Languages) != 2 {
		t.Errorf("Languages count = %d, want 2", len(stats.Languages))
	}

	if stats.Languages["Go"].Lines != 80 { // 50 + 30
		t.Errorf("Go lines = %d, want 80", stats.Languages["Go"].Lines)
	}

	if stats.Languages["Lua"].Lines != 20 {
		t.Errorf("Lua lines = %d, want 20", stats.Languages["Lua"].Lines)
	}

	// Verify projects
	if len(stats.Projects) != 2 {
		t.Errorf("Projects count = %d, want 2", len(stats.Projects))
	}

	// Verify session time
	// Heartbeat 1 at 0min, Heartbeat 2 at 5min, Heartbeat 3 at 20min
	// Gaps: 5min (300s) + 15min (900s) = 20min (1200s)
	// Both gaps are within 15min threshold, so all count as same session
	expectedTime := int64(1200)
	tolerance := int64(60) // Allow 1 minute tolerance

	if stats.TotalTime < expectedTime-tolerance || stats.TotalTime > expectedTime+tolerance {
		t.Errorf("TotalTime = %d, want ~%d (±%d)", stats.TotalTime, expectedTime, tolerance)
	}
}
