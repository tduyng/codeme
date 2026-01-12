package core

import (
	"os"
	"testing"
)

func TestOpenDB(t *testing.T) {
	db, err := OpenDB()
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	defer db.Close()

	// Verify table exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='heartbeats'").Scan(&count)
	if err != nil || count != 1 {
		t.Errorf("heartbeats table not created")
	}
}

func TestSaveHeartbeat(t *testing.T) {
	db, _ := OpenDB()
	defer db.Close()
	defer os.RemoveAll(os.ExpandEnv("$HOME/.local/share/codeme"))

	hb := Heartbeat{
		File:     "/test/file.go",
		Language: "Go",
		Project:  "test",
		Lines:    10,
	}

	err := SaveHeartbeat(db, hb)
	if err != nil {
		t.Errorf("SaveHeartbeat() error = %v", err)
	}
}
