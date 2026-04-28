# План 72: Origin-фронт — pixel-perfect клон (главный)

**Дата**: 2026-04-28
**Статус**: Скелет (детали допишет агент-реализатор при старте)
**Зависимости**: блокируется планами **64, 65, 66, 67, 68, 69, 70**
(вся backend-готовность); **57** (mail/TipTap для чата + почты);
**75** (путь освобождён). Может быть **частично** запущен раньше —
экраны, backend которых уже готов.
**Связанные документы**:
- [62-origin-on-nova-feasibility.md](62-origin-on-nova-feasibility.md)
- [docs/research/origin-vs-nova/origin-ui-replication.md](../research/origin-vs-nova/origin-ui-replication.md) —
  **главный референс**: S-001..S-055 экраны + layout + ассеты + типовые таблицы
- [docs/research/origin-vs-nova/roadmap-report.md](../research/origin-vs-nova/roadmap-report.md) —
  R1-R5 + раздел плана 72 (включая разбиение на 5 спринтов)

---

## Цель

Построить **новый фронт origin** — отдельный Vite-bundle в
`projects/game-origin/frontend/` (путь освобождён планом 75), —
который **визуально pixel-perfect** воспроизводит legacy
game-origin-php (тема standard) на современном стеке React/TS, но
**функционально работает на nova-API** (без backend-адаптеров).

Это самый большой план серии (~3-4 месяца).

---

## Что делаем

### Bootstrap

- `projects/game-origin/frontend/` — новый Vite-проект.
- Стек по nova-конвенциям: TypeScript strict, TanStack Query,
  Zustand, OpenAPI-сгенерированный клиент к nova-API, TipTap
  для чата + почты.
- Тема — переписать `public/css/` legacy в современный CSS
  (CSS-переменные, без legacy-хаков но с тем же визуалом).
- Перенос ассетов из `projects/game-origin-php/public/`:
  - `images/` — SVG/PNG спрайты (с проверкой лицензий).
  - `fonts/` — шрифты (с проверкой лицензий — это критично, см.
    Risk register).

### Layout

3-frame воспроизведение (как в origin):
- Top header (баланс, юзернейм, language, logout)
- Left menu (collapsible)
- Main content area
- Footer (юр-ссылки + 12+ маркировка плана 50)

### Реализация всех 50 prod-экранов

S-NNN записи из `origin-ui-replication.md`. Разбиение на 5 спринтов
(см. roadmap-report.md план 72):

- **Spring 1** (~7 экранов): Main, Constructions, Research,
  Shipyard, Galaxy, Mission, Empire.
- **Spring 2** (~10 экранов): Alliance (12 шаблонов), Resource,
  Market, Repair, Battlestats, Fleet operations.
- **Spring 3** (~10 экранов): Artefacts, ArtefactMarket,
  ArtefactInfo, BuildingInfo, UnitInfo, Techtree, Records,
  Statistics, Achievements, Daily quests.
- **Spring 4** (~13 экранов): Friends, MSG, Chat, ChatAlly,
  Notepad, Search, Officer, Profession, Settings, Tutorial,
  UserAgreement, Changelog, Support, Widgets, AdvTechCalculator.
- **Spring 5** (~5 экранов): Simulator, RocketAttack, MonitorPlanet,
  ResTransferStats, Stock/Exchange (зависит от плана 68).

### i18n

Только **русский** в первой итерации. Английский / другие — после
старта.

### Чат

BBCode origin **выкидывается** — заменяется TipTap (план 57). Это
единственное **намеренное визуальное расхождение** с legacy
(BBCode-теги в сообщениях больше не отображаются как `[b]...[/b]`,
а как rich-text TipTap; но контент сохраняется).

### Что НЕ делаем в первой итерации

- Адаптив (mobile / tablet) — после старта.
- Тёмная тема — после старта (legacy тема — единственная).
- Перепроектирование экранов / новые UX-фичи — это новшества,
  делаются после стабилизации pixel-perfect клона.

## Этапы (детали — при старте)

- Ф.1. Bootstrap проекта + ассеты + layout + тема.
- Ф.2-Ф.6. Spring 1-5 (по экранам).
- Ф.7. i18n русский.
- Ф.8. TipTap-чат (зависит от плана 57).
- Ф.9. Финализация.

## Конвенции (R1-R5)

- Стек React/TS strict + TanStack Query + Zustand + TipTap.
- OpenAPI-клиент к nova-API (R2).
- React-компоненты — функциональные, hook'и.
- CSS — переменные темы, не хардкод цветов (несмотря на pixel-perfect,
  переменные позволяют менять в будущем).
- Никаких backend-адаптеров — origin-фронт сразу на nova-имена API.
- Pixel-perfect (R5) — визуальный паритет проверяется планом 73
  (screenshot-diff CI).

## Объём

12-16 недель (3-4 месяца). Самый большой план серии.

## References

- origin-ui-replication.md — все S-NNN с file:line, layout,
  типовые таблицы.
- nova-ui-backlog.md U-NNN + X-NNN — фичи и UX, которые могут
  понадобиться.
- План 57 — TipTap для чата.
- План 75 — путь `projects/game-origin/frontend/` освобождён.
- `projects/game-origin-php/public/` + `src/templates/standard/` —
  визуальный референс.
