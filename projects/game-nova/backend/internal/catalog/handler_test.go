package catalog_test

// Тесты catalog endpoint'ов (план 72 Ф.4 Spring 3).
//
// Catalog handler не делает IO кроме чтения in-memory config.Catalog —
// поэтому тесты полные round-trip без БД.
//
// Покрытие:
//   - 401 без userID в контексте.
//   - 404 для неизвестного type-параметра.
//   - 200 + структура для известного building/ship/defense/research/artefact
//     с проверкой preview-таблицы для buildings/research.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"oxsar/game-nova/internal/auth/authtest"
	"oxsar/game-nova/internal/catalog"
	"oxsar/game-nova/internal/config"
)

func newCat() *config.Catalog {
	mineRate := 30.0
	mineEnergy := 10.0
	solarEnergy := 20.0
	cap := int64(5000)
	cargo := int64(5000)
	return &config.Catalog{
		Units: config.UnitsCatalog{
			Buildings: []config.UnitEntry{
				{ID: 1, Key: "metal_mine", Name: "Metal Mine"},
				{ID: 4, Key: "solar_plant", Name: "Solar Plant"},
				{ID: 9, Key: "metal_storage", Name: "Metal Storage"},
			},
			Fleet: []config.UnitEntry{
				{ID: 202, Key: "small_transporter", Name: "Small Transporter"},
			},
			Defense: []config.UnitEntry{
				{ID: 401, Key: "rocket_launcher", Name: "Rocket Launcher"},
			},
			Research: []config.UnitEntry{
				{ID: 14, Key: "computer_tech", Name: "Computer Tech"},
			},
		},
		Buildings: config.BuildingCatalog{
			Buildings: map[string]config.BuildingSpec{
				"metal_mine": {
					ID:              1,
					CostBase:        config.ResCost{Metal: 60, Silicon: 15, Hydrogen: 0},
					CostFactor:      1.5,
					TimeBaseSeconds: 60,
					BaseRatePerHour: &mineRate,
					EnergyPerLevel:  &mineEnergy,
					MaxLevel:        40,
				},
				"solar_plant": {
					ID:                   4,
					CostBase:             config.ResCost{Metal: 75, Silicon: 30, Hydrogen: 0},
					CostFactor:           1.5,
					TimeBaseSeconds:      60,
					EnergyOutputPerLevel: &solarEnergy,
					MaxLevel:             40,
				},
				"metal_storage": {
					ID:              9,
					CostBase:        config.ResCost{Metal: 1000, Silicon: 0, Hydrogen: 0},
					CostFactor:      2.0,
					TimeBaseSeconds: 60,
					CapacityBase:    &cap,
					MaxLevel:        40,
				},
			},
		},
		Ships: config.ShipCatalog{
			Ships: map[string]config.ShipSpec{
				"small_transporter": {
					ID:     202,
					Attack: 5,
					Shield: 10,
					Shell:  4000,
					Cargo:  cargo,
					Speed:  5000,
					Fuel:   10,
					Cost:   config.ResCost{Metal: 2000, Silicon: 2000, Hydrogen: 0},
				},
			},
		},
		Defense: config.DefenseCatalog{
			Defense: map[string]config.DefenseSpec{
				"rocket_launcher": {
					ID:     401,
					Attack: 80,
					Shield: 20,
					Shell:  200,
					Cost:   config.ResCost{Metal: 2000, Silicon: 0, Hydrogen: 0},
				},
			},
		},
		Research: config.ResearchCatalog{
			Research: map[string]config.ResearchSpec{
				"computer_tech": {
					ID:         14,
					CostBase:   config.ResCost{Metal: 0, Silicon: 400, Hydrogen: 600},
					CostFactor: 2.0,
				},
			},
		},
		Rapidfire: config.RapidfireCatalog{
			Rapidfire: map[int]map[int]int{
				202: {210: 5},
			},
		},
		Artefacts: config.ArtefactCatalog{
			Artefacts: map[string]config.ArtefactSpec{
				"catalyst": {
					ID:              3001,
					Name:            "Catalyst",
					Stackable:       false,
					LifetimeSeconds: 604800,
					Effect: config.ArtefactEffect{
						Type:  "factor_all_planets",
						Field: "produce_factor",
						Op:    "add",
						Value: 0.10,
					},
				},
			},
		},
	}
}

// callRouted запускает chi-router и эмулирует запрос с {type} URL-param.
func callRouted(t *testing.T, h func(http.ResponseWriter, *http.Request),
	method, pattern, urlPath string, ctx context.Context) *httptest.ResponseRecorder {
	t.Helper()
	r := chi.NewRouter()
	r.Method(method, pattern, http.HandlerFunc(h))
	req := httptest.NewRequest(method, urlPath, nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func authedCtx() context.Context {
	return authtest.WithUserID(context.Background(),
		"00000000-0000-0000-0000-000000000001")
}

func TestBuilding_Unauthorized(t *testing.T) {
	t.Parallel()
	h := catalog.NewHandler(newCat())
	rr := callRouted(t, h.BuildingByType, http.MethodGet,
		"/api/buildings/catalog/{type}", "/api/buildings/catalog/1", context.Background())
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("got %d want 401: %s", rr.Code, rr.Body.String())
	}
}

func TestBuilding_NotFound(t *testing.T) {
	t.Parallel()
	h := catalog.NewHandler(newCat())
	rr := callRouted(t, h.BuildingByType, http.MethodGet,
		"/api/buildings/catalog/{type}", "/api/buildings/catalog/9999", authedCtx())
	if rr.Code != http.StatusNotFound {
		t.Fatalf("got %d want 404: %s", rr.Code, rr.Body.String())
	}
}

func TestBuilding_OK_ByID(t *testing.T) {
	t.Parallel()
	h := catalog.NewHandler(newCat())
	rr := callRouted(t, h.BuildingByType, http.MethodGet,
		"/api/buildings/catalog/{type}", "/api/buildings/catalog/1", authedCtx())
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want 200: %s", rr.Code, rr.Body.String())
	}
	var entry catalog.BuildingCatalogEntry
	if err := json.Unmarshal(rr.Body.Bytes(), &entry); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if entry.ID != 1 || entry.Key != "metal_mine" {
		t.Fatalf("entry: %+v", entry)
	}
	if entry.CostBase.Metal != 60 || entry.CostFactor != 1.5 {
		t.Fatalf("cost mismatch: %+v", entry)
	}
	// Preview: уровень 1 должен иметь cost_base, более высокие — растущая стоимость.
	if len(entry.Preview) == 0 {
		t.Fatalf("empty preview")
	}
	if entry.Preview[0].Level != 1 || entry.Preview[0].Cost.Metal != 60 {
		t.Fatalf("preview[0]: %+v", entry.Preview[0])
	}
	// Preview-уровень 5 — стоимость floor(60 * 1.5^4) = 303 (не 60).
	var lvl5 *catalog.BuildingPreviewRow
	for i := range entry.Preview {
		if entry.Preview[i].Level == 5 {
			lvl5 = &entry.Preview[i]
			break
		}
	}
	if lvl5 == nil {
		t.Fatalf("preview level 5 missing")
	}
	if lvl5.Cost.Metal <= entry.Preview[0].Cost.Metal {
		t.Fatalf("level5 cost not growing: %d vs %d", lvl5.Cost.Metal, entry.Preview[0].Cost.Metal)
	}
	if lvl5.ProductionPerHr <= 0 {
		t.Fatalf("metal_mine should have production_per_hour > 0; got %v", lvl5.ProductionPerHr)
	}
}

func TestBuilding_OK_ByKey(t *testing.T) {
	t.Parallel()
	h := catalog.NewHandler(newCat())
	rr := callRouted(t, h.BuildingByType, http.MethodGet,
		"/api/buildings/catalog/{type}", "/api/buildings/catalog/solar_plant", authedCtx())
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want 200: %s", rr.Code, rr.Body.String())
	}
	var entry catalog.BuildingCatalogEntry
	if err := json.Unmarshal(rr.Body.Bytes(), &entry); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if entry.Key != "solar_plant" {
		t.Fatalf("key=%q", entry.Key)
	}
	// Solar plant должен иметь energy_output > 0 в preview (не demand).
	var any bool
	for _, p := range entry.Preview {
		if p.EnergyOutput > 0 {
			any = true
			break
		}
	}
	if !any {
		t.Fatalf("solar_plant preview should report energy_output for some level")
	}
}

func TestUnit_Ship_OK(t *testing.T) {
	t.Parallel()
	h := catalog.NewHandler(newCat())
	rr := callRouted(t, h.UnitByType, http.MethodGet,
		"/api/units/catalog/{type}", "/api/units/catalog/202", authedCtx())
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want 200: %s", rr.Code, rr.Body.String())
	}
	var entry catalog.UnitCatalogEntry
	if err := json.Unmarshal(rr.Body.Bytes(), &entry); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if entry.Kind != "ship" || entry.ID != 202 {
		t.Fatalf("entry: %+v", entry)
	}
	if entry.Attack == nil || *entry.Attack != 5 {
		t.Fatalf("attack: %+v", entry.Attack)
	}
	if len(entry.Rapidfire) != 1 || entry.Rapidfire[0].TargetID != 210 {
		t.Fatalf("rapidfire: %+v", entry.Rapidfire)
	}
}

func TestUnit_Defense_OK(t *testing.T) {
	t.Parallel()
	h := catalog.NewHandler(newCat())
	rr := callRouted(t, h.UnitByType, http.MethodGet,
		"/api/units/catalog/{type}", "/api/units/catalog/rocket_launcher", authedCtx())
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want 200: %s", rr.Code, rr.Body.String())
	}
	var entry catalog.UnitCatalogEntry
	if err := json.Unmarshal(rr.Body.Bytes(), &entry); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if entry.Kind != "defense" {
		t.Fatalf("kind=%q", entry.Kind)
	}
}

func TestUnit_Research_OK_WithPreview(t *testing.T) {
	t.Parallel()
	h := catalog.NewHandler(newCat())
	rr := callRouted(t, h.UnitByType, http.MethodGet,
		"/api/units/catalog/{type}", "/api/units/catalog/14", authedCtx())
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want 200: %s", rr.Code, rr.Body.String())
	}
	var entry catalog.UnitCatalogEntry
	if err := json.Unmarshal(rr.Body.Bytes(), &entry); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if entry.Kind != "research" {
		t.Fatalf("kind=%q", entry.Kind)
	}
	if entry.CostFactor == nil || *entry.CostFactor != 2.0 {
		t.Fatalf("cost_factor: %+v", entry.CostFactor)
	}
	if len(entry.Preview) == 0 {
		t.Fatalf("research preview missing")
	}
	if entry.Preview[0].Cost.Silicon != 400 {
		t.Fatalf("preview[0] silicon = %d, want 400", entry.Preview[0].Cost.Silicon)
	}
}

func TestUnit_NotFound(t *testing.T) {
	t.Parallel()
	h := catalog.NewHandler(newCat())
	rr := callRouted(t, h.UnitByType, http.MethodGet,
		"/api/units/catalog/{type}", "/api/units/catalog/9999", authedCtx())
	if rr.Code != http.StatusNotFound {
		t.Fatalf("got %d want 404: %s", rr.Code, rr.Body.String())
	}
}

func TestArtefact_OK(t *testing.T) {
	t.Parallel()
	h := catalog.NewHandler(newCat())
	rr := callRouted(t, h.ArtefactByType, http.MethodGet,
		"/api/artefacts/catalog/{type}", "/api/artefacts/catalog/3001", authedCtx())
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want 200: %s", rr.Code, rr.Body.String())
	}
	var entry catalog.ArtefactCatalogEntry
	if err := json.Unmarshal(rr.Body.Bytes(), &entry); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if entry.ID != 3001 || entry.Key != "catalyst" {
		t.Fatalf("entry: %+v", entry)
	}
	if entry.Effect.Type != "factor_all_planets" || entry.Effect.Value != 0.10 {
		t.Fatalf("effect: %+v", entry.Effect)
	}
}

func TestArtefact_NotFound(t *testing.T) {
	t.Parallel()
	h := catalog.NewHandler(newCat())
	rr := callRouted(t, h.ArtefactByType, http.MethodGet,
		"/api/artefacts/catalog/{type}", "/api/artefacts/catalog/9999", authedCtx())
	if rr.Code != http.StatusNotFound {
		t.Fatalf("got %d want 404: %s", rr.Code, rr.Body.String())
	}
}
