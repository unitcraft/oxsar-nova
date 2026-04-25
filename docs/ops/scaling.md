---
title: Multi-instance scaling
date: 2026-04-26
---

# Multi-instance scaling

Backend готов к запуску в N≥2 инстансов. Этот документ описывает, что
именно сделано, как поднять scale-сетап для smoke-теста и на что
смотреть в логах/метриках.

## Что готово

| Компонент | Поведение при N≥2 |
|---|---|
| HTTP API (chi handlers) | stateless, БД-операции в транзакциях — race-safe |
| Auth rate-limiter | Redis-based, общий счётчик |
| LastSeen middleware | UPSERT, race-safe |
| WebSocket чат (`chat.Hub`) | Redis pub/sub (`chat:*`); fan-out через subscriber-горутину |
| Главный event-loop (`event/worker.go`) | `FOR UPDATE SKIP LOCKED` — несколько worker'ов конкурируют без дублей |
| Singleton-job'ы (alien_spawn, inactivity_reminders, expire_temp_planets, event_pruner, score_recalc_all) | scheduler с advisory lock — выполняет ровно один инстанс |
| MetricsUpdater (15s ticker) | read-only gauge.Set, idempotent — может тикать на всех инстансах |

Подробности — в [плане 32](../plans/32-multi-instance.md).

## Запуск 2 worker'а локально

```bash
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.scaling.yml up --build
```

Что должно произойти:

1. Postgres + Redis поднимаются как обычно.
2. Migrate отрабатывает один раз.
3. **Backend × 2** и **worker × 2** стартуют параллельно.
4. Каждый worker регистрирует scheduler-job'ы и пишет в логи:
   ```
   scheduler job registered job=alien_spawn schedule="0 */6 * * *"
   scheduler started jobs=5
   ```
5. На cron-tick'е **один** из worker'ов залогирует `scheduler job ok`,
   второй — `scheduler job skipped (lock held by another instance)`
   (на debug-уровне; counter `oxsar_scheduler_job_runs_total{status="skip"}`
   инкрементится).

## Что проверять

### Singleton-задачи (НЕ должны дублироваться)

В логах на cron-tick'е каждой job'ы — ровно один `status=ok` через все
инстансы:

```bash
docker compose ... logs worker --since 1m | grep "scheduler job"
# scheduler job ok job=alien_spawn duration=120ms      ← один инстанс
# (нет второй строки от того же job'а)
```

Метрика:

```promql
sum(oxsar_scheduler_job_runs_total{job="alien_spawn", status="ok"}) ==
  rate(oxsar_scheduler_job_runs_total{job="alien_spawn"}[1h]) * 3600 / 6
# (job выполняется 4 раза в день; status="ok" совпадает; status="skip" растёт от второго инстанса)
```

### Чат (должен синхронизироваться между инстансами)

Если открыть две вкладки браузера, проксируя в разные backend-инстансы,
и написать в global-чат с одной — другая должна увидеть. До плана 32
такого не было: in-memory Hub видел только локальных клиентов.

В логах backend'ов:

```
chat: redis subscriber active pattern=chat:*
```

Если Redis недоступен — fallback на local-only:

```
chat: redis psubscribe failed, retrying
```

Чат продолжит работать в degradation-режиме (только клиенты одного
инстанса видят друг друга), полного отказа нет.

### Главный event-loop

`FOR UPDATE SKIP LOCKED` уже работает. Метрика
`oxsar_events_processed_total` растёт суммарно по всем worker'ам без
дублей.

## Метрики для алертов

| Метрика | Что значит | Алерт |
|---|---|---|
| `oxsar_scheduler_job_runs_total{status="error"}` | scheduler-job упала с ошибкой | `rate > 0` за 1ч |
| `oxsar_scheduler_job_last_run_timestamp{job="alien_spawn"}` | unix-ts последнего запуска (любой статус) | `time() - last_run > 7h` (job каждые 6h) |
| `oxsar_scheduler_job_duration_seconds{job}` | длительность не-skip запусков | p99 > 5min для daily-job |

## ENV-настройки

Worker читает расписание из `configs/schedule.yaml`. Можно переопределить:

| ENV | Эффект |
|---|---|
| `SCHEDULE_FILE=/path/to/schedule.yaml` | альтернативный путь к YAML |
| `SCHEDULER_<JOB>_CRON="0 */3 * * *"` | заменить расписание конкретной job'ы |
| `SCHEDULER_<JOB>_ENABLED=false` | временно отключить job'у без правки YAML |
| `SCHEDULER_DISABLE_ALL=true` | kill-switch для всего scheduler'а (job'ы не запускаются) |

Имя `<JOB>` пишется в верхнем регистре (`SCHEDULER_ALIEN_SPAWN_CRON`),
маппится на ключ из YAML в lower_case.

## Известные ограничения

- **Catch-up пропущенных запусков не поддерживается.** Если worker лежал
  во время cron-tick'а (например, 04:00 UTC для score_recalc_all),
  следующий запуск будет в 04:00 следующего дня. Если потеря критична —
  отдельная задача (план 32 §5).
- **Sticky session для WebSocket не нужны** — Redis pub/sub делает их
  ненужными. Если reverse proxy всё равно шлёт sticky — это не вредит.
- **chat.Hub при недоступном Redis** деградирует до single-instance:
  клиенты разных backend-инстансов перестают видеть друг друга, но
  внутри одного инстанса чат работает. Полного отказа нет.
- **Postgres advisory lock привязан к connection.** Если процесс
  worker'а упал OOM-killer'ом — Postgres сам сбросит lock через
  session disconnect.

## Откат

Если нужно срочно вернуться на single-instance:

```bash
docker compose -f deploy/docker-compose.yml up    # без overlay
```

Или: `SCHEDULER_DISABLE_ALL=true` остановит все scheduler-job'ы (но
останутся handler'ы события `KindScoreRecalcAll` для legacy wait-events,
если они есть в БД).

## Ссылки

- [План 32 — Multi-instance readiness](../plans/32-multi-instance.md)
- [План 31 — zero-downtime deploy + feature flags](../plans/31-zero-downtime-deploy.md) (зависимость)
- [backend/internal/scheduler/](../../backend/internal/scheduler/) — пакет scheduler
- [backend/internal/locks/](../../backend/internal/locks/) — advisory locks
- [configs/schedule.yaml](../../configs/schedule.yaml) — расписание
