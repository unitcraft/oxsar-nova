# Промпт: выполнить план 67 (расширение alliance-системы)

**Дата создания**: 2026-04-28
**План**: [docs/plans/67-remaster-alliance-extension.md](../plans/67-remaster-alliance-extension.md)
**Зависимости**: нет критичных (можно параллелить с 64).
**Объём**: 2-3 нед, ~600-1000 строк Go + ~400-800 строк frontend.

---

```
Задача: выполнить план 67 (ремастер) — расширение alliance-системы
nova до паритета с oxsar2-classic + общий знаменатель для всех
вселенных.

ВАЖНОЕ:
- Применимо ко ВСЕМ вселенным (uni01/uni02/origin) — это «общий
  знаменатель», не origin-only.
- Можно делать параллельно с другими планами серии (64, 68, 71).

ПЕРЕД НАЧАЛОМ:

1) git status --short.

2) Прочитай:
   - docs/plans/67-remaster-alliance-extension.md (твоё ТЗ)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5»
     (R0-R15)
   - docs/research/origin-vs-nova/divergence-log.md D-014, D-040, D-041
   - docs/research/origin-vs-nova/nova-ui-backlog.md U-004, U-005,
     U-006, U-012, U-013, U-015

3) Прочитай выборочно:
   - projects/game-nova/backend/internal/alliance/ — текущая
     реализация (1314 строк, 16 endpoints).
   - projects/game-origin-php/src/game/page/Alliance.class.php
     (1413 строк) и AllianceList.class.php (294) — referenc.

ЧТО НУЖНО СДЕЛАТЬ:

1. БД-миграции (R1: snake_case, _at для timestamps, _id для FK):
   - alliances: добавить description_external, description_internal,
     description_apply (TEXT, nullable).
   - alliance_ranks: новая таблица с permissions JSONB:
     can_invite, can_kick, can_send_global_mail, can_manage_diplomacy,
     can_change_description, can_propose_relations.
   - alliance_audit_log: журнал активности альянса (по образцу
     admin_audit_log плана 14).
   - alliance_relations: enum дипстатусов расширить до 5 значений:
     friend / neutral / hostile_neutral / nap / war (B1 решение).
     Миграция nova-3-статусов: nova-friend→friend, nova-neutral→
     neutral, nova-war→war. hostile_neutral и nap — новые.
   - user_buddies: новая таблица buddy-list (user_id, buddy_user_id,
     created_at, is_mutual). По R10: per-universe (universe_id) ИЛИ
     глобальная — решить, склонность к per-universe.

2. Backend handlers + RBAC:
   - 3 описания: PATCH /api/alliances/{id}/descriptions.
   - Передача лидерства: POST /api/alliances/{id}/transfer-leadership/{userId}
     + email-подтверждение через identity (как D-003), Idempotency-Key (R9).
   - Гранулярные права: middleware проверяет permission JSONB при
     каждом alliance-действии.
   - Расширенные дипстатусы — endpoints AcceptRelation/ProposeRelation
     поддерживают все 5.
   - Полнотекстовый поиск: GET /api/alliances?search=... + фильтры
     (тип/размер/открытость) через Postgres tsvector.
   - Альянсный лог: writes в alliance_audit_log на каждое действие.
   - Buddy-list: GET/POST/DELETE /api/users/{id}/buddies.

3. Frontend (game-nova-frontend):
   - Расширить features/alliance/ под новые поля.
   - Новый экран buddy-list (features/buddies/).

ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА:

R0: общий знаменатель — добавление новой функциональности, не
правка существующих чисел/механик.
R1: все имена snake_case по nova-стилю (description_external, не
descExt; alliance_ranks мн.ч.; can_invite в JSONB).
R2: OpenAPI первым.
R8: Prometheus метрики.
R9: Idempotency-Key для transfer-leadership.
R10: per-universe изоляция (universe_id в alliance_audit_log).
R12: i18n — grep nova-bundle перед созданием новых ключей.
R15: без упрощений — тесты, обработка ошибок, метрики.

GIT-ИЗОЛЯЦИЯ:
- Свои пути: internal/alliance/, internal/buddies/ (новый),
  migrations/00NN_*, frontend/src/features/alliance/,
  frontend/src/features/buddies/, openapi.yaml, тесты,
  docs/plans/67-..., divergence-log.md.

ЧЕГО НЕ ДЕЛАТЬ:
- НЕ переносить bbcode (план 57 TipTap заменит).
- НЕ вводить custom logo альянса (U-011) — отдельный план после
  storage/moderation.
- НЕ делать global-mail рассылку без плана 57 (mail-service).

КОММИТЫ:

2-3 коммита:
1. feat(alliance): миграции БД + 3 описания + buddy-list таблицы
2. feat(alliance): handlers + RBAC + диплома́тия 5 статусов
3. feat(alliance): frontend расширения + buddy UI

УСПЕШНЫЙ ИСХОД:
- D-014, D-040, D-041 закрыты.
- U-004, U-005, U-006, U-012, U-013, U-015 закрыты.
- 5 enum дипстатусов в alliance_relations.
- buddy-list работает.
- Все existing nova-тесты зелёные (R0).

Стартуй.
```
