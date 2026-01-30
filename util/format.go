// util/format.go
package util

import "fmt"

func FormatDuration(seconds float64) string {
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

func FormatBytes(bytes int64) string {
	const unit = 1024

	sign := ""
	if bytes < 0 {
		sign = "-"
		bytes = -bytes
	}

	if bytes < unit {
		return fmt.Sprintf("%s%d B", sign, bytes)
	}

	div := float64(unit)
	exp := 0

	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}

	for bytes >= int64(div)*unit && exp < len(units)-1 {
		div *= unit
		exp++
	}

	value := float64(bytes) / div
	return fmt.Sprintf("%s%.1f %s", sign, value, units[exp])
}
