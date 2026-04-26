---
title: 33 — Вынос строк интерфейса в i18n-модуль
date: 2026-04-26
status: planned
---

# План 33: Перенос пользовательских строк в языковой модуль

**Цель**: перенести все user-facing строки из кода (frontend и backend) в
существующие словари [configs/i18n/ru.yml](../../configs/i18n/ru.yml) /
[configs/i18n/en.yml](../../configs/i18n/en.yml). Базовый язык — русский,
английский подключим позже как ещё одну локаль (инфраструктура уже готова).

**Сейчас имеем**:
- Backend i18n-пакет
  ([backend/internal/i18n/i18n.go](../../backend/internal/i18n/i18n.go)):
  `Bundle.Tr(lang, group, key, args...)`, fallback на ru, маркер
  `[group.key]` при отсутствии. Загружается в
  [backend/cmd/server/main.go:242-254](../../backend/cmd/server/main.go#L242-L254),
  раздаётся через `GET /api/i18n` и `GET /api/i18n/{lang}`.
- Frontend i18n
  ([frontend/src/i18n/i18n.tsx](../../frontend/src/i18n/i18n.tsx)):
  `useTranslation()` с `t(group, key, ...args)` и `tf(...)` (с дефолтом),
  кеш Query + localStorage по `LOCALE_VERSION`.
- Словари — 1504 строки в ru.yml, 1486 в en.yml (en — заготовка после
  `import-phrases`, многие ключи равны ru). 23 группы (после Ф.1bis
  все будут в `lowerCamelCase`): `achievements`, `administrator`,
  `alliance`, `artefactInfo`, `assaultReport`, `autoMessages`,
  `buddylist`, `espionageReport`, `galaxy`, `main`, `message`,
  `payment`, `prefs`, `registration`, `resource`, `statistics`,
  `unitInfo`, `userAgreement`, `buildings`, `error`, `global`, `info`,
  `mission`.

**Состояние использования**:
- Frontend: `useTranslation` используется в 3 файлах из 46 содержащих
  кириллицу (App.tsx, BattleSimScreen.tsx, сам i18n.tsx). Остальное —
  хардкод. Грубая оценка: **~1300 строк-литералов** в .tsx + .ts.
- Backend: `Bundle.Tr()` не используется ни в одном прод-файле. В коде
  ~199 хардкод-строк-литералов с кириллицей: subject/body
  inbox-сообщений, тексты кнопок Confirm на legacy-страницах, ошибки в
  `httpx.WriteError`, статусы alien/expedition/spy.

**Вне зоны этого плана**:
- Wiki-страницы (`docs/wiki/*.md`) и pages таблицы — отдельный pipeline,
  трогать не будем.
- Игровые конфиги в `configs/units/*.yml`, `configs/buildings/*.yml` —
  имена/описания берутся через `unitInfo`, `buildings` группы; здесь
  нужно проверить, что описания юнитов уже идут через i18n (см. Ф.4).

---

## 1. Принципы

1. **Никакого «нового i18n»**: используем существующий `Bundle.Tr` и
   `useTranslation`. Ключи добавляем в `configs/i18n/ru.yml`, в
   `en.yml` дублируем тот же текст (русский) с пометкой `# TODO: en` —
   когда сядем за английскую локализацию, переводчик пройдётся по
   маркерам.
2. **Группы — по доменам**, не по экранам.
   - Уже существующие группы используем в первую очередь
     (`galaxy`, `mission`, `error`, `message`, `unitInfo`, …).
   - Новые группы заводим только когда не хватает: предлагаются
     `fleet`, `rocket`, `alien`, `expedition`, `colonize`, `paymentUi`,
     `adminUi`, `wikiUi`, `chat`, `authUi`, `confirm` (для общих
     кнопок).
3. **Naming-конвенция ключей**: `lowerCamelCase` с точечной иерархией
   внутри группы. Соответствует i18next/современной практике, читается
   ровно, не конфликтует с YAML. **Конвенция единая для всего
   проекта** — на legacy-формат `SCREAMING_SNAKE_CASE` не
   ориентируемся, переименовываем разово в Ф.1bis (см. ниже).
   - Группа = домен (`fleet`, `rocket`, `auth`, `error`, …) в
     `lowerCamelCase` — верхний уровень YAML. Существующие группы
     `Achievements`, `Galaxy`, `UnitInfo` и т.п. тоже переименуются
     (`achievements`, `galaxy`, `unitInfo`).
   - Ключ = действие/контекст в `lowerCamelCase`, точка отделяет
     подкатегории: `recall.toastTitle`, `recall.toastBody`,
     `raidWarning.subject`, `raidWarning.body`.
   - Параметры — именованные `{{name}}` (детали в п.4 ниже).
   - Lint-правило (Ф.8) проверяет конвенцию **для всех групп и
     ключей** без исключений.
4. **Плейсхолдеры**: `{{name}}` — именованные, единый формат для
   backend и frontend. Читабельнее в YAML, безопаснее при переводе
   (переводчик не сломает порядок), позволяет менять порядок слов в
   en-локали без `%1$s`-синтаксиса. **`%s`/`%d` не поддерживаются —
   стандарт проекта един**.
   - Legacy `%s`/`%d` в `ru.yml` (~64 ключа) конвертируем разово в
     `{{name}}` в Ф.1 (см. ниже). `en.yml` плейсхолдеров не содержит
     (там голый русский, переведут после Ф.6). Конвертация ru.yml —
     ручная, ~1-2 часа: имена параметров читаются из контекста
     (`Игрок %s исключен` → `{{username}}`, `%s металла` →
     `{{metal}}`). Callsites сейчас не вызывают `Tr` для этих ключей,
     значит миграция ключей не ломает работающий код — он подключится
     уже на новом формате в Ф.3-Ф.4.
   - Lint (Ф.8) запрещает `%s`/`%d` во всех YAML-файлах i18n.
   - `Bundle.Tr` и `sprintfLite` поддерживают **только** `{{name}}` —
     старая variadic-перегрузка `Tr(lang, group, key, args ...any)`
     удаляется, остаётся одна подпись с `vars map[string]string`.
5. **Никаких ручных русских строк в коде**. Запрет проверяется CI
   (см. Ф.8: lint-правило).
6. **Маркер `[group.key]`** при отсутствии ключа — это
   нормальный signal, видим в UI сразу. Не падаем.
7. **`tf(group, key, fallback, ...args)`** допустим только в редких
   местах, где ключ может отсутствовать в legacy-словаре. По умолчанию
   используем `t()` и добавляем ключ.

---

## 2. План фаз

### Ф.1 — Аудит ключей и подготовка инструментов (≤ 1 день)

**Что сделать**:
1. Скрипт `backend/cmd/tools/i18n-audit/main.go`:
   - Сканирует `frontend/src` и `backend` на string-литералы с
     кириллицей.
   - Исключает: комментарии (`//`, `/* */`, JSDoc), test-файлы
     (`*_test.go`, `*.test.ts`, `*.spec.tsx`), `console.error`,
     `slog.*`, `fmt.Errorf` с %w (внутренние ошибки), `panic(...)`.
   - Для каждой строки эвристически предлагает `group.key` (по имени
     папки/файла) и проверяет, есть ли точное совпадение в `ru.yml`.
   - Выход: `docs/plans/33-i18n-audit-report.md` с таблицей
     «файл:строка → литерал → предложенный ключ → найдено в ru.yml?».
2. Дополнить `Makefile`: `make i18n-audit` запускает скрипт и пишет
   отчёт.
3. Прогнать сейчас — получить точное число строк (оценка ~1300+199),
   распределение по файлам, % уже покрытых легаси-ключей.
4. **Конвертация `ru.yml`: `%s`/`%d` → `{{name}}`**. 64 ключа
   (`grep -nE '%[sd]' configs/i18n/ru.yml`), ручной проход. Имена
   параметров — по смыслу: `username`, `metal`, `silicon`, `hydrogen`,
   `coords`, `count`, `percent`, `days`, `min`, `max` и т.п. Документ
   `docs/plans/33-i18n-placeholder-glossary.md` фиксирует список
   стандартных имён параметров — чтобы в новых ключах не плодить
   синонимы (`user` vs `username` vs `name`).
5. CI-проверка `i18n_no_printf_test.go`: `grep` по обоим YAML — если
   найден `%s`/`%d`, тест красный.

**Готово, когда**: отчёт сгенерирован, в `ru.yml`/`en.yml` нет ни
одного `%s`/`%d`, glossary опубликован.

---

### Ф.1bis — Массовое переименование словарей в новую конвенцию (1 день)

**Контекст**: 1481 ключ × 23 группы в `SCREAMING_SNAKE_CASE` —
наследие импорта из legacy `na_phrases`. Делаем разовую трансформацию
до начала Ф.3-Ф.4, чтобы новый код сразу писался на финальных именах.

**Что делаем**:

1. **Скрипт `backend/cmd/tools/i18n-rename/main.go`**:
   - Читает `configs/i18n/ru.yml`, `en.yml`.
   - Группы: `Achievements` → `achievements`, `ArtefactInfo` →
     `artefactInfo`, `AssaultReport` → `assaultReport`,
     `EspionageReport` → `espionageReport`, `UnitInfo` → `unitInfo`,
     `UserAgreement` → `userAgreement`, `AutoMessages` →
     `autoMessages`. Уже-`lowerCamelCase` (`buildings`, `error`,
     `global`, `info`, `mission`) — без изменений.
   - Ключи: `FIRST_METAL` → `firstMetal`, `MOON_BONUS_ONLY` →
     `moonBonusOnly`, `BMAT_ATTER_DAMAGE` → `bmatAtterDamage`,
     `RECYLCING_REPORT_SUBJECT` → `recylcingReportSubject`. Алгоритм:
     split по `_`, первое слово lower, остальные — capitalize-first.
     Цифры остаются как есть (`FIRST_DAY_2` → `firstDay2`).
   - Опечатки legacy (`RECYLCING` вместо `RECYCLING`,
     `ESPIOANGE` вместо `ESPIONAGE`) — не исправляем в этом проходе,
     это отдельная зачистка после.
   - Output: новые `ru.yml` / `en.yml` + `i18n-rename-map.json`
     (старое имя → новое, для построения rename-таблицы в коде).
2. **Прогон**: `make i18n-rename`. Diff в YAML огромный
   (1481 переименование) — но это **один коммит**, пересекать с
   другими работами нельзя. После прогона смержить сразу.
3. **Поднимаем `LOCALE_VERSION` в
   [frontend/src/i18n/i18n.tsx](../../frontend/src/i18n/i18n.tsx)** —
   инвалидируем localStorage у всех сессий.
4. **Использования в коде**: на момент Ф.1bis Tr ещё не подключен
   нигде кроме теста ([i18n_test.go](../../backend/internal/i18n/i18n_test.go))
   и 3 фронтовых файлов. Их обновить вручную (≤10 мест) по
   rename-map.
5. **Удалить старые ключи из glossary Ф.1**, если попали — там
   только имена параметров, не ключей, но проверить.

**Риски**:
- **Конфликты регистра** на case-insensitive файловых системах — нет,
  YAML-ключи sensitive, конфликта не будет. Но проверить, что после
  переименования нет дубликатов (`FOO_BAR` и `FoO_bAr` оба → `fooBar`).
  Скрипт падает на дубликате с понятным сообщением.
- **Цифры в начале сегмента**: `2_DAYS_LEFT` → `2DaysLeft` —
  невалидный JS-идентификатор. Если такие есть — вставлять префикс
  `key_` или переформулировать руками. Скрипт ругается, не
  конвертирует молча.

**Готово, когда**:
- `ru.yml`/`en.yml` целиком в новой конвенции.
- `i18n-rename-map.json` сохранён в репо как audit-trail.
- Тесты i18n зелёные.
- 3 фронтовых вызова `t()` обновлены.

---

### Ф.2 — Backend: Tr доступен в сервисах и хендлерах (1-2 дня)

**Контекст**: сейчас `Bundle` живёт только в HTTP-слое. Сервисам, что
формируют subject/body inbox-сообщений (`fleet`, `rocket`, `alien`,
`colonize`, `expedition`, `officer`), Tr недоступен.

**Что сделать**:
1. `internal/i18n/Bundle` инжектируется в сервисы через DI. Аккуратно:
   язык получателя зависит от `users.lang` (поле уже есть? — проверить
   миграции; если нет — добавить с дефолтом `ru`). Передавать `lang`
   первым аргументом в Tr.
2. Helper `func (b *Bundle) ForUser(ctx, userID) Lang` — читает один раз,
   кеширует в `context` через ключ. Чтобы не делать лишних запросов в
   hot-path (например, `fleet/spy.go` пишет 3 сообщения).
3. Где `httpx.WriteError(w, r, code, "русский текст")` — заменяем на
   `httpx.WriteErrorTr(w, r, code, group, key)`, где `WriteErrorTr`
   читает `lang` из middleware (см. п.4) и форматирует.
4. Middleware `i18n.LangMiddleware` — определяет язык запроса:
   - JWT-claim `lang` (для авторизованных),
   - cookie `oxsar.lang` (для гостевых /login),
   - Accept-Language (fallback),
   - default `ru`.
   Кладёт в request-context.
5. Проверить, что worker (background-job'ы) тоже получает Bundle —
   передаётся в `event.Worker` через конструктор.
6. **Переписать `Bundle.Tr` под `{{name}}`**: единственная подпись —
   `Tr(lang, group, key string, vars map[string]string) string`.
   Подстановка через `strings.NewReplacer` (`{{name}}` → значение).
   Старая variadic-форма `args ...any` с `fmt.Sprintf` удаляется
   вместе с тестом
   [i18n_test.go](../../backend/internal/i18n/i18n_test.go) на `%s`
   (тест переписать под новый формат). `vars=nil` допустим — для
   ключей без плейсхолдеров.
7. **Frontend `sprintfLite` → `interpolate`**: переписать на
   `{{name}}` (regex `\{\{(\w+)\}\}` → значение из vars-объекта).
   `%s`/`%d` поддержку выкинуть. `useTranslation` форма:
   `t('group', 'key', { name: 'value' })`. Текущие 3 callsites
   `useTranslation` — обновить.

**Готово, когда**: сервисы вызывают `b.Tr(lang, group, key, vars)`,
worker-events тоже, фронтовый `t()` принимает объект `vars`,
`%s`/`%d` нет ни в коде i18n, ни в YAML.

**Риски**: language detection при write-only worker'ах (alien spawn,
colonize). У них нет HTTP-запроса. Решение: язык получателя берём из
`users.lang` в момент формирования сообщения, через
`bundle.ForUser(ctx, userID)`.

---

### Ф.3 — Backend: вынос строк по доменам (3-5 дней, серия PR)

Каждый PR — один пакет, ≤400 строк diff (см. CLAUDE.md правила).
Очерёдность по «грязности» (где больше всего хардкода) и риску:

| # | Пакет / файлы | Группа в YAML | Кол-во строк | Заметки |
|---|---|---|---|---|
| 1 | [fleet/colonize.go](../../backend/internal/fleet/colonize.go) | `colonize` (новая) | ~12 | inbox-сообщения, конкатенация %s |
| 2 | [fleet/spy.go](../../backend/internal/fleet/spy.go), [fleet/raid_warning.go](../../backend/internal/fleet/raid_warning.go) | `espionageReport` (есть) | ~10 | переиспользовать существующие ключи |
| 3 | [fleet/expedition.go](../../backend/internal/fleet/expedition.go) | `expedition` (новая) | ~15 | ветви ‒ outcomes |
| 4 | [fleet/events.go](../../backend/internal/fleet/events.go), [fleet/attack.go](../../backend/internal/fleet/attack.go), [fleet/acs_attack.go](../../backend/internal/fleet/acs_attack.go), [fleet/moon_destruction.go](../../backend/internal/fleet/moon_destruction.go), [fleet/stargate.go](../../backend/internal/fleet/stargate.go) | `mission` + `assaultReport` (есть) | ~25 | проверить существующие ключи по тексту |
| 5 | [rocket/events.go](../../backend/internal/rocket/events.go) | `rocket` (новая) | ~12 | %d + конкат |
| 6 | [alien/alien.go](../../backend/internal/alien/alien.go), [alien/holding.go](../../backend/internal/alien/holding.go), [alien/pay.go](../../backend/internal/alien/pay.go) | `alien` (новая) | ~15 | "Инопланетяне" → `alien.raceName` |
| 7 | [officer/service.go](../../backend/internal/officer/service.go) | `officer` (новая) | ~6 | renewal-сообщения |
| 8 | [payment/packages.go](../../backend/internal/payment/packages.go) | `payment` (есть) | ~5 | Label пакетов вытащить в YAML |
| 9 | [aiadvisor/service.go](../../backend/internal/aiadvisor/service.go) | `aiadvisor` (новая) | ~8 | системные prompts (см. ниже) |
| 10 | [referral/service.go](../../backend/internal/referral/service.go), [achievement/service.go](../../backend/internal/achievement/service.go), [galaxyevent/service.go](../../backend/internal/galaxyevent/service.go), [dailyquest/service.go](../../backend/internal/dailyquest/service.go), [goal/notifier.go](../../backend/internal/goal/notifier.go) | соответствующие | ~15 | мелочи |
| 11 | [admin/*.go](../../backend/internal/admin/) | `admin` (новая) | ~10 | админ-сообщения, видны только админу — но всё равно через i18n |
| 12 | [wiki/service.go](../../backend/internal/wiki/service.go), [chat/hub.go](../../backend/internal/chat/hub.go), [search/handler.go](../../backend/internal/search/handler.go) | `wiki` / `chat` / `global` | ~8 | системные сообщения чата |

**Особый случай — aiadvisor**: системный промпт LLM включает русский
текст. Это **не** UI-строка, но всё равно зависит от языка пользователя
(совет на русском пользователю-русскому). Промпт выносим в
`configs/i18n/<lang>.yml` группа `aiadvisor`, ключ `systemPrompt` —
переводчик-разработчик потом сделает английскую версию когда понадобится.

**Готово, когда**: `grep -nE '"[^"]*[А-Яа-яЁё][^"]*"' backend/internal/`
по списку выше возвращает 0 результатов (с учётом исключений из Ф.1).

---

### Ф.4 — Frontend: вынос по экранам (5-8 дней, серия PR)

**Оценка**: ~46 .tsx-файлов с кириллицей, медиана ~30 строк литералов.
Около **15 PR** по 3-4 экрана. Каждый PR ≤400 строк diff.

**Приоритет** (по частоте посещения экрана):
1. `LoginScreen`, `RegisterScreen`, `App.tsx` (header/menu) → `authUi`,
   `main`.
2. `OverviewScreen`, `ResourceScreen`, `BuildingsScreen`,
   `ResearchScreen`, `ShipyardScreen` → `buildings`, `resource`, `main`.
3. `FleetScreen`, `GalaxyScreen` → `mission`, `galaxy`.
4. `MessagesScreen`, `BattleSimScreen`, `BattlestatsScreen` → `message`,
   `assaultReport` (уже частично).
5. `MarketScreen`, `ArtefactMarketScreen`, `RepairScreen` → `resource`,
   `artefactInfo`.
6. Остальное (alliance, profile, achievements, wiki, admin, …).

**Шаблон работы для одного экрана**:
1. Добавить хук `const { t } = useTranslation('group')` на верхнем
   уровне компонента.
2. Заменить все литералы `'Текст'` / `\`Текст ${x}\`` на `t('key')` /
   `t('key', { x })`.
3. Если literal — JSX-текст (`<button>Войти</button>`), оставить как
   `<button>{t('loginButton')}</button>`.
4. Если literal — атрибут (`placeholder="..."`), завернуть в `t()`.
5. Toast'ы: `toast.show('success', t('toast.fleetSent.title'),
   t('toast.fleetSent.body', { mission, g, s, pos }))` — все три
   параметра через i18n.
6. Все новые ключи добавить в `configs/i18n/ru.yml` (текстом из кода с
   плейсхолдерами `{{mission}}`, `{{g}}`, …) и `en.yml` (тем же
   русским текстом, см. Ф.6).
7. Проверить локально: F5, экран рендерится без `[group.key]` маркеров.

**Особые случаи**:
- **MISSION_LABELS, FLEET_STATUS_LABELS** в FleetScreen.tsx и
  аналогичные map'ы — выносим в `configs/i18n/ru.yml` группа
  `mission`, ключи `byId.6`, `byId.7`, … и читаем через
  `t('mission.byId.' + missionType)`.
- **Числа с единицами**: `${secs}с`, `${m}м`, `${h}ч ${m}м` —
  отдельный helper `formatDuration(secs, t)`, ключи `time.unit.sec`,
  `time.unit.min`, `time.unit.hour`, `time.unit.day` и шаблоны
  `time.duration.hms = '%d ч %d мин %d с'`. Переводчик сам решит
  порядок и суффиксы.
- **`toLocaleString('ru-RU')`** на цифрах — оставляем
  `toLocaleString(lang)` где `lang` берётся из `useTranslation()`.

**Готово, когда**: фронт сборка проходит ESLint-правилом «no Cyrillic
literals in JSX» (Ф.8), и `make i18n-audit` показывает 0 unmapped в
`frontend/src`.

---

### Ф.5 — Чистка: общие тексты, error messages, описания юнитов (1-2 дня)

1. **`httpx.WriteError`** возвращает структурированный JSON
   `{ error: { code, message } }`. Сейчас message — захардкоженный
   русский. Меняем на схему: код возвращает `{ code: 'fleet.noShips' }`,
   фронт сам делает `t('error.' + code)`. Список кодов
   фиксируется в `backend/internal/httpx/errors.go`.
2. **Confirm-диалоги** ([frontend/src/ui/Confirm.tsx](../../frontend/src/ui/Confirm.tsx)):
   props `title`, `body`, `confirmText`, `cancelText` — везде вызовы
   передают русский текст. Меняем на ключ:
   `<Confirm titleKey="fleet.recall.confirmTitle" .../>` или передаём
   уже переведённую строку (выбор сделать в PR).
3. **Описания юнитов / зданий**: проверить, что
   [frontend/src/api/catalog.ts](../../frontend/src/api/catalog.ts)
   подтягивает имена через i18n, а не хранит в JSON. Если хранит —
   удалить дубль, читать из группы `unitInfo`.
4. **Wiki заголовки** ([backend/internal/wiki/service.go:47](../../backend/internal/wiki/service.go#L47)):
   `Title string // "Здания"` — вытащить в i18n.

---

### Ф.6 — Английская локализация: процесс перевода (≤ 1 день, без переводов)

> **Этот шаг не делает английских переводов.** Только готовит инфру и
> процесс — переводчик пройдётся отдельно, когда придёт время.

1. После всех Ф.3-Ф.5 ключи в `ru.yml` и `en.yml` совпадают по
   структуре. В новых местах `en.yml` содержит русский текст
   с пометкой комментария `# TODO: en` над ключом.
2. Скрипт `make i18n-en-todo` — печатает все `# TODO: en` ключи и
   количество. Это и есть TODO-лист переводчика.
3. CI-проверка `i18n_consistency_test.go`: множества ключей в `ru.yml`
   и `en.yml` совпадают. Дрейф ловим в PR.
4. UI: переключатель языка в `SettingsScreen` уже есть? — проверить.
   Если нет, добавить (тоже в этой фазе, ≤30 строк).

---

### Ф.7 — automsg-шаблоны в i18n, удаление admin-эндпоинта (1-2 дня)

Шаблоны `automsg_defs` (title, body) сейчас живут в БД через миграции
[0016](../../migrations/0016_automsg.sql),
[0019](../../migrations/0019_automsg_inactivity.sql) и редактируются
админом через `GET/PUT /api/admin/automsgs`
([admin/handler.go:205-259](../../backend/internal/admin/handler.go#L205-L259)).
Решено: **возможность править тексты в рантайме админу не нужна** —
шаблонов всего 4, правки идут через PR (history + review).

**Что делаем**:

1. **Тексты переезжают в `configs/i18n/<lang>.yml` группа
   `autoMessages`**. Группа уже существует с родственными ключами;
   добавляем 4 наших (`welcome`, `starterGuide`, `firstAttackReceived`,
   `inactivityReminder`) парами `*.title` / `*.body`.

2. **Плейсхолдеры `{{name}}` остаются как есть** — это и есть единый
   формат проекта (см. п.4 раздела «Принципы»). Конвертация шаблонов
   в YAML — копипаст из миграций 0016/0019 без изменений.

3. **`automsg_defs` (таблица) — удалить**. Миграция `down`-вариант:
   таблица + INSERT'ы базовых шаблонов (для отката).

4. **`automsg_sent` (таблица идемпотентности) — оставить**, она про
   факт отправки, не про текст. FK `key → automsg_defs(key)` снимаем
   (целевой таблицы больше нет), `key` остаётся `text` без FK. Список
   допустимых ключей живёт в коде как `var KnownKeys = []string{...}`,
   проверяется на `Send` (defensive — не критично).

5. **`folder` (целевая папка inbox)** — переезжает в `var
   AutomsgFolders = map[string]int{"WELCOME": 2, ...}` рядом с
   `KnownKeys` в `internal/automsg/`. Список из 4 строк, отдельная
   таблица под это излишня.

6. **Сервис `automsg.Send` переписать**: вместо `SELECT FROM
   automsg_defs` → `bundle.Tr(lang, "autoMessages", key+".title",
   vars)` и `..body`. Получает `*i18n.Bundle` через DI (уже
   подготовлено в Ф.2). **Сигнатура остаётся прежней**: `Send(ctx,
   tx, userID, key, vars map[string]string)` — `{{name}}`-подстановка
   совместима с текущим вызовом без переделок.

7. **Существующие вызовы `Send`** ([auth/service.go:151-152](../../backend/internal/auth/service.go#L151-L152)
   и др.) — менять не нужно, передача vars остаётся как есть. Только
   добавляется `lang` получателя через `bundle.ForUser(ctx, userID)`
   внутри сервиса.

8. **Удалить admin-эндпоинт**:
   - `ListAutomsgs` / `UpdateAutomsg` в
     [admin/handler.go:205-259](../../backend/internal/admin/handler.go#L205-L259) — снести.
   - Роуты `/api/admin/automsgs` в роутере — снести.
   - Frontend admin-страницу (если есть) — снести вместе.
   - Из `api/openapi.yaml` — удалить пути.

9. **Миграция `0NN_drop_automsg_defs.sql`**:
   ```sql
   -- +goose Up
   ALTER TABLE automsg_sent DROP CONSTRAINT IF EXISTS automsg_sent_key_fkey;
   DROP TABLE IF EXISTS automsg_defs;
   -- +goose Down
   -- (восстановление возможно из 0016 + 0019, дублировать здесь не будем)
   ```

**Готово, когда**:
- `automsg_defs` нет ни в БД, ни в коде.
- Тексты 4 шаблонов в `configs/i18n/ru.yml` группа `autoMessages`,
  с пометкой `# TODO: en` в `en.yml`.
- Admin-роуты `/api/admin/automsgs` удалены, smoke-test админки
  не отваливается.
- `make test` зелёный (тест в [auth/service_test.go](../../backend/internal/auth/service.go) проверяет welcome —
  убедиться, что mock-Bundle ему отдаётся).

---

### Ф.8 — CI и lint-правила, чтобы регрессии не вернулись (≤ 1 день)

1. **Backend**: golangci-lint custom-rule `no-cyrillic-literals` —
   запрещает кириллицу в string-литералах прод-файлов. Whitelist:
   тестовые файлы, миграции (там фикстуры), tools/import-phrases.
2. **Frontend**: eslint-rule `no-cyrillic-literals` (или regex через
   `no-restricted-syntax`) — запрещает кириллицу в JSXText, StringLiteral,
   TemplateLiteral вне `i18n.tsx`, `*.test.tsx`, `*.spec.tsx`.
3. **YAML-lint** `i18n_no_printf_test.go` (из Ф.1) — `%s`/`%d` в
   `configs/i18n/*.yml` запрещены, тест красный.
4. **CI step** `make i18n-check` — запускает все lint + audit, падает
   если что-то осталось.
5. **PR-template**: пункт «Все новые user-facing строки добавлены в
   `configs/i18n/ru.yml` (и en.yml с пометкой)?».

---

## 3. Риски и mitigation

- **Большой diff во фронте**: дробим по экранам, ≤400 строк diff на PR
  (CLAUDE.md). 15+ PR — но каждый ревью простой (механический
  refactor).
- **Перепутали ключи**: маркер `[group.key]` в UI делает ошибку
  видимой при первом же рендере. Включить smoke-тест после каждого PR
  (открыть экран в dev-сервере).
- **Дрейф ru.yml ↔ en.yml**: CI-тест `i18n_consistency_test.go` ловит.
- **Performance**: `useTranslation` создаёт `t` на каждый рендер. На
  массовых списках (Galaxy, Fleet) проверить — если заметно, мемоизируем
  через `useMemo(() => ({ t }), [t])`.
- **localStorage кеш ломается при изменении ключей**: поднимаем
  `LOCALE_VERSION` в `i18n.tsx` после каждой большой фазы.
- **Перевод и порядок слов**: `{{name}}` решает проблему — переводчик
  свободно меняет порядок плейсхолдеров в en. Нет нужды в позиционных
  `%1$s`/`%2$d`.
- **Конвертация 64 legacy-ключей в Ф.1**: имена параметров
  выбираются из glossary (`username`, `metal`, …), но в редких ключах
  смысл параметра неочевиден (`SHIPS_EXIST: "в наличии %s"` —
  количество? список?). Если упёрлись — открыть legacy-вызов в
  `d:\Sources\oxsar2` для уточнения, записать в glossary.

---

## 4. Оценка трудозатрат

| Фаза | Дни | Комментарий |
|---|---|---|
| Ф.1 audit | 1 | tools/i18n-audit + отчёт + конвертация %s→{{name}} |
| Ф.1bis rename | 1 | tools/i18n-rename + прогон + LOCALE_VERSION++ |
| Ф.2 backend infra | 1-2 | Bundle в сервисах, middleware, lang в users |
| Ф.3 backend strings | 3-5 | 12 PR по доменам |
| Ф.4 frontend strings | 5-8 | ~15 PR по экранам |
| Ф.5 cleanup | 1-2 | error codes, Confirm, юниты |
| Ф.6 en process | ≤1 | без самих переводов |
| Ф.7 automsg | 1-2 | переезд в YAML + удаление admin-эндпоинта |
| Ф.8 CI rules | ≤1 | linter rules + make-target |
| **Итого** | **15-23 дня** | человеко-дни в фокусе; календарно ×1.5-2 (ревью, CI, переключения), реально 2-3 спринта |

---

## 5. Definition of Done

- [ ] `make i18n-audit` показывает 0 unmapped строк в `frontend/src` и
      в backend (вне whitelist'а).
- [ ] CI-rule `no-cyrillic-literals` зелёный.
- [ ] CI-rule `i18n_consistency` (одинаковые ключи в ru/en) зелёный.
- [ ] Все subject/body inbox-сообщений идут через `Bundle.Tr` с языком
      получателя.
- [ ] Переключатель языка в `SettingsScreen` сохраняет выбор в
      `users.lang` и в cookie.
- [ ] При выборе `en` UI и inbox показывают русский текст (он же
      «исходный»), `[group.key]` нигде не светится.
- [ ] Маркер `# TODO: en` в `en.yml` — единственное, что отличает его
      от `ru.yml`.
- [ ] Таблица `automsg_defs` удалена, тексты 4 шаблонов в
      `configs/i18n/<lang>.yml` группа `autoMessages`. Эндпоинт
      `/api/admin/automsgs` снесён, openapi.yaml обновлён.
- [ ] Запись в [docs/simplifications.md](../simplifications.md) если
      что-то отложили.
- [ ] Запись итерации в `docs/project-creation.txt`.

---

## 6. Связанные документы

- [oxsar-spec.txt §10.3](../oxsar-spec.txt) — дизайн i18n.
- [docs/plans/](.) — план идёт после 32 (multi-instance) и до
  следующих gameplay-итераций.
- [backend/internal/i18n/](../../backend/internal/i18n/) — инфра
  бэкенда.
- [frontend/src/i18n/i18n.tsx](../../frontend/src/i18n/i18n.tsx) —
  инфра фронта.
- [configs/i18n/ru.yml](../../configs/i18n/ru.yml),
  [configs/i18n/en.yml](../../configs/i18n/en.yml) — словари.
