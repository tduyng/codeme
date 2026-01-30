// core/tracker.go
package core

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Tracker struct {
	storage  Storage
	detector *Detector
}

func NewTracker(storage Storage) *Tracker {
	return &Tracker{
		storage:  storage,
		detector: NewDetector(),
	}
}

func (t *Tracker) TrackFileActivity(filePath, language, editor string, linesChanged int, isWrite bool) error {
	now := time.Now()

	if language == "" || language == "unknown" {
		language = t.detector.DetectLanguage(filePath)
	}

	project := t.detector.DetectProject(filePath)

	if editor == "" {
		editor = "neovim"
	}

	activity := Activity{
		ID:        uuid.New().String(),
		Timestamp: now,
		Duration:  0,
		Lines:     linesChanged,
		Language:  language,
		Project:   project,
		Editor:    editor,
		File:      filePath,
		IsWrite:   isWrite,
	}

	if err := t.storage.SaveActivity(activity); err != nil {
		return fmt.Errorf("failed to save activity: %w", err)
	}

	return nil
}

func (t *Tracker) Close() error {
	return t.storage.Close()
}
