package core

import (
	"testing"
)

func TestTrack(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name         string
		file         string
		language     string
		linesChanged int
		linesTotal   int
		wantErr      bool
	}{
		{
			name:         "Track Go file",
			file:         "/test/main.go",
			language:     "Go",
			linesChanged: 50,
			linesTotal:   100,
			wantErr:      false,
		},
		{
			name:         "Track Lua file",
			file:         "/test/init.lua",
			language:     "Lua",
			linesChanged: 30,
			linesTotal:   80,
			wantErr:      false,
		},
		{
			name:         "Track with zero lines",
			file:         "/test/empty.txt",
			language:     "Text",
			linesChanged: 0,
			linesTotal:   0,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Track(db, tt.file, tt.language, tt.linesChanged, tt.linesTotal)
			if (err != nil) != tt.wantErr {
				t.Errorf("Track() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify the heartbeat was saved
			if !tt.wantErr {
				var count int
				db.QueryRow("SELECT COUNT(*) FROM heartbeats WHERE file = ?", tt.file).Scan(&count)
				if count != 1 {
					t.Errorf("Heartbeat not saved for file %s", tt.file)
				}

				// Verify lines_changed and lines_total were saved
				var linesChanged, linesTotal int
				db.QueryRow("SELECT lines_changed, lines_total FROM heartbeats WHERE file = ? ORDER BY timestamp DESC LIMIT 1",
					tt.file).Scan(&linesChanged, &linesTotal)

				if linesChanged != tt.linesChanged {
					t.Errorf("lines_changed = %d, want %d", linesChanged, tt.linesChanged)
				}
				if linesTotal != tt.linesTotal {
					t.Errorf("lines_total = %d, want %d", linesTotal, tt.linesTotal)
				}
			}
		})
	}
}

func TestTrackWithChangedLines(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Track initial file
	err := Track(db, "/test/evolving.go", "Go", 100, 100)
	if err != nil {
		t.Fatalf("First Track() error = %v", err)
	}

	// Track changes to the same file
	err = Track(db, "/test/evolving.go", "Go", 10, 110) // Added 10 lines
	if err != nil {
		t.Fatalf("Second Track() error = %v", err)
	}

	// Verify both heartbeats exist
	var count int
	db.QueryRow("SELECT COUNT(*) FROM heartbeats WHERE file = ?", "/test/evolving.go").Scan(&count)
	if count != 2 {
		t.Errorf("Expected 2 heartbeats, got %d", count)
	}

	// Verify the second heartbeat has correct values
	var linesChanged, linesTotal int
	db.QueryRow("SELECT lines_changed, lines_total FROM heartbeats WHERE file = ? ORDER BY timestamp DESC LIMIT 1",
		"/test/evolving.go").Scan(&linesChanged, &linesTotal)

	if linesChanged != 10 {
		t.Errorf("lines_changed = %d, want 10", linesChanged)
	}
	if linesTotal != 110 {
		t.Errorf("lines_total = %d, want 110", linesTotal)
	}
}
