// Package building управляет очередью строительства зданий.
//
// По ТЗ (§5.3): один слот очереди на планете + один на луне.
// Стоимость/время — из configs/buildings.yml. Отмена очереди:
// в первые 15 сек возврат 100%, далее — до 95% (EV_ABORT_MAX_BUILD_PERCENT).
package building

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/economy"
	"oxsar/game-nova/internal/planet"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/internal/requirements"
	"oxsar/game-nova/pkg/ids"
)

var (
	ErrQueueBusy         = errors.New("building: queue busy")
	ErrNotEnoughRes      = errors.New("building: not enough resources")
	ErrUnknownUnit       = errors.New("building: unknown unit")
	ErrPlanetOwnership   = errors.New("building: planet does not belong to user")
	ErrQueueItemNotFound = errors.New("building: queue item not found")
	ErrMoonOnly          = errors.New("building: this building is only available on moons")
	ErrPlanetOnly        = errors.New("building: this building is not available on moons")
	ErrMaxLevelReached   = errors.New("building: max level reached")
	ErrFieldsExhausted   = errors.New("building: no free fields on planet")
	// План 72.1.33: попытка demolish здания на уровне 0.
	ErrLevelZero         = errors.New("building: level is zero, nothing to demolish")
	// План 72.1.40: legacy `Constructions::index` строки 22, 247 блокируют
	// при umode; строки 64-66, 386 — при observer.
	ErrUmodeBlocked      = errors.New("building: blocked in vacation mode")
	ErrObserverBlocked   = errors.New("building: blocked in observer mode")
	// План 72.1.44: VIP-instant errors.
	ErrNotEnoughCredit   = errors.New("building: not enough credit for VIP start")
	ErrVIPAlreadyStarted = errors.New("building: task already started, VIP not applicable")
)

type Service struct {
	db      repo.Exec
	planets *planet.Service
	catalog *config.Catalog
	reqs    *requirements.Checker
	gameSpd float64
}

func NewService(db repo.Exec, planets *planet.Service, cat *config.Catalog, reqs *requirements.Checker, gameSpeed float64) *Service {
	if gameSpeed <= 0 {
		gameSpeed = 1
	}
	return &Service{db: db, planets: planets, catalog: cat, reqs: reqs, gameSpd: gameSpeed}
}

// QueueItem — задача в очереди строительства.
type QueueItem struct {
	ID          string    `json:"id"`
	PlanetID    string    `json:"planet_id"`
	UnitID      int       `json:"unit_id"`
	TargetLevel int       `json:"target_level"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	Status      string    `json:"status"`
}

// Enqueue ставит здание в очередь. Возвращает созданный QueueItem.
// Вся операция атомарна: списание ресурсов + создание задачи + событие
// завершения происходят в одной транзакции.
func (s *Service) Enqueue(ctx context.Context, userID, planetID string, unitID int) (QueueItem, error) {
	key, spec, ok := s.lookupBuilding(unitID)
	if !ok {
		return QueueItem{}, ErrUnknownUnit
	}

	// Сначала догоняем тик, чтобы списание шло от актуального баланса.
	p, err := s.planets.Get(ctx, planetID)
	if err != nil {
		return QueueItem{}, err
	}
	if p.UserID != userID {
		return QueueItem{}, ErrPlanetOwnership
	}
	if spec.MoonOnly && !p.IsMoon {
		return QueueItem{}, ErrMoonOnly
	}
	if !spec.MoonOnly && p.IsMoon {
		return QueueItem{}, ErrPlanetOnly
	}

	// План 72.1.40: блокировка umode/observer (legacy строки 22,247,
	// 64-66, 386).
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
		// Проверка: нет активной очереди.
		var busy int
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(*) FROM construction_queue
			WHERE planet_id = $1 AND status IN ('queued','running') AND end_at > NOW()
		`, p.ID).Scan(&busy); err != nil {
			return fmt.Errorf("check queue: %w", err)
		}
		if busy > 0 {
			return ErrQueueBusy
		}

		// Проверка зависимостей (если в requirements.yml есть запись
		// для этого здания — проверяем). Пустые зависимости пропускаются.
		if err := s.reqs.Check(ctx, tx, key, userID, p.ID); err != nil {
			return err
		}

		// Текущий уровень + стоимость следующего.
		curLevel, err := currentLevel(ctx, tx, p.ID, unitID)
		if err != nil {
			return err
		}
		targetLevel := curLevel + 1
		if spec.MaxLevel > 0 && targetLevel > spec.MaxLevel {
			return ErrMaxLevelReached
		}

		// План 23: проверка лимита полей. Новое здание (curLevel==0)
		// занимает поле; апгрейд существующего — нет. used_fields
		// обновляется воркером при завершении постройки (не здесь),
		// но для резервирования слота считаем pending-билды тоже.
		if curLevel == 0 {
			bm, err := buildingLevelsTx(ctx, tx, p.ID)
			if err != nil {
				return err
			}
			maxF := planet.MaxFields(&p, bm, planet.DefaultFieldConsts)
			if p.UsedFields+1 > maxF {
				return ErrFieldsExhausted
			}
		}
		cost := economy.CostForLevel(economy.Cost{
			Metal:    spec.CostBase.Metal,
			Silicon:  spec.CostBase.Silicon,
			Hydrogen: spec.CostBase.Hydrogen,
		}, spec.CostFactor, targetLevel)

		if int64(p.Metal) < cost.Metal || int64(p.Silicon) < cost.Silicon || int64(p.Hydrogen) < cost.Hydrogen {
			return ErrNotEnoughRes
		}

		// Снимаем ресурсы.
		if _, err := tx.Exec(ctx, `
			UPDATE planets
			SET metal = metal - $1, silicon = silicon - $2, hydrogen = hydrogen - $3
			WHERE id = $4
		`, cost.Metal, cost.Silicon, cost.Hydrogen, p.ID); err != nil {
			return fmt.Errorf("charge resources: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO res_log (user_id, planet_id, reason, delta_metal, delta_silicon, delta_hydrogen)
			VALUES ($1, $2, 'build', $3, $4, $5)
		`, userID, p.ID, -cost.Metal, -cost.Silicon, -cost.Hydrogen); err != nil {
			return fmt.Errorf("res_log: %w", err)
		}

		robo, err := currentLevel(ctx, tx, p.ID, s.catalog.Buildings.Buildings["robotic_factory"].ID)
		if err != nil {
			return err
		}
		var nano int
		if nanoSpec, ok := s.catalog.Buildings.Buildings["nano_factory"]; ok {
			nano, _ = currentLevel(ctx, tx, p.ID, nanoSpec.ID)
		}

		start := time.Now().UTC()
		dur := economy.BuildDuration(spec.TimeBaseSeconds, cost, robo, nano, s.gameSpd)
		end := start.Add(dur)

		id := ids.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO construction_queue (id, planet_id, unit_id, unit_type, target_level,
			                                start_at, end_at, cost_metal, cost_silicon, cost_hydrogen, status)
			VALUES ($1, $2, $3, 'building', $4, $5, $6, $7, $8, $9, 'running')
		`, id, p.ID, unitID, targetLevel, start, end, cost.Metal, cost.Silicon, cost.Hydrogen); err != nil {
			return fmt.Errorf("insert queue: %w", err)
		}

		// Событие завершения — воркер подхватит и применит эффект.
		if _, err := tx.Exec(ctx, `
			INSERT INTO events (id, user_id, planet_id, kind, state, fire_at, payload)
			VALUES ($1, $2, $3, 1, 'wait', $4, $5)
		`, ids.New(), userID, p.ID, end,
			fmt.Sprintf(`{"queue_id":"%s","unit_id":%d,"target_level":%d}`, id, unitID, targetLevel)); err != nil {
			return fmt.Errorf("insert event: %w", err)
		}

		item = QueueItem{
			ID: id, PlanetID: p.ID, UnitID: unitID, TargetLevel: targetLevel,
			StartAt: start, EndAt: end, Status: "running",
		}
		return nil
	})
	return item, err
}

// StartVIP — план 72.1.44 cross-cut. Legacy
// `EventHandler::startConstructionEventVIP`: мгновенный старт уже
// поставленной в очередь задачи за credits.
//
// Действия:
//  1. Найти queue.id, status='running'/'queued', start_at > now+5s.
//  2. cost = economy.VIPCostConstruction(target_level).
//  3. SELECT users.credit FOR UPDATE; if credit < cost → ErrNotEnoughCredit.
//  4. UPDATE users.credit -= cost.
//  5. UPDATE construction_queue.start_at = now, end_at = now + (end-start).
//  6. UPDATE events.fire_at = end_at; pending дальние задачи в очереди
//     планеты сдвигаются (legacy строки 1973-2010).
//
// Возвращает обновлённый QueueItem.
func (s *Service) StartVIP(ctx context.Context, userID, queueID string) (QueueItem, error) {
	var out QueueItem
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var (
			planetID    string
			targetLevel int
			startAt     time.Time
			endAt       time.Time
			ownerID     string
			status      string
		)
		err := tx.QueryRow(ctx, `
			SELECT cq.planet_id, cq.target_level, cq.start_at, cq.end_at, p.user_id, cq.status
			FROM construction_queue cq
			JOIN planets p ON p.id = cq.planet_id
			WHERE cq.id = $1 AND cq.unit_type = 'building'
			  AND cq.status IN ('queued','running')
			FOR UPDATE
		`, queueID).Scan(&planetID, &targetLevel, &startAt, &endAt, &ownerID, &status)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrQueueItemNotFound
			}
			return fmt.Errorf("select queue: %w", err)
		}
		if ownerID != userID {
			return ErrPlanetOwnership
		}
		// VIP применяется только к ещё не стартовавшим заданиям
		// (legacy: `start > time()+5`).
		if !startAt.After(time.Now().UTC().Add(5 * time.Second)) {
			return ErrVIPAlreadyStarted
		}

		cost := economy.VIPCostConstruction(targetLevel)

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
			UPDATE construction_queue SET start_at=$1, end_at=$2, status='running'
			WHERE id=$3
		`, now, newEnd, queueID); err != nil {
			return fmt.Errorf("update queue: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE events SET fire_at=$1
			WHERE kind=1 AND state='wait' AND user_id=$2
			  AND payload @> jsonb_build_object('queue_id', $3::text)
		`, newEnd, userID, queueID); err != nil {
			return fmt.Errorf("update event: %w", err)
		}

		out = QueueItem{
			ID: queueID, PlanetID: planetID, UnitID: 0, TargetLevel: targetLevel,
			StartAt: now, EndAt: newEnd, Status: "running",
		}
		return nil
	})
	return out, err
}

// EnqueueDemolish ставит здание в очередь на снос (план 72.1.33,
// legacy `BuildingInfo::DEMOLISH_NOW` + `EventHandler::demolish`).
//
// Семантика:
//   - target_level = curLevel - 1 (только на 1 уровень за раз).
//   - cost = (1 / spec.Demolish) × cost_at_current_level.
//   - duration = build duration уровня curLevel × 0.5 (стандарт OGame).
//   - event Kind=2 (KindDemolishConstruction).
//
// Ошибки:
//   - ErrQueueBusy — уже есть активная стройка/снос.
//   - ErrNotEnoughRes — не хватает ресурсов на cost demolish.
//   - ErrUnknownUnit — id не существует или demolish не задан.
//   - ErrLevelZero — здание уже на 0.
func (s *Service) EnqueueDemolish(ctx context.Context, userID, planetID string, unitID int) (QueueItem, error) {
	_, spec, ok := s.lookupBuilding(unitID)
	if !ok {
		return QueueItem{}, ErrUnknownUnit
	}
	if spec.Demolish <= 0 {
		return QueueItem{}, ErrUnknownUnit
	}

	p, err := s.planets.Get(ctx, planetID)
	if err != nil {
		return QueueItem{}, err
	}
	if p.UserID != userID {
		return QueueItem{}, ErrPlanetOwnership
	}

	// План 72.1.40: blocked при umode/observer (то же что Enqueue).
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
		// Только одна задача на планету.
		var busy int
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(*) FROM construction_queue
			WHERE planet_id = $1 AND status IN ('queued','running') AND end_at > NOW()
		`, p.ID).Scan(&busy); err != nil {
			return fmt.Errorf("check queue: %w", err)
		}
		if busy > 0 {
			return ErrQueueBusy
		}

		curLevel, err := currentLevel(ctx, tx, p.ID, unitID)
		if err != nil {
			return err
		}
		if curLevel <= 0 {
			return ErrLevelZero
		}
		targetLevel := curLevel - 1

		// Cost demolish = (1/factor) × cost_at_current_level. Legacy
		// `parseChargeFormula` возвращает basic × factor^(level-1) — у нас
		// economy.CostForLevel делает то же самое.
		baseCost := economy.CostForLevel(economy.Cost{
			Metal:    spec.CostBase.Metal,
			Silicon:  spec.CostBase.Silicon,
			Hydrogen: spec.CostBase.Hydrogen,
		}, spec.CostFactor, curLevel)
		inv := 1.0 / spec.Demolish
		demoCost := economy.Cost{
			Metal:    int64(float64(baseCost.Metal) * inv),
			Silicon:  int64(float64(baseCost.Silicon) * inv),
			Hydrogen: int64(float64(baseCost.Hydrogen) * inv),
		}

		if int64(p.Metal) < demoCost.Metal || int64(p.Silicon) < demoCost.Silicon || int64(p.Hydrogen) < demoCost.Hydrogen {
			return ErrNotEnoughRes
		}

		// Списываем ресурсы.
		if _, err := tx.Exec(ctx, `
			UPDATE planets
			SET metal = metal - $1, silicon = silicon - $2, hydrogen = hydrogen - $3
			WHERE id = $4
		`, demoCost.Metal, demoCost.Silicon, demoCost.Hydrogen, p.ID); err != nil {
			return fmt.Errorf("charge resources: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO res_log (user_id, planet_id, reason, delta_metal, delta_silicon, delta_hydrogen)
			VALUES ($1, $2, 'demolish', $3, $4, $5)
		`, userID, p.ID, -demoCost.Metal, -demoCost.Silicon, -demoCost.Hydrogen); err != nil {
			return fmt.Errorf("res_log: %w", err)
		}

		// Длительность demolish = build_duration × 0.5 (стандарт OGame).
		robo, _ := currentLevel(ctx, tx, p.ID, s.catalog.Buildings.Buildings["robotic_factory"].ID)
		var nano int
		if nanoSpec, ok := s.catalog.Buildings.Buildings["nano_factory"]; ok {
			nano, _ = currentLevel(ctx, tx, p.ID, nanoSpec.ID)
		}
		start := time.Now().UTC()
		buildDur := economy.BuildDuration(spec.TimeBaseSeconds, demoCost, robo, nano, s.gameSpd)
		dur := buildDur / 2
		end := start.Add(dur)

		id := ids.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO construction_queue (id, planet_id, unit_id, unit_type, target_level,
			                                start_at, end_at, cost_metal, cost_silicon, cost_hydrogen, status)
			VALUES ($1, $2, $3, 'building', $4, $5, $6, $7, $8, $9, 'running')
		`, id, p.ID, unitID, targetLevel, start, end, demoCost.Metal, demoCost.Silicon, demoCost.Hydrogen); err != nil {
			return fmt.Errorf("insert queue: %w", err)
		}

		// Event Kind=2 = KindDemolishConstruction.
		if _, err := tx.Exec(ctx, `
			INSERT INTO events (id, user_id, planet_id, kind, state, fire_at, payload)
			VALUES ($1, $2, $3, 2, 'wait', $4, $5)
		`, ids.New(), userID, p.ID, end,
			fmt.Sprintf(`{"queue_id":"%s","unit_id":%d,"target_level":%d}`, id, unitID, targetLevel)); err != nil {
			return fmt.Errorf("insert event: %w", err)
		}

		item = QueueItem{
			ID: id, PlanetID: p.ID, UnitID: unitID, TargetLevel: targetLevel,
			StartAt: start, EndAt: end, Status: "running",
		}
		return nil
	})
	return item, err
}

// List возвращает текущую очередь планеты (включая running).
func (s *Service) List(ctx context.Context, planetID string) ([]QueueItem, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, planet_id, unit_id, target_level, start_at, end_at, status
		FROM construction_queue
		WHERE planet_id = $1 AND status IN ('queued','running') AND end_at > NOW()
		ORDER BY start_at
	`, planetID)
	if err != nil {
		return nil, fmt.Errorf("list queue: %w", err)
	}
	defer rows.Close()
	var out []QueueItem
	for rows.Next() {
		var q QueueItem
		if err := rows.Scan(&q.ID, &q.PlanetID, &q.UnitID, &q.TargetLevel, &q.StartAt, &q.EndAt, &q.Status); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

// Cancel отменяет задачу, возвращает процент ресурсов согласно §5.3.
func (s *Service) Cancel(ctx context.Context, userID, queueID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var (
			planetID string
			startAt  time.Time
			cm, cs, ch int64
		)
		err := tx.QueryRow(ctx, `
			SELECT planet_id, start_at, cost_metal, cost_silicon, cost_hydrogen
			FROM construction_queue
			WHERE id = $1 AND status IN ('queued','running')
			FOR UPDATE
		`, queueID).Scan(&planetID, &startAt, &cm, &cs, &ch)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrQueueItemNotFound
			}
			return fmt.Errorf("select queue: %w", err)
		}

		refundFactor := 0.95
		if time.Since(startAt) < 15*time.Second {
			refundFactor = 1.0
		}

		rm := int64(float64(cm) * refundFactor)
		rs := int64(float64(cs) * refundFactor)
		rh := int64(float64(ch) * refundFactor)

		if _, err := tx.Exec(ctx, `
			UPDATE planets SET metal = metal + $1, silicon = silicon + $2, hydrogen = hydrogen + $3
			WHERE id = $4
		`, rm, rs, rh, planetID); err != nil {
			return fmt.Errorf("refund: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO res_log (user_id, planet_id, reason, delta_metal, delta_silicon, delta_hydrogen)
			VALUES ($1, $2, 'refund', $3, $4, $5)
		`, userID, planetID, rm, rs, rh); err != nil {
			return fmt.Errorf("res_log: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE construction_queue SET status='cancelled' WHERE id=$1
		`, queueID); err != nil {
			return fmt.Errorf("update queue: %w", err)
		}
		return nil
	})
}

func (s *Service) lookupBuilding(unitID int) (string, config.BuildingSpec, bool) {
	for key, spec := range s.catalog.Buildings.Buildings {
		if spec.ID == unitID {
			return key, spec, true
		}
	}
	return "", config.BuildingSpec{}, false
}

// Levels возвращает map unit_id → level для всех построенных зданий планеты.
func (s *Service) Levels(ctx context.Context, planetID string) (map[int]int, error) {
	rows, err := s.db.Pool().Query(ctx,
		`SELECT unit_id, level FROM buildings WHERE planet_id=$1 AND level>0`,
		planetID)
	if err != nil {
		return nil, fmt.Errorf("building levels: %w", err)
	}
	defer rows.Close()
	out := make(map[int]int)
	for rows.Next() {
		var uid, lvl int
		if err := rows.Scan(&uid, &lvl); err != nil {
			return nil, err
		}
		out[uid] = lvl
	}
	return out, rows.Err()
}

// BuildSecondsMap возвращает время постройки следующего уровня каждого здания
// в секундах, с учётом robotic_factory, nano_factory на планете и скорости игры.
func (s *Service) BuildSecondsMap(ctx context.Context, planetID string, levels map[int]int) (map[int]int, error) {
	roboSpec, ok := s.catalog.Buildings.Buildings["robotic_factory"]
	if !ok {
		return map[int]int{}, nil
	}
	var roboLevel int
	_ = s.db.Pool().QueryRow(ctx,
		`SELECT level FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
		planetID, roboSpec.ID,
	).Scan(&roboLevel)

	var nanoLevel int
	if nanoSpec, ok := s.catalog.Buildings.Buildings["nano_factory"]; ok {
		_ = s.db.Pool().QueryRow(ctx,
			`SELECT level FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
			planetID, nanoSpec.ID,
		).Scan(&nanoLevel)
	}

	out := make(map[int]int, len(s.catalog.Buildings.Buildings))
	for _, spec := range s.catalog.Buildings.Buildings {
		curLvl := levels[spec.ID]
		nextLvl := curLvl + 1
		cost := economy.CostForLevel(economy.Cost{
			Metal:   spec.CostBase.Metal,
			Silicon: spec.CostBase.Silicon,
		}, spec.CostFactor, nextLvl)
		dur := economy.BuildDuration(spec.TimeBaseSeconds, cost, roboLevel, nanoLevel, s.gameSpd)
		out[spec.ID] = int(dur.Seconds())
	}
	return out, nil
}

// BuildingCost — стоимость постройки следующего уровня здания.
type BuildingCost struct {
	Metal    int64 `json:"metal"`
	Silicon  int64 `json:"silicon"`
	Hydrogen int64 `json:"hydrogen"`
}

// BuildCostsMap возвращает стоимость следующего уровня каждого здания.
// Pixel-perfect клон legacy required_res_table (план 72.1 ч.20).
func (s *Service) BuildCostsMap(levels map[int]int) map[int]BuildingCost {
	out := make(map[int]BuildingCost, len(s.catalog.Buildings.Buildings))
	for _, spec := range s.catalog.Buildings.Buildings {
		curLvl := levels[spec.ID]
		nextLvl := curLvl + 1
		cost := economy.CostForLevel(economy.Cost{
			Metal:    spec.CostBase.Metal,
			Silicon:  spec.CostBase.Silicon,
			Hydrogen: spec.CostBase.Hydrogen,
		}, spec.CostFactor, nextLvl)
		out[spec.ID] = BuildingCost{
			Metal:    cost.Metal,
			Silicon:  cost.Silicon,
			Hydrogen: cost.Hydrogen,
		}
	}
	return out
}

// RequirementsUnmet возвращает map unitKey → []UnmetItem для всех зданий
// у которых не выполнены пререквизиты на данной планете.
func (s *Service) RequirementsUnmet(ctx context.Context, userID, planetID string) (map[string][]requirements.UnmetItem, error) {
	out := make(map[string][]requirements.UnmetItem)
	for key := range s.catalog.Buildings.Buildings {
		unmet, err := s.reqs.UnmetForTarget(ctx, s.db.Pool(), key, userID, planetID)
		if err != nil {
			return nil, err
		}
		if len(unmet) > 0 {
			out[key] = unmet
		}
	}
	return out, nil
}

func currentLevel(ctx context.Context, tx pgx.Tx, planetID string, unitID int) (int, error) {
	var lvl int
	err := tx.QueryRow(ctx,
		`SELECT level FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
		planetID, unitID,
	).Scan(&lvl)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("level: %w", err)
	}
	return lvl, nil
}

// buildingLevelsTx читает все постройки планеты как map[unit_id]level.
// Используется для планет-wide расчётов (MaxFields и т.п.), где нужен
// уровень нескольких ключевых зданий сразу (terra_former, moon_lab).
func buildingLevelsTx(ctx context.Context, tx pgx.Tx, planetID string) (map[int]int, error) {
	rows, err := tx.Query(ctx,
		`SELECT unit_id, level FROM buildings WHERE planet_id=$1`,
		planetID)
	if err != nil {
		return nil, fmt.Errorf("building levels: %w", err)
	}
	defer rows.Close()
	out := make(map[int]int)
	for rows.Next() {
		var uid, lvl int
		if err := rows.Scan(&uid, &lvl); err != nil {
			return nil, fmt.Errorf("scan level: %w", err)
		}
		out[uid] = lvl
	}
	return out, rows.Err()
}
