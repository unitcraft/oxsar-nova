# План 20: Консистентность конфигов — research.yml, defense.yml, buildings.yml

## Проблема

Несколько независимых несогласованностей в YAML-конфигах.

---

## 20.1 research.yml — устаревший placeholder

`research.yml` содержит base-стоимости, расходящиеся с `construction.yml` в 8–16×:

| Технология | research.yml base | construction.yml basic | Разница |
|------------|-------------------|------------------------|---------|
| ballistics_tech | 4000/8000/4000 | 400/500/0 | ~10–16× |
| masking_tech | 4000/8000/4000 | 400/500/0 (примерно) | ~10× |

`construction.yml` — источник истины (данные из `na_construction` MySQL).
`research.yml` — OGame-placeholder, аналогично `ships.yml` до плана 17.

**Решение:** research.yml хранит только то, чего нет в construction.yml
(factor, cost_factor для формулы). Стоимость (basic) читается из construction.yml.
Аналогично паттерну плана 17 для кораблей.

**Шаги:**
1. Проверить все технологии: сравнить `research.yml` base vs `construction.yml` basic для mode=2.
2. Удалить поле `cost_base` из `ResearchSpec` (или пометить `yaml:"-"`).
3. В `LoadCatalog` заполнять `ResearchSpec.CostBase` из `Construction.Buildings` (аналогично `ShipSpec.Cost`).
4. Найти все вызовы `spec.CostBase` в коде — убедиться что работают без изменений.

Файлы: `backend/internal/config/catalog.go`, `backend/internal/research/service.go`.

---

## 20.2 ion_gun отсутствует в defense.yml

`units.yml` и `construction.yml` содержат `ion_gun` (id=46), но `defense.yml`
не имеет записи — нет `attack/shield/shell` статов.

Параметры из `na_ship_datasheet` (§Таблица na_ship_datasheet в legacy-game-reference.md):

| unitid | attack | shield | front | ballistics | masking |
|--------|--------|--------|-------|------------|---------|
| 46 | 150 | 500 | 10 | 1 | 1 |

`shell` = basic_metal + basic_silicon = 2000 + 6000 = **8000**.

**Шаги:**
1. Добавить в `configs/defense.yml`:
```yaml
ion_gun:
  id: 46
  attack: 150
  shield: 500
  shell: 8000
```

---

## 20.3 max_level для зданий отсутствует

`configs/buildings.yml` не имеет поля `max_level`. Теоретически можно строить
Metal Mine до уровня 999. В legacy (`$GLOBALS["MAX_UNIT_LEVELS"]`) лимиты есть
для отдельных зданий, остальные ограничены `MAX_BUILDING_LEVEL = 40`.

Лимиты из legacy (из `docs/legacy-game-reference.md`):

| Здание | max_level |
|--------|-----------|
| По умолчанию (все здания) | 40 |
| Moon Hydrogen Lab (326) | 10 |
| Nano Factory (7) | 12 |
| Star Gate (56) | 15 |
| Gravi (28) | 10 |

**Шаги:**
1. Добавить поле `MaxLevel int` в `BuildingSpec` (`catalog.go`), `yaml:"max_level,omitempty"`.
2. В `buildings.yml` проставить `max_level: 40` для всех стандартных зданий.
3. Добавить исключения: `nano_factory: 12`, `gravi: 10`.
4. В `building/service.go` (enqueue): проверить `level+1 <= spec.MaxLevel` (если MaxLevel > 0).
5. Аналогично в `construction.yml` — поле `max_level` уже есть? Проверить.

---

## 20.4 exchange/exch_office не обрабатываются building/service.go

`units.yml` включает `exchange` (id=107) и `exch_office` (id=108).
`construction.yml` содержит их данные (mode=1). Но `buildings.yml` не имеет
формул — непонятно как `building/service.go` считает стоимость и время для них.

**Шаги:**
1. Проверить в `building/service.go`: читает ли он стоимость из `Construction.Buildings`
   (тогда OK), или только из `Buildings.Buildings` (тогда exchange не работает).
2. Если только из `Buildings` — добавить fallback на `Construction`.
3. Написать интеграционный тест: попытка поставить в очередь `exchange` не возвращает `ErrUnknownUnit`.

---

## Порядок выполнения

1. **20.2** — самое маленькое, одна строка в defense.yml
2. **20.3** — конфиг + одна проверка в service.go
3. **20.4** — требует проверки кода, минимальные изменения
4. **20.1** — самое крупное, аналог плана 17 для research

## Проверка готовности

- [ ] `ion_gun` в `defense.yml` с корректными статами из na_ship_datasheet
- [ ] `buildings.yml`: все здания имеют `max_level`, лимиты соответствуют legacy
- [ ] `building/service.go`: enqueue отклоняет уровень выше max_level
- [ ] `exchange`/`exch_office` строятся без ошибки (или задокументировано почему нет)
- [ ] `research.yml` стоимости синхронизированы с `construction.yml`
- [ ] `make test` зелёный
