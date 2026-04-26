package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

// ExpirePlanetPayload — payload KindExpirePlanet.
type ExpirePlanetPayload struct {
	PlanetID string `json:"planet_id"`
}

// HandleExpirePlanet — handler для KindExpirePlanet.
// Идемпотентно удаляет (hard-delete) планету, если её expires_at
// всё ещё в прошлом. Если expires_at был сдвинут (игрок купил
// постоянство) — ничего не делает.
func HandleExpirePlanet(ctx context.Context, tx pgx.Tx, e Event) error {
	var pl ExpirePlanetPayload
	if err := json.Unmarshal(e.Payload, &pl); err != nil {
		return fmt.Errorf("expire_planet: parse: %w", err)
	}
	tag, err := tx.Exec(ctx, `
		DELETE FROM planets
		WHERE id = $1
		  AND expires_at IS NOT NULL
		  AND expires_at <= now()
	`, pl.PlanetID)
	if err != nil {
		return fmt.Errorf("expire_planet: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		slog.DebugContext(ctx, "expire_planet_skipped",
			slog.String("planet_id", pl.PlanetID))
	}
	return nil
}
