package engine

import (
	"fmt"
	"strconv"
	"strings"
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

// Custom unmarshaler for LiturgicalClass to read Roman numerals from XML
func (c *LiturgicalClass) UnmarshalXMLAttr(attr attrType) error {
	switch attr.Value {
	case "I":
		*c = ClassI
	case "II":
		*c = ClassII
	case "III":
		*c = ClassIII
	case "IV":
		*c = ClassIV
	default:
		*c = ClassIII // fallback
	}
	return nil
}

// Helper structures for XML parsing attributes
type attrType struct {
	Value string
}

type LiturgicalColor string

const (
	ColorWhite  LiturgicalColor = "WHITE"
	ColorRed    LiturgicalColor = "RED"
	ColorGreen  LiturgicalColor = "GREEN"
	ColorViolet LiturgicalColor = "VIOLET"
	ColorBlack  LiturgicalColor = "BLACK"
	ColorRose   LiturgicalColor = "ROSE"
)

type LiturgicalDay struct {
	Name            string          `json:"name"`
	NameResID       string          `json:"name_res_id,omitempty"`
	NameArgs        []any           `json:"name_args,omitempty"`
	LiturgicalClass LiturgicalClass `json:"class_code"`
	Color           LiturgicalColor `json:"color"`
	IsLordFeast     bool            `json:"is_lord_feast"`
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

type LiturgicalDayJSON struct {
	Name        string `json:"name"`
	ClassCode   string `json:"class_code"`
	ClassName   string `json:"class_name"`
	Color       string `json:"color"`
	IsLordFeast bool   `json:"is_lord_feast"`
}

func (d LiturgicalDay) ToJSON(translations map[string]string) LiturgicalDayJSON {
	name := d.Name
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
		name = formatAndroidString(rawTemplate, resolvedArgs...)
	}

	classResID := d.LiturgicalClass.NameResID()
	className := fmt.Sprintf("%s Class", d.LiturgicalClass.String())
	if val, ok := translations[classResID]; ok {
		className = val
	}

	return LiturgicalDayJSON{
		Name:        name,
		ClassCode:   d.LiturgicalClass.String(),
		ClassName:   className,
		Color:       string(d.Color),
		IsLordFeast: d.IsLordFeast,
	}
}

type LiturgicalResult struct {
	MainDay        LiturgicalDay
	Commemorations []LiturgicalDay
}

func (r LiturgicalResult) ToJSON(translations map[string]string) LiturgicalResultJSON {
	mainDayJSON := r.MainDay.ToJSON(translations)
	var commsJSON []LiturgicalDayJSON
	for _, c := range r.Commemorations {
		commsJSON = append(commsJSON, c.ToJSON(translations))
	}
	if commsJSON == nil {
		commsJSON = []LiturgicalDayJSON{}
	}
	return LiturgicalResultJSON{
		MainDay:        mainDayJSON,
		Commemorations: commsJSON,
	}
}

type LiturgicalResultJSON struct {
	MainDay        LiturgicalDayJSON   `json:"main_day"`
	Commemorations []LiturgicalDayJSON `json:"commemorations"`
}
