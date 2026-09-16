package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cm-manaus/liturgical-calendar-engine/engine"
)

func setupTestRouter(t *testing.T) http.Handler {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		// Traverse up if running tests inside api/
		if _, err := os.Stat("./data"); err == nil {
			dataDir = "./data"
		} else if _, err := os.Stat("../data"); err == nil {
			dataDir = "../data"
		} else {
			dataDir = filepath.Join("..", "data")
		}
	}

	eng := engine.NewLiturgicalEngine(dataDir)
	loc := engine.NewLocalizationManager(dataDir)
	h := NewHandler(eng, loc)
	return NewRouter(h)
}

func TestHealthz(t *testing.T) {
	router := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if body["status"] != "healthy" {
		t.Errorf("Expected status healthy, got %v", body["status"])
	}
}

func TestRoot(t *testing.T) {
	router := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if body["message"] != "Welcome to the Go Liturgical Day API" {
		t.Errorf("Unexpected message: %v", body["message"])
	}
}

func TestGetLiturgicalDay1962(t *testing.T) {
	router := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/api/v1/liturgical-day?date=2026-10-12&lang=pt-br", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var resp LiturgicalResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.CalendarVersion != "tridentine_1962" {
		t.Errorf("Expected default calendar 1962, got %s", resp.CalendarVersion)
	}
	if resp.MainDay.ClassCode != "I" {
		t.Errorf("Expected Class I for Oct 12, got %s", resp.MainDay.ClassCode)
	}
}

func TestGetLiturgicalDay1954(t *testing.T) {
	router := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/api/v1/liturgical-day?date=2026-01-06&calendar=1954&lang=pt-br", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var resp LiturgicalResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.CalendarVersion != "divino_afflatu_1954" {
		t.Errorf("Expected calendar divino_afflatu_1954, got %s", resp.CalendarVersion)
	}
	if resp.MainDay.RankCode != "D1Cl" {
		t.Errorf("Expected Rank D1Cl for Epiphany 1954, got %s", resp.MainDay.RankCode)
	}
}

func TestPostLiturgicalDay(t *testing.T) {
	router := setupTestRouter(t)

	payload := LiturgicalDayRequest{
		Date:     "2026-01-06",
		Lang:     "pt-br",
		Calendar: "1954",
	}
	jsonBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/liturgical-day", bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var resp LiturgicalResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.CalendarVersion != "divino_afflatu_1954" {
		t.Errorf("Expected calendar divino_afflatu_1954, got %s", resp.CalendarVersion)
	}
}

func TestGetLiturgicalMonth(t *testing.T) {
	router := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/api/v1/liturgical-month?year=2026&month=8&calendar=1954&lang=pt-br", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var results []LiturgicalResponse
	if err := json.NewDecoder(w.Body).Decode(&results); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(results) != 31 {
		t.Errorf("Expected 31 days in August, got %d", len(results))
	}
}

func TestPanicRecovery(t *testing.T) {
	panickingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("simulated critical runtime error")
	})

	protectedHandler := RecoveryMiddleware(panickingHandler)

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Expected status 500 after recovered panic, got %d", w.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if body["error"] != "internal_server_error" {
		t.Errorf("Expected error internal_server_error, got %s", body["error"])
	}
}

func TestDocsAndOpenAPI(t *testing.T) {
	router := setupTestRouter(t)

	// Test GET /docs (Scalar interactive UI)
	reqDocs := httptest.NewRequest("GET", "/docs", nil)
	wDocs := httptest.NewRecorder()
	router.ServeHTTP(wDocs, reqDocs)

	if wDocs.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for /docs, got %d", wDocs.Code)
	}
	if !bytes.Contains(wDocs.Body.Bytes(), []byte("@scalar/api-reference")) {
		t.Errorf("Expected /docs to include Scalar script reference")
	}

	// Test GET /openapi.json
	reqOpenAPI := httptest.NewRequest("GET", "/openapi.json", nil)
	wOpenAPI := httptest.NewRecorder()
	router.ServeHTTP(wOpenAPI, reqOpenAPI)

	if wOpenAPI.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for /openapi.json, got %d", wOpenAPI.Code)
	}
	var spec map[string]any
	if err := json.NewDecoder(wOpenAPI.Body).Decode(&spec); err != nil {
		t.Fatalf("Failed to parse OpenAPI JSON: %v", err)
	}
	if spec["openapi"] != "3.1.0" {
		t.Errorf("Expected OpenAPI version 3.1.0, got %v", spec["openapi"])
	}

	// Test GET /swagger redirect
	reqSwagger := httptest.NewRequest("GET", "/swagger", nil)
	wSwagger := httptest.NewRecorder()
	router.ServeHTTP(wSwagger, reqSwagger)

	if wSwagger.Code != http.StatusMovedPermanently {
		t.Fatalf("Expected status 301 for /swagger redirect, got %d", wSwagger.Code)
	}
	if loc := wSwagger.Header().Get("Location"); loc != "/docs" {
		t.Errorf("Expected redirect to /docs, got %s", loc)
	}
}

func TestCacheControlHeaders(t *testing.T) {
	router := setupTestRouter(t)

	// Explicit date should have long cache with stale-if-error
	reqExplicit := httptest.NewRequest("GET", "/api/v1/liturgical-day?date=2026-09-08", nil)
	wExplicit := httptest.NewRecorder()
	router.ServeHTTP(wExplicit, reqExplicit)

	ccExplicit := wExplicit.Header().Get("Cache-Control")
	if !bytes.Contains([]byte(ccExplicit), []byte("max-age=604800")) {
		t.Errorf("Expected max-age=604800 for explicit date, got: %s", ccExplicit)
	}
	if !bytes.Contains([]byte(ccExplicit), []byte("stale-if-error")) {
		t.Errorf("Expected stale-if-error header for edge resilience, got: %s", ccExplicit)
	}

	// Healthz should never be cached
	reqHealthz := httptest.NewRequest("GET", "/healthz", nil)
	wHealthz := httptest.NewRecorder()
	router.ServeHTTP(wHealthz, reqHealthz)

	ccHealthz := wHealthz.Header().Get("Cache-Control")
	if ccHealthz != "no-cache, no-store, must-revalidate" {
		t.Errorf("Expected no-cache for /healthz, got: %s", ccHealthz)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	router := setupTestRouter(t)

	// Issue a sample request to increment metrics
	reqSample := httptest.NewRequest("GET", "/healthz", nil)
	wSample := httptest.NewRecorder()
	router.ServeHTTP(wSample, reqSample)

	// Fetch /metrics
	reqMetrics := httptest.NewRequest("GET", "/metrics", nil)
	wMetrics := httptest.NewRecorder()
	router.ServeHTTP(wMetrics, reqMetrics)

	if wMetrics.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for /metrics, got %d", wMetrics.Code)
	}

	body := wMetrics.Body.String()
	requiredSubstrings := []string{
		"tesouro_http_requests_total",
		"go_goroutines",
		"go_memstats_alloc_bytes",
		"tesouro_uptime_seconds",
	}

	for _, sub := range requiredSubstrings {
		if !bytes.Contains([]byte(body), []byte(sub)) {
			t.Errorf("Expected /metrics output to contain %q", sub)
		}
	}
}

func TestCalendarExportXLS(t *testing.T) {
	router := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/api/v1/calendar/export?year=2026&month=10&format=xls", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for export xls, got %d: %s", w.Code, w.Body.String())
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/vnd.ms-excel") {
		t.Errorf("Expected Content-Type application/vnd.ms-excel, got %s", ct)
	}

	cd := w.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "calendario_outubro_2026.xls") {
		t.Errorf("Expected Content-Disposition containing calendario_outubro_2026.xls, got %s", cd)
	}

	body := w.Body.String()
	expectedSubstrings := []string{
		"Calendário Litúrgico Tradicional — Outubro de 2026",
		"Calendário litúrgico e mariano",
		"Dia",
		"Calendário Litúrgico",
		"Calendário Mariano",
		"Liturgia",
		"Nossa Senhora Medianeira de Todas as Graças",
		"Abstinência de carne",
		"Primeira sexta do mês",
		"Primeiro sábado do mês",
		"Branco",
		"Verde",
		"Glória",
		"Prefácio",
	}

	for _, sub := range expectedSubstrings {
		if !strings.Contains(body, sub) {
			t.Errorf("Expected export body to contain %q", sub)
		}
	}
}

func TestCalendarExportHTML(t *testing.T) {
	router := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/api/v1/calendar/export?year=2026&month=10&format=html", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for export html, got %d: %s", w.Code, w.Body.String())
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Expected Content-Type text/html, got %s", ct)
	}

	cd := w.Header().Get("Content-Disposition")
	if cd != "" {
		t.Errorf("Expected empty Content-Disposition for html format, got %s", cd)
	}

	body := w.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Errorf("Expected HTML doctype in response")
	}
}

func TestCalendarExportValidation(t *testing.T) {
	router := setupTestRouter(t)

	// Invalid month
	reqBadMonth := httptest.NewRequest("GET", "/api/v1/calendar/export?year=2026&month=13", nil)
	wBadMonth := httptest.NewRecorder()
	router.ServeHTTP(wBadMonth, reqBadMonth)

	if wBadMonth.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for month=13, got %d", wBadMonth.Code)
	}

	// Invalid year
	reqBadYear := httptest.NewRequest("GET", "/api/v1/calendar/export?year=abc&month=10", nil)
	wBadYear := httptest.NewRecorder()
	router.ServeHTTP(wBadYear, reqBadYear)

	if wBadYear.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid year, got %d", wBadYear.Code)
	}
}

func TestCalendarExportDirectURLExtensions(t *testing.T) {
	router := setupTestRouter(t)

	// Test GET /api/v1/calendar/export.xls without format query param
	reqXLS := httptest.NewRequest("GET", "/api/v1/calendar/export.xls?year=2026&month=10", nil)
	wXLS := httptest.NewRecorder()
	router.ServeHTTP(wXLS, reqXLS)

	if wXLS.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for export.xls, got %d", wXLS.Code)
	}
	if ct := wXLS.Header().Get("Content-Type"); !strings.Contains(ct, "application/vnd.ms-excel") {
		t.Errorf("Expected application/vnd.ms-excel for export.xls, got %s", ct)
	}
	if cd := wXLS.Header().Get("Content-Disposition"); !strings.Contains(cd, "calendario_outubro_2026.xls") {
		t.Errorf("Expected attachment for export.xls, got %s", cd)
	}

	// Test GET /api/v1/calendar/export.html without format query param
	reqHTML := httptest.NewRequest("GET", "/api/v1/calendar/export.html?year=2026&month=10", nil)
	wHTML := httptest.NewRecorder()
	router.ServeHTTP(wHTML, reqHTML)

	if wHTML.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for export.html, got %d", wHTML.Code)
	}
	if ct := wHTML.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Expected text/html for export.html, got %s", ct)
	}
}

func TestCalendarExportAliases(t *testing.T) {
	router := setupTestRouter(t)

	testCases := []struct {
		url         string
		expectedSub string
	}{
		{"/api/v1/calendar/export?mes=10&ano=2026", "Outubro de 2026"},
		{"/api/v1/calendar/export?mes=outubro&ano=2026", "Outubro de 2026"},
		{"/api/v1/calendar/export?mês=10&ano=2026", "Outubro de 2026"},
		{"/api/v1/calendar/export?m=10&y=2026", "Outubro de 2026"},
		{"/api/v1/calendar/export?month=october&year=2026", "Outubro de 2026"},
		{"/api/v1/liturgical-month?mes=10&ano=2026", "2026-10-01"},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest("GET", tc.url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 for %s, got %d: %s", tc.url, w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), tc.expectedSub) {
			t.Errorf("Expected response for %s to contain %q", tc.url, tc.expectedSub)
		}
	}
}

func TestOctoberReferenceAlignment(t *testing.T) {
	router := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/api/v1/calendar/export?mes=10&ano=2026&format=html", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	os.WriteFile("/tmp/local_outubro_2026.html", []byte(body), 0644)

	assertions := []struct {
		desc     string
		expected string
	}{
		{"Day 1 Marian White", "Branco<br>Glória • Sem Credo<br>Prefácio de Nossa Senhora"},
		{"Day 1 Mediatrix Proper Readings", "Is 55, 1-3,5 • Jo 19, 25-27"},
		{"Day 3 St Therese Proper Epistle", "Is 66, 12-14 • Mt 18, 1-4"},
		{"Day 5 Feria Readings from 19th Sunday", "Ef 4, 23-28 • Mt 22, 1-14"},
		{"Day 7 Rosary Marian White and Credo", "Branco<br>Glória • Credo<br>Prefácio de Nossa Senhora"},
		{"Day 7 Rosary Readings", "Pr 8, 22-24; 32-35 • Lc 1, 26-38"},
		{"Day 8 St Bridget Proper Epistle", "I Tm 5, 3-10 • Mt 13, 44-52"},
		{"Day 9 St John Leonardi Proper Readings", "2 Cor 4, 1-6; 15-18 • Lc 10, 1-9"},
		{"Day 10 St Francis Borgia Readings", "Eclo 45, 1-6 • Mt 19, 27-29"},
		{"Day 12 Aparecida White and Credo", "Branco<br>Glória • Credo<br>Prefácio de Nossa Senhora"},
		{"Day 12 Aparecida Proper Readings", "Ap 12, 1; 5; 14 e 15-16 • Lc 1, 26-28"},
		{"Day 13 St Edward King Readings", "Sb 31, 8-11 • Lc 12, 35-40"},
		{"Day 14 St Callistus Proper Readings", "1 Pe 5, 1-4; 10-11 • Mt 16, 13-19"},
		{"Day 15 St Teresa of Avila Proper Epistle", "2 Cor 10, 17-18; 11, 1-2 • Mt 25, 1-13"},
		{"Day 21 Feria Readings from 21st Sunday", "Ef 6, 10-17 • Mt 18, 23-35"},
		{"Day 24 St Raphael Full Gospel", "Tb 12, 7-15 • Jo 5, 1-15"},
		{"Day 25 Christ the King Preface", "Prefácio de Cristo Rei"},
		{"Day 25 Christ the King Readings", "Cl 1, 12-20 • Jo 18, 33-37"},
		{"Day 26 Feria Readings from 22nd Sunday", "Flp 1, 6-11 • Mt 22, 15-21"},
		{"Day 28 Apostles Credo and Preface", "Vermelho<br>Glória • Credo<br>Prefácio dos Apóstolos"},
		{"Day 28 Apostles Readings", "Ef 4, 7-13 • Jo 15, 17-25"},
		{"Day 31 Saturday BVM White and Gloria", "Branco<br>Glória • Sem Credo<br>Prefácio de Nossa Senhora"},
		{"Day 31 Saturday BVM Readings", "Eclo 24,14-16 • Lc 11,27-28"},
	}

	for _, a := range assertions {
		if !strings.Contains(body, a.expected) {
			t.Errorf("Assertion failed for %s: body does not contain %q", a.desc, a.expected)
		}
	}
}


