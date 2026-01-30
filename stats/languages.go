package stats

import "strings"

var LanguageClassification = map[string]string{
	"ada": "code", "apex": "code", "assembly": "code", "bash": "code", "beef": "code",
	"blitzbasic": "code", "c": "code", "clojure": "code", "cobol": "code", "coffeescript": "code",
	"cpp": "code", "crystal": "code", "csharp": "code", "dart": "code", "delphi": "code",
	"dlang": "code", "elixir": "code", "elm": "code", "erlang": "code", "fennel": "code",
	"fortran": "code", "gleam": "code", "go": "code", "groovy": "code", "hack": "code",
	"haskell": "code", "idris": "code", "java": "code", "javascript": "code", "julia": "code",
	"kotlin": "code", "lua": "code", "mojo": "code", "matlab": "code", "nim": "code",
	"nix": "code", "objectivec": "code", "objectivecplus": "code", "ocaml": "code", "perl": "code",
	"php": "code", "powershell": "code", "python": "code", "racket": "code", "reasonml": "code",
	"ruby": "code", "rust": "code", "scala": "code", "scheme": "code", "solidity": "code",
	"swift": "code", "typescript": "code", "v": "code", "vala": "code", "wolfram": "code",
	"zig": "code", "astro": "code", "svelte": "code", "vue": "code", "javascriptreact": "code",
	"typescriptreact": "code", "fish": "code", "make": "code", "makefile": "code", "nu": "code",
	"sh": "code", "zsh": "code", "just": "code", "cue": "code", "hcl": "code",
	"terraform": "code", "conf": "config", "dockerfile": "config", "env": "config", "ini": "config",
	"properties": "config", "toml": "config", "yaml": "config", "yml": "config", "csv": "data",
	"graphql": "data", "json": "data", "json5": "data", "jsonc": "data", "parquet": "data",
	"protobuf": "data", "sql": "data", "sqlite": "data", "xml": "data", "css": "markup",
	"html": "markup", "less": "markup", "scss": "markup", "asciidoc": "doc", "md": "doc",
	"markdown": "doc", "rst": "doc", "bazel": "meta", "cmake": "meta", "gitconfig": "meta",
	"gitignore": "meta", "lock": "meta", "meson": "meta", "ninja": "meta",
}

func GetLanguageClass(lang string) string {
	lower := strings.ToLower(strings.TrimSpace(lang))
	lower = strings.ReplaceAll(lower, ".", "")
	class, ok := LanguageClassification[lower]
	if !ok {
		return "other"
	}
	return class
}

func IsCodeLanguage(lang string) bool {
	return GetLanguageClass(lang) == "code"
}

func IsValidLanguage(lang string) bool {
	lang = strings.TrimSpace(strings.ToLower(lang))
	invalidLangs := map[string]bool{
		"": true, "n/a": true, "na": true, "unknown": true,
		"undefined": true, "null": true, "none": true,
	}
	return !invalidLangs[lang]
}

func NormalizeLanguage(lang string) string {
	return strings.TrimSpace(strings.ToLower(lang))
}

func CalculateProficiency(hours float64) string {
	switch {
	case hours >= 10000:
		return "Master"
	case hours >= 5000:
		return "Expert"
	case hours >= 1000:
		return "Advanced"
	case hours >= 500:
		return "Intermediate"
	case hours >= 50:
		return "Beginner+"
	default:
		return "Beginner"
	}
}
