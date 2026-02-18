package stats

import (
	"sort"
	"time"

	"github.com/tduyng/codeme/core"
	"github.com/tduyng/codeme/util"
)

type LanguageAgg struct {
	Time  float64
	Lines int
	Files util.StringSet
}

type ProjectAgg struct {
	Time  float64
	Lines int
	Files util.StringSet
}

type EditorAgg struct {
	Time float64
}

type FileAgg struct {
	Time       float64
	Lines      int
	LastEdited time.Time
}

type DayAgg struct {
	Date         string
	Time         float64
	Lines        int
	Files        util.StringSet
	Languages    util.StringSet
	Projects     util.StringSet
	SessionCount int
}

type HourAgg struct {
	Duration float64
}

func AggregateByFile(activities []core.Activity) map[string]*FileAgg {
	agg := make(map[string]*FileAgg)

	for _, a := range activities {
		if a.File == "" {
			continue
		}

		if agg[a.File] == nil {
			agg[a.File] = &FileAgg{
				LastEdited: a.Timestamp,
			}
		}

		agg[a.File].Time += a.Duration
		agg[a.File].Lines += a.Lines
		agg[a.File].LastEdited = a.Timestamp
	}

	return agg
}

func AggregateByDay(activities []core.Activity, tz *time.Location) map[string]*DayAgg {
	agg := make(map[string]*DayAgg)

	for _, a := range activities {
		date := util.DateString(a.Timestamp, tz)

		if agg[date] == nil {
			agg[date] = &DayAgg{
				Date:      date,
				Files:     util.NewStringSet(),
				Languages: util.NewStringSet(),
				Projects:  util.NewStringSet(),
			}
		}

		agg[date].Time += a.Duration
		agg[date].Lines += a.Lines
		agg[date].Files.Add(a.File)

		if IsValidLanguage(a.Language) {
			agg[date].Languages.Add(a.Language)
		}
		if a.Project != "" {
			agg[date].Projects.Add(a.Project)
		}
	}

	return agg
}

func AggregateByHour(activities []core.Activity, tz *time.Location) [24]HourAgg {
	var agg [24]HourAgg
	if tz == nil {
		tz = time.UTC
	}

	for _, a := range activities {
		hour := a.Timestamp.In(tz).Hour()
		agg[hour].Duration += a.Duration
	}

	return agg
}

func TopFiles(fileAgg map[string]*FileAgg, total float64, n int) []APIFileStats {
	files := make([]APIFileStats, 0, len(fileAgg))

	for name, data := range fileAgg {
		pct := 0.0
		if total > 0 {
			pct = (data.Time / total) * 100
		}

		files = append(files, APIFileStats{
			Name:         name,
			Time:         data.Time,
			Lines:        data.Lines,
			PercentTotal: pct,
			LastEdited:   data.LastEdited,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Time > files[j].Time
	})

	if len(files) > n {
		files = files[:n]
	}

	return files
}
