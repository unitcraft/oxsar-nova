---
title: 34 — Grafana-дашборд "Operator view" для админа
date: 2026-04-26
status: draft
---

# План 33: операторский дашборд в Grafana

**Цель**: сделать **один экран**, на который заходишь раз в день и за
30 секунд понимаешь — игра здорова или нет. Без знания PromQL, без
переключения вкладок.

**Контекст**: после плана 32 у нас есть Prometheus + Grafana и
auto-provision. Текущий dashboard `oxsar-overview.json` — технический
(rate, p95, queue lag) — полезный для DevOps, но непонятный для
геймдиз/админа. Нужен второй дашборд: **«всё в порядке»** vs
**«надо смотреть логи»**.

---

## 1. Что показывать (5 секций по 4-5 панелей)

### 1.1 Top row: «жив или нет» (5 stat-панелей × 4×4 grid)

Все цветные: зелёный = ok, жёлтый = warning, красный = алерт.

| Панель | Метрика / запрос | Зелёный когда |
|---|---|---|
| **Backend** | `up{job="backend"}` | =1 |
| **Workers (live)** | `count(up{job="worker"} == 1)` | =2 (наши replicas) |
| **Postgres** | `up{job="postgres"}` | =1 |
| **Redis (через chat sub)** | `oxsar_chat_subscriber_active` *(новая, см. §3)* | =1 |
| **Errors / 1h** | `sum(increase(oxsar_events_processed_total{state="error"}[1h]))` | =0 |

### 1.2 Игроки (4 stat-панели)

Эти метрики **новые** — экспортируем из БД (см. §3).

| Панель | Что | PromQL после §3 |
|---|---|---|
| **Игроков всего** | `oxsar_players_total` | мгновенное |
| **Онлайн (≤15 мин)** | `oxsar_players_online` | мгновенное |
| **Регистраций сегодня** | `increase(oxsar_players_total[24h])` | за день |
| **В отпуске (umode)** | `oxsar_players_umode` | мгновенное |

### 1.3 Здоровье event-loop (timeseries × 2)

Уже есть метрики, но в красивой подаче.

| Панель | Запрос | Алерт-зона |
|---|---|---|
| **События/мин по статусу** | `sum by (state) (rate(oxsar_events_processed_total[1m])) * 60` | error > 0 — красная заливка |
| **Очередь wait + lag** | `sum(oxsar_events_queue_depth{state="wait"})` (левая Y) и `max(oxsar_events_lag_seconds)` (правая Y) | lag > 60s — красный |

### 1.4 Бой и экономика (4 stat за период + 1 timeseries)

«За день/час сколько произошло событий».

| Панель | Метрика |
|---|---|
| **Боёв за 24ч** | `oxsar_battles_total_24h` *(новая)* |
| **Экспедиций за 24ч** | `oxsar_expeditions_total_24h` *(новая)* |
| **Сообщений в чате /1ч** | `oxsar_chat_messages_total_1h` *(новая)* |
| **Кредитов куплено /24ч** | `sum(increase(oxsar_credit_purchases_credits[24h]))` *(новая)* |
| **Боёв / час (timeseries)** | `rate(oxsar_battles_total[1h])` |

### 1.5 Scheduler (1 table + 1 timeseries)

Центральный показатель того, что план 32 работает.

| Панель | Что | Подача |
|---|---|---|
| **Last run, sec ago** | `time() - oxsar_scheduler_job_last_run_timestamp` | table, threshold по cron-периоду каждой job (alien 6ч, daily 24ч) |
| **Runs by status** | `sum by (job, status) (increase(oxsar_scheduler_job_runs_total[24h]))` | stacked bar — видно ok/skip/error |

### 1.6 Ресурсы хоста (2 timeseries)

Чтобы понять, упирается ли VPS.

| Панель | Запрос |
|---|---|
| **RAM по контейнерам** | `process_resident_memory_bytes{job=~"backend\|worker"}` |
| **Postgres connections** | `pg_stat_database_numbackends{datname="oxsar"}` |

---

## 2. Визуальный стиль

- **Тема**: dark (Grafana default).
- **Refresh**: 30s (быстрее не нужно — scrape_interval=15s).
- **Time range default**: last 24h.
- **Variable**: `$instance` для фильтра по worker'ам (опционально).
- **Цвета**: зелёный → жёлтый → красный по threshold'ам, без
  кастомных палитр.
- **Подписи**: все RU, не EN. «События в очереди», не «Events queue».
- **Иконки/эмодзи**: не используем (CLAUDE.md: «only if user explicitly
  requests»).

---

## 3. Новые метрики, которые нужно добавить

Текущий backend экспортирует только event-related + scheduler. Для
бизнес-панелей §1.2 и §1.4 нужны **gauges, обновляемые из БД**.

**Где живёт**: расширяем уже существующий `event/worker.go::RunMetricsUpdater`
(15s ticker), либо отдельная горутина в server.

### 3.1 Метрики игроков (gauges)

```go
// pkg/metrics/metrics.go
PlayersTotal  prometheus.Gauge   // SELECT count(*) FROM users WHERE deleted_at IS NULL
PlayersOnline prometheus.Gauge   // ... AND last_seen_at > now() - 15min
PlayersUmode  prometheus.Gauge   // ... AND umode = true
```

Updater (раз в минуту достаточно) — server-side, чтобы не дёргать БД с
каждого worker'а.

### 3.2 Метрики бизнес-событий (counters)

Инкрементить в handler'ах, **не** обновлять из SELECT — counter
правильнее, чем gauge на исторические данные.

```go
BattlesTotal       *prometheus.CounterVec // labels: result (won|lost|draw)
ExpeditionsTotal   *prometheus.CounterVec // labels: outcome (gain|loss|nothing)
ChatMessagesTotal  prometheus.Counter     // в chat.handler.send
CreditPurchasesTotal *prometheus.CounterVec // labels: provider, status — уже есть в payment? проверить
```

Дёргаются:
- `BattlesTotal` — в `transport.AttackHandler` после успешного боя.
- `ExpeditionsTotal` — в `transport.ExpeditionHandler`.
- `ChatMessagesTotal` — в `chat.Handler.Send` и WS-receive после INSERT.
- `CreditPurchasesTotal` — в payment success-callback.

### 3.3 Chat Redis subscriber active gauge

Уже есть `pubSubRunning atomic.Bool` в `chat.Hub`. Экспортировать как
gauge:

```go
ChatSubscriberActive prometheus.Gauge // 1 = subscriber connected, 0 = degradation
```

Обновлять в `runSubscriber` при `Store(true/false)`.

---

## 4. Фазы

### Фаза 1 — новые метрики и updater (M, ~1 день)

- [ ] `pkg/metrics/metrics.go`: добавить 7 новых метрик (см. §3).
- [ ] `cmd/server/main.go`: запустить горутину `RunBusinessMetrics`
      (1m ticker, читает users-таблицу, пишет gauges).
- [ ] Инкременты в:
  - `internal/fleet/transport.go::AttackHandler` — `BattlesTotal`.
  - `internal/fleet/transport.go::ExpeditionHandler` — `ExpeditionsTotal`.
  - `internal/chat/handler.go::Send` + WS-receive — `ChatMessagesTotal`.
  - `internal/payment/...success...` — `CreditPurchasesTotal`.
  - `internal/chat/hub.go::runSubscriber` — `ChatSubscriberActive`.
- [ ] Юнит-тесты: counter инкрементируется в handler'ах (mock-counter
      или `testutil.ToFloat64`).

**Готовность**: 1 PR, ~150 строк.

### Фаза 2 — дашборд "Operator view" (M, ~0.5 дня)

- [ ] `deploy/grafana/provisioning/dashboards/oxsar-operator.json`:
      6 секций × 3-5 панелей по §1.
- [ ] Threshold'ы и заливка по «зелёный/жёлтый/красный».
- [ ] Dashboard tags: `oxsar`, `operator` — чтобы найти в списке.
- [ ] uid: `oxsar-operator` (стабильный, не GUID).

**Готовность**: 1 PR, ~600 строк JSON.

### Фаза 3 — старый "events + scheduler" → "Tech view" (S, ~0.1 дня)

- [ ] Переименовать `oxsar-overview.json` title → "oxsar-nova: tech
      view (events + scheduler)" — для DevOps-отладки.
- [ ] tags: `oxsar`, `tech`.
- [ ] В README дашборда (markdown-комментарий в первой панели):
      «Operator view = здоровье; Tech view = детали для дебага».

**Готовность**: 5 правок в JSON.

### Фаза 4 — алерты (опционально, S, ~0.5 дня)

- [ ] `deploy/grafana/provisioning/alerting/rules.yml`: 3-4 правила:
  - error-rate > 0 за 5 мин,
  - workers up < 2 за 1 мин,
  - lag > 120s за 2 мин,
  - scheduler last_run > 2× cron-period.
- [ ] Notification policy: email или Telegram-bot
      (`GRAFANA_NOTIFY_*` env).
- [ ] Без contact point оставлять не имеет смысла — UI-алерты в
      dashboard'е и так видны.

**Готовность**: 1 PR, ~80 строк YAML + ENV в monitoring.yml.

### Фаза 5 — документация (S, ~0.2 дня)

- [ ] `docs/ops/monitoring.md`:
  - что есть (Operator + Tech view),
  - как читать (зелёный/жёлтый/красный),
  - частые сценарии («ошибки растут — куда смотреть»),
  - как запустить, где логин/пароль,
  - troubleshooting (что делать, если метрики не видны).

**Готовность**: 1 markdown.

---

## 5. Что НЕ делаем

- **Не делаем** custom Grafana plugins или экзотические виджеты —
  стандартных stat/timeseries/table хватит.
- **Не делаем** Loki/log-aggregation — `docker logs` + grep пока хватает.
- **Не делаем** distributed tracing (Tempo/Jaeger) — не тот масштаб.
- **Не делаем** SLO/SLI с error budget'ом — нет SLA-обещаний игрокам.
- **Не делаем** мобильное Grafana-app или экспорт PDF — для DevOps
  достаточно браузера.
- **Не делаем** node-exporter (host-уровень CPU/disk) — у нас один
  VPS, его и без графика видно через `htop`.

---

## 6. Риски и митигации

1. **Cardinality blow-up** на counters с user_id label.  
   *Митигация*: НЕ добавляем user_id ни в одну метрику. Только статусы
   и kind'ы.
2. **Нагрузка на БД от 1m polling в business-metrics.**  
   *Митигация*: 1 минута интервал, индексы по `last_seen_at` есть, COUNT(*)
   на 10к users — миллисекунды.
3. **Дашборд показывает «всё ок», но что-то реально сломано** — метрики
   не покрывают всех bugs.  
   *Митигация*: дашборд **дополняет** логи, не заменяет. В каждой
   секции — link «логи за этот период» (Grafana data link на kibana
   когда появится).
4. **Цвета threshold'ов не соответствуют реальности** (например,
   1 ошибка в час нормальная для нас, но дашборд красный).  
   *Митигация*: thresholds задаются по факту, после 1-2 недель работы
   подстраиваем.

---

## 7. Definition of done

- В Grafana 2 дашборда: **"Operator view"** (новый, главный) и
  **"Tech view"** (текущий, переименованный).
- Operator view загружается в 1 запросе и читается за 30 секунд:
  «зелёный — спим, желтый — глянуть после кофе, красный — лезем
  в логи».
- Все 7 новых метрик появляются в `/metrics` после первого
  использования (battle, expedition, chat, etc).
- 1m business-metrics-updater работает в server, не в worker.
- `docs/ops/monitoring.md` объясняет, как пользоваться, без знания
  PromQL.
- Все backend-тесты зелёные.

---

## 8. Зависимости

- **План 32** (multi-instance + monitoring stack) — выполнен.
- Новых Go-зависимостей нет (всё на стандартном `prometheus/client_golang`).
- Grafana 11.1.0 уже в `docker-compose.monitoring.yml`.

---

## 9. Что сделать в первую очередь

Если выбирать одну фазу — **Фаза 2 без Фазы 1**: даже без новых
бизнес-метрик можно сделать читаемый дашборд из того, что уже есть
(`up`, `oxsar_events_*`, `pg_stat_*`). Это уберёт «не понимаю что
смотреть» сразу.

Но без §1.2 (игроки) и §1.4 (бой/экономика) дашборд будет техническим —
поэтому в идеале Фаза 1 → Фаза 2 в одну итерацию.

Минимальный MVP за 0.5 дня: **Фаза 2 + ChatSubscriberActive из Фазы 1**
(остальные новые метрики оставим на потом).
