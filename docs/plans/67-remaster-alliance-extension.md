# План 67 (ремастер): Расширение alliance-системы

**Дата**: 2026-04-28
**Статус**: Ф.0 дельта-аудит ✅ 2026-04-28; Ф.1 миграции ✅ 2026-04-28
(коммит 99895ae230); Ф.2 backend handlers + RBAC + дипстатусы + audit-log
✅ 2026-04-28 (этот коммит); Ф.3-Ф.6 — отдельные сессии.
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

## Ф.0 Дельта-аудит (2026-04-28)

Сверка плана 67 с фактическим состоянием схемы nova и кода
`internal/alliance/` + `internal/friends/`. Цель — выкинуть
дубликаты и зафиксировать что реально нужно делать в Ф.1.

| Фича | План 67 | Текущее состояние | Решение |
|---|---|---|---|
| 3 описания альянса (D-041, U-015) | `description_external/internal/apply` | `alliances.description TEXT` (миграция 0017) — одно поле | **Ф.1**: добавить 3 nullable TEXT-поля; legacy `description` оставить как есть (Ф.2 решит маппинг). |
| Передача лидерства (D-040, U-004) | endpoint `transfer-leadership` | endpoint отсутствует; в БД нет следов | **Ф.1**: добавить `alliances.leadership_transferred_at` (audit-метка для UI). Сам handler — Ф.3 (с email-подтверждением через identity, Idempotency-Key). |
| Гранулярные права рангов (D-014, U-005) | таблица `alliance_ranks` + `permissions JSONB` | `alliance_members.rank TEXT` (`owner`\|`member`) + `rank_name TEXT` (миграция 0034, свободный текст без прав) | **Ф.1**: новая таблица `alliance_ranks (id, alliance_id, name, position, permissions JSONB)` + `alliance_members.rank_id` FK (nullable — fallback на builtin `rank`). |
| Расширенные дипстатусы (D-014, B1) | enum `friend / neutral / hostile_neutral / nap / war` | enum `alliance_relation` = `('nap','war','ally')` (миграция 0028) | **Ф.1**: расширить enum. Маппинг: `ally → friend` (rename), `nap`/`war` остаются, добавляем `neutral`/`hostile_neutral`. **NB**: промпт говорил «nova-friend→friend» — но в nova `friend` не было; фактически мигрируем `ally→friend`. |
| Альянсный лог (U-013) | таблица `alliance_audit_log` | отсутствует | **Ф.1**: создать по образцу `admin_audit_log` (миграция 0059). **R10**: nova однобазная (universe = отдельный инстанс БД), `universe_id` не добавляю. |
| Полнотекстовый поиск (U-012) | tsvector + фильтры | отсутствует | **Откладывается на Ф.2**: индекс без handler-потребителя — преждевременная оптимизация. Миграция (GIN-индекс по name/tag) сделается одновременно с поиск-handler. |
| Buddy-list (U-006/U-008 в backlog) | таблица `user_buddies (..., is_mutual)` | ✅ **закрыто в иной форме**: таблица `friends` (миграция 0053), `internal/friends/handler.go` (125 строк), `frontend/src/features/friends/FriendsScreen.tsx`. Endpoints `GET/POST/DELETE /api/friends{,/{userId}}` работают. Односторонний (без `is_mutual`) — намеренное упрощение (см. doccomment в handler.go). | **Отказ от дубликата** `user_buddies`. План 67 в этой части закрыт. Если потребуется `is_mutual` — отдельный план. |

**Итог Ф.1**: 4 миграции (descriptions+leadership-метка, ranks,
audit-log, relations-extend). Backend handlers / OpenAPI / RBAC
middleware / frontend — фазы Ф.2-Ф.5 в следующих сессиях.

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
| D-014, U-005 | Гранулярные права рангов | Таблица `alliance_ranks` с `permissions JSONB`: `can_invite`, `can_kick`, `can_send_global_mail`, `can_manage_diplomacy`, `can_change_description`, `can_propose_relations` (snake_case по R1) |
| D-014 | Расширенные дипломатические статусы | enum `friend / neutral / hostile_neutral / nap / war` (5 значений, B1). TEXT с CHECK по R1, не int magic numbers. Старые nova `friend / neutral / war` мигрируются: nova-friend→friend, nova-neutral→neutral, nova-war→war (без потерь). Новые `hostile_neutral` и `nap` — origin-инспирированы, доступны во всех вселенных. |
| U-012 | Полнотекстовый поиск альянсов | Расширить `GET /api/alliances` фильтрами (тип, размер, открытость) + полнотекст по name/tag (Postgres tsvector) |
| U-013 | Альянсный лог активности | Таблица `alliance_audit_log` (по образцу `admin_audit_log` плана 14) |
| U-006 | Buddy-list (друзья игрока) | Таблица `user_buddies` (`user_id`, `buddy_user_id`, `created_at`, `is_mutual`). Endpoints: `GET/POST/DELETE /api/users/{id}/buddies`. Применимо ко всем вселенным (J2 — общий знаменатель). По R10: per-universe изоляция (если дружба per-universe) ИЛИ глобальная (если дружба cross-universe — решение при реализации, склонность к per-universe для простоты). |

**Не входит в этот план** (отдельные задачи):
- U-011 (custom logo альянса) — нужны storage + moderation, после
  плана 57 (mail) или отдельно.
- Global mail членам — после плана 57 (mail-service с TipTap).

---

## Что НЕ делаем

- Не переносим bbcode — он выкидывается, заменяется TipTap (план 57).
- Не вводим премиум-альянс / лимиты — отдельная задача.

## Этапы (детали — при старте)

- Ф.1. ✅ Миграции БД (3 описания, alliance_ranks с permissions JSONB,
  alliance_audit_log, расширение enum до 5 значений + миграция данных).
  Коммит 99895ae230.
- Ф.2. ✅ Backend handlers + RBAC permissions (через explicit guards,
  не middleware — см. ниже) + расширенные дипстатусы + audit-log.
- Ф.3. Передача лидерства с email-подтверждением.
- Ф.4. Полнотекстовый поиск (Postgres `tsvector`).
- Ф.5. Frontend в game-nova (для uni01/uni02).
- Ф.6. Финализация.

## Ф.2 Реализация (2026-04-28)

**Новые файлы (backend):**
- `internal/alliance/permissions.go` — модель прав, builtin-fallback
  (owner=все true, member=все false), JSONB-резолвер. Helpers
  `LoadMembership` / `Has` / `HasViaPool`. 7 permissions:
  `can_invite`, `can_kick`, `can_send_global_mail`, `can_manage_diplomacy`,
  `can_change_description`, `can_propose_relations`, `can_manage_ranks`.
- `internal/alliance/audit.go` — writer `writeAuditTx` + `ListAudit`.
  18 action-констант (alliance_created, member_kicked, …),
  4 target_kind. Доступ к логу — любой член альянса.
- `internal/alliance/ranks.go` — CRUD кастомных рангов
  (List/Create/Update/Delete) + `AssignMemberRank`.
- `internal/alliance/permissions_test.go` — unit-тесты RBAC и
  normalize-helpers (8 кейсов permissionInJSON, 5 normalizeRelation,
  5 relationNeedsAccept, +3 группы).
- `pkg/metrics/alliance.go` — Prometheus counter
  `oxsar_alliance_actions_total{action,status}` (status=ok|forbidden|error).

**Изменённые файлы:**
- `internal/alliance/service.go` — расширен:
  - `GetDescriptions` / `UpdateDescriptions` (3 описания, viewer-контекст
    member|applicant|outsider); `description` legacy остаётся для R0.
  - `Kick` (с `PermKick` guard, защита от kick'а owner и self).
  - `ProposeRelation` переписан под 5 enum (ally→friend alias),
    `relationNeedsAccept` для war/hostile_neutral.
  - `AcceptRelation` / `RejectRelation` — `PermManageDiplomacy` guard.
  - audit-writes в `Create`, `SetOpen`, `Approve`, `Reject`, `Leave`,
    `ProposeRelation`, `AcceptRelation`, `RejectRelation`.
- `internal/alliance/handler.go` — 9 новых endpoints + Idempotency-Key
  на PATCH/POST + Prometheus-метрики.
- `cmd/server/main.go` — роутинг новых endpoints, `WithRedis(rdb)`
  для idempotency.
- `api/openapi.yaml` — 6 новых path-секций + 5 новых schemas
  (AllianceDescriptionView, AlliancePermissions, AllianceRank,
  AllianceAuditEntry, AllianceRelation).
- `pkg/metrics/metrics.go` — вызов `RegisterAlliance()` в `Register()`.

**Дизайн-решения:**

1. **Permissions через guards, не middleware.** Continuation-промпт
   просил «middleware-decorator», но в чистом виде это не работает:
   для большинства действий нужно сначала прочитать `alliance_id` из
   `alliance_members` (member может быть не указан в URL — например,
   `Leave` или `UpdateDescriptions`), а это работа сервиса. Поэтому
   проверка `Has` вызывается из сервисных методов внутри транзакции
   (через `LoadMembership`). Builtin owner (alliance_members.rank='owner')
   всегда имеет все права — fallback независимо от rank_id.
2. **`audit-log` пишется внутри транзакции** (`writeAuditTx`), best-effort
   (ошибка логируется, не пробрасывается). При rollback запись audit
   тоже откатится, что корректно.
3. **`alliance_audit_log` для disband не пишется**: ON DELETE CASCADE
   удалит все записи о альянсе вместе с самим альянсом — запись была
   бы немедленно потеряна. Если в будущем потребуется глобальный лог
   действий — отдельная таблица или soft-delete.
4. **`relation_needs_accept`**: war и hostile_neutral односторонние
   (можно атаковать/насильно обозначить); friend/neutral/nap двусторонние.
5. **Idempotency-Key (R9)** на PATCH/POST description, ranks CRUD,
   AssignMemberRank — критичные мутации. Не на Kick/DeleteRank
   (DELETE-операции уже идемпотентны по семантике).
6. **`viewer` поле в DescriptionView**: фронт может использовать для
   подсказки «вы видите внешнее описание; для членов есть отдельное».

**Ограничения / в Ф.3-Ф.5:**
- Endpoint `transfer-leadership` — Ф.3 (требует identity-flow).
- Полнотекстовый поиск — Ф.4 (отдельная миграция 0078 с GIN-индексом).
- UI/frontend — Ф.5.
- Финализация (divergence-log + ui-backlog) — Ф.6.

**Закрыто этой фазой:**
- D-014 (расширенные дипстатусы): backend ✅, UI откладывается на Ф.5.
- D-041, U-015 (3 описания): backend ✅, UI откладывается на Ф.5.
- U-005 (гранулярные права): backend ✅, UI откладывается на Ф.5.
- U-013 (audit-log): backend ✅, UI откладывается на Ф.5.

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
