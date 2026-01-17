package stats

import "testing"

func TestCalculateAchievements(t *testing.T) {
	tests := []struct {
		name     string
		stats    Stats
		validate func(t *testing.T, achievements []Achievement)
	}{
		{
			name: "no achievements",
			stats: Stats{
				TotalLines: 0,
				TotalTime:  0,
				Streak:     0,
			},
			validate: func(t *testing.T, achievements []Achievement) {
				for _, a := range achievements {
					if a.Unlocked {
						t.Errorf("Achievement %s should not be unlocked", a.ID)
					}
				}
			},
		},
		{
			name: "lines achievement",
			stats: Stats{
				TotalLines: 1500,
			},
			validate: func(t *testing.T, achievements []Achievement) {
				found := false
				for _, a := range achievements {
					if a.ID == "lines_1000" && a.Unlocked {
						found = true
					}
				}
				if !found {
					t.Error("Expected lines_1000 achievement to be unlocked")
				}
			},
		},
		{
			name: "streak achievement",
			stats: Stats{
				Streak: 7,
			},
			validate: func(t *testing.T, achievements []Achievement) {
				found := false
				for _, a := range achievements {
					if a.ID == "streak_5" && a.Unlocked {
						found = true
					}
				}
				if !found {
					t.Error("Expected streak_5 achievement to be unlocked")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateAchievements(tt.stats)
			tt.validate(t, result)
		})
	}
}
