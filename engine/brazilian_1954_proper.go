package engine

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// Brazilian1954ProperEntry represents a source-backed Brazilian proper in the 1954 local calendar layer.
type Brazilian1954ProperEntry struct {
	ID                 string
	Name               string
	NameResID          string
	NameArgs           []any
	Date               string
	RelativeTo         string
	RelativeOffsetDays *int
	OctaveEndOffset    *int
	Pre55Rank          Pre55Rank
	Color              LiturgicalColor
	IsLordFeast        bool
	Epistle            string
	Gospel             string
	Metadata           CalendarObservanceMetadata
}

func (e Brazilian1954ProperEntry) ToLiturgicalDay() LiturgicalDay {
	meta := e.Metadata
	return LiturgicalDay{
		ID:               e.ID,
		Name:             e.Name,
		NameResID:        e.NameResID,
		NameArgs:         e.NameArgs,
		LiturgicalClass:  ClassIV,
		Color:            e.Color,
		IsLordFeast:      e.IsLordFeast,
		Epistle:          e.Epistle,
		Gospel:           e.Gospel,
		CalendarVersion:  Calendar1954,
		Pre55Rank:        &e.Pre55Rank,
		CalendarMetadata: &meta,
	}
}

func (e Brazilian1954ProperEntry) AsOctaveDay(dayNumber int, date time.Time) Brazilian1954ProperEntry {
	isOctaveDay := dayNumber == 8
	rank := RankSD
	if isOctaveDay {
		rank = RankDMaj
	}

	targetName := e.Name
	anchorID := e.Name
	if e.NameResID != "" {
		anchorID = e.NameResID
	}

	var generatedName string
	var nameResID string
	var nameArgs []any

	if isOctaveDay {
		generatedName = fmt.Sprintf("Octave Day of %s", targetName)
		nameResID = "brazilian_octave_day_8"
		nameArgs = []any{anchorID}
	} else {
		generatedName = fmt.Sprintf("%d%s day within the Octave of %s", dayNumber, ordinalSuffix(dayNumber), targetName)
		nameResID = "brazilian_octave_day"
		nameArgs = []any{fmt.Sprintf("ord_%d", dayNumber), anchorID}
	}

	observanceKind := "within_octave"
	octaveStatus := "day_within"
	if isOctaveDay {
		observanceKind = "octave_day"
		octaveStatus = "octave_day"
	}

	rankVal := rank.Info().Rank
	meta := e.Metadata
	meta.ObservanceKind = observanceKind
	meta.Privileged = boolPtr(false)
	sourceRef := fmt.Sprintf("%s#octave-day-%d", e.ID, dayNumber)
	if e.Metadata.SourceRef != "" {
		sourceRef = fmt.Sprintf("%s#octave-day-%d", e.Metadata.SourceRef, dayNumber)
	}
	meta.SourceRef = sourceRef
	meta.SourceRank = &rankVal
	meta.Precedence = &rankVal
	meta.FirstVespers = boolPtr(false)
	meta.Occurrence = "commemorate"
	meta.Concurrence = "none"
	meta.OctaveDay = &dayNumber
	meta.OctaveStatus = octaveStatus
	meta.VigilKind = "none"
	meta.VigilPrivileged = boolPtr(false)
	meta.VigilAnticipated = boolPtr(false)
	meta.TransferStatus = "none"
	meta.Transferable = boolPtr(false)
	meta.TransferTarget = ""
	meta.MassGloria = boolPtr(true)
	meta.MassCredo = boolPtr(true)

	dateKey := fmt.Sprintf("%02d-%02d", date.Month(), date.Day())
	return Brazilian1954ProperEntry{
		ID:                 fmt.Sprintf("%s_octave_day_%d", e.ID, dayNumber),
		Name:               generatedName,
		NameResID:          nameResID,
		NameArgs:           nameArgs,
		Date:               dateKey,
		Pre55Rank:          rank,
		Color:              e.Color,
		IsLordFeast:        e.IsLordFeast,
		Epistle:            e.Epistle,
		Gospel:             e.Gospel,
		Metadata:           meta,
	}
}

func (e Brazilian1954ProperEntry) HasCommonOctave() bool {
	return e.Date != "" &&
		e.Metadata.OctaveType != nil &&
		*e.Metadata.OctaveType == OctaveCommon &&
		e.Metadata.OctaveID != "" &&
		e.Metadata.OctaveID != "none" &&
		e.OctaveEndOffset != nil &&
		*e.OctaveEndOffset > 0
}

func ordinalSuffix(value int) string {
	if value%100 >= 11 && value%100 <= 13 {
		return "th"
	}
	switch value % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

// Brazilian1954ProperCalendar manages offline Brazilian propers for the 1954 profile.
type Brazilian1954ProperCalendar struct {
	entriesByDate   map[string][]Brazilian1954ProperEntry
	entries         []Brazilian1954ProperEntry
	relativeEntries []Brazilian1954ProperEntry
	loaded          bool
}

func NewBrazilian1954ProperCalendar(path string) *Brazilian1954ProperCalendar {
	cal := &Brazilian1954ProperCalendar{
		entriesByDate: make(map[string][]Brazilian1954ProperEntry),
	}
	if path != "" {
		cal.Load(path)
	}
	return cal
}

func (c *Brazilian1954ProperCalendar) IsLoaded() bool {
	return c.loaded
}

func (c *Brazilian1954ProperCalendar) Load(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return
	}
	c.LoadFromBytes(bytes)
}

func (c *Brazilian1954ProperCalendar) LoadFromBytes(data []byte) {
	var doc XML1954SanctoralDoc
	if err := xml.Unmarshal(data, &doc); err != nil {
		return
	}

	c.entries = nil
	c.relativeEntries = nil
	c.entriesByDate = make(map[string][]Brazilian1954ProperEntry)

	for _, d := range doc.Feasts {
		entry := c.parseEntry(d)
		c.entries = append(c.entries, entry)
		if entry.Date != "" {
			c.entriesByDate[entry.Date] = append(c.entriesByDate[entry.Date], entry)
		} else {
			c.relativeEntries = append(c.relativeEntries, entry)
		}
	}
	c.loaded = true
}

func (c *Brazilian1954ProperCalendar) parseEntry(d XML1954Day) Brazilian1954ProperEntry {
	observanceKind := d.ObservanceKind
	if observanceKind == "" {
		observanceKind = "local_proper"
	}
	rank := Pre55RankFromAsset(d.Pre55Grade, observanceKind, d.Privileged == "true")

	color := ParseColor(d.Color)
	if d.Color == "" {
		color = ColorWhite
	}

	metadata := xmlDayToMetadata(d)

	var octaveEndOffset *int
	if d.Octave != nil && d.Octave.EndOffset != "" {
		if val, err := strconv.Atoi(d.Octave.EndOffset); err == nil {
			octaveEndOffset = &val
		}
	}

	var relativeOffsetDays *int
	if d.OffsetDays != "" {
		if val, err := strconv.Atoi(d.OffsetDays); err == nil {
			relativeOffsetDays = &val
		}
	}

	epistle, gospel := "", ""
	if d.Readings != nil {
		epistle = d.Readings.Epistle
		gospel = d.Readings.Gospel
	}

	nameArgs := parseNameArgs(d.NameArg, d.Name)

	return Brazilian1954ProperEntry{
		ID:                 d.ID,
		Name:               d.Name,
		NameResID:          d.NameResID,
		NameArgs:           nameArgs,
		Date:               d.Date,
		RelativeTo:         d.RelativeTo,
		RelativeOffsetDays: relativeOffsetDays,
		OctaveEndOffset:    octaveEndOffset,
		Pre55Rank:          rank,
		Color:              color,
		IsLordFeast:        d.IsLordFeast == "true",
		Epistle:            epistle,
		Gospel:             gospel,
		Metadata:           metadata,
	}
}

func (c *Brazilian1954ProperCalendar) EntriesForDate(date time.Time) []Brazilian1954ProperEntry {
	if !c.loaded {
		return nil
	}
	dateUTC := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dateKey := fmt.Sprintf("%02d-%02d", dateUTC.Month(), dateUTC.Day())

	var result []Brazilian1954ProperEntry
	if entries, ok := c.entriesByDate[dateKey]; ok {
		result = append(result, entries...)
	}

	for _, entry := range c.relativeEntries {
		if c.relativeMatches(entry, dateUTC) {
			result = append(result, entry)
		}
	}

	for _, anchor := range c.entries {
		if !anchor.HasCommonOctave() {
			continue
		}
		start := parseFixedDate(dateUTC.Year(), anchor.Date)
		if start == nil {
			continue
		}
		elapsed := int(dateUTC.Sub(*start).Hours() / 24)
		if elapsed >= 1 && elapsed <= *anchor.OctaveEndOffset {
			result = append(result, anchor.AsOctaveDay(elapsed+1, dateUTC))
		}
	}

	return result
}

func (c *Brazilian1954ProperCalendar) relativeMatches(entry Brazilian1954ProperEntry, date time.Time) bool {
	switch entry.RelativeTo {
	case "sacred_heart":
		if entry.RelativeOffsetDays == nil {
			return false
		}
		target := CalculateEaster(date.Year()).AddDate(0, 0, 68+*entry.RelativeOffsetDays)
		return target.Year() == date.Year() && target.Month() == date.Month() && target.Day() == date.Day()
	default:
		return false
	}
}

func parseFixedDate(year int, val string) *time.Time {
	parts := strings.Split(val, "-")
	if len(parts) != 2 {
		return nil
	}
	m, err1 := strconv.Atoi(parts[0])
	d, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return nil
	}
	t := time.Date(year, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	return &t
}
