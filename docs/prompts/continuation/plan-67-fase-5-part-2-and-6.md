# Промпт: выполнить план 67 Ф.5 ч.2 + Ф.6 (frontend audit/search/transfer + финализация)

**Дата создания**: 2026-04-28
**План**: [docs/plans/67-remaster-alliance-extension.md](../../plans/67-remaster-alliance-extension.md)
**Зависимости**: ✅ Ф.1-Ф.4 (backend закрыт), ✅ Ф.5 ч.1 (3 описания + ranks + дипстатусы UI, commit 669af55dae).
**Объём**: ~600-1000 строк frontend + ~50 строк docs, 1 коммит.

---

```
Задача: выполнить план 67 Ф.5 ч.2 (оставшийся frontend) + Ф.6
(финализация плана).

КОНТЕКСТ:

План 67 backend закрыт (Ф.1-Ф.4: миграции 0073-0077+0079+0080,
handlers, RBAC, дипстатусы 5 enum, audit-log, transfer-leadership,
полнотекстовый поиск). Frontend Ф.5 ч.1 закрыт коммитом 669af55dae:
DescriptionsPanel + RanksPanel + DiplomacyPanel (1221 строк).

Ф.5 ч.2 — оставшийся frontend:
- **Audit-log UI** (U-013): экран лога активности альянса с
  фильтрами по action_kind / target_kind / actor.
- **Поиск с фильтрами** (U-012): расширение списка альянсов
  фильтрами тип/размер/открытость + полнотекст по name/tag.
- **Transfer-leadership UI** (U-004): кнопка передачи лидерства +
  модалка с email-кодом подтверждения (через identity).

Ф.6 — финализация: divergence-log записи D-014/D-040/D-041,
ui-backlog U-004/U-005/U-012/U-013/U-015, шапка плана 67,
project-creation.txt.

R12 (i18n) — ОБЯЗАТЕЛЬНО grep по `projects/game-nova/configs/i18n/`
**и** `projects/game-nova/frontends/nova/src/i18n/` перед созданием новых
ключей. Цель — максимальное переиспользование. Ф.5 ч.1 показала ~95%
переиспользования в плане 71 — целиться в такое же число.

ПЕРЕД НАЧАЛОМ:

1) git status --short. cat docs/active-sessions.md.

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/67-remaster-alliance-extension.md (твоё ТЗ — Ф.5 ч.2 + Ф.6)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5» R0-R15
   - projects/game-nova/api/openapi.yaml — секции:
     - GET /api/alliances/{id}/audit
     - GET /api/alliances?filters... (search)
     - POST /api/alliances/{id}/transfer-leadership
     - POST /api/alliances/{id}/transfer-leadership/confirm
     (что возвращает каждый endpoint — DTO).
   - projects/game-nova/frontends/nova/src/features/alliance/DescriptionsPanel.tsx
     (эталон UI-стиля Ф.5 ч.1, commit 669af55dae)
   - projects/game-nova/frontends/nova/src/features/alliance/RanksPanel.tsx
     (эталон permissions UI)
   - projects/game-nova/frontends/nova/src/features/alliance/DiplomacyPanel.tsx
     (эталон списка с действиями)

3) Прочитай выборочно:
   - commit 669af55dae полностью (как структурированы 3 панели Ф.5 ч.1)
   - commit 556baa8483 (transfer-leadership backend — какие ошибки
     показывать в UI)
   - commit d7988572c4 (search backend — какие фильтры есть)
   - projects/game-nova/frontends/nova/src/features/admin/AuditLogPage.tsx
     или подобный (если есть, как образец audit-UI)

4) Добавь свою строку в docs/active-sessions.md:
   | <slot> | План 67 Ф.5 ч.2 + Ф.6 | projects/game-nova/frontends/nova/src/features/alliance/ docs/research/origin-vs-nova/{divergence-log,nova-ui-backlog}.md docs/plans/67-... | <дата-время> | feat(alliance,frontend): audit + search + transfer-leadership UI (план 67 Ф.5 ч.2 + Ф.6) |

ЧТО НУЖНО СДЕЛАТЬ:

### Ф.5 ч.2 — Frontend

1. **AuditLogPanel.tsx** (U-013):
   - Подключить TanStack Query: GET /api/alliances/{id}/audit
     с параметрами `?cursor=&limit=50&action=&target_kind=&actor_id=`.
   - Список записей: actor (username + ссылка) → action (i18n
     key из 18 action-констант: alliance_created, member_kicked,
     ...) → target_kind+target_id → created_at (relative).
   - Фильтры: dropdown по action_kind, dropdown target_kind,
     поиск по actor (типа autocomplete над members).
   - Пагинация — cursor-based (по существующему backend-API).
   - Доступ: любой member (backend уже проверяет).
   - i18n: используй `alliance.audit.action.<action_constant>`
     (если уже есть — переиспользуй; иначе добавь все 18).

2. **AllianceSearchPanel.tsx / расширить ListAlliancesPage.tsx**
   (U-012):
   - Существующий список альянсов получает фильтры:
     - input «Поиск» (name + tag, full-text).
     - select по типу (если есть в схеме).
     - range-input по размеру (members from/to).
     - checkbox «Только открытые» (open_for_join).
   - GET /api/alliances?q=...&min_members=...&open_only=true.
   - Debounce input 300ms.
   - i18n: используй существующие `alliance.list.*` ключи.

3. **TransferLeadershipDialog.tsx** (U-004):
   - В RanksPanel или AllianceMembersPanel — кнопка «Передать
     лидерство» (видна только owner'у).
   - Модалка: список member'ов с радио-выбором + кнопка «Запросить
     код подтверждения».
   - POST /api/alliances/{id}/transfer-leadership с
     `{new_owner_user_id}` → backend шлёт код на email через identity.
   - Поле «Введите код из email» + POST .../transfer-leadership/confirm
     с `{code, idempotency_key}`.
   - При успехе — query-invalidation alliance + members.
   - Отображение ошибок: invalid_code, expired_code,
     not_a_member, billing_unavailable.
   - i18n: `alliance.transferLeadership.*`.

4. **Интеграция в AllianceScreen** (или эквивалент):
   - Добавь Tab/Section «Журнал», «Передача лидерства» (если
     текущая структура не имеет).
   - Сохрани UI-консистентность с Ф.5 ч.1 (DescriptionsPanel +
     RanksPanel + DiplomacyPanel).

5. **Тесты frontend**:
   - Минимум по 1-2 unit-теста на каждый новый компонент через
     Vitest + React Testing Library:
     - AuditLogPanel: рендер списка, фильтр по action.
     - SearchPanel: debounce + query-invalidation.
     - TransferDialog: 2-step flow (request code → confirm).
   - Property-based — N/A для UI; основное покрытие —
     interaction-test.

### Ф.6 — Финализация

6. **divergence-log**:
   - D-014 (расширенные дипстатусы): закрыто backend Ф.2,
     UI Ф.5 ч.1 → статус ✅.
   - D-040 (transfer-leadership): backend Ф.3 + UI Ф.5 ч.2 → ✅.
   - D-041 (3 описания): backend Ф.2 + UI Ф.5 ч.1 → ✅.

7. **nova-ui-backlog**:
   - U-004 (transfer-leadership UI): ✅ Ф.5 ч.2.
   - U-005 (granular permissions UI): ✅ Ф.5 ч.1.
   - U-012 (full-text search UI): ✅ Ф.5 ч.2.
   - U-013 (audit-log UI): ✅ Ф.5 ч.2.
   - U-015 (3 описания UI): ✅ Ф.5 ч.1.

8. **Шапка плана 67**: Ф.5 ✅, Ф.6 ✅, весь план ЗАКРЫТ.

9. **project-creation.txt**: запись итерации 67 Ф.5 ч.2 + Ф.6.

══════════════════════════════════════════════════════════════════
ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА R0-R15 — см. блок rules-r0-r15.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/rules-r0-r15.md
══════════════════════════════════════════════════════════════════

══════════════════════════════════════════════════════════════════
GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ — см. блок git-isolation.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/git-isolation.md

Свои пути для CC_AGENT_PATHS:
- projects/game-nova/frontends/nova/src/features/alliance/AuditLogPanel.tsx
- projects/game-nova/frontends/nova/src/features/alliance/AuditLogPanel.test.tsx
- projects/game-nova/frontends/nova/src/features/alliance/AllianceSearchPanel.tsx
- projects/game-nova/frontends/nova/src/features/alliance/AllianceSearchPanel.test.tsx
- projects/game-nova/frontends/nova/src/features/alliance/TransferLeadershipDialog.tsx
- projects/game-nova/frontends/nova/src/features/alliance/TransferLeadershipDialog.test.tsx
- projects/game-nova/frontends/nova/src/features/alliance/AllianceScreen.tsx (если меняешь интеграцию)
- projects/game-nova/frontends/nova/src/features/alliance/ListAlliancesPage.tsx (если меняешь поиск)
- projects/game-nova/frontends/nova/src/i18n/ru.ts (только alliance.* ключи)
- projects/game-nova/frontends/nova/src/i18n/en.ts (только alliance.* ключи)
- projects/game-nova/frontends/nova/src/api/alliance.ts (если меняешь типы)
- docs/research/origin-vs-nova/divergence-log.md
- docs/research/origin-vs-nova/nova-ui-backlog.md
- docs/plans/67-remaster-alliance-extension.md
- docs/project-creation.txt
- docs/active-sessions.md
══════════════════════════════════════════════════════════════════

КОММИТЫ:

Один коммит: feat(alliance,frontend): audit + search + transfer-leadership UI (план 67 Ф.5 ч.2 + Ф.6)

Trailer: Generated-with: Claude Code

ВСЕГДА:
git commit -m "..." -- $CC_AGENT_PATHS

ЧЕГО НЕ ДЕЛАТЬ:

- НЕ трогать backend плана 67 — он закрыт. Если найдёшь баг —
  отдельная сессия / отдельный коммит.
- НЕ создавать новые i18n-ключи без grep — целиться в ~95%
  переиспользование.
- НЕ забывать про a11y (jsx-a11y eslint plugin строгий).
- НЕ использовать `any` (R1 + tsconfig strict).
- НЕ забывать про -- в git commit.

УСПЕШНЫЙ ИСХОД:

- 3 новых компонента (Audit + Search + Transfer) + ≥6 тестов.
- Интегрированы в AllianceScreen / ListAlliancesPage.
- D-014/D-040/D-041 в divergence-log → ✅.
- U-004/U-005/U-012/U-013/U-015 в ui-backlog → ✅.
- Шапка плана 67 → весь план ЗАКРЫТ ✅.
- Запись в docs/project-creation.txt.
- Удалена строка из docs/active-sessions.md.

Стартуй.
```
