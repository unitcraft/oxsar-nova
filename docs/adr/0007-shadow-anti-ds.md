# ADR-0007: Shadow Ship — anti-DS-роль (отклонение от legacy)

- Status: Accepted
- Date: 2026-04-25
- План: [docs/plans/27-unit-rebalance-deep.md](../plans/27-unit-rebalance-deep.md) блок 27-B

## Context

В аудите 2026-04-25 (план 27) Shadow Ship (id=325) определён как
**декоративный юнит**: attack=30 при стоимости 5k metal-eq, masking=4.
Шпионаж лучше делать через Espionage Sensor (probe), а низкий attack
делает Shadow бесполезным в бою.

При этом в оригинальной legacy у Shadow есть rapidfire против всего
крупного флота (LT×15, LF×5, Frigate×20, Bmb×25, SD×15, **DS×70**),
но при attack=30 эти rapidfire срабатывают слишком слабо.

Endgame oxsar-nova сейчас имеет **только один путь** — массовое DS-
производство. Контр через BS+SD mix есть, но требует тех же ресурсов
что DS. Альтернатива через stealth-флот была бы интересной для PvP.

## Decision

Изменить параметры Shadow Ship:

| Параметр | Было | Стало |
|---|---:|---:|
| attack | 30 | **200** |
| masking | 4 | 5 (уже было в construction.yml) |
| ballistics | 3 | 5 (уже было в construction.yml) |
| front | 9 | 7 (уже было в construction.yml) |
| shield | 30 | 30 |
| shell | 4000 | 4000 |
| cost | 1k/3k/1k | без изменений |

Файлы:
- `configs/ships.yml` — `attack: 30 → 200`
- `frontend/src/api/catalog.ts` — синхронизация (там были устаревшие
  значения cost 8k/4k/1.5k вместо 1k/3k/1k, masking=4 вместо 5).

## Consequences

### Геймплей

- **Stealth-мета** — 50 Shadow (250k metal-eq) образуют невидимый
  ударный кулак. Masking=5 даёт ~83% уклонения от ballistics=0.
- **Anti-DS контр** — Shadow×70 vs DS теперь работает: 100 Shadow
  (500k metal-eq) с rapidfire×70 = 7000 виртуальных выстрелов за
  раунд. Альтернатива BS+SD-mix.
- **Шпионаж + удар** — один тип юнита для двух ролей. Игроку проще
  планировать.

### Симуляция

После применения 27-B запустить:
```bash
cd backend && go run ./cmd/tools/battle-sim -scenario shadow-vs-ds -runs 50 -configs ../configs
```

Целевой exchange ratio: 1.5–3.0 для Shadow-fleet vs Death Star.
Сценарий нужно добавить в `cmd/tools/battle-sim/main.go` (отдельная
итерация).

### Плюсы

- Открытие альтернативного endgame-пути.
- Shadow получает функциональную роль — не декорация.
- Rapidfire-таблица Shadow становится осмысленной.
- Legacy-rapidfire не меняется — только attack-стат.

### Минусы

- **Отклонение от legacy-баланса** — attack 30 был исходным значением.
  Меняем сознательно.
- Тесты `battle/golden` нужно перегенерировать (если есть с участием Shadow).
- Существующие игроки с большим Shadow-флотом получат «бесплатный buff».
  Актуально для dev-seed; продакшн-запуска ещё не было.

### Альтернативы отвергнуты

1. **Поднять цену Shadow до 5k+15k+5k metal-eq** при сохранении attack=30 —
   не решает проблемы, Shadow остаётся декорацией.
2. **Добавить новые rapidfire-записи Shadow → ...** — все нужные
   записи уже есть в legacy-таблице (план 18 портировал).
3. **Удалить Shadow из игры** — роняет идентичность с legacy и
   убирает потенциал для future-stealth-мета.

## Lancer-cost (27-A) — отдельно

В плане 27 был блок 27-A (Lancer cost ×1.4). После повторной симуляции
2026-04-25 на текущем балансе:

| Сценарий | Exchange |
|---|---:|
| Lancer vs Cruiser | 0.05 (Cruiser рвёт Lancer) |
| Lancer vs mixed | 0.21 (атакующий теряет ресурсы) |

Lancer уже невыгоден атакующему (ADR-0004 закрыл BA-002). Дополнительный
нерф не нужен.

## Откат

Изменить в `configs/ships.yml` `attack: 30` обратно. Frontend
`catalog.ts` синхронизировать.
