package engine

import (
	"strings"
)

// CalendarVersion identifies which liturgical calendar ruleset is active.
type CalendarVersion string

const (
	Calendar1962 CalendarVersion = "tridentine_1962"
	Calendar1954 CalendarVersion = "divino_afflatu_1954"
)

func (v CalendarVersion) DisplayName() string {
	switch v {
	case Calendar1954:
		return "1954 (Divino Afflatu)"
	default:
		return "1962"
	}
}

func (v CalendarVersion) AssetID() string {
	switch v {
	case Calendar1954:
		return "divino_afflatu_1954"
	default:
		return "tridentine_1962"
	}
}

func ParseCalendarVersion(input string) CalendarVersion {
	clean := strings.ToLower(strings.TrimSpace(input))
	switch clean {
	case "1954", "54", "pre55", "pre-55", "divino_afflatu", "divino_afflatu_1954", "divinoafflatu1954":
		return Calendar1954
	case "1962", "62", "tridentine", "tridentine_1962", "tridentine1962":
		return Calendar1962
	default:
		return Calendar1962
	}
}

// Pre55Rank represents the pre-1955 liturgical rank hierarchy (Divino Afflatu, 1911-1954).
type Pre55Rank string

const (
	RankD1Cl              Pre55Rank = "D1Cl"
	RankD2Cl              Pre55Rank = "D2Cl"
	RankDMaj              Pre55Rank = "DMaj"
	RankD                 Pre55Rank = "D"
	RankSD                Pre55Rank = "SD"
	RankS                 Pre55Rank = "S"
	RankFeriaMajor        Pre55Rank = "FeriaMajor"
	RankFeriaMinor        Pre55Rank = "FeriaMinor"
	RankFeriaPrivilegiata Pre55Rank = "FeriaPrivilegiata"
)

type Pre55RankInfo struct {
	Rank             float64
	Code             string
	LatinName        string
	NameResID        string
	HasFirstVespers  bool
	HasCommemoration bool
}

var pre55RankTable = map[Pre55Rank]Pre55RankInfo{
	RankD1Cl: {
		Rank:             6.5,
		Code:             "D1Cl",
		LatinName:        "Duplex I Classis",
		NameResID:        "pre55_d1cl",
		HasFirstVespers:  true,
		HasCommemoration: false,
	},
	RankD2Cl: {
		Rank:             5.0,
		Code:             "D2Cl",
		LatinName:        "Duplex II Classis",
		NameResID:        "pre55_d2cl",
		HasFirstVespers:  true,
		HasCommemoration: false,
	},
	RankDMaj: {
		Rank:             4.0,
		Code:             "DMaj",
		LatinName:        "Duplex Maius",
		NameResID:        "pre55_dmaj",
		HasFirstVespers:  false,
		HasCommemoration: false,
	},
	RankD: {
		Rank:             3.0,
		Code:             "D",
		LatinName:        "Duplex",
		NameResID:        "pre55_d",
		HasFirstVespers:  false,
		HasCommemoration: false,
	},
	RankSD: {
		Rank:             2.2,
		Code:             "SD",
		LatinName:        "Semiduplex",
		NameResID:        "pre55_sd",
		HasFirstVespers:  false,
		HasCommemoration: true,
	},
	RankS: {
		Rank:             1.2,
		Code:             "S",
		LatinName:        "Simplex",
		NameResID:        "pre55_s",
		HasFirstVespers:  false,
		HasCommemoration: true,
	},
	RankFeriaMajor: {
		Rank:             1.1,
		Code:             "FeriaMajor",
		LatinName:        "Feria Major",
		NameResID:        "pre55_feria_major",
		HasFirstVespers:  false,
		HasCommemoration: false,
	},
	RankFeriaMinor: {
		Rank:             1.0,
		Code:             "Feria",
		LatinName:        "Feria",
		NameResID:        "pre55_feria_minor",
		HasFirstVespers:  false,
		HasCommemoration: false,
	},
	RankFeriaPrivilegiata: {
		Rank:             7.0,
		Code:             "FeriaPrivilegiata",
		LatinName:        "Feria Privilegiata",
		NameResID:        "pre55_feria_privilegiata",
		HasFirstVespers:  false,
		HasCommemoration: false,
	},
}

func (r Pre55Rank) Info() Pre55RankInfo {
	if info, ok := pre55RankTable[r]; ok {
		return info
	}
	return pre55RankTable[RankFeriaMinor]
}

func (r Pre55Rank) IsFeria() bool {
	return r == RankFeriaMajor || r == RankFeriaMinor || r == RankFeriaPrivilegiata
}

func (r Pre55Rank) PrecedenceIndex() int {
	switch r {
	case RankFeriaPrivilegiata:
		return -1
	case RankD1Cl:
		return 0
	case RankD2Cl:
		return 1
	case RankDMaj:
		return 2
	case RankD:
		return 3
	case RankSD:
		return 4
	case RankS:
		return 5
	case RankFeriaMajor:
		return 6
	case RankFeriaMinor:
		return 7
	default:
		return 8
	}
}

func (r Pre55Rank) Precedes(other Pre55Rank) bool {
	return r.PrecedenceIndex() < other.PrecedenceIndex()
}

func Pre55RankFromCode(code string) (Pre55Rank, bool) {
	switch code {
	case "D1Cl":
		return RankD1Cl, true
	case "D2Cl":
		return RankD2Cl, true
	case "DMaj":
		return RankDMaj, true
	case "D":
		return RankD, true
	case "SD":
		return RankSD, true
	case "S":
		return RankS, true
	case "FMaj", "FeriaMajor":
		return RankFeriaMajor, true
	case "F", "Feria", "FeriaMinor":
		return RankFeriaMinor, true
	case "FPriv", "FeriaPrivilegiata":
		return RankFeriaPrivilegiata, true
	}
	for rank, info := range pre55RankTable {
		if info.Code == code || string(rank) == code {
			return rank, true
		}
	}
	return RankFeriaMinor, false
}

func Pre55RankFromAsset(code, observanceKind string, privileged bool) Pre55Rank {
	parsed, ok := Pre55RankFromCode(code)
	if !ok || observanceKind != "feria" || parsed.IsFeria() {
		return parsed
	}
	if privileged {
		return RankFeriaPrivilegiata
	}
	if parsed == RankSD {
		return RankFeriaMinor
	}
	return RankFeriaMajor
}

// Pre55OctaveType categorizes octaves in the pre-1955 rubrics.
type Pre55OctaveType string

const (
	OctavePrivilegedI   Pre55OctaveType = "privileged_i"
	OctavePrivilegedII  Pre55OctaveType = "privileged_ii"
	OctavePrivilegedIII Pre55OctaveType = "privileged_iii"
	OctaveCommon        Pre55OctaveType = "common"
	OctaveSimple        Pre55OctaveType = "simple"
)

type Pre55OctaveInfo struct {
	Code        string
	DisplayName string
	NameResID   string
}

var octaveInfoTable = map[Pre55OctaveType]Pre55OctaveInfo{
	OctavePrivilegedI: {
		Code:        "privileged_i",
		DisplayName: "Oitava Privilegiada de 1ª Ordem",
		NameResID:   "pre55_octave_privileged_i",
	},
	OctavePrivilegedII: {
		Code:        "privileged_ii",
		DisplayName: "Oitava Privilegiada de 2ª Ordem",
		NameResID:   "pre55_octave_privileged_ii",
	},
	OctavePrivilegedIII: {
		Code:        "privileged_iii",
		DisplayName: "Oitava Privilegiada de 3ª Ordem",
		NameResID:   "pre55_octave_privileged_iii",
	},
	OctaveCommon: {
		Code:        "common",
		DisplayName: "Oitava Comum",
		NameResID:   "pre55_octave_common",
	},
	OctaveSimple: {
		Code:        "simple",
		DisplayName: "Oitava Simples",
		NameResID:   "pre55_octave_simple",
	},
}

func (o Pre55OctaveType) Info() Pre55OctaveInfo {
	if info, ok := octaveInfoTable[o]; ok {
		return info
	}
	return Pre55OctaveInfo{Code: string(o), DisplayName: string(o), NameResID: ""}
}

func Pre55OctaveTypeFromCode(code string) (Pre55OctaveType, bool) {
	switch code {
	case "privileged_i", "privileged_1", "privileged_first":
		return OctavePrivilegedI, true
	case "privileged_ii", "privileged_2", "privileged_second":
		return OctavePrivilegedII, true
	case "privileged_iii", "privileged_3", "privileged_third":
		return OctavePrivilegedIII, true
	case "common":
		return OctaveCommon, true
	case "simple":
		return OctaveSimple, true
	default:
		return "", false
	}
}

func Pre55OctaveTypeFromOctaveID(octaveID string) (Pre55OctaveType, bool) {
	switch octaveID {
	case "Easter", "Pentecost":
		return OctavePrivilegedI, true
	case "Epiphany", "CorpusChristi":
		return OctavePrivilegedII, true
	case "Christmas", "Ascension", "SacredHeart":
		return OctavePrivilegedIII, true
	case "Assumption", "SaintJoseph", "SaintsPeterPaul", "PeterPaul", "StJohnBaptist", "AllSaints", "ImmaculateConception":
		return OctaveCommon, true
	case "NativityBVM", "StLawrence":
		return OctaveSimple, true
	default:
		return "", false
	}
}
