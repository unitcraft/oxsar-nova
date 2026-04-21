package main

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/oxsar/nova/backend/pkg/sqldump"
)

// na_ship_datasheet: unitid, capicity, speed, consume, attack, shield.
// Там написано "capicity" (опечатка в legacy-схеме), сохраняем как есть
// при чтении, но наружу в YAML пишем "capacity" — это не баланс, это
// имя поля, и исправление не ломает поиск по гайдам (§18.22 ТЗ).

type shipRow struct {
	UnitID  int64
	Cargo   int64
	Speed   int64
	Consume int64
	Attack  int64
	Shield  int64
}

type shipYAML struct {
	ID      int64 `yaml:"id"`
	Cargo   int64 `yaml:"cargo"`
	Speed   int64 `yaml:"speed"`
	Consume int64 `yaml:"consume"`
	Attack  int64 `yaml:"attack"`
	Shield  int64 `yaml:"shield"`
}

// convertShips читает na_ship_datasheet.sql и дописывает
// configs/ships_generated.yml. НЕ перезаписываем configs/ships.yml —
// там есть вручную подобранный баланс из первых версий (§§5.5, 18.11).
// Имя _generated — маркер: этот файл формируется тулом, правки руками
// потеряются при следующем прогоне.
//
// Mapping unit_id -> ключ берём из configs/units.yml через side-import:
// в текущей реализации ключ собираем как "unit_<id>". Это некрасиво,
// но разблокирует M0.1 без дополнительных зависимостей. Когда будет
// единый registry (§18.11), ключи подтянутся оттуда.
func convertShips(inputDir, outputDir string) error {
	src, err := readInputSQL(inputDir, "na_ship_datasheet.sql")
	if err != nil {
		return err
	}
	data, err := sqldump.ParseInserts(src, "na_ship_datasheet")
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	if len(data.Rows) == 0 {
		return nil
	}
	col := sqldump.IndexColumns(data.Columns)

	out := map[string]shipYAML{}
	for _, row := range data.Rows {
		r := shipRow{}
		if err := sqldump.AssignInt(col, row, "unitid", &r.UnitID); err != nil {
			return err
		}
		if err := sqldump.AssignInt(col, row, "capicity", &r.Cargo); err != nil {
			return err
		}
		if err := sqldump.AssignInt(col, row, "speed", &r.Speed); err != nil {
			return err
		}
		if err := sqldump.AssignInt(col, row, "consume", &r.Consume); err != nil {
			return err
		}
		if err := sqldump.AssignInt(col, row, "attack", &r.Attack); err != nil {
			return err
		}
		if err := sqldump.AssignInt(col, row, "shield", &r.Shield); err != nil {
			return err
		}
		key := fmt.Sprintf("unit_%d", r.UnitID)
		out[key] = shipYAML{
			ID: r.UnitID, Cargo: r.Cargo, Speed: r.Speed,
			Consume: r.Consume, Attack: r.Attack, Shield: r.Shield,
		}
	}

	outPath := filepath.Join(outputDir, "ships_generated.yml")
	return writeYAMLSorted(outPath, out, sort.Strings, "ships")
}

