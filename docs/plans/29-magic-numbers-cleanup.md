---
title: 29 — Замена magic numbers на именованные константы
date: 2026-04-26
status: in-progress
---

> **Статус 2026-04-26**: Ф.1-4 ✅ внедрены. Ф.5 (Mode-enum) — отложено
> до решения, нужно ли. Boy-scout правило — действует для будущих PR.

# План 29: Magic numbers → именованные константы

**Цель**: устранить «магические числа» в Go-коде (event kinds, unit IDs,
mission types, mode), заменив их на существующие именованные константы из
`event/kinds.go` и `economy/ids.go`. Где констант не хватает —
добавить.

**Без изменения логики и поведения** — только улучшение читаемости и
maintainability.

---

## 1. Контекст

В проекте уже есть полный набор именованных констант:

- **`backend/internal/event/kinds.go`** — `event.KindBuildConstruction = 1`,
  `KindBuildFleet = 4`, `KindBuildDefense = 5`, `KindRocketAttack = 16`,
  `KindAttackSingle = 10`, `KindSpy = 11`, `KindAttackAlliance = 12`,
  `KindRecycling = 9`, `KindExpedition = 15`, `KindHolding = 17`,
  `KindReturn = 20`, `KindDeliveryUnits = 21`, `KindDeliveryResources = 22`,
  `KindMoonDestruction = 14`, `KindAttackDestroyMoon = 25`,
  `KindAttackAllianceDestroyMoon = 27`, `KindStargateTransport = 28`,
  `KindStargateJump = 32`, `KindAlienFlyUnknown = 33`, `KindAlienHolding = 34`,
  `KindAlienAttack = 35`, `KindAlienHalt = 36`, `KindAlienGrabCredit = 37`,
  `KindRepair = 50`, `KindDisassemble = 51`, `KindArtefactExpire = 60`,
  `KindRaidWarning = 64`, `KindExpirePlanet = 65`, `KindScoreRecalcAll = 70`,
  `KindAlienHoldingAI = 80`, `KindAlienChangeMissionAI = 81`.

- **`backend/internal/economy/ids.go`** — `IDMetalmine = 1`, `IDSiliconLab = 2`,
  `IDHydrogenLab = 3`, `IDSolarPlant = 4`, `IDHydrogenPlant = 5`,
  `IDRoboticFactory = 6`, `IDNanoFactory = 7`, `IDShipyard = 8`,
  `IDMetalStorage = 9`, `IDSiliconStorage = 10`, `IDHydrogenStorage = 11`,
  `IDResearchLab = 12`, `IDGravi = 28`, `IDSolarSatellite = 39`,
  `IDDefenseFactory = 100`. Tech: `IDTechGun = 15`, `IDTechShield = 16`,
  `IDTechShell = 17`, `IDTechEnergy = 18`, `IDTechLaser = 23`,
  `IDTechSilicon = 24`, `IDTechHydrogen = 25`, `IDTechBallistics = 103`,
  `IDTechMasking = 104`.

**Проблема**: эти константы **не везде используются** — в SQL-запросах
и Go-коде сохранились магические числа.

---

## 2. Найденные случаи (по аудиту 2026-04-26)

### 2.1 Топ-горячих файлов

| Файл | Кол-во | Содержание |
|---|---:|---|
| `backend/internal/achievement/service.go` | **9** | unit_id и mission в SQL для проверки достижений |
| `backend/internal/fleet/transport.go` | **5+** | mission в `WHERE` clauses |
| `backend/internal/planet/service.go` | 2 | unit_id IN (58, 350) для terraformer/moon_lab |
| `backend/internal/score/event.go` | 1 | `kind = 70` для пересчёта очков |
| `backend/internal/shipyard/service.go` | 1 | `kind = 5` для defense build |
| `backend/internal/alien/alien.go` | 1 | `kind IN (33,34,35,36)` |
| Прочие | ~5 | разные места |

**Итого**: ~25 магических чисел в SQL/Go-коде.

### 2.2 Что заменять (критичные случаи)

#### Event kinds в SQL

| Файл:строка | Было | Станет |
|---|---|---|
| `shipyard/service.go:146` | `kind = 5` | `int(event.KindBuildDefense)` |
| `score/event.go:84` | `WHERE kind = 70` | `WHERE kind = $1` (param) |
| `achievement/service.go:157` | `WHERE kind = 16` | через параметр |
| `fleet/transport.go:562` | `WHERE kind = 7` | `event.KindTransport` |
| `fleet/transport.go:570` | `WHERE kind = 20` | `event.KindReturn` |
| `alien/alien.go:86` | `kind IN (33,34,35,36)` | через параметр или CTE |

#### Unit IDs в SQL

| Файл:строка | Было | Станет |
|---|---|---|
| `achievement/service.go:111` | `unit_id = 1` | `economy.IDMetalmine` |
| `achievement/service.go:117` | `unit_id = 2` | `economy.IDSiliconLab` |
| `achievement/service.go:193` | `unit_id = 3` | `economy.IDHydrogenLab` |
| `achievement/service.go:199` | `unit_id = 4` | `economy.IDSolarPlant` |
| `achievement/service.go:205` | `unit_id = 21` | `economy.IDImpulseEngine` (нужно добавить) |
| `achievement/service.go:211` | `unit_id = 22` | `economy.IDHyperspaceEngine` (нужно добавить) |
| `planet/service.go:134` | `IN (58, 350)` | `IDTerraformer` + `IDMoonLab` (добавить) |

#### Mission в SQL

| Файл:строка | Было | Станет |
|---|---|---|
| `achievement/service.go:130` | `mission = 15` | `event.KindExpedition` |
| `achievement/service.go:148` | `mission = 11` | `event.KindSpy` |
| `achievement/service.go:152` | `mission = 9` | `event.KindRecycling` |
| `achievement/service.go:225` | `mission = 10` | `event.KindAttackSingle` |
| `fleet/transport.go:485` | `mission IN (10, 12)` | `KindAttackSingle, KindAttackAlliance` |
| `fleet/transport.go:632` | `mission NOT IN (15, 29)` | `KindExpedition`, `29` (см. ниже) |
| `fleet/transport.go:664` | `mission = 15` | `event.KindExpedition` |

### 2.3 Лакуны: констант не хватает

| ID | Что | Куда добавить |
|---:|---|---|
| 21 | impulse_engine (research) | `economy/ids.go` → `IDImpulseEngine` |
| 22 | hyperspace_engine (research) | `economy/ids.go` → `IDHyperspaceEngine` |
| 58 | terra_former (building) | `economy/ids.go` → `IDTerraformer` |
| 350 | moon_lab (moon building) | `economy/ids.go` → `IDMoonLab` |
| 29 | mission artefact-delivery? | `event/kinds.go` → проверить, что это за число |

### 2.4 Mode в YAML/Go (опционально)

В `units.yml`/`construction.yml`/`buildings.yml` поле `mode`:
- 1 = building
- 2 = research
- 3 = ship
- 4 = defense
- 5 = moon-only building
- 6 = artefact / special
- 7 = tutorial

В Go-коде сейчас нет констант `economy.ModeBuilding = 1` и т.д. Если
будут конструкции `if spec.Mode == 3` — стоит добавить. Проверить
`ConstructionSpec.Mode` использование.

---

## 3. Стратегия

### Принцип SQL-параметризации

В Go SQL-запросах **не вшивать константу как литерал**, а передавать
через `$1`:

```go
// ❌ Плохо — magic number в SQL литерале
rows, err := tx.Query(ctx, `SELECT ... WHERE kind = 70`)

// ✅ Хорошо — константа Go передана параметром
rows, err := tx.Query(ctx, `SELECT ... WHERE kind = $1`, event.KindScoreRecalcAll)
```

Это даёт:
- Читабельный SQL: `kind = $1` + явный параметр.
- Compile-time safety (если константа переименована — компилятор словит).
- Подготовленные запросы PostgreSQL (минимальная оптимизация).

### Принцип Go struct literals

```go
// ❌ Плохо
event := Event{Kind: 4, ...}

// ✅ Хорошо
event := Event{Kind: int(event.KindBuildFleet), ...}
```

### Не трогаем

- **YAML-конфиги** (`mode: 3` остаётся, это data, не код).
- **`legacy_name: METALMINE`** — историческая ссылка, оставить.
- **Тестовые значения** (`Quantity: 100` в test fixtures) — это not magic.
- **Математические литералы в формулах** (`math.Pow(1.1, level)`).

---

## 4. План внедрения по фазам

### Фаза 1: добавить отсутствующие константы

**Файл**: `backend/internal/economy/ids.go`

Добавить:
- `IDImpulseEngine = 21`
- `IDHyperspaceEngine = 22`
- `IDTerraformer = 58`
- `IDMoonLab = 350`

Проверить, есть ли `event.KindXxx = 29` (mission artefact-delivery). Если
нет — выяснить, что это за mission, добавить в `event/kinds.go`.

**Готовность**: 1 PR, ~10 строк.

### Фаза 2: shipyard, score, alien, fleet — простые случаи

| Файл | Действие |
|---|---|
| `shipyard/service.go:146` | `kind := 4` → `kind := int(event.KindBuildFleet)`, `kind = 5` → `kind = int(event.KindBuildDefense)` |
| `score/event.go:84` | `WHERE kind = 70` → `WHERE kind = $1`, передать `event.KindScoreRecalcAll` |
| `alien/alien.go:86` | `kind IN (33,34,35,36)` → `IN ($1,$2,$3,$4)` с явными константами |
| `fleet/transport.go` | 5 случаев `WHERE mission/kind = N` → параметры |

**Готовность**: 1 PR, ~30 строк правок в 4 файлах.

### Фаза 3: achievement/service.go — большая чистка

Этот файл — главный «нарушитель» (9 magic numbers). Все запросы
переписать с использованием `economy.ID*` и `event.Kind*`.

**Готовность**: 1 PR, ~50 строк правок в одном файле.

### Фаза 4: planet/service.go и оставшееся

Заменить `unit_id IN (58, 350)` на параметры с `economy.IDTerraformer`
и `economy.IDMoonLab`.

**Готовность**: 1 PR, ~10 строк.

### Фаза 5: Mode-enum (опционально)

Если в Go-коде есть `ConstructionSpec.Mode == 3` — добавить:

```go
package economy

type Mode int

const (
    ModeBuilding   Mode = 1
    ModeResearch   Mode = 2
    ModeShip       Mode = 3
    ModeDefense    Mode = 4
    ModeMoon       Mode = 5
    ModeArtefact   Mode = 6
    ModeTutorial   Mode = 7
)
```

Проверить `ConstructionSpec.Mode` использование — если только в loader
(сравнения нет), то Mode-enum не нужен.

**Готовность**: 0-1 PR, ~10 строк (если применимо).

---

## 5. Boy-scout правило (для будущего)

Когда трогаешь файл с magic numbers — **заменяешь на константы**, даже
если задача про другое. Это плавная очистка без больших sweep-PR'ов.

В CLAUDE.md / docs/coding-style.md (если будет) — записать:

> **No magic numbers**: event kinds, unit IDs, tech IDs, HTTP коды
> используют именованные константы из `event/kinds.go`, `economy/ids.go`,
> `net/http`. SQL-запросы передают константы через параметры (`$1`),
> не вшивают как литералы.

---

## 6. Чек-лист

- [x] Ф.1 — добавить недостающие ID в `economy/ids.go`:
       `IDImpulseEngine=21`, `IDHyperspaceEngine=22`, `IDTerraformer=58`,
       `IDMoonLab=350`, `IDCombustionEngine=20`, `IDShipyard` (упорядочен).
- [x] Ф.2 — простые случаи (shipyard, score, alien, fleet).
- [x] Ф.3 — `achievement/service.go` (главный нарушитель). Замечено:
       `STARTER_BUILD_SOLARPLANT/METALLURGY/SHIPYARD/LAB` исторически
       проверяют unit_id 3/4/21/22, что соответствует HydrogenLab/
       SolarPlant/ImpulseEngine/HyperspaceEngine — имена не совпадают
       с реальными ID. Это **отдельный баг логики достижений**, не
       исправлено (выходит за рамки magic-numbers cleanup); сохранено
       оригинальное поведение через явные константы.
- [x] Ф.4 — `planet/service.go` (`IN (58, 350)` → `IN ($2, $3)` через
       `IDTerraformer`/`IDMoonLab`).
- [ ] Ф.5 — Mode-enum (если применимо). **Отложено**: проверить, есть
       ли в Go-коде `spec.Mode == 3` сравнения. Если нет — не нужно.
- [ ] Записать правило в CLAUDE.md / coding-style.

---

## 7. Что НЕ делаем

- **Не меняем** значения констант (1, 2, 3, ...) — это совместимость с БД.
- **Не трогаем** SQL-схемы — ID-числа остаются в БД как есть.
- **Не вводим** новые слои абстракции (Repository, ORM-like) — только
  замена литералов на константы.
- **Не правим** YAML-конфиги (`mode: 3`).
- **Не правим** `legacy_name: METALMINE` — это reference на legacy.

---

## 8. Риски

1. **Опечатки при замене** — компилятор Go словит, тесты подтвердят.
2. **Параметризация SQL может изменить план запроса** — pgx обычно
   prepares одинаково, но проверить EXPLAIN на горячих запросах.
3. **Конфликт с другими PR** — если кто-то параллельно правит SQL в
   `achievement/service.go`, будет merge conflict. Сделать в один заход.

---

## 9. Связь с другими планами

- **План 22 (configs-cleanup)** — закрыт, YAML-схема ОК.
- **План 26 (units-audit-sync)** — закрыт, ID-mapping выверен.
- **План 28 (configs-simplification)** — частично закрыт (Ф.1-4),
  ConstructionCatalog удалён.

Этот план **независим** от других — pure code cleanup.
