// Package officer — временные подписки-модификаторы факторов.
//
// Officer работает симметрично артефакту: при Activate → UPDATE
// factor-полей users/planets (add Delta), при Expire (event-handler) →
// revert (add -Delta). Стоимость в credit списывается при Activate.
//
// Текущие 4 officer'а:
//   ADMIRAL   (user.build_factor +0.1)    — +10% скорости сборки флота.
//   GEOLOGIST (planets.produce_factor +0.1) — +10% добычи.
//   ENGINEER  (planets.build_factor +0.25) — +25% скорости построек.
//   MERCHANT  (user.exchange_rate -0.2)   — честный паритет market.
//
// Идемпотентность: повторная активация того же officer'а, пока
// прошлая не истекла — ошибка ErrAlreadyActive (legacy тоже не
// даёт «продлить за credit», только купить снова после expire).
package officer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/event"
	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/pkg/ids"
)

type Service struct {
	db repo.Exec
}

func NewService(db repo.Exec) *Service { return &Service{db: db} }

var (
	ErrOfficerNotFound  = errors.New("officer: not found")
	ErrAlreadyActive    = errors.New("officer: already active")
	ErrGroupActive      = errors.New("officer: another officer in the same group is already active")
	ErrNotEnoughCredit  = errors.New("officer: not enough credit")
)

// Def — каталожная запись.
type Def struct {
	Key          string          `json:"key"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	DurationDays int             `json:"duration_days"`
	CostCredit   int64           `json:"cost_credit"`
	Effect       json.RawMessage `json:"effect"`
	GroupKey     *string         `json:"group_key,omitempty"`
}

// Active — активная подписка.
type Active struct {
	OfficerKey  string    `json:"officer_key"`
	ActivatedAt time.Time `json:"activated_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// Entry — элемент для UI: def + активная запись (если есть).
type Entry struct {
	Key          string     `json:"key"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	DurationDays int        `json:"duration_days"`
	CostCredit   int64      `json:"cost_credit"`
	ActivatedAt  *time.Time `json:"activated_at,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// List возвращает defs + флаг active для userID.
func (s *Service) List(ctx context.Context, userID string) ([]Entry, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT d.key, d.title, d.description, d.duration_days, d.cost_credit,
		       a.activated_at, a.expires_at
		FROM officer_defs d
		LEFT JOIN officer_active a
		  ON a.officer_key = d.key AND a.user_id = $1
		ORDER BY d.cost_credit DESC, d.key ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("officers list: %w", err)
	}
	defer rows.Close()
	var out []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.Key, &e.Title, &e.Description,
			&e.DurationDays, &e.CostCredit, &e.ActivatedAt, &e.ExpiresAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// factorChange — payload effect JSONB.
type factorChange struct {
	Scope string  `json:"scope"` // user | all_planets
	Field string  `json:"field"`
	Op    string  `json:"op"`    // add | set
	Delta float64 `json:"delta"`
}

// allowedField защищает SQL-инъекцию через кривой YAML.
func allowedField(f string) bool {
	switch f {
	case "exchange_rate", "research_factor",
		"build_factor", "produce_factor", "energy_factor", "storage_factor":
		return true
	}
	return false
}

// Activate покупает officer'а за credit и применяет эффект.
// При успехе создаёт event KindOfficerExpire=62 на expires_at.
// autoRenew=true — при истечении срока автоматически продлевает подписку,
// если у игрока хватает credit.
func (s *Service) Activate(ctx context.Context, userID, key string, autoRenew bool) (Entry, error) {
	var out Entry
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Def.
		var def Def
		err := tx.QueryRow(ctx, `
			SELECT key, title, description, duration_days, cost_credit, effect, group_key
			FROM officer_defs WHERE key = $1
		`, key).Scan(&def.Key, &def.Title, &def.Description,
			&def.DurationDays, &def.CostCredit, &def.Effect, &def.GroupKey)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrOfficerNotFound
			}
			return fmt.Errorf("read def: %w", err)
		}
		// Parse effect.
		var eff factorChange
		if err := json.Unmarshal(def.Effect, &eff); err != nil {
			return fmt.Errorf("parse effect: %w", err)
		}
		if !allowedField(eff.Field) {
			return fmt.Errorf("unsupported field %q", eff.Field)
		}
		if eff.Op != "add" {
			return fmt.Errorf("unsupported op %q (officer: only add)", eff.Op)
		}
		// Not already active.
		var exists bool
		if err := tx.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM officer_active WHERE user_id=$1 AND officer_key=$2)`,
			userID, key).Scan(&exists); err != nil {
			return fmt.Errorf("check active: %w", err)
		}
		if exists {
			return ErrAlreadyActive
		}
		// Group exclusivity: если у офицера group_key, проверяем нет ли другого
		// активного офицера из той же группы.
		if def.GroupKey != nil && *def.GroupKey != "" {
			var groupActive bool
			if err := tx.QueryRow(ctx, `
				SELECT EXISTS (
					SELECT 1 FROM officer_active a
					JOIN officer_defs d ON d.key = a.officer_key
					WHERE a.user_id = $1 AND d.group_key = $2
				)
			`, userID, *def.GroupKey).Scan(&groupActive); err != nil {
				return fmt.Errorf("check group: %w", err)
			}
			if groupActive {
				return ErrGroupActive
			}
		}
		// Списываем credit.
		var credit int64
		if err := tx.QueryRow(ctx,
			`SELECT credit FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&credit); err != nil {
			return fmt.Errorf("read credit: %w", err)
		}
		if credit < def.CostCredit {
			return ErrNotEnoughCredit
		}
		if _, err := tx.Exec(ctx,
			`UPDATE users SET credit = credit - $1 WHERE id = $2`,
			def.CostCredit, userID); err != nil {
			return fmt.Errorf("debit credit: %w", err)
		}
		// Применяем factor.
		if err := applyFactor(ctx, tx, userID, eff, +eff.Delta); err != nil {
			return err
		}
		// INSERT active.
		now := time.Now().UTC()
		exp := now.Add(time.Duration(def.DurationDays) * 24 * time.Hour)
		if _, err := tx.Exec(ctx, `
			INSERT INTO officer_active (user_id, officer_key, activated_at, expires_at, auto_renew)
			VALUES ($1, $2, $3, $4, $5)
		`, userID, key, now, exp, autoRenew); err != nil {
			return fmt.Errorf("insert active: %w", err)
		}
		// Event на expire.
		payload, _ := json.Marshal(map[string]any{
			"user_id":      userID,
			"officer_key":  key,
			"effect":       def.Effect,
			"cost_credit":  def.CostCredit,
			"duration_days": def.DurationDays,
			"auto_renew":   autoRenew,
		})
		if _, err := tx.Exec(ctx, `
			INSERT INTO events (id, user_id, kind, state, fire_at, payload)
			VALUES ($1, $2, $3, 'wait', $4, $5)
		`, ids.New(), userID, event.KindOfficerExpire, exp, payload); err != nil {
			return fmt.Errorf("insert event: %w", err)
		}
		out = Entry{
			Key: def.Key, Title: def.Title, Description: def.Description,
			DurationDays: def.DurationDays, CostCredit: def.CostCredit,
			ActivatedAt: &now, ExpiresAt: &exp,
		}
		return nil
	})
	return out, err
}

// applyFactor — UPDATE factor-поля в соответствии со scope.
// delta >0 — активация, <0 — revert (зеркально).
func applyFactor(ctx context.Context, tx pgx.Tx, userID string, eff factorChange, delta float64) error {
	if !allowedField(eff.Field) {
		return fmt.Errorf("unsupported field %q", eff.Field)
	}
	switch eff.Scope {
	case "user":
		_, err := tx.Exec(ctx,
			`UPDATE users SET `+eff.Field+` = `+eff.Field+` + $1 WHERE id = $2`,
			delta, userID)
		return err
	case "all_planets":
		_, err := tx.Exec(ctx, `
			UPDATE planets SET `+eff.Field+` = `+eff.Field+` + $1
			WHERE user_id = $2 AND destroyed_at IS NULL
		`, delta, userID)
		return err
	default:
		return fmt.Errorf("unsupported scope %q", eff.Scope)
	}
}

// ExpireHandler — event.Handler для KindOfficerExpire=62.
// Revert factor + DELETE active-row. Идемпотентно через DELETE
// по (user, key): если row уже удалена, revert не повторится.
// Если auto_renew=true и у игрока хватает credit — продлевает автоматически.
func (s *Service) ExpireHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		var pl struct {
			UserID       string          `json:"user_id"`
			OfficerKey   string          `json:"officer_key"`
			Effect       json.RawMessage `json:"effect"`
			CostCredit   int64           `json:"cost_credit"`
			DurationDays int             `json:"duration_days"`
			AutoRenew    bool            `json:"auto_renew"`
		}
		if err := json.Unmarshal(e.Payload, &pl); err != nil {
			return fmt.Errorf("officer expire: payload: %w", err)
		}
		// Проверяем, что запись ещё существует (идемпотентность).
		tag, err := tx.Exec(ctx,
			`DELETE FROM officer_active WHERE user_id = $1 AND officer_key = $2`,
			pl.UserID, pl.OfficerKey)
		if err != nil {
			return fmt.Errorf("delete active: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return nil // уже удалена — второй запуск
		}
		var eff factorChange
		if err := json.Unmarshal(pl.Effect, &eff); err != nil {
			return fmt.Errorf("parse effect: %w", err)
		}

		// Auto-renew: если флаг установлен и у игрока хватает credit — продлеваем.
		if pl.AutoRenew && pl.CostCredit > 0 && pl.DurationDays > 0 {
			var credit int64
			if err := tx.QueryRow(ctx,
				`SELECT credit FROM users WHERE id=$1 FOR UPDATE`, pl.UserID).Scan(&credit); err == nil &&
				credit >= pl.CostCredit {
				if _, err := tx.Exec(ctx,
					`UPDATE users SET credit = credit - $1 WHERE id = $2`,
					pl.CostCredit, pl.UserID); err != nil {
					return fmt.Errorf("auto_renew debit: %w", err)
				}
				now := time.Now().UTC()
				exp := now.Add(time.Duration(pl.DurationDays) * 24 * time.Hour)
				if _, err := tx.Exec(ctx, `
					INSERT INTO officer_active (user_id, officer_key, activated_at, expires_at, auto_renew)
					VALUES ($1, $2, $3, $4, true)
				`, pl.UserID, pl.OfficerKey, now, exp); err != nil {
					return fmt.Errorf("auto_renew insert: %w", err)
				}
				newPayload, _ := json.Marshal(map[string]any{
					"user_id":       pl.UserID,
					"officer_key":   pl.OfficerKey,
					"effect":        pl.Effect,
					"cost_credit":   pl.CostCredit,
					"duration_days": pl.DurationDays,
					"auto_renew":    true,
				})
				if _, err := tx.Exec(ctx, `
					INSERT INTO events (id, user_id, kind, state, fire_at, payload)
					VALUES ($1, $2, $3, 'wait', $4, $5)
				`, ids.New(), pl.UserID, event.KindOfficerExpire, exp, newPayload); err != nil {
					return fmt.Errorf("auto_renew event: %w", err)
				}
				// factor не меняется (был активен → остаётся активен).
				if _, err := tx.Exec(ctx, `
					INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
					VALUES ($1, $2, NULL, 13, $3, $4)
				`, ids.New(), pl.UserID,
					fmt.Sprintf("Officer %s продлён автоматически", pl.OfficerKey),
					fmt.Sprintf("Подписка продлена. Списано %d кредитов.", pl.CostCredit),
				); err != nil {
					return fmt.Errorf("auto_renew notify: %w", err)
				}
				return nil
			}
			// Недостаточно credit — сбрасываем factor и уведомляем.
		}

		if err := applyFactor(ctx, tx, pl.UserID, eff, -eff.Delta); err != nil {
			return fmt.Errorf("revert factor: %w", err)
		}
		// Уведомление.
		subject := fmt.Sprintf("Officer %s истёк", pl.OfficerKey)
		body := "Срок подписки закончился. Активируйте снова, если нужно."
		if pl.AutoRenew {
			body = "Срок подписки закончился. Недостаточно кредитов для авто-продления."
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body)
			VALUES ($1, $2, NULL, 13, $3, $4)
		`, ids.New(), pl.UserID, subject, body); err != nil {
			return fmt.Errorf("notify: %w", err)
		}
		return nil
	}
}
