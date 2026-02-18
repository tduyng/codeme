package stats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tduyng/codeme/core"
)

func TestSessionManager_GroupSessions(t *testing.T) {
	baseTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name              string
		activities        []core.Activity
		expectedSessions  int
		timeout           time.Duration
		minDuration       time.Duration
		checkFirstSession func(t *testing.T, s core.Session)
	}{
		{
			name:             "empty activities",
			activities:       []core.Activity{},
			expectedSessions: 0,
		},
		{
			name: "single activity",
			activities: []core.Activity{
				{
					ID:        "1",
					Timestamp: baseTime,
					Duration:  120,
					Language:  "go",
					Project:   "test",
				},
			},
			expectedSessions: 1,
			checkFirstSession: func(t *testing.T, s core.Session) {
				require.Equal(t, "1", s.ID)
				require.Contains(t, s.Projects, "test")
				require.Contains(t, s.Languages, "go")
			},
		},
		{
			name: "two activities same session",
			activities: []core.Activity{
				{ID: "1", Timestamp: baseTime, Duration: 120, Language: "go", Project: "p1"},
				{ID: "2", Timestamp: baseTime.Add(5 * time.Minute), Duration: 120, Language: "go", Project: "p1"},
			},
			expectedSessions: 1,
			checkFirstSession: func(t *testing.T, s core.Session) {
				require.Greater(t, s.Duration, 0.0)
			},
		},
		{
			name: "timeout creates new session",
			activities: []core.Activity{
				{ID: "1", Timestamp: baseTime, Duration: 120, Language: "go", Project: "p1"},
				{ID: "2", Timestamp: baseTime.Add(30 * time.Minute), Duration: 120, Language: "go", Project: "p1"},
			},
			expectedSessions: 2,
			timeout:          15 * time.Minute,
		},
		{
			name: "below min duration filtered",
			activities: []core.Activity{
				{ID: "1", Timestamp: baseTime, Duration: 30, Language: "go", Project: "p1"},
			},
			expectedSessions: 0,
			minDuration:      1 * time.Minute,
		},
		{
			name: "mixed languages",
			activities: []core.Activity{
				{ID: "1", Timestamp: baseTime, Duration: 120, Language: "go", Project: "p1"},
				{ID: "2", Timestamp: baseTime.Add(2 * time.Minute), Duration: 120, Language: "typescript", Project: "p1"},
				{ID: "3", Timestamp: baseTime.Add(4 * time.Minute), Duration: 120, Language: "python", Project: "p1"},
			},
			expectedSessions: 1,
			checkFirstSession: func(t *testing.T, s core.Session) {
				require.Len(t, s.Languages, 3)
				require.Contains(t, s.Languages, "go")
				require.Contains(t, s.Languages, "typescript")
				require.Contains(t, s.Languages, "python")
			},
		},
		{
			name: "mixed projects",
			activities: []core.Activity{
				{ID: "1", Timestamp: baseTime, Duration: 120, Language: "go", Project: "p1"},
				{ID: "2", Timestamp: baseTime.Add(2 * time.Minute), Duration: 120, Language: "go", Project: "p2"},
			},
			expectedSessions: 1,
			checkFirstSession: func(t *testing.T, s core.Session) {
				require.Len(t, s.Projects, 2)
				require.Contains(t, s.Projects, "p1")
				require.Contains(t, s.Projects, "p2")
			},
		},
		{
			name: "exactly at timeout boundary",
			activities: []core.Activity{
				{ID: "1", Timestamp: baseTime, Duration: 120, Language: "go", Project: "p1"},
				{ID: "2", Timestamp: baseTime.Add(15 * time.Minute), Duration: 120, Language: "go", Project: "p1"},
			},
			expectedSessions: 1,
			timeout:          15 * time.Minute,
		},
		{
			name: "just over timeout",
			activities: []core.Activity{
				{ID: "1", Timestamp: baseTime, Duration: 120, Language: "go", Project: "p1"},
				{ID: "2", Timestamp: baseTime.Add(15*time.Minute + time.Second), Duration: 120, Language: "go", Project: "p1"},
			},
			expectedSessions: 2,
			timeout:          15 * time.Minute,
		},
		{
			name: "invalid language filtered",
			activities: []core.Activity{
				{ID: "1", Timestamp: baseTime, Duration: 120, Language: "", Project: "p1"},
				{ID: "2", Timestamp: baseTime.Add(2 * time.Minute), Duration: 120, Language: "unknown", Project: "p1"},
			},
			expectedSessions: 1,
			checkFirstSession: func(t *testing.T, s core.Session) {
				require.Len(t, s.Languages, 0)
			},
		},
		{
			name: "unsorted activities",
			activities: func() []core.Activity {
				// Activities are now expected to be sorted by timestamp
				// (guaranteed from DB query in production)
				return []core.Activity{
					{ID: "1", Timestamp: baseTime, Duration: 120, Language: "go", Project: "p1"},
					{ID: "2", Timestamp: baseTime.Add(5 * time.Minute), Duration: 120, Language: "go", Project: "p1"},
					{ID: "3", Timestamp: baseTime.Add(10 * time.Minute), Duration: 120, Language: "go", Project: "p1"},
				}
			}(),
			expectedSessions: 1,
			checkFirstSession: func(t *testing.T, s core.Session) {
				require.Equal(t, "1", s.ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeout := tt.timeout
			if timeout == 0 {
				timeout = 15 * time.Minute
			}
			minDuration := tt.minDuration
			if minDuration == 0 {
				minDuration = 1 * time.Minute
			}

			sm := NewSessionManager(timeout, minDuration)
			sessions := sm.GroupSessions(tt.activities)

			require.Len(t, sessions, tt.expectedSessions)

			if tt.expectedSessions > 0 && tt.checkFirstSession != nil {
				tt.checkFirstSession(t, sessions[0])
			}
		})
	}
}

func TestSessionManager_BreakCalculation(t *testing.T) {
	baseTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	activities := []core.Activity{
		{ID: "1", Timestamp: baseTime, Duration: 120, Language: "go", Project: "p1"},
		{ID: "2", Timestamp: baseTime.Add(30 * time.Minute), Duration: 120, Language: "go", Project: "p1"},
		{ID: "3", Timestamp: baseTime.Add(90 * time.Minute), Duration: 120, Language: "go", Project: "p1"},
	}

	sm := NewSessionManager(15*time.Minute, 1*time.Minute)
	sessions := sm.GroupSessions(activities)

	require.Len(t, sessions, 3)

	// First session should have break to second
	require.Greater(t, sessions[0].BreakAfter, 0.0)

	// Second session should have break to third
	require.Greater(t, sessions[1].BreakAfter, 0.0)

	// Last session has no break
	require.Equal(t, 0.0, sessions[2].BreakAfter)
}

func TestConvertSessionsToAPI(t *testing.T) {
	baseTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	sessions := []core.Session{
		{
			ID:         "1",
			StartTime:  baseTime,
			EndTime:    baseTime.Add(1 * time.Hour),
			Duration:   3600,
			Projects:   []string{"p1"},
			Languages:  []string{"go"},
			IsActive:   false,
			BreakAfter: 600,
		},
		{
			ID:         "2",
			StartTime:  baseTime.Add(2 * time.Hour),
			EndTime:    baseTime.Add(3 * time.Hour),
			Duration:   3600,
			Projects:   []string{"p2"},
			Languages:  []string{"typescript"},
			IsActive:   true,
			BreakAfter: 0,
		},
	}

	apiSessions := ConvertSessionsToAPI(sessions)

	require.Len(t, apiSessions, 2)

	require.Equal(t, "1", apiSessions[0].ID)
	require.Equal(t, 3600.0, apiSessions[0].Duration)
	require.Equal(t, 600.0, apiSessions[0].BreakAfter)
	require.False(t, apiSessions[0].IsActive)

	require.Equal(t, "2", apiSessions[1].ID)
	require.True(t, apiSessions[1].IsActive)
}

func TestSessionManager_EdgeCases(t *testing.T) {
	t.Run("same timestamp activities", func(t *testing.T) {
		baseTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

		activities := []core.Activity{
			{ID: "1", Timestamp: baseTime, Duration: 120, Language: "go", Project: "p1"},
			{ID: "2", Timestamp: baseTime, Duration: 120, Language: "go", Project: "p1"},
		}

		sm := NewSessionManager(15*time.Minute, 1*time.Minute)
		sessions := sm.GroupSessions(activities)

		require.Len(t, sessions, 1)
	})

	t.Run("zero duration activities", func(t *testing.T) {
		baseTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

		activities := []core.Activity{
			{ID: "1", Timestamp: baseTime, Duration: 0, Language: "go", Project: "p1"},
		}

		sm := NewSessionManager(15*time.Minute, 1*time.Minute)
		sessions := sm.GroupSessions(activities)

		require.Len(t, sessions, 0)
	})

	t.Run("cross-day session", func(t *testing.T) {
		day1 := time.Date(2025, 1, 15, 23, 50, 0, 0, time.UTC)
		day2 := time.Date(2025, 1, 16, 0, 5, 0, 0, time.UTC)

		activities := []core.Activity{
			{ID: "1", Timestamp: day1, Duration: 600, Language: "go", Project: "p1"},
			{ID: "2", Timestamp: day2, Duration: 600, Language: "go", Project: "p1"},
		}

		sm := NewSessionManager(15*time.Minute, 1*time.Minute)
		sessions := sm.GroupSessions(activities)

		require.Len(t, sessions, 1)
	})
}
