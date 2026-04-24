package alien

import "time"

// Константы, портированные из oxsar2 consts.php (строки 752–770).
//
// Семантика и значения 1-в-1 с legacy: менять их без ADR запрещено
// (§17 / §18 ТЗ, CLAUDE.md).
const (
	// AlienAttackInterval — минимальный промежуток между alien-событиями
	// для одного игрока. В legacy: ALIEN_ATTACK_INTERVAL = 6 дней.
	AlienAttackInterval = 6 * 24 * time.Hour

	// Четверг: ×5 спавнов, сила флота × rand(1.5, 2.0). В legacy —
	// ALIEN_NORMAL_FLEETS_NUMBER=50 и ALIEN_ATTACK_TIME_FLEETS_NUMBER=250
	// на весь мир; у нас в nova масштаб иной (Spawn берёт до 5 игроков
	// за тик), поэтому интерпретируем как множитель кандидатов.
	ThursdayCandidateMultiplier = 5

	ThursdayPowerMin = 1.5
	ThursdayPowerMax = 2.0

	// HOLDING lifecycle (legacy consts.php: ALIEN_HALTING_*).
	AlienHaltingMinTime     = 12 * time.Hour
	AlienHaltingMaxTime     = 24 * time.Hour
	AlienHaltingMaxRealTime = 15 * 24 * time.Hour // cap продления платежом

	// HOLDING payment: 2 часа за 50 кредитов.
	AlienHoldingPaySecondsPerCredit = 2 * 60 * 60.0 / 50.0 // = 144.0 сек/кредит
)
