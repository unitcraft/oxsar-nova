package alien

import (
	"context"

	"oxsar/game-nova/internal/economy"
)

// Loader — интерфейс загрузки данных, нужных AlienAI.
//
// Helper-функции пакета чисто-функциональны; всё IO собрано здесь.
// Тесты подменяют Loader на in-memory stub. Боевая реализация —
// PgxLoader (Ф.3 плана 66, после плана 65 / R8/R9 паттернов).
type Loader interface {
	// LoadAttackCandidates возвращает список потенциальных целей
	// для findTarget. SQL-проверки активности/umode/recent-event
	// делаются на стороне loader'а; pure-функция PickAttackTarget
	// применяет финальный random + satellite-фильтр.
	LoadAttackCandidates(ctx context.Context, cfg Config) ([]TargetCandidate, error)

	// LoadCreditCandidates — то же для findCreditTarget.
	LoadCreditCandidates(ctx context.Context, cfg Config) ([]TargetCandidate, error)

	// LoadPlanetShips возвращает Fleet — флот на планете
	// (без UNIT_SOLAR_SATELLITE, как PHP loadPlanetShips:269).
	LoadPlanetShips(ctx context.Context, planetID string) ([]TargetUnit, error)

	// LoadUserResearches возвращает 10 ключевых тех-уровней игрока
	// (origin loadUserResearches:281).
	LoadUserResearches(ctx context.Context, userID string) (TechProfile, error)

	// LoadActiveAlienMissionsCount — сколько уже активных alien-миссий
	// в мире. Origin: count(*) WHERE mode IN
	// (FLY_UNKNOWN/HOLDING/ATTACK/HALT) AND processed=WAIT.
	// Используется в checkAlientNeeds для лимита FleetsNumber*.
	LoadActiveAlienMissionsCount(ctx context.Context) (int64, error)
}

// AlienResearchTechIDs — список techID, которые AI «узнаёт» о цели.
// Источник: AlienAI::loadUserResearches (PHP:280-281).
//
// Используется в SQL `IN (...)` запросе loader'а и тестах.
var AlienResearchTechIDs = []int{
	27,                       // UNIT_EXPO_TECH
	economy.IDTechGun,        // 15
	economy.IDTechShield,     // 16
	economy.IDTechShell,      // 17
	13,                       // UNIT_SPYWARE
	economy.IDTechBallistics, // 103
	economy.IDTechMasking,    // 104
	economy.IDTechLaser,      // 23
	economy.IDTechSilicon,    // 24 — UNIT_ION_TECH (см. internal/economy/ids.go)
	economy.IDTechHydrogen,   // 25 — UNIT_PLASMA_TECH
}

// PgxLoader реализующий Loader живёт в loader_pgx.go.
