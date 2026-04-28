package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

// BuildingPayload — payload события завершения стройки здания или
// исследования. Структура одинаковая, различается только Kind события
// и таблица, куда применяется уровень.
//
// Для KindDemolishConstruction TargetLevel — желаемый уровень ПОСЛЕ
// сноса (обычно curLevel-1, может быть 0). См. HandleDemolishConstruction.
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

	// План 23: инкрементируем used_fields только при первой постройке
	// здания (cur==0, target==1). Апгрейд того же здания поля не занимает.
	// Solar satellite и ракеты не считаются «зданиями» на полях в legacy
	// (см. Planet.class.php:717 — getFields), но поскольку их нельзя
	// построить через construction_queue как buildings (они через
	// shipyard), здесь не фильтруем.
	if cur == 0 && pl.TargetLevel == 1 {
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET used_fields = used_fields + 1 WHERE id = $1`,
			*e.PlanetID); err != nil {
			return fmt.Errorf("inc used_fields: %w", err)
		}
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

// HandleDemolishConstruction понижает уровень здания на планете до
// TargetLevel (обычно curLevel-1, допускается 0 = полное удаление).
// Зеркалит HandleBuildConstruction.
//
// Семантика origin (EventHandler::demolish, EventHandler.class.php:2257):
//   - level > 0 → UPDATE building level = TargetLevel.
//   - level == 0 → DELETE строки building (легаси).
//
// В nova таблица buildings без UNIQUE-NOT-NULL на level, поэтому DELETE
// эквивалентен UPDATE level=0 (см. SELECT ниже — отсутствие строки
// читается как 0). Используем UPDATE: данные о факте «когда-то было
// построено» можно восстановить через events_dead. Альтернатива (DELETE)
// прокатилась бы, но усложняет audit-замеры «сколько построек игрок
// демонтировал за период».
//
// Идемпотентность: если cur <= TargetLevel — событие уже применено
// (или применено раньше воркером, или заявка перезатёрта новой). В этом
// случае только закрываем очередь и возвращаемся без ошибки.
//
// Поля поля планеты: при demolish здания **до 0** возвращаем 1 used_field
// (зеркалит HandleBuildConstruction:69). При понижении уровня (>0) — нет.
//
// Очки: пересчитываются батчем (KindScoreRecalcAll, Kind=70) — здесь не
// трогаем. Это отличается от legacy oxsar2 (инкремент UPDATE user.points
// в той же транзакции), но в nova очки derived state, восстанавливаемые
// из buildings/research/ships.
//
// Audit: пишем структурированный slog (R3) с полями event_id, planet_id,
// unit_id, level_from, level_to. Отдельной audit-таблицы для player-action
// в nova нет — slog уезжает в централизованный лог-агрегатор и достаточен
// для построения временного ряда «снос построек». Если понадобится
// SQL-доступ к истории — события остаются в events / events_dead с
// исходным payload.
func HandleDemolishConstruction(ctx context.Context, tx pgx.Tx, e Event) error {
	var pl BuildingPayload
	if err := json.Unmarshal(e.Payload, &pl); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}
	if e.PlanetID == nil {
		return fmt.Errorf("demolish event without planet_id")
	}
	if pl.TargetLevel < 0 {
		return fmt.Errorf("demolish target_level must be >=0, got %d", pl.TargetLevel)
	}

	var cur int
	err := tx.QueryRow(ctx,
		`SELECT level FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
		*e.PlanetID, pl.UnitID,
	).Scan(&cur)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("select level: %w", err)
	}
	// Идемпотентность: уже применён demolish (или ниже).
	if cur <= pl.TargetLevel {
		_, _ = tx.Exec(ctx, `UPDATE construction_queue SET status='done' WHERE id=$1`, pl.QueueID)
		slog.InfoContext(ctx, "event_demolish_skip_idempotent",
			slog.String("event_id", e.ID),
			slog.String("planet_id", *e.PlanetID),
			slog.Int("unit_id", pl.UnitID),
			slog.Int("level_current", cur),
			slog.Int("level_target", pl.TargetLevel))
		return nil
	}

	if _, err := tx.Exec(ctx, `
		UPDATE buildings SET level=$3
		WHERE planet_id=$1 AND unit_id=$2
	`, *e.PlanetID, pl.UnitID, pl.TargetLevel); err != nil {
		return fmt.Errorf("downgrade building: %w", err)
	}

	// План 23 (зеркало HandleBuildConstruction): полностью снесённое
	// здание (target=0) освобождает поле планеты. Понижение уровня (target>0)
	// поле не освобождает — само здание остаётся.
	if pl.TargetLevel == 0 {
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET used_fields = GREATEST(used_fields - 1, 0) WHERE id = $1`,
			*e.PlanetID); err != nil {
			return fmt.Errorf("dec used_fields: %w", err)
		}
	}

	if _, err := tx.Exec(ctx, `UPDATE construction_queue SET status='done' WHERE id=$1`, pl.QueueID); err != nil {
		return fmt.Errorf("close queue: %w", err)
	}

	slog.InfoContext(ctx, "event_demolish_applied",
		slog.String("event_id", e.ID),
		slog.String("planet_id", *e.PlanetID),
		slog.Int("unit_id", pl.UnitID),
		slog.Int("level_from", cur),
		slog.Int("level_to", pl.TargetLevel))
	return nil
}
