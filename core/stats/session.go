// core/stats/session.go
package stats

import (
	"sort"
	"time"

	"github.com/tduyng/codeme/util"
)

// SessionManager handles coding session logic
type SessionManager struct {
	timeout     time.Duration
	minDuration time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager(timeout, minDuration time.Duration) *SessionManager {
	if timeout == 0 {
		timeout = 15 * time.Minute
	}
	if minDuration == 0 {
		minDuration = 1 * time.Minute
	}
	return &SessionManager{
		timeout:     timeout,
		minDuration: minDuration,
	}
}

// ShouldStartNewSession determines if a new session should be started
func (sm *SessionManager) ShouldStartNewSession(lastActivity time.Time, currentTime time.Time) bool {
	return currentTime.Sub(lastActivity) > sm.timeout
}

// IsValidSession checks if a session meets minimum requirements
func (sm *SessionManager) IsValidSession(session Session) bool {
	return session.Duration >= sm.minDuration.Seconds()
}

// GroupActivitiesIntoSessions groups activities into sessions
func (sm *SessionManager) GroupActivitiesIntoSessions(activities []Activity) []Session {
	if len(activities) == 0 {
		return nil
	}

	// Sort chronologically
	sorted := make([]Activity, len(activities))
	copy(sorted, activities)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	var sessions []Session
	var current *Session
	var projects, languages util.StringSet

	start := func(a Activity) {
		projects = util.NewStringSet()
		languages = util.NewStringSet()
		projects.Add(a.Project)
		if IsValidLanguage(a.Language) {
			languages.Add(NormalizeLanguage(a.Language))
		}

		current = &Session{
			ID:         a.ID,
			StartTime:  a.Timestamp,
			EndTime:    a.Timestamp,
			Activities: []Activity{a},
			IsActive:   false,
		}
	}

	finalize := func() {
		// CRITICAL: Session duration = sum of activity durations (not time span!)
		sessionDuration := 0.0
		for _, act := range current.Activities {
			sessionDuration += act.Duration
		}
		current.Duration = sessionDuration

		// Set end time to last activity timestamp (not extended)
		if len(current.Activities) > 0 {
			current.EndTime = current.Activities[len(current.Activities)-1].Timestamp
		}

		// Populate projects and languages
		current.Projects = projects.ToSortedSlice()
		current.Languages = languages.ToSortedSlice()

		if sm.IsValidSession(*current) {
			sessions = append(sessions, *current)
		}
	}

	for _, a := range sorted {
		if current == nil {
			start(a)
			continue
		}

		if a.Timestamp.Sub(current.EndTime) > sm.timeout {
			// Gap too large - finalize current session and start new one
			finalize()
			start(a)
			continue
		}

		// Continue current session
		current.Activities = append(current.Activities, a)
		current.EndTime = a.Timestamp
		projects.Add(a.Project)
		if IsValidLanguage(a.Language) {
			languages.Add(NormalizeLanguage(a.Language))
		}
	}

	// Finalize last session
	if current != nil {
		finalize()
	}

	// Calculate breaks between sessions
	for i := 0; i < len(sessions)-1; i++ {
		sessions[i].BreakAfter = sessions[i+1].StartTime.Sub(
			sessions[i].EndTime,
		).Seconds()
	}

	return sessions
}
