---
title: 30 — Единый Goal Engine (рефакторинг achievement + dailyquest)
date: 2026-04-26
status: in-progress
---

> **Статус 2026-04-26**:
> - Ф.1 (backend-engine + YAML-каталог + миграция БД) — ✅ done.
> - Ф.2 (HTTP API) — pending.
> - Ф.3 (worker hook) — pending.
> - Ф.4 (перенос данных) — pending.
> - Ф.5 (UI) — pending.
> - Ф.6-7 (расширения, удаление старого) — pending.

# План 30: Goal Engine — единый движок целей

**Цель**: заменить две отдельные системы (`achievement`, `dailyquest`) на
один backend-движок «целей» (goals) с гибкими параметрами. UI остаётся
двухэкранным, но за ним — общая логика.

**Контекст**: обе системы сейчас в раннем состоянии. Legacy oxsar2
реализован «ужасно» (по словам пользователя), но содержит несколько
хороших идей (граф зависимостей, custom requirements, разнообразные
награды). DailyQuest в legacy не было — это полностью наша фича.

**Зависит от плана 31** (Zero-downtime deploy + feature flags). Без
feature flags рефакторинг рискованный — каждая фаза должна катиться за
флагом, чтобы можно было откатить за 30 секунд (toggle + restart).

> **Решение по архитектуре данных (после обсуждения 2026-04-26)**:
> определения целей живут в **YAML** (`configs/goals.yml`), не в БД.
> БД хранит только пользовательский state (`goal_progress`,
> `goal_rewards_log`). Тип condition — ключ в Go-registry; параметры —
> поле в YAML. Это согласуется с архитектурой проекта (план 28 убрал
> данные из БД в YAML), даёт review через `git diff`, type safety при
> загрузке и единый паттерн с другим контентом (units.yml, defense.yml).

---

## 1. Что не так сейчас

### Achievement
- 15 SQL-checks **захардкожены в Go** (`CheckAll` в service.go) →
  новое достижение требует deploy.
- `CheckAll` запускается **после каждого event** (`withAchievement` в
  worker) → 15 SELECT-запросов на каждый build/attack/recycle/spy.
- Награды примитивны: только фиксированное число credits +
  message в inbox. Нет ресурсов, юнитов, артефактов.
- STARTER-достижения «висят» как обычные — нет UI tutorial-progress
  для онбординга.
- Нет toast/badge при unlock — игрок не узнаёт, пока не зайдёт на
  экран «Достижения».
- Был обнаружен баг (исправлен в ad7009e/08be516): SHIPYARD/LAB
  никогда не разблокировались. Симптом того, что **SQL-литералы в
  нескольких файлах легко рассинхронизируются**.

### DailyQuest
- Условия в БД через JSONB (хорошо), но **matcher-callback в Go-домене**
  (transport.go строит closure для условия `mission`) — дубль логики:
  каждый домен сам интерпретирует свой формат `condition_value`.
- Не интегрирован с i18n (русский хардкод).
- Только 9 quest-определений в seed, нет weekly/seasonal.
- Нет «вчерашних»/streak-механик.

### Общее
- Дублирующийся pattern «после event → побочное действие»
  (`withScore`, `withAchievement`, `withDailyQuest`). Растёт хаос.
- **Награждение раздроблено** — credits в users.credit, ресурсы на
  home-планету, артефакты в artefacts_user. Каждая система пишет
  свою логику зачисления.
- Нет единого лога «событий-наград» — игрок не может посмотреть, что
  получил вчера.

---

## 2. Архитектура: YAML-контент + минимум таблиц

**Решение (после обсуждения 2026-04-26)**: определения целей живут в
**YAML** (`configs/goals.yml`), как и весь остальной content проекта
(units, defense, rapidfire, requirements). БД хранит **только пользовательский
state**.

Это согласуется с архитектурой проекта (план 28 убрал данные из БД в
YAML), даёт review через `git diff`, type safety при загрузке (один
раз `yaml.Unmarshal` в типизированную struct), удобные массовые правки.

### YAML: configs/goals.yml

```yaml
goals:
  FIRST_METAL:
    title: "Первая шахта"          # i18n-ключ или прямой текст
    description: "Построить metal_mine."
    category: achievement           # achievement|starter|daily|weekly|event
    lifecycle: permanent            # permanent|one-time|daily|weekly|seasonal
    points: 1
    icon: "⛏️"
    condition:
      type: building_level          # ключ в Go-registry (см. §3)
      params: { unit_id: 1, min_level: 1 }
    target: 1                        # для прогресс-баров (default 1)
    reward:
      credits: 10
      metal: 1000

  STARTER_BUILD_SOLARPLANT:
    title: "Солнечная энергия"
    category: starter
    lifecycle: one-time
    points: 1
    icon: "☀️"
    condition:
      type: building_level
      params: { unit_id: 4, min_level: 1 }
    requires:                        # граф зависимостей (опц.)
      - STARTER_BUILD_METALMINE
    reward:
      credits: 15

  daily_send_attack:
    title: "Атаковать игрока"
    category: daily
    lifecycle: daily                 # period_key = 'YYYY-MM-DD' (UTC)
    icon: "⚔️"
    random_weight: 100               # для weighted-pick из daily-pool
    condition:
      type: event_count
      params: { event_kind: 10 }     # KindAttackSingle
    target: 1
    reward:
      credits: 25

  weekly_score_5000:
    title: "Набрать 5000 очков за неделю"
    category: weekly
    lifecycle: weekly                # period_key = 'YYYY-Www'
    condition:
      type: score_total
      params: { min: 5000 }
    target: 5000
    reward:
      credits: 200

  spring_event_2026:
    title: "Весенний турнир"
    category: event
    lifecycle: seasonal
    active_from: "2026-04-01T00:00:00Z"
    active_until: "2026-05-01T00:00:00Z"
    condition:
      type: battle_won
      params: { side: attackers, min_count: 5 }
    target: 5
    reward:
      credits: 500
      metal: 50000
```

**Правила валидации** при загрузке:
- `condition.type` должен быть зарегистрирован в Go-registry.
- `requires` — все ключи существуют, нет циклов.
- `lifecycle=daily/weekly` требует `category=daily/weekly` и обычно
  `random_weight > 0`.
- `lifecycle=seasonal` требует `active_from`/`active_until`.

### БД: только user-state

```
goal_progress (per-user state):
    user_id      UUID FK users(id) ON DELETE CASCADE
    goal_key     TEXT NOT NULL          -- ссылка на YAML по строковому ключу
    period_key   TEXT NOT NULL          -- '' permanent/one-time/seasonal
                                        -- 'YYYY-MM-DD' daily
                                        -- 'YYYY-Www' weekly
    progress     INT  DEFAULT 0
    completed_at TIMESTAMPTZ NULL
    claimed_at   TIMESTAMPTZ NULL       -- NULL = не взял награду
    seen_at      TIMESTAMPTZ NULL       -- NULL = не видел toast
    PK (user_id, goal_key, period_key)
    INDEX (user_id, period_key)        -- для List
    INDEX (user_id) WHERE completed AND NOT seen -- для badge

goal_rewards_log (опц., аудит-лог):
    id          UUID PK
    user_id     UUID FK users(id) ON DELETE CASCADE
    goal_key    TEXT
    period_key  TEXT
    granted_at  TIMESTAMPTZ DEFAULT now()
    reward      JSONB                  -- snapshot что выдано
```

**Изменения относительно первой версии**:
- ❌ нет `goal_defs` (всё в YAML).
- ❌ нет `goal_dependencies` (поле `requires` в YAML).
- ✅ `goal_key` — string-ссылка на YAML (а не FK на serial id).
- ✅ две таблицы вместо четырёх.

**Trade-off за string-key вместо FK**:
- Невозможно DB-уровневой проверкой отловить «прогресс на удалённую
  goal». Решение: при load YAML валидируем, что для всех
  `goal_progress.goal_key` есть запись в YAML (или маркируем сиротские
  как deprecated, но не удаляем — лог истории игрока сохраняется).

### Унификация
- **Achievement** = goal с `lifecycle='permanent'`, `period_key=''`.
- **Daily quest** = goal с `lifecycle='daily'`, `period_key='2026-04-26'`,
  `random_weight>0`.
- **Weekly quest** = goal с `lifecycle='weekly'`, `period_key='2026-W17'`.
- **Seasonal event** = goal с `lifecycle='seasonal'`, `active_from/until`.
- **Tutorial step** = goal с `category='starter'`, `lifecycle='one-time'`,
  + запись в `goal_dependencies` для линейного flow.

Один движок, одна таблица progress, одна логика claim.

---

## 3. Condition Engine: Go-registry + JSONB-параметры

Вместо «matcher-callback в каждом домене» (текущая dailyquest) или
«полный JSONB-evaluator» (хорошо для гибкости, но сложно типизировать) —
**типизированные Go-функции, зарегистрированные в реестре**, плюс
**параметры** условия в JSONB-поле БД.

### Принцип

```go
// backend/internal/goal/conditions/registry.go

type SnapshotCondition func(ctx context.Context, tx pgx.Tx, userID string, params json.RawMessage) (progress int, err error)
type CounterMatcher func(eventKind int, payload []byte, params json.RawMessage) bool

var snapshotRegistry = map[string]SnapshotCondition{}
var counterRegistry  = map[string]CounterMatcher{}

func RegisterSnapshot(typ string, fn SnapshotCondition) { snapshotRegistry[typ] = fn }
func RegisterCounter(typ string, fn CounterMatcher) { counterRegistry[typ] = fn }
```

Каждое условие — **отдельный файл** в `backend/internal/goal/conditions/`,
с типизированной структурой params:

```go
// conditions/building_level.go (snapshot-тип)

type BuildingLevelParams struct {
    UnitID   int `json:"unit_id"`
    MinLevel int `json:"min_level"`
}

func init() {
    goal.RegisterSnapshot("building_level", func(ctx, tx, userID, raw) (int, error) {
        var p BuildingLevelParams
        if err := json.Unmarshal(raw, &p); err != nil { return 0, err }
        var lvl int
        err := tx.QueryRow(ctx, `
            SELECT COALESCE(MAX(b.level), 0) FROM buildings b
            JOIN planets p ON p.id = b.planet_id
            WHERE p.user_id = $1 AND b.unit_id = $2
        `, userID, p.UnitID).Scan(&lvl)
        if err != nil { return 0, err }
        if lvl > p.MinLevel { return p.MinLevel, nil }
        return lvl, nil
    })
}
```

```go
// conditions/event_count.go (counter-тип)

type EventCountParams struct {
    EventKind int             `json:"event_kind"`
    Filter    json.RawMessage `json:"filter,omitempty"`  // опц. фильтр по payload
}

func init() {
    goal.RegisterCounter("event_count", func(eventKind int, payload []byte, raw json.RawMessage) bool {
        var p EventCountParams
        if err := json.Unmarshal(raw, &p); err != nil { return false }
        if eventKind != p.EventKind { return false }
        if p.Filter != nil {
            // matchPayloadFilter(payload, p.Filter) — небольшой helper для fields
            return matchPayloadFilter(payload, p.Filter)
        }
        return true
    })
}
```

### Хранение

Тип condition и параметры — **в YAML** (см. §2), а не в БД:
```yaml
condition:
  type: building_level
  params: { unit_id: 1, min_level: 1 }
```
При load YAML — `yaml.Unmarshal` парсит params в типизированную
`json.RawMessage`, которая на этапе `Recompute` десериализуется в
конкретный struct (`BuildingLevelParams` и т.п.) внутри функции
registry.

### Engine

```go
package goal

// Catalog — загруженный YAML-набор GoalDef. Immutable после Load.
type Catalog struct {
    byKey map[string]GoalDef
    // вспомогательные индексы:
    byCategory map[string][]string // {"daily": [...keys]}
    byEventKind map[int][]string   // counter-цели на event-kind
}

func LoadCatalog(path string) (*Catalog, error)

type Engine struct {
    catalog  *Catalog
    db       repo.Exec
    rewarder Rewarder
    notifier Notifier
}

// Recompute — пересчитать snapshot-цель.
// goalKey — string-ключ из YAML; period — '' для permanent.
// Идемпотентно: completed/claimed повторно не меняются.
func (e *Engine) Recompute(ctx, tx, userID, goalKey, period string) error {
    def, ok := e.catalog.byKey[goalKey]
    if !ok { return ErrUnknownGoal }
    fn, ok := snapshotRegistry[def.Condition.Type]
    if !ok { return ErrUnknownConditionType }
    progress, err := fn(ctx, tx, userID, def.Condition.Params)
    if err != nil { return err }
    return e.updateProgress(ctx, tx, userID, def, period, progress)
}

// OnEvent — counter-цели реагируют на event.
// Используется withGoal-обёрткой в worker.
func (e *Engine) OnEvent(ctx, tx, userID string, eventKind int, payload []byte) error {
    // Прямо в memory: какие counter-цели интересуются этим event_kind?
    keys := e.catalog.byEventKind[eventKind]
    for _, key := range keys {
        def := e.catalog.byKey[key]
        m := counterRegistry[def.Condition.Type]
        if !m(eventKind, payload, def.Condition.Params) { continue }
        // increment progress в goal_progress (period вычисляется по lifecycle).
        period := computePeriodKey(def.Lifecycle, time.Now().UTC())
        if err := e.incrementProgress(ctx, tx, userID, def, period, 1); err != nil {
            return err
        }
    }
    return nil
}

// Claim — забрать награду.
func (e *Engine) Claim(ctx, userID, goalKey, period string) (Reward, error)

// MarkSeen — пометить toast как показанный.
func (e *Engine) MarkSeen(ctx, userID, goalKey, period string) error

// List — для UI.
func (e *Engine) List(ctx, userID string, filter ListFilter) ([]GoalView, error)
```

**Преимущества доставания def из памяти**:
- 0 SQL-запросов на «загрузить определение цели» (vs `JOIN goal_progress
  + goal_defs`).
- `byEventKind`-индекс — `O(1)` вместо SELECT с JSONB-фильтром.

### Преимущества подхода

| Свойство | Текущий хардкод | JSONB-evaluator | **Go-registry** |
|---|---|---|---|
| Type safety params | ❌ литералы в SQL | ❌ runtime parsing | ✅ Go struct |
| Тестируемость | ⚠ интеграционные тесты | ⚠ нужна тестовая инфра | ✅ unit-тесты на каждое условие |
| Гибкость новых условий | ❌ изменение CheckAll | ✅ INSERT в БД | ✅ добавить файл + INSERT |
| Без deploy для нового goal? | ❌ | ✅ | ⚠ если новый condition-type — нужен deploy |
| Производительность snapshot | ✅ прямой SQL | ⚠ зависит от eval | ✅ прямой SQL |

Минус Go-registry: **новый тип** condition требует deploy. Но (см.
плана 31) deploy безболезненный, и **типы редко меняются** — обычно
добавляются новые goals с **существующими** типами.

### Список типов в MVP

**Snapshot**:
- `building_level` — есть здание уровня N.
- `research_level` — есть исследование уровня N.
- `planet_count` — N планет (с фильтром: всё/колонии/луны).
- `fleet_count` — N кораблей (по сумме на всех планетах).
- `score_total` — N очков.
- `artefact_count` — N артефактов в `state IN [...]`.

**Counter**:
- `event_count` — N событий заданного `event_kind` (с опц. фильтром
  по payload).
- `battle_won` — N побед в боях (через battle_reports).

**Композиция (опц., если нужно)**:
- `and` — все вложенные condition должны быть completed.
- `or` — любое.

В MVP вряд ли нужна композиция — большинство достижений
односложные. Если потребуется — добавим отдельным итерированием.

---

## 4. Hook-pattern: единый event-bus

Вместо `withScore`, `withAchievement`, `withDailyQuest` —
**один `withGoal`-hook**, который дёшево пушит event в очередь
обработки goals (или сразу инкрементирует counter, если goal активна).

```go
// В worker/main.go:

w.Register(event.KindBuildConstruction,
    withGoal(event.KindBuildConstruction)(
        withScore(event.HandleBuildConstruction)))
```

`withGoal` для каждого event:
1. Передаёт event-kind + payload в `goal.Engine.OnEvent(userID, eventKind, payload)`.
2. Engine находит **активные** goals с counter-условием на этот kind
   (один SELECT с JSONB-фильтром по `condition->>'event_kind'`).
3. Для каждой увеличивает progress; при достижении target — completed.
4. Snapshot-цели **не** триггерятся через event-bus (их пересчитывает
   GET /api/goals при заходе на UI или периодический job).

**Преимущество** vs текущего `withAchievement`:
- 1 SELECT (поиск активных counter-целей) вместо 15 проверок.
- Snapshot-цели проверяются **lazy** при GET (не на каждый event).

---

## 5. Reward — единая система награждения

```go
package goal

type Reward struct {
    Credits      int                `json:"credits,omitempty"`
    Metal        int64              `json:"metal,omitempty"`
    Silicon      int64              `json:"silicon,omitempty"`
    Hydrogen     int64              `json:"hydrogen,omitempty"`
    Units        []UnitReward       `json:"units,omitempty"`
    Buildings    []BuildingReward   `json:"buildings,omitempty"`
    Researches   []ResearchReward   `json:"researches,omitempty"`
    Artefacts    []string           `json:"artefacts,omitempty"`
}

type UnitReward struct {
    UnitID   int    `json:"unit_id"`
    Quantity int64  `json:"quantity"`
}

type Rewarder interface {
    Grant(ctx, tx, userID, planetID, reward Reward) error
}
```

`Grant` — атомарно (в одной транзакции):
- Кредиты: `UPDATE users SET credit = credit + N`.
- Ресурсы: на home-планету (или planetID если задан).
- Юниты: `INSERT INTO ships (planet_id, unit_id, count)`.
- Здания/исследования: `UPDATE buildings/research SET level = max(level, N)`.
- Артефакты: `INSERT INTO artefacts_user`.

**Все награды через один интерфейс** → достаточно его расширить, и любая
goal автоматически получит новый тип reward.

---

## 6. Notifier — toast / badge / inbox-message

```go
type Notifier interface {
    OnGoalCompleted(ctx, userID, goal Goal) error
}
```

Реализации:
- **inbox-message** (как сейчас в achievement.UnlockIfNew).
- **toast queue** — записывается в `goal_progress.seen_at = NULL` →
  frontend проверяет «новые с прошлого раза» при /api/goals.
- **badge counter** — `GET /api/goals/badge` возвращает
  `{ unread_count: N }` для UI-плашки в меню.

---

## 7. Tutorial-flow (использование graph)

С `goal_dependencies` появляется граф зависимостей. Для tutorial:

```
goal: TUTORIAL_BUILD_METALMINE
  category: starter
  condition: { type: building_level, params: { unit_id: 1, min_level: 1 } }
  reward: { metal: 1000, silicon: 1000 }

goal: TUTORIAL_BUILD_SILICONLAB
  category: starter
  depends_on: [TUTORIAL_BUILD_METALMINE]
  condition: { type: building_level, params: { unit_id: 2, min_level: 1 } }
```

UI вычисляет «следующий незавершённый tutorial-step» как первый
goal в `category=starter`, у которого все `depends_on` выполнены, но
сам он `not completed` → показывает на главном экране как «Следующий
шаг: построить Silicon Lab».

---

## 8. UI: оба экрана сохраняются, но через одну API

### `/api/goals` — единый endpoint

```
GET /api/goals?category=achievement
GET /api/goals?category=daily
GET /api/goals?category=starter

Response:
{
  goals: [
    {
      key, title, description, category, lifecycle,
      target_progress, progress,
      completed: bool, claimed: bool, seen: bool,
      reward: {...},
      depends_on: [...],
      unlocked_at: "2026-04-26T10:00Z" | null
    }
  ]
}
```

### UI-экраны
- **AchievementsScreen** — фильтрует `category in ('achievement', 'starter')`.
- **DailyQuestScreen** — `category='daily'`, фильтр `period_key=сегодня`.
- **TutorialPanel** (новый, на главном экране) — `category='starter'`,
  показывает следующий незавершённый.

### Notifications
- Toast при completion (через `seen_at IS NULL` flag).
- Badge на иконке меню «Достижения» / «Задания» — счётчик новых.

---

## 9. Миграция существующих данных

1. **Создать таблицы** `goal_progress`, `goal_rewards_log` (только две, без
   goal_defs/goal_dependencies — они в YAML).
2. **Создать `configs/goals.yml`** с записями, эквивалентными:
   - 15 `achievement_defs` (FIRST_METAL и т.п.) → `category=achievement`,
     `lifecycle=permanent`. Conditions вытащены из текущего
     `CheckAll` (Go-код).
   - 8 starter_defs (`STARTER_BUILD_*`) → `category=starter`,
     `lifecycle=one-time`. Уже исправленные unit_id (см. коммит
     08be516).
   - 9 daily_quest_defs (`send_transport`, `do_spy` и т.п.) →
     `category=daily`, `lifecycle=daily`. Condition format похож,
     просто переносим `condition_value.mission` в `params.event_kind`.
3. **Перенести `goal_progress`** (одна миграция-функция):
   - `achievements_user (user_id, achievement, unlocked_at)` →
     `goal_progress (user_id, goal_key=achievement, period_key='',
     completed_at=unlocked_at, claimed_at=unlocked_at, seen_at=
     unlocked_at)`.
   - `daily_quests (user_id, def_id, date, progress, completed_at,
     claimed_at)` → `goal_progress (user_id, goal_key=def.key,
     period_key=date, ...)`.
4. **Удалить старые таблицы** (`achievement_defs`, `achievements_user`,
   `daily_quest_defs`, `daily_quests`) через 1-2 deploy-цикла (когда
   уверены в стабильности).

**Этап 2 (миграция goal_progress)** — Go-tool в `cmd/tools/goals-migrate`,
запускается один раз вручную. Идемпотентен через `ON CONFLICT DO NOTHING`.

---

## 10. План внедрения по фазам

### Фаза 1: backend-engine + YAML-каталог + минимум миграций ✅ DONE

- [x] Migration `0065_goal_progress.sql`: таблицы `goal_progress`,
      `goal_rewards_log` (без goal_defs — определения в YAML).
- [x] Пакет `backend/internal/goal/`:
      - `types.go`: `GoalDef`, `Reward`, `View`, `ConditionSpec.DecodeParams`.
      - `catalog.go`: `LoadCatalog`/`ParseCatalog` с валидацией
        (категория, lifecycle, requires-граф, циклы), индексы
        `byCategory` и `byEventKind`.
      - `registry.go`: `snapshotRegistry`/`counterRegistry`,
        `RegisterSnapshot`/`RegisterCounter` с защитой от дублей.
      - `period.go`: `PeriodKey(lifecycle, now)` для daily/weekly UTC.
      - `engine.go`: `Engine.Recompute`, `OnEvent`, `Claim`,
        `MarkSeen`, `List`.
      - `rewarder.go`: `SimpleRewarder` (credits + ресурсы на home).
      - `notifier.go`: `InboxNotifier` + `NoopNotifier`.
- [x] Conditions:
      - `conditions/building_level.go` (snapshot, max-level через JOIN).
      - `conditions/event_count.go` (counter по `event_kind`).
- [x] `configs/goals.yml`: smoke-test с FIRST_METAL.
- [x] Feature flag `goal_engine = false` в `configs/features.yaml`.
- [x] Unit-тесты: catalog (8 кейсов: empty, sample, requires-valid/missing/cycle,
      invalid-category, seasonal-window, byEventKind, byCategorySorted),
      period (UTC normalization).

**Tests**: 29 пакетов зелёные.

**Что осталось до полной интеграции** (Ф.2-7):
- HTTP API (`GET /api/goals`, `POST /api/goals/{key}/claim`, `seen`).
- Worker hook `withGoal` вместо `withAchievement` + `withDailyQuest`.
- Перенос всех 32 целей из старых таблиц в YAML.
- Миграция `goal_progress` из существующих данных.
- UI переключить на `/api/goals`.
- Удаление старых пакетов achievement + dailyquest.

### Фаза 2: HTTP API

- [ ] `GET /api/goals` (с фильтром по category).
- [ ] `POST /api/goals/{key}/claim`.
- [ ] `POST /api/goals/{key}/seen` (mark toast).

### Фаза 3: Worker-hook

- [ ] Заменить `withAchievement` + `withDailyQuest` на единый
      `withGoal` (под флагом).
- [ ] Передавать event-kind + minimal payload в Engine.

### Фаза 4: миграция данных и контента

- [ ] Заполнить `configs/goals.yml` всеми существующими целями:
      15 achievement + 8 starter + 9 daily.
- [ ] Go-tool `cmd/tools/goals-migrate`: переносит `achievements_user`
      + `daily_quests` → `goal_progress`. Идемпотентен.

### Фаза 5: UI

- [ ] Сохранить два экрана, переключить на `/api/goals`.
- [ ] Добавить toast при unlock (через `seen_at`).
- [ ] Добавить badge counter в меню.
- [ ] Tutorial-panel на главном экране.

### Фаза 6: расширение

- [ ] Условия: `score_total`, `planet_count`, `fleet_count`,
      `artefact_acquired`, `battle_won`, `and`/`or`.
- [ ] Награды: юниты, здания, артефакты.
- [ ] Weekly quests (lifecycle='weekly', period_key из ISO-week).
- [ ] Seasonal events (active_from/active_until).
- [ ] Tutorial dependencies (граф через goal_dependencies).

### Фаза 7: удаление старого

- [ ] Удалить пакеты `backend/internal/achievement/` и `dailyquest/`.
- [ ] Удалить таблицы `achievement_defs`, `achievements_user`,
      `daily_quest_defs`, `daily_quests` (через 1-2 цикла deploy).

---

## 11. Что НЕ делаем

- **Не копируем** legacy-схему (`na_achievement_datasheet` со 100+
  столбцами `req_*`, `bonus_*`). Используем JSONB.
- **Не делаем** «states» (ALERT/HIDDEN/PROCESSED) в первой версии —
  достаточно `completed_at + claimed_at + seen_at`.
- **Не реализуем** custom-requirements через PHP-method-naming
  (legacy `checkReq{Name}()`). Все условия — declarative JSONB.
- **Не пишем** time-based validity (`time` поле в legacy для
  истечения условия). Если потребуется — `active_until` уже есть.
- **Не делаем** многоуровневые achievements в одной записи (`quantity`
  в legacy). Repeatable достижения — отдельные goal с
  `lifecycle='repeatable'` + `period_key='2026-04-26-1'` (счётчик).

---

## 12. Заимствованное из legacy (только хорошее)

| Идея | Как используем |
|---|---|
| Граф зависимостей (`na_requirements`) | `goal_dependencies` для tutorial-flow |
| Custom requirements (расширяемость) | Declarative conditions JSONB + новые типы добавляются в Evaluator |
| Награды разнообразным контентом (юниты/здания/арты) | `Reward` struct + `Rewarder` интерфейс |
| Repeatable через `quantity` | `lifecycle='repeatable'` + `period_key` |

## 13. Что **не берём** из legacy

| Анти-паттерн | Почему |
|---|---|
| 30+ полей `req_*` / `bonus_*` в datasheet | Schema rigidity. JSONB гибче и проще |
| `processAchievements()` после каждого event | Slow-path с 15 SELECT'ов. Заменяем counter+lazy |
| 4 состояния (ALERT/HIDDEN/PROCESSED/...) | Усложняет UX. completed/claimed/seen — достаточно |
| `custom_req_1`/`_2` строковые ключи + PHP `checkReq*()` | Code injection-style. Заменяем typed conditions |
| `granted_planet_id`/`bonus_planet_id` | Лишнее. Награда даётся на home-planet или указанную |
| Tutorial через jQuery-селекторы (`arrow_name`, `menu_div`) | Связан с конкретной DOM-структурой. UI-логика отдельно |

---

## 14. Связь с другими планами

- **План 17 D** (Daily Quests) — закрыт реализацией dailyquest. План 30
  фактически переписывает его.
- **План 18-25** — не пересекается, чисто backend-логика.
- **План 28** (configs simplification) — не пересекается, goals в БД, не YAML.
- **План 29** (magic numbers) — поможет: вместо `mission = 11` будет
  `event.KindSpy` в JSONB-condition.

---

## 15. Риски

1. **Объём работы** — это не 1 коммит. Минимум 4-5 PR'ов через 2 фазы
   (backend → UI). Делать **за feature flag** (план 31): новый код
   мёртв, пока flag=false. Включаем для тестирования на своём аккаунте,
   потом для всех.
2. **Migration данных** — нужно перенести existing achievements/quests
   игроков без потерь. Backup БД перед migration обязателен. Делать
   **shadow-режим** через два флага:
   - `goal_engine_writes`: пишет в новые таблицы параллельно со
     старыми → данные синхронизированы, можно сравнивать.
   - `goal_engine`: переключает чтение на новые таблицы.
3. **Производительность snapshot-checks** — composite indexes,
   проверить EXPLAIN на горячих запросах (`buildings(planet_id, unit_id)`,
   `research(user_id, unit_id)` уже должны быть).
4. **Backward compat HTTP** — старые клиенты могут ходить на
   `/api/achievements` / `/api/daily-quests`. Сохранить proxy-endpoints
   на 1-2 версии (proxy на `/api/goals?category=...`).
5. **Условия в Go-registry vs БД** — если добавляется новый тип
   condition (`battle_won`, `composite_and`), нужен deploy. План 31
   делает deploy безболезненным, поэтому это **не риск**.

---

## 16. Открытые вопросы

1. **Tutorial UI** — где показывать? Главный экран? Отдельная панель?
   Сайдбар?
2. **Daily quest reset time** — UTC midnight (как сейчас) или
   per-user timezone?
3. **Repeatable goals** — нужны ли в MVP? Или после weekly?
4. **Seasonal events** — нужен ли admin-UI для активации, или INSERT
   из миграции достаточно?
5. **i18n** — где хранить локализацию goal-текстов? В `goal_defs.title`
   как i18n-ключ + словарь, или прямо текст для каждого языка в JSONB?
