package aiadvisor

import (
	"context"
	"testing"

	"oxsar/game-nova/internal/config"
)

// План 06.1 Ф.1а: unit-тесты для каркаса автопилота.
// Тесты с реальной БД (Enqueue / Compute / Result / Execute end-to-end)
// добавляются в Ф.1д.

func TestStrategy_Validate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		s    Strategy
		want bool
	}{
		{StrategyEconomy, true},
		{StrategyMilitary, true},
		{StrategyDefense, true},
		{StrategyExpansion, true},
		{StrategyAuto, true},
		{Strategy(""), false},
		{Strategy("foo"), false},
		{Strategy("ECONOMY"), false}, // case-sensitive
	}
	for _, tc := range cases {
		got := tc.s.Validate()
		if got != tc.want {
			t.Errorf("Strategy(%q).Validate() = %v, want %v", string(tc.s), got, tc.want)
		}
	}
}

// scoreCandidates без catalog → пустой slice (нет данных о юнитах).
func TestScoreCandidates_NoCatalog(t *testing.T) {
	t.Parallel()
	snap := PlayerSnapshot{}
	for _, st := range []Strategy{StrategyEconomy, StrategyMilitary, StrategyDefense, StrategyExpansion, StrategyAuto} {
		got := scoreCandidates(snap, st, scoringInputs{})
		if len(got) != 0 {
			t.Errorf("strategy=%s: expected empty slice, got %d items", st, len(got))
		}
	}
}

func TestDetectPhase_Default(t *testing.T) {
	t.Parallel()
	// Ф.5 заменит на эвристику фазы; пока тест фиксирует дефолт.
	got := detectPhase(PlayerSnapshot{})
	if got != StrategyEconomy {
		t.Errorf("detectPhase default = %s, want %s", got, StrategyEconomy)
	}
}

func TestExecuteRecommendation_UnsupportedCategory(t *testing.T) {
	t.Parallel()
	cases := []string{"", "mission", "exchange", "profession", "acs", "unknown"}
	for _, cat := range cases {
		_, err := executeRecommendation(context.Background(),
			executorDeps{}, "user-1",
			Recommendation{Category: cat})
		if err == nil {
			t.Errorf("category=%q: expected error, got nil", cat)
			continue
		}
		// На этом этапе любая категория без service возвращает
		// ErrUnsupportedCategory. После Ф.2 категория "mission" будет
		// поддержана при наличии fleet svc, и тест надо будет обновить.
		if err != ErrUnsupportedCategory {
			t.Errorf("category=%q: want ErrUnsupportedCategory, got %v", cat, err)
		}
	}
}

func TestExecuteRecommendation_BuildingNilService(t *testing.T) {
	t.Parallel()
	// building без deps.Building — должен вернуть ErrUnsupportedCategory,
	// чтобы deps=zero не панически кастился к building.Service.
	_, err := executeRecommendation(context.Background(),
		executorDeps{}, "user-1",
		Recommendation{Category: "building", PlanetID: "p1", UnitID: 1})
	if err != ErrUnsupportedCategory {
		t.Fatalf("nil building svc: want ErrUnsupportedCategory, got %v", err)
	}
}

func TestExecuteRecommendation_ResearchNilService(t *testing.T) {
	t.Parallel()
	_, err := executeRecommendation(context.Background(),
		executorDeps{}, "user-1",
		Recommendation{Category: "research", PlanetID: "p1", UnitID: 1})
	if err != ErrUnsupportedCategory {
		t.Fatalf("nil research svc: want ErrUnsupportedCategory, got %v", err)
	}
}

func TestAutopilotConstants(t *testing.T) {
	t.Parallel()
	// Эти строки попадают в БД и в JSON фронту. Защищаем от случайного
	// переименования — миграция и фронт должны меняться синхронно.
	if AutopilotSource != "autopilot" {
		t.Errorf("AutopilotSource = %q, want %q", AutopilotSource, "autopilot")
	}
	for _, want := range []string{"pending", "ready", "executed", "expired"} {
		switch want {
		case AutopilotStatusPending, AutopilotStatusReady, AutopilotStatusExecuted, AutopilotStatusExpired:
			// ok
		default:
			t.Errorf("missing autopilot status constant for %q", want)
		}
	}
}

func TestAutopilotErrors_NotNil(t *testing.T) {
	t.Parallel()
	for _, err := range []error{
		ErrInvalidStrategy,
		ErrJobNotFound,
		ErrJobNotReady,
		ErrRecommendationNotFound,
		ErrUnsupportedCategory,
		ErrRecommendationStale,
	} {
		if err == nil {
			t.Errorf("sentinel error must not be nil")
		}
	}
}

func TestBuildSnapshot_NilDeps(t *testing.T) {
	t.Parallel()
	// DB/PlanetSvc/ScoreSvc обязательны; их отсутствие — ошибка.
	// BuildingSvc/ResearchSvc опциональны (nil → пустые карты).
	_, err := buildSnapshot(context.Background(), SnapshotDeps{}, "user-1")
	if err == nil {
		t.Fatal("expected error for nil DB/PlanetSvc/ScoreSvc")
	}
}

// TestPlanetSnapshot_DefaultMaps — sanity-check для PlanetSnapshot,
// чтобы scoring не паниковал на nil-картах при тестировании
// scoreCandidates с ручным snapshot'ом.
func TestPlanetSnapshot_DefaultMaps(t *testing.T) {
	t.Parallel()
	ps := PlanetSnapshot{}
	// Buildings объявлено как map → zero-value = nil; чтение не паникует,
	// записи требуют явной инициализации в коде.
	if ps.Buildings != nil {
		t.Errorf("zero PlanetSnapshot.Buildings should be nil, got %v", ps.Buildings)
	}
	// Resources — value-тип, zero = все нули. Это допустимо.
	if ps.Resources.Metal != 0 {
		t.Errorf("zero PlanetSnapshot.Resources.Metal != 0")
	}
}

// fixtureCatalog возвращает минимальный каталог зданий для scoring-тестов:
// metal_mine (производственное) и robotic_factory (без производства).
//
// Параметры приближены к продовым (для проверки реалистичности score),
// но точные значения не критичны — тесты проверяют относительный
// порядок и условия отсева, а не абсолютные числа.
func fixtureCatalog() *config.Catalog {
	rate := 30.0 // metal_mine BaseRatePerHour, как в configs/buildings.yml
	return &config.Catalog{
		Buildings: config.BuildingCatalog{
			Buildings: map[string]config.BuildingSpec{
				"metal_mine": {
					ID:              1,
					CostBase:        config.ResCost{Metal: 60, Silicon: 15},
					CostFactor:      1.5,
					TimeBaseSeconds: 60,
					BaseRatePerHour: &rate,
					MaxLevel:        40,
				},
				"robotic_factory": {
					ID:              14,
					CostBase:        config.ResCost{Metal: 400, Silicon: 120, Hydrogen: 200},
					CostFactor:      2.0,
					TimeBaseSeconds: 60,
					MaxLevel:        20,
				},
				"moon_base": {
					ID:              41,
					CostBase:        config.ResCost{Metal: 20000, Silicon: 40000, Hydrogen: 20000},
					CostFactor:      2.0,
					TimeBaseSeconds: 60,
					MoonOnly:        true,
					MaxLevel:        15,
				},
			},
		},
	}
}

func fixturePlanet(id, name string, freeFields int) PlanetSnapshot {
	return PlanetSnapshot{
		ID:        id,
		Name:      name,
		IsMoon:    false,
		MaxFields: 10 + freeFields,
		UsedFields: 10,
		Resources: Resources{Metal: 1000, Silicon: 1000, Hydrogen: 1000},
		Capacity:  Resources{Metal: 100000, Silicon: 100000, Hydrogen: 100000},
		Rates:     Resources{Metal: 0.1, Silicon: 0.05, Hydrogen: 0.01},
		Buildings: map[int]int{},
		// Прогноз 1ч хватит на постройку metal_mine ур.1.
		ResourcesIn1h: Resources{Metal: 1360, Silicon: 1180, Hydrogen: 1036},
	}
}

func fixtureScoringInputs(cat *config.Catalog, planetIDs []string) scoringInputs {
	bs := make(map[string]map[int]int, len(planetIDs))
	for _, pid := range planetIDs {
		// Фиксированное время 60 сек для всех — упрощение для теста;
		// относительный порядок score не зависит от единого делителя.
		seconds := map[int]int{}
		for _, spec := range cat.Buildings.Buildings {
			seconds[spec.ID] = 60
		}
		bs[pid] = seconds
	}
	return scoringInputs{
		Catalog:              cat,
		PointsK:              config.PointsCoefficients{Building: 0.00005, Research: 0.0005, Unit: 0.002},
		BuildSecondsByPlanet: bs,
	}
}

func TestScoreBuildings_EconomyPicksProducer(t *testing.T) {
	t.Parallel()
	cat := fixtureCatalog()
	planet := fixturePlanet("p1", "Earth", 5)
	planet.Buildings = map[int]int{1: 2, 14: 1} // metal_mine ур.2, robo_factory ур.1
	snap := PlayerSnapshot{Planets: []PlanetSnapshot{planet}}

	recs := scoreCandidates(snap, StrategyEconomy, fixtureScoringInputs(cat, []string{"p1"}))
	if len(recs) == 0 {
		t.Fatal("expected at least one recommendation")
	}
	// metal_mine — производственное → должен быть #1 при Economy.
	if recs[0].UnitID != 1 {
		t.Errorf("Economy first pick = %d, want metal_mine (id=1)", recs[0].UnitID)
	}
	if recs[0].Category != "building" {
		t.Errorf("first rec category = %q, want building", recs[0].Category)
	}
}

func TestScoreBuildings_BusyQueueSkipped(t *testing.T) {
	t.Parallel()
	cat := fixtureCatalog()
	planet := fixturePlanet("p1", "Earth", 5)
	planet.Buildings = map[int]int{1: 2}
	planet.BuildingQueueBusy = true // очередь занята → пропуск всей планеты
	snap := PlayerSnapshot{Planets: []PlanetSnapshot{planet}}

	recs := scoreCandidates(snap, StrategyEconomy, fixtureScoringInputs(cat, []string{"p1"}))
	if len(recs) != 0 {
		t.Errorf("expected empty recs (busy queue), got %d", len(recs))
	}
}

func TestScoreBuildings_MoonOnlyOnPlanetSkipped(t *testing.T) {
	t.Parallel()
	cat := fixtureCatalog()
	planet := fixturePlanet("p1", "Earth", 5)
	planet.Buildings = map[int]int{1: 2}
	snap := PlayerSnapshot{Planets: []PlanetSnapshot{planet}}

	recs := scoreCandidates(snap, StrategyEconomy, fixtureScoringInputs(cat, []string{"p1"}))
	for _, r := range recs {
		if r.UnitID == 41 { // moon_base.ID
			t.Errorf("moon_base should not appear on planet (not moon)")
		}
	}
}

func TestScoreBuildings_NoFreeFieldsSkipped(t *testing.T) {
	t.Parallel()
	cat := fixtureCatalog()
	planet := fixturePlanet("p1", "Earth", 0) // UsedFields == MaxFields
	planet.UsedFields = planet.MaxFields
	planet.Buildings = map[int]int{1: 2}
	snap := PlayerSnapshot{Planets: []PlanetSnapshot{planet}}

	recs := scoreCandidates(snap, StrategyEconomy, fixtureScoringInputs(cat, []string{"p1"}))
	if len(recs) != 0 {
		t.Errorf("expected empty recs (no free fields), got %d", len(recs))
	}
}

func TestScoreBuildings_MaxLevelSkipped(t *testing.T) {
	t.Parallel()
	cat := fixtureCatalog()
	planet := fixturePlanet("p1", "Earth", 5)
	// metal_mine на максимальном уровне.
	planet.Buildings = map[int]int{1: 40}
	planet.ResourcesIn1h = Resources{Metal: 1e15, Silicon: 1e15, Hydrogen: 1e15} // ресурсов навалом
	snap := PlayerSnapshot{Planets: []PlanetSnapshot{planet}}

	recs := scoreCandidates(snap, StrategyEconomy, fixtureScoringInputs(cat, []string{"p1"}))
	for _, r := range recs {
		if r.UnitID == 1 {
			t.Errorf("metal_mine at max level should not be recommended")
		}
	}
}

func TestScoreBuildings_Top3Limit(t *testing.T) {
	t.Parallel()
	// Шесть кандидатов (все доступны) → топ-3.
	cat := fixtureCatalog()
	// Дополнительные здания, чтобы гарантированно > 3 кандидатов.
	cat.Buildings.Buildings["silicon_lab"] = config.BuildingSpec{
		ID: 2, CostBase: config.ResCost{Metal: 48, Silicon: 24}, CostFactor: 1.6,
		TimeBaseSeconds: 60, MaxLevel: 40,
	}
	cat.Buildings.Buildings["hydrogen_synth"] = config.BuildingSpec{
		ID: 3, CostBase: config.ResCost{Metal: 225, Silicon: 75}, CostFactor: 1.5,
		TimeBaseSeconds: 60, MaxLevel: 40,
	}
	cat.Buildings.Buildings["solar_plant"] = config.BuildingSpec{
		ID: 4, CostBase: config.ResCost{Metal: 75, Silicon: 30}, CostFactor: 1.5,
		TimeBaseSeconds: 60, MaxLevel: 40,
	}

	planet := fixturePlanet("p1", "Earth", 5)
	planet.Buildings = map[int]int{1: 1, 2: 1, 3: 1, 4: 1, 14: 1}
	planet.ResourcesIn1h = Resources{Metal: 1e9, Silicon: 1e9, Hydrogen: 1e9}
	snap := PlayerSnapshot{Planets: []PlanetSnapshot{planet}}

	recs := scoreCandidates(snap, StrategyEconomy, fixtureScoringInputs(cat, []string{"p1"}))
	if len(recs) > 3 {
		t.Errorf("topN limit broken: got %d recs", len(recs))
	}
	// recs отсортирован по убыванию score
	for i := 1; i < len(recs); i++ {
		if recs[i-1].Score < recs[i].Score {
			t.Errorf("recs not sorted: idx=%d score=%.3f < idx=%d score=%.3f",
				i-1, recs[i-1].Score, i, recs[i].Score)
		}
	}
}

func TestScoreBuildings_AffordabilityCheck(t *testing.T) {
	t.Parallel()
	cat := fixtureCatalog()
	planet := fixturePlanet("p1", "Earth", 5)
	planet.Buildings = map[int]int{1: 2}
	// Прогноз через 1ч даёт мало — на metal_mine ур.3 должно хватить
	// (cost_metal = 60 * 1.5^2 = 135, cost_silicon = 15 * 1.5^2 ~ 34),
	// но для robotic_factory ур.2 — нет (cost_metal = 400 * 2 = 800).
	planet.ResourcesIn1h = Resources{Metal: 200, Silicon: 100, Hydrogen: 50}
	snap := PlayerSnapshot{Planets: []PlanetSnapshot{planet}}

	recs := scoreCandidates(snap, StrategyEconomy, fixtureScoringInputs(cat, []string{"p1"}))
	// Должен быть только metal_mine (robotic_factory недоступна по ресурсам).
	hasMine, hasRobo := false, false
	for _, r := range recs {
		switch r.UnitID {
		case 1:
			hasMine = true
		case 14:
			hasRobo = true
		}
	}
	if !hasMine {
		t.Error("metal_mine should be affordable")
	}
	if hasRobo {
		t.Error("robotic_factory should NOT fit in 1h forecast")
	}
}

func TestFormatDuration(t *testing.T) {
	t.Parallel()
	cases := []struct {
		seconds int
		want    string
	}{
		{30, "30с"},
		{59, "59с"},
		{60, "1мин"},
		{120, "2мин"},
		{3599, "59мин"},
		{3600, "1ч"},
		{3660, "1ч 1мин"},
		{15600, "4ч 20мин"},
	}
	for _, c := range cases {
		got := formatDuration(c.seconds)
		if got != c.want {
			t.Errorf("formatDuration(%d) = %q, want %q", c.seconds, got, c.want)
		}
	}
}
