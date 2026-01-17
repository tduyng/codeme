package stats

import "strings"

var LanguageClassification = map[string]string{
	// =========================================================
	// CODE
	// =========================================================
	// A
	"ada":      "code",
	"apex":     "code",
	"assembly": "code",

	// B
	"bash":       "code",
	"beef":       "code",
	"blitzbasic": "code",

	// C
	"c":            "code",
	"clojure":      "code",
	"cobol":        "code",
	"coffeescript": "code",
	"cpp":          "code",
	"crystal":      "code",
	"csharp":       "code",

	// D
	"dart":   "code",
	"delphi": "code",
	"dlang":  "code",

	// E
	"elixir": "code",
	"elm":    "code",
	"erlang": "code",

	// F
	"fennel":  "code",
	"fortran": "code",

	// G
	"gleam":  "code",
	"go":     "code",
	"groovy": "code",

	// H
	"hack":    "code",
	"haskell": "code",

	// I
	"idris": "code",

	// J
	"java":       "code",
	"javascript": "code",
	"julia":      "code",

	// K
	"kotlin": "code",

	// L
	"lua": "code",

	// M
	"mojo":   "code",
	"matlab": "code",

	// N
	"nim": "code",
	"nix": "code",

	// O
	"objectivec":     "code",
	"objectivecplus": "code",
	"ocaml":          "code",

	// P
	"perl":       "code",
	"php":        "code",
	"powershell": "code",
	"python":     "code",

	// R
	"racket":   "code",
	"reasonml": "code",
	"ruby":     "code",
	"rust":     "code",

	// S
	"scala":    "code",
	"scheme":   "code",
	"solidity": "code",
	"swift":    "code",

	// T
	"typescript": "code",

	// V
	"v":    "code",
	"vala": "code",

	// W
	"wolfram": "code",

	// Z
	"zig": "code",

	// =========================================================
	// WEB / UI FRAMEWORK LANGUAGES (COUNT AS CODE)
	// =========================================================
	"astro":           "code",
	"svelte":          "code",
	"vue":             "code",
	"javascriptreact": "code",
	"typescriptreact": "code",

	// =========================================================
	// SHELL / SCRIPTING (COUNT AS CODE)
	// =========================================================
	"fish":     "code",
	"make":     "code",
	"makefile": "code",
	"nu":       "code",
	"sh":       "code",
	"zsh":      "code",
	"just":     "code",

	// =========================================================
	// INFRA / DSLs (COUNT AS CODE)
	// =========================================================
	"cue":       "code",
	"hcl":       "code",
	"terraform": "code",

	// =========================================================
	// CONFIGURATION (NOT IDENTITY)
	// =========================================================
	"conf":       "config",
	"dockerfile": "config",
	"env":        "config",
	"ini":        "config",
	"properties": "config",
	"toml":       "config",
	"yaml":       "config",
	"yml":        "config",

	// =========================================================
	// DATA / QUERY / SERIALIZATION (NOT IDENTITY)
	// =========================================================
	"csv":      "data",
	"graphql":  "data",
	"json":     "data",
	"json5":    "data",
	"jsonc":    "data",
	"parquet":  "data",
	"protobuf": "data",
	"sql":      "data",
	"sqlite":   "data",
	"xml":      "data",

	// =========================================================
	// MARKUP / STYLING (NOT IDENTITY)
	// =========================================================
	"css":  "markup",
	"html": "markup",
	"less": "markup",
	"scss": "markup",

	// =========================================================
	// DOCUMENTATION (NOT IDENTITY)
	// =========================================================
	"asciidoc": "doc",
	"md":       "doc",
	"markdown": "doc",
	"rst":      "doc",

	// =========================================================
	// META / TOOLING (NEVER IDENTITY)
	// =========================================================
	"bazel":     "meta",
	"cmake":     "meta",
	"gitconfig": "meta",
	"gitignore": "meta",
	"lock":      "meta",
	"meson":     "meta",
	"ninja":     "meta",
}

// CalculateProficiency returns proficiency level based on hours
func CalculateProficiency(hoursTotal int) string {
	if hoursTotal >= 2000 {
		return "Master"
	} else if hoursTotal >= 1000 {
		return "Expert"
	} else if hoursTotal >= 500 {
		return "Advanced"
	} else if hoursTotal >= 200 {
		return "Intermediate"
	} else if hoursTotal >= 50 {
		return "Beginner+"
	}
	return "Beginner"
}

// IsTrending checks if language is growing
func IsTrending(growth string) bool {
	return strings.Contains(growth, "↗") || strings.Contains(growth, "+")
}

// IsCodeLanguage checks if language should count as programming
func IsCodeLanguage(lang string) bool {
	lower := strings.ToLower(strings.TrimSpace(lang))
	lower = strings.ReplaceAll(lower, ".", "")
	kind, ok := LanguageClassification[lower]
	return ok && kind == "code"
}
