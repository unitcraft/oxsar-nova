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

	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/economy"
	"oxsar/game-nova/internal/planet"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/internal/requirements"
	"oxsar/game-nova/pkg/ids"
)

var (
	ErrQueueBusy         = errors.New("research: user already researching")
	ErrNotEnoughRes      = errors.New("research: not enough resources")
	ErrUnknownUnit       = errors.New("research: unknown unit")
	ErrPlanetOwnership   = errors.New("research: planet not owned by user")
	ErrNoResearchLab     = errors.New("research: planet has no research lab")
	ErrQueueItemNotFound = errors.New("research: queue item not found")
	// План 72.1.39: legacy `Research.class.php` блокировки (umode/observer)
	// + MAX_RESEARCH_LEVEL.
	ErrUmodeBlocked      = errors.New("research: blocked in vacation mode")
	ErrObserverBlocked   = errors.New("research: blocked in observer mode")
	ErrMaxLevelReached   = errors.New("research: max level reached")
)

// MaxResearchLevel — legacy `MAX_RESEARCH_LEVEL` (consts.php:305), для
// non-deathmatch. DEATHMATCH=35 — отдельная ветка для будущего.
const MaxResearchLevel = 40

type Service struct {
	db          repo.Exec
	planets     *planet.Service
	catalog     *config.Catalog
	reqs        *requirements.Checker
	gameSpd     float64
	researchSpd float64 // RESEARCH_SPEED_FACTOR
}

func NewService(db repo.Exec, p *planet.Service, cat *config.Catalog, reqs *requirements.Checker, gameSpeed float64) *Service {
	return NewServiceWithFactors(db, p, cat, reqs, gameSpeed, 1)
}

func NewServiceWithFactors(db repo.Exec, p *planet.Service, cat *config.Catalog, reqs *requirements.Checker, gameSpeed, researchSpeedFactor float64) *Service {
	if gameSpeed <= 0 {
		gameSpeed = 1
	}
	if researchSpeedFactor <= 0 {
		researchSpeedFactor = 1
	}
	return &Service{db: db, planets: p, catalog: cat, reqs: reqs, gameSpd: gameSpeed, researchSpd: researchSpeedFactor}
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

	// План 72.1.39: legacy `Research.class.php` строки 76-78 блокирует
	// исследование при umode/observer.
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

		// План 20 Ф.8: IGR network. При igr_level >= 1 лаба-источник
		// объединяется с топ-N других лабораторий игрока (N = igr_level).
		// Формула effective = sum(top (igr+1) lab-levels DESC).
		effectiveLab, err := s.effectiveLabLevel(ctx, tx, userID, planetID, labLvl)
		if err != nil {
			return fmt.Errorf("effective lab: %w", err)
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
		// План 72.1.39: legacy MAX_RESEARCH_LEVEL=40 (consts.php:305).
		if targetLevel > MaxResearchLevel {
			return ErrMaxLevelReached
		}
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
		//    t = (m+s) / (1000 * (1 + effective_lab)) секунд, / GAMESPEED / RESEARCH_SPEED_FACTOR.
		// effective_lab = sum(top (1+igr_level) labs DESC).
		resSum := float64(cost.Metal + cost.Silicon)
		raw := resSum / (1000.0 * float64(1+effectiveLab))
		if s.gameSpd > 0 {
			raw /= s.gameSpd
		}
		if s.researchSpd > 0 {
			raw /= s.researchSpd
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

// Cancel отменяет research-задачу. План 72.1.39 / правило 1:1 для
// /research (legacy `Research::abort` через `abortConstructionEvent`).
// Refund 95% (или 100% если <15 сек), удаление события + пометка
// queue.status='cancelled'.
//
// Возвращает ErrQueueItemNotFound если задачи нет.
func (s *Service) Cancel(ctx context.Context, userID, queueID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var (
			planetID   string
			startAt    time.Time
			cm, cs, ch int64
			ownerID    string
		)
		err := tx.QueryRow(ctx, `
			SELECT cq.planet_id, cq.start_at, cq.cost_metal, cq.cost_silicon, cq.cost_hydrogen,
			       p.user_id
			FROM construction_queue cq
			JOIN planets p ON p.id = cq.planet_id
			WHERE cq.id = $1 AND cq.unit_type = 'research'
			  AND cq.status IN ('queued','running')
			FOR UPDATE
		`, queueID).Scan(&planetID, &startAt, &cm, &cs, &ch, &ownerID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrQueueItemNotFound
			}
			return fmt.Errorf("select queue: %w", err)
		}
		if ownerID != userID {
			return ErrPlanetOwnership
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
		// Cancel pending event (его handler не должен apply'ить уровень).
		if _, err := tx.Exec(ctx, `
			UPDATE events SET state='cancelled'
			WHERE kind=3 AND state='wait' AND user_id=$1
			  AND payload @> jsonb_build_object('queue_id', $2::text)
		`, userID, queueID); err != nil {
			return fmt.Errorf("cancel event: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE construction_queue SET status='cancelled' WHERE id=$1`,
			queueID); err != nil {
			return fmt.Errorf("update queue: %w", err)
		}
		return nil
	})
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

// ResearchSecondsMap возвращает время исследования следующего уровня каждой технологии
// в секундах. Использует максимальный уровень research_lab среди всех планет пользователя.
func (s *Service) ResearchSecondsMap(ctx context.Context, userID string, levels map[int]int) (map[int]int, error) {
	labSpec, ok := s.catalog.Buildings.Buildings["research_lab"]
	if !ok {
		return map[int]int{}, nil
	}
	// План 20 Ф.8: effective lab = sum top-(1+igr) уровней. Если igr=0
	// — эквивалент max(level), как и было раньше. Один SQL:
	// igr_level в подзапросе LIMIT.
	var labLvl int
	_ = s.db.Pool().QueryRow(ctx, `
		SELECT COALESCE(SUM(level), 0) FROM (
			SELECT b.level
			FROM buildings b
			JOIN planets p ON p.id = b.planet_id
			WHERE p.user_id = $1 AND b.unit_id = $2 AND p.destroyed_at IS NULL
			ORDER BY b.level DESC
			LIMIT 1 + COALESCE(
				(SELECT level FROM research WHERE user_id = $1 AND unit_id = $3),
				0
			)
		) t
	`, userID, labSpec.ID, unitIGRTech).Scan(&labLvl)

	out := make(map[int]int, len(s.catalog.Research.Research))
	for _, spec := range s.catalog.Research.Research {
		curLvl := levels[spec.ID]
		nextLvl := curLvl + 1
		cost := economy.CostForLevel(economy.Cost{
			Metal:   spec.CostBase.Metal,
			Silicon: spec.CostBase.Silicon,
		}, spec.CostFactor, nextLvl)
		resSum := float64(cost.Metal + cost.Silicon)
		raw := resSum / (1000.0 * float64(1+labLvl))
		if s.gameSpd > 0 {
			raw /= s.gameSpd
		}
		if s.researchSpd > 0 {
			raw /= s.researchSpd
		}
		if raw < 1 {
			raw = 1
		}
		out[spec.ID] = int(raw)
	}
	return out, nil
}

// ResearchCost — стоимость исследования (металл/кремний/водород).
type ResearchCost struct {
	Metal    int64 `json:"metal"`
	Silicon  int64 `json:"silicon"`
	Hydrogen int64 `json:"hydrogen"`
}

// ResearchCostsMap возвращает стоимость следующего уровня каждой технологии.
// Используется фронтендом для preview cost-таблицы (план 72.1 ч.20).
func (s *Service) ResearchCostsMap(levels map[int]int) map[int]ResearchCost {
	out := make(map[int]ResearchCost, len(s.catalog.Research.Research))
	for _, spec := range s.catalog.Research.Research {
		curLvl := levels[spec.ID]
		nextLvl := curLvl + 1
		cost := economy.CostForLevel(economy.Cost{
			Metal:    spec.CostBase.Metal,
			Silicon:  spec.CostBase.Silicon,
			Hydrogen: spec.CostBase.Hydrogen,
		}, spec.CostFactor, nextLvl)
		out[spec.ID] = ResearchCost{
			Metal:    cost.Metal,
			Silicon:  cost.Silicon,
			Hydrogen: cost.Hydrogen,
		}
	}
	return out
}

func (s *Service) lookupResearch(unitID int) (string, config.ResearchSpec, bool) {
	for key, spec := range s.catalog.Research.Research {
		if spec.ID == unitID {
			return key, spec, true
		}
	}
	return "", config.ResearchSpec{}, false
}

// unitIGRTech — id Intergalactic Research Network (план 20 Ф.8).
const unitIGRTech = 113

// effectiveLabLevel — сумма топ-(1+igr_level) уровней research_lab
// по всем планетам игрока (DESC). При igr=0 → 1 лаба = max planet's
// research_lab (то же поведение что и до плана 20 Ф.8).
//
// План 20 Ф.8: позволяет игроку использовать пул лабораторий,
// если он развивает их на нескольких планетах. Стимул не строить
// одну гигантскую, а распределять.
//
// minLab — уровень лаборатории на планете-источнике, чтобы при
// деградации (одна планета удалена) effective не упал ниже её.
//
// Один SQL: igr_level берётся подзапросом в LIMIT, чтобы не делать
// два round-trip'а к БД.
func (s *Service) effectiveLabLevel(ctx context.Context, tx pgx.Tx,
	userID, planetID string, minLab int) (int, error) {
	labID := s.catalog.Buildings.Buildings["research_lab"].ID
	var sum int
	err := tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(level), 0) FROM (
			SELECT b.level FROM buildings b
			JOIN planets p ON p.id = b.planet_id
			WHERE p.user_id = $1 AND p.destroyed_at IS NULL
			  AND b.unit_id = $2
			ORDER BY b.level DESC
			LIMIT 1 + COALESCE(
				(SELECT level FROM research WHERE user_id = $1 AND unit_id = $3),
				0
			)
		) t
	`, userID, labID, unitIGRTech).Scan(&sum)
	if err != nil {
		return 0, err
	}
	if sum < minLab {
		sum = minLab
	}
	return sum, nil
}
