// main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/tduyng/codeme/core"
	"github.com/tduyng/codeme/stats"
)

var (
	version   = "dev"
	buildTime = "unknown"
	commit    = "unknown"
)

const LOOKBACK_DAYS = 365

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	cmd := os.Args[1]

	switch cmd {
	case "track":
		handleTrack(os.Args[2:])
	case "stats":
		handleStats(os.Args[2:])
	case "today":
		handleToday()
	case "projects":
		handleProjects()
	case "api":
		handleAPI(os.Args[2:])
	case "optimize":
		handleOptimize()
	case "rebuild-summaries":
		handleRebuildSummaries()
	case "info":
		handleInfo()
	case "version", "-v", "--version":
		printVersion()
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Printf("codeme %s\n", version)
	fmt.Printf("  Build time: %s\n", buildTime)
	fmt.Printf("  Commit: %s\n", commit)
}

func printHelp() {
	fmt.Println("codeme - Zero-config coding activity tracker")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  codeme <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  track      Track a file activity")
	fmt.Println("  stats      Show statistics (pretty printed)")
	fmt.Println("  today      Show today's activity")
	fmt.Println("  projects   Show project breakdown")
	fmt.Println("  api        Output JSON for external tools (Neovim, etc)")
	fmt.Println("  optimize   Optimize database (run monthly)")
	fmt.Println("  info       Show database information")
	fmt.Println("  version    Show version information")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  codeme track --file main.go --lang go --lines 10")
	fmt.Println("  codeme stats")
	fmt.Println("  codeme stats --today")
	fmt.Println("  codeme today")
	fmt.Println("  codeme api              # JSON output for Neovim")
	fmt.Println("  codeme api --compact    # Minified JSON")
	fmt.Println("  codeme api --days=30    # Load last 30 days only")
	fmt.Println("  codeme optimize         # Vacuum and analyze database")
	fmt.Println()
	fmt.Println("For more information, visit: https://github.com/tduyng/codeme")
}

func handleTrack(args []string) {
	fs := flag.NewFlagSet("track", flag.ExitOnError)

	file := fs.String("file", "", "File path")
	lang := fs.String("lang", "", "Language")
	editor := fs.String("editor", "", "Editor name (e.g. neovim, vscode)")
	lines := fs.Int("lines", 0, "Lines changed")

	fs.Parse(args)

	if *file == "" {
		fmt.Println("Error: --file required")
		os.Exit(1)
	}

	if *editor == "" {
		*editor = "neovim"
	}

	dbPath, err := core.GetDefaultDBPath()
	if err != nil {
		fmt.Printf("Error resolving DB path: %v\n", err)
		os.Exit(1)
	}

	storage, err := core.NewSQLiteStorage(dbPath)
	if err != nil {
		fmt.Printf("Error initializing storage: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	tracker := core.NewTracker(storage)

	if err := tracker.TrackFileActivity(*file, *lang, *editor, *lines, true); err != nil {
		fmt.Printf("Error tracking: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Activity tracked successfully")
}

func handleStats(args []string) {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	todayOnly := fs.Bool("today", false, "Show only today's stats")
	fs.Parse(args)

	dbPath, err := core.GetDefaultDBPath()
	if err != nil {
		fmt.Printf("Error resolving DB path: %v\n", err)
		os.Exit(1)
	}

	storage, err := core.OpenReadOnlyStorage(dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	calc := stats.NewCalculator(time.Local)
	apiStats, err := calc.CalculateAPI(storage, stats.APIOptions{
		LoadRecentDays: LOOKBACK_DAYS,
	})
	if err != nil {
		fmt.Printf("Error calculating stats: %v\n", err)
		os.Exit(1)
	}

	if *todayOnly {
		printTodayStats(apiStats)
	} else {
		printAllStats(apiStats)
	}
}

func handleToday() {
	dbPath, err := core.GetDefaultDBPath()
	if err != nil {
		fmt.Printf("Error resolving DB path: %v\n", err)
		os.Exit(1)
	}

	storage, err := core.OpenReadOnlyStorage(dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	calc := stats.NewCalculator(time.Local)
	apiStats, err := calc.CalculateAPI(storage, stats.APIOptions{
		LoadRecentDays: 2,
	})
	if err != nil {
		fmt.Printf("Error calculating stats: %v\n", err)
		os.Exit(1)
	}

	printTodayStats(apiStats)
}

func handleProjects() {
	dbPath, err := core.GetDefaultDBPath()
	if err != nil {
		fmt.Printf("Error resolving DB path: %v\n", err)
		os.Exit(1)
	}

	storage, err := core.OpenReadOnlyStorage(dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	calc := stats.NewCalculator(time.Local)
	apiStats, err := calc.CalculateAPI(storage, stats.APIOptions{
		LoadRecentDays: 90,
	})
	if err != nil {
		fmt.Printf("Error calculating stats: %v\n", err)
		os.Exit(1)
	}

	printProjectStats(apiStats)
}

func handleAPI(args []string) {
	fs := flag.NewFlagSet("api", flag.ExitOnError)
	compact := fs.Bool("compact", false, "Output compact JSON (no indentation)")
	days := fs.Int("days", LOOKBACK_DAYS, "Load activities from last N days (default: 365)")
	fs.Parse(args)

	dbPath, err := core.GetDefaultDBPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving DB path: %v\n", err)
		os.Exit(1)
	}

	storage, err := core.OpenReadOnlyStorage(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	calc := stats.NewCalculator(time.Local)
	apiStats, err := calc.CalculateAPI(storage, stats.APIOptions{
		LoadRecentDays: *days,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calculating stats: %v\n", err)
		os.Exit(1)
	}

	encoder := json.NewEncoder(os.Stdout)
	if !*compact {
		encoder.SetIndent("", "  ")
	}

	if err := encoder.Encode(apiStats); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}

func handleOptimize() {
	fmt.Println("🔧 Optimizing database...")

	dbPath, err := core.GetDefaultDBPath()
	if err != nil {
		fmt.Printf("Error resolving DB path: %v\n", err)
		os.Exit(1)
	}

	storage, err := core.NewSQLiteStorage(dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	startTime := time.Now()

	if err := storage.Optimize(); err != nil {
		fmt.Printf("❌ Error optimizing: %v\n", err)
		os.Exit(1)
	}

	duration := time.Since(startTime)
	fmt.Printf("✓ Database optimized in %.2fs\n", duration.Seconds())
	fmt.Println("  • Rebuilt indexes")
	fmt.Println("  • Reclaimed space")
	fmt.Println("  • Analyzed query patterns")
}

func handleRebuildSummaries() {
	fmt.Println("📊 Rebuilding summary tables...")

	dbPath, err := core.GetDefaultDBPath()
	if err != nil {
		fmt.Printf("Error resolving DB path: %v\n", err)
		os.Exit(1)
	}

	storage, err := core.NewSQLiteStorage(dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	startTime := time.Now()

	if err := storage.RebuildSummaries(); err != nil {
		fmt.Printf("❌ Error rebuilding summaries: %v\n", err)
		os.Exit(1)
	}

	duration := time.Since(startTime)
	fmt.Printf("✓ Summary tables rebuilt in %.2fs\n", duration.Seconds())
	fmt.Println("  • daily_summary")
	fmt.Println("  • daily_language_summary")
	fmt.Println("  • daily_project_summary")
}

func handleInfo() {
	dbPath, err := core.GetDefaultDBPath()
	if err != nil {
		fmt.Printf("Error resolving DB path: %v\n", err)
		os.Exit(1)
	}

	storage, err := core.NewSQLiteStorage(dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	count, err := storage.GetActivityCount()
	if err != nil {
		fmt.Printf("Error getting activity count: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(dbPath)
	var dbSize int64 = 0
	if err == nil {
		dbSize = info.Size()
	}

	fmt.Println("\n╭────────────────────────────────────╮")
	fmt.Println("│       Database Information         │")
	fmt.Println("╰────────────────────────────────────╯")
	fmt.Printf("\n  📍 Location: %s\n", dbPath)
	fmt.Printf("  📊 Total Activities: %d\n", count)
	fmt.Printf("  💾 Database Size: %.2f MB\n", float64(dbSize)/(1024*1024))

	if count > 0 {
		avgSize := float64(dbSize) / float64(count)
		fmt.Printf("  📏 Avg per Activity: %.2f KB\n", avgSize/1024)
		estimatedDays := count / 200
		fmt.Printf("  📅 Estimated Days: ~%d days\n", estimatedDays)
	}

	fmt.Println("\n  💡 Tip: Run 'codeme optimize' monthly to maintain performance")
	fmt.Println()
}

func printTodayStats(s *stats.APIStats) {
	today := s.Today

	fmt.Println("\n╭────────────────────────────────────╮")
	fmt.Println("│         Today's Activity           │")
	fmt.Println("╰────────────────────────────────────╯")

	fmt.Printf("\n  ⏱  Time: %s\n", formatDuration(today.TotalTime))
	fmt.Printf("  📝 Lines: %d\n", today.TotalLines)

	if len(today.Sessions) > 0 {
		fmt.Printf("  🎯 Sessions: %d (Focus: %d%%)\n", len(today.Sessions), today.FocusScore)
	}

	if len(today.Languages) > 0 {
		fmt.Println("\n  Languages:")
		for i, lang := range today.Languages {
			if i >= 5 {
				break
			}
			fmt.Printf("    %-15s %s (%.1f%%)\n", lang.Name, formatDuration(lang.Time), lang.PercentTotal)
		}
	}

	if len(today.Projects) > 0 {
		fmt.Println("\n  Projects:")
		for i, proj := range today.Projects {
			if i >= 5 {
				break
			}
			fmt.Printf("    %-15s %s\n", proj.Name, formatDuration(proj.Time))
		}
	}

	if today.DailyGoals.TimeGoal > 0 {
		fmt.Println("\n  Daily Goals:")
		fmt.Printf("    Time:  %.1f%% of %s\n",
			today.DailyGoals.TimeProgress,
			formatDuration(today.DailyGoals.TimeGoal))
		fmt.Printf("    Lines: %.1f%% of %d\n",
			today.DailyGoals.LinesProgress,
			today.DailyGoals.LinesGoal)
		if today.DailyGoals.OnTrack {
			fmt.Println("    ✓ On track!")
		}
	}

	if s.Meta.QueryTimeMs > 0 {
		fmt.Printf("\n  ⚡ Query: %.0fms (%d activities)\n", s.Meta.QueryTimeMs, s.Meta.LoadedActivities)
	}

	fmt.Println()
}

func printAllStats(s *stats.APIStats) {
	fmt.Println("\n╭────────────────────────────────────╮")
	fmt.Println("│       CodeMe Statistics            │")
	fmt.Println("╰────────────────────────────────────╯")

	fmt.Printf("\n  📊 Overview\n")
	fmt.Printf("  ─────────────────────────────────\n")
	fmt.Printf("  Today:      %s (%d lines)\n", formatDuration(s.Today.TotalTime), s.Today.TotalLines)
	fmt.Printf("  This Week:  %s (%d lines)\n", formatDuration(s.ThisWeek.TotalTime), s.ThisWeek.TotalLines)
	fmt.Printf("  All Time:   %s (%d lines)\n", formatDuration(s.AllTime.TotalTime), s.AllTime.TotalLines)

	if s.StreakInfo.Current > 0 {
		fmt.Printf("\n  🔥 Streak\n")
		fmt.Printf("  ─────────────────────────────────\n")
		fmt.Printf("  Current: %d days\n", s.StreakInfo.Current)
		fmt.Printf("  Longest: %d days\n", s.StreakInfo.Longest)
		if s.StreakInfo.IsActive {
			fmt.Printf("  ✓ Active today!\n")
		}
	}

	if len(s.AllTime.Languages) > 0 {
		fmt.Printf("\n  💬 Top Languages (All Time)\n")
		fmt.Printf("  ─────────────────────────────────\n")
		for i, lang := range s.AllTime.Languages {
			if i >= 5 {
				break
			}
			profBadge := ""
			if lang.Proficiency != "" {
				profBadge = fmt.Sprintf(" [%s]", lang.Proficiency)
			}
			fmt.Printf("  %-15s %s%s\n", lang.Name, formatDuration(lang.Time), profBadge)
		}
	}

	if len(s.AllTime.Projects) > 0 {
		fmt.Printf("\n  📁 Top Projects (All Time)\n")
		fmt.Printf("  ─────────────────────────────────\n")
		for i, proj := range s.AllTime.Projects {
			if i >= 5 {
				break
			}
			fmt.Printf("  %-20s %s (%s)\n", proj.Name, formatDuration(proj.Time), proj.MainLanguage)
		}
	}

	unlockedCount := 0
	for _, ach := range s.Achievements {
		if ach.Unlocked {
			unlockedCount++
		}
	}

	if unlockedCount > 0 {
		fmt.Printf("\n  🏆 Achievements\n")
		fmt.Printf("  ─────────────────────────────────\n")
		fmt.Printf("  Unlocked: %d/%d\n", unlockedCount, len(s.Achievements))

		shown := 0
		for _, ach := range s.Achievements {
			if ach.Unlocked && shown < 3 {
				fmt.Printf("  %s %s\n", ach.Icon, ach.Name)
				shown++
			}
		}
	}

	if s.Meta.QueryTimeMs > 0 {
		fmt.Printf("\n  ⚡ Performance\n")
		fmt.Printf("  ─────────────────────────────────\n")
		fmt.Printf("  Query Time: %.0fms\n", s.Meta.QueryTimeMs)
		fmt.Printf("  Loaded: %d/%d activities\n", s.Meta.LoadedActivities, s.Meta.TotalActivities)
		fmt.Printf("  Window: %s\n", s.Meta.DataWindow)
	}

	fmt.Println()
}

func printProjectStats(s *stats.APIStats) {
	fmt.Println("\n╭────────────────────────────────────╮")
	fmt.Println("│           Projects                 │")
	fmt.Println("╰────────────────────────────────────╯")

	if len(s.AllTime.Projects) == 0 {
		fmt.Println("\n  No projects tracked yet")
		fmt.Println()
		return
	}

	for i, proj := range s.AllTime.Projects {
		if i >= 10 {
			break
		}
		fmt.Printf("\n  %d. %s\n", i+1, proj.Name)
		fmt.Printf("     Time:     %s\n", formatDuration(proj.Time))
		fmt.Printf("     Lines:    %d\n", proj.Lines)
		fmt.Printf("     Files:    %d\n", proj.Files)
		fmt.Printf("     Language: %s\n", proj.MainLanguage)
	}
	fmt.Println()
}

func formatDuration(seconds float64) string {
	if seconds < 60 {
		return fmt.Sprintf("%.0fs", seconds)
	}
	minutes := int(seconds) / 60
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	mins := minutes % 60
	if mins > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dh", hours)
}
