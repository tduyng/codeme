package stats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tduyng/codeme/core"
)

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
