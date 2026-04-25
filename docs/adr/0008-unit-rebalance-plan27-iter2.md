# ADR-0008: Глубокая ребалансировка юнитов (план 27, итерация 2)

- Status: Accepted
- Date: 2026-04-25
- План: [docs/plans/27-unit-rebalance-deep.md](../plans/27-unit-rebalance-deep.md)
- Заменяет: ADR-0007 (Shadow attack 30→200) — параметр обновлён до 520
  (см. 27-F).

## Context

После итерации 1 плана 27 (применение ADR-0007 для Shadow) были найдены
системные перекосы (см. план 27 §10):

1. **Shadow ADR-0007 не работал** — attack=200 ≤ DS ignoreAttack=500
   (= shield/100), выстрелы Shadow поглощались щитом DS.
2. **Bomber-роль не работала** на endgame-обороне — нет RF против
   Gauss/Plasma, exchange 0.64 vs Plasma.
3. **Strong Fighter — trap-юнит** без ниши.
4. **Defense слаба** для endgame — 3% потерь у атакующего vs 36% у
   защитника.
5. **SSat-эксплойт** реален: 50k SSat = ловушка для DS.
6. **Front-тюнинг** не использовался для Lancer/Shadow/Recycler/Probe/
   Shields/SD.
7. **Per-unit ballistics/masking** — мёртвые поля в YAML, движок их не
   читает.

Паритет с legacy oxsar2/Java JAR больше **не цель** (см. memory
`feedback_no_legacy_parity.md`), поэтому изменения принимаются по
симуляции.

## Decision

Применены 9 групп изменений (subtags 27-F..27-U):

### 27-F/V — Shadow Ship rebalance

| Параметр | Было | 27-F | 27-V (финал) |
|---|---:|---:|---:|
| attack | 200 | **520** | 520 |
| front | 7 | **5** | 5 |
| cost | 5k (1+3+1) | 5k | **15k (3+9+3)** |

`configs/ships.yml`.

**27-F**: attack 200→520. Причина: 520 > DS ignoreAttack=500 → активирует
RF×70 vs DS. front=5 усиливает stealth-эффект в смешанной атаке
(weight = 2^5 = 32 vs 2^10 = 1024 у обычного юнита).

**27-V**: cost 5k → 15k (×3). После прогона матрицы 1v1 при равной
metal-eq (план 27 §17) Shadow доминировал чрезмерно: vs LF **90.91**,
vs Frigate **104.68**, vs Bomber **83.25**, vs SD **51.28**. Причина:
Shadow с cost=5k был самым дешёвым юнитом с attack=520 — на 10M
metal-eq получалось 2000 Shadow-юнитов, что просто переламывало любой
мирный флот. cost ×3 снижает «эффективную массу» в 3×, anti-DS-роль
сохраняется (нужно меньше юнитов, но они дороже).

**Эффект 27-V на матрицу 1v1**:

| Shadow vs | До 27-V | После 27-V |
|---|---:|---:|
| LF | 90.91 | **7.57** |
| Frigate | 104.68 | **30.14** |
| Bomber | 83.25 | **27.75** |
| SD | 51.28 | **11.90** |
| DS | 6.67 | **2.22** (целевой коридор 1.5-3.0) |
| Defense (RL/LL/SL/IG) | 1.25-1.94 | **0.07-0.13** |
| Plasma | 3.71 | 0.27 |

Shadow стал **специализированным**: anti-fleet/anti-DS, проигрывает
defense — больше не «универсальная wunderwaffe».

### 27-G — Bomber RF vs Gauss/Plasma

`configs/rapidfire.yml`:
- Bomber → Gauss Gun (47): RF×3
- Bomber → Plasma Gun (48): RF×3

Восстанавливает anti-defense-роль Bomber на endgame.

### 27-I — Strong Fighter rebalance

| Параметр | Было | Стало |
|---|---:|---:|
| attack | 150 | **120** |
| cost | 6k+4k = 10k | **4k+2k = 6k** |

`configs/ships.yml`. Причина: SF был «trap-юнитом». Теперь дешёвый
anti-LF (sf-vs-lf exchange 1.57 → 2.05).

### 27-J/W — Lancer rebalance (двухходовка)

| Параметр | Было | 27-J | 27-W (финал) |
|---|---:|---:|---:|
| **attack** | 5500 | 4000 | **5000** |
| fuel | 100 | **400** | 400 |
| speed | 8000 | **6000** | 6000 |
| front | 8 | **6** | 6 |
| **max_per_planet** | — | — | **50** (game-механика) |

`configs/ships.yml` + `backend/internal/shipyard/service.go` (cap-проверка).

**Эволюция решения**:
- **27-J (промежуточно)**: attack 5500→4000 + fuel/speed nerf + front
  8→6. После прогона матрицы 1v1 (план 27 §17) выяснилось, что
  attack=4000 делает Lancer **trap-юнитом**: vs LF 0.04, vs SF 0.02,
  vs Cru 0.04 — проигрывает всем 1v1.
- **27-W (финал)**: attack 4000→**5000** возвращает Lancer боеспособность
  (vs LF 0.04, vs Cru 0.05, vs Frigate 0.66 — Lancer всё ещё слабый
  одиночный юнит, но не trap). При attack=5000 Lancer-spam vs lite
  возвращается (lancer-vs-mixed exchange 0.19, attacker_wins 100%) —
  поэтому добавлен **cap=50 Lancer/планета** (game-механика).

**Cap=50 Lancer/планета** реализован через:
- Поле `max_per_planet: 50` в `configs/ships.yml`.
- Поле `MaxPerPlanet` в `config.ShipSpec`.
- Проверка в `shipyard.Service.Enqueue`: считает `existing + in_queue + new ≤ cap`.
- Новая ошибка `ErrPlanetCapExceeded` → HTTP 400.

С cap=50 максимальный Lancer-флот собирается с 10 планет (500 Lancer),
для атаки на дальнюю цель — fuel-bottleneck (400/Lancer + speed=6000).
Lancer-spam как raid-инструмент **экономически невозможен**.

**Trade-off**: Lancer-как-anti-DS-mass ослаблен (300 Lancer vs 1 DS:
def-clean — Lancer теряет всё). Anti-DS-роль теперь основная у
Shadow (27-F), Lancer **в комбо** с другими юнитами.

### 27-K — Defense shell ×1.5

`configs/defense.yml`. Все defense-юниты получают shell ×1.5:
- RL/LL: 2000 → 3000
- SL/IG: 8000 → 12000
- Gauss: 35000 → 52000
- Plasma: 100000 → 150000
- Small Shield: 20000 → 30000
- Large Shield: 100000 → 150000

Цена не меняется. Защитник получает чистый buff.

### 27-L — Solar Satellite shell+cost

| Параметр | Было | Стало |
|---|---:|---:|
| shell | 2000 | **5000** |
| cost (Si) | 2000 | **3000** |
| cost (H) | 500 | **1000** |

`configs/ships.yml`. Закрывает SSat-эксплойт: 50k SSat теперь стоят
200M (вместо 125M) и при потере 37.5% защитник теряет 75M — дороже,
чем raw value DS-атаки.

### 27-N..T — Front-тюнинг

| Юнит | Front было | Стало | Тег |
|---|---:|---:|---|
| Lancer | 8 | **6** | 27-N |
| Bomber | 10 | **8** | 27-O |
| Shadow | 7 | **5** | 27-P |
| Small Transporter | 10 | **6** | 27-Q |
| Large Transporter | 10 | **6** | 27-Q |
| Recycler | 10 | **6** | 27-Q |
| Espionage Sensor | 10 | **5** | 27-R |
| Small Shield | 16 | **13** | 27-S |
| Large Shield | 17 | **14** | 27-S |
| Star Destroyer | 10 | **9** | 27-T |

Семантика: логистика (transports/recycler/probe) больше не «принимает
огонь», капитал-юниты (Lancer/Bomber/Shadow) скрываются за более
многочисленным флотом, shields умеренно приоритетны.

### 27-U — Удаление per-unit ballistics/masking

`configs/ships.yml`, `configs/defense.yml`, `backend/internal/config/catalog.go`.

Удалены поля `ballistics`/`masking` per-unit (DS=4, Lancer=3, Shadow=5/5,
Paladin=5/1, Corvette=2, Frigate(alien)=1, Torpedo=4, Interplanetary=10).
Движок никогда их не читал — использует только `Side.Tech.Ballistics/Masking`
(уровни research). Удалены поля `Ballistics`/`Masking` из `ShipSpec` и
`DefenseSpec`.

### Отвергнуто: 27-H (DS ignoreAttack /150)

Слишком инвазивно — меняет порог пробивания для всех юнитов глобально.
Anti-DS-путь открыт через 27-F (Shadow). Дополнительный канал через
Cruiser сейчас не нужен.

## Consequences

### Симуляция (battle-sim --runs=20, до → после)

**Ключевые улучшения**:

| Сценарий | Before | After | Изменение |
|---|---:|---:|---:|
| shadow-mass-vs-ds | 0.00 | **6.67** | +6.67 |
| shadow-bs-mix-vs-cruiser | 2.39 | **4.60** | +2.21 |
| bomber-vs-gauss | 1.13 | **3.84** | +2.71 |
| bomber-vs-plasma | 0.64 | **1.44** | +0.80 |
| lancer-bs-combo-vs-mixed | 0.22 | **0.79** | +0.57 |
| sf-vs-lf | 1.57 | **2.05** | +0.48 |
| recycler-bs-mix | 1.70 | 1.93 | +0.23 |

**Defense buff** (атаковать стало дороже):

| Сценарий | Before | After |
|---|---:|---:|
| bomber-vs-rl | 16.67 | 11.11 |
| cruiser-vs-rl | 8.62 | 4.57 |
| bs-vs-plasma | 1.81 | 1.17 |
| lf-mass-vs-rl-ll | 0.32 | 0.21 |
| large-shield-coverage | 5.20 | 3.46 |
| plasma-wall-vs-bs-mass | 2.43 | 1.53 |

**Mirror-сценарии** (без изменений, как ожидалось):
- huge-fleet-vs-huge-fleet: 1.00 → 1.00
- ds-fleet-vs-ds-fleet: 1.00 → 1.00
- frigate-vs-cruiser: 3.27 → 3.27
- frigate-vs-bs: 8.21 → 8.21

### Известные ограничения (не починено)

- `ds-fleet-vs-defended-planet`: 3.51 → 3.70 — defense buff (27-K)
  не сильно помог против DS-флота. DS остаётся endgame-доминирующим
  юнитом. Радикальный фикс (27-H, ignoreAttack /150) отвергнут.
- `ds-as-defense-vs-fleet`: 0.00 → 0.00 — DS-как-защитник всё ещё
  неуязвим для BS+SD. Контр — только Shadow-mass или Lancer (27-F).
- `lancer-vs-mixed`: 0.21 → 0.21 — Lancer-spam vs lite не починен
  (fuel/speed не влияют на бой, только на миссии).
- `shadow-vs-ds` (100 Shadow): 0.00 — мало массы, нужны 1000+ для
  пробивания DS.

### Тесты

- `go test ./... -count=1`: все 26 пакетов зелёные.
- battle-sim 35+ сценариев: побитово стабильны при повторных прогонах
  (детерминированный seed).

### Frontend

`frontend/src/api/catalog.ts` имеет hardcoded SHIPS/DEFENSE с
устаревшими значениями (Shadow attack=200, SF cost=10k, Lancer fuel=100).
**Нужна синхронизация** или ожидать план 28 Ф.5 (`/api/catalog`).

## Альтернативы отвергнуты

1. **27-H (DS ignoreAttack /150)** — слишком инвазивно. Открывает
   Cruiser anti-DS, но меняет балансы для всех юнитов.
2. **Lancer cap per planet** — лимит 50-100 Lancer на планету. Решает
   raid-проблему, но требует новой game-механики (тикет на будущее).
3. **DS speed 100 → 50** — ещё больше зафиксировать DS как защитника.
   Сейчас не критично.
4. **Strong Fighter удалить** — слишком радикально для итерации.

## Откат

Каждое изменение — точечное в YAML/коде:
- 27-F: `configs/ships.yml` shadow_ship.attack: 520 → 200, front: 5 → 7.
- 27-G: убрать строки `47: 3` и `48: 3` в `configs/rapidfire.yml` под
  Bomber.
- 27-I: SF cost/attack/...
- 27-K: shell-значения в `configs/defense.yml`.
- 27-L: SSat shell/cost.
- 27-N..T: front-значения в `configs/ships.yml` и `configs/defense.yml`.
- 27-U: вернуть поля Ballistics/Masking в Go-структуры, но это бессмысленно
  — движок их не читал.
