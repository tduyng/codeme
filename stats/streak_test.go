package stats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tduyng/codeme/core"
)

func TestStreakCalculator_Calculate(t *testing.T) {
	// Use a fixed date in the past so tests are deterministic
	baseDate := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		activities      []core.Activity
		expectedCurrent int
		expectedActive  bool
	}{
		{
			name:            "empty activities",
			activities:      []core.Activity{},
			expectedCurrent: 0,
			expectedActive:  false,
		},
		{
			name: "single day",
			activities: []core.Activity{
				{Timestamp: baseDate, Duration: 100},
			},
			expectedCurrent: 1,
			expectedActive:  false,
		},
		{
			name: "consecutive days",
			activities: []core.Activity{
				{Timestamp: baseDate, Duration: 100},
				{Timestamp: baseDate.AddDate(0, 0, -1), Duration: 100},
				{Timestamp: baseDate.AddDate(0, 0, -2), Duration: 100},
			},
			expectedCurrent: 3,
			expectedActive:  false,
		},
		{
			name: "broken streak",
			activities: []core.Activity{
				{Timestamp: baseDate, Duration: 100},
				{Timestamp: baseDate.AddDate(0, 0, -1), Duration: 100},
				{Timestamp: baseDate.AddDate(0, 0, -3), Duration: 100},
				{Timestamp: baseDate.AddDate(0, 0, -4), Duration: 100},
			},
			expectedCurrent: 2,
			expectedActive:  false,
		},
		{
			name: "multiple activities same day",
			activities: []core.Activity{
				{Timestamp: baseDate, Duration: 100},
				{Timestamp: baseDate.Add(2 * time.Hour), Duration: 100},
				{Timestamp: baseDate.AddDate(0, 0, -1), Duration: 100},
			},
			expectedCurrent: 2,
			expectedActive:  false,
		},
		{
			name: "long streak",
			activities: func() []core.Activity {
				var acts []core.Activity
				for i := 0; i < 30; i++ {
					acts = append(acts, core.Activity{
						Timestamp: baseDate.AddDate(0, 0, -i),
						Duration:  100,
					})
				}
				return acts
			}(),
			expectedCurrent: 30,
			expectedActive:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewStreakCalculator(time.UTC)
			result := calc.Calculate(tt.activities)

			require.Equal(t, tt.expectedCurrent, result.Current, "current streak mismatch")
			require.GreaterOrEqual(t, result.Longest, result.Current, "longest should be >= current")
			require.Equal(t, tt.expectedActive, result.IsActive, "active status mismatch")

			if len(tt.activities) > 0 {
				require.NotZero(t, result.LastActivity)
			}
		})
	}
}

func TestStreakCalculator_ActiveStreak(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)

	tests := []struct {
		name           string
		activities     []core.Activity
		expectedActive bool
	}{
		{
			name: "active today",
			activities: []core.Activity{
				{Timestamp: today, Duration: 100},
			},
			expectedActive: true,
		},
		{
			name: "active yesterday",
			activities: []core.Activity{
				{Timestamp: yesterday, Duration: 100},
			},
			expectedActive: true,
		},
		{
			name: "not active",
			activities: []core.Activity{
				{Timestamp: today.AddDate(0, 0, -2), Duration: 100},
			},
			expectedActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewStreakCalculator(time.UTC)
			result := calc.Calculate(tt.activities)

			require.Equal(t, tt.expectedActive, result.IsActive)
		})
	}
}

func TestStreakCalculator_Timezone(t *testing.T) {
	est, _ := time.LoadLocation("America/New_York")
	baseDate := time.Date(2025, 1, 15, 12, 0, 0, 0, est)

	activities := []core.Activity{
		{Timestamp: baseDate, Duration: 100},
		{Timestamp: baseDate.AddDate(0, 0, -1), Duration: 100},
		{Timestamp: baseDate.AddDate(0, 0, -2), Duration: 100},
	}

	t.Run("UTC timezone", func(t *testing.T) {
		calc := NewStreakCalculator(time.UTC)
		result := calc.Calculate(activities)
		require.Greater(t, result.Current, 0)
	})

	t.Run("EST timezone", func(t *testing.T) {
		calc := NewStreakCalculator(est)
		result := calc.Calculate(activities)
		require.Equal(t, 3, result.Current)
	})

	t.Run("nil timezone defaults to UTC", func(t *testing.T) {
		calc := NewStreakCalculator(nil)
		result := calc.Calculate(activities)
		require.Greater(t, result.Current, 0)
	})
}

func TestStreakCalculator_EdgeCases(t *testing.T) {
	t.Run("midnight boundary", func(t *testing.T) {
		midnight := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		justBefore := time.Date(2025, 1, 14, 23, 59, 59, 0, time.UTC)

		activities := []core.Activity{
			{Timestamp: midnight, Duration: 100},
			{Timestamp: justBefore, Duration: 100},
		}

		calc := NewStreakCalculator(time.UTC)
		result := calc.Calculate(activities)

		require.Equal(t, 2, result.Current)
	})

	t.Run("unsorted activities", func(t *testing.T) {
		baseDate := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

		activities := []core.Activity{
			{Timestamp: baseDate.AddDate(0, 0, -2), Duration: 100},
			{Timestamp: baseDate, Duration: 100},
			{Timestamp: baseDate.AddDate(0, 0, -1), Duration: 100},
		}

		calc := NewStreakCalculator(time.UTC)
		result := calc.Calculate(activities)

		require.Equal(t, 3, result.Current)
	})

	t.Run("year boundary", func(t *testing.T) {
		newYear := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
		oldYear := time.Date(2024, 12, 31, 12, 0, 0, 0, time.UTC)

		activities := []core.Activity{
			{Timestamp: newYear, Duration: 100},
			{Timestamp: oldYear, Duration: 100},
		}

		calc := NewStreakCalculator(time.UTC)
		result := calc.Calculate(activities)

		require.Equal(t, 2, result.Current)
	})

	t.Run("maximum historic streak", func(t *testing.T) {
		baseDate := time.Date(2025, 1, 31, 12, 0, 0, 0, time.UTC)

		// Create 400 days of activities (more than 365 lookback)
		var activities []core.Activity
		for i := range 400 {
			activities = append(activities, core.Activity{
				Timestamp: baseDate.AddDate(0, 0, -i),
				Duration:  100,
			})
		}

		calc := NewStreakCalculator(time.UTC)
		result := calc.Calculate(activities)

		// Should calculate longest within 365 day window
		require.LessOrEqual(t, result.Longest, 365)
	})
}

func TestStreakCalculator_LongestVsCurrent(t *testing.T) {
	baseDate := time.Date(2025, 1, 31, 12, 0, 0, 0, time.UTC)

	// Create pattern: 5 days on, 2 days off, 10 days on
	activities := []core.Activity{
		// Current streak: 10 days
		{Timestamp: baseDate, Duration: 100},
		{Timestamp: baseDate.AddDate(0, 0, -1), Duration: 100},
		{Timestamp: baseDate.AddDate(0, 0, -2), Duration: 100},
		{Timestamp: baseDate.AddDate(0, 0, -3), Duration: 100},
		{Timestamp: baseDate.AddDate(0, 0, -4), Duration: 100},
		{Timestamp: baseDate.AddDate(0, 0, -5), Duration: 100},
		{Timestamp: baseDate.AddDate(0, 0, -6), Duration: 100},
		{Timestamp: baseDate.AddDate(0, 0, -7), Duration: 100},
		{Timestamp: baseDate.AddDate(0, 0, -8), Duration: 100},
		{Timestamp: baseDate.AddDate(0, 0, -9), Duration: 100},
		// Gap of 2 days
		// Previous streak: 5 days
		{Timestamp: baseDate.AddDate(0, 0, -12), Duration: 100},
		{Timestamp: baseDate.AddDate(0, 0, -13), Duration: 100},
		{Timestamp: baseDate.AddDate(0, 0, -14), Duration: 100},
		{Timestamp: baseDate.AddDate(0, 0, -15), Duration: 100},
		{Timestamp: baseDate.AddDate(0, 0, -16), Duration: 100},
	}

	calc := NewStreakCalculator(time.UTC)
	result := calc.Calculate(activities)

	require.Equal(t, 10, result.Current)
	require.GreaterOrEqual(t, result.Longest, 10) // Longest is at least as long as current
}
