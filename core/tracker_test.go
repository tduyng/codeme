package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockStorage struct {
	activities []Activity
	closed     bool
}

func (m *mockStorage) SaveActivity(a Activity) error {
	m.activities = append(m.activities, a)
	return nil
}

func (m *mockStorage) GetActivitiesSince(since time.Time) ([]Activity, error) {
	var result []Activity
	for _, a := range m.activities {
		if !a.Timestamp.Before(since) {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockStorage) GetActivityCount() (int, error) {
	return len(m.activities), nil
}

func (m *mockStorage) Optimize() error {
	return nil
}

func (m *mockStorage) Close() error {
	m.closed = true
	return nil
}

func TestTracker_TrackFileActivity(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		language     string
		editor       string
		linesChanged int
		isWrite      bool
		checkFields  func(t *testing.T, a Activity)
	}{
		{
			name:         "basic tracking",
			filePath:     "/project/main.go",
			language:     "go",
			editor:       "neovim",
			linesChanged: 10,
			isWrite:      true,
			checkFields: func(t *testing.T, a Activity) {
				require.Equal(t, "go", a.Language)
				require.Equal(t, "neovim", a.Editor)
				require.Equal(t, 10, a.Lines)
				require.True(t, a.IsWrite)
			},
		},
		{
			name:         "auto-detect language",
			filePath:     "/project/app.ts",
			language:     "",
			editor:       "vscode",
			linesChanged: 5,
			isWrite:      false,
			checkFields: func(t *testing.T, a Activity) {
				require.Equal(t, "typescript", a.Language)
				require.Equal(t, "vscode", a.Editor)
				require.Equal(t, 5, a.Lines)
				require.False(t, a.IsWrite)
			},
		},
		{
			name:         "default editor",
			filePath:     "/project/script.py",
			language:     "python",
			editor:       "",
			linesChanged: 0,
			isWrite:      true,
			checkFields: func(t *testing.T, a Activity) {
				require.Equal(t, "neovim", a.Editor)
				require.Equal(t, 0, a.Lines)
			},
		},
		{
			name:         "unknown language",
			filePath:     "/project/file.xyz",
			language:     "",
			editor:       "vim",
			linesChanged: 100,
			isWrite:      true,
			checkFields: func(t *testing.T, a Activity) {
				require.Equal(t, "xyz", a.Language) // Auto-detected from extension
				require.Equal(t, 100, a.Lines)
			},
		},
		{
			name:         "negative lines",
			filePath:     "/project/test.go",
			language:     "go",
			editor:       "neovim",
			linesChanged: -5,
			isWrite:      true,
			checkFields: func(t *testing.T, a Activity) {
				require.Equal(t, -5, a.Lines)
				require.NotEmpty(t, a.Project) // Should detect project
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &mockStorage{}
			tracker := NewTracker(storage)

			err := tracker.TrackFileActivity(
				tt.filePath,
				tt.language,
				tt.editor,
				tt.linesChanged,
				tt.isWrite,
			)

			require.NoError(t, err)
			require.Len(t, storage.activities, 1)

			activity := storage.activities[0]
			require.NotEmpty(t, activity.ID)
			require.NotZero(t, activity.Timestamp)
			require.Equal(t, tt.filePath, activity.File)
			require.NotEmpty(t, activity.Project)

			if tt.checkFields != nil {
				tt.checkFields(t, activity)
			}
		})
	}
}

func TestTracker_Close(t *testing.T) {
	storage := &mockStorage{}
	tracker := NewTracker(storage)

	err := tracker.Close()
	require.NoError(t, err)
	require.True(t, storage.closed)
}

func TestTracker_MultipleActivities(t *testing.T) {
	storage := &mockStorage{}
	tracker := NewTracker(storage)

	files := []struct {
		path  string
		lang  string
		lines int
	}{
		{"/p1/a.go", "go", 10},
		{"/p1/b.ts", "typescript", 20},
		{"/p2/c.py", "python", 30},
	}

	for _, f := range files {
		err := tracker.TrackFileActivity(f.path, f.lang, "neovim", f.lines, true)
		require.NoError(t, err)
	}

	require.Len(t, storage.activities, 3)

	// Verify all activities are tracked
	totalLines := 0
	for _, a := range storage.activities {
		totalLines += a.Lines
	}
	require.Equal(t, 60, totalLines)
}

func TestTracker_ProjectDetection(t *testing.T) {
	storage := &mockStorage{}
	tracker := NewTracker(storage)

	err := tracker.TrackFileActivity("/some/path/project/main.go", "go", "vim", 5, true)
	require.NoError(t, err)

	require.Len(t, storage.activities, 1)
	require.NotEmpty(t, storage.activities[0].Project)
}
