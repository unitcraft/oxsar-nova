package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/oxsar/nova/backend/pkg/formula"
	"github.com/oxsar/nova/backend/pkg/sqldump"
)

// Модель колонок na_construction (см. §1.4 ТЗ и
// d:/Sources/oxsar2/sql/table_dump/na_construction.sql).
// mode: 1=construction, 2=research, 3=fleet, 4=defense, 5=moon_construction.
type constructionRow struct {
	BuildingID    int64
	Race          int64
	Mode          int64
	Name          string // UPPER_SNAKE_CASE из legacy
	Front         int64
	Ballistics    int64
	Masking       int64
	BasicMetal    int64
	BasicSilicon  int64
	BasicHydrogen int64
	BasicEnergy   int64
	BasicCredit   int64
	ProdMetal     string
	ProdSilicon   string
	ProdHydrogen  string
	ProdEnergy    string
	ConsMetal     string
	ConsSilicon   string
	ConsHydrogen  string
	ConsEnergy    string
	ChargeMetal   string
	ChargeSilicon string
	ChargeHydro   string
	ChargeEnergy  string
	ChargeCredit  string
	Special       string
	Demolish      float64
	DisplayOrder  int64
}

// YAML-структура — такая же, какую будет читать config.Catalog на M1.
// Legacy-name (METALMINE) преобразуем в snake_case (metal_mine), чтобы
// совпадало с units.yml и не ломало существующие Go-импорты по ключу.
type constructionYAML struct {
	Buildings map[string]buildingYAML `yaml:"buildings"`
}

type buildingYAML struct {
	ID           int64      `yaml:"id"`
	Mode         int64      `yaml:"mode"`
	Name         string     `yaml:"legacy_name"`
	Front        int64      `yaml:"front,omitempty"`
	Ballistics   int64      `yaml:"ballistics,omitempty"`
	Masking      int64      `yaml:"masking,omitempty"`
	Basic        costYAML   `yaml:"basic,omitempty"`
	Prod         formulasY  `yaml:"prod,omitempty"`
	Cons         formulasY  `yaml:"cons,omitempty"`
	Charge       formulasY  `yaml:"charge,omitempty"`
	ChargeCredit string     `yaml:"charge_credit,omitempty"`
	Demolish     float64    `yaml:"demolish,omitempty"`
	DisplayOrder int64      `yaml:"display_order,omitempty"`
}

type costYAML struct {
	Metal    int64 `yaml:"metal,omitempty"`
	Silicon  int64 `yaml:"silicon,omitempty"`
	Hydrogen int64 `yaml:"hydrogen,omitempty"`
	Energy   int64 `yaml:"energy,omitempty"`
	Credit   int64 `yaml:"credit,omitempty"`
}

type formulasY struct {
	Metal    string `yaml:"metal,omitempty"`
	Silicon  string `yaml:"silicon,omitempty"`
	Hydrogen string `yaml:"hydrogen,omitempty"`
	Energy   string `yaml:"energy,omitempty"`
}

// convertConstruction читает na_construction.sql и пишет
// configs/construction.yml. Параллельно валидирует каждую формулу
// через pkg/formula — если хотя бы одна не парсится, падаем.
func convertConstruction(inputDir, outputDir string) error {
	src, err := readInputSQL(inputDir, "na_construction.sql")
	if err != nil {
		return err
	}
	data, err := sqldump.ParseInserts(src, "na_construction")
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	if len(data.Rows) == 0 {
		return fmt.Errorf("na_construction: no rows parsed")
	}

	col := sqldump.IndexColumns(data.Columns)
	out := constructionYAML{Buildings: map[string]buildingYAML{}}

	for _, row := range data.Rows {
		r, err := parseConstructionRow(col, row)
		if err != nil {
			return err
		}
		if err := validateFormulas(r); err != nil {
			return fmt.Errorf("building %s (id=%d): %w", r.Name, r.BuildingID, err)
		}
		key := snakeCase(r.Name)
		out.Buildings[key] = buildingYAML{
			ID:           r.BuildingID,
			Mode:         r.Mode,
			Name:         r.Name,
			Front:        r.Front,
			Ballistics:   r.Ballistics,
			Masking:      r.Masking,
			Basic: costYAML{
				Metal:    r.BasicMetal,
				Silicon:  r.BasicSilicon,
				Hydrogen: r.BasicHydrogen,
				Energy:   r.BasicEnergy,
				Credit:   r.BasicCredit,
			},
			Prod:         formulasY{r.ProdMetal, r.ProdSilicon, r.ProdHydrogen, r.ProdEnergy},
			Cons:         formulasY{r.ConsMetal, r.ConsSilicon, r.ConsHydrogen, r.ConsEnergy},
			Charge:       formulasY{r.ChargeMetal, r.ChargeSilicon, r.ChargeHydro, r.ChargeEnergy},
			ChargeCredit: r.ChargeCredit,
			Demolish:     r.Demolish,
			DisplayOrder: r.DisplayOrder,
		}
	}

	outPath := filepath.Join(outputDir, "construction.yml")
	return writeYAMLSorted(outPath, out.Buildings, func(keys []string) {
		sort.Strings(keys)
	}, "buildings")
}

// parseConstructionRow достаёт одну строку na_construction по индексу
// колонок. Если в дампе формат поменялся — здесь это сразу всплывёт.
func parseConstructionRow(col map[string]int, row []sqldump.Value) (constructionRow, error) {
	get := func(name string) (sqldump.Value, error) {
		i, ok := col[name]
		if !ok {
			return sqldump.Value{}, fmt.Errorf("column %q missing", name)
		}
		if i >= len(row) {
			return sqldump.Value{}, fmt.Errorf("row too short for column %q", name)
		}
		return row[i], nil
	}
	getStr := func(name string) (string, error) {
		v, err := get(name)
		if err != nil {
			return "", err
		}
		return v.Raw, nil
	}
	getInt := func(name string) (int64, error) {
		v, err := get(name)
		if err != nil {
			return 0, err
		}
		return v.AsInt()
	}
	getFloat := func(name string) (float64, error) {
		v, err := get(name)
		if err != nil {
			return 0, err
		}
		return v.AsFloat()
	}

	r := constructionRow{}
	fields := []struct {
		col string
		ptr any
	}{
		{"buildingid", &r.BuildingID},
		{"race", &r.Race},
		{"mode", &r.Mode},
		{"name", &r.Name},
		{"front", &r.Front},
		{"ballistics", &r.Ballistics},
		{"masking", &r.Masking},
		{"basic_metal", &r.BasicMetal},
		{"basic_silicon", &r.BasicSilicon},
		{"basic_hydrogen", &r.BasicHydrogen},
		{"basic_energy", &r.BasicEnergy},
		{"basic_credit", &r.BasicCredit},
		{"prod_metal", &r.ProdMetal},
		{"prod_silicon", &r.ProdSilicon},
		{"prod_hydrogen", &r.ProdHydrogen},
		{"prod_energy", &r.ProdEnergy},
		{"cons_metal", &r.ConsMetal},
		{"cons_silicon", &r.ConsSilicon},
		{"cons_hydrogen", &r.ConsHydrogen},
		{"cons_energy", &r.ConsEnergy},
		{"charge_metal", &r.ChargeMetal},
		{"charge_silicon", &r.ChargeSilicon},
		{"charge_hydrogen", &r.ChargeHydro},
		{"charge_energy", &r.ChargeEnergy},
		{"charge_credit", &r.ChargeCredit},
		{"special", &r.Special},
		{"demolish", &r.Demolish},
		{"display_order", &r.DisplayOrder},
	}
	for _, f := range fields {
		var err error
		switch p := f.ptr.(type) {
		case *int64:
			*p, err = getInt(f.col)
		case *float64:
			*p, err = getFloat(f.col)
		case *string:
			*p, err = getStr(f.col)
		}
		if err != nil {
			return r, fmt.Errorf("field %s: %w", f.col, err)
		}
	}
	return r, nil
}

// validateFormulas прогоняет каждую формулу через formula.Parse — это
// золотой тест «SQL-дамп не испортился, DSL его понимает».
func validateFormulas(r constructionRow) error {
	all := []struct {
		name, src string
	}{
		{"prod_metal", r.ProdMetal},
		{"prod_silicon", r.ProdSilicon},
		{"prod_hydrogen", r.ProdHydrogen},
		{"prod_energy", r.ProdEnergy},
		{"cons_metal", r.ConsMetal},
		{"cons_silicon", r.ConsSilicon},
		{"cons_hydrogen", r.ConsHydrogen},
		{"cons_energy", r.ConsEnergy},
		{"charge_metal", r.ChargeMetal},
		{"charge_silicon", r.ChargeSilicon},
		{"charge_hydrogen", r.ChargeHydro},
		{"charge_energy", r.ChargeEnergy},
		{"charge_credit", r.ChargeCredit},
	}
	for _, f := range all {
		if _, err := formula.Parse(f.src); err != nil {
			return fmt.Errorf("formula %s=%q: %w", f.name, f.src, err)
		}
	}
	return nil
}

// snakeCase превращает METALMINE -> metal_mine (как в configs/units.yml).
// Правила: знаки `_` сохраняются; буквы приводятся к нижнему регистру.
// Legacy-имена уже в SNAKE_UPPER, так что достаточно lowercase.
func snakeCase(legacy string) string {
	return strings.ToLower(legacy)
}

// writeYAMLSorted сериализует карту с сортированными ключами.
// yaml.v3 по умолчанию не гарантирует порядок, а для git-diff'а нам
// нужна стабильность.
func writeYAMLSorted[V any](path string, m map[string]V, sortFn func([]string), topKey string) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	if sortFn != nil {
		sortFn(keys)
	}
	node := &yaml.Node{Kind: yaml.MappingNode}
	innerNode := &yaml.Node{Kind: yaml.MappingNode}
	for _, k := range keys {
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: k}
		valNode := &yaml.Node{}
		if err := valNode.Encode(m[k]); err != nil {
			return err
		}
		innerNode.Content = append(innerNode.Content, keyNode, valNode)
	}
	node.Content = []*yaml.Node{
		{Kind: yaml.ScalarNode, Value: topKey},
		innerNode,
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(node); err != nil {
		return err
	}
	return enc.Close()
}
