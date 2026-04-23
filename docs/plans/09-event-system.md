# План 09 — Event System: доработка и улучшение

## Контекст

Event-система — сердце игрового времени. Через неё идут постройки, исследования, верфь, полёты флота, бои, экспедиции, алиены, ракеты, артефакты, офицеры, ремонт, авто-сообщения, истечения и декораторы (achievements, score). 25 зарегистрированных Kind, ~19 активных handler'ов.

Данный план — результат аудита текущей реализации (апрель 2026). Система **работает** и **в целом безопасна** (транзакции, SKIP LOCKED, идемпотентность), но имеет системные риски для стадии, когда пользовательская база вырастет или вырастет нагрузка.

## Текущая реализация — якоря в коде

- [backend/cmd/worker/main.go](../../backend/cmd/worker/main.go) — bootstrap, регистрация handler'ов, декораторы withScore / withAchievement.
- [backend/internal/event/worker.go](../../backend/internal/event/worker.go) — `Worker.Run` / `tick` / `process`.
- [backend/internal/event/kinds.go](../../backend/internal/event/kinds.go) — Kind enum и `KindBatchProcessIntervalSecond = 10`.
- [migrations/0004_fleet_events.sql](../../migrations/0004_fleet_events.sql) — схема `events` + индекс `ix_events_fire_at` на `(fire_at) WHERE state='wait'`.

## Аудит: сильные стороны

1. **Атомарность**. `Worker.tick` оборачивает выборку batch + вызов всех handler'ов + финализацию state в одну транзакцию через `repo.InTx`. Откат консистентен.
2. **Конкурентность корректна**. `SELECT ... FOR UPDATE SKIP LOCKED` позволяет запустить N воркеров без риска двойной обработки.
3. **Идемпотентность handler'ов**. Каждый handler проверяет pre-condition (`level >= target`, `status = 'done'`, `state != 'outbound'`) и возвращает nil, если состояние уже применено. Это закреплено в docstring Kind.
4. **Понятная регистрация**. Явный `worker.Register(Kind, Handler)` в [cmd/worker/main.go:130-151](../../backend/cmd/worker/main.go#L130-L151), никакого скрытого авто-reflection.
5. **Декораторы**. `withScore` и `withAchievement` оборачивают handler'ы без модификации — чисто по принципу open/closed.
6. **Частичное покрытие e2e**. [fleet/attack_test.go](../../backend/internal/fleet/attack_test.go), [fleet/expedition_test.go](../../backend/internal/fleet/expedition_test.go) косвенно упражняют handler-логику.

## Аудит: проблемы

### Критические (данные / корректность)

- **C1. Весь batch в одной транзакции**. Если 99-й event из 100 упадёт с транзиентной ошибкой БД — транзакция может откатить все 100. Сейчас ошибка handler'а ловится и помечает state='error', но ошибка в `tx.Exec("UPDATE ... state='error'")` или `rows.Scan` откатит **всю** пачку ([worker.go:110-117](../../backend/internal/event/worker.go#L110-L117)). Это увеличивает blast radius по мере роста batch.
- **C2. Ошибка handler'а = финальный `state='error'`, никакого retry**. Сетевая флаки, deadlock, transient timeout — event мёртв навсегда. Для boot-критичных событий (ArriveHandler, ExpireArtefact) это риск потери игрового состояния. Docstring говорит «защита от шторма», но отсутствие даже одного retry со backoff перегибает в обратную сторону.
- **C3. Нет dead-letter observability**. События в `state='error'` никто не мониторит — ни алерт, ни admin-UI, ни ручной replay. Они просто копятся.
- **C4. `ErrSkip` зарезервирован, но не используется**. Кейс «зависимость ещё не готова» (постройка требует ресурсов, которых не хватило из-за гонки с атакой) решается через ошибку — event умирает вместо переноса.

### Масштабируемость

- **S1. Один воркер-процесс — SPOF**. `make worker-run` запускает один instance. Падение → вся игра замораживается на время restart. SKIP LOCKED позволяет запуск N штук, но ни compose, ни деплой это не делают.
- **S2. Batch=100 и interval=10s жёстко зашиты**. При 1000+ готовых events обработка отстаёт от реального времени (макс ~600 events/мин на воркер). Должно быть конфигурируемо + адаптивное (если batch полный — не ждать 10с, тикать сразу).
- **S3. Индекс `ix_events_fire_at` на `(fire_at) WHERE state='wait'`** — хорош до ~100k waiting events, при большем объёме `state` discrimination через WHERE теряет эффективность на bloat (dead tuples от updated rows). Нужен `partial + include(kind)` для planner и регулярный `VACUUM (INDEX_CLEANUP)`.
- **S4. Нет шардирования**. Все воркеры конкурируют за весь pool events. Для географически разнесённого деплоя или выделенных shard'ов (например, battle-heavy vs. economy) нужна фильтрация по `hash(planet_id) % N` или по Kind-category.
- **S5. jsonb payload без схем**. В payload напихано всё что угодно (TransportPayload, BuildingPayload, ArtefactPayload, …). Нет версионирования, нет JSON Schema. При миграциях формата — риск протухших wait-events с несовместимым форматом.

### Наблюдаемость

- **O1. Нет `trace_id` в events**. Нельзя cross-reference action пользователя (POST /fleet/send) с будущим handler (arrive в 2 часа). Slog в handler'е не наследует request context.
- **O2. Нет метрик**. Ни `events_processed_total`, ни `events_failed_total{kind=…}`, ни latency гистограмм. Невозможно понять нагрузку, spike'и, kind-mix без SQL в БД.
- **O3. Нет алертов**. `state='error'` count растёт — никто не узнает. Lag `max(now - fire_at) WHERE state='wait'` может расти до часов — никто не узнает.
- **O4. Нет admin-UI для events**. Админ не может: посмотреть очередь, переопубликовать error, отменить wait, увидеть гистограмму kind'ов.

### Тестовое покрытие

- **T1. `event/` — ноль unit-тестов**. Нет теста на: повторный Register = panic, tick-с-ошибкой, handler-с-ошибкой = state='error', batch с concurrent worker'ом.
- **T2. Нет property-теста идемпотентности**. Заявлено в ТЗ §6.1 и §18.3, но не проверяется на фаззинге.
- **T3. Нет golden-теста состояния events** (snapshot через итерацию tick).

### Частные gaps (из simplifications.md)

- **G1. `KindExpirePlanet` отсутствует**. `planets.expires_at` выставляется в экспедиции (extra_planet), но удалять никто не удаляет. Утечка записей.
- **G2. `credit_purchases` таблицы нет**. `buyCredit=0` всегда — формула экспедиции теряет смысл при монетизации.
- **G3. Welcome-automsg вне транзакции регистрации** (не event-system, но связанный race).

## Оценка по 10-балльной шкале (текущее состояние)

| Ось | Балл | Комментарий |
|---|---|---|
| Качество | 6/10 | Идемпотентность и транзакции есть; тестов, trace_id, retry — нет. |
| Масштабируемость | 5/10 | SKIP LOCKED готов, но SPOF, фиксированный batch, нет шардирования. |
| Стабильность | 7/10 | Транзакции атомарны; но ошибки финальны и невидимы — тихие потери. |

## План доработок

Пять фаз, от самого ценного к наименее. Каждая фаза — отдельный PR (≤ 400 LOC, feature-flag где уместно). Закрытые пункты переносить в «Закрытые» как обычно.

### Фаза 1 — Надёжность (H, ~1 спринт)

1. **Per-event транзакция вместо batch-транзакции.** `tick` выбирает batch в read-only snapshot, но каждый handler вызывается в отдельном `InTx`. Ошибка одного event не откатывает остальные. Финализация state ('ok'/'error') — в той же per-event транзакции.
   - Риск: `SELECT FOR UPDATE SKIP LOCKED` в отдельной транзакции на каждый event. Нужно выбрать ID's, затем цикл: `BEGIN; SELECT ... WHERE id=$1 FOR UPDATE; handler; UPDATE state; COMMIT`.
   - Альтернатива: savepoint'ы внутри batch-транзакции — проще, но не освобождает lock на другие events при долгом handler'е.
2. **Retry с backoff.** Добавить `attempt int` и `next_retry_at timestamptz` колонки. При ошибке handler'а: если `attempt < maxAttempts (default 3)` — `state='wait'`, `fire_at = now() + backoff(attempt)` (10s → 60s → 300s), `attempt+=1`, `last_error text`. Только на `attempt >= max` → `state='error'`.
   - Ортогонально `ErrSkip`: `ErrSkip` не инкрементирует attempt, просто сдвигает `fire_at`.
   - Whitelist kind'ов, которые **не** ретраятся (те, где побочки дороже дубликата — но таких не должно быть, т.к. handler'ы идемпотентны).
3. **Dead-letter таблица.** `events_dead` с той же схемой + `last_error text`, `failed_at timestamptz`. Cron moves `state='error' AND processed_at < now() - interval '7 days'` туда. Оригинальная `events` не пухнет.
4. **Admin endpoint для replay.** `POST /admin/events/{id}/retry` → `UPDATE state='wait', attempt=0, fire_at=now()`. `POST /admin/events/{id}/cancel` → soft-delete. AdminOnly middleware уже есть.

### Фаза 2 — Наблюдаемость (H, ~0.5 спринта)

1. **`trace_id` в events.** `events.trace_id uuid nullable`. API-handler'ы при вставке event проставляют trace_id из request context (уже есть в slog middleware?). Handler логирует trace_id — связь «пользователь нажал → worker обработал» готова.
2. **Prometheus-метрики.** `pkg/metrics` (если нет — добавить `github.com/prometheus/client_golang`). Экспорт на `:9090/metrics` в worker process:
   - `events_processed_total{kind,state}` counter
   - `event_handler_duration_seconds{kind}` histogram
   - `events_queue_depth{state}` gauge (обновляется в tick через COUNT)
   - `events_lag_seconds` gauge (`now - min(fire_at) WHERE state='wait' AND fire_at <= now()`)
3. **Slog-поля.** В handler'ах вместо голых сообщений — structured: `kind`, `event_id`, `user_id`, `planet_id`, `attempt`, `trace_id`.
4. **Grafana dashboard JSON.** Положить в `deploy/grafana/event-system.json`. Panels: QPS by kind, error rate by kind, p50/p95/p99 handler duration, queue depth, lag.
5. **Admin-UI «Events monitor».** Новая вкладка в AdminScreen: queue depth by state, top-10 kind'ов по ошибкам (за 24h), список последних error-events с кнопкой replay.
6. **Rate-limit на admin replay.** `POST /admin/events/{id}/retry` — не более 10 req/min per admin (Redis sliding-window, уже есть для auth). Для массового replay — отдельный `POST /admin/events/replay` с body `{kind, max_count}` и жёстким лимитом `max_count ≤ 1000` за вызов. Защита от кнопки-storm.

### Фаза 3 — Масштабируемость (M, ~1 спринт)

Порядок подпунктов важен: **3.0 → 3.1 → 3.2 → 3.3 → 3.4 → 3.5 → 3.6 → 3.7**. Запуск ≥2 воркеров (3.4) ДО реализации шардирования (3.3) приведёт к deadlock'ам на ACS — см. раздел «Порядок обработки».

1. **3.0. Tie-breaker в `ORDER BY`.** Сейчас `ORDER BY fire_at` — при одинаковых `fire_at` порядок недетерминирован. Изменить на `ORDER BY fire_at, id`. Дешёво, защищает от плавающего поведения при bursts и clock skew между нодами.
2. **3.1. Конфигурируемые `batch` и `interval`.** Через ENV: `WORKER_BATCH=100`, `WORKER_INTERVAL_SEC=10`, `WORKER_MAX_BATCH=1000` (для burst-дренажа).
3. **3.2. Адаптивный tick.** Если `tick` обработал полный batch — не ждать следующий Ticker, запустить следующий tick сразу (до пустого batch или `MAX_BATCH` в итерации). Снимает SLA-lag в пиках. Безопасно при 1 воркере.
4. **3.3. Шардирование по `user_id`.** Добавить ENV `WORKER_SHARD_ID=0`, `WORKER_SHARD_COUNT=1`. В `tick`: `WHERE (user_id IS NULL AND $shard_id = 0) OR hashtext(user_id::text) % $shard_count = $shard_id`. Для ACS: события группы должны шардироваться по общему ключу — добавить колонку `shard_key text nullable` (по умолчанию `user_id::text`, для ACS = `group_id::text`). Handler'ы при создании event'ов ACS проставляют её явно.
5. **3.4. Deploy ≥2 воркеров.** Обновить `deploy/docker-compose.prod.yml` — worker replicas=N с разными `WORKER_SHARD_ID`. Добавить секцию в `docs/deployment.md`. **Только после 3.3.**
6. **3.5. Graceful shutdown.** Текущий `Run` завершается по ctx.Done(). Добавить: в `tick` проверка `ctx.Err()` между events, чтобы не начинать новый handler при SIGTERM. В `main.go` — signal handler с 30s grace.
7. **3.6. Schema для payload.** Тип-safe wrapper: `pkg/event/payload/` с версионированными структурами `BuildingPayloadV1`, `TransportPayloadV1`. Handler'ы разбирают через typed parser с version-check. Миграции payload — явные с версиями.
8. **3.7. Partitioning `events` по `fire_at`.** При росте до миллионов wait-events VACUUM и индекс становятся узким местом. PG native partitioning по месяцу `fire_at` (`CREATE TABLE events PARTITION BY RANGE (fire_at)`). Drop старой партиции = мгновенный cleanup ok-events без VACUUM FULL. Поддерживается в PG 14+.
9. **3.8. Autovacuum tuning + HOT updates.** Таблица с постоянным UPDATE (wait→ok) пухнет. Настроить per-table: `ALTER TABLE events SET (autovacuum_vacuum_scale_factor = 0.05, autovacuum_vacuum_cost_delay = 2)`. Проверить, что UPDATE не трогает индексированные колонки (`state`, `fire_at`) в happy path — если меняется только `state`, это не HOT; если `fire_at` меняется только при retry — тоже ок. Документировать, что `fill_factor = 80` помогает HOT.

### Фаза 4 — Тесты (M, ~0.5 спринта)

1. **`event/worker_test.go`** с testcontainers (PostgreSQL):
   - tick с 0 events → no-op.
   - tick с wait-event, handler=nil → state='error'.
   - tick с wait-event, handler=err → retry scheduled (фаза 1).
   - Два параллельных воркера, 100 events → все обработаны ровно один раз, нет дубликатов.
   - Tie-breaker: 10 events с одинаковым `fire_at` и разными `id` — порядок обработки строго по `id`.
2. **Deadlock-тест ACS.** Создать ACS-группу с 3 флотами → 3 события с одинаковым `fire_at` → запустить 2 воркера без шардирования → зафиксировать deadlock (ожидаемое поведение **до** 3.3). Повторить **с** шардированием по `shard_key=group_id` → все 3 события обрабатывает один воркер, deadlock'а нет.
3. **Race-тест конфликтующих событий одного юзера.** `BuildConstruction` и `AttackSingle` на одной планете с одинаковым `fire_at` → при sequential-режиме результат определяется `ORDER BY fire_at, id`. Проверить: повтор теста 100× даёт один и тот же финальный state планеты.
4. **Property-тест идемпотентности.** Для каждого kind: применить handler дважды на одно event → состояние совпадает. Использовать `rapid` (уже в зависимостях для battle).
5. **Snapshot-тест декораторов.** `withScore` / `withAchievement` вызывают декорированное строго после успешного handler'а, не при ошибке.

### Фаза 5 — Закрытие gap'ов (L-M, ~0.5 спринта)

1. **`KindExpirePlanet`.** При создании extra_planet в expedition — сразу планируется event на `expires_at`. Handler: `DELETE FROM planets WHERE id=$1 AND expires_at IS NOT NULL AND expires_at <= now()`. Авто-сообщение владельцу (folder=13).
2. **`credit_purchases` таблица.** Под план 07 (payments), но event-ручка: `KindCreditRefund` (на случай отмены платежа) — чистый event с идемпотентным refund.
3. **Welcome-send внутри регистрации.** Refactor `auth.Register` в одну `InTx`. automsg.Send принимает `tx` (уже поддерживается).
4. **Score RecalcAll: убрать 5-минутный ticker, заменить на batch-SQL раз в сутки.** Гибридная схема (см. отдельный раздел ниже).

## Оптимизация: Score RecalcAll

### Проблема

Текущая [score/service.go:RecalcAll](../../backend/internal/score/service.go#L402-L430) итерирует всех активных юзеров и на каждого делает 5 запросов (`calcBuildings`/`calcResearch`/`calcUnits`/`calcAchievements` + UPDATE). На 10k юзеров это ~50 000 round-trip'ов, плюс Go-side цикл `for lvl := 1; lvl <= level; lvl++` для cost-accumulation. Bottleneck #1 из [docs/vps-sizing.md](../vps-sizing.md).

При этом `withScore` decorator ([cmd/worker/main.go](../../backend/cmd/worker/main.go)) **уже** обновляет очки incrementally после build/research/fleet/defense events — 5-минутный RecalcAll дублирует эту работу.

### Решение: гибридная схема

**1. Incremental (основной механизм, уже работает)**:
- `withScore` на KindBuildConstruction, KindResearch, KindBuildFleet, KindBuildDefense — вызывает `RecalcUser(userID)` после каждого успешного handler'а.
- Оптимизация одного RecalcUser: заменить цикл `for lvl := 1; lvl <= level; lvl++` формулой геометрической прогрессии `cost_base × (factor^level − 1) / (factor − 1)` (при factor ≠ 1) или `cost_base × level` (при factor = 1). O(1) вместо O(level). Ускоряет и incremental-путь, и любое будущее использование RecalcUser.

**2. Batch-сверка раз в сутки (ночной cron)**:
- Один SQL-запрос через CTE обновляет очки **всех** юзеров за ~секунды на 10k юзеров.
- Запускается в низкий по трафику час (например, 4:00 UTC).
- Новый `KindScoreRecalcAll` event с `fire_at = next_midnight_utc + 4h` — handler выполняет один UPDATE, планирует следующий на +24h.

**3. Удалить 5-минутный ticker полностью**. Оставить ручной `/admin/score/recalc` для on-demand сверки.

### SQL для batch-сверки

```sql
WITH b AS (
  SELECT p.user_id,
         SUM(bs.cost_base_sum *
             CASE WHEN bs.cost_factor = 1
                  THEN b.level
                  ELSE (POWER(bs.cost_factor, b.level) - 1) / (bs.cost_factor - 1)
             END) AS bp
  FROM buildings b
  JOIN planets p ON p.id = b.planet_id AND p.destroyed_at IS NULL
  JOIN building_specs bs ON bs.unit_id = b.unit_id
  GROUP BY p.user_id
),
r AS (
  SELECT r.user_id,
         SUM(rs.cost_base_sum *
             CASE WHEN rs.cost_factor = 1
                  THEN r.level
                  ELSE (POWER(rs.cost_factor, r.level) - 1) / (rs.cost_factor - 1)
             END) AS rp
  FROM research r
  JOIN research_specs rs ON rs.unit_id = r.unit_id
  GROUP BY r.user_id
),
s AS (
  SELECT p.user_id, SUM(sh.count * ss.cost_sum) AS sp
  FROM ships sh
  JOIN planets p ON p.id = sh.planet_id AND p.destroyed_at IS NULL
  JOIN ship_specs ss ON ss.unit_id = sh.unit_id
  WHERE sh.count > 0
  GROUP BY p.user_id
),
d AS (
  SELECT p.user_id, SUM(def.count * ds.cost_sum) AS dp
  FROM defense def
  JOIN planets p ON p.id = def.planet_id AND p.destroyed_at IS NULL
  JOIN def_specs ds ON ds.unit_id = def.unit_id
  WHERE def.count > 0
  GROUP BY p.user_id
),
a AS (
  SELECT user_id, SUM(points) AS ap
  FROM achievements
  GROUP BY user_id
)
UPDATE users u SET
  b_points = ROUND(COALESCE(b.bp, 0) * $1, 2),
  r_points = ROUND(COALESCE(r.rp, 0) * $2, 2),
  u_points = ROUND((COALESCE(s.sp, 0) + COALESCE(d.dp, 0)) * $3, 2),
  a_points = ROUND(COALESCE(a.ap, 0), 2),
  points   = ROUND(
    COALESCE(b.bp, 0) * $1 +
    COALESCE(r.rp, 0) * $2 +
    (COALESCE(s.sp, 0) + COALESCE(d.dp, 0)) * $3, 2)
FROM (SELECT id FROM users WHERE umode = false) usrs
LEFT JOIN b ON b.user_id = usrs.id
LEFT JOIN r ON r.user_id = usrs.id
LEFT JOIN s ON s.user_id = usrs.id
LEFT JOIN d ON d.user_id = usrs.id
LEFT JOIN a ON a.user_id = usrs.id
WHERE u.id = usrs.id;
```

`$1=kBld`, `$2=kRes`, `$3=kUnt` — коэффициенты из конфига.

### Миграция `*_specs` таблиц

Cost-данные живут в YAML ([configs/construction.yml](../../configs/construction.yml), [configs/research.yml](../../configs/research.yml), [configs/ships.yml](../../configs/ships.yml), [configs/defense.yml](../../configs/defense.yml)). В SQL нужны таблицы с `cost_base_sum` и `cost_factor`:

```sql
CREATE TABLE building_specs (
  unit_id int PRIMARY KEY,
  cost_base_sum numeric NOT NULL,  -- metal + silicon + hydrogen
  cost_factor numeric NOT NULL
);
-- + research_specs, ship_specs, def_specs (для ship/def — только cost_sum без factor)
```

Заполнение при старте: backend читает YAML, делает `INSERT ... ON CONFLICT (unit_id) DO UPDATE`. Идемпотентно, никакого drift'а между YAML и БД. Либо отдельный CLI `cmd/tools/sync-specs` — вызывается при деплое после миграций.

### Acceptance criteria

- [ ] В [score/service.go](../../backend/internal/score/service.go) добавлен `RecalcAllBatch(ctx) error` — один SQL вместо цикла.
- [ ] Старый `RecalcAll(ctx, log)` помечен deprecated и больше не вызывается из worker.
- [ ] 5-минутный ticker в [cmd/worker/main.go](../../backend/cmd/worker/main.go) удалён.
- [ ] Добавлен `KindScoreRecalcAll` с handler'ом, перепланирующим следующий запуск на +24h.
- [ ] Миграция создаёт `*_specs` таблицы; bootstrap backend'а делает upsert из каталога.
- [ ] `calcBuildings` / `calcResearch` в Go переписаны на формулу геометрической прогрессии (O(1) вместо O(level)).
- [ ] Тест: RecalcUser и RecalcAllBatch на одном dataset дают идентичные значения (±0.01 на округление) для 100 юзеров.
- [ ] Тест: RecalcAllBatch на 1000 юзеров укладывается в 1s на локальном PG.

## Приоритеты и последовательность

Рекомендуемый порядок: **1 → 2 → 4 → 3 → 5**. Надёжность + наблюдаемость закрывают 80% реальных рисков. Тесты страхуют фазы 3 и 5. Масштабируемость (3) важна только при росте DAU; пока 1+2+4 достаточно для текущей стадии.

Внутри фазы 3 обязателен порядок **3.0 → 3.1 → 3.2 → 3.3 → 3.4 → 3.5 → 3.6 → 3.7 → 3.8**. Запуск ≥2 воркеров (3.4) **до** шардирования (3.3) = deadlock на ACS. Partitioning (3.7) имеет смысл только при >1M events в таблице — можно отложить.

## Definition of Done по фазам

Фаза считается закрытой, когда выполнены **все** критерии её DoD. Закрытые пункты переносятся в «Закрытые» в [simplifications.md](../simplifications.md).

**Фаза 1 — Надёжность**
- Транзиентная ошибка одного event не откатывает batch (per-event tx).
- Transient error → retry через backoff, `attempt` инкрементится; после `maxAttempts` → `state='error'`.
- Таблица `events_dead` создана; cron перекладывает туда error-events старше 7 дней.
- Admin endpoint'ы `/admin/events/{id}/retry` и `/cancel` работают, проверены AdminOnly middleware.

**Фаза 2 — Наблюдаемость**
- `trace_id` проставляется в каждое новое event при создании, логируется в slog handler'а.
- Метрики `events_processed_total`, `events_failed_total`, `event_handler_duration_seconds`, `events_queue_depth`, `events_lag_seconds` экспортируются на `:9090/metrics`.
- Grafana-дашборд `deploy/grafana/event-system.json` импортируется и показывает данные.
- Admin-UI «Events monitor» показывает queue depth, error-top, кнопку replay.
- Rate-limit на replay endpoint (10 req/min + batch max_count=1000).

**Фаза 3 — Масштабируемость**
- ENV-переменные `WORKER_BATCH`, `WORKER_INTERVAL_SEC`, `WORKER_MAX_BATCH`, `WORKER_SHARD_ID`, `WORKER_SHARD_COUNT` работают.
- `ORDER BY fire_at, id` в SELECT воркера.
- Адаптивный tick не ждёт Ticker при полном batch.
- `shard_key` колонка в events; ACS-события проставляют `group_id`.
- N воркеров с разными `SHARD_ID` работают в prod-compose; deadlock-тест (фаза 4.2) проходит.
- Graceful shutdown с 30s timeout; SIGTERM не рвёт handler посередине.
- Typed payload wrappers с версиями; handler'ы отклоняют неизвестную версию с понятной ошибкой.
- Autovacuum tuning применён к `events`; partitioning — опционально при >1M events.

**Фаза 4 — Тесты**
- `go test ./backend/internal/event/...` покрывает tick, retry, deadlock ACS, race-тест конфликтующих событий, идемпотентность.
- CI зелёный, тесты не флаки (100 прогонов подряд).

**Фаза 5 — Закрытие gap'ов**
- Планеты с `expires_at <= now()` удаляются через `KindExpirePlanet`.
- `credit_purchases` таблица создана; `KindCreditRefund` handler реализован.
- `auth.Register` выполняет welcome-send внутри одной `InTx`.

## Сколько воркеров запускать

План 09 делает количество воркеров конфигурируемым (`WORKER_SHARD_COUNT`, `WORKER_SHARD_ID`), но практический потолок определяется не воркерами, а **PostgreSQL как single-writer по доменным таблицам**. Воркеры читают очередь параллельно через SKIP LOCKED, но UPDATE'ы в `planets`/`fleets`/`ships`/`users` упираются в один primary PG.

### Пропускная способность одного воркера

- Обычный handler (Build/Research/Transport): 5–50ms per-event tx.
- AttackSingle c боем: 50–200ms.
- ACS AttackAlliance (multi-row locking, battle report): 100–500ms.
- При batch=100 и среднем 30ms/event → **~3000 events/min на воркер**. Линейно до упора в БД.

### Ограничения сверху

1. **Connection pool**. pgxpool обычно 10–25 conn на процесс. N воркеров × 15 = 15N. API-сервер ещё 20–30. При `max_connections=100` (дефолт) **потолок ~4 воркера без PgBouncer**.
2. **WAL / fsync**. При `synchronous_commit=on` обычный SSD даёт ~1000–3000 commit/s. При 6000 events/min × 3 write/event = 300 w/s — запас есть, но при 8+ воркерах и ACS-тяжёлом геймплее можно приблизиться.
3. **Hotspot contention**. Кросс-шардовые гонки (две атаки на одну планету из разных шардов) сериализуются PG через `FOR UPDATE`. Количество воркеров не помогает — лечится только геймплейной логикой или региональным шардированием.

### Рекомендуемая шкала

| DAU | Воркеров | Доп. требования |
|---|---|---|
| < 1k | **1** | Текущее состояние. После фаз 1-2 — стабильно. |
| 1k – 10k | **2** | `WORKER_SHARD_COUNT=2`. Покрывает SPOF. |
| 10k – 50k | **4** | Обязателен **PgBouncer** (transaction pooling). Фазы 3.7 (partitioning) и 3.8 (autovacuum) желательны. |
| 50k – 100k | **8** | PgBouncer + partitioning events (3.7) обязательны. Read-replica для domain-queries (building queue, fleet ETA) — разгрузить primary. |
| > 100k | **>8** | Выходит за рамки плана 09: региональное шардирование PG, возможно отдельные PG-инстансы на event-shard. Отдельная итерация. |

### Метрики для принятия решения

Не подбирать число воркеров по DAU вслепую — смотреть на экспортированные метрики (фаза 2):

- **`events_lag_seconds` p99 < 5s** — воркеров хватает. Если растёт → добавить воркер.
- **`pg_stat_activity` idle_in_transaction < 10%** — пул не затыкается. Если растёт → не добавлять воркер, ставить PgBouncer или увеличивать `max_connections`.
- **`event_handler_duration_seconds` p95 стабилен** — нет row-level contention. Если растёт вместе с числом воркеров → упираемся в hotspot'ы, доп. воркеры не помогут.

**Правило**: если растёт только lag — добавляем воркеры. Если idle_in_transaction или p95 handler duration — упираемся в БД, добавление воркеров ухудшит ситуацию.

### Deploy-конфиг

`deploy/docker-compose.prod.yml` (после 3.3 + 3.4) шаблон:

```yaml
worker:
  deploy:
    replicas: ${WORKER_REPLICAS:-2}
  environment:
    WORKER_SHARD_COUNT: ${WORKER_REPLICAS:-2}
    WORKER_SHARD_ID: # проставляется через entrypoint из hostname/replica index
```

Замечание: `docker-compose` не даёт из коробки per-replica ENV. Варианты:
- Вручную описать N сервисов (`worker-0`, `worker-1`, …) с фиксированным `SHARD_ID`. Надёжно, но неудобно при изменении N.
- Entrypoint-скрипт читает `HOSTNAME` (типа `worker_1`), извлекает индекс. Требует чтобы compose давал предсказуемые имена.
- Перейти на k8s StatefulSet — там из коробки `POD_NAME` с индексом.

Для MVP на 2–4 воркерах — достаточно варианта 1 (явные сервисы). При переходе на >4 — k8s.

## Миграции БД — план накатки

Таблица `events` живая (на ней в любой момент лежат wait-events игроков). Миграции не могут ломать формат существующих записей или блокировать таблицу надолго. План по фазам:

**Фаза 1 (миграция 00NN_events_retry.sql):**
```sql
ALTER TABLE events
  ADD COLUMN attempt smallint NOT NULL DEFAULT 0,
  ADD COLUMN next_retry_at timestamptz,
  ADD COLUMN last_error text;
```
PG 11+ поддерживает `ADD COLUMN ... DEFAULT` без переписывания таблицы (stored as metadata) — быстро даже на больших events. `next_retry_at` и `last_error` nullable — backfill не нужен. Handler'ы фазы 1 читают `attempt` с `COALESCE(attempt, 0)` на случай, если миграция ещё не накатилась в momentом развёртывания (версионный зазор worker vs. schema).

**Фаза 1 (миграция 00NN_events_dead.sql):**
```sql
CREATE TABLE events_dead (LIKE events INCLUDING ALL);
ALTER TABLE events_dead ADD COLUMN moved_at timestamptz NOT NULL DEFAULT now();
```

**Фаза 2 (миграция 00NN_events_trace.sql):**
```sql
ALTER TABLE events ADD COLUMN trace_id uuid;
CREATE INDEX CONCURRENTLY ix_events_trace_id ON events(trace_id) WHERE trace_id IS NOT NULL;
```
Старые wait-events остаются с `trace_id IS NULL` — grafana-панели обрабатывают NULL (группа «pre-tracing»).

**Фаза 3.3 (миграция 00NN_events_shard_key.sql):**
```sql
ALTER TABLE events ADD COLUMN shard_key text;
UPDATE events SET shard_key = user_id::text WHERE shard_key IS NULL AND user_id IS NOT NULL;
CREATE INDEX CONCURRENTLY ix_events_shard ON events(hashtext(shard_key), fire_at, id) WHERE state='wait';
```
`UPDATE ... WHERE shard_key IS NULL` на большом `events` может быть тяжёлым — выполнять **в окне low-traffic** или батчами по 10k строк (отдельный CLI в `backend/cmd/tools/`). **До накатки** — worker должен поддерживать оба режима (fallback на `user_id::text` если `shard_key IS NULL`).

**Фаза 3.7 (миграция 00NN_events_partitioning.sql):**
Отдельный план, требует `pg_dump` + пересоздание таблицы. Не делать до >1M events. Описать как ADR.

**Общий принцип:**
1. Сначала накатываем миграцию (добавляем колонки nullable + default).
2. Деплоим worker, который умеет **читать и писать** новое поле, но **не требует** его наличия (graceful fallback).
3. Делаем backfill старых записей (если нужен).
4. Делаем worker, который **требует** нового поля (уже все записи заполнены).
5. Следующий релиз может удалить fallback-код.

Это стандартный expand-contract pattern. Без него любой wait-event, созданный старой версией и обработанный новой, падает на `NULL` в новой колонке.

## Порядок обработки: нужна ли строгая последовательность?

Отдельный вопрос: обязаны ли events выполняться подряд (по `fire_at`), или часть можно параллелить? От ответа зависит, имеют ли смысл пункты 3.2 (адаптивный tick), 3.3 (шардирование) и 3.4 (≥2 воркеров).

### Матрица «Kind → затрагиваемые таблицы» (сжато)

| Kind | Пишет |
|---|---|
| BuildConstruction (1), Research (3) | buildings/research, construction_queue |
| BuildFleet (4), BuildDefense (5) | ships/defense, shipyard_queue |
| Transport (7), Return (20), Colonize (8) | fleets, planets (ресурсы) |
| AttackSingle (10) | fleets, fleet_ships, ships, defense, planets, battle_reports, messages, debris_fields, artefacts_user |
| **AttackAlliance (12)** | **все фloты ACS-группы**, ships, defense, planets, battle_reports, messages |
| Recycling (9) | debris_fields, fleets, planets |
| AlienAttack (35), RocketAttack (16) | planets, defense, ships, messages |
| Repair (50), Disassemble (51) | ships, defense, repair_queue, planets |
| ArtefactExpire (60), ArtefactActivate (63) | artefacts_user, (planet/research/user в зависимости от эффекта) |
| OfficerExpire (62) | officer_active, users.credit |
| Expedition (15) | fleets, events, (планеты/ресурсы/карма) |

Коды якорей: [fleet/attack.go](../../backend/internal/fleet/attack.go), [fleet/acs_attack.go](../../backend/internal/fleet/acs_attack.go), [fleet/events.go](../../backend/internal/fleet/events.go), [artefact/expire.go](../../backend/internal/artefact/expire.go), [event/handlers.go](../../backend/internal/event/handlers.go).

### Где порядок КРИТИЧЕН

1. **BuildConstruction + AttackSingle на одной планете с близким `fire_at`.** Если обрабатываются параллельно, две транзакции конкурируют за строку `planets` и `buildings`. Row-lock сериализует их, но результат зависит от того, кто первым взял lock — игрок может либо потерять только что достроенное здание, либо сохранить. Правильнее: строго по `fire_at` (тот, чьё `fire_at` раньше, применяется первым).
2. **ArtefactExpire + AttackSingle того же игрока.** Attack читает активные артефакты для бонуса ([fleet/attack.go:94-97](../../backend/internal/fleet/attack.go#L94-L97)). Если Expire успел раньше — бонус не применён; если позже — применён. При параллельной обработке один handler может прочитать row, который другой только что пометил `deleted`. Row-lock защищает от разрыва, но **результат становится недетерминированным** относительно `fire_at`.
3. **Recycling + AttackSingle на одних координатах.** Если Attack раньше — создаёт новые обломки, которые Recycling подхватит. Если Recycling раньше — подберёт старые, Attack добавит свои, сборщик уже улетел.
4. **ACS AttackAlliance (Kind=12) — самый опасный.** Handler в одной транзакции читает **все** флоты группы и берёт `FOR UPDATE` на каждый ([acs_attack.go:70-77](../../backend/internal/fleet/acs_attack.go#L70-L77)). Если воркер A обрабатывает group=G через fleet_1, а воркер B — через fleet_2 (у них отдельные события с одним и тем же `fire_at`) — **гарантированный deadlock**: A залочил fleet_1, ждёт fleet_2; B залочил fleet_2, ждёт fleet_1. PG обнаружит deadlock и убьёт одного — event пойдёт в `state='error'`.

### Где порядок НЕ критичен

- События **разных игроков на разных планетах без общих целей** — полностью независимы.
- BuildFleet / Research / Repair / OfficerExpire / ArtefactActivate разных игроков — идут параллельно без конфликтов.
- AttackSingle игрока A на планету P + BuildConstruction игрока B на свою планету Q — не пересекаются ни по одной строке.
- Два transport'а разных игроков в разные направления — независимы.

### Что даёт row-level locking PG «бесплатно»

Handler'ы уже используют `SELECT ... FOR UPDATE` на доменные строки (planets, fleets, ships). Это значит: **даже при параллельном запуске воркеров корректность каждого отдельного handler'а не нарушается** — PG сериализует доступ к строке. Но:

- Это не даёт гарантии **порядка по `fire_at`** между конкурирующими events — побеждает тот, кто первым взял lock.
- Это не спасает от deadlock на multi-row locking (ACS, кросс-планетные события алиенов).
- Это не даёт serializable-поведение для read-then-write паттерна (Attack читает `planets.building_level`, потом рассчитывает урон) — без `REPEATABLE READ` / `SERIALIZABLE` возможно чтение stale значения относительно commit'а параллельного BuildConstruction.

### Вывод

**Нужна последовательность, но не глобальная — а по ключу.**

Три уровня параллелизма, от безопасного к опасному:

| Модель | Безопасность | Скорость | Сложность |
|---|---|---|---|
| 1 воркер, все события подряд по `fire_at` (текущее) | ✅ безопасно, детерминировано | ❌ SPOF, bottleneck | ✅ просто |
| N воркеров, **шардирование по `user_id`** (events одного юзера — строго в одном воркере и по `fire_at`) | ✅ безопасно для user-scope, но ACS/алиены/recycling-cross-user требуют особой обработки | ✅ масштабируется линейно | 🟡 средне |
| N воркеров, только SKIP LOCKED без шардирования | ❌ deadlock на ACS, недетерминизм по `fire_at` для конфликтующих событий одного юзера | ✅ быстро | ✅ просто |

**Рекомендация для плана 09** (согласована с нумерацией фазы 3):

- **Фаза 1** (надёжность, per-event транзакция) — безопасна при 1 воркере. Порядок по `fire_at, id` сохраняется.
- **Пункт 3.3 (шардирование) обязателен ДО пункта 3.4 (≥2 воркера)**. Иначе получим deadlock на первой же ACS-атаке.
  - Схема: `WHERE (user_id IS NULL AND $shard_id = 0) OR hashtext(shard_key) % $N = $shard_id`.
  - События с `user_id IS NULL` (системные — AlienAttack spawn, глобальные cron'ы) обрабатываются шардом 0.
  - Для ACS: все события группы проставляют `shard_key = group_id::text` при создании — попадают в один шард.
- **Кросс-юзерские конфликты** (Attack игрока A на планету игрока B, Recycling на координатах чужого боя) решаются row-level locking на `planets` и `debris_fields` — PG сериализует. Для редкого кейса «две атаки на одну планету в одну и ту же секунду» `FOR UPDATE` на planets + идемпотентность handler'ов достаточны.

**До включения шардирования — оставить 1 воркер**. Адаптивный tick (3.2) полезен и на одном воркере: пока batch полный — не ждём Ticker.

### Обновлённые оценки в свете этого анализа

- Масштабируемость 5/10 → после 3.2 (адаптивный tick) станет 6/10, после 3.3+3.4 (shard + ≥2 воркера) — 8/10.
- Риск deadlock'а есть **уже сейчас**, если кто-то случайно запустит второй воркер (например, через scale-up в compose). Добавить в документацию деплоя явное «worker replicas = 1, пока не реализованы 3.0 (tie-breaker) и 3.3 (shard)».

## Выбор технологии: почему остаёмся на PostgreSQL

Вопрос «не переписать ли на Kafka / Redis Streams / очередь-менеджер» поднимался отдельно. Вердикт: **остаёмся на PG + SKIP LOCKED**. Ниже — разбор.

### Почему не Kafka

Kafka — это **лог сообщений**, а у нас **очередь отложенных задач с произвольным `fire_at` в будущем**. Принципиально разные вещи:

- Kafka не умеет «доставить через 3 часа 47 минут». Приходится городить delay-топики (`delay-5m`, `delay-1h`, …) с relay-consumer'ами, перекидывающими сообщение между ними. Антипаттерн, но именно так delayed jobs в Kafka и делают.
- Kafka не умеет **отменить** запланированное сообщение. У нас отмена — повседневная механика (recall флота, отмена постройки, cancel lot). В PG это `DELETE FROM events WHERE id=$1`.
- Kafka не даёт **транзакционно** связать изменение домена (планета/флот/ресурсы) и постановку задачи. Нужен outbox + relay — ещё один движущийся компонент. В PG — одна `InTx`.
- Операционная нагрузка: +Kafka/KRaft, свои метрики/алерты/backup. Для соло-проекта ощутимо.

Kafka оправдан при: throughput десятки-сотни тысяч msg/sec, несколько независимых consumer-групп на один поток, event-sourcing как primary. У нас ни одного из трёх.

### Разбор альтернатив

| Решение | Когда оправдано у нас | Вердикт |
|---|---|---|
| **PG + SKIP LOCKED (текущее)** | до десятков тысяч активных юзеров, ~1–10 events/sec/юзер | ✅ оставить |
| **Redis Streams / Sorted Set** | нужен sub-100ms latency на очередь | ❌ теряем транзакционность с доменом |
| **River / Asynq / Neoq** (Go-либы поверх PG) | готовый retry/dead-letter/dashboard из коробки | 🟡 рассмотреть вместо самописных фаз 1-2 |
| **RabbitMQ с delayed-exchange** | кросс-сервисные очереди, fanout | ❌ overkill, один consumer |
| **Temporal / Cadence** | долгоживущие workflow со сложным состоянием (многошаговый alien AI, ACS-координация) | 🟡 возможно для будущего state-machine, не для замены текущего |
| **Kafka / NATS JetStream** | высокий throughput, fan-out consumers | ❌ не наш кейс |

### Единственная реальная альтернатива: River

[River](https://github.com/riverqueue/river) — Go-библиотека поверх PG, ровно под нашу модель (SKIP LOCKED + jsonb payload), но с готовыми retry+backoff, scheduled jobs, уникальными jobs (идемпотентность по ключу), periodic jobs, dead-letter, web UI, per-job timeout, priority queues. Это ровно то, что фазы 1-2 плана пишут руками.

**Почему всё равно НЕ мигрируем**:

- River навязывает свою таблицу `river_job`. Наши доменные колонки `user_id` / `planet_id` / `kind` не маппятся 1:1 — всё пришлось бы пихать в payload.
- **Убийственный аргумент**: игровой UI постоянно делает domain-query по очереди: «что в building queue», «когда прилетит флот», «когда освободится лаборатория». Это SELECT-ы с фильтрами по `user_id`/`planet_id`/`kind`. River спроектирован под чёрный ящик job queue, domain-query-friendly модели не даёт — эти запросы пойдут против зерна.
- Миграция 25 Kind + 19 handler'ов + слой совместимости — ~1-2 недели работы, взамен — ровно функционал из плана 09, но с архитектурным несоответствием домену.

### Когда пересмотреть

- DAU > ~50k и один PG-инстанс перестаёт справляться → сначала read-replicas и партицирование `events` по `fire_at`, смена технологии — **только если это не помогло**.
- Появится второй consumer на тот же поток (аналитика, внешние интеграции) → outbox → Kafka/NATS как **внешняя** шина, PG остаётся primary.
- Появится длинный workflow со сложным state-machine (полноценный alien AI HALT-cycle, ACS с задержкой боя) → тогда точечно Temporal для этого workflow, остальное остаётся в PG.

## Out of scope

- Миграция event-loop на Redis Streams / NATS / Kafka — см. раздел «Выбор технологии». Избыточно для игровой PvE/PvP-экономики с ~1–10 events/sec на активного игрока. PG + SKIP LOCKED хватает до десятков тысяч активных юзеров.
- CRDT / distributed state — одна PG, одна правда.
- Event sourcing full-stack — наш `events` это taskqueue, не event-log в смысле DDD.

## Ссылки

- [CLAUDE.md](../../CLAUDE.md) — §17 правила кода, §6 идемпотентность.
- [oxsar-spec.txt §6, §18](../../oxsar-spec.txt) — требования к event-loop.
- [docs/status.md](../status.md) — Event-loop worker ✅ M3.
- [docs/simplifications.md](../simplifications.md) — существующие trade-offs.
