// Package universe предоставляет реестр вселенных из configs/universes.yaml.
package universe

// DUPLICATE: этот файл скопирован между Go-модулями oxsar/game-nova,
// oxsar/identity, oxsar/portal и oxsar/billing. При любом изменении
// синхронизируйте КОПИИ:
//   - projects/game-nova/backend/internal/universe/registry.go
//   - projects/portal/backend/internal/universe/registry.go
// Причина дубля: каждый домен — отдельный go.mod, без shared-модуля.

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Universe описывает одну игровую вселенную. План 72.1 часть 12:
// все per-universe параметры баланса хранятся здесь — это единственный
// источник истины. Никаких env-переопределений в compose.
type Universe struct {
	ID          string `yaml:"id"          json:"id"`
	Name        string `yaml:"name"        json:"name"`
	Description string `yaml:"description" json:"description"`
	Subdomain   string `yaml:"subdomain"   json:"subdomain"`
	// DevURL — full URL фронтенда в dev-окружении (план 36 Ф.8).
	// Если задано, switcher редиректит сюда вместо production
	// https://<subdomain>.oxsar-nova.ru. На проде — пустая строка.
	DevURL string `yaml:"dev_url"     json:"dev_url,omitempty"`
	Status string `yaml:"status"      json:"status"` // active|maintenance|upcoming|retired

	// Игровые параметры (все обязательны в YAML — fail-fast при загрузке,
	// см. validate() ниже).
	Speed                  float64 `yaml:"speed"                     json:"speed"`
	Deathmatch             bool    `yaml:"deathmatch"                json:"deathmatch"`
	MaxPlanets             int     `yaml:"max_planets"               json:"max_planets"`
	BashingPeriod          int     `yaml:"bashing_period"            json:"bashing_period"`
	BashingMaxAttacks      int     `yaml:"bashing_max_attacks"       json:"bashing_max_attacks"`
	ProtectionPeriod       int     `yaml:"protection_period"         json:"protection_period"`
	NumGalaxies            int     `yaml:"num_galaxies"              json:"num_galaxies"`
	NumSystems             int     `yaml:"num_systems"               json:"num_systems"`
	StorageFactor          float64 `yaml:"storage_factor"            json:"-"`
	ResearchSpeedFactor    float64 `yaml:"research_speed_factor"     json:"-"`
	EnergyProductionFactor float64 `yaml:"energy_production_factor"  json:"-"`
	TeleportCostOxsars     int64   `yaml:"teleport_cost_oxsars"      json:"-"`
	TeleportCooldownHours  int     `yaml:"teleport_cooldown_hours"   json:"-"`
	TeleportDurationMin    int     `yaml:"teleport_duration_minutes" json:"-"`

	LaunchedAt time.Time `yaml:"launched_at"  json:"launched_at"`

	// Поля, заполняемые при рантайм-обогащении (не из YAML).
	OnlinePlayers int `yaml:"-" json:"online_players,omitempty"`
	TotalPlayers  int `yaml:"-" json:"total_players,omitempty"`
}

// universeRaw — DTO для YAML-парсинга. Указатели для числовых/булевых
// полей различают «отсутствует в YAML» от «явно 0/false». Без этого
// невозможно отличить max_planets:0 (валидное «computer_tech+1») от
// «забыли указать поле».
type universeRaw struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Subdomain   string `yaml:"subdomain"`
	DevURL      string `yaml:"dev_url"`
	Status      string `yaml:"status"`

	Speed                  *float64 `yaml:"speed"`
	Deathmatch             *bool    `yaml:"deathmatch"`
	MaxPlanets             *int     `yaml:"max_planets"`
	BashingPeriod          *int     `yaml:"bashing_period"`
	BashingMaxAttacks      *int     `yaml:"bashing_max_attacks"`
	ProtectionPeriod       *int     `yaml:"protection_period"`
	NumGalaxies            *int     `yaml:"num_galaxies"`
	NumSystems             *int     `yaml:"num_systems"`
	StorageFactor          *float64 `yaml:"storage_factor"`
	ResearchSpeedFactor    *float64 `yaml:"research_speed_factor"`
	EnergyProductionFactor *float64 `yaml:"energy_production_factor"`
	TeleportCostOxsars     *int64   `yaml:"teleport_cost_oxsars"`
	TeleportCooldownHours  *int     `yaml:"teleport_cooldown_hours"`
	TeleportDurationMin    *int     `yaml:"teleport_duration_minutes"`

	LaunchedAt time.Time `yaml:"launched_at"`
}

type registryFile struct {
	Universes []universeRaw `yaml:"universes"`
}

// Load читает конфиг вселенных из YAML-файла. Все 15 балансных полей
// обязательны хотя бы в одной вселенной (uni01 как fallback-эталон) —
// для остальных при отсутствии поля применяется fallback-merge с uni01.
// Если поле отсутствует и в uni01 — fail-fast.
func Load(path string) ([]Universe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("universe: read %s: %w", path, err)
	}
	var rf registryFile
	if err := yaml.Unmarshal(data, &rf); err != nil {
		return nil, fmt.Errorf("universe: parse %s: %w", path, err)
	}
	if len(rf.Universes) == 0 {
		return nil, fmt.Errorf("universe: %s contains no universes", path)
	}

	// uni01 — fallback-эталон для 6 новых полей. Сначала находим её.
	var uni01 *universeRaw
	for i := range rf.Universes {
		if rf.Universes[i].ID == "uni01" {
			uni01 = &rf.Universes[i]
			break
		}
	}
	if uni01 == nil {
		return nil, fmt.Errorf("universe: %s: uni01 not found (fallback-эталон обязателен)", path)
	}

	out := make([]Universe, 0, len(rf.Universes))
	for _, raw := range rf.Universes {
		u, err := buildUniverse(raw, uni01)
		if err != nil {
			return nil, fmt.Errorf("universe %q: %w", raw.ID, err)
		}
		out = append(out, u)
	}
	return out, nil
}

// buildUniverse материализует Universe из raw, применяя fallback на
// uni01-поля для 6 новых параметров (StorageFactor, ResearchSpeedFactor,
// EnergyProductionFactor, TeleportCostOxsars, TeleportCooldownHours,
// TeleportDurationMin), и валидирует остальные 9 как required.
func buildUniverse(r universeRaw, fallback *universeRaw) (Universe, error) {
	if r.ID == "" {
		return Universe{}, fmt.Errorf("id: required")
	}
	if r.Name == "" {
		return Universe{}, fmt.Errorf("name: required")
	}
	if r.Status == "" {
		return Universe{}, fmt.Errorf("status: required")
	}

	// Required без fallback (специфичные для каждой вселенной).
	speed, err := requireFloat(r.Speed, "speed")
	if err != nil {
		return Universe{}, err
	}
	dm, err := requireBool(r.Deathmatch, "deathmatch")
	if err != nil {
		return Universe{}, err
	}
	maxPlanets, err := requireInt(r.MaxPlanets, "max_planets")
	if err != nil {
		return Universe{}, err
	}
	bashingPeriod, err := requireInt(r.BashingPeriod, "bashing_period")
	if err != nil {
		return Universe{}, err
	}
	bashingMax, err := requireInt(r.BashingMaxAttacks, "bashing_max_attacks")
	if err != nil {
		return Universe{}, err
	}
	protection, err := requireInt(r.ProtectionPeriod, "protection_period")
	if err != nil {
		return Universe{}, err
	}
	numGal, err := requireInt(r.NumGalaxies, "num_galaxies")
	if err != nil {
		return Universe{}, err
	}
	numSys, err := requireInt(r.NumSystems, "num_systems")
	if err != nil {
		return Universe{}, err
	}

	// Required с fallback на uni01 (общие балансные коэффициенты, обычно
	// одинаковые — в uni01 указаны явно, остальные наследуют).
	storageF, err := requireFloatFallback(r.StorageFactor, fallback.StorageFactor, "storage_factor", r.ID == "uni01")
	if err != nil {
		return Universe{}, err
	}
	researchF, err := requireFloatFallback(r.ResearchSpeedFactor, fallback.ResearchSpeedFactor, "research_speed_factor", r.ID == "uni01")
	if err != nil {
		return Universe{}, err
	}
	energyF, err := requireFloatFallback(r.EnergyProductionFactor, fallback.EnergyProductionFactor, "energy_production_factor", r.ID == "uni01")
	if err != nil {
		return Universe{}, err
	}
	teleCost, err := requireInt64Fallback(r.TeleportCostOxsars, fallback.TeleportCostOxsars, "teleport_cost_oxsars", r.ID == "uni01")
	if err != nil {
		return Universe{}, err
	}
	teleCooldown, err := requireIntFallback(r.TeleportCooldownHours, fallback.TeleportCooldownHours, "teleport_cooldown_hours", r.ID == "uni01")
	if err != nil {
		return Universe{}, err
	}
	teleDur, err := requireIntFallback(r.TeleportDurationMin, fallback.TeleportDurationMin, "teleport_duration_minutes", r.ID == "uni01")
	if err != nil {
		return Universe{}, err
	}

	return Universe{
		ID:                     r.ID,
		Name:                   r.Name,
		Description:            r.Description,
		Subdomain:              r.Subdomain,
		DevURL:                 r.DevURL,
		Status:                 r.Status,
		Speed:                  speed,
		Deathmatch:             dm,
		MaxPlanets:             maxPlanets,
		BashingPeriod:          bashingPeriod,
		BashingMaxAttacks:      bashingMax,
		ProtectionPeriod:       protection,
		NumGalaxies:            numGal,
		NumSystems:             numSys,
		StorageFactor:          storageF,
		ResearchSpeedFactor:    researchF,
		EnergyProductionFactor: energyF,
		TeleportCostOxsars:     teleCost,
		TeleportCooldownHours:  teleCooldown,
		TeleportDurationMin:    teleDur,
		LaunchedAt:             r.LaunchedAt,
	}, nil
}

func requireFloat(v *float64, name string) (float64, error) {
	if v == nil {
		return 0, fmt.Errorf("%s: required", name)
	}
	return *v, nil
}

func requireInt(v *int, name string) (int, error) {
	if v == nil {
		return 0, fmt.Errorf("%s: required", name)
	}
	return *v, nil
}

func requireBool(v *bool, name string) (bool, error) {
	if v == nil {
		return false, fmt.Errorf("%s: required", name)
	}
	return *v, nil
}

// requireFloatFallback — если значение задано → берём его; если nil и
// это НЕ эталонная вселенная → берём fallback (значение из uni01); если
// nil и это эталон → ошибка.
func requireFloatFallback(v, fallback *float64, name string, isFallbackUniverse bool) (float64, error) {
	if v != nil {
		return *v, nil
	}
	if isFallbackUniverse {
		return 0, fmt.Errorf("%s: required (uni01 — fallback-эталон)", name)
	}
	if fallback == nil {
		return 0, fmt.Errorf("%s: required (отсутствует и в uni01)", name)
	}
	return *fallback, nil
}

func requireIntFallback(v, fallback *int, name string, isFallbackUniverse bool) (int, error) {
	if v != nil {
		return *v, nil
	}
	if isFallbackUniverse {
		return 0, fmt.Errorf("%s: required (uni01 — fallback-эталон)", name)
	}
	if fallback == nil {
		return 0, fmt.Errorf("%s: required (отсутствует и в uni01)", name)
	}
	return *fallback, nil
}

func requireInt64Fallback(v, fallback *int64, name string, isFallbackUniverse bool) (int64, error) {
	if v != nil {
		return *v, nil
	}
	if isFallbackUniverse {
		return 0, fmt.Errorf("%s: required (uni01 — fallback-эталон)", name)
	}
	if fallback == nil {
		return 0, fmt.Errorf("%s: required (отсутствует и в uni01)", name)
	}
	return *fallback, nil
}

// Registry хранит загруженный список вселенных.
type Registry struct {
	universes []Universe
}

// NewRegistry создаёт реестр из файла.
func NewRegistry(path string) (*Registry, error) {
	us, err := Load(path)
	if err != nil {
		return nil, err
	}
	return &Registry{universes: us}, nil
}

// NewRegistryFromSlice создаёт реестр из готового среза (для тестов и fallback).
func NewRegistryFromSlice(us []Universe) (*Registry, error) {
	if us == nil {
		us = []Universe{}
	}
	return &Registry{universes: us}, nil
}

// All возвращает все вселенные.
func (r *Registry) All() []Universe {
	return r.universes
}

// ByID ищет вселенную по идентификатору.
func (r *Registry) ByID(id string) (Universe, bool) {
	for _, u := range r.universes {
		if u.ID == id {
			return u, true
		}
	}
	return Universe{}, false
}

// Active возвращает вселенные со статусом "active".
func (r *Registry) Active() []Universe {
	var out []Universe
	for _, u := range r.universes {
		if u.Status == "active" {
			out = append(out, u)
		}
	}
	return out
}
