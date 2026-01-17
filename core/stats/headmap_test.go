package stats

import (
	"testing"
	"time"
)

func TestGenerateWeeklyHeatmap(t *testing.T) {
	now := time.Now()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	tests := []struct {
		name     string
		daily    map[string]DailyStat
		weeks    int
		validate func(t *testing.T, heatmap []HeatmapDay)
	}{
		{
			name:  "empty data",
			daily: map[string]DailyStat{},
			weeks: 1,
			validate: func(t *testing.T, heatmap []HeatmapDay) {
				if len(heatmap) != 7 {
					t.Errorf("Expected 7 days for 1 week, got %d", len(heatmap))
				}
			},
		},
		{
			name: "with activity",
			daily: map[string]DailyStat{
				today:     {Lines: 100, Time: 3600},
				yesterday: {Lines: 50, Time: 1800},
			},
			weeks: 2,
			validate: func(t *testing.T, heatmap []HeatmapDay) {
				if len(heatmap) != 14 {
					t.Errorf("Expected 14 days for 2 weeks, got %d", len(heatmap))
				}
				// Check that at least one day has activity
				hasActivity := false
				for _, day := range heatmap {
					if day.Lines > 0 {
						hasActivity = true
						break
					}
				}
				if !hasActivity {
					t.Error("Expected at least one day with activity")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateWeeklyHeatmap(tt.daily, tt.weeks)
			tt.validate(t, result)
		})
	}
}
