package rocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/event"
	"github.com/oxsar/nova/backend/pkg/ids"
)

// ImpactHandler — event.Handler для KindRocketAttack=16. Идемпотентен
// через отметку events.state воркером.
//
// Модель урона (M5):
//   totalDamage = count × missileDamage
//   для каждого defense-стека планеты:
//     share = stack.count × shell / Σ(count × shell)
//     damaged = floor((totalDamage × share) / shell)
//     new_count = max(0, stack.count - damaged)
//
// Ракеты бьют только defense. Щитов и hp-такого защита не имеет
// (legacy-модель: defense имеет только `count` и `cost`).
func (s *Service) ImpactHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl struct {
			ImpactID   string `json:"impact_id"`
			AttackerID string `json:"attacker_id"`
			SrcPlanet  string `json:"src_planet"`
			Dst        struct {
				Galaxy   int  `json:"galaxy"`
				System   int  `json:"system"`
				Position int  `json:"position"`
				IsMoon   bool `json:"is_moon"`
			} `json:"dst"`
			Count int64 `json:"count"`
		}
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("rocket impact: payload: %w", err)
		}

		// Находим цель.
		var (
			planetID     string
			targetOwner  string
		)
		err := tx.QueryRow(ctx, `
			SELECT id, user_id FROM planets
			WHERE galaxy=$1 AND system=$2 AND position=$3 AND is_moon=$4
			  AND destroyed_at IS NULL
			FOR UPDATE
		`, pl.Dst.Galaxy, pl.Dst.System, pl.Dst.Position, pl.Dst.IsMoon).
			Scan(&planetID, &targetOwner)
		if err != nil {
			if err == pgx.ErrNoRows {
				// Цели нет — ракеты «улетели в пустоту». Пишем сообщение
				// атакующему и выходим.
				_, _ = tx.Exec(ctx, `
					INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
					VALUES ($1, $2, NULL, 2, $3, $4)
				`, ids.New(), pl.AttackerID,
					fmt.Sprintf("Ракетный удар %d:%d:%d провалился",
						pl.Dst.Galaxy, pl.Dst.System, pl.Dst.Position),
					"Цель не найдена (уничтожена или не существовала).")
				return nil
			}
			return fmt.Errorf("rocket impact: find target: %w", err)
		}

		// Читаем defense-таблицу цели.
		type defStack struct {
			UnitID int
			Count  int64
			Shell  int
		}
		var stacks []defStack
		rows, err := tx.Query(ctx,
			`SELECT unit_id, count FROM defense WHERE planet_id=$1 AND count>0 FOR UPDATE`,
			planetID)
		if err != nil {
			return fmt.Errorf("rocket impact: read defense: %w", err)
		}
		for rows.Next() {
			var d defStack
			if err := rows.Scan(&d.UnitID, &d.Count); err != nil {
				rows.Close()
				return err
			}
			// shell берём из каталога.
			for _, spec := range s.catalog.Defense.Defense {
				if spec.ID == d.UnitID {
					d.Shell = spec.Shell
					break
				}
			}
			if d.Shell <= 0 {
				d.Shell = 1000 // fallback
			}
			stacks = append(stacks, d)
		}
		rows.Close()

		totalDamage := pl.Count * int64(missileDamage)

		if len(stacks) == 0 {
			// У цели нет defense — пишем сообщение «оборона отсутствует,
			// урон пропал».
			subj := fmt.Sprintf("Ракетный удар %d:%d:%d", pl.Dst.Galaxy, pl.Dst.System, pl.Dst.Position)
			body := fmt.Sprintf("%d ракет долетели. Оборона отсутствует — урон %d пропал.",
				pl.Count, totalDamage)
			return notifyBoth(ctx, tx, pl.AttackerID, targetOwner, subj, body)
		}

		// Σ(count × shell) для нормировки.
		var totalPool int64
		for _, d := range stacks {
			totalPool += d.Count * int64(d.Shell)
		}
		if totalPool <= 0 {
			return nil // не должно, но защитимся
		}

		type loss struct {
			UnitID int
			Lost   int64
		}
		var losses []loss
		for _, d := range stacks {
			share := float64(d.Count*int64(d.Shell)) / float64(totalPool)
			dmg := int64(float64(totalDamage) * share)
			killed := dmg / int64(d.Shell)
			if killed > d.Count {
				killed = d.Count
			}
			if killed <= 0 {
				continue
			}
			newCount := d.Count - killed
			if _, err := tx.Exec(ctx,
				`UPDATE defense SET count=$1 WHERE planet_id=$2 AND unit_id=$3`,
				newCount, planetID, d.UnitID); err != nil {
				return fmt.Errorf("rocket impact: update defense: %w", err)
			}
			losses = append(losses, loss{UnitID: d.UnitID, Lost: killed})
		}

		// Сообщение обеим сторонам.
		subj := fmt.Sprintf("Ракетный удар %d:%d:%d", pl.Dst.Galaxy, pl.Dst.System, pl.Dst.Position)
		lossStr := ""
		for _, l := range losses {
			lossStr += fmt.Sprintf("\n- юнит #%d: -%d", l.UnitID, l.Lost)
		}
		body := fmt.Sprintf("%d ракет долетели (общий урон %d). Потери обороны:%s",
			pl.Count, totalDamage, lossStr)
		return notifyBoth(ctx, tx, pl.AttackerID, targetOwner, subj, body)
	}
}

func notifyBoth(ctx context.Context, tx pgx.Tx, attUID, defUID, subject, body string) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
		VALUES ($1, $2, $3, 2, $4, $5)
	`, ids.New(), attUID, defUID, subject, body); err != nil {
		return fmt.Errorf("rocket notify attacker: %w", err)
	}
	if defUID != "" && defUID != attUID {
		if _, err := tx.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
			VALUES ($1, $2, $3, 2, $4, $5)
		`, ids.New(), defUID, attUID, subject, body); err != nil {
			return fmt.Errorf("rocket notify defender: %w", err)
		}
	}
	return nil
}
