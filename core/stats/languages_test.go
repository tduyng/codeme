package stats

import "testing"

func TestCalculateProficiency(t *testing.T) {
	tests := []struct {
		name       string
		hoursTotal int
		expected   string
	}{
		{"beginner", 20, "Beginner"},
		{"beginner plus", 100, "Beginner+"},
		{"intermediate", 300, "Intermediate"},
		{"advanced", 600, "Advanced"},
		{"expert", 1200, "Expert"},
		{"master", 2500, "Master"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateProficiency(tt.hoursTotal)
			if result != tt.expected {
				t.Errorf("CalculateProficiency(%d) = %s, want %s", tt.hoursTotal, result, tt.expected)
			}
		})
	}
}

func TestIsTrending(t *testing.T) {
	tests := []struct {
		name     string
		growth   string
		expected bool
	}{
		{"trending up", "↗ +50%", true},
		{"positive growth", "+10%", true},
		{"stable", "→ Stable", false},
		{"declining", "↘ -20%", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTrending(tt.growth)
			if result != tt.expected {
				t.Errorf("IsTrending(%s) = %v, want %v", tt.growth, result, tt.expected)
			}
		})
	}
}

func TestIsCodeLanguage(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		expected bool
	}{
		{"go is code", "Go", true},
		{"python is code", "Python", true},
		{"json is not code", "JSON", false},
		{"yaml is not code", "yaml", false},
		{"markdown is not code", "markdown", false},
		{"javascript is code", "JavaScript", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCodeLanguage(tt.lang)
			if result != tt.expected {
				t.Errorf("IsCodeLanguage(%s) = %v, want %v", tt.lang, result, tt.expected)
			}
		})
	}
}
