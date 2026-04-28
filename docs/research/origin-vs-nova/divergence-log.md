# Divergence Log: origin ↔ nova

**Дата сборки**: 2026-04-28
**Контекст**: артефакт плана 62 Ф.3.5. Журнал технических
расхождений между game-origin (PHP/MySQL) и game-nova
(Go/PostgreSQL). Каждая запись D-NNN — кандидат на вынос
параметра в `configs/balance/legacy.yaml` или новую механику в
nova-backend для поддержки legacy-вселенной.

**Принципы** (из плана 62):
- Каждая запись формулируется как «какой ключ конфигурации/код-путь
  нужен в nova», а не «вернуть число к origin».
- Для современных вселенных (uni01/uni02) баланс nova **не
  меняется**. Параметризуется только для legacy01.
- Тривиальные имена (`user_id` vs `userid`) НЕ записываются.
- Семантически идентичные поля с разными именами решаются на этапе
  реализации origin-фронта (он сразу пишется на nova-API).

**Цвета**:
- 🟡 Жёлтый: правка конфига или флаг (≤50 строк)
- 🟠 Оранжевый: фича только в origin (нужно реализовать в nova)
- 🔴 Красный: формула/механика расходится (нужен code-path)
- 🟣 Фиолетовый: тихая семантика (одинаковые имена, разная семантика)

---

## Категория «домен» (D-001 .. D-025)

### D-001. Множественность очков (dm_points / points / max_points)

✅ **ЗАКРЫТО** (план 69, 2026-04-28). Поле `users.max_points numeric(20,4)
NOT NULL DEFAULT 0` добавлено миграцией 0072. `internal/score/service.go`
обновляет `max_points = GREATEST(max_points, total)` при каждом
`RecalcUser` (исторический пик, никогда не убывает).
Категории `dm_points` / `be_points` / `of_points` — **отказ** (YAGNI:
6 текущих категорий u/r/b/a/e_points достаточно для UX).

- **Категория**: домен
- **Цвет**: 🟣
- **Origin**:
  - Файл: `migrations/001_schema.sql` (DDL `na_user`)
  - Что: 9 разных полей очков — `dm_points`, `points`, `max_points`,
    `u_points`, `r_points`, `b_points`, `e_points`, `be_points`,
    `of_points`
  - Конкретно: `dm_points` — внутренний рейтинг, `points` —
    отображаемый, `max_points` — исторический пик, остальные —
    разрезы (units / research / buildings / espionage /
    battle-experience / officer)
- **Nova**:
  - Файл: миграция `0001_users.sql` + 0066+
  - Что: `points`, `u_points`, `r_points`, `b_points`, `a_points`,
    `e_points` (numeric)
  - Конкретно: 6 полей, нет `dm_points`, нет `max_points`,
    нет `be_points`, нет `of_points`
- **Разница**: Origin различает «внутренний» рейтинг от
  отображаемого; nova хранит только текущие очки.
  `max_points` теряется при миграции (для топа «когда-либо»).
- **Как сделать nova-конфигурируемой**: добавить миграцию
  `users.max_points DOUBLE` в nova; для legacy01 заполняется при
  каждом росте `points`. Опционально `dm_points`, `be_points`,
  `of_points` — только если решим выносить отдельные рейтинги.
- **Объём правки nova**: миграция (5 строк) + триггер на UPDATE +
  legacy.yaml; 1-2 дня
- **Риски**: тихая регрессия рейтинга при миграции уни01/uni02 на
  новую модель — добавление nullable поля безопасно

### D-002. Vacation mode (umode/umodemin vs vacation_since/last_end)

- **Категория**: домен
- **Цвет**: 🟣
- **Origin**:
  - Файл: `na_user.umode TINYINT(1)`, `umodemin INT(10)`
  - Что: `umode` = флаг включён/нет, `umodemin` = unix timestamp
    окончания vacation
- **Nova**:
  - Файл: миграция `0045_vacation.sql`
  - Что: `vacation_since timestamptz`, `vacation_last_end timestamptz`
  - Конкретно: «с какого момента в отпуске» + «когда последний раз
    закончился»
- **Разница**: Полностью противоположная семантика. Origin: «до
  когда»; nova: «с когда + когда закончился прошлый». При миграции
  тихо потеряется `umodemin` или будет интерпретирован
  неправильно.
- **Как сделать nova-конфигурируемой**: либо добавить
  `vacation_until timestamptz` (синоним `umodemin`), либо явный
  трансформер при миграции уни-данных (legacy01 умеет читать оба).
  Предпочтительно — добавить поле, оставить оба представления.
- **Объём**: миграция + helper; 1 день
- **Риски**: 🟣 ТИХАЯ СЕМАНТИКА — при миграции игроков из origin
  в nova vacation-таймеры могут сбиться

### D-003. Account deletion (delete timestamp vs deletion codes)

✅ **ЗАКРЫТО** (план 69 Ф.0, 2026-04-28; основная реализация — миграция
0051). Архитектурно лучше плана: `account_deletion_codes (user_id,
code_hash, expires_at, attempts)` — email-confirm flow вместо
простого `scheduled_at`. Соответствует roadmap-report «Часть III»
для всех вселенных.

- **Категория**: домен
- **Цвет**: 🔴
- **Origin**:
  - Файл: `na_user.delete INT(10)` (unix timestamp удаления)
  - Что: простое отложенное удаление по таймстампу
- **Nova**:
  - Файл: миграция `0051_account_deletion_codes.sql`
  - Что: `account_deletion_codes (user_id, code_hash, expires_at,
    attempts)`
  - Конкретно: код подтверждения по email, multi-step flow
- **Разница**: Совершенно разные механики (auto-delete vs
  email-confirm).
- **Как сделать nova-конфигурируемой**: добавить
  `users.account_deletion_scheduled_at timestamptz nullable`. В
  legacy01 поведение — старое (без email-кода). В uni01/uni02 —
  через коды.
- **Объём**: миграция + ветка в handler по
  `universe.deletion_flow`; 2-3 дня
- **Риски**: разные UX для разных вселенных, потребитель должен
  читать `universe.deletion_flow`

### D-004. Protection time (новичковая защита)

✅ **ЗАКРЫТО** (план 69, 2026-04-28). Поле `users.protected_until_at
timestamptz` добавлено миграцией 0072. `internal/fleet/attack.go`
расширен: проверка защиты срабатывает по global protectionPeriod ИЛИ
per-user `protected_until_at` ИЛИ `is_observer = true` (флот возвращается
без боя). Spy / acs / moon-destruction handlers НЕ обновлены — это
существующая дыра pre-69 (защита новичков отсутствовала и там); вне
scope плана 69, отдельный план при необходимости.

- **Категория**: домен / механика
- **Цвет**: 🔴
- **Origin**: `na_user.protection_time INT(10)` — unix таймстамп до
  которого новичок защищён от атак
- **Nova**: ОТСУТСТВУЕТ
- **Разница**: Нет механики защиты новичков
- **Как сделать nova-конфигурируемой**: добавить
  `users.protected_until timestamptz` + проверки в `fleet/attack.go`
- **Объём**: миграция (5 строк) + проверка в attack-логике;
  3-5 дней
- **Риски**: нужен фоновый процесс снятия защиты или проверка по
  каждому attack

### D-005. Observer flag (привилегированный статус)

✅ **ЗАКРЫТО** (план 69, 2026-04-28). Поле `users.is_observer boolean
NOT NULL DEFAULT false` добавлено миграцией 0072 (домен-флаг, не
RBAC-роль — `users.role` удалён ранее миграцией 0070, роли мигрировали
в identity-сервис). Фильтр `is_observer = false` добавлен везде, где
был фильтр `umode = false` для публичных рейтингов: `score.Top`,
`score.TopAlliances`, `score.PlayerRank`, `records.topScore`,
`score/handler.go` online-статистика, `galaxy/repository.go` рейтинг
в галактика-обзоре. Не добавлен в `score.RecalcAll` (observer всё
равно играет, очки считаем) и `score.VacationPlayers`.

- **Категория**: домен
- **Цвет**: 🔴
- **Origin**: `na_user.observer TINYINT(1)`
- **Nova**: `users.role ENUM ('player'|'support'|'admin'|'superadmin')`
- **Разница**: Observer как флаг отдельной семантики (наблюдатель,
  не участвует в рейтинге)
- **Как сделать nova-конфигурируемой**: либо ввести роль
  `'observer'` в enum, либо добавить boolean `is_observer`
- **Объём**: миграция + ветки в score/highscore queries; 2 дня
- **Риски**: фильтрация observer из рейтинга в нескольких местах

### D-006. Координаты (umi vs galaxy/system/position)

- **Категория**: домен
- **Цвет**: 🟣
- **Origin**: `na_planet.umi FLOAT` (формула:
  `galaxy*1000000+system*1000+position`)
- **Nova**: явные `galaxy INT, system INT, position INT` с CHECK
  constraints
- **Разница**: Float-encoding имеет precision errors при больших
  galaxy. В origin все запросы используют умножение/деление; в
  nova — индекс по 3 полям.
- **Как сделать nova-конфигурируемой**: helper-функция
  `decodeUmi(float) → (g, s, p)` для миграции legacy-данных. В
  nova схема не меняется.
- **Объём**: helper в legacy migration script; 1 день
- **Риски**: 🟣 ТИХАЯ — при миграции legacy01 округление float

### D-007. UI customization (templatepackage/theme/skin)

🚫 **ОТКАЗ** (план 69, 2026-04-28). YAGNI: у nova одна тема, у origin
одна тема (pixel-perfect клон `standard`). Свободный выбор тем — отдельным
планом если понадобится. В legacy-PHP-схеме поля присутствуют,
но в первой итерации oxsar-nova не реплицируются.

- **Категория**: домен
- **Цвет**: 🔴
- **Origin**: `templatepackage`, `imagepackage`, `theme`,
  `user_bg_style`, `user_table_style`, `skin_type` (varbinary в
  na_user)
- **Nova**: ОТСУТСТВУЕТ
- **Разница**: Origin позволяет каждому игроку сменить тему,
  фоны, стиль таблиц.
- **Как сделать nova-конфигурируемой**: добавить
  `users.ui_theme TEXT` + `ui_pack TEXT` (с whitelist в
  `configs/themes.yml`)
- **Объём**: миграция + handler + frontend выбор темы; 1 неделя
- **Риски**: ассеты тем для legacy01 — отдельный набор
  (`projects/origin-frontend/themes/`)

### D-008. Profession (профессия игрока)

✅ **ЗАКРЫТО** (миграция 0046, до плана 69). Поле `users.profession TEXT
NOT NULL DEFAULT 'none'` + `users.profession_changed_at TIMESTAMPTZ`
+ полный handler с 14-day cooldown (`internal/profession/service.go`,
`ErrChangeTooSoon`). План 69 Ф.0 (дельта-аудит) подтвердил: уже
закрыто, в Ф.1 не включаем.

- **Категория**: домен / механика
- **Цвет**: 🔴
- **Origin**: `na_user.profession TINYINT(3)` + `prof_time INT(10)`
- **Nova**: `users.profession TEXT` (миграция 0046), но
  `prof_time` потеряна
- **Разница**: Origin хранит время последней смены — нужно для
  rate-limit (мин. интервал 14 дней + стоимость 1000 кредитов)
- **Как сделать nova-конфигурируемой**: добавить
  `users.profession_changed_at timestamptz` в миграцию 0046
- **Объём**: миграция + проверка в handler; 1-2 дня
- **Риски**: legacy01 без таймера — игроки могут менять свободно

### D-009. Email/password activation tokens

- **Категория**: домен / инфра
- **Цвет**: 🔴
- **Origin**: `na_user.activation`, `password_activation`,
  `email_activation` (varbinary коды)
- **Nova**: email и password управляются identity-service
  (план 36/52); токены не в users
- **Разница**: Origin self-contained, nova зависит от identity
- **Как сделать nova-конфигурируемой**: для legacy01 — все вызовы
  через identity API. Если identity не поддерживает определённый
  flow — расширить identity, не возвращать токены в game-БД.
- **Объём**: уточнить, какие flow legacy не покрыты identity;
  1 день анализ
- **Риски**: блокирует если identity не реализует часть legacy
  flow

### D-010. Last activity (last vs last_seen)

- **Категория**: домен
- **Цвет**: 🟡
- **Origin**: `na_user.last INT(10)` — любая активность
- **Nova**: `users.last_seen TIMESTAMPTZ` — page visits
- **Разница**: гранулярность может отличаться (origin записывает
  ЛЮБУЮ активность, nova — только страницы)
- **Как сделать nova-конфигурируемой**: документировать в схеме
  семантику; синхронизировать при миграции
- **Объём**: 0.5 дня доку
- **Риски**: alien AI требует «активен в последние 30 мин» — должен
  работать одинаково

### D-011. Battle reports (нормализованные vs JSONB)

- **Категория**: домен / API
- **Цвет**: 🟣
- **Origin**: `na_assault` (44 поля) + `na_assaultparticipant`
  (отдельная таблица)
- **Nova**: `battle_reports (report JSONB + 8 summary полей)`
- **Разница**: Origin нормализованный, nova денормализованный —
  при миграции данных нужен трансформер
- **Как сделать nova-конфигурируемой**: миграционный скрипт
  origin → nova. В nova схема не меняется. Опционально добавить
  `battle_report_participants` для аналитики, если нужно.
- **Объём**: трансформер 1-2 дня
- **Риски**: legacy-отчёты могут не отображаться корректно если
  payload отличается

### D-012. Espionage reports (events.mode=11 vs отдельная таблица)

- **Категория**: домен / API
- **Цвет**: 🟡
- **Origin**: SPY-отчёты как event mode=11 в `na_events.data
  MEDIUMBLOB`
- **Nova**: `espionage_reports` отдельная таблица с `ratio`,
  `probes`, `report JSONB`
- **Разница**: одна таблица событий с blob vs нормализованная
- **Как сделать**: миграционный скрипт парсит origin
  `mode=11.data` → nova `espionage_reports`
- **Объём**: 1-2 дня
- **Риски**: PHP-сериализация blob нужна расшифровка

### D-013. Event-loop (mode-based vs kind-based)

- **Категория**: event-loop
- **Цвет**: 🟣
- **Origin**: `na_events (mode INT(2), processed TINYINT,
  processed_mode INT, data MEDIUMBLOB)` — статус через 4 поля
- **Nova**: `events (kind INT, state ENUM('wait'|'start'|'ok'|'error'),
  payload JSONB)`
- **Разница**: разные state-машины. У origin
  `EVENT_PROCESSED_WAIT/START/ERROR/OK`, у nova `state ENUM`. Маппинг
  есть, но нужен трансформер.
- **Как сделать**: маппинг таблица + миграционный скрипт.
  В nova схема не меняется.
- **Объём**: 2-3 дня
- **Риски**: при миграции running events — race-условия

### D-014. Alliance ranks (битовые права vs enum)

- **Категория**: домен / механика
- **Цвет**: 🔴
- **Origin**: `na_allyrank` с битовыми правами
  (`CAN_SEE_MEMBERLIST`, `CAN_MANAGE`, `CAN_BAN_MEMBER`,
  `CAN_DIPLOMACY`, `CAN_GLOBAL_MAIL`, ...)
- **Nova**: `alliance_members.rank TEXT` ('owner'|'member' +
  `rank_name` свободная строка из 0034)
- **Разница**: Origin полная система разрешений, nova
  минималистична
- **Как сделать nova-конфигурируемой**: новая таблица
  `alliance_ranks (id, alliance_id, name, permissions JSONB,
  position INT)`. `alliance_members.rank_id` FK. Endpoint'ы:
  `GET/POST/PUT/DELETE /api/alliances/{id}/ranks`.
- **Объём**: 1-2 недели (миграция + 4 endpoint'а + frontend
  manage_ranks.tpl-аналог)
- **Риски**: совместимость с существующим `rank='owner'` — оставить
  как builtin role

### D-015. Officer units (юниты vs subscription)

- **Категория**: домен / механика
- **Цвет**: 🔴
- **Origin**: `na_officer (of_1, of_2, of_3, of_4 = юнит-IDs,
  of_level, of_points)` — офицеры как боевые юниты на планете
- **Nova**: `officer_defs (key, ...)` + `officer_active (user_id,
  expires_at)` — офицеры как временные эффекты-buff'ы
- **Разница**: концептуально разные модели. Origin — **юниты**,
  которые могут участвовать в бою; nova — **подписка** на бонус.
- **Как сделать nova-конфигурируемой**: документировать как
  **намеренное расхождение**. Для legacy01 — реализовать
  origin-стиль (отдельный пакет `internal/legacy/officers/` под
  флагом). Альтернатива: оставить nova-модель и для legacy01.
  **Рекомендация**: оставить nova-модель — legacy-механика
  устаревшая.
- **Объём**: 0 (если оставляем nova) или 2 недели
- **Риски**: legacy-игроки не получат привычные офицер-юниты

### D-016. Planet teleport rate-limiting

✅ **ЗАКРЫТО** (план 69, 2026-04-28). Поле `users.last_planet_teleport_at
timestamptz` добавлено миграцией 0072. Helper для проверки cooldown'а
будет добавлен в плане 72 (origin-фронт смены home-планеты) — поле
готово к использованию. **Не дублирует** `stargate_cooldowns`
(миграция 0062): stargate — это прыжок флота между лунами per
planet_id; teleport — смена «домашней» планеты per user_id.

- **Категория**: домен
- **Цвет**: 🟡
- **Origin**: `na_user.planet_teleport_time INT(11)`
- **Nova**: ОТСУТСТВУЕТ (т.к. сама механика TELEPORT_PLANET
  отсутствует)
- **Разница**: связан с D-NNN-TELEPORT
- **Как сделать**: при добавлении teleport-механики (см.
  U-009) — `users.last_planet_teleport_at timestamptz`
- **Объём**: ~5 дней (вместе с механикой)
- **Риски**: rate-limit нужен — иначе спам теле-портации

### D-017. Achievements (32 поля условий vs 3 поля)

- **Категория**: домен / механика
- **Цвет**: 🔴
- **Origin**: `na_achievement_datasheet (req_points,
  req_u_points, ..., bonus_metal, bonus_*_unit)` — 32 поля условий
  и наград, ~100+ ачивок
- **Nova**: `achievement_defs (key, title, description, points)` +
  goal engine — 5 базовых ачивок
- **Разница**: nova ачивки минималистичны, origin богатые
- **Как сделать nova-конфигурируемой**: расширить
  `achievement_defs (requirements JSONB, rewards JSONB)` или
  использовать существующий goal engine (план 65) — добавить
  legacy-цели в `configs/goals.yml`.
- **Объём**: 1-2 недели (загрузка origin данных + UI)
- **Риски**: рефакторинг goal engine может сломать существующие

### D-018. Asteroid slots (creature creation limit)

- **Категория**: домен / механика
- **Цвет**: 🔴
- **Origin**: `na_user.asteroid INT(10)` — лимит на создание
  астероидов
- **Nova**: ОТСУТСТВУЕТ
- **Разница**: legacy-механика отсутствует
- **Как сделать**: `users.available_asteroid_slots INT` если
  механика будет
- **Объём**: ~3 дня (если делаем механику)
- **Риски**: новая механика — нужна декомпозиция

### D-019. Home planet (hp + curplanet vs только cur_planet_id)

🚫 **ОТКАЗ** (план 69, 2026-04-28). Не вводим `home_planet_id` ни в
каком виде. Обоснование (вариант «в»): в nova нет UI/UX концепции
«домашней планеты» — игрок переключается через `cur_planet_id` /
session-state. R10 (users cross-universe) + per-universe planets
создали бы конфликт. Когда понадобится cross-universe state
(preferred-планета per вселенная) — отдельный план с
`user_universe_state`, не сейчас.

- **Категория**: домен
- **Цвет**: 🟡
- **Origin**: `hp INT(10)` (главная планета — куда возвращаются
  миссии без явного origin), `curplanet INT(10)` (просмотр)
- **Nova**: `cur_planet_id UUID` (только просмотр), home_planet =
  первая по `created_at`
- **Разница**: nova отбрасывает «мою главную» как отдельное
  понятие
- **Как сделать**: `users.home_planet_id UUID` + endpoint
  `POST /api/me/set-home/{planet_id}`
- **Объём**: миграция + endpoint; 1 день

### D-020. Chat read tracking (last_chat / last_chatally / chat_languageid)

✅ **ЗАКРЫТО** (план 69, 2026-04-28). Поля
`users.last_global_chat_read_at`, `users.last_ally_chat_read_at`
(timestamptz, nullable) добавлены миграцией 0072. Endpoints
`POST /api/chat/{kind}/read` и `GET /api/chat/{kind}/unread` в
`internal/chat/handler.go` (kind = global/ally). `chat_languageid` —
**отказ** (YAGNI: `users.language` уже есть с миграции 0001, отдельный
язык чата избыточен).

- **Категория**: домен
- **Цвет**: 🔴
- **Origin**: `last_chat`, `last_chatally`, `chat_languageid`
- **Nova**: ОТСУТСТВУЮТ в users
- **Разница**: nova-чат не помечает «прочитано до сюда»
- **Как сделать**: миграция users +
  `last_global_chat_read_at timestamptz`,
  `last_ally_chat_read_at timestamptz`, `chat_language TEXT`
- **Объём**: миграция + frontend updates; 2-3 дня

### D-021. Race (раса персонажа)

🚫 **ОТКАЗ** (план 69, 2026-04-28). Мёртвое поле в legacy-PHP-схеме:
`na_user.race` есть с дефолтом 1, но в коде origin не используется
(архитектурная заглушка без поведения). См. roadmap-report «Часть V».
Не реплицируем.

- **Категория**: домен / механика
- **Цвет**: 🔴
- **Origin**: `na_user.race TINYINT(3) DEFAULT 1`
- **Nova**: ОТСУТСТВУЕТ
- **Разница**: legacy-фича выбора расы
- **Как сделать**: `users.race TEXT` + `configs/races.yml`
- **Объём**: миграция + UI; ~3 дня (без бонусов)
- **Риски**: если есть game-effects от расы — больше работы

### D-022. Building production override (prod_factor)

✅ **ЗАКРЫТО** (план 64, 2026-04-28). Параметризация production
коэффициентов через `internal/balance.Globals` теперь возможна. Полная
per-planet `production_factor` колонка — отдельный план (когда понадобится
конкретно эта механика; loader-инфраструктура готова).

- **Категория**: домен / формула
- **Цвет**: 🟡
- **Origin**: `na_building2planet.prod_factor INT(3) DEFAULT 100`
  — per-planet процент производства
- **Nova**: `buildings` без production_factor
- **Разница**: нет per-planet override
- **Как сделать**: `buildings.production_factor REAL DEFAULT 1.0`,
  читать в economy
- **Объём**: миграция + формула + UI; 1-2 дня

### D-023. Event payload serialization (PHP serialized blob vs JSONB)

- **Категория**: event-loop
- **Цвет**: 🟡
- **Origin**: `na_events.data MEDIUMBLOB` (PHP serialized)
- **Nova**: `events.payload JSONB`
- **Разница**: при миграции running events нужен трансформер
- **Как сделать**: миграционный PHP-скрипт
  `unserialize → json_encode`
- **Объём**: 1 день
- **Риски**: PHP serialize-формат с custom-classами может не
  декодироваться

### D-024. Event chains (parent_eventid, ally_eventid)

- **Категория**: event-loop
- **Цвет**: 🟡
- **Origin**: `na_events.parent_eventid`, `ally_eventid`
- **Nova**: ОТСУТСТВУЮТ явные ссылки
- **Разница**: в nova цепочки только в коде, не в БД
- **Как сделать**: `events.parent_event_id UUID FK`,
  `events.related_event_ids JSONB` для ACS
- **Объём**: миграция + использование в alien chains; 2-3 дня

### D-025. User agreement tracking

- **Категория**: домен / инфра
- **Цвет**: 🟡
- **Origin**: `na_user.user_agreement_read INT(10)`
- **Nova**: ОТСУТСТВУЕТ как поле (план 47 — terms_accepted в
  identity)
- **Разница**: для legacy01 нужно знать дату принятия last
  agreement
- **Как сделать**: использовать identity-service `terms_accepted_at`
  (план 47) — легко
- **Объём**: 0 если identity покрывает

---

## Категория «формула / баланс» (D-026 .. D-030)

### D-026. Источник истины формул (БД-строки vs YAML-числа)

✅ **ЗАКРЫТО** (план 64, 2026-04-28). origin balance теперь живёт в
`configs/balance/origin.yaml` (override-файл, ~395 строк автогенерации).
Импортёр `cmd/tools/import-legacy-balance` парсит DSL-формулы из live
docker-mysql-1, статика предвычисляется в таблицы (1..50 уровней),
динамика (`{temp}`, `{tech=N}`) реализуется в Go в `internal/origin/
economy/` через bundle.Globals. Pixel-perfect совпадение с PHP eval()
проверено golden-тестами (verify 2026-04-28).

- **Категория**: формула
- **Цвет**: 🟣
- **Origin**: `na_construction.prod_*, cons_*, charge_*` —
  varbinary(255) DSL-строки парсятся PHP `eval()`. Источник
  истины — **продакшн БД**.
- **Nova**: `configs/buildings.yml`, `units.yml`, формулы в
  Go-коде. Источник — **репо**.
- **Разница**: фундаментально разное хранение баланса.
- **Как сделать**: предвычислить формулы origin по уровням и
  сохранить в `configs/balance/legacy.yaml` как числа. Динамику
  prod_* (с `{temp}`, `{tech=N}`) реализовать в Go в
  `internal/legacy/economy/` под флагом
  `universe.balance_profile = legacy`. **B1+B3 гибрид**.
- **Объём**: 2 недели (импорт формул + Go-формулы для динамики)
- **Риски**: расхождение в округлении PHP `eval round()` vs Go
  `math.Round` — несколько единиц на больших уровнях
- **Связь**: см. `formula-dsl.md`

### D-027. RF-таблица (rapidfire) — алиен-юниты

✅ **ЗАКРЫТО** (план 64, 2026-04-28). RF алиен/спец-юнитов
(200, 201, 202, 203, 204, 102 Lancer, 325 Shadow, 348 Transmitter,
352 Transplantator, 353 Collector) импортированы из oxsar2-mysql-1
legacy в `configs/rapidfire.yml` (R0-исключение: применяется во всех
вселенных, +38 строк). Debug-юнит 358 (×900 ко всему) и 348
(TRANSMITTER, не в nova) исключены. Существующие nova-RF числа не
тронуты (R0 inplace-merge).

- **Категория**: формула
- **Цвет**: 🟠
- **Origin**: `na_rapidfire` содержит entries для UNIT_A_*
  (200-204) — RF алиенов между собой и против игрока
- **Nova**: `configs/rapidfire.yml` — нет UNIT_A_* (см. также
  game-reference.md § «Параметры юнитов из БД легаси»)
- **Разница**: в nova нет алиен-кораблей вообще
- **Как сделать**: добавить алиен-юниты в `configs/units.yml` и
  RF в `configs/rapidfire.yml` под флагом
  `universe.legacy_alien_units = true`
- **Объём**: 1-2 дня (числа известны из game-reference)
- **Связь**: D-028, alien-ai-comparison.md

### D-028. Юниты UNIT_A_* (id 200-204) и Lancer/Shadow/etc

✅ **ЗАКРЫТО** (план 64, 2026-04-28). Алиен-флот UNIT_A_* (200-204)
+ planet shields (354/355) уже были в nova `configs/units.yml` через
план 22 (с именами `unit_a_corvette..torpedocarier`). Lancer (102) и
Shadow (325) — там же с balance из плана 22+ADR-0007/0008 (отличается
от origin для modern). Origin override восстанавливает legacy-числа
для них (configs/balance/origin.yaml::ships). Новые добавки в дефолт:
`ship_transplantator` (352), `ship_collector` (353), `armored_terran`
(358) — отсутствовали в nova, теперь в дефолтных units.yml +
ships.yml (R0-исключение для всех вселенных).

- **Категория**: формула / домен
- **Цвет**: 🟠
- **Origin**: `na_ship_datasheet` содержит unitid 200, 201, 202,
  203, 204 (Alien*), 102 (Lancer Ship), 325 (Shadow Ship), 352
  (Ship Transplantator), 353 (Ship Collector), 354 (Small Planet
  Shield), 355 (Large Planet Shield), 358 (Armored Terran)
- **Nova**: `configs/ships.yml` содержит только базовые юниты
- **Разница**: ~10 «специальных» legacy-юнитов отсутствуют
- **Как сделать**: добавить в `configs/units.yml` секцию
  `legacy:` с этими юнитами под флагом
- **Объём**: 2-3 дня (числа в game-reference.md)
- **Связь**: D-027

### D-029. Температура влияет на производство водорода

✅ **ЗАКРЫТО** (план 64, 2026-04-28). Verify против live origin
docker-mysql-1: `na_construction.prod_hydrogen` HYDROGEN_LAB =
`floor(10 * {level} * pow(1.1+{tech=25}*0.0008, {level}) *
(-0.002*{temp} + 1.28))`. nova `internal/economy/HydrogenLabProdHydrogen`
УЖЕ реализует это с того же plan-03 (origin-формулы и nova-формулы
совпали по факту). План 64 параметризовал коэффициенты через
`balance.Globals.HydrogenTempCoefficient/HydrogenTempIntercept` —
теперь origin override может изменить их, не меняя кода. Закрыто
golden-тестом TestGolden_HydrogenLabProduction (level={1,5,10,20,30}
× temp={-150..150} × tech={0,5,12} = 105 точек, pixel-perfect совпадение
Go vs PHP eval).

- **Категория**: формула
- **Цвет**: 🔴
- **Origin**: формула `prod_hydrogen` для Hydrogen Lab содержит
  `(-0.002 * {temp} + 1.28)` — холодные планеты производят больше
- **Nova**: вероятно нет температурного модификатора (нужна
  верификация)
- **Разница**: сила водородного производства зависит от температуры
- **Как сделать**: реализовать в Go-функции
  `economy.HydrogenProduction(level, tech, temp)` с условием
  `if balance_profile == "legacy" { applyTempModifier }`
- **Объём**: 2-3 дня (включая тесты)
- **Связь**: D-026, formula-dsl.md

### D-030. Charge_* экспоненты (×1.5 vs ×1.6 vs другие)

✅ **ЗАКРЫТО** (план 64, 2026-04-28). nova `BuildingSpec.CostFactor`
поддерживает per-building factor (verify в configs/buildings.yml:
metal_mine 1.5, silicon_lab 1.6, robotic_factory 2.0, hydrogen_plant
1.8 — все совпадают с origin). Импортёр инфериит cost_factor из
charge_metal-формул origin через эмпирическое отношение (level=10
vs 11 при basic=10000). Все cost_factor в origin.yaml override
сгенерированы автоматически и совпадают с nova-defaults в большинстве
случаев (origin = nova по cost_factor для зданий; ships override
переопределяет cost_base для специальных юнитов Lancer/Shadow).

- **Категория**: формула
- **Цвет**: 🟡
- **Origin**: charge формулы у разных зданий разные:
  Metal Mine `pow(1.5, level-1)`, Silicon Lab `pow(1.6, level-1)`,
  Solar Plant простая `50 * pow(1.5, level)`
- **Nova**: `configs/buildings.yml` имеет `cost_factor`
- **Разница**: nova может иметь унифицированный `cost_factor`
  vs origin per-building
- **Как сделать**: проверить, что nova поддерживает per-building
  factor; если нет — расширить `buildings.yml`
- **Объём**: 1 день верификация
- **Связь**: D-026

---

## Категория «event-loop / механика» (D-031 .. D-038)

### D-031. EVENT_TOURNAMENT_* (3 типа)

- **Категория**: event-loop / механика
- **Цвет**: 🟠
- **Origin**: EVENT_TOURNAMENT_SCHEDULE, RESCHEDULE, PARTICIPANT
  объявлены, но обработчики не реализованы (зарезервировано)
- **Nova**: ОТСУТСТВУЮТ
- **Разница**: legacy фича не реализована даже в origin
- **Как сделать**: новая фича в nova-backend (план для всех
  вселенных как опция). Связано с U-002.
- **Объём**: 3-4 недели (полностью новая фича)

### D-031b. EVENT_DEMOLISH_CONSTRUCTION (handler-stub)

- **Категория**: event-loop / inventory-bug
- **Цвет**: 🟢 (закрыт планом 65 Ф.1, 2026-04-28)
- **Origin**: `EventHandler::demolish` (PHP, EventHandler.class.php:2257)
  понижает уровень здания (`building2planet.level = data.level + added`),
  при level=0 — DELETE строки. Уменьшает игроку `b_points`.
- **Nova до фикса**: `KindDemolishConstruction (Kind=2)` объявлен в
  [kinds.go:17](../../projects/game-nova/backend/internal/event/kinds.go),
  но handler в `cmd/worker/main.go` не зарегистрирован → событие
  валилось бы в `error` со статусом «no handler for kind 2».
- **Nova после фикса**: `HandleDemolishConstruction` в
  [handlers.go](../../projects/game-nova/backend/internal/event/handlers.go) —
  идемпотентный handler, зеркалит `HandleBuildConstruction`.
  Очки не инкрементятся (отличие от legacy: в nova очки derived
  state, пересчитываются `withScore` decorator + `KindScoreRecalcAll`).
- **Разница vs legacy oxsar2**: убран DELETE строки при level=0
  (UPDATE level=0 эквивалентен по чтению, но даёт audit-trail).
- **Объём**: реализовано (~120 строк handler + ~290 строк тестов).

### D-032. EVENT_TELEPORT_PLANET

- **Категория**: event-loop / механика
- **Цвет**: 🟠
- **Origin**: артефакт перемещает планету
- **Nova**: ОТСУТСТВУЕТ
- **Как сделать**: `KindTeleportPlanet` + handler +
  `users.last_planet_teleport_at` (D-016)
- **Объём**: 5-7 дней
- **Связь**: U-009, D-016

### D-033. EVENT_TEMP_PLANET_DISAPEAR

- **Категория**: event-loop / механика
- **Цвет**: 🟠
- **Origin**: временные планеты живут TTL и исчезают
- **Nova**: `KindExpirePlanet (65)` есть — soft-delete по `expires_at`.
  **Совпадает по семантике**, верифицировать payload-схему
- **Цвет**: пересмотрел на 🟢 — фактически реализовано
- **Объём**: 0-1 день верификация
- **Связь**: возможно НЕ расхождение

### D-034. EVENT_RUN_SIM_ASSAULT (отложенный симулятор)

- **Категория**: event-loop
- **Цвет**: 🟠
- **Origin**: запускает симуляцию боя как событие
- **Nova**: симулятор синхронный (`POST /api/battle-sim`)
- **Разница**: origin может ставить симуляцию в очередь
- **Как сделать**: для legacy01 опционально — Kind +
  endpoint-альтернатива
- **Объём**: 2 дня
- **Риски**: вероятно не нужно — синхронный симулятор удобнее

### D-035. EVENT_DELIVERY_ARTEFACTS (доставка артефактов флотом)

- **Категория**: event-loop / механика
- **Цвет**: 🟢 (закрыто 2026-04-28, план 65 Ф.2)
- **Origin**: доставка артефакта между планетами через флот
  (`EventHandler::transport` ветка `EVENT_DELIVERY_ARTEFACTS` +
  `Artefact::onOwnerChange`).
- **Nova**: `KindDeliveryArtefacts Kind = 23` + handler
  `HandleDeliveryArtefacts` в
  [internal/event/handlers.go](../../../projects/game-nova/backend/internal/event/handlers.go).
- **Решение**:
  - typed payload `DeliveryArtefactsPayload{FleetID, ArtefactIDs[]}`
    (R13); получатель и планета назначения берутся из
    `e.UserID`/`e.PlanetID` (как у demolish — не дублируем в payload);
  - семантика: UPDATE `artefacts_user.user_id/planet_id` на получателя,
    флот → `returning`, ресурсы НЕ трогаем (отличие от обычного
    TRANSPORT — см. EventHandler.class.php:2688 «if mode != DELIVERY_ARTEFACTS»);
  - `active → held` с обнулением `activated_at`/`expire_at`. Revert
    эффектов не зовём синхронно — nova вычисляет effect-стек по списку
    активных артефактов на каждом чтении (см.
    [simplifications.md](../../simplifications.md));
  - per-universe (R10): обе стороны (sender, recipient) проверяются в
    одной вселенной через `users.universe_id` JOIN — защита от багов
    биржи плана 68;
  - идемпотентность: артефакт уже у получателя → skip; флот
    `state ≠ outbound` → no-op (ArriveHandler-паттерн).
- **Регистрация**: `withAchievement(HandleDeliveryArtefacts)` (без
  `withScore` — артефакты не дают очков; без `withDailyQuest` — нет
  такого квеста в дизайне).
- **Тесты**:
  [delivery_artefacts_test.go](../../../projects/game-nova/backend/internal/event/delivery_artefacts_test.go) —
  round-trip JSON, rapid-property на skip-decision, 5 golden-сценариев
  через `TEST_DATABASE_URL` (single/three artefacts delivered, idempotent
  replay, active reset to held, fleet-not-outbound noop), 5 negative
  payload-validation кейсов.
- **Объём фактический**: ~210 строк handler + ~430 строк тестов.
- **Связь**: план 65 Ф.2.

### D-035b. KindExchangeExpire / KindExchangeBan — перенос в план 68

- **Категория**: event-loop / scope
- **Цвет**: ⚪ (отложено — перенесено)
- **Решение 2026-04-28**: stub-handler с `ErrSkip` нарушал бы R15
  (без TODO/MVP-сокращений). Концептуально оба Kind'а — биржевые и
  должны жить рядом со своим service'ом в `internal/exchange/`, не в
  общем `internal/event/handlers.go`.
- **Где будут реализованы**: план 68 (биржа артефактов) — handler'ы
  + регистрация в `cmd/worker/main.go` + миграция таблиц
  `exchange_lots`/`exchange_bans`.
- **Снижение scope плана 65** с 6 Kind'ов до 5 (Demolish ✅,
  DeliveryArtefacts ✅, AttackDestroyBuilding, AttackAllianceDestroyBuilding,
  AllianceAttackAdditional, TeleportPlanet).

### D-036. EVENT_ALIEN_FLY_UNKNOWN / GRAB_CREDIT / CHANGE_MISSION_AI

- **Категория**: event-loop / механика
- **Цвет**: 🟠 (в работе — план 66)
- **Origin**: 3 типа алиен-событий с богатой логикой (см.
  `alien-ai-comparison.md`)
- **Nova**: KindAlienFlyUnknown (33) и KindAlienChangeMissionAI (81)
  объявлены без обработчиков; KindAlienGrabCredit (37) встроено в
  Attack
- **Как сделать**: реализовать все 3 как отдельные обработчики в
  `internal/origin/alien/`. Подробнее — `alien-ai-comparison.md` записи
  A4-A11
- **Объём**: 3 недели (3 итерации плана 66)
- **Связь**: alien-ai-comparison.md A1-A14, план 66
- **Прогресс**:
  - 2026-04-28 (Ф.1+Ф.2 плана 66): создан пакет
    `internal/origin/alien/` с state machine, Config (1-в-1 с
    consts.php:752-770), helper'ами `GenerateFleet`/`PickAttackTarget`/
    `PickCreditTarget`/`ShuffleKeyValues`/`ApplyShuffledTechWeakening`/
    `CalcGrabAmount`/`CalcGiftAmount`/`HoldingExtension` и
    интерфейсом Loader.
  - 2026-04-28 (Ф.3 плана 66): добавлены handlers
    `FlyUnknownHandler` / `GrabCreditHandler` / `ChangeMissionAIHandler`
    + `MissionPayload` / `ChangeMissionPayload` (R13 typed) +
    pgx-реализация `Loader`. Зарегистрированы в `cmd/worker/main.go`
    рядом с существующими alien-handler'ами. Применён эталонный
    паттерн от plan-65 Ф.1 (KindDemolishConstruction, 9a3992a384):
    typed payload, slog audit (R3), R8 метрики автоматом на уровне
    worker, idempotency через FOR UPDATE SKIP LOCKED + state-machine.
  - 2026-04-28 (Ф.4 плана 66): `KindAlienHoldingAI` расширен с
    50/50 random extract/unload до 8 sub-phases как в origin
    (`AlienAI.class.php:924-1014`): 2 активных
    (`SubphaseExtractAlienShips`, `SubphaseUnloadAlienResources`)
    + 6 заглушек (`SubphaseRepairUserUnits` /`AddUserUnits` /
    `AddCredits` / `AddArtefact` / `GenerateAsteroid` /
    `FindPlanetAfterBattle`) — пустые тела как в origin
    (PHP:1086-1124). Добавлены `HoldingAIPayload` (R13 typed) +
    `control_times++`, продление `parent.fire_at` платежом
    (`2h × paid / 50`, capped HaltingMaxRealTime=15d), длительность
    следующего тика растёт с control_times
    (`HoldingAISubphaseDuration`). Регистрация worker'а перенесена
    с `internal/alien.HoldingAIHandler` на
    `originAlienSvc.HoldingAIHandler()`. Закрывает A5 / A14 в
    alien-ai-comparison.md.
  - **Применимо**: ко всем вселенным (R0-исключение 2026-04-28,
    решение пользователя). Пакет `origin/alien/` — источник кода,
    не таргет вселенной.

### D-037. EVENT_ATTACK_DESTROY_BUILDING / ALLIANCE_DESTROY_BUILDING ✅ (закрыто 2026-04-28, план 65 Ф.3+Ф.4)

- **Категория**: event-loop / механика
- **Цвет**: 🟢
- **Origin**: атака с целью разрушения постройки (не луны, а
  здания)
- **Nova (было)**: атака разрушения только луны
  (`KindAttackDestroyMoon`, план 20 Ф.6)
- **Nova (стало)**: реализовано — `KindAttackDestroyBuilding=26` (single)
  и `KindAttackAllianceDestroyBuilding=29` (ACS); общая логика разрушения
  здания вынесена в [fleet/destroy_building.go](../../../projects/game-nova/backend/internal/fleet/destroy_building.go),
  payload расширен опциональным `target_building_id` (omitempty); при
  отсутствии — random-выбор из buildings планеты (с фильтром
  UNIT_EXCHANGE/UNIT_NANO_FACTORY как в legacy). Сообщения через i18n
  `assaultReport.buildingDestroyed*` / `enemyBuildingDestroyed*`.
- **Сознательное упрощение**: legacy-эвристика «у атакующего должно
  быть здание сравнимого уровня» (`DESTROY_BUILD_RESULT_MIN_OFFS_LEVEL`)
  не реализована — в nova random-выбор из всех eligible зданий, без
  тонкой балансировки ([simplifications.md](../../simplifications.md)).
- **Связь**: backend готов; UI выбора здания — отдельным планом, когда
  дойдёт оригинальный Mission UI.

### D-038. EVENT_ALIEN_ATTACK_CUSTOM (admin-инициируемая)

- **Категория**: event-loop
- **Цвет**: 🟡
- **Origin**: admin может инициировать атаку с custom параметрами
- **Nova**: ОТСУТСТВУЕТ
- **Как сделать**: admin endpoint
  `POST /api/admin/alien/custom-attack`, переиспользует
  `KindAlienAttack` с custom payload
- **Объём**: 1-2 дня

---

## Категория «API» (D-039 .. D-045)

### D-039. Биржа артефактов (Exchange / Stock / StockNew)

- **Категория**: api / механика
- **Цвет**: 🟠 (главное расхождение)
- **Origin**: 3 контроллера
  (`Exchange.class.php` 1220 стр, `Stock.class.php` 757 стр,
  `StockNew.class.php` 850 стр) + 5+ шаблонов + таблицы
  `na_exchange*`
- **Nova**: ОТСУТСТВУЕТ как player-to-player биржа. Только
  `artmarket` (продажа от системы).
- **Как сделать**: новый модуль `internal/exchange/` ~2000 стр
  Go + 5+ endpoint'ов:
  - `GET /api/exchange/lots` — список лотов
  - `POST /api/exchange/lots` — создать
  - `GET /api/exchange/lots/{id}` — детали
  - `POST /api/exchange/lots/{id}/buy` — купить
  - `DELETE /api/exchange/lots/{id}` — отозвать
- **Объём**: 2-3 недели backend + frontend
- **Связь**: U-001

### D-040. Передача лидерства альянса (abandonAlly)

- **Категория**: api
- **Цвет**: 🟠
- **Origin**: `Alliance.class.php::abandonAlly`,
  `referFounderStatus`
- **Nova**: ОТСУТСТВУЕТ endpoint
- **Как сделать**:
  `POST /api/alliances/{id}/transfer-ownership/{userID}`
- **Объём**: 1 день backend + frontend кнопка
- **Связь**: U-004

### D-041. Три описания альянса (external/internal/application)

- **Категория**: api / домен
- **Цвет**: 🟠
- **Origin**: `Alliance.class.php::updateAllyPrefs` устанавливает
  3 поля
- **Nova**: одно `description`
- **Как сделать**: миграция `alliances` —
  `description_external, description_internal, description_apply
  TEXT`. Backend + DTO + frontend.
- **Объём**: миграция + handler + frontend; 2-3 дня
- **Связь**: U-015

### D-042. Global mail членам альянса

- **Категория**: api / механика
- **Цвет**: 🟠 (заблокировано планом 57)
- **Origin**: `Alliance.class.php::globalMail`
- **Nova**: ОТСУТСТВУЕТ (после плана 57 mail-service —
  сделаем легко)
- **Как сделать**: после плана 57 — endpoint
  `POST /api/alliances/{id}/global-mail` с TipTap-payload
- **Объём**: после 57 — 3-5 дней
- **Связь**: U-006

### D-043. Phalanx vs MonitorPlanet

- **Категория**: api
- **Цвет**: 🟡
- **Origin**: `MonitorPlanet.class.php` показывает события на
  чужой планете (через сенсорную фалангу)
- **Nova**: `GET /api/phalanx` есть
- **Разница**: верифицировать паритет (UI и payload)
- **Объём**: 1 день верификации

### D-044. ResTransferStats

- **Категория**: api
- **Цвет**: 🟡
- **Origin**: `ResTransferStats.class.php` — статистика передачи
  ресурсов между членами альянса
- **Nova**: таблица `resource_transfers` есть, endpoint`а нет
- **Как сделать**: `GET /api/alliances/{id}/resource-transfers`
- **Объём**: 2 дня

### D-045. ExchangeOpts (auto-exchange при переполнении)

- **Категория**: api / механика
- **Цвет**: 🟡
- **Origin**: `ExchangeOpts.class.php` — настройки автоматического
  обмена при переполнении хранилища
- **Nova**: ОТСУТСТВУЕТ
- **Как сделать**: настройки в `users` (поля `auto_exchange_*`) +
  обработка в economy
- **Объём**: 3-5 дней
- **Риски**: авто-операции могут ломать ожидания игрока

---

## Категория «assets» (D-046)

### D-046. Runtime-генерируемые ассеты артефактов

- **Категория**: assets
- **Цвет**: 🟠
- **Origin**: `public/artefact-image.php` (153 строки PHP-GD) —
  композитные PNG артефактов
- **Nova**: статические иконки
- **Как сделать**: Go-генератор в `internal/artefact/image.go`
  через `image/draw`. Кеш на диск.
- **Объём**: 100-200 строк Go + endpoint + кеш; 1 неделя
- **Риски**: лицензии шрифтов и базовых иконок (план 40)

---

## Сводка

**Всего записей**: 46 (D-001 .. D-046).

По категориям:
| Категория | Кол-во | Цвета |
|---|---|---|
| Домен (поля) | 25 | 🟡×8, 🟠×0, 🔴×11, 🟣×6 |
| Формула / баланс | 5 | 🟡×2, 🟠×2, 🔴×1, 🟣×0 |
| Event-loop / механика | 8 | 🟡×1, 🟠×7, 🔴×0, 🟣×0 |
| API | 7 | 🟡×4, 🟠×3, 🔴×0, 🟣×0 |
| Assets | 1 | 🟠×1 |

По цветам:
- 🟡 Жёлтых: 15 (правка конфига или флаг)
- 🟠 Оранжевых: 13 (нужно реализовать в nova)
- 🔴 Красных: 12 (формула/механика расходятся)
- 🟣 Фиолетовых: 6 (тихая семантика)

**Минимальный порог плана 62 (≥30 записей)**: ✅ выполнен (46).

## Связь с другими артефактами

- `alien-ai-comparison.md` записи A1-A14 — расширение D-036
- `nova-ui-backlog.md` U-NNN ↔ D-NNN: U-001↔D-039, U-004↔D-040,
  U-005↔D-014, U-006↔D-042, U-009↔D-032, U-015↔D-041
- `formula-dsl.md` § «Архитектурное расхождение» ↔ D-026
- `origin-ui-replication.md` S-NNN — что воспроизводится
  pixel-perfect (отличается от D-NNN, который про backend)

## Правила использования

1. Каждое D-NNN получает **рабочий статус** при начале реализации:
   `🔵 в работе`, `✅ закрыто`, `🚫 отказано`.
2. При смене стратегии через 6-12 месяцев журнал **дополняется/
   верифицируется**, не пересоставляется с нуля.
3. Записи 🟣 при реализации **обязательно покрываются
   golden/property-based тестами на паритет**.
4. Группировка в будущие планы 63+ — см. `roadmap-report.md` (Ф.5).
