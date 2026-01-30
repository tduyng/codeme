package stats

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetLanguageClass(t *testing.T) {
	tests := []struct {
		lang     string
		expected string
	}{
		// Code languages
		{"go", "code"},
		{"python", "code"},
		{"javascript", "code"},
		{"typescript", "code"},
		{"rust", "code"},
		{"java", "code"},

		// Config files
		{"yaml", "config"},
		{"toml", "config"},
		{"ini", "config"},
		{"conf", "config"},

		// Data formats
		{"json", "data"},
		{"xml", "data"},
		{"csv", "data"},
		{"sql", "data"},

		// Markup
		{"html", "markup"},
		{"css", "markup"},
		{"scss", "markup"},

		// Documentation
		{"markdown", "doc"},
		{"md", "doc"},
		{"rst", "doc"},

		// Edge cases
		{"", "other"},
		{"unknown", "other"},
		{"xyz", "other"},

		// Case insensitivity
		{"Go", "code"},
		{"PYTHON", "code"},
		{"JavaScript", "code"},

		// With dots
		{".go", "code"},
		{".yaml", "config"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			result := GetLanguageClass(tt.lang)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestIsCodeLanguage(t *testing.T) {
	tests := []struct {
		lang     string
		expected bool
	}{
		{"go", true},
		{"python", true},
		{"javascript", true},
		{"yaml", false},
		{"json", false},
		{"html", false},
		{"markdown", false},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			result := IsCodeLanguage(tt.lang)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidLanguage(t *testing.T) {
	tests := []struct {
		lang     string
		expected bool
	}{
		{"go", true},
		{"python", true},
		{"", false},
		{"n/a", false},
		{"na", false},
		{"unknown", false},
		{"undefined", false},
		{"null", false},
		{"none", false},

		// Case insensitive
		{"N/A", false},
		{"Unknown", false},
		{"NONE", false},

		// With whitespace
		{"  go  ", true},
		{"  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			result := IsValidLanguage(tt.lang)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Go", "go"},
		{"PYTHON", "python"},
		{"JavaScript", "javascript"},
		{"  rust  ", "rust"},
		{"", ""},
		{"  ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeLanguage(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateProficiency(t *testing.T) {
	tests := []struct {
		hours    float64
		expected string
	}{
		{0, "Beginner"},
		{10, "Beginner"},
		{49, "Beginner"},
		{50, "Beginner+"},
		{100, "Beginner+"},
		{499, "Beginner+"},
		{500, "Intermediate"},
		{999, "Intermediate"},
		{1000, "Advanced"},
		{4999, "Advanced"},
		{5000, "Expert"},
		{9999, "Expert"},
		{10000, "Master"},
		{50000, "Master"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := CalculateProficiency(tt.hours)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestLanguageClassification_Coverage(t *testing.T) {
	// Ensure all classifications are accounted for
	classes := map[string]bool{
		"code":   false,
		"config": false,
		"data":   false,
		"markup": false,
		"doc":    false,
		"meta":   false,
	}

	for _, class := range LanguageClassification {
		classes[class] = true
	}

	for class, found := range classes {
		require.True(t, found, "class %s should be in use", class)
	}
}

func TestLanguageEdgeCases(t *testing.T) {
	t.Run("nil or empty classification", func(t *testing.T) {
		result := GetLanguageClass("")
		require.Equal(t, "other", result)
	})

	t.Run("normalization consistency", func(t *testing.T) {
		variants := []string{"go", "Go", "GO", "  go  ", "  GO  "}
		expected := "go"

		for _, v := range variants {
			result := NormalizeLanguage(v)
			require.Equal(t, expected, result)
		}
	})

	t.Run("proficiency boundaries", func(t *testing.T) {
		// Test exact boundaries
		require.Equal(t, "Beginner+", CalculateProficiency(50))
		require.Equal(t, "Intermediate", CalculateProficiency(500))
		require.Equal(t, "Advanced", CalculateProficiency(1000))
		require.Equal(t, "Expert", CalculateProficiency(5000))
		require.Equal(t, "Master", CalculateProficiency(10000))
	})
}
