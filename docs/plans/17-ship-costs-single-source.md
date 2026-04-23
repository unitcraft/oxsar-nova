# План 17: Единый источник истины для стоимостей кораблей

## Проблема

Существуют два конфига с данными кораблей:

- `configs/ships.yml` — содержит `cost` (стоимость), `attack`, `shield`, `shell`, `cargo`, `speed`, `fuel`
- `configs/construction.yml` — содержит `basic` (стоимость) для кораблей (mode=3), плюс стоимости зданий/исследований/обороны

Оба содержат `cost`/`basic` для кораблей, и они **расходятся**. `construction.yml` является первичным источником (данные из `na_construction` MySQL), тогда как `ships.yml` — устаревший placeholder на основе OGame.

## Конкретные расхождения

### Конфликт значений (оба файла содержат корабль, но разные цифры)

| Корабль | ships.yml (cost) | construction.yml (basic) | Правильное (legacy DB) |
|---------|-----------------|--------------------------|------------------------|
| `battle_ship` (34) | metal=45000, silicon=15000, H=0 | metal=40000, silicon=20000 | metal=40000, silicon=20000 |

### Корабли только в construction.yml (отсутствуют в ships.yml)

| Корабль | construction.yml basic |
|---------|------------------------|
| `frigate` (35) | metal=30000, silicon=40000, H=15000 |
| `recycler` (37) | metal=10000, silicon=6000, H=2000 |
| `solar_satellite` (39) | silicon=2000, H=500 |
| `bomber` (40) | metal=50000, silicon=25000, H=15000 |
| `star_destroyer` (41) | metal=60000, silicon=50000, H=15000 |

Эти корабли полностью отсутствуют в `ships.yml`, поэтому их боевые параметры (attack/shield/shell/cargo/speed/fuel) тоже не заданы.

### Боевые параметры отсутствующих кораблей (из na_ship_datasheet)

| unitid | Корабль | cargo | speed | fuel | attack | shield | front | ballistics | masking |
|--------|---------|-------|-------|------|--------|--------|-------|------------|---------|
| 35 | Frigate | 750 | 10 000 | 250 | 700 | 400 | 10 | 0 | 0 |
| 37 | Recycler | 20 000 | 2 000 | 300 | 1 | 10 | 10 | 0 | 0 |
| 39 | Solar Satellite | 0 | 0 | 0 | 0 | 2 | 10 | 0 | 0 |
| 40 | Bomber | 500 | 4 000 | 1 000 | 900 | 550 | 10 | 0 | 0 |
| 41 | Star Destroyer | 2 000 | 5 000 | 1 000 | 2 000 | 500 | 10 | 0 | 0 |

`shell` (броня) вычисляется по формуле: `(basic_metal + basic_silicon) / коэффициент`. Точный коэффициент — из legacy-кода (`na_construction.charge_*`). Принять `shell = metal + silicon` (стандарт OGame/oxsar) до уточнения по legacy.

## Решение

**Единственный источник стоимостей кораблей — `configs/construction.yml`** (поле `basic`, mode=3). `configs/ships.yml` хранит только боевые и ходовые параметры (attack/shield/shell/cargo/speed/fuel) и не содержит `cost`.

Чтение стоимости корабля в Go-коде: `configs.GetConstruction(shipKey).Basic`.

## Шаги реализации

### Шаг 1 — Удалить `cost` из ships.yml

Убрать поле `cost` у всех кораблей в `configs/ships.yml`. Стоимость теперь читается только из `construction.yml`.

### Шаг 2 — Добавить недостающие корабли в ships.yml

Добавить записи с боевыми параметрами (без cost) для кораблей, которых нет в `ships.yml`:

```yaml
frigate:
  id: 35
  attack: 700
  shield: 400
  shell: 70000      # (30000+40000) — броня = сумма стоимостей
  cargo: 750
  speed: 10000
  fuel: 250

recycler:
  id: 37
  attack: 1
  shield: 10
  shell: 16000      # (10000+6000)
  cargo: 20000
  speed: 2000
  fuel: 300

solar_satellite:
  id: 39
  attack: 0
  shield: 2
  shell: 2000       # (0+2000)
  cargo: 0
  speed: 0
  fuel: 0

bomber:
  id: 40
  attack: 900
  shield: 550
  shell: 75000      # (50000+25000)
  cargo: 500
  speed: 4000
  fuel: 1000

star_destroyer:
  id: 41
  attack: 2000
  shield: 500
  shell: 110000     # (60000+50000)
  cargo: 2000
  speed: 5000
  fuel: 1000
```

### Шаг 3 — Исправить battle_ship в ships.yml

`battle_ship.cost` уже удалён (шаг 1). Боевые параметры остаются без изменений — они совпадают с legacy DB.

### Шаг 4 — Обновить загрузчик конфигов

Найти код, который читает `cost` из `ships.yml` (если есть) и переориентировать на `construction.yml`.

Места для поиска:
- `backend/internal/*/config.go` или аналог
- любые вызовы типа `ship.Cost`, `cfg.Ships[k].Cost`

Новый паттерн:

```go
// Стоимость корабля
cost := constructionCfg.Buildings[shipKey].Basic

// Боевые параметры корабля
stats := shipsCfg.Ships[shipKey]
```

### Шаг 5 — Обновить тесты и golden-файлы

Проверить тесты на предмет захардкоженных стоимостей кораблей. Обновить при необходимости.

### Шаг 6 — Проверить rapidfire для новых кораблей

`frigate`, `bomber`, `star_destroyer` имеют rapidfire в legacy (см. `docs/legacy-game-reference.md`, раздел `na_rapidfire`). Убедиться что `configs/rapidfire.yml` содержит их записи или создать задачу на отдельный план.

## Что НЕ меняется

- Формулы `charge_*` в `construction.yml` остаются как есть
- `shell` в `ships.yml` — расчётное значение, не из БД (формула legacy — броня = сумма металл+кремний; уточнить при необходимости в отдельном ADR)
- Боевые параметры существующих кораблей в `ships.yml` не трогаем — они совпадают с legacy DB

## Проверка готовности

- [x] `configs/ships.yml` не содержит поля `cost`
- [x] `configs/ships.yml` содержит все корабли из `construction.yml` (mode=3, ids: 29-42, 52, 102, 325; unit_exch_support_range id=105 — не боевой, пропущен намеренно)
- [x] Код читает стоимость строго из `construction.yml`
- [x] `make test` зелёный
- [x] `make lint` зелёный
