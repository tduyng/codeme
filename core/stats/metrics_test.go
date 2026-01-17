package stats

import (
	"strings"
	"testing"
)

func TestCalculateFocusScore(t *testing.T) {
	tests := []struct {
		name     string
		sessions []Session
		validate func(t *testing.T, score int)
	}{
		{
			name:     "no sessions",
			sessions: []Session{},
			validate: func(t *testing.T, score int) {
				if score != 0 {
					t.Errorf("Expected 0 for no sessions, got %d", score)
				}
			},
		},
		{
			name: "long focused sessions",
			sessions: []Session{
				{Duration: 7200}, // 2 hours
				{Duration: 7200},
			},
			validate: func(t *testing.T, score int) {
				if score < 80 {
					t.Errorf("Expected high score for long sessions, got %d", score)
				}
			},
		},
		{
			name: "short fragmented sessions",
			sessions: []Session{
				{Duration: 600}, // 10 minutes
				{Duration: 600},
				{Duration: 600},
				{Duration: 600},
				{Duration: 600},
				{Duration: 600},
				{Duration: 600},
				{Duration: 600},
			},
			validate: func(t *testing.T, score int) {
				if score > 50 {
					t.Errorf("Expected low score for fragmented sessions, got %d", score)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateFocusScore(tt.sessions)
			tt.validate(t, result)
		})
	}
}

func TestCalculateProductivityTrend(t *testing.T) {
	tests := []struct {
		name          string
		dailyActivity map[string]DailyStat
		thisWeekStart string
		lastWeekStart string
		lastWeekEnd   string
		expectedTrend string
	}{
		{
			name:          "no data",
			dailyActivity: map[string]DailyStat{},
			thisWeekStart: "2026-01-13",
			lastWeekStart: "2026-01-06",
			lastWeekEnd:   "2026-01-12",
			expectedTrend: "📊 Building Data",
		},
		{
			name: "improving trend",
			dailyActivity: map[string]DailyStat{
				"2026-01-13": {Time: 10000},
				"2026-01-14": {Time: 10000},
				"2026-01-06": {Time: 5000},
			},
			thisWeekStart: "2026-01-13",
			lastWeekStart: "2026-01-06",
			lastWeekEnd:   "2026-01-12",
			expectedTrend: "↗ Improving",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateProductivityTrend(tt.dailyActivity, tt.thisWeekStart, tt.lastWeekStart, tt.lastWeekEnd)
			if result != tt.expectedTrend {
				t.Errorf("CalculateProductivityTrend() = %s, want %s", result, tt.expectedTrend)
			}
		})
	}
}

func TestCalculateGrowth(t *testing.T) {
	tests := []struct {
		name     string
		validate func(t *testing.T, growth string)
	}{
		{
			name: "new project",
			validate: func(t *testing.T, growth string) {
				if !strings.Contains(growth, "New") {
					t.Errorf("Expected 'New' indicator, got %s", growth)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dailyActivity := map[string]DailyStat{
				"2026-01-13": {Time: 10000},
			}
			result := CalculateGrowth("test", dailyActivity, "2026-01-13", "2026-01-06", "2026-01-12")
			tt.validate(t, result)
		})
	}
}
