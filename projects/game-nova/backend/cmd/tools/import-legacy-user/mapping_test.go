package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// configsRoot ищет projects/game-nova/configs/ относительно cwd теста.
// go test запускается в директории пакета, поэтому идём ../../../../configs.
func configsRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// .../projects/game-nova/backend/cmd/tools/import-legacy-user → .../configs
	return filepath.Join(wd, "..", "..", "..", "..", "configs")
}

type catalogEntry struct {
	ID  int    `yaml:"id"`
	Key string `yaml:"key"`
}

// loadIDsFromUnitsYAML читает projects/game-nova/configs/units.yml и
// возвращает множество всех id юнитов по группам. units.yml — единый
// источник истины для всех unit_id (см. файл-комментарий в YAML).
func loadIDsFromUnitsYAML(t *testing.T) map[string]map[int]bool {
	t.Helper()
	path := filepath.Join(configsRoot(t), "units.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read units.yml: %v", err)
	}
	type doc struct {
		Buildings     []catalogEntry `yaml:"buildings"`
		MoonBuildings []catalogEntry `yaml:"moon_buildings"`
		Research      []catalogEntry `yaml:"research"`
		Fleet         []catalogEntry `yaml:"fleet"`
		Defense       []catalogEntry `yaml:"defense"`
	}
	var d doc
	if err := yaml.Unmarshal(raw, &d); err != nil {
		t.Fatalf("parse units.yml: %v", err)
	}
	return map[string]map[int]bool{
		"buildings":      asSet(d.Buildings),
		"moon_buildings": asSet(d.MoonBuildings),
		"research":       asSet(d.Research),
		"fleet":          asSet(d.Fleet),
		"defense":        asSet(d.Defense),
	}
}

func asSet(es []catalogEntry) map[int]bool {
	m := make(map[int]bool, len(es))
	for _, e := range es {
		m[e.ID] = true
	}
	return m
}

// loadArtefactIDs парсит artefacts.yml и собирает множество id'ов.
// Структура: top-level `artefacts:` map, каждый ключ — артефакт со полем `id`.
func loadArtefactIDs(t *testing.T) map[int]bool {
	t.Helper()
	path := filepath.Join(configsRoot(t), "artefacts.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read artefacts.yml: %v", err)
	}
	type artefactEntry struct {
		ID int `yaml:"id"`
	}
	type doc struct {
		Artefacts map[string]artefactEntry `yaml:"artefacts"`
	}
	var d doc
	if err := yaml.Unmarshal(raw, &d); err != nil {
		t.Fatalf("parse artefacts.yml: %v", err)
	}
	m := make(map[int]bool, len(d.Artefacts))
	for _, a := range d.Artefacts {
		m[a.ID] = true
	}
	return m
}

// loadProfessionKeys парсит professions.yml и возвращает все известные ключи.
func loadProfessionKeys(t *testing.T) map[string]bool {
	t.Helper()
	path := filepath.Join(configsRoot(t), "professions.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read professions.yml: %v", err)
	}
	type doc struct {
		Professions map[string]any `yaml:"professions"`
	}
	var d doc
	if err := yaml.Unmarshal(raw, &d); err != nil {
		t.Fatalf("parse professions.yml: %v", err)
	}
	m := make(map[string]bool, len(d.Professions)+1)
	for k := range d.Professions {
		m[k] = true
	}
	// "none" — допустимое значение DEFAULT в схеме (см. миграция 0046),
	// в professions.yml не описан как отдельная запись.
	m["none"] = true
	return m
}

func TestBuildingMapping_TargetsExistInCatalog(t *testing.T) {
	groups := loadIDsFromUnitsYAML(t)
	for legacyID, novaID := range BuildingMapping {
		if !groups["buildings"][novaID] {
			t.Errorf("BuildingMapping[%d]=%d: nova unit_id not present in configs/units.yml buildings",
				legacyID, novaID)
		}
	}
}

func TestMoonBuildingMapping_TargetsExistInCatalog(t *testing.T) {
	groups := loadIDsFromUnitsYAML(t)
	for legacyID, novaID := range MoonBuildingMapping {
		if !groups["moon_buildings"][novaID] {
			t.Errorf("MoonBuildingMapping[%d]=%d: nova unit_id not present in configs/units.yml moon_buildings",
				legacyID, novaID)
		}
	}
}

func TestResearchMapping_TargetsExistInCatalog(t *testing.T) {
	groups := loadIDsFromUnitsYAML(t)
	for legacyID, novaID := range ResearchMapping {
		if !groups["research"][novaID] {
			t.Errorf("ResearchMapping[%d]=%d: nova unit_id not present in configs/units.yml research",
				legacyID, novaID)
		}
	}
}

func TestFleetMapping_TargetsExistInCatalog(t *testing.T) {
	groups := loadIDsFromUnitsYAML(t)
	for legacyID, novaID := range FleetMapping {
		if !groups["fleet"][novaID] {
			t.Errorf("FleetMapping[%d]=%d: nova unit_id not present in configs/units.yml fleet",
				legacyID, novaID)
		}
	}
}

func TestDefenseMapping_TargetsExistInCatalog(t *testing.T) {
	groups := loadIDsFromUnitsYAML(t)
	for legacyID, novaID := range DefenseMapping {
		if !groups["defense"][novaID] {
			t.Errorf("DefenseMapping[%d]=%d: nova unit_id not present in configs/units.yml defense",
				legacyID, novaID)
		}
	}
}

func TestRocketMapping_TargetsExistInDefense(t *testing.T) {
	// Ракеты (51, 52) в nova живут в defense-секции units.yml (см.
	// комментарий в configs/units.yml ~line 96).
	groups := loadIDsFromUnitsYAML(t)
	for legacyID, novaID := range RocketMapping {
		if !groups["defense"][novaID] {
			t.Errorf("RocketMapping[%d]=%d: nova unit_id not present in configs/units.yml defense",
				legacyID, novaID)
		}
	}
}

func TestPlanetShieldMapping_TargetsExistInDefense(t *testing.T) {
	groups := loadIDsFromUnitsYAML(t)
	for legacyID, novaID := range PlanetShieldMapping {
		if !groups["defense"][novaID] {
			t.Errorf("PlanetShieldMapping[%d]=%d: nova unit_id not present in configs/units.yml defense",
				legacyID, novaID)
		}
	}
}

func TestArtefactMapping_TargetsExistInCatalog(t *testing.T) {
	known := loadArtefactIDs(t)
	for legacyID, novaID := range ArtefactMapping {
		if !known[novaID] {
			t.Logf("ArtefactMapping[%d]=%d: nova unit_id NOT present in configs/artefacts.yml — "+
				"импортёр пропустит такие записи с warning (см. simplifications.md)",
				legacyID, novaID)
		}
	}
}

func TestProfessionMapping_TargetsAreValid(t *testing.T) {
	known := loadProfessionKeys(t)
	for legacyID, novaKey := range ProfessionMapping {
		if !known[novaKey] {
			t.Errorf("ProfessionMapping[%d]=%q: not present in configs/professions.yml",
				legacyID, novaKey)
		}
	}
}

func TestPlanetTypeMapping_NoEmptyValues(t *testing.T) {
	for picture, ptype := range LegacyPlanetTypeMapping {
		if strings.TrimSpace(ptype) == "" {
			t.Errorf("LegacyPlanetTypeMapping[%q]: empty target", picture)
		}
	}
}

// TestMapPlanetType_StripsTrailingDigits — проверяет хелпер, который
// убирает суффикс-номер из legacy picture-строки.
func TestMapPlanetType_StripsTrailingDigits(t *testing.T) {
	cases := []struct {
		in    string
		moon  bool
		want  string
	}{
		{"dschjungelplanet05", false, "dschjungelplanet"},
		{"wasserplanet09", false, "wasserplanet"},
		{"normaltempplanet02", false, "normaltempplanet"},
		{"mond", false, "moon"},
		{"mond", true, "moon"},
		{"unknown_picture_xyz", false, "normaltempplanet"}, // fallback на planet
		{"unknown", true, "moon"},                          // fallback на moon
		{"", true, "moon"},
	}
	for _, c := range cases {
		got := mapPlanetType(c.in, c.moon)
		if got != c.want {
			t.Errorf("mapPlanetType(%q, moon=%v) = %q, want %q", c.in, c.moon, got, c.want)
		}
	}
}
