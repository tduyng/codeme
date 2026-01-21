package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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
	return "n/a"
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

var langMap = map[string]string{
	"go":                   "Go",
	"js":                   "JavaScript",
	"ts":                   "TypeScript",
	"jsx":                  "JSX",
	"tsx":                  "TSX",
	"py":                   "Python",
	"rb":                   "Ruby",
	"java":                 "Java",
	"c":                    "C",
	"h":                    "C Header",
	"cpp":                  "C++",
	"hpp":                  "C++ Header",
	"cs":                   "C#",
	"rs":                   "Rust",
	"php":                  "PHP",
	"swift":                "Swift",
	"kt":                   "Kotlin",
	"lua":                  "Lua",
	"vim":                  "Vim script",
	"sh":                   "Shell",
	"bash":                 "Bash",
	"zsh":                  "Zsh",
	"fish":                 "Fish",
	"md":                   "Markdown",
	"json":                 "JSON",
	"yaml":                 "YAML",
	"yml":                  "YAML",
	"toml":                 "TOML",
	"html":                 "HTML",
	"htm":                  "HTML",
	"css":                  "CSS",
	"scss":                 "SCSS",
	"sass":                 "Sass",
	"less":                 "Less",
	"sql":                  "SQL",
	"mysql":                "MySQL",
	"psql":                 "PostgreSQL",
	"plsql":                "PL/SQL",
	"jl":                   "Julia",
	"rkt":                  "Racket",
	"clj":                  "Clojure",
	"cljs":                 "ClojureScript",
	"scm":                  "Scheme",
	"lisp":                 "Lisp",
	"el":                   "Emacs Lisp",
	"erl":                  "Erlang",
	"ex":                   "Elixir",
	"exs":                  "Elixir",
	"hs":                   "Haskell",
	"nim":                  "Nim",
	"crystal":              "Crystal",
	"scala":                "Scala",
	"sbt":                  "SBT",
	"fs":                   "F#",
	"fsi":                  "F# Script",
	"ml":                   "OCaml",
	"mli":                  "OCaml Interface",
	"re":                   "Reason",
	"dart":                 "Dart",
	"flutter":              "Flutter",
	"web":                  "WebAssembly",
	"wat":                  "WAT",
	"zig":                  "Zig",
	"v":                    "V",
	"odin":                 "Odin",
	"bun":                  "Bun",
	"deno":                 "Deno",
	"bunjs":                "Bun JavaScript",
	"vue":                  "Vue",
	"svelte":               "Svelte",
	"svelte-":              "Svelte",
	"astro":                "Astro",
	"solid":                "SolidJS",
	"qml":                  "QML",
	"rsh":                  "ReScript",
	"res":                  "ReScript",
	"elm":                  "Elm",
	"purs":                 "PureScript",
	"gleam":                "Gleam",
	"mojo":                 "Mojo",
	"vala":                 "Vala",
	"d":                    "D",
	"pas":                  "Pascal",
	"pascal":               "Pascal",
	"ada":                  "Ada",
	"fortran":              "Fortran",
	"77":                   "Fortran 77",
	"f90":                  "Fortran 90",
	"asm":                  "Assembly",
	"s":                    "Assembly",
	"nasm":                 "NASM",
	"objdump":              "Object Dump",
	"bat":                  "Batch",
	"cmd":                  "CMD",
	"ps1":                  "PowerShell",
	"psm1":                 "PowerShell Module",
	"ahk":                  "AutoHotkey",
	"applescript":          "AppleScript",
	"scpt":                 "AppleScript",
	"tex":                  "LaTeX",
	"latex":                "LaTeX",
	"rnoweb":               "Rnoweb",
	"rtex":                 "R TeX",
	"xml":                  "XML",
	"xhtml":                "XHTML",
	"svg":                  "SVG",
	"graphql":              "GraphQL",
	"gql":                  "GraphQL",
	"proto":                "Protocol Buffers",
	"thrift":               "Thrift",
	"capnp":                "Cap'n Proto",
	"dockerfile":           "Dockerfile",
	"docker":               "Docker",
	"dockerignore":         "Docker Ignore",
	"compose":              "Docker Compose",
	"editorconfig":         "EditorConfig",
	"gitignore":            "Gitignore",
	"gitattributes":        "Git Attributes",
	"prettierrc":           "Prettier RC",
	"eslintrc":             "ESLint RC",
	"tsconfig":             "TSConfig",
	"jsconfig":             "JSConfig",
	"package.json":         "Package JSON",
	"package-lock.json":    "Package Lock",
	"pnpm-lock.yaml":       "PNPM Lock",
	"yarn.lock":            "Yarn Lock",
	"cargotoml":            "Cargo TOML",
	"cargo.lock":           "Cargo Lock",
	"composer.json":        "Composer JSON",
	"pubspec.yaml":         "Pubspec YAML",
	"build.gradle":         "Gradle",
	"gradle.properties":    "Gradle Properties",
	"pom.xml":              "Maven POM",
	"makefile":             "Makefile",
	"make":                 "Makefile",
	"meson.build":          "Meson",
	"cmake":                "CMake",
	"ninja":                "Ninja",
	"vcpkg.json":           "VCPKG",
	"justfile":             "Justfile",
	"taskfile.yml":         "Taskfile",
	"wolfram":              "Wolfram Language",
	"wl":                   "Wolfram Language",
	"mathematica":          "Mathematica",
	"nb":                   "Mathematica Notebook",
	"ipynb":                "Jupyter Notebook",
	"rmd":                  "R Markdown",
	"quarto":               "Quarto",
	"org":                  "Org-mode",
	"rst":                  "reStructuredText",
	"adoc":                 "AsciiDoc",
	"txt":                  "Plain Text",
	"log":                  "Log File",
	"cfg":                  "Config",
	"conf":                 "Config",
	"ini":                  "INI",
	"env":                  ".env",
	"graphql.config.json":  "GraphQL Config",
	"tailwind.config.js":   "Tailwind Config",
	"next.config.js":       "Next.js Config",
	"nuxt.config.ts":       "Nuxt Config",
	"vite.config.ts":       "Vite Config",
	"rollup.config.js":     "Rollup Config",
	"webpack.config.js":    "Webpack Config",
	"jest.config.js":       "Jest Config",
	"cypress.config.ts":    "Cypress Config",
	"playwright.config.ts": "Playwright Config",
}
