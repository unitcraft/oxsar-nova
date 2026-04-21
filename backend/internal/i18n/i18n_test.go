package i18n

import (
	"os"
	"path/filepath"
	"testing"
)

// helper: пишет словари в temp-папку и возвращает Bundle.
func loadTestBundle(t *testing.T, files map[string]string) *Bundle {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	b, err := Load(dir, LangRu)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return b
}

func TestTr_Hit(t *testing.T) {
	t.Parallel()
	b := loadTestBundle(t, map[string]string{
		"ru.yml": "global:\n  ATTACK: \"Атака\"\n",
	})
	if got := b.Tr(LangRu, "global", "ATTACK"); got != "Атака" {
		t.Fatalf("got %q", got)
	}
}

func TestTr_FallbackToRu(t *testing.T) {
	t.Parallel()
	b := loadTestBundle(t, map[string]string{
		"ru.yml": "global:\n  ATTACK: \"Атака\"\n",
		"en.yml": "global: {}\n",
	})
	if got := b.Tr(LangEn, "global", "ATTACK"); got != "Атака" {
		t.Fatalf("expected ru fallback, got %q", got)
	}
}

func TestTr_MissingKeyMarker(t *testing.T) {
	t.Parallel()
	b := loadTestBundle(t, map[string]string{
		"ru.yml": "global: {}\n",
	})
	if got := b.Tr(LangRu, "global", "MISSING"); got != "[global.MISSING]" {
		t.Fatalf("got %q", got)
	}
}

func TestTr_Sprintf(t *testing.T) {
	t.Parallel()
	b := loadTestBundle(t, map[string]string{
		"ru.yml": "mission:\n  ARRIVED: \"Флот прибыл на %s в %d:%d:%d\"\n",
	})
	got := b.Tr(LangRu, "mission", "ARRIVED", "Moon", 1, 42, 8)
	want := "Флот прибыл на Moon в 1:42:8"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestLoad_NoFallbackError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// только en
	if err := os.WriteFile(filepath.Join(dir, "en.yml"), []byte("x: {}\n"), 0o644); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err := Load(dir, LangRu); err == nil {
		t.Fatalf("expected error when fallback missing")
	}
}

func TestHas(t *testing.T) {
	t.Parallel()
	b := loadTestBundle(t, map[string]string{
		"ru.yml": "a:\n  X: \"x\"\n",
	})
	if !b.Has(LangRu, "a", "X") {
		t.Fatalf("Has returned false for existing key")
	}
	if b.Has(LangRu, "a", "Y") {
		t.Fatalf("Has returned true for missing key")
	}
}

func TestLanguages(t *testing.T) {
	t.Parallel()
	b := loadTestBundle(t, map[string]string{
		"ru.yml": "a:\n  X: \"x\"\n",
		"en.yml": "a:\n  X: \"x\"\n",
	})
	langs := b.Languages()
	if len(langs) != 2 {
		t.Fatalf("expected 2 languages, got %d: %v", len(langs), langs)
	}
}

func TestLocale_KnownLang(t *testing.T) {
	t.Parallel()
	b := loadTestBundle(t, map[string]string{
		"ru.yml": "grp:\n  KEY: \"значение\"\n",
	})
	d := b.Locale(LangRu)
	if d == nil {
		t.Fatal("Locale(LangRu) returned nil")
	}
	if d["grp"]["KEY"] != "значение" {
		t.Errorf("Locale value = %q, want значение", d["grp"]["KEY"])
	}
}

func TestLocale_UnknownLangFallsBack(t *testing.T) {
	t.Parallel()
	b := loadTestBundle(t, map[string]string{
		"ru.yml": "grp:\n  KEY: \"значение\"\n",
	})
	// LangEn not loaded → fallback to LangRu.
	d := b.Locale(LangEn)
	if d == nil {
		t.Fatal("Locale fallback returned nil")
	}
	if d["grp"]["KEY"] != "значение" {
		t.Errorf("Locale fallback value = %q, want значение", d["grp"]["KEY"])
	}
}
