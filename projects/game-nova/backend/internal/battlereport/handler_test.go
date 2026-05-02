package battlereport

// План 72.1.10 wave 2: unit-тесты для clampDateRange и sortClause.
// Не дёргают БД — проверяют только pure-функции. Полный
// integration-тест (Pool().Query) вне scope wave 2.

import (
	"testing"
	"time"
)

func TestClampDateRange_Empty_DefaultWindow(t *testing.T) {
	now := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	dMin, dMax := clampDateRange("", "", now)

	wantMax := now.Add(-3 * 24 * time.Hour)
	wantMin := now.Add(-18 * 24 * time.Hour)
	if !dMax.Equal(wantMax) {
		t.Errorf("dMax = %s, want %s", dMax, wantMax)
	}
	if !dMin.Equal(wantMin) {
		t.Errorf("dMin = %s, want %s", dMin, wantMin)
	}
}

func TestClampDateRange_TooWide_Clamped(t *testing.T) {
	now := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	// Клиент шлёт диапазон [now-100d .. now] — оба за пределами
	// разрешённого окна [now-18d .. now-3d].
	tooEarly := now.Add(-100 * 24 * time.Hour).Format(time.RFC3339)
	tooLate := now.Format(time.RFC3339)
	dMin, dMax := clampDateRange(tooEarly, tooLate, now)

	wantMax := now.Add(-3 * 24 * time.Hour)
	wantMin := now.Add(-18 * 24 * time.Hour)
	if !dMin.Equal(wantMin) {
		t.Errorf("dMin = %s, want clamped to %s", dMin, wantMin)
	}
	if !dMax.Equal(wantMax) {
		t.Errorf("dMax = %s, want clamped to %s", dMax, wantMax)
	}
}

func TestClampDateRange_WithinWindow_Preserved(t *testing.T) {
	now := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	want1 := now.Add(-10 * 24 * time.Hour)
	want2 := now.Add(-5 * 24 * time.Hour)
	dMin, dMax := clampDateRange(want1.Format(time.RFC3339), want2.Format(time.RFC3339), now)

	if !dMin.Equal(want1) {
		t.Errorf("dMin = %s, want %s (preserved)", dMin, want1)
	}
	if !dMax.Equal(want2) {
		t.Errorf("dMax = %s, want %s (preserved)", dMax, want2)
	}
}

func TestClampDateRange_InvalidString_Defaults(t *testing.T) {
	now := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	dMin, dMax := clampDateRange("not-a-date", "garbage", now)

	wantMax := now.Add(-3 * 24 * time.Hour)
	wantMin := now.Add(-18 * 24 * time.Hour)
	if !dMin.Equal(wantMin) || !dMax.Equal(wantMax) {
		t.Errorf("invalid input must fall back to default window: got [%s..%s], want [%s..%s]",
			dMin, dMax, wantMin, wantMax)
	}
}

func TestClampDateRange_Inverted_Swapped(t *testing.T) {
	now := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	// Клиент перепутал: max < min. Функция должна свопнуть.
	earlier := now.Add(-15 * 24 * time.Hour).Format(time.RFC3339)
	later := now.Add(-5 * 24 * time.Hour).Format(time.RFC3339)
	dMin, dMax := clampDateRange(later, earlier, now)
	if dMin.After(dMax) {
		t.Errorf("after swap: dMin=%s > dMax=%s — must be swapped", dMin, dMax)
	}
}

func TestSortClause_AllFields(t *testing.T) {
	// План 72.1.50 ч.5 (72.1.10 wave 3): SQL получил JOIN planets,
	// поэтому col-фрагменты префиксированы `b.` (battle_reports) или
	// `p.` (planets).
	cases := []struct {
		field, order string
		wantCol      string
		wantDir      string
	}{
		{"", "", "b.at", "DESC"},
		{"date", "", "b.at", "DESC"},
		{"date", "asc", "b.at", "ASC"},
		{"rounds", "desc", "b.rounds", "DESC"},
		{"debris", "asc", "(b.debris_metal + b.debris_silicon)", "ASC"},
		{"loot", "desc", "(b.loot_metal + b.loot_silicon + b.loot_hydrogen)", "DESC"},
		{"outcome", "asc", "b.winner", "ASC"},
		{"moon", "desc", "b.is_moon", "DESC"},
		// План 72.1.50 ч.5 wave 3: planet_name → planets.name (через JOIN).
		{"planet_name", "asc", "p.name", "ASC"},
		{"planet_name", "desc", "p.name", "DESC"},
		// Неизвестное поле → дефолт `b.at`.
		{"unknown", "asc", "b.at", "ASC"},
		// SQL-инъекция через field — должна быть проигнорирована (whitelist).
		{"; DROP TABLE users;--", "asc", "b.at", "ASC"},
		// Неизвестное направление → DESC.
		{"date", "ascending", "b.at", "DESC"},
	}
	for _, c := range cases {
		col, dir := sortClause(c.field, c.order)
		if col != c.wantCol || dir != c.wantDir {
			t.Errorf("sortClause(%q,%q) = (%q,%q), want (%q,%q)",
				c.field, c.order, col, dir, c.wantCol, c.wantDir)
		}
	}
}
