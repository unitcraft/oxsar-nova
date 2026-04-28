# План 69 (ремастер): Расширение domain-полей в nova

**Дата**: 2026-04-28
**Статус**: Ф.0 (дельта-аудит) выполнена. R10 (home_planet_id) — отказ. Ф.1 в работе.
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

- **Ф.0. Дельта-аудит** ✅ (выполнен 2026-04-28, см. ниже).
- Ф.1. Миграция БД дельты (только реально отсутствующие поля).
- Ф.2. Обновить sqlc-модели + регенерация.
- Ф.3. Handler-обновления для тех endpoint'ов, где поля отдаются /
  читаются.
- Ф.4. Защитная логика protected_until_at в attack-handler.
- Ф.5. Cooldown teleport / профессии.
- Ф.6. Endpoint для notes (GET/PUT `/api/users/me/notes`) — **переоценить**:
  notes уже реализованы как отдельная таблица `user_notepad`
  (миграция 0050). Возможно, endpoint существует или достаточно
  расширения существующего модуля.
- Ф.7. Финализация.

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
