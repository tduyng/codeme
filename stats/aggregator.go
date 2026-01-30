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

func AggregateByLanguage(activities []core.Activity) map[string]*LanguageAgg {
	agg := make(map[string]*LanguageAgg)

	for _, a := range activities {
		if !IsValidLanguage(a.Language) {
			continue
		}

		lang := NormalizeLanguage(a.Language)
		if agg[lang] == nil {
			agg[lang] = &LanguageAgg{
				Files: util.NewStringSet(),
			}
		}

		agg[lang].Time += a.Duration
		agg[lang].Lines += a.Lines
		agg[lang].Files.Add(a.File)
	}

	return agg
}

func AggregateByProject(activities []core.Activity) map[string]*ProjectAgg {
	agg := make(map[string]*ProjectAgg)

	for _, a := range activities {
		if agg[a.Project] == nil {
			agg[a.Project] = &ProjectAgg{
				Files: util.NewStringSet(),
			}
		}

		agg[a.Project].Time += a.Duration
		agg[a.Project].Lines += a.Lines
		agg[a.Project].Files.Add(a.File)
	}

	return agg
}

func AggregateByEditor(activities []core.Activity) map[string]*EditorAgg {
	agg := make(map[string]*EditorAgg)

	for _, a := range activities {
		if agg[a.Editor] == nil {
			agg[a.Editor] = &EditorAgg{}
		}
		agg[a.Editor].Time += a.Duration
	}

	return agg
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

func TopLanguages(langAgg map[string]*LanguageAgg, lifetimeHours map[string]float64, total float64, n int) []APILanguageStats {
	langs := make([]APILanguageStats, 0, len(langAgg))

	for name, data := range langAgg {
		hours := lifetimeHours[name]
		proficiency := CalculateProficiency(hours)

		pct := 0.0
		if total > 0 {
			pct = (data.Time / total) * 100
		}

		langs = append(langs, APILanguageStats{
			Name:         name,
			Time:         data.Time,
			Lines:        data.Lines,
			Files:        data.Files.Len(),
			PercentTotal: pct,
			Proficiency:  proficiency,
			HoursTotal:   hours,
		})
	}

	sort.Slice(langs, func(i, j int) bool {
		return langs[i].Time > langs[j].Time
	})

	if len(langs) > n {
		langs = langs[:n]
	}

	return langs
}

func TopProjects(projAgg map[string]*ProjectAgg, projectLangs map[string]map[string]float64, total float64, n int) []APIProjectStats {
	projs := make([]APIProjectStats, 0, len(projAgg))

	for name, data := range projAgg {
		mainLang := "Mixed"
		maxTime := 0.0
		for lang, time := range projectLangs[name] {
			if time > maxTime {
				maxTime = time
				mainLang = lang
			}
		}

		pct := 0.0
		if total > 0 {
			pct = (data.Time / total) * 100
		}

		projs = append(projs, APIProjectStats{
			Name:         name,
			Time:         data.Time,
			Lines:        data.Lines,
			Files:        data.Files.Len(),
			PercentTotal: pct,
			MainLanguage: mainLang,
		})
	}

	sort.Slice(projs, func(i, j int) bool {
		return projs[i].Time > projs[j].Time
	})

	if len(projs) > n {
		projs = projs[:n]
	}

	return projs
}

func TopEditors(editorAgg map[string]*EditorAgg, total float64, n int) []APIEditorStats {
	editors := make([]APIEditorStats, 0, len(editorAgg))

	for name, data := range editorAgg {
		pct := 0.0
		if total > 0 {
			pct = (data.Time / total) * 100
		}

		editors = append(editors, APIEditorStats{
			Name:         name,
			Time:         data.Time,
			PercentTotal: pct,
		})
	}

	sort.Slice(editors, func(i, j int) bool {
		return editors[i].Time > editors[j].Time
	})

	if len(editors) > n {
		editors = editors[:n]
	}

	return editors
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
