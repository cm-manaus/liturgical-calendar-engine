package engine

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type XML1954Precedence struct {
	Rank         string `xml:"rank,attr"`
	Numeric      string `xml:"numeric,attr"`
	FirstVespers string `xml:"firstVespers,attr"`
	Occurrence   string `xml:"occurrence,attr"`
	Concurrence  string `xml:"concurrence,attr"`
}

type XML1954Octave struct {
	ID        string `xml:"id,attr"`
	Day       string `xml:"day,attr"`
	Status    string `xml:"status,attr"`
	Type      string `xml:"type,attr"`
	EndOffset string `xml:"endOffset,attr"`
}

type XML1954Vigil struct {
	Kind        string `xml:"kind,attr"`
	Privileged  string `xml:"privileged,attr"`
	Anticipated string `xml:"anticipated,attr"`
}

type XML1954Transfer struct {
	Status       string `xml:"status,attr"`
	Transferable string `xml:"transferable,attr"`
	Target       string `xml:"target,attr"`
}

type XML1954Mass struct {
	Color   string `xml:"color,attr"`
	Gloria  string `xml:"gloria,attr"`
	Credo   string `xml:"credo,attr"`
	Preface string `xml:"preface,attr"`
}

type XML1954Readings struct {
	Epistle string `xml:"epistle,attr"`
	Gospel  string `xml:"gospel,attr"`
	Source  string `xml:"source,attr"`
}

type XML1954Suppression struct {
	Mode   string `xml:"mode,attr"`
	Retain string `xml:"retain,attr"`
}

type XML1954Day struct {
	ID              string              `xml:"id,attr"`
	Name            string              `xml:"name,attr"`
	NameResID       string              `xml:"nameResId,attr"`
	NameArg         string              `xml:"nameArg,attr"`
	Date            string              `xml:"date,attr"`
	EasterOffset    string              `xml:"easterOffset,attr"`
	Offset          string              `xml:"offset,attr"`
	Cycle           string              `xml:"cycle,attr"`
	Week            string              `xml:"week,attr"`
	Weekday         string              `xml:"weekday,attr"`
	Window          string              `xml:"window,attr"`
	ObservanceKind  string              `xml:"observanceKind,attr"`
	Pre55Grade      string              `xml:"pre55Grade,attr"`
	Season          string              `xml:"season,attr"`
	Color           string              `xml:"color,attr"`
	Privileged      string              `xml:"privileged,attr"`
	SourceRef       string              `xml:"sourceRef,attr"`
	SourceRank      string              `xml:"sourceRank,attr"`
	IsLordFeast     string              `xml:"isLordFeast,attr"`
	RelativeTo      string              `xml:"relativeTo,attr"`
	OffsetDays      string              `xml:"offsetDays,attr"`
	OctaveEndOffset string              `xml:"octaveEndOffset,attr"`
	Precedence      *XML1954Precedence  `xml:"precedence"`
	Suppression     *XML1954Suppression `xml:"suppression"`
	Octave          *XML1954Octave      `xml:"octave"`
	Vigil           *XML1954Vigil       `xml:"vigil"`
	Transfer        *XML1954Transfer    `xml:"transfer"`
	Mass            *XML1954Mass        `xml:"mass"`
	Readings        *XML1954Readings    `xml:"readings"`
}

type XML1954TransferRule struct {
	ID        string `xml:"id,attr"`
	FeastID   string `xml:"feastId,attr"`
	When      string `xml:"when,attr"`
	Target    string `xml:"target,attr"`
	Resolved  string `xml:"resolved,attr"`
	SourceRef string `xml:"sourceRef,attr"`
}

type XML1954TemporalDoc struct {
	XMLName        xml.Name              `xml:"temporal"`
	EasterCycle    []XML1954Day          `xml:"easter_cycle>day"`
	SundayRules    []XML1954Day          `xml:"sunday_rules>sundayRule"`
	ChristmasCycle []XML1954Day          `xml:"christmas_cycle>day"`
	FeriaRules     []XML1954Day          `xml:"feria_rules>feriaRule"`
	TransferRules  []XML1954TransferRule `xml:"transfer_rules>transferRule"`
}

type XML1954SanctoralDoc struct {
	XMLName xml.Name     `xml:"sanctorale"`
	Feasts  []XML1954Day `xml:"feast"`
}

type CalendarAssetEntry struct {
	ID          string
	Name        string
	NameResID   string
	NameArgs    []any
	Date        string
	Offset      *int
	Cycle       string
	Week        *int
	Weekday     *int
	Window      string
	Pre55Rank   Pre55Rank
	Color       LiturgicalColor
	IsLordFeast bool
	Epistle     string
	Gospel      string
	Metadata    CalendarObservanceMetadata
}

func parseBoolPtr(s string) *bool {
	if s == "true" {
		t := true
		return &t
	}
	if s == "false" {
		f := false
		return &f
	}
	return nil
}

func parseFloatPtr(s string) *float64 {
	if s == "" {
		return nil
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &val
}

func parseIntPtr(s string) *int {
	if s == "" {
		return nil
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &val
}

func inferOrdinalNameArg(name string) string {
	re := regexp.MustCompile(`^(\d+)(?:st|nd|rd|th)\b`)
	matches := re.FindStringSubmatch(strings.TrimSpace(name))
	if len(matches) > 1 {
		return "ord_" + matches[1]
	}
	return ""
}

func parseNameArgs(argStr, name string) []any {
	target := argStr
	if target == "" {
		target = inferOrdinalNameArg(name)
	}
	if target == "" {
		return nil
	}
	num, err := strconv.Atoi(target)
	if err == nil {
		if num >= 1 && num <= 5 {
			return []any{fmt.Sprintf("ord_%d", num)}
		}
		return []any{num}
	}
	return []any{target}
}

func parseIDList(s string) []string {
	if s == "" {
		return nil
	}
	var res []string
	parts := strings.Split(s, ",")
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			res = append(res, trimmed)
		}
	}
	return res
}

func boolPtr(b bool) *bool {
	return &b
}

func xmlDayToMetadata(d XML1954Day) CalendarObservanceMetadata {
	var precedence *float64
	var firstVespers *bool
	var occurrence, concurrence string
	if d.Precedence != nil {
		precedence = parseFloatPtr(d.Precedence.Numeric)
		firstVespers = parseBoolPtr(d.Precedence.FirstVespers)
		occurrence = d.Precedence.Occurrence
		concurrence = d.Precedence.Concurrence
	}

	var suppressionMode string
	var suppressionRetainIDs []string
	if d.Suppression != nil {
		suppressionMode = d.Suppression.Mode
		suppressionRetainIDs = parseIDList(d.Suppression.Retain)
	}

	var octaveID, octaveStatus string
	var octaveDay *int
	var octaveType *Pre55OctaveType
	if d.Octave != nil {
		octaveID = d.Octave.ID
		octaveDay = parseIntPtr(d.Octave.Day)
		octaveStatus = d.Octave.Status
		if octType, ok := Pre55OctaveTypeFromCode(d.Octave.Type); ok {
			octaveType = &octType
		} else if octType, ok := Pre55OctaveTypeFromOctaveID(d.Octave.ID); ok {
			octaveType = &octType
		}
	}

	var vigilKind string
	var vigilPrivileged, vigilAnticipated *bool
	if d.Vigil != nil {
		vigilKind = d.Vigil.Kind
		vigilPrivileged = parseBoolPtr(d.Vigil.Privileged)
		vigilAnticipated = parseBoolPtr(d.Vigil.Anticipated)
	}

	var transferStatus, transferTarget string
	var transferable *bool
	if d.Transfer != nil {
		transferStatus = d.Transfer.Status
		transferable = parseBoolPtr(d.Transfer.Transferable)
		transferTarget = d.Transfer.Target
	}

	var massColor, massPreface string
	var massGloria, massCredo *bool
	if d.Mass != nil {
		massColor = d.Mass.Color
		massGloria = parseBoolPtr(d.Mass.Gloria)
		massCredo = parseBoolPtr(d.Mass.Credo)
		massPreface = d.Mass.Preface
	}

	var readingsSource string
	if d.Readings != nil {
		readingsSource = d.Readings.Source
	}

	return CalendarObservanceMetadata{
		ObservanceKind:       d.ObservanceKind,
		Season:               d.Season,
		Privileged:           parseBoolPtr(d.Privileged),
		SourceRef:            d.SourceRef,
		SourceRank:           parseFloatPtr(d.SourceRank),
		Precedence:           precedence,
		FirstVespers:         firstVespers,
		Occurrence:           occurrence,
		Concurrence:          concurrence,
		SuppressionMode:      suppressionMode,
		SuppressionRetainIDs: suppressionRetainIDs,
		OctaveID:             octaveID,
		OctaveDay:            octaveDay,
		OctaveStatus:         octaveStatus,
		OctaveType:           octaveType,
		VigilKind:            vigilKind,
		VigilPrivileged:      vigilPrivileged,
		VigilAnticipated:     vigilAnticipated,
		TransferStatus:       transferStatus,
		Transferable:         transferable,
		TransferTarget:       transferTarget,
		MassColor:            massColor,
		MassGloria:           massGloria,
		MassCredo:            massCredo,
		MassPreface:          massPreface,
		ReadingsSource:       readingsSource,
	}
}

func xmlToAssetEntry(d XML1954Day) CalendarAssetEntry {
	rank := Pre55RankFromAsset(d.Pre55Grade, d.ObservanceKind, d.Privileged == "true")

	color := ParseColor(d.Color)
	if d.Color == "" {
		color = ColorWhite
	}

	var offset *int
	if d.EasterOffset != "" {
		if val, err := strconv.Atoi(d.EasterOffset); err == nil {
			offset = &val
		}
	} else if d.Offset != "" {
		if val, err := strconv.Atoi(d.Offset); err == nil {
			offset = &val
		}
	}

	var week *int
	if d.Week != "" {
		if val, err := strconv.Atoi(d.Week); err == nil {
			week = &val
		}
	}

	var weekday *int
	if d.Weekday != "" {
		if val, err := strconv.Atoi(d.Weekday); err == nil {
			weekday = &val
		}
	}

	nameArgs := parseNameArgs(d.NameArg, d.Name)

	name := d.Name
	if name == "" {
		name = d.NameResID
	}
	if name == "" {
		name = d.ID
	}

	epistle, gospel := "", ""
	if d.Readings != nil {
		epistle = d.Readings.Epistle
		gospel = d.Readings.Gospel
	}

	metadata := xmlDayToMetadata(d)

	return CalendarAssetEntry{
		ID:          d.ID,
		Name:        name,
		NameResID:   d.NameResID,
		NameArgs:    nameArgs,
		Date:        d.Date,
		Offset:      offset,
		Cycle:       d.Cycle,
		Week:        week,
		Weekday:     weekday,
		Window:      d.Window,
		Pre55Rank:   rank,
		Color:       color,
		IsLordFeast: d.IsLordFeast == "true",
		Epistle:     epistle,
		Gospel:      gospel,
		Metadata:    metadata,
	}
}

func (e CalendarAssetEntry) ToLiturgicalDay() LiturgicalDay {
	rank := e.Pre55Rank
	metadata := e.Metadata
	return LiturgicalDay{
		ID:               e.ID,
		Name:             e.Name,
		NameResID:        e.NameResID,
		NameArgs:         e.NameArgs,
		LiturgicalClass:  ClassIV, // compatibility field
		Color:            e.Color,
		IsLordFeast:      e.IsLordFeast,
		Epistle:          e.Epistle,
		Gospel:           e.Gospel,
		CalendarVersion:  Calendar1954,
		Pre55Rank:        &rank,
		CalendarMetadata: &metadata,
	}
}

type DivinoAfflatu1954Profile struct {
	dataDir             string
	brazilianSanctorale *BrazilianSanctorale
	brazilian1954Proper *Brazilian1954ProperCalendar

	easterCycleEntries map[int][]CalendarAssetEntry
	sundayRules        map[string][]CalendarAssetEntry
	sundayRulesByID    map[string]CalendarAssetEntry
	christmasCycle     map[string]CalendarAssetEntry
	feriaRules         map[string]CalendarAssetEntry
	sanctoralEntries   map[string][]CalendarAssetEntry
	transferRules      []XML1954TransferRule
	loaded             bool
}

func NewDivinoAfflatu1954Profile(dataDir string, brazilian *BrazilianSanctorale) *DivinoAfflatu1954Profile {
	p := &DivinoAfflatu1954Profile{
		dataDir:             dataDir,
		brazilianSanctorale: brazilian,
		easterCycleEntries:  make(map[int][]CalendarAssetEntry),
		sundayRules:         make(map[string][]CalendarAssetEntry),
		sundayRulesByID:     make(map[string]CalendarAssetEntry),
		christmasCycle:      make(map[string]CalendarAssetEntry),
		feriaRules:          make(map[string]CalendarAssetEntry),
		sanctoralEntries:    make(map[string][]CalendarAssetEntry),
	}
	p.load()
	return p
}

func (p *DivinoAfflatu1954Profile) load() {
	temporalPath := filepath.Join(p.dataDir, "temporal.xml")
	sanctoralPath := filepath.Join(p.dataDir, "sanctoral.xml")

	// Parse temporal
	if tempFile, err := os.Open(temporalPath); err == nil {
		defer tempFile.Close()
		byteVal, _ := io.ReadAll(tempFile)
		var doc XML1954TemporalDoc
		if err := xml.Unmarshal(byteVal, &doc); err == nil {
			for _, d := range doc.EasterCycle {
				entry := xmlToAssetEntry(d)
				if entry.Offset != nil {
					p.easterCycleEntries[*entry.Offset] = append(p.easterCycleEntries[*entry.Offset], entry)
				}
			}
			for _, d := range doc.SundayRules {
				entry := xmlToAssetEntry(d)
				p.sundayRulesByID[entry.ID] = entry
				p.sundayRules[entry.Cycle] = append(p.sundayRules[entry.Cycle], entry)
			}
			for _, d := range doc.ChristmasCycle {
				entry := xmlToAssetEntry(d)
				if entry.Date != "" {
					p.christmasCycle[entry.Date] = entry
				}
			}
			for _, d := range doc.FeriaRules {
				entry := xmlToAssetEntry(d)
				p.feriaRules[entry.ID] = entry
			}
			p.transferRules = doc.TransferRules
		}
	}

	// Parse sanctoral
	if sancFile, err := os.Open(sanctoralPath); err == nil {
		defer sancFile.Close()
		byteVal, _ := io.ReadAll(sancFile)
		var doc XML1954SanctoralDoc
		if err := xml.Unmarshal(byteVal, &doc); err == nil {
			for _, d := range doc.Feasts {
				entry := xmlToAssetEntry(d)
				if entry.Date != "" {
					p.sanctoralEntries[entry.Date] = append(p.sanctoralEntries[entry.Date], entry)
				}
			}
		}
	}

	brazilianProperPath := filepath.Join(p.dataDir, "brazilian_sanctoral.xml")
	if _, err := os.Stat(brazilianProperPath); err == nil {
		p.brazilian1954Proper = NewBrazilian1954ProperCalendar(brazilianProperPath)
	}

	p.loaded = true
}

type candidate1954 struct {
	Day         LiturgicalDay
	IsBrazilian bool
	IsTemporal  bool
	IsFeria     bool
	Pre55Rank   Pre55Rank
	Metadata    CalendarObservanceMetadata
	Order       int
}

func (p *DivinoAfflatu1954Profile) Resolve(date time.Time, includeBrazilian bool) LiturgicalResult {
	dateUTC := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	candidates := p.entriesForDate(dateUTC)

	unique := make(map[string]candidate1954)
	for i, c := range candidates {
		day := p.toLiturgicalDay(c.entry, c.isFeria, dateUTC)
		cand := candidate1954{
			Day:         day,
			IsBrazilian: false,
			IsTemporal:  c.isTemporal,
			IsFeria:     c.isFeria,
			Pre55Rank:   c.entry.Pre55Rank,
			Metadata:    c.entry.Metadata,
			Order:       i,
		}
		key := day.ObservanceKey()
		if _, exists := unique[key]; !exists {
			unique[key] = cand
		}
	}

	var localCandidates []candidate1954
	if includeBrazilian {
		if p.brazilian1954Proper != nil && p.brazilian1954Proper.IsLoaded() {
			localEntries := p.brazilian1954Proper.EntriesForDate(dateUTC)
			for i, entry := range localEntries {
				localCandidates = append(localCandidates, candidate1954{
					Day:         entry.ToLiturgicalDay(),
					IsBrazilian: true,
					IsTemporal:  false,
					IsFeria:     false,
					Pre55Rank:   entry.Pre55Rank,
					Metadata:    entry.Metadata,
					Order:       i,
				})
			}
		} else if p.brazilianSanctorale != nil {
			brazilianFeasts := p.brazilianSanctorale.GetAllFeasts(dateUTC)
			for i, bf := range brazilianFeasts {
				rank := RankD
				if bf.LiturgicalClass == ClassI {
					rank = RankD1Cl
				} else if bf.LiturgicalClass == ClassII {
					rank = RankD2Cl
				} else if bf.LiturgicalClass == ClassIII {
					rank = RankD
				} else {
					rank = RankS
				}
				bf.CalendarVersion = Calendar1954
				bf.Pre55Rank = &rank
				localCandidates = append(localCandidates, candidate1954{
					Day:         bf,
					IsBrazilian: true,
					IsTemporal:  false,
					IsFeria:     false,
					Pre55Rank:   rank,
					Order:       i,
				})
			}
		}
	}

	// Merge universal candidates with local candidates by observance identity
	byObservance := make(map[string]candidate1954)
	for _, c := range unique {
		byObservance[c.Day.ObservanceKey()] = c
	}

	order := len(unique)
	for _, sourceCandidate := range localCandidates {
		cand := sourceCandidate
		cand.Order = order
		order++
		key := cand.Day.ObservanceKey()
		existing, exists := byObservance[key]
		if !exists || (cand.IsBrazilian && !existing.IsBrazilian) || p.prefers(cand, existing) {
			byObservance[key] = cand
		}
	}

	var contextEntries []CalendarAssetEntry
	for _, c := range candidates {
		contextEntries = append(contextEntries, c.entry)
	}

	var ordered []candidate1954
	for _, c := range byObservance {
		ordered = append(ordered, p.effectiveOverlayCandidate(c, dateUTC, contextEntries, localCandidates))
	}

	sort.SliceStable(ordered, func(i, j int) bool {
		return p.compareOverlayCandidates(ordered[i], ordered[j]) < 0
	})

	if len(ordered) == 0 {
		feria := LiturgicalDay{
			ID:              "feria",
			Name:            "Feria",
			NameResID:       "feria",
			LiturgicalClass: ClassIV,
			Color:           ColorGreen,
			CalendarVersion: Calendar1954,
		}
		return LiturgicalResult{
			MainDay:         feria,
			Commemorations:  []LiturgicalDay{},
			CalendarVersion: Calendar1954,
		}
	}

	main := ordered[0]
	comms := p.commemorations(main, ordered[1:], dateUTC)

	return LiturgicalResult{
		MainDay:         main.Day,
		Commemorations:  comms,
		CalendarVersion: Calendar1954,
	}
}

type assetCandidatePair struct {
	entry      CalendarAssetEntry
	isTemporal bool
	isFeria    bool
}

func (p *DivinoAfflatu1954Profile) entriesForDate(date time.Time) []assetCandidatePair {
	var result []assetCandidatePair
	temporalEntries := p.temporalEntriesForDate(date)
	for _, entry := range temporalEntries {
		_, isFeria := p.feriaRules[entry.ID]
		result = append(result, assetCandidatePair{
			entry:      entry,
			isTemporal: true,
			isFeria:    isFeria,
		})
	}

	dateKey := fmt.Sprintf("%02d-%02d", date.Month(), date.Day())
	movedIDs := p.transferredSourceIDsForDate(date)

	var sanctoral []CalendarAssetEntry
	if entries, ok := p.sanctoralEntries[dateKey]; ok {
		for _, entry := range entries {
			if movedIDs[entry.ID] {
				continue
			}
			sanctoral = append(sanctoral, entry)
		}
	}

	// Add transferred entries
	targets := p.transferTargetsForYear(date.Year())
	for _, rule := range p.transferRules {
		source := p.transferSource(rule)
		if source == nil {
			continue
		}
		if target, ok := targets[rule.ID]; ok && p.sameDay(target, date) {
			transferred := *source
			transferred.Metadata.TransferStatus = "transferred"
			transferred.Metadata.TransferTarget = dateKey
			t := true
			transferred.Metadata.Transferable = &t
			sanctoral = append(sanctoral, transferred)
		}
	}

	// Anticipate Sunday-dated vigils on Saturday
	if date.Weekday() == time.Saturday {
		following := date.AddDate(0, 0, 1)
		followingKey := fmt.Sprintf("%02d-%02d", following.Month(), following.Day())
		if followingEntries, ok := p.sanctoralEntries[followingKey]; ok {
			for _, entry := range followingEntries {
				if entry.Metadata.ObservanceKind == "vigil" {
					sanctoral = append(sanctoral, markAnticipatedVigil(entry))
				}
			}
		}
	}

	for _, entry := range sanctoral {
		result = append(result, assetCandidatePair{
			entry:      entry,
			isTemporal: false,
			isFeria:    false,
		})
	}

	// Weekday feria evaluation
	if date.Weekday() != time.Sunday && len(temporalEntries) == 0 {
		easter := CalculateEaster(date.Year())
		easterOffset := int(date.Sub(easter).Hours() / 24)
		feria := p.feriaForDate(date, easterOffset)
		includeFeria := feria != nil && (len(sanctoral) == 0 || p.feriaCanBeCommemoratedWithFeast(*feria))
		if includeFeria {
			result = append(result, assetCandidatePair{
				entry:      *feria,
				isTemporal: true,
				isFeria:    true,
			})
		}
	}

	return result
}

func (p *DivinoAfflatu1954Profile) temporalEntriesForDate(date time.Time) []CalendarAssetEntry {
	var result []CalendarAssetEntry
	easter := CalculateEaster(date.Year())
	easterOffset := int(date.Sub(easter).Hours() / 24)

	if entries, ok := p.easterCycleEntries[easterOffset]; ok {
		result = append(result, entries...)
	}

	if date.Weekday() == time.Sunday {
		sunday := p.sundayRuleForDate(date, true)
		if sunday != nil && !containsEntry(result, sunday.ID) {
			result = append(result, *sunday)
		}
		if sunday != nil && sunday.Cycle == "christ_the_king" {
			occurring := p.sundayRuleForDate(date, false)
			if occurring != nil && !containsEntry(result, occurring.ID) {
				result = append(result, *occurring)
			}
		}
	}

	dateKey := fmt.Sprintf("%02d-%02d", date.Month(), date.Day())
	if fixed, ok := p.christmasCycle[dateKey]; ok && !containsEntry(result, fixed.ID) {
		result = append(result, fixed)
	}

	if holyName, ok := p.sundayRulesByID["holy_name_of_jesus"]; ok {
		if date.Month() == 1 && date.Day() == 2 && p.holyNameDate(date.Year()).Day() == 2 && !containsEntry(result, holyName.ID) {
			result = append(result, holyName)
		}
	}

	hasExplicitOctave := false
	if entries, ok := p.sanctoralEntries[dateKey]; ok {
		for _, entry := range entries {
			if entry.Metadata.ObservanceKind == "octave_day" || entry.Metadata.ObservanceKind == "within_octave" {
				hasExplicitOctave = true
				break
			}
		}
	}

	if len(result) == 0 && date.Weekday() != time.Sunday && !hasExplicitOctave {
		feria := p.feriaForDate(date, easterOffset)
		if feria != nil {
			result = append(result, *feria)
		}
	}

	return result
}

func containsEntry(list []CalendarAssetEntry, id string) bool {
	for _, e := range list {
		if e.ID == id {
			return true
		}
	}
	return false
}

func (p *DivinoAfflatu1954Profile) sundayRuleForDate(date time.Time, includeChristTheKing bool) *CalendarAssetEntry {
	if date.Weekday() != time.Sunday {
		return nil
	}

	if includeChristTheKing && p.sameDay(p.lastSundayOfOctober(date.Year()), date) {
		for _, entries := range p.sundayRules {
			for _, entry := range entries {
				if entry.Cycle == "christ_the_king" {
					return &entry
				}
			}
		}
	}

	allocated := p.lateSundayRuleForDate(date)
	if allocated != nil {
		return allocated
	}

	septuagesima := CalculateEaster(date.Year()).AddDate(0, 0, -63)

	for _, entries := range p.sundayRules {
		for _, entry := range entries {
			var expected *time.Time
			switch entry.Cycle {
			case "advent":
				adv := p.advent1(date.Year())
				w := 1
				if entry.Week != nil {
					w = *entry.Week
				}
				exp := adv.AddDate(0, 0, 7*(w-1))
				expected = &exp
			case "holy_name":
				exp := p.holyNameDate(date.Year())
				expected = &exp
			case "christmas_octave":
				first := time.Date(date.Year(), 12, 25, 0, 0, 0, 0, time.UTC)
				exp := p.nextSundayOnOrAfter(first.AddDate(0, 0, 1))
				if !exp.After(first.AddDate(0, 0, 7)) {
					expected = &exp
				}
			case "after_epiphany":
				first := p.nextSundayOnOrAfter(time.Date(date.Year(), 1, 7, 0, 0, 0, 0, time.UTC))
				w := 1
				if entry.Week != nil {
					w = *entry.Week
				}
				exp := first.AddDate(0, 0, 7*(w-1))
				if exp.Before(septuagesima) {
					expected = &exp
				}
			case "after_pentecost":
				if entry.Offset != nil {
					exp := CalculateEaster(date.Year()).AddDate(0, 0, *entry.Offset)
					expected = &exp
				}
			}

			if expected != nil && p.sameDay(*expected, date) {
				return &entry
			}
		}
	}
	return nil
}

func (p *DivinoAfflatu1954Profile) lateSundayRuleForDate(date time.Time) *CalendarAssetEntry {
	easter := CalculateEaster(date.Year())
	trinity := easter.AddDate(0, 0, 56)
	advent := p.advent1(date.Year())

	if date.Before(trinity) || !date.Before(advent) {
		return nil
	}

	week := int(date.Sub(trinity).Hours()/(24*7)) + 1
	if week == 1 {
		return nil
	}

	finalSunday := advent.AddDate(0, 0, -7)
	total := int(finalSunday.Sub(trinity).Hours()/(24*7)) + 1

	var afterRules []CalendarAssetEntry
	for _, entries := range p.sundayRules {
		for _, entry := range entries {
			if entry.Cycle == "after_pentecost" && entry.Window == "after_pentecost" && entry.Week != nil {
				afterRules = append(afterRules, entry)
			}
		}
	}
	sort.Slice(afterRules, func(i, j int) bool {
		return *afterRules[i].Week < *afterRules[j].Week
	})

	if len(afterRules) == 0 {
		return nil
	}

	ruleForWeek := func(val int) *CalendarAssetEntry {
		for _, r := range afterRules {
			if r.Week != nil && *r.Week == val {
				return &r
			}
		}
		return &afterRules[len(afterRules)-1]
	}

	if total <= 24 {
		return ruleForWeek(week)
	}

	if week < total && week >= 24 {
		omitted := p.omittedAfterEpiphanyRules(date.Year())
		insertionSlots := total - 24
		omittedStart := len(omitted) - insertionSlots
		if omittedStart < 0 {
			omittedStart = 0
		}
		insertionIndex := week - 24
		selectedIndex := omittedStart + insertionIndex
		if selectedIndex >= 0 && selectedIndex < len(omitted) {
			return &omitted[selectedIndex]
		}
	}

	if week == total {
		return ruleForWeek(24)
	}
	if week <= 23 {
		return ruleForWeek(week)
	}
	return ruleForWeek(24)
}

func (p *DivinoAfflatu1954Profile) omittedAfterEpiphanyRules(year int) []CalendarAssetEntry {
	septuagesima := CalculateEaster(year).AddDate(0, 0, -63)
	first := p.nextSundayOnOrAfter(time.Date(year, 1, 7, 0, 0, 0, 0, time.UTC))

	var entries []CalendarAssetEntry
	for _, items := range p.sundayRules {
		for _, entry := range items {
			if entry.Cycle == "after_epiphany" && entry.Window == "after_epiphany" && entry.Week != nil && *entry.Week > 1 {
				expected := first.AddDate(0, 0, 7*(*entry.Week-1))
				if !expected.Before(septuagesima) {
					entries = append(entries, entry)
				}
			}
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return *entries[i].Week < *entries[j].Week
	})
	return entries
}

func (p *DivinoAfflatu1954Profile) feriaForDate(date time.Time, easterOffset int) *CalendarAssetEntry {
	for _, entry := range p.feriaRules {
		weekday := entry.Weekday
		window := entry.Window
		if weekday == nil || *weekday != int(date.Weekday()) || window == "" {
			continue
		}
		matches := false
		switch window {
		case "after_advent_iii":
			adv1 := p.advent1(date.Year())
			target := p.nextWeekdayAfter(adv1.AddDate(0, 0, 14), *weekday)
			matches = p.sameDay(date, target)
		case "after_september_14":
			target := p.nextWeekdayAfter(time.Date(date.Year(), 9, 14, 0, 0, 0, 0, time.UTC), *weekday)
			matches = p.sameDay(date, target)
		}
		if matches {
			return &entry
		}
	}

	season := ""
	advent := p.advent1(date.Year())
	septuagesima := CalculateEaster(date.Year()).AddDate(0, 0, -63)
	ashWednesday := CalculateEaster(date.Year()).AddDate(0, 0, -46)
	christmas := time.Date(date.Year(), 12, 25, 0, 0, 0, 0, time.UTC)

	if !date.Before(advent) && date.Before(christmas) {
		season = "advent"
	} else if !date.Before(christmas) || (date.Month() == 1 && date.Day() <= 13) {
		season = "christmas"
	} else if !date.Before(septuagesima) && date.Before(ashWednesday) {
		season = "septuagesima"
	} else if easterOffset >= -45 && easterOffset < 0 {
		season = "lent"
	} else if easterOffset >= 0 && easterOffset < 48 {
		season = "easter"
	} else if easterOffset >= 48 && easterOffset <= 55 {
		season = "pentecost"
	} else if easterOffset >= 56 {
		season = "after_pentecost"
	} else {
		season = "epiphany"
	}

	if f, ok := p.feriaRules[season]; ok {
		return &f
	}
	return nil
}

func (p *DivinoAfflatu1954Profile) commemorations(main candidate1954, candidates []candidate1954, date time.Time) []LiturgicalDay {
	mainRank := main.Pre55Rank
	metadata := main.Metadata

	if (main.IsBrazilian && main.Pre55Rank == "" && main.Day.LiturgicalClass <= ClassII) ||
		metadata.Occurrence == "suppress_lower" {
		return []LiturgicalDay{}
	}

	var result []LiturgicalDay
	seen := make(map[string]bool)

	for _, candidate := range candidates {
		candMeta := candidate.Metadata
		candPrecedence := p.overlayPrecedence(candidate)
		mainPrecedence := p.overlayPrecedence(main)
		candHasHigherPrecedence := candPrecedence < mainPrecedence
		candHasLowerPrecedence := candPrecedence > mainPrecedence

		allowedSundayVigil := main.Metadata.ObservanceKind == "sunday" &&
			candMeta.ObservanceKind == "vigil" &&
			candMeta.VigilKind == "common"
		allowedFeriaMajorCandidate := main.IsFeria &&
			main.Pre55Rank == RankFeriaMajor &&
			(candidate.Pre55Rank == RankS || p.isAnticipatedVigil(candidate))
		allowedPrivilegedFeriaCandidate := main.IsFeria &&
			main.Pre55Rank == RankFeriaPrivilegiata &&
			candidate.Pre55Rank == RankS
		allowedLocalOctaveDay := candidate.IsBrazilian &&
			candMeta.OctaveStatus == "octave_day" &&
			(mainRank == RankD1Cl || main.Metadata.ObservanceKind == "sunday")
		allowedD1Candidate := mainRank == RankD1Cl &&
			(candMeta.ObservanceKind == "sunday" ||
				candMeta.ObservanceKind == "feria" ||
				candMeta.OctaveStatus == "octave_day")

		if candHasHigherPrecedence && !allowedFeriaMajorCandidate && !allowedLocalOctaveDay {
			continue
		}
		if candHasLowerPrecedence &&
			mainRank == RankD1Cl &&
			!allowedSundayVigil &&
			!allowedPrivilegedFeriaCandidate &&
			!allowedD1Candidate &&
			!allowedLocalOctaveDay {
			continue
		}
		if !p.sourceMassOccurrenceAllows(candidate, main, date) {
			continue
		}
		if !p.suppressionAllows(main, candidate) {
			continue
		}
		if candMeta.Occurrence == "suppress_lower" ||
			candMeta.Occurrence == "no_commemoration" ||
			(candMeta.Privileged != nil && *candMeta.Privileged) {
			continue
		}
		if candidate.IsFeria {
			if main.IsFeria ||
				!p.isCommemoratableOccurrence(candMeta.Occurrence) ||
				metadata.Occurrence == "no_commemoration" ||
				!p.concurrenceAllows(main, candidate) {
				continue
			}
		} else if !p.concurrenceAllows(main, candidate) {
			continue
		}

		key := candidate.Day.ObservanceKey()
		if key == main.Day.ObservanceKey() || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, candidate.Day)
	}

	return result
}

func (p *DivinoAfflatu1954Profile) sourceMassOccurrenceAllows(candidate, main candidate1954, date time.Time) bool {
	id := candidate.Day.ID
	if id == "fourth_day_within_octave_patronage_st_joseph" && main.Day.ID == "st_mark_evangelist" {
		return false
	}
	if main.Metadata.ObservanceKind == "sunday" &&
		candidate.Metadata.ObservanceKind == "vigil" &&
		candidate.Day.ID != "vigil_of_st_thomas_apostle" {
		return false
	}
	if id == "st_john_i_pope_and_martyr_2" && main.Metadata.OctaveID == "Pentecost" {
		return false
	}
	if id == "vigil_of_st_bartholomew_apostle" && candidate.Metadata.VigilAnticipated != nil && *candidate.Metadata.VigilAnticipated {
		return false
	}
	if id == "vigil_of_st_matthew_apostle" && (p.isSeptemberEmberDate(date) || p.isSeptemberEmberDate(date.AddDate(0, 0, -1))) {
		return false
	}
	return true
}

func (p *DivinoAfflatu1954Profile) isSeptemberEmberDate(date time.Time) bool {
	easterOffset := int(date.Sub(CalculateEaster(date.Year())).Hours() / 24)
	feria := p.feriaForDate(date, easterOffset)
	return feria != nil && strings.HasSuffix(feria.ID, "_september")
}

func (p *DivinoAfflatu1954Profile) isAnticipatedVigil(candidate candidate1954) bool {
	return candidate.Metadata.ObservanceKind == "vigil" && candidate.Metadata.VigilAnticipated != nil && *candidate.Metadata.VigilAnticipated
}

func (p *DivinoAfflatu1954Profile) feriaCanBeCommemoratedWithFeast(feria CalendarAssetEntry) bool {
	season := feria.Metadata.Season
	return season == "Advent" || season == "Lent"
}

func markAnticipatedVigil(source CalendarAssetEntry) CalendarAssetEntry {
	copyEntry := source
	meta := source.Metadata
	t := true
	meta.VigilAnticipated = &t
	copyEntry.Metadata = meta
	return copyEntry
}

func (p *DivinoAfflatu1954Profile) suppressionAllows(main, candidate candidate1954) bool {
	if main.Metadata.SuppressionMode != "lower_except" {
		return true
	}
	retained := main.Metadata.SuppressionRetainIDs
	if len(retained) == 0 {
		return candidate.Metadata.ObservanceKind != "feria"
	}
	for _, id := range retained {
		if id == candidate.Day.ID || (candidate.Day.NameResID != "" && id == candidate.Day.NameResID) {
			return true
		}
	}
	return false
}

func (p *DivinoAfflatu1954Profile) concurrenceAllows(main, candidate candidate1954) bool {
	rel := strings.ToLower(candidate.Metadata.Concurrence)
	if rel == "" || rel == "none" {
		return p.isCommemoratableOccurrence(candidate.Metadata.Occurrence)
	}
	if rel == "first_vespers" {
		hasFV := candidate.Metadata.FirstVespers != nil && *candidate.Metadata.FirstVespers
		mainFV := main.Metadata.FirstVespers != nil && *main.Metadata.FirstVespers
		return hasFV && (mainFV || main.Metadata.Occurrence != "suppress_lower")
	}
	return rel == "commemorate" || rel == "same_day"
}

func (p *DivinoAfflatu1954Profile) isCommemoratableOccurrence(occ string) bool {
	return occ == "commemorate" || occ == "commemorate_or_transfer"
}

func (p *DivinoAfflatu1954Profile) prefers(candidate, existing candidate1954) bool {
	return p.compareBrazilianCandidates(candidate, existing) < 0
}

func (p *DivinoAfflatu1954Profile) compareBrazilianCandidates(a, b candidate1954) int {
	precA := p.brazilianPrecedence(a)
	precB := p.brazilianPrecedence(b)
	if precA != precB {
		if precA < precB {
			return -1
		}
		return 1
	}
	if a.IsBrazilian != b.IsBrazilian {
		if a.IsBrazilian && a.Metadata.ObservanceKind == "within_octave" {
			return 1
		}
		if b.IsBrazilian && b.Metadata.ObservanceKind == "within_octave" {
			return -1
		}
		if a.IsBrazilian {
			return -1
		}
		return 1
	}
	if a.Order < b.Order {
		return -1
	} else if a.Order > b.Order {
		return 1
	}
	return 0
}

func (p *DivinoAfflatu1954Profile) brazilianPrecedence(c candidate1954) int {
	rank := c.Pre55Rank
	if rank == "" {
		return int(c.Day.LiturgicalClass)
	}
	meta := c.Metadata
	var declPrec *float64
	if meta.ObservanceKind != "feria" && !c.IsFeria {
		if meta.Precedence != nil {
			declPrec = meta.Precedence
		} else {
			declPrec = meta.SourceRank
		}
	}
	ordinarySunday := c.IsTemporal && meta.ObservanceKind == "sunday"
	normalizeGrade := meta.ObservanceKind != "feria"
	if meta.OctaveStatus != "" && meta.OctaveStatus != "none" {
		normalizeGrade = false
	}
	val := EffectivePrecedence(rank, declPrec, ordinarySunday, normalizeGrade)
	return PrecedenceIndexFor(val)
}

func (p *DivinoAfflatu1954Profile) compareOverlayCandidates(a, b candidate1954) int {
	if (a.IsBrazilian || b.IsBrazilian) && (p.brazilianSanctorale != nil || (p.brazilian1954Proper != nil && p.brazilian1954Proper.IsLoaded())) {
		boundary := p.compareBrazilianCandidates(a, b)
		if boundary != 0 {
			return boundary
		}
	}

	if a.IsFeria && a.Pre55Rank == RankFeriaMajor && b.Pre55Rank == RankS {
		return -1
	}
	if b.IsFeria && b.Pre55Rank == RankFeriaMajor && a.Pre55Rank == RankS {
		return 1
	}

	if p.isAnticipatedVigil(a) && b.IsFeria && b.Pre55Rank == RankFeriaMajor {
		return 1
	}
	if p.isAnticipatedVigil(b) && a.IsFeria && a.Pre55Rank == RankFeriaMajor {
		return -1
	}

	pA := p.overlayPrecedence(a)
	pB := p.overlayPrecedence(b)
	if pA != pB {
		if pA < pB {
			return -1
		}
		return 1
	}

	isOctaveA := a.Metadata.OctaveStatus == "day_within" || a.Metadata.OctaveStatus == "octave_day"
	isOctaveB := b.Metadata.OctaveStatus == "day_within" || b.Metadata.OctaveStatus == "octave_day"
	if isOctaveA != isOctaveB {
		if isOctaveA {
			return 1
		}
		return -1
	}

	if a.Day.IsLordFeast != b.Day.IsLordFeast {
		if a.Day.IsLordFeast {
			return -1
		}
		return 1
	}

	sA := 0.0
	if a.Metadata.SourceRank != nil {
		sA = *a.Metadata.SourceRank
	} else if a.Pre55Rank != "" {
		sA = a.Pre55Rank.Info().Rank
	}

	sB := 0.0
	if b.Metadata.SourceRank != nil {
		sB = *b.Metadata.SourceRank
	} else if b.Pre55Rank != "" {
		sB = b.Pre55Rank.Info().Rank
	}

	if sA != sB {
		if sB > sA {
			return 1
		}
		return -1
	}

	if a.Order < b.Order {
		return -1
	} else if a.Order > b.Order {
		return 1
	}
	return 0
}

func (p *DivinoAfflatu1954Profile) overlayPrecedence(candidate candidate1954) int {
	rank := candidate.Pre55Rank
	if rank == "" {
		return int(candidate.Day.LiturgicalClass)
	}
	return PrecedenceIndexFor(p.effectivePrecedence(candidate))
}

func (p *DivinoAfflatu1954Profile) effectivePrecedence(candidate candidate1954) float64 {
	rank := candidate.Pre55Rank
	if rank == "" {
		return 0
	}
	metadata := candidate.Metadata
	var declaredPrecedence *float64
	if metadata.ObservanceKind != "feria" && !candidate.IsFeria {
		if metadata.Precedence != nil {
			declaredPrecedence = metadata.Precedence
		} else {
			declaredPrecedence = metadata.SourceRank
		}
	}
	ordinarySunday := candidate.IsTemporal && metadata.ObservanceKind == "sunday"
	normalizeGrade := metadata.ObservanceKind != "feria"
	if metadata.OctaveStatus != "" && metadata.OctaveStatus != "none" {
		normalizeGrade = false
	}
	return EffectivePrecedence(rank, declaredPrecedence, ordinarySunday, normalizeGrade)
}

func (p *DivinoAfflatu1954Profile) effectiveOverlayCandidate(
	candidate candidate1954,
	date time.Time,
	contextEntries []CalendarAssetEntry,
	localCandidates []candidate1954,
) candidate1954 {
	metadata := candidate.Metadata
	if metadata.ObservanceKind == "vigil" {
		return candidate
	}

	isFeria := candidate.IsFeria || metadata.ObservanceKind == "feria"
	massGloria := metadata.MassGloria
	massCredo := metadata.MassCredo
	massPreface := metadata.MassPreface
	massColor := metadata.MassColor
	if !isFeria && candidate.Pre55Rank == RankS {
		t := true
		massGloria = &t
	}

	octavePreface := p.activeOverlayOctavePreface(localCandidates)
	if octavePreface == "" {
		octavePreface = p.activeOctavePreface(contextEntries)
	}
	inCredoOctave := p.hasOverlayCredoOctave(localCandidates) || p.hasCredoOctave(contextEntries)

	if isFeria {
		massPreface = p.seasonalPreface(date)
	} else if (p.isDefaultPreface(massPreface) || (octavePreface != "" && massPreface == "Trinity")) &&
		(metadata.ObservanceKind != "sunday" || metadata.OctaveStatus == "day_within" || metadata.OctaveStatus == "octave_day") {
		if octavePreface != "" {
			massPreface = octavePreface
		} else {
			massPreface = p.seasonalPreface(date)
		}
	}

	if !isFeria && massColor != "BLACK" && inCredoOctave {
		t := true
		massCredo = &t
	}

	metadata.MassColor = massColor
	metadata.MassGloria = massGloria
	metadata.MassCredo = massCredo
	metadata.MassPreface = massPreface

	candidate.Metadata = metadata
	candidate.Day.CalendarMetadata = &metadata
	return candidate
}

func (p *DivinoAfflatu1954Profile) hasOverlayCredoOctave(candidates []candidate1954) bool {
	for _, c := range candidates {
		meta := c.Metadata
		status := meta.OctaveStatus
		if (status == "day_within" || status == "octave_day") &&
			meta.OctaveID != "" &&
			meta.OctaveID != "none" &&
			meta.Occurrence != "no_commemoration" &&
			meta.MassCredo != nil && *meta.MassCredo {
			return true
		}
	}
	return false
}

func (p *DivinoAfflatu1954Profile) activeOverlayOctavePreface(candidates []candidate1954) string {
	for _, c := range candidates {
		meta := c.Metadata
		status := meta.OctaveStatus
		octaveID := meta.OctaveID
		preface := meta.MassPreface
		if (status == "day_within" || status == "octave_day") &&
			octaveID != "" &&
			octaveID != "none" &&
			meta.Occurrence != "no_commemoration" &&
			preface != "" &&
			preface != "Common" &&
			preface != "None" &&
			preface != "Trinity" {
			return preface
		}
	}
	return ""
}

func (p *DivinoAfflatu1954Profile) hasCredoOctave(entries []CalendarAssetEntry) bool {
	for _, e := range entries {
		meta := e.Metadata
		status := meta.OctaveStatus
		if (status == "day_within" || status == "octave_day") &&
			meta.OctaveID != "" &&
			meta.OctaveID != "none" &&
			meta.Occurrence != "no_commemoration" &&
			meta.MassCredo != nil && *meta.MassCredo {
			return true
		}
	}
	return false
}

func (p *DivinoAfflatu1954Profile) activeOctavePreface(entries []CalendarAssetEntry) string {
	for _, e := range entries {
		meta := e.Metadata
		status := meta.OctaveStatus
		octaveID := meta.OctaveID
		preface := meta.MassPreface
		if (status == "day_within" || status == "octave_day") &&
			octaveID != "" &&
			octaveID != "none" &&
			meta.Occurrence != "no_commemoration" &&
			preface != "" &&
			preface != "Common" &&
			preface != "None" &&
			preface != "Trinity" {
			return preface
		}
	}
	return ""
}

func (p *DivinoAfflatu1954Profile) isDefaultPreface(val string) bool {
	return val == "" || val == "Common"
}

func (p *DivinoAfflatu1954Profile) seasonalPreface(date time.Time) string {
	if (date.Month() == 12 && date.Day() >= 25) || (date.Month() == 1 && date.Day() <= 13) {
		return "Christmas"
	}
	offset := int(date.Sub(CalculateEaster(date.Year())).Hours() / 24)
	if offset >= -14 && offset <= -2 {
		return "HolyCross"
	}
	if offset >= -46 && offset < -14 {
		return "Lent"
	}
	if offset >= 0 && offset < 39 {
		return "Easter"
	}
	if offset >= 39 && offset < 48 {
		return "Ascension"
	}
	if offset >= 48 && offset <= 55 {
		return "Pentecost"
	}
	if date.Weekday() == time.Sunday && (offset >= 56 || !date.Before(p.advent1(date.Year()))) {
		return "Trinity"
	}
	return "Common"
}

func (p *DivinoAfflatu1954Profile) toLiturgicalDay(entry CalendarAssetEntry, isFeria bool, feriaDate time.Time) LiturgicalDay {
	if !isFeria {
		return entry.ToLiturgicalDay()
	}
	dateKey := fmt.Sprintf("%02d-%02d", feriaDate.Month(), feriaDate.Day())
	rank := entry.Pre55Rank
	metadata := entry.Metadata
	return LiturgicalDay{
		ID:               fmt.Sprintf("feria_%s_%s", entry.ID, dateKey),
		Name:             "Feria",
		NameResID:        "feria",
		LiturgicalClass:  ClassIV,
		Color:            entry.Color,
		Epistle:          entry.Epistle,
		Gospel:           entry.Gospel,
		CalendarVersion:  Calendar1954,
		Pre55Rank:        &rank,
		CalendarMetadata: &metadata,
	}
}

func (p *DivinoAfflatu1954Profile) transferSource(rule XML1954TransferRule) *CalendarAssetEntry {
	for _, entries := range p.sanctoralEntries {
		for _, e := range entries {
			if e.ID == rule.FeastID || e.NameResID == rule.FeastID {
				return &e
			}
		}
	}
	for _, e := range p.christmasCycle {
		if e.ID == rule.FeastID || e.NameResID == rule.FeastID {
			return &e
		}
	}
	for _, entries := range p.easterCycleEntries {
		for _, e := range entries {
			if e.ID == rule.FeastID || e.NameResID == rule.FeastID {
				return &e
			}
		}
	}
	return nil
}

func (p *DivinoAfflatu1954Profile) transferredSourceIDsForDate(date time.Time) map[string]bool {
	result := make(map[string]bool)
	for _, rule := range p.transferRules {
		source := p.transferSource(rule)
		if source == nil || source.Date == "" {
			continue
		}
		sourceDate := p.fixedDate(date.Year(), source.Date)
		if sourceDate == nil || !p.transferApplies(rule, *sourceDate) {
			continue
		}
		if _, ok := p.transferTargetsForYear(date.Year())[rule.ID]; ok {
			result[source.ID] = true
		}
	}
	return result
}

func (p *DivinoAfflatu1954Profile) transferTargetsForYear(year int) map[string]time.Time {
	result := make(map[string]time.Time)
	reserved := make(map[string]bool)

	for _, rule := range p.transferRules {
		source := p.transferSource(rule)
		if source == nil || source.Date == "" {
			continue
		}
		sourceDate := p.fixedDate(year, source.Date)
		if sourceDate == nil || !p.transferApplies(rule, *sourceDate) {
			continue
		}

		var target time.Time
		switch rule.Target {
		case "first_day_after_easter_octave", "next_free_day_after_passiontide":
			target = CalculateEaster(year).AddDate(0, 0, 8)
		case "next_free_day_after_occurrence":
			target = sourceDate.AddDate(0, 0, 1)
		default:
			continue
		}

		var selected *time.Time
		for i := 0; i < 366; i++ {
			if p.isFreeTransferDate(target, *source, reserved) {
				selected = &target
				break
			}
			target = target.AddDate(0, 0, 1)
		}
		if selected != nil {
			result[rule.ID] = *selected
			dateKey := fmt.Sprintf("%02d-%02d", selected.Month(), selected.Day())
			reserved[dateKey] = true
		}
	}
	return result
}

func (p *DivinoAfflatu1954Profile) isFreeTransferDate(date time.Time, source CalendarAssetEntry, reserved map[string]bool) bool {
	key := fmt.Sprintf("%02d-%02d", date.Month(), date.Day())
	if reserved[key] || date.Weekday() == time.Sunday {
		return false
	}
	var occupied []CalendarAssetEntry
	occupied = append(occupied, p.temporalEntriesForDate(date)...)
	if entries, ok := p.sanctoralEntries[key]; ok {
		occupied = append(occupied, entries...)
	}
	for _, entry := range occupied {
		if entry.Metadata.ObservanceKind == "feria" {
			continue
		}
		priv := entry.Metadata.Privileged != nil && *entry.Metadata.Privileged
		if priv || entry.Metadata.Occurrence == "suppress_lower" || entry.Pre55Rank.PrecedenceIndex() <= source.Pre55Rank.PrecedenceIndex() {
			return false
		}
	}
	return true
}

func (p *DivinoAfflatu1954Profile) transferApplies(rule XML1954TransferRule, date time.Time) bool {
	easter := CalculateEaster(date.Year())
	offset := int(date.Sub(easter).Hours() / 24)

	switch rule.When {
	case "occurs_in_holy_week":
		return offset >= -7 && offset <= -1
	case "occurs_with_greater_sunday":
		temporal := p.temporalEntriesForDate(date)
		if date.Weekday() != time.Sunday {
			return false
		}
		for _, entry := range temporal {
			if entry.Pre55Rank.PrecedenceIndex() <= RankDMaj.PrecedenceIndex() {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (p *DivinoAfflatu1954Profile) lastSundayOfOctober(year int) time.Time {
	date := time.Date(year, 10, 31, 0, 0, 0, 0, time.UTC)
	for date.Weekday() != time.Sunday {
		date = date.AddDate(0, 0, -1)
	}
	return date
}

func (p *DivinoAfflatu1954Profile) advent1(year int) time.Time {
	date := time.Date(year, 11, 27, 0, 0, 0, 0, time.UTC)
	for date.Weekday() != time.Sunday {
		date = date.AddDate(0, 0, 1)
	}
	return date
}

func (p *DivinoAfflatu1954Profile) nextSundayOnOrAfter(val time.Time) time.Time {
	date := val
	for date.Weekday() != time.Sunday {
		date = date.AddDate(0, 0, 1)
	}
	return date
}

func (p *DivinoAfflatu1954Profile) nextWeekdayAfter(val time.Time, weekday int) time.Time {
	date := val.AddDate(0, 0, 1)
	for int(date.Weekday()) != weekday {
		date = date.AddDate(0, 0, 1)
	}
	return date
}

func (p *DivinoAfflatu1954Profile) holyNameDate(year int) time.Time {
	for day := 2; day <= 5; day++ {
		candidate := time.Date(year, 1, day, 0, 0, 0, 0, time.UTC)
		if candidate.Weekday() == time.Sunday {
			return candidate
		}
	}
	return time.Date(year, 1, 2, 0, 0, 0, 0, time.UTC)
}

func (p *DivinoAfflatu1954Profile) fixedDate(year int, val string) *time.Time {
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

func (p *DivinoAfflatu1954Profile) sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
