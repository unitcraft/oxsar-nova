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

	Buildings map[int]int // unitID → level

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
type SnapshotDeps struct {
	DB          repo.Exec
	PlanetSvc   *planet.Service
	ScoreSvc    *score.Service
	BuildingSvc *building.Service
	ResearchSvc *research.Service
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

		snap.Planets = append(snap.Planets, ps)
	}

	// Загрузка флотов и соседей — в Ф.2.

	return snap, nil
}

func minF(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
