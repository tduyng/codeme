package util

import (
	"sort"
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
