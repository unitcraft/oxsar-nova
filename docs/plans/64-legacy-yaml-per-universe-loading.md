# План 64: origin.yaml override + per-universe balance loading

**Дата**: 2026-04-28
**Статус**: Активный
**Зависимости**: нет блокирующих. Стартовый план серии ремастера 64-74.
**Связанные документы**:
- [62-origin-on-nova-feasibility.md](62-origin-on-nova-feasibility.md) —
  исследование, источник журнала D-NNN.
- [docs/research/origin-vs-nova/roadmap-report.md](../research/origin-vs-nova/roadmap-report.md) —
  Часть I.5 «Сквозные правила реализации» (R1-R5) и Часть II
  (декомпозиция, см. план 64).
- [docs/research/origin-vs-nova/divergence-log.md](../research/origin-vs-nova/divergence-log.md) —
  записи D-022, D-026..D-030 (закрываются этим планом).
- [docs/research/origin-vs-nova/formula-dsl.md](../research/origin-vs-nova/formula-dsl.md) —
  спецификация DSL legacy-формул (`parseChargeFormula`,
  `parseSpecialFormula`).
- [docs/legacy/game-origin-access.md](../legacy/game-origin-access.md) —
  доступ к запущенному origin (источник реальных значений).

---

## Цель

Параметризовать game-nova так, чтобы один и тот же backend мог
обслуживать вселенные с разным балансом:

- **uni01, uni02** (и любые будущие modern-вселенные) — текущие
  YAML / Go-формулы nova, никаких изменений в значениях, ребалансы
  планов 17/18/20/21 остаются как есть.
- **origin** (вселенная-ремастер) — балансовые числа из game-origin
  (`na_construction` + `consts.php`).

Архитектурный приём — **override-файлы**: дефолтные значения для
всех вселенных живут в существующих `configs/buildings.yml` и др.
(modern-баланс), а вселенная origin получает override-файл
`configs/balance/origin.yaml`, перекрывающий нужные значения.

Никаких новых полей в БД (`universes.balance_profile` не
вводится — modern-вселенные пользуются дефолтом, не нужен флаг).
Идентификация вселенной идёт по существующему полю `universes.code`.

---

## Контекст

### Почему override, а не флаг профиля

Альтернатива (которая сначала рассматривалась) — поле
`universes.balance_profile = 'modern'|'legacy'`. Отвергнута:

- В 99% времени это поле было бы `'modern'`. YAGNI.
- Modern — основной случай, origin — частный. Override-схема
  это отражает.
- При добавлении новой modern-вселенной `uni03` ничего не нужно —
  она автоматически наследует дефолт. С полем нужно явно ставить
  `balance_profile='modern'`.
- Меньше сущностей в схеме = меньше миграций, тестов, проверок.

Override-схема: backend проверяет, есть ли файл
`configs/balance/<universe.code>.yaml`. Если есть — применяет
поверх дефолта. Если нет — работает на дефолтных YAML.

### Стратегия ремастера

Зафиксирована планом 62: один движок (game-nova), несколько
вселенных. Современные на дефолте, **origin** — со своим
override.

**Балансовая правда per-universe** (R1 секция «Особый случай:
игровая валюта» — концептуально применимо ко всему балансу):
nova-числа правильны для своих вселенных и **не пересматриваются**.
Расхождения с origin = «вынести в конфиг» / «параметризовать
код-путь», не «исправить число в nova».

### Почему origin.yaml + Go-формулы (B1+B3 гибрид)

Балансовые данные origin живут в трёх местах:

1. `na_construction` (БД, varbinary-строки DSL) — `prod_*`,
   `cons_*`, `charge_*` для зданий/юнитов/исследований/защиты.
2. `config/consts.php` (PHP define) — глобальные коэффициенты
   (`METAL_BASIC_PROD=20` и т.п.).
3. `config/params.php` — конфиги приложения, начальные ресурсы.

Парсеры:
- `Functions.inc.php:41` — `parseChargeFormula()` (без контекста
  планеты).
- `Planet.class.php:592` — `parseSpecialFormula()` (с переменными
  `{temp}`, `{tech=N}`, `{level}`).

DSL поддерживает: переменные, операторы `+ - * /`, функции
`pow`, `floor`, `round`, `min`, `max`. Подробнее — `formula-dsl.md`.

**Не реализуем DSL-парсер на Go.** Вместо этого:

- **B1 — статика в YAML.** Числа `basic_*`, плюс **предвычисленные
  таблицы** для `charge_*` (стоимости по уровням 1..50) и
  `prod_*`/`cons_*` (производство по уровню при стандартной
  температуре `temp=0` и без тех-бонусов). Один раз импортировано
  скриптом — лежит в `configs/balance/origin.yaml` как факт.
- **B3 — динамика в Go.** Те формулы, которые зависят от
  **рантайм-контекста** (температура планеты, уровни
  research-технологий), реализуются **функциями на Go** в
  `internal/origin/economy/`, читающими коэффициенты из
  `origin.yaml` и применяющими их.

Этот гибрид:
- Не создаёт долгоживущей DSL-зависимости в Go-коде.
- Не теряет точность (числа берутся из реального legacy через
  скрипт-импортёр, а не от руки).
- Делает origin-вселенную конфигурируемой через те же механизмы
  YAML, что nova.

---

## Что меняем

### 1. БД — без изменений

**Никакой миграции не требуется.** Идентификация вселенной идёт по
существующему полю `universes.code`. Override-файл подбирается по
имени:

- `universes.code = 'uni01'` → нет файла `configs/balance/uni01.yaml`
  → работает на дефолтных YAML.
- `universes.code = 'origin'` → есть файл `configs/balance/origin.yaml`
  → применяется поверх дефолта.

Имя поддомена origin (например, `origin.oxsar-nova.ru` /
`classic.oxsar-nova.ru`) определяется ADR-0010, но `universes.code`
в БД фиксируется отдельно (короткий идентификатор, snake_case).

### 2. Конфиг `configs/balance/origin.yaml`

Новый файл. Структура — по аналогии с существующими
`configs/buildings.yml`, `configs/units.yml`, `configs/rapidfire.yml`,
но с **числами origin**.

Минимальная структура (детали — в Ф.2):

```yaml
# configs/balance/origin.yaml
# Override-файл для вселенной origin (oxsar2-classic balance).
# Применяется поверх дефолтных configs/{buildings,units,rapidfire}.yml.
# Источник: импорт из projects/game-origin-php/migrations/002_data.sql
# таблиц na_construction, na_rapidfire, na_ship_datasheet + na_options.
# Сгенерировано скриптом cmd/tools/import-legacy-balance/main.go.
# Дата импорта: <YYYY-MM-DD>.

version: 1
universe: origin

# Глобальные коэффициенты (из config/consts.php)
globals:
  metal_basic_prod: 20
  silicon_basic_prod: 10
  hydrogen_basic_prod: 0
  energy_basic_prod: 20
  # ...

# Здания: базовые стоимости + предвычисленные стоимости по уровням
buildings:
  metal_mine:
    basic:
      metal: 60
      silicon: 15
      hydrogen: 0
      energy: 0
    # Предвычисленный charge: cost(level) для level 1..50
    charge_table:
      metal:  [60, 90, 135, 202, 303, 455, 683, 1024, ...]
      silicon: [15, 22, 33, 50, 75, ...]
      # ...
    # Динамика — реализуется в Go (см. internal/origin/economy/buildings.go)
    has_dynamic_production: true  # производство зависит от temp/tech

  hydrogen_lab:
    basic: { metal: 225, silicon: 75, hydrogen: 0, energy: 0 }
    charge_table: { ... }
    has_dynamic_production: true
    # Hint для Go: используется температурный модификатор
    # (-0.002 * temp + 1.28). См. D-029.

  # ... все остальные здания

# Юниты (включая алиен-флот UNIT_A_*, Lancer, Shadow и др.)
units:
  light_fighter:
    cost: { metal: 3000, silicon: 1000, hydrogen: 0 }
    attack: 50
    shield: 10
    hull: 4000
    speed: 12500
    cargo: 50
    fuel_consumption: 20
  # ... остальные

  # Алиен-юниты, отсутствующие в nova (D-027, D-028)
  alien_unit_1:  # UNIT_A_1
    cost: { metal: 0, silicon: 0, hydrogen: 0 }  # не строятся игроком
    attack: 1000
    shield: 200
    hull: 50000
    is_alien: true
  # ... UNIT_A_2..UNIT_A_4

  # Спец-юниты (D-028)
  lancer_ship:
    cost: { metal: ..., silicon: ..., hydrogen: ... }
    # ...
  shadow_ship: { ... }
  ship_transplantator: { ... }
  ship_collector: { ... }
  small_planet_shield: { ... }
  large_planet_shield: { ... }
  armored_terran: { ... }

# Rapidfire — алиен и legacy-спецы
rapidfire:
  light_fighter:
    espionage_probe: 5
    solar_satellite: 5
  # ... все RF из na_rapidfire включая UNIT_A_*

# Технологии — стоимости и формулы прироста
research:
  computer_tech:
    basic: { metal: 0, silicon: 400, hydrogen: 600 }
    charge_table: { ... }
  # ...
```

Файл **большой** (несколько тысяч строк после полного импорта),
но генерируется автоматически (Ф.2), не пишется руками.

### 3. Скрипт импорта `cmd/tools/import-legacy-balance/main.go`

Новый Go CLI, который:

1. Подключается к локальному MySQL `docker-mysql-1` (origin БД).
2. Читает `na_construction`, `na_rapidfire`, `na_ship_datasheet`,
   `na_options`.
3. Для каждой записи парсит `prod_*`/`cons_*`/`charge_*` строки
   по DSL из `formula-dsl.md` — реализован простой evaluator
   только для статических формул (без `{temp}`, `{tech}`).
4. Если формула содержит динамические переменные — записывает
   `has_dynamic_production: true` и hint-комментарий.
5. Для статических — предвычисляет таблицы стоимостей по уровням
   1..50 (50 — потолок legacy).
6. Записывает результат в `configs/balance/origin.yaml`.

```bash
# Запуск (один раз при импорте + регенерация если данные origin меняются):
go run ./cmd/tools/import-legacy-balance \
  --mysql-dsn="root:pass@tcp(localhost:3306)/oxsar_db" \
  --output=configs/balance/origin.yaml
```

CLI **не запускается в продакшене** — это импорт-инструмент.
Сгенерированный YAML коммитится в репо.

### 4. Loader баланса с override-схемой

Расширить `internal/config/` (или создать `internal/balance/`):

```go
// internal/balance/loader.go
package balance

type Bundle struct {
    UniverseCode string
    Buildings    map[string]Building
    Units        map[string]Unit
    Rapidfire    map[string]map[string]int
    Research     map[string]Tech
    Globals      Globals
}

// LoadDefaults — загружает дефолтный bundle (текущие configs/buildings.yml,
// units.yml, rapidfire.yml — modern-вселенные).
func LoadDefaults() (*Bundle, error) { ... }

// LoadFor — загружает bundle для конкретной вселенной по её code.
// Если для universeCode существует configs/balance/<code>.yaml —
// он применяется поверх дефолта (override). Иначе возвращается
// чистый дефолт.
func LoadFor(universeCode string) (*Bundle, error) { ... }
```

**Override-применение**: для каждой ключевой секции (`buildings`,
`units`, `rapidfire`, `research`, `globals`) — deep merge: ключи
из override-файла перекрывают одноимённые в дефолте. Ключи, которых
нет в override, остаются дефолтными. Это позволяет origin переопределять
только то, что отличается, не дублируя весь набор.

**Кешируется in-memory** per-universe (один раз при первом обращении;
инвалидация — через restart процесса; в будущем — admin-API).

### 5. Использование Bundle в коде

Любой Go-сервис, который сейчас читает `configs/buildings.yml`
напрямую, переключается на:

```go
bundle := balance.LoadFor(ctx, db, universe.ID)
cost := bundle.Buildings["metal_mine"].ChargeAt(level)
```

**Минимум** для этого плана — обновить:
- `internal/economy/` — функции производства/стоимости
- `internal/building/` — стоимости постройки + время
- `internal/research/` — стоимости исследований
- `internal/shipyard/` — стоимости юнитов

Полное переключение всех потребителей — допускается отложить
до плана 65/66/69 если объём слишком большой; на старте обязательно
переключить только горячие пути (производство, очередь
строительства).

### 6. Динамические формулы (B3 — Go-функции)

Для origin-вселенной — модуль `internal/origin/economy/`:

```go
// internal/origin/economy/buildings.go

// HydrogenProduction — закрывает D-029.
// formula: prod = base_prod * level * pow(1.1, level) * (-0.002*temp + 1.28) * (1 + 0.1*tech)
func HydrogenProduction(bundle *balance.Bundle, level int, planetTemp int, hydroTech int) float64 { ... }

// SolarPlantProduction — энергия от солнечной электростанции
// formula: prod = 20 * level * pow(1.1, level)
func SolarPlantProduction(bundle *balance.Bundle, level int) float64 { ... }

// ... остальные buildings с has_dynamic_production:true
```

Каждая функция:
- читает базовые числа из `bundle.Buildings[name].Globals`
- применяет formula с runtime-параметрами
- покрывается golden-тестами против реальных значений origin
  (сверка через `docker exec docker-mysql-1` запросом к live-планете)

### 7. Тесты

#### 7.1. Unit-тесты импорт-скрипта
- Парсинг простых формул `pow(1.5, level-1)` → правильные числа
- Detection динамических формул → `has_dynamic_production: true`
- Skip формул без leading `_` (служебные поля)

#### 7.2. Golden-тесты `internal/origin/economy/`
- Для каждой динамической функции — golden-таблица «(level, temp, tech) → expected_value»
- Эталоны генерируются из live-origin: SQL-запрос к `na_planet` +
  `na_user`, расчёт через `Planet::updateProduction()`, фиксация
  результата в JSON.
- Golden-файлы в `internal/origin/economy/testdata/*.json`

#### 7.3. Integration-тест per-universe loader
- Создать две universe-записи с разными profile
- Загрузить bundle для каждой — убедиться, что разные числа
- При изменении profile в БД и пересборке кеша — bundle меняется

#### 7.4. Property-based (rapid)
- Для economy-функций: invariants (производство положительное,
  стоимость растёт с уровнем монотонно)

### 8. OpenAPI

Новых публичных endpoint'ов план не вводит. Внутренние loader'ы —
не в OpenAPI.

Если в admin-API понадобится «посмотреть текущий профиль вселенной»
или «пересобрать кеш баланса» — добавить в `/api/admin/universes/`
(не сейчас, по необходимости).

### 9. Документация

- `configs/balance/README.md` — короткое описание: что есть,
  как импортировать, как добавить профиль.
- Пример загрузки + использования Bundle в `internal/balance/doc.go`.

---

## Чего НЕ делаем

- **Не меняем существующие YAML** (`configs/buildings.yml`,
  `units.yml`, `rapidfire.yml`). Они остаются «modern profile».
- **Не реализуем полноценный DSL-evaluator на Go.** Только
  простой парсер для импорта (`pow`, `floor`, `round`, операторы) —
  и только в `cmd/tools/import-legacy-balance/`. В рантайме
  Go использует числа.
- **Не вводим UI** для смены override-файла (через админку или
  игровой UI). Override определяется наличием файла в репо, на
  старте процесса не пересматривается.
- **Не мигрируем данные** между вселенными (игроки uni01 не
  переходят в origin и обратно — у них разные вселенные).
- **Не оптимизируем** (кеширование bundle делаем самым простым —
  один in-memory, перезагрузка через restart). Шардинг по
  процессам, hot-reload и прочее — после необходимости.
- **Не покрываем** D-029 полностью (там больше уровней
  расхождений) — Hydrogen Lab как первый representative-пример,
  остальные динамики — по аналогии в плане 65.

---

## Этапы

### Ф.1. Скаффолд loader без БД-миграций

- Создать скелет `internal/balance/loader.go` с типом `Bundle` +
  функцией `LoadDefaults()` — должен вернуть текущий nova-bundle
  через существующие YAML-loader'ы (рефакторинг существующего кода).
- Никаких миграций БД (override-схема, см. секцию «Что меняем» №1).
- Тест: `LoadDefaults()` возвращает ту же конфигурацию, что
  раньше читалась прямым доступом к YAML. Все existing nova-тесты
  остаются зелёными.

### Ф.2. Импорт-скрипт

- Создать `cmd/tools/import-legacy-balance/main.go`.
- Реализовать parser DSL для статических формул (без `{temp}`,
  `{tech}`):
  - `pow(base, exp)`, `floor(x)`, `round(x)`, `min(a,b)`, `max(a,b)`
  - арифметика `+ - * / ( )`
  - переменная `level`
- Подключение к `docker-mysql-1` через mysql-client-Go (или
  `database/sql` + `go-sql-driver/mysql`).
- SELECT из `na_construction`, `na_rapidfire`, `na_ship_datasheet`,
  `na_options` (базовая часть `consts.php` тоже хочется, но это
  PHP-define — копируем руками в YAML `globals:`).
- Генерация `configs/balance/origin.yaml`.
- Запуск: должен выдать YAML на ~3000-5000 строк (override —
  только то, что отличается от дефолта).
- **Verify**: спот-проверка нескольких значений вручную против
  live-origin через `docker exec docker-mysql-1 mysql -e
  "SELECT name, basic_metal FROM na_construction WHERE name='metal_mine';"`.

### Ф.3. Override-loader для origin

- Расширить `internal/balance/loader.go` — функция
  `LoadFor(universeCode string)`:
  - Если есть `configs/balance/<universeCode>.yaml` → deep merge
    поверх дефолта.
  - Иначе → возвращает дефолт.
- In-memory кеш per-universe.
- Тест: для `LoadFor("uni01")` возвращается дефолт; для
  `LoadFor("origin")` — bundle с origin-числами
  (`bundle.Buildings["metal_mine"].Basic.Metal == 60`).

### Ф.4. Динамические формулы — `internal/origin/economy/`

- Создать модуль с минимальным набором динамических функций
  (приоритет — production, потому что они вызываются на каждый
  тик):
  - `MetalMineProduction(bundle, level, energy_factor) → float64`
  - `SiliconMineProduction`
  - `HydrogenLabProduction(bundle, level, planet_temp, hydro_tech)`
    — закрывает D-029
  - `SolarPlantProduction(bundle, level)`
  - `EnergyConsumption(bundle, level)` для energy-потребителей
- Golden-тесты с эталонами из live-origin (Ф.4.1).

### Ф.4.1. Сбор golden-эталонов

- Запустить дев-стенд origin (:8092 + docker-mysql-1).
- SELECT планет из `na_planet` (5-10 разных temp), их `na_user`.
- Для каждой планеты — выписать `Planet::updateProduction()`
  результат через PHP CLI (новый script `tools/dump-planet-prod.php`):
  ```php
  // tools/dump-planet-prod.php
  // Запуск: php tools/dump-planet-prod.php <planet_id>
  // Output: JSON с полями prod_metal/silicon/hydrogen/energy для всех зданий 1..30
  ```
- Сохранить эти JSON-файлы в
  `internal/origin/economy/testdata/golden_planet_*.json`.
- Go-golden-тест читает эти JSON и сверяет с результатами своих
  функций. Допуск: **точное совпадение** для целых чисел,
  абсолютная погрешность <= 1 единица для дробных (PHP eval может
  округлять иначе чем Go math.Round в редких случаях — допуск
  обоснован в комментарии теста, не превышать).

### Ф.5. Использование bundle в существующих сервисах

- `internal/economy/`, `internal/building/`, `internal/research/`,
  `internal/shipyard/` — переключить чтение на `bundle`.
- Пройти все места, которые сейчас обращаются к
  `configs/buildings.yml`, и заменить на `bundle.Buildings[...]`.
- Убедиться, что **все сценарии modern по-прежнему работают**:
  тесты game-nova должны быть зелёные.
- Создать первый сценарий origin: тест-вселенная с
  `code='origin'` и файлом `configs/balance/origin.yaml`,
  проверить, что у неё свои числа.

### Ф.6. E2E + Smoke

- Поднять dev-стенд game-nova с двумя вселенными:
  `uni01` (modern, без override) и `origin` (с
  `configs/balance/origin.yaml`).
- Создать пользователя в каждой, поставить здание в очередь.
- Verify: стоимость постройки в `origin` соответствует
  `origin.yaml` (Metal Mine lvl 1 → 60/15/0/0), в `uni01` —
  текущий nova.
- Verify: тик производства даёт origin-числа в origin-вселенной.

### Ф.7. Финализация

1. Шапка плана 64 → ✅ Завершён <дата>.
2. Запись в `docs/project-creation.txt` — итерация 64.
3. В `docs/research/origin-vs-nova/divergence-log.md` — пометить
   D-022, D-026, D-027, D-028, D-029, D-030 как ✅ ЗАКРЫТО (план 64)
   с ссылкой на коммит.
4. Обновить `docs/research/origin-vs-nova/roadmap-report.md` —
   план 64 ✅, разблокирует план 65.
5. Коммит(ы) (см. ниже КОММИТЫ).

---

## Тестирование

- Unit-тесты импорт-скрипта (DSL parser).
- Unit + golden-тесты в `internal/origin/economy/`.
- Integration-тест per-universe loader.
- Property-based (rapid) для invariants production.
- E2E smoke по Ф.6.
- Вся существующая game-nova test suite — зелёная (modern не
  сломан).

Покрытие изменённых строк ≥ 85% (R4).

---

## Объём

- Один-три коммита (см. КОММИТЫ).
- Backend: ~800-1500 строк нового Go-кода (loader, importer,
  legacy/economy, тесты).
- YAML: ~3000-5000 строк автогенерации (`origin.yaml`).
- Миграция: ~10 строк SQL.
- Скрипт PHP для golden-эталонов: ~50 строк.

**Время выполнения**: ~2 недели агента в активном темпе.

---

## КОММИТЫ

Рекомендую разбить на три:

1. **`feat(balance): per-universe balance loader + миграция (план 64 Ф.1+Ф.3)`**
   — миграция БД + скелет loader'а с поддержкой modern profile
   (рефакторинг текущего nova-чтения через Bundle), без legacy.
   Если modern продолжает работать — основа надёжная.

2. **`feat(balance): импорт origin → configs/balance/origin.yaml (план 64 Ф.2)`**
   — CLI-импортёр + сгенерированный YAML.

3. **`feat(origin/economy): динамические формулы + golden-тесты (план 64 Ф.4-Ф.6)`**
   — `internal/origin/economy/`, golden-эталоны, переключение
   потребителей на bundle, smoke на двух вселенных.

Каждый коммит — conventional, с trailer `Generated-with: Claude Code`,
ссылка на план 64.

---

## КОНВЕНЦИИ ИМЕНОВАНИЯ (R1)

При создании новых полей строго следовать R1 из roadmap-report.md
Часть I.5. Конкретно для этого плана:

- В БД новых колонок не вводим (override-схема).
- Имя override-файла = `configs/balance/<universes.code>.yaml`
  (snake_case, как и сам code).
- Названия зданий/юнитов в YAML: `metal_mine`, `light_fighter` —
  snake_case, английский, полные слова.
- Алиен-юниты: `alien_unit_1` (НЕ `UNIT_A_1` / `na_alien_1`).
- Структуры в Go: `type Bundle struct { ... }` — английский,
  PascalCase для типов.

Поля под валюту (если их затронем) — по ADR-0009 / план 58
(`oxsar`, `oxsarit`, без постфиксов). В этом плане валюту не
трогаем.

---

## Известные риски

| Риск | Митигация |
|---|---|
| Расхождение в округлении PHP `eval round()` vs Go `math.Round` (D-026 риск) | Golden-тесты с допуском <=1 ед. для дробных, точное совпадение для целых; описать допуск в тесте явно |
| Импортёр пропустил формулу (например, новый оператор в DSL) | После генерации `origin.yaml` — diff-тест: сравнить количество записей с `SELECT COUNT FROM na_construction`; alert при несоответствии |
| Loader кеширует устаревший bundle (после изменения профиля) | На MVP — restart сервера; в будущем (после плана 64) — admin-API `/api/admin/balance/reload` |
| Динамическая формула в Go реализована неточно | Golden-тесты против live-origin — обязательны для каждой динамической функции |
| Большой YAML (~5000 строк) тяжело ревьюить | Не ревьюим построчно — это автогенерация. Спот-проверка нескольких ключевых значений вручную + diff-тест против live-origin |
| Сломали modern (uni01/uni02) при рефакторинге loader'а | Все existing game-nova тесты должны оставаться зелёными после Ф.1 (это критерий приёма Ф.1) |
| origin БД недоступна (нет docker-stack) | Импорт-скрипт читает из `migrations/002_data.sql` mysqldump fallback'ом, если нет live-БД |

---

## Что после плана 64

- Game-nova может работать с двумя балансовыми профилями.
- `origin.yaml` — фиксированная база origin-чисел в репо.
- Динамические производственные формулы реализованы для origin.
- Разблокированы планы:
  - **65** (расширение event-loop) — теперь есть откуда брать
    origin-стоимости постройки.
  - **66** (AlienAI до полного паритета) — алиен-юниты
    уже в `origin.yaml`.
  - **69** (расширение domain-полей) — может опираться на
    `origin.yaml` для origin-значений.

---

## References

- ADR-0009 (currency rebranding) — концептуальный прецедент
  «per-universe конфигурация».
- План 03 (`docs/plans/03-economy-config.md`) — текущая модель
  экономики nova, источник стиля YAML-конфигов.
- План 18 (`docs/plans/18-unit-rebalance.md`) — пример того, как
  ребалансы nova остаются в modern и **не переносятся** в origin.
- `formula-dsl.md` — спецификация origin DSL для импорт-скрипта.
- `divergence-log.md` D-022, D-026..D-030 — конкретные расхождения,
  которые этот план закрывает.
