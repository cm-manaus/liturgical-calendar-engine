package engine

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

type XMLTemporal struct {
	XMLName        xml.Name          `xml:"temporal"`
	EasterCycle    XMLEasterCycle    `xml:"easter_cycle"`
	ChristmasCycle XMLChristmasCycle `xml:"christmas_cycle"`
}

type XMLEasterCycle struct {
	Days []XMLTemporalDay `xml:"day"`
}

type XMLChristmasCycle struct {
	Days []XMLTemporalDay `xml:"day"`
}

type XMLTemporalDay struct {
	Offset      string `xml:"offset,attr"`
	Date        string `xml:"date,attr"`
	Name        string `xml:"name,attr"`
	NameResID   string `xml:"nameResId,attr"`
	Class       string `xml:"class,attr"`
	Color       string `xml:"color,attr"`
	IsLordFeast string `xml:"isLordFeast,attr"`
	NameArg     string `xml:"nameArg,attr"`
}

type TemporalCycle struct {
	EasterCycle    map[int]LiturgicalDay
	ChristmasCycle map[string]LiturgicalDay
}

func parseClass(s string) LiturgicalClass {
	switch s {
	case "I":
		return ClassI
	case "II":
		return ClassII
	case "III":
		return ClassIII
	case "IV":
		return ClassIV
	default:
		return ClassIII
	}
}

func parseColor(s string) LiturgicalColor {
	switch s {
	case "WHITE":
		return ColorWhite
	case "RED":
		return ColorRed
	case "GREEN":
		return ColorGreen
	case "VIOLET":
		return ColorViolet
	case "BLACK":
		return ColorBlack
	case "ROSE":
		return ColorRose
	default:
		return ColorWhite
	}
}

func NewTemporalCycle(xmlPath string) *TemporalCycle {
	tc := &TemporalCycle{
		EasterCycle:    make(map[int]LiturgicalDay),
		ChristmasCycle: make(map[string]LiturgicalDay),
	}
	tc.parse(xmlPath)
	return tc
}

func (tc *TemporalCycle) parse(xmlPath string) {
	file, err := os.Open(xmlPath)
	if err != nil {
		fmt.Printf("Error opening temporal cycle %s: %v\n", xmlPath, err)
		return
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading temporal cycle %s: %v\n", xmlPath, err)
		return
	}

	var xmlTemp XMLTemporal
	err = xml.Unmarshal(byteValue, &xmlTemp)
	if err != nil {
		fmt.Printf("Error unmarshaling temporal cycle %s: %v\n", xmlPath, err)
		return
	}

	for _, d := range xmlTemp.EasterCycle.Days {
		offset, err := strconv.Atoi(d.Offset)
		if err == nil {
			tc.EasterCycle[offset] = tc.parseDay(d)
		}
	}

	for _, d := range xmlTemp.ChristmasCycle.Days {
		if d.Date != "" {
			tc.ChristmasCycle[d.Date] = tc.parseDay(d)
		}
	}
}

func (tc *TemporalCycle) parseDay(d XMLTemporalDay) LiturgicalDay {
	isLordFeast := d.IsLordFeast == "true"
	var nameArgs []any

	if d.NameArg != "" {
		val, err := strconv.Atoi(d.NameArg)
		if err == nil {
			if val >= 1 && val <= 5 {
				nameArgs = []any{fmt.Sprintf("ord_%d", val)}
			} else {
				nameArgs = []any{val}
			}
		} else {
			nameArgs = []any{d.NameArg}
		}
	}

	return LiturgicalDay{
		Name:            d.Name,
		NameResID:       d.NameResID,
		NameArgs:        nameArgs,
		LiturgicalClass: parseClass(d.Class),
		Color:           parseColor(d.Color),
		IsLordFeast:     isLordFeast,
	}
}

func (tc *TemporalCycle) GetDay(date time.Time) *LiturgicalDay {
	year := date.Year()
	easter := CalculateEaster(year)
	diffFromEaster := int(date.Sub(easter).Hours() / 24)

	// Keep differences precise for negative offsets
	// Go duration calculation has slight rounding differences on DST transitions,
	// so it is better to calculate difference using Date arithmetic:
	y1, m1, d1 := date.Date()
	y2, m2, d2 := easter.Date()
	t1 := time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC)
	diffFromEaster = int(t1.Sub(t2).Hours() / 24)

	septuagesima := easter.AddDate(0, 0, -63)
	christmas := time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC)

	// advent1 logic
	nov27 := time.Date(year, 11, 27, 0, 0, 0, 0, time.UTC)
	var advent1 time.Time
	if nov27.Weekday() == time.Sunday {
		advent1 = nov27
	} else {
		advent1 = nov27.AddDate(0, 0, 7-int(nov27.Weekday()))
	}

	// 5. Sundays after Epiphany
	epiphany := time.Date(year, 1, 6, 0, 0, 0, 0, time.UTC)
	var sundayAfterEpiphany1 time.Time
	if epiphany.Weekday() == time.Sunday {
		sundayAfterEpiphany1 = epiphany.AddDate(0, 0, 7)
	} else {
		sundayAfterEpiphany1 = epiphany.AddDate(0, 0, 7-int(epiphany.Weekday()))
	}

	if date.Equal(sundayAfterEpiphany1) && date.Before(septuagesima) {
		return &LiturgicalDay{
			Name:            "Feast of the Holy Family",
			NameResID:       "holy_family",
			LiturgicalClass: ClassII,
			Color:           ColorWhite,
		}
	}

	// 1. Easter Cycle (from XML)
	if fromEaster, ok := tc.EasterCycle[diffFromEaster]; ok {
		return &fromEaster
	}

	// 2. Christmas Cycle (from XML fixed dates and Sunday within Octave)
	if date.Month() == 12 && date.Day() >= 26 && date.Day() <= 31 {
		var sundayWithinOctave time.Time
		for i := 1; i <= 6; i++ {
			d := christmas.AddDate(0, 0, i)
			if d.Weekday() == time.Sunday {
				sundayWithinOctave = d
				break
			}
		}
		if !sundayWithinOctave.IsZero() && date.Equal(sundayWithinOctave) {
			return &LiturgicalDay{
				Name:            "Sunday within the Octave of Christmas",
				NameResID:       "sunday_octave_christmas",
				LiturgicalClass: ClassII,
				Color:           ColorWhite,
			}
		}
	}

	christmasKey := fmt.Sprintf("%02d-%02d", date.Month(), date.Day())
	if fromChristmas, ok := tc.ChristmasCycle[christmasKey]; ok {
		if fromChristmas.NameResID == "day_octave_nativity" {
			dayNum := date.Day() - 24
			var arg any = dayNum
			if dayNum >= 1 && dayNum <= 5 {
				arg = fmt.Sprintf("ord_%d", dayNum)
			}
			return &LiturgicalDay{
				Name:            fromChristmas.Name,
				NameResID:       fromChristmas.NameResID,
				NameArgs:        []any{arg},
				LiturgicalClass: fromChristmas.LiturgicalClass,
				Color:           fromChristmas.Color,
				IsLordFeast:     fromChristmas.IsLordFeast,
			}
		}
		return &fromChristmas
	}

	// 3. Septuagesima Time Logic
	// septuagesima time: from septuagesima Sunday until Ash Wednesday (-46 offset)
	ashWednesdayOffset := -46
	if (date.After(septuagesima) || date.Equal(septuagesima)) && diffFromEaster < ashWednesdayOffset {
		if date.Weekday() != time.Sunday {
			return &LiturgicalDay{
				Name:            "Feria of Septuagesima",
				NameResID:       "feria",
				LiturgicalClass: ClassIV,
				Color:           ColorViolet,
			}
		}
	}

	// 4. Lenten Ferias (III Class)
	if diffFromEaster < 0 && diffFromEaster > -46 && date.Weekday() != time.Sunday {
		return &LiturgicalDay{
			Name:            "Feria of Lent",
			NameResID:       "feria",
			LiturgicalClass: ClassIII,
			Color:           ColorViolet,
		}
	}

	// Epiphany Sundays (2nd to 6th)
	for i := 2; i <= 6; i++ {
		sunday := sundayAfterEpiphany1.AddDate(0, 0, 7*int(i-1))
		if date.Equal(sunday) && date.Before(septuagesima) {
			var arg any = i
			if i >= 1 && i <= 5 {
				arg = fmt.Sprintf("ord_%d", i)
			}
			return &LiturgicalDay{
				Name:            fmt.Sprintf("%s Sunday after Epiphany", getEnglishOrdinal(i)),
				NameResID:       "sunday_after_epiphany",
				NameArgs:        []any{arg},
				LiturgicalClass: ClassII,
				Color:           ColorGreen,
			}
		}
	}

	// 6. Sundays after Pentecost
	pentecostSunday := easter.AddDate(0, 0, 49) // offset 49
	_ = pentecostSunday
	if diffFromEaster > 56 && date.Weekday() == time.Sunday && date.Before(advent1) {
		// Total sundays between Pentecost+56 (Trinity Sunday + 7) and Advent 1
		trinityPlus7 := easter.AddDate(0, 0, 56)
		totalDays := int(advent1.Sub(trinityPlus7).Hours()/24) - 1
		totalSundaysAfterPentecost := totalDays/7 + 1
		sundayNum := (diffFromEaster-56)/7 + 1

		if sundayNum == totalSundaysAfterPentecost {
			return &LiturgicalDay{
				Name:            "24th Sunday after Pentecost",
				NameResID:       "sunday_after_pentecost",
				NameArgs:        []any{"ord_24"},
				LiturgicalClass: ClassII,
				Color:           ColorGreen,
			}
		}

		if totalSundaysAfterPentecost > 24 {
			numExtra := totalSundaysAfterPentecost - 24
			if sundayNum > 23 && sundayNum <= 23+numExtra {
				// Resumed Sunday after Epiphany
				septOffset := int(septuagesima.Sub(sundayAfterEpiphany1).Hours() / 24)
				lastEpiphanySunday := (septOffset-1)/7 + 1
				omittedStart := lastEpiphanySunday + 1
				currentOmitted := omittedStart + (sundayNum - 24)
				if currentOmitted >= 3 && currentOmitted <= 6 {
					var arg any = currentOmitted
					if currentOmitted >= 1 && currentOmitted <= 5 {
						arg = fmt.Sprintf("ord_%d", currentOmitted)
					}
					return &LiturgicalDay{
						Name:            fmt.Sprintf("%s Sunday after Epiphany (Resumed)", getEnglishOrdinal(currentOmitted)),
						NameResID:       "sunday_after_epiphany",
						NameArgs:        []any{arg},
						LiturgicalClass: ClassII,
						Color:           ColorGreen,
					}
				}
			}

			if sundayNum > 23+numExtra {
				return &LiturgicalDay{
					Name:            "24th Sunday after Pentecost",
					NameResID:       "sunday_after_pentecost",
					NameArgs:        []any{"ord_24"},
					LiturgicalClass: ClassII,
					Color:           ColorGreen,
				}
			}
		}

		if sundayNum <= 23 {
			var arg any = sundayNum
			if sundayNum >= 1 && sundayNum <= 5 {
				arg = fmt.Sprintf("ord_%d", sundayNum)
			}
			return &LiturgicalDay{
				Name:            fmt.Sprintf("%s Sunday after Pentecost", getEnglishOrdinal(sundayNum)),
				NameResID:       "sunday_after_pentecost",
				NameArgs:        []any{arg},
				LiturgicalClass: ClassII,
				Color:           ColorGreen,
			}
		}
	}

	// 7. Ember Days of September (following the 3rd Sunday of September)
	sept1 := time.Date(year, 9, 1, 0, 0, 0, 0, time.UTC)
	var septSunday1 time.Time
	if sept1.Weekday() == time.Sunday {
		septSunday1 = sept1
	} else {
		septSunday1 = sept1.AddDate(0, 0, 7-int(sept1.Weekday()))
	}
	septSunday3 := septSunday1.AddDate(0, 0, 14)

	if date.Equal(septSunday3.AddDate(0, 0, 3)) {
		return &LiturgicalDay{
			Name:            "Ember Wednesday of September",
			NameResID:       "ember_wednesday",
			NameArgs:        []any{"of_september"},
			LiturgicalClass: ClassII,
			Color:           ColorViolet,
		}
	}
	if date.Equal(septSunday3.AddDate(0, 0, 5)) {
		return &LiturgicalDay{
			Name:            "Ember Friday of September",
			NameResID:       "ember_friday",
			NameArgs:        []any{"of_september"},
			LiturgicalClass: ClassII,
			Color:           ColorViolet,
		}
	}
	if date.Equal(septSunday3.AddDate(0, 0, 6)) {
		return &LiturgicalDay{
			Name:            "Ember Saturday of September",
			NameResID:       "ember_saturday",
			NameArgs:        []any{"of_september"},
			LiturgicalClass: ClassII,
			Color:           ColorViolet,
		}
	}

	// 8. Holy Name of Jesus
	// Sunday between Jan 2 and Jan 5, otherwise Jan 2
	var holyNameDate time.Time
	for i := 2; i <= 5; i++ {
		d := time.Date(year, 1, i, 0, 0, 0, 0, time.UTC)
		if d.Weekday() == time.Sunday {
			holyNameDate = d
			break
		}
	}
	if holyNameDate.IsZero() {
		holyNameDate = time.Date(year, 1, 2, 0, 0, 0, 0, time.UTC)
	}

	if date.Equal(holyNameDate) {
		return &LiturgicalDay{
			Name:            "Most Holy Name of Jesus",
			NameResID:       "holy_name_jesus",
			LiturgicalClass: ClassII,
			Color:           ColorWhite,
			IsLordFeast:     true,
		}
	}

	// 9. Advent Cycle Logic
	diffFromAdvent1 := int(date.Sub(advent1).Hours() / 24)
	// Recalculate using date arithmetic to bypass DST
	y1, m1, d1 = date.Date()
	y2, m2, d2 = advent1.Date()
	t1 = time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC)
	t2 = time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC)
	diffFromAdvent1 = int(t1.Sub(t2).Hours() / 24)

	if date.Before(christmas) {
		if date.Equal(advent1) {
			return &LiturgicalDay{
				Name:            "1st Sunday of Advent",
				NameResID:       "sunday_of_advent",
				NameArgs:        []any{"ord_1"},
				LiturgicalClass: ClassI,
				Color:           ColorViolet,
			}
		}
		if diffFromAdvent1 > 0 {
			sundayNum := diffFromAdvent1/7 + 1
			if date.Weekday() == time.Sunday {
				color := ColorViolet
				if sundayNum == 3 {
					color = ColorRose
				}
				var arg any = sundayNum
				if sundayNum >= 1 && sundayNum <= 5 {
					arg = fmt.Sprintf("ord_%d", sundayNum)
				}
				return &LiturgicalDay{
					Name:            fmt.Sprintf("%s Sunday of Advent", getEnglishOrdinal(sundayNum)),
					NameResID:       "sunday_of_advent",
					NameArgs:        []any{arg},
					LiturgicalClass: ClassI,
					Color:           color,
				}
			}

			if sundayNum == 3 {
				dow := date.Weekday()
				if dow == time.Wednesday {
					return &LiturgicalDay{
						Name:            "Ember Wednesday of Advent",
						NameResID:       "ember_wednesday",
						NameArgs:        []any{"of_advent"},
						LiturgicalClass: ClassII,
						Color:           ColorViolet,
					}
				} else if dow == time.Friday {
					return &LiturgicalDay{
						Name:            "Ember Friday of Advent",
						NameResID:       "ember_friday",
						NameArgs:        []any{"of_advent"},
						LiturgicalClass: ClassII,
						Color:           ColorViolet,
					}
				} else if dow == time.Saturday {
					return &LiturgicalDay{
						Name:            "Ember Saturday of Advent",
						NameResID:       "ember_saturday",
						NameArgs:        []any{"of_advent"},
						LiturgicalClass: ClassII,
						Color:           ColorViolet,
					}
				}
			}
		}
	}

	return nil
}

func getEnglishOrdinal(n int) string {
	switch n {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	case 4:
		return "4th"
	}
	return strconv.Itoa(n) + "th"
}
