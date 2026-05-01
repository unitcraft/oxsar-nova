// План 72.1.42: handler KindArtMarketExpire = 91 — auto-снятие
// артефакт-маркет лота через TTL (legacy `ads.lifetime`).
//
// При срабатывании:
//   1. SELECT offer_id из payload.
//   2. Если оффер ещё существует — UPDATE artefacts_user.state='held'
//      (revert), DELETE artefact_offers.
//   3. Если оффер уже удалён (купили или cancel) — no-op (idempotent).

package artmarket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/event"
)

// ExpireEvent — фабрика event.Handler для KindArtMarketExpire (91).
//
// Использование (cmd/worker/main.go):
//
//	w.Register(event.KindArtMarketExpire, artMarketSvc.ExpireEvent())
func (s *Service) ExpireEvent() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl struct {
			OfferID    string `json:"offer_id"`
			ArtefactID string `json:"artefact_id"`
		}
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("artmarket expire: parse payload: %w", err)
		}
		if pl.OfferID == "" {
			return fmt.Errorf("artmarket expire: empty offer_id")
		}

		// Idempotent: если offer уже удалён (купили или cancel'нули) —
		// просто выходим.
		var stillExists bool
		if err := tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM artefact_offers WHERE id = $1)`,
			pl.OfferID,
		).Scan(&stillExists); err != nil {
			return fmt.Errorf("artmarket expire: check exists: %w", err)
		}
		if !stillExists {
			return nil
		}

		// Revert state артефакта 'listed' → 'held'.
		if pl.ArtefactID != "" {
			if _, err := tx.Exec(ctx,
				`UPDATE artefacts_user SET state='held' WHERE id=$1 AND state='listed'`,
				pl.ArtefactID); err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("artmarket expire: revert state: %w", err)
			}
			}
		}

		// Удаляем оффер.
		if _, err := tx.Exec(ctx,
			`DELETE FROM artefact_offers WHERE id=$1`, pl.OfferID); err != nil {
			return fmt.Errorf("artmarket expire: delete offer: %w", err)
		}
		return nil
	}
}
