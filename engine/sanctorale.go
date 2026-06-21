package engine

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"time"
)

type XMLSanctorale struct {
	XMLName xml.Name       `xml:"sanctorale"`
	Feasts  []XMLFeastElem `xml:"feast"`
}

type XMLFeastElem struct {
	Date        string `xml:"date,attr"`
	Name        string `xml:"name,attr"`
	NameResID   string `xml:"nameResId,attr"`
	Class       string `xml:"class,attr"`
	Color       string `xml:"color,attr"`
	IsLordFeast string `xml:"isLordFeast,attr"`
}

type Sanctorale struct {
	Feasts map[string]LiturgicalDay
}

func NewSanctorale(xmlPath string) *Sanctorale {
	s := &Sanctorale{
		Feasts: make(map[string]LiturgicalDay),
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
		isLordFeast := f.IsLordFeast == "true"
		s.Feasts[f.Date] = LiturgicalDay{
			Name:            f.Name,
			NameResID:       f.NameResID,
			LiturgicalClass: parseClass(f.Class),
			Color:           parseColor(f.Color),
			IsLordFeast:     isLordFeast,
		}
	}
}

func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func (s *Sanctorale) GetDay(date time.Time) *LiturgicalDay {
	year := date.Year()
	month := date.Month()
	day := date.Day()

	// Christ the King is the Sunday preceding the feast of All Saints (Nov 1)
	// which is the last Sunday of October (same as Oct 31 or the Sunday before it)
	oct31 := time.Date(year, 10, 31, 0, 0, 0, 0, time.UTC)
	var christTheKing time.Time
	if oct31.Weekday() == time.Sunday {
		christTheKing = oct31
	} else {
		christTheKing = oct31.AddDate(0, 0, -int(oct31.Weekday()))
	}

	if date.Year() == christTheKing.Year() && date.Month() == christTheKing.Month() && date.Day() == christTheKing.Day() {
		return &LiturgicalDay{
			Name:            "Feast of Christ the King",
			NameResID:       "christ_the_king",
			LiturgicalClass: ClassI,
			Color:           ColorWhite,
			IsLordFeast:     true,
		}
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

	if feast, ok := s.Feasts[key]; ok {
		return &feast
	}
	return nil
}

type BrazilianSanctorale struct {
	Feasts map[string]LiturgicalDay
}

func NewBrazilianSanctorale(xmlPath string) *BrazilianSanctorale {
	bs := &BrazilianSanctorale{
		Feasts: make(map[string]LiturgicalDay),
	}
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
		bs.Feasts[f.Date] = LiturgicalDay{
			Name:            f.Name,
			NameResID:       f.NameResID,
			LiturgicalClass: parseClass(f.Class),
			Color:           parseColor(f.Color),
			IsLordFeast:     isLordFeast,
		}
	}
}

func (bs *BrazilianSanctorale) GetDay(date time.Time) *LiturgicalDay {
	key := fmt.Sprintf("%02d-%02d", date.Month(), date.Day())
	if feast, ok := bs.Feasts[key]; ok {
		return &feast
	}
	return nil
}
