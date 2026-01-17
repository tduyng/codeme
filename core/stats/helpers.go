package stats

import (
	"sort"
	"time"
)

// MapKeysToSlice converts map keys to sorted string slice
func MapKeysToSlice(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ParseWeekday returns weekday name from date string
func ParseWeekday(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "Unknown"
	}
	return t.Weekday().String()
}

// CalculateLanguageGrowth calculates week-over-week growth for a language
func CalculateLanguageGrowth(language string, dailyActivity map[string]DailyStat, thisWeekStart, lastWeekStart, lastWeekEnd string) string {
	// Similar to CalculateGrowth but would need language-specific daily data
	// For now, return a placeholder - would need more detailed tracking
	return ""
}
