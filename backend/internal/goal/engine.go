package goal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/pkg/ids"
)

// Engine — основной сервис.
type Engine struct {
	catalog  *Catalog
	db       repo.Exec
	rewarder Rewarder
	notifier Notifier
	now      func() time.Time // подменяется в тестах
}

// NewEngine собирает движок. catalog не должен быть nil; rewarder/
// notifier — могут быть nil (используется no-op по умолчанию).
func NewEngine(catalog *Catalog, db repo.Exec, rewarder Rewarder, notifier Notifier) *Engine {
	if catalog == nil {
		catalog = &Catalog{byKey: map[string]GoalDef{}}
	}
	if rewarder == nil {
		rewarder = NewSimpleRewarder()
	}
	if notifier == nil {
		notifier = NoopNotifier{}
	}
	return &Engine{
		catalog:  catalog,
		db:       db,
		rewarder: rewarder,
		notifier: notifier,
		now:      time.Now,
	}
}

// Catalog — getter для доступа к списку целей (UI).
func (e *Engine) Catalog() *Catalog { return e.catalog }

// Recompute — пересчитать прогресс snapshot-цели для пользователя.
// Идемпотентно: completed/claimed повторно не меняются.
//
// goalKey — ключ из YAML.
// period   — '' для permanent/one-time/seasonal; для daily/weekly
//
//	вычисляется автоматически по now() (передавать пустой ОК).
func (e *Engine) Recompute(ctx context.Context, userID, goalKey string) error {
	def, ok := e.catalog.Get(goalKey)
	if !ok {
		return ErrUnknownGoal
	}
	now := e.now()
	if !def.Active(now) {
		return nil // sezonal-цель неактивна — тихо пропускаем
	}
	fn, ok := snapshotByType(def.Condition.Type)
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownConditionType, def.Condition.Type)
	}
	period := PeriodKey(def.Lifecycle, now)

	return e.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		progress, err := fn(ctx, tx, userID, def.Condition)
		if err != nil {
			return err
		}
		return e.applyProgress(ctx, tx, userID, def, period, progress, false /*=increment*/)
	})
}

// OnEvent — counter-цели реагируют на event. Вызывается из worker
// через withGoal-обёртку (план 30 Ф.3).
//
// payload передаётся для будущих фильтров (по полям event payload),
// в MVP не используется.
func (e *Engine) OnEvent(ctx context.Context, tx pgx.Tx, userID string, eventKind int, payload []byte) error {
	now := e.now()
	keys := e.catalog.ByEventKind(eventKind)
	for _, key := range keys {
		def, _ := e.catalog.Get(key)
		if !def.Active(now) {
			continue
		}
		fn, ok := counterByType(def.Condition.Type)
		if !ok {
			continue
		}
		if !fn(eventKind, payload, def.Condition) {
			continue
		}
		period := PeriodKey(def.Lifecycle, now)
		// Прогресс читается+обновляется в той же транзакции, что и event.
		// Если не существует — INSERT progress=1; если есть — UPDATE +1.
		if err := e.incrementCounter(ctx, tx, userID, def, period); err != nil {
			return err
		}
	}
	return nil
}

// Claim — забрать награду за completed-цель.
// Возвращает выданное (для UI toast).
func (e *Engine) Claim(ctx context.Context, userID, goalKey, period string) (Reward, error) {
	def, ok := e.catalog.Get(goalKey)
	if !ok {
		return Reward{}, ErrUnknownGoal
	}
	// period может быть '' от клиента — для daily/weekly fix-up.
	if period == "" {
		period = PeriodKey(def.Lifecycle, e.now())
	}

	var reward Reward
	err := e.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var completedAt, claimedAt *time.Time
		err := tx.QueryRow(ctx, `
			SELECT completed_at, claimed_at FROM goal_progress
			WHERE user_id = $1 AND goal_key = $2 AND period_key = $3
			FOR UPDATE
		`, userID, goalKey, period).Scan(&completedAt, &claimedAt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotCompleted
			}
			return fmt.Errorf("claim select: %w", err)
		}
		if completedAt == nil {
			return ErrNotCompleted
		}
		if claimedAt != nil {
			return ErrAlreadyClaimed
		}
		// Mark claimed.
		if _, err := tx.Exec(ctx, `
			UPDATE goal_progress SET claimed_at = now()
			WHERE user_id = $1 AND goal_key = $2 AND period_key = $3
		`, userID, goalKey, period); err != nil {
			return fmt.Errorf("claim update: %w", err)
		}
		// Grant reward.
		if err := e.rewarder.Grant(ctx, tx, userID, def.Reward); err != nil {
			return fmt.Errorf("claim grant: %w", err)
		}
		// Audit log.
		rewardJSON, _ := marshalReward(def.Reward)
		if _, err := tx.Exec(ctx, `
			INSERT INTO goal_rewards_log (id, user_id, goal_key, period_key, reward)
			VALUES ($1, $2, $3, $4, $5)
		`, ids.New(), userID, goalKey, period, rewardJSON); err != nil {
			return fmt.Errorf("claim log: %w", err)
		}
		reward = def.Reward
		return nil
	})
	return reward, err
}

// MarkSeen — пометить toast как показанный (seen_at = now).
// Идемпотентно: повторный вызов не меняет seen_at.
func (e *Engine) MarkSeen(ctx context.Context, userID, goalKey, period string) error {
	def, ok := e.catalog.Get(goalKey)
	if !ok {
		return ErrUnknownGoal
	}
	if period == "" {
		period = PeriodKey(def.Lifecycle, e.now())
	}
	_, err := e.db.Pool().Exec(ctx, `
		UPDATE goal_progress SET seen_at = COALESCE(seen_at, now())
		WHERE user_id = $1 AND goal_key = $2 AND period_key = $3
	`, userID, goalKey, period)
	return err
}

// List — список целей с прогрессом для UI. Фильтр по category.
//
// Возвращает Views для всех целей в catalog (для категорий permanent/
// achievement даже без записи в goal_progress показываем как «закрытые»).
func (e *Engine) List(ctx context.Context, userID string, cat Category) ([]View, error) {
	now := e.now()
	keys := e.catalog.ByCategory(cat)

	// Один SELECT: все progress пользователя в категории.
	type progRow struct {
		key, period string
		progress    int
		completed   *time.Time
		claimed     *time.Time
		seen        *time.Time
	}
	progs := map[string]progRow{} // key by goal_key + period_key
	if len(keys) > 0 {
		rows, err := e.db.Pool().Query(ctx, `
			SELECT goal_key, period_key, progress, completed_at, claimed_at, seen_at
			FROM goal_progress
			WHERE user_id = $1 AND goal_key = ANY($2)
		`, userID, keys)
		if err != nil {
			return nil, fmt.Errorf("list query: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var pr progRow
			if err := rows.Scan(&pr.key, &pr.period, &pr.progress, &pr.completed, &pr.claimed, &pr.seen); err != nil {
				return nil, err
			}
			progs[pr.key+"|"+pr.period] = pr
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	out := make([]View, 0, len(keys))
	for _, key := range keys {
		def, _ := e.catalog.Get(key)
		if !def.Active(now) {
			continue
		}
		period := PeriodKey(def.Lifecycle, now)
		pr, ok := progs[key+"|"+period]
		v := View{GoalDef: def}
		if ok {
			v.Progress = pr.progress
			v.Completed = pr.completed != nil
			v.Claimed = pr.claimed != nil
			v.Seen = pr.seen != nil
			v.CompletedAt = pr.completed
		}
		out = append(out, v)
	}
	return out, nil
}

// applyProgress — обновить goal_progress по результату snapshot-evaluator'а.
//
// progress — сырое значение из условия (например, текущий уровень).
// Если progress >= target, цель completed (если ещё не была).
// При первой пометке completed — вызывается Notifier.
func (e *Engine) applyProgress(
	ctx context.Context, tx pgx.Tx,
	userID string, def GoalDef, period string, progress int, _ bool,
) error {
	target := def.EffectiveTarget()
	if progress > target {
		progress = target
	}
	completedNow := progress >= target

	var prevCompletedAt *time.Time
	err := tx.QueryRow(ctx, `
		SELECT completed_at FROM goal_progress
		WHERE user_id = $1 AND goal_key = $2 AND period_key = $3
		FOR UPDATE
	`, userID, def.Key, period).Scan(&prevCompletedAt)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		// Первая запись.
		var completedAt any
		if completedNow {
			completedAt = time.Now()
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO goal_progress (user_id, goal_key, period_key, progress, completed_at)
			VALUES ($1, $2, $3, $4, $5)
		`, userID, def.Key, period, progress, completedAt); err != nil {
			return fmt.Errorf("insert progress: %w", err)
		}
		if completedNow {
			return e.notifier.OnCompleted(ctx, tx, userID, def)
		}
		return nil
	case err != nil:
		return fmt.Errorf("read progress: %w", err)
	}

	// Запись существует. Если уже completed — ничего.
	if prevCompletedAt != nil {
		// Может потребоваться обновить progress (например снимочное условие
		// выросло), но completed/claimed не трогаем.
		_, err := tx.Exec(ctx, `
			UPDATE goal_progress SET progress = $1
			WHERE user_id = $2 AND goal_key = $3 AND period_key = $4
		`, progress, userID, def.Key, period)
		return err
	}
	// Не completed: обновляем progress и, если прошли порог, mark completed.
	if completedNow {
		if _, err := tx.Exec(ctx, `
			UPDATE goal_progress
			SET progress = $1, completed_at = now()
			WHERE user_id = $2 AND goal_key = $3 AND period_key = $4
		`, progress, userID, def.Key, period); err != nil {
			return fmt.Errorf("update completed: %w", err)
		}
		return e.notifier.OnCompleted(ctx, tx, userID, def)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE goal_progress SET progress = $1
		WHERE user_id = $2 AND goal_key = $3 AND period_key = $4
	`, progress, userID, def.Key, period); err != nil {
		return fmt.Errorf("update progress: %w", err)
	}
	return nil
}

// incrementCounter — +1 к counter-цели; если достигли target — mark completed.
// Должен вызываться внутри транзакции (которая уже обрабатывает event).
func (e *Engine) incrementCounter(ctx context.Context, tx pgx.Tx, userID string, def GoalDef, period string) error {
	target := def.EffectiveTarget()

	// UPSERT: вставка с progress=1, либо increment если уже есть и не
	// completed. completed_at выставляется в той же UPDATE через CASE.
	var newProgress int
	var completedAt *time.Time
	err := tx.QueryRow(ctx, `
		INSERT INTO goal_progress (user_id, goal_key, period_key, progress, completed_at)
		VALUES ($1, $2, $3, 1, CASE WHEN 1 >= $4 THEN now() ELSE NULL END)
		ON CONFLICT (user_id, goal_key, period_key) DO UPDATE
		SET progress = LEAST(goal_progress.progress + 1, $4),
		    completed_at = CASE
		        WHEN goal_progress.completed_at IS NOT NULL THEN goal_progress.completed_at
		        WHEN goal_progress.progress + 1 >= $4 THEN now()
		        ELSE NULL
		    END
		RETURNING progress, completed_at
	`, userID, def.Key, period, target).Scan(&newProgress, &completedAt)
	if err != nil {
		return fmt.Errorf("counter upsert: %w", err)
	}
	// Если только что стал completed — notify.
	// Проверка completedAt != nil + progress == target гарантирует, что
	// это первый момент перехода (UPDATE clause для уже-completed
	// сохраняет старое completed_at, мы не различим — но это
	// идемпотентно: notifier надо делать сам идемпотентным, либо
	// дублей не будет, потому что пользователь не получит двойной
	// inbox-message — INSERT уникальный по id).
	if completedAt != nil && newProgress == target {
		// Чтобы notifier вызывался один раз на момент перехода —
		// проверяем, что completedAt в пределах последней секунды
		// (грубо). В прод-коде безопаснее делать через returning
		// `xmax`/old_completed; для MVP достаточно.
		if time.Since(*completedAt) < time.Second {
			return e.notifier.OnCompleted(ctx, tx, userID, def)
		}
	}
	return nil
}

// marshalReward сериализует reward в JSONB для goal_rewards_log.
func marshalReward(r Reward) ([]byte, error) {
	return jsonMarshal(r)
}
