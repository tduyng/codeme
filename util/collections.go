// util/collections.go
package util

import "sort"

type StringSet map[string]struct{}

func NewStringSet() StringSet {
	return make(StringSet)
}

func (s StringSet) Add(v string) {
	if v != "" {
		s[v] = struct{}{}
	}
}

func (s StringSet) Contains(v string) bool {
	_, exists := s[v]
	return exists
}

func (s StringSet) Len() int {
	return len(s)
}

func (s StringSet) ToSlice() []string {
	result := make([]string, 0, len(s))
	for v := range s {
		result = append(result, v)
	}
	return result
}

func (s StringSet) ToSortedSlice() []string {
	result := s.ToSlice()
	sort.Strings(result)
	return result
}

func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func MapValues[K comparable, V any, R any](m map[K]V, fn func(V) R) []R {
	result := make([]R, 0, len(m))
	for _, v := range m {
		result = append(result, fn(v))
	}
	return result
}

func Filter[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

func UniqueStrings(input []string) []string {
	set := NewStringSet()
	for _, s := range input {
		set.Add(s)
	}
	return set.ToSortedSlice()
}
