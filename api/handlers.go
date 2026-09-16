package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cm-manaus/liturgical-calendar-engine/engine"
)

// Handler encapsulates HTTP endpoints and dependencies with clean dependency injection.
type Handler struct {
	engine    *engine.LiturgicalEngine
	locMgr    *engine.LocalizationManager
	startTime time.Time
}

// NewHandler constructs an initialized Handler with its required domain engines.
func NewHandler(eng *engine.LiturgicalEngine, loc *engine.LocalizationManager) *Handler {
	return &Handler{
		engine:    eng,
		locMgr:    loc,
		startTime: time.Now(),
	}
}

// HandleRoot provides API metadata, supported calendars, and route documentation.
func (h *Handler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message":  "Welcome to the Go Liturgical Day API",
		"docs_url": "/docs",
		"calendars": []map[string]string{
			{"id": "1962", "name": "1962 (Tridentine)"},
			{"id": "1954", "name": "1954 (Divino Afflatu / Pre-55)"},
		},
		"endpoints": map[string]string{
			"health":           "/healthz",
			"docs":             "/docs",
			"openapi":          "/openapi.json",
			"liturgical_day":   "/api/v1/liturgical-day",
			"liturgical_month": "/api/v1/liturgical-month",
			"calendar_export":  "/api/v1/calendar/export",
		},
	})
}

// HandleHealthz reports system health and readiness to Docker, orchestrators, and monitoring agents.
func (h *Handler) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	if h.engine == nil || h.locMgr == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "reason": "engine_not_initialized"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":         "healthy",
		"uptime_seconds": int(time.Since(h.startTime).Seconds()),
		"version":        "2.4.0",
		"calendars":      []string{"1962", "1954"},
	})
}

// HandleGetLiturgicalDay resolves the liturgy of a given date via query parameters.
func (h *Handler) HandleGetLiturgicalDay(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	lang := r.URL.Query().Get("lang")
	calendarStr := r.URL.Query().Get("calendar")
	if calendarStr == "" {
		calendarStr = r.URL.Query().Get("version")
	}
	if calendarStr == "" {
		calendarStr = r.URL.Query().Get("calendar_version")
	}

	includeBrazilianStr := r.URL.Query().Get("include_brazilian")
	includeBrazilian := true
	if includeBrazilianStr == "false" {
		includeBrazilian = false
	}
	acceptLanguage := r.Header.Get("Accept-Language")

	resp, err := h.resolveLiturgicalDay(dateStr, lang, acceptLanguage, calendarStr, includeBrazilian)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if dateStr != "" {
		w.Header().Set("Cache-Control", "public, max-age=604800, stale-while-revalidate=86400, stale-if-error=2592000")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=300, stale-while-revalidate=60, stale-if-error=3600")
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// HandlePostLiturgicalDay resolves the liturgy of a given date via JSON payload.
func (h *Handler) HandlePostLiturgicalDay(w http.ResponseWriter, r *http.Request) {
	var req LiturgicalDayRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	includeBrazilian := true
	if req.IncludeBrazilian != nil {
		includeBrazilian = *req.IncludeBrazilian
	}
	acceptLanguage := r.Header.Get("Accept-Language")

	calStr := req.Calendar
	if calStr == "" {
		calStr = req.Version
	}

	resp, err := h.resolveLiturgicalDay(req.Date, req.Lang, acceptLanguage, calStr, includeBrazilian)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// HandleGetLiturgicalMonth resolves all liturgical days of a specified month.
func (h *Handler) HandleGetLiturgicalMonth(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")
	lang := r.URL.Query().Get("lang")
	calendarStr := r.URL.Query().Get("calendar")
	if calendarStr == "" {
		calendarStr = r.URL.Query().Get("version")
	}
	if calendarStr == "" {
		calendarStr = r.URL.Query().Get("calendar_version")
	}

	includeBrazilianStr := r.URL.Query().Get("include_brazilian")
	includeBrazilian := true
	if includeBrazilianStr == "false" {
		includeBrazilian = false
	}
	acceptLanguage := r.Header.Get("Accept-Language")

	var year, month int
	var err error
	if yearStr != "" {
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			http.Error(w, "invalid year format", http.StatusBadRequest)
			return
		}
	} else {
		year = time.Now().Year()
	}

	if monthStr != "" {
		month, err = strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			http.Error(w, "invalid month format", http.StatusBadRequest)
			return
		}
	} else {
		month = int(time.Now().Month())
	}

	tNext := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC)
	daysInMonth := tNext.Day()

	results := make([]LiturgicalResponse, 0, daysInMonth)
	for day := 1; day <= daysInMonth; day++ {
		dateStr := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
		resp, err := h.resolveLiturgicalDay(dateStr, lang, acceptLanguage, calendarStr, includeBrazilian)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	if yearStr != "" && monthStr != "" {
		w.Header().Set("Cache-Control", "public, max-age=604800, stale-while-revalidate=86400, stale-if-error=2592000")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=3600, stale-while-revalidate=300, stale-if-error=86400")
	}
	_ = json.NewEncoder(w).Encode(results)
}

func (h *Handler) resolveLiturgicalDay(dateStr, lang, acceptLanguage, calendarStr string, includeBrazilian bool) (LiturgicalResponse, error) {
	var targetDate time.Time
	var err error

	if dateStr != "" {
		targetDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return LiturgicalResponse{}, fmt.Errorf("invalid date format. Expected YYYY-MM-DD")
		}
	} else {
		targetDate = time.Now()
	}

	selectedLang := "en"
	if lang != "" {
		selectedLang = lang
	} else if acceptLanguage != "" {
		parts := strings.Split(acceptLanguage, ",")
		if len(parts) > 0 {
			first := strings.Split(parts[0], ";")[0]
			selectedLang = strings.TrimSpace(first)
		}
	}

	calVersion := engine.ParseCalendarVersion(calendarStr)
	result := h.engine.Resolve(targetDate, calVersion, includeBrazilian)
	translations := h.locMgr.GetTranslations(selectedLang)

	resolvedLang := selectedLang
	if len(translations) == 0 {
		translations = h.locMgr.GetTranslations("en")
		resolvedLang = "en"
	}

	jsonResult := result.ToJSON(translations, targetDate)

	return LiturgicalResponse{
		MainDay:          jsonResult.MainDay,
		Commemorations:   jsonResult.Commemorations,
		Date:             targetDate.Format("2006-01-02"),
		RequestedLang:    selectedLang,
		ResolvedLang:     resolvedLang,
		CalendarVersion:  calVersion.AssetID(),
		CalendarName:     calVersion.DisplayName(),
		IncludeBrazilian: includeBrazilian,
		HasAbstinence:         jsonResult.MainDay.HasAbstinence,
		IsAbstinenceDispensed: jsonResult.MainDay.IsAbstinenceDispensed,
	}, nil
}
