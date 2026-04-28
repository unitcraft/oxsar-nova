package alien

import (
	"oxsar/game-nova/pkg/rng"
)

// TargetCandidate — кандидат на роль цели alien-миссии.
// Заполняется repo-слоем (см. repo.go); pure-логика отбора —
// в PickAttackTarget / PickCreditTarget ниже.
//
// Семантически 1-в-1 с строкой результата SQL findTarget()
// (AlienAI.class.php:336-369).
type TargetCandidate struct {
	UserID            string
	PlanetID          string
	Galaxy            int
	System            int
	Position          int
	Metal             int64
	Silicon           int64
	Hydrogen          int64
	Credit            int64
	UserShipCount     int64 // u.u_count в PHP
	PlanetShipCount   int64 // sum(ship.quantity) per planet
	HasOnlySatellites bool  // ship.unitid = solar_satellite

	// LastActiveSeconds — секунд с last_seen у юзера. PHP проверяет
	// `u.last > unixtime() - 60*30` (активен последние 30 мин).
	LastActiveSeconds int64

	// InUmode — игрок в режиме отпуска (umode != 0).
	InUmode bool

	// HasRecentAlienEvent — у игрока есть alien-событие (FLY_UNKNOWN /
	// ATTACK / HOLDING / HALT) за последние AttackInterval (6 дней).
	HasRecentAlienEvent bool

	// HasRecentGrabEvent — у игрока есть GRAB_CREDIT за последние
	// GrabCreditInterval (10 дней).
	HasRecentGrabEvent bool
}

// PickAttackTarget — порт AlienAI::findTarget (AlienAI.class.php:336-369).
//
// Из переданного списка кандидатов отбирает подходящих и возвращает
// случайного. Возвращает nil если подходящих нет.
//
// Критерии (origin:336-369):
//   - last_seen < 30 мин назад
//   - !umode
//   - user_ship_count > 1000  (FindTargetUserShipsMin)
//   - planet_ship_count > 100 (FindTargetPlanetShipsMin)
//   - !has_recent_alien_event
//   - 10% (SolarSatelliteTargetChance) — берём планеты только с
//     solar_satellite; 90% — исключаем «satellite-only» планеты.
//
// caller передаёт уже отфильтрованный список (по lastActive/umode
// и т.п. — это проще для SQL); PickAttackTarget применяет финальный
// random выбор и satellite-фильтр.
func PickAttackTarget(candidates []TargetCandidate, cfg Config, r *rng.R) *TargetCandidate {
	if len(candidates) == 0 {
		return nil
	}
	useSatellites := r.IntN(100) < cfg.SolarSatelliteTargetChance

	filtered := make([]TargetCandidate, 0, len(candidates))
	for _, c := range candidates {
		if !attackTargetEligible(c, cfg) {
			continue
		}
		if useSatellites != c.HasOnlySatellites {
			continue
		}
		filtered = append(filtered, c)
	}
	if len(filtered) == 0 {
		// Fallback: relax satellite-фильтр (origin полагается на
		// LIMIT 1 и в редких случаях возвращает null — мы воспроизводим
		// эту вероятность отказа через len(filtered)==0).
		return nil
	}
	idx := r.IntN(len(filtered))
	t := filtered[idx]
	return &t
}

// PickCreditTarget — порт AlienAI::findCreditTarget
// (AlienAI.class.php:299-334).
//
// Критерии:
//   - active (last_seen < 30 мин)
//   - !umode
//   - credit > 100_000 (GrabMinCredit)
//   - user_ship_count > 300_000 (FindCreditTargetUserShipsMin)
//   - planet_ship_count > 10_000 (FindCreditTargetPlanetShipsMin)
//   - !has_recent_grab_event
func PickCreditTarget(candidates []TargetCandidate, cfg Config, r *rng.R) *TargetCandidate {
	if len(candidates) == 0 {
		return nil
	}
	filtered := make([]TargetCandidate, 0, len(candidates))
	for _, c := range candidates {
		if !creditTargetEligible(c, cfg) {
			continue
		}
		filtered = append(filtered, c)
	}
	if len(filtered) == 0 {
		return nil
	}
	idx := r.IntN(len(filtered))
	t := filtered[idx]
	return &t
}

func attackTargetEligible(c TargetCandidate, cfg Config) bool {
	switch {
	case c.InUmode:
		return false
	case c.LastActiveSeconds > 30*60:
		return false
	case c.UserShipCount <= cfg.FindTargetUserShipsMin:
		return false
	case c.PlanetShipCount <= cfg.FindTargetPlanetShipsMin:
		return false
	case c.HasRecentAlienEvent:
		return false
	}
	return true
}

func creditTargetEligible(c TargetCandidate, cfg Config) bool {
	switch {
	case c.InUmode:
		return false
	case c.LastActiveSeconds > 30*60:
		return false
	case c.Credit <= cfg.GrabMinCredit:
		return false
	case c.UserShipCount <= cfg.FindCreditTargetUserShipsMin:
		return false
	case c.PlanetShipCount <= cfg.FindCreditTargetPlanetShipsMin:
		return false
	case c.HasRecentGrabEvent:
		return false
	}
	return true
}
