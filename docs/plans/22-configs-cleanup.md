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

### Ф.2.2 Мёртвые legacy-юниты — ✅ РЕШЕНО 2026-04-25 ([ADR-0006](../adr/0006-orphan-units-deferred.md))

**Не реализуем в v1.0**. Все 7 orphan-юнитов остаются в `units.yml` для
совместимости с legacy-дампом, без игрового эффекта. `knownOrphans`
теперь = roadmap для v1.x.

Главные мотивы:
- `planet_shield` пара — самые ценные (H), но требуют ad-hoc shield-pool
  несовместимый с battle-engine. Скоп L, не M.
- `moon_lab + ign` — после реализации IGR-tech (ADR-0005) частично
  избыточны.
- `moon_hydrogen_lab + lancer_ship` — won't fix (никогда не имели
  effect'а в legacy).

После релиза: priority 1 = planet_shield (если симуляция/playtesting
потребует), priority 2 = moon_lab.



Данные из legacy PHP (oxsar2/www + oxsar2/www/ext) и БД, поиск
2026-04-24. Детали в `knownOrphans` (backend/internal/config/
catalog_validate_test.go).

Список кандидатов на реализацию с оценкой «сложность/ценность»:

| Юнит | Legacy-эффект | Сложность | Ценность |
|------|--------------|-----------|----------|
| `terra_former` (58) | +5 полей / уровень к max_fields планеты | **M** — ждёт базового field-limit | M — расширяет late-game (больше слотов под здания) |
| `small_planet_shield` (354) | +10 HP shield pool / штука, 4 слота | M (новое поле planet.shield_pool + интеграция с battle-engine) | H — классический late-game щит |
| `large_planet_shield` (355) | +40 HP shield pool / штука, 8 слотов | то же, что small | H |
| `moon_lab` (350) | альтернативная research-лаба на луне + 5 полей | M (формула в research.Service) | M |
| `moon_repair_factory` (351) | лунный repair factory | M (copy-paste логики repair под луну) | L — мало игроков с лунами в M1 |
| `ign` (26) | virtual lab network: +lab'ы других планет | L (изменяет research-тайминг, нужны тесты баланса) | M — крупная механика альянсов |
| `gravi` (28) | weapon scale в бою, `getGraviWeaponScale()` | L (правка battle-engine, рискованно для баланса) | L — легко обойти оставив requirement без effect |
| `moon_hydrogen_lab` (326) | **effect не реализован даже в legacy** | — | — |
| `lancer_ship` (102) | **effect не реализован даже в legacy** | — | у нас в AlienAI отдельно |

Сложность:
- **S** ≤ 2 часа, одиночный файл + миграция
- **M** 2-6 часов, требуется интеграция с сервисом/engine + тесты
- **L** > 6 часов, затрагивает баланс / битвы, нужны golden-тесты

**Открытие при попытке реализовать `terra_former`**: в backend'е
**нет проверки лимита полей при постройке**. `diameter` и `used_fields`
есть в модели, показываются в UI, но `building.Service.StartBuild`
не блокирует постройку при превышении. То есть terra_former сейчас
бесполезен — даже без него можно строить бесконечно. Сначала нужен
базовый field-limit, потом terra_former имеет смысл.

**Рекомендация по очередности (если решим реализовать):**

0. **(prerequisite для #1)** field-limit в building.Service — блокировать
   постройку, если `used_fields >= max_fields(diameter)`. Отдельный
   маленький план.
1. `terra_former` — S после prerequisite, ценность M.
2. `planet_shield` пара — M/H. Добавляет late-game вариативность обороны, ожидаемая механика из OGame-like.
3. `moon_lab` — M/M. Люди уже строят луны, но лаба на луне сейчас «мёртвая».
4. `moon_repair_factory` — M/L. По желанию после луны.
5. `ign` — L/M. Сложно и меняет research-тайминг всей партии, осторожно.
6. `gravi` — L/L. Или эффект, или просто оставить как требование для DS.

**Что можно сразу выкинуть из units.yml** (нулевой эффект даже в legacy):
- `moon_hydrogen_lab` (326)
- `lancer_ship` (102) — если мы не используем в AlienAI как визуальный ship

Но удаление из `units.yml` потребует `import-datasheets` prepatch
(он при импорте воссоздаст записи из legacy-дампа). Проще
оставить в реестре + закомментировать в `knownOrphans`.

**Оригинальное описание (для истории):**


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
