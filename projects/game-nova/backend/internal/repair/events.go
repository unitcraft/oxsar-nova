package repair

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/event"
)

// DisassembleHandler — обработчик KindDisassemble=51. Идемпотентен:
// если очередь уже 'done', ничего не делает (страховка от повторного
// запуска после сбоя воркера).
func (s *Service) DisassembleHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl struct {
			QueueID string `json:"queue_id"`
			Mode    string `json:"mode"`
		}
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("disassemble: parse payload: %w", err)
		}
		var (
			status                                     string
			planetID, userID                           string
			retMetal, retSilicon, retHydrogen, qCount  int64
		)
		err := tx.QueryRow(ctx, `
			SELECT status, planet_id, user_id,
			       return_metal, return_silicon, return_hydrogen, count
			FROM repair_queue WHERE id = $1 FOR UPDATE
		`, pl.QueueID).Scan(&status, &planetID, &userID,
			&retMetal, &retSilicon, &retHydrogen, &qCount)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil // очередь удалена — считаем событие погашенным
			}
			return fmt.Errorf("disassemble: select queue: %w", err)
		}
		if status == "done" {
			return nil
		}

		if _, err := tx.Exec(ctx, `
			UPDATE planets SET metal = metal + $1, silicon = silicon + $2, hydrogen = hydrogen + $3
			WHERE id = $4
		`, retMetal, retSilicon, retHydrogen, planetID); err != nil {
			return fmt.Errorf("disassemble: credit return: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO res_log (user_id, planet_id, reason, delta_metal, delta_silicon, delta_hydrogen)
			VALUES ($1, $2, 'disassemble_return', $3, $4, $5)
		`, userID, planetID, retMetal, retSilicon, retHydrogen); err != nil {
			return fmt.Errorf("disassemble: res_log: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE repair_queue SET status='done' WHERE id=$1`, pl.QueueID,
		); err != nil {
			return fmt.Errorf("disassemble: mark done: %w", err)
		}
		return nil
	}
}

// RepairHandler — обработчик KindRepair=50. При срабатывании уменьшает
// damaged_count на qCount (то, что чинили) у ships или defense для
// данного unit_id. Если итоговый damaged ≤ 0 — также сбрасывает
// shell_percent (стек полностью восстановлен).
//
// План 72.1.56 B6: legacy 1:1 partial-repair поддерживается; раньше
// handler сбрасывал damaged=0 безусловно (некорректно при quantity <
// damaged + не различал ships/defense).
func (s *Service) RepairHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl struct {
			QueueID string `json:"queue_id"`
			Mode    string `json:"mode"`
		}
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("repair: parse payload: %w", err)
		}
		var (
			status    string
			planetID  string
			unitID    int
			qCount    int64
			isDefense bool
		)
		err := tx.QueryRow(ctx, `
			SELECT status, planet_id, unit_id, count, is_defense
			FROM repair_queue WHERE id = $1 FOR UPDATE
		`, pl.QueueID).Scan(&status, &planetID, &unitID, &qCount, &isDefense)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil
			}
			return fmt.Errorf("repair: select queue: %w", err)
		}
		if status == "done" {
			return nil
		}

		stockTable := "ships"
		if isDefense {
			stockTable = "defense"
		}
		// Уменьшаем damaged_count на чиненое qCount (clamp ≥ 0).
		// Если итоговый damaged_count = 0 → сбрасываем shell_percent
		// (стек полностью восстановлен).
		if _, err := tx.Exec(ctx, `
			UPDATE `+stockTable+`
			SET damaged_count = GREATEST(damaged_count - $3, 0),
			    shell_percent = CASE
			        WHEN damaged_count - $3 <= 0 THEN 0
			        ELSE shell_percent
			    END
			WHERE planet_id = $1 AND unit_id = $2
		`, planetID, unitID, qCount); err != nil {
			return fmt.Errorf("repair: apply damaged: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE repair_queue SET status='done' WHERE id=$1`, pl.QueueID,
		); err != nil {
			return fmt.Errorf("repair: mark done: %w", err)
		}
		return nil
	}
}
