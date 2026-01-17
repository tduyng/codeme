package stats

import (
	"reflect"
	"testing"
	"time"
)

func TestCalculateMostActiveHour(t *testing.T) {
	tests := []struct {
		name     string
		hourly   map[int]int
		expected int
	}{
		{
			name:     "empty map",
			hourly:   map[int]int{},
			expected: 0,
		},
		{
			name: "single hour",
			hourly: map[int]int{
				14: 10,
			},
			expected: 14,
		},
		{
			name: "multiple hours",
			hourly: map[int]int{
				9:  5,
				14: 15,
				18: 10,
			},
			expected: 14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateMostActiveHour(tt.hourly)
			if result != tt.expected {
				t.Errorf("CalculateMostActiveHour() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestCalculatePeakHours(t *testing.T) {
	tests := []struct {
		name     string
		hourly   map[int]int
		count    int
		expected []int
	}{
		{
			name:     "empty map",
			hourly:   map[int]int{},
			count:    3,
			expected: []int{},
		},
		{
			name: "top 3 hours",
			hourly: map[int]int{
				9:  5,
				10: 3,
				14: 15,
				16: 10,
				18: 8,
			},
			count:    3,
			expected: []int{14, 16, 18}, // Top 3 by count: 14(15), 16(10), 18(8), sorted by hour
		},
		{
			name: "request more than available",
			hourly: map[int]int{
				9:  5,
				14: 10,
			},
			count:    5,
			expected: []int{9, 14},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePeakHours(tt.hourly, tt.count)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("CalculatePeakHours() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateMostActiveDay(t *testing.T) {
	tests := []struct {
		name         string
		dayTime      map[time.Weekday]int64
		expectedDay  string
		expectedTime int64
	}{
		{
			name:         "empty map",
			dayTime:      map[time.Weekday]int64{},
			expectedDay:  "",
			expectedTime: 0,
		},
		{
			name: "monday most active",
			dayTime: map[time.Weekday]int64{
				time.Monday:    10000,
				time.Wednesday: 5000,
				time.Friday:    8000,
			},
			expectedDay:  "Monday",
			expectedTime: 10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			day, timeVal := CalculateMostActiveDay(tt.dayTime)
			if day != tt.expectedDay || timeVal != tt.expectedTime {
				t.Errorf("CalculateMostActiveDay() = (%s, %d), want (%s, %d)",
					day, timeVal, tt.expectedDay, tt.expectedTime)
			}
		})
	}
}
