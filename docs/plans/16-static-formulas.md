# План 16: Перевод formula-DSL на статические Go-функции

## Статус: НЕ НАЧАТ

## Контекст

Сейчас формулы производства, потребления и стоимости построек хранятся как строки в
`configs/construction.yml` и вычисляются во время выполнения через самописный
DSL-движок (`backend/pkg/formula/`). У этого подхода есть плюсы (формулы как данные,
без перекомпиляции), но и минусы:

- Парсинг строк при старте + синхронизация кэша (`formulaCache` с `sync.RWMutex`).
- Дополнительный слой абстракции, который усложняет трассировку и отладку.
- Нет явных сигнатур — неясно какие аргументы нужны формуле без чтения строки.
- Ошибки в формулах всплывают только в runtime.

**Цель плана** — заменить DSL статическими Go-функциями, сохранив источник истины
(legacy `na_construction.sql` → `configs/construction.yml`) для документации и
валидации, но убрав runtime-интерпретацию.

**Затронутые файлы:**
- `backend/pkg/formula/` — DSL-движок (удалить или оставить только для валидации)
- `configs/construction.yml` — остаётся как документация и источник для кодогенерации
- `backend/internal/economy/formulas.go` — сюда переносятся все формулы
- `backend/internal/planet/service.go` — убрать `evalProd`/`evalCons`/`formulaCache`

---

## Анализ: что именно переводить

### Prod-формулы (8 зданий)

| Здание | Ресурс | Формула | Зависимости |
|--------|--------|---------|-------------|
| metalmine | metal | `floor(30 * L * pow(1.1+T23*0.0006, L))` | level, tech[23] |
| silicon_lab | silicon | `floor(20 * L * pow(1.1+T24*0.0007, L))` | level, tech[24] |
| hydrogen_lab | hydrogen | `floor(10 * L * pow(1.1+T25*0.0008, L) * (-0.002*temp+1.28))` | level, tech[25], temp |
| moon_hydrogen_lab | hydrogen | `floor(100 * L * pow(1.1+T25*0.0008, L) * (-0.002*temp+1.28))` | level, tech[25], temp |
| solar_plant | energy | `floor(20 * L * pow(1.1+T18*0.0005, L))` | level, tech[18] |
| hydrogen_plant | energy | `floor(50 * L * pow(1.15+T18*0.005, L))` | level, tech[18] |
| solar_satellite | energy | (формула из YAML) | level |
| gravi | energy | `{basic} * pow(3, (L-1))` | level, basic |

### Cons-формулы (5 зданий)

| Здание | Ресурс | Формула | Зависимости |
|--------|--------|---------|-------------|
| metalmine | energy | `floor(10 * L * pow(1.1-T18*0.0005, L))` | level, tech[18] |
| silicon_lab | energy | `floor(20 * L * pow(1.1-T18*0.0005, L))` | level, tech[18] |
| hydrogen_lab | energy | `floor(20 * L * pow(1.1-T18*0.0005, L))` | level, tech[18] |
| hydrogen_plant | energy | `floor(10 * L * pow(1.1-T18*0.0005, L))` | level, tech[18] |
| moon_hydrogen_lab | energy | `floor(200 * L * pow(1.1-T18*0.0005, L))` | level, tech[18] |

### Charge-формулы (4 базовых паттерна, ~80+ зданий)

| Паттерн | Здания (примеры) | Формула |
|---------|-----------------|---------|
| factor=1.5 | metalmine, exch_office | `floor(basic * pow(1.5, L-1))` |
| factor=1.6 | silicon_lab | `floor(basic * pow(1.6, L-1))` |
| factor=1.8 | hydrogen_plant, robotic_factory | `floor(basic * pow(1.8, L-1))` |
| factor=2.0 | большинство зданий | `basic * pow(2, L-1)` |
| factor=3.0 | gravi | `basic * pow(3, L-1)` |
| factor=1.2 | exchange | `floor(basic * pow(1.2, L-1))` |
| прямые константы | solar_plant | `50 * pow(1.5, L)` |

---

## Задачи

### Задача 1: Добавить статические prod/cons-функции в economy/formulas.go

Добавить в `backend/internal/economy/formulas.go` функции для каждого здания:

```go
// Prod-функции
func MetalmineProdMetal(level, techLaser int) float64
func SiliconLabProdSilicon(level, techSilicon int) float64
func HydrogenLabProdHydrogen(level, techHydrogen, tempC int) float64
func MoonHydrogenLabProdHydrogen(level, techHydrogen, tempC int) float64
func SolarPlantProdEnergy(level, techEnergy int) float64
func HydrogenPlantProdEnergy(level, techEnergy int) float64
func SolarSatelliteProdEnergy(level int) float64
func GraviProdEnergy(level int, basicEnergy int64) float64

// Cons-функции
func MetalmineConsEnergy(level, techEnergy int) float64
func SiliconLabConsEnergy(level, techEnergy int) float64
func HydrogenLabConsEnergy(level, techEnergy int) float64
func HydrogenPlantConsEnergy(level, techEnergy int) float64
func MoonHydrogenLabConsEnergy(level, techEnergy int) float64
```

Каждая функция — прямой перенос формулы из YAML без DSL. Сигнатуры явные,
проверяются компилятором.

**Файл:** `backend/internal/economy/formulas.go`
**Тесты:** `backend/internal/economy/formulas_test.go` — добавить тест для каждой
новой функции с эталонными значениями из `formula_test.go::TestRealLegacyFormulas`.

---

### Задача 2: Обобщённая charge-функция с явным factor

`CostForLevel` уже существует (`base * factor^(level-1)`). Нужно:

1. Убедиться, что все charge-паттерны из YAML покрываются существующей сигнатурой.
2. Добавить `CostForLevelRound(base int64, factor float64, level int) int64` — вариант
   с `floor()` для зданий, где YAML явно использует `floor({basic} * pow(factor, L-1))`.
3. Задокументировать в `construction.yml` какой factor у каждого здания (уже есть в
   charge-строке, перенести в числовое поле `charge_factor` при следующем импорте).

**Файл:** `backend/internal/economy/formulas.go`

---

### Задача 3: Переписать planet/service.go — убрать evalProd/evalCons/formulaCache

Заменить динамические вызовы DSL прямыми вызовами из economy:

**Было:**
```go
metalPerHour := s.evalProd("metal_mine", "metal", levels, ctxBase)
```

**Станет:**
```go
metalPerHour := economy.MetalmineProdMetal(levels[MetalmineID], tech[TechLaserID])
```

Убрать:
- `formulaCache` struct + методы (`compile`, `evalProd`, `evalCons`)
- импорт `backend/pkg/formula` из planet/service.go

ID-константы (здания и технологии) вынести в `backend/internal/catalog/ids.go`
(или `backend/internal/economy/ids.go`) — без магических чисел.

**Файлы:**
- `backend/internal/planet/service.go`
- новый `backend/internal/economy/ids.go` (или `catalog/ids.go`)

---

### Задача 4: Переписать ResourceReport — убрать formula.Parse в calcBuildingProduction

`planet/service.go:724` — `calcBuildingProduction` и `calcBuildingConsumption` парсят
формулы повторно (вне кэша) для отчёта. Заменить на те же static-функции из задачи 1.

**Файл:** `backend/internal/planet/service.go` (строки 724–755)

---

### Задача 5: Инлайнить валидацию в import-tool и удалить пакет formula

После задач 3 и 4 `backend/pkg/formula/` используется только в двух местах:
- `convert_construction.go::validateFormulas` — парсит формулы при импорте из SQL-дампа.
- `formula_test.go` — тесты самого пакета.

Оба use-case находятся в `import-datasheets`. Вместо сохранения целого DSL-пакета:

1. Перенести логику `validateFormulas` в сам `convert_construction.go` — простая
   проверка на допустимые символы/функции через regexp или минимальный локальный парсер
   (20–30 строк), либо просто убрать валидацию формул (они уже зафиксированы как
   статические Go-функции и больше не интерпретируются).
2. Удалить `backend/pkg/formula/` полностью — все 5 файлов (context.go, lexer.go,
   parser.go, nodes.go, formula_test.go).
3. Удалить импорт `backend/pkg/formula` из всех файлов (после задач 3–4 импортов
   остаться не должно — проверить `grep -r "pkg/formula"`).

**Файлы к удалению:**
```
backend/pkg/formula/context.go
backend/pkg/formula/lexer.go
backend/pkg/formula/parser.go
backend/pkg/formula/nodes.go
backend/pkg/formula/formula_test.go
```

---

### Задача 6: Обновить тесты

1. `backend/internal/economy/formulas_test.go` — тесты для каждой новой prod/cons-функции.
   Эталонные значения взять из `formula_test.go::TestRealLegacyFormulas` (уже есть
   конкретные числа + контекст).
2. `backend/internal/planet/production_test.go` — `TestProductionRatesDSL_MetalMine`
   переименовать в `TestProductionRatesMetal` и убедиться что проходит после смены
   реализации.
3. `backend/pkg/formula/formula_test.go` — оставить как есть (тестирует валидатор).

---

## Что НЕ делаем в этом плане

- Не удаляем `configs/construction.yml` — он остаётся документацией и источником
  для charge-факторов, base-значений и ID-маппинга.
- Не меняем charge-логику в building/research service — `CostForLevel` уже статическая.
- Не меняем баланс — коэффициенты переносятся 1:1 из YAML-строк.

---

## Риски

| Риск | Вероятность | Митигация |
|------|------------|-----------|
| Опечатка в коэффициенте при переносе формулы | Средняя | Тест с эталонными значениями из TestRealLegacyFormulas |
| Нарушение fallback-пути (productionRatesApprox) | Низкая | Fallback не трогаем |
| Пропущенное здание | Средняя | Grep по `evalProd`/`evalCons` после рефакторинга |
| tech ID-константы в нескольких местах | Низкая | Один файл ids.go |

---

## Порядок выполнения

1 → 2 → 6 (тесты для новых функций) → 3 → 4 → 5

Задачи 1 и 2 можно делать параллельно. Задачи 3 и 4 зависят от 1.
Задача 5 (удаление пакета formula) выполняется последней — только после того, как
`grep -r "pkg/formula"` вернёт ноль совпадений за пределами import-datasheets.
