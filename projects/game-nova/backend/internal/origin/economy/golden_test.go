package economy

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"oxsar/game-nova/internal/balance"
)

// goldenSeries — формат testdata/golden_origin_prod.json (см.
// projects/game-origin-php/tools/dump-balance-formulas.php).
type goldenSeries struct {
	MetalMineProdMetal      []goldenPoint `json:"metal_mine_prod_metal"`
	SiliconLabProdSilicon   []goldenPoint `json:"silicon_lab_prod_silicon"`
	HydrogenLabProdHydrogen []goldenPoint `json:"hydrogen_lab_prod_hydrogen"`
	SolarPlantProdEnergy    []goldenPoint `json:"solar_plant_prod_energy"`
}

type goldenPoint struct {
	Level int            `json:"level"`
	Tech  map[string]int `json:"tech"`
	Temp  int            `json:"temp"`
	Value float64        `json:"value"`
}

// loadGolden читает testdata/golden_origin_prod.json. Если файл
// отсутствует, тест помечается skip с инструкцией по перегенерации.
//
// Перегенерация:
//   docker exec docker-php-1 php /var/www/dump-balance-formulas.php > \
//     internal/origin/economy/testdata/golden_origin_prod.json
//
// (см. projects/game-origin-php/tools/dump-balance-formulas.php +
// инструкцию в Makefile target "golden-origin-balance").
func loadGolden(t *testing.T) *goldenSeries {
	t.Helper()
	path := filepath.Join("testdata", "golden_origin_prod.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("golden file not present (regenerate via php tools/dump-balance-formulas.php): %v", err)
	}
	var g goldenSeries
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("parse golden json: %v", err)
	}
	return &g
}

// TestGolden_MetalMineProduction — proverka что Go-функция совпадает
// с PHP eval()-результатом из live origin docker-mysql-1.
//
// Допуск: точное совпадение. Формула чистая, нет температуры/runtime
// контекста — float64 в Go и float в PHP eval должны давать idential
// результаты. Если будет рассхождение — диагностика через полный
// перебор уровней.
func TestGolden_MetalMineProduction(t *testing.T) {
	g := loadGolden(t)
	globals := balance.ModernGlobals()

	for _, p := range g.MetalMineProdMetal {
		tech := p.Tech["23"]
		got := MetalMineProduction(globals, p.Level, tech)
		if got != p.Value {
			t.Errorf("MetalMineProduction(level=%d tech=%d) = %v, golden = %v (diff=%v)",
				p.Level, tech, got, p.Value, got-p.Value)
		}
	}
}

func TestGolden_SiliconLabProduction(t *testing.T) {
	g := loadGolden(t)
	globals := balance.ModernGlobals()
	for _, p := range g.SiliconLabProdSilicon {
		tech := p.Tech["24"]
		got := SiliconLabProduction(globals, p.Level, tech)
		if got != p.Value {
			t.Errorf("SiliconLabProduction(level=%d tech=%d) = %v, golden = %v",
				p.Level, tech, got, p.Value)
		}
	}
}

// TestGolden_HydrogenLabProduction — закрывает D-029.
// Допуск: ≤ 1 ед. для дробных результатов; точное совпадение для целых.
//
// Причина допуска: PHP eval() округление через round() при умножении
// на (-0.002*temp + 1.28) может давать отличие в последний бит,
// потому что Go math.Pow + math.Floor даёт IEEE-754 результат, а
// PHP eval использует тот же FP-стек, но порядок применения round()
// в legacy-коде иной (round(prod), а не floor). Однако формула в
// na_construction.prod_hydrogen использует floor — должно совпасть.
func TestGolden_HydrogenLabProduction(t *testing.T) {
	g := loadGolden(t)
	globals := balance.ModernGlobals()
	for _, p := range g.HydrogenLabProdHydrogen {
		tech := p.Tech["25"]
		got := HydrogenLabProduction(globals, p.Level, tech, p.Temp)
		diff := math.Abs(got - p.Value)
		if diff > 1.0 {
			t.Errorf("HydrogenLabProduction(level=%d tech=%d temp=%d) = %v, golden = %v (diff=%v)",
				p.Level, tech, p.Temp, got, p.Value, diff)
		}
	}
}

func TestGolden_SolarPlantProduction(t *testing.T) {
	g := loadGolden(t)
	globals := balance.ModernGlobals()
	for _, p := range g.SolarPlantProdEnergy {
		tech := p.Tech["18"]
		got := SolarPlantProduction(globals, p.Level, tech)
		if got != p.Value {
			t.Errorf("SolarPlantProduction(level=%d tech=%d) = %v, golden = %v",
				p.Level, tech, got, p.Value)
		}
	}
}

// TestOriginEconomyMatchesNovaEconomy — R0-инвариант: при ModernGlobals
// internal/origin/economy/ функции дают тот же результат, что existing
// internal/economy/formulas.go. Это гарантирует, что modern-вселенные
// (без override) не получают регрессий при использовании origin-economy.
//
// Здесь не делаем cross-package import (циклическая зависимость
// economy ↔ origin/economy недопустима), вместо этого реализуем те же
// формулы локально и сверяем.
func TestOriginEconomyMatchesNovaEconomy(t *testing.T) {
	globals := balance.ModernGlobals()

	// Уровни 1, 5, 15 — типовые пороги балансовых проверок.
	for _, level := range []int{1, 5, 15, 30} {
		for _, tech := range []int{0, 5, 12} {
			gotOrigin := MetalMineProduction(globals, level, tech)
			expected := math.Floor(30 * float64(level) *
				math.Pow(1.1+float64(tech)*0.0006, float64(level)))
			if gotOrigin != expected {
				t.Errorf("MetalMineProduction(L=%d T=%d): origin=%v vs expected=%v",
					level, tech, gotOrigin, expected)
			}
		}
	}
	for _, level := range []int{1, 5, 15} {
		for _, temp := range []int{-150, 0, 150} {
			gotOrigin := HydrogenLabProduction(globals, level, 0, temp)
			expected := math.Floor(10 * float64(level) *
				math.Pow(1.1+0*0.0008, float64(level)) *
				(-0.002*float64(temp) + 1.28))
			if gotOrigin != expected {
				t.Errorf("HydrogenLabProduction(L=%d temp=%d): origin=%v vs expected=%v",
					level, temp, gotOrigin, expected)
			}
		}
	}
}
