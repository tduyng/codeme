package stats

import (
	"time"
)

const SessionGapMinutes = 15

// DetectSessionGap returns true if timestamps are within session threshold
func DetectSessionGap(prev, current time.Time) bool {
	return !prev.IsZero() && current.Sub(prev).Minutes() <= SessionGapMinutes
}

// CalculateAvgSessionLength computes average session duration
func CalculateAvgSessionLength(sessions []Session) int64 {
	if len(sessions) == 0 {
		return 0
	}

	total := int64(0)
	for _, session := range sessions {
		total += session.Duration
	}

	return total / int64(len(sessions))
}

// CalculateSessionBreaks adds break duration between sessions
func CalculateSessionBreaks(sessions []Session) []Session {
	for i := 1; i < len(sessions); i++ {
		prevEnd, _ := time.Parse(time.RFC3339, sessions[i-1].End)
		currentStart, _ := time.Parse(time.RFC3339, sessions[i].Start)
		breakDuration := int64(currentStart.Sub(prevEnd).Seconds())
		sessions[i-1].BreakAfter = breakDuration
	}
	return sessions
}
