package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"tesouro-backend/engine"
)

type LiturgicalDayRequest struct {
	Date             string `json:"date"`
	Lang             string `json:"lang"`
	Calendar         string `json:"calendar"`
	Version          string `json:"version"`
	IncludeBrazilian *bool  `json:"include_brazilian"`
}

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

var (
	litEngine *engine.LiturgicalEngine
	locMgr    *engine.LocalizationManager
)

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

func main() {
	// Load environment variables
	loadDotEnv(".env")

	// Determine data directory
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	// Initialize engine and localization
	litEngine = engine.NewLiturgicalEngine(dataDir)
	locMgr = engine.NewLocalizationManager(dataDir)

	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("GET /", handleRoot)

	mux.HandleFunc("GET /liturgical-day", handleGetLiturgicalDay)
	mux.HandleFunc("POST /liturgical-day", handlePostLiturgicalDay)
	mux.HandleFunc("GET /api/v1/liturgical-day", handleGetLiturgicalDay)
	mux.HandleFunc("POST /api/v1/liturgical-day", handlePostLiturgicalDay)
	mux.HandleFunc("GET /api/v1/liturgical-month", handleGetLiturgicalMonth)

	// PORT configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server with CORS
	fmt.Printf("Go Liturgical Backend listening on port %s...\n", port)
	err := http.ListenAndServe(":"+port, corsMiddleware(mux))
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept-Language")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message":  "Welcome to the Go Liturgical Day API",
		"docs_url": "/docs",
		"calendars": []map[string]string{
			{"id": "1962", "name": "1962 (Tridentine)"},
			{"id": "1954", "name": "1954 (Divino Afflatu / Pre-55)"},
		},
		"endpoints": map[string]string{
			"liturgical_day":   "/api/v1/liturgical-day",
			"liturgical_month": "/api/v1/liturgical-month",
		},
	})
}

func resolveLiturgicalDay(dateStr, lang, acceptLanguage, calendarStr string, includeBrazilian bool) (LiturgicalResponse, error) {
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

	// Resolve language
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
	result := litEngine.Resolve(targetDate, calVersion, includeBrazilian)
	translations := locMgr.GetTranslations(selectedLang)

	resolvedLang := selectedLang
	if len(translations) == 0 {
		translations = locMgr.GetTranslations("en")
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
	}, nil
}

func handleGetLiturgicalDay(w http.ResponseWriter, r *http.Request) {
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

	resp, err := resolveLiturgicalDay(dateStr, lang, acceptLanguage, calendarStr, includeBrazilian)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleGetLiturgicalMonth(w http.ResponseWriter, r *http.Request) {
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

	// Calculate number of days in that month
	tNext := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC)
	daysInMonth := tNext.Day()

	results := make([]LiturgicalResponse, 0, daysInMonth)
	for day := 1; day <= daysInMonth; day++ {
		dateStr := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
		resp, err := resolveLiturgicalDay(dateStr, lang, acceptLanguage, calendarStr, includeBrazilian)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func handlePostLiturgicalDay(w http.ResponseWriter, r *http.Request) {
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

	resp, err := resolveLiturgicalDay(req.Date, req.Lang, acceptLanguage, calStr, includeBrazilian)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
