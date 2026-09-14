package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"tesouro-backend/engine"
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


