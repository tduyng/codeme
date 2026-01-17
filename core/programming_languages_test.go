package core

import (
	"testing"
	"time"

	"github.com/tduyng/codeme/core/stats"
)

func TestProgrammingLanguageFiltering(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test data with mix of code and non-code languages
	testData := []struct {
		language string
		lines    int
	}{
		{"javascript", 100}, // code
		{"typescript", 150}, // code
		{"go", 200},         // code
		{"lua", 75},         // code
		{"markdown", 50},    // markup
		{"yaml", 25},        // config
		{"json", 30},        // data
		{"gitignore", 5},    // meta
	}

	now := time.Now()
	for _, data := range testData {
		hb := Heartbeat{
			Timestamp: now,
			File:      "/test/file." + data.language,
			Language:  data.language,
			Project:   "test",
			Branch:    "main",
			Lines:     data.lines,
		}
		err := SaveHeartbeat(db, hb)
		if err != nil {
			t.Fatalf("Failed to save heartbeat: %v", err)
		}
	}

	// Calculate calStats
	calStats, err := stats.CalculateStats(db, false)
	if err != nil {
		t.Fatalf("Failed to calculate stats: %v", err)
	}

	// Verify all languages are in Languages map
	expectedAllLanguages := []string{"javascript", "typescript", "go", "lua", "markdown", "yaml", "json", "gitignore"}
	for _, lang := range expectedAllLanguages {
		if _, exists := calStats.Languages[lang]; !exists {
			t.Errorf("Language %s not found in Languages map", lang)
		}
	}

	// Verify only programming languages are in ProgrammingLanguages map
	expectedProgrammingLanguages := []string{"javascript", "typescript", "go", "lua"}
	excludedLanguages := []string{"markdown", "yaml", "json", "gitignore"}

	for _, lang := range expectedProgrammingLanguages {
		if _, exists := calStats.ProgrammingLanguages[lang]; !exists {
			t.Errorf("Programming language %s not found in ProgrammingLanguages map", lang)
		}
	}

	for _, lang := range excludedLanguages {
		if _, exists := calStats.ProgrammingLanguages[lang]; exists {
			t.Errorf("Non-programming language %s should not be in ProgrammingLanguages map", lang)
		}
	}

	// Verify line counts match
	expectedProgrammingLines := 100 + 150 + 200 + 75 // js + ts + go + lua = 525
	actualProgrammingLines := 0
	for _, stat := range calStats.ProgrammingLanguages {
		actualProgrammingLines += stat.Lines
	}

	if actualProgrammingLines != expectedProgrammingLines {
		t.Errorf("Expected %d programming language lines, got %d", expectedProgrammingLines, actualProgrammingLines)
	}

	// Verify total lines include everything
	expectedTotalLines := 100 + 150 + 200 + 75 + 50 + 25 + 30 + 5 // 635
	if calStats.TotalLines != expectedTotalLines {
		t.Errorf("Expected %d total lines, got %d", expectedTotalLines, calStats.TotalLines)
	}

	t.Logf("✓ Languages map has %d entries", len(calStats.Languages))
	t.Logf("✓ ProgrammingLanguages map has %d entries", len(calStats.ProgrammingLanguages))
	t.Logf("✓ Programming languages: %d lines", actualProgrammingLines)
	t.Logf("✓ Total lines: %d lines", calStats.TotalLines)
}
