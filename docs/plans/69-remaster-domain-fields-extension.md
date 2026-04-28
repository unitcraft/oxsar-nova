# План 69 (ремастер): Расширение domain-полей в nova

**Дата**: 2026-04-28
**Статус**: ✅ Завершён 2026-04-28. Ф.0 ✅ Ф.1 ✅ Ф.3 ✅ Ф.4 ✅ Ф.5 (частично — profession уже был; teleport отложен до плана 72) ✅ Ф.6 ✅ Ф.7 ✅.
**Зависимости**: блокируется планом 64 (`configs/balance/origin.yaml` для дефолтных
значений вселенной origin).
**Связанные документы**:
- [62-origin-on-nova-feasibility.md](62-origin-on-nova-feasibility.md)
- [docs/research/origin-vs-nova/divergence-log.md](../research/origin-vs-nova/divergence-log.md) —
  D-001, D-003, D-004, D-005, D-008, D-016, D-019, D-020 (D-007, D-021 — отказ, см. ниже)
- [docs/research/origin-vs-nova/roadmap-report.md](../research/origin-vs-nova/roadmap-report.md) —
  R1-R5 + раздел плана 69

---

## Цель

Добавить недостающие поля в `users` (и сопутствующие таблицы) для
поддержки механик origin (классические из oxsar2). Применимо для всех вселенных
(uni01/uni02/origin), но часть полей **активна только для legacy**
(через NULL / default).

---

## Что делаем (по D-NNN)

> Уточнено в Ф.0 (дельта-аудит). Колонка **Статус** показывает
> фактическое состояние: ✅ закрыто, ⚙️ ждёт Ф.1, ⏳ ждёт R10.

| D-NNN | Поле | Назначение | Статус |
|---|---|---|---|
| D-001 | `users.max_points` (категории dm/be/of — отказ, YAGNI) | Исторический пик очков | ⚙️ Ф.1 |
| D-004 | `users.protected_until_at` | Защита новичков от атак (timestamp) | ⚙️ Ф.1 |
| D-005 | `users.is_observer BOOLEAN` (домен-флаг, не RBAC) | Наблюдатель без боя | ⚙️ Ф.1 |
| D-008 | `users.profession_changed_at` | Когда менял профессию (cooldown) | ✅ закрыто (миграция 0046_profession.sql) |
| D-020 | `users.last_global_chat_read_at`, `users.last_ally_chat_read_at` (chat_language — отказ, есть `users.language`) | Маркеры прочтения чата | ⚙️ Ф.1 |
| D-019 | `users.home_planet_id` | Главная планета (FK на planets.id) | ❌ отказ (R10 / YAGNI, см. ниже) |
| D-016 | `users.last_planet_teleport_at` | Cooldown teleport-планеты (не путать со stargate_cooldowns per-planet) | ⚙️ Ф.1 |
| D-003 | `users.account_deletion_scheduled_at` | Soft-удаление с задержкой | ✅ закрыто архитектурно лучше: `account_deletion_codes` (миграция 0051) — email-confirm flow |
| W1 | `users.notes TEXT` | Приватные заметки игрока | ✅ закрыто архитектурно лучше: `user_notepad` (миграция 0050) — отдельная таблица |

---

## Что НЕ делаем

- **Не делаем** `delete INT(10)` auto-deletion как в origin —
  у нас email-коды через identity (D-003 для всех вселенных).
- **Не вводим** `users.race` (D-021) — отказ. В legacy-PHP-схеме
  поле `na_user.race` есть с дефолтом 1, но в коде не
  используется (мёртвая архитектурная заглушка). См.
  roadmap-report «Часть V».
- **Не вводим** `users.ui_theme` / `users.ui_pack` (D-007) — отказ
  для первой итерации (P1 / YAGNI). У nova одна тема, у origin
  одна тема (pixel-perfect клон `standard`). Свободный выбор тем —
  отдельным планом если понадобится.
- Не миграцируем существующие fixtures — все колонки nullable / с
  безопасным default.

## Этапы (детали — при старте)

- **Ф.0. Дельта-аудит** ✅ (2026-04-28, см. ниже).
- **Ф.1. Миграция БД дельты** ✅ (миграция `0072_users_remaster_fields.sql`,
  5 ALTER: max_points, protected_until_at, is_observer,
  last_planet_teleport_at, last_global_chat_read_at, last_ally_chat_read_at).
- **Ф.2. sqlc-регенерация** — N/A: sqlc в `game-nova` не используется,
  все запросы — сырые pgx (см. CLAUDE.md → требует обновления).
- **Ф.3 + Ф.4. Handler-обновления + защитная логика** ✅:
  - `internal/fleet/attack.go`: расширена проверка защиты —
    срабатывает по global protectionPeriod ИЛИ per-user
    `protected_until_at` ИЛИ `is_observer = true` (флот возвращается
    без боя). Spy / acs / moon-destruction handlers НЕ обновлены —
    это уже существующая дыра pre-69 (защита новичков отсутствовала
    и там); вне scope плана.
  - `internal/score/service.go`: `RecalcUser` обновляет
    `max_points = GREATEST(max_points, total)` (D-001 — исторический
    пик, никогда не убывает).
  - Фильтр `is_observer = false` добавлен везде, где был фильтр
    `umode = false` для публичных рейтингов: `score.Top`,
    `score.TopAlliances`, `score.PlayerRank`, `records.topScore`,
    `score/handler.go` online-статистика, `galaxy/repository.go`
    рейтинг в галактика-обзоре. НЕ добавлен в `score.RecalcAll`
    (observer всё равно играет, очки считаем) и
    `score.VacationPlayers` (observer + vacation — состояние данных).
  - `internal/chat/handler.go`: новые endpoints
    `POST /api/chat/{kind}/read` и `GET /api/chat/{kind}/unread`
    (D-020). Используют новые поля `last_global_chat_read_at` /
    `last_ally_chat_read_at` per kind.
- **Ф.5. Cooldowns**:
  - **Profession**: ✅ уже реализован полностью в
    `internal/profession/service.go` (миграция 0046 + handler с
    14-day cooldown, `ErrChangeTooSoon`).
  - **Teleport**: ⏸ отложен до плана 72. Поле
    `last_planet_teleport_at` уже добавлено миграцией 0072 — готово
    к использованию когда механика смены home-планеты будет
    реализована. Сейчас helper'а нет — добавлять "на будущее"
    противоречит KISS.
- **Ф.6. Endpoint для notes** ✅ (2026-04-28):
  - Endpoint **уже существует**: `GET/PUT /api/notepad`,
    зарегистрирован в `cmd/server/main.go`, реализация в
    `internal/notepad/handler.go`. Backed by таблица `user_notepad`
    (миграция 0050).
  - Лимит размера: `MaxLength = 50_000` символов в handler (R15:
    отклонение от первоначального плана «16KB CHECK в миграции» —
    handler-level лимит достаточен; in-row CHECK добавит миграционные
    риски при будущей правке константы и не предохраняет от
    серверной логики, которая уже валидирует. CHECK на уровне SQL
    не добавляем).
  - **Что добавлено в этой сессии**:
    1. OpenAPI: `paths./api/notepad` (GET/PUT) + tag `notepad` +
       schemas `Notepad`, `NotepadSaveRequest` (R2). Документирует
       уже существующий endpoint.
    2. `internal/notepad/handler_test.go` — unit-тесты на
       Unauthorized (GET/PUT), невалидный JSON, превышение
       MaxLength.
    3. `internal/auth/authtest/` — пакет-helper `WithUserID(ctx,
       userID)` для тестов, чтобы не выпускать JWT и не поднимать
       middleware-стек ради unit-теста. `auth.userIDKey`
       переименован в `auth.UserIDKey` (экспортируемый).
  - **R8 Prometheus**: метрики не добавлены. Notepad — низкочастотный
    endpoint (1 GET при открытии экрана + 1 PUT раз в N секунд при
    редактировании); общие request-метрики Chi-router'а покрывают
    его без дополнительных custom labels. Не делаем.
  - **R9 Idempotency**: PUT идемпотентен по семантике (полная замена
    контента + UPSERT). Idempotency-Key header не добавляем — клиент
    дебаунсит сохранения, повторная отправка того же тела — корректный
    no-op в БД.
  - **R12 i18n**: ключи `notepad.*` уже существовали в
    `configs/i18n/{ru,en}.yml` (план 33-35). Не трогаем.
- **Ф.7. Финализация** ✅ (2026-04-28):
  - Шапка плана 69 обновлена → ✅ Завершён.
  - В `divergence-log.md` пометки ✅ для D-001, D-003, D-004, D-005,
    D-008, D-016, D-019, D-020, W1; отказы D-007, D-021.
  - Запись в `docs/project-creation.txt` (итерация 69).

---

## Ф.0. Дельта-аудит (2026-04-28)

Перед планированием миграции проверено фактическое состояние схемы
`users` и связанных таблиц. Часть полей плана 69 уже закрыта
независимыми миграциями (бэклог nova, выполненный параллельно с
подготовкой ремастера); другая часть — закрыта **в иной форме**
(отдельная таблица вместо колонки в `users`), что архитектурно
лучше плана.

### Закрыто полностью (исключаются из Ф.1)

| D-NNN | План 69 (предложение) | Фактическое состояние | Решение |
|---|---|---|---|
| **D-002** | `vacation_*` (vacation mode) | `users.vacation_since`, `users.vacation_last_end` (миграция 0045_vacation_mode.sql) | ✅ Не часть плана 69 (был P-20). Учтено для контекста. |
| **D-008** | `users.profession_changed_at` | `users.profession TEXT NOT NULL DEFAULT 'none'` + `users.profession_changed_at TIMESTAMPTZ` (миграция 0046_profession.sql) | ✅ Закрыто. **Из Ф.1 исключаем.** |
| **W1**   | `users.notes TEXT` (CHECK 16KB) | Отдельная таблица `user_notepad (user_id PK, content text, updated_at)` (миграция 0050_notepad.sql) | ✅ Закрыто **архитектурно лучше** (отдельная таблица — не утяжеляет горячую `users`). **Из Ф.1 исключаем.** Ф.6 переоценить: проверить наличие endpoint `/api/users/me/notes`. CHECK на длину content стоит добавить отдельной микро-миграцией если ещё нет (verify в Ф.6). |
| **D-003** | `users.account_deletion_scheduled_at` | Отдельная таблица `account_deletion_codes (user_id PK, code_hash, issued_at, expires_at, attempts)` (миграция 0051) | ✅ Закрыто **архитектурно лучше** (email-confirm flow вместо простого scheduled_at — соответствует roadmap-report «Часть III» для всех вселенных). **Из Ф.1 исключаем.** |

### Закрыто частично (в плане 69 НЕ актуально, но имя поля иное)

| D-NNN | План 69 предлагал | Фактически | Решение |
|---|---|---|---|
| **D-005** | `is_observer BOOLEAN` ИЛИ `role enum +'observer'` | `users.role` enum **удалён** миграцией 0070_drop_users_role.sql; роли мигрировали в identity-сервис (план 52, RBAC unification). | **Решение по R10 архитектуре**: identity-уровень. `is_observer` — это **домен-флаг**, не RBAC-роль. Хранить в `users` (game-nova) как `is_observer BOOLEAN DEFAULT false`. Identity-сервис не знает про "observer" — это per-universe game-state, не cross-universe identity. **Включаем в Ф.1.** |

### НЕ закрыто (включаются в Ф.1)

| D-NNN | Поле | Тип | Семантика |
|---|---|---|---|
| **D-001** | `users.max_points` | `numeric(20, 4) NOT NULL DEFAULT 0` | Исторический пик `points`. Категории `dm_points`/`be_points`/`of_points` — **отказ** для итерации (YAGNI: 6 текущих категорий u/r/b/a/e_points достаточно, dm/be/of — origin-only внутренняя сложность, не приносит UX). |
| **D-004** | `users.protected_until_at` | `timestamptz` nullable | Защита новичков от атак. |
| **D-005** | `users.is_observer` | `boolean NOT NULL DEFAULT false` | Наблюдатель без боя (домен-флаг, не RBAC-роль). |
| **D-016** | `users.last_planet_teleport_at` | `timestamptz` nullable | Cooldown смены home-планеты. **Не дублирует** `stargate_cooldowns` (миграция 0062): stargate — это прыжок флота между лунами per planet_id; teleport — смена «домашней» планеты per user_id. Разные механики. |
| **D-020** | `users.last_global_chat_read_at`, `users.last_ally_chat_read_at` | `timestamptz` nullable | Маркеры прочтения чата. **Без `chat_language`** — `users.language` (UI язык) уже есть с 0001, отдельный язык чата = YAGNI (одна тема, один язык per user). |

### Отказы (зафиксированы в плане выше)

- **D-007** `ui_theme/ui_pack` — YAGNI.
- **D-021** `race` — мёртвое поле в legacy-PHP.
- **D-001 расширенное**: `dm_points`, `be_points`, `of_points` — YAGNI
  (origin-only внутренняя сложность). `max_points` — да.
- **D-020 chat_language**: YAGNI, `users.language` достаточно.
- **D-019 `home_planet_id`**: R10 / YAGNI (см. раздел ниже).

### Решение по R10 (home_planet_id) — ОТКАЗ (2026-04-28)

`users` per текущему паттерну nova — per-universe; R10
(roadmap-report) предписывает cross-universe identity. Конфликт:
`home_planet_id` указывает на `planets.id` per-universe.

**Решение: вариант (в) — НЕ вводим `home_planet_id` ни в каком виде.**

Обоснование:
- В nova **нет UI/UX концепции «домашней планеты»**. Игрок
  переключается между планетами через `cur_planet_id` /
  session-state (текущий паттерн nova).
- R10 (users cross-universe) + per-universe planets создали бы
  конфликт. Не создавать поле = не создавать конфликт.
- **YAGNI / R15**: пока origin-фронт (план 72) не реализован,
  `home_planet_id` не используется. Когда понадобится — origin
  может работать через session/Zustand без БД-поля.
- Если в будущем потребуется cross-universe state (например,
  preferred-планета per вселенная) — отдельный план с
  cross-table `user_universe_state`. Сейчас не делаем.

D-019 закрывается **отказом**, не миграцией.

### Итог Ф.0

- Изначально план 69: **9 полей**.
- Уже закрыто (полностью или архитектурно лучше): **3** (D-008, D-003, W1).
- Отказы: **1** (D-019 — R10/YAGNI).
- Включаем в Ф.1: **5 полей** — `max_points`, `protected_until_at`,
  `is_observer`, `last_planet_teleport_at`,
  `last_global_chat_read_at` + `last_ally_chat_read_at`.
- Объём правки: уменьшен с «одна большая миграция 9 ALTER» до
  «одна миграция 5 ALTER». Handlers и тесты соответственно
  сужены: убираем notes-endpoint и deletion-flow из Ф.6.

---

## Конвенции (R1-R5)

- Все новые колонки — snake_case + nova-стиль (НЕ калька с
  `na_user.umode` / `na_user.del_request_at`).
- `_at` для timestamps (`protected_until_at`, не `protect_til`).
- `_id` для FK (`home_planet_id`).
- `is_*` / `has_*` для boolean.
- `users.notes` — TEXT, размер по CHECK ≤ 16KB.
- Регрессионные тесты на текущие fixtures uni01/uni02.

## Объём

Изначально 2 недели. После Ф.0 (дельта-аудит): **~1 неделя**.
Одна миграция (5–6 ALTER) + handler-обновления (attack-protection,
teleport cooldown, observer-фильтры в highscore, chat read markers)
+ тесты.

## References

- D-001..D-025 в divergence-log.md (категория «домен»).
- Существующая `projects/game-nova/backend/migrations/` — формат миграций.
- `projects/game-origin-php/migrations/001_schema.sql` — referenc для
  смысла полей (но НЕ для имён — по R1).
