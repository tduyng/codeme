package stats

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tduyng/codeme/core"
	"github.com/tduyng/codeme/util"
)

func TestAggregateByLanguage(t *testing.T) {
	baseTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		activities []core.Activity
		check      func(t *testing.T, result map[string]*LanguageAgg)
	}{
		{
			name:       "empty activities",
			activities: []core.Activity{},
			check: func(t *testing.T, result map[string]*LanguageAgg) {
				require.Empty(t, result)
			},
		},
		{
			name: "single language",
			activities: []core.Activity{
				{Timestamp: baseTime, Duration: 100, Lines: 10, Language: "go", File: "/a.go"},
				{Timestamp: baseTime.Add(1 * time.Hour), Duration: 200, Lines: 20, Language: "go", File: "/b.go"},
			},
			check: func(t *testing.T, result map[string]*LanguageAgg) {
				require.Len(t, result, 1)
				require.Equal(t, 300.0, result["go"].Time)
				require.Equal(t, 30, result["go"].Lines)
				require.Equal(t, 2, result["go"].Files.Len())
			},
		},
		{
			name: "multiple languages",
			activities: []core.Activity{
				{Timestamp: baseTime, Duration: 100, Lines: 10, Language: "go", File: "/a.go"},
				{Timestamp: baseTime, Duration: 150, Lines: 15, Language: "typescript", File: "/b.ts"},
				{Timestamp: baseTime, Duration: 200, Lines: 20, Language: "python", File: "/c.py"},
			},
			check: func(t *testing.T, result map[string]*LanguageAgg) {
				require.Len(t, result, 3)
			},
		},
		{
			name: "case normalization",
			activities: []core.Activity{
				{Timestamp: baseTime, Duration: 100, Lines: 10, Language: "Go", File: "/a.go"},
				{Timestamp: baseTime, Duration: 100, Lines: 10, Language: "go", File: "/b.go"},
				{Timestamp: baseTime, Duration: 100, Lines: 10, Language: "GO", File: "/c.go"},
			},
			check: func(t *testing.T, result map[string]*LanguageAgg) {
				require.Len(t, result, 1)
				require.Equal(t, 300.0, result["go"].Time)
			},
		},
		{
			name: "invalid languages filtered",
			activities: []core.Activity{
				{Timestamp: baseTime, Duration: 100, Lines: 10, Language: "", File: "/a"},
				{Timestamp: baseTime, Duration: 100, Lines: 10, Language: "unknown", File: "/b"},
				{Timestamp: baseTime, Duration: 100, Lines: 10, Language: "go", File: "/c.go"},
			},
			check: func(t *testing.T, result map[string]*LanguageAgg) {
				require.Contains(t, result, "go")
			},
		},
		{
			name: "duplicate files counted once",
			activities: []core.Activity{
				{Timestamp: baseTime, Duration: 100, Lines: 10, Language: "go", File: "/same.go"},
				{Timestamp: baseTime.Add(1 * time.Hour), Duration: 100, Lines: 10, Language: "go", File: "/same.go"},
			},
			check: func(t *testing.T, result map[string]*LanguageAgg) {
				require.Equal(t, 1, result["go"].Files.Len())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AggregateByLanguage(tt.activities)
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestAggregateByProject(t *testing.T) {
	baseTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		activities []core.Activity
		checkFunc  func(t *testing.T, agg map[string]*ProjectAgg)
	}{
		{
			name: "single project",
			activities: []core.Activity{
				{Timestamp: baseTime, Duration: 100, Lines: 10, Project: "p1", File: "/a"},
				{Timestamp: baseTime.Add(1 * time.Hour), Duration: 200, Lines: 20, Project: "p1", File: "/b"},
			},
			checkFunc: func(t *testing.T, agg map[string]*ProjectAgg) {
				require.Len(t, agg, 1)
				require.Contains(t, agg, "p1")
				require.Equal(t, 300.0, agg["p1"].Time)
				require.Equal(t, 30, agg["p1"].Lines)
				require.Equal(t, 2, agg["p1"].Files.Len())
			},
		},
		{
			name: "multiple projects",
			activities: []core.Activity{
				{Timestamp: baseTime, Duration: 100, Lines: 10, Project: "p1", File: "/a"},
				{Timestamp: baseTime, Duration: 150, Lines: 15, Project: "p2", File: "/b"},
				{Timestamp: baseTime, Duration: 200, Lines: 20, Project: "p3", File: "/c"},
			},
			checkFunc: func(t *testing.T, agg map[string]*ProjectAgg) {
				require.Len(t, agg, 3)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AggregateByProject(tt.activities)
			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestAggregateByDay(t *testing.T) {
	baseDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		activities []core.Activity
		tz         *time.Location
		check      func(t *testing.T, result map[string]*DayAgg)
	}{
		{
			name:       "empty",
			activities: []core.Activity{},
			tz:         time.UTC,
			check: func(t *testing.T, result map[string]*DayAgg) {
				require.Empty(t, result)
			},
		},
		{
			name: "single day",
			activities: []core.Activity{
				{Timestamp: baseDate.Add(10 * time.Hour), Duration: 100, Lines: 10, File: "/a", Language: "go", Project: "p1"},
				{Timestamp: baseDate.Add(14 * time.Hour), Duration: 200, Lines: 20, File: "/b", Language: "ts", Project: "p2"},
			},
			tz: time.UTC,
			check: func(t *testing.T, result map[string]*DayAgg) {
				require.Len(t, result, 1)
				require.Contains(t, result, "2025-01-15")
			},
		},
		{
			name: "multiple days",
			activities: []core.Activity{
				{Timestamp: baseDate, Duration: 100, Lines: 10, File: "/a", Language: "go", Project: "p1"},
				{Timestamp: baseDate.AddDate(0, 0, 1), Duration: 200, Lines: 20, File: "/b", Language: "ts", Project: "p2"},
				{Timestamp: baseDate.AddDate(0, 0, 2), Duration: 150, Lines: 15, File: "/c", Language: "py", Project: "p3"},
			},
			tz: time.UTC,
			check: func(t *testing.T, result map[string]*DayAgg) {
				require.Len(t, result, 3)
			},
		},
		{
			name: "midnight boundary",
			activities: []core.Activity{
				{Timestamp: baseDate.Add(23*time.Hour + 59*time.Minute), Duration: 100, Lines: 10, File: "/a"},
				{Timestamp: baseDate.AddDate(0, 0, 1), Duration: 100, Lines: 10, File: "/b"},
			},
			tz: time.UTC,
			check: func(t *testing.T, result map[string]*DayAgg) {
				require.Len(t, result, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AggregateByDay(tt.activities, tt.tz)
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestAggregateByHour(t *testing.T) {
	baseDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	activities := []core.Activity{
		{Timestamp: baseDate.Add(9 * time.Hour), Duration: 100},
		{Timestamp: baseDate.Add(9*time.Hour + 30*time.Minute), Duration: 200},
		{Timestamp: baseDate.Add(14 * time.Hour), Duration: 150},
		{Timestamp: baseDate.Add(23 * time.Hour), Duration: 50},
	}

	result := AggregateByHour(activities, time.UTC)

	require.Len(t, result, 24)
	require.Equal(t, 300.0, result[9].Duration)
	require.Equal(t, 150.0, result[14].Duration)
	require.Equal(t, 50.0, result[23].Duration)
	require.Equal(t, 0.0, result[0].Duration)
}

func TestTopLanguages(t *testing.T) {
	langAgg := map[string]*LanguageAgg{
		"go":         {Time: 1000, Lines: 100, Files: mockStringSet(5)},
		"typescript": {Time: 800, Lines: 80, Files: mockStringSet(3)},
		"python":     {Time: 600, Lines: 60, Files: mockStringSet(2)},
	}

	lifetimeHours := map[string]float64{
		"go":         1000, // Advanced
		"typescript": 300,
		"python":     100,
	}

	total := 2400.0

	result := TopLanguages(langAgg, lifetimeHours, total, 5)

	require.Len(t, result, 3)
	require.Equal(t, "go", result[0].Name)
	require.Equal(t, "Advanced", result[0].Proficiency)
	require.InDelta(t, 41.67, result[0].PercentTotal, 0.01)
}

func TestTopProjects(t *testing.T) {
	projAgg := map[string]*ProjectAgg{
		"p1": {Time: 1000, Lines: 100, Files: mockStringSet(5)},
		"p2": {Time: 800, Lines: 80, Files: mockStringSet(3)},
		"p3": {Time: 600, Lines: 60, Files: mockStringSet(2)},
	}

	projectLangs := map[string]map[string]float64{
		"p1": {"go": 600, "typescript": 400},
		"p2": {"python": 800},
		"p3": {"rust": 600},
	}

	total := 2400.0

	result := TopProjects(projAgg, projectLangs, total, 5)

	require.Len(t, result, 3)
	require.Equal(t, "p1", result[0].Name)
	require.Equal(t, "go", result[0].MainLanguage)
	require.Equal(t, "p2", result[1].Name)
	require.Equal(t, "python", result[1].MainLanguage)
}

func TestTopEditors(t *testing.T) {
	editorAgg := map[string]*EditorAgg{
		"neovim": {Time: 1500},
		"vscode": {Time: 900},
	}

	total := 2400.0

	result := TopEditors(editorAgg, total, 5)

	require.Len(t, result, 2)
	require.Equal(t, "neovim", result[0].Name)
	require.InDelta(t, 62.5, result[0].PercentTotal, 0.01)
	require.Equal(t, "vscode", result[1].Name)
	require.InDelta(t, 37.5, result[1].PercentTotal, 0.01)
}

func TestAggregation_EdgeCases(t *testing.T) {
	t.Run("zero total time", func(t *testing.T) {
		langAgg := map[string]*LanguageAgg{
			"go": {Time: 0, Lines: 0},
		}
		result := TopLanguages(langAgg, nil, 0, 5)
		require.Len(t, result, 1)
		require.Equal(t, 0.0, result[0].PercentTotal)
	})

	t.Run("limit larger than results", func(t *testing.T) {
		langAgg := map[string]*LanguageAgg{
			"go": {Time: 100, Lines: 10},
		}
		result := TopLanguages(langAgg, nil, 100, 100)
		require.Len(t, result, 1)
	})

	t.Run("empty aggregation", func(t *testing.T) {
		result := TopLanguages(map[string]*LanguageAgg{}, nil, 0, 5)
		require.Empty(t, result)
	})
}

func mockStringSet(size int) util.StringSet {
	s := util.NewStringSet()
	for i := 0; i < size; i++ {
		s.Add(fmt.Sprintf("f%d", i))
	}
	return s
}
