package main

import (
	"fmt"
	"path/filepath"
	"sort"

	"oxsar/game-nova/pkg/sqldump"
)

// na_artefact_datasheet: (typeid, buyable, auto_active, movable, unique,
// usable, trophy_chance, delay, use_times, use_duration, lifetime,
// effect_type, max_active, quota).
//
// ВАЖНО: Эти поля описывают МЕТАДАННЫЕ артефакта (можно ли покупать,
// когда активируется, cколько живёт), но НЕ описывают конкретный
// эффект. Эффект хранится в PHP-switch в game/Artefact.class.php
// (см. §5.10.1 ТЗ).
//
// Поэтому мы генерим artefacts_meta_generated.yml — метаданные, которые
// слипаются с нашим ручным configs/artefacts.yml (эффекты). Слияние
// происходит на этапе загрузки каталога.

type artefactMetaRow struct {
	TypeID       int64
	Buyable      int64
	AutoActive   int64
	Movable      int64
	Unique       int64
	Usable       int64
	TrophyChance int64
	Delay        int64
	UseTimes     int64
	UseDuration  int64
	Lifetime     int64
	EffectType   int64
	MaxActive    int64
	Quota        float64
}

type artefactMetaYAML struct {
	ID           int64   `yaml:"id"`
	Buyable      bool    `yaml:"buyable"`
	AutoActive   bool    `yaml:"auto_active,omitempty"`
	Movable      bool    `yaml:"movable"`
	Unique       bool    `yaml:"unique,omitempty"`
	Usable       bool    `yaml:"usable,omitempty"`
	TrophyChance int64   `yaml:"trophy_chance"`
	Delay        int64   `yaml:"delay,omitempty"`
	UseTimes     int64   `yaml:"use_times,omitempty"`
	UseDuration  int64   `yaml:"use_duration,omitempty"`
	Lifetime     int64   `yaml:"lifetime,omitempty"`
	EffectType   int64   `yaml:"effect_type"`
	MaxActive    int64   `yaml:"max_active,omitempty"`
	Quota        float64 `yaml:"quota,omitempty"`
}

func convertArtefacts(inputDir, outputDir string) error {
	src, err := readInputSQL(inputDir, "na_artefact_datasheet.sql")
	if err != nil {
		return err
	}
	data, err := sqldump.ParseInserts(src, "na_artefact_datasheet")
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	if len(data.Rows) == 0 {
		return nil
	}
	col := sqldump.IndexColumns(data.Columns)

	out := map[string]artefactMetaYAML{}
	for _, row := range data.Rows {
		r := artefactMetaRow{}
		intMap := []struct {
			col string
			ptr *int64
		}{
			{"typeid", &r.TypeID},
			{"buyable", &r.Buyable},
			{"auto_active", &r.AutoActive},
			{"movable", &r.Movable},
			{"unique", &r.Unique},
			{"usable", &r.Usable},
			{"trophy_chance", &r.TrophyChance},
			{"delay", &r.Delay},
			{"use_times", &r.UseTimes},
			{"use_duration", &r.UseDuration},
			{"lifetime", &r.Lifetime},
			{"effect_type", &r.EffectType},
			{"max_active", &r.MaxActive},
		}
		for _, f := range intMap {
			if err := sqldump.AssignInt(col, row, f.col, f.ptr); err != nil {
				return err
			}
		}
		// quota — float
		if i, ok := col["quota"]; ok && i < len(row) {
			if v, err := row[i].AsFloat(); err == nil {
				r.Quota = v
			}
		}

		key := fmt.Sprintf("artefact_%d", r.TypeID)
		out[key] = artefactMetaYAML{
			ID:           r.TypeID,
			Buyable:      r.Buyable == 1,
			AutoActive:   r.AutoActive == 1,
			Movable:      r.Movable == 1,
			Unique:       r.Unique == 1,
			Usable:       r.Usable == 1,
			TrophyChance: r.TrophyChance,
			Delay:        r.Delay,
			UseTimes:     r.UseTimes,
			UseDuration:  r.UseDuration,
			Lifetime:     r.Lifetime,
			EffectType:   r.EffectType,
			MaxActive:    r.MaxActive,
			Quota:        r.Quota,
		}
	}

	outPath := filepath.Join(outputDir, "artefacts_meta_generated.yml")
	return writeYAMLSorted(outPath, out, sort.Strings, "artefacts")
}
