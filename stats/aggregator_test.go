package stats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tduyng/codeme/core"
)

func TestAggregateByDay(t *testing.T) {
	baseDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		activities []core.Activity
		tz         *time.Location
		check      func(t *testing.T, result map[string]*DayAgg)
	}{
		{
			name:       "empty",
			activities: []core.Activity{},
			tz:         time.UTC,
			check: func(t *testing.T, result map[string]*DayAgg) {
				require.Empty(t, result)
			},
		},
		{
			name: "single day",
			activities: []core.Activity{
				{Timestamp: baseDate.Add(10 * time.Hour), Duration: 100, Lines: 10, File: "/a", Language: "go", Project: "p1"},
				{Timestamp: baseDate.Add(14 * time.Hour), Duration: 200, Lines: 20, File: "/b", Language: "ts", Project: "p2"},
			},
			tz: time.UTC,
			check: func(t *testing.T, result map[string]*DayAgg) {
				require.Len(t, result, 1)
				require.Contains(t, result, "2025-01-15")
			},
		},
		{
			name: "multiple days",
			activities: []core.Activity{
				{Timestamp: baseDate, Duration: 100, Lines: 10, File: "/a", Language: "go", Project: "p1"},
				{Timestamp: baseDate.AddDate(0, 0, 1), Duration: 200, Lines: 20, File: "/b", Language: "ts", Project: "p2"},
				{Timestamp: baseDate.AddDate(0, 0, 2), Duration: 150, Lines: 15, File: "/c", Language: "py", Project: "p3"},
			},
			tz: time.UTC,
			check: func(t *testing.T, result map[string]*DayAgg) {
				require.Len(t, result, 3)
			},
		},
		{
			name: "midnight boundary",
			activities: []core.Activity{
				{Timestamp: baseDate.Add(23*time.Hour + 59*time.Minute), Duration: 100, Lines: 10, File: "/a"},
				{Timestamp: baseDate.AddDate(0, 0, 1), Duration: 100, Lines: 10, File: "/b"},
			},
			tz: time.UTC,
			check: func(t *testing.T, result map[string]*DayAgg) {
				require.Len(t, result, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AggregateByDay(tt.activities, tt.tz)
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestAggregateByHour(t *testing.T) {
	baseDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	activities := []core.Activity{
		{Timestamp: baseDate.Add(9 * time.Hour), Duration: 100},
		{Timestamp: baseDate.Add(9*time.Hour + 30*time.Minute), Duration: 200},
		{Timestamp: baseDate.Add(14 * time.Hour), Duration: 150},
		{Timestamp: baseDate.Add(23 * time.Hour), Duration: 50},
	}

	result := AggregateByHour(activities, time.UTC)

	require.Len(t, result, 24)
	require.Equal(t, 300.0, result[9].Duration)
	require.Equal(t, 150.0, result[14].Duration)
	require.Equal(t, 50.0, result[23].Duration)
	require.Equal(t, 0.0, result[0].Duration)
}
