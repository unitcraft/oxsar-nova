# Continuation: план 67 — Ф.5 frontend + Ф.6 финализация

**Применение**: вставить блок ниже в новую сессию Claude Code.
**Объём**: ~1 нед, ~400-800 строк frontend + финализация. 1-2 коммита.

---

```
Задача: завершить план 67 (ремастер) — Ф.5 frontend и Ф.6
финализация. Backend закрыт полностью (Ф.2-Ф.4).

КОНТЕКСТ — что закрыто в плане 67:

- Ф.0 дельта-аудит (коммит 99895ae230) — нашёл что buddy-list уже
  есть как `friends` (миграция 0053), отказ от дубликата user_buddies.
- Ф.1 миграции (коммит 99895ae230 + последующие):
  · 0073 alliance_descriptions (3 поля description_external/internal/apply)
  · 0074 alliance_ranks (с permissions JSONB)
  · 0075 alliance_audit_log
  · 0076 alliance_relations_extend (enum 5: friend/neutral/hostile_neutral/nap/war)
  · 0077 alliance_relations_rename_ally
  · 0079 alliance_leadership_codes (transfer-leadership, переименована
    из 0078 после конфликта с пост-фиксом плана 69)
  · 0080 alliance GIN-индекс полнотекстового поиска
- Ф.2 (коммит 2fd010cd87) — handlers + RBAC permissions через guards
  + 5 дипстатусов + audit-log writes (~870 строк, 4 trade-off в
  simplifications.md).
- Ф.3 (коммит 556baa8483) — transfer-leadership с email-кодом через
  identity + Idempotency-Key (transfer.go, transfer_handler.go,
  transfer_test.go, OpenAPI, i18n).
- Ф.4 (коммит d7988572c4) — полнотекстовый поиск (GIN-индекс,
  ListFilters, тесты).

Backend плана 67 ПОЛНОСТЬЮ закрыт. D-014, D-040, D-041 закрыты.
U-005, U-013, U-015 закрыты backend-частью. Осталось:
- **Ф.5 frontend** (~400-800 строк React в game-nova-frontend) —
  UI для всего, что добавлено в backend.
- **Ф.6 финализация** — пометки в divergence-log, nova-ui-backlog,
  шапка плана 67 ✅.

ПЕРЕД НАЧАЛОМ:

1) git status --short — если есть чужие правки от параллельных
   сессий (план 77 backend, план 71 UX-микрологика, план 66 golden),
   бери только свои файлы. ВСЕГДА `git commit -m "..." -- path1
   path2` с двойным тире (4 раза прецедентов в memory).

2) Прочитай:
   - docs/plans/67-remaster-alliance-extension.md (своё ТЗ + Ф.0
     отчёт + 4 trade-off из Ф.2 в simplifications.md)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5»
     (R0-R15 + R15 уточнено: пропуск vs trade-off)
   - КОММИТЫ 2fd010cd87 (Ф.2 handlers/RBAC), 556baa8483 (Ф.3 transfer),
     d7988572c4 (Ф.4 GIN-поиск) — эталон для frontend-API.

3) Прочитай выборочно:
   - projects/game-nova/api/openapi.yaml — все новые alliance-endpoints
     уже описаны (генерация TS-клиента — образец).
   - projects/game-nova/frontend/src/features/alliance/ — текущий
     UI alliance, расширяешь.
   - projects/game-nova/frontend/src/features/ — стиль компонентов
     (TanStack Query, Zustand, TS strict).
   - projects/game-nova/configs/i18n/ — обязательный grep перед
     созданием новых ключей (R12 уточнено).
   - projects/game-origin-php/src/templates/standard/ally*.tpl —
     UX-референс (НЕ для копирования кода, для понимания поведения).

ЧТО НУЖНО СДЕЛАТЬ:

Ф.5. Frontend в projects/game-nova/frontend/src/features/alliance/:

5.1. Три описания альянса (D-041, U-015):
- Tab-компонент в AlliancePage: External / Internal / Apply.
- Видимость по контексту:
  · Не член — только External.
  · Член — External + Internal.
  · При apply — External + Apply.
- PATCH /api/alliances/{id}/descriptions с Idempotency-Key.
- Permission check на frontend: только если can_change_description.

5.2. Ranks UI с permissions checkbox-формой (D-014, U-005):
- Новый экран AllianceRanksPage (`features/alliance/ranks/`).
- Список рангов (table) с position-сортировкой.
- Создание ранга: form с input "name" + 7 checkbox'ов
  (can_invite, can_kick, can_send_global_mail,
   can_manage_diplomacy, can_change_description,
   can_propose_relations, can_manage_ranks).
- Редактирование ранга: тот же form, prefilled.
- Удаление с confirm-modal.
- Permission check: только can_manage_ranks.

5.3. Дипломатия 5 статусов (D-014, B1):
- Расширить AllianceDiplomacyPage:
  · Tabs или filter: Friends / Neutral / Hostile-Neutral / NAP / War.
  · Status-badge с цветом per-статус (зелёный/серый/жёлтый/синий/красный).
  · Action-кнопки: Propose / Accept / Reject / Break (зависит от
    текущего статуса и permissions can_manage_diplomacy).
- POST /api/alliances/{id}/relations + accept/reject endpoints.

5.4. Audit log UI (U-013):
- Новый экран AllianceAuditPage (`features/alliance/audit/`).
- Таблица с timestamp, actor (user), action, target, details.
- Cursor-pagination (R6).
- Filter по action-type (создание, изгнание, передача, etc.).
- Permission check: видно только членам альянса.

5.5. Поиск с фильтрами (U-012):
- Расширить AllianceListPage:
  · Search-input (полнотекстовый по name+tag через
    GET /api/alliances?search=...).
  · Filter dropdowns: тип / размер / открытость.
  · Cursor-pagination для длинных списков.

5.6. Transfer-leadership UI (D-040, U-004):
- На AlliancePage (для owner) — кнопка "Передать лидерство".
- Modal: выбор member из dropdown → POST /transfer-leadership/code
  → отображение "код отправлен на email".
- Form для confirmation: input "code" → POST /transfer-leadership.
- Idempotency-Key (R9) — обязательно.

Все компоненты:
- TanStack Query (queryKey, mutationFn).
- Zustand для UI-state (modal open/close, фильтры).
- TS strict, никакого any.
- Импорт DTO из openapi-генератора (R2 первым).
- Tailwind + nova-tema (CSS-переменные, не хардкод цветов).

R12 ОБЯЗАТЕЛЬНО:
- Перед созданием каждой строки grep по configs/i18n/ru.yml +
  configs/i18n/en.yml. Многие alliance-строки УЖЕ ЕСТЬ.
- Идентификаторы legacy-PHP na_phrases НЕ переносить — ключи в
  nova-стиле (alliance.transfer.confirm и т.п.).
- В коммите указать соотношение переиспользовано/новых ключей.

Ф.6. Финализация:

- Шапка плана 67 → ✅ Завершён <дата>.
- docs/project-creation.txt — итерация 67.
- В divergence-log.md ✅:
  · D-014 (granular ranks + 5 дипстатусов)
  · D-040 (transfer-leadership)
  · D-041 (3 описания)
- В nova-ui-backlog.md ✅:
  · U-004 (transfer-leadership UI)
  · U-005 (ranks UI)
  · U-012 (полнотекстовый поиск)
  · U-013 (audit log UI)
  · U-015 (3 описания UI)
- При желании — отметить в roadmap-report «Часть V.5» или Part VIII
  что план 67 закрыт.

ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА (R0-R15):

R0: общий знаменатель — UI работает для всех вселенных
(uni01/uni02/origin), не только origin.
R2: DTO из OpenAPI-сгенерированного клиента (никакие manual types).
R5: pixel-perfect только в плане 72; здесь nova-стиль.
R12: i18n переиспользование — критично для большой alliance-фичи.
   Большая часть строк уже есть в bundle.
R15: без упрощений — все 5 sub-экранов реализованы. Не «MVP с
   3 описаниями, остальное потом».

R15 УТОЧНЕНО:
🚫 НЕ КЛАССИФИЦИРУЙ КАК TRADE-OFF:
- Хардкод строки (R12) — обязателен Tr() из i18n bundle.
- Пропуск Idempotency-Key на mutation (PATCH/POST/DELETE) — нет.
- Пропуск Permission-check на frontend (потому что backend всё
  равно проверит) — нет, frontend должен скрывать кнопки которые
  не позволены, иначе плохой UX.
- Пропуск тестов на новые компоненты (Vitest для logic, Playwright
  для smoke) — нет.

✅ TRADE-OFF (можно с обоснованием):
- Аккордеон вместо tabs — UI-выбор, ок если обоснован.
- Один большой modal vs несколько маленьких — UI-выбор.
- Если frontend не успевает за одну сессию — split на 2 коммита
  (например, 5.1+5.2+5.3 → коммит 1, 5.4+5.5+5.6 → коммит 2).

GIT-ИЗОЛЯЦИЯ (4 раза граблей в memory!):
- Свои пути:
  · projects/game-nova/frontend/src/features/alliance/ (всё)
  · projects/game-nova/frontend/src/features/alliance/ranks/ (новое)
  · projects/game-nova/frontend/src/features/alliance/audit/ (новое)
  · projects/game-nova/configs/i18n/ru.yml + en.yml (только новые
    ключи если потребуются)
  · docs/plans/67-..., docs/project-creation.txt,
    docs/research/origin-vs-nova/divergence-log.md,
    docs/research/origin-vs-nova/nova-ui-backlog.md
- ВСЕГДА `git commit -m "..." -- path1 path2 path3` с двойным
  тире.
- Перед commit: git status --short + git diff --cached --name-only.
- НИКОГДА git add . / git add -A / git commit -m без `--`.

КОММИТЫ:

1-2 коммита:
1. feat(alliance,frontend): UI расширения для ремастера
   (план 67 Ф.5).
2. (опц.) docs(plan-67): финализация (план 67 Ф.6).

ИЛИ один: feat(alliance,frontend): UI + финализация (план 67 Ф.5+Ф.6).

В коммит-сообщении — соотношение переиспользовано/новых i18n-ключей.

ЧЕГО НЕ ДЕЛАТЬ:
- НЕ переписывать существующие alliance-страницы (Create,
  AlliancePage базовый, Members) — расширяй tabs/секциями.
- НЕ создавать дубликат buddy-list — это already есть как `friends`
  (см. план 67 Ф.0 дельта-аудит).
- НЕ дублировать i18n-ключи — обязательный grep R12.
- НЕ менять modern-числа nova (R0).
- НЕ забывать про `git commit -- path` (4 раза граблей).

УСПЕШНЫЙ ИСХОД:
- 5 sub-экранов реализованы и протестированы.
- D-014, D-040, D-041 закрыты ✅.
- U-004, U-005, U-012, U-013, U-015 закрыты ✅.
- В коммите указано соотношение i18n-ключей.
- Все existing nova-тесты зелёные (R0).
- План 67 ✅ полностью.

Стартуй.
```
