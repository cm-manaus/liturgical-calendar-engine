package api

import "github.com/cm-manaus/liturgical-calendar-engine/engine"

// LiturgicalDayRequest represents the incoming JSON payload for POST requests.
type LiturgicalDayRequest struct {
	Date             string `json:"date"`
	Lang             string `json:"lang"`
	Calendar         string `json:"calendar"`
	Version          string `json:"version"`
	IncludeBrazilian *bool  `json:"include_brazilian"`
}

// LiturgicalResponse represents the standard serialized response for liturgical day queries.
type LiturgicalResponse struct {
	MainDay          engine.LiturgicalDayJSON   `json:"main_day"`
	Commemorations   []engine.LiturgicalDayJSON `json:"commemorations"`
	Date             string                     `json:"date"`
	RequestedLang    string                     `json:"requested_lang"`
	ResolvedLang     string                     `json:"resolved_lang"`
	CalendarVersion  string                     `json:"calendar_version"`
	CalendarName     string                     `json:"calendar_name"`
	IncludeBrazilian bool                       `json:"include_brazilian"`
}
