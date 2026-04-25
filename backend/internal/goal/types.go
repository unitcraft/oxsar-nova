// Package goal — единый движок целей (achievements + daily quests + tutorial).
//
// Определения целей — в configs/goals.yml. БД хранит только
// goal_progress (per-user state) и goal_rewards_log (аудит).
//
// Архитектурно:
//   - Catalog (in-memory) — YAML загружается при старте.
//   - Engine — операции над goal_progress: Recompute, OnEvent, Claim,
//     MarkSeen, List.
//   - snapshotRegistry / counterRegistry — типы условий, регистрируются
//     init-функциями файлов в conditions/.
//   - Rewarder — атомарное награждение (credits + ресурсы).
//   - Notifier — toast/inbox-уведомление при completion.
//
// План 30 Ф.1.
package goal

import (
	"encoding/json"
	"errors"
	"time"
)

// Sentinel errors.
var (
	ErrUnknownGoal          = errors.New("goal: unknown goal key")
	ErrUnknownConditionType = errors.New("goal: unknown condition type")
	ErrNotCompleted         = errors.New("goal: not completed yet")
	ErrAlreadyClaimed       = errors.New("goal: already claimed")
)

// Category — высокоуровневая группа цели для UI.
type Category string

const (
	CategoryAchievement Category = "achievement"
	CategoryStarter     Category = "starter"
	CategoryDaily       Category = "daily"
	CategoryWeekly      Category = "weekly"
	CategoryEvent       Category = "event"
)

// Lifecycle описывает, как формируется period_key и как часто goal
// «перезагружается» для пользователя.
type Lifecycle string

const (
	LifecyclePermanent  Lifecycle = "permanent"   // period_key = ''
	LifecycleOneTime    Lifecycle = "one-time"    // period_key = '' (для starter-цепочек)
	LifecycleDaily      Lifecycle = "daily"       // period_key = 'YYYY-MM-DD' (UTC)
	LifecycleWeekly     Lifecycle = "weekly"      // period_key = 'YYYY-Www' (ISO week, UTC)
	LifecycleSeasonal   Lifecycle = "seasonal"    // period_key = '' + active_from/until
	LifecycleRepeatable Lifecycle = "repeatable"  // period_key = счётчик (зарезервировано)
)

// ConditionSpec — что проверяет цель. Тип — ключ в registry; Params —
// параметры конкретного типа (хранятся как map; функция registry
// маршалит их в свою struct через DecodeParams).
type ConditionSpec struct {
	Type   string         `yaml:"type"   json:"type"`
	Params map[string]any `yaml:"params" json:"params,omitempty"`
}

// DecodeParams — типобезопасное чтение Params в любую struct через JSON.
// Используется в registry-функциях:
//
//	var p BuildingLevelParams
//	if err := cond.DecodeParams(&p); err != nil { ... }
func (c ConditionSpec) DecodeParams(dst any) error {
	if c.Params == nil {
		return nil
	}
	data, err := json.Marshal(c.Params)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

// Reward — что выдаётся при claim. Все поля опциональны.
type Reward struct {
	Credits  int   `yaml:"credits,omitempty"  json:"credits,omitempty"`
	Metal    int64 `yaml:"metal,omitempty"    json:"metal,omitempty"`
	Silicon  int64 `yaml:"silicon,omitempty"  json:"silicon,omitempty"`
	Hydrogen int64 `yaml:"hydrogen,omitempty" json:"hydrogen,omitempty"`
}

// Empty — true, если все поля нулевые (нечего выдавать).
func (r Reward) Empty() bool {
	return r.Credits == 0 && r.Metal == 0 && r.Silicon == 0 && r.Hydrogen == 0
}

// GoalDef — определение цели из YAML. Immutable после Load.
type GoalDef struct {
	Key         string    `yaml:"-"           json:"key"`
	Title       string    `yaml:"title"       json:"title"`
	Description string    `yaml:"description" json:"description,omitempty"`
	Category    Category  `yaml:"category"    json:"category"`
	Lifecycle   Lifecycle `yaml:"lifecycle"   json:"lifecycle"`

	Condition ConditionSpec `yaml:"condition" json:"condition"`
	Target    int           `yaml:"target,omitempty" json:"target"` // default 1

	Reward Reward `yaml:"reward,omitempty" json:"reward,omitempty"`

	// Для daily/weekly random-pool. weight=0 — не выбирается случайно.
	RandomWeight int `yaml:"random_weight,omitempty" json:"random_weight,omitempty"`

	// Для seasonal-целей.
	ActiveFrom  *time.Time `yaml:"active_from,omitempty"  json:"active_from,omitempty"`
	ActiveUntil *time.Time `yaml:"active_until,omitempty" json:"active_until,omitempty"`

	// Граф зависимостей (tutorial-flow).
	Requires []string `yaml:"requires,omitempty" json:"requires,omitempty"`

	// UI-meta.
	Points    int    `yaml:"points,omitempty"     json:"points,omitempty"`
	Icon      string `yaml:"icon,omitempty"       json:"icon,omitempty"`
	SortOrder int    `yaml:"sort_order,omitempty" json:"sort_order,omitempty"`
}

// EffectiveTarget — Target с дефолтом 1 (для прогресс-баров).
func (g GoalDef) EffectiveTarget() int {
	if g.Target <= 0 {
		return 1
	}
	return g.Target
}

// Active — true если goal активна сейчас (для seasonal проверяет окно).
func (g GoalDef) Active(now time.Time) bool {
	if g.ActiveFrom != nil && now.Before(*g.ActiveFrom) {
		return false
	}
	if g.ActiveUntil != nil && now.After(*g.ActiveUntil) {
		return false
	}
	return true
}

// Progress — состояние одной цели одного пользователя на конкретный
// period (соответствует строке в goal_progress).
type Progress struct {
	UserID      string     `json:"user_id"`
	GoalKey     string     `json:"goal_key"`
	PeriodKey   string     `json:"period_key"`
	Progress    int        `json:"progress"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ClaimedAt   *time.Time `json:"claimed_at,omitempty"`
	SeenAt      *time.Time `json:"seen_at,omitempty"`
}

// View — combined view для UI: GoalDef + Progress + computed-поля.
type View struct {
	GoalDef
	Progress    int        `json:"progress"`
	Completed   bool       `json:"completed"`
	Claimed     bool       `json:"claimed"`
	Seen        bool       `json:"seen"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
