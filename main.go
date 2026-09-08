package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
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
	litEngine       *engine.LiturgicalEngine
	locMgr          *engine.LocalizationManager
	serverStartTime time.Time
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

type statusRecorder struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytesWritten += int64(n)
	return n, err
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		// Do not log spammy /healthz checks if 200 OK
		if r.URL.Path == "/healthz" && rec.statusCode == http.StatusOK {
			return
		}

		slog.Info("http_request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_addr", r.RemoteAddr,
			"bytes", rec.bytesWritten,
		)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic_recovered",
					"panic", fmt.Sprintf("%v", rec),
					"path", r.URL.Path,
					"method", r.Method,
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error":   "internal_server_error",
					"message": "An unexpected server error occurred",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept-Language")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handleRoot)
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /api/v1/healthz", handleHealthz)

	mux.HandleFunc("GET /liturgical-day", handleGetLiturgicalDay)
	mux.HandleFunc("POST /liturgical-day", handlePostLiturgicalDay)
	mux.HandleFunc("GET /api/v1/liturgical-day", handleGetLiturgicalDay)
	mux.HandleFunc("POST /api/v1/liturgical-day", handlePostLiturgicalDay)
	mux.HandleFunc("GET /api/v1/liturgical-month", handleGetLiturgicalMonth)

	return mux
}

func main() {
	serverStartTime = time.Now()

	// Initialize structured JSON logging (Google/Enterprise standard)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	loadDotEnv(".env")

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	slog.Info("initializing_engine", "data_dir", dataDir)
	litEngine = engine.NewLiturgicalEngine(dataDir)
	locMgr = engine.NewLocalizationManager(dataDir)

	mux := setupRoutes()
	handler := corsMiddleware(recoveryMiddleware(loggingMiddleware(mux)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// P0: Protected HTTP Server with explicit connection timeouts (Anti-Slowloris)
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB header limit
	}

	// P0: Graceful Shutdown listener
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		slog.Info("server_listening", "port", port, "version", "2.1.0")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server_listen_error", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("shutting_down_server_gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server_shutdown_failed", "error", err)
	} else {
		slog.Info("server_stopped_cleanly")
	}
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	if litEngine == nil || locMgr == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "reason": "engine_not_initialized"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":         "healthy",
		"uptime_seconds": int(time.Since(serverStartTime).Seconds()),
		"version":        "2.1.0",
		"calendars":      []string{"1962", "1954"},
	})
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message":  "Welcome to the Go Liturgical Day API",
		"docs_url": "/docs",
		"calendars": []map[string]string{
			{"id": "1962", "name": "1962 (Tridentine)"},
			{"id": "1954", "name": "1954 (Divino Afflatu / Pre-55)"},
		},
		"endpoints": map[string]string{
			"health":           "/healthz",
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
	_ = json.NewEncoder(w).Encode(resp)
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
	_ = json.NewEncoder(w).Encode(results)
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
	_ = json.NewEncoder(w).Encode(resp)
}
