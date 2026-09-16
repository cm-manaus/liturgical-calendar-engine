package engine

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type LiturgicalClass int

const (
	ClassI   LiturgicalClass = 1
	ClassII  LiturgicalClass = 2
	ClassIII LiturgicalClass = 3
	ClassIV  LiturgicalClass = 4
)

func (c LiturgicalClass) Precedes(other LiturgicalClass) bool {
	return c < other
}

func (c LiturgicalClass) NameResID() string {
	return "class_" + strconv.Itoa(int(c))
}

func (c LiturgicalClass) String() string {
	switch c {
	case ClassI:
		return "I"
	case ClassII:
		return "II"
	case ClassIII:
		return "III"
	case ClassIV:
		return "IV"
	}
	return "IV"
}

func ParseClass(s string) LiturgicalClass {
	switch strings.TrimSpace(strings.ToUpper(s)) {
	case "I", "1":
		return ClassI
	case "II", "2":
		return ClassII
	case "III", "3":
		return ClassIII
	case "IV", "4":
		return ClassIV
	default:
		return ClassIII
	}
}

type LiturgicalColor string

const (
	ColorWhite  LiturgicalColor = "WHITE"
	ColorRed    LiturgicalColor = "RED"
	ColorGreen  LiturgicalColor = "GREEN"
	ColorViolet LiturgicalColor = "VIOLET"
	ColorBlack  LiturgicalColor = "BLACK"
	ColorRose   LiturgicalColor = "ROSE"
	ColorBlue   LiturgicalColor = "BLUE"
)

func ParseColor(s string) LiturgicalColor {
	switch strings.TrimSpace(strings.ToUpper(s)) {
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
	case "BLUE":
		return ColorBlue
	default:
		return ColorWhite
	}
}

// CalendarObservanceMetadata contains optional rubrical and source metadata from XML.
type CalendarObservanceMetadata struct {
	ObservanceKind       string           `json:"observance_kind,omitempty"`
	Season               string           `json:"season,omitempty"`
	Privileged           *bool            `json:"privileged,omitempty"`
	SourceRef            string           `json:"source_ref,omitempty"`
	SourceRank           *float64         `json:"source_rank,omitempty"`
	Precedence           *float64         `json:"precedence,omitempty"`
	FirstVespers         *bool            `json:"first_vespers,omitempty"`
	Occurrence           string           `json:"occurrence,omitempty"`
	Concurrence          string           `json:"concurrence,omitempty"`
	SuppressionMode      string           `json:"suppression_mode,omitempty"`
	SuppressionRetainIDs []string         `json:"suppression_retain_ids,omitempty"`
	OctaveID             string           `json:"octave_id,omitempty"`
	OctaveDay            *int             `json:"octave_day,omitempty"`
	OctaveStatus         string           `json:"octave_status,omitempty"`
	OctaveType           *Pre55OctaveType `json:"octave_type,omitempty"`
	VigilKind            string           `json:"vigil_kind,omitempty"`
	VigilPrivileged      *bool            `json:"vigil_privileged,omitempty"`
	VigilAnticipated     *bool            `json:"vigil_anticipated,omitempty"`
	TransferStatus       string           `json:"transfer_status,omitempty"`
	Transferable         *bool            `json:"transferable,omitempty"`
	TransferTarget       string           `json:"transfer_target,omitempty"`
	MassColor            string           `json:"mass_color,omitempty"`
	MassGloria           *bool            `json:"mass_gloria,omitempty"`
	MassCredo            *bool            `json:"mass_credo,omitempty"`
	MassPreface          string           `json:"mass_preface,omitempty"`
	ReadingsSource       string           `json:"readings_source,omitempty"`
}

type LiturgyInfo struct {
	Gloria  string `json:"gloria"`
	Credo   string `json:"credo"`
	Preface string `json:"preface"`
	Epistle string `json:"epistle"`
	Gospel  string `json:"gospel"`
}

type LiturgicalDay struct {
	ID               string
	Name             string
	NameResID        string
	NameArgs         []any
	LiturgicalClass  LiturgicalClass
	Color            LiturgicalColor
	IsLordFeast      bool
	Epistle          string
	Gospel           string
	CalendarVersion  CalendarVersion
	Pre55Rank        *Pre55Rank
	CalendarMetadata *CalendarObservanceMetadata
}

func (d LiturgicalDay) IsPre55() bool {
	return d.CalendarVersion == Calendar1954
}

func (d LiturgicalDay) ObservanceKey() string {
	if d.NameResID != "" {
		argsStr := ""
		if len(d.NameArgs) > 0 {
			var args []string
			for _, arg := range d.NameArgs {
				args = append(args, fmt.Sprintf("%v", arg))
			}
			argsStr = strings.Join(args, ",")
		}
		return fmt.Sprintf("res:%s:%s", d.NameResID, argsStr)
	}
	return "name:" + strings.ToLower(strings.TrimSpace(d.Name))
}

func (d LiturgicalDay) isJosephFeast() bool {
	return d.NameResID == "st_joseph" ||
		d.NameResID == "st_joseph_worker" ||
		d.NameResID == "st_joseph_calasanctius" ||
		d.NameResID == "st_joseph_cupertino"
}

func (d LiturgicalDay) isMarianFeast(localizedName string) bool {
	lowerName := strings.ToLower(localizedName)
	saintNamePrefixes := []string{
		"st. ",
		"st ",
		"sts. ",
		"sts ",
		"s. ",
		"ss. ",
		"saint ",
		"saints ",
		"san ",
		"santa ",
		"santo ",
		"sainte ",
		"ste. ",
		"sta. ",
		"hl. ",
	}
	for _, prefix := range saintNamePrefixes {
		if strings.HasPrefix(lowerName, prefix) {
			return false
		}
	}

	isVirginNamedMaria := strings.Contains(lowerName, "maria") &&
		(strings.Contains(lowerName, "virgem") || strings.Contains(lowerName, "virgin")) &&
		!strings.Contains(lowerName, "nossa senhora") &&
		!strings.Contains(lowerName, "our lady") &&
		!strings.Contains(lowerName, "bem-aventurada virgem maria") &&
		!strings.Contains(lowerName, "b.m.v") &&
		!strings.Contains(lowerName, "b.v.m") &&
		!strings.Contains(lowerName, "bvm")

	if strings.Contains(lowerName, "joseph") ||
		strings.Contains(lowerName, "jose") ||
		strings.Contains(lowerName, "josé") ||
		strings.Contains(lowerName, "confessor") ||
		strings.Contains(lowerName, "vigil") ||
		strings.Contains(lowerName, "vigília") ||
		strings.Contains(lowerName, "véspera") ||
		strings.Contains(lowerName, "magdala") ||
		strings.Contains(lowerName, "madalena") ||
		strings.Contains(lowerName, "magdalene") ||
		isVirginNamedMaria {
		return false
	}

	return strings.Contains(lowerName, "our lady") ||
		strings.Contains(lowerName, "bem-aventurada virgem maria") ||
		strings.Contains(lowerName, "nossa senhora") ||
		strings.Contains(lowerName, "b.m.v.") ||
		strings.Contains(lowerName, "b.v.m.") ||
		strings.Contains(lowerName, "bvm") ||
		strings.Contains(lowerName, "holy name of mary") ||
		strings.Contains(lowerName, "name of mary") ||
		strings.Contains(lowerName, "nome de maria") ||
		strings.Contains(lowerName, "nombre de mar") ||
		strings.Contains(lowerName, "nom de marie") ||
		strings.Contains(lowerName, "mariä namen") ||
		strings.Contains(lowerName, "immaculate") ||
		strings.Contains(lowerName, "assumption") ||
		strings.Contains(lowerName, "annunciation") ||
		strings.Contains(lowerName, "maria") ||
		strings.Contains(lowerName, "marian")
}

func (d LiturgicalDay) displayColor(baseColor, localizedName string) string {
	if d.isJosephFeast() {
		return string(ColorWhite)
	}
	if d.isMarianFeast(localizedName) || (d.IsPre55() && d.isMarianFeast(d.Name)) {
		return string(ColorBlue)
	}
	return baseColor
}

type LiturgicalDayJSON struct {
	Name            string       `json:"name"`
	ID              string       `json:"id,omitempty"`
	NameResID       string       `json:"name_res_id,omitempty"`
	ObservanceKey   string       `json:"observance_key,omitempty"`
	CalendarVersion string       `json:"calendar_version,omitempty"`
	CalendarName    string       `json:"calendar_name,omitempty"`
	ClassCode       string       `json:"class_code,omitempty"`
	ClassName       string       `json:"class_name,omitempty"`
	Pre55Grade      string       `json:"pre55_grade,omitempty"`
	RankCode        string       `json:"rank_code,omitempty"`
	RankName        string       `json:"rank_name,omitempty"`
	OctaveTypeName  string       `json:"octave_type_name,omitempty"`
	Color                 string       `json:"color"`
	IsLordFeast           bool         `json:"is_lord_feast"`
	HasAbstinence         bool         `json:"has_abstinence"`
	IsAbstinenceDispensed bool         `json:"is_abstinence_dispensed"`
	Date                  string       `json:"date,omitempty"`
	Liturgy               *LiturgyInfo `json:"liturgy,omitempty"`

	// Optional 1954 metadata exposed when present
	ObservanceKind  string   `json:"observance_kind,omitempty"`
	Season          string   `json:"season,omitempty"`
	Privileged      *bool    `json:"privileged,omitempty"`
	Precedence      *float64 `json:"precedence,omitempty"`
	FirstVespers    *bool    `json:"first_vespers,omitempty"`
	Occurrence      string   `json:"occurrence,omitempty"`
	Concurrence     string   `json:"concurrence,omitempty"`
	SuppressionMode string   `json:"suppression_mode,omitempty"`
	OctaveID        string   `json:"octave_id,omitempty"`
	OctaveDay       *int     `json:"octave_day,omitempty"`
	OctaveStatus    string   `json:"octave_status,omitempty"`
	TransferStatus  string   `json:"transfer_status,omitempty"`
	TransferTarget  string   `json:"transfer_target,omitempty"`
}

// CalculateAbstinenceInfo determines whether abstinence from meat applies to the given liturgical day,
// and whether a Friday observance has a canonical dispensation (e.g. Feast of I Class).
// Under Catholic canonical rules (CIC 1917 can. 1252 § 1, 1962 Code of Rubrics, Pre-55):
// 1. Every Friday throughout the year is a day of abstinence from meat, EXCEPT when a Feast of the I Class
//    (or Double of I Class in Pre-55) falls on that Friday (e.g. Christmas, Sacred Heart, Annunciation, etc.).
// 2. Good Friday is always a day of fast and abstinence.
// 3. Ash Wednesday is always a day of fast and abstinence.
// 4. Thursdays and Saturdays after Ash Wednesday are NOT abstinence days.
func CalculateAbstinenceInfo(date time.Time, day LiturgicalDay, finalName string) (hasAbstinence bool, isDispensed bool) {
	if date.IsZero() {
		return false, false
	}

	nameLower := strings.ToLower(day.Name)
	finalLower := strings.ToLower(finalName)

	if date.Weekday() == time.Friday {
		// Good Friday always requires abstinence (never dispensed)
		if strings.Contains(nameLower, "parasceve") ||
			strings.Contains(nameLower, "paixão") ||
			strings.Contains(nameLower, "good friday") ||
			strings.Contains(finalLower, "parasceve") ||
			strings.Contains(finalLower, "paixão") ||
			strings.Contains(finalLower, "good friday") ||
			day.ID == "passion_friday" ||
			day.ID == "good_friday" {
			return true, false
		}

		// In 1962: Class I feast ceases abstinence.
		if day.LiturgicalClass == ClassI {
			return false, true
		}

		// In 1954 / Pre-55: Rank D1Cl / I Class ceases abstinence.
		if day.Pre55Rank != nil && *day.Pre55Rank == RankD1Cl {
			return false, true
		}

		if day.CalendarMetadata != nil && day.CalendarMetadata.Precedence != nil && *day.CalendarMetadata.Precedence >= 6.0 {
			return false, true
		}

		return true, false
	}

	// Days after Ash Wednesday (explicitly not abstinence)
	if strings.Contains(nameLower, "quinta-feira depois das cinzas") ||
		strings.Contains(nameLower, "sábado depois das cinzas") ||
		strings.Contains(nameLower, "thursday after ash") ||
		strings.Contains(nameLower, "saturday after ash") ||
		strings.Contains(finalLower, "quinta-feira depois das cinzas") ||
		strings.Contains(finalLower, "sábado depois das cinzas") {
		return false, false
	}

	// Ash Wednesday
	if strings.Contains(nameLower, "cinzas") ||
		strings.Contains(nameLower, "ash wednesday") ||
		strings.Contains(nameLower, "cinerum") ||
		strings.Contains(finalLower, "cinzas") ||
		strings.Contains(finalLower, "ash wednesday") ||
		day.ID == "ash_wednesday" ||
		day.NameResID == "ash_wednesday" {
		return true, false
	}

	return false, false
}

// CalculateAbstinence determines whether abstinence from meat applies to the given liturgical day.
func CalculateAbstinence(date time.Time, day LiturgicalDay, finalName string) bool {
	hasAbstinence, _ := CalculateAbstinenceInfo(date, day, finalName)
	return hasAbstinence
}

func (d LiturgicalDay) ToJSON(translations map[string]string, targetDate time.Time, isCommemoration bool) LiturgicalDayJSON {
	finalName := d.Name
	if d.NameResID != "" {
		var resolvedArgs []string
		for _, arg := range d.NameArgs {
			switch v := arg.(type) {
			case int:
				ordGeneric := "%1$sth"
				if val, ok := translations["ord_generic"]; ok {
					ordGeneric = val
				}
				resolvedArgs = append(resolvedArgs, formatAndroidString(ordGeneric, strconv.Itoa(v)))
			case string:
				if translated, ok := translations[v]; ok {
					resolvedArgs = append(resolvedArgs, translated)
				} else if _, err := strconv.Atoi(v); err == nil {
					ordGeneric := "%1$sth"
					if val, ok := translations["ord_generic"]; ok {
						ordGeneric = val
					}
					resolvedArgs = append(resolvedArgs, formatAndroidString(ordGeneric, v))
				} else {
					resolvedArgs = append(resolvedArgs, v)
				}
			default:
				resolvedArgs = append(resolvedArgs, fmt.Sprintf("%v", v))
			}
		}

		rawTemplate := d.Name
		if val, ok := translations[d.NameResID]; ok {
			rawTemplate = val
		}
		finalName = formatAndroidString(rawTemplate, resolvedArgs...)
	}

	dateStr := ""
	if !targetDate.IsZero() {
		dateStr = fmt.Sprintf("%02d-%02d", targetDate.Month(), targetDate.Day())
	}

	if d.IsPre55() {
		return d.toPre55JSON(translations, finalName, dateStr, targetDate, isCommemoration)
	}

	// 1962 Resolution JSON
	classResID := d.LiturgicalClass.NameResID()
	className := fmt.Sprintf("%s Class", d.LiturgicalClass.String())
	if val, ok := translations[classResID]; ok {
		className = val
	}

	finalColorName := d.displayColor(string(d.Color), finalName)

	liturgy := build1962LiturgyInfo(
		finalName,
		finalColorName,
		d.LiturgicalClass,
		d.IsLordFeast,
		translations,
		targetDate,
		isCommemoration,
		d.Epistle,
		d.Gospel,
		d.Name,
	)

	hasAbstinence, isDispensed := CalculateAbstinenceInfo(targetDate, d, finalName)

	return LiturgicalDayJSON{
		Name:                  finalName,
		ID:                    d.ID,
		NameResID:             d.NameResID,
		ObservanceKey:         d.ObservanceKey(),
		CalendarVersion:       d.CalendarVersion.AssetID(),
		CalendarName:          d.CalendarVersion.DisplayName(),
		ClassCode:             d.LiturgicalClass.String(),
		ClassName:             className,
		Color:                 finalColorName,
		IsLordFeast:           d.IsLordFeast,
		HasAbstinence:         hasAbstinence,
		IsAbstinenceDispensed: isDispensed,
		Date:                  dateStr,
		Liturgy:               liturgy,
	}
}

func (d LiturgicalDay) toPre55JSON(translations map[string]string, finalName, dateStr string, targetDate time.Time, isCommemoration bool) LiturgicalDayJSON {
	rank := d.Pre55Rank
	if rank == nil {
		r := RankFeriaMinor
		rank = &r
	}
	info := rank.Info()
	rankName := info.LatinName
	if val, ok := translations[info.NameResID]; ok {
		rankName = val
	}

	metadata := d.CalendarMetadata
	baseColorName := string(d.Color)
	if metadata != nil && metadata.MassColor != "" {
		baseColorName = metadata.MassColor
	}
	finalColorName := d.displayColor(baseColorName, finalName)

	var octaveTypeName string
	if metadata != nil && metadata.OctaveType != nil {
		octInfo := metadata.OctaveType.Info()
		octaveTypeName = octInfo.DisplayName
		if val, ok := translations[octInfo.NameResID]; ok {
			octaveTypeName = val
		}
	}

	var liturgy *LiturgyInfo
	if metadata != nil {
		preface := localizedPreface(metadata.MassPreface, translations)
		liturgy = &LiturgyInfo{
			Gloria:  liturgyBoolLabel(metadata.MassGloria, translations, isCommemoration, true),
			Credo:   liturgyBoolLabel(metadata.MassCredo, translations, isCommemoration, false),
			Preface: preface,
			Epistle: d.Epistle,
			Gospel:  d.Gospel,
		}
	}

	hasAbstinence, isDispensed := CalculateAbstinenceInfo(targetDate, d, finalName)

	res := LiturgicalDayJSON{
		Name:                  finalName,
		ID:                    d.ID,
		NameResID:             d.NameResID,
		ObservanceKey:         d.ObservanceKey(),
		CalendarVersion:       d.CalendarVersion.AssetID(),
		CalendarName:          d.CalendarVersion.DisplayName(),
		Pre55Grade:            info.Code,
		RankCode:              info.Code,
		RankName:              rankName,
		OctaveTypeName:        octaveTypeName,
		Color:                 finalColorName,
		IsLordFeast:           d.IsLordFeast,
		HasAbstinence:         hasAbstinence,
		IsAbstinenceDispensed: isDispensed,
		Date:                  dateStr,
		Liturgy:               liturgy,
	}

	if metadata != nil {
		res.ObservanceKind = metadata.ObservanceKind
		res.Season = metadata.Season
		res.Privileged = metadata.Privileged
		res.Precedence = metadata.Precedence
		res.FirstVespers = metadata.FirstVespers
		res.Occurrence = metadata.Occurrence
		res.Concurrence = metadata.Concurrence
		res.SuppressionMode = metadata.SuppressionMode
		res.OctaveID = metadata.OctaveID
		res.OctaveDay = metadata.OctaveDay
		res.OctaveStatus = metadata.OctaveStatus
		res.TransferStatus = metadata.TransferStatus
		res.TransferTarget = metadata.TransferTarget
	}

	return res
}

func liturgyBoolLabel(val *bool, translations map[string]string, isCommemoration, isGloria bool) string {
	posKey := "gloria"
	negKey := "no_gloria"
	posFallback := "Gloria"
	negFallback := "Without Gloria"
	if !isGloria {
		posKey = "credo"
		negKey = "no_credo"
		posFallback = "Credo"
		negFallback = "Without Credo"
	}
	if isCommemoration {
		if v, ok := translations[negKey]; ok {
			return v
		}
		return negFallback
	}
	if val == nil {
		if v, ok := translations["missing_liturgy_data"]; ok {
			return v
		}
		return "Unavailable"
	}
	if *val {
		if v, ok := translations[posKey]; ok {
			return v
		}
		return posFallback
	}
	if v, ok := translations[negKey]; ok {
		return v
	}
	return negFallback
}

func localizedPreface(val string, translations map[string]string) string {
	if val == "" {
		if v, ok := translations["missing_liturgy_data"]; ok {
			return v
		}
		return "Unavailable"
	}
	clean := strings.ToLower(val)
	clean = strings.ReplaceAll(clean, " ", "")
	clean = strings.ReplaceAll(clean, "_", "")

	key := ""
	switch clean {
	case "common":
		key = "preface_common"
	case "advent":
		key = "preface_advent"
	case "lent":
		key = "preface_lent"
	case "easter":
		key = "preface_easter"
	case "holycross":
		key = "preface_holy_cross"
	case "pentecost":
		key = "preface_pentecost"
	case "trinity":
		key = "preface_trinity"
	case "blessedvirginmary", "ourlady":
		key = "preface_our_lady"
	case "nativity", "christmas":
		key = "preface_nativity"
	case "epiphany":
		key = "preface_epiphany"
	case "apostles":
		key = "preface_apostles"
	case "sacredheart":
		key = "preface_sacred_heart"
	case "ascension":
		key = "preface_ascension"
	case "christtheking":
		key = "preface_christ_the_king"
	case "saintjoseph", "stjoseph", "joseph":
		key = "preface_joseph"
	case "johnbaptist", "baptista":
		key = "preface_john_baptist"
	case "dedication", "dedicatio":
		key = "preface_dedication"
	case "dead", "defunctorum":
		key = "preface_dead"
	case "holynamejesus":
		key = "preface_holy_name_jesus"
	case "none":
		key = "preface_none"
	}

	if key != "" {
		if v, ok := translations[key]; ok {
			return v
		}
	}
	if v, ok := translations["missing_liturgy_data"]; ok {
		return v
	}
	return "Unavailable"
}

func build1962LiturgyInfo(
	feastName, colorName string,
	liturgicalClass LiturgicalClass,
	isLordFeast bool,
	translations map[string]string,
	date time.Time,
	isCommemoration bool,
	epistle, gospel, englishName string,
) *LiturgyInfo {
	lowerName := strings.ToLower(feastName)
	isSunday := strings.Contains(strings.ToLower(englishName), "sunday") || strings.Contains(lowerName, "sunday")

	isPenitentialOrBlack := colorName == "VIOLET" || colorName == "ROSE" || colorName == "BLACK"

	var offsetFromEaster *int
	if !date.IsZero() {
		easter := CalculateEaster(date.Year())
		y1, m1, d1 := date.Date()
		y2, m2, d2 := easter.Date()
		t1 := time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC)
		t2 := time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC)
		diff := int(t1.Sub(t2).Hours() / 24)
		offsetFromEaster = &diff
	}

	isChristmasTimeFeria := !date.IsZero() && ((date.Month() == 12 && date.Day() >= 25) || (date.Month() == 1 && date.Day() <= 13))
	isEasterTimeFeria := offsetFromEaster != nil && *offsetFromEaster >= 0 && *offsetFromEaster <= 49

	hasGloria := false
	if !isCommemoration {
		if isPenitentialOrBlack {
			hasGloria = false
		} else if liturgicalClass == ClassI || liturgicalClass == ClassII || liturgicalClass == ClassIII {
			hasGloria = true
		} else if liturgicalClass == ClassIV {
			hasGloria = strings.Contains(lowerName, "saturday") || isChristmasTimeFeria || isEasterTimeFeria
		}
	}

	hasCredo := false
	if !isCommemoration {
		if liturgicalClass == ClassI {
			hasCredo = true
		} else if isSunday {
			hasCredo = true
		} else if liturgicalClass == ClassII && isLordFeast {
			hasCredo = true
		}
	}

	prefaceKey := "preface_common"
	if isLordFeast {
		prefaceKey = "preface_trinity"
	}
	if isSunday {
		prefaceKey = "preface_trinity"
	}
	if strings.Contains(lowerName, "nossa senhora") || strings.Contains(lowerName, "our lady") || strings.Contains(lowerName, "virgem maria") {
		prefaceKey = "preface_our_lady"
	}
	if offsetFromEaster != nil {
		if *offsetFromEaster >= 0 && *offsetFromEaster <= 48 {
			prefaceKey = "preface_easter"
		} else if *offsetFromEaster >= 49 && *offsetFromEaster <= 55 {
			prefaceKey = "preface_pentecost"
		} else if *offsetFromEaster >= -46 && *offsetFromEaster < 0 {
			prefaceKey = "preface_lent"
		}
	}

	prefaceName := translations[prefaceKey]
	if prefaceName == "" {
		prefaceName = translations["preface_common"]
		if prefaceName == "" {
			prefaceName = "Prefácio Comum"
		}
	}

	return &LiturgyInfo{
		Gloria:  liturgyBoolLabel(&hasGloria, translations, isCommemoration, true),
		Credo:   liturgyBoolLabel(&hasCredo, translations, isCommemoration, false),
		Preface: prefaceName,
		Epistle: epistle,
		Gospel:  gospel,
	}
}

type LiturgicalResult struct {
	MainDay         LiturgicalDay
	Commemorations  []LiturgicalDay
	CalendarVersion CalendarVersion
}

type LiturgicalResultJSON struct {
	MainDay        LiturgicalDayJSON   `json:"main_day"`
	Commemorations []LiturgicalDayJSON `json:"commemorations"`
}

func (r LiturgicalResult) ToJSON(translations map[string]string, targetDate time.Time) LiturgicalResultJSON {
	mainDayJSON := r.MainDay.ToJSON(translations, targetDate, false)
	var commsJSON []LiturgicalDayJSON
	for _, c := range r.Commemorations {
		commsJSON = append(commsJSON, c.ToJSON(translations, targetDate, true))
	}
	if commsJSON == nil {
		commsJSON = []LiturgicalDayJSON{}
	}
	return LiturgicalResultJSON{
		MainDay:        mainDayJSON,
		Commemorations: commsJSON,
	}
}
