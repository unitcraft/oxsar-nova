package alien

import (
	"errors"
	"testing"
	"time"
)

// TestPayFormula_SecondsPerCredit — документирует легаси-формулу
// (AlienAI.class.php:993): 2ч на каждые 50 кредитов = 144 сек/кредит.
// Guard против молчаливого изменения константы.
func TestPayFormula_SecondsPerCredit(t *testing.T) {
	t.Parallel()
	// 50 кредитов должны давать ровно 2 часа продления.
	ext := time.Duration(50*AlienHoldingPaySecondsPerCredit) * time.Second
	if ext != 2*time.Hour {
		t.Errorf("50 credit extension = %v, want 2h", ext)
	}
	// 1 кредит = 144 сек.
	ext = time.Duration(1*AlienHoldingPaySecondsPerCredit) * time.Second
	if ext != 144*time.Second {
		t.Errorf("1 credit extension = %v, want 144s", ext)
	}
	// 500 кредитов = 20 часов.
	ext = time.Duration(500*AlienHoldingPaySecondsPerCredit) * time.Second
	if ext != 20*time.Hour {
		t.Errorf("500 credit extension = %v, want 20h", ext)
	}
}

// TestPayErrors_Distinct — типы ошибок должны быть различимы через
// errors.Is, чтобы HTTP-слой мог ответить корректным статусом.
func TestPayErrors_Distinct(t *testing.T) {
	t.Parallel()
	all := []error{
		ErrHoldingNotFound,
		ErrHoldingNotOwner,
		ErrInsufficientCred,
		ErrHoldingAtCap,
		ErrPayAmountInvalid,
	}
	// Каждая ошибка равна себе и не равна другим.
	for i, a := range all {
		if !errors.Is(a, a) {
			t.Errorf("err %d not equal to self", i)
		}
		for j, b := range all {
			if i == j {
				continue
			}
			if errors.Is(a, b) {
				t.Errorf("err %d wrongly matches err %d", i, j)
			}
		}
	}
}

// TestCapBoundary — cap продления = start + 15 дней. Проверяем арифметику.
func TestCapBoundary(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
	want := start.Add(15 * 24 * time.Hour)
	got := start.Add(AlienHaltingMaxRealTime)
	if !got.Equal(want) {
		t.Errorf("cap = %v, want %v", got, want)
	}
}
