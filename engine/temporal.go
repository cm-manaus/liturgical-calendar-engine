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
	Offset      string          `xml:"offset,attr"`
	Date        string          `xml:"date,attr"`
	Name        string          `xml:"name,attr"`
	NameResID   string          `xml:"nameResId,attr"`
	Class       string          `xml:"class,attr"`
	Color       string          `xml:"color,attr"`
	IsLordFeast string          `xml:"isLordFeast,attr"`
	NameArg     string          `xml:"nameArg,attr"`
	Liturgia    *XMLDayLiturgia `xml:"liturgia"`
	Epistle     string          `xml:"epistle"`
	Gospel      string          `xml:"gospel"`
}

type XMLDayLiturgia struct {
	Epistle string `xml:"epistle"`
	Gospel  string `xml:"gospel"`
}

type TemporalCycle struct {
	EasterCycle    map[int]LiturgicalDay
	ChristmasCycle map[string]LiturgicalDay
}

// Proper readings for Sundays after Pentecost in 1962
var sundayReadingsAfterPentecost = map[int][2]string{
	1:  {"Rm 11, 33-36", "Mt 28, 18-20"}, // Holy Trinity
	2:  {"1Jo 3, 13-18", "Lc 14, 16-24"},
	3:  {"1Pd 5, 6-11", "Lc 15, 1-10"},
	4:  {"Rm 8, 18-23", "Lc 5, 1-11"},
	5:  {"1Pd 3, 8-15a", "Mt 5, 20-24"},
	6:  {"Rm 6, 3-11", "Mc 8, 1-9"},
	7:  {"Rm 6, 19-23", "Mt 7, 15-21"},
	8:  {"Rm 8, 12-17", "Lc 16, 1-9"},
	9:  {"1Co 10, 6-13", "Lc 19, 41-47"},
	10: {"1Co 12, 2-11", "Lc 18, 9-14"},
	11: {"1Co 15, 1-10", "Mc 7, 31-37"},
	12: {"2Co 3, 4-9", "Lc 10, 23-37"},
	13: {"Gl 3, 16-22", "Lc 17, 11-19"},
	14: {"Gl 5, 16-24", "Mt 6, 24-33"},
	15: {"Gl 5, 25-26; 6, 1-10", "Lc 7, 11-16"},
	16: {"Ef 3, 13-21", "Lc 14, 1-11"},
	17: {"Ef 4, 1-6", "Mt 22, 34-46"},
	18: {"1Co 1, 4-8", "Mt 9, 1-8"},
	19: {"Ef 4, 23-28", "Mt 22, 1-14"},
	20: {"Ef 5, 15-21", "Jo 4, 46-53"},
	21: {"Ef 6, 10-17", "Mt 18, 23-35"},
	22: {"Flp 1, 6-11", "Mt 22, 15-21"},
	23: {"Flp 3, 17-21; 4, 1-3", "Mt 9, 18-26"},
	24: {"Cl 1, 9-14", "Mt 24, 15-35"},
}

func getAdvent1(year int) time.Time {
	nov27 := time.Date(year, 11, 27, 0, 0, 0, 0, time.UTC)
	dow := int(nov27.Weekday())
	if dow == 0 { // Sunday
		return nov27
	}
	return nov27.AddDate(0, 0, 7-dow)
}

func getFerialReadings1962(date time.Time) (string, string, bool) {
	daysBack := int(date.Weekday())
	if daysBack == 0 {
		daysBack = 7
	}
	prevSunday := date.AddDate(0, 0, -daysBack)

	easterDate := CalculateEaster(date.Year())
	pentecost := easterDate.AddDate(0, 0, 49)
	advent1 := getAdvent1(date.Year())

	if prevSunday.After(pentecost) && prevSunday.Before(advent1) {
		diffFromEaster := int(prevSunday.Sub(easterDate).Hours() / 24)
		sundayNum := ((diffFromEaster - 56) / 7) + 1
		if readings, ok := sundayReadingsAfterPentecost[sundayNum]; ok {
			return readings[0], readings[1], true
		}
	}

	if (date.Equal(advent1) || date.After(advent1)) && date.Before(time.Date(date.Year(), 12, 25, 0, 0, 0, 0, time.UTC)) {
		return "Rm 13, 11-14", "Lc 21, 25-33", true
	}

	return "", "", false
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

	epistle := d.Epistle
	gospel := d.Gospel
	if d.Liturgia != nil {
		if epistle == "" {
			epistle = d.Liturgia.Epistle
		}
		if gospel == "" {
			gospel = d.Liturgia.Gospel
		}
	}

	return LiturgicalDay{
		ID:              d.NameResID,
		Name:            d.Name,
		NameResID:       d.NameResID,
		NameArgs:        nameArgs,
		LiturgicalClass: ParseClass(d.Class),
		Color:           ParseColor(d.Color),
		IsLordFeast:     isLordFeast,
		Epistle:         epistle,
		Gospel:          gospel,
		CalendarVersion: Calendar1962,
	}
}

func (tc *TemporalCycle) GetDay(date time.Time) *LiturgicalDay {
	year := date.Year()
	easter := CalculateEaster(year)

	y1, m1, d1 := date.Date()
	y2, m2, d2 := easter.Date()
	t1 := time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC)
	diffFromEaster := int(t1.Sub(t2).Hours() / 24)

	septuagesima := easter.AddDate(0, 0, -63)
	christmas := time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC)

	advent1 := getAdvent1(year)

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
			ID:              "holy_family",
			Name:            "Feast of the Holy Family",
			NameResID:       "holy_family",
			LiturgicalClass: ClassII,
			Color:           ColorWhite,
			CalendarVersion: Calendar1962,
		}
	}

	// 1. Easter Cycle (from XML)
	if fromEaster, ok := tc.EasterCycle[diffFromEaster]; ok {
		day := fromEaster
		if day.NameResID == "feria" {
			if ep, gosp, found := getFerialReadings1962(t1); found {
				day.Epistle = ep
				day.Gospel = gosp
			}
		}
		return &day
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
				ID:              "sunday_octave_christmas",
				Name:            "Sunday within the Octave of Christmas",
				NameResID:       "sunday_octave_christmas",
				LiturgicalClass: ClassII,
				Color:           ColorWhite,
				CalendarVersion: Calendar1962,
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
				ID:              fromChristmas.ID,
				Name:            fromChristmas.Name,
				NameResID:       fromChristmas.NameResID,
				NameArgs:        []any{arg},
				LiturgicalClass: fromChristmas.LiturgicalClass,
				Color:           fromChristmas.Color,
				IsLordFeast:     fromChristmas.IsLordFeast,
				Epistle:         fromChristmas.Epistle,
				Gospel:          fromChristmas.Gospel,
				CalendarVersion: Calendar1962,
			}
		}
		day := fromChristmas
		return &day
	}

	// 3. Septuagesima Time Logic
	ashWednesdayOffset := -46
	if (date.After(septuagesima) || date.Equal(septuagesima)) && diffFromEaster < ashWednesdayOffset {
		if date.Weekday() != time.Sunday {
			return &LiturgicalDay{
				ID:              "feria",
				Name:            "Feria of Septuagesima",
				NameResID:       "feria",
				LiturgicalClass: ClassIV,
				Color:           ColorViolet,
				CalendarVersion: Calendar1962,
			}
		}
	}

	// 4. Lenten Ferias (III Class)
	if diffFromEaster < 0 && diffFromEaster > -46 && date.Weekday() != time.Sunday {
		return &LiturgicalDay{
			ID:              "feria",
			Name:            "Feria of Lent",
			NameResID:       "feria",
			LiturgicalClass: ClassIII,
			Color:           ColorViolet,
			CalendarVersion: Calendar1962,
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
				ID:              "sunday_after_epiphany",
				Name:            fmt.Sprintf("%s Sunday after Epiphany", getEnglishOrdinal(i)),
				NameResID:       "sunday_after_epiphany",
				NameArgs:        []any{arg},
				LiturgicalClass: ClassII,
				Color:           ColorGreen,
				CalendarVersion: Calendar1962,
			}
		}
	}

	// 6. Sundays after Pentecost
	if diffFromEaster > 56 && date.Weekday() == time.Sunday && date.Before(advent1) {
		trinityPlus7 := easter.AddDate(0, 0, 56)
		totalDays := int(advent1.Sub(trinityPlus7).Hours()/24) - 1
		totalSundaysAfterPentecost := totalDays/7 + 1
		sundayNum := (diffFromEaster-56)/7 + 1

		if sundayNum == totalSundaysAfterPentecost {
			ep, gosp := "", ""
			if r, ok := sundayReadingsAfterPentecost[24]; ok {
				ep, gosp = r[0], r[1]
			}
			return &LiturgicalDay{
				ID:              "sunday_after_pentecost",
				Name:            "24th Sunday after Pentecost",
				NameResID:       "sunday_after_pentecost",
				NameArgs:        []any{"ord_24"},
				LiturgicalClass: ClassII,
				Color:           ColorGreen,
				Epistle:         ep,
				Gospel:          gosp,
				CalendarVersion: Calendar1962,
			}
		}

		if totalSundaysAfterPentecost > 24 {
			numExtra := totalSundaysAfterPentecost - 24
			if sundayNum > 23 && sundayNum <= 23+numExtra {
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
						ID:              "sunday_after_epiphany",
						Name:            fmt.Sprintf("%s Sunday after Epiphany (Resumed)", getEnglishOrdinal(currentOmitted)),
						NameResID:       "sunday_after_epiphany",
						NameArgs:        []any{arg},
						LiturgicalClass: ClassII,
						Color:           ColorGreen,
						CalendarVersion: Calendar1962,
					}
				}
			}

			if sundayNum > 23+numExtra {
				ep, gosp := "", ""
				if r, ok := sundayReadingsAfterPentecost[24]; ok {
					ep, gosp = r[0], r[1]
				}
				return &LiturgicalDay{
					ID:              "sunday_after_pentecost",
					Name:            "24th Sunday after Pentecost",
					NameResID:       "sunday_after_pentecost",
					NameArgs:        []any{"ord_24"},
					LiturgicalClass: ClassII,
					Color:           ColorGreen,
					Epistle:         ep,
					Gospel:          gosp,
					CalendarVersion: Calendar1962,
				}
			}
		}

		if sundayNum <= 23 {
			var arg any = sundayNum
			if sundayNum >= 1 && sundayNum <= 5 {
				arg = fmt.Sprintf("ord_%d", sundayNum)
			}
			ep, gosp := "", ""
			if r, ok := sundayReadingsAfterPentecost[sundayNum]; ok {
				ep, gosp = r[0], r[1]
			}
			return &LiturgicalDay{
				ID:              "sunday_after_pentecost",
				Name:            fmt.Sprintf("%s Sunday after Pentecost", getEnglishOrdinal(sundayNum)),
				NameResID:       "sunday_after_pentecost",
				NameArgs:        []any{arg},
				LiturgicalClass: ClassII,
				Color:           ColorGreen,
				Epistle:         ep,
				Gospel:          gosp,
				CalendarVersion: Calendar1962,
			}
		}
	}

	// 7. Ember Days of September (First Wednesday, Friday, Saturday after the 3rd Sunday of September in 1962 rubrics)
	thirdSundayOfSept := thirdSundayOfSeptember(year)
	emberWednesday := nextWeekdayAfter(thirdSundayOfSept, int(time.Wednesday))
	emberFriday := nextWeekdayAfter(thirdSundayOfSept, int(time.Friday))
	emberSaturday := nextWeekdayAfter(thirdSundayOfSept, int(time.Saturday))

	if date.Equal(emberWednesday) {
		return &LiturgicalDay{
			ID:              "ember_wednesday_september",
			Name:            "Ember Wednesday of September",
			NameResID:       "ember_wednesday",
			NameArgs:        []any{"of_september"},
			LiturgicalClass: ClassII,
			Color:           ColorViolet,
			Epistle:         "Am 9,13-15; Ne 8,1-10",
			Gospel:          "Mc 9,16-28",
			CalendarVersion: Calendar1962,
		}
	}
	if date.Equal(emberFriday) {
		return &LiturgicalDay{
			ID:              "ember_friday_september",
			Name:            "Ember Friday of September",
			NameResID:       "ember_friday",
			NameArgs:        []any{"of_september"},
			LiturgicalClass: ClassII,
			Color:           ColorViolet,
			Epistle:         "Os 14,2-10",
			Gospel:          "Lc 7,36-50",
			CalendarVersion: Calendar1962,
		}
	}
	if date.Equal(emberSaturday) {
		return &LiturgicalDay{
			ID:              "ember_saturday_september",
			Name:            "Ember Saturday of September",
			NameResID:       "ember_saturday",
			NameArgs:        []any{"of_september"},
			LiturgicalClass: ClassII,
			Color:           ColorViolet,
			Epistle:         "Lv 23,26-32; Lv 23,39-43; Miq 7,14; 7,16; 7,18-20; Za 8,14-19; Dn 3,47-51; 3,52-59; Hb 9,2-12",
			Gospel:          "Lc 13,6-17",
			CalendarVersion: Calendar1962,
		}
	}

	// 8. Holy Name of Jesus
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
			ID:              "holy_name_jesus",
			Name:            "Most Holy Name of Jesus",
			NameResID:       "holy_name_jesus",
			LiturgicalClass: ClassII,
			Color:           ColorWhite,
			IsLordFeast:     true,
			CalendarVersion: Calendar1962,
		}
	}

	// 9. Advent Cycle Logic
	diffFromAdvent1 := int(t1.Sub(advent1).Hours() / 24)

	if date.Before(christmas) {
		if date.Equal(advent1) {
			return &LiturgicalDay{
				ID:              "sunday_of_advent_1",
				Name:            "1st Sunday of Advent",
				NameResID:       "sunday_of_advent",
				NameArgs:        []any{"ord_1"},
				LiturgicalClass: ClassI,
				Color:           ColorViolet,
				CalendarVersion: Calendar1962,
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
					ID:              fmt.Sprintf("sunday_of_advent_%d", sundayNum),
					Name:            fmt.Sprintf("%s Sunday of Advent", getEnglishOrdinal(sundayNum)),
					NameResID:       "sunday_of_advent",
					NameArgs:        []any{arg},
					LiturgicalClass: ClassI,
					Color:           color,
					CalendarVersion: Calendar1962,
				}
			}

			if sundayNum == 3 {
				dow := date.Weekday()
				if dow == time.Wednesday {
					return &LiturgicalDay{
						ID:              "ember_wednesday_advent",
						Name:            "Ember Wednesday of Advent",
						NameResID:       "ember_wednesday",
						NameArgs:        []any{"of_advent"},
						LiturgicalClass: ClassII,
						Color:           ColorViolet,
						Epistle:         "Is 2,2-5; Is 7,10-15",
						Gospel:          "Lc 1,26-38",
						CalendarVersion: Calendar1962,
					}
				} else if dow == time.Friday {
					return &LiturgicalDay{
						ID:              "ember_friday_advent",
						Name:            "Ember Friday of Advent",
						NameResID:       "ember_friday",
						NameArgs:        []any{"of_advent"},
						LiturgicalClass: ClassII,
						Color:           ColorViolet,
						Epistle:         "Is 11,1-5",
						Gospel:          "Lc 1,39-47",
						CalendarVersion: Calendar1962,
					}
				} else if dow == time.Saturday {
					return &LiturgicalDay{
						ID:              "ember_saturday_advent",
						Name:            "Ember Saturday of Advent",
						NameResID:       "ember_saturday",
						NameArgs:        []any{"of_advent"},
						LiturgicalClass: ClassII,
						Color:           ColorViolet,
						Epistle:         "Is 19,20-22; Is 35,1-7; Is 40,9-11; Is 45,1-8; Dn 3,47-51; 3,52-59; 2Ts 2,1-8",
						Gospel:          "Lc 3,1-6",
						CalendarVersion: Calendar1962,
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

func thirdSundayOfSeptember(year int) time.Time {
	september1 := time.Date(year, 9, 1, 0, 0, 0, 0, time.UTC)
	daysUntilFirstSunday := (int(time.Sunday) - int(september1.Weekday()) + 7) % 7
	return september1.AddDate(0, 0, daysUntilFirstSunday+14)
}

func nextWeekdayAfter(val time.Time, weekday int) time.Time {
	target := val.AddDate(0, 0, 1)
	for int(target.Weekday()) != weekday {
		target = target.AddDate(0, 0, 1)
	}
	return target
}
