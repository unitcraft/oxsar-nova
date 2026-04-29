package planet

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/miner"
)

// applyMinerTick — обновление of_points / of_level / credit при добыче
// (план 72.1 ч.17). Вызывается из applyTickInTx после успешного UPDATE
// planets. Транзакция та же — атомарность гарантирована.
//
// Логика (legacy `Functions.inc.php:1633-1665`):
//  1. Читаем (FOR UPDATE) текущие of_points / of_level.
//  2. miner.LevelUp накапливает points, считает level-ups и награду.
//  3. Если ничего не изменилось — выходим без записи.
//  4. Один UPDATE: новый of_points, of_level, credit (если есть награда).
//
// `addedPoints` — фактически добытая сумма ресурсов (после clamp на cap),
// округлённая до int64. Отрицательные / нулевые значения игнорируются.
func applyMinerTick(ctx context.Context, tx pgx.Tx, userID string, addedPoints int64) error {
	if addedPoints <= 0 {
		return nil
	}
	var curLevel int
	var curPoints int64
	if err := tx.QueryRow(ctx,
		`SELECT of_level, of_points::bigint FROM users WHERE id=$1 FOR UPDATE`,
		userID,
	).Scan(&curLevel, &curPoints); err != nil {
		return fmt.Errorf("read of_*: %w", err)
	}

	res := miner.LevelUp(curLevel, curPoints, addedPoints)

	// Если уровень не изменился и award=0 — экономим один UPDATE.
	// Поле of_points всё равно меняется (curPoints != res.NewPoints когда
	// add>0 и нет level-up'а), поэтому в этом случае писать обязательно.
	if res.LevelUps == 0 && res.NewLevel == curLevel && res.NewPoints == curPoints {
		return nil
	}

	if res.CreditsAwarded > 0 {
		_, err := tx.Exec(ctx, `
			UPDATE users
			SET of_points = $1,
			    of_level = $2,
			    credit = credit + $3
			WHERE id = $4
		`, res.NewPoints, res.NewLevel, res.CreditsAwarded, userID)
		if err != nil {
			return fmt.Errorf("update miner: %w", err)
		}
		return nil
	}

	_, err := tx.Exec(ctx, `
		UPDATE users
		SET of_points = $1, of_level = $2
		WHERE id = $3
	`, res.NewPoints, res.NewLevel, userID)
	if err != nil {
		return fmt.Errorf("update miner: %w", err)
	}
	return nil
}
