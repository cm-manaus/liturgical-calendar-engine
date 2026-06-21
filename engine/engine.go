package engine

import (
	"path/filepath"
	"time"
)

type LiturgicalEngine struct {
	temporalCycle       *TemporalCycle
	sanctorale          *Sanctorale
	brazilianSanctorale *BrazilianSanctorale
}

func NewLiturgicalEngine(dataDir string) *LiturgicalEngine {
	temporalPath := filepath.Join(dataDir, "temporal_cycle.xml")
	universalPath := filepath.Join(dataDir, "universal_sanctoral.xml")
	brazilianPath := filepath.Join(dataDir, "brazilian_sanctoral.xml")

	return &LiturgicalEngine{
		temporalCycle:       NewTemporalCycle(temporalPath),
		sanctorale:          NewSanctorale(universalPath),
		brazilianSanctorale: NewBrazilianSanctorale(brazilianPath),
	}
}

func (le *LiturgicalEngine) Resolve(date time.Time, includeBrazilian bool) LiturgicalResult {
	temporal := le.temporalCycle.GetDay(date)
	universalSanctoral := le.sanctorale.GetDay(date)
	var brazilianSanctoral *LiturgicalDay
	if includeBrazilian {
		brazilianSanctoral = le.brazilianSanctorale.GetDay(date)
	}

	var sanctoral *LiturgicalDay
	if brazilianSanctoral != nil {
		if universalSanctoral == nil || brazilianSanctoral.LiturgicalClass.Precedes(universalSanctoral.LiturgicalClass) {
			sanctoral = brazilianSanctoral
		} else {
			sanctoral = universalSanctoral
		}
	} else {
		sanctoral = universalSanctoral
	}

	// Our Lady on Saturday rule
	if date.Weekday() == time.Saturday && (temporal == nil || temporal.NameResID == "feria") && sanctoral == nil {
		seasonColor := le.GetSeasonColor(date)
		if seasonColor == ColorGreen || seasonColor == ColorWhite {
			return LiturgicalResult{
				MainDay: LiturgicalDay{
					Name:            "Our Lady on Saturday",
					NameResID:       "our_lady_saturday",
					LiturgicalClass: ClassIV,
					Color:           ColorWhite,
				},
			}
		}
	}

	if temporal == nil && sanctoral == nil {
		seasonColor := le.GetSeasonColor(date)
		return LiturgicalResult{
			MainDay: LiturgicalDay{
				Name:            "Feria",
				NameResID:       "feria",
				LiturgicalClass: ClassIV,
				Color:           seasonColor,
			},
		}
	}

	if temporal != nil && sanctoral == nil {
		return LiturgicalResult{MainDay: *temporal}
	}

	if temporal == nil && sanctoral != nil {
		return LiturgicalResult{MainDay: *sanctoral}
	}

	// Both temporal and sanctoral are present
	t := *temporal
	s := *sanctoral

	// Rule 1: Higher class wins
	if t.LiturgicalClass.Precedes(s.LiturgicalClass) {
		return resultWithFilteredComms(t, []LiturgicalDay{s})
	} else if s.LiturgicalClass.Precedes(t.LiturgicalClass) {
		return resultWithFilteredComms(s, []LiturgicalDay{t})
	}

	// Rule 2: Same class
	if date.Weekday() == time.Sunday {
		if s.IsLordFeast {
			return resultWithFilteredComms(s, []LiturgicalDay{t})
		}
		if t.LiturgicalClass == ClassI {
			return resultWithFilteredComms(t, []LiturgicalDay{s})
		}
		return resultWithFilteredComms(t, []LiturgicalDay{s})
	}

	return resultWithFilteredComms(t, []LiturgicalDay{s})
}

func resultWithFilteredComms(main LiturgicalDay, comms []LiturgicalDay) LiturgicalResult {
	var filtered []LiturgicalDay
	seen := make(map[string]bool)
	for _, c := range comms {
		if c.LiturgicalClass == ClassIV {
			continue
		}
		if c.ObservanceKey() == main.ObservanceKey() {
			continue
		}
		key := c.ObservanceKey()
		if !seen[key] {
			seen[key] = true
			filtered = append(filtered, c)
		}
	}
	if filtered == nil {
		filtered = []LiturgicalDay{}
	}
	return LiturgicalResult{
		MainDay:        main,
		Commemorations: filtered,
	}
}

func (le *LiturgicalEngine) GetSeasonColor(date time.Time) LiturgicalColor {
	year := date.Year()
	easter := CalculateEaster(year)

	y1, m1, d1 := date.Date()
	y2, m2, d2 := easter.Date()
	t1 := time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC)
	diffFromEaster := int(t1.Sub(t2).Hours() / 24)

	if diffFromEaster >= 0 && diffFromEaster <= 55 {
		return ColorWhite
	}
	if diffFromEaster >= -70 && diffFromEaster <= -1 {
		return ColorViolet
	}

	christmas := time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC)
	christmasDow := int(christmas.Weekday())
	daysBefore := christmasDow
	if christmasDow == 0 {
		daysBefore = 7
	}
	sundayBeforeChristmas := christmas.AddDate(0, 0, -daysBefore)
	advent1 := sundayBeforeChristmas.AddDate(0, 0, -21)

	dateUTC := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	if (dateUTC.After(advent1) || dateUTC.Equal(advent1)) && dateUTC.Before(christmas) {
		return ColorViolet
	}

	if dateUTC.Equal(christmas) || dateUTC.After(christmas) {
		return ColorWhite
	}
	jan13 := time.Date(year, 1, 13, 0, 0, 0, 0, time.UTC)
	if dateUTC.Before(jan13) || dateUTC.Equal(jan13) {
		return ColorWhite
	}

	return ColorGreen
}
