package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var langMap = map[string]string{
	"go": "Go", "js": "JavaScript", "ts": "TypeScript", "py": "Python",
	"rb": "Ruby", "java": "Java", "c": "C", "cpp": "C++", "cs": "C#",
	"rs": "Rust", "php": "PHP", "swift": "Swift", "kt": "Kotlin",
	"lua": "Lua", "vim": "Vim", "sh": "Shell", "bash": "Shell",
	"md": "Markdown", "json": "JSON", "yaml": "YAML", "toml": "TOML",
	"html": "HTML", "css": "CSS", "scss": "SCSS", "sql": "SQL",
}

func DetectProject(filePath string) string {
	// Try git root first
	dir := filepath.Dir(filePath)
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	if output, err := cmd.Output(); err == nil {
		return filepath.Base(strings.TrimSpace(string(output)))
	}

	// Fallback to directory name
	abs, _ := filepath.Abs(filePath)
	parts := strings.Split(filepath.Dir(abs), string(os.PathSeparator))
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown"
}

func DetectLanguage(filePath string) string {
	ext := strings.TrimPrefix(filepath.Ext(filePath), ".")
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return ext
}

func DetectBranch(filePath string) string {
	dir := filepath.Dir(filePath)
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output))
	}
	return ""
}
