package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// BuildingPayload — payload события завершения стройки здания или
// исследования. Структура одинаковая, различается только Kind события
// и таблица, куда применяется уровень.
type BuildingPayload struct {
	QueueID     string `json:"queue_id"`
	UnitID      int    `json:"unit_id"`
	TargetLevel int    `json:"target_level"`
}

// ShipyardPayload — payload события окончания постройки кораблей/обороны.
// Здесь важнее count, а не target_level.
type ShipyardPayload struct {
	QueueID string `json:"queue_id"`
	UnitID  int    `json:"unit_id"`
	Count   int64  `json:"count"`
	IsDefense bool `json:"is_defense"`
}

// HandleBuildConstruction повышает уровень здания на планете,
// закрывает запись в construction_queue. Идемпотентен: если уровень уже
// >= target, ничего не делает.
func HandleBuildConstruction(ctx context.Context, tx pgx.Tx, e Event) error {
	var pl BuildingPayload
	if err := json.Unmarshal(e.Payload, &pl); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}
	if e.PlanetID == nil {
		return fmt.Errorf("building event without planet_id")
	}

	var cur int
	err := tx.QueryRow(ctx,
		`SELECT level FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
		*e.PlanetID, pl.UnitID,
	).Scan(&cur)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("select level: %w", err)
	}
	if cur >= pl.TargetLevel {
		// уже применено ранее — идемпотентность
		_, _ = tx.Exec(ctx, `UPDATE construction_queue SET status='done' WHERE id=$1`, pl.QueueID)
		return nil
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO buildings (planet_id, unit_id, level)
		VALUES ($1, $2, $3)
		ON CONFLICT (planet_id, unit_id) DO UPDATE SET level = EXCLUDED.level
	`, *e.PlanetID, pl.UnitID, pl.TargetLevel); err != nil {
		return fmt.Errorf("upsert building: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE construction_queue SET status='done' WHERE id=$1`, pl.QueueID); err != nil {
		return fmt.Errorf("close queue: %w", err)
	}
	return nil
}

// HandleResearch повышает уровень research у игрока. Идемпотентен.
func HandleResearch(ctx context.Context, tx pgx.Tx, e Event) error {
	var pl BuildingPayload
	if err := json.Unmarshal(e.Payload, &pl); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}
	if e.UserID == nil {
		return fmt.Errorf("research event without user_id")
	}
	var cur int
	err := tx.QueryRow(ctx,
		`SELECT level FROM research WHERE user_id=$1 AND unit_id=$2`,
		*e.UserID, pl.UnitID,
	).Scan(&cur)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("select research: %w", err)
	}
	if cur >= pl.TargetLevel {
		_, _ = tx.Exec(ctx, `UPDATE construction_queue SET status='done' WHERE id=$1`, pl.QueueID)
		return nil
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO research (user_id, unit_id, level)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, unit_id) DO UPDATE SET level = EXCLUDED.level
	`, *e.UserID, pl.UnitID, pl.TargetLevel); err != nil {
		return fmt.Errorf("upsert research: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE construction_queue SET status='done' WHERE id=$1`, pl.QueueID); err != nil {
		return fmt.Errorf("close queue: %w", err)
	}
	return nil
}

// HandleBuildFleet применяет постройку корабля.
//
// В отличие от постройки здания/исследования, тут порция кораблей
// (Count) добавляется к существующему запасу. Идемпотентность
// обеспечивается через проверку статуса очереди: если status=done,
// событие уже было обработано ранее, ничего не делаем.
func HandleBuildFleet(ctx context.Context, tx pgx.Tx, e Event) error {
	var pl ShipyardPayload
	if err := json.Unmarshal(e.Payload, &pl); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}
	if e.PlanetID == nil {
		return fmt.Errorf("fleet event without planet_id")
	}

	var status string
	err := tx.QueryRow(ctx,
		`SELECT status FROM shipyard_queue WHERE id=$1`, pl.QueueID,
	).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil // уже удалено — идемпотентность
		}
		return fmt.Errorf("select queue: %w", err)
	}
	if status == "done" {
		return nil
	}

	targetTable := "ships"
	if pl.IsDefense {
		targetTable = "defense"
	}

	if _, err := tx.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (planet_id, unit_id, count)
		VALUES ($1, $2, $3)
		ON CONFLICT (planet_id, unit_id) DO UPDATE SET count = %s.count + EXCLUDED.count
	`, targetTable, targetTable),
		*e.PlanetID, pl.UnitID, pl.Count,
	); err != nil {
		return fmt.Errorf("upsert %s: %w", targetTable, err)
	}
	if _, err := tx.Exec(ctx, `UPDATE shipyard_queue SET status='done' WHERE id=$1`, pl.QueueID); err != nil {
		return fmt.Errorf("close shipyard queue: %w", err)
	}
	return nil
}
