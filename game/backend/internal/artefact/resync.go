package artefact

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/config"
)

// Дефолтные значения factor-полей (см. §5.10.1 ТЗ и
// Artefact.class.php::resyncUser).
const (
	DefaultExchangeRate  = 1.2
	DefaultFactorValue   = 1.0 // research_factor / build_factor / produce_factor / energy_factor / storage_factor
)

// ResyncUser сбрасывает все factor-поля игрока и всех его планет в
// дефолт, затем переприменяет все активные артефакты.
//
// Используется:
//   - при подозрении на рассинхронизацию (админ-утилита, cron);
//   - при старте нового сезона/universe_reset.
//
// Идемпотентно: два подряд вызова дают один и тот же результат.
func (s *Service) ResyncUser(ctx context.Context, userID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. Сбросить users.{exchange_rate, research_factor}.
		if _, err := tx.Exec(ctx, `
			UPDATE users
			SET exchange_rate = $1, research_factor = $2
			WHERE id = $3
		`, DefaultExchangeRate, DefaultFactorValue, userID); err != nil {
			return fmt.Errorf("reset user factors: %w", err)
		}

		// 2. Сбросить planets.*_factor у всех планет игрока.
		if _, err := tx.Exec(ctx, `
			UPDATE planets
			SET build_factor = $1, produce_factor = $1,
			    energy_factor = $1, storage_factor = $1
			WHERE user_id = $2
		`, DefaultFactorValue, userID); err != nil {
			return fmt.Errorf("reset planet factors: %w", err)
		}

		// 3. Прочитать активные артефакты в порядке активации и
		// переприменить каждый. Порядок важен только для set-эффектов
		// (последний set выигрывает); для add он коммутативен.
		rows, err := tx.Query(ctx, `
			SELECT id, user_id, planet_id, unit_id, state, acquired_at, activated_at, expire_at
			FROM artefacts_user
			WHERE user_id = $1 AND state = $2
			ORDER BY activated_at ASC NULLS LAST, acquired_at ASC
		`, userID, StateActive)
		if err != nil {
			return fmt.Errorf("list active: %w", err)
		}
		defer rows.Close()

		type active struct {
			rec  Record
			spec config.ArtefactSpec
		}
		var actives []active
		for rows.Next() {
			var r Record
			if err := rows.Scan(&r.ID, &r.UserID, &r.PlanetID, &r.UnitID, &r.State,
				&r.AcquiredAt, &r.ActivatedAt, &r.ExpireAt); err != nil {
				return fmt.Errorf("scan: %w", err)
			}
			spec, ok := s.lookupByID(r.UnitID)
			if !ok {
				// артефакт неизвестен каталогу — пропускаем;
				// возможно, если YAML был урезан между деплоями.
				continue
			}
			actives = append(actives, active{rec: r, spec: spec})
		}

		for _, a := range actives {
			change, err := computeChanges(a.spec, dirApply)
			if err != nil && !errors.Is(err, ErrUnsupported) {
				return err
			}
			if change == nil {
				continue
			}
			if err := applyChange(ctx, tx, *change, userID, a.rec.PlanetID); err != nil {
				return fmt.Errorf("reapply %s: %w", a.rec.ID, err)
			}
		}
		return nil
	})
}
