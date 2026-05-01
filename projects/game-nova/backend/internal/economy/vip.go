// VIP-instant-старт: legacy `EventHandler::startConstructionEventVIP`
// (план 72.1.44 cross-cut). Платный мгновенный старт ожидающей задачи
// в очереди (Constructions/Research/Shipyard/Repair).
//
// Формулы 1:1 из legacy `Functions.inc.php` (строки 875-924):
//
//   getCreditImmStartConstructions(level) =
//     max(1, round(pow(level, level<30 ? 1.3 : level<35 ? 2 : 2.5) × 5,  -1))
//
//   getCreditImmStartResearch(level) =
//     max(1, round(pow(level, level<30 ? 1.3 : level<35 ? 2 : 2.5) × 10, -1))
//
//   getCreditImmStartShipyard(quantity) =
//     clamp(round(pow(quantity, 0.8), -1), 10, 100000)
//
//   getCreditImmStartRepair(quantity)      = getCreditImmStartShipyard
//   getCreditImmStartDisassemble(quantity) = getCreditImmStartShipyard
//
// `round(x, -1)` в PHP — округление до десятков (10's place). В Go
// эквивалент: round(x/10) × 10.

package economy

import "math"

// VIPCostConstruction возвращает credit-стоимость VIP-старта здания
// уровня `level`.
func VIPCostConstruction(level int) int64 {
	exp := vipExponent(level)
	raw := math.Pow(float64(level), exp) * 5
	c := roundToTens(raw)
	if c < 1 {
		c = 1
	}
	return int64(c)
}

// VIPCostResearch — credit-стоимость VIP-старта исследования.
func VIPCostResearch(level int) int64 {
	exp := vipExponent(level)
	raw := math.Pow(float64(level), exp) * 10
	c := roundToTens(raw)
	if c < 1 {
		c = 1
	}
	return int64(c)
}

// VIPCostShipyard — credit-стоимость VIP-старта shipyard-задачи на
// `quantity` юнитов. То же самое для repair/disassemble.
func VIPCostShipyard(quantity int64) int64 {
	if quantity <= 0 {
		return 10
	}
	raw := math.Pow(float64(quantity), 0.8)
	c := roundToTens(raw)
	if c < 10 {
		c = 10
	}
	if c > 100_000 {
		c = 100_000
	}
	return int64(c)
}

// vipExponent — legacy-формула «1.3 / 2 / 2.5» в зависимости от уровня.
func vipExponent(level int) float64 {
	switch {
	case level < 30:
		return 1.3
	case level < 35:
		return 2.0
	default:
		return 2.5
	}
}

// roundToTens — PHP-стиль round(x, -1) — до ближайших 10. PHP по
// умолчанию HALF_UP, эквивалент math.Round.
func roundToTens(x float64) float64 {
	return math.Round(x/10) * 10
}
