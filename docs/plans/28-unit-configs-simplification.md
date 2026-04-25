---
title: 28 — Симплификация конфигов юнитов
date: 2026-04-25
status: in-progress
---

> **Статус 2026-04-25**:
> - Ф.1-4 (backend) — **✅ done**: construction.yml удалён, поля
>   мигрированы в ships.yml/defense.yml/research.yml/buildings.yml,
>   ConstructionCatalog/ConstructionSpec удалены из Go,
>   battle-sim читает front из ShipSpec, regression-тесты идентичны.
> - Ф.5 (frontend через `/api/catalog`) — pending.
> - Ф.6-7 (dead-поля + валидатор) — pending.

# План 28: Симплификация конфигов юнитов

**Цель**: устранить дубли и разрозненность параметров юнитов между
`configs/*.yml`, Go-loader (`backend/internal/config/`) и frontend
(`frontend/src/api/catalog.ts`). **Без изменения логики и игрового
поведения** — только архитектурная очистка.

**Не путать с**:
- План 22 (configs-cleanup) — уже сделал YAML-cleanup (orphans, dead
  fields, валидатор).
- План 26 (units-audit-sync) — уже сделал сверку слоёв по ID/имени.
- План 27 (unit-rebalance-deep) — баланс (числа), не структура.

**Этот план — про структуру: где хранить, как читать, как избавиться
от дублей.**

**Важный факт (2026-04-25)**: формулы производства/потребления/
стоимости **уже живут в коде** в [backend/internal/economy/formulas.go](../../backend/internal/economy/formulas.go).
Поля `prod`/`cons`/`charge` в `construction.yml` — **dead-fields**
(grep подтверждает: ни одно использование `.Prod.`/`.Cons.`/`.Charge.`
в Go-коде). Комментарий в `catalog.go:39-40` явно говорит: «Поля
prod/cons/charge хранят формулы как строки — только для документации;
runtime-расчёт идёт через статические функции economy». Это
**радикально упрощает** план: после миграции `construction.yml`
**полностью удаляется**, не остаётся «формулы зданий».

---

## 1. Текущая картина (по аудиту 2026-04-25)

### 1.1 Где живут параметры юнита (на примере Battleship, id=34)

| Параметр | Файл / место | Формат |
|---|---|---|
| `attack`, `shield`, `shell` | `configs/ships.yml` | int |
| `cargo`, `speed`, `fuel` | `configs/ships.yml` | int |
| `cost` (metal/silicon/hydrogen) | `configs/construction.yml` (basic) | int |
| `front`, `ballistics`, `masking` | `configs/construction.yml` | int (per-unit) |
| `mode`, `legacy_name`, `display_order` | `configs/construction.yml` | meta |
| `requirements` | `configs/requirements.yml` | list |
| `wiki description` | `configs/wiki-descriptions.yml` | text |
| `name`, `key`, `id` | `configs/units.yml` | meta |
| **Frontend копия всех боевых статов** | `frontend/src/api/catalog.ts` SHIPS[] | TS |
| **Frontend копия cost** | `frontend/src/api/catalog.ts` SHIPS[] | TS |
| **Frontend копия rapidfire** | `frontend/src/api/catalog.ts` SHIPS[].rapidfire | TS |
| `rapidfire` (backend) | `configs/rapidfire.yml` | int×int → int |

### 1.2 Дубли (короткое резюме аудита)

| Параметр | Источник истины (intended) | Реально хранится в |
|---|---|---|
| attack/shield/shell (ship) | `ships.yml` | ships.yml + catalog.ts |
| attack/shield/shell (defense) | `defense.yml` | defense.yml + catalog.ts |
| cost (ship) | `construction.yml` | construction.yml + catalog.ts |
| cost (defense) | `defense.yml` (НЕ construction!) | defense.yml + catalog.ts |
| front/ballistics/masking | `construction.yml` | construction.yml + catalog.ts |
| rapidfire | `rapidfire.yml` | rapidfire.yml + catalog.ts |
| requirements | `requirements.yml` | requirements.yml + catalog.ts |

**Главная проблема**: `frontend/src/api/catalog.ts` — **полная копия
всех данных** в TS-литералах, без связи с YAML. При изменении
`ships.yml` или `construction.yml` фронтенд **молча рассинхронизируется**.

### 1.3 Дополнительные перекосы

1. **Cost кораблей** — из `construction.yml` (через `LoadCatalog`
   подтягивает в `ShipSpec.Cost`).
2. **Cost обороны** — из `defense.yml` напрямую (НЕ из construction).
   Несимметрично.
3. **front/ballistics/masking** — Go читает из `construction.yml`, но
   в `ShipSpec`/`DefenseSpec` этих полей **нет** — они читаются
   отдельным запросом `cat.Construction.Buildings[key]`. Потребители
   (battle-sim, движок через Input) должны знать о двух источниках.
4. **per-unit `ballistics`/`masking` — мёртвые поля** (см. план 27 §16):
   движок их игнорирует, но `construction.yml` хранит.
5. **`legacy_name` в construction.yml** — не используется в коде, только
   для документации. Безопасно удалить (но требует решения).

### 1.4 Комментарии-долги в коде

- `backend/internal/config/catalog.go:122-124` — «Cost заполняется из
  Construction.Buildings, в ships.yml отсутствует — источник истины
  construction.yml».
- `frontend/src/api/catalog.ts:2` — `// TODO: сгенерировать из YAML на
  этапе gen:api`.
- `backend/internal/config/catalog.go:26-33` — «ConstructionCatalog
  предпочтительнее BuildingCatalog».

---

## 2. Принципы симплификации

1. **Один параметр — один источник истины.** Если параметр есть в
   `construction.yml`, то его не должно быть в `ships.yml`/`defense.yml`/
   `catalog.ts`.
2. **Frontend не дублирует balance-данные** — получает их через API
   (gen-from-OpenAPI или dedicated endpoint).
3. **Без изменения логики** — каждое перемещение параметра
   подкрепляется тестом «было → стало», runtime-поведение идентично.
4. **Слои данных явно разделены**:
   - **Identity** (id, key, name, mode) — `units.yml`.
   - **Balance** (attack, shield, shell, cost, front, ...) —
     `ships.yml` / `defense.yml` / `buildings.yml` / `research.yml`.
   - **Mechanics** (rapidfire, requirements) — собственные файлы.
   - **Presentation** (icons, wiki text) — `wiki-descriptions.yml` +
     icons в `frontend/public/`.
5. **YAML — единственный человеко-редактируемый источник.** Frontend
   и backend читают из него (через API).

---

## 3. Целевая архитектура

```
configs/
├── units.yml              # IDENTITY: id, key, name, mode
├── ships.yml              # BALANCE: attack, shield, shell, cargo, speed, fuel, cost, front, ballistics, masking
├── defense.yml            # BALANCE: то же для обороны
├── buildings.yml          # BALANCE: cost_base, cost_factor, time, energy, demolish, display_order, charge_credit
├── research.yml           # BALANCE: cost_base, cost_factor, demolish (если есть)
├── rapidfire.yml          # MECHANICS: shooter × target → multiplier
├── requirements.yml       # MECHANICS: dependencies
├── wiki-descriptions.yml  # PRESENTATION: тексты для wiki
├── artefacts.yml          # IDENTITY+BALANCE: артефакты
└── professions.yml        # BALANCE: профессии

# construction.yml — УДАЛЁН. Все поля разъехались по другим файлам.
# Формулы prod/cons/charge — в backend/internal/economy/formulas.go.

backend/internal/config/
├── catalog.go             # Loader, **БЕЗ** logic копирования cost из construction
├── catalog_validate_test.go  # Расширенные проверки

api/openapi.yaml
└── /api/catalog endpoint  # Возвращает units+balance в JSON для frontend

frontend/src/api/
└── catalog.ts             # ТОЛЬКО icons, key→imageMap. Balance из API.
```

### 3.1 Изменения в YAML

**ships.yml** — расширить, чтобы стал полным источником истины
для кораблей:

```yaml
ships:
  battle_ship:
    id: 34
    # combat
    attack: 1000
    shield: 200
    shell: 60000
    front: 10
    ballistics: 0      # default, но явно
    masking: 0
    # logistics
    cargo: 1500
    speed: 10000
    fuel: 500
    # economy
    cost:
      metal: 40000
      silicon: 20000
      hydrogen: 0
```

**defense.yml** — то же:

```yaml
defense:
  rocket_launcher:
    id: 43
    attack: 80
    shield: 20
    shell: 2000
    front: 10
    cost:
      metal: 2000
      silicon: 0
      hydrogen: 0
```

**construction.yml** — **полностью удалить** после миграции:
- `id`/`mode`/`legacy_name` — дубль с `units.yml`, остаётся в `units.yml`.
- `front`/`ballistics`/`masking` — переехали в `ships.yml`/`defense.yml` (Ф.1).
- `basic` (cost) — переехал в `ships.yml`/`defense.yml` (Ф.1).
- `display_order`/`demolish`/`charge_credit` — переехали в
  `buildings.yml`/`research.yml`.
- `prod`/`cons`/`charge` (строки-формулы) — **dead-документация**,
  удаляются. Если нужен «текстовый референс» — вынести в
  `docs/balance/formulas-reference.md` (опционально).

### 3.2 Изменения в Go-loader

`catalog.go`:
- `ShipSpec` получает поля `Front`, `Ballistics`, `Masking`, `Cost`
  (cost больше не yaml:"-", читается из ships.yml).
- Удалить блок «Заполнить ShipSpec.Cost из Construction.Buildings» (lines 278+).
- `ConstructionSpec` — оставить только для зданий/исследований
  (formulas).
- Аналогично `DefenseSpec` (front уже близок, добавить Front).

### 3.3 Изменения во frontend

- Добавить endpoint `GET /api/catalog` (или сгенерировать из OpenAPI).
- `frontend/src/api/catalog.ts` оставляет:
  - `KEY_MAP` для иконок.
  - `imageOf` / `imageOfId`.
  - Хелперы для UI.
- Удалить hardcoded `BUILDINGS`, `MOON_BUILDINGS`, `RESEARCH`, `SHIPS`,
  `DEFENSE`, `ARTEFACTS` массивы — заменить на TanStack Query из
  `/api/catalog`.

### 3.4 OpenAPI endpoint

Новый ресурс в `api/openapi.yaml`:

```yaml
paths:
  /api/catalog:
    get:
      summary: Полный catalog юнитов и балансных параметров
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Catalog'

components:
  schemas:
    Catalog:
      type: object
      properties:
        ships: { type: array, items: { $ref: '#/components/schemas/ShipSpec' } }
        defense: { type: array, items: { $ref: '#/components/schemas/DefenseSpec' } }
        buildings: { ... }
        research: { ... }
        rapidfire: { ... }
        requirements: { ... }
```

Эндпоинт чисто read-only, безопасно кэшируется на CDN/edge.

---

## 4. План внедрения по фазам

### Фаза 1 — Расширение ships.yml / defense.yml (без поломок)

**Не удалять старые данные**, только дублировать вперёд для безопасности.

1. Добавить в `ships.yml` поля `front`, `ballistics`, `masking`, `cost`
   (значения копируются из текущего `construction.yml`).
2. Добавить в `defense.yml` поле `front`.
3. Загрузчик читает **из ships.yml** (приоритет), при отсутствии
   падает на construction.yml (legacy fallback).
4. Тесты: golden-comparison `cat.Ships.Ships[k]` до/после = идентичны.

**Изменения в коде**:
- `backend/internal/config/catalog.go`: расширить `ShipSpec`/`DefenseSpec`,
  добавить fallback-логику.
- `backend/internal/config/catalog_validate_test.go`: проверка, что
  значения в ships.yml совпадают с construction.yml (для миграции).

**Готовность**: 1 PR, ~150 строк diff.

### Фаза 2 — Battle-sim и движок переходят на ships.yml

1. `battle-sim/main.go` `constructionMeta()` → `shipMeta()` (читает
   front из ships.yml).
2. Любые другие потребители front/ballistics/masking — на ships.yml.
3. Тесты: `battle-sim --all` даёт **идентичные** результаты до и после.

**Готовность**: 1 PR, ~80 строк.

### Фаза 3 — Перенести оставшиеся поля из construction.yml

Поля, которые **используются** в коде, но всё ещё живут в construction.yml:

1. **`display_order`** — для UI-сортировки. Перенести в
   `buildings.yml`/`research.yml` (для обороны/кораблей — в
   соответствующие файлы).
2. **`demolish`** (множитель сноса) — перенести в `buildings.yml`.
3. **`charge_credit`** (для exch_office/exchange) — перенести в
   `buildings.yml`.
4. Удалить `basic` (cost) из `construction.yml` — уже в ships.yml/defense.yml
   после Ф.1.
5. Удалить `front`/`ballistics`/`masking` — уже в ships.yml/defense.yml.
6. **Удалить `prod`/`cons`/`charge`** (строки-документация) —
   они никем не читаются (grep подтверждает). Если нужен референс —
   опционально вынести в `docs/balance/formulas-reference.md`.

После Ф.3 в `construction.yml` остаётся только `id`/`mode`/`legacy_name`,
которые дублируют `units.yml`.

### Фаза 4 — Удаление construction.yml

1. Перенести `legacy_name` (если ещё нужен для legacy-сравнения)
   в `units.yml` как опциональное поле.
2. Удалить `configs/construction.yml`.
3. Удалить `ConstructionCatalog`/`ConstructionSpec`/`ConstructionBasic`/
   `ConstructionFormulas` из `backend/internal/config/catalog.go`.
4. Обновить `LoadCatalog` — убрать чтение `construction.yml` и блоки
   копирования cost (lines 278+).
5. Обновить тесты в `catalog_validate_test.go`.

**Готовность**: 1 PR, удаление ~400 строк (YAML+Go).

### Фаза 5 — Frontend через API

1. Создать backend handler `GET /api/catalog` в `backend/internal/catalog/`
   (новый домен) или extension к существующему.
2. Сгенерировать TS-клиент из OpenAPI.
3. Заменить hardcoded `SHIPS`, `DEFENSE`, `BUILDINGS`, `RESEARCH`
   в `frontend/src/api/catalog.ts` на TanStack Query.
4. Оставить `KEY_MAP` и `imageOf` (presentation).
5. e2e-тесты UI: симулятор боя, экраны fleet/buildings/research должны
   показывать те же значения, что и раньше.

**Готовность**: 2-3 PR (backend handler + TS-генерация + UI-замены).
~400-600 строк.

### Фаза 6 — Удаление мёртвых полей

1. `per-unit ballistics`/`masking` в movement (план 27 §16): движок
   игнорирует. Удалить из ships.yml/defense.yml/construction.yml
   ИЛИ включить в движок (план 27 ADR-V).
2. `mode=7` псевдо-юниты (LOGIN, FIRST_*, FIVE_ATTACKER) — это tutorial-
   маркеры. Перенести из construction.yml в отдельный
   `configs/tutorial-flags.yml` или удалить, если не используются.

**Готовность**: 1 PR, ~50 строк.

### Фаза 7 — Валидатор согласованности (CI)

Расширить `backend/internal/config/catalog_validate_test.go`:
- Каждый id в `ships.yml` имеет id в `units.yml`.
- Каждый id в `defense.yml` имеет id в `units.yml`.
- Каждый ship/defense с `cost` (metal+silicon+hydrogen > 0) имеет
  `attack` или `shield > 0` (не «бесплатный пустой юнит»).
- Каждый id в `rapidfire.yml` (shooter+target) присутствует в
  ships.yml/defense.yml.
- Frontend OpenAPI-схема покрывает все поля бэкенда.

**Готовность**: 1 PR, ~100 строк тестов.

---

## 5. Что НЕ делаем (out of scope)

- **Не меняем числа** — план 27 отдельно.
- **Не реализуем новые юниты** — orphans остаются по ADR-0006.
- **Не меняем формулу боя** — план 18/21 закрыты.
- **Не вводим новые YAML-файлы** (кроме опционального
  `tutorial-flags.yml` в Ф.6) — текущий список достаточен.
- **Не трогаем legacy schema** (`oxsar2/sql/`) — только nova.
- **Не пишем миграционных скриптов** — это правки YAML в репо.
- **Не оптимизируем загрузку** — кэширование catalog уже работает.

---

## 6. Риски

1. **Frontend десинк во время миграции** — если backend выкатил
   `/api/catalog` раньше, чем frontend переключился. Mitigation:
   feature-flag, временно держать оба источника.
2. **Тесты battle/golden зависят от construction.yml-loader-логики** —
   если есть assert на cost из ConstructionSpec, нужно обновить.
3. **Legacy migration tools** — если в `backend/cmd/tools/` есть
   утилита, читающая construction.yml для legacy-сравнения, она
   ломается. Проверить и обновить.
4. **YAML lint в CI** — расширение схемы ships.yml требует
   обновления валидаторов (если есть JSON-Schema в `.golangci.yml`
   или `pre-commit`).
5. **Размер `/api/catalog`** — может быть большим (все юниты + балансы).
   ~50 KB JSON. Не критично.

---

## 7. Предполагаемая экономия

**Строки YAML** (после Ф.3): −200 (убираем basic+front+ballistics+
masking из construction.yml для кораблей/defense).

**Строки TS** (после Ф.5): −600..−800 (массивы SHIPS/DEFENSE/
BUILDINGS/RESEARCH в catalog.ts).

**Когнитивная нагрузка**: разработчик хочет изменить attack у
Battleship — открывает `ships.yml`, меняет одну строку, всё. Сейчас:
открывает `ships.yml`, потом помнит, что cost в `construction.yml`,
потом помнит, что front тоже в construction, потом проверяет, не
скопирован ли стат во frontend `catalog.ts`.

**Riski десинка** между backend и frontend → 0 (после Ф.5).

---

## 8. Чек-лист прогресса

- [x] Ф.1 — Расширение ships.yml/defense.yml/research.yml/buildings.yml
       (cost, front, ballistics, masking, cost_base, display_order,
       demolish, charge_credit). Расширены ShipSpec/DefenseSpec/
       ResearchSpec/BuildingSpec в Go.
- [x] Ф.2 — Battle-sim читает front из ShipSpec; флаг `--front=key=N`
       работает на ships/defense.
- [x] Ф.3 — поля display_order/demolish/charge_credit/basic/front/
       ballistics/masking перенесены в целевые YAML (на этапе Ф.1
       идемпотентно). prod/cons/charge были удалены раньше.
- [x] Ф.4 — `construction.yml` удалён, `ConstructionCatalog` /
       `ConstructionSpec` / `ConstructionBasic` удалены из Go.
       `productionRatesApprox` (fallback) удалён, `getBuildingName`
       переключён на `cat.Units` (UnitsCatalog). `convert_construction.go`
       удалён, `writeYAMLSorted` переехал в `helpers.go`.
- [ ] Ф.5.1 — Backend handler `/api/catalog`.
- [ ] Ф.5.2 — OpenAPI + gen TS-клиент.
- [ ] Ф.5.3 — Замена hardcoded в catalog.ts на TanStack Query.
- [ ] Ф.6 — Удаление dead-полей (per-unit ballistics/masking — план 27
       решает отдельно).
- [ ] Ф.7 — Валидатор согласованности + CI-тесты.

---

## 9. Связь с другими планами

- **22 (configs-cleanup)** — предшественник, сделал YAML-cleanup
  (orphans, валидатор Ф.3). Этот план — продолжение архитектурное.
- **26 (units-audit-sync)** — сверка слоёв по ID/имени. Готова
  Ф.1, остатки в Ф.2/Ф.3 этого плана покрываются здесь.
- **27 (unit-rebalance-deep)** — баланс (числа). Идёт параллельно,
  не конфликтует.
- **27 §16** — ADR-U (удалить per-unit ballistics/masking) учитывает
  Ф.6 этого плана.

---

## 10. Открытые вопросы

1. **Defense cost — переехать в construction.yml для симметрии с
   ships, или оставить в defense.yml?** Текущее решение: **оставить
   в defense.yml** (cost — балансный параметр, ему место рядом с
   attack/shield).
2. **`legacy_name` в construction.yml** — оставить (для wiki/legacy-
   reference) или удалить? Решение: **оставить**, мало байт.
3. **Single-file vs multi-file YAML** — может, всё в один
   `units-balance.yml`? Решение: **multi-file** (разделение
   ship/defense/building/research удобнее для PR-review).
   Причины подробнее (рассмотрено 2026-04-25):
   - Single-file ~4500+ строк (150 юнитов × 30 полей) — сложнее
     навигировать.
   - Разные роли правщиков: балансер (cifры) vs гейм-дизайнер (wiki-
     текст) → разные жизненные циклы → отдельные файлы.
   - **Rapidfire по природе матричный** (shooter × target). Per-unit
     формат дублирует записи; плоская таблица в `rapidfire.yml`
     читается как матрица.
   - **Requirements** — граф зависимостей, читается как граф, не как
     поле юнита.
   - Merge-конфликты при параллельных балансовых правках разносятся
     между файлами.
   - **Единственная полезная консолидация — удаление
     `construction.yml`** (Ф.3-4 этого плана). После неё не остаётся
     дублей между файлами; оставшаяся multi-file структура корректна.
4. **Когда делать?** Перед или после плана 27? Рекомендация:
   **после 27** (баланс не должен ехать поверх архитектурных правок).
   Если 27 будет тянуться долго, можно делать Ф.1-2 параллельно
   (они без поломок).
