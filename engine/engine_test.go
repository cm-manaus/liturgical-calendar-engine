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

func Test1954BrazilianPropers(t *testing.T) {
	eng := NewLiturgicalEngine(getTestDataDir())
	loc := NewLocalizationManager(getTestDataDir())
	trans := loc.GetTranslations("pt-br")

	expected := []struct {
		date     time.Time
		id       string
		rankCode string
	}{
		{time.Date(2026, 2, 13, 0, 0, 0, 0, time.UTC), "st_john_brito_brazil_1954", "D"},
		{time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC), "st_turibius_brazil_1954", "D2Cl"},
		{time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC), "finding_cross_brazil_1954", "D1Cl"},
		{time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), "dedication_own_church_brazil_1954", "D1Cl"},
		{time.Date(2026, 6, 13, 0, 0, 0, 0, time.UTC), "st_anthony_padua_brazil_1954", "D2Cl"},
		{time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC), "eucharistic_heart_brazil_1954", "DMaj"},
		{time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC), "our_lady_queen_peace_brazil_1954", "DMaj"},
		{time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC), "bl_ignatius_azevedo_brazil_1954", "DMaj"},
		{time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC), "st_rose_lima_brazil_1954", "D1Cl"},
		{time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC), "exaltation_cross_brazil_1954", "D2Cl"},
		{time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), "bvm_mediatrix_brazil_1954", "DMaj"},
		{time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC), "our_lady_aparecida_brazil_1954", "D1Cl"},
		{time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC), "st_peter_alcantara_brazil_1954", "D1Cl"},
		{time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC), "holy_relics_brazil_1954", "DMaj"},
		{time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC), "bl_roque_gonzalez_brazil_1954", "DMaj"},
		{time.Date(2026, 12, 12, 0, 0, 0, 0, time.UTC), "our_lady_guadalupe_brazil_1954", "D1Cl"},
	}

	for _, tc := range expected {
		res := eng.Resolve(tc.date, Calendar1954, true)
		jsonRes := res.ToJSON(trans, tc.date)
		if jsonRes.MainDay.ID != tc.id {
			t.Errorf("Date %s expected ID %s, got %s", tc.date.Format("2006-01-02"), tc.id, jsonRes.MainDay.ID)
		}
		if jsonRes.MainDay.RankCode != tc.rankCode {
			t.Errorf("Date %s expected rank %s, got %s", tc.date.Format("2006-01-02"), tc.rankCode, jsonRes.MainDay.RankCode)
		}
	}

	// Verify St. Rose of Lima display color is WHITE (not incorrectly classified as Marian)
	roseDate := time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC)
	roseRes := eng.Resolve(roseDate, Calendar1954, true)
	roseJSON := roseRes.ToJSON(trans, roseDate)
	if roseJSON.MainDay.Color != "WHITE" {
		t.Errorf("Expected St. Rose display color to be WHITE, got %s", roseJSON.MainDay.Color)
	}

	// Verify Eucharistic Heart proper preface
	ehDate := time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC)
	ehRes := eng.Resolve(ehDate, Calendar1954, true)
	if ehRes.MainDay.CalendarMetadata == nil || ehRes.MainDay.CalendarMetadata.MassPreface != "SacredHeart" {
		t.Errorf("Expected Eucharistic Heart preface SacredHeart, got %v", ehRes.MainDay.CalendarMetadata)
	}
}

func Test1954BrazilianOctaves(t *testing.T) {
	eng := NewLiturgicalEngine(getTestDataDir())
	loc := NewLocalizationManager(getTestDataDir())
	trans := loc.GetTranslations("pt-br")

	// 1. Dedication octave day 2 commemorated on June 2
	june2 := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
	resJune2 := eng.Resolve(june2, Calendar1954, true)
	jsonJune2 := resJune2.ToJSON(trans, june2)
	foundOctaveDay2 := false
	for _, comm := range jsonJune2.Commemorations {
		if comm.ID == "dedication_own_church_brazil_1954_octave_day_2" {
			foundOctaveDay2 = true
			break
		}
	}
	if !foundOctaveDay2 {
		t.Errorf("Expected dedication_own_church_brazil_1954_octave_day_2 commemorated on June 2")
	}

	// 2. Finding of Holy Cross octave day 8 on May 10
	may10 := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	resMay10 := eng.Resolve(may10, Calendar1954, true)
	jsonMay10 := resMay10.ToJSON(trans, may10)
	if jsonMay10.MainDay.ID != "finding_cross_brazil_1954_octave_day_8" {
		t.Errorf("Expected finding_cross_brazil_1954_octave_day_8 on May 10, got %s", jsonMay10.MainDay.ID)
	}

	// 3. Guadalupe octave day 5 on Dec 16
	dec16 := time.Date(2026, 12, 16, 0, 0, 0, 0, time.UTC)
	resDec16 := eng.Resolve(dec16, Calendar1954, true)
	jsonDec16 := resDec16.ToJSON(trans, dec16)
	foundGuadalupeOctave := false
	for _, comm := range jsonDec16.Commemorations {
		if comm.ID == "our_lady_guadalupe_brazil_1954_octave_day_5" {
			foundGuadalupeOctave = true
			break
		}
	}
	if !foundGuadalupeOctave {
		t.Errorf("Expected Guadalupe octave day 5 commemorated on Dec 16")
	}

	// 4. Dec 17: Particular octaves suspended
	dec17 := time.Date(2026, 12, 17, 0, 0, 0, 0, time.UTC)
	resDec17 := eng.Resolve(dec17, Calendar1954, true)
	jsonDec17 := resDec17.ToJSON(trans, dec17)
	for _, comm := range jsonDec17.Commemorations {
		if comm.ID == "our_lady_guadalupe_brazil_1954_octave_day_6" {
			t.Errorf("Guadalupe octave should be suspended on and after Dec 17")
		}
	}
}

func TestSeptemberEmberDays(t *testing.T) {
	eng := NewLiturgicalEngine(getTestDataDir())
	loc := NewLocalizationManager(getTestDataDir())
	trans := loc.GetTranslations("pt-br")

	// In 2026, Sept 14 is Monday. Next Wed is Sep 16, Fri is Sep 18, Sat is Sep 19.
	emberWed := time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC)
	resWed := eng.Resolve(emberWed, Calendar1962, false)
	jsonWed := resWed.ToJSON(trans, emberWed)
	if jsonWed.MainDay.ID != "ember_wednesday_september" {
		t.Errorf("Expected ember_wednesday_september on 2026-09-16, got %s", jsonWed.MainDay.ID)
	}
	if jsonWed.MainDay.Liturgy == nil || jsonWed.MainDay.Liturgy.Epistle == "" || jsonWed.MainDay.Liturgy.Gospel == "" {
		t.Errorf("Expected readings on Ember Wednesday of September")
	}

	emberFri := time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC)
	resFri := eng.Resolve(emberFri, Calendar1962, false)
	jsonFri := resFri.ToJSON(trans, emberFri)
	if jsonFri.MainDay.ID != "ember_friday_september" {
		t.Errorf("Expected ember_friday_september on 2026-09-18, got %s", jsonFri.MainDay.ID)
	}

	emberSat := time.Date(2026, 9, 19, 0, 0, 0, 0, time.UTC)
	resSat := eng.Resolve(emberSat, Calendar1962, false)
	jsonSat := resSat.ToJSON(trans, emberSat)
	if jsonSat.MainDay.ID != "ember_saturday_september" {
		t.Errorf("Expected ember_saturday_september on 2026-09-19, got %s", jsonSat.MainDay.ID)
	}
}

func TestStJosephPatronageOctave1954(t *testing.T) {
	eng := NewLiturgicalEngine(getTestDataDir())

	// 2026-04-25: St. Mark Evangelist suppresses 4th day within octave of St. Joseph
	apr25 := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	resApr25 := eng.Resolve(apr25, Calendar1954, false)
	if resApr25.MainDay.ID != "st_mark_evangelist" {
		t.Errorf("Expected st_mark_evangelist on 2026-04-25, got %s", resApr25.MainDay.ID)
	}
	for _, comm := range resApr25.Commemorations {
		if comm.ID == "fourth_day_within_octave_patronage_st_joseph" {
			t.Errorf("4th day within octave of St. Joseph should be omitted on St. Mark feast")
		}
	}

	// 2026-04-26: 3rd Sunday after Easter commemorates 5th day within octave of St. Joseph
	apr26 := time.Date(2026, 4, 26, 0, 0, 0, 0, time.UTC)
	resApr26 := eng.Resolve(apr26, Calendar1954, false)
	found5thDay := false
	for _, comm := range resApr26.Commemorations {
		if comm.ID == "fifth_day_within_octave_patronage_st_joseph" {
			found5thDay = true
			break
		}
	}
	if !found5thDay {
		t.Errorf("Expected fifth_day_within_octave_patronage_st_joseph commemorated on 2026-04-26")
	}
}

func BenchmarkEasterCalculation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = CalculateEaster(2026)
	}
}

func Benchmark1962Resolution(b *testing.B) {
	eng := NewLiturgicalEngine(getTestDataDir())
	d := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.Resolve(d, Calendar1962, true)
	}
}

func Benchmark1954Resolution(b *testing.B) {
	eng := NewLiturgicalEngine(getTestDataDir())
	d := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.Resolve(d, Calendar1954, false)
	}
}


