# План 26: Аудит юнитов — сверка с легаси и синхронизация слоёв

**Дата**: 2026-04-25
**Статус**: Ф.1 — ✅ done. Ф.2 — выявлен список к решению (см. ниже).

## Контекст

На скриншоте wiki юниты #35, #52, #102 показаны числами вместо имён —
сигнал, что `frontend/src/api/catalog.ts` рассинхронизирован с
`configs/units.yml` и balance YAML (`ships.yml`, `defense.yml`,
`research.yml`, `buildings.yml`).

Цель плана — единая сверка всех слоёв определения юнитов:

1. **Legacy** — `oxsar2/www/new_game/protected/config/consts.php` (UNIT_*)
   и `oxsar2/www/ext/` (override-логика, не определения).
2. **`configs/units.yml`** — мастер-список ID + ключей nova.
3. **Balance YAML** — `buildings.yml` / `ships.yml` / `defense.yml` /
   `research.yml` (параметры и формулы).
4. **`configs/wiki-descriptions.yml`** — описания для wiki-gen.
5. **`configs/requirements.yml`** — требования.
6. **`frontend/src/api/catalog.ts`** — runtime UI (имена, иконки,
   характеристики для симулятора).
7. **`docs/wiki/ru/`** — генерируемые wiki-страницы.

## Принцип

oxsar-nova — **не строгий 1:1 порт legacy**. Часть юнитов сознательно
не реализована (см. [ADR-0006](../adr/0006-orphan-units-deferred.md)).
Главное правило: **в каждом слое юнит должен либо иметь полные данные,
либо отсутствовать целиком**. Потерянных юнитов с частичными
параметрами быть не должно.

---

## Ф.1: Сверка catalog.ts с balance YAML — ✅ done 2026-04-25

### Что было пропущено

В catalog.ts отсутствовали:

| ID  | Ключ                  | Тип       | Статус legacy | Реализован в nova |
|-----|-----------------------|-----------|---------------|--------------------|
| 26  | `ign`                 | research  | ✅            | ✅ (требование DS) |
| 28  | `gravi`               | research  | ✅            | ✅ (требование DS) |
| 35  | `frigate`             | ship      | ✅            | ✅                 |
| 41  | `star_destroyer`      | ship      | ✅            | ✅                 |
| 46  | `ion_gun`             | defense   | ✅            | ✅                 |
| 52  | `interplanetary_missile` | ship   | ✅ (id=52)    | ✅ (rocket, kind=16)|
| 102 | `lancer_ship`         | ship      | ✅            | ✅                 |
| 112 | `astro_tech`          | research  | ✅            | ✅                 |
| 113 | `igr_tech`            | research  | ✅            | ✅                 |
| 200-204 | alien-юниты      | ship      | ✅            | ✅ (AI-only)       |
| 325 | `shadow_ship`         | ship      | ✅            | ✅                 |

### Симптом

`nameOf(35)` → `#35` вместо `Фрегат`. Wiki-сайдбар показывал числовые ID,
ссылки `[[unit:N]]` рендерились как `unit N`. UnitInfoScreen / симулятор
не находили описаний и параметров.

### Фикс

Добавлены в catalog.ts:
- `RESEARCH`: 26, 28, 112, 113
- `SHIPS`: 35, 41, 52, 102, 200-204, 325
- `DEFENSE`: 46

`units.yml` — добавлены alien-корабли (200-204) в секцию `fleet:`,
теперь все 6 слоёв согласованы.

---

## Ф.2: Decoupled-юниты — нужны ADR

Юниты, объявленные в legacy, но **намеренно не реализованные** в nova.
Часть уже зафиксирована в [ADR-0006](../adr/0006-orphan-units-deferred.md):

| ID  | Ключ                       | Legacy | Решение                    |
|-----|----------------------------|--------|----------------------------|
| 51  | `interceptor_rocket`       | ✅     | ⏳ stub (только references) |
| 105 | `exch_support_range`       | ✅     | ⏳ ADR — не реализуем      |
| 106 | `exch_support_slot`        | ✅     | ⏳ ADR — не реализуем      |
| 110 | `alien_tech`               | ✅     | ⏳ ADR — не реализуем      |
| 111 | `artefacts_tech`           | ✅     | ⏳ ADR — заменено системой artefacts_market |
| 326 | `moon_hydrogen_lab`        | ✅     | ⏳ план 22 Ф.2.2           |
| 350 | `moon_lab`                 | ✅     | ⏳ план 22 Ф.2.2           |
| 351 | `moon_repair_factory`      | ✅     | ⏳ план 22 Ф.2.2           |
| 352 | `ship_transplantator`      | ✅     | ⏳ ADR — не реализуем      |
| 353 | `ship_collector`           | ✅     | ⏳ ADR — не реализуем      |
| 354 | `small_planet_shield`      | ✅     | ⏳ план 22 Ф.2.2 (planetary shields) |
| 355 | `large_planet_shield`      | ✅     | ⏳ план 22 Ф.2.2           |
| 358 | `ship_armored_terran`      | ✅     | ⏳ ADR — не реализуем      |

### Действие

Эти юниты **не должны** появляться:
- В `catalog.ts` (frontend не должен знать о них)
- В `wiki-descriptions.yml` (нет описания)
- В `docs/wiki/ru/` (нет страницы)

Если юнит присутствует только в `units.yml` — это допустимо, но требует
комментария в YAML. ADR-0006 уже фиксирует knownOrphans в коде.

---

## Ф.3: Валидатор согласованности — ⏳ todo

Нужен CI-тест, который при сборке проверяет:

1. Каждый ID из `ships.yml`/`defense.yml`/`buildings.yml`/`research.yml`
   присутствует в `units.yml`.
2. Каждый ID из `units.yml` (кроме knownOrphans) имеет балансные параметры.
3. Каждый ID из `units.yml` (кроме knownOrphans) имеет имя в catalog.ts.
4. Каждый ID из catalog.ts (кроме артефактов) имеет описание в
   `wiki-descriptions.yml` или wiki-страницу.

Существующий `backend/internal/config/catalog_validate_test.go`
покрывает (1) и частично (2). Расширить под frontend и wiki.

---

## Связанные планы

- [22 Ф.2.2](22-configs-cleanup.md) — planetary shields (354, 355)
  и moon_* (326, 350, 351). Требует ADR.
- [ADR-0006](../adr/0006-orphan-units-deferred.md) — решение
  по orphan-юнитам.
- [18 Фаза 1](18-unit-rebalance.md) — балансные правки (rapidfire,
  Lancer cost), не пересекается напрямую.

---

## Что НЕ делаем

- Не реализуем юниты из Ф.2 без ADR.
- Не меняем балансные параметры существующих юнитов (это план 18).
- Не трогаем legacy ext/ — только источник для сверки.
