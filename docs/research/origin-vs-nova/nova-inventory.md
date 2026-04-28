# Инвентарь game-nova (Ф.2 плана 62)

**Дата сборки**: 2026-04-28
**Источники**: `projects/game-nova/`, `projects/identity/`,
БД-миграции `projects/game-nova/migrations/`.

Симметричный к `origin-inventory.md` инвентарь современной
кодовой базы game-nova (Go + React + PostgreSQL). Не дублирует
[docs/balance/analysis.md](../../balance/analysis.md) — даёт
структурный обзор для сравнения с origin.

---

## Backend: доменные пакеты `internal/` (54 пакета)

| Пакет | Строк | Файлов | Назначение |
|---|---|---|---|
| auth | 673 | 6 | HTTP-адаптер `/api/me`, password utils, JWKS, vacation mode |
| planet | 1645 | 6 | Планеты, создание, переименование, ресурсы, стартовые настройки |
| fleet | 5327 | 14 | Флот и миссии: транспорт, атака, ACS, экспедиция, stargate, шпионаж |
| battle | 757 | 2 | Боевой движок (engine.go 648 + types.go 109) |
| event | 780 | 5 | Event-loop, kinds.go, dispatcher, retry/DLQ |
| admin | 1963 | 9 | Админ-эндпоинты (users, planets, audit, dead events) |
| alliance | 1207 | 2 | Альянсы, членство, отношения (NAP/WAR/ALLY) |
| market | 1148 | 3 | Обмен ресурсов и флотов, курсы |
| repair | 908 | 3 | Ремонт + разборка повреждённых юнитов |
| goal | 997 | 8 | Единый goal engine (achievements + daily quests + tutorial) |
| score | 752 | 3 | Рейтинг, личный rank |
| artefact | 854 | 6 | Артефакты, активация, эффекты, ресинхронизация |
| building | 509 | 2 | Постройка зданий |
| research | 439 | 2 | Исследования |
| shipyard | 414 | 2 | Верфь |
| rocket | 608 | 4 | Межпланетные ракеты |
| officer | 428 | 2 | Офицеры (активация, продление) |
| message | 605 | 2 | Личные сообщения с soft-delete |
| chat | 677 | 2 | WebSocket чат с edit/delete |
| galaxy | 326 | 3 | Чтение системы галактики (15 позиций) |
| economy | 323 | 3 | Формулы производства/стоимостей/времени |
| config | 437 | 2 | Загрузка YAML-справочников |
| settings | 409 | 2 | Настройки аккаунта, удаление |
| achievement | 406 | 2 | Достижения |
| **alien** | 1662 | 7 | Чужеземцы (план 15) — Halt/Holding/Attack/HoldingAI |
| artmarket | 369 | 2 | Рынок артефактов |
| dailyquest | 425 | 2 | Ежедневные задания |
| httpx | 258 | 5 | HTTP-обвязки (логгер, error-writer, recovery) |
| scheduler | 357 | 2 | Обёртка над robfig/cron |
| aiadvisor | 448 | 3 | AI советник (LLM-интеграция) |
| requirements | 181 | 1 | Проверка пред-условий построек |
| i18n | 171 | 2 | GET /api/i18n/{lang} |
| techtree | 202 | 1 | Дерево требований с прогрессом |
| empire | 176 | 1 | Сводная таблица всех планет |
| battlestats | 183 | 1 | История боёв с фильтрами |
| galaxyevent | 200 | 2 | Галактические события |
| profession | 222 | 2 | Профессии и их бонусы |
| referral | 227 | 2 | Реферальная программа (план 59) |
| automsg | 197 | 1 | Шаблонные системные сообщения |
| search | 146 | 1 | Глобальный поиск |
| records | 146 | 1 | Топ-1 рекорды по категориям |
| health | 150 | 1 | Runtime-состояние (live/draining/ready) |
| wiki | 307 | 2 | Вики (категории, страницы, markdown) |
| universe | 108 | 1 | Реестр вселенных (configs/universes.yaml) |
| moderation | 129 | 1 | UGC-модерация (никнеймы, чат, описания) |
| notepad | 82 | 1 | Личный блокнот игрока |
| friends | 125 | 1 | Список друзей (односторонний) |
| storage | 96 | 1 | Инкапсуляция БД и кеша |
| locks | 90 | 1 | Distributed locking через advisory lock |
| repo | 66 | 1 | Обёртка над pgxpool |
| features | 162 | 2 | Feature flags |
| universeswitcher | 203 | 2 | Переключение между вселенными |
| user | 327 | 1 | Основной HTTP-адаптер профиля |

---

## Миграции (71 файл, PostgreSQL)

Группы:
- **0001-0003**: users, planets, galaxy_cells, buildings, research, ships, defense
- **0004-0005**: fleets, events, messages, chat
- **0006-0013**: market, artefacts, battle_reports, debris_fields, espionage, expedition
- **0014-0020**: achievements, officers, automsg, alliances, tutorial, ACS
- **0021-0035**: market_lots, achievement_categories, building_factors, vacation, alliance_ranks, bans
- **0036-0051**: chat edit/delete, planet settings, resource_transfers, friends, notepad, account deletion
- **0052-0065**: galaxy_events, daily_quests, goal_engine
- **0066-0071**: cleanup-миграции, nullable fields, dropped tables

---

## Таблицы PostgreSQL (~50)

`users`, `planets`, `galaxy_cells`, `buildings`, `research`, `ships`,
`defense`, `construction_queue`, `shipyard_queue`, `fleets`,
`fleet_ships`, `events`, `events_dead`, `res_log`, `messages`,
`chat_messages`, `artefacts_user`, `repair_queue`, `battle_reports`,
`debris_fields`, `espionage_reports`, `expedition_reports`,
`expedition_visits`, `artefact_offers`, `achievement_defs`,
`achievements_user`, `officer_defs`, `officer_active`, `automsg_defs`,
`automsg_sent`, `alliances`, `alliance_members`,
`alliance_applications`, `alliance_relationships`, `daily_quest_defs`,
`daily_quests`, `goal_progress`, `goal_rewards_log`, `friends`,
`user_notepad`, `account_deletion_codes`, `credit_purchases`,
`resource_transfers`, `stargate_cooldowns`, `galaxy_events`,
`admin_audit_log`, `ai_advisor_log`.

---

## Конфиги (`configs/`, 14 YAML + 2 директории)

| Файл | Описание |
|---|---|
| `artefacts.yml` | Каталог артефактов (порт из oxsar2) |
| `buildings.yml` | Баланс зданий (cost_base, cost_factor, time) |
| `defense.yml` | Характеристики защиты |
| `features.yaml` | Feature flags (goal_engine и др.) |
| `goals.yml` | Каталог целей для unified goal engine |
| `professions.yml` | Профессии и их бонусы |
| `rapidfire.yml` | RF-таблица |
| `requirements.yml` | Требования для построек |
| `research.yml` | Баланс исследований |
| `schedule.yaml` | Cron-задачи (scheduler) |
| `ships.yml` | Характеристики кораблей |
| `units.yml` | Объединённый каталог юнитов |
| `universes.yaml` | Реестр вселенных (GAMESPEED, NAME, STATUS) |
| `wiki-descriptions.yml` | Описания для вики |
| `i18n/` | Языковые файлы переводов |
| `moderation/` | Blacklist и moderation rules |

---

## Entry-points (`backend/cmd/`)

| Команда | Что запускает |
|---|---|
| `server` | HTTP/WS API (порт 8080) — REST + WebSocket чат |
| `worker` | Фоновый event-loop (отдельный процесс) |
| `tools` | Утилиты: battle-sim, i18n-audit, i18n-rename, import-datasheets, import-phrases, resync, seed, testseed, wiki-gen |

---

## API-роуты (OpenAPI 3.1.0, 80+ эндпоинтов)

### Auth
`POST /api/auth/register`, `/login`, `/refresh`

### Планеты и ресурсы
`GET/PATCH/DELETE /api/planets{,/{id},/{id}/set-home,/{id}/resource-{report,update}}`

### Строительство
`POST /api/planets/{id}/buildings`,
`GET/DELETE /api/planets/{id}/buildings/queue{,/{taskId}}`

### Исследования
`POST /api/planets/{id}/research`, `GET /api/research`

### Верфь
`POST /api/planets/{id}/shipyard`,
`GET /api/planets/{id}/shipyard/{queue,inventory}`

### Ремонт
`POST /api/planets/{id}/repair/{disassemble,repair}`,
`GET /api/planets/{id}/repair/{damaged,queue}`

### Флот
`GET/POST /api/fleet`, `POST /api/fleet/{id}/recall`,
`POST /api/stargate`, `GET /api/phalanx`

### Артефакты
`GET /api/artefacts`,
`POST /api/artefacts/{id}/{activate,deactivate,sell}`,
`GET/POST /api/artefact-market/offers{,/{id}/buy}`

### Галактика
`GET /api/galaxy/{g}/{s}`

### Квесты, события
`GET /api/daily-quests`, `POST /api/daily-quests/{id}/claim`,
`GET /api/galaxy-event`,
`POST/DELETE /api/admin/galaxy-events{,/{id}}`

### Сообщения, чат
`GET/POST /api/messages`, `GET /api/messages/unread-count`,
`DELETE/POST /api/messages/{id}{,/read}`, `WS /ws/chat`

### Отчёты
`GET /api/{battle,espionage,expedition}-reports/{id}`,
`POST /api/battle-sim`

### Боевая статистика
`GET /api/battlestats`

### Достижения, офицеры, рынок ресурсов, ракеты
`GET /api/achievements`,
`GET /api/officers`, `POST /api/officers/{key}/activate`,
`GET /api/market/rates`, `POST /api/planets/{id}/market/exchange`,
`GET /api/planets/{id}/rockets`, `POST /api/planets/{id}/rockets/launch`

### Рейтинг
`GET /api/highscore{,/me}`

### Альянсы (17 endpoint'ов — см. секцию ниже)

### Wiki, поиск, друзья, профили
`GET /api/wiki{,/{category}{,/{slug}}}`, `GET /api/search`,
`GET/POST/DELETE /api/friends{,/{id}}`, `GET /api/professions`,
`GET /api/records`

### Локализация, feature flags
`GET /api/i18n{,/{lang}}`, `GET /api/features`

### Платежи
`GET /api/payment/{packages,history}`,
`POST /api/payment/{order,webhook}`

### Админ-панель
`GET/POST /api/admin/users{,/{id}{,/role,/ban,/unban,/credit,/resources,/restore,/artefacts/{grant,{aid}}}}`,
`GET/POST /api/admin/planets{,/{id}{,/rename,/transfer}}`,
`GET /api/admin/{audit,stats,events/dead}`,
`POST /api/admin/events/dead/{id}/resurrect`

---

## Frontend (`projects/game-nova/frontends/nova/src/`)

### Features (41 вертикальный срез)

`auth`, `overview`, `buildings`, `research`, `shipyard`, `fleet`,
`galaxy`, `repair`, `messages`, `market`, `artefacts`, `artmarket`,
`rockets`, `officers`, `achievements`, `score`, **`alliance`**,
`chat`, `battle-sim`, `admin`, `planet-options`, `resource`,
`unit-info`, `payment`, `profession`, `empire`, `settings`,
`referral`, `notepad`, `search`, `techtree`, `battlestats`,
`friends`, `records`, `dailyquest`, `galaxyevent`, `alien`,
`wiki`, `universes`, `billing`.

### URL-схема

Hash-based: `#<tab>` (например `#alliance`, `#buildings`,
`#fleet`). 41+ роутов, по компоненту на feature.

### Компоненты top-level
- `AgeRating.tsx` — возрастная проверка (152-ФЗ)
- `ReportButton.tsx` — кнопка жалобы (moderation)

### API-клиент (`frontend/src/api/`)
- `types.ts` (3.6 KB) — ручные типы (User, Tokens, Planet)
- `catalog.ts` (31 KB) — каталог юнитов (BUILDINGS, MOON_BUILDINGS,
  RESEARCH, SHIPS, DEFENSE) с метаданными
- `client.ts` — HTTP-клиент

---

## События (Kind* в `internal/event/kinds.go`)

### Полный список реализованных Kind-типов

| Kind | Значение | Обработчик (file:line) | Что делает |
|---|---|---|---|
| KindBuildConstruction | 1 | `handlers.go:32` | Повышает уровень здания |
| KindResearch | 3 | `handlers.go:83` | Повышает уровень research |
| KindBuildFleet | 4 | `handlers.go:122` | Добавляет корабли к запасам |
| KindBuildDefense | 5 | `handlers.go:122` | Тот же handler что KindBuildFleet |
| KindPosition | 6 | `fleet/events.go:218` | PositionArriveHandler |
| KindTransport | 7 | `fleet/events.go:31` | ArriveHandler — списание ресурсов |
| KindColonize | 8 | `fleet/events.go` | ColonizeHandler |
| KindRecycling | 9 | `fleet/events.go:113` | RecyclingHandler — обломки в ресурсы |
| KindAttackSingle | 10 | `fleet/attack.go:61` | AttackHandler |
| KindSpy | 11 | `fleet/events.go` | SpyHandler |
| KindAttackAlliance | 12 | `fleet/acs_attack.go:39` | ACSAttackHandler |
| KindExpedition | 15 | `fleet/events.go` | ExpeditionHandler |
| KindRocketAttack | 16 | `rocket/events.go:32` | ImpactHandler |
| KindReturn | 20 | `fleet/events.go:333` | ReturnHandler |
| KindAttackDestroyMoon | 25 | `fleet/attack.go:61` | AttackHandler с moon destruction (план 20 Ф.6) |
| KindAttackAllianceDestroyMoon | 27 | `fleet/acs_attack.go:39` | ACS-вариант moon destruction |
| KindRepair | 50 | `repair/events.go:70` | RepairHandler |
| KindDisassemble | 51 | `repair/events.go:16` | DisassembleHandler |
| KindArtefactExpire | 60 | `artefact/expire.go` | ExpireEvent |
| KindArtefactDelay | 63 | `artefact/delay.go` | DelayEvent |
| KindOfficerExpire | 62 | `officer/service.go:270` | ExpireHandler |
| KindRaidWarning | 64 | `fleet/raid_warning.go` | Уведомление защитнику за 10 мин |
| KindExpirePlanet | 65 | `event/expire_planet.go:21` | Soft-удаление временной планеты |
| KindScoreRecalcAll | 70 | `score/event.go` | Ежедневный batch-пересчёт очков |

### Alien* события (план 15 Этап 3)

| Kind | Значение | file:line | Что делает | Origin-аналог |
|---|---|---|---|---|
| **KindAlienFlyUnknown** | 33 | — (нет обработчика) | задекларирован | EVENT_ALIEN_FLY_UNKNOWN ❌ |
| KindAlienHalt | 36 | `alien/holding.go:220` | HaltHandler — переход 12-24ч → HOLDING+HOLDING_AI | EVENT_ALIEN_HALT ✅ |
| KindAlienHolding | 34 | `alien/holding.go:297` | HoldingHandler — финальное сообщение, уход | EVENT_ALIEN_HOLDING ✅ |
| KindAlienAttack | 35 | `alien/alien.go:192` | AttackHandler — бой, лут 30%, grab 0.08-0.1%, drop arts | EVENT_ALIEN_ATTACK ✅ |
| **KindAlienGrabCredit** | 37 | — (вложено в Attack) | вложено в EVENT_ALIEN_ATTACK | EVENT_ALIEN_GRAB_CREDIT ⚠️ частично |
| KindAlienHoldingAI | 80 | `alien/holding.go:344` | 50/50: unloadResources (7-10%) или extractShips (1%) | EVENT_ALIEN_HOLDING_AI ⚠️ упрощён (2 действия из 8) |
| **KindAlienChangeMissionAI** | 81 | — (нет обработчика) | задекларирован | EVENT_ALIEN_CHANGE_MISSION_AI ❌ |

### Заявленные но НЕ реализованные Kind (объявлены, нет handler'а)

- KindDemolishConstruction (2)
- KindHalt (13), KindHolding (17) — legacy
- KindMoonDestruction (14)
- KindStargateTransport (28), KindStargateJump (32) — план 20 Ф.5
- KindDeliveryUnits (21), KindDeliveryResources (22)
- KindArtefactDisappear (61)

### Worker / dispatcher

**Файл**: `backend/cmd/worker/main.go:49-368`

- **Main loop**: `Run(ctx)` / `RunWithGrace(ctx, grace)` —
  `time.Ticker` с интервалом `KindBatchProcessIntervalSecond = 10s`
- **Адаптивный режим**: до `maxBatch=1000` событий за цикл
- **Fetch batch**: `SELECT FROM events WHERE state='wait' AND
  fire_at <= now() LIMIT batch FOR UPDATE SKIP LOCKED`
  — это и есть основной механизм дедупа
- **Идемпотентность**: каждый handler обязан быть идемпотентным
- **Retry**: backoff 10s → 60s → 300s, max 3 попытки, после
  `state='error'`
- **DLQ**: `events_dead` таблица; `PruneErrors()` переносит
  error-события старше 7 дней раз в сутки
- **Метрики**: `EventsQueue`, `EventsLagSec`, `EventHandlerSec`,
  `EventsProcessed` (по Kind и статусу) — обновляются каждые 15s

### Параметры (env)

| Переменная | Default | Назначение |
|---|---|---|
| `WORKER_INTERVAL` | 10s | Интервал основного тикера |
| `WORKER_BATCH` | 100 | Размер батча fetch |
| `WORKER_MAX_BATCH` | 1000 | Верхняя граница адаптивного режима |
| `WORKER_MAX_ATTEMPTS` | 3 | Максимум попыток |
| `WORKER_SHUTDOWN_GRACE` | 30s | Grace при остановке |

---

## Alliance подсистема — детально

### Файлы

- `backend/internal/alliance/service.go` (814 строк) — бизнес-логика
- `backend/internal/alliance/handler.go` (393 строки) — HTTP-адаптер
- `backend/internal/alliance/service_test.go` (107 строк)

### Endpoints (17)

| HTTP | URL | Действие |
|---|---|---|
| GET | `/api/alliances` | List (top-50) |
| GET | `/api/alliances/{id}` | Get с членами |
| GET | `/api/alliances/me` | Мой альянс |
| POST | `/api/alliances` | Create |
| POST | `/api/alliances/{id}/join` | Join (или заявка) |
| PATCH | `/api/alliances/{id}/open` | SetOpen |
| GET | `/api/alliances/{id}/applications` | Список заявок |
| POST | `/api/alliances/applications/{appID}/approve` | Approve |
| DELETE | `/api/alliances/applications/{appID}` | Reject |
| POST | `/api/alliances/leave` | Leave |
| DELETE | `/api/alliances/{id}` | Disband |
| PATCH | `/api/alliances/{id}/members/{userID}/rank` | SetMemberRank |
| GET | `/api/alliances/{id}/relations` | GetRelations |
| PUT | `/api/alliances/{id}/relations/{target_id}` | ProposeRelation |
| POST | `/api/alliances/{id}/relations/{initiator_id}/accept` | AcceptRelation |
| DELETE | `/api/alliances/{id}/relations/{initiator_id}` | RejectRelation |

### Схема БД

```sql
-- alliances (0017)
id uuid PK, tag text UNIQUE (3-5 символов), name text UNIQUE,
description text, owner_id uuid → users(id), created_at timestamptz

-- alliance_members (0017)
alliance_id, user_id PK, rank ('owner'|'member'),
rank_name text (произвольный, добавлен 0034), joined_at

-- alliance_applications (0024)
id, alliance_id, user_id, message, created_at,
UNIQUE (alliance_id, user_id)

-- alliance_relationships (0028)
alliance_id, target_alliance_id PK,
relation ENUM ('nap','war','ally'),
status text ('pending'|'active', добавлено 0031),
set_at timestamptz,
CHECK (alliance_id <> target_alliance_id)

-- denormalization (0017)
ALTER users ADD alliance_id uuid → alliances(id) ON DELETE SET NULL
```

### Что в origin есть, в nova **нет** (предварительно)

- Три текстовых описания (external/internal/application) → в nova
  одно `description`
- Передача лидерства (`abandonAlly`) → в nova нет endpoint'а
- Гранулярные права рангов (битовые) → в nova `rank` = enum
- Global mail членам → блокирован планом 57 (mail-service)
- BBCode → в nova не нужно (TipTap)

---

## Battle подсистема

- `backend/internal/battle/engine.go` (648 строк) — симуляция боёв
- `backend/internal/battle/types.go` (109 строк)
- Используется в `fleet/attack.go` и `fleet/acs_attack.go`
- **Battle-sim CLI отсутствует** в `cmd/tools/` (упоминание в memory
  устарело — нужно верифицировать). UI-симулятор —
  `frontend/src/features/battle-sim/`

---

## Domain entities

### User
`projects/game-nova/backend/internal/auth/`,
`projects/game-nova/backend/internal/user/`
```go
type User struct {
    ID, Username, Email, PasswordHash string
    // vacation, ref_code, alliance_id, ...
}
```

### Planet (`internal/planet/model.go`)
```go
type Planet struct {
    ID, UserID, Name string
    IsMoon bool
    Galaxy, System, Position, Diameter, UsedFields, MaxFields int
    Metal, Silicon, Hydrogen float64
    MetalPerSec, SiliconPerSec, HydrogenPerSec float64
    MetalCap, SiliconCap, HydrogenCap float64
    EnergyProd, EnergyCons, EnergyRemaining float64
    BuildFactor, ResearchFactor, ProduceFactor, EnergyFactor, StorageFactor float64
    LastResUpdate time.Time
}
```

### Building / Research
Уровни в отдельных таблицах `buildings` (planet × type → level)
и `research` (user × type → level).

### Fleet / Mission (`internal/fleet/`)
Типы миссий: attack, transport, expedition, spy, colonize,
stargate, ACSAttack.

```go
type Mission struct {
    ID, FleetID string
    Type string
    SourcePlanetID, TargetPlanetID string
    TargetCoords (galaxy, system, position) int
    Units map[int]int
    Payload (metal, silicon, hydrogen) float64
    ArrivalAt time.Time
}
```

---

## Identity-service (`projects/identity/backend/`)

Отдельный микросервис.

### Структура
- `internal/identitysvc/` (8 файлов, 76 KB):
  - `service.go` (17 KB) — Register, Login, JWT, RBAC
  - `handler.go` (11 KB) — `/auth/{register,login,refresh,logout}`
  - `rbac.go` (13 KB) — роли и permissions
  - `rbac_handler.go` (9.9 KB) — HTTP endpoints для RBAC
- `internal/auth/`:
  - `password.go` — bcrypt/scrypt
- `pkg/jwtrs/` — Issuer + Verifier (RS256/ES256, JWKS)

### Endpoints
- `POST /auth/register` — регистрация (с consent_accepted, terms_accepted — план 44, 47)
- `POST /auth/login` — access + refresh токены
- `POST /auth/refresh`
- `POST /auth/logout` (JTI Blacklist в Redis)
- `GET /.well-known/jwks.json` — план 52

### Интеграция с game-nova
- `game-nova/internal/auth/jwksloader.go` (62 строки)
- `game-nova/internal/auth/middleware.go` — извлечение userID из JWT

---

## Универсальное знание для сравнения с origin

| Элемент | Origin | Nova |
|---|---|---|
| Стек | PHP 8.3 + nginx + MySQL 5.7 + memcached | Go 1.23 + PostgreSQL 16 + Redis |
| Identity | inline `na_user`+`na_password` MD5 | отдельный identity-service, JWT/JWKS |
| Event-loop | EventHandler.class.php (3573 стр) с `case`-веткой | dispatcher по `Kind` (24 реализованных) |
| Дедуп | `NS::isFirstRun` memcached TTL=2s | PostgreSQL advisory lock + `FOR UPDATE SKIP LOCKED` |
| Балансовые формулы | DSL-строки в `na_construction` (eval) | Go-формулы + YAML-числа в `configs/*.yml` |
| Auth-схема | sessions + IPCheck | JWT (access+refresh) с blacklist |
| Шаблоны | Smarty .tpl (125 файлов) | React + TS strict, hash-based routing |
| WebSocket | нет (memcached polling) | `/ws/chat` (план 32) |
| Admin | через permissions в БД | отдельный `internal/admin/` + audit_log |
| RBAC | usergroups + permissions (legacy) | план 51 (ролевая модель) |

---

## References

- `projects/game-nova/backend/internal/` — все доменные пакеты
- `projects/game-nova/migrations/` — 71 миграция
- `projects/game-nova/api/openapi.yaml` — контракт REST API
- `projects/game-nova/frontends/nova/src/features/` — 41 frontend-feature
- [docs/balance/analysis.md](../../balance/analysis.md) — анализ
  балансных формул nova
- [docs/status.md](../../status.md) — матрица готовности модулей
