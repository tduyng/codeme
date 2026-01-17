package stats

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
			name:     "no activity",
			daily:    map[string]DailyStat{},
			expected: 0,
		},
		{
			name: "current streak of 3 days",
			daily: map[string]DailyStat{
				time.Now().Format("2006-01-02"):                   {Lines: 100, Time: 3600},
				time.Now().AddDate(0, 0, -1).Format("2006-01-02"): {Lines: 50, Time: 1800},
				time.Now().AddDate(0, 0, -2).Format("2006-01-02"): {Lines: 75, Time: 2400},
				time.Now().AddDate(0, 0, -4).Format("2006-01-02"): {Lines: 25, Time: 600},
			},
			expected: 3,
		},
		{
			name: "broken streak",
			daily: map[string]DailyStat{
				time.Now().AddDate(0, 0, -2).Format("2006-01-02"): {Lines: 100, Time: 3600},
				time.Now().AddDate(0, 0, -3).Format("2006-01-02"): {Lines: 50, Time: 1800},
			},
			expected: 0,
		},
		{
			name: "today only",
			daily: map[string]DailyStat{
				time.Now().Format("2006-01-02"): {Lines: 50, Time: 1800},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateStreak(tt.daily)
			if result != tt.expected {
				t.Errorf("CalculateStreak() = %d, want %d", result, tt.expected)
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
			name:     "no activity",
			daily:    map[string]DailyStat{},
			expected: 0,
		},
		{
			name: "longest streak is 5 days",
			daily: map[string]DailyStat{
				time.Now().Format("2006-01-02"):                   {Lines: 10, Time: 600},
				time.Now().AddDate(0, 0, -1).Format("2006-01-02"): {Lines: 20, Time: 1200},
				time.Now().AddDate(0, 0, -5).Format("2006-01-02"): {Lines: 30, Time: 1800},
				time.Now().AddDate(0, 0, -6).Format("2006-01-02"): {Lines: 40, Time: 2400},
				time.Now().AddDate(0, 0, -7).Format("2006-01-02"): {Lines: 50, Time: 3000},
				time.Now().AddDate(0, 0, -8).Format("2006-01-02"): {Lines: 60, Time: 3600},
				time.Now().AddDate(0, 0, -9).Format("2006-01-02"): {Lines: 70, Time: 4200},
			},
			expected: 5,
		},
		{
			name: "single day",
			daily: map[string]DailyStat{
				time.Now().Format("2006-01-02"): {Lines: 100, Time: 7200},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateLongestStreak(tt.daily)
			if result != tt.expected {
				t.Errorf("CalculateLongestStreak() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestCalculateStreakInfo(t *testing.T) {
	tests := []struct {
		name     string
		daily    map[string]DailyStat
		validate func(t *testing.T, info StreakInfo)
	}{
		{
			name:  "empty data",
			daily: map[string]DailyStat{},
			validate: func(t *testing.T, info StreakInfo) {
				if info.Current != 0 || info.Longest != 0 {
					t.Errorf("Expected zero values for empty data")
				}
			},
		},
		{
			name: "active streak",
			daily: map[string]DailyStat{
				time.Now().Format("2006-01-02"):                   {Lines: 100, Time: 3600},
				time.Now().AddDate(0, 0, -1).Format("2006-01-02"): {Lines: 50, Time: 1800},
			},
			validate: func(t *testing.T, info StreakInfo) {
				if info.Current != 2 {
					t.Errorf("Current streak = %d, want 2", info.Current)
				}
				if info.StartDate == "" {
					t.Error("StartDate should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateStreakInfo(tt.daily)
			tt.validate(t, result)
		})
	}
}
