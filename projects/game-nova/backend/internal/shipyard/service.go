// Package shipyard — постройка кораблей и обороны.
//
// В отличие от зданий/исследований, здесь единица производства — штука,
// а не уровень. Игрок ставит задачу «построить N cruiser'ов», задача
// выполняется последовательно (одна штука за per_unit_seconds).
//
// По ТЗ (§5.5): nano_factory делит время постройки корабля, robotic
// фабрика делит время ПОСТРОЙКИ ЗДАНИЙ (не кораблей, хотя в OGame есть
// нюанс с Shipyard-level влиянием). Мы следуем модели OGame classic.
package shipyard

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/economy"
	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/internal/planet"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/internal/requirements"
	"oxsar/game-nova/pkg/ids"
)

var (
	ErrUnknownUnit       = errors.New("shipyard: unknown unit")
	ErrNotEnoughRes      = errors.New("shipyard: not enough resources")
	ErrPlanetOwnership   = errors.New("shipyard: planet not owned by user")
	ErrNoShipyard        = errors.New("shipyard: shipyard required")
	ErrInvalidCount      = errors.New("shipyard: invalid count")
	ErrQueueItemNotFound = errors.New("shipyard: queue item not found")
	ErrAlreadyDone       = errors.New("shipyard: queue item already completed")
	// План 72.1.41: legacy `Shipyard.class.php` строки 390, 394 блокируют
	// при umode/observer.
	ErrUmodeBlocked      = errors.New("shipyard: blocked in vacation mode")
	ErrObserverBlocked   = errors.New("shipyard: blocked in observer mode")
	// План 72.1.44: VIP-instant.
	ErrNotEnoughCredit   = errors.New("shipyard: not enough credit for VIP start")
	ErrVIPAlreadyStarted = errors.New("shipyard: task already started, VIP not applicable")
)

type Service struct {
	db      repo.Exec
	planets *planet.Service
	catalog *config.Catalog
	reqs    *requirements.Checker
	gameSpd float64
}

func NewService(db repo.Exec, p *planet.Service, cat *config.Catalog, reqs *requirements.Checker, gameSpeed float64) *Service {
	if gameSpeed <= 0 {
		gameSpeed = 1
	}
	return &Service{db: db, planets: p, catalog: cat, reqs: reqs, gameSpd: gameSpeed}
}

type QueueItem struct {
	ID             string    `json:"id"`
	PlanetID       string    `json:"planet_id"`
	UnitID         int       `json:"unit_id"`
	Count          int64     `json:"count"`
	PerUnitSeconds int       `json:"per_unit_seconds"`
	StartAt        time.Time `json:"start_at"`
	EndAt          time.Time `json:"end_at"`
	Status         string    `json:"status"`
}

// Enqueue ставит задачу «построить count юнитов". Unit может быть как
// корабль, так и оборона — разделяем по наличию в каталоге.
func (s *Service) Enqueue(ctx context.Context, userID, planetID string, unitID int, count int64) (QueueItem, error) {
	if count <= 0 {
		return QueueItem{}, ErrInvalidCount
	}
	key, costPerUnit, isDefense, ok := s.lookupShipOrDefense(unitID)
	if !ok {
		return QueueItem{}, ErrUnknownUnit
	}

	p, err := s.planets.Get(ctx, planetID)
	if err != nil {
		return QueueItem{}, err
	}
	if p.UserID != userID {
		return QueueItem{}, ErrPlanetOwnership
	}

	// План 72.1.41: блок umode/observer (legacy Shipyard.class.php:390, 394).
	var umode, isObs bool
	if err := s.db.Pool().QueryRow(ctx,
		`SELECT umode, is_observer FROM users WHERE id = $1`, userID,
	).Scan(&umode, &isObs); err != nil {
		return QueueItem{}, fmt.Errorf("read user state: %w", err)
	}
	if umode {
		return QueueItem{}, ErrUmodeBlocked
	}
	if isObs {
		return QueueItem{}, ErrObserverBlocked
	}

	var item QueueItem
	err = s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. Есть ли верфь.
		var shipyardLvl int
		err := tx.QueryRow(ctx,
			`SELECT level FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
			planetID, s.catalog.Buildings.Buildings["shipyard"].ID,
		).Scan(&shipyardLvl)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("shipyard level: %w", err)
		}
		if shipyardLvl < 1 {
			return ErrNoShipyard
		}

		// 2. Зависимости юнита.
		if err := s.reqs.Check(ctx, tx, key, userID, planetID); err != nil {
			return err
		}

		// 2b. План 72.1.41: capacity-check для shield/rocket юнитов
		// (legacy `Shipyard` строки 41-51, 470-490).
		if err := s.checkCapacity(ctx, tx, userID, planetID, unitID, count); err != nil {
			return err
		}

		// 3. Стоимость × count. Масштаб ресурсов — int64.
		totalMetal := costPerUnit.Metal * count
		totalSilicon := costPerUnit.Silicon * count
		totalHydrogen := costPerUnit.Hydrogen * count
		if int64(p.Metal) < totalMetal || int64(p.Silicon) < totalSilicon || int64(p.Hydrogen) < totalHydrogen {
			return ErrNotEnoughRes
		}
		if _, err := tx.Exec(ctx, `
			UPDATE planets SET metal = metal - $1, silicon = silicon - $2, hydrogen = hydrogen - $3
			WHERE id = $4
		`, totalMetal, totalSilicon, totalHydrogen, planetID); err != nil {
			return fmt.Errorf("charge: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO res_log (user_id, planet_id, reason, delta_metal, delta_silicon, delta_hydrogen)
			VALUES ($1, $2, 'fleet_cost', $3, $4, $5)
		`, userID, planetID, -totalMetal, -totalSilicon, -totalHydrogen); err != nil {
			return fmt.Errorf("res_log: %w", err)
		}

		// 4. Per-unit time: базовая формула economy.BuildDuration,
		//    но учитываем только shipyard (и nano, когда он появится).
		perUnit := economy.BuildDuration(1, economy.Cost{
			Metal: costPerUnit.Metal, Silicon: costPerUnit.Silicon, Hydrogen: costPerUnit.Hydrogen,
		}, shipyardLvl, 0, s.gameSpd)
		perUnitSec := int(math.Max(1, math.Round(perUnit.Seconds())))
		totalDur := time.Duration(perUnitSec) * time.Duration(count) * time.Second
		start := time.Now().UTC()
		end := start.Add(totalDur)

		id := ids.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO shipyard_queue (id, planet_id, unit_id, count, per_unit_seconds,
			                            start_at, end_at, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'running')
		`, id, planetID, unitID, count, perUnitSec, start, end); err != nil {
			return fmt.Errorf("insert queue: %w", err)
		}

		// 5. Событие завершения постройки.
		kind := int(event.KindBuildFleet)
		if isDefense {
			kind = int(event.KindBuildDefense)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO events (id, user_id, planet_id, kind, state, fire_at, payload)
			VALUES ($1, $2, $3, $4, 'wait', $5, $6)
		`, ids.New(), userID, planetID, kind, end,
			fmt.Sprintf(`{"queue_id":"%s","unit_id":%d,"count":%d,"is_defense":%t}`,
				id, unitID, count, isDefense),
		); err != nil {
			return fmt.Errorf("insert event: %w", err)
		}

		item = QueueItem{
			ID: id, PlanetID: planetID, UnitID: unitID, Count: count,
			PerUnitSeconds: perUnitSec, StartAt: start, EndAt: end, Status: "running",
		}
		return nil
	})
	return item, err
}

// List возвращает активные задания на планете.
func (s *Service) List(ctx context.Context, planetID string) ([]QueueItem, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, planet_id, unit_id, count, per_unit_seconds, start_at, end_at, status
		FROM shipyard_queue
		WHERE planet_id=$1 AND status IN ('queued','running') AND end_at > NOW()
		ORDER BY start_at
	`, planetID)
	if err != nil {
		return nil, fmt.Errorf("list queue: %w", err)
	}
	defer rows.Close()
	var out []QueueItem
	for rows.Next() {
		var q QueueItem
		if err := rows.Scan(&q.ID, &q.PlanetID, &q.UnitID, &q.Count,
			&q.PerUnitSeconds, &q.StartAt, &q.EndAt, &q.Status); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

// Inventory возвращает корабли и оборону на планете.
func (s *Service) Inventory(ctx context.Context, planetID string) (ships, defense map[int]int64, err error) {
	ships = map[int]int64{}
	defense = map[int]int64{}
	rows, err := s.db.Pool().Query(ctx,
		`SELECT unit_id, count FROM ships WHERE planet_id = $1`, planetID)
	if err != nil {
		return nil, nil, fmt.Errorf("ships: %w", err)
	}
	for rows.Next() {
		var id int
		var c int64
		if err := rows.Scan(&id, &c); err != nil {
			rows.Close()
			return nil, nil, err
		}
		ships[id] = c
	}
	rows.Close()

	rows, err = s.db.Pool().Query(ctx,
		`SELECT unit_id, count FROM defense WHERE planet_id = $1`, planetID)
	if err != nil {
		return nil, nil, fmt.Errorf("defense: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var c int64
		if err := rows.Scan(&id, &c); err != nil {
			return nil, nil, err
		}
		defense[id] = c
	}
	return ships, defense, rows.Err()
}

// StartVIP — план 72.1.44 cross-cut. Legacy
// `EventHandler::startConstructionEventVIP` для UNIT_TYPE_FLEET/DEFENSE:
// мгновенный старт ожидающей задачи в shipyard_queue за credits.
//
// cost = economy.VIPCostShipyard(count).
func (s *Service) StartVIP(ctx context.Context, userID, planetID, queueID string) (QueueItem, error) {
	var out QueueItem
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var (
			unitID         int
			count          int64
			perUnitSeconds int
			startAt        time.Time
			endAt          time.Time
			ownerID        string
			status         string
		)
		err := tx.QueryRow(ctx, `
			SELECT sq.unit_id, sq.count, sq.per_unit_seconds, sq.start_at, sq.end_at, p.user_id, sq.status
			FROM shipyard_queue sq
			JOIN planets p ON p.id = sq.planet_id
			WHERE sq.id = $1 AND sq.planet_id = $2 AND sq.status IN ('queued','running')
			FOR UPDATE
		`, queueID, planetID).Scan(&unitID, &count, &perUnitSeconds, &startAt, &endAt, &ownerID, &status)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrQueueItemNotFound
			}
			return fmt.Errorf("select queue: %w", err)
		}
		if ownerID != userID {
			return ErrPlanetOwnership
		}
		if !startAt.After(time.Now().UTC().Add(5 * time.Second)) {
			return ErrVIPAlreadyStarted
		}

		cost := economy.VIPCostShipyard(count)

		var credit int64
		if err := tx.QueryRow(ctx,
			`SELECT credit FROM users WHERE id=$1 FOR UPDATE`, userID,
		).Scan(&credit); err != nil {
			return fmt.Errorf("select credit: %w", err)
		}
		if credit < cost {
			return ErrNotEnoughCredit
		}

		now := time.Now().UTC()
		duration := endAt.Sub(startAt)
		newEnd := now.Add(duration)

		if _, err := tx.Exec(ctx,
			`UPDATE users SET credit = credit - $1 WHERE id = $2`,
			cost, userID,
		); err != nil {
			return fmt.Errorf("debit credit: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE shipyard_queue SET start_at=$1, end_at=$2, status='running'
			WHERE id=$3
		`, now, newEnd, queueID); err != nil {
			return fmt.Errorf("update queue: %w", err)
		}
		// Shipyard event может быть KindBuildFleet (4) или KindBuildDefense (5);
		// меняем оба варианта.
		if _, err := tx.Exec(ctx, `
			UPDATE events SET fire_at=$1
			WHERE kind IN (4, 5) AND state='wait' AND user_id=$2
			  AND payload @> jsonb_build_object('queue_id', $3::text)
		`, newEnd, userID, queueID); err != nil {
			return fmt.Errorf("update event: %w", err)
		}

		out = QueueItem{
			ID: queueID, PlanetID: planetID, UnitID: unitID, Count: count,
			PerUnitSeconds: perUnitSeconds,
			StartAt: now, EndAt: newEnd, Status: "running",
		}
		return nil
	})
	return out, err
}

// Cancel отменяет задание очереди верфи и возвращает ресурсы на планету.
func (s *Service) Cancel(ctx context.Context, userID, planetID, queueID string) error {
	p, err := s.planets.Get(ctx, planetID)
	if err != nil {
		return err
	}
	if p.UserID != userID {
		return ErrPlanetOwnership
	}

	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var unitID int
		var count int64
		var status string
		err := tx.QueryRow(ctx,
			`SELECT unit_id, count, status FROM shipyard_queue WHERE id=$1 AND planet_id=$2`,
			queueID, planetID,
		).Scan(&unitID, &count, &status)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrQueueItemNotFound
		}
		if err != nil {
			return fmt.Errorf("select queue: %w", err)
		}
		if status == "done" {
			return ErrAlreadyDone
		}

		_, costPerUnit, _, ok := s.lookupShipOrDefense(unitID)
		if !ok {
			return ErrUnknownUnit
		}

		// Вернуть ресурсы (100%).
		refundMetal := costPerUnit.Metal * count
		refundSilicon := costPerUnit.Silicon * count
		refundHydrogen := costPerUnit.Hydrogen * count
		if _, err := tx.Exec(ctx, `
			UPDATE planets SET metal = metal + $1, silicon = silicon + $2, hydrogen = hydrogen + $3
			WHERE id = $4
		`, refundMetal, refundSilicon, refundHydrogen, planetID); err != nil {
			return fmt.Errorf("refund: %w", err)
		}

		// Удалить задание и связанное событие.
		if _, err := tx.Exec(ctx,
			`DELETE FROM shipyard_queue WHERE id=$1`, queueID,
		); err != nil {
			return fmt.Errorf("delete queue: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`DELETE FROM events WHERE planet_id=$1 AND state='wait' AND payload::jsonb->>'queue_id'=$2`,
			planetID, queueID,
		); err != nil {
			return fmt.Errorf("delete event: %w", err)
		}
		return nil
	})
}

// UnitCost — стоимость постройки одного корабля/обороны.
type UnitCost struct {
	Metal    int64 `json:"metal"`
	Silicon  int64 `json:"silicon"`
	Hydrogen int64 `json:"hydrogen"`
}

// CostsMap возвращает per-unit стоимость для всех ships и defense.
// Pixel-perfect клон legacy (план 72.1 ч.20.3).
func (s *Service) CostsMap() (ships, defense map[int]UnitCost) {
	ships = make(map[int]UnitCost, len(s.catalog.Ships.Ships))
	for _, spec := range s.catalog.Ships.Ships {
		ships[spec.ID] = UnitCost{
			Metal: spec.Cost.Metal, Silicon: spec.Cost.Silicon, Hydrogen: spec.Cost.Hydrogen,
		}
	}
	defense = make(map[int]UnitCost, len(s.catalog.Defense.Defense))
	for _, spec := range s.catalog.Defense.Defense {
		defense[spec.ID] = UnitCost{
			Metal: spec.Cost.Metal, Silicon: spec.Cost.Silicon, Hydrogen: spec.Cost.Hydrogen,
		}
	}
	return ships, defense
}

// SecondsMap возвращает per-unit время постройки для всех ships и defense
// на данной планете (с учётом shipyard / nano_factory уровней).
func (s *Service) SecondsMap(ctx context.Context, planetID string) (ships, defense map[int]int, err error) {
	var shipyardLvl int
	if shSpec, ok := s.catalog.Buildings.Buildings["shipyard"]; ok {
		_ = s.db.Pool().QueryRow(ctx,
			`SELECT level FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
			planetID, shSpec.ID,
		).Scan(&shipyardLvl)
	}
	var nanoLvl int
	if nanoSpec, ok := s.catalog.Buildings.Buildings["nano_factory"]; ok {
		_ = s.db.Pool().QueryRow(ctx,
			`SELECT level FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
			planetID, nanoSpec.ID,
		).Scan(&nanoLvl)
	}
	ships = make(map[int]int, len(s.catalog.Ships.Ships))
	for _, spec := range s.catalog.Ships.Ships {
		dur := economy.BuildDuration(1, economy.Cost{
			Metal: spec.Cost.Metal, Silicon: spec.Cost.Silicon, Hydrogen: spec.Cost.Hydrogen,
		}, shipyardLvl, nanoLvl, s.gameSpd)
		secs := int(math.Max(1, math.Round(dur.Seconds())))
		ships[spec.ID] = secs
	}
	defense = make(map[int]int, len(s.catalog.Defense.Defense))
	for _, spec := range s.catalog.Defense.Defense {
		dur := economy.BuildDuration(1, economy.Cost{
			Metal: spec.Cost.Metal, Silicon: spec.Cost.Silicon, Hydrogen: spec.Cost.Hydrogen,
		}, shipyardLvl, nanoLvl, s.gameSpd)
		secs := int(math.Max(1, math.Round(dur.Seconds())))
		defense[spec.ID] = secs
	}
	return ships, defense, nil
}

func (s *Service) lookupShipOrDefense(unitID int) (key string, cost config.ResCost, isDefense bool, ok bool) {
	for k, spec := range s.catalog.Ships.Ships {
		if spec.ID == unitID {
			return k, spec.Cost, false, true
		}
	}
	for k, spec := range s.catalog.Defense.Defense {
		if spec.ID == unitID {
			return k, spec.Cost, true, true
		}
	}
	return "", config.ResCost{}, false, false
}
