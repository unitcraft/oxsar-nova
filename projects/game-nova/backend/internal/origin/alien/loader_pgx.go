package alien

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/event"
)

// PgxLoader — pgxpool-реализация Loader.
//
// Запросы — порт SQL из AlienAI.class.php:
//   - LoadAttackCandidates ↔ findTarget (PHP:336-369)
//   - LoadCreditCandidates ↔ findCreditTarget (PHP:299-334)
//   - LoadPlanetShips ↔ loadPlanetShips (PHP:266-276)
//   - LoadUserResearches ↔ loadUserResearches (PHP:278-297)
//   - LoadActiveAlienMissionsCount ↔ count(*) в checkAlientNeeds
//     (PHP:191)
//
// R10 (per-universe изоляция): в nova users/planets уже изолированы
// через FK на universe_id; loader не добавляет явный фильтр —
// caller передаёт `universeID`, и SQL фильтрует по нему.
//
// R0-исключение: loader работает одинаково для uni01/uni02/origin
// (см. doc.go).
type PgxLoader struct {
	Pool       *pgxpool.Pool
	UniverseID string // фильтр R10
}

// NewPgxLoader — конструктор.
func NewPgxLoader(pool *pgxpool.Pool, universeID string) *PgxLoader {
	return &PgxLoader{Pool: pool, UniverseID: universeID}
}

// LoadAttackCandidates — порт findTarget (AlienAI.class.php:336-369).
//
// Origin SQL (упрощённо):
//
//   SELECT u.userid, p.metal, p.silicon, p.hydrogen,
//          ship.planetid, sum(ship.quantity) as quantity
//   FROM user u JOIN planet p ON ... JOIN unit2shipyard ship ON ...
//   WHERE u.last > now() - 30min
//     AND u.umode = 0
//     AND u.u_count > 1000  -- FindTargetUserShipsMin
//     AND u.userid NOT IN (
//       SELECT userid FROM events WHERE mode IN
//         (FLY_UNKNOWN, ATTACK, HOLDING, HALT)
//         AND start > now() - 6 days
//     )
//   GROUP BY ship.planetid
//   HAVING sum(ship.quantity) > 100  -- FindTargetPlanetShipsMin
//
// Pure-функция PickAttackTarget применяет финальный random + satellite-
// фильтр. Поэтому loader возвращает ВСЕХ кандидатов, не LIMIT 1.
//
// Различие с origin:
//   - PHP `ORDER BY rand() LIMIT 1` — pgx-loader возвращает все,
//     pure-функция `PickAttackTarget` делает random выбор. Это
//     даёт большую тестируемость и детерминизм через rng.R.
//   - PHP `ship.unitid IN/NOT IN solar_satellite` — мы возвращаем
//     флаг `HasOnlySatellites` и pure-функция фильтрует.
func (p *PgxLoader) LoadAttackCandidates(ctx context.Context, cfg Config) ([]TargetCandidate, error) {
	const q = `
SELECT
  u.id::text                                       AS user_id,
  pl.id::text                                      AS planet_id,
  pl.galaxy, pl.system, pl.position,
  COALESCE(pl.metal, 0)::bigint                    AS metal,
  COALESCE(pl.silicon, 0)::bigint                  AS silicon,
  COALESCE(pl.hydrogen, 0)::bigint                 AS hydrogen,
  COALESCE(u.credit, 0)::bigint                    AS credit,
  COALESCE(uc.user_ships, 0)::bigint               AS user_ships,
  COALESCE(ps.planet_ships, 0)::bigint             AS planet_ships,
  COALESCE(ps.has_only_sats, false)                AS has_only_sats,
  EXTRACT(EPOCH FROM (now() - u.last_seen_at))::bigint
                                                   AS last_active_sec,
  u.umode                                          AS in_umode,
  EXISTS (
    SELECT 1 FROM events e
    JOIN planets p2 ON p2.id = e.planet_id AND p2.user_id = u.id
    WHERE e.kind = ANY($3::int[])
      AND e.fire_at > now() - $4::interval
  )                                                AS has_recent_alien
FROM users u
JOIN planets pl ON pl.user_id = u.id AND pl.destroyed_at IS NULL
                AND pl.is_moon = false
JOIN LATERAL (
  SELECT SUM(s.count) AS user_ships
  FROM ships s
  JOIN planets p2 ON p2.id = s.planet_id
  WHERE p2.user_id = u.id AND p2.destroyed_at IS NULL
) uc ON true
JOIN LATERAL (
  SELECT SUM(s.count) AS planet_ships,
         BOOL_AND(s.unit_id = $5)::bool AS has_only_sats
  FROM ships s
  WHERE s.planet_id = pl.id AND s.count > 0
) ps ON true
WHERE u.universe_id = $1::uuid
  AND u.banned_at IS NULL
  AND u.last_seen_at > now() - interval '30 minutes'
  AND u.umode = false
  AND COALESCE(uc.user_ships, 0) > $2
  AND COALESCE(ps.planet_ships, 0) > 0
ORDER BY u.id, pl.id
`
	kinds := []int32{
		int32(event.KindAlienFlyUnknown),
		int32(event.KindAlienHolding),
		int32(event.KindAlienAttack),
		int32(event.KindAlienHalt),
	}
	rows, err := p.Pool.Query(ctx, q,
		p.UniverseID,
		cfg.FindTargetUserShipsMin,
		kinds,
		cfg.AttackInterval,
		UnitSolarSatellite,
	)
	if err != nil {
		return nil, fmt.Errorf("alien.PgxLoader.LoadAttackCandidates: %w", err)
	}
	defer rows.Close()

	var out []TargetCandidate
	for rows.Next() {
		var c TargetCandidate
		if err := rows.Scan(
			&c.UserID, &c.PlanetID,
			&c.Galaxy, &c.System, &c.Position,
			&c.Metal, &c.Silicon, &c.Hydrogen,
			&c.Credit, &c.UserShipCount, &c.PlanetShipCount,
			&c.HasOnlySatellites, &c.LastActiveSeconds,
			&c.InUmode, &c.HasRecentAlienEvent,
		); err != nil {
			return nil, fmt.Errorf("alien.LoadAttackCandidates: scan: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// LoadCreditCandidates — порт findCreditTarget (AlienAI.class.php:299-334).
func (p *PgxLoader) LoadCreditCandidates(ctx context.Context, cfg Config) ([]TargetCandidate, error) {
	const q = `
SELECT
  u.id::text                                       AS user_id,
  pl.id::text                                      AS planet_id,
  pl.galaxy, pl.system, pl.position,
  COALESCE(pl.metal, 0)::bigint                    AS metal,
  COALESCE(pl.silicon, 0)::bigint                  AS silicon,
  COALESCE(pl.hydrogen, 0)::bigint                 AS hydrogen,
  COALESCE(u.credit, 0)::bigint                    AS credit,
  COALESCE(uc.user_ships, 0)::bigint               AS user_ships,
  COALESCE(ps.planet_ships, 0)::bigint             AS planet_ships,
  false                                            AS has_only_sats,
  EXTRACT(EPOCH FROM (now() - u.last_seen_at))::bigint
                                                   AS last_active_sec,
  u.umode                                          AS in_umode,
  EXISTS (
    SELECT 1 FROM events e
    JOIN planets p2 ON p2.id = e.planet_id AND p2.user_id = u.id
    WHERE e.kind = $5
      AND e.fire_at > now() - $6::interval
  )                                                AS has_recent_grab
FROM users u
JOIN planets pl ON pl.user_id = u.id AND pl.destroyed_at IS NULL
                AND pl.is_moon = false
JOIN LATERAL (
  SELECT SUM(s.count) AS user_ships
  FROM ships s JOIN planets p2 ON p2.id = s.planet_id
  WHERE p2.user_id = u.id AND p2.destroyed_at IS NULL
) uc ON true
JOIN LATERAL (
  SELECT SUM(s.count) AS planet_ships
  FROM ships s
  WHERE s.planet_id = pl.id AND s.count > 0
) ps ON true
WHERE u.universe_id = $1::uuid
  AND u.banned_at IS NULL
  AND u.last_seen_at > now() - interval '30 minutes'
  AND u.umode = false
  AND COALESCE(u.credit, 0) > $2
  AND COALESCE(uc.user_ships, 0) > $3
  AND COALESCE(ps.planet_ships, 0) > $4
ORDER BY u.id, pl.id
`
	rows, err := p.Pool.Query(ctx, q,
		p.UniverseID,
		cfg.GrabMinCredit,
		cfg.FindCreditTargetUserShipsMin,
		cfg.FindCreditTargetPlanetShipsMin,
		int32(event.KindAlienGrabCredit),
		cfg.GrabCreditInterval,
	)
	if err != nil {
		return nil, fmt.Errorf("alien.PgxLoader.LoadCreditCandidates: %w", err)
	}
	defer rows.Close()

	var out []TargetCandidate
	for rows.Next() {
		var c TargetCandidate
		if err := rows.Scan(
			&c.UserID, &c.PlanetID,
			&c.Galaxy, &c.System, &c.Position,
			&c.Metal, &c.Silicon, &c.Hydrogen,
			&c.Credit, &c.UserShipCount, &c.PlanetShipCount,
			&c.HasOnlySatellites, &c.LastActiveSeconds,
			&c.InUmode, &c.HasRecentGrabEvent,
		); err != nil {
			return nil, fmt.Errorf("alien.LoadCreditCandidates: scan: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// LoadPlanetShips — порт loadPlanetShips (AlienAI.class.php:266-276).
//
// Возвращает []TargetUnit БЕЗ ShipSpec.Attack/Shield/BasicMetal/Silicon —
// caller должен заполнить spec из catalog. Это разделение помогает
// тестам работать без полного catalog.
//
// PHP исключает UNIT_SOLAR_SATELLITE; мы тоже.
func (p *PgxLoader) LoadPlanetShips(ctx context.Context, planetID string) ([]TargetUnit, error) {
	const q = `
SELECT unit_id, count
FROM ships
WHERE planet_id = $1::uuid AND count > 0 AND unit_id <> $2
ORDER BY unit_id
`
	rows, err := p.Pool.Query(ctx, q, planetID, UnitSolarSatellite)
	if err != nil {
		return nil, fmt.Errorf("alien.LoadPlanetShips: %w", err)
	}
	defer rows.Close()

	var out []TargetUnit
	for rows.Next() {
		var unitID int
		var count int64
		if err := rows.Scan(&unitID, &count); err != nil {
			return nil, fmt.Errorf("alien.LoadPlanetShips: scan: %w", err)
		}
		out = append(out, TargetUnit{
			Spec:     ShipSpec{UnitID: unitID},
			Quantity: count,
		})
	}
	return out, rows.Err()
}

// LoadUserResearches — порт loadUserResearches (AlienAI.class.php:278-297).
//
// Возвращает map[techID]level для 10 ключевых tech'ов
// (см. AlienResearchTechIDs). Отсутствующие — 0.
func (p *PgxLoader) LoadUserResearches(ctx context.Context, userID string) (TechProfile, error) {
	out := make(TechProfile, len(AlienResearchTechIDs))
	for _, id := range AlienResearchTechIDs {
		out[id] = 0
	}
	rows, err := p.Pool.Query(ctx, `
SELECT unit_id, level
FROM research
WHERE user_id = $1::uuid AND unit_id = ANY($2::int[])
`, userID, AlienResearchTechIDs)
	if err != nil {
		return nil, fmt.Errorf("alien.LoadUserResearches: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var uid, lvl int
		if err := rows.Scan(&uid, &lvl); err != nil {
			return nil, fmt.Errorf("alien.LoadUserResearches: scan: %w", err)
		}
		out[uid] = lvl
	}
	return out, rows.Err()
}

// LoadActiveAlienMissionsCount — порт PHP:191
//
//   SELECT count(*) FROM events WHERE mode IN
//     (FLY_UNKNOWN, HOLDING, ATTACK, HALT) AND processed=WAIT
func (p *PgxLoader) LoadActiveAlienMissionsCount(ctx context.Context) (int64, error) {
	kinds := []int32{
		int32(event.KindAlienFlyUnknown),
		int32(event.KindAlienHolding),
		int32(event.KindAlienAttack),
		int32(event.KindAlienHalt),
	}
	var count int64
	err := p.Pool.QueryRow(ctx, `
SELECT count(*)::bigint FROM events e
JOIN planets p ON p.id = e.planet_id
JOIN users u ON u.id = p.user_id
WHERE e.kind = ANY($1::int[])
  AND e.state = 'wait'
  AND u.universe_id = $2::uuid
`, kinds, p.UniverseID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("alien.LoadActiveAlienMissionsCount: %w", err)
	}
	return count, nil
}

// Compile-time assertion: PgxLoader реализует Loader.
var _ Loader = (*PgxLoader)(nil)

// PgxLoaderFromExec — конструктор из repo.Exec (для consistency со
// Service-конструкторами в internal/alien/, internal/economy/ и пр.).
type pgxPoolGetter interface {
	Pool() *pgxpool.Pool
}

func PgxLoaderFromExec(db pgxPoolGetter, universeID string) *PgxLoader {
	return &PgxLoader{Pool: db.Pool(), UniverseID: universeID}
}

// Заглушка времени — переопределяется в тестах.
var nowFn = time.Now

// _ pgx — keep import (используется в loader_pgx_test.go).
var _ = pgx.ErrNoRows
