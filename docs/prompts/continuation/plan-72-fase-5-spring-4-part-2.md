# Промпт: выполнить план 72 Ф.5 Spring 4 ч.2 — premium + static + utilities + финализация

**Дата создания**: 2026-04-28
**План**: [docs/plans/72-remaster-origin-frontend-pixel-perfect.md](../../plans/72-remaster-origin-frontend-pixel-perfect.md)
**Зависимости**: ✅ Ф.5 Spring 4 ч.1 (716b5a518b — 7 экранов
communication+notes+search+settings + полное OpenAPI для всех 14
Spring 4 экранов).
**Объём**: ~1200-1700 строк TS + i18n + simplifications, 1 коммит,
3-5 часов одной сессии.

---

```
Задача: завершить Spring 4 — реализовать оставшиеся 6 экранов
origin-фронта + закрыть S-046 Widgets как дубликат S-001 Main.
После этого финализировать шапку плана 72 Ф.5 (Spring 4 ✅).

КОНТЕКСТ:

Spring 4 ч.1 закрыт коммитом 716b5a518b (7 экранов: Friends/MSG/
Chat/ChatAlly/Notepad/Search/Settings + 620 строк openapi-расширения
для ВСЕХ 14 экранов). Это значит для ч.2 OpenAPI **уже
задокументирован** — endpoints для officers, professions, и
прочее уже есть в openapi.yaml (см. строки ~1317, 3251, 3268,
3370 для officers/professions). Тебе НЕ надо расширять openapi.yaml
в ч.2 (за исключением мелких корректировок если что-то реально
отсутствует).

Оставшиеся 6 экранов Spring 4:

Premium (2):
- S-040 Officer (наём officer-типов за credit/oxsariты)
- S-041 Profession (выбор профессии раз в N дней)

Static (3):
- S-043 UserAgreement (статическая страница пользовательского
  соглашения)
- S-044 Changelog (история обновлений)
- S-045 Support (форма обращения в поддержку — CROSS-SERVICE на
  portal-backend, не game-nova)

Utilities (1):
- S-047 AdvTechCalculator (advanced tech calculator — клиентская
  утилита, без backend)

И финализация:
- S-046 Widgets — записать в simplifications.md как «закрыт через
  S-001 Main» (skip-trade-off, не реализуется как отдельный экран).
- Шапка плана 72: Ф.5 ✅ полностью (после ч.1+ч.2 = 13 экранов
  реализованы, 14-й = S-046 Widgets закрыт через дубликат).
- project-creation.txt запись итерации Ф.5 ч.2.

ВАЖНО ПРО S-045 SUPPORT — CROSS-SERVICE:

План 56 (✅ закрыт коммитами 37ae65b430+63c27b1bed+86cc35d355+
6cff366aaf) **перенёс** reports из game-nova в portal-backend.
endpoint /api/reports НЕ существует в game-nova-backend.

Origin-фронт делает кросс-сервисный POST на
`${VITE_PORTAL_BASE_URL}/api/reports` — точно так же как nova-фронт
делает в ReportButton.tsx (см. план 56 коммит 37ae65b430).

В origin-фронте нужно:
1. Vite ENV `VITE_PORTAL_BASE_URL`. Если ещё не объявлена в
   Ф.1 Bootstrap — добавь в `.env.example` и vite.config.ts.
2. В src/api/support.ts — fetch на portal-host, не на game-nova-host.
3. БЕЗ Idempotency-Key (portal-backend сам управляет
   дедупликацией).
4. В openapi.yaml game-nova **НЕ** добавляй /api/reports — это
   не его endpoint. Если хочешь зафиксировать кросс-сервисную
   зависимость — комментарий в support.ts сверху + запись в
   simplifications.md как разъяснение (не trade-off).

ВАЖНО ПРО S-046 WIDGETS:

Legacy templates/widgets.tpl существует, но семантически это
**сборка на главной из других данных** (производство, очередь
стройки, последние сообщения, статистика империи). В origin-фронте
эти данные уже агрегированы в S-001 Main (Spring 1).

Уточни templates/widgets.tpl при разведке. Если действительно
дубликат Main — записать в simplifications.md:

```
## P72.S4.WIDGETS — S-046 Widgets закрыт через S-001 Main

**Where**: legacy templates/widgets.tpl содержит виджеты которые
в origin-фронте уже агрегированы в S-001 MainScreen (Spring 1
плана 72, коммит 47d1f0ef65).

**Why**: дубликат — нет смысла иметь /widgets и / отдельными
маршрутами с одинаковым контентом. Современный паттерн —
единая «главная» с виджетами на ней.

**Trade-off (R15 ✅, не упрощение)**: визуальное расхождение с
legacy. В Spring 1 уже зафиксировано «pixel-perfect только в
рамках реализуемых экранов; semantic equivalence важнее
визуальной точности дубликатов».

**Where to apply**: routes для /widgets либо НЕ создаются (404),
либо redirect на /. Лучше — redirect с notice в console.dev
«S-046 deprecated, see S-001».
```

Если templates/widgets.tpl при чтении окажется НЕ дубликатом —
реализуй экран и не записывай в simplifications.md.

ВАЖНО ПРО S-044 CHANGELOG:

Уточни в legacy templates/changelog.tpl — это:
(а) статический MD/HTML файл во frontend (markdown bundled at
    build-time, как docs/release-notes-style),
(б) endpoint /api/changelog который возвращает список релизов
    из БД,
(в) hardcoded в *.tpl с версиями.

В nova-стеке скорее всего (а) или (в). Если backend endpoint
отсутствует — реализуй как frontend-bundled MD-файл (markdown в
src/features/changelog/CHANGELOG.md, рендеринг через `react-
markdown` или существующий MD-renderer; лимит ~500 строк, не
endpoint).

Запись в simplifications.md как «P72.S4.CHANGELOG: статический
markdown в frontend, не backend-endpoint» — это не trade-off
(это правильный паттерн для редко-меняющегося контента), просто
факт.

ВАЖНО ПРО S-043 UserAgreement:

Юр-документ. Источник истины — портал (portal-frontend). В
origin-фронте либо:
(а) cross-link на портальную страницу `${VITE_PORTAL_BASE_URL}/
    user-agreement`,
(б) дубликат текста в origin-фронте (markdown bundled).

Я бы выбрал **(а)** — централизация юр-документов на портале,
один источник истины. В origin-фронте /user-agreement рисует
inline-iframe или редиректит. **Уточни предпочтение пользователя**
если сомневаешься (либо реализуй (а) и пометь в simplifications
как «cross-portal link, единственный источник»).

ПЕРЕД НАЧАЛОМ:

ПЕРВЫМ ДЕЙСТВИЕМ (до любого чтения плана):

1) git status --short. cat docs/active-sessions.md.

2) ОБЯЗАТЕЛЬНО добавь свою строку в раздел «Активные сессии»:
   | <N> | План 72 Ф.5 Spring 4 ч.2 (6 экранов origin: premium+static+utilities + Widgets-skip + финализация) | projects/game-nova/frontends/origin/, docs/plans/72-remaster-origin-frontend-pixel-perfect.md, docs/active-sessions.md, docs/project-creation.txt, docs/simplifications.md | <дата-время> | feat(origin/frontend): Ф.5 Spring 4 ч.2 — premium+static+utilities (план 72) |

3) Закоммить slot отдельным коммитом (правило _blocks/git-isolation.md):
   git add docs/active-sessions.md
   git commit -m "chore(sessions): slot N — план 72 Ф.5 Spring 4 ч.2" -- docs/active-sessions.md

ТОЛЬКО ПОСЛЕ — переходи к чтению:

4) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/72-remaster-origin-frontend-pixel-perfect.md
   - docs/research/origin-vs-nova/origin-ui-replication.md секции
     S-040, S-041, S-043, S-044, S-045, S-046, S-047
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5» R0-R15

5) Прочитай выборочно:
   - projects/game-nova/api/openapi.yaml — секции
     /api/officers (1317-1334), /api/professions (3251, 3268).
     Профession endpoint протокол: GET /api/professions (список),
     GET /api/professions/me (текущий выбор), POST /api/professions
     (выбрать). Уточни DTO.
   - projects/game-legacy-php/src/templates/standard/officer.tpl
   - projects/game-legacy-php/src/templates/standard/profession.tpl
   - projects/game-legacy-php/src/templates/standard/user_agreemet.tpl
     (typo в legacy — `agreemet` без `n`, не `agreement`)
   - projects/game-legacy-php/src/templates/standard/changelog.tpl
   - projects/game-legacy-php/src/templates/standard/support.tpl
   - projects/game-legacy-php/src/templates/standard/widgets.tpl
   - projects/game-nova/frontends/origin/src/main.tsx (router)
   - projects/game-nova/frontends/origin/src/api/ — паттерн
     API-клиента (см. ч.1 коммит 716b5a518b: friends.ts, chat.ts,
     messages.ts, settings.ts).
   - projects/game-nova/frontends/origin/src/features/main/MainScreen.tsx
     — для проверки виджетов в S-001 (для подтверждения дубликата
     с S-046).
   - projects/game-nova/frontends/origin/.env.example или vite.config.ts
     — есть ли VITE_PORTAL_BASE_URL.
   - projects/game-nova/frontends/origin/src/features/exchange/
     CreateLotPage.tsx (если он там в Spring 3 — пример Idempotency-
     Key UUID generation; pattern для Officer/Profession activate).

ЧТО НУЖНО СДЕЛАТЬ:

### Архитектурно

1. **Routes** в роутере origin-фронта (расширить main.tsx или
   router.tsx):
   - /officer → OfficerScreen
   - /profession → ProfessionScreen
   - /user-agreement → UserAgreementScreen
   - /changelog → ChangelogScreen
   - /support → SupportScreen
   - /tools/tech-calc → AdvTechCalculatorScreen
   - /widgets → redirect на / (S-046 закрыт через S-001)

2. **API-модули** в src/api/:
   - officer.ts — list/activate. Endpoints уже в openapi.yaml.
   - profession.ts — list/me/choose. Endpoints в openapi.yaml.
   - support.ts — POST на portal-backend (cross-service).
   - changelog.ts — bundled markdown loader (если (а)) или
     fetch на endpoint (если backend есть).
   - tech-calc — pure-функции, без API.
   - user-agreement.ts — cross-link конструктор URL portal'а.

3. **OpenAPI** — НЕ расширять (уже сделано в ч.1). Если найдёшь
   что какой-то endpoint реально отсутствует — это маленькое
   дополнение через Edit, не блокирующее.

4. **i18n (R12)** — grep configs/i18n/{ru,en}.yml на
   officer|profession|user-agreement|changelog|support|techcalc|
   widget. Цель ≥95% переиспользования.

5. **Pixel-perfect** — HTML+CSS-классы зеркалят legacy *.tpl.

6. **Тесты** — vitest + RTL, 1-2 теста на каждый экран. Spring 4
   ч.2 = 6 экранов = 8-12 тестов минимум.

7. **Idempotency-Key (R9)** на mutation'ах:
   - officer.activate (POST).
   - profession.choose (POST).
   - support.submit (POST) — если portal-backend требует
     (см. план 56).

### Экран за экраном

#### S-040 Officer (OfficerScreen)

- TanStack Query → GET /api/officers (список доступных).
- Каждый officer-card: имя, описание, эффект, цена в credit
  (= оксариты по ADR-0009), длительность, активный/неактивный.
- Кнопка «Нанять» → POST /api/officers/{key}/activate с
  Idempotency-Key.
- Возможные ошибки: 402 insufficient credit, 409 already active,
  503 billing unavailable.
- Активный officer: счётчик «осталось X дней».

#### S-041 Profession (ProfessionScreen)

- TanStack Query → GET /api/professions (список) +
  GET /api/professions/me (текущий выбор).
- Карточки профессий (фермер / пират / шахтёр / etc.) с
  описанием эффектов.
- Текущая профессия — подсветка + disabled-кнопка «Уже выбрано».
- Кнопка «Выбрать» → POST /api/professions с Idempotency-Key.
- Cooldown: если backend сообщает, что выбор закрыт N дней
  (через 409 или GET /me с полем `change_available_at`), показать
  «Доступно через X дней».

#### S-043 UserAgreement (UserAgreementScreen)

- Cross-link на portal: `${VITE_PORTAL_BASE_URL}/user-agreement`.
- Реализация: либо `<a href={...} target="_blank">` (открыть в
  новой вкладке) либо `<iframe src={...}>` (inline).
- Я бы выбрал target="_blank" для простоты и консистентности с
  internet-conventions.
- Если VITE_PORTAL_BASE_URL не задана — fallback на относительный
  /user-agreement (404 в dev — норма, в prod portal обслуживает
  тот же домен).

#### S-044 Changelog (ChangelogScreen)

- Markdown bundled в frontend: `src/features/changelog/CHANGELOG.md`.
- Источник: либо скопировать из portal-frontend если он там есть,
  либо вручную нарастить из git-истории основных коммитов.
- Renderer: `react-markdown` если уже подключен, либо простой
  custom-loader на `?raw` Vite-импорт + `<pre>`.
- Если в Vite пока нет MD-loader — добавь
  (это часть скоупа этого экрана).

#### S-045 Support (SupportScreen)

- Form: subject (text), category (select: game-bug/payment/
  account/other), body (textarea), email (если у юзера не
  залогинен — иначе берём из profile).
- POST на `${VITE_PORTAL_BASE_URL}/api/reports` с body
  `{subject, category, body, source: 'origin', meta: {...}}`.
- Точный shape body — посмотри в nova-фронте ReportButton.tsx
  или portal-backend handler.
- На успех — Toast + redirect на / или /msg/inbox (где придёт
  ответ от support).

#### S-046 Widgets — НЕ реализуем, пишем simplifications

См. блок выше «ВАЖНО ПРО S-046 WIDGETS».

#### S-047 AdvTechCalculator (AdvTechCalculatorScreen)

- Pure client-side утилита.
- Form:
  - select Research (из catalog Spring 3 — endpoint
    GET /api/research/tree уже есть).
  - input current_level (целое 0..max).
  - input target_level (целое 1..max, > current).
- Output:
  - таблица: уровень → cost (metal/silicon/hydrogen) → time.
  - суммарная стоимость до target_level.
  - влияние RoboticFactory на time (если есть параметр).
- Опционально: comparison table нескольких research одновременно.
- Backend не нужен — все формулы в catalog Spring 3 + frontend
  catalog state из TanStack Query кеша.

### Финализация Ф.5

- Шапка плана 72: после ч.2 → Ф.5 ✅ полностью (Spring 4 закрыт).
  Ф.6 (Spring 5), Ф.7 (i18n рус), Ф.8 (TipTap), Ф.9 (финал) —
  следующие сессии.
- Запись итерации в docs/project-creation.txt
  («72 Ф.5 ч.2 — Spring 4 завершён»).
- В коммите указать i18n переиспользовано/новых.

══════════════════════════════════════════════════════════════════
ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА R0-R15 — см. блок rules-r0-r15.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/rules-r0-r15.md

Особо важно:
- R0: backend nova не меняем (только origin-frontend + опционально
  мелкие openapi-фиксы).
- R5: pixel-perfect для origin-фронта.
- R9: Idempotency-Key на activate/choose/submit.
- R12: i18n grep сначала, цель 95%.
- R15: S-046 Widgets-skip — это TRADE-OFF (✅ дубликат S-001 Main),
  записать в simplifications.md, не пропуск.
══════════════════════════════════════════════════════════════════

══════════════════════════════════════════════════════════════════
GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ — см. блок git-isolation.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/git-isolation.md

Свои пути для CC_AGENT_PATHS:
- projects/game-nova/frontends/origin/

  (вся папка целиком — твой territory)

- projects/game-nova/api/openapi.yaml (только если нужны мелкие
  доправки — основное сделано в ч.1)
- docs/plans/72-remaster-origin-frontend-pixel-perfect.md
- docs/active-sessions.md
- docs/project-creation.txt
- docs/simplifications.md (P72.S4.WIDGETS, P72.S4.CHANGELOG,
  и любые новые)

ВНИМАНИЕ: НЕ трогай:
- projects/game-nova/frontends/nova/ — там работает nova-фронт.
- projects/game-nova/backend/ — backend закрыт.
- projects/game-legacy-php/ — только читаешь как источник истины.
══════════════════════════════════════════════════════════════════

КОММИТЫ:

Один коммит:

feat(origin/frontend): Ф.5 Spring 4 ч.2 — premium + static +
utilities + финализация (план 72)

Trailer: Generated-with: Claude Code
ВСЕГДА с двойным тире:
git commit -m "..." -- $CC_AGENT_PATHS

ЧЕГО НЕ ДЕЛАТЬ:

- НЕ менять nova-фронт.
- НЕ менять backend handlers.
- НЕ дублировать openapi-расширение из ч.1 (уже сделано).
- НЕ создавать /api/reports в game-nova-backend (он в portal,
  cross-service).
- НЕ реализовывать S-046 Widgets как отдельный экран — записать
  в simplifications.md.
- НЕ закрывать весь план 72 — Ф.5 это Spring 4. После ч.2 →
  Ф.5 ✅, остаются Ф.6-Ф.9.
- НЕ забывай Idempotency-Key.
- НЕ забывай про -- в git commit.

УСПЕШНЫЙ ИСХОД:

- 6 экранов работают (router → каждый рендерится; S-046 redirect
  на /).
- typecheck + build + tests зелёные.
- Все экраны pixel-perfect клоны legacy.
- i18n: 95%+ переиспользование.
- 2-3 trade-off в simplifications.md (P72.S4.WIDGETS, опционально
  CHANGELOG, USER_AGREEMENT cross-link).
- Шапка плана 72 Ф.5 ✅ полностью.
- Запись в docs/project-creation.txt.
- Удалена строка из docs/active-sessions.md.

Стартуй.
```
