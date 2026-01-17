package stats

import (
	"reflect"
	"testing"
)

func TestMapKeysToSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]bool
		expected []string
	}{
		{
			name:     "empty map",
			input:    map[string]bool{},
			expected: []string{},
		},
		{
			name: "single key",
			input: map[string]bool{
				"key1": true,
			},
			expected: []string{"key1"},
		},
		{
			name: "multiple keys sorted",
			input: map[string]bool{
				"zebra":  true,
				"apple":  true,
				"banana": true,
			},
			expected: []string{"apple", "banana", "zebra"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapKeysToSlice(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("MapKeysToSlice() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseWeekday(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		expected string
	}{
		{
			name:     "valid monday",
			date:     "2026-01-19",
			expected: "Monday",
		},
		{
			name:     "valid sunday",
			date:     "2026-01-18",
			expected: "Sunday",
		},
		{
			name:     "invalid date",
			date:     "invalid",
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseWeekday(tt.date)
			if result != tt.expected {
				t.Errorf("ParseWeekday() = %v, want %v", result, tt.expected)
			}
		})
	}
}
