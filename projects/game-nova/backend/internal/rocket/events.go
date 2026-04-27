package rocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/pkg/ids"
)

// interceptorRocketUnitID — unit_id юнита "interceptor_rocket" (§5.16).
const interceptorRocketUnitID = 51

// ImpactHandler — event.Handler для KindRocketAttack=16. Идемпотентен
// через отметку events.state воркером.
//
// Модель урона:
//   intercepted = min(abm_count, rocket_count)
//   surviving = rocket_count - intercepted
//   abm_count -= intercepted   (перехватчики расходуются)
//   totalDamage = surviving × missileDamage
//   для каждого defense-стека планеты:
//     share = stack.count × shell / Σ(count × shell)
//     killed = floor(totalDamage × share / shell)
//     new_count = max(0, stack.count - killed)
//
// Ракеты бьют только defense. Щитов и hp-такого защита не имеет
// (legacy-модель: defense имеет только `count` и `cost`).
func (s *Service) ImpactHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl struct {
			ImpactID     string `json:"impact_id"`
			AttackerID   string `json:"attacker_id"`
			SrcPlanet    string `json:"src_planet"`
			Dst          struct {
				Galaxy   int  `json:"galaxy"`
				System   int  `json:"system"`
				Position int  `json:"position"`
				IsMoon   bool `json:"is_moon"`
			} `json:"dst"`
			Count        int64 `json:"count"`
			TargetUnitID int   `json:"target_unit_id"` // 0 = не указана (равномерно)
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
				coords := fmt.Sprintf("%d:%d:%d", pl.Dst.Galaxy, pl.Dst.System, pl.Dst.Position)
				_, _ = tx.Exec(ctx, `
					INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
					VALUES ($1, $2, NULL, 2, $3, $4)
				`, ids.New(), pl.AttackerID,
					s.tr("rocket", "failed.title", map[string]string{"coords": coords}),
					s.tr("rocket", "failed.body", nil))
				return nil
			}
			return fmt.Errorf("rocket impact: find target: %w", err)
		}

		// Читаем defense-таблицу цели.
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

		// Anti-ballistic missile interception.
		survivingRockets := pl.Count
		var abmIntercepted int64
		for i, d := range stacks {
			if d.UnitID != interceptorRocketUnitID {
				continue
			}
			abmIntercepted = d.Count
			if abmIntercepted > survivingRockets {
				abmIntercepted = survivingRockets
			}
			survivingRockets -= abmIntercepted
			newABM := d.Count - abmIntercepted
			if _, err := tx.Exec(ctx,
				`UPDATE defense SET count=$1 WHERE planet_id=$2 AND unit_id=$3`,
				newABM, planetID, interceptorRocketUnitID); err != nil {
				return fmt.Errorf("rocket impact: update abm: %w", err)
			}
			// Remove ABM from stacks so it's not included in damage calc.
			stacks = append(stacks[:i], stacks[i+1:]...)
			break
		}

		totalDamage := survivingRockets * int64(missileDamage)

		coords := fmt.Sprintf("%d:%d:%d", pl.Dst.Galaxy, pl.Dst.System, pl.Dst.Position)
		if len(stacks) == 0 || survivingRockets == 0 {
			// Все ракеты сбиты или нет defense — пишем сообщение.
			subj := s.tr("rocket", "impact.title", map[string]string{"coords": coords})
			var body string
			if survivingRockets == 0 {
				body = s.tr("rocket", "impact.allIntercepted", map[string]string{
					"intercepted": fmt.Sprintf("%d", abmIntercepted),
					"abm":         fmt.Sprintf("%d", abmIntercepted),
				})
			} else {
				body = s.tr("rocket", "impact.noDefense", map[string]string{
					"count":  fmt.Sprintf("%d", survivingRockets),
					"damage": fmt.Sprintf("%d", totalDamage),
				})
			}
			return notifyBoth(ctx, tx, pl.AttackerID, targetOwner, subj, body)
		}

		// Если задана приоритетная цель — бьём её первой, остаток урона
		// распределяем по оставшимся стекам.
		losses := applyRocketDamage(survivingRockets, stacks, pl.TargetUnitID)

		// Применяем потери в БД.
		for _, l := range losses {
			var cur int64
			for _, d := range stacks {
				if d.UnitID == l.UnitID {
					cur = d.Count
					break
				}
			}
			if _, err := tx.Exec(ctx,
				`UPDATE defense SET count=$1 WHERE planet_id=$2 AND unit_id=$3`,
				cur-l.Lost, planetID, l.UnitID); err != nil {
				return fmt.Errorf("rocket impact: update defense unit %d: %w", l.UnitID, err)
			}
		}

		// Сообщение обеим сторонам.
		subj := s.tr("rocket", "impact.title", map[string]string{"coords": coords})
		lossStr := ""
		for _, l := range losses {
			lossStr += s.tr("rocket", "impact.unitLoss", map[string]string{
				"unitId": fmt.Sprintf("%d", l.UnitID),
				"lost":   fmt.Sprintf("%d", l.Lost),
			})
		}
		abmNote := ""
		if abmIntercepted > 0 {
			abmNote = s.tr("rocket", "impact.abmNote", map[string]string{
				"count": fmt.Sprintf("%d", abmIntercepted),
			})
		}
		body := s.tr("rocket", "impact.body", map[string]string{
			"surviving": fmt.Sprintf("%d", survivingRockets),
			"total":     fmt.Sprintf("%d", pl.Count),
			"abmNote":   abmNote,
			"damage":    fmt.Sprintf("%d", totalDamage),
			"losses":    lossStr,
		})
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
