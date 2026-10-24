package engine

import (
	"testing"
	"time"
)

// TestResumedSundaysAfterEpiphany trava a numeração do fim do tempo depois de
// Pentecostes. Pela rubrica de 1962, os domingos entre o 23º e o último são
// nomeados pela POSIÇÃO ordinal (24º, 25º...) e rezam a Missa do Domingo depois
// de Epifania retomado (o 6º imediatamente antes do Último). O último é o
// "Último Domingo depois de Pentecostes" (Missa do 24º). Regressão do bug que
// numerava os retomados como 4º/5º e do placeholder "ord_24" não resolvido.
func TestResumedSundaysAfterEpiphany(t *testing.T) {
	eng := NewLiturgicalEngine(getTestDataDir())
	loc := NewLocalizationManager(getTestDataDir())
	trans := loc.GetTranslations("en")

	cases := []struct {
		date time.Time
		want string
	}{
		// 2026: Páscoa 05/04, Advento 29/11 -> 2 domingos retomados (posições 24, 25; último = 26).
		{time.Date(2026, 11, 8, 0, 0, 0, 0, time.UTC), "24th Sunday after Pentecost (Mass of the 5th Sunday after Epiphany)"},
		{time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC), "25th Sunday after Pentecost (Mass of the 6th Sunday after Epiphany)"},
		{time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC), "Last Sunday after Pentecost"},
		// 2024: Páscoa 31/03 (cedo) -> 3 retomados (posições 24, 25, 26; último = 27).
		{time.Date(2024, 11, 10, 0, 0, 0, 0, time.UTC), "25th Sunday after Pentecost (Mass of the 5th Sunday after Epiphany)"},
		{time.Date(2024, 11, 17, 0, 0, 0, 0, time.UTC), "26th Sunday after Pentecost (Mass of the 6th Sunday after Epiphany)"},
		{time.Date(2024, 11, 24, 0, 0, 0, 0, time.UTC), "Last Sunday after Pentecost"},
	}
	for _, tc := range cases {
		got := eng.Resolve(tc.date, Calendar1962, false).ToJSON(trans, tc.date).MainDay.Name
		if got != tc.want {
			t.Errorf("%s: esperado %q, obtido %q", tc.date.Format("2006-01-02"), tc.want, got)
		}
	}
}

// TestNormalSundayAfterPentecostNumbering garante que a numeração comum (sem
// domingos retomados) continua correta.
func TestNormalSundayAfterPentecostNumbering(t *testing.T) {
	eng := NewLiturgicalEngine(getTestDataDir())
	loc := NewLocalizationManager(getTestDataDir())
	trans := loc.GetTranslations("en")

	d := time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC) // 2º Domingo depois de Pentecostes
	got := eng.Resolve(d, Calendar1962, false).ToJSON(trans, d).MainDay.Name
	if got != "2nd Sunday after Pentecost" {
		t.Errorf("2026-06-07: esperado %q, obtido %q", "2nd Sunday after Pentecost", got)
	}
}
