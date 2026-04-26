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

	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/internal/economy"
	"github.com/oxsar/nova/backend/internal/planet"
	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/internal/requirements"
	"github.com/oxsar/nova/backend/pkg/ids"
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
