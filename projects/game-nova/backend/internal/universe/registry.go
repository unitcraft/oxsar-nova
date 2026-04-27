// Package universe предоставляет реестр вселенных из configs/universes.yaml.
package universe

// DUPLICATE: этот файл скопирован между Go-модулями oxsar/game-nova,
// oxsar/auth, oxsar/portal и oxsar/billing. При любом изменении
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

// Universe описывает одну игровую вселенную.
type Universe struct {
	ID                string    `yaml:"id"         json:"id"`
	Name              string    `yaml:"name"        json:"name"`
	Description       string    `yaml:"description" json:"description"`
	Subdomain         string    `yaml:"subdomain"   json:"subdomain"`
	Status            string    `yaml:"status"      json:"status"` // active|maintenance|upcoming|retired
	Speed             float64   `yaml:"speed"       json:"speed"`
	Deathmatch        bool      `yaml:"deathmatch"  json:"deathmatch"`
	MaxPlanets        int       `yaml:"max_planets" json:"max_planets"`
	BashingPeriod     int       `yaml:"bashing_period"      json:"bashing_period"`
	BashingMaxAttacks int       `yaml:"bashing_max_attacks" json:"bashing_max_attacks"`
	ProtectionPeriod  int       `yaml:"protection_period"   json:"protection_period"`
	NumGalaxies       int       `yaml:"num_galaxies" json:"num_galaxies"`
	NumSystems        int       `yaml:"num_systems"  json:"num_systems"`
	LaunchedAt        time.Time `yaml:"launched_at"  json:"launched_at"`

	// Поля, заполняемые при рантайм-обогащении (не из YAML).
	OnlinePlayers int `yaml:"-" json:"online_players,omitempty"`
	TotalPlayers  int `yaml:"-" json:"total_players,omitempty"`
}

type registryFile struct {
	Universes []Universe `yaml:"universes"`
}

// Load читает конфиг вселенных из YAML-файла.
func Load(path string) ([]Universe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("universe: read %s: %w", path, err)
	}
	var rf registryFile
	if err := yaml.Unmarshal(data, &rf); err != nil {
		return nil, fmt.Errorf("universe: parse %s: %w", path, err)
	}
	return rf.Universes, nil
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
