package main

import (
	"fmt"
	"path/filepath"
	"sort"

	"oxsar/game-nova/pkg/sqldump"
)

// na_requirements: (requirementid, buildingid, needs, level, level_limit).
// Семантика: для юнита buildingid нужна (needs) на уровне level.
// level_limit (может быть NULL) — предельный уровень юнита, до которого
// требование применяется. Нас интересуют требования базовые (без limit).
//
// Поле needs — это ID из таблицы na_construction (здание ИЛИ исследование,
// зависит от mode ссылаемого unit'а).

type reqRow struct {
	BuildingID int64
	Needs      int64
	Level      int64
	LevelLimit int64 // 0 = NULL
}

type requirementsYAML struct {
	// key = "unit_<id>", value = список требований
	Requirements map[string][]reqItemYAML `yaml:"requirements"`
}

type reqItemYAML struct {
	// Ссылаемся на ID юнита из na_construction. Конкретный `kind`
	// (building | research) мы здесь не определяем — это знание из
	// того же na_construction.mode и обрабатывается на этапе загрузки
	// каталога в приложении.
	Needs      int64 `yaml:"needs"`
	Level      int64 `yaml:"level"`
	LevelLimit int64 `yaml:"level_limit,omitempty"`
}

func convertRequirements(inputDir, outputDir string) error {
	src, err := readInputSQL(inputDir, "na_requirements.sql")
	if err != nil {
		return err
	}
	data, err := sqldump.ParseInserts(src, "na_requirements")
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	if len(data.Rows) == 0 {
		return nil
	}
	col := sqldump.IndexColumns(data.Columns)

	buckets := map[string][]reqItemYAML{}
	for _, row := range data.Rows {
		r := reqRow{}
		if err := sqldump.AssignInt(col, row, "buildingid", &r.BuildingID); err != nil {
			return err
		}
		if err := sqldump.AssignInt(col, row, "needs", &r.Needs); err != nil {
			return err
		}
		if err := sqldump.AssignInt(col, row, "level", &r.Level); err != nil {
			return err
		}
		// level_limit — NULL-able.
		if i, ok := col["level_limit"]; ok && i < len(row) && !row[i].IsNull {
			v, err := row[i].AsInt()
			if err != nil {
				return fmt.Errorf("level_limit: %w", err)
			}
			r.LevelLimit = v
		}
		key := fmt.Sprintf("unit_%d", r.BuildingID)
		buckets[key] = append(buckets[key], reqItemYAML{
			Needs: r.Needs, Level: r.Level, LevelLimit: r.LevelLimit,
		})
	}

	// Стабилизируем порядок требований внутри каждого юнита.
	for k := range buckets {
		sort.SliceStable(buckets[k], func(i, j int) bool {
			a, b := buckets[k][i], buckets[k][j]
			if a.Needs != b.Needs {
				return a.Needs < b.Needs
			}
			return a.Level < b.Level
		})
	}

	outPath := filepath.Join(outputDir, "requirements_generated.yml")
	return writeYAMLSorted(outPath, buckets, sort.Strings, "requirements")
}
