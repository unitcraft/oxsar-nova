# Continuation: план 67 — Ф.2-Ф.6 (handlers, RBAC, transfer-leadership, frontend, финал)

**Применение**: вставить блок ниже в новую сессию Claude Code.
**Объём**: ~2 нед, ~600-1000 строк Go + ~400-800 строк frontend.

---

```
Задача: завершить план 67 (ремастер) — Ф.2-Ф.6. Backend handlers,
RBAC middleware, transfer-leadership, frontend, финализация.

КОНТЕКСТ:

Ф.0 дельта-аудит и Ф.1 миграции закрыты коммитом 99895ae230 —
5 миграций 0073-0077:
- 0073 alliance_descriptions (3 поля description_external/internal/apply)
- 0074 alliance_leadership_transferred_at
- 0075 alliance_ranks (с permissions JSONB)
- 0076 alliance_relations_extend (enum до 5: friend/neutral/hostile_neutral/nap/war)
- 0077 alliance_relations_rename_ally (ally → friend mapping)
+ alliance_audit_log из миграции (если в комитах есть, проверить)

Ф.0 нашёл что buddy-list уже есть как `friends` (миграция 0053) —
дубликат не создан, U-006 закрыт в иной форме.

Эта сессия: backend handlers + RBAC + transfer-leadership +
полнотекст + frontend. Финализация плана.

ПЕРЕД НАЧАЛОМ:

1) git status --short.

2) Прочитай:
   - docs/plans/67-remaster-alliance-extension.md (своё ТЗ + Ф.0
     отчёт)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5»
     (R0-R15)
   - КОММИТ 9a3992a384 (план 65 Ф.1) — эталон handler-паттерна:
     typed payload, Idempotency, Prometheus, golden, audit-log.
   - КОММИТ 99895ae230 (Ф.1 миграции) — твоя работа, что в БД.

3) Выборочно:
   - projects/game-nova/backend/internal/alliance/ — текущая
     реализация (1314 строк).
   - projects/game-nova/backend/internal/identity/ или auth/ для
     email-подтверждения (как D-003 в плане 44).
   - projects/game-nova/frontend/src/features/alliance/ — текущий
     UI.

ЧТО НУЖНО СДЕЛАТЬ:

Ф.2. Backend handlers + RBAC middleware:
- Расширить service.go под 3 описания (description_external/
  internal/apply): GET/PUT endpoints.
- alliance_ranks middleware: проверка permissions JSONB при каждом
  alliance-действии (can_invite/kick/send_global_mail/manage_diplomacy/
  change_description/propose_relations).
- Расширенные дипстатусы: enum 5 значений работает в
  AcceptRelation/ProposeRelation.
- alliance_audit_log: writes на каждое значимое действие
  (создание, изменение, передача лидерства, изгнание, и т.д.).

Ф.3. Transfer-leadership с email-подтверждением:
- Endpoint POST /api/alliances/{id}/transfer-leadership/{userId}.
- Email-подтверждение через identity-сервис (как D-003,
  account_deletion_codes).
- Idempotency-Key (R9).
- Запись в alliance_audit_log.
- alliance_leadership_transferred_at (миграция 0074) обновляется.

Ф.4. Полнотекстовый поиск:
- GIN-индекс по name+tag (миграция 0078).
- Расширить GET /api/alliances фильтрами (тип/размер/открытость) +
  полнотекст через Postgres tsvector.

Ф.5. Frontend (game-nova-frontend):
- features/alliance/: 3 описания (UI таб External/Internal/Apply).
- features/alliance/ranks/: UI создания кастомных рангов с правами
  (checkbox-форма).
- features/alliance/diplomacy/: расширенные дипстатусы (5 значений).
- features/alliance/audit/: лог активности.
- Поиск с фильтрами в alliance list.

Ф.6. Финализация:
- Шапка плана 67 → ✅ Завершён <дата>.
- docs/project-creation.txt — итерация 67.
- divergence-log.md: D-014, D-040, D-041 ✅.
- nova-ui-backlog.md: U-004, U-005, U-012, U-013, U-015 ✅.

ПРАВИЛА (R0-R15):
- R0: общий знаменатель — добавление новой функциональности, не
  правка существующих чисел.
- R1: snake_case (description_external, alliance_ranks, can_invite).
- R2: OpenAPI первым.
- R6: REST API с нуля.
- R8: Prometheus метрики.
- R9: Idempotency-Key для transfer-leadership.
- R10: per-universe изоляция (universe_id в alliance_audit_log).
- R12: i18n — grep nova-bundle перед новыми ключами.
- R15: без упрощений.

GIT-ИЗОЛЯЦИЯ:
- Свои пути: internal/alliance/, migrations/0078_alliance_search.sql,
  frontend/src/features/alliance/, openapi.yaml, тесты,
  docs/plans/67-..., divergence-log.md, nova-ui-backlog.md.

КОММИТЫ:

3-4 коммита:
1. feat(alliance): handlers + RBAC middleware + audit-log (план 67 Ф.2).
2. feat(alliance): transfer-leadership с email + Idempotency (план 67 Ф.3).
3. feat(alliance): полнотекстовый поиск + GIN индекс (план 67 Ф.4).
4. feat(alliance,frontend): UI расширения + финализация (план 67 Ф.5+Ф.6).

ЧЕГО НЕ ДЕЛАТЬ:
- НЕ дублировать buddy-list (есть как friends, миграция 0053).
- НЕ переносить bbcode (план 57 TipTap).
- НЕ вводить custom logo альянса (отдельный план).
- НЕ делать global-mail без плана 57.

УСПЕШНЫЙ ИСХОД:
- D-014, D-040, D-041 закрыты.
- U-004, U-005, U-012, U-013, U-015 закрыты.
- Все existing nova-тесты зелёные (R0).
- План 67 ✅.

Стартуй.
```
