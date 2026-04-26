// Package rng содержит детерминированный ГПСЧ для боевой системы.
//
// Требование §5.7 ТЗ: один и тот же seed + один и тот же BattleInput
// должны давать бит-в-бит идентичный отчёт. Поэтому НЕ используем
// math/rand с глобальным state — вместо этого свой независимый
// xorshift64*, быстрый и воспроизводимый.
//
// Алгоритм совместим по семантике с Java-движком oxsar2-java
// (см. ADR-0002): одинаковое распределение на равных seed'ах.
// Точная бит-совместимость с java.util.Random достижима отдельным
// адаптером JavaRandom — добавим, когда запустим cross-verification
// с Assault.jar (см. §14.4 ТЗ).
package rng

// R — детерминированный RNG. Не потокобезопасен. Если нужно несколько
// потоков, используй несколько независимых R-объектов с разным seed.
type R struct {
	state uint64
}

// New возвращает генератор, инициализированный seed'ом. Нулевой seed
// недопустим (вырождение xorshift), подменяем на константу.
func New(seed uint64) *R {
	if seed == 0 {
		seed = 0x9E3779B97F4A7C15 // golden ratio, ненулевая константа
	}
	return &R{state: seed}
}

// Uint64 возвращает следующее псевдослучайное 64-битное число.
func (r *R) Uint64() uint64 {
	x := r.state
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	r.state = x
	return x * 0x2545F4914F6CDD1D
}

// IntN возвращает целое в [0, n). Для n <= 0 возвращает 0.
func (r *R) IntN(n int) int {
	if n <= 0 {
		return 0
	}
	return int(r.Uint64() % uint64(n))
}

// Float64 возвращает число в [0, 1).
func (r *R) Float64() float64 {
	return float64(r.Uint64()>>11) / (1 << 53)
}

// Roll возвращает true с вероятностью p (0..1).
func (r *R) Roll(p float64) bool {
	if p <= 0 {
		return false
	}
	if p >= 1 {
		return true
	}
	return r.Float64() < p
}
