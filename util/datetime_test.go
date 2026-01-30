package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStartOfDay(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		tz       *time.Location
		expected time.Time
	}{
		{
			name:     "UTC noon",
			input:    time.Date(2025, 1, 15, 12, 30, 45, 0, time.UTC),
			tz:       time.UTC,
			expected: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "UTC midnight",
			input:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			tz:       time.UTC,
			expected: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "UTC just before midnight",
			input:    time.Date(2025, 1, 15, 23, 59, 59, 999999999, time.UTC),
			tz:       time.UTC,
			expected: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "nil timezone defaults to UTC",
			input:    time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
			tz:       nil,
			expected: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartOfDay(tt.input, tt.tz)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestStartOfDay_Timezone(t *testing.T) {
	est, _ := time.LoadLocation("America/New_York")
	pst, _ := time.LoadLocation("America/Los_Angeles")

	// Same UTC time, different timezones
	utcTime := time.Date(2025, 1, 15, 5, 0, 0, 0, time.UTC)

	t.Run("EST", func(t *testing.T) {
		result := StartOfDay(utcTime, est)
		// 5 AM UTC = midnight EST
		expected := time.Date(2025, 1, 15, 0, 0, 0, 0, est)
		require.Equal(t, expected, result)
	})

	t.Run("PST", func(t *testing.T) {
		result := StartOfDay(utcTime, pst)
		// 5 AM UTC = 9 PM PST previous day, so start of day is Jan 14
		expected := time.Date(2025, 1, 14, 0, 0, 0, 0, pst)
		require.Equal(t, expected, result)
	})
}

func TestStartOfWeek(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		tz       *time.Location
		expected time.Time
	}{
		{
			name:     "Monday",
			input:    time.Date(2025, 1, 13, 12, 0, 0, 0, time.UTC),
			tz:       time.UTC,
			expected: time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC), // Monday
		},
		{
			name:     "Sunday",
			input:    time.Date(2025, 1, 12, 12, 0, 0, 0, time.UTC),
			tz:       time.UTC,
			expected: time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC), // previous Monday
		},
		{
			name:     "Saturday",
			input:    time.Date(2025, 1, 18, 12, 0, 0, 0, time.UTC),
			tz:       time.UTC,
			expected: time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC), // Monday
		},
		{
			name:     "Friday",
			input:    time.Date(2025, 1, 17, 23, 59, 59, 0, time.UTC),
			tz:       time.UTC,
			expected: time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC), // Monday
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartOfWeek(tt.input, tt.tz)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestStartOfWeek_ISO(t *testing.T) {
	tz, _ := time.LoadLocation("Europe/Paris")

	// Sunday
	tm := time.Date(2026, 2, 1, 12, 0, 0, 0, tz)
	start := StartOfWeek(tm, tz)

	require.Equal(t, "2026-01-26", DateString(start, tz))
}

func TestStartOfMonth(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		tz       *time.Location
		expected time.Time
	}{
		{
			name:     "mid month",
			input:    time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
			tz:       time.UTC,
			expected: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "first day",
			input:    time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			tz:       time.UTC,
			expected: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "last day",
			input:    time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
			tz:       time.UTC,
			expected: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "February leap year",
			input:    time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC),
			tz:       time.UTC,
			expected: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartOfMonth(tt.input, tt.tz)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestDateString(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		tz       *time.Location
		expected string
	}{
		{
			name:     "UTC time",
			input:    time.Date(2025, 1, 15, 12, 30, 0, 0, time.UTC),
			tz:       time.UTC,
			expected: "2025-01-15",
		},
		{
			name:     "with nanoseconds",
			input:    time.Date(2025, 1, 15, 12, 30, 45, 123456789, time.UTC),
			tz:       time.UTC,
			expected: "2025-01-15",
		},
		{
			name:     "nil timezone",
			input:    time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
			tz:       nil,
			expected: "2025-01-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DateString(tt.input, tt.tz)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestDateString_Timezone(t *testing.T) {
	est, _ := time.LoadLocation("America/New_York")

	// 11 PM EST = 4 AM UTC next day
	estTime := time.Date(2025, 1, 15, 23, 0, 0, 0, est)

	t.Run("EST date string", func(t *testing.T) {
		result := DateString(estTime, est)
		require.Equal(t, "2025-01-15", result)
	})

	t.Run("UTC date string", func(t *testing.T) {
		result := DateString(estTime, time.UTC)
		require.Equal(t, "2025-01-16", result)
	})
}

func TestParseWeekday(t *testing.T) {
	tests := []struct {
		dateStr  string
		tz       *time.Location
		expected string
	}{
		{"2025-01-15", time.UTC, "Wednesday"},
		{"2025-01-12", time.UTC, "Sunday"},
		{"2025-01-18", time.UTC, "Saturday"},
		{"2025-01-13", time.UTC, "Monday"},
		{"invalid", time.UTC, ""},
	}

	for _, tt := range tests {
		t.Run(tt.dateStr, func(t *testing.T) {
			result := ParseWeekday(tt.dateStr, tt.tz)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestDateFrom(t *testing.T) {
	tests := []struct {
		name     string
		dateStr  string
		tz       *time.Location
		expected time.Time
	}{
		{
			name:     "valid date",
			dateStr:  "2025-01-15",
			tz:       time.UTC,
			expected: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "invalid date returns zero",
			dateStr:  "invalid",
			tz:       time.UTC,
			expected: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DateFrom(tt.dateStr, tt.tz)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestDaysBetween(t *testing.T) {
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected int
	}{
		{
			name:     "same day",
			start:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 1, 15, 23, 59, 59, 0, time.UTC),
			expected: 0,
		},
		{
			name:     "one day",
			start:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
		{
			name:     "one week",
			start:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 1, 22, 0, 0, 0, 0, time.UTC),
			expected: 7,
		},
		{
			name:     "negative (end before start)",
			start:    time.Date(2025, 1, 22, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			expected: -7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DaysBetween(tt.start, tt.end)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestIsToday(t *testing.T) {
	now := time.Now().In(time.UTC)
	todayStart := StartOfDay(now, time.UTC)

	tests := []struct {
		name     string
		input    time.Time
		tz       *time.Location
		expected bool
	}{
		{
			name:     "now is today",
			input:    now,
			tz:       time.UTC,
			expected: true,
		},
		{
			name:     "start of today",
			input:    todayStart,
			tz:       time.UTC,
			expected: true,
		},
		{
			name:     "yesterday",
			input:    todayStart.AddDate(0, 0, -1),
			tz:       time.UTC,
			expected: false,
		},
		{
			name:     "tomorrow",
			input:    todayStart.AddDate(0, 0, 1),
			tz:       time.UTC,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsToday(tt.input, tt.tz)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestGetWeekBounds(t *testing.T) {
	input := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC) // Wednesday

	start, end := GetWeekBounds(input, time.UTC)

	require.Equal(t, time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC), start) // Monday
	require.Equal(t, time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC), end)   // Next Monday
	require.Equal(t, 7, DaysBetween(start, end))
}

func TestGetMonthBounds(t *testing.T) {
	input := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	start, end := GetMonthBounds(input, time.UTC)

	require.Equal(t, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), start)
	require.Equal(t, time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC), end)
}

func TestGetMonthBounds_EdgeCases(t *testing.T) {
	t.Run("February leap year", func(t *testing.T) {
		input := time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC)
		start, end := GetMonthBounds(input, time.UTC)
		require.Equal(t, time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), start)
		require.Equal(t, time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), end)
	})

	t.Run("December to January", func(t *testing.T) {
		input := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
		start, end := GetMonthBounds(input, time.UTC)
		require.Equal(t, time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC), start)
		require.Equal(t, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), end)
	})
}

func TestDefaultLocation(t *testing.T) {
	est, _ := time.LoadLocation("America/New_York")

	tests := []struct {
		name     string
		input    *time.Location
		expected *time.Location
	}{
		{
			name:     "non-nil location",
			input:    est,
			expected: est,
		},
		{
			name:     "nil location",
			input:    nil,
			expected: time.UTC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DefaultLocation(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestDatetime_CrossTimezone(t *testing.T) {
	est, _ := time.LoadLocation("America/New_York")
	pst, _ := time.LoadLocation("America/Los_Angeles")

	// 11:30 PM EST on Jan 15 = 8:30 PM PST same day = 4:30 AM UTC Jan 16
	estTime := time.Date(2025, 1, 15, 23, 30, 0, 0, est)

	t.Run("EST date", func(t *testing.T) {
		require.Equal(t, "2025-01-15", DateString(estTime, est))
	})

	t.Run("PST date", func(t *testing.T) {
		require.Equal(t, "2025-01-15", DateString(estTime, pst))
	})

	t.Run("UTC date", func(t *testing.T) {
		require.Equal(t, "2025-01-16", DateString(estTime, time.UTC))
	})
}
