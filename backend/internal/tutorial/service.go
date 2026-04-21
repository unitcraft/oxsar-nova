// Package tutorial — стартовый квест-туториал (§5.20 ТЗ).
//
// 6 шагов: build metal_mine → solar_plant → research_lab →
// research computer_tech → build first ship → send expedition.
// Прогресс хранится в users.tutorial_state (0 = шаг 1 ещё не выполнен,
// 6 = всё завершено). Каждый выполненный шаг даёт 10 кредитов
// (идемпотентно через tutorial_rewards).
package tutorial

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/repo"
)

// Step — один шаг туториала.
type Step struct {
	Index       int    `json:"index"`       // 1..6
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

const totalSteps = 6
const rewardPerStep = 10 // кредиты

// stepResources — ресурсы на стартовую планету за каждый шаг (metal, silicon, hydrogen).
var stepResources = [totalSteps][3]float64{
	{500, 200, 0},    // шаг 1: металл за шахту
	{300, 300, 100},  // шаг 2: ресурсы за электростанцию
	{500, 500, 200},  // шаг 3: ресурсы за лабораторию
	{1000, 500, 300}, // шаг 4: ресурсы за исследование
	{2000, 1000, 500},// шаг 5: ресурсы за первый корабль
	{5000, 3000, 1000},// шаг 6: финальная награда за экспедицию
}

// steps — статичный каталог шагов (текст на русском, i18n-ключи можно
// добавить позже). Условия проверяются в checkStep.
var steps = []Step{
	{1, "Шахта металла", "Постройте шахту металла до уровня 1.", false},
	{2, "Солнечная электростанция", "Постройте солнечную электростанцию до уровня 1.", false},
	{3, "Лаборатория исследований", "Постройте лабораторию исследований до уровня 1.", false},
	{4, "Компьютерные технологии", "Исследуйте компьютерные технологии до уровня 1.", false},
	{5, "Первый корабль", "Постройте любой корабль.", false},
	{6, "Экспедиция", "Отправьте экспедиционный флот (позиция 16 в системе).", false},
}

type Service struct {
	db repo.Exec
}

func NewService(db repo.Exec) *Service { return &Service{db: db} }

// Status возвращает текущее состояние туториала для игрока.
// Lazy-check: проверяем все шаги до текущего состояния,
// автоматически продвигаем вперёд и выдаём награды.
func (s *Service) Status(ctx context.Context, userID string) ([]Step, int, error) {
	var state int
	if err := s.db.Pool().QueryRow(ctx,
		`SELECT tutorial_state FROM users WHERE id = $1`, userID).Scan(&state); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, fmt.Errorf("tutorial: user not found")
		}
		return nil, 0, fmt.Errorf("tutorial: read state: %w", err)
	}

	// Проверяем следующий невыполненный шаг.
	newState := state
	for newState < totalSteps {
		done, err := checkStep(ctx, s.db, userID, newState+1)
		if err != nil {
			return nil, 0, err
		}
		if !done {
			break
		}
		newState++
		if err := s.advanceAndReward(ctx, userID, newState); err != nil {
			return nil, 0, err
		}
	}

	if newState != state {
		state = newState
	}

	out := make([]Step, totalSteps)
	for i, st := range steps {
		cp := st
		cp.Done = i+1 <= state
		out[i] = cp
	}
	return out, state, nil
}

// advanceAndReward атомарно продвигает tutorial_state, выдаёт кредиты
// и начисляет ресурсы на первую (стартовую) планету игрока.
func (s *Service) advanceAndReward(ctx context.Context, userID string, step int) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			INSERT INTO tutorial_rewards (user_id, step)
			VALUES ($1, $2) ON CONFLICT DO NOTHING
		`, userID, step)
		if err != nil {
			return fmt.Errorf("tutorial: insert reward: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return nil // уже выдавали
		}
		if _, err := tx.Exec(ctx, `
			UPDATE users SET tutorial_state = $1, credit = credit + $2
			WHERE id = $3 AND tutorial_state < $1
		`, step, rewardPerStep, userID); err != nil {
			return fmt.Errorf("tutorial: advance state: %w", err)
		}
		// Ресурсы на стартовую планету (первая созданная, не луна).
		res := stepResources[step-1]
		if res[0] > 0 || res[1] > 0 || res[2] > 0 {
			if _, err := tx.Exec(ctx, `
				UPDATE planets
				SET metal    = metal    + $1,
				    silicon  = silicon  + $2,
				    hydrogen = hydrogen + $3
				WHERE id = (
					SELECT id FROM planets
					WHERE user_id = $4 AND destroyed_at IS NULL AND is_moon = false
					ORDER BY created_at ASC LIMIT 1
				)
			`, res[0], res[1], res[2], userID); err != nil {
				return fmt.Errorf("tutorial: resource reward step %d: %w", step, err)
			}
		}
		return nil
	})
}

// checkStep проверяет, выполнен ли шаг номер step (1-based).
func checkStep(ctx context.Context, db repo.Exec, userID string, step int) (bool, error) {
	var ok bool
	var err error
	switch step {
	case 1: // metal_mine (unit_id=1) level >= 1
		err = db.Pool().QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM buildings b
				JOIN planets p ON p.id = b.planet_id
				WHERE p.user_id = $1 AND b.unit_id = 1 AND b.level >= 1
			)`, userID).Scan(&ok)
	case 2: // solar_plant (unit_id=3) level >= 1
		err = db.Pool().QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM buildings b
				JOIN planets p ON p.id = b.planet_id
				WHERE p.user_id = $1 AND b.unit_id = 3 AND b.level >= 1
			)`, userID).Scan(&ok)
	case 3: // research_lab (unit_id=8) level >= 1
		err = db.Pool().QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM buildings b
				JOIN planets p ON p.id = b.planet_id
				WHERE p.user_id = $1 AND b.unit_id = 8 AND b.level >= 1
			)`, userID).Scan(&ok)
	case 4: // computer_tech (unit_id=106) level >= 1
		err = db.Pool().QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM research
				WHERE user_id = $1 AND unit_id = 106 AND level >= 1
			)`, userID).Scan(&ok)
	case 5: // любой корабль (unit_id 31..42) построен
		err = db.Pool().QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM ships s
				JOIN planets p ON p.id = s.planet_id
				WHERE p.user_id = $1 AND s.unit_id BETWEEN 31 AND 42 AND s.count > 0
			)`, userID).Scan(&ok)
	case 6: // отправлена хотя бы одна экспедиция (fleet mission=15)
		err = db.Pool().QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM fleets
				WHERE owner_user_id = $1 AND mission = 15
			)`, userID).Scan(&ok)
	default:
		return false, fmt.Errorf("tutorial: unknown step %d", step)
	}
	if err != nil {
		return false, fmt.Errorf("tutorial: check step %d: %w", step, err)
	}
	return ok, nil
}
