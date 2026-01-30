package stats

import (
	"os"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/stretchr/testify/require"
	"github.com/tduyng/codeme/core"
)

func TestMain(m *testing.M) {
	time.Local = time.UTC
	os.Exit(m.Run())
}

func TestCalculateAchievements(t *testing.T) {

	tests := []struct {
		name       string
		allTime    APIPeriodStats
		activities []core.Activity
		streakInfo StreakInfo
		checkFunc  func(t *testing.T, achievements []Achievement)
		snapshot   bool
	}{
		{
			name: "no achievements",
			allTime: APIPeriodStats{
				TotalTime:  100,
				TotalLines: 10,
				Languages:  []APILanguageStats{},
				Sessions:   []APISession{},
			},
			activities: []core.Activity{},
			streakInfo: StreakInfo{Current: 0, Longest: 0},
			checkFunc: func(t *testing.T, achievements []Achievement) {
				for _, ach := range achievements {
					require.False(t, ach.Unlocked)
				}
			},
			snapshot: false,
		},
		{
			name: "streak achievements",
			allTime: APIPeriodStats{
				TotalTime:  1000,
				TotalLines: 100,
			},
			activities: []core.Activity{},
			streakInfo: StreakInfo{Current: 30, Longest: 30},
			checkFunc: func(t *testing.T, achievements []Achievement) {
				streakAchievements := filterAchievements(achievements, "streak")
				require.True(t, findAchievement(streakAchievements, "streak_5").Unlocked)
				require.True(t, findAchievement(streakAchievements, "streak_30").Unlocked)
				require.False(t, findAchievement(streakAchievements, "streak_90").Unlocked)
			},
			snapshot: true,
		},
		{
			name: "lines achievements",
			allTime: APIPeriodStats{
				TotalTime:  1000,
				TotalLines: 15000,
			},
			activities: []core.Activity{},
			streakInfo: StreakInfo{Current: 0, Longest: 0},
			checkFunc: func(t *testing.T, achievements []Achievement) {
				linesAchievements := filterAchievements(achievements, "lines")
				require.True(t, findAchievement(linesAchievements, "lines_1000").Unlocked)
				require.True(t, findAchievement(linesAchievements, "lines_10000").Unlocked)
				require.False(t, findAchievement(linesAchievements, "lines_50000").Unlocked)
			},
			snapshot: false,
		},
		{
			name: "hours achievements",
			allTime: APIPeriodStats{
				TotalTime:  200000, // ~55 hours
				TotalLines: 100,
			},
			activities: []core.Activity{},
			streakInfo: StreakInfo{Current: 0, Longest: 0},
			checkFunc: func(t *testing.T, achievements []Achievement) {
				hoursAchievements := filterAchievements(achievements, "hours")
				require.True(t, findAchievement(hoursAchievements, "hours_50").Unlocked)
				require.False(t, findAchievement(hoursAchievements, "hours_1000").Unlocked)
			},
			snapshot: false,
		},
		{
			name: "polyglot achievements",
			allTime: APIPeriodStats{
				TotalTime:  1000,
				TotalLines: 100,
				Languages: []APILanguageStats{
					{Name: "go"},
					{Name: "typescript"},
					{Name: "python"},
					{Name: "rust"},
					{Name: "java"},
					{Name: "c"},
				},
			},
			activities: []core.Activity{},
			streakInfo: StreakInfo{Current: 0, Longest: 0},
			checkFunc: func(t *testing.T, achievements []Achievement) {
				polyglotAchievements := filterAchievements(achievements, "languages")
				require.True(t, findAchievement(polyglotAchievements, "polyglot_2").Unlocked)
				require.True(t, findAchievement(polyglotAchievements, "polyglot_5").Unlocked)
				require.False(t, findAchievement(polyglotAchievements, "polyglot_10").Unlocked)
			},
			snapshot: false,
		},
		{
			name: "early bird achievement",
			allTime: APIPeriodStats{
				TotalTime:  1000,
				TotalLines: 100,
			},
			activities: []core.Activity{
				{Timestamp: time.Date(2025, 1, 15, 5, 0, 0, 0, time.UTC), Duration: 100}, // 5 AM UTC
			},
			streakInfo: StreakInfo{Current: 0, Longest: 0},
			checkFunc: func(t *testing.T, achievements []Achievement) {
				earlyBird := findAchievement(achievements, "early_bird")
				require.True(t, earlyBird.Unlocked)
			},
			snapshot: false,
		},
		{
			name: "night owl achievement",
			allTime: APIPeriodStats{
				TotalTime:  1000,
				TotalLines: 100,
			},
			activities: []core.Activity{
				{Timestamp: time.Date(2025, 1, 15, 1, 0, 0, 0, time.UTC), Duration: 100}, // 1 AM UTC
			},
			streakInfo: StreakInfo{Current: 0, Longest: 0},
			checkFunc: func(t *testing.T, achievements []Achievement) {
				nightOwl := findAchievement(achievements, "night_owl")
				require.True(t, nightOwl.Unlocked)
			},
			snapshot: false,
		},
		{
			name: "session achievements",
			allTime: APIPeriodStats{
				TotalTime:  50000,
				TotalLines: 1000,
				Sessions: []APISession{
					{Duration: 7200},  // 2 hours
					{Duration: 14400}, // 4 hours
					{Duration: 21600}, // 6 hours
				},
			},
			activities: []core.Activity{},
			streakInfo: StreakInfo{Current: 0, Longest: 0},
			checkFunc: func(t *testing.T, achievements []Achievement) {
				require.True(t, findAchievement(achievements, "session_2h").Unlocked)
				require.True(t, findAchievement(achievements, "session_4h").Unlocked)
				require.True(t, findAchievement(achievements, "session_6h").Unlocked)
				require.False(t, findAchievement(achievements, "session_8h").Unlocked)
			},
			snapshot: false,
		},
		{
			name: "comprehensive unlock",
			allTime: APIPeriodStats{
				TotalTime:  4000000, // ~1111 hours
				TotalLines: 60000,
				Languages: []APILanguageStats{
					{Name: "go"}, {Name: "typescript"}, {Name: "python"},
					{Name: "rust"}, {Name: "java"}, {Name: "c"},
					{Name: "cpp"}, {Name: "ruby"}, {Name: "php"},
					{Name: "kotlin"}, {Name: "swift"},
				},
				Sessions: []APISession{
					{Duration: 43200}, // 12 hours
				},
			},
			activities: []core.Activity{
				{Timestamp: time.Date(2025, 1, 15, 5, 0, 0, 0, time.UTC), Duration: 100}, // 5 AM
				{Timestamp: time.Date(2025, 1, 15, 1, 0, 0, 0, time.UTC), Duration: 100}, // 1 AM
			},
			streakInfo: StreakInfo{Current: 100, Longest: 100},
			snapshot:   true,
		},
	}

	snapshotter := cupaloy.New(cupaloy.SnapshotSubdirectory("testdata/snapshots"))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			achievements := CalculateAchievements(tt.allTime, tt.activities, tt.streakInfo)

			require.NotEmpty(t, achievements)

			if tt.checkFunc != nil {
				tt.checkFunc(t, achievements)
			}

			if tt.snapshot {
				snapshotter.SnapshotT(t, achievements)
			}
		})
	}
}

func TestAchievementConfigs(t *testing.T) {
	// Verify all achievement configs are valid
	require.NotEmpty(t, AchievementConfigs)

	seenIDs := make(map[string]bool)
	for _, cfg := range AchievementConfigs {
		// No duplicate IDs
		require.False(t, seenIDs[cfg.ID], "duplicate achievement ID: %s", cfg.ID)
		seenIDs[cfg.ID] = true

		// Required fields
		require.NotEmpty(t, cfg.ID)
		require.NotEmpty(t, cfg.Name)
		require.NotEmpty(t, cfg.Description)
		require.NotEmpty(t, cfg.Type)
		require.NotEmpty(t, cfg.Icon)

		// Type-specific validation
		switch cfg.Type {
		case "streak", "lines", "hours", "languages":
			require.Greater(t, cfg.Threshold, 0, "achievement %s missing threshold", cfg.ID)
		case "habit":
			require.NotEmpty(t, cfg.Hours, "achievement %s missing hours", cfg.ID)
		case "session":
			require.Greater(t, cfg.MinSession, 0, "achievement %s missing min_session", cfg.ID)
		default:
			t.Fatalf("unknown achievement type: %s", cfg.Type)
		}
	}
}

func TestAchievements_EdgeCases(t *testing.T) {
	t.Run("exactly at threshold", func(t *testing.T) {
		allTime := APIPeriodStats{
			TotalLines: 1000,
		}
		achievements := CalculateAchievements(allTime, nil, StreakInfo{})
		lines1k := findAchievement(achievements, "lines_1000")
		require.True(t, lines1k.Unlocked)
	})

	t.Run("non-code languages excluded from polyglot", func(t *testing.T) {
		allTime := APIPeriodStats{
			Languages: []APILanguageStats{
				{Name: "go"},
				{Name: "yaml"},     // config
				{Name: "markdown"}, // doc
			},
		}
		achievements := CalculateAchievements(allTime, nil, StreakInfo{})
		polyglot2 := findAchievement(achievements, "polyglot_2")
		require.False(t, polyglot2.Unlocked) // Only 1 code language
	})

	t.Run("midnight hour activity", func(t *testing.T) {
		baseTime := time.Date(2025, 1, 15, 0, 30, 0, 0, time.UTC)
		activities := []core.Activity{
			{Timestamp: baseTime, Duration: 100},
		}
		achievements := CalculateAchievements(APIPeriodStats{}, activities, StreakInfo{})
		nightOwl := findAchievement(achievements, "night_owl")
		require.True(t, nightOwl.Unlocked)
	})
}

// Helper functions
func filterAchievements(achievements []Achievement, achievementType string) []Achievement {
	var result []Achievement
	for _, ach := range achievements {
		// Find in configs
		for _, cfg := range AchievementConfigs {
			if cfg.ID == ach.ID && cfg.Type == achievementType {
				result = append(result, ach)
				break
			}
		}
	}
	return result
}

func findAchievement(achievements []Achievement, id string) Achievement {
	for _, ach := range achievements {
		if ach.ID == id {
			return ach
		}
	}
	return Achievement{}
}
