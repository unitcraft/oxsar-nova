# План 72 (ремастер): Origin-фронт — pixel-perfect клон (главный план серии)

**Дата**: 2026-04-28
**Статус**: Ф.1 ✅ (Bootstrap, коммит 54fabbdf46, 2026-04-28),
Ф.2 ✅ Spring 1 (7 главных экранов: Main, Constructions, Research,
Shipyard, Galaxy, Mission, Empire — каркас pixel-perfect HTML+CSS,
2026-04-28), Ф.3 ✅ Spring 2 (12 alliance-экранов S-008..S-019 +
5 экранов S-020..S-024: resource-market/market/repair/battlestats/
fleet-operations — коммиты 48ef07cf19 + следующий, 2026-04-28),
Ф.4 ✅ Spring 3 (7 экранов S-013/S-014/S-018/S-019/S-021/S-024/
S-023: artefacts + 3 info-страницы + techtree + records + ranking;
backend пакет `internal/catalog/` с 3 catalog endpoints, real
`/api/techtree` и `/api/records` подключены — реальные данные, не
моки; catalog отдаёт current-universe-only данные modern-вселенной,
см. simplifications P72.S3.A, 2026-04-28),
Ф.5 🟡 Spring 4 ч.1 ✅ (7 экранов S-034 Friends / S-035 MSG /
S-036 Chat / S-037 ChatAlly / S-038 Notepad / S-039 Search /
S-042 Settings — communication+notes+search+settings; полное
openapi-расширение для всех 14 экранов Spring 4 включая ч.2 —
закрытие R2 тех-долга; BBCode → plain text trade-off P72.S4.BBCODE,
TipTap отложен на Ф.8; legacy-only Settings поля P72.S4.SETTINGS;
2026-04-28); Spring 4 ч.2 в очереди (6 экранов: Officer / Profession /
UserAgreement / Changelog / Support / AdvTechCalc + Widgets-skip),
Ф.6-Ф.9 — отдельными сессиями.
**Зависимости** (актуализировано 2026-04-28):
- ✅ План 64 — закрыт (override-схема балансов).
- ✅ План 65 — закрыт (Ф.1-Ф.6, включая teleport).
- ✅ План 66 — закрыт (Ф.1-Ф.7, выкуп удержания + golden).
- 🟡 План 67 — backend закрыт (Ф.1-Ф.4), frontend Ф.5 ч.1 закрыт
  (коммит 669af55dae); Ф.5 ч.2 + Ф.6 идут параллельно в nova-фронте,
  origin-фронт это не блокирует (отдельная папка).
- ✅ План 68 — закрыт.
- ✅ План 69 — закрыт.
- ✅ План 71 — закрыт (UX-микрологика nova-фронта).
- ✅ План 75 — закрыт (путь `frontends/origin/` освобождён).
- ✅ План 78 — закрыт (раскладка фронтов + переименование
  game-origin-php → game-legacy-php).
- 🚫 План 57 (mail) — справочный документ-эпик, не выполняется
  как обязательная фаза; для Bootstrap не блокер.

**Старт Ф.1 безопасен ✅.** План 70 (achievements) выведен из
обязательных — отложен до пост-запуска (см. roadmap-report
«Часть V»).
**Связанные документы**:
- [62-origin-on-nova-feasibility.md](62-origin-on-nova-feasibility.md)
- [docs/research/origin-vs-nova/origin-ui-replication.md](../research/origin-vs-nova/origin-ui-replication.md) —
  **главный референс**: S-001..S-055 экраны + layout + ассеты + типовые таблицы
- [docs/research/origin-vs-nova/roadmap-report.md](../research/origin-vs-nova/roadmap-report.md) —
  R1-R5 + раздел плана 72 (включая разбиение на 5 спринтов)

---

## Цель

Построить **новый фронт origin** — отдельный Vite-bundle в
`projects/game-nova/frontends/origin/` (путь освобождён планом 75), —
который **визуально pixel-perfect** воспроизводит legacy
game-legacy-php (тема standard) на современном стеке React/TS, но
**функционально работает на nova-API** (без backend-адаптеров).

Это самый большой план серии (~3-4 месяца).

---

## Что делаем

### Bootstrap

- `projects/game-nova/frontends/origin/` — новый Vite-проект.
- Стек по nova-конвенциям: TypeScript strict, TanStack Query,
  Zustand, OpenAPI-сгенерированный клиент к nova-API, TipTap
  для чата + почты.
- Тема — переписать `public/css/` legacy в современный CSS
  (CSS-переменные, без legacy-хаков но с тем же визуалом).
- Перенос ассетов из `projects/game-legacy-php/public/`:
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
- **Spring 3** (~8 экранов): Artefacts, ArtefactMarket,
  ArtefactInfo, BuildingInfo, UnitInfo, Techtree, Records,
  Statistics, Daily quests.
  - **Исключены из первой итерации**: Achievements (см. ниже).
- **Spring 4** (~12 экранов): Friends, MSG, Chat, ChatAlly,
  Notepad, Search, Officer, Profession, Settings,
  UserAgreement, Changelog, Support, Widgets, AdvTechCalculator.
  - **Исключён из первой итерации**: Tutorial (см. ниже).
- **Spring 5** (~5 экранов): Simulator, RocketAttack, MonitorPlanet,
  ResTransferStats, Stock/Exchange (зависит от плана 68).

### i18n

Только **русский** в первой итерации. Английский / другие — после
старта.

**Переиспользование nova-bundle (R12):**

В `projects/game-nova/configs/i18n/` уже есть значительный
i18n-набор для существующих фич (бой, экономика, дипломатия, чат,
уведомления). Перед созданием **любой** новой строки в origin-фронте:

1. **Grep по nova-bundle** на похожий ключ или текст.
2. **Совпало логически + группа подходит** → переиспользовать.
3. **Совпало логически, разные группы** → оценить, стоит ли
   обобщить (например, `error.common.not_enough_resources`).
4. **Логически совпадает, но смысл тоньше** → завести отдельный
   ключ.
5. **Аналога нет** → завести новый по nova-конвенции
   (snake_case, namespace).

**Идентификаторы legacy-PHP `na_phrases` НЕ переносить** — там
PHP-ключи вроде `MSG_INV_PASS`. Ищем существующий nova-ключ или
заводим новый по nova-конвенции. Тексты на русском
переиспользуем как значения, ключи — всегда nova-стиль.

В коммитах плана 72 указывать: **сколько nova-ключей переиспользовано
vs новых** (метрика качества). Цель — максимизировать переиспользование.

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
- **Achievements (S-Achievements)** — экран ачивок не реализуется
  в origin-фронте на старте. В nova ачивки уже есть (план 17,
  goal engine), для origin реализуем позже отдельным планом
  (см. план 70 — отложен). Пункт меню «Достижения» в шапке/левом
  меню в первой итерации скрыт ИЛИ ведёт на заглушку «Скоро».
  Эталонный screenshot этого экрана для CI (план 73) не снимается.
- **Tutorial (S-Tutorial)** — экран обучения не реализуется на
  старте. nova имеет свой onboarding-flow, для origin вернёмся
  отдельным планом после старта. Игроки приходят в origin через
  portal с уже пройденным onboarding identity-сервиса (планы 36, 51).
- **Баннеры и рекламные тексты** из legacy-PHP **не переносятся**
  в новый фронт. В `projects/game-legacy-php/` есть рекламные блоки
  (промо-тексты, баннеры в шапке/футере/между секциями) — они
  **не копируются** ни визуально, ни функционально. На старте
  origin-фронт чист от рекламы. Решение по монетизации origin
  принимается отдельно после запуска (если будет). Это **намеренное
  визуальное расхождение** с legacy-PHP — отдельные «дыры»
  на этих местах допустимы (или пустые контейнеры заполняются
  нейтральным контентом — игровой статистикой, новостями portal,
  etc.).

## Этапы (детали — при старте)

- Ф.1. Bootstrap проекта + ассеты + layout + тема.
- Ф.2-Ф.6. Spring 1-5 (по экранам).
- Ф.7. i18n русский.
- Ф.8. TipTap-чат (зависит от плана 57).
- Ф.9. Финализация.

### Backend-расширения по требованию (R15-обязательные)

origin-фронт открывает legacy-механики которых nova-фронт не
использовал, и для некоторых экранов в `nova-backend` нет нужных
endpoint'ов. По R15 (без MVP-сокращений / mock'ов «как для прода»)
**отсутствующий endpoint = расширение backend в той же сессии**,
не TODO-заглушка.

При обнаружении отсутствующего endpoint'а в processe Spring N агент:
1. Проверяет что endpoint реально нужен (не дублирует существующий
   с другим именем).
2. Если нужен — расширяет backend (обычно 50-200 строк Go: handler +
   запись в openapi.yaml + тесты).
3. R8 (Prometheus), R12 (i18n), R10 (universe-aware если применимо),
   R9 (Idempotency-Key для мутаций — для GET-каталогов не нужен).
4. **Не записывает** это как trade-off в simplifications.md —
   это полная реализация, не упрощение.

Зафиксированные расширения backend в фазах:
- **Ф.4 (Spring 3)** — добавляются catalog-endpoints (отсутствовали
  в openapi.yaml на момент Spring 3):
  - `GET /api/artefacts/catalog/{type}` — описание артефакта.
  - `GET /api/buildings/catalog/{type}` — параметры здания + pre-
    computed таблица.
  - `GET /api/units/catalog/{type}` — параметры юнита + pre-
    computed статы для нескольких уровней техники.
  - `GET /api/research/tree` — граф технологий с зависимостями.
  - Альтернативно: один общий `GET /api/catalog` для (1)+(2)+(3)
    (агент решает по простоте; tree и records — отдельные).

  **Важно про источник данных и форму ответа catalog-endpoints:**

  Параметры (`cost_base`, `cost_factor`, `base_rate_per_hour`,
  `max_level`, ...) живут в YAML `projects/game-nova/configs/
  {buildings,units,research}.yml` (читаются через
  `internal/balance/loader.go`, план 64). **Формулы** живут в **Go-
  коде** — для modern (nova) в `internal/economy/formulas.go`,
  для origin в `internal/origin/economy/*.go`. Они **не
  сериализуются** в YAML и не отдаются как «формула» через endpoint.

  Что catalog-endpoint отдаёт:
  - **params** из YAML — как есть (для отображения «base_cost: 60»,
    «factor: 1.5», «max_level: 40»).
  - **pre-computed** таблица результатов для нескольких ключевых
    уровней (например, 1, 5, 10, 20, max_level) — рассчитанных
    через Go-функции из economy/ (cost_at_level, production_at_level,
    time_at_level и пр.). Frontend рендерит таблицу «уровень →
    стоимость → производство → время → энергия».
  - **formula description** как строка для UI (например,
    «base_cost × factor^(level-1)») — справочно, не как код.

  Это правильное разделение: формула живёт в Go (как вычисляется),
  endpoint сериализует **результат** для UI.

  **Universe-context — сознательно НЕ реализуется в Ф.4:**

  Backend в плане 64 поддерживает override per-universe (modern vs
  origin может иметь разные числа). Но catalog-endpoint в Ф.4 Spring 3
  **НЕ принимает universe-параметр** — отдаёт modern (nova)
  параметры из `internal/economy/formulas.go` + `configs/{buildings,
  units,research}.yml`. Origin-вселенная ещё не запущена ни для кого
  (план 74 — публичный запуск, ещё не сделан), поэтому current-
  universe-only catalog **корректен** для текущего состояния продукта.

  Universe-aware расширение catalog-endpoint (через `?universe=origin`
  query-param или per-user JWT-routing к `internal/origin/economy/`) —
  отдельная архитектурная задача после плана 74. Создан черновик
  [план 83 «Universe-aware catalog-endpoints»](83-catalog-endpoints-universe-aware.md)
  — триггер запуска: после плана 74 либо при жалобах от origin-
  тестеров. Это **не упрощение** текущей реализации (catalog данные
  **полные**, real numbers, не mock), а сознательно отложенное
  расширение области применения.

  Это **не R15-нарушение**: catalog отдаёт реальные данные, frontend
  получает реальные числа. Universe-aware — это про **routing**,
  не про **полноту данных**.

- **Records (S-031) — отдельный план 82**, не Ф.4 Spring 3.
  Per-unit record holders (top-1 по типу здания/исследования/корабля)
  — отсутствующий backend-домен в nova (агрегации не считаются
  нигде). Это **не упрощение существующей фичи**, а отсутствующий
  домен с собственным объёмом ~400-600 строк (миграция +
  materialized view + cron + golden + handler).

  В Ф.4 Spring 3 реализуется **endpoint-skeleton** для S-031:
  - `GET /api/records?type=...` возвращает корректный DTO
    `{records: [], type, category}` с пустым массивом.
  - R8 Prometheus подключён.
  - Frontend `RecordsScreen` рендерит empty-state «Рекорды
    появятся в планируемом обновлении».
  - Когда план 82 реализован — handler начинает возвращать данные,
    frontend **не меняется**.

  Это **TRADE-OFF по R15** (✅ «Не реализованная фича из плана,
  явно отложенная до Ф.X»), а не пропуск. Записывается в
  `simplifications.md` как P72.S3.X.

- **Ф.5 (Spring 4)** — TBD при реализации.
- **Ф.6 (Spring 5)** — TBD при реализации.

Если по ходу обнаруживается что нужен значительно более сложный
endpoint (>200 строк, требует миграции, нового домена) — агент
останавливается и обсуждает с пользователем (отдельный sub-план
или вынос в новый план — как сделано с records → план 82).

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
- План 75 — путь `projects/game-nova/frontends/origin/` освобождён.
- `projects/game-legacy-php/public/` + `src/templates/standard/` —
  визуальный референс.
