// Package battlestats: WriteBattleReport — общий helper для INSERT
// в `battle_reports` для всех типов боёв (PvP attack, ACS attack,
// alien-raid, expedition).
//
// План 72.1.55 Task H (P72.1.1.NO_BATTLE_REPORT_FOR_RAIDS 1:1):
// раньше alien-рейды и экспедиции вызывали ApplyBattleResult с
// пустым battleID, что ломало UNIQUE-constraint на user_experience
// при event re-process (NULL != NULL в SQL → дубликат опыта).
// Теперь все типы боёв пишут запись в battle_reports и передают её
// id в ApplyBattleResult — UNIQUE (battle_id, user_id, is_atter)
// блокирует дубль.
//
// Helper не дублирует inline INSERT в attack/acs_attack — те уже
// работают и не трогаются (R0). Helper используется в новых call
// sites (alien.go + expedition.go).
package battlestats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/battle"
	"oxsar/game-nova/pkg/ids"
)

// ReportFlags — флаги для фильтров /battlestats (миграция 0085).
type ReportFlags struct {
	HasAliens    bool
	MoonCreated  bool
	IsMoon       bool
}

// WriteBattleReport вставляет запись в battle_reports и возвращает
// её id. attackerUserID/defenderUserID могут быть пустыми (для NPC
// сторон или анонимных боёв). planetID может быть пустым (например,
// для экспедиции). loot/debris передаются как разложенные int64 —
// fleet/lootAmount package-private, поэтому helper не зависит от него.
func WriteBattleReport(
	ctx context.Context,
	tx pgx.Tx,
	rep battle.Report,
	attackerUserID, defenderUserID, planetID string,
	lootMetal, lootSilicon, lootHydrogen, debrisMetal, debrisSilicon int64,
	flags ReportFlags,
) (string, error) {
	reportJSON, err := json.Marshal(rep)
	if err != nil {
		return "", fmt.Errorf("write battle report: marshal: %w", err)
	}
	reportID := ids.New()
	atkArg := nullableUUID(attackerUserID)
	defArg := nullableUUID(defenderUserID)
	plArg := nullableUUID(planetID)
	if _, err := tx.Exec(ctx, `
		INSERT INTO battle_reports (id, attacker_user_id, defender_user_id, planet_id,
		                            seed, winner, rounds,
		                            debris_metal, debris_silicon,
		                            loot_metal, loot_silicon, loot_hydrogen,
		                            report,
		                            has_aliens, moon_created, is_moon)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, reportID, atkArg, defArg, plArg,
		int64(rep.Seed), rep.Winner, rep.Rounds,
		debrisMetal, debrisSilicon,
		lootMetal, lootSilicon, lootHydrogen,
		reportJSON,
		flags.HasAliens, flags.MoonCreated, flags.IsMoon,
	); err != nil {
		return "", fmt.Errorf("write battle report: insert: %w", err)
	}
	return reportID, nil
}

// nullableUUID возвращает nil для пустой строки (чтобы PostgreSQL
// записал NULL вместо пустой строки в uuid-колонку).
func nullableUUID(s string) any {
	if s == "" {
		return nil
	}
	return s
}
