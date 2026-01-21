# CodeMe

> Zero-config coding activity tracker written in Go

Track your coding activity with automatic detection of projects, languages, and branches. Designed to work seamlessly with editor integrations.

## Installation

```bash
# Using go install
go install github.com/tduyng/codeme@latest

# Or build from source
git clone https://github.com/tduyng/codeme
cd codeme
just install
```

## Usage

### Quick Start

```bash
# Show help
codeme help

# Show version
codeme version

# View your stats
codeme stats

# Today's activity only
codeme today

# List projects
codeme projects

# JSON output (for integrations)
codeme stats --json
```

### Automatic Tracking (Recommended)

Use the [codeme.nvim](https://github.com/tduyng/codeme.nvim) plugin for automatic tracking in Neovim. It tracks when you:

- Open files
- Save files
- Switch back to Neovim

See "Integration with Editors" section below for setup.

### Track Activity Manually (Advanced)

The track command is mainly used by editor integrations. You rarely need to call it manually.

```bash
# Track a file (called by codeme.nvim automatically)
codeme track --file /path/to/file.go --lines 100

# Specify language manually
codeme track --file /path/to/file.txt --lang plaintext --lines 50
```

## Features

- Zero configuration - works out of the box
- Auto-detection - project, language, and git branch
- Session tracking - 15-minute idle timeout
- Streak calculation - maintain your coding momentum
- SQLite storage - all data stored locally
- JSON API - easy integration with editors and tools
- Neovim plugin - beautiful dashboard with [codeme.nvim](https://github.com/tduyng/codeme.nvim)

## Data Storage

All data is stored locally and persists across versions:

```
~/.local/share/codeme/codeme.db
```

## Architecture

```
codeme/
├── main.go           # CLI entry point
├── core/
│   ├── storage.go    # SQLite operations
│   ├── tracker.go    # Activity tracking
│   ├── detector.go   # Auto-detection logic
│   └── stats.go      # Statistics calculation
└── go.mod
```

## Development

### Build and Install

```bash
# Install development version
just install

# Run tests
just test

# Run locally
go run . stats
```

### Debug from database
```bash
# Check if activities exist
sqlite3 ~/.local/share/codeme/codeme.db "SELECT COUNT(*) FROM activities;"

# If > 0, check API output
codeme api | jq '.this_week.total_time'

# If 0, use debug version
# Replace bridge.go with debug_bridge artifact
go build && ./codeme api 2>&1 | grep DEBUG
```

### Integration with Editors

#### Neovim

Install [codeme.nvim](https://github.com/tduyng/codeme.nvim) for a beautiful dashboard:

```lua
{
  "tduyng/codeme.nvim",
  dependencies = { "nvzone/volt" },
  config = function()
    require("codeme").setup()
  end,
}
```

The plugin automatically tracks files and provides:

- 3-tab dashboard - Overview, Languages, Activity
- GitHub-style activity heatmap - 7 months of coding history
- Language breakdown - with visual bar graphs
- Streak tracking - maintain momentum
- Auto-tracking - on file open, save, and focus

Commands:

- `:CodeMe` - Open dashboard
- `:CodeMeTrack` - Manually track current file
- `:CodeMeToday` - Show today's stats
- `:CodeMeProjects` - Show project breakdown

#### Other Editors

The CLI is designed to be editor-agnostic. Integrate by calling:

```bash
# On file save
codeme track --file "$FILE_PATH" --lang "$LANGUAGE" --lines "$LINE_COUNT"

# Get stats (JSON format)
codeme stats --json
```

## License

MIT
