package engine

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

type XMLSanctorale struct {
	XMLName xml.Name       `xml:"sanctorale"`
	Feasts  []XMLFeastElem `xml:"feast"`
}

type XMLFeastElem struct {
	Date        string `xml:"date,attr"`
	RelativeTo  string `xml:"relativeTo,attr"`
	OffsetDays  string `xml:"offsetDays,attr"`
	Name        string `xml:"name,attr"`
	NameResID   string `xml:"nameResId,attr"`
	Class       string `xml:"class,attr"`
	Color       string `xml:"color,attr"`
	IsLordFeast string `xml:"isLordFeast,attr"`
	Epistle     string `xml:"epistle"`
	Gospel      string `xml:"gospel"`
}

type Sanctorale struct {
	Feasts map[string][]LiturgicalDay
}

func NewSanctorale(xmlPath string) *Sanctorale {
	s := &Sanctorale{
		Feasts: make(map[string][]LiturgicalDay),
	}
	s.parse(xmlPath)
	return s
}

func (s *Sanctorale) parse(xmlPath string) {
	file, err := os.Open(xmlPath)
	if err != nil {
		fmt.Printf("Error opening sanctorale %s: %v\n", xmlPath, err)
		return
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading sanctorale %s: %v\n", xmlPath, err)
		return
	}

	var xmlSanc XMLSanctorale
	err = xml.Unmarshal(byteValue, &xmlSanc)
	if err != nil {
		fmt.Printf("Error unmarshaling sanctorale %s: %v\n", xmlPath, err)
		return
	}

	for _, f := range xmlSanc.Feasts {
		if f.Date == "" {
			continue
		}
		isLordFeast := f.IsLordFeast == "true"
		day := LiturgicalDay{
			ID:              f.NameResID,
			Name:            f.Name,
			NameResID:       f.NameResID,
			LiturgicalClass: ParseClass(f.Class),
			Color:           ParseColor(f.Color),
			IsLordFeast:     isLordFeast,
			Epistle:         f.Epistle,
			Gospel:          f.Gospel,
			CalendarVersion: Calendar1962,
		}
		s.Feasts[f.Date] = append(s.Feasts[f.Date], day)
	}
}

func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func (s *Sanctorale) GetAllFeasts(date time.Time) []LiturgicalDay {
	year := date.Year()
	month := date.Month()
	day := date.Day()

	var result []LiturgicalDay

	// Christ the King is the last Sunday of October
	oct31 := time.Date(year, 10, 31, 0, 0, 0, 0, time.UTC)
	var christTheKing time.Time
	if oct31.Weekday() == time.Sunday {
		christTheKing = oct31
	} else {
		christTheKing = oct31.AddDate(0, 0, -int(oct31.Weekday()))
	}

	if date.Year() == christTheKing.Year() && date.Month() == christTheKing.Month() && date.Day() == christTheKing.Day() {
		result = append(result, LiturgicalDay{
			ID:              "christ_the_king",
			Name:            "Feast of Christ the King",
			NameResID:       "christ_the_king",
			LiturgicalClass: ClassI,
			Color:           ColorWhite,
			IsLordFeast:     true,
			CalendarVersion: Calendar1962,
		})
	}

	key := ""
	if isLeapYear(year) && month == time.February {
		if day == 24 {
			key = "02-24-LEAP-NONE"
		} else if day == 25 {
			key = "02-24"
		} else {
			key = fmt.Sprintf("%02d-%02d", month, day)
		}
	} else {
		key = fmt.Sprintf("%02d-%02d", month, day)
	}

	if feasts, ok := s.Feasts[key]; ok {
		result = append(result, feasts...)
	}

	return result
}

func (s *Sanctorale) GetDay(date time.Time) *LiturgicalDay {
	feasts := s.GetAllFeasts(date)
	if len(feasts) > 0 {
		return &feasts[0]
	}
	return nil
}

type BrazilianFeastEntry struct {
	Day        LiturgicalDay
	FixedDate  string
	RelativeTo string
	OffsetDays int
}

type BrazilianSanctorale struct {
	Entries []BrazilianFeastEntry
}

func NewBrazilianSanctorale(xmlPath string) *BrazilianSanctorale {
	bs := &BrazilianSanctorale{}
	bs.parse(xmlPath)
	return bs
}

func (bs *BrazilianSanctorale) parse(xmlPath string) {
	file, err := os.Open(xmlPath)
	if err != nil {
		fmt.Printf("Error opening brazilian sanctorale %s: %v\n", xmlPath, err)
		return
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading brazilian sanctorale %s: %v\n", xmlPath, err)
		return
	}

	var xmlSanc XMLSanctorale
	err = xml.Unmarshal(byteValue, &xmlSanc)
	if err != nil {
		fmt.Printf("Error unmarshaling brazilian sanctorale %s: %v\n", xmlPath, err)
		return
	}

	for _, f := range xmlSanc.Feasts {
		isLordFeast := f.IsLordFeast == "true"
		day := LiturgicalDay{
			ID:              f.NameResID,
			Name:            f.Name,
			NameResID:       f.NameResID,
			LiturgicalClass: ParseClass(f.Class),
			Color:           ParseColor(f.Color),
			IsLordFeast:     isLordFeast,
			Epistle:         f.Epistle,
			Gospel:          f.Gospel,
			CalendarVersion: Calendar1962,
		}

		offset := 0
		if f.OffsetDays != "" {
			offset, _ = strconv.Atoi(f.OffsetDays)
		}

		bs.Entries = append(bs.Entries, BrazilianFeastEntry{
			Day:        day,
			FixedDate:  f.Date,
			RelativeTo: f.RelativeTo,
			OffsetDays: offset,
		})
	}
}

func (bs *BrazilianSanctorale) GetAllFeasts(date time.Time) []LiturgicalDay {
	var result []LiturgicalDay
	key := fmt.Sprintf("%02d-%02d", date.Month(), date.Day())

	for _, entry := range bs.Entries {
		if entry.FixedDate != "" && entry.FixedDate == key {
			result = append(result, entry.Day)
			continue
		}

		if entry.RelativeTo == "sacred_heart" {
			// Sacred Heart is Easter + 68 days (Friday after Corpus Christi octave / 19 days after Pentecost)
			easter := CalculateEaster(date.Year())
			sacredHeart := easter.AddDate(0, 0, 68)
			target := sacredHeart.AddDate(0, 0, entry.OffsetDays)
			if date.Year() == target.Year() && date.Month() == target.Month() && date.Day() == target.Day() {
				result = append(result, entry.Day)
			}
		}
	}
	return result
}

func (bs *BrazilianSanctorale) GetDay(date time.Time) *LiturgicalDay {
	feasts := bs.GetAllFeasts(date)
	if len(feasts) > 0 {
		return &feasts[0]
	}
	return nil
}
