package util

import "sort"

// StringSet is a simple set of strings
type StringSet map[string]struct{}

// NewStringSet returns an empty StringSet
func NewStringSet() StringSet {
	return make(StringSet)
}

func (s StringSet) Len() int {
	return len(s)
}

// Add inserts a string into the set
func (s StringSet) Add(v string) {
	if v != "" {
		s[v] = struct{}{}
	}
}

// ToSortedSlice converts the set into a sorted slice
func (s StringSet) ToSortedSlice() []string {
	out := make([]string, 0, len(s))
	for v := range s {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}
