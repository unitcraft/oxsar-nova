package goal

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Rewarder — атомарное награждение пользователя за completed goal.
//
// MVP: credits + ресурсы на home-планету. Дальше расширяется (юниты,
// артефакты, исследования) без изменения интерфейса.
type Rewarder interface {
	Grant(ctx context.Context, tx pgx.Tx, userID string, reward Reward) error
}

// SimpleRewarder — реализация по умолчанию: credits в users.credit,
// ресурсы — на самую старую (home) планету пользователя.
//
// Все мутации в одной транзакции (передаётся снаружи).
type SimpleRewarder struct{}

func NewSimpleRewarder() *SimpleRewarder { return &SimpleRewarder{} }

// Grant выдаёт reward. Должен вызываться внутри уже открытой
// транзакции (atomic с UPDATE goal_progress.claimed_at).
func (r *SimpleRewarder) Grant(ctx context.Context, tx pgx.Tx, userID string, reward Reward) error {
	if reward.Empty() {
		return nil
	}
	if reward.Credits > 0 {
		if _, err := tx.Exec(ctx,
			`UPDATE users SET credit = credit + $1 WHERE id = $2`,
			reward.Credits, userID,
		); err != nil {
			return fmt.Errorf("grant credits: %w", err)
		}
	}
	if reward.Metal > 0 || reward.Silicon > 0 || reward.Hydrogen > 0 {
		// Home-планета: самая старая по created_at, не уничтожена, не луна.
		var planetID string
		err := tx.QueryRow(ctx, `
			SELECT id FROM planets
			WHERE user_id = $1 AND destroyed_at IS NULL AND is_moon = false
			ORDER BY created_at ASC LIMIT 1
		`, userID).Scan(&planetID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// Нет планет — кредиты выданы, ресурсы тихо пропадают.
				// Должно быть невозможно для активного игрока, но
				// защищаемся от corner-case.
				return nil
			}
			return fmt.Errorf("grant find home planet: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE planets
			SET metal = metal + $1, silicon = silicon + $2, hydrogen = hydrogen + $3
			WHERE id = $4
		`, reward.Metal, reward.Silicon, reward.Hydrogen, planetID); err != nil {
			return fmt.Errorf("grant resources: %w", err)
		}
	}
	return nil
}
