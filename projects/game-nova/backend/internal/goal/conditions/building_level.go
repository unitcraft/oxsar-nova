// Package conditions — реализации condition-типов для Goal Engine.
//
// Каждый файл регистрирует один тип в init(). Импорт пакета подключает
// все типы (см. backend/internal/goal/conditions/init.go).
package conditions

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/goal"
)

// BuildingLevelParams — параметры snapshot-условия "building_level".
//
// Цель completed когда у пользователя есть здание unit_id хотя бы на
// одной планете уровня >= MinLevel.
//
// Прогресс = max(level) у этого здания через все планеты пользователя
// (capped at MinLevel в движке). Это даёт UI прогресс-бар «3/5» для
// многоуровневых целей.
type BuildingLevelParams struct {
	UnitID   int `json:"unit_id"`
	MinLevel int `json:"min_level"`
}

func init() {
	goal.RegisterSnapshot("building_level", evalBuildingLevel)
}

func evalBuildingLevel(ctx context.Context, tx pgx.Tx, userID string, cond goal.ConditionSpec) (int, error) {
	var p BuildingLevelParams
	if err := cond.DecodeParams(&p); err != nil {
		return 0, fmt.Errorf("building_level params: %w", err)
	}
	if p.UnitID <= 0 || p.MinLevel <= 0 {
		return 0, fmt.Errorf("building_level: unit_id and min_level must be > 0")
	}

	var maxLevel int
	err := tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(b.level), 0)
		FROM buildings b
		JOIN planets p ON p.id = b.planet_id
		WHERE p.user_id = $1 AND b.unit_id = $2
	`, userID, p.UnitID).Scan(&maxLevel)
	if err != nil {
		return 0, fmt.Errorf("building_level query: %w", err)
	}
	return maxLevel, nil
}
