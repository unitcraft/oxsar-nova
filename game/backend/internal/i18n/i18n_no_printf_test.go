package i18n_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestNoPrintfPlaceholders проверяет, что в configs/i18n/*.yml нет
// позиционных плейсхолдеров %s/%d. Все плейсхолдеры должны быть
// именованными: {{name}}.
func TestNoPrintfPlaceholders(t *testing.T) {
	rePrintf := regexp.MustCompile(`%[sd]`)

	// Ищем configs/i18n/ относительно этого файла (backend/internal/i18n/).
	dir := filepath.Join("..", "..", "..", "configs", "i18n")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("cannot read %s: %v", dir, err)
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yml") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if rePrintf.MatchString(line) {
				t.Errorf("%s:%d содержит %%s/%%d: %s", e.Name(), i+1, strings.TrimSpace(line))
			}
		}
	}
}
