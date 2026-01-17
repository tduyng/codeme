package stats

import (
	"fmt"
	"math"
)

// CalculateFocusScore computes a focus score based on session patterns
func CalculateFocusScore(sessions []Session) int {
	if len(sessions) == 0 {
		return 0
	}

	totalDuration := int64(0)
	sessionCount := len(sessions)

	for _, session := range sessions {
		totalDuration += session.Duration
	}

	avgSession := totalDuration / int64(sessionCount)

	// === BASE SCORE: Average Session Length ===
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

	// === BONUS: Session Count (healthy rhythm) ===
	var sessionBonus int
	if sessionCount == 1 {
		// Single long session = deep focus
		if avgSession >= 5400 { // 1.5+ hours
			sessionBonus = 10 // Reward deep focus
		} else {
			sessionBonus = -5 // Penalize short single session
		}
	} else if sessionCount >= 2 && sessionCount <= 5 {
		// Healthy rhythm (2-5 sessions with breaks)
		sessionBonus = 10
	} else if sessionCount >= 6 && sessionCount <= 8 {
		// Many sessions (could be good with breaks, or scattered)
		sessionBonus = 0
	} else if sessionCount > 8 {
		// Too fragmented
		sessionBonus = -10
	}

	// === BONUS: Session Consistency ===
	// Calculate variance - are sessions similar length?
	var variance float64
	for _, session := range sessions {
		diff := float64(session.Duration) - float64(avgSession)
		variance += diff * diff
	}
	variance = variance / float64(sessionCount)
	stdDev := math.Sqrt(variance)

	var consistencyBonus int
	// Low variance = consistent sessions = good focus
	if stdDev < float64(avgSession)*0.3 { // Within 30% of average
		consistencyBonus = 10
	} else if stdDev < float64(avgSession)*0.5 { // Within 50%
		consistencyBonus = 5
	} else {
		// High variance = scattered focus
		consistencyBonus = -5
	}

	// === PENALTY: Break Analysis ===
	breakPenalty := 0
	if sessionCount > 1 {
		longBreaks := 0  // Breaks > 2 hours
		shortBreaks := 0 // Breaks < 15 minutes

		for i := 0; i < len(sessions)-1; i++ {
			breakDuration := sessions[i].BreakAfter
			if breakDuration > 7200 { // > 2 hours
				longBreaks++
			} else if breakDuration < 900 && breakDuration > 0 { // < 15 min
				shortBreaks++
			}
		}

		// Too many long breaks = distracted
		if longBreaks > sessionCount/2 {
			breakPenalty = -10
		}
		// Too many short breaks = context switching
		if shortBreaks > sessionCount/2 {
			breakPenalty = -5
		}
	}

	// === FINAL SCORE ===
	finalScore := baseScore + sessionBonus + consistencyBonus + breakPenalty

	// Clamp to 0-100
	return int(math.Max(0, math.Min(100, float64(finalScore))))
}

// CalculateProductivityTrend compares this week vs last week
func CalculateProductivityTrend(dailyActivity map[string]DailyStat, thisWeekStart, lastWeekStart, lastWeekEnd string) string {
	thisWeekTime := int64(0)
	lastWeekTime := int64(0)

	for date, stat := range dailyActivity {
		if date >= thisWeekStart {
			thisWeekTime += stat.Time
		} else if date >= lastWeekStart && date <= lastWeekEnd {
			lastWeekTime += stat.Time
		}
	}

	if lastWeekTime == 0 {
		if thisWeekTime > 0 {
			return "🚀 Starting Strong"
		}
		return "📊 Building Data"
	}

	growthPercent := ((float64(thisWeekTime) - float64(lastWeekTime)) / float64(lastWeekTime)) * 100

	if growthPercent > 10 {
		return "↗ Improving"
	} else if growthPercent < -10 {
		return "↘ Declining"
	}
	return "→ Stable"
}

// CalculateGrowth computes week-over-week growth for a project
func CalculateGrowth(project string, dailyActivity map[string]DailyStat, thisWeekStart, lastWeekStart, lastWeekEnd string) string {
	thisWeekTime := int64(0)
	lastWeekTime := int64(0)

	for date, stat := range dailyActivity {
		if date >= thisWeekStart {
			thisWeekTime += stat.Time
		} else if date >= lastWeekStart && date <= lastWeekEnd {
			lastWeekTime += stat.Time
		}
	}

	if lastWeekTime == 0 {
		if thisWeekTime > 0 {
			return "🆕 New"
		}
		return ""
	}

	growthPercent := ((float64(thisWeekTime) - float64(lastWeekTime)) / float64(lastWeekTime)) * 100

	if growthPercent > 5 {
		return fmt.Sprintf("↗ +%.0f%%", growthPercent)
	} else if growthPercent < -5 {
		return fmt.Sprintf("↘ %.0f%%", growthPercent)
	}
	return "→ Stable"
}
