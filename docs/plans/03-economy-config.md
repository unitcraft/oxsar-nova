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

Четыре независимых несогласованности. Порядок: C.1.1 → C.1.2 → C.1.4 → C.1.3.

#### C.1.1 ion_gun отсутствует в defense.yml

`units.yml` и `construction.yml` содержат `ion_gun` (id=46), но `defense.yml` не имеет
боевых статов. Данные из `na_ship_datasheet` (§13.7 balance-analysis.md):

| unitid | attack | shield | shell (броня) |
|--------|--------|--------|---------------|
| 46     | 150    | 500    | 8 000         |

Добавить в `configs/defense.yml`:
```yaml
ion_gun:
  id: 46
  attack: 150
  shield: 500
  shell: 8000
```

Примечание: броня (`shell`) вычисляется из `na_construction` по формуле charge_metal+charge_silicon,
значение 8000 соответствует стоимости 8000/2000/500 (из na_construction для id=46).

#### C.1.2 max_level для зданий

`configs/buildings.yml` не имеет поля `max_level` — можно строить до уровня 999.
Лимиты из legacy `consts.php` (`MAX_BUILDING_LEVEL = 40`, `$GLOBALS["MAX_UNIT_LEVELS"]`):

| Здание | max_level | Источник |
|--------|-----------|---------|
| По умолчанию (все) | 40 | `MAX_BUILDING_LEVEL = 40` |
| moon_hydrogen_lab (326) | 10 | `MAX_UNIT_LEVELS` |
| moon_repair_factory | 9 | `MAX_UNIT_LEVELS` |
| moon_lab | 5 | `MAX_UNIT_LEVELS` |
| nano_factory (7) | 12 | `MAX_UNIT_LEVELS` |
| star_gate (56) | 15 | `MAX_UNIT_LEVELS` |
| gravi (28) | 10 | `MAX_UNIT_LEVELS` |

Шаги:
1. Добавить `MaxLevel int` в `BuildingSpec` (`catalog.go`), `yaml:"max_level,omitempty"`
2. В `buildings.yml` проставить `max_level: 40` для всех зданий + исключения выше
3. В `building/service.go` (enqueue): `level+1 <= spec.MaxLevel` (если MaxLevel > 0)

#### C.1.3 research.yml — устаревшие placeholder-стоимости

`research.yml` содержит base-стоимости, расходящиеся с `construction.yml` в 8–16×.
Пример: `ballistics_tech` — research.yml: 4000/8000/4000, construction.yml: 400/500/0.
`construction.yml` — источник истины. Аналогично паттерну плана 17 для кораблей:

1. Убрать поле `cost_base` из `ResearchSpec` (или `yaml:"-"`)
2. В `LoadCatalog` заполнять `ResearchSpec.CostBase` из `Construction.Buildings` (mode=2)
3. Проверить все вызовы `spec.CostBase` в `research/service.go`

Также: `MAX_RESEARCH_LEVEL = 40` из legacy consts.php — применить аналогично C.1.2
для исследований если поле отсутствует.

#### C.1.4 exchange/exch_office в building/service.go

`units.yml` включает `exchange` (id=107) и `exch_office` (id=108), `construction.yml` содержит их (mode=1),
но `buildings.yml` не имеет записей — неясно как считается стоимость/время.

1. Проверить читает ли `building/service.go` стоимость из `Construction.Buildings` (тогда OK)
2. Если только из `Buildings` — добавить fallback
3. Интеграционный тест: enqueue `exchange` не возвращает `ErrUnknownUnit`

**Параметры биржи из legacy consts.php** (для будущей реализации):
- `EXCH_LEVEL_SLOTS = 15` слотов на уровень биржи
- `EXCH_MERCHANT_COMMISSION = 13%` (без премиума), `10%` (с премиумом)
- `EXCH_MAX_TTL = 7 дней`, `EXCH_MIN_TTL = 3 дня`
- `EXCH_SELLER_MAX_PROFIT = 1000%` — макс. наценка

**Проверка готовности C.1:**
- [ ] `ion_gun` в `defense.yml` с корректными статами
- [ ] `buildings.yml`: все здания имеют `max_level`
- [ ] `building/service.go`: enqueue отклоняет уровень выше max_level
- [ ] `exchange`/`exch_office` строятся без ошибки
- [ ] `research.yml` стоимости синхронизированы с `construction.yml`
- [ ] `make test` зелёный

---

### ✅ C.2 Параметры игрового инстанса (план 18, итерация 18)
- `config.go`: StorageFactor, ResearchSpeedFactor, EnergyProductionFactor, MaxPlanets,
  BashingPeriod, BashingMaxAttacks, ProtectionPeriod — все читаются из ENV
- `planet/service.go`: NewServiceWithFactors(storageFactor, energyProductionFactor)
- `research/service.go`: NewServiceWithFactors(researchSpeedFactor)
- `fleet/colonize.go`: MaxPlanets проверяется при колонизации
- `fleet/attack.go`: BashingMaxAttacks / ProtectionPeriod проверяются при атаке

---

### ✅ C.3 Система профессий (план 19, итерация 19)
- `migrations/0046_profession.sql` — колонки `profession`, `profession_changed_at` в `users`
- `configs/professions.yml` — 4 профессии: miner/attacker/defender/tank
- `ProfessionCatalog` / `ProfessionSpec` в `config/catalog.go`, загрузка из professions.yml
- `economy/ids.go`: `ProfessionKeyToID` — карта строковых ключей → unit ID
- `profession/service.go`: `Change` (1000 кр, интервал 14 дней), `Get`, `List`, `BonusFromKey`
- `profession/handler.go`: `GET /api/professions`, `GET /api/professions/me`, `POST /api/professions/me`
- `planet/service.go`: бонусы профессии применяются к tech-карте при расчёте производства
- `fleet/attack.go`, `fleet/acs_attack.go`: бонусы профессии применяются к `battle.Tech`

---

### ✅ C.4 Рефакторинг formula DSL → статические функции (план 16)
`backend/pkg/formula/` удалён. `grep -r "pkg/formula" backend/` — нет совпадений.
