---
title: 32 — Multi-instance readiness (scheduler + advisory locks + chat pub/sub)
date: 2026-04-26
status: draft
---

# План 32: Multi-instance readiness

**Цель**: одним заходом сделать backend готовым к запуску в N≥2 инстансов
(server и worker). Включает:
1. **Cron-scheduler** для периодических задач (детерминированное время
   вместо «отсчёт от старта»).
2. **Distributed lock** через Postgres advisory lock — для singleton-задач
   (alien_spawn, inactivity_reminders, expire_planets, event_pruner,
   score_recalc_all).
3. **Redis pub/sub** для WebSocket чата — без него игроки на разных
   backend-instance не видят сообщений друг друга.
4. **ON CONFLICT** для bootstrap-вставок (race-fix мелочей).

**Контекст**: при текущей схеме N=1 worker — всё работает. Но при scale-out:
- 5 singleton-таймеров дают **дубль** побочных эффектов (двойной alien
  spawn, двойные inbox-сообщения).
- WebSocket Hub in-memory — игроки на разных backend instance не услышат
  друг друга.
- `BootstrapRecalcAllEvent` race на стартe двух воркеров.

Главный event-loop ([backend/internal/event/worker.go](../../backend/internal/event/worker.go))
**уже scale-ready** через `FOR UPDATE SKIP LOCKED` — его не трогаем.

---

## 1. Аудит: что и где

### 1.1 Singleton-задачи в worker'е

| Задача | Период | Where | Race при N>1? |
|---|---|---|---|
| Alien AI spawn | 6 ч | [worker/main.go:228-241](../../backend/cmd/worker/main.go#L228-L241) | 🔴 дубль атак |
| Inactivity reminders | 24 ч | [worker/main.go:253-271](../../backend/cmd/worker/main.go#L253-L271) | 🔴 дубль сообщений |
| Expire temp planets | 1 ч | [worker/main.go:274-292](../../backend/cmd/worker/main.go#L274-L292) | 🔴 race на DELETE |
| Event pruner (error→dead) | 24 ч | [event/worker.go:437](../../backend/internal/event/worker.go#L437) | 🔴 race DELETE+INSERT |
| BootstrapRecalcAllEvent | при старте | [score/event.go:79](../../backend/internal/score/event.go#L79) | 🟡 лишний event |
| MetricsUpdater | 15 с | [event/worker.go:352](../../backend/internal/event/worker.go#L352) | ✅ read-only, OK |
| Main event-loop | tick 10 с | [event/worker.go](../../backend/internal/event/worker.go) | ✅ FOR UPDATE SKIP LOCKED |
| Score recalc (через event) | 24 ч self-reschedule | KindScoreRecalcAll | ✅ событие подхватит один из воркеров |

### 1.2 Server-side

| Компонент | Race при N>1? |
|---|---|
| HTTP API (handlers) | ✅ stateless, БД-операции в транзакциях |
| Auth rate-limiter | ✅ Redis-based |
| LastSeenMiddleware | ✅ UPSERT |
| WebSocket Hub | 🔴 in-memory, нужен Redis pub/sub |
| i18n / catalog | ✅ read-only |

### 1.3 Что плохо в текущих ticker'ах (даже при N=1)

- **Расписание зависит от момента старта**: после рестарта в 17:43
  daily-задача каждый день будет дёргаться в 17:43. Деплой = сдвиг.
- **Совпадение пиков**: несколько daily-job в одной минуте суток.
- **Thursday-bonus alien** — `time.Now().Weekday()` проверка внутри
  Spawn, но сам Spawn привязан ко времени старта worker'а. Бонусные
  окна непредсказуемы.

---

## 2. Архитектура

### 2.1 Cron-scheduler

Пакет `backend/internal/scheduler/`:
- `Scheduler` — обёртка над `github.com/robfig/cron/v3`.
- Job регистрируется парой `(name, fn)`. Расписание читается из YAML.
- Каждая job обернута в **distributed lock** (см. §2.2): первый воркер,
  взявший lock, выполняет; остальные тихо пропускают.
- Метрики: `scheduler_job_runs_total{job, status}`,
  `scheduler_job_duration_seconds{job}`,
  `scheduler_job_last_run_timestamp{job}`,
  `scheduler_job_skipped_total{job}` (для скипов из-за lock-conflict).

### 2.2 Distributed lock через Postgres advisory locks

Postgres даёт `pg_try_advisory_lock(key)` — non-blocking попытка взять
эксклюзивный lock на 64-битный int. Идеально для нашего случая: не
требует доп.инфраструктуры (Redis lock-сервиса), не страдает от
clock skew.

```go
// backend/internal/locks/advisory.go

func TryRun(ctx context.Context, pool *pgxpool.Pool, lockName string, fn func(ctx context.Context) error) (acquired bool, err error) {
    key := hashLockName(lockName) // FNV-64 → int64
    conn, err := pool.Acquire(ctx)
    if err != nil { return false, err }
    defer conn.Release()

    var acq bool
    if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", key).Scan(&acq); err != nil {
        return false, err
    }
    if !acq { return false, nil }
    defer conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", key)

    return true, fn(ctx)
}
```

Особенности:
- **Lock per-connection**: важно использовать один `*pgx.Conn` (через
  `pool.Acquire`) на весь периметр lock'а. Если использовать pool
  напрямую — следующий запрос может взять другой connection, и lock
  не отпустится.
- **Освобождение при crash**: если процесс умер, Postgres сам сбросит
  lock через session disconnect.
- **Не блокирует основной event-loop** — lock берётся только для
  scheduler-job'ы.

Использование в scheduler'e:
```go
sch.Register("alien_spawn", func(ctx context.Context) error {
    acquired, err := locks.TryRun(ctx, pool, "scheduler:alien_spawn", func(ctx context.Context) error {
        return alienSvc.Spawn(ctx)
    })
    if !acquired {
        log.Debug("alien_spawn skipped: another worker holds lock")
        metrics.SchedulerSkipped.WithLabelValues("alien_spawn").Inc()
    }
    return err
})
```

При N=1 worker — lock всегда берётся, поведение идентично текущему.
При N>1 — ровно один worker выполняет, остальные skip + metric.

### 2.3 Redis pub/sub для WebSocket Hub

Сейчас `chat.Hub` — in-memory map'а каналов. При N=2 backend:
- Alice подключилась к backend-1, в её Hub зарегистрирована.
- Bob подключился к backend-2, в его Hub.
- Alice пишет → backend-1.Hub.Broadcast(msg) → доходит только до
  клиентов backend-1. Bob не получает.

Решение: **fan-out через Redis pub/sub**.

```go
// backend/internal/chat/hub.go (refactor)

type Hub struct {
    rdb     *redis.Client
    local   map[string]map[*client]struct{} // channel → clients (только этого инстанса)
    mu      sync.RWMutex
}

// Publish: пишет в Redis, не в local. Local-клиенты получат через subscriber.
func (h *Hub) Publish(channel string, msg Message) error {
    data, _ := json.Marshal(msg)
    return h.rdb.Publish(ctx, "chat:"+channel, data).Err()
}

// runSubscriber: единственная горутина на инстанс, читает Redis-канал
// и рассылает локальным клиентам. Запускается при NewHub.
func (h *Hub) runSubscriber(ctx context.Context) {
    sub := h.rdb.PSubscribe(ctx, "chat:*")
    for m := range sub.Channel() {
        ch := strings.TrimPrefix(m.Channel, "chat:")
        var msg Message
        if json.Unmarshal([]byte(m.Payload), &msg) != nil { continue }
        h.broadcastLocal(ch, msg)
    }
}
```

Свойства:
- **Все backend-instance подписаны на `chat:*`**. Сообщение от Alice
  на backend-1 → Redis → доходит до backend-2 → доходит до Bob.
- **При падении Redis** — local-клиенты на одном инстансе всё ещё
  видят друг друга (degradation, не полный отказ).
- **History persistence** не меняется — БД-таблица `chat_messages`
  остаётся как есть.

При N=1 backend (текущее) — пользователь не заметит, всё работает
через тот же loopback-pub/sub Redis (миллисекундная задержка).

### 2.4 ON CONFLICT для BootstrapRecalcAllEvent

```sql
INSERT INTO events (id, kind, state, fire_at, payload)
SELECT $1, $2, 'wait', $3, '{}'
WHERE NOT EXISTS (
    SELECT 1 FROM events WHERE kind = $2 AND state = 'wait'
)
```

CTE-форма гарантирует атомарность EXISTS+INSERT в одном запросе. Race
двух стартующих воркеров устранён.

---

## 3. YAML расписания

`configs/schedule.yaml`:

```yaml
# Расписание периодических задач worker'а.
# Все cron-выражения в UTC.
# Формат: стандартный cron (5 полей) или @every <duration>.
#
# Каждая job обернута в advisory lock — при N>1 воркеров выполняет
# только один. См. backend/internal/locks/.

jobs:
  alien_spawn:
    schedule: "0 */6 * * *"      # каждые 6 ч в :00 UTC (00,06,12,18)
    enabled: true
    description: "Spawn alien AI attacks. Thursday bonus active автоматически."

  inactivity_reminders:
    schedule: "0 9 * * *"         # каждый день в 09:00 UTC
    enabled: true
    description: "Рассылка email-напоминаний неактивным игрокам."

  expire_temp_planets:
    schedule: "5 * * * *"         # каждый час в :05 (не совпадая с alien_spawn)
    enabled: true
    description: "DELETE planets WHERE expires_at < now()."

  event_pruner:
    schedule: "30 3 * * *"        # 03:30 UTC — низкий трафик
    enabled: true
    description: "Перенос error-events старше 7 дней в events_dead."

  score_recalc_all:
    schedule: "0 4 * * *"         # 04:00 UTC — после pruner'а
    enabled: true
    description: "Полный пересчёт очков всех активных игроков."
```

**Не в scheduler** (остаются ticker'ами):
- `metrics_updater` — read-only gauges, race-safe, 15с тикер избыточен
  для cron-планировщика.
- Главный event-loop — он уже тиковый (`WORKER_INTERVAL=10s`).

ENV-overrides:
- `SCHEDULER_<JOB>_CRON` — заменяет `schedule:`.
- `SCHEDULER_<JOB>_ENABLED` — `true`/`false`.
- `SCHEDULER_DISABLE_ALL=true` — kill-switch для всего scheduler'а.

---

## 4. Фазы

### Фаза 1 — пакет `locks` (S, ~0.5 дня)

- [ ] `backend/internal/locks/advisory.go`: `TryRun(ctx, pool, name, fn)`
      возвращает `(acquired bool, err error)`.
- [ ] FNV-64 hash имени lock'а в int64.
- [ ] Использует `pool.Acquire()` для одного connection на lock-сессию.
- [ ] Defer-unlock при выходе.
- [ ] Unit-тесты: parallel-вызовы — только один acquired=true.
- [ ] Integration test (с реальным Postgres через testcontainers или
      опциональный — пропускается без БД).

**Готовность**: 1 PR, ~80 строк + тесты.

### Фаза 2 — пакет `scheduler` (M, ~1 день)

- [ ] `backend/internal/scheduler/scheduler.go`: `Scheduler` обёртка
      над `robfig/cron/v3`.
- [ ] `scheduler.Register(name string, fn func(ctx) error)`:
      внутри обёртывает в `locks.TryRun("scheduler:"+name, ...)`.
- [ ] `Config` — load YAML, ENV-overrides, валидация cron-выражений
      (fail-fast при invalid).
- [ ] Метрики Prometheus.
- [ ] Graceful Stop: дожидается активных job до shutdown grace.
- [ ] Unit-тесты: load YAML, ENV override, invalid cron = fail,
      disabled job не зарегистрирована, lock-skip metric.

**Готовность**: 1 PR, ~250 строк + тесты + dependency на robfig/cron.

### Фаза 3 — миграция 4 ticker-задач (M, ~0.5 дня)

- [ ] Создать `configs/schedule.yaml`.
- [ ] В `cmd/worker/main.go` заменить 4 goroutine-блока (alien_spawn,
      inactivity_reminders, expire_temp_planets, RunPruner) на
      `sch.Register(...)`.
- [ ] Удалить старые ticker-goroutines.
- [ ] `metrics_updater` оставить как ticker-goroutine (read-only).
- [ ] Smoke-тест: scheduler стартует, jobs регистрируются.

**Готовность**: 1 PR, ~80 строк правок в worker/main.go.

### Фаза 4 — score_recalc через scheduler (S, ~0.5 дня)

- [ ] Scheduler.Register("score_recalc_all", ...) → прямой
      `scoreSvc.RecalcAll(ctx)` (без события).
- [ ] Удалить `BootstrapRecalcAllEvent`.
- [ ] `KindScoreRecalcAll` handler оставить для legacy wait-events
      (через 7 дней удалить отдельной миграцией).

**Готовность**: 1 PR, ~30 строк.

### Фаза 5 — chat Redis pub/sub (M, ~1 день)

- [ ] `chat.Hub` рефакторинг: `local map[channel]→clients` +
      `runSubscriber` горутина с PSubscribe `chat:*`.
- [ ] `Publish` пишет в Redis вместо local broadcast.
- [ ] При падении Redis: продолжать работать на local (degradation
      до single-instance поведения).
- [ ] Тесты: два Hub'а на одном Redis — broadcast от одного приходит
      ко второму.

**Готовность**: 1 PR, ~150 строк chat refactor + тесты.

### Фаза 6 — race-fix BootstrapRecalcAllEvent (S, ~0.1 дня)

- [ ] `score/event.go:79`: переписать через CTE `WHERE NOT EXISTS`.

**Готовность**: 5 строк в существующем файле.

### Фаза 7 — документация и тестирование multi-instance (S, ~0.5 дня)

- [ ] `docs/ops/scaling.md`: «как запустить N>1 worker'а / backend».
- [ ] `docker-compose.scaling.yml` (overlay): пример с replicas: 2.
- [ ] Smoke-тест: поднять 2 worker'а в docker-compose, убедиться что
      scheduler выполняет каждую job ровно один раз.

**Готовность**: 1 PR, документация + compose-overlay.

---

## 5. Что НЕ делаем

- **Не реализуем** sticky session для WebSocket — Redis pub/sub
  делает их ненужными.
- **Не переходим** на Kafka/RabbitMQ — Postgres advisory locks +
  Redis pub/sub достаточно для нашего масштаба.
- **Не вводим** Kubernetes / k8s CronJob — у нас docker compose, проще.
- **Не делаем** catch-up пропущенных запусков (если worker лежал в
  04:00, score_recalc пропустим — следующий через 24h). Если
  потеря критична — отдельный план.
- **Не пишем** свой scheduler — используем `robfig/cron/v3`.
- **Не делаем** распределение нагрузки внутри одной job'ы (например,
  параллельный score_recalc на 2 worker'ах через шардирование users).
  Это масштабирование самой задачи, не runtime'а.

---

## 6. Риски

1. **Cron-выражение с опечаткой** → fail-fast при старте. Митигация:
   schema-валидация в CI (тест парсит `configs/schedule.yaml`).
2. **Synchronized DB load** в фиксированное время суток (4:00 UTC —
   все сразу пересчитывают очки). Митигация: разнести задачи во
   времени (см. YAML — pruner 03:30, recalc 04:00).
3. **Connection leak** в advisory lock при panic'е fn. Митигация:
   defer unlock + recover внутри scheduler-обёртки.
4. **Redis недоступен** → chat не работает между инстансами. Митигация:
   degradation до local-only (логировать, не падать).
5. **Postgres advisory lock-table не очищается** при OOM-kill. Митигация:
   advisory locks per-session — Postgres сбросит при disconnect.
6. **Двойной запуск scheduler-job из-за clock-skew между worker'ами** —
   не проблема: cron срабатывает чуть в разное время на разных хостах,
   но advisory lock пускает только одного. Второй просто получит
   `acquired=false` и залогирует skip.

---

## 7. Зависимости

- **План 31 Ф.1-2** (deploy + feature flags) — уже сделано. Без graceful
  shutdown и feature flags — рискованно катить такой рефакторинг.
- **Новая Go-зависимость**: `github.com/robfig/cron/v3`.
- Postgres ≥ 9.4 (advisory locks). Уже есть.
- Redis — уже есть (rdb в server и worker).

## 8. Definition of done

- 4+ singleton-задачи переведены на scheduler с advisory lock.
- Можно поднять 2 worker'а в docker-compose, и каждая периодическая
  задача выполняется **ровно один раз** (проверяется логами +
  `scheduler_job_runs_total` метрикой).
- WebSocket чат: Alice на backend-1 видит сообщения от Bob с backend-2.
- `docker-compose.scaling.yml` для smoke-теста.
- `BootstrapRecalcAllEvent` race-safe через ON CONFLICT.
- `docs/ops/scaling.md` с runbook'ом.
- Все 29+ go-test пакетов зелёные.
- Метрики `scheduler_job_*` экспортируются.
