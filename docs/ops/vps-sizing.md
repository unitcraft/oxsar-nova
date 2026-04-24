# VPS sizing: железо под разное количество пользователей

## Цель

Дать конкретные рекомендации по VPS-конфигурациям для oxsar-nova на разных этапах роста DAU. Все цифры — от кода, а не из воздуха.

## Методология

Расчёты основаны на аудите кодовой базы (апрель 2026):
- 38 таблиц, 49 миграций, размер per-player измерен по DDL ([migrations/](../../migrations/)).
- Handler duration замерен по типу события (Build/Research: 5-50ms, Attack: 50-200ms, ACS: 100-500ms).
- HTTP QPS выведен из TanStack Query `staleTime=30s` (дефолт) × 226 useQuery в [frontend/src/](../../frontend/src/).
- WS-характеристики из [backend/internal/chat/hub.go](../../backend/internal/chat/hub.go) (32-slot buffer, in-memory map).
- Сверка с планом 09 (event-system) — потолок воркеров определяется PG-инстансом.

**CCU/DAU ratio** = 5–10% (стандарт для OGame-подобных). Будем считать **8%** как рабочее значение.

## Ключевые факты о нагрузке

### На одного активного игрока

**БД footprint**:
- Hot state (users + planets + buildings + research): ~66 KB.
- Warm state (fleets + events + queues): ~27 KB.
- Cold state (messages + battle_reports + logs за 6 мес): ~100-200 KB.
- **Суточный прирост: 13–20 KB/player/day** (в основном res_log, messages, events).

**Трафик**:
- Passive viewing: 0.13 req/sec.
- Active play: 0.4 req/sec.
- War/raid mode: 1.5 req/sec.

**Events**:
- 4–6 events/hour/player (постройки, полёты, верфь, repair) = ~0.0015 events/sec.

**WebSocket**:
- ~95% CCU подключены к global chat (у многих открыта вкладка чата).
- Память: 500 B/connection.

### Критические bottleneck'и (отдельно)

1. **Score recalc каждые 5 мин по всем юзерам** ([backend/cmd/worker/main.go](../../backend/cmd/worker/main.go)) — при 10k+ DAU один цикл не укладывается в 5 мин. **Решение — гибридная схема**: incremental через `withScore` (уже работает) + batch-SQL раз в сутки через CTE (один UPDATE вместо N×5 round-trip'ов) + удаление 5-минутного ticker. Детально в [plans/09-event-system.md#оптимизация-score-recalcall](09-event-system.md#оптимизация-score-recalcall).
2. **res_log растёт неограниченно** — 5.5 KB/player/day audit-log без retention. При 10k DAU это 55 MB/день = **1.6 GB/мес**. Нужен 90-дневный TTL через cron или партицирование.
3. **Chat hub in-memory** ([chat/hub.go](../../backend/internal/chat/hub.go)) — при 2+ backend-инстансах сообщения не распространятся между ними. Нужен Redis pub/sub.

Эти три пункта должны быть закрыты **до** перехода к большим конфигурациям (10k+ DAU). Они не архитектурные — правятся за 1-2 спринта.

## Профили VPS

### Профиль 0 — Dev / localhost

Одна машина разработчика. Не обсуждается в sizing.

### Профиль 1 — Soft launch (< 500 DAU, < 50 CCU)

**Один VPS, всё вместе.** Окупается, пока PG не упирается в IOPS.

| Компонент | RAM | CPU |
|---|---|---|
| nginx (SPA + proxy) | 30 MB | минимум |
| backend:server | 200 MB | 0.3 core |
| backend:worker (1 шт) | 150 MB | 0.2 core |
| postgres 16 | 500 MB | 0.5 core |
| redis 7 | 100 MB | минимум |
| **Итого** | **~1 GB активно** | **~1 core под нагрузкой** |

**Рекомендация**:
- **2 vCPU / 4 GB RAM / 40 GB NVMe SSD.**
- Hetzner CX22 (~€4/мес), DigitalOcean $24/мес, Selectel шeap tier, Timeweb Cloud.
- Достаточно запаса для пиков и для PostgreSQL page cache.

**Настройки PG**: `shared_buffers=1GB`, `effective_cache_size=2GB`, `work_mem=8MB`, `max_connections=100`, `synchronous_commit=on`.

**БД объём**: старт ~100 MB, через 3 месяца при 500 DAU — ~3-5 GB. 40 GB диска хватает надолго.

### Профиль 2 — Ранняя стадия (500–2000 DAU, 40–160 CCU)

**Всё ещё один VPS**, но с запасом.

| Компонент | RAM | CPU |
|---|---|---|
| nginx | 50 MB | минимум |
| backend:server | 400 MB | 1 core (200 HTTP req/sec) |
| backend:worker (1-2 шт) | 200 MB | 0.5 core |
| postgres | 2 GB | 1-2 core |
| redis | 200 MB | минимум |
| **Итого** | **~3 GB** | **~3 core** |

**Рекомендация**:
- **4 vCPU / 8 GB RAM / 80 GB NVMe.**
- Hetzner CPX31 (~€14/мес), Selectel, Timeweb, netcup RS 2000 G10.
- Один воркер — достаточно (из плана 09, см. таблицу). Добавить второй только ради HA.

**Уже пора сделать**:
- Автобэкап PG раз в сутки на отдельное хранилище (S3-совместимое).
- Настроить fail2ban или Cloudflare перед nginx — браузерные MMO ловят DDoS.
- Мониторинг (Grafana Cloud free tier или Netdata).

### Профиль 3 — Рост (2k–10k DAU, 160–800 CCU)

**Разделение на app и db.** Один PG уже не делит память с backend'ом.

| Роль | Конфиг | Зачем |
|---|---|---|
| App VPS | 4 vCPU / 8 GB | nginx + backend × 2-3 + worker × 2 |
| DB VPS | 4 vCPU / 16 GB / 200 GB NVMe | PostgreSQL + Redis (или Redis на app) |

**Рекомендация**:
- **App**: Hetzner CPX31 (€14) или CX42 (€22).
- **DB**: Hetzner CCX23 dedicated (€29) — dedicated vCPU критичен для PG под нагрузкой, burstable инстансы не подходят.
- **Итого ~€50-60/мес**.

**Сетевой layout**: private network между app и db (обязательно, latency <1ms), публичный доступ к PG закрыт.

**Что обязательно должно быть сделано в коде** перед переходом на Профиль 3:
- Incremental score recalc вместо глобального (см. bottleneck #1).
- res_log retention (см. bottleneck #2).
- Фазы 1-2 плана 09 (retry + observability).

**PG-настройки**: `shared_buffers=4GB`, `effective_cache_size=12GB`, `work_mem=16MB`, `max_connections=200`, WAL на отдельный диск (если есть возможность).

**БД объём**: ~10-30 GB (зависит от ретеншена). Backup — ежедневно + WAL-архивирование.

### Профиль 4 — Масштаб (10k–50k DAU, 800–4000 CCU)

**PgBouncer обязателен. Несколько app-нод. Read-replica PG.**

| Роль | Конфиг | Количество |
|---|---|---|
| App | 4 vCPU / 8 GB | 3-4 шт |
| Worker | 2 vCPU / 4 GB | 4 шт (с шардированием из плана 09) |
| DB primary | 8 vCPU / 32 GB / 500 GB NVMe dedicated | 1 |
| DB replica (readonly) | 4 vCPU / 16 GB | 1 |
| Redis | 2 vCPU / 4 GB | 1 |
| Load balancer | — | HAProxy на app или managed LB |

**Рекомендация**:
- Hetzner: app × 4 (CPX31 €14 × 4 = €56) + worker × 4 (CPX21 €8 × 4 = €32) + DB primary CCX33 €59 + replica CCX23 €29 + Redis CX22 €4 + LB €6 = **~€190/мес**.
- Либо AWS EC2: ~$800-1000/мес (ожидаемо дороже).

**Обязательно в коде**:
- Фаза 3 плана 09 полностью (shard-by-user, адаптивный tick, partitioning events).
- PgBouncer в transaction-pooling режиме перед PG.
- Chat hub через Redis pub/sub (иначе сообщения не долетят между app-нодами).
- Idempotency keys для всех mutation'ов (уже частично есть через Redis).

**Мониторинг**: обязательны Prometheus + Grafana + алерты на `events_lag_seconds`, `pg_stat_activity idle_in_transaction`, `pg_replication_lag`.

**БД объём**: 50-200 GB. Партицирование по месяцам для `events`, `res_log`, `messages`.

### Профиль 5 — Высокая нагрузка (50k–200k DAU, 4k–16k CCU)

Выходит за рамки «один PG». Нужен редизайн:

- Региональное шардирование PG по `galaxy_id` или `user_id % N_shards` (отдельная крупная работа).
- CDN для статики (Cloudflare / bunny.net).
- Отдельные PG-инстансы на shard.
- Managed PostgreSQL (AWS RDS, Yandex Managed PostgreSQL) имеет смысл — операционная нагрузка растёт.
- Kubernetes вместо docker-compose (автоматический failover, rollout).

Ориентировочно **€800-2000/мес** или $2000-5000/мес на cloud-провайдерах. Без редизайна кода на этот профиль не переходить.

### Профиль 6 — >200k DAU

Не рассчитывается в этом документе. Архитектура (единая PG) и код потребуют значительных переработок. Отдельный ADR.

## Сводная таблица

| Профиль | DAU | CCU | App | DB | Redis | Цена (EU VPS) | Готовность кода |
|---|---|---|---|---|---|---|---|
| 1. Soft launch | < 500 | < 50 | 2 vCPU / 4 GB | вместе | вместе | **€4-10/мес** | Текущее состояние ок |
| 2. Ранняя | 500-2k | 40-160 | 4 vCPU / 8 GB | вместе | вместе | **€14-22/мес** | + бэкапы, + мониторинг |
| 3. Рост | 2k-10k | 160-800 | 4 vCPU / 8 GB | 4 vCPU / 16 GB | share app | **€50-80/мес** | + score-incremental, + res_log TTL, + фазы 1-2 плана 09 |
| 4. Масштаб | 10k-50k | 800-4k | 4 × (4/8) + 4 × (2/4) | 8/32 + replica | 2/4 | **€180-250/мес** | + фаза 3 плана 09, + PgBouncer, + chat Redis pub/sub |
| 5. Высокая | 50k-200k | 4k-16k | N × app | PG shards | cluster | **€800-2000/мес** | Редизайн: sharding, CDN, k8s |

## Storage-планирование

Прирост на DAU (без retention):

| DAU | Месяц | Полгода | Год |
|---|---|---|---|
| 500 | 240 MB | 1.5 GB | 3 GB |
| 2 000 | 1 GB | 6 GB | 12 GB |
| 10 000 | 5 GB | 30 GB | 60 GB |
| 50 000 | 25 GB | 150 GB | 300 GB |

**С 90-дневным retention на res_log + messages** (рекомендуется с Профиля 3):
- Прирост снижается ~в 3 раза для активных пишущих таблиц.
- 10k DAU → ~10-15 GB/год в «hot» объёме.

Backup-хранилище обычно 3× от БД (ежедневные копии × 7 + еженедельные × 4 + месячные × 6, зависит от политики).

## Сетевой трафик (egress)

HTTP JSON-ответ ~2-10 KB + gzip. На одного активного игрока:
- 0.4 req/sec × 5 KB avg = 2 KB/sec = **~170 MB/day/player**.

| DAU | Трафик/день | Трафик/мес |
|---|---|---|
| 500 | 85 GB | 2.5 TB |
| 2 000 | 340 GB | 10 TB |
| 10 000 | 1.7 TB | 50 TB |

У большинства VPS включён 20 TB/мес (Hetzner), дальше — €1/TB. У российских провайдеров (Selectel, Timeweb) трафик обычно безлимитный в рамках канала 100-200 Mbit.

**CDN для статики снижает egress на backend в 5-10 раз** — Cloudflare Free tier решает проблему статики бесплатно.

## Конкретные провайдеры (апрель 2026)

Для oxsar-nova с учётом того, что проект в РФ-сегменте (legacy на oxsar.ru):

| Провайдер | Плюсы | Минусы | Для профилей |
|---|---|---|---|
| **Hetzner Cloud (EU)** | дёшево, быстрое NVMe, 20 TB трафика | EU-лат до РФ 40-60ms | 1-4 |
| **Selectel** | РФ, managed PG, обёртки | дороже Hetzner в ~2× | 1-5 |
| **Timeweb Cloud** | РФ, простой биллинг | меньше вариантов configs | 1-3 |
| **netcup (EU)** | очень дёшево | latency, overselling | 1-2 |
| **AWS / GCP** | managed всё | дорого, vendor lock-in | 4-5 при cloud-first |
| **Yandex Cloud** | РФ, managed PG/Redis | дороже Selectel на k8s | 4-5 |

Личная рекомендация для старта: **Hetzner CX22/CPX31** до 2k DAU, затем **Selectel или Yandex Cloud** для managed PG.

## Что сделать ДО перехода между профилями

### Перед Профилем 3 (>2k DAU)
- [ ] Фазы 1-2 плана 09 (retry + observability).
- [ ] res_log retention (90 дней) — иначе БД растёт как снежный ком.
- [ ] Гибридный score recalc: incremental (есть) + batch-SQL раз в сутки + удалить 5-минутный ticker ([план 09 фаза 5.4](plans/09-event-system.md#фаза-5--закрытие-gapов-l-m-05-спринта)).
- [ ] Настроить pg_dump + WAL-архивирование на отдельное хранилище.
- [ ] CDN/Cloudflare перед frontend (Free tier достаточно).

### Перед Профилем 4 (>10k DAU)
- [ ] Фаза 3 плана 09 целиком (включая shard-by-user + partitioning events).
- [ ] PgBouncer в transaction pooling перед primary PG.
- [ ] Chat hub → Redis pub/sub (иначе при 2+ backend сообщения не долетят).
- [ ] Read-replica PG для domain-queries (building queue, fleet ETA).
- [ ] Алерты Prometheus + on-call-практика.

### Перед Профилем 5 (>50k DAU)
- [ ] Шардирование PG по galaxy или user_id (крупная отдельная работа, ADR).
- [ ] Kubernetes (опционально, для HA и rollout без downtime).
- [ ] Managed PG (SaaS) имеет смысл — operational cost растёт.

## Что НЕ нужно на старте

Частые заблуждения:

- **Kubernetes на 1k DAU** — операционная сложность > выгоды. `docker-compose` на одном VPS.
- **Managed Redis на старте** — проект fail-open на Redis, `redis:7-alpine` в docker-compose достаточно.
- **3+ backend-ноды на 2k DAU** — один воркер и один backend уложатся в <50% CPU.
- **Dedicated DB-машина на 500 DAU** — PG на общей машине с backend'ом работает отлично до 2k DAU.
- **CloudFront / AWS managed LB** до 10k DAU — Cloudflare Free + nginx на app решают те же задачи за €0.

## Ссылки

- [plans/09-event-system.md](plans/09-event-system.md) — shard-by-user, worker sizing, детали масштабирования event-loop.
- [deploy/docker-compose.prod.yml](../../deploy/docker-compose.prod.yml) — текущий prod-конфиг.
- [docs/status.md](status.md) — матрица готовности модулей.
- [oxsar-spec.txt §16](oxsar-spec.txt) — milestones (без SLO/DAU-таргетов).
