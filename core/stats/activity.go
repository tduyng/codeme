package stats

import "time"

// CalculateMostActiveHour finds the hour with most activity
func CalculateMostActiveHour(hourly map[int]int) int {
	maxHour := 0
	maxCount := 0

	for hour, count := range hourly {
		if count > maxCount {
			maxCount = count
			maxHour = hour
		}
	}

	return maxHour
}

// CalculateMostActiveDay finds the day of week with most time
func CalculateMostActiveDay(dayTime map[time.Weekday]int64) (string, int64) {
	if len(dayTime) == 0 {
		return "", 0
	}

	dayNames := map[time.Weekday]string{
		time.Sunday:    "Sunday",
		time.Monday:    "Monday",
		time.Tuesday:   "Tuesday",
		time.Wednesday: "Wednesday",
		time.Thursday:  "Thursday",
		time.Friday:    "Friday",
		time.Saturday:  "Saturday",
	}

	var maxDay time.Weekday
	var maxTime int64
	first := true

	for day, t := range dayTime {
		if first || t > maxTime {
			maxTime = t
			maxDay = day
			first = false
		}
	}

	return dayNames[maxDay], maxTime
}

// CalculatePeakHours returns top N hours by activity
func CalculatePeakHours(hourly map[int]int, count int) []int {
	type hourCount struct {
		hour  int
		count int
	}

	hours := make([]hourCount, 0, len(hourly))
	for hour, cnt := range hourly {
		hours = append(hours, hourCount{hour: hour, count: cnt})
	}

	// Sort by count descending
	for i := 0; i < len(hours); i++ {
		for j := i + 1; j < len(hours); j++ {
			if hours[j].count > hours[i].count {
				hours[i], hours[j] = hours[j], hours[i]
			}
		}
	}

	result := make([]int, 0, count)
	for i := 0; i < count && i < len(hours); i++ {
		result = append(result, hours[i].hour)
	}

	// Sort result by hour
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j] < result[i] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}
