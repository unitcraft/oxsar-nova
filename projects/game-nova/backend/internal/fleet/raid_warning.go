// RAID_WARNING (kind=64) — уведомление защитнику за 10 минут до прибытия атакующего флота.
package fleet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/pkg/ids"
)

type raidWarningPayload struct {
	FleetID string `json:"fleet_id"`
}

// RaidWarningHandler шлёт сообщение владельцу целевой планеты о приближающемся флоте.
func (s *TransportService) RaidWarningHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl raidWarningPayload
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("raid warning: parse payload: %w", err)
		}

		// Читаем флот — нужны координаты и миссия.
		var ownerUserID string
		var g, sys, pos int
		var isMoon bool
		var state string
		err := tx.QueryRow(ctx, `
			SELECT owner_user_id, dst_galaxy, dst_system, dst_position, dst_is_moon, state
			FROM fleets WHERE id=$1
		`, pl.FleetID).Scan(&ownerUserID, &g, &sys, &pos, &isMoon, &state)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil // флот удалён или отозван
			}
			return fmt.Errorf("raid warning: read fleet: %w", err)
		}
		if state != "outbound" {
			return nil // флот уже отозван
		}

		// Определяем владельца целевой планеты.
		var defenderUserID string
		err = tx.QueryRow(ctx, `
			SELECT user_id FROM planets
			WHERE galaxy=$1 AND system=$2 AND position=$3 AND is_moon=$4 AND destroyed_at IS NULL
		`, g, sys, pos, isMoon).Scan(&defenderUserID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil // планета исчезла или ничья
			}
			return fmt.Errorf("raid warning: read planet: %w", err)
		}
		if defenderUserID == ownerUserID {
			return nil // атакующий атакует свою же планету
		}

		subject := s.tr("mission", "raidWarnSubject", nil)
		body := s.tr("mission", "raidWarnBody", map[string]string{
			"g": strconv.Itoa(g), "s": strconv.Itoa(sys), "pos": strconv.Itoa(pos),
		})
		if _, err := tx.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
			VALUES ($1, $2, NULL, 13, $3, $4)
		`, ids.New(), defenderUserID, subject, body); err != nil {
			return fmt.Errorf("raid warning: insert message: %w", err)
		}
		return nil
	}
}
