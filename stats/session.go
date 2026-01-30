package stats

import (
	"sort"
	"time"

	"github.com/tduyng/codeme/core"
	"github.com/tduyng/codeme/util"
)

type SessionManager struct {
	timeout     time.Duration
	minDuration time.Duration
}

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

func (sm *SessionManager) GroupSessions(activities []core.Activity) []core.Session {
	if len(activities) == 0 {
		return nil
	}

	sorted := make([]core.Activity, len(activities))
	copy(sorted, activities)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	var sessions []core.Session
	var current *core.Session
	var projects, languages util.StringSet

	start := func(a core.Activity) {
		projects = util.NewStringSet()
		languages = util.NewStringSet()
		projects.Add(a.Project)
		if IsValidLanguage(a.Language) {
			languages.Add(NormalizeLanguage(a.Language))
		}

		current = &core.Session{
			ID:        a.ID,
			StartTime: a.Timestamp,
			EndTime:   a.Timestamp,
			IsActive:  false,
		}
	}

	finalize := func() {
		sessionDuration := 0.0
		for i := range sorted {
			if sorted[i].Timestamp.Before(current.StartTime) {
				continue
			}
			if sorted[i].Timestamp.After(current.EndTime) {
				break
			}
			sessionDuration += sorted[i].Duration
		}

		current.Duration = sessionDuration
		current.Projects = projects.ToSortedSlice()
		current.Languages = languages.ToSortedSlice()

		if current.Duration >= sm.minDuration.Seconds() {
			sessions = append(sessions, *current)
		}
	}

	for _, a := range sorted {
		if current == nil {
			start(a)
			continue
		}

		if a.Timestamp.Sub(current.EndTime) > sm.timeout {
			finalize()
			start(a)
			continue
		}

		current.EndTime = a.Timestamp
		projects.Add(a.Project)
		if IsValidLanguage(a.Language) {
			languages.Add(NormalizeLanguage(a.Language))
		}
	}

	if current != nil {
		finalize()
	}

	for i := 0; i < len(sessions)-1; i++ {
		sessions[i].BreakAfter = sessions[i+1].StartTime.Sub(sessions[i].EndTime).Seconds()
	}

	return sessions
}

func ConvertSessionsToAPI(sessions []core.Session) []APISession {
	result := make([]APISession, len(sessions))

	for i, s := range sessions {
		result[i] = APISession{
			ID:         s.ID,
			StartTime:  s.StartTime,
			EndTime:    s.EndTime,
			Duration:   s.Duration,
			Projects:   s.Projects,
			Languages:  s.Languages,
			IsActive:   s.IsActive,
			BreakAfter: s.BreakAfter,
		}
	}

	return result
}
