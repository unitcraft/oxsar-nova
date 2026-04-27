package artefact

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/event"
)

// ExpireEvent формирует event.Handler, регистрируемый в воркере для
// KindArtefactExpire. Почему так, а не регистрация изнутри event:
// event не должен импортировать artefact (цикл), а artefact
// формально зависит от event только типами — это допустимо.
func (s *Service) ExpireEvent() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl struct {
			ArtefactID string `json:"artefact_id"`
		}
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("parse artefact expire payload: %w", err)
		}

		// Читаем и проверяем, что артефакт всё ещё активен: если его
		// уже деактивировали вручную — событие ничего не делает.
		var rec Record
		err := tx.QueryRow(ctx, `
			SELECT id, user_id, planet_id, unit_id, state, acquired_at, activated_at, expire_at
			FROM artefacts_user WHERE id = $1 FOR UPDATE
		`, pl.ArtefactID).Scan(&rec.ID, &rec.UserID, &rec.PlanetID, &rec.UnitID,
			&rec.State, &rec.AcquiredAt, &rec.ActivatedAt, &rec.ExpireAt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil // артефакт удалён — идемпотентно
			}
			return fmt.Errorf("load artefact: %w", err)
		}
		if rec.State != StateActive {
			return nil
		}

		spec, ok := s.lookupByID(rec.UnitID)
		if !ok {
			return fmt.Errorf("unknown artefact id %d", rec.UnitID)
		}
		change, err := computeChanges(spec, dirRevert)
		if err != nil && !errors.Is(err, ErrUnsupported) {
			return err
		}
		if change != nil {
			if err := applyChange(ctx, tx, *change, rec.UserID, rec.PlanetID); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx,
			`UPDATE artefacts_user SET state = $1 WHERE id = $2`,
			StateExpired, rec.ID); err != nil {
			return fmt.Errorf("expire update: %w", err)
		}
		// Системное сообщение об истечении (folder=7 ARTEFACTS).
		if s.automsg != nil {
			title := s.tr("artefact", "expired.title", nil)
			body := s.tr("artefact", "expired.body", map[string]string{
				"artefactName": spec.Name,
			})
			_ = s.automsg.SendDirect(ctx, tx, rec.UserID, 7, title, body)
		}
		return nil
	}
}
