# План 69 (ремастер): Расширение domain-полей в nova

**Дата**: 2026-04-28
**Статус**: Скелет (детали допишет агент-реализатор при старте)
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

| D-NNN | Поле | Назначение | R1 (имя) |
|---|---|---|---|
| D-001 | `users.max_points`, опц. `dm_points`, `be_points`, `of_points` | Достижения и категории очков | snake_case ОК |
| D-004 | `users.protected_until_at` | Защита новичков от атак (timestamp) | `_at` суффикс по R1 |
| D-005 | `users.role` enum + value `observer` ИЛИ `users.is_observer bool` | Наблюдатель без боя | `is_observer` если простая бинарная |
| D-008 | `users.profession_changed_at` | Когда менял профессию (cooldown) | `_at` ОК |
| D-020 | `users.last_global_chat_read_at`, `users.last_ally_chat_read_at`, `users.chat_language` | Маркеры прочтения чата | `_at` для timestamp, `chat_language` для locale |
| D-019 | `users.home_planet_id` | Главная планета (FK на planets.id) | `_id` для FK по R1 |
| D-016 | `users.last_planet_teleport_at` | Cooldown телепорта | `_at` ОК |
| D-003 | `users.account_deletion_scheduled_at` | Soft-удаление с задержкой (для всех вселенных) | `_at` ОК |
| W1 | `users.notes TEXT` | Приватные заметки игрока (notepad из legacy-PHP, S-Notepad экран в плане 72) | TEXT, nullable, размер ≤ 16KB по CHECK |

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

- Ф.1. Миграция БД (одна большая migration с 9 ALTER TABLE).
- Ф.2. Обновить sqlc-модели + регенерация.
- Ф.3. Handler-обновления для тех endpoint'ов, где поля отдаются /
  читаются.
- Ф.4. Защитная логика protected_until_at в attack-handler.
- Ф.5. Cooldown teleport / профессии.
- Ф.6. Endpoint для notes (GET/PUT `/api/users/me/notes`).
- Ф.7. Финализация.

## Конвенции (R1-R5)

- Все новые колонки — snake_case + nova-стиль (НЕ калька с
  `na_user.umode` / `na_user.del_request_at`).
- `_at` для timestamps (`protected_until_at`, не `protect_til`).
- `_id` для FK (`home_planet_id`).
- `is_*` / `has_*` для boolean.
- `users.notes` — TEXT, размер по CHECK ≤ 16KB.
- Регрессионные тесты на текущие fixtures uni01/uni02.

## Объём

2 недели. Одна большая миграция + handler-обновления + тесты.

## References

- D-001..D-025 в divergence-log.md (категория «домен»).
- Существующая `projects/game-nova/backend/migrations/` — формат миграций.
- `projects/game-origin-php/migrations/001_schema.sql` — referenc для
  смысла полей (но НЕ для имён — по R1).
