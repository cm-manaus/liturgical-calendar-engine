package engine

import (
	"path/filepath"
	"testing"
	"time"
)

func getTestDataDir() string {
	return filepath.Join("..", "data")
}

func TestEasterCalculation(t *testing.T) {
	cases := []struct {
		year     int
		expected string
	}{
		{2024, "2024-03-31"},
		{2025, "2025-04-20"},
		{2026, "2026-04-05"},
		{2027, "2027-03-28"},
	}

	for _, tc := range cases {
		easter := CalculateEaster(tc.year)
		actual := easter.Format("2006-01-02")
		if actual != tc.expected {
			t.Errorf("For year %d expected Easter %s, got %s", tc.year, tc.expected, actual)
		}
	}
}

func Test1962Resolution(t *testing.T) {
	eng := NewLiturgicalEngine(getTestDataDir())
	loc := NewLocalizationManager(getTestDataDir())
	trans := loc.GetTranslations("pt-br")

	// Test 1: Oct 12, 2026 (Nossa Senhora Aparecida - Class I, White/Blue)
	d1 := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	res1 := eng.Resolve(d1, Calendar1962, true)
	json1 := res1.ToJSON(trans, d1)
	if json1.MainDay.ClassCode != "I" {
		t.Errorf("Expected Class I for Oct 12, got %s", json1.MainDay.ClassCode)
	}
	if json1.MainDay.Color != "BLUE" {
		t.Errorf("Expected BLUE color for Marian feast on Oct 12, got %s", json1.MainDay.Color)
	}

	// Test 2: March 25, 2024 (Annunciation in Holy Week -> Transferred in 1962)
	// In 2024, Easter is March 31. March 25 is Monday of Holy Week.
	// Annunciation is transferred to Monday after Low Sunday (April 8, 2024).
	dMar25_2024 := time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC)
	resMar25 := eng.Resolve(dMar25_2024, Calendar1962, true)
	if resMar25.MainDay.NameResID == "annunciation" {
		t.Errorf("March 25, 2024 should not be Annunciation (must be transferred)")
	}

	dApr8_2024 := time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC)
	resApr8 := eng.Resolve(dApr8_2024, Calendar1962, true)
	if resApr8.MainDay.NameResID != "annunciation" {
		t.Errorf("April 8, 2024 should be transferred Annunciation, got %s", resApr8.MainDay.NameResID)
	}
}

func Test1954Resolution(t *testing.T) {
	eng := NewLiturgicalEngine(getTestDataDir())
	loc := NewLocalizationManager(getTestDataDir())
	trans := loc.GetTranslations("pt-br")

	// Test 1: Circumcision (Jan 1) - D2Cl in 1954
	dJan1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	resJan1 := eng.Resolve(dJan1, Calendar1954, true)
	jsonJan1 := resJan1.ToJSON(trans, dJan1)
	if jsonJan1.MainDay.RankCode != "D2Cl" {
		t.Errorf("Jan 1 1954 expected D2Cl, got %s", jsonJan1.MainDay.RankCode)
	}
	if jsonJan1.MainDay.Liturgy == nil || jsonJan1.MainDay.Liturgy.Gloria == "" {
		t.Errorf("Expected Liturgy Info on Jan 1 1954")
	}

	// Test 2: Epiphany (Jan 6) - D1Cl
	dJan6 := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	resJan6 := eng.Resolve(dJan6, Calendar1954, true)
	jsonJan6 := resJan6.ToJSON(trans, dJan6)
	if jsonJan6.MainDay.RankCode != "D1Cl" {
		t.Errorf("Jan 6 1954 expected D1Cl, got %s", jsonJan6.MainDay.RankCode)
	}

	// Test 3: Multiple Commemorations (Jan 19)
	dJan19 := time.Date(2026, 1, 19, 0, 0, 0, 0, time.UTC)
	resJan19 := eng.Resolve(dJan19, Calendar1954, true)
	jsonJan19 := resJan19.ToJSON(trans, dJan19)
	t.Logf("Jan 19 1954 Main: %s, comms count: %d", jsonJan19.MainDay.Name, len(jsonJan19.Commemorations))
}
