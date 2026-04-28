# Промпт: выполнить план 72 (origin-фронт pixel-perfect)

**Дата создания**: 2026-04-28
**План**: [docs/plans/72-remaster-origin-frontend-pixel-perfect.md](../plans/72-remaster-origin-frontend-pixel-perfect.md)
**Зависимости**: блокируется планами 64-69 (backend) + 71 (UX) + 57
(mail/TipTap для чата) + 75 (путь освобождён — ✅).
**Объём**: 12-16 нед, 5 спринтов. Самый большой план серии.

---

```
Задача: выполнить план 72 (ремастер) — построить новый origin-фронт
в projects/game-nova/frontends/origin/ как pixel-perfect клон визуала
legacy-PHP origin (тема standard) на современном стеке React/TS,
работающий на nova-API.

ВАЖНОЕ:
- Это **главный план серии ремастера**, ~3-4 месяца работы.
- Зависит от завершения планов 64-69, 71. Можно начинать с
  Bootstrap (Ф.1) если backend ещё не полностью готов, и постепенно
  добавлять экраны по мере готовности эндпоинтов.
- legacy = только d:\Sources\oxsar2 / oxsar2-java / game-legacy-php.
  origin = projects/game-origin/. Не путать.

ПЕРЕД НАЧАЛОМ:

1) git status --short.

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/72-remaster-origin-frontend-pixel-perfect.md
   - docs/research/origin-vs-nova/origin-ui-replication.md —
     **главный референс**: 55 экранов S-NNN с file:line, layout,
     типовые таблицы.
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5»
     (R0-R15, особенно R5 pixel-perfect, R12 i18n переиспользование,
     R15 без упрощений).
   - docs/plans/57-mail-service.md (TipTap-конфигурация).

3) Выборочно:
   - projects/game-legacy-php/src/templates/standard/ — UX/layout
     референс (НЕ для копирования кода, для воспроизведения визуала).
   - projects/game-legacy-php/public/css/ — палитра, layout-системы.
   - projects/game-legacy-php/public/images/ — иконки, спрайты
     (с проверкой лицензий).
   - projects/game-nova/frontends/nova/ — стек/конвенции nova (для
     общности подходов, НЕ для переиспользования компонентов).
   - projects/game-nova/configs/i18n/ — **обязательно grep**
     перед созданием каждой новой строки (R12).

ЧТО НУЖНО СДЕЛАТЬ:

Bootstrap (Ф.1):
- Создать projects/game-nova/frontends/origin/ — отдельный Vite-проект.
- Стек: TypeScript strict, TanStack Query, Zustand, OpenAPI-клиент
  к nova-API, TipTap (MIT) для чата + почты.
- CSS-тема — переписать legacy-PHP public/css/ в современный CSS
  (CSS-переменные, без legacy-хаков, но с тем же визуалом).
- Перенос ассетов из projects/game-legacy-php/public/:
  · images/ — SVG/PNG спрайты (с проверкой лицензий, R4 от
    plan 41).
  · fonts/ — шрифты (с проверкой лицензий — критично).

Layout (Ф.1):
- 3-frame воспроизведение (как в legacy-PHP):
  · Top header (баланс, юзернейм, language, logout).
  · Left menu (collapsible).
  · Main content area.
  · Footer (юр-ссылки + 12+ маркировка плана 50).

5 спринтов экранов (Ф.2-Ф.6):
- Spring 1 (~7): Main, Constructions, Research, Shipyard, Galaxy,
  Mission, Empire.
- Spring 2 (~10): Alliance (12 шаблонов), Resource, Market, Repair,
  Battlestats, Fleet operations.
- Spring 3 (~8): Artefacts, ArtefactMarket, ArtefactInfo,
  BuildingInfo, UnitInfo, Techtree, Records, Statistics, Daily quests.
  ИСКЛЮЧЕНО: Achievements (план 70 отложен).
- Spring 4 (~12): Friends, MSG, Chat, ChatAlly, Notepad, Search,
  Officer, Profession, Settings, UserAgreement, Changelog, Support,
  Widgets, AdvTechCalculator.
  ИСКЛЮЧЕНО: Tutorial.
- Spring 5 (~5): Simulator, RocketAttack, MonitorPlanet,
  ResTransferStats, Stock/Exchange (зависит от плана 68).

i18n (Ф.7) — только русский:
- ⚠️ Перед созданием **каждой** новой строки grep по
  projects/game-nova/configs/i18n/ (R12). Цель — максимальное
  переиспользование.
- Идентификаторы legacy-PHP na_phrases НЕ переносить. Тексты —
  переиспользуем как значения, ключи — nova-конвенция (snake_case
  namespace).
- В коммитах указывать соотношение переиспользовано/новых ключей.

Чат (Ф.8):
- BBCode origin **выкидывается** — заменяется TipTap (план 57).
- Это единственное намеренное визуальное расхождение с legacy-PHP.

Финализация (Ф.9):
- Шапка плана 72 → ✅.
- docs/project-creation.txt — итерация 72.
- divergence-log.md / nova-ui-backlog.md / origin-ui-replication.md —
  пометить S-NNN/U-NNN/X-NNN как ✅.

ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА:

R0: не трогать nova-чисел/механик. Origin-фронт работает на
nova-API; если обнаруживается несовпадение — это новая фича для
backend (отдельный план, не правка nova).
R1: snake_case JSON, _at для timestamps, английский в API.
R5: pixel-perfect клон. Threshold ≤ 0.5% diff (план 73 проверит).
R6: REST API с нуля — origin-фронт сразу пишется на nova-имена,
без backend-адаптеров.
R12: i18n переиспользование — критично для плана 72.
R15 УТОЧНЕНО (см. roadmap-report.md "Часть I.5 / R15"):
🚫 Пропуск R12 i18n (хардкод строки) — обязателен grep nova-bundle
   и Tr() со старта.
🚫 Пропуск screenshot-diff (план 73) для готовых экранов — нельзя.
✅ Trade-off: разбиение на 5 спринтов (по плану) — это R15-совместимый
   способ выполнить большой объём качественно.

R15: без упрощений — все экраны полностью функциональны со старта,
не «MVP пробросом».

ИСКЛЮЧЕНИЯ И ОТКАЗЫ (см. roadmap-report «Часть V»):
- ❌ Achievements экран — отложен до плана 70 реактивации. Пункт
  меню скрыт ИЛИ ведёт на заглушку «Скоро».
- ❌ Tutorial экран — отложен. Игроки приходят через portal с
  пройденным onboarding identity.
- ❌ Баннеры/реклама — не переносятся из legacy-PHP. Чистый фронт.
- ❌ Реферальный экран — кнопка ведёт на portal в новой вкладке
  (план 59 на portal-frontend), без своего экрана в origin.

GIT-ИЗОЛЯЦИЯ:
- Свои пути: projects/game-nova/frontends/origin/ (всё),
  projects/game-nova/configs/i18n/ (если новые ключи —
  СОВМЕСТНО с nova-frontend, осторожно),
  docs/plans/72-..., docs/research/origin-vs-nova/* (S-NNN/U-NNN
  пометки).

КОММИТЫ (5 спринтов = 5+ коммитов):

1. feat(origin-frontend): bootstrap + layout + theme (план 72 Ф.1).
2-6. feat(origin-frontend): Spring N — N экранов (план 72 Ф.N).
   Каждый коммит указывает: сколько i18n-ключей nova-bundle
   переиспользовано / создано новых.
N+1. feat(origin-frontend): TipTap-чат + финализация.

ЧЕГО НЕ ДЕЛАТЬ:
- НЕ копировать .tpl-код 1:1 — это референс визуала, переписываем
  React-эквивалентами.
- НЕ делать адаптив / тёмную тему / новые UX-фичи в первой
  итерации (R5: pixel-perfect только).
- НЕ переиспользовать nova-frontend компоненты для origin —
  у origin свой набор компонентов, визуально клонированных.
- НЕ дублировать i18n-ключи — обязательный grep (R12).

УСПЕШНЫЙ ИСХОД:
- Все 50 prod-экранов реализованы (без Achievements/Tutorial).
- Pixel-perfect screenshot-diff threshold ≤ 0.5% (план 73 проверит).
- TipTap-чат вместо BBCode.
- Только русский в первой итерации.
- В коммитах указано соотношение переиспользовано/новых i18n-ключей.
- Все S-NNN/U-NNN/X-NNN, относящиеся к плану 72, помечены ✅.

Стартуй.
```
