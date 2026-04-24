# План 22: Очистка и согласование configs/

**Цель:** привести YAML-справочники в `configs/` в согласованное состояние.
Реестр `units.yml` — источник истины по ID, все балансные файлы
(`buildings.yml`, `research.yml`, `ships.yml`, `defense.yml`,
`requirements.yml`) должны с ним совпадать. Убрать мёртвый код и дубли.

**Scope:** только YAML-файлы и связанный аудит. Не трогаем движок, формулы
боя и runtime-логику. Изменения баланса — отдельный план 18.

**Связанное:**
- [Первоначальный аудит](../configs-audit.md) — что грузится, где дыры
- [План 18](./18-unit-rebalance.md) — балансные правки (пересекаются по файлам)
- [План 20](./20-legacy-port.md) — порт из легаси (если реализуем дыры, а не вычеркиваем)

---

## Статус

**Ф.1 (быстрые фиксы)** — ✅ коммит `badf588` (2026-04-24).
**Ф.2.1 (defense_factory)** — ✅ коммит `8b60f19`.
**Ф.2.2 (planetary shields)** — ⏳ ADR отложен.
**Ф.2.3 (ракеты в defense-группе)** — ✅ решено «оставить как есть» (см. ниже).
**Ф.3 (валидатор)** — ✅ этот коммит. 4 теста, нашёл ещё 2 бага (rocket_station/missile_silo рассинхрон, terra_former/moon_* orphans).

---

## Ф.1 Быстрые фиксы (HIGH) — ✅ done

Данные взяты из legacy MySQL (`na_construction`, `na_requirements`),
применены 1:1 за редкими исключениями (комментарии в файлах).

- [x] Удалены `*_generated.yml` (3 файла, 12 КБ) — не читаются
      runtime'ом. Добавлены в `.gitignore`.
- [x] `research.yml`: добавлены `ign` (id=26, cost_factor=2.0)
      и `gravi` (id=28, cost_factor=3.0). Последний уже требовался
      `death_star`, но `requirements.Check` молча возвращал nil.
- [x] `requirements.yml`: добавлены записи для `star_destroyer` (41),
      `shadow_ship` (325), `ion_gun` (46). До этого юниты строились
      без проверки требований — фактический баг.

**Известные расхождения с легаси в Ф.1:**
- `ion_gun` в легаси требует `defense_factory level 4`. У нас этого здания
  нет в `buildings.yml` (хотя id=101 зарегистрирован в `units.yml`).
  Временно заменено на `shipyard level 4`. См. Ф.2.1.

---

## Ф.2 ADR-решения (MEDIUM) — нужен геймдизайн

### Ф.2.1 `defense_factory` (id=101) — ✅ done

Добавлен в `buildings.yml` с параметрами из legacy `na_construction`:
cost (350, 200, 100), factor 2.0, time_base_seconds 180, max_level 40.
В `requirements.yml` `ion_gun` теперь ссылается на `defense_factory 4`
(было shipyard 4 как временная замена в Ф.1). UI в BuildingsScreen
подхватит автоматически через BuildingCatalog.

**Оригинальное описание фазы (сохранено для контекста):**


- Объявлен в `units.yml buildings`, но отсутствует в `buildings.yml`
  (нет статов/формул, ни в `construction.yml` — запись есть, но не
  используется runtime'ом).
- Используется как requirement для `ion_gun`, `planet_shield` в легаси.

**Варианты:**
- (A) Реализовать как здание аналогично `shipyard`: формулы
  производства/cost из `construction.yml` (запись id=101 уже есть),
  UI-карточку в BuildingsScreen. Стоимость постройки defense-юнитов
  снизится с повышением уровня. Это правильная legacy-механика.
- (B) Удалить из `units.yml` и навсегда заменить требования
  на `shipyard`. Теряем legacy-механику ускорения defense-построек.

**Рекомендация:** (A). Это здание — часть исходного баланса.

### Ф.2.2 Planetary shields (354, 355)

- В `units.yml defense` объявлены `small_planet_shield`, `large_planet_shield`.
- Легаси-статы из БД: small `shell=300k, shield=100`; large
  `shell=1M, shield=100k` (очень высокие параметры).
- Требования: shield_tech 5+ (small), shield_tech 10 + gun_tech 10 +
  ballistics 10 (large).
- **Нет реализации:** battle-engine не знает о планетарных щитах,
  UI-карточки нет, защитный эффект на планете не применяется.

**Варианты:**
- (A) Реализовать как особый defense-юнит с уникальным эффектом
  (поглощение входящего урона перед ablation). Нужен backend
  (новый параметр planet.shield_pool), UI (карточка в defense), ADR
  про механику.
- (B) Удалить из `units.yml`. Теряем legacy-механику; балансно не
  обязательная, игра без них жива.

**Рекомендация:** (B) на ближайший релиз, (A) когда понадобится
late-game сдерживание. Legacy планетарные щиты делают DS-раши
менее болезненными — это классическая late-game механика OGame.

### Ф.2.3 Ракеты в `units.yml defense`-группе — ✅ решено «оставить»

Проверил использование `UnitsCatalog` в коде: **никем не используется**
на runtime'e. `DefenseCatalog` грузится из `defense.yml` напрямую, там
ракет нет. `units.yml` — только реестр для `import-datasheets` и
ментальной модели.

Перенос 51/52 в fleet-группу потребует правок CLI импортёра и ломает
публичный контракт реестра. Компромисс: оставить в defense-группе с
комментарием, что это legacy-совместимость. В `configs/units.yml`
добавлен поясняющий комментарий (commit `[текущий]`).

**Оригинальное описание:**


- Записи `interceptor_rocket` (51), `interplanetary_rocket` (52)
  зарегистрированы как `defense` в `units.yml`.
- Фактически работают как `fleet` через `ships.yml` + kind=16
  (interplanetary rocket impact event из плана 05).
- В `requirements.yml` уже есть требования (shipyard 1 + rocket_station 2/4).

**Проблема:** `units.yml` как реестр врёт про категорию. Поиск по
категории `defense` в UI найдёт ракеты, которые в defense-очередь
не попадают.

**Вариант:** перенести 51/52 в `units.yml fleet`-группу. Это меняет
публичную семантику реестра — возможно, ломает API-ответы/frontend.

**Требует:** проверить где `units.yml` отдаётся в API и как используется
категория «defense»/«fleet» во frontend (BuildingsScreen/ShipyardScreen).

---

## Ф.3 Валидатор конфигов — ✅ done

Файл `backend/internal/config/catalog_validate_test.go`, 4 теста:

- [x] `TestValidate_AllUnitsHaveBalance` — каждый key из `units.yml`
      есть в балансном файле (кроме `knownOrphans` со списком известных
      пробелов с обоснованием).
- [x] `TestValidate_RequirementsReferenceExistingUnits` — все `key` в
      requirements ссылаются на реальные юниты; `kind` только building
      или research.
- [x] `TestValidate_RapidfireReferenceExistingShips` — from/to в
      rapidfire.yml — существующие id.
- [x] `TestValidate_NoDuplicateIDs` — нет дублей id между группами
      `units.yml`.

**Находки при первом запуске** (закрыты в этом коммите):
1. `rocket_station` в `units.yml`, но в `buildings.yml` записано как
   `missile_silo` — **реальный рассинхрон ключа**. Переименовано в
   buildings.yml к `rocket_station`; старое имя осталось алиасом во
   frontend.
2. Orphan-юниты (`terra_former`, `moon_hydrogen_lab`, `moon_lab`,
   `moon_repair_factory`, `lancer_ship`) — добавлены в `knownOrphans`
   с обоснованием. Реализация — отдельные планы, не Ф.3.

`knownOrphans` — это белый список сознательных пропусков. Удаление
записи оттуда (без добавления реализации) сразу ломает тест — защита
от регрессии.

---

## Что НЕ делать в этом плане

- **Не менять баланс.** Цифры берутся 1:1 из легаси или оставляются
  как есть. Балансные правки — план 18.
- **Не удалять `construction.yml`** (24 КБ). Это источник истины legacy
  na_construction-формул, `ConstructionCatalog` — fallback/reference
  (§1.4 ТЗ).
- **Не переносить balance-данные из `construction.yml` в `buildings.yml`
  и т.д. автоматически.** Сначала решить судьбу orphan-id (см. Ф.2),
  иначе генератор затянет их как пустышки.

---

## Порядок реализации

1. **Ф.1** — ✅ сделано (коммит `badf588`).
2. **Ф.2.1** defense_factory — прочитать `construction.yml` запись id=101,
   добавить в `buildings.yml`, интегрировать в BuildingsScreen. ~60 мин.
3. **Ф.2.2** planetary shields — ADR: делаем или нет. Если да — новая
   механика в battle-engine, ~2-3 часа + тесты.
4. **Ф.2.3** категории ракет — проверить использование `units.yml`
   группировки в API/UI, перенести 51/52 в fleet. ~30 мин.
5. **Ф.3** валидатор — 150 LOC теста, ~45 мин.
