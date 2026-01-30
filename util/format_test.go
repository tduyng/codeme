package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		seconds  float64
		expected string
	}{
		// Seconds
		{"zero", 0, "0s"},
		{"one second", 1, "1s"},
		{"59 seconds", 59, "59s"},
		{"fractional seconds", 45.7, "46s"},

		// Minutes
		{"one minute", 60, "1m"},
		{"two minutes", 120, "2m"},
		{"59 minutes", 3540, "59m"},
		{"90 seconds", 90, "1m"},

		// Hours
		{"one hour", 3600, "1h"},
		{"one hour one minute", 3660, "1h 1m"},
		{"two hours", 7200, "2h"},
		{"two hours 30 minutes", 9000, "2h 30m"},
		{"exact hour no minutes", 7200, "2h"},

		// Large durations
		{"10 hours", 36000, "10h"},
		{"24 hours", 86400, "24h"},
		{"100 hours 15 minutes", 360900, "100h 15m"},

		// Edge cases
		{"59 minutes 59 seconds", 3599, "59m"},
		{"1 hour 59 minutes", 7140, "1h 59m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.seconds)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		// Bytes
		{"zero", 0, "0 B"},
		{"one byte", 1, "1 B"},
		{"1023 bytes", 1023, "1023 B"},

		// Kilobytes
		{"1 KB", 1024, "1.0 KB"},
		{"1.5 KB", 1536, "1.5 KB"},
		{"1023 KB", 1047552, "1023.0 KB"},

		// Megabytes
		{"1 MB", 1048576, "1.0 MB"},
		{"1.5 MB", 1572864, "1.5 MB"},
		{"100 MB", 104857600, "100.0 MB"},

		// Gigabytes
		{"1 GB", 1073741824, "1.0 GB"},
		{"2.5 GB", 2684354560, "2.5 GB"},

		// Terabytes
		{"1 TB", 1099511627776, "1.0 TB"},
		{"10 TB", 10995116277760, "10.0 TB"},

		// Edge cases
		{"1024 exactly", 1024, "1.0 KB"},
		{"large number", 1234567890, "1.1 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDuration_EdgeCases(t *testing.T) {
	t.Run("negative duration", func(t *testing.T) {
		result := FormatDuration(-100)
		require.Contains(t, result, "-")
	})

	t.Run("very small fraction", func(t *testing.T) {
		result := FormatDuration(0.1)
		require.Equal(t, "0s", result)
	})

	t.Run("very large duration", func(t *testing.T) {
		result := FormatDuration(1000000)
		require.Contains(t, result, "h")
	})

	t.Run("exactly 60 seconds", func(t *testing.T) {
		result := FormatDuration(60)
		require.Equal(t, "1m", result)
	})

	t.Run("exactly 3600 seconds", func(t *testing.T) {
		result := FormatDuration(3600)
		require.Equal(t, "1h", result)
	})
}

func TestFormatBytes_EdgeCases(t *testing.T) {
	t.Run("negative bytes", func(t *testing.T) {
		result := FormatBytes(-1024)
		require.Contains(t, result, "-")
	})

	t.Run("exactly 1024", func(t *testing.T) {
		result := FormatBytes(1024)
		require.Equal(t, "1.0 KB", result)
	})

	t.Run("max int64", func(t *testing.T) {
		result := FormatBytes(9223372036854775807)
		// Should not panic and should be in TB or PB range
		require.Contains(t, result, "B") // Has some byte unit
	})
}

func TestFormatting_Precision(t *testing.T) {
	t.Run("bytes precision one decimal", func(t *testing.T) {
		result := FormatBytes(1536)
		require.Equal(t, "1.5 KB", result)
	})

	t.Run("bytes precision rounds", func(t *testing.T) {
		result := FormatBytes(1587)
		require.Equal(t, "1.5 KB", result)
	})

	t.Run("duration truncates minutes", func(t *testing.T) {
		// 90.9 seconds should be 1 minute (90 / 60 = 1)
		result := FormatDuration(90.9)
		require.Equal(t, "1m", result)
	})
}

func TestFormatDuration_RealWorldExamples(t *testing.T) {
	tests := []struct {
		name     string
		seconds  float64
		expected string
	}{
		{"quick edit", 30, "30s"},
		{"short session", 900, "15m"},
		{"medium session", 3600, "1h"},
		{"long session", 7200, "2h"},
		{"work day", 28800, "8h"},
		{"marathon session", 43200, "12h"},
		{"coding with breaks", 14400, "4h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.seconds)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatBytes_RealWorldExamples(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"small config", 1024, "1.0 KB"},
		{"medium document", 102400, "100.0 KB"},
		{"database", 10485760, "10.0 MB"},
		{"video file", 1073741824, "1.0 GB"},
		{"large backup", 10737418240, "10.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			require.Equal(t, tt.expected, result)
		})
	}
}
