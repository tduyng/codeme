package core

import (
	"database/sql"
	"path/filepath"
	"testing"
)

// setupTestDB creates an isolated test database in a temporary directory
// This helper is shared across all test files in the core package
func setupTestDB(t *testing.T) *sql.DB {
	tmpDir := t.TempDir()
	testDB := filepath.Join(tmpDir, "test.db")
	db, err := OpenDBWithPath(testDB)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	return db
}
