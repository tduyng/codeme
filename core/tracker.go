package core

import (
	"database/sql"
	"time"
)

func Track(db *sql.DB, file, language string, linesChanged, linesTotal int) error {
	project := DetectProject(file)
	branch := DetectBranch(file)

	hb := Heartbeat{
		Timestamp:    time.Now(),
		File:         file,
		Language:     language,
		Project:      project,
		Branch:       branch,
		Lines:        linesChanged, // For backward compatibility
		LinesChanged: linesChanged,
		LinesTotal:   linesTotal,
	}

	return SaveHeartbeat(db, hb)
}
