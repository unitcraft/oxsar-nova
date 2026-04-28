package alien_test

// План 66 Ф.5: unit-тесты buyout_handler.
//
// Покрытые сценарии (без TEST_DATABASE_URL — pure-логика + property):
//   - TestCalcBuyoutCost_Determinism: cost(missionID) детерминирован,
//     не зависит от mission_id (новая фича — fixed-price).
//   - TestCalcBuyoutCost_PositiveOnDefault: дефолт > 0 (защита от
//     misconfiguration, R15).
//   - TestBuyoutBilling_Compatibility: *billingclient.Client реализует
//     BuyoutBilling (compile-time, дублирует _ = check в файле, но
//     даёт явный test-failure с чтением).

import (
	"testing"

	"pgregory.net/rapid"

	billingclient "oxsar/game-nova/internal/billing/client"
	originalien "oxsar/game-nova/internal/origin/alien"
)

// TestCalcBuyoutCost_Determinism — property (R4): для любого mission_id
// и любого Config.BuyoutBaseOxsars > 0 функция возвращает ровно
// BuyoutBaseOxsars (текущая реализация — fixed-price). Это закрепляет
// контракт: «cost не зависит от mission_id», который потребуется при
// future-расширении формулы (если ввести зависимость от tier — этот
// тест должен сломаться, привлекая внимание ревьюера).
func TestCalcBuyoutCost_Determinism(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		base := rapid.Int64Range(1, 1_000_000).Draw(t, "base_oxsars")
		mid1 := rapid.String().Draw(t, "mid1")
		mid2 := rapid.String().Draw(t, "mid2")
		cfg := originalien.DefaultConfig()
		cfg.BuyoutBaseOxsars = base

		c1 := originalien.CalcBuyoutCost(cfg, mid1)
		c2 := originalien.CalcBuyoutCost(cfg, mid2)
		if c1 != base || c2 != base {
			t.Fatalf("CalcBuyoutCost not equal to base: c1=%d c2=%d base=%d", c1, c2, base)
		}
	})
}

// TestCalcBuyoutCost_PositiveOnDefault — DefaultConfig() возвращает
// положительный BuyoutBaseOxsars. Защита от случайного 0/отрицательного
// в DefaultConfig (Buyout() вернёт ошибку, но эта проверка локализует
// баг к конфигу).
func TestCalcBuyoutCost_PositiveOnDefault(t *testing.T) {
	cfg := originalien.DefaultConfig()
	if cfg.BuyoutBaseOxsars <= 0 {
		t.Fatalf("BuyoutBaseOxsars must be positive, got %d", cfg.BuyoutBaseOxsars)
	}
	mid := "00000000-0000-0000-0000-000000000001"
	if got := originalien.CalcBuyoutCost(cfg, mid); got != cfg.BuyoutBaseOxsars {
		t.Fatalf("CalcBuyoutCost = %d, want %d", got, cfg.BuyoutBaseOxsars)
	}
}

// TestBuyoutBilling_Compatibility — compile-time check, что
// *billingclient.Client удовлетворяет интерфейсу BuyoutBilling.
//
// Дублирует `var _ BuyoutBilling = (*billingclient.Client)(nil)` в
// buyout_handler.go, но этот test даёт явное сообщение об ошибке если
// в будущем сигнатура Spend изменится — `go test` точечно укажет на
// этот файл, а не на буква-в-букву идентичный файл с компиляцией.
func TestBuyoutBilling_Compatibility(t *testing.T) {
	var _ originalien.BuyoutBilling = (*billingclient.Client)(nil)
}
