package artefact

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/pkg/ids"
)

// DelayEvent формирует event.Handler для KindArtefactDelay (63).
// Срабатывает, когда истёк delay_seconds артефакта: переводит
// state=delayed → active и применяет эффекты.
func (s *Service) DelayEvent() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl struct {
			ArtefactID string `json:"artefact_id"`
		}
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("parse artefact delay payload: %w", err)
		}

		var rec Record
		err := tx.QueryRow(ctx, `
			SELECT id, user_id, planet_id, unit_id, state, acquired_at, activated_at, expire_at
			FROM artefacts_user WHERE id = $1 FOR UPDATE
		`, pl.ArtefactID).Scan(&rec.ID, &rec.UserID, &rec.PlanetID, &rec.UnitID,
			&rec.State, &rec.AcquiredAt, &rec.ActivatedAt, &rec.ExpireAt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("load artefact: %w", err)
		}
		// Если игрок уже активировал/деактивировал — пропускаем.
		if rec.State != StateDelayed {
			return nil
		}

		spec, ok := s.lookupByID(rec.UnitID)
		if !ok {
			return fmt.Errorf("unknown artefact id %d", rec.UnitID)
		}

		change, err := computeChanges(spec, dirApply)
		if err != nil && !errors.Is(err, ErrUnsupported) {
			return err
		}
		if change != nil {
			if change.Scope == "planet" && rec.PlanetID == nil {
				return ErrPlanetRequired
			}
			if err := applyChange(ctx, tx, *change, rec.UserID, rec.PlanetID); err != nil {
				return err
			}
		}

		now := time.Now().UTC()
		var expire *time.Time
		if spec.LifetimeSeconds > 0 {
			t := now.Add(time.Duration(spec.LifetimeSeconds) * time.Second)
			expire = &t
		}
		if _, err := tx.Exec(ctx, `
			UPDATE artefacts_user SET state = $1, expire_at = $2 WHERE id = $3
		`, StateActive, expire, rec.ID); err != nil {
			return fmt.Errorf("activate after delay: %w", err)
		}

		if expire != nil {
			if _, err := tx.Exec(ctx, `
				INSERT INTO events (id, user_id, planet_id, kind, state, fire_at, payload)
				VALUES ($1, $2, $3, 60, 'wait', $4, $5)
			`, ids.New(), rec.UserID, rec.PlanetID, *expire,
				fmt.Sprintf(`{"artefact_id":"%s"}`, rec.ID)); err != nil {
				return fmt.Errorf("insert expire event: %w", err)
			}
		}
		return nil
	}
}
