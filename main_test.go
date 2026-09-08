package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"tesouro-backend/engine"
)

func setupTestServer() http.Handler {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	litEngine = engine.NewLiturgicalEngine(dataDir)
	locMgr = engine.NewLocalizationManager(dataDir)

	mux := setupRoutes()
	return corsMiddleware(recoveryMiddleware(loggingMiddleware(mux)))
}

func TestHTTPHealthz(t *testing.T) {
	handler := setupTestServer()

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

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

func TestHTTPRoot(t *testing.T) {
	handler := setupTestServer()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

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

func TestHTTPGetLiturgicalDay1962(t *testing.T) {
	handler := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/liturgical-day?date=2026-10-12&lang=pt-br", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

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

func TestHTTPGetLiturgicalDay1954(t *testing.T) {
	handler := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/liturgical-day?date=2026-01-06&calendar=1954&lang=pt-br", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

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

func TestHTTPPostLiturgicalDay(t *testing.T) {
	handler := setupTestServer()

	payload := LiturgicalDayRequest{
		Date:     "2026-01-06",
		Lang:     "pt-br",
		Calendar: "1954",
	}
	jsonBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/liturgical-day", bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

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

func TestHTTPGetLiturgicalMonth(t *testing.T) {
	handler := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/liturgical-month?year=2026&month=8&calendar=1954&lang=pt-br", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

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
