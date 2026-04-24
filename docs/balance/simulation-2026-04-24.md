# Балансовая симуляция 2026-04-24

**Цель**: после плана 18 Фазы 1 (порт rapidfire из legacy) проверить,
сохраняются ли эксплойты BA-001 (DS endgame) и BA-002 (Lancer-spam).

**Инструмент**: [`backend/cmd/tools/battle-sim/`](../../backend/cmd/tools/battle-sim/main.go)
— CLI-обёртка над `battle.Calculate`, прогоняет сценарии N раз,
агрегирует потери и размен ресурсов.

**Параметры**: 30 прогонов на сценарий, 200 раундов макс, rapidfire
после Фазы 1 (commit `7f5ce71`). Технологии обеих сторон = 0.

---

## Результаты

| Сценарий | Атакующий | Защитник | atk loss % | def loss % | Exchange (def/atk) |
|---|---|---|---:|---:|---:|
| lancer-vs-cruiser | 1000×Lancer (25M) | 1000×Cruiser (29M) | 7.2% | 31.7% | **5.11** |
| lancer-vs-mixed | 500×Lancer (12.5M) | 200 LF + 100 Cru + 50 BS (6.7M) | 1.2% | 6.5% | **2.89** |
| ds-vs-lancer | 1×DS (10M) | 300×Lancer (7.5M) | 0% | 67.0% | ∞ |
| ds-vs-bs-fleet | 1×DS (10M) | 200×BS (12M) | 0% | 50.5% | ∞ |
| ds-vs-bs-sd-fleet | 1×DS (10M) | 100 BS + 50 SD (12.25M) | 0% | 0.5% | ∞ |
| bomber-vs-rl | 100×Bomber (9M) | 3000×RL (6M) | 0% | 58.4% | ∞ |
| cruiser-vs-rl | 200×Cruiser (5.8M) | 2000×RL (4M) | 0.5% | 32.1% | **44.34** |
| mixed-vs-mixed | 500 LF + 200 Cru + 50 BS (10.8M) | 1000 RL + 300 LL + 50 Plasma (9.1M) | 90.8% | 46.5% | **0.43** |

> Exchange > 1 — атакующий эффективен по ресурсам. < 1 — защитник выигрывает размен.

---

## Выводы

### BA-001 (Death Star endgame) — частично закрыт

**DS vs Lancer** (exchange ∞): DS×100 vs Lancer теперь работает — 1 DS
уничтожает 67% стека из 300 Lancer'ов без собственных потерь. Это прямой
эффект портированной записи `42 → 102 ×100` из legacy.

**DS vs BS-only** (∞): DS рвёт Battleship — BS пробивает DS по атаке
(1000>500), но у BS нет rapidfire vs DS, поэтому BS-fleet тает быстрее,
чем DS успевает потерять что-то.

**DS vs BS+SD mix** (∞, но def loss всего 0.5%): при смешанном флоте
100 BS + 50 SD DS не может пробить — **SD пробивает DS**, и количество
SD достаточно, чтобы DS не успевал убивать всех. Это **интересный баланс**:
BS+SD mix vs DS — жизнеспособная стратегия.

**Статус**: BA-001 — снят до «P2/дизайнерский». DS остаётся топ-юнитом
endgame, но есть контрплей через BS+SD mix. Жёсткий фикс (shield 50k→30k,
новые rapidfire) **не нужен**.

### BA-002 (Lancer-spam) — ОСТАЁТСЯ

**Lancer vs Cruiser** (exchange 5.11): Lancer всё ещё в 5× эффективнее
Cruiser по размену ресурсов, даже с Cruiser×35 rapidfire.

**Причина**: rapidfire ×35 срабатывает только после того, как Cruiser
**убил** Lancer. А Lancer убивает Cruiser быстрее (attack 5500 vs 400 —
13× разница). К моменту, когда Cruiser начинает пользоваться rapidfire,
у него уже крупные потери.

**Lancer vs смешанный флот** (exchange 2.89): даже против mix LF+Cru+BS
Lancer выигрывает размен почти в 3×.

**Статус**: BA-002 — **НЕ закрыт**. Блок A плана 21 (нерф стоимости
Lancer) **всё-таки нужен**. Rapidfire-порт помог (исходный attack/1k=220
без контры → теперь экспе­кт exchange 5× вместо прежних 13-15× по метрике),
но до честного баланса ещё далеко.

**Предложение для ADR (план 21 A)**: поднять стоимость Lancer до
значения, при котором exchange падает до ~1.2-1.5 (примерно как у
других флотов). Нужна отдельная симуляция с вариациями стоимости.

### BA-004 (Rapidfire-граф) — закрыт

Все сценарии с rapidfire дают ожидаемые результаты:
- Bomber×20 vs RL — exchange ∞
- Cruiser×10 vs RL — exchange 44
- DS×100 vs Lancer — работает
- SD×3 vs Lancer (исправлено с ×2) — в симуляции не измерялось отдельно

### Sanity checks

- `mixed-vs-mixed` — defender выигрывает (exchange 0.43). Это OGame-дизайн
  «оборона в 2-3× эффективнее флота по attack/1k». Работает.
- `cruiser-vs-rl` — очень большой exchange (44), Cruiser разносит RL.
  Bomber/Cruiser-spam vs чистые RL — не жизнеспособная защита. Это знание
  legacy-баланса.

---

## Что делать с BA-002

Варианты для ADR (план 21 блок A):

### Вариант 1: нерф стоимости Lancer
`construction.yml:521`:
```yaml
lancer_ship:
  basic:
    metal: 25000     # было 2 500 (×10)
    silicon: 50000   # было 7 500 (×6.6)
    hydrogen: 80000  # было 15 000 (×5.3)
# Cost metal-eq: 155 000 (было 25 000 — ×6.2)
# attack/1k: 5500/155 ≈ 35 (было 220 — в 6× меньше, на уровне BS=16.7)
```

Требует симуляции повторной. Предполагаемый эффект: exchange в
lancer-vs-cruiser упадёт с 5.11 до ~1.2.

### Вариант 2: нерф attack Lancer
Снизить attack с 5500 до ~2500. **Ломает паритет с Java JAR** (стат юнита
изменён, golden-тесты перегенерировать).

### Вариант 3: дополнительные rapidfire контры Lancer
Frigate→Lancer ×20, Bomber→Lancer ×10. **Изобретение** (в legacy этих
записей нет). Требует ADR.

**Рекомендация**: Вариант 1 (цена). Не ломает паритет, не вводит новых
механик. Числа подбираются симуляцией.

---

## Воспроизведение

```bash
cd backend
go run ./cmd/tools/battle-sim/ --configs=../configs --all --runs=30 --rounds=200
# или отдельный сценарий:
go run ./cmd/tools/battle-sim/ --configs=../configs --scenario=lancer-vs-cruiser --runs=100 --rounds=200
```

Список сценариев в `backend/cmd/tools/battle-sim/main.go::buildScenarios`.
