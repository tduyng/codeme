package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/tduyng/codeme/core"
)

var (
	version   = "dev"
	buildTime = "unknown"
	commit    = "unknown"
)

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
	fmt.Println("  stats      Show statistics")
	fmt.Println("  today      Show today's activity")
	fmt.Println("  projects   Show project breakdown")
	fmt.Println("  version    Show version information")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  codeme track --file main.go --lang go --lines 10 --total 100")
	fmt.Println("  codeme stats")
	fmt.Println("  codeme stats --json")
	fmt.Println("  codeme today")
	fmt.Println()
	fmt.Println("For more information, visit: https://github.com/tduyng/codeme")
}

func handleTrack(args []string) {
	fs := flag.NewFlagSet("track", flag.ExitOnError)
	file := fs.String("file", "", "File path")
	lang := fs.String("lang", "", "Language")
	lines := fs.Int("lines", 0, "Lines changed")
	total := fs.Int("total", 0, "Total lines in file")
	fs.Parse(args)

	if *file == "" {
		fmt.Println("Error: --file required")
		os.Exit(1)
	}

	db, err := core.OpenDB()
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if *lang == "" {
		*lang = core.DetectLanguage(*file)
	}

	// Default total to lines if not provided (backward compatibility)
	if *total == 0 {
		*total = *lines
	}

	if err := core.Track(db, *file, *lang, *lines, *total); err != nil {
		fmt.Printf("Error tracking: %v\n", err)
		os.Exit(1)
	}
}

func handleStats(args []string) {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "Output as JSON")
	todayOnly := fs.Bool("today", false, "Today's stats only")
	fs.Parse(args)

	db, err := core.OpenDB()
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	stats, err := core.CalculateStats(db, *todayOnly)
	if err != nil {
		fmt.Printf("Error calculating stats: %v\n", err)
		os.Exit(1)
	}

	if *asJSON {
		json.NewEncoder(os.Stdout).Encode(stats)
	} else {
		printStats(stats)
	}
}

func handleToday() {
	db, err := core.OpenDB()
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	stats, err := core.CalculateStats(db, true)
	if err != nil {
		fmt.Printf("Error calculating stats: %v\n", err)
		os.Exit(1)
	}

	printStats(stats)
}

func handleProjects() {
	db, err := core.OpenDB()
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	stats, err := core.CalculateStats(db, false)
	if err != nil {
		fmt.Printf("Error calculating stats: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nProjects:")
	fmt.Println("─────────────────────────────────")
	for project, ps := range stats.Projects {
		fmt.Printf("  %-20s %s (%d lines)\n", project, formatTime(ps.Time), ps.Lines)
	}
}

func printStats(stats core.Stats) {
	fmt.Println("\n╭────────────────────────────────────╮")
	fmt.Println("│         CodeMe Statistics          │")
	fmt.Println("╰────────────────────────────────────╯")
	fmt.Printf("\n  Total Time: %s\n", formatTime(stats.TotalTime))
	fmt.Printf("  Total Lines: %d\n", stats.TotalLines)
	fmt.Printf("  Total Files: %d\n", stats.TotalFiles)
	fmt.Printf("  Current Streak: %d days\n", stats.Streak)
	fmt.Printf("  Longest Streak: %d days\n", stats.LongestStreak)

	fmt.Printf("\n  Today: %s (%d lines)\n", formatTime(stats.TodayTime), stats.TodayLines)

	if len(stats.Languages) > 0 {
		fmt.Println("\n  Top Languages:")
		for lang, ls := range stats.Languages {
			fmt.Printf("    %-15s %d lines (%d files, %s)\n", lang, ls.Lines, ls.Files, formatTime(ls.Time))
		}
	}

	if len(stats.Projects) > 0 {
		fmt.Println("\n  Projects:")
		for project, ps := range stats.Projects {
			fmt.Printf("    %-15s %d lines (%d files, %s)\n", project, ps.Lines, ps.Files, formatTime(ps.Time))
		}
	}

	if len(stats.TopFiles) > 0 {
		fmt.Println("\n  Most Edited Files:")
		count := 5
		if len(stats.TopFiles) < 5 {
			count = len(stats.TopFiles)
		}
		for i := 0; i < count; i++ {
			f := stats.TopFiles[i]
			fmt.Printf("    %s (%d lines)\n", f.Path, f.Lines)
		}
	}
	fmt.Println()
}

func formatTime(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%dh %dm", hours, mins)
}
