package aiadvisor

import (
	"context"
	"testing"
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

// На этом этапе scoreCandidates — каркас и возвращает пустой slice.
// Тест фиксирует именно это поведение, чтобы при реализации Ф.1в
// (building scoring) рост массива был осознанным изменением.
func TestScoreCandidates_Empty(t *testing.T) {
	t.Parallel()
	snap := PlayerSnapshot{}
	for _, st := range []Strategy{StrategyEconomy, StrategyMilitary, StrategyDefense, StrategyExpansion, StrategyAuto} {
		got := scoreCandidates(snap, st, scoringInputs{})
		if got == nil {
			// nil-slice допустим (sort.SliceStable на nil — no-op),
			// но проверим что итоговая длина 0.
			continue
		}
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
