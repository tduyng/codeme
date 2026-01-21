package stats

import (
	"math"
)

// CalculateFocusScore computes a focus score based on session patterns
func CalculateFocusScore(sessions []Session) int {
	if len(sessions) == 0 {
		return 0
	}

	totalDuration := 0.0
	sessionCount := len(sessions)

	for _, session := range sessions {
		totalDuration += session.Duration
	}

	avgSession := totalDuration / float64(sessionCount)

	// Base score from average session length
	var baseScore int
	if avgSession >= 7200 { // 2+ hours
		baseScore = 90
	} else if avgSession >= 5400 { // 1.5+ hours
		baseScore = 80
	} else if avgSession >= 3600 { // 1+ hour
		baseScore = 70
	} else if avgSession >= 2700 { // 45+ minutes
		baseScore = 60
	} else if avgSession >= 1800 { // 30+ minutes
		baseScore = 50
	} else if avgSession >= 900 { // 15+ minutes
		baseScore = 40
	} else {
		baseScore = 30
	}

	// Session count bonus
	var sessionBonus int
	if sessionCount == 1 {
		if avgSession >= 5400 {
			sessionBonus = 10
		} else {
			sessionBonus = -5
		}
	} else if sessionCount >= 2 && sessionCount <= 5 {
		sessionBonus = 10
	} else if sessionCount > 8 {
		sessionBonus = -10
	}

	// Consistency bonus
	var variance float64
	for _, session := range sessions {
		diff := session.Duration - avgSession
		variance += diff * diff
	}
	variance = variance / float64(sessionCount)
	stdDev := math.Sqrt(variance)

	var consistencyBonus int
	if stdDev < avgSession*0.3 {
		consistencyBonus = 10
	} else if stdDev < avgSession*0.5 {
		consistencyBonus = 5
	} else {
		consistencyBonus = -5
	}

	// Break penalty
	breakPenalty := 0
	if sessionCount > 1 {
		longBreaks := 0
		shortBreaks := 0

		for i := 0; i < len(sessions)-1; i++ {
			breakDuration := sessions[i].BreakAfter
			if breakDuration > 7200 {
				longBreaks++
			} else if breakDuration < 900 && breakDuration > 0 {
				shortBreaks++
			}
		}

		if longBreaks > sessionCount/2 {
			breakPenalty = -10
		}
		if shortBreaks > sessionCount/2 {
			breakPenalty = -5
		}
	}

	finalScore := baseScore + sessionBonus + consistencyBonus + breakPenalty
	return int(math.Max(0, math.Min(100, float64(finalScore))))
}
