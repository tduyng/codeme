package stats

import (
	"time"

	"github.com/tduyng/codeme/core"
	"github.com/tduyng/codeme/util"
)

const (
	DefaultIdleCap        = 2 * time.Minute
	DefaultSessionTimeout = 15 * time.Minute
	DefaultMinSession     = 1 * time.Minute
)

type SessionManager struct {
	timeout     time.Duration
	minDuration time.Duration
	idleCap     time.Duration
}

func NewSessionManager(timeout, minDuration time.Duration) *SessionManager {
	if timeout == 0 {
		timeout = DefaultSessionTimeout
	}
	if minDuration == 0 {
		minDuration = DefaultMinSession
	}
	return &SessionManager{
		timeout:     timeout,
		minDuration: minDuration,
		idleCap:     DefaultIdleCap,
	}
}

func (sm *SessionManager) GroupAndCalculate(activities []core.Activity) ([]core.Activity, []core.Session) {
	if len(activities) == 0 {
		return activities, nil
	}

	idleCapSeconds := sm.idleCap.Seconds()
	timeoutSeconds := sm.timeout.Seconds()
	minSessionSeconds := sm.minDuration.Seconds()

	var sessions []core.Session
	var sessionStart int
	var sessionDuration float64
	projects := util.NewStringSet()
	languages := util.NewStringSet()

	for i := range activities {
		var gap float64
		if i < len(activities)-1 {
			gap = activities[i+1].Timestamp.Sub(activities[i].Timestamp).Seconds()
		}

		isLastActivity := i == len(activities)-1
		isSessionEnd := isLastActivity || gap > timeoutSeconds

		if isSessionEnd {
			activities[i].Duration = idleCapSeconds
		} else {
			activities[i].Duration = gap
		}

		sessionDuration += activities[i].Duration
		projects.Add(activities[i].Project)
		if IsValidLanguage(activities[i].Language) {
			languages.Add(NormalizeLanguage(activities[i].Language))
		}

		if isSessionEnd {
			if sessionDuration >= minSessionSeconds {
				s := core.Session{
					ID:         activities[sessionStart].ID,
					StartTime:  activities[sessionStart].Timestamp,
					EndTime:    activities[i].Timestamp,
					Duration:   sessionDuration,
					Projects:   projects.ToSortedSlice(),
					Languages:  languages.ToSortedSlice(),
					IsActive:   isLastActivity && i == len(activities)-1,
					BreakAfter: 0,
				}
				sessions = append(sessions, s)
			}
			sessionStart = i + 1
			sessionDuration = 0
			projects = util.NewStringSet()
			languages = util.NewStringSet()
		}
	}

	for i := 0; i < len(sessions)-1; i++ {
		sessions[i].BreakAfter = sessions[i+1].StartTime.Sub(sessions[i].EndTime).Seconds()
	}

	return activities, sessions
}

func (sm *SessionManager) GroupSessions(activities []core.Activity) []core.Session {
	_, sessions := sm.GroupAndCalculate(activities)
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
