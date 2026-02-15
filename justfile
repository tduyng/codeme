bin_name := "codeme"
changelog := "CHANGELOG.md"
version_file := "VERSION"
commit := `git rev-parse --short HEAD 2>/dev/null`
# Use git describe for version: if on tag, shows tag; if after tag, shows "v0.1.0-3-gabcd1234"; if no tags, shows "dev-abcd1234"
version := `git describe --tags --match "v*" 2>/dev/null || echo "dev-$(git rev-parse --short HEAD)"`
build_time := `date -u +"%Y-%m-%dT%H:%M:%SZ"`

# Display all commands
default:
    @just --list

build:
    @go build -ldflags "-X main.version={{ version }} -X main.commit={{ commit }} -X main.buildTime={{ build_time }}" -o {{ bin_name }} .

# Build and install
install: build
    @mv {{ bin_name }} $(go env GOPATH)/bin
    @echo "✓ Installed {{ bin_name }} {{ version }} to $(go env GOPATH)/bin"

uninstall:
    @rm -rf $(go env GOPATH)/bin/codeme
    @echo "✓ Uninstalled {{ bin_name }} from $(go env GOPATH)/bin"

# Run linter (requires golangci-lint)
lint:
    @echo -e "{{ YELLOW }}Running linter...{{ NORMAL }}"
    @golangci-lint run

fmt:
    @echo -e "{{ YELLOW }}Formatting code...{{ NORMAL }}"
    @go fmt ./...

# Run tests
test:
    @go test ./...

# Coverage report
coverage:
    @mkdir -p coverage
    @go test -coverprofile=coverage/coverage.out ./...
    @go tool cover -html=coverage/coverage.out -o coverage/index.html
    @echo "Coverage: coverage/index.html"

# Clean artifacts
clean:
    @rm -rf dist/ {{ bin_name }} coverage/

tag VERSION:
    @echo "Creating tag v{{ VERSION }}..."
    @git checkout main
    @git pull origin main
    @echo "Updating {{ version_file }}..."
    @echo "{{ VERSION }}" > {{ version_file }}
    @echo "→ Generating changelog..."
    @git cliff --unreleased --tag "v{{ VERSION }}" --prepend {{ changelog }}
    @git cliff --unreleased --tag "v{{ VERSION }}" --strip all > RELEASE_NOTES.md
    @git add {{ changelog }} {{ version_file }}
    @git commit -m "chore: release v{{ VERSION }}"
    @git tag -a v{{ VERSION }} -m "Release v{{ VERSION }}"
    @echo "Push: git push && git push origin v{{ VERSION }}"
