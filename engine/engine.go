package engine

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type LiturgicalEngine struct {
	temporalCycle       *TemporalCycle
	sanctorale          *Sanctorale
	brazilianSanctorale *BrazilianSanctorale
	profile1954         *DivinoAfflatu1954Profile
}

func NewLiturgicalEngine(dataDir string) *LiturgicalEngine {
	temporalPath := filepath.Join(dataDir, "temporal.xml")
	if _, err := os.Stat(temporalPath); os.IsNotExist(err) {
		temporalPath = filepath.Join(dataDir, "temporal_cycle.xml")
	}

	universalPath := filepath.Join(dataDir, "sanctoral.xml")
	if _, err := os.Stat(universalPath); os.IsNotExist(err) {
		universalPath = filepath.Join(dataDir, "universal_sanctoral.xml")
	}

	brazilianPath := filepath.Join(dataDir, "brazilian_sanctoral.xml")

	brazilian := NewBrazilianSanctorale(brazilianPath)
	profile1954Dir := filepath.Join(dataDir, "1954")

	return &LiturgicalEngine{
		temporalCycle:       NewTemporalCycle(temporalPath),
		sanctorale:          NewSanctorale(universalPath),
		brazilianSanctorale: brazilian,
		profile1954:         NewDivinoAfflatu1954Profile(profile1954Dir, brazilian),
	}
}

var annunciation1962 = LiturgicalDay{
	ID:              "annunciation",
	Name:            "Annunciation of the B.V.M.",
	NameResID:       "annunciation",
	LiturgicalClass: ClassI,
	Color:           ColorWhite,
	IsLordFeast:     true,
	CalendarVersion: Calendar1962,
}

func (le *LiturgicalEngine) Resolve(date time.Time, version CalendarVersion, includeBrazilian bool) LiturgicalResult {
	if version == Calendar1954 {
		return le.profile1954.Resolve(date, includeBrazilian)
	}
	return le.resolve1962(date, includeBrazilian)
}

func (le *LiturgicalEngine) resolve1962(date time.Time, includeBrazilian bool) LiturgicalResult {
	dateUTC := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	temporal := le.temporalCycle.GetDay(dateUTC)
	if temporal != nil && temporal.NameResID == "feria" {
		if ep, gosp, found := getFerialReadings1962(dateUTC); found {
			temporal.Epistle = ep
			temporal.Gospel = gosp
		}
	}

	universalFeasts := le.sanctorale.GetAllFeasts(dateUTC)
	var brazilianFeasts []LiturgicalDay
	if includeBrazilian && le.brazilianSanctorale != nil {
		brazilianFeasts = le.brazilianSanctorale.GetAllFeasts(dateUTC)
	}

	// Annunciation transfer (1960 Code of Rubrics §96a)
	easterDate := CalculateEaster(dateUTC.Year())
	easterOctaveEnd := easterDate.AddDate(0, 0, 7)
	transferDate := easterOctaveEnd.AddDate(0, 0, 1)

	isHolyWeek := dateUTC.After(easterDate.AddDate(0, 0, -8)) && dateUTC.Before(easterDate)
	isEasterOctave := !dateUTC.Before(easterDate) && !dateUTC.After(easterOctaveEnd)

	var effectiveUniversal []LiturgicalDay
	for _, f := range universalFeasts {
		if dateUTC.Month() == 3 && dateUTC.Day() == 25 && (isHolyWeek || isEasterOctave) {
			if f.NameResID == "annunciation" {
				continue
			}
		}
		effectiveUniversal = append(effectiveUniversal, f)
	}

	if dateUTC.Equal(transferDate) {
		hasAnnunciation := false
		for _, f := range effectiveUniversal {
			if f.NameResID == "annunciation" {
				hasAnnunciation = true
				break
			}
		}
		if !hasAnnunciation {
			march25 := time.Date(dateUTC.Year(), 3, 25, 0, 0, 0, 0, time.UTC)
			march25InHolyWeek := march25.After(easterDate.AddDate(0, 0, -8)) && march25.Before(easterDate)
			march25InEasterOctave := !march25.Before(easterDate) && !march25.After(easterOctaveEnd)
			if march25InHolyWeek || march25InEasterOctave {
				effectiveUniversal = append(effectiveUniversal, annunciation1962)
			}
		}
	}

	mergedSanctoral := le.mergeSanctoralFeasts1962(effectiveUniversal, brazilianFeasts)
	var sanctoral *LiturgicalDay
	var sanctoralComms []LiturgicalDay
	if len(mergedSanctoral) > 0 {
		sanctoral = &mergedSanctoral[0]
		if len(mergedSanctoral) > 1 {
			sanctoralComms = mergedSanctoral[1:]
		}
	}

	// Helper for filtered commemorations per 1960 Code §108
	resultWithFilteredComms := func(main LiturgicalDay, comms []LiturgicalDay) LiturgicalResult {
		nameResID := main.NameResID
		isPenitentialDay := nameResID == "ash_wednesday" ||
			nameResID == "lent_sunday" ||
			nameResID == "lent_sunday_laetare" ||
			nameResID == "passion_sunday_1" ||
			nameResID == "palm_sunday_passion" ||
			nameResID == "monday_holy_week" ||
			nameResID == "tuesday_holy_week" ||
			nameResID == "wednesday_holy_week" ||
			nameResID == "holy_thursday" ||
			nameResID == "holy_saturday" ||
			nameResID == "easter_sunday" ||
			strings.HasPrefix(nameResID, "easter_") ||
			nameResID == "low_sunday" ||
			nameResID == "sunday_after_easter" ||
			nameResID == "sunday_after_easter_good_shepherd" ||
			nameResID == "sunday_after_ascension" ||
			nameResID == "finding_cross" ||
			nameResID == "vigil_ascension" ||
			nameResID == "pentecost_monday" ||
			nameResID == "pentecost_tuesday" ||
			nameResID == "pentecost_thursday" ||
			strings.HasPrefix(nameResID, "ember_")

		var minCommClass *LiturgicalClass
		if isPenitentialDay {
			if main.LiturgicalClass == ClassI {
				lowerName := strings.ToLower(main.Name)
				isSunday := strings.Contains(lowerName, "domingo") || strings.Contains(lowerName, "sunday")
				if isSunday {
					c := ClassII
					minCommClass = &c
				} else {
					c := ClassI
					minCommClass = &c
				}
			} else if main.LiturgicalClass == ClassII {
				c := ClassII
				minCommClass = &c
			}
		}

		if minCommClass == nil && (main.LiturgicalClass == ClassI || main.LiturgicalClass == ClassII) {
			return LiturgicalResult{
				MainDay:         main,
				Commemorations:  []LiturgicalDay{},
				CalendarVersion: Calendar1962,
			}
		}

		var filtered []LiturgicalDay
		seen := make(map[string]bool)
		for _, c := range comms {
			if c.ObservanceKey() == main.ObservanceKey() {
				continue
			}
			if minCommClass != nil && c.LiturgicalClass > *minCommClass {
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
			MainDay:         main,
			Commemorations:  filtered,
			CalendarVersion: Calendar1962,
		}
	}

	// 1960 Code §78: Our Lady on Saturday
	if dateUTC.Weekday() == time.Saturday &&
		(temporal == nil || temporal.NameResID == "feria") &&
		(sanctoral == nil || sanctoral.LiturgicalClass == ClassIV) {
		seasonColor := le.GetSeasonColor(dateUTC)
		if seasonColor == ColorGreen || seasonColor == ColorWhite {
			var comms []LiturgicalDay
			if sanctoral != nil {
				comms = append(comms, *sanctoral)
			}
			ourLady := LiturgicalDay{
				ID:              "our_lady_saturday",
				Name:            "Our Lady on Saturday",
				NameResID:       "our_lady_saturday",
				LiturgicalClass: ClassIV,
				Color:           ColorWhite,
				Epistle:         "Eclo 24,14-16",
				Gospel:          "Lc 11,27-28",
				CalendarVersion: Calendar1962,
			}
			return resultWithFilteredComms(ourLady, comms)
		}
	}

	if temporal == nil && sanctoral == nil {
		seasonColor := le.GetSeasonColor(dateUTC)
		ep, gosp, _ := getFerialReadings1962(dateUTC)
		return LiturgicalResult{
			MainDay: LiturgicalDay{
				ID:              "feria",
				Name:            "Feria",
				NameResID:       "feria",
				LiturgicalClass: ClassIV,
				Color:           seasonColor,
				Epistle:         ep,
				Gospel:          gosp,
				CalendarVersion: Calendar1962,
			},
			Commemorations:  []LiturgicalDay{},
			CalendarVersion: Calendar1962,
		}
	}

	if temporal != nil && sanctoral == nil {
		return LiturgicalResult{
			MainDay:         *temporal,
			Commemorations:  []LiturgicalDay{},
			CalendarVersion: Calendar1962,
		}
	}

	if temporal == nil && sanctoral != nil {
		seasonColor := le.GetSeasonColor(dateUTC)
		ep, gosp, _ := getFerialReadings1962(dateUTC)
		feria := LiturgicalDay{
			ID:              "feria",
			Name:            "Feria",
			NameResID:       "feria",
			LiturgicalClass: ClassIV,
			Color:           seasonColor,
			Epistle:         ep,
			Gospel:          gosp,
			CalendarVersion: Calendar1962,
		}
		if sanctoral.LiturgicalClass == ClassIV {
			var allComms []LiturgicalDay
			allComms = append(allComms, *sanctoral)
			allComms = append(allComms, sanctoralComms...)
			return resultWithFilteredComms(feria, allComms)
		}
		return resultWithFilteredComms(*sanctoral, sanctoralComms)
	}

	t := *temporal
	s := *sanctoral
	isChristTheKing := s.NameResID == "christ_the_king"

	var comms []LiturgicalDay
	comms = append(comms, s)
	comms = append(comms, sanctoralComms...)

	var commsIfSWins []LiturgicalDay
	if !isChristTheKing {
		commsIfSWins = append(commsIfSWins, t)
	}
	commsIfSWins = append(commsIfSWins, sanctoralComms...)

	if s.LiturgicalClass == ClassIV {
		return resultWithFilteredComms(t, comms)
	}

	if t.LiturgicalClass.Precedes(s.LiturgicalClass) {
		return resultWithFilteredComms(t, comms)
	} else if s.LiturgicalClass.Precedes(t.LiturgicalClass) {
		return resultWithFilteredComms(s, commsIfSWins)
	}

	// Same class precedence
	if t.IsLordFeast {
		return resultWithFilteredComms(t, comms)
	}
	if s.IsLordFeast {
		return resultWithFilteredComms(s, commsIfSWins)
	}
	if s.NameResID == "immaculate_conception" {
		return resultWithFilteredComms(s, commsIfSWins)
	}

	if t.NameResID == "feria" && s.NameResID != "feria" {
		return resultWithFilteredComms(s, commsIfSWins)
	}
	if s.NameResID == "feria" && t.NameResID != "feria" {
		return resultWithFilteredComms(t, comms)
	}

	if t.LiturgicalClass == ClassI {
		return resultWithFilteredComms(t, comms)
	}
	return resultWithFilteredComms(t, comms)
}

type sanctoralCandidate1962 struct {
	feast       LiturgicalDay
	isBrazilian bool
	order       int
}

func (le *LiturgicalEngine) mergeSanctoralFeasts1962(universal, brazilian []LiturgicalDay) []LiturgicalDay {
	byObservance := make(map[string]sanctoralCandidate1962)
	order := 0
	for _, feast := range universal {
		byObservance[feast.ObservanceKey()] = sanctoralCandidate1962{
			feast:       feast,
			isBrazilian: false,
			order:       order,
		}
		order++
	}
	for _, feast := range brazilian {
		cand := sanctoralCandidate1962{
			feast:       feast,
			isBrazilian: true,
			order:       order,
		}
		order++
		existing, ok := byObservance[feast.ObservanceKey()]
		if !ok || le.isPreferred1962(cand, existing) {
			byObservance[feast.ObservanceKey()] = cand
		}
	}

	var candidates []sanctoralCandidate1962
	for _, c := range byObservance {
		candidates = append(candidates, c)
	}

	sort.Slice(candidates, func(i, j int) bool {
		a := candidates[i]
		b := candidates[j]
		if a.feast.LiturgicalClass != b.feast.LiturgicalClass {
			return a.feast.LiturgicalClass < b.feast.LiturgicalClass
		}
		if a.isBrazilian != b.isBrazilian {
			return a.isBrazilian
		}
		return a.order < b.order
	})

	var result []LiturgicalDay
	for _, c := range candidates {
		result = append(result, c.feast)
	}
	return result
}

func (le *LiturgicalEngine) isPreferred1962(candidate, existing sanctoralCandidate1962) bool {
	if candidate.feast.LiturgicalClass != existing.feast.LiturgicalClass {
		return candidate.feast.LiturgicalClass < existing.feast.LiturgicalClass
	}
	if candidate.isBrazilian != existing.isBrazilian {
		return candidate.isBrazilian
	}
	return candidate.order < existing.order
}

func (le *LiturgicalEngine) GetSeasonColor(date time.Time) LiturgicalColor {
	year := date.Year()
	easter := CalculateEaster(year)

	y1, m1, d1 := date.Date()
	y2, m2, d2 := easter.Date()
	t1 := time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC)
	diffFromEaster := int(t1.Sub(t2).Hours() / 24)

	// White from Easter Vigil to before Pentecost Vigil
	if diffFromEaster >= 0 && diffFromEaster <= 48 {
		return ColorWhite
	}
	// Red during Pentecost Octave
	if diffFromEaster >= 49 && diffFromEaster <= 55 {
		return ColorRed
	}
	if diffFromEaster >= -63 && diffFromEaster <= -1 {
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
