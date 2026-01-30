package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetector_DetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		// Golden paths
		{"go file", "/project/main.go", "go"},
		{"typescript", "/src/app.ts", "typescript"},
		{"python", "script.py", "python"},

		// Edge cases
		{"no extension", "README", ""},
		{"hidden file", ".gitignore", "gitignore"},
		{"empty string", "", ""},
		{"multiple dots", "config.test.js", "javascript"},
		{"uppercase ext", "Main.GO", "GO"}, // Function preserves case
		{"path with spaces", "/my files/test.rb", "ruby"},

		// Config files
		{"dockerfile", "Dockerfile", ""},
		{"makefile", "Makefile", ""},

		// Unknown extension
		{"unknown", "file.xyz", "xyz"},
	}

	d := NewDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := d.DetectLanguage(tt.path)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestDetector_DetectLanguage_Caching(t *testing.T) {
	d := NewDetector()

	path := "/test/main.go"

	// First call
	lang1 := d.DetectLanguage(path)

	// Second call should use cache
	lang2 := d.DetectLanguage(path)

	require.Equal(t, lang1, lang2)
	require.Equal(t, "go", lang1)
}

func TestDetector_DetectProject(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectEmpty bool
	}{
		{"simple path", "/home/user/project/main.go", false},
		{"nested", "/home/user/projects/myapp/src/app.go", false},
		{"root file", "/main.go", true}, // Root files may not have project
		{"empty", "", false},
		{"relative", "main.go", false},
	}

	d := NewDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := d.DetectProject(tt.path)
			if tt.expectEmpty {
				require.Empty(t, result)
			} else {
				require.NotEmpty(t, result)
			}
		})
	}
}

func TestDetector_DetectProject_Caching(t *testing.T) {
	d := NewDetector()

	path := "/test/project/main.go"

	proj1 := d.DetectProject(path)
	proj2 := d.DetectProject(path)

	require.Equal(t, proj1, proj2)
	require.NotEmpty(t, proj1)
}

func TestLanguageFromExtension(t *testing.T) {
	tests := []struct {
		ext      string
		expected string
	}{
		// Common languages
		{"go", "go"},
		{"js", "javascript"},
		{"ts", "typescript"},
		{"py", "python"},

		// Edge cases
		{"", ""},
		{"unknown", "unknown"},

		// Case sensitivity
		{"GO", "GO"}, // Function doesn't lowercase

		// Markup
		{"html", "html"},
		{"css", "css"},

		// Config
		{"json", "json"},
		{"yaml", "yaml"},
		{"toml", "toml"},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := languageFromExtension(tt.ext)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestDetector_ConcurrentAccess(t *testing.T) {
	d := NewDetector()

	paths := []string{
		"/project/main.go",
		"/project/app.ts",
		"/project/script.py",
	}

	// Simulate concurrent access
	done := make(chan bool)

	for _, path := range paths {
		go func(p string) {
			_ = d.DetectLanguage(p)
			_ = d.DetectProject(p)
			done <- true
		}(path)
	}

	for range paths {
		<-done
	}

	// Verify cache works after concurrent access
	for _, path := range paths {
		lang := d.DetectLanguage(path)
		require.NotEmpty(t, lang)
	}
}
