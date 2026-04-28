// Package economy — production/consumption формулы для вселенной origin
// (план 64 Ф.4).
//
// Это надстройка над internal/economy/, которая в качестве коэффициентов
// читает bundle.Globals (origin-override). Modern-вселенные (uni01,
// uni02 и любые без override) используют ModernGlobals, и формулы ведут
// себя идентично internal/economy/formulas.go — это R0-инвариант
// (тестируется в TestOriginEconomyMatchesNovaEconomy).
//
// Когда использовать:
//   - В код-путях, которым нужно поведение origin-вселенной с её
//     коэффициентами (если в будущем они разойдутся с nova).
//   - В тестах, которые сравнивают origin-числа с live origin
//     (golden_*.json).
//
// Когда НЕ использовать:
//   - Modern-вселенные: им достаточно internal/economy/ напрямую,
//     поскольку nova-коэффициенты захардкожены и не зависят от bundle.
//
// Текущий статус (план 64 Ф.4): origin-формулы pixel-perfect совпадают
// с nova internal/economy/formulas.go (verify 2026-04-28 против live
// docker-mysql-1: METALMINE prod = floor(30*L*pow(1.1+tech*0.0006,L)),
// HYDROGEN_LAB prod включает (-0.002*temp + 1.28) — оба совпадают с
// existing nova-функциями). Поэтому функции этого пакета — тонкие
// обёртки, читающие коэффициенты из bundle. Future divergence — добавить
// тут отдельные origin-формулы без затрагивания internal/economy/.
//
// Закрывает D-022 (production_factor — параметризуется через
// bundle.Globals при необходимости), D-026 (источник истины формул —
// теперь bundle, не хардкод), D-029 (температурный множитель водорода
// уже работал в nova; теперь явно параметризован).
package economy
