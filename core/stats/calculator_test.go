package stats

import (
	"database/sql"
	"testing"
	"time"
)

// Helper to create test database
func setupTestDB(t *testing.T) *sql.DB {
	// Your existing test DB setup
	return nil // Placeholder
}

func TestCalculateStats(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("Test DB not configured")
		return
	}
	defer db.Close()

	// Test full integration
	stats, err := CalculateStats(db, false)
	if err != nil {
		t.Fatalf("CalculateStats() error = %v", err)
	}

	// Verify key fields are populated
	if stats.DailyGoals.TimeGoal == 0 {
		t.Error("Expected DailyGoals to be initialized")
	}
}

func TestCalculateStatsWithTodayFilter(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("Test DB not configured")
		return
	}
	defer db.Close()

	stats, err := CalculateStats(db, true)
	if err != nil {
		t.Fatalf("CalculateStats() error = %v", err)
	}

	// Verify today-only filtering worked
	today := time.Now().Format("2006-01-02")
	if len(stats.DailyActivity) > 1 {
		t.Error("Expected only today's activity when todayOnly=true")
	}

	// Check that today's data exists if there should be activity
	if todayActivity, exists := stats.DailyActivity[today]; !exists && stats.TodayTime > 0 {
		t.Error("Expected today's activity to be present")
	} else if exists && todayActivity.Time != stats.TodayTime {
		t.Errorf("Expected today activity time %d, got %d", stats.TodayTime, todayActivity.Time)
	}
}
