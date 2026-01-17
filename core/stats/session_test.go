package stats

import (
	"testing"
	"time"
)

func TestDetectSessionGap(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		prev     time.Time
		current  time.Time
		expected bool
	}{
		{
			name:     "zero prev time",
			prev:     time.Time{},
			current:  now,
			expected: false,
		},
		{
			name:     "within 15 minutes",
			prev:     now,
			current:  now.Add(10 * time.Minute),
			expected: true,
		},
		{
			name:     "exactly 15 minutes",
			prev:     now,
			current:  now.Add(15 * time.Minute),
			expected: true,
		},
		{
			name:     "over 15 minutes",
			prev:     now,
			current:  now.Add(20 * time.Minute),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectSessionGap(tt.prev, tt.current)
			if result != tt.expected {
				t.Errorf("DetectSessionGap() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateAvgSessionLength(t *testing.T) {
	tests := []struct {
		name     string
		sessions []Session
		expected int64
	}{
		{
			name:     "no sessions",
			sessions: []Session{},
			expected: 0,
		},
		{
			name: "single session",
			sessions: []Session{
				{Duration: 3600},
			},
			expected: 3600,
		},
		{
			name: "multiple sessions",
			sessions: []Session{
				{Duration: 3600},
				{Duration: 7200},
				{Duration: 1800},
			},
			expected: 4200, // (3600 + 7200 + 1800) / 3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateAvgSessionLength(tt.sessions)
			if result != tt.expected {
				t.Errorf("CalculateAvgSessionLength() = %d, want %d", result, tt.expected)
			}
		})
	}
}
