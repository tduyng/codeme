// core/detector.go
package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type Detector struct {
	langCache    sync.Map
	projectCache sync.Map
}

func NewDetector() *Detector {
	return &Detector{}
}

func (d *Detector) DetectLanguage(path string) string {
	if cached, ok := d.langCache.Load(path); ok {
		return cached.(string)
	}

	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	lang := languageFromExtension(ext)

	d.langCache.Store(path, lang)
	return lang
}

func (d *Detector) DetectProject(path string) string {
	if cached, ok := d.projectCache.Load(path); ok {
		return cached.(string)
	}

	var project string

	dir := filepath.Dir(path)
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir

	if output, err := cmd.Output(); err == nil {
		project = filepath.Base(strings.TrimSpace(string(output)))
	} else {
		abs, _ := filepath.Abs(path)
		parts := strings.Split(filepath.Dir(abs), string(os.PathSeparator))
		if len(parts) > 0 {
			project = parts[len(parts)-1]
		} else {
			project = "unknown"
		}
	}

	d.projectCache.Store(path, project)
	return project
}

func languageFromExtension(ext string) string {
	lang, ok := langMap[ext]
	if !ok {
		return ext
	}
	return lang
}

var langMap = map[string]string{
	"go":                   "go",
	"js":                   "javascript",
	"ts":                   "typescript",
	"jsx":                  "jsx",
	"tsx":                  "tsx",
	"py":                   "python",
	"rb":                   "ruby",
	"java":                 "java",
	"c":                    "c",
	"h":                    "c header",
	"cpp":                  "c++",
	"hpp":                  "c++ header",
	"cs":                   "c#",
	"rs":                   "rust",
	"php":                  "php",
	"swift":                "swift",
	"kt":                   "kotlin",
	"lua":                  "lua",
	"vim":                  "vim script",
	"sh":                   "shell",
	"bash":                 "bash",
	"zsh":                  "zsh",
	"fish":                 "fish",
	"md":                   "markdown",
	"json":                 "json",
	"yaml":                 "yaml",
	"yml":                  "yaml",
	"toml":                 "toml",
	"html":                 "html",
	"htm":                  "html",
	"css":                  "css",
	"scss":                 "scss",
	"sass":                 "sass",
	"less":                 "less",
	"sql":                  "sql",
	"mysql":                "mysql",
	"psql":                 "postgresql",
	"plsql":                "pl/sql",
	"jl":                   "julia",
	"rkt":                  "racket",
	"clj":                  "clojure",
	"cljs":                 "clojurescript",
	"scm":                  "scheme",
	"lisp":                 "lisp",
	"el":                   "emacs lisp",
	"erl":                  "erlang",
	"ex":                   "elixir",
	"exs":                  "elixir",
	"hs":                   "haskell",
	"nim":                  "nim",
	"crystal":              "crystal",
	"scala":                "scala",
	"sbt":                  "sbt",
	"fs":                   "f#",
	"fsi":                  "f# script",
	"ml":                   "ocaml",
	"mli":                  "ocaml interface",
	"re":                   "reason",
	"dart":                 "dart",
	"flutter":              "flutter",
	"web":                  "webassembly",
	"wat":                  "wat",
	"zig":                  "zig",
	"v":                    "v",
	"odin":                 "odin",
	"bun":                  "bun",
	"deno":                 "deno",
	"bunjs":                "bun javascript",
	"vue":                  "vue",
	"svelte":               "svelte",
	"svelte-":              "svelte",
	"astro":                "astro",
	"solid":                "solidjs",
	"qml":                  "qml",
	"rsh":                  "rescript",
	"res":                  "rescript",
	"elm":                  "elm",
	"purs":                 "purescript",
	"gleam":                "gleam",
	"mojo":                 "mojo",
	"vala":                 "vala",
	"d":                    "d",
	"pas":                  "pascal",
	"pascal":               "pascal",
	"ada":                  "ada",
	"fortran":              "fortran",
	"77":                   "fortran 77",
	"f90":                  "fortran 90",
	"asm":                  "assembly",
	"s":                    "assembly",
	"nasm":                 "nasm",
	"objdump":              "object dump",
	"bat":                  "batch",
	"cmd":                  "cmd",
	"ps1":                  "powershell",
	"psm1":                 "powershell module",
	"ahk":                  "autohotkey",
	"applescript":          "applescript",
	"scpt":                 "applescript",
	"tex":                  "latex",
	"latex":                "latex",
	"rnoweb":               "rnoweb",
	"rtex":                 "r tex",
	"xml":                  "xml",
	"xhtml":                "xhtml",
	"svg":                  "svg",
	"graphql":              "graphql",
	"gql":                  "graphql",
	"proto":                "protocol buffers",
	"thrift":               "thrift",
	"capnp":                "cap'n proto",
	"dockerfile":           "dockerfile",
	"docker":               "docker",
	"dockerignore":         "docker ignore",
	"compose":              "docker compose",
	"editorconfig":         "editorconfig",
	"gitignore":            "gitignore",
	"gitattributes":        "git attributes",
	"prettierrc":           "prettier rc",
	"eslintrc":             "eslint rc",
	"tsconfig":             "tsconfig",
	"jsconfig":             "jsconfig",
	"package.json":         "package json",
	"package-lock.json":    "package lock",
	"pnpm-lock.yaml":       "pnpm lock",
	"yarn.lock":            "yarn lock",
	"cargotoml":            "cargo toml",
	"cargo.lock":           "cargo lock",
	"composer.json":        "composer json",
	"pubspec.yaml":         "pubspec yaml",
	"build.gradle":         "gradle",
	"gradle.properties":    "gradle properties",
	"pom.xml":              "maven pom",
	"makefile":             "makefile",
	"make":                 "makefile",
	"meson.build":          "meson",
	"cmake":                "cmake",
	"ninja":                "ninja",
	"vcpkg.json":           "vcpkg",
	"justfile":             "justfile",
	"taskfile.yml":         "taskfile",
	"wolfram":              "wolfram language",
	"wl":                   "wolfram language",
	"mathematica":          "mathematica",
	"nb":                   "mathematica notebook",
	"ipynb":                "jupyter notebook",
	"rmd":                  "r markdown",
	"quarto":               "quarto",
	"org":                  "org-mode",
	"rst":                  "restructuredtext",
	"adoc":                 "asciidoc",
	"txt":                  "plain text",
	"log":                  "log file",
	"cfg":                  "config",
	"conf":                 "config",
	"ini":                  "ini",
	"env":                  ".env",
	"graphql.config.json":  "graphql config",
	"tailwind.config.js":   "tailwind config",
	"next.config.js":       "next.js config",
	"nuxt.config.ts":       "nuxt config",
	"vite.config.ts":       "vite config",
	"rollup.config.js":     "rollup config",
	"webpack.config.js":    "webpack config",
	"jest.config.js":       "jest config",
	"cypress.config.ts":    "cypress config",
	"playwright.config.ts": "playwright config",
}
