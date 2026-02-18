package core

import "time"

type Activity struct {
	ID        string
	Timestamp time.Time
	Duration  float64
	Lines     int
	Language  string
	Project   string
	Editor    string
	File      string
	Branch    string
	IsWrite   bool
}

type Session struct {
	ID         string
	StartTime  time.Time
	EndTime    time.Time
	Duration   float64
	Projects   []string
	Languages  []string
	IsActive   bool
	BreakAfter float64
}

type Storage interface {
	SaveActivity(Activity) error
	GetActivitiesSince(time.Time) ([]Activity, error)
	GetActivityCount() (int, error)
	Optimize() error
	RebuildSummaries() error
	Close() error
}
