package aiadvisor

import (
	"context"
	"fmt"
	"time"

	"oxsar/game-nova/internal/building"
	"oxsar/game-nova/internal/planet"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/internal/research"
	"oxsar/game-nova/internal/score"
)

// PlayerSnapshot — мгновенное состояние игрока, на котором основано
// scoring кандидатов. Собирается в Compute (внутри транзакции воркера)
// и передаётся в scoreCandidates как чистая функция.
//
// Snapshot включает только то, что нужно для оценки ходов;
// например историю боёв или подробные коэффициенты we-don't-need.
type PlayerSnapshot struct {
	UserID  string
	Now     time.Time
	Credits float64
	Score   float64

	Planets   []PlanetSnapshot
	Research  map[int]int // unitID → level (research2user)
	Fleets    []FleetSnapshot
	Neighbors []NeighborSnapshot

	// Флаги среды.
	Umode      bool
	IsObserver bool
}

// PlanetSnapshot — состояние одной планеты на момент snapshot.
type PlanetSnapshot struct {
	ID        string
	Name      string
	IsMoon    bool
	Galaxy    int
	System    int
	Position  int
	Diameter  int
	UsedFields int
	MaxFields  int

	Resources Resources
	Capacity  Resources
	Rates     Resources // в секунду

	Buildings map[int]int  // unitID → level
	Ships     map[int]int64 // unitID → количество кораблей на планете
	Defense   map[int]int64 // unitID → количество defense установок

	// Очередь занята (true → нельзя стартовать новую постройку/ресёрч на этой планете).
	BuildingQueueBusy bool
	ResearchQueueBusy bool // глобальный флаг (research один на игрока), но дублируется для удобства

	// Прогноз через 1 час (учитывает производство и cap).
	ResourcesIn1h Resources
}

// Resources — re-export planet.Resources для удобства (без импорта planet в scoring).
type Resources struct {
	Metal    float64 `json:"metal"`
	Silicon  float64 `json:"silicon"`
	Hydrogen float64 `json:"hydrogen"`
}

// FleetSnapshot — флот игрока (дома или в полёте).
type FleetSnapshot struct {
	ID            string
	OriginPlanet  string
	TargetPlanet  string
	Mission       int
	StartTime     time.Time
	ArrivalTime   time.Time
	ReturnTime    time.Time
	Ships         map[int]int
	IsHome        bool // true если флот стоит на планете
	IsReturning   bool
}

// NeighborSnapshot — игрок-сосед в радиусе сканирования.
// Используется для выбора целей атаки/шпионажа в Ф.2.
type NeighborSnapshot struct {
	UserID         string
	Galaxy         int
	System         int
	Position       int
	MilitaryScore  float64
	TotalScore     float64
	HasProtection  bool
	Umode          bool
	IsObserver     bool
}

// SnapshotDeps — зависимости buildSnapshot (DI для тестов).
//
// BuildingSvc / ResearchSvc — для чтения уровней и очередей.
// Если nil — Buildings/Research остаются пустыми, флаги очередей false.
// Это допустимо в тестах, но не в проде: server/main.go и worker/main.go
// обязаны передать все.
//
// ProtectionPeriod — секунды защиты новичка (cfg.Game.ProtectionPeriod).
// Используется в readNeighbors для пометки HasProtection.
type SnapshotDeps struct {
	DB               repo.Exec
	PlanetSvc        *planet.Service
	ScoreSvc         *score.Service
	BuildingSvc      *building.Service
	ResearchSvc      *research.Service
	ProtectionPeriod int
}

// buildSnapshot собирает PlayerSnapshot для пользователя.
//
// Не модифицирует БД (только чтение). Внешняя транзакция не нужна:
// между чтениями возможны малые гонки с другими handler-ами,
// но это допустимо — рекомендация всё равно валидируется при Execute
// (building.Enqueue и т.п. сами проверяют ресурсы и очередь).
//
// Расширения по фазам:
//   Ф.1а: планеты + ресурсы + прогноз 1ч + score + кредиты + флаги среды.
//   Ф.1б: уровни зданий per-planet, уровни исследований, флаги очередей.
//   Ф.2:  флоты (дом + в полёте) и список соседей в радиусе.
func buildSnapshot(ctx context.Context, deps SnapshotDeps, userID string) (PlayerSnapshot, error) {
	if deps.DB == nil || deps.PlanetSvc == nil || deps.ScoreSvc == nil {
		return PlayerSnapshot{}, fmt.Errorf("autopilot: buildSnapshot: nil dependency")
	}

	snap := PlayerSnapshot{
		UserID:   userID,
		Now:      time.Now().UTC(),
		Research: map[int]int{},
	}

	// Чтение баланса кредитов и флагов umode/observer одним запросом.
	if err := deps.DB.Pool().QueryRow(ctx, `
		SELECT COALESCE(credit, 0), COALESCE(umode, false), COALESCE(is_observer, false)
		FROM users WHERE id = $1
	`, userID).Scan(&snap.Credits, &snap.Umode, &snap.IsObserver); err != nil {
		return PlayerSnapshot{}, fmt.Errorf("autopilot: buildSnapshot: read user: %w", err)
	}

	// Текущие очки. PlayerScore может вернуть ошибку, если пользователь
	// ещё не в users_score; snapshot всё равно валиден, score=0
	// не критично для scoring (score дальше используется как ratio,
	// а не абсолют).
	if pts, err := deps.ScoreSvc.PlayerScore(ctx, userID, "total"); err == nil {
		snap.Score = pts
	}

	// Уровни исследований и активная очередь исследования.
	// Research один на игрока — флаг ResearchQueueBusy глобален и
	// дублируется в каждый PlanetSnapshot для удобства scoring.
	researchBusy := false
	if deps.ResearchSvc != nil {
		levels, _, err := deps.ResearchSvc.Levels(ctx, userID)
		if err != nil {
			return PlayerSnapshot{}, fmt.Errorf("autopilot: buildSnapshot: research levels: %w", err)
		}
		for unitID, lvl := range levels {
			snap.Research[unitID] = lvl
		}
		queue, err := deps.ResearchSvc.List(ctx, userID)
		if err != nil {
			return PlayerSnapshot{}, fmt.Errorf("autopilot: buildSnapshot: research queue: %w", err)
		}
		researchBusy = len(queue) > 0
	}

	// Планеты + ресурсы.
	planets, err := deps.PlanetSvc.ListByUser(ctx, userID)
	if err != nil {
		return PlayerSnapshot{}, fmt.Errorf("autopilot: buildSnapshot: list planets: %w", err)
	}
	for _, p := range planets {
		ps := PlanetSnapshot{
			ID:         p.ID,
			Name:       p.Name,
			IsMoon:     p.IsMoon,
			Galaxy:     p.Galaxy,
			System:     p.System,
			Position:   p.Position,
			Diameter:   p.Diameter,
			UsedFields: p.UsedFields,
			MaxFields:  p.MaxFields,
			Resources: Resources{
				Metal:    p.Metal,
				Silicon:  p.Silicon,
				Hydrogen: p.Hydrogen,
			},
			Capacity: Resources{
				Metal:    p.MetalCap,
				Silicon:  p.SiliconCap,
				Hydrogen: p.HydrogenCap,
			},
			Rates: Resources{
				Metal:    p.MetalPerSec,
				Silicon:  p.SiliconPerSec,
				Hydrogen: p.HydrogenPerSec,
			},
			Buildings:         map[int]int{},
			ResearchQueueBusy: researchBusy,
		}
		// Прогноз через 1 час (cap-aware).
		ps.ResourcesIn1h = Resources{
			Metal:    minF(p.Metal+p.MetalPerSec*3600, p.MetalCap),
			Silicon:  minF(p.Silicon+p.SiliconPerSec*3600, p.SiliconCap),
			Hydrogen: minF(p.Hydrogen+p.HydrogenPerSec*3600, p.HydrogenCap),
		}

		if deps.BuildingSvc != nil {
			levels, err := deps.BuildingSvc.Levels(ctx, p.ID)
			if err != nil {
				return PlayerSnapshot{}, fmt.Errorf("autopilot: buildSnapshot: building levels for %s: %w", p.ID, err)
			}
			for unitID, lvl := range levels {
				ps.Buildings[unitID] = lvl
			}
			queue, err := deps.BuildingSvc.List(ctx, p.ID)
			if err != nil {
				return PlayerSnapshot{}, fmt.Errorf("autopilot: buildSnapshot: building queue for %s: %w", p.ID, err)
			}
			ps.BuildingQueueBusy = len(queue) > 0
		}

		// Корабли и оборона на планете — простой SQL без shipyard.Service,
		// чтобы не тащить ещё одну зависимость в autopilot. Чтение
		// неконкурентно (ровно те же таблицы, что использует shipyard).
		ships, defense, err := readShipsAndDefense(ctx, deps.DB, p.ID)
		if err != nil {
			return PlayerSnapshot{}, fmt.Errorf("autopilot: buildSnapshot: ships for %s: %w", p.ID, err)
		}
		ps.Ships = ships
		ps.Defense = defense

		snap.Planets = append(snap.Planets, ps)
	}

	// Флоты в полёте — нужны для отсева занятого тоннажа при выборе
	// миссий (особенно экспедиция: лимит = computer_tech уровень).
	fleets, err := readFleets(ctx, deps.DB, userID)
	if err != nil {
		return PlayerSnapshot{}, fmt.Errorf("autopilot: buildSnapshot: fleets: %w", err)
	}
	snap.Fleets = fleets

	// Соседи (Ф.2.2): берём всех игроков в той же galaxy
	// в радиусе ±2 систем от любой планеты игрока.
	if len(snap.Planets) > 0 {
		neighbors, err := readNeighbors(ctx, deps.DB, userID, snap.Planets, deps.ProtectionPeriod)
		if err != nil {
			return PlayerSnapshot{}, fmt.Errorf("autopilot: buildSnapshot: neighbors: %w", err)
		}
		snap.Neighbors = neighbors
	}

	return snap, nil
}

// readShipsAndDefense читает inventory планеты (корабли + оборона).
//
// Возвращает пустые карты если на планете ничего нет; nil-pgx-rows.Err()
// тоже не считаем ошибкой.
func readShipsAndDefense(ctx context.Context, db repo.Exec, planetID string) (map[int]int64, map[int]int64, error) {
	ships := map[int]int64{}
	rows, err := db.Pool().Query(ctx,
		`SELECT unit_id, count FROM ships WHERE planet_id = $1 AND count > 0`,
		planetID)
	if err != nil {
		return nil, nil, fmt.Errorf("ships query: %w", err)
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
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	defense := map[int]int64{}
	rows, err = db.Pool().Query(ctx,
		`SELECT unit_id, count FROM defense WHERE planet_id = $1 AND count > 0`,
		planetID)
	if err != nil {
		return nil, nil, fmt.Errorf("defense query: %w", err)
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

// readNeighbors читает игроков в той же галактике в радиусе ±2 систем
// от любой планеты пользователя.
//
// Поля:
//   - umode/is_observer/banned — нельзя атаковать;
//   - protection_until > now() — игрок в защите новичка;
//   - points (total) — для оценки силы соседа.
//
// Не возвращает: deleted_at != NULL (soft-deleted), сам пользователь,
// игроки без планет в радиусе.
//
// Радиус ±2 систем — компромисс: достаточно чтобы покрыть ближайших
// соседей, но не перегружает scoring сотней целей. Координаты планет
// берутся в Go (а не одним SQL с UNION), чтобы запрос остался простым.
func readNeighbors(ctx context.Context, db repo.Exec, userID string, planets []PlanetSnapshot, protectionPeriod int) ([]NeighborSnapshot, error) {
	if len(planets) == 0 {
		return nil, nil
	}
	const sysRadius = 2

	// Собираем интервалы (galaxy, [system_min, system_max]).
	galaxies := make([]int, len(planets))
	sysMin := make([]int, len(planets))
	sysMax := make([]int, len(planets))
	for i, p := range planets {
		galaxies[i] = p.Galaxy
		sysMin[i] = p.System - sysRadius
		sysMax[i] = p.System + sysRadius
	}

	// HasProtection — true если игрок защищён по любому из условий
	// (legacy fleet/attack.go protectionCheck):
	//   - регистрация в пределах ProtectionPeriod секунд назад,
	//   - protected_until_at > now(),
	//   - is_observer.
	rows, err := db.Pool().Query(ctx, `
		WITH ranges AS (
			SELECT g, smin, smax
			FROM unnest($2::int[], $3::int[], $4::int[]) AS t(g, smin, smax)
		)
		SELECT DISTINCT ON (u.id)
			u.id, p.galaxy, p.system, p.position,
			COALESCE(u.points, 0),
			COALESCE(u.umode, false),
			COALESCE(u.is_observer, false),
			(
				($5 > 0 AND u.created_at > NOW() - ($5 || ' seconds')::interval)
				OR (u.protected_until_at IS NOT NULL AND u.protected_until_at > NOW())
				OR u.is_observer
			) AS has_protection
		FROM planets p
		JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
		JOIN ranges r ON r.g = p.galaxy AND p.system BETWEEN r.smin AND r.smax
		WHERE u.id <> $1
	`, userID, galaxies, sysMin, sysMax, protectionPeriod)
	if err != nil {
		return nil, fmt.Errorf("neighbors query: %w", err)
	}
	defer rows.Close()

	var out []NeighborSnapshot
	for rows.Next() {
		var n NeighborSnapshot
		if err := rows.Scan(&n.UserID, &n.Galaxy, &n.System, &n.Position,
			&n.TotalScore, &n.Umode, &n.IsObserver, &n.HasProtection); err != nil {
			return nil, err
		}
		// MilitaryScore не вычисляем (нужны u_points/b_points): оставляем 0
		// и используем TotalScore как proxy в scoring.
		out = append(out, n)
	}
	return out, rows.Err()
}

// readFleets читает все активные флоты пользователя (state IN
// ('flying','holding','returning')).
//
// Возвращает упрощённую структуру FleetSnapshot — для scoring
// достаточно знать сколько слотов и какой тоннаж в полёте.
func readFleets(ctx context.Context, db repo.Exec, userID string) ([]FleetSnapshot, error) {
	rows, err := db.Pool().Query(ctx, `
		SELECT id, src_planet_id, mission, state, depart_at, arrive_at, return_at
		FROM fleets
		WHERE owner_user_id = $1 AND state IN ('flying','holding','returning')
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FleetSnapshot
	for rows.Next() {
		var (
			f          FleetSnapshot
			returnAt   *time.Time
			state      string
		)
		if err := rows.Scan(&f.ID, &f.OriginPlanet, &f.Mission, &state,
			&f.StartTime, &f.ArrivalTime, &returnAt); err != nil {
			return nil, err
		}
		if returnAt != nil {
			f.ReturnTime = *returnAt
		}
		f.IsReturning = state == "returning"
		f.IsHome = false // в полёте — не дома
		out = append(out, f)
	}
	return out, rows.Err()
}

func minF(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
