package planet

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/internal/economy"
	"github.com/oxsar/nova/backend/internal/repo"
)

// Service — бизнес-логика планет. Содержит ApplyTick — расчёт добычи
// за прошедшее время с последней синхронизации (§6.1 ТЗ: при любом
// чтении состояния планеты сервер догоняет тики).
type Service struct {
	db                     repo.Exec
	repo                   *Repository
	catalog                *config.Catalog
	storageFactor          float64 // global STORAGE_FACTOR multiplier
	energyProductionFactor float64 // global ENEGRY_PRODUCTION_FACTOR multiplier
	galaxyEvent            GalaxyEventReader // план 17 F — может быть nil
}

// GalaxyEventReader — узкий интерфейс к galaxyevent.Service.
// Используется для чтения мультипликатора production. Если nil —
// мультипликатор всегда 1.0.
type GalaxyEventReader interface {
	MetalMultiplier(ctx context.Context) float64
}

// SetGalaxyEventReader — wire-up из server/main.go.
func (s *Service) SetGalaxyEventReader(r GalaxyEventReader) {
	s.galaxyEvent = r
}

func NewService(db repo.Exec, r *Repository, cat *config.Catalog) *Service {
	return &Service{db: db, repo: r, catalog: cat, storageFactor: 1, energyProductionFactor: 1}
}

func NewServiceWithFactors(db repo.Exec, r *Repository, cat *config.Catalog, storageFactor, energyProductionFactor float64) *Service {
	if storageFactor <= 0 {
		storageFactor = 1
	}
	if energyProductionFactor <= 0 {
		energyProductionFactor = 1
	}
	return &Service{db: db, repo: r, catalog: cat, storageFactor: storageFactor, energyProductionFactor: energyProductionFactor}
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
	if err := s.fillMaxFields(ctx, &p); err != nil {
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
		if err := s.fillMaxFields(ctx, &planets[i]); err != nil {
			return nil, err
		}
	}
	return planets, nil
}

// fillMaxFields вычисляет p.MaxFields по формуле из fields.go (legacy).
// Читает только terra_former (58) и moon_lab (350), больше ничего из
// зданий не нужно.
func (s *Service) fillMaxFields(ctx context.Context, p *Planet) error {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT unit_id, level FROM buildings
		WHERE planet_id = $1 AND unit_id IN (58, 350)
	`, p.ID)
	if err != nil {
		return fmt.Errorf("max_fields buildings: %w", err)
	}
	defer rows.Close()
	b := make(map[int]int, 2)
	for rows.Next() {
		var uid, lvl int
		if err := rows.Scan(&uid, &lvl); err != nil {
			return err
		}
		b[uid] = lvl
	}
	p.MaxFields = MaxFields(p, b, DefaultFieldConsts)
	return nil
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
		eProd, eCons := s.energyStats(p, levels, techLevels)

		// План 20 Ф.1: при отпуске игрока производство = 0 (legacy
		// setProdOfUser → prod_factor=0 во всех планетах).
		if p.UserID != "" {
			var onVacation bool
			if err := tx.QueryRow(ctx,
				`SELECT vacation_since IS NOT NULL FROM users WHERE id=$1`,
				p.UserID).Scan(&onVacation); err == nil && onVacation {
				rates.metalPerSec = 0
				rates.siliconPerSec = 0
				rates.hydrogenPerSec = 0
			}
		}

		// План 17 F: галактические события. Активный 'meteor_storm'
		// даёт +30% (или иной множитель) к metal production.
		if s.galaxyEvent != nil {
			if mult := s.galaxyEvent.MetalMultiplier(ctx); mult != 1.0 {
				rates.metalPerSec *= mult
			}
		}

		// Ограничиваем ёмкостью хранилища: при заполнении cap добыча
		// конкретного ресурса замирает (§5.2 ТЗ, как OGame).
		gainM := rates.metalPerSec * elapsed
		gainS := rates.siliconPerSec * elapsed
		gainH := rates.hydrogenPerSec * elapsed
		p.Metal = clampAdd(p.Metal, gainM, caps.metal)
		p.Silicon = clampAdd(p.Silicon, gainS, caps.silicon)
		p.Hydrogen = clampAdd(p.Hydrogen, gainH, caps.hydrogen)
		p.LastResUpdate = now
		p.MetalPerSec = rates.metalPerSec
		p.SiliconPerSec = rates.siliconPerSec
		p.HydrogenPerSec = rates.hydrogenPerSec
		p.MetalCap = caps.metal
		p.SiliconCap = caps.silicon
		p.HydrogenCap = caps.hydrogen
		p.EnergyProd = eProd
		p.EnergyCons = eCons
		p.EnergyRemaining = eProd - eCons

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
// (ConstructionCatalog), используются статические формулы из economy —
// бит-в-бит паритет с oxsar2 (§5.2.1 ТЗ).
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

	temp := (p.TempMin + p.TempMax) / 2
	techE := tech[economy.IDTechEnergy]

	metalPerHour := economy.MetalmineProdMetal(levels[economy.IDMetalmine], tech[economy.IDTechLaser])
	silPerHour := economy.SiliconLabProdSilicon(levels[economy.IDSiliconLab], tech[economy.IDTechSilicon])
	hydPerHour := economy.HydrogenLabProdHydrogen(levels[economy.IDHydrogenLab], tech[economy.IDTechHydrogen], temp)
	// Лунный синтезатор — если есть на планете.
	hydPerHour += economy.MoonHydrogenLabProdHydrogen(levels[economy.IDMoonHydrogenLab], tech[economy.IDTechHydrogen], temp)

	factor := float64(p.ProduceFactor) * ratio
	_ = techE // используется ниже в energyStats
	return rates{
		metalPerSec:    metalPerHour * factor / 3600.0,
		siliconPerSec:  silPerHour * factor / 3600.0,
		hydrogenPerSec: hydPerHour * factor / 3600.0,
	}
}

// energyDemand считает суммарное потребление энергии шахтами.
func energyDemand(levels map[int]int, tech map[int]int) float64 {
	techE := tech[economy.IDTechEnergy]
	return economy.MineConsEnergy(10, levels[economy.IDMetalmine], techE) +
		economy.MineConsEnergy(10, levels[economy.IDSiliconLab], techE) +
		economy.MineConsEnergy(20, levels[economy.IDHydrogenLab], techE) +
		economy.MineConsEnergy(200, levels[economy.IDMoonHydrogenLab], techE)
}

// energyStats возвращает (prod, cons) энергии в абсолютных единицах.
func (s *Service) energyStats(p *Planet, levels map[int]int, tech map[int]int) (prod, cons float64) {
	temp := (p.TempMin + p.TempMax) / 2
	techE := tech[economy.IDTechEnergy]
	cons = energyDemand(levels, tech)
	prod = economy.SolarPlantProdEnergy(levels[economy.IDSolarPlant], techE)
	// Вклад солнечных спутников: per-unit * count.
	satCount := levels[economy.IDSolarSatellite]
	if satCount > 0 {
		prod += economy.SolarSatelliteProdEnergy(temp, techE) * float64(satCount)
	}
	// Гравитрон.
	if graviLvl := levels[economy.IDGravi]; graviLvl > 0 {
		prod += economy.GraviProdEnergy(graviLvl, 300000)
	}
	// Синтезатор водорода (если есть).
	prod += economy.HydrogenPlantProdEnergy(levels[economy.IDHydrogenPlant], techE)
	prod *= float64(p.EnergyFactor) * s.energyProductionFactor
	return prod, cons
}

// energyRatio считает долю удовлетворённой энергии (0..1).
func (s *Service) energyRatio(p *Planet, levels map[int]int, tech map[int]int) float64 {
	temp := (p.TempMin + p.TempMax) / 2
	techE := tech[economy.IDTechEnergy]
	demand := energyDemand(levels, tech)
	output := economy.SolarPlantProdEnergy(levels[economy.IDSolarPlant], techE)
	satCount := levels[economy.IDSolarSatellite]
	if satCount > 0 {
		output += economy.SolarSatelliteProdEnergy(temp, techE) * float64(satCount)
	}
	if graviLvl := levels[economy.IDGravi]; graviLvl > 0 {
		output += economy.GraviProdEnergy(graviLvl, 300000)
	}
	output += economy.HydrogenPlantProdEnergy(levels[economy.IDHydrogenPlant], techE)
	ratio := economy.EnergyRatio(output*s.energyProductionFactor, demand) * float64(p.EnergyFactor)
	if ratio > 1 {
		ratio = 1
	}
	return ratio
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

// researchLevels читает уровни всех исследований игрока внутри транзакции.
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

// researchLevelsDirect читает уровни исследований без транзакции (для read-only запросов).
func researchLevelsDirect(ctx context.Context, pool interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}, userID string) (map[int]int, error) {
	rows, err := pool.Query(ctx, `SELECT unit_id, level FROM research WHERE user_id = $1`, userID)
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

// profBonusForUser читает профессию пользователя и возвращает карту
// смещений уровней (ключ профессии → ID из economy.ProfessionKeyToID).
// При ошибке чтения возвращает nil (не блокирует производство).
func profBonusForUser(ctx context.Context, pool interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}, userID string, cat *config.Catalog) map[int]int {
	var prof string
	if err := pool.QueryRow(ctx, `SELECT profession FROM users WHERE id=$1`, userID).Scan(&prof); err != nil {
		return nil
	}
	if prof == "" || prof == "none" {
		return nil
	}
	spec, ok := cat.Professions.Professions[prof]
	if !ok {
		return nil
	}
	out := make(map[int]int)
	for k, v := range spec.Bonus {
		if id, ok := economy.ProfessionKeyToID[k]; ok {
			out[id] += v
		}
	}
	for k, v := range spec.Malus {
		if id, ok := economy.ProfessionKeyToID[k]; ok {
			out[id] += v
		}
	}
	return out
}

// applyProfessionBonus добавляет смещения профессии к карте tech-уровней.
func applyProfessionBonus(tech map[int]int, bonus map[int]int) {
	for id, delta := range bonus {
		tech[id] += delta
	}
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

	factor := float64(p.StorageFactor) * s.storageFactor
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

// Rename переименовывает планету. Валидация: имя 1–50 символов,
// планета принадлежит юзеру.
func (s *Service) Rename(ctx context.Context, userID, planetID, name string) error {
	name = trimSpace(name)
	if len(name) < 1 || len(name) > 50 {
		return fmt.Errorf("planet: invalid name length (must be 1–50 chars): %w", ErrInvalidInput)
	}

	p, err := s.repo.GetByID(ctx, planetID)
	if err != nil {
		return err
	}
	if p.UserID != userID {
		return ErrNotFound
	}

	return s.repo.Rename(ctx, planetID, name)
}

// Reorder устанавливает sort_order планет по указанному порядку ID.
// Все ID должны принадлежать userID, иначе вся операция откатится.
func (s *Service) Reorder(ctx context.Context, userID string, planetIDs []string) error {
	if len(planetIDs) == 0 {
		return nil
	}
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verify ownership.
		var count int
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(*) FROM planets
			WHERE user_id = $1 AND id = ANY($2) AND destroyed_at IS NULL
		`, userID, planetIDs).Scan(&count); err != nil {
			return fmt.Errorf("verify ownership: %w", err)
		}
		if count != len(planetIDs) {
			return fmt.Errorf("planet: reorder ownership mismatch: %w", ErrNotFound)
		}
		for i, id := range planetIDs {
			if _, err := tx.Exec(ctx,
				`UPDATE planets SET sort_order = $1 WHERE id = $2 AND user_id = $3`,
				i, id, userID); err != nil {
				return fmt.Errorf("update sort_order: %w", err)
			}
		}
		return nil
	})
}

// SetHome устанавливает планету домашней. Проверка: не луна, принадлежит юзеру.
func (s *Service) SetHome(ctx context.Context, userID, planetID string) error {
	p, err := s.repo.GetByID(ctx, planetID)
	if err != nil {
		return err
	}
	if p.UserID != userID {
		return ErrNotFound
	}
	if p.IsMoon {
		return fmt.Errorf("planet: cannot set moon as home: %w", ErrMoonRestricted)
	}

	return s.repo.SetHome(ctx, userID, planetID)
}

// Abandon удаляет (мягко) планету. Проверка: не луна, не единственная,
// не домашняя (или есть другая для замены).
func (s *Service) Abandon(ctx context.Context, userID, planetID string) error {
	p, err := s.repo.GetByID(ctx, planetID)
	if err != nil {
		return err
	}
	if p.UserID != userID {
		return ErrNotFound
	}
	if p.IsMoon {
		return fmt.Errorf("planet: cannot abandon moon: %w", ErrMoonRestricted)
	}

	// Проверить что есть хотя бы 2 планеты (не луны).
	planets, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return err
	}
	nonMoons := 0
	for _, pp := range planets {
		if !pp.IsMoon {
			nonMoons++
		}
	}
	if nonMoons < 2 {
		return fmt.Errorf("planet: cannot abandon only planet: %w", ErrOnlyPlanet)
	}

	// Проверить что это не домашняя планета.
	// Home planet = cur_planet_id в таблице users.
	var curPlanetID string
	err = s.db.Pool().QueryRow(ctx, `SELECT cur_planet_id FROM users WHERE id = $1`, userID).
		Scan(&curPlanetID)
	if err != nil {
		return fmt.Errorf("planet: check home: %w", err)
	}
	if curPlanetID == planetID {
		return fmt.Errorf("planet: cannot abandon home planet: %w", ErrCannotAbandonHome)
	}

	return s.repo.Abandon(ctx, planetID)
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\n' || s[0] == '\r') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t' || s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
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

// ResourceReportDTO — отчёт о производстве ресурсов планеты.
type ResourceReportDTO struct {
	PlanetID           string                  `json:"planet_id"`
	PlanetName         string                  `json:"planet_name"`
	Buildings          []ResourceBuildingDTO   `json:"buildings"`
	MetalTotal         float64                 `json:"metal_total"`
	SiliconTotal       float64                 `json:"silicon_total"`
	HydrogenTotal      float64                 `json:"hydrogen_total"`
	MetalPerHour       float64                 `json:"metal_per_hour"`
	SiliconPerHour     float64                 `json:"silicon_per_hour"`
	HydrogenPerHour    float64                 `json:"hydrogen_per_hour"`
	BasicMetal         float64                 `json:"basic_metal"`
	BasicSilicon       float64                 `json:"basic_silicon"`
	BasicHydrogen      float64                 `json:"basic_hydrogen"`
	StorageMetal       float64                 `json:"storage_metal"`
	StorageSilicon     float64                 `json:"storage_silicon"`
	StorageHydrogen    float64                 `json:"storage_hydrogen"`
	DailyMetal         float64                 `json:"daily_metal"`
	DailySilicon       float64                 `json:"daily_silicon"`
	DailyHydrogen      float64                 `json:"daily_hydrogen"`
	WeeklyMetal        float64                 `json:"weekly_metal"`
	WeeklySilicon      float64                 `json:"weekly_silicon"`
	WeeklyHydrogen     float64                 `json:"weekly_hydrogen"`
	TotalEnergy        float64                 `json:"total_energy"`
}

type ResourceBuildingDTO struct {
	UnitID       int     `json:"unit_id"`
	Name         string  `json:"name"`
	Level        int     `json:"level"`
	ProdMetal    float64 `json:"prod_metal"`
	ProdSilicon  float64 `json:"prod_silicon"`
	ProdHydrogen float64 `json:"prod_hydrogen"`
	ConsEnergy   float64 `json:"cons_energy"`
	Factor       int     `json:"factor"`
	AllowFactor  bool    `json:"allow_factor"`
}

// ResourceReport возвращает отчёт о производстве ресурсов для планеты.
func (s *Service) ResourceReport(ctx context.Context, userID, planetID string) (*ResourceReportDTO, error) {
	p, err := s.repo.GetByID(ctx, planetID)
	if err != nil {
		return nil, err
	}
	if p.UserID != userID {
		return nil, ErrNotFound
	}

	// Прочитаем здания планеты.
	buildings, err := s.repo.GetBuildings(ctx, planetID)
	if err != nil {
		return nil, fmt.Errorf("get buildings: %w", err)
	}

	// Уровни исследований игрока — нужны для статических формул производства.
	tech, err := researchLevelsDirect(ctx, s.db.Pool(), p.UserID)
	if err != nil {
		return nil, fmt.Errorf("get research levels: %w", err)
	}
	// Применяем виртуальные уровни профессии к tech-карте.
	applyProfessionBonus(tech, profBonusForUser(ctx, s.db.Pool(), p.UserID, s.catalog))

	report := &ResourceReportDTO{
		PlanetID:   p.ID,
		PlanetName: p.Name,
		Buildings:  []ResourceBuildingDTO{},
	}

	// Базовое (естественное) производство — константа из legacy oxsar2,
	// не зависит от зданий (metal=10/h, silicon=5/h, hydrogen=0/h).
	report.BasicMetal = 10
	report.BasicSilicon = 5
	report.BasicHydrogen = 0

	// Ёмкости хранилищ — пересчитываем по текущим уровням зданий,
	// не из кеша p.MetalCap (который обновляется только при тике).
	lvls := make(map[int]int, len(buildings))
	for _, b := range buildings {
		lvls[b.UnitID] = b.Level
	}
	caps := s.storageCap(&p, lvls)
	report.StorageMetal = caps.metal
	report.StorageSilicon = caps.silicon
	report.StorageHydrogen = caps.hydrogen

	tempC := (p.TempMin + p.TempMax) / 2

	// Расчёт производства по зданиям.
	totalMetal, totalSilicon, totalHydrogen := 0.0, 0.0, 0.0
	totalEnergy := 0.0

	for _, b := range buildings {
		bp := buildingProdStatic(b.UnitID, b.Level, tech, tempC)

		// apply produce_factor для ресурсных шахт.
		metalHour := bp.metal * p.ProduceFactor
		silHour := bp.silicon * p.ProduceFactor
		hydHour := bp.hydrogen * p.ProduceFactor
		netEnergy := bp.energy

		// Применить factor для итоговых сумм.
		factor := float64(b.Factor) / 100.0
		totalMetal += metalHour * factor
		totalSilicon += silHour * factor
		totalHydrogen += hydHour * factor
		totalEnergy += netEnergy * factor

		// allow_factor: здания с ненулевым производством ресурсов или энергии.
		allowFactor := bp.metal != 0 || bp.silicon != 0 || bp.hydrogen != 0 || bp.energy != 0

		buildingName := s.getBuildingName(b.UnitID)

		report.Buildings = append(report.Buildings, ResourceBuildingDTO{
			UnitID:       b.UnitID,
			Name:         buildingName,
			Level:        b.Level,
			ProdMetal:    metalHour,
			ProdSilicon:  silHour,
			ProdHydrogen: hydHour,
			ConsEnergy:   netEnergy,
			Factor:       b.Factor,
			AllowFactor:  allowFactor,
		})
	}

	// Почасовое производство.
	report.MetalPerHour = totalMetal
	report.SiliconPerHour = totalSilicon
	report.HydrogenPerHour = totalHydrogen

	// Сводные значения (текущий запас).
	report.MetalTotal = float64(p.Metal)
	report.SiliconTotal = float64(p.Silicon)
	report.HydrogenTotal = float64(p.Hydrogen)
	report.TotalEnergy = totalEnergy

	// Дневное производство (24 часа).
	report.DailyMetal = report.MetalPerHour * 24
	report.DailySilicon = report.SiliconPerHour * 24
	report.DailyHydrogen = report.HydrogenPerHour * 24

	// Недельное производство (7 дней).
	report.WeeklyMetal = report.DailyMetal * 7
	report.WeeklySilicon = report.DailySilicon * 7
	report.WeeklyHydrogen = report.DailyHydrogen * 7

	return report, nil
}

// UpdateResourceFactors обновляет факторы производства для зданий планеты (батч-операция).
func (s *Service) UpdateResourceFactors(ctx context.Context, userID, planetID string, factors map[string]int) error {
	p, err := s.repo.GetByID(ctx, planetID)
	if err != nil {
		return err
	}
	if p.UserID != userID {
		return ErrNotFound
	}

	// Парсинг и валидация факторов (0-100%).
	intFactors := make(map[int]int)
	for unitIDStr, factor := range factors {
		if factor < 0 || factor > 100 {
			return fmt.Errorf("invalid factor: %d: %w", factor, ErrInvalidInput)
		}
		var unitID int
		if _, err := fmt.Sscanf(unitIDStr, "%d", &unitID); err != nil {
			return fmt.Errorf("invalid unit_id: %s: %w", unitIDStr, ErrInvalidInput)
		}
		intFactors[unitID] = factor
	}

	// Обновить все факторы в одном батч-запросе.
	if err := s.repo.UpdateBuildingFactors(ctx, planetID, intFactors); err != nil {
		return fmt.Errorf("update building factors: %w", err)
	}

	return nil
}

// getBuildingName возвращает имя здания по его ID.
func (s *Service) getBuildingName(unitID int) string {
	for key, spec := range s.catalog.Construction.Buildings {
		if int(spec.ID) == unitID {
			return key
		}
	}
	return ""
}

type buildingProd struct {
	metal    float64
	silicon  float64
	hydrogen float64
	energy   float64 // net = prod - cons
}

// buildingProdStatic возвращает почасовое производство и потребление здания
// через статические формулы economy (без DSL).
func buildingProdStatic(unitID, level int, tech map[int]int, tempC int) buildingProd {
	techE := tech[economy.IDTechEnergy]
	switch unitID {
	case economy.IDMetalmine:
		prod := economy.MetalmineProdMetal(level, tech[economy.IDTechLaser])
		cons := economy.MineConsEnergy(10, level, techE)
		return buildingProd{metal: prod, energy: -cons}
	case economy.IDSiliconLab:
		prod := economy.SiliconLabProdSilicon(level, tech[economy.IDTechSilicon])
		cons := economy.MineConsEnergy(10, level, techE)
		return buildingProd{silicon: prod, energy: -cons}
	case economy.IDHydrogenLab:
		prod := economy.HydrogenLabProdHydrogen(level, tech[economy.IDTechHydrogen], tempC)
		cons := economy.MineConsEnergy(20, level, techE)
		return buildingProd{hydrogen: prod, energy: -cons}
	case economy.IDMoonHydrogenLab:
		prod := economy.MoonHydrogenLabProdHydrogen(level, tech[economy.IDTechHydrogen], tempC)
		cons := economy.MineConsEnergy(200, level, techE)
		return buildingProd{hydrogen: prod, energy: -cons}
	case economy.IDSolarPlant:
		prod := economy.SolarPlantProdEnergy(level, techE)
		return buildingProd{energy: prod}
	case economy.IDHydrogenPlant:
		prod := economy.HydrogenPlantProdEnergy(level, techE)
		cons := economy.HydrogenPlantConsHydrogen(level, techE)
		return buildingProd{hydrogen: -cons, energy: prod}
	case economy.IDSolarSatellite:
		prod := economy.SolarSatelliteProdEnergy(tempC, techE) * float64(level)
		return buildingProd{energy: prod}
	case economy.IDGravi:
		prod := economy.GraviProdEnergy(level, 300000)
		return buildingProd{energy: prod}
	}
	return buildingProd{}
}
