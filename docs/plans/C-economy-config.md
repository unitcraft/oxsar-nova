# План C: Экономика и конфиги

---

## История (завершено)

### ✅ construction.yml — единый источник истины для стоимостей (план 16-статика, итерация 16)
- `configs/construction.yml` сгенерирован из `na_construction.sql`
- DSL-формулы производства (`evalProd`/`evalCons`) заменены на статические Go-функции
  в `backend/internal/economy/formulas.go`
- Удалён `backend/pkg/formula/` (DSL-движок)
- `planet/service.go`: убраны `formulaCache`, `evalProd`, `evalCons`

### ✅ Единый источник для стоимостей кораблей (план 17, итерация ~50)
- `ShipSpec.Cost yaml:"-"` — заполняется в `LoadCatalog` из `Construction.Buildings`
- Удалено дублирование `cost:` из `configs/ships.yml`
- Добавлены отсутствующие корабли: frigate(35), recycler(37), solar_satellite(39),
  bomber(40), star_destroyer(41), lancer_ship(102), shadow_ship(325),
  unit_a_corvette(200), unit_a_screen(201), unit_a_paladin(202),
  unit_a_frigate(203), unit_a_torpedocarier(204)

### ✅ Стоимости в buildings/research — корректный CostForLevel (итерация ~30)
- `building/service.go`, `research/service.go`: используют `Construction.Buildings`
- `CostForLevel(base, factor, level)` — статическая формула

### ✅ Реальное время постройки по уровням зданий (план 13, итерация ~40)
- `construction.yml` содержит `time_factor` для каждого здания
- Время строительства учитывает robotic_factory + nano_factory уровни
- `BuildTime(base, level, roboLvl, nanoLvl, gameSpeed)` в formulas.go

---

## Открытые задачи

### C.1 Конфиг-консистентность (план 20, приоритет: HIGH)

Четыре независимых несогласованности.

#### C.1.1 ion_gun отсутствует в defense.yml (самое маленькое)

`units.yml` и `construction.yml` содержат `ion_gun` (id=46), но `defense.yml` не имеет
боевых статов. Данные из `na_ship_datasheet`:

| unitid | attack | shield | shell |
|--------|--------|--------|-------|
| 46     | 150    | 500    | 8000  |

Добавить в `configs/defense.yml`:
```yaml
ion_gun:
  id: 46
  attack: 150
  shield: 500
  shell: 8000
```

#### C.1.2 max_level для зданий

`configs/buildings.yml` не имеет поля `max_level` — можно строить до уровня 999.
Legacy лимиты (`$GLOBALS["MAX_UNIT_LEVELS"]`):

| Здание | max_level |
|--------|-----------|
| По умолчанию (все) | 40 |
| moon_hydrogen_lab (326) | 10 |
| nano_factory (7) | 12 |
| star_gate (56) | 15 |
| gravi (28) | 10 |

Шаги:
1. Добавить `MaxLevel int` в `BuildingSpec` (`catalog.go`), `yaml:"max_level,omitempty"`
2. В `buildings.yml` проставить `max_level: 40` для всех зданий + исключения
3. В `building/service.go` (enqueue): `level+1 <= spec.MaxLevel` (если MaxLevel > 0)

#### C.1.3 research.yml — устаревшие placeholder-стоимости

`research.yml` содержит base-стоимости расходящиеся с `construction.yml` в 8–16×.
`construction.yml` — источник истины. Аналогично паттерну плана 17 для кораблей:

1. Убрать поле `cost_base` из `ResearchSpec` (или `yaml:"-"`)
2. В `LoadCatalog` заполнять `ResearchSpec.CostBase` из `Construction.Buildings` (mode=2)
3. Проверить все вызовы `spec.CostBase` в `research/service.go`

#### C.1.4 exchange/exch_office в building/service.go

`units.yml` включает `exchange` (id=107) и `exch_office` (id=108), `construction.yml` содержит их (mode=1),
но `buildings.yml` не имеет записей — неясно как считается стоимость/время.

1. Проверить читает ли `building/service.go` стоимость из `Construction.Buildings` (тогда OK)
2. Если только из `Buildings` — добавить fallback
3. Интеграционный тест: enqueue `exchange` не возвращает `ErrUnknownUnit`

**Порядок:** C.1.1 → C.1.2 → C.1.4 → C.1.3 (по возрастанию сложности)

**Проверка готовности:**
- [ ] `ion_gun` в `defense.yml` с корректными статами
- [ ] `buildings.yml`: все здания имеют `max_level`
- [ ] `building/service.go`: enqueue отклоняет уровень выше max_level
- [ ] `exchange`/`exch_office` строятся без ошибки
- [ ] `research.yml` стоимости синхронизированы с `construction.yml`
- [ ] `make test` зелёный

---

### C.2 Рефакторинг formula DSL → статические функции (план 16, приоритет: LOW)

> **Статус:** Частично сделано в итерации 16 (prod/cons функции).
> Проверить что пакет `backend/pkg/formula/` не используется за пределами import-datasheets.

Если `grep -r "pkg/formula" backend/` возвращает hits вне `cmd/tools/import-datasheets` — удалить.

**Проверка:**
- [ ] `grep -r "pkg/formula" backend/` — только import-datasheets
- [ ] Удалить `backend/pkg/formula/` если больше не нужен
