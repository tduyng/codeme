package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringSet_Add(t *testing.T) {
	s := NewStringSet()

	s.Add("a")
	s.Add("b")
	s.Add("a") // duplicate

	require.Equal(t, 2, s.Len())
	require.True(t, s.Contains("a"))
	require.True(t, s.Contains("b"))
	require.False(t, s.Contains("c"))
}

func TestStringSet_AddEmpty(t *testing.T) {
	s := NewStringSet()

	s.Add("")
	s.Add("a")

	// Empty strings should not be added
	require.Equal(t, 1, s.Len())
	require.False(t, s.Contains(""))
}

func TestStringSet_ToSlice(t *testing.T) {
	s := NewStringSet()
	s.Add("c")
	s.Add("a")
	s.Add("b")

	result := s.ToSlice()
	require.Len(t, result, 3)
	require.Contains(t, result, "a")
	require.Contains(t, result, "b")
	require.Contains(t, result, "c")
}

func TestStringSet_ToSortedSlice(t *testing.T) {
	s := NewStringSet()
	s.Add("c")
	s.Add("a")
	s.Add("b")

	result := s.ToSortedSlice()
	require.Equal(t, []string{"a", "b", "c"}, result)
}

func TestStringSet_Empty(t *testing.T) {
	s := NewStringSet()

	require.Equal(t, 0, s.Len())
	require.False(t, s.Contains("anything"))
	require.Empty(t, s.ToSlice())
	require.Empty(t, s.ToSortedSlice())
}

func TestKeys(t *testing.T) {
	m := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	keys := Keys(m)
	require.Len(t, keys, 3)
	require.Contains(t, keys, "a")
	require.Contains(t, keys, "b")
	require.Contains(t, keys, "c")
}

func TestKeys_Empty(t *testing.T) {
	m := map[string]int{}
	keys := Keys(m)
	require.Empty(t, keys)
}

func TestValues(t *testing.T) {
	m := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	values := Values(m)
	require.Len(t, values, 3)
	require.Contains(t, values, 1)
	require.Contains(t, values, 2)
	require.Contains(t, values, 3)
}

func TestMapValues(t *testing.T) {
	m := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	result := MapValues(m, func(v int) int {
		return v * 2
	})

	require.Len(t, result, 3)
	require.Contains(t, result, 2)
	require.Contains(t, result, 4)
	require.Contains(t, result, 6)
}

func TestMapValues_Empty(t *testing.T) {
	m := map[string]int{}
	result := MapValues(m, func(v int) int {
		return v * 2
	})
	require.Empty(t, result)
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		expected  []int
	}{
		{
			name:  "filter evens",
			input: []int{1, 2, 3, 4, 5, 6},
			predicate: func(n int) bool {
				return n%2 == 0
			},
			expected: []int{2, 4, 6},
		},
		{
			name:  "filter greater than 3",
			input: []int{1, 2, 3, 4, 5},
			predicate: func(n int) bool {
				return n > 3
			},
			expected: []int{4, 5},
		},
		{
			name:  "filter none",
			input: []int{1, 2, 3},
			predicate: func(n int) bool {
				return n > 10
			},
			expected: []int{},
		},
		{
			name:  "filter all",
			input: []int{1, 2, 3},
			predicate: func(n int) bool {
				return true
			},
			expected: []int{1, 2, 3},
		},
		{
			name:  "empty input",
			input: []int{},
			predicate: func(n int) bool {
				return true
			},
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Filter(tt.input, tt.predicate)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFilter_Strings(t *testing.T) {
	input := []string{"apple", "banana", "apricot", "berry"}
	result := Filter(input, func(s string) bool {
		return s[0] == 'a'
	})
	require.Equal(t, []string{"apple", "apricot"}, result)
}

func TestUniqueStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "single element",
			input:    []string{"a"},
			expected: []string{"a"},
		},
		{
			name:     "all same",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
		{
			name:     "with empty strings",
			input:    []string{"a", "", "b", ""},
			expected: []string{"a", "b"},
		},
		{
			name:     "sorted output",
			input:    []string{"z", "a", "m"},
			expected: []string{"a", "m", "z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UniqueStrings(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestCollections_EdgeCases(t *testing.T) {
	t.Run("StringSet concurrent add safety", func(t *testing.T) {
		// Note: StringSet is map[string]struct{} which is NOT concurrency-safe
		// This test just documents current behavior
		s := NewStringSet()
		s.Add("test")
		require.True(t, s.Contains("test"))
	})

	t.Run("nil map to Keys", func(t *testing.T) {
		var m map[string]int
		keys := Keys(m)
		require.Empty(t, keys)
	})

	t.Run("nil slice to Filter", func(t *testing.T) {
		var slice []int
		result := Filter(slice, func(i int) bool { return true })
		require.Empty(t, result)
	})
}

func TestStringSet_MultipleOperations(t *testing.T) {
	s := NewStringSet()

	// Build set
	items := []string{"go", "typescript", "python", "rust", "go"}
	for _, item := range items {
		s.Add(item)
	}

	require.Equal(t, 4, s.Len())

	// Convert to sorted
	sorted := s.ToSortedSlice()
	require.Equal(t, []string{"go", "python", "rust", "typescript"}, sorted)

	// Check contains
	require.True(t, s.Contains("go"))
	require.False(t, s.Contains("java"))
}

func TestGenericFunctions(t *testing.T) {
	t.Run("Keys with different types", func(t *testing.T) {
		intMap := map[int]string{1: "a", 2: "b"}
		keys := Keys(intMap)
		require.Len(t, keys, 2)
	})

	t.Run("Values with different types", func(t *testing.T) {
		stringMap := map[int]string{1: "a", 2: "b"}
		values := Values(stringMap)
		require.Len(t, values, 2)
	})

	t.Run("MapValues type conversion", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}
		result := MapValues(m, func(v int) string {
			if v == 1 {
				return "one"
			}
			return "two"
		})
		require.Contains(t, result, "one")
		require.Contains(t, result, "two")
	})
}
