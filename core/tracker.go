// core/tracker.go
package core

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tduyng/codeme/core/stats"
)

// Tracker manages activity tracking
type Tracker struct {
	storage    Storage
	calculator *stats.Calculator
}

// TrackerConfig holds tracker configuration
type TrackerConfig struct {
	Storage  Storage
	Timezone *time.Location
}

// NewTracker creates a new activity tracker
func NewTracker(config TrackerConfig) *Tracker {
	if config.Timezone == nil {
		config.Timezone = time.UTC
	}

	return &Tracker{
		storage:    config.Storage,
		calculator: stats.NewCalculator(config.Timezone),
	}
}

// TrackFileActivity records a file edit activity
// NOTE: We only save activities, not sessions
// Sessions are calculated on-demand from activities
func (t *Tracker) TrackFileActivity(filePath, language, editor string, linesChanged int, isWrite bool) error {
	now := time.Now()

	// Detect language if not provided
	if language == "" || language == "Unknown" {
		language = DetectLanguage(filePath)
	}

	// Detect project
	project := DetectProject(filePath)

	// Create activity WITHOUT duration
	// Duration will be calculated later from gaps between activities
	activity := stats.Activity{
		ID:        uuid.New().String(),
		Timestamp: now,
		Duration:  0, // Will be calculated from gaps in calculator.go
		Lines:     linesChanged,
		Language:  language,
		Project:   project,
		Editor:    editor,
		File:      filePath,
		IsWrite:   isWrite,
	}

	// Save activity only - sessions will be calculated on-demand
	if err := t.storage.SaveActivity(activity); err != nil {
		return fmt.Errorf("failed to save activity: %w", err)
	}

	return nil
}

// Close performs cleanup
func (t *Tracker) Close() error {
	return t.storage.Close()
}

func OpenStorage() (Storage, error) {
	dbPath, err := GetDefaultDBPath()
	if err != nil {
		return nil, err
	}
	return NewSQLiteStorage(dbPath)
}
