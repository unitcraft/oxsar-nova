package planet

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/pkg/ids"
	"github.com/oxsar/nova/backend/pkg/rng"
)

// StartingResources — ресурсы на старте (как у OGame classic).
// Храним в int64 (integer-резерв на старте, чтобы в Postgres ушло точное
// число без плавающей точки). В runtime-расчётах экономики ресурсы —
// float64 (см. model.go Resources), здесь нам нужны только целые числа
// для единичного INSERT.
var StartingResources = struct {
	Metal    int64
	Silicon  int64
	Hydrogen int64
}{
	Metal:    500,
	Silicon:  500,
	Hydrogen: 0,
}

// Starter создаёт первую планету игрока сразу после регистрации
// (§2.2 ТЗ: «первая планета», §5.13 защита новичков).
//
// Выделено в отдельную службу, чтобы auth-пакет не зависел от всей
// планетарной логики, а только от одной функции.
type Starter struct {
	db repo.Exec
}

func NewStarter(db repo.Exec) *Starter { return &Starter{db: db} }

// Assign создаёт планету на случайной свободной позиции и делает её
// текущей для пользователя. Возвращает id созданной планеты.
//
// Алгоритм: крутим координаты, пока не найдём свободную. Для 1..8
// галактик × 1..500 систем × 1..15 позиций в начале игры свободных
// позиций десятки миллионов, гонок практически не бывает.
// На случай полной вселенной — возвращаем ошибку после 100 попыток.
func (s *Starter) Assign(ctx context.Context, userID string) (string, error) {
	r := rng.New(seedFromUserID(userID))
	var planetID string

	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		for attempt := 0; attempt < 100; attempt++ {
			g := r.IntN(8) + 1         // 1..8
			sys := r.IntN(500) + 1     // 1..500
			pos := r.IntN(13) + 2      // 2..14 (1 и 15 — «крайности», оставим на колонизацию)

			taken, err := coordTaken(ctx, tx, g, sys, pos, false)
			if err != nil {
				return err
			}
			if taken {
				continue
			}

			id := ids.New()
			diameter := 12800 + r.IntN(2000) // 12800..14800 (стандарт OGame для pos 4-9)
			tempMax := -40 + r.IntN(80)      // ±
			tempMin := tempMax - 40

			_, err = tx.Exec(ctx, `
				INSERT INTO planets (id, user_id, is_moon, name, galaxy, system, position,
				                     diameter, used_fields, temperature_min, temperature_max,
				                     metal, silicon, hydrogen)
				VALUES ($1, $2, false, $3, $4, $5, $6, $7, 0, $8, $9, $10, $11, $12)
			`, id, userID, "Homeworld", g, sys, pos, diameter, tempMin, tempMax,
				StartingResources.Metal, StartingResources.Silicon, StartingResources.Hydrogen)
			if err != nil {
				return fmt.Errorf("insert starter planet: %w", err)
			}

			if _, err := tx.Exec(ctx,
				`UPDATE users SET cur_planet_id = $1 WHERE id = $2`, id, userID); err != nil {
				return fmt.Errorf("set cur_planet: %w", err)
			}

			// Стартовый набор зданий: по одному уровню mines + solar.
			// Без solar_plant экономический тик сразу уходит в минус по
			// энергии (storage-капа уже добавляет потребление, а
			// шахты тоже хотят энергию). Это мешает новичку и не
			// соответствует OGame-поведению, где старт — baseline 1/1/1/1.
			//
			// unit_id из configs/buildings.yml:
			//   1 = metal_mine, 2 = silicon_lab, 3 = hydrogen_lab, 4 = solar_plant.
			// hydrogen_lab оставляем на 0 (OGame classic: он стартует с 0).
			for _, unit := range []int{1, 2, 4} {
				if _, err := tx.Exec(ctx, `
					INSERT INTO buildings (planet_id, unit_id, level)
					VALUES ($1, $2, 1)
				`, id, unit); err != nil {
					return fmt.Errorf("starter building %d: %w", unit, err)
				}
			}

			// Базовый journal entry — чтобы первая запись в res_log была
			// «стартовый грант».
			if _, err := tx.Exec(ctx, `
				INSERT INTO res_log (user_id, planet_id, reason, delta_metal, delta_silicon, delta_hydrogen)
				VALUES ($1, $2, 'admin_gift', $3, $4, $5)
			`, userID, id,
				StartingResources.Metal, StartingResources.Silicon, StartingResources.Hydrogen,
			); err != nil {
				return fmt.Errorf("starter res_log: %w", err)
			}

			planetID = id
			return nil
		}
		return errors.New("planet: no free coords after 100 attempts")
	})
	if err != nil {
		return "", err
	}
	return planetID, nil
}

func coordTaken(ctx context.Context, tx pgx.Tx, g, sys, pos int, isMoon bool) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM planets
			WHERE galaxy = $1 AND system = $2 AND position = $3 AND is_moon = $4
			  AND destroyed_at IS NULL
		)
	`, g, sys, pos, isMoon).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("coord check: %w", err)
	}
	return exists, nil
}

// seedFromUserID — детерминированный seed для генератора стартовых
// координат. Разные пользователи получают разные последовательности,
// но один и тот же пользователь при retry (повторный вызов Assign) —
// ту же. Это не критично, просто меньше случайности в тестах.
func seedFromUserID(userID string) uint64 {
	// FNV-1a, чтобы не тянуть hash/fnv в этот файл ради одной строки.
	var h uint64 = 14695981039346656037
	for i := 0; i < len(userID); i++ {
		h ^= uint64(userID[i])
		h *= 1099511628211
	}
	return h
}
