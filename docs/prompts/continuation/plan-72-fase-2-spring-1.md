# Промпт: выполнить план 72 Ф.2 — Spring 1 (главные игровые экраны origin-фронта)

**Дата создания**: 2026-04-28
**План**: [docs/plans/72-remaster-origin-frontend-pixel-perfect.md](../../plans/72-remaster-origin-frontend-pixel-perfect.md)
**Зависимости**: ✅ Ф.1 Bootstrap (коммит 54fabbdf46), ✅ план 64
(override-схема), ✅ план 65 (event-loop), ✅ план 66 (AlienAI),
✅ план 67 (alliance), ✅ план 78 (раскладка), ✅ ADR-0011 (display
name «Oxsar Classic»).
**Объём**: ~1500-2500 строк TS + CSS + i18n, 7 экранов в одном
коммите по Spring 1.

---

```
Задача: реализовать Spring 1 плана 72 — pixel-perfect клон 7 главных
игровых экранов origin-фронта (Main, Constructions, Research,
Shipyard, Galaxy, Mission, Empire) на каркасе Ф.1 Bootstrap.

КОНТЕКСТ:

Ф.1 закрыта коммитом 54fabbdf46. Каркас в
projects/game-nova/frontends/origin/ работает (Vite+TS+TanStack+Zustand,
3-frame layout, theme.css из legacy 1:1, i18n через nova-bundle,
auth-store с namespace 'oxsar-origin-auth'). Сейчас в App.tsx —
Bootstrap-заглушка, которую нужно заменить router'ом и 7 экранами.

Spring 1 экраны (из docs/research/origin-vs-nova/origin-ui-replication.md):
- S-001 Main (главная страница после логина — обзор империи)
- S-002 Constructions (строительство зданий на планете)
- S-003 Research (исследовательская лаборатория)
- S-004 Shipyard (верфь — производство кораблей и обороны)
- S-005 Galaxy (карта галактики — навигация по системам)
- S-006 Mission (отправка миссий: атака, шпионаж, экспедиция)
- S-007 Empire (обзор всех планет империи)

Это **pixel-perfect клон визуала** legacy oxsar2. Каждый экран
зеркалит расположение блоков, цвета, шрифты, кнопки legacy. Источник
истины:
- projects/game-legacy-php/templates/<screen>.tpl — HTML-структура
- projects/game-legacy-php/public/style.css — CSS (уже частично
  перенесён в Ф.1 в theme.css)
- скриншоты legacy — снимаются с running legacy-стека
  (см. docs/legacy/game-legacy-access.md, либо взаимодействие
  с pixel-перфект через ручную сверку)

ВАЖНО ПРО R5 (pixel-perfect ТОЛЬКО в плане 72):
- В origin-фронте — pixel-perfect клон legacy.
- В nova-фронте — современный nova-стиль (туда не лезем).
- В этой сессии работаем ИСКЛЮЧИТЕЛЬНО в
  projects/game-nova/frontends/origin/.

ВАЖНО ПРО ADR-0011 (display name):
- Технический slug `origin` — в URL, импортах, configs (не меняем).
- Display name в UI = «Oxsar Classic» (en) / «Оксар Классик» (ru).
- Header origin-фронта показывает «Oxsar Classic», не «origin».
- В footer и заголовке окна — «Oxsar Classic».

ПЕРЕД НАЧАЛОМ:

1) git status --short. cat docs/active-sessions.md.
   Параллелится с планом 68 (биржа артефактов backend) — разные
   папки, конфликта нет.

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/72-remaster-origin-frontend-pixel-perfect.md
   - docs/research/origin-vs-nova/origin-ui-replication.md секции
     S-001..S-007 (детальное описание каждого экрана)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5» R0-R15
   - docs/adr/0011-universe-display-naming.md
   - projects/game-nova/frontends/origin/src/main.tsx
     (router-точка из Ф.1)
   - projects/game-nova/frontends/origin/src/layout/AppShell.tsx
     (3-frame caркас)
   - projects/game-nova/frontends/origin/src/styles/theme.css
     (CSS-переменные из legacy)
   - projects/game-nova/frontends/origin/src/styles/layout.css
     (3-frame layout)

3) Прочитай выборочно:
   - projects/game-legacy-php/templates/main.tpl (S-001)
   - projects/game-legacy-php/templates/constructions.tpl (S-002)
   - projects/game-legacy-php/templates/research.tpl (S-003)
   - projects/game-legacy-php/templates/shipyard.tpl (S-004)
   - projects/game-legacy-php/templates/galaxy.tpl (S-005)
   - projects/game-legacy-php/templates/mission.tpl (S-006)
   - projects/game-legacy-php/templates/empire.tpl (S-007)
   - projects/game-nova/api/openapi.yaml — все используемые endpoint'ы
     (GET /api/planet, GET /api/buildings, GET /api/research,
     GET /api/shipyard, GET /api/galaxy/{g}/{s}, POST /api/missions/...,
     GET /api/empire/planets и др.)
   - projects/game-nova/frontends/nova/src/features/<domain>/ —
     ТОЛЬКО как референс на TanStack Query patterns
     (query-keys, mutation, invalidation). НЕ копировать визуально —
     nova-стиль ≠ origin-стиль.

4) Добавь свою строку в docs/active-sessions.md:
   | <N> | План 72 Ф.2 Spring 1 (7 экранов origin-фронта) | projects/game-nova/frontends/origin/ | <дата-время> | feat(origin/frontend): Ф.2 Spring 1 — 7 главных экранов |

ЧТО НУЖНО СДЕЛАТЬ:

### Архитектурно (общее для всех 7 экранов)

1. **Router** в `src/main.tsx`: react-router v6 с маршрутами:
   - `/` → Main (S-001)
   - `/constructions/:planetId?` → Constructions (S-002)
   - `/research/:planetId?` → Research (S-003)
   - `/shipyard/:planetId?` → Shipyard (S-004)
   - `/galaxy/:galaxy?/:system?` → Galaxy (S-005)
   - `/mission/:planetId?` → Mission (S-006)
   - `/empire` → Empire (S-007)
   - `/login` (placeholder, реальный экран — Ф.3)
   - 404 → Main с redirect.

2. **API-клиент** в `src/api/`:
   - `client.ts` — fetch wrapper с Bearer + Idempotency-Key
     (генерация UUID на каждую mutation), 401 → logout
     (уже было в Ф.1 — расширь под endpoint'ы Spring 1).
   - `planet.ts`, `buildings.ts`, `research.ts`, `shipyard.ts`,
     `galaxy.ts`, `mission.ts`, `empire.ts` — каждый экспортирует
     query-функции и mutation-функции.
   - Типы из openapi.yaml — используй `openapi-typescript` если он
     уже подключен (см. Ф.1), иначе **руками** (МИНИМАЛЬНО, R1
     snake_case полей в API-моделях).

3. **TanStack Query** keys в `src/api/query-keys.ts`:
   - `['planet', planetId]`, `['buildings', planetId]`,
     `['research', planetId]`, `['shipyard', planetId]`,
     `['galaxy', galaxy, system]`, `['empire']`.
   - Invalidation: при mutation (start build, start research) —
     invalidate соответствующий ключ + `planet` (ресурсы изменятся).

4. **i18n (R12, КРИТИЧНО)**:
   - Перед созданием нового ключа — grep по
     `projects/game-nova/configs/i18n/{ru,en}.yml`. Цель ≥95%
     переиспользование (как в плане 71).
   - origin-фронт читает тот же `/api/i18n/{lang}` что и nova
     (см. Ф.1 I18nProvider).
   - Новые ключи только для legacy-специфичных текстов которых нет
     в nova: например legacy-названия зданий вроде «mtg-сооружение»,
     «rocket_station».
   - Идентификаторы legacy-PHP (na_phrases типа `MSG_INV_PASS`) НЕ
     переносим — переименовываем в nova-конвенцию.
   - В коммите указать соотношение **переиспользовано/новых**.

5. **Pixel-perfect соответствие**:
   - Каждый экран — отдельная папка `src/features/<screen>/`.
   - HTML-структура зеркалит legacy *.tpl.
   - CSS-классы с теми же именами что в legacy (`.headLine`,
     `.contentBlock`, `.alphaTable` и пр.) — для лёгкости сверки.
   - Цвета/шрифты/отступы — через CSS-переменные из theme.css.
   - **Нет** новых дизайн-элементов «как было бы лучше». Это клон.

6. **Доступность (R12 a11y)**:
   - jsx-a11y eslint строгий — соблюдай.
   - aria-label на кнопках без текста, alt на img.

7. **Что НЕ делаем в Spring 1**:
   - Achievements, Tutorial — исключены из первой итерации (см. план 72).
   - Mobile-адаптив — после старта.
   - Тёмная тема — после старта.
   - Реклама/баннеры из legacy — НЕ переносить.
   - Логин/Register экраны — это Ф.3 (Spring 2).

### Экран за экраном

#### S-001 Main (Главная)

`src/features/main/MainScreen.tsx`:
- Обзор империи: текущая планета, ресурсы, очередь строительства,
  последние события (миссии в прогрессе).
- Layout legacy: блоки «Производство», «Очередь стройки», «Текущие
  миссии», «Сообщения непрочитанные».
- Endpoints: GET /api/planet, GET /api/empire/overview (если есть),
  GET /api/missions?status=active, GET /api/messages/unread-count.
- TanStack Query с auto-refresh на ресурсы (interval=10s).

#### S-002 Constructions (Строительство)

`src/features/constructions/ConstructionsScreen.tsx`:
- Список доступных зданий с уровнем, ценой, временем стройки.
- Кнопка «Построить» → POST /api/buildings/{type}/start
  (Idempotency-Key обязателен).
- Очередь стройки с прогрессом + кнопка «Отмена» → DELETE
  соответствующего event.
- Endpoints: GET /api/buildings/{planetId}, POST .../start,
  DELETE /api/events/{id}.

#### S-003 Research (Исследования)

`src/features/research/ResearchScreen.tsx`:
- Tech-tree с уровнями, ценой, временем.
- Зависимости (требуемые здания/исследования) — disabled state с
  hint.
- Кнопка «Исследовать» → POST /api/research/{type}/start.
- Endpoints: GET /api/research/{planetId}, POST .../start.

#### S-004 Shipyard (Верфь)

`src/features/shipyard/ShipyardScreen.tsx`:
- Список юнитов (корабли + оборона) с ценой, временем.
- Поле «Количество» (max = по ресурсам) + кнопка «Построить» →
  POST /api/shipyard/build.
- Очередь производства с прогрессом.
- Endpoints: GET /api/shipyard/{planetId}, POST /api/shipyard/build.

#### S-005 Galaxy (Карта галактики)

`src/features/galaxy/GalaxyScreen.tsx`:
- Таблица 15 позиций (или сколько в legacy) на текущей системе.
- Навигация: input galaxy/system + кнопки «вперёд/назад/перейти».
- Каждая занятая позиция — username, alliance-tag, статус
  (активен/inactive/banned).
- Действия: «Шпионаж», «Атака», «Передача», «Сообщение» —
  navigate к Mission или Message-форме.
- Endpoints: GET /api/galaxy/{galaxy}/{system}.

#### S-006 Mission (Миссия)

`src/features/mission/MissionScreen.tsx`:
- Форма отправки миссии: цель (galaxy:system:position), флот, тип
  миссии (атака/шпионаж/экспедиция/транспорт/...).
- Селектор кораблей по типам (доступные / отправляемые).
- Расчёт времени, стоимости топлива.
- Кнопка «Отправить» → POST /api/missions (Idempotency-Key).
- Endpoints: GET /api/planet, GET /api/galaxy/.../{position},
  POST /api/missions.

#### S-007 Empire (Империя)

`src/features/empire/EmpireScreen.tsx`:
- Таблица всех планет игрока: имя, координаты, ресурсы, поля,
  очередь стройки.
- Сортировка по столбцам.
- Клик по планете → navigate(`/`+planet).
- Endpoints: GET /api/empire/planets.

### Тесты

- Vitest + React Testing Library — минимум по 1-2 теста на экран:
  - Main: рендер блоков, корректное отображение ресурсов.
  - Constructions: клик «Построить» вызывает mutation с
    Idempotency-Key.
  - Research: tech-tree disabled состояние.
  - Shipyard: лимит quantity по ресурсам.
  - Galaxy: навигация galaxy/system.
  - Mission: форма с валидацией.
  - Empire: сортировка таблицы.

### Финализация Ф.2

- Шапка плана 72: Ф.2 ✅, Spring 1 закрыт.
- НЕ закрываешь весь план 72 — впереди Ф.3-Ф.9.
- Запись итерации в docs/project-creation.txt («72 Ф.2 — Spring 1»).
- В коммите указать соотношение i18n **переиспользовано/новых**.

══════════════════════════════════════════════════════════════════
ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА R0-R15 — см. блок rules-r0-r15.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/rules-r0-r15.md

Особо важно:
- R0: модерн-формулы NOVA не трогаем. Origin-фронт читает игровое
  состояние ИЗ nova-API, никаких backend-изменений в этой сессии.
- R5: pixel-perfect — это план 72. Соответствие визуалу legacy
  буквальное.
- R12: i18n grep сначала, цель 95% переиспользования.
- R7: API-схемы можно расширять (если нужны новые endpoint'ы для
  origin), но в первую очередь используй существующие. Если нет
  endpoint'а — отметь в simplifications.md и используй mock с TODO.
══════════════════════════════════════════════════════════════════

══════════════════════════════════════════════════════════════════
GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ — см. блок git-isolation.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/git-isolation.md

Свои пути для CC_AGENT_PATHS:
- projects/game-nova/frontends/origin/

  (вся папка целиком — это твой territory; план 68 параллельно
   работает в backend, не пересечётся; план 67 nova-фронт уже
   закрыт)

- docs/plans/72-remaster-origin-frontend-pixel-perfect.md
- docs/active-sessions.md
- docs/project-creation.txt (запись итерации)

ВНИМАНИЕ: НЕ трогай:
- projects/game-nova/frontends/nova/ — это nova-фронт, R5 запрещает
  pixel-perfect там.
- projects/game-nova/backend/ — backend этой сессии не задеваем.
- projects/game-nova/api/openapi.yaml — если нужно расширить, отметь
  в simplifications.md и пометь как ожидающее расширения backend.
- projects/game-legacy-php/ — только читаем как источник истины.
══════════════════════════════════════════════════════════════════

КОММИТЫ:

Один большой коммит на весь Spring 1 (7 экранов органически
связаны через router и общую инфру):

feat(origin/frontend): Ф.2 Spring 1 — 7 главных экранов (план 72)

Trailer: Generated-with: Claude Code
ВСЕГДА с двойным тире:
git commit -m "..." -- $CC_AGENT_PATHS

Если объём окажется > 3000 строк — раздели на 2 коммита:
1) feat(origin/frontend): Ф.2 Spring 1 ч.1 — Main+Construction+
   Research+Shipyard
2) feat(origin/frontend): Ф.2 Spring 1 ч.2 — Galaxy+Mission+Empire

ЧЕГО НЕ ДЕЛАТЬ:

- НЕ менять nova-фронт.
- НЕ менять backend (если не хватает endpoint'а — simplifications.md).
- НЕ переносить рекламу/баннеры из legacy.
- НЕ делать Achievements / Tutorial экраны.
- НЕ хардкодить тексты — только через i18n (R12).
- НЕ использовать `any` в TS (R1 + tsconfig strict).
- НЕ забывать про Idempotency-Key на mutation'ах (R9).
- НЕ забывать про -- в git commit.

УСПЕШНЫЙ ИСХОД:

- 7 экранов работают (npm run dev → переходишь по router'у,
  каждый рендерится).
- typecheck зелёный (npm run typecheck).
- build зелёный (npm run build).
- Vitest зелёный (npm run test).
- Все экраны pixel-perfect клоны legacy (визуальная сверка с
  скриншотами / running legacy-стеком).
- i18n: 95%+ переиспользование, в коммите указано.
- Шапка плана 72: Ф.2 ✅ (Spring 1).
- Запись в docs/project-creation.txt.
- Удалена строка из docs/active-sessions.md.

Стартуй.
```
