package planet

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/internal/economy"
	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/pkg/formula"
)

// Service — бизнес-логика планет. Содержит ApplyTick — расчёт добычи
// за прошедшее время с последней синхронизации (§6.1 ТЗ: при любом
// чтении состояния планеты сервер догоняет тики).
type Service struct {
	db      repo.Exec
	repo    *Repository
	catalog *config.Catalog

	// formulaCache — кеш разобранных Expr по строковому исходнику.
	// Парсим формулы один раз при первом запросе; на горячем пути
	// economy tick'а работаем только с готовым AST. Мьютекс даёт
	// безопасный конкурентный доступ — тик читается параллельно
	// несколькими горутинами.
	formulaCache struct {
		sync.RWMutex
		m map[string]*formula.Expr
	}
}

func NewService(db repo.Exec, r *Repository, cat *config.Catalog) *Service {
	s := &Service{db: db, repo: r, catalog: cat}
	s.formulaCache.m = map[string]*formula.Expr{}
	return s
}

// compile парсит формулу, кеширует и возвращает готовое AST.
// Пустая/невалидная формула → nil без ошибки: в economy tick это
// означает «ресурс не производится этим зданием».
func (s *Service) compile(src string) *formula.Expr {
	if src == "" {
		return nil
	}
	s.formulaCache.RLock()
	if e, ok := s.formulaCache.m[src]; ok {
		s.formulaCache.RUnlock()
		return e
	}
	s.formulaCache.RUnlock()

	e, err := formula.Parse(src)
	if err != nil {
		// Логично было бы прокинуть ошибку наружу, но тогда тик
		// падал бы из-за одной некорректной формулы, и никто бы
		// её не заметил сразу. Вместо этого: null-handler плюс
		// предупреждение через slog (добавим при интеграции
		// наблюдаемости в M8). Сейчас просто нулим — ожидается,
		// что все формулы проходят валидацию в import-datasheets.
		return nil
	}
	s.formulaCache.Lock()
	s.formulaCache.m[src] = e
	s.formulaCache.Unlock()
	return e
}

// Get возвращает планету с уже применённым тиком.
func (s *Service) Get(ctx context.Context, id string) (Planet, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Planet{}, err
	}
	if err := s.applyTickInTx(ctx, &p); err != nil {
		return Planet{}, err
	}
	return p, nil
}

// ListByUser возвращает все планеты игрока с применённым тиком каждой.
func (s *Service) ListByUser(ctx context.Context, userID string) ([]Planet, error) {
	planets, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range planets {
		if err := s.applyTickInTx(ctx, &planets[i]); err != nil {
			return nil, err
		}
	}
	return planets, nil
}

// applyTickInTx догоняет ресурсы до now() и сохраняет в БД. Транзакция
// обеспечивает консистентность с очередями — иначе гонка «стройка
// закончилась + тик добыл ресурсы» могла бы потерять добычу.
func (s *Service) applyTickInTx(ctx context.Context, p *Planet) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		now := time.Now().UTC()
		elapsed := now.Sub(p.LastResUpdate).Seconds()
		if elapsed <= 0 {
			return nil
		}

		// Уровни зданий читаем один раз — используются и для rates,
		// и для storage cap.
		levels, err := buildingLevels(ctx, tx, p.ID)
		if err != nil {
			return err
		}

		// Уровни исследований игрока — нужны для {tech=N} в DSL
		// (energy_tech влияет на cons_energy, laser_tech — на metal,
		// и т.д. см. §5.2.1 ТЗ).
		techLevels, err := researchLevels(ctx, tx, p.UserID)
		if err != nil {
			return err
		}

		rates := s.productionRates(p, levels, techLevels)
		caps := s.storageCap(p, levels)

		// Ограничиваем ёмкостью хранилища: при заполнении cap добыча
		// конкретного ресурса замирает (§5.2 ТЗ, как OGame).
		gainM := rates.metalPerSec * elapsed
		gainS := rates.siliconPerSec * elapsed
		gainH := rates.hydrogenPerSec * elapsed
		p.Metal = clampAdd(p.Metal, gainM, caps.metal)
		p.Silicon = clampAdd(p.Silicon, gainS, caps.silicon)
		p.Hydrogen = clampAdd(p.Hydrogen, gainH, caps.hydrogen)
		p.LastResUpdate = now

		_, err = tx.Exec(ctx, `
			UPDATE planets
			SET metal=$1, silicon=$2, hydrogen=$3, last_res_update=$4
			WHERE id=$5
		`, p.Metal, p.Silicon, p.Hydrogen, p.LastResUpdate, p.ID)
		if err != nil {
			return fmt.Errorf("update planet tick: %w", err)
		}
		return nil
	})
}

type rates struct {
	metalPerSec    float64
	siliconPerSec  float64
	hydrogenPerSec float64
}

type caps struct {
	metal    float64
	silicon  float64
	hydrogen float64
}

// productionRates считает скорость добычи per-second на основе
// уровней зданий и множителей планеты.
//
// Ветка 1 (предпочтительная): если загружен configs/construction.yml
// (ConstructionCatalog), формулы берутся из legacy-парадигмы
// na_construction через pkg/formula — это бит-в-бит паритет с
// oxsar2 (§5.2.1 ТЗ).
//
// Ветка 2 (fallback): старое приближение base_rate_per_hour из
// configs/buildings.yml. Оставлено для режима «мы ещё не прогнали
// import-datasheets».
//
// Не делает IO — все данные уже прочитаны вызывающим.
func (s *Service) productionRates(p *Planet, levels map[int]int, tech map[int]int) rates {
	if len(s.catalog.Construction.Buildings) > 0 {
		return s.productionRatesDSL(p, levels, tech)
	}
	return s.productionRatesApprox(p, levels)
}

// productionRatesDSL — основной путь с legacy-формулами.
func (s *Service) productionRatesDSL(p *Planet, levels map[int]int, tech map[int]int) rates {
	ratio := s.energyRatio(p, levels, tech)
	ctxBase := formula.Context{
		Temperature: (p.TempMin + p.TempMax) / 2,
		Tech:        tech,
	}

	metalPerHour := s.evalProd("metal_mine", "metal", levels, ctxBase)
	silPerHour := s.evalProd("silicon_lab", "silicon", levels, ctxBase)
	hydPerHour := s.evalProd("hydrogen_lab", "hydrogen", levels, ctxBase)

	factor := float64(p.ProduceFactor) * ratio
	return rates{
		metalPerSec:    metalPerHour * factor / 3600.0,
		siliconPerSec:  silPerHour * factor / 3600.0,
		hydrogenPerSec: hydPerHour * factor / 3600.0,
	}
}

// evalProd вычисляет prod-формулу одного здания для конкретного уровня.
// key — ключ в ConstructionCatalog (lowercase legacy-имя, см.
// import-datasheets::snakeCase). resource — metal|silicon|hydrogen|energy.
func (s *Service) evalProd(key, resource string, levels map[int]int, base formula.Context) float64 {
	spec, ok := s.catalog.Construction.Buildings[key]
	if !ok {
		return 0
	}
	var src string
	switch resource {
	case "metal":
		src = spec.Prod.Metal
	case "silicon":
		src = spec.Prod.Silicon
	case "hydrogen":
		src = spec.Prod.Hydrogen
	case "energy":
		src = spec.Prod.Energy
	}
	expr := s.compile(src)
	if expr == nil {
		return 0
	}
	ctx := base
	ctx.Level = levels[int(spec.ID)]
	// basic-field для prod не используется в legacy-формулах, но
	// выставляем: разные поля могут понадобиться в будущих
	// (hydrogen_plant prod зависит от basic_energy).
	ctx.Basic = spec.Basic.Metal // arbitrary, not used in prod formulas
	v, err := expr.Eval(ctx)
	if err != nil {
		return 0
	}
	return v
}

// energyRatio считает долю удовлетворённой энергии (0..1) через
// cons_energy формулы шахт + prod_energy солнечной станции.
func (s *Service) energyRatio(p *Planet, levels map[int]int, tech map[int]int) float64 {
	base := formula.Context{
		Temperature: (p.TempMin + p.TempMax) / 2,
		Tech:        tech,
	}
	demand := s.evalCons("metal_mine", "energy", levels, base) +
		s.evalCons("silicon_lab", "energy", levels, base) +
		s.evalCons("hydrogen_lab", "energy", levels, base)
	output := s.evalProd("solar_plant", "energy", levels, base)
	ratio := economy.EnergyRatio(output, demand) * float64(p.EnergyFactor)
	if ratio > 1 {
		ratio = 1
	}
	return ratio
}

// evalCons вычисляет cons-формулу одного здания.
// resource поддерживает "energy" (для шахт), остальные (metal/silicon/
// hydrogen) доступны, но в прод-формулах шахт не встречаются.
func (s *Service) evalCons(key, resource string, levels map[int]int, base formula.Context) float64 {
	spec, ok := s.catalog.Construction.Buildings[key]
	if !ok {
		return 0
	}
	var src string
	switch resource {
	case "energy":
		src = spec.Cons.Energy
	case "metal":
		src = spec.Cons.Metal
	case "silicon":
		src = spec.Cons.Silicon
	case "hydrogen":
		src = spec.Cons.Hydrogen
	}
	expr := s.compile(src)
	if expr == nil {
		return 0
	}
	// hack: evalProd пишет неиспользуемый Basic; cons тоже basic не
	// использует в legacy-формулах.
	ctx := base
	ctx.Level = levels[int(spec.ID)]
	v, err := expr.Eval(ctx)
	if err != nil {
		return 0
	}
	return v
}

// productionRatesApprox — fallback-путь на приближённых формулах
// из configs/buildings.yml. Используется до прогона import-datasheets.
func (s *Service) productionRatesApprox(p *Planet, levels map[int]int) rates {
	mine := s.catalog.Buildings.Buildings["metal_mine"]
	lab := s.catalog.Buildings.Buildings["silicon_lab"]
	synth := s.catalog.Buildings.Buildings["hydrogen_lab"]
	solar := s.catalog.Buildings.Buildings["solar_plant"]

	metalBase := floatOr(mine.BaseRatePerHour, 30)
	silBase := floatOr(lab.BaseRatePerHour, 20)
	hydBase := floatOr(synth.BaseRatePerHour, 10)

	metalLvl := levels[mine.ID]
	silLvl := levels[lab.ID]
	hydLvl := levels[synth.ID]

	demand := economy.EnergyDemand(floatOr(mine.EnergyPerLevel, 10), metalLvl) +
		economy.EnergyDemand(floatOr(lab.EnergyPerLevel, 10), silLvl) +
		economy.EnergyDemand(floatOr(synth.EnergyPerLevel, 20), hydLvl)
	output := economy.EnergyOutput(floatOr(solar.EnergyOutputPerLevel, 20), levels[solar.ID])
	ratio := economy.EnergyRatio(output, demand) * float64(p.EnergyFactor)
	if ratio > 1 {
		ratio = 1
	}

	f := float64(p.ProduceFactor) * ratio
	return rates{
		metalPerSec:    economy.ProductionPerHour(metalBase, metalLvl, f) / 3600.0,
		siliconPerSec:  economy.ProductionPerHour(silBase, silLvl, f) / 3600.0,
		hydrogenPerSec: economy.ProductionPerHour(hydBase, hydLvl, f) / 3600.0,
	}
}

// researchLevels читает уровни всех исследований игрока.
// Результат — unit_id → level; отсутствие записи означает уровень 0.
// Передаётся в formula.Context как {tech=N}.
func researchLevels(ctx context.Context, tx pgx.Tx, userID string) (map[int]int, error) {
	rows, err := tx.Query(ctx, `SELECT unit_id, level FROM research WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("read research levels: %w", err)
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

// storageCap возвращает ёмкость трёх хранилищ на основе уровней
// складов и множителя storage_factor (§5.10.1 — может быть > 1 от
// активного артефакта ATOMIC_DENSIFIER).
func (s *Service) storageCap(p *Planet, levels map[int]int) caps {
	mStorage := s.catalog.Buildings.Buildings["metal_storage"]
	sStorage := s.catalog.Buildings.Buildings["silicon_storage"]
	hStorage := s.catalog.Buildings.Buildings["hydrogen_storage"]

	baseM := int64OrDefault(mStorage.CapacityBase, 5000)
	baseS := int64OrDefault(sStorage.CapacityBase, 5000)
	baseH := int64OrDefault(hStorage.CapacityBase, 5000)

	factor := float64(p.StorageFactor)
	return caps{
		metal:    economy.StorageCapacity(baseM, levels[mStorage.ID], factor),
		silicon:  economy.StorageCapacity(baseS, levels[sStorage.ID], factor),
		hydrogen: economy.StorageCapacity(baseH, levels[hStorage.ID], factor),
	}
}

func int64OrDefault(p *int64, def int64) int64 {
	if p == nil {
		return def
	}
	return *p
}

// buildingLevels читает уровни всех зданий планеты.
func buildingLevels(ctx context.Context, tx pgx.Tx, planetID string) (map[int]int, error) {
	rows, err := tx.Query(ctx, `SELECT unit_id, level FROM buildings WHERE planet_id = $1`, planetID)
	if err != nil {
		return nil, fmt.Errorf("read building levels: %w", err)
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

func floatOr(p *float64, def float64) float64 {
	if p == nil {
		return def
	}
	return *p
}

func clampAdd(cur, delta, max float64) float64 {
	sum := cur + delta
	if sum > max {
		return max
	}
	if sum < 0 {
		return 0
	}
	return sum
}
