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

### ✅ C.1 Конфиг-консистентность (итерации ~48)
- C.1.1: `ion_gun` добавлен в `defense.yml` (attack:150, shield:500, shell:8000)
- C.1.2: `buildings.yml` — все здания имеют `max_level` (дефолт 40, исключения: nano_factory=12, star_gate=15); `building/service.go` отклоняет уровень выше max_level
- C.1.3: `research.yml` — `cost_base yaml:"-"`, заполняется из `construction.yml` в LoadCatalog
- C.1.4: `buildings.yml` дополнен `exchange` (107) и `exch_office` (108)

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
