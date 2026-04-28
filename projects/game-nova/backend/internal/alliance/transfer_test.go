package alliance

// План 67 Ф.3: unit-тесты для transfer-кода (без БД).
//
// Полный round-trip RequestTransferCode → ConfirmTransferLeadership
// требует TEST_DATABASE_URL и FK на users/alliances; integration-тесты
// будут отдельно (как demolish_test.go в плане 65).

import (
	"strings"
	"testing"
)

func TestGenerateTransferCode_LengthAndAlphabet(t *testing.T) {
	t.Parallel()
	for i := 0; i < 50; i++ {
		code, err := generateTransferCode()
		if err != nil {
			t.Fatalf("generateTransferCode: %v", err)
		}
		if len(code) != transferCodeLen {
			t.Errorf("len = %d, want %d", len(code), transferCodeLen)
		}
		for _, r := range code {
			if !strings.ContainsRune(string(transferCodeAlphabet), r) {
				t.Errorf("rune %q not in alphabet", r)
			}
		}
	}
}

func TestGenerateTransferCode_NotConstant(t *testing.T) {
	t.Parallel()
	// Простая проверка: 100 кодов — больше 1 уникального.
	// Не сильная гарантия, но ловит «всегда возвращает один и тот
	// же код» (например, если seed забыли).
	seen := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		c, err := generateTransferCode()
		if err != nil {
			t.Fatalf("generateTransferCode: %v", err)
		}
		seen[c] = struct{}{}
	}
	if len(seen) < 50 {
		t.Errorf("expected >50 unique codes out of 100, got %d", len(seen))
	}
}
