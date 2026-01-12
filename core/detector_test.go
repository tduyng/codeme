package core

import "testing"

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		file string
		want string
	}{
		{"test.go", "Go"},
		{"test.js", "JavaScript"},
		{"test.py", "Python"},
		{"test.unknown", "unknown"},
	}

	for _, tt := range tests {
		got := DetectLanguage(tt.file)
		if got != tt.want {
			t.Errorf("DetectLanguage(%q) = %q, want %q", tt.file, got, tt.want)
		}
	}
}
