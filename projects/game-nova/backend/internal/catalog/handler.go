// Package catalog отдаёт публичные описания юнитов из configs/*.yml
// + pre-computed таблицы стоимости/времени/статов для нескольких
// ключевых уровней. Используется origin-фронтом (план 72 Ф.4 Spring 3:
// info-страницы S-013/S-018/S-019, S-021).
//
// Endpoints:
//   GET /api/buildings/catalog/{type}  → BuildingCatalogEntry
//   GET /api/units/catalog/{type}      → UnitCatalogEntry  (ship | defense | research)
//   GET /api/artefacts/catalog/{type}  → ArtefactCatalogEntry
//
// Source of truth — `internal/config.Catalog` (загружен из YAML на старте).
// Формулы стоимости/времени/производства — `internal/economy`.
//
// Universe-context: на 2026-04-28 catalog отдаёт modern (nova) данные.
// Universe-aware override (план 64) и origin-вселенная (план 74) —
// расширение query-param `?universe=...` отдельным планом.
// См. simplifications.md P72.S3.A.
package catalog

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/economy"
	"oxsar/game-nova/internal/httpx"
)

// PreviewLevels — уровни, для которых pre-computed таблица возвращается
// в ответе building/research catalog. Выбраны как «человеческая оптимальная
// плотность»: 1 (старт), 5 (ранний game), 10 (mid-game), 20 (late-game),
// 30 (end-game), 40 (max).
var PreviewLevels = []int{1, 5, 10, 20, 30, 40}

// Handler — единая точка входа для catalog endpoints.
type Handler struct {
	cat *config.Catalog
}

func NewHandler(cat *config.Catalog) *Handler {
	return &Handler{cat: cat}
}

// ---- Building catalog ----

// BuildingCatalogEntry — публичное описание здания (params + pre-computed).
type BuildingCatalogEntry struct {
	ID                int                   `json:"id"`
	Key               string                `json:"key"`
	Name              string                `json:"name"`
	CostBase          ResCost               `json:"cost_base"`
	CostFactor        float64               `json:"cost_factor"`
	TimeBaseSeconds   int                   `json:"time_base_seconds"`
	BaseRatePerHour   *float64              `json:"base_rate_per_hour,omitempty"`
	EnergyPerLevel    *float64              `json:"energy_per_level,omitempty"`
	EnergyOutputPer   *float64              `json:"energy_output_per_level,omitempty"`
	CapacityBase      *int64                `json:"capacity_base,omitempty"`
	MoonOnly          bool                  `json:"moon_only,omitempty"`
	MaxLevel          int                   `json:"max_level"`
	Preview           []BuildingPreviewRow  `json:"preview"`
}

// ResCost — троица металл/кремний/водород.
type ResCost struct {
	Metal    int64 `json:"metal"`
	Silicon  int64 `json:"silicon"`
	Hydrogen int64 `json:"hydrogen"`
}

// BuildingPreviewRow — pre-computed строка таблицы для уровня L.
type BuildingPreviewRow struct {
	Level           int     `json:"level"`
	Cost            ResCost `json:"cost"`
	BuildSeconds    int     `json:"build_seconds"`
	ProductionPerHr float64 `json:"production_per_hour,omitempty"`
	EnergyDemand    float64 `json:"energy_demand,omitempty"`
	EnergyOutput    float64 `json:"energy_output,omitempty"`
}

// BuildingByType GET /api/buildings/catalog/{type}.
// `type` — числовой unit_id ИЛИ строковый key из buildings.yml.
func (h *Handler) BuildingByType(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	typeParam := chi.URLParam(r, "type")
	key, spec, ok := h.lookupBuilding(typeParam)
	if !ok {
		incCatalogReq("building", "not_found")
		httpx.WriteError(w, r, httpx.ErrNotFound)
		return
	}

	out := BuildingCatalogEntry{
		ID:              spec.ID,
		Key:             key,
		Name:            h.unitName(spec.ID, "building"),
		CostBase:        ResCost{Metal: spec.CostBase.Metal, Silicon: spec.CostBase.Silicon, Hydrogen: spec.CostBase.Hydrogen},
		CostFactor:      spec.CostFactor,
		TimeBaseSeconds: spec.TimeBaseSeconds,
		BaseRatePerHour: spec.BaseRatePerHour,
		EnergyPerLevel:  spec.EnergyPerLevel,
		EnergyOutputPer: spec.EnergyOutputPerLevel,
		CapacityBase:    spec.CapacityBase,
		MoonOnly:        spec.MoonOnly,
		MaxLevel:        spec.MaxLevel,
	}
	out.Preview = buildBuildingPreview(spec)
	incCatalogReq("building", "ok")
	httpx.WriteJSON(w, r, http.StatusOK, out)
}

func buildBuildingPreview(spec config.BuildingSpec) []BuildingPreviewRow {
	rows := make([]BuildingPreviewRow, 0, len(PreviewLevels))
	baseCost := economy.Cost{
		Metal:    spec.CostBase.Metal,
		Silicon:  spec.CostBase.Silicon,
		Hydrogen: spec.CostBase.Hydrogen,
	}
	for _, lvl := range PreviewLevels {
		if spec.MaxLevel > 0 && lvl > spec.MaxLevel {
			break
		}
		c := economy.CostForLevelFloor(baseCost, spec.CostFactor, lvl)
		// build_seconds для baseline: robotic=0, nano=0, gameSpeed=1.
		dur := economy.BuildDuration(spec.TimeBaseSeconds, c, 0, 0, 1)
		row := BuildingPreviewRow{
			Level:        lvl,
			Cost:         ResCost{Metal: c.Metal, Silicon: c.Silicon, Hydrogen: c.Hydrogen},
			BuildSeconds: int(dur.Seconds()),
		}
		if spec.BaseRatePerHour != nil {
			row.ProductionPerHr = economy.ProductionPerHour(*spec.BaseRatePerHour, lvl, 1.0)
		}
		if spec.EnergyPerLevel != nil {
			row.EnergyDemand = economy.EnergyDemand(*spec.EnergyPerLevel, lvl)
		}
		if spec.EnergyOutputPerLevel != nil {
			row.EnergyOutput = economy.EnergyOutput(*spec.EnergyOutputPerLevel, lvl)
		}
		rows = append(rows, row)
	}
	return rows
}

// ---- Units catalog (ship | defense | research) ----

// UnitCatalogEntry — публичное описание корабля / обороны / исследования.
type UnitCatalogEntry struct {
	ID         int      `json:"id"`
	Key        string   `json:"key"`
	Name       string   `json:"name"`
	Kind       string   `json:"kind"` // ship | defense | research
	Cost       ResCost  `json:"cost"`
	CostFactor *float64 `json:"cost_factor,omitempty"` // только для research
	Attack     *int     `json:"attack,omitempty"`
	Shield     *int     `json:"shield,omitempty"`
	Shell      *int     `json:"shell,omitempty"`
	Cargo      *int64   `json:"cargo,omitempty"`
	Speed      *int     `json:"speed,omitempty"`
	Fuel       *int     `json:"fuel,omitempty"`
	Front      *int     `json:"front,omitempty"`
	Rapidfire  []RfPair `json:"rapidfire,omitempty"`

	// Для research — preview cost по уровням.
	Preview []ResearchPreviewRow `json:"preview,omitempty"`
}

// RfPair — rapid-fire против target юнита.
type RfPair struct {
	TargetID int `json:"target_id"`
	Multiplier int `json:"multiplier"`
}

// ResearchPreviewRow — pre-computed cost для конкретного уровня research.
type ResearchPreviewRow struct {
	Level int     `json:"level"`
	Cost  ResCost `json:"cost"`
}

// UnitByType GET /api/units/catalog/{type}.
// type = unit_id (числовой) или key.
func (h *Handler) UnitByType(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	typeParam := chi.URLParam(r, "type")

	if key, spec, ok := h.lookupShip(typeParam); ok {
		out := UnitCatalogEntry{
			ID:   spec.ID,
			Key:  key,
			Name: h.unitName(spec.ID, "fleet"),
			Kind: "ship",
			Cost: ResCost{Metal: spec.Cost.Metal, Silicon: spec.Cost.Silicon, Hydrogen: spec.Cost.Hydrogen},
			Attack: ptrInt(spec.Attack),
			Shield: ptrInt(spec.Shield),
			Shell:  ptrInt(spec.Shell),
			Cargo:  ptrInt64(spec.Cargo),
			Speed:  ptrInt(spec.Speed),
			Fuel:   ptrInt(spec.Fuel),
			Front:  ptrIntIfNonZero(spec.Front),
		}
		out.Rapidfire = h.rapidfire(spec.ID)
		incCatalogReq("ship", "ok")
		httpx.WriteJSON(w, r, http.StatusOK, out)
		return
	}
	if key, spec, ok := h.lookupDefense(typeParam); ok {
		out := UnitCatalogEntry{
			ID:   spec.ID,
			Key:  key,
			Name: h.unitName(spec.ID, "defense"),
			Kind: "defense",
			Cost: ResCost{Metal: spec.Cost.Metal, Silicon: spec.Cost.Silicon, Hydrogen: spec.Cost.Hydrogen},
			Attack: ptrInt(spec.Attack),
			Shield: ptrInt(spec.Shield),
			Shell:  ptrInt(spec.Shell),
			Front:  ptrIntIfNonZero(spec.Front),
		}
		incCatalogReq("defense", "ok")
		httpx.WriteJSON(w, r, http.StatusOK, out)
		return
	}
	if key, spec, ok := h.lookupResearch(typeParam); ok {
		factor := spec.CostFactor
		out := UnitCatalogEntry{
			ID:         spec.ID,
			Key:        key,
			Name:       h.unitName(spec.ID, "research"),
			Kind:       "research",
			Cost:       ResCost{Metal: spec.CostBase.Metal, Silicon: spec.CostBase.Silicon, Hydrogen: spec.CostBase.Hydrogen},
			CostFactor: &factor,
		}
		out.Preview = buildResearchPreview(spec)
		incCatalogReq("research", "ok")
		httpx.WriteJSON(w, r, http.StatusOK, out)
		return
	}
	incCatalogReq("unit", "not_found")
	httpx.WriteError(w, r, httpx.ErrNotFound)
}

func buildResearchPreview(spec config.ResearchSpec) []ResearchPreviewRow {
	base := economy.Cost{
		Metal:    spec.CostBase.Metal,
		Silicon:  spec.CostBase.Silicon,
		Hydrogen: spec.CostBase.Hydrogen,
	}
	rows := make([]ResearchPreviewRow, 0, len(PreviewLevels))
	for _, lvl := range PreviewLevels {
		c := economy.CostForLevel(base, spec.CostFactor, lvl)
		rows = append(rows, ResearchPreviewRow{
			Level: lvl,
			Cost:  ResCost{Metal: c.Metal, Silicon: c.Silicon, Hydrogen: c.Hydrogen},
		})
	}
	return rows
}

// ---- Artefact catalog ----

// ArtefactCatalogEntry — публичное описание артефакта.
type ArtefactCatalogEntry struct {
	ID              int                  `json:"id"`
	Key             string               `json:"key"`
	Name            string               `json:"name"`
	Effect          ArtefactEffectEntry  `json:"effect"`
	Stackable       bool                 `json:"stackable"`
	MaxStacks       int                  `json:"max_stacks,omitempty"`
	LifetimeSeconds int                  `json:"lifetime_seconds"`
	DelaySeconds    int                  `json:"delay_seconds,omitempty"`
}

// ArtefactEffectEntry — публичная проекция artefact effect.
type ArtefactEffectEntry struct {
	Type           string  `json:"type"`
	Field          string  `json:"field,omitempty"`
	Op             string  `json:"op,omitempty"`
	Value          float64 `json:"value,omitempty"`
	ActiveValue    float64 `json:"active_value,omitempty"`
	InactiveValue  float64 `json:"inactive_value,omitempty"`
	BattleAttack   float64 `json:"battle_attack,omitempty"`
	BattleShield   float64 `json:"battle_shield,omitempty"`
	BattleShell    float64 `json:"battle_shell,omitempty"`
}

// ArtefactByType GET /api/artefacts/catalog/{type}.
func (h *Handler) ArtefactByType(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	typeParam := chi.URLParam(r, "type")
	key, spec, ok := h.lookupArtefact(typeParam)
	if !ok {
		incCatalogReq("artefact", "not_found")
		httpx.WriteError(w, r, httpx.ErrNotFound)
		return
	}
	out := ArtefactCatalogEntry{
		ID:              spec.ID,
		Key:             key,
		Name:            spec.Name,
		Stackable:       spec.Stackable,
		MaxStacks:       spec.MaxStacks,
		LifetimeSeconds: spec.LifetimeSeconds,
		DelaySeconds:    spec.DelaySeconds,
		Effect: ArtefactEffectEntry{
			Type:          spec.Effect.Type,
			Field:         spec.Effect.Field,
			Op:            spec.Effect.Op,
			Value:         spec.Effect.Value,
			ActiveValue:   spec.Effect.ActiveValue,
			InactiveValue: spec.Effect.InactiveValue,
			BattleAttack:  spec.Effect.BattleAttack,
			BattleShield:  spec.Effect.BattleShield,
			BattleShell:   spec.Effect.BattleShell,
		},
	}
	incCatalogReq("artefact", "ok")
	httpx.WriteJSON(w, r, http.StatusOK, out)
}

// ---- helpers ----

func (h *Handler) lookupBuilding(param string) (string, config.BuildingSpec, bool) {
	if id, err := strconv.Atoi(param); err == nil {
		for k, s := range h.cat.Buildings.Buildings {
			if s.ID == id {
				return k, s, true
			}
		}
	}
	if s, ok := h.cat.Buildings.Buildings[param]; ok {
		return param, s, true
	}
	return "", config.BuildingSpec{}, false
}

func (h *Handler) lookupShip(param string) (string, config.ShipSpec, bool) {
	if id, err := strconv.Atoi(param); err == nil {
		for k, s := range h.cat.Ships.Ships {
			if s.ID == id {
				return k, s, true
			}
		}
	}
	if s, ok := h.cat.Ships.Ships[param]; ok {
		return param, s, true
	}
	return "", config.ShipSpec{}, false
}

func (h *Handler) lookupDefense(param string) (string, config.DefenseSpec, bool) {
	if id, err := strconv.Atoi(param); err == nil {
		for k, s := range h.cat.Defense.Defense {
			if s.ID == id {
				return k, s, true
			}
		}
	}
	if s, ok := h.cat.Defense.Defense[param]; ok {
		return param, s, true
	}
	return "", config.DefenseSpec{}, false
}

func (h *Handler) lookupResearch(param string) (string, config.ResearchSpec, bool) {
	if id, err := strconv.Atoi(param); err == nil {
		for k, s := range h.cat.Research.Research {
			if s.ID == id {
				return k, s, true
			}
		}
	}
	if s, ok := h.cat.Research.Research[param]; ok {
		return param, s, true
	}
	return "", config.ResearchSpec{}, false
}

func (h *Handler) lookupArtefact(param string) (string, config.ArtefactSpec, bool) {
	if id, err := strconv.Atoi(param); err == nil {
		for k, s := range h.cat.Artefacts.Artefacts {
			if s.ID == id {
				return k, s, true
			}
		}
	}
	if s, ok := h.cat.Artefacts.Artefacts[param]; ok {
		return param, s, true
	}
	return "", config.ArtefactSpec{}, false
}

// unitName ищет имя в units.yml (UnitsCatalog) — это «английское» техническое
// имя для отладки, фронт всё равно резолвит i18n.info.{key} по своему словарю.
func (h *Handler) unitName(id int, kind string) string {
	var entries []config.UnitEntry
	switch kind {
	case "building":
		entries = h.cat.Units.Buildings
	case "fleet":
		entries = h.cat.Units.Fleet
	case "defense":
		entries = h.cat.Units.Defense
	case "research":
		entries = h.cat.Units.Research
	}
	for _, e := range entries {
		if e.ID == id {
			return e.Name
		}
	}
	return ""
}

func (h *Handler) rapidfire(shooterID int) []RfPair {
	row, ok := h.cat.Rapidfire.Rapidfire[shooterID]
	if !ok {
		return nil
	}
	out := make([]RfPair, 0, len(row))
	for target, mult := range row {
		out = append(out, RfPair{TargetID: target, Multiplier: mult})
	}
	return out
}

func ptrInt(v int) *int {
	return &v
}
func ptrInt64(v int64) *int64 {
	return &v
}
func ptrIntIfNonZero(v int) *int {
	if v == 0 {
		return nil
	}
	return &v
}
