// Package dailyquest — ежедневные задания (план 17 D).
//
// Lazy-generation: при первом GET'е в новый день для пользователя
// выбираются 3 случайных quest_def, INSERT в daily_quests с
// PK=(user, def, date). ON CONFLICT DO NOTHING защищает от race
// (одновременные запросы дают одинаковый результат).
//
// Прогресс инкрементируется hook-вызовами из доменов (research,
// fleet, planet) при event'ах. См. service.IncrementProgress.
package dailyquest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ConditionType — типы quest-условий.
const (
	ConditionResourceEarn  = "resource_earn"  // value: {"resource":"metal|silicon|hydrogen"}, прогресс = накопленные единицы
	ConditionFleetMission  = "fleet_mission"  // value: {"mission":N}, прогресс = кол-во отправленных
	ConditionResearchDone  = "research_done"  // прогресс = кол-во завершённых исследований
	ConditionBuildingDone  = "building_done"  // прогресс = кол-во достроенных зданий
)

// Quest — активный quest игрока на день.
type Quest struct {
	DefID           int       `json:"def_id"`
	Key             string    `json:"key"`
	Title           string    `json:"title"`
	ConditionType   string    `json:"condition_type"`
	ConditionValue  json.RawMessage `json:"condition_value"`
	TargetProgress  int       `json:"target_progress"`
	Progress        int       `json:"progress"`
	RewardCredits   int       `json:"reward_credits"`
	RewardMetal     int64     `json:"reward_metal"`
	RewardSilicon   int64     `json:"reward_silicon"`
	RewardHydrogen  int64     `json:"reward_hydrogen"`
	Date            time.Time `json:"date"`
	Completed       bool      `json:"completed"`
	Claimed         bool      `json:"claimed"`
}

// Sentinel errors.
var (
	ErrQuestNotFound  = errors.New("daily quest: not found")
	ErrNotCompleted   = errors.New("daily quest: not completed yet")
	ErrAlreadyClaimed = errors.New("daily quest: reward already claimed")
)

// Service — основной сервис.
type Service struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// QuestsPerDay — сколько quest'ов выдаётся игроку в день.
const QuestsPerDay = 3

// today — UTC-дата без времени.
func today() time.Time {
	t := time.Now().UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// List возвращает quest'ы игрока на сегодня. Если их ещё нет в БД —
// генерирует lazy.
func (s *Service) List(ctx context.Context, userID string) ([]Quest, error) {
	d := today()

	if err := s.ensureForToday(ctx, userID, d); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT q.def_id, d.key, d.title, d.condition_type, d.condition_value,
		       d.target_progress, q.progress,
		       d.reward_credits, d.reward_metal, d.reward_silicon, d.reward_hydrogen,
		       q.date, q.completed_at IS NOT NULL, q.claimed_at IS NOT NULL
		FROM daily_quests q
		JOIN daily_quest_defs d ON d.id = q.def_id
		WHERE q.user_id = $1 AND q.date = $2
		ORDER BY q.def_id
	`, userID, d)
	if err != nil {
		return nil, fmt.Errorf("daily quest list: %w", err)
	}
	defer rows.Close()

	var out []Quest
	for rows.Next() {
		var q Quest
		if err := rows.Scan(
			&q.DefID, &q.Key, &q.Title, &q.ConditionType, &q.ConditionValue,
			&q.TargetProgress, &q.Progress,
			&q.RewardCredits, &q.RewardMetal, &q.RewardSilicon, &q.RewardHydrogen,
			&q.Date, &q.Completed, &q.Claimed,
		); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

// ensureForToday — генерирует QuestsPerDay quest'ов для пользователя
// на сегодня, если их ещё нет. Идемпотентно: ON CONFLICT DO NOTHING.
func (s *Service) ensureForToday(ctx context.Context, userID string, d time.Time) error {
	// Быстрый путь: уже есть.
	var cnt int
	if err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM daily_quests WHERE user_id=$1 AND date=$2`,
		userID, d).Scan(&cnt); err != nil {
		return fmt.Errorf("count daily: %w", err)
	}
	if cnt >= QuestsPerDay {
		return nil
	}

	// Загрузим все def-ы с их весами.
	rows, err := s.db.Query(ctx,
		`SELECT id, weight FROM daily_quest_defs WHERE weight > 0`)
	if err != nil {
		return fmt.Errorf("load defs: %w", err)
	}
	var defs []defRow
	for rows.Next() {
		var dr defRow
		if err := rows.Scan(&dr.id, &dr.weight); err != nil {
			rows.Close()
			return err
		}
		defs = append(defs, dr)
	}
	rows.Close()
	if len(defs) == 0 {
		return errors.New("daily quest: no defs in database")
	}

	// Детерминированный seed по (userID, дата) — два запроса в один день
	// не дадут разный набор.
	seed := int64(0)
	for _, c := range userID {
		seed = seed*31 + int64(c)
	}
	seed ^= d.Unix()
	r := rand.New(rand.NewSource(seed))

	chosen := pickWeighted(defs, QuestsPerDay, r)
	for _, dr := range chosen {
		_, err := s.db.Exec(ctx, `
			INSERT INTO daily_quests (user_id, def_id, date)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id, def_id, date) DO NOTHING
		`, userID, dr.id, d)
		if err != nil {
			return fmt.Errorf("insert quest def=%d: %w", dr.id, err)
		}
	}
	return nil
}

// defRow для weighted-pick.
type defRow struct {
	id     int
	weight int
}

// pickWeighted — выбирает n уникальных def-ов с учётом weight.
func pickWeighted(items []defRow, n int, r *rand.Rand) []defRow {
	if n >= len(items) {
		out := make([]defRow, len(items))
		copy(out, items)
		return out
	}
	pool := make([]defRow, len(items))
	copy(pool, items)
	out := make([]defRow, 0, n)
	for k := 0; k < n && len(pool) > 0; k++ {
		total := 0
		for _, it := range pool {
			total += it.weight
		}
		pick := r.Intn(total)
		acc := 0
		idx := 0
		for i, it := range pool {
			acc += it.weight
			if pick < acc {
				idx = i
				break
			}
		}
		out = append(out, pool[idx])
		pool = append(pool[:idx], pool[idx+1:]...)
	}
	return out
}

// IncrementProgress — увеличить прогресс по condition_type/value на
// delta. Вызывается из доменных handler'ов. Тихо игнорирует, если у
// игрока нет подходящего quest на сегодня.
//
// matcher — функция, которая для конкретного quest-condition_value
// решает, подходит ли событие (например, mission совпадает).
func (s *Service) IncrementProgress(ctx context.Context, userID, conditionType string,
	delta int, matcher func(condValue json.RawMessage) bool) error {
	if delta <= 0 {
		return nil
	}
	d := today()

	// Найдём подходящие quest'ы на сегодня.
	rows, err := s.db.Query(ctx, `
		SELECT q.def_id, d.condition_value, d.target_progress, q.progress
		FROM daily_quests q
		JOIN daily_quest_defs d ON d.id = q.def_id
		WHERE q.user_id=$1 AND q.date=$2 AND q.completed_at IS NULL
		  AND d.condition_type=$3
	`, userID, d, conditionType)
	if err != nil {
		return fmt.Errorf("incr scan: %w", err)
	}
	type match struct {
		defID    int
		curProgr int
		target   int
	}
	var matches []match
	for rows.Next() {
		var m match
		var cv json.RawMessage
		if err := rows.Scan(&m.defID, &cv, &m.target, &m.curProgr); err != nil {
			rows.Close()
			return err
		}
		if matcher == nil || matcher(cv) {
			matches = append(matches, m)
		}
	}
	rows.Close()

	for _, m := range matches {
		newProgr := m.curProgr + delta
		if newProgr >= m.target {
			newProgr = m.target
			_, err := s.db.Exec(ctx, `
				UPDATE daily_quests
				SET progress=$1, completed_at=now()
				WHERE user_id=$2 AND def_id=$3 AND date=$4 AND completed_at IS NULL
			`, newProgr, userID, m.defID, d)
			if err != nil {
				return fmt.Errorf("complete: %w", err)
			}
		} else {
			_, err := s.db.Exec(ctx, `
				UPDATE daily_quests SET progress=$1
				WHERE user_id=$2 AND def_id=$3 AND date=$4
			`, newProgr, userID, m.defID, d)
			if err != nil {
				return fmt.Errorf("update progress: %w", err)
			}
		}
	}
	return nil
}

// Claim — забрать награду за completed quest. Транзакция: проверка +
// кредиты + ресурсы (на home-планету) + claimed_at=now.
func (s *Service) Claim(ctx context.Context, userID string, defID int) (rewardCredits int, rewardMetal, rewardSilicon, rewardHydrogen int64, err error) {
	d := today()
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("claim begin: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var (
		completedAt *time.Time
		claimedAt   *time.Time
		rc          int
		rm, rs, rh  int64
	)
	err = tx.QueryRow(ctx, `
		SELECT q.completed_at, q.claimed_at,
		       d.reward_credits, d.reward_metal, d.reward_silicon, d.reward_hydrogen
		FROM daily_quests q
		JOIN daily_quest_defs d ON d.id = q.def_id
		WHERE q.user_id=$1 AND q.def_id=$2 AND q.date=$3
		FOR UPDATE
	`, userID, defID, d).Scan(&completedAt, &claimedAt, &rc, &rm, &rs, &rh)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, 0, 0, ErrQuestNotFound
		}
		return 0, 0, 0, 0, fmt.Errorf("claim select: %w", err)
	}
	if completedAt == nil {
		return 0, 0, 0, 0, ErrNotCompleted
	}
	if claimedAt != nil {
		return 0, 0, 0, 0, ErrAlreadyClaimed
	}
	if _, err = tx.Exec(ctx,
		`UPDATE daily_quests SET claimed_at = now()
		 WHERE user_id=$1 AND def_id=$2 AND date=$3`,
		userID, defID, d); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("claim update: %w", err)
	}
	if rc > 0 {
		if _, err = tx.Exec(ctx,
			`UPDATE users SET credit = credit + $1 WHERE id=$2`,
			rc, userID); err != nil {
			return 0, 0, 0, 0, fmt.Errorf("claim credits: %w", err)
		}
	}
	if rm > 0 || rs > 0 || rh > 0 {
		var planetID string
		if err = tx.QueryRow(ctx, `
			SELECT id FROM planets
			WHERE user_id=$1 AND destroyed_at IS NULL AND is_moon=false
			ORDER BY created_at ASC LIMIT 1
		`, userID).Scan(&planetID); err != nil {
			return 0, 0, 0, 0, fmt.Errorf("claim find planet: %w", err)
		}
		if _, err = tx.Exec(ctx, `
			UPDATE planets
			SET metal = metal + $1, silicon = silicon + $2, hydrogen = hydrogen + $3
			WHERE id=$4
		`, rm, rs, rh, planetID); err != nil {
			return 0, 0, 0, 0, fmt.Errorf("claim resources: %w", err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("claim commit: %w", err)
	}
	return rc, rm, rs, rh, nil
}
