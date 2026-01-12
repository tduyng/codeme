bin_name := "codeme"
changelog := "CHANGELOG.md"
version_file := "VERSION"
commit := `git rev-parse --short HEAD 2>/dev/null`
version := `cat VERSION 2>/dev/null || echo "0.0.1"`
build_time := `date -u +"%Y-%m-%dT%H:%M:%SZ"`

# Display all commands
default:
    @just --list

# Build and install
install:
    @go build -ldflags "-X main.version={{ version }} -X main.commit={{ commit }} -X main.buildTime={{ build_time }}" -o {{ bin_name }} .
    @mv {{ bin_name }} $(go env GOPATH)/bin
    @echo "✓ Installed {{ bin_name }} {{ version }} to $(go env GOPATH)/bin"

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

changelog:
    @echo "→ Generating changelog with git‑cliff…"
    @git cliff --latest --prepend {{ changelog }}
    @git cliff --latest --strip all --output LATEST_CHANGELOG.md
    @echo "Changelog written to {{ changelog }}"

tag VERSION:
    @echo "Creating tag v{{ VERSION }}..."
    @git checkout main
    @git pull origin main
    @echo "Updating {{ version_file }}..."
    @echo "{{ VERSION }}" > {{ version_file }}
    @echo "→ Generating changelog for v{{ VERSION }}…"
    @git cliff --unreleased --tag "v{{ VERSION }}" --prepend {{ changelog }}
    @git cliff --unreleased --tag "v{{ VERSION }}" --strip all --output LATEST_CHANGELOG.md
    @git add {{ changelog }} LATEST_CHANGELOG.md {{ version_file }}
    @git commit -m "chore: release v{{ VERSION }}"
    @git tag -a v{{ VERSION }} -m "Release v{{ VERSION }}"
    @echo "Tag v{{ VERSION }} created. Push with: git push && git push origin v{{ VERSION }}"
