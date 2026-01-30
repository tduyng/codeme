# CodeMe

> Private coding activity tracker that just works

Track your coding sessions automatically. No account, no cloud, no config, just insights.

## Install

### Download prebuilt binary (recommended)

No dependencies needed. Download the latest release for your platform:

→ [GitHub Releases](https://github.com/tduyng/codeme/releases/latest)

| Platform       | File                         |
| -------------- | ---------------------------- |
| macOS (ARM)    | `codeme_darwin_arm64.tar.gz` |
| macOS (Intel)  | `codeme_darwin_amd64.tar.gz` |
| Linux (x86_64) | `codeme_linux_amd64.tar.gz`  |
| Linux (ARM64)  | `codeme_linux_arm64.tar.gz`  |

### Go install

Requires [Go 1.25+](https://go.dev/dl/) and a C compiler (`gcc` or `clang`):

```bash
go install github.com/tduyng/codeme@latest
```

### Build from source

Same requirements as above: Go 1.25+ and a C compiler. The C compiler is needed
because the SQLite library ([mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)) uses cgo.
The resulting binary bundles SQLite, no runtime dependency.

```bash
git clone https://github.com/tduyng/codeme
cd codeme && just install # must have justfile installed
```

## Quick Start

```bash
codeme stats      # View your coding stats
codeme today      # Today's activity
codeme projects   # Project breakdown
codeme api        # JSON output for integrations
```

## Neovim Plugin

Install [codeme.nvim](https://github.com/tduyng/codeme.nvim) for automatic tracking and a beautiful dashboard:

```lua
{ "tduyng/codeme.nvim" }
```

The plugin tracks automatically when you:

- Open and save files
- Switch back to Neovim
- Work across different projects

No manual tracking needed.

This tracks:

- Sessions - 15min idle timeout groups your work
- Projects - Auto-detected from git repos
- Languages - Detected from file extensions
- Streaks - Keep your momentum going
- Branches - Know what you worked on

## Your Data

Everything stays on your machine:

```
~/.local/share/codeme/codeme.db
```

No telemetry. No accounts. No cloud sync. Just you and your code.

## Manual Tracking

Mostly used by editor integrations, but you can track manually:

```bash
codeme track --file main.go --lines 50
codeme track --file script.py --lang python --lines 100
```

## Development

```bash
just test       # Run tests
just install    # Build and install
go run . stats  # Run locally
```

## License

MIT
