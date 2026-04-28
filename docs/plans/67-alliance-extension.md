# План 67: Расширение alliance-системы

**Дата**: 2026-04-28
**Статус**: Скелет (детали допишет агент-реализатор при старте)
**Зависимости**: нет критичных (можно параллелить).
**Связанные документы**:
- [62-origin-on-nova-feasibility.md](62-origin-on-nova-feasibility.md)
- [docs/research/origin-vs-nova/divergence-log.md](../research/origin-vs-nova/divergence-log.md) —
  D-014, D-040, D-041
- [docs/research/origin-vs-nova/nova-ui-backlog.md](../research/origin-vs-nova/nova-ui-backlog.md) —
  U-004, U-005, U-012, U-013, U-015
- [docs/research/origin-vs-nova/roadmap-report.md](../research/origin-vs-nova/roadmap-report.md) —
  R1-R5 + раздел плана 67

---

## Цель

Добавить в game-nova alliance-фичи, отсутствующие сейчас, но
присутствующие в origin. Применимо для **всех вселенных**
(uni01/uni02/origin) — это «общий знаменатель», не origin-only
(R1 секция «Общий знаменатель»).

---

## Что делаем

| ID | Фича | Что |
|---|---|---|
| D-041, U-015 | 3 описания альянса | Поля `description_external`, `description_internal`, `description_apply` (snake_case по R1) — публичное / для членов / для заявок |
| D-040, U-004 | Передача лидерства | `POST /api/alliances/{id}/transfer-leadership/{userId}` + email-подтверждение через identity (как в D-003) |
| D-014, U-005 | Гранулярные права рангов | Таблица `alliance_ranks` с `permissions JSONB` (см. R1: snake_case колонки, JSONB для флагов прав) |
| U-012 | Полнотекстовый поиск альянсов | Расширить `GET /api/alliances` фильтрами (тип, размер, открытость) + полнотекст по name/tag |
| U-013 | Альянсный лог активности | Таблица `alliance_audit_log` (по образцу `admin_audit_log` плана 14) |

**Не входит в этот план** (отдельные задачи):
- U-011 (custom logo альянса) — нужны storage + moderation, после
  плана 57 (mail) или отдельно.
- Global mail членам — после плана 57 (mail-service с TipTap).

---

## Что НЕ делаем

- Не переносим bbcode — он выкидывается, заменяется TipTap (план 57).
- Не вводим премиум-альянс / лимиты — отдельная задача.

## Этапы (детали — при старте)

- Ф.1. Миграции БД (3 описания, alliance_ranks с permissions JSONB,
  alliance_audit_log).
- Ф.2. Backend handler'ы + RBAC права рангов в middleware.
- Ф.3. Передача лидерства с email-подтверждением.
- Ф.4. Полнотекстовый поиск (Postgres `tsvector`).
- Ф.5. Frontend в game-nova (для uni01/uni02).
- Ф.6. Финализация.

## Конвенции (R1-R5)

- Все новые колонки — snake_case, без префикса таблицы.
- `description_external/internal/apply` (а не `descExt`/`int_desc`/etc.).
- `permissions` в alliance_ranks — JSONB с ключами в snake_case
  (`can_invite`, `can_kick`, `can_send_global_mail`, `can_manage_diplomacy`,
  `can_change_description`, `can_propose_relations`).
- Timestamps — `_at` суффикс (`leadership_transferred_at`).
- Таблица — `alliance_ranks` (мн. ч.), не `alliance_rank`.
- OpenAPI первым (R2).

## Объём

2-3 недели. ~600-1000 строк Go + ~400-800 строк frontend (game-nova).

## References

- D-014, D-040, D-041 + U-004, U-005, U-012, U-013, U-015.
- Существующий `internal/alliance/` в game-nova-backend.
- `projects/game-origin-php/src/game/page/Alliance.class.php` —
  origin-референс (1413 строк, 30 действий).
