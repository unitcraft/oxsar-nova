# Промпт: выполнить план 69 (расширение domain-полей)

**Дата создания**: 2026-04-28
**План**: [docs/plans/69-remaster-domain-fields-extension.md](../plans/69-remaster-domain-fields-extension.md)
**Зависимости**: блокируется планом 64 (`configs/balance/origin.yaml`).
**Объём**: 2 нед, одна большая миграция + handler-обновления + тесты.

---

```
Задача: выполнить план 69 (ремастер) — расширение domain-полей в
users-таблице nova под механики origin (общий знаменатель — поля
доступны во всех вселенных, активны через NULL/default где требуется).

ПЕРЕД НАЧАЛОМ:

1) git status --short.

2) Прочитай:
   - docs/plans/69-remaster-domain-fields-extension.md
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5»
   - docs/research/origin-vs-nova/divergence-log.md
     D-001, D-003, D-004, D-005, D-008, D-016, D-019, D-020

3) Выборочно:
   - projects/game-nova/backend/migrations/ — формат миграций.
   - projects/game-origin-php/migrations/001_schema.sql — referenc
     для смысла полей (НЕ для имён, R1).

ЧТО НУЖНО СДЕЛАТЬ:

Одна большая миграция migrations/00NN_extend_users_for_origin.sql:
- users.max_points (BIGINT, опц. dm_points / be_points / of_points
  для категорий) — D-001.
- users.protected_until_at (TIMESTAMPTZ, nullable) — D-004,
  защита новичков от атак.
- users.is_observer (BOOLEAN default false) — D-005, role observer.
- users.profession_changed_at (TIMESTAMPTZ, nullable) — D-008,
  cooldown.
- users.last_global_chat_read_at, last_ally_chat_read_at,
  chat_language (TIMESTAMPTZ, TEXT) — D-020, маркеры прочтения чата.
- users.home_planet_id (UUID FK на planets.id) — D-019.
- users.last_planet_teleport_at (TIMESTAMPTZ, nullable) — D-016.
- users.account_deletion_scheduled_at (TIMESTAMPTZ, nullable) —
  D-003, soft-удаление с задержкой (для всех вселенных).
- users.notes (TEXT, nullable, CHECK length ≤ 16384) — W1, notepad
  из legacy-PHP, S-Notepad экран в плане 72.

ОТКАЗ (НЕ вводим):
- users.race (D-021) — мёртвое поле в legacy-PHP, не используется.
- users.ui_theme / users.ui_pack (D-007) — YAGNI, одна тема.

Backend handlers:
- Обновить SQL-запросы / Go-модели в internal/users/ под новые
  колонки (sqlc в проекте не используется, несмотря на упоминание
  в CLAUDE.md — это историческое расхождение, поправлено 2026-04-28).
- Защитная логика protected_until_at в attack-handler.
- Cooldown teleport / профессии.
- GET/PUT /api/users/me/notes (R6 REST).

ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА:

R0: общий знаменатель — поля для всех вселенных, не правка
существующих чисел. modern-вселенные используют те же поля но в
своих сценариях.
R1: snake_case + _at для timestamps + _id для FK + is_*/has_* для
boolean (is_observer, не observer_flag). UUID для FK (R1).
R2: OpenAPI первым.
R8: Prometheus метрики для новых endpoints.
R10: per-universe изоляция — users — это identity-уровень,
cross-universe (R10 говорит users НЕ per-universe). Но
home_planet_id — это per-universe (planets per-universe). Решить
в миграции: либо home_planet_id хранится в отдельной cross-table,
либо допустить что разные вселенные = разные users (текущий
паттерн nova).
R12: i18n — grep nova-bundle.
R15: без упрощений — все колонки nullable / с безопасным default,
регрессионные тесты на текущие fixtures.

GIT-ИЗОЛЯЦИЯ:
- Свои пути: migrations/, internal/users/, handlers,
  openapi.yaml, тесты, docs/plans/69-..., divergence-log.md.

КОММИТЫ:

1-2 коммита:
1. feat(users): миграция расширения полей для ремастера (план 69).
2. feat(users): handlers (notes/teleport/cooldown) + тесты.

ЧЕГО НЕ ДЕЛАТЬ:
- НЕ вводить race / ui_theme — отказ зафиксирован.
- НЕ делать `delete INT(10)` auto-deletion — у нас email-коды.
- НЕ менять существующие fixtures (все колонки nullable / default).

УСПЕШНЫЙ ИСХОД:
- D-001, D-003, D-004, D-005, D-008, D-016, D-019, D-020 закрыты.
- W1 (notes) реализован.
- Все existing nova-тесты зелёные.

Стартуй.
```
