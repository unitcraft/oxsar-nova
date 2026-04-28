# Промпт: выполнить план 72 Ф.3 — Spring 2 (10 alliance/resource/market экранов origin-фронта)

**Дата создания**: 2026-04-28
**План**: [docs/plans/72-remaster-origin-frontend-pixel-perfect.md](../../plans/72-remaster-origin-frontend-pixel-perfect.md)
**Зависимости**: ✅ Ф.1 Bootstrap (54fabbdf46), ✅ Ф.2 Spring 1 (47d1f0ef65 —
7 главных экранов готовы), ✅ план 67 (alliance backend), ✅ план 78
(раскладка). Параллелится с планом 76 (nova exchange UI) — разные
папки.
**Объём**: ~2000-3000 строк TS + CSS + i18n, 1-2 коммита.

---

```
Задача: реализовать Spring 2 плана 72 — 10 экранов origin-фронта,
группа alliance/resource/market/repair/battlestats/fleet, в pixel-
perfect клоне legacy.

КОНТЕКСТ:

Ф.1 Bootstrap закрыт (54fabbdf46). Ф.2 Spring 1 закрыт коммитом
47d1f0ef65 — 7 главных экранов работают (Main, Constructions,
Research, Shipyard, Galaxy, Mission, Empire), router + API-инфра
на месте, 8 simplifications.md записей (P72.S1.A-H) фиксируют
недостающие openapi-endpoints как backend-долг.

Spring 2 экраны (из docs/research/origin-vs-nova/origin-ui-replication.md):
- S-008..S-019 — alliance (12 шаблонов в legacy):
  - Alliance overview / list / search
  - Alliance create / join / apply
  - Alliance members / ranks / permissions
  - Alliance descriptions (3 типа: external/internal/apply)
  - Alliance diplomacy (5 enum статусов)
  - Alliance audit-log
  - Alliance transfer-leadership
- S-020 Resource (рынок ресурсов — обмен металл/кристалл/водород)
- S-021 Market (артефактный market — старый, не биржа из плана 68;
  это EXT_MODE рынок из legacy за credit)
- S-022 Repair (ремонт повреждённых юнитов на планете)
- S-023 Battlestats (статистика боёв игрока)
- S-024 Fleet operations (управление активным флотом — отзыв,
  переадресация)

ВАЖНОЕ ПРО ПЛАН 67:

Backend alliance ПОЛНОСТЬЮ закрыт планом 67 (Ф.1-Ф.6, последний
коммит a149594306). Все нужные endpoint'ы существуют в openapi.yaml:
- 3 описания (D-041)
- ranks с granular permissions JSONB (U-005)
- audit-log (U-013)
- transfer-leadership с email-кодом (U-004)
- полнотекстовый поиск (U-012)
- 5 enum дипстатусов (D-014)

Используй их напрямую. Не нужно описывать заглушки в
simplifications.md (как было для S-001..S-007 где backend
не доделан) — здесь backend готов.

В nova-фронте (frontends/nova) уже есть 6 alliance-компонентов
от плана 67 — DescriptionsPanel, RanksPanel, DiplomacyPanel,
AuditLogPanel, AllianceSearchPanel, TransferLeadershipDialog.
Это **референс** для логики (TanStack Query keys, endpoints, UX-flow),
НО не визуальный референс — origin-фронт зеркалит legacy *.tpl,
не nova-стиль.

ВАЖНО ПРО R5 + ADR-0011:
- pixel-perfect клон legacy — HTML-структура и CSS-классы зеркалят
  *.tpl + style.css.
- Display name «Oxsar Classic» (применено в Spring 1 в TopHeader/
  Footer/title). Ничего нового добавлять не нужно.

ПЕРЕД НАЧАЛОМ:

1) git status --short. cat docs/active-sessions.md. Параллельно
   запускается план 76 (nova exchange UI) — разные папки. Не
   пересекайся с slot 76 на openapi.yaml (вы оба не должны его
   трогать).

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/72-remaster-origin-frontend-pixel-perfect.md
   - docs/research/origin-vs-nova/origin-ui-replication.md секции
     S-008..S-024 (детальное описание каждого экрана)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5» R0-R15
   - projects/game-nova/api/openapi.yaml — секции:
     - /api/alliances/* (план 67 backend)
     - /api/resources/* и /api/market/* (для S-020, S-021)
     - /api/units/repair (S-022)
     - /api/battlestats (S-023)
     - /api/missions (S-024 fleet operations)

3) Прочитай выборочно:
   - projects/game-nova/frontends/origin/src/main.tsx (роутер)
   - projects/game-nova/frontends/origin/src/api/ (API-инфра Spring 1)
   - projects/game-nova/frontends/origin/src/features/main/ (эталон)
   - projects/game-nova/frontends/nova/src/features/alliance/ —
     ВСЕ 6 компонентов (DescriptionsPanel, RanksPanel, DiplomacyPanel,
     AuditLogPanel, AllianceSearchPanel, TransferLeadershipDialog) —
     эталон ЛОГИКИ для alliance-экранов (НЕ визуала)
   - projects/game-legacy-php/templates/alliance*.tpl (12 шаблонов)
   - projects/game-legacy-php/templates/{resource,market,repair,
     battlestats,fleet*}.tpl

4) Добавь свою строку в docs/active-sessions.md:
   | <N> | План 72 Ф.3 Spring 2 (10 alliance+ экранов origin) | projects/game-nova/frontends/origin/ | <дата-время> | feat(origin/frontend): Ф.3 Spring 2 — alliance + resource + market |

ЧТО НУЖНО СДЕЛАТЬ:

### Группа 1: Alliance (12 шаблонов — S-008..S-019)

Каркас в `src/features/alliance/`:
- `AllianceOverviewScreen.tsx` (S-008) — главный экран альянса
  для члена: 3 описания (член видит external + internal),
  список членов с рангами, дипстатусы, audit-feed (последние N
  записей).
- `AllianceListScreen.tsx` (S-009) — список всех альянсов
  с фильтрами (полнотекстовый поиск plus filters: тип/размер/
  открытость).
- `AllianceCreateScreen.tsx` (S-010) — форма создания.
- `AllianceJoinScreen.tsx` / `AllianceApplyScreen.tsx` (S-011) —
  заявка на вступление (с external description + apply description).
- `AllianceMembersScreen.tsx` (S-012) — таблица членов с рангами.
- `AllianceRanksScreen.tsx` (S-013) — управление рангами +
  permissions JSONB (только owner/админ).
- `AllianceDescriptionsScreen.tsx` (S-014) — редактирование 3
  описаний (только с правом can_change_description).
- `AllianceDiplomacyScreen.tsx` (S-015) — список дипстатусов с
  другими альянсами + действия (propose/accept/reject) с правом
  can_manage_diplomacy / can_propose_relations.
- `AllianceAuditLogScreen.tsx` (S-016) — журнал с фильтрами
  (action, target_kind, actor).
- `AllianceTransferLeadershipScreen.tsx` (S-017) — двушаговая
  форма (выбор member → email-код подтверждения).
- `AllianceSettingsScreen.tsx` (S-018) — настройки (open_for_join
  toggle, и пр.).
- `AllianceSearchScreen.tsx` (S-019) — расширенный поиск (если
  не помещается в S-009).

API: используй endpoints плана 67 как есть. Логику смотри в
nova-аналогах (frontends/nova/src/features/alliance/) — она там
покрыта. Visualизация: pixel-perfect legacy *.tpl.

### Группа 2: Resource market (S-020)

`src/features/resource/ResourceMarketScreen.tsx`:
- Обмен ресурсов металл/кристалл/водород/тёмная материя по фикс-
  курсу или выставление лотов (см. legacy resource.tpl).
- API: GET /api/resources/exchange-rates, POST /api/resources/exchange.

### Группа 3: Artefact market — старый legacy (S-021)

`src/features/market/MarketScreen.tsx`:
- Это **EXT_MODE legacy market** — продажа артефактов за `credit`
  (см. миграцию 0013_artefact_market.sql, internal/market в
  game-nova). Это НЕ та биржа из плана 68 — это другой механизм
  (фиксированный pricelist, не P2P).
- В legacy и в nova реально работает. Реализуй UI как клон
  legacy market.tpl.
- API: GET /api/market/offers, POST /api/market/buy.

### Группа 4: Repair (S-022)

`src/features/repair/RepairScreen.tsx`:
- Ремонт повреждённых юнитов на планете.
- API: GET /api/units/{planetId}/damaged, POST /api/units/repair.
- Idempotency-Key обязателен (R9).

### Группа 5: Battlestats (S-023)

`src/features/battlestats/BattlestatsScreen.tsx`:
- Личная статистика игрока: побед/поражений, уничтожено юнитов,
  потеряно, очков получено и т.д.
- API: GET /api/battlestats (или /api/users/{id}/battlestats).

### Группа 6: Fleet operations (S-024)

`src/features/fleet/FleetOperationsScreen.tsx`:
- Управление активными миссиями: список, отзыв, переадресация.
- API: GET /api/missions?status=active, POST /api/missions/{id}/recall.
- Idempotency-Key.

### Архитектурно (общее)

1. **Routes** — добавь в роутере `/alliance`, `/alliance/:id`,
   `/alliance/list`, `/alliance/create`, ..., `/resource-market`,
   `/market`, `/repair`, `/battlestats`, `/fleet-operations`.

2. **API-клиент** — расширь существующие модули src/api/ либо
   создай новые: alliance.ts, resource.ts, market.ts, repair.ts,
   battlestats.ts, fleet-operations.ts. Query-keys в query-keys.ts.

3. **i18n (R12)** — grep по configs/i18n/{ru,en}.yml на нужные
   термины. План 67 backend уже добавил много alliance-ключей.
   Цель ≥ 95% переиспользование.

4. **Pixel-perfect** — HTML+CSS-классы зеркалят legacy *.tpl. Точная
   визуальная сверка отложена на план 73 (screenshot-diff CI).

5. **Тесты** — vitest + RTL, минимум 1-2 теста на каждый экран
   (рендер + ключевое действие). Spring 2 = 10 экранов = ~12-15
   тестов минимум.

6. **simplifications.md** — если backend endpoint отсутствует:
   отметь как P72.S2.A/B/C... mock + TODO в коде.

### Финализация Ф.3

- Шапка плана 72: Ф.3 ✅, Spring 2 закрыт.
- Запись итерации в docs/project-creation.txt («72 Ф.3 — Spring 2»).
- В коммите: соотношение i18n переиспользовано/новых.
- НЕ закрываешь весь план 72 — впереди Ф.4 (Spring 3),
  Ф.5 (Spring 4), Ф.6 (Spring 5), Ф.7 (i18n рус), Ф.8 (TipTap чат),
  Ф.9 (финал).

══════════════════════════════════════════════════════════════════
ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА R0-R15 — см. блок rules-r0-r15.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/rules-r0-r15.md

Особо важно:
- R0: backend nova не меняем (только frontend этой сессии).
- R5: pixel-perfect для origin-фронта.
- R9: Idempotency-Key на repair и fleet-operations.
- R12: i18n grep сначала, цель 95%.
══════════════════════════════════════════════════════════════════

══════════════════════════════════════════════════════════════════
GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ — см. блок git-isolation.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/git-isolation.md

Свои пути для CC_AGENT_PATHS:
- projects/game-nova/frontends/origin/

  (вся папка целиком — твой territory; план 76 параллельно работает
   в frontends/nova/, не пересечётся; backend не задеваем)

- docs/plans/72-remaster-origin-frontend-pixel-perfect.md
- docs/active-sessions.md
- docs/project-creation.txt (запись итерации)
- docs/simplifications.md (если есть новые P72.S2.* записи)

ВНИМАНИЕ: НЕ трогай:
- projects/game-nova/frontends/nova/ — план 76 идёт параллельно.
- projects/game-nova/backend/ — не нужно.
- projects/game-nova/api/openapi.yaml — backend закрыт, расширения
  отметить в simplifications.md.
══════════════════════════════════════════════════════════════════

КОММИТЫ:

1-2 коммита:

1) feat(origin/frontend): Ф.3 Spring 2 — alliance + resource + market
   (план 72)

ИЛИ если объём > 3000 строк:

1) feat(origin/frontend): Ф.3 Spring 2 ч.1 — 12 alliance экранов
2) feat(origin/frontend): Ф.3 Spring 2 ч.2 — resource + market +
   repair + battlestats + fleet-ops + финализация

Trailer: Generated-with: Claude Code
ВСЕГДА с двойным тире:
git commit -m "..." -- $CC_AGENT_PATHS

ЧЕГО НЕ ДЕЛАТЬ:

- НЕ менять nova-фронт (там работает план 76).
- НЕ менять backend.
- НЕ менять openapi.yaml.
- НЕ переносить рекламу/баннеры.
- НЕ закрывать весь план 72 — закрываешь Ф.3.
- НЕ забывай Idempotency-Key.
- НЕ забывай про -- в git commit.

УСПЕШНЫЙ ИСХОД:

- 10+ экранов работают (router → переход → каждый рендерится).
- typecheck + build + tests зелёные.
- Все экраны pixel-perfect клоны legacy.
- i18n: 95%+ переиспользования.
- Шапка плана 72: Ф.3 ✅ (Spring 2).
- Запись в docs/project-creation.txt.
- Удалена строка из docs/active-sessions.md.

Стартуй ТОЛЬКО когда план 68 финализирован (Ф.7) — иначе backend
exchange может быть не до конца готов и frontends/nova/exchange
(параллельный план 76) залипнет на этом.
```
