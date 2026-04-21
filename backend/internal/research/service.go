// Package research — очередь исследований.
//
// Ресурсы тратятся на конкретной планете, уровни технологий хранятся
// у игрока (таблица research). Одновременно допускается одно
// исследование (§5.4 ТЗ); параллельные слоты через Astrophysics
// придут в M7.
//
// По структуре очередь лежит в той же таблице construction_queue
// (unit_type = 'research'), потому что workflow идентичен: список
// задач со сроком выполнения + событие для воркера.
package research

import (
	"context"
	"errors"
	"fmt"
	"math"
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
	ErrQueueBusy         = errors.New("research: user already researching")
	ErrNotEnoughRes      = errors.New("research: not enough resources")
	ErrUnknownUnit       = errors.New("research: unknown unit")
	ErrPlanetOwnership   = errors.New("research: planet not owned by user")
	ErrNoResearchLab     = errors.New("research: planet has no research lab")
	ErrQueueItemNotFound = errors.New("research: queue item not found")
)

type Service struct {
	db       repo.Exec
	planets  *planet.Service
	catalog  *config.Catalog
	reqs     *requirements.Checker
	gameSpd  float64
}

func NewService(db repo.Exec, p *planet.Service, cat *config.Catalog, reqs *requirements.Checker, gameSpeed float64) *Service {
	if gameSpeed <= 0 {
		gameSpeed = 1
	}
	return &Service{db: db, planets: p, catalog: cat, reqs: reqs, gameSpd: gameSpeed}
}

type QueueItem struct {
	ID          string    `json:"id"`
	PlanetID    string    `json:"planet_id"`
	UnitID      int       `json:"unit_id"`
	TargetLevel int       `json:"target_level"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	Status      string    `json:"status"`
}

// Enqueue ставит исследование на планете planetID. Планета нужна только
// как источник ресурсов и для проверки «есть ли research_lab».
func (s *Service) Enqueue(ctx context.Context, userID, planetID string, unitID int) (QueueItem, error) {
	key, spec, ok := s.lookupResearch(unitID)
	if !ok {
		return QueueItem{}, ErrUnknownUnit
	}

	// Тик экономики + проверка владения планетой.
	p, err := s.planets.Get(ctx, planetID)
	if err != nil {
		return QueueItem{}, err
	}
	if p.UserID != userID {
		return QueueItem{}, ErrPlanetOwnership
	}

	var item QueueItem
	err = s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. Проверка, что у игрока сейчас нет другого исследования
		//    (§5.4 ТЗ, одно исследование одновременно).
		var busy int
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(*) FROM construction_queue cq
			JOIN planets pl ON pl.id = cq.planet_id
			WHERE pl.user_id = $1 AND cq.unit_type = 'research'
			  AND cq.status IN ('queued','running')
		`, userID).Scan(&busy); err != nil {
			return fmt.Errorf("check busy: %w", err)
		}
		if busy > 0 {
			return ErrQueueBusy
		}

		// 2. Проверка зависимостей.
		if err := s.reqs.Check(ctx, tx, key, userID, planetID); err != nil {
			return err
		}

		// 3. Research lab должен быть хотя бы 1 уровня (иначе не можем
		//    исследовать в принципе — это частный случай requirements,
		//    но полезно явно).
		var labLvl int
		err := tx.QueryRow(ctx, `
			SELECT level FROM buildings WHERE planet_id = $1 AND unit_id = $2
		`, planetID, s.catalog.Buildings.Buildings["research_lab"].ID).Scan(&labLvl)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("lab level: %w", err)
		}
		if labLvl < 1 {
			return ErrNoResearchLab
		}

		// 4. Текущий уровень исследования (у игрока, не у планеты).
		curLevel := 0
		err = tx.QueryRow(ctx,
			`SELECT level FROM research WHERE user_id = $1 AND unit_id = $2`,
			userID, unitID).Scan(&curLevel)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("current level: %w", err)
		}
		targetLevel := curLevel + 1
		cost := economy.CostForLevel(economy.Cost{
			Metal:    spec.CostBase.Metal,
			Silicon:  spec.CostBase.Silicon,
			Hydrogen: spec.CostBase.Hydrogen,
		}, spec.CostFactor, targetLevel)

		if int64(p.Metal) < cost.Metal || int64(p.Silicon) < cost.Silicon || int64(p.Hydrogen) < cost.Hydrogen {
			return ErrNotEnoughRes
		}

		// 5. Снять ресурсы.
		if _, err := tx.Exec(ctx, `
			UPDATE planets
			SET metal = metal - $1, silicon = silicon - $2, hydrogen = hydrogen - $3
			WHERE id = $4
		`, cost.Metal, cost.Silicon, cost.Hydrogen, planetID); err != nil {
			return fmt.Errorf("charge: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO res_log (user_id, planet_id, reason, delta_metal, delta_silicon, delta_hydrogen)
			VALUES ($1, $2, 'research', $3, $4, $5)
		`, userID, planetID, -cost.Metal, -cost.Silicon, -cost.Hydrogen); err != nil {
			return fmt.Errorf("res_log: %w", err)
		}

		// 6. Длительность. Формула исследования:
		//    t = (m+s) / (1000 * (1 + lab_level)) секунд, / GAMESPEED.
		resSum := float64(cost.Metal + cost.Silicon)
		raw := resSum / (1000.0 * float64(1+labLvl))
		if s.gameSpd > 0 {
			raw /= s.gameSpd
		}
		if raw < 1 {
			raw = 1
		}
		dur := time.Duration(math.Round(raw * float64(time.Second)))
		start := time.Now().UTC()
		end := start.Add(dur)

		id := ids.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO construction_queue (id, planet_id, unit_id, unit_type, target_level,
			                                start_at, end_at, cost_metal, cost_silicon, cost_hydrogen, status)
			VALUES ($1, $2, $3, 'research', $4, $5, $6, $7, $8, $9, 'running')
		`, id, planetID, unitID, targetLevel, start, end, cost.Metal, cost.Silicon, cost.Hydrogen); err != nil {
			return fmt.Errorf("insert queue: %w", err)
		}

		// 7. Событие завершения.
		if _, err := tx.Exec(ctx, `
			INSERT INTO events (id, user_id, planet_id, kind, state, fire_at, payload)
			VALUES ($1, $2, $3, 3, 'wait', $4, $5)
		`, ids.New(), userID, planetID, end,
			fmt.Sprintf(`{"queue_id":"%s","unit_id":%d,"target_level":%d}`, id, unitID, targetLevel)); err != nil {
			return fmt.Errorf("insert event: %w", err)
		}

		item = QueueItem{
			ID: id, PlanetID: planetID, UnitID: unitID, TargetLevel: targetLevel,
			StartAt: start, EndAt: end, Status: "running",
		}
		return nil
	})
	return item, err
}

// List возвращает текущие исследования пользователя (по всем его планетам).
func (s *Service) List(ctx context.Context, userID string) ([]QueueItem, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT cq.id, cq.planet_id, cq.unit_id, cq.target_level, cq.start_at, cq.end_at, cq.status
		FROM construction_queue cq
		JOIN planets pl ON pl.id = cq.planet_id
		WHERE pl.user_id = $1 AND cq.unit_type = 'research'
		  AND cq.status IN ('queued','running')
		ORDER BY cq.start_at
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list research: %w", err)
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

// Levels возвращает текущие уровни всех исследований пользователя.
// Пригодится для UI Research-экрана.
func (s *Service) Levels(ctx context.Context, userID string) (map[int]int, error) {
	rows, err := s.db.Pool().Query(ctx,
		`SELECT unit_id, level FROM research WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("levels: %w", err)
	}
	defer rows.Close()
	out := map[int]int{}
	for rows.Next() {
		var id, lvl int
		if err := rows.Scan(&id, &lvl); err != nil {
			return nil, err
		}
		out[id] = lvl
	}
	return out, rows.Err()
}

func (s *Service) lookupResearch(unitID int) (string, config.ResearchSpec, bool) {
	for key, spec := range s.catalog.Research.Research {
		if spec.ID == unitID {
			return key, spec, true
		}
	}
	return "", config.ResearchSpec{}, false
}
