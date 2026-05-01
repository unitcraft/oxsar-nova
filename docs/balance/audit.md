# Balance audit — реестр игровых дыр и эксплойтов

Живой документ. Каждая найденная дыра — запись с id, датой, расчётами,
серьёзностью, статусом и ссылкой на план исправления.

**Связанные документы**:
- [docs/balance/analysis.md](analysis.md) — разбор балансных формул
- [docs/simplifications.md](../simplifications.md) — реестр упрощений при порте
- [docs/plans/](../plans/) — планы исправлений

---

## Методология

Перед добавлением записи — считаем по единому протоколу:

### Юниты (ships.yml, defense.yml)

- **attack per 1k metal-equiv**: `attack / (metal + silicon + hydrogen) × 1000`
  - metal-equiv = сумма ресурсов (без курсов рынка — упрощение)
- **shell per 1k metal-equiv**: аналогично
- **Rapidfire coverage**: для каждого юнита — кого контрит (x≥3) и кого боится
- **Порог игнорирования щита**: `attacker_attack ≤ defender_shield / 100` → атака поглощается
- **Юнит-бог**: attack/1k или shell/1k в 3+ раза выше медианы

### Экспедиции (fleet/expedition.go)

- **EV (expected value)**: `sum(weight_i / total_weight × reward_i)` по всем исходам
- **EV vs fleet cost**: если EV > 2× cost — фарм-эксплойт
- **Lost coverage**: вероятность `outcome="lost"` и что она делает (полное/частичное уничтожение)

### Экономика

- **Round-trip potential**: металл → кремний → водород → металл через рынок.
  Если финальный больше стартового — арбитраж
- **ROI зданий**: сколько часов до окупаемости на уровне 1

### Бой (battle/engine.go)

- **Щит cap**: проверка что `ignoreAttack` не растёт бесконтрольно с tech
- **Rapidfire loop**: нет ли циклов (A контрит B, B контрит A с большим коэффициентом)
- **Moral/retreat**: условия выхода из боя без потерь

---

## Серьёзность

- **P0 (крит)** — явная доминантная стратегия, ломает мету
- **P1 (высокая)** — сильный перекос, но можно играть иначе
- **P2 (средняя)** — эксплойт в узком сценарии или при high-tech
- **P3 (низкая)** — теоретический или минорный дисбаланс

---

## Статус

- **open** — найдено, не исправлено
- **planned** — есть план, ещё не реализовано (ссылка в колонке «План»)
- **patched** — исправлено, проверено
- **won't fix** — осознанно оставлено (объяснение в записи)

---

## Реестр дыр

### BA-001: Death Star доминирует в эндгейме

- **Дата находки**: 2026-04-24
- **Update 2026-04-24** (переисследование): пересмотрена диагностика.
  Серьёзность понижена до P1, основная причина — недостающий rapidfire-граф,
  а не shield/attack DS сами по себе.
- **Серьёзность**: P1 (было P0)
- **Статус**: **patched** 2026-04-24 (план 18 Фаза 1 применена, симуляция подтвердила — см. [simulation-2026-04-24.md](simulation-2026-04-24.md)). DS endgame остался, но появился жизнеспособный контрплей через BS+SD mix (в симуляции DS vs 100 BS + 50 SD — defender loss всего 0.5%). Serьёзность понижена до P2/дизайнерский.

**Файлы**: [configs/ships.yml:126-133](../../configs/ships.yml), [configs/rapidfire.yml](../../configs/rapidfire.yml)

**Фактические числа (проверено в `configs/`)**:
- Death Star (id=42): attack=200 000, shell=9 000 000, shield=50 000
- Стоимость: 5M metal + 4M silicon + 1M hydrogen = **10M metal-eq** (было ошибочно написано 14M)
- `ignoreAttack = shield/100 = 500`
- ~~**В legacy, но НЕ в nova**: DS → Frigate ×15, DS → Colony ×250, **DS → Lancer ×100**, DS → Shadow ×300~~ — **портировано** 2026-04-24
- ~~**В nova, но НЕ в legacy**: DS → Plasma ×50~~ — **удалено** 2026-04-24

**Корректная картина** (после переисследования):
- BS (attack=1000) **уже пробивает** DS (1000 > ignoreAttack=500). Прежнее утверждение «нет контры тяжёлыми кораблями» — ошибочно.
- Пробивают DS: Bomber (900), BS (1000), SD (2000), Lancer (5500), Frigate (700).
- Реальная проблема: DS в nova не убивает Lancer пачками (отсутствует DS→Lancer ×100). В legacy это было главным анти-Lancer инструментом, который уравновешивал Lancer-spam против DS.

**Фикс** ([план 18 Фаза 1](../plans/18-unit-rebalance.md)): портировать 38 недостающих rapidfire-записей из legacy (включая DS→Lancer ×100, DS→Frigate ×15, DS→Colony ×250), удалить 2 записи-изобретения. Доп. изменения (shield 50k→30k, новые rapidfire vs DS) — только после симуляции и ADR.

---

### BA-002: Lancer Ship — аномальное attack-per-resource

- **Дата находки**: 2026-04-24
- **Update 2026-04-24** (переисследование): исходные числа оказались на порядок
  неверны. Фактическая стоимость — 2.5k/7.5k/15k = 25k metal-eq (не 47.5k),
  что делает attack/1k ещё хуже: 220 (не 115). Главная причина эксплойта —
  отсутствующие в nova rapidfire-записи против Lancer. См. [план 18](../plans/18-unit-rebalance.md).
- **Серьёзность**: P0
- **Статус**: **patched** 2026-04-24. [ADR-0004](../adr/0004-lancer-cost.md): Lancer cost 2.5k/7.5k/15k → **15k/35k/60k** (metal-eq 25k → 110k). Симуляция подтверждает exchange упал с 5.11 до **1.16** (Lancer vs Cruiser), с 2.89 до **0.64** (Lancer vs mix).

**История фикса**:
1. **Rapidfire-порт** (план 18 Фаза 1, commit `7f5ce71`): exchange 13-15× (теоретически по attack/1k) → 5.11× (по симуляции). Помог, но не закрыл.
2. **Нерф стоимости** (ADR-0004, план 21 блок A): exchange 5.11× → 1.16×. Lancer-spam мёртв.

Причина, почему rapidfire не хватило: ×35 срабатывает только после kill, а Lancer убивает Cruiser быстрее (attack 5500 vs 400). К моменту, когда Cruiser получает rapidfire-bonus, у него уже критические потери.

Симуляция подтверждает: DS×100 vs Lancer продолжает работать как контра (defender loss 67%, atk loss 0).

**Файлы**: [configs/ships.yml:150-157](../../configs/ships.yml), [configs/construction.yml:521](../../configs/construction.yml), [configs/rapidfire.yml](../../configs/rapidfire.yml)

**Фактические числа** (attack per 1k metal-equiv, проверено):

| Юнит | Attack | Cost (M+Si+H) | Attack/1k |
|---|---|---|---|
| Light Fighter | 50 | 4 000 | 12.5 |
| Cruiser | 400 | 29 000 | 13.8 |
| Battleship | 1 000 | 60 000 | 16.7 |
| **Lancer** | **5 500** | **25 000** | **220** ⚠️ |

Lancer в **13–15× эффективнее** любого базового юнита — главная аномалия.

**Контры Lancer в legacy** (✅ все портированы 2026-04-24):
- LF → Lancer ×20 ✅
- SF → Lancer ×20 ✅
- Cruiser → Lancer ×35 ✅
- DS → Lancer ×100 ✅
- Bomber → Lancer ×5 ✅

Пять недостающих записей делают Lancer бессмертным для обычного флота. В legacy 1 Cruiser за ход убивает 35 Lancer'ов — поэтому там Lancer не доминировал несмотря на высокий attack/1k.

**Фикс — двухступенчатый**:
1. **План 18 Фаза 1**: портировать недостающие rapidfire → Lancer (LF×20, SF×20, Cru×35, DS×100). Возможно этого одного достаточно.
2. **План 21 Блок A** (если 18 не хватит): повысить стоимость Lancer. Решение по результатам симуляции.

**Альтернатива** (не выбрана): снизить attack 5500→3500. Ломает golden-тесты паритета с Java JAR сильнее, чем изменение rapidfire.

---

### BA-003: Экспедиции — фарм ресурсов минимальным флотом

- **Дата находки**: 2026-04-24
- **Серьёзность**: P0
- **Статус**: **patched** 2026-04-24. Применены: план 17 B1 (полное уничтожение флота при outcome=lost вместо 5-20%), план 21 B1 (min fleet ≥ 50k metal-eq при отправке), план 21 B2 (lost_weight растёт с exp_power, было 10 фикс → `10×(1+ep×0.1)`), план 21 B4 (reward cap ≤ fleet_value × 3 в expResources/expAsteroid). Ожидает [плана 20 Ф.7](../plans/20-legacy-port.md) для окончательного закрытия (astro_tech как лимит слотов экспедиций).

**Файлы**: [backend/internal/fleet/expedition.go](../../backend/internal/fleet/expedition.go) (expLoss, calcExpWeights, expResources, expAsteroid, fleetValueMetalEq), [backend/internal/fleet/transport.go:146-156](../../backend/internal/fleet/transport.go) (min fleet validation)

**Числа** для новичка (expo_tech=5, 3 часа):
- `exp_power = 5 + 6 + 0 = 11`
- Веса: `resource=100×1.22^11=810`, `ship=20×1.25^11=232`, `lost=10` (фикс)
- Сумма весов ≈ 5240
- Вероятность lost: **10/5240 = 0.19%** независимо от power
- Вероятность resource: **15.4%**, средняя награда ~5M metal-eq

**EV для 1× Light Fighter (5k cost) на 3ч**:
- `EV ≈ 0.154 × 5M + 0.056 × 5k = 770k metal-eq`
- **EV/cost ≈ 154×** — фарм бесплатных ресурсов

**Усугубляет**: `expLoss` снимает только 5-20% флота
([simplifications.md EXPEDITION expeditionLost](simplifications.md)),
хотя в legacy флот уничтожается полностью (`sendBack=false`).

**Почему это дыра**: оптимальная стратегия для нового игрока —
гонять 10× LF в экспедиции, игнорировать экономику и PvP.

**Фикс** (три уровня):
1. **План 17 B1**: `outcome="lost"` → полное уничтожение флота (порт legacy)
2. **План 20 Ф.7**: `expedition_slots = max(1, floor(sqrt(astro_level)))`
   при стартовом astro=2 → 1 слот, нельзя параллелить
3. **План 21 B**: внутри одной экспедиции — минимум флота (≥10% EV),
   `lost_weight` растёт с power, cap награды ≤3× стоимости флота

---

### BA-004: Rapidfire-граф — переформулировано

- **Дата находки**: 2026-04-24
- **Update 2026-04-24** (переисследование): изначальная формулировка была
  неверной. BS пробивает DS по атаке (1000 > ignoreAttack=500) — rapidfire
  не обязателен. Реальная проблема — **не нехватка anti-DS rapidfire**,
  а **массовый недопорт rapidfire-графа в целом**: 38 записей legacy
  отсутствуют в nova. См. план 18 §3.1.
- **Серьёзность**: P1 → **замена на BA-007** (общий порт rapidfire)
- **Статус**: **patched** 2026-04-24 (план 18 Фаза 1 применена: 38 legacy-записей добавлены, 2 изобретения удалены, 2 числовых расхождения исправлены)

**Почему переформулировано**: прежнее утверждение «Star Destroyer vs DS: x5»
неверно — в nova этой записи нет (только в legacy её тоже нет). «Bomber vs DS
×0» — у нас есть ×25 (как в legacy). Таблица в старой версии — фантазия.

Все реальные пробелы rapidfire-графа перечислены в [плане 18 §3.1](../plans/18-unit-rebalance.md).

---

### BA-005: Щиты близки к неуязвимости при high-tech

- **Дата находки**: 2026-04-24
- **Серьёзность**: P2 (не подтверждено симуляцией)
- **Статус**: planned → [план 21 Блок C](../plans/21-gameplay-hardening.md)

**Файлы**: [backend/internal/battle/engine.go:481-491](../../backend/internal/battle/engine.go)

**Механика**:
```go
ignoreAttack := unitShield / 100.0
if attack > 0 && attack <= ignoreAttack {
    // выстрел полностью поглощается щитом, урон 0
}
```

**Числа**:
- Small Shield: base shield=2000
- Shield Tech ×1.1^level: на уровне 10 → `2000 × 2.59 = 5180`
- `ignoreAttack = 5180/100 = 51.8`
- Light Fighter (attack=50) → **полное поглощение каждого выстрела**
- Защита `shieldDestroyFactor >= 0.01` есть, но баланс в массовом бою не проверен

**Почему это потенциальная дыра**: с точки зрения игрока-турели — выгодно
качать Shield Tech до 10+ и строить Small Shield, а не прокачивать оборону
в ширину. Turtle-стратегия становится слишком сильной.

**Фикс** (план 21 C):
1. C1. Golden-тест 100× Small Shield (tech=10) vs 10000× LF — за сколько
   раундов падает оборона
2. C2 (если C1 подтвердит): `ignoreAttack = baseShield / 100` без
   tech-множителя. Tech по-прежнему увеличивает абсорбцию, но не
   абсолютный порог игнора

---

### BA-006: Отсутствие антибашинга

- **Дата находки**: 2026-04-24
- **Серьёзность**: P1
- **Статус**: planned → [план 17 A1](../plans/17-gameplay-improvements.md)

**Файлы**: [backend/internal/fleet/attack.go](../../backend/internal/fleet/attack.go) — нет проверки

**Механика в legacy**:
- `consts.dm.local.php`: BASHING_PERIOD=18000 (5ч), BASHING_MAX_ATTACKS=4
- `NS.class.php:2285`: считает все атаки attacker → все планеты defender за 5ч
  (pending + finished). Блок если ≥4

**Почему это дыра**: в nova один игрок может атаковать другого бесконечно.
Для legacy это было исправлено ещё в oxsar2, у нас — не портировано.

**Фикс** (план 17 A1): порт legacy-значений 4/5h, два COUNT по `events`
(pending wait + finished ≤5h), `ErrAttackBashingLimit`.

---

## Audit движка `internal/battle/` (план 87, 2026-05-01)

Глубокий обзор движка `projects/game-nova/backend/internal/battle/`
(engine.go ~1040 строк, types.go, simstats.go) + 5 call sites
(`internal/fleet/{attack,acs_attack,expedition}.go`, `internal/alien/alien.go`,
`internal/simulator/handler.go`) и сверка с reference-Java (`d:\Sources\oxsar2-java\`).

Сводка находок: **3 critical**, **6 high**, **3 medium**, **3 low/info**.
Большинство critical — пропавшие фичи относительно legacy (debris formula,
moon chance, building destroy, ракетная атака). Самый болезненный
регресс — **BA-007** (опыт занижен в 2× в обычных боях из-за путаницы
семантики `IsMoon` vs `hasPlanet`).

### BA-007: Опыт за бой занижен в 2× для обычных атак (IsMoon vs hasPlanet)

- **Дата находки**: 2026-05-01
- **Серьёзность**: **P0** (затрагивает каждую атаку планеты — большинство боёв)
- **Статус**: **open**
- **Категория**: расхождение семантики с legacy

**Файлы**: [engine.go:140](../../projects/game-nova/backend/internal/battle/engine.go#L140), [engine.go:191-193](../../projects/game-nova/backend/internal/battle/engine.go#L191).

**Что не так**: в `Calculate` параметр `hasPlanet` для `computeExperience`
берётся из `in.IsMoon`:

```go
atkExp, defExp := computeExperience(report.Winner, report.Rounds,
    atkPower, defPower, in.IsMoon)
```

В Java [Assault.java:838](d:/Sources/oxsar2-java/assault/src/assault/Assault.java#L838) условие — `if (planetid == 0)
battlePowerCoefficient *= 0.5`. То есть «×0.5 опыта **только** если бой
без планеты-цели» (бой в полёте, ACS-перехват). У Go условие
семантически противоположно — `if !IsMoon`, что значит «×0.5 если цель
**не луна**»:

| Сценарий боя | Java (правильно) | Go (текущий баг) |
|---|---|---|
| Атака на планету (`IsMoon=false`, planetid≠0) | ×1 | **×0.5 (БАГ)** |
| Атака на луну (`IsMoon=true`, moonid≠0) | ×1 | ×1 |
| Бой в полёте (без цели) | ×0.5 | ×0.5 |
| Симулятор боя (planetid=SIM_PLANET_ID=1) | ×1 | **×0.5 (БАГ)** |

В обычной атаке планеты опыт **в 2 раза меньше нужного**. Тест
[engine_test.go:721](../../projects/game-nova/backend/internal/battle/engine_test.go#L721) `TestExperience_NoPlanet_HalfCoeff` **не ловит** баг,
потому что вызывает `computeExperience` напрямую с заранее заданным
`hasPlanet`, минуя Calculate.

**Фикс**: завести в `Input` отдельный флаг `HasPlanet bool` (по
умолчанию `true`, ставится `false` явно для боя в полёте) и передавать
его, а не `IsMoon`.

---

### BA-008: Debris formula расходится с legacy (30% vs 50%/1%)

- **Дата находки**: 2026-05-01
- **Серьёзность**: P1 (меняет экономику обломков для всех боёв)
- **Статус**: **open**
- **Категория**: расхождение формулы с legacy

**Файлы**: [attack.go:328-353](../../projects/game-nova/backend/internal/fleet/attack.go#L328), [Participant.java:773-778](d:/Sources/oxsar2-java/assault/src/assault/Participant.java#L773).

**Что не так**: в Go обломки = `30% от (metal+silicon)` потерянных
**кораблей**, оборона исключается целиком. В Java — bulk factor зависит
от типа юнита: **50% для флота, 1% для обороны** ([Assault.java:246-255](d:/Sources/oxsar2-java/assault/src/assault/Assault.java#L246)).

Эффект: при больших морских боях у нас **в ~1.67× меньше обломков от
флота**, и при разрушении обороны не падает 1% от её стоимости как в
legacy. Это меняет EV/ROI рециклеров и стимулы к атаке.

**Фикс**: либо вынести debris-расчёт в `engine.go` (заполнять
`Report.DebrisMetal/Silicon` по правилам Java), либо привести
`calcDebris` в `attack.go` к 50/1.

---

### BA-009: Moon chance — упрощённая формула (debris/100k vs композитная)

- **Дата находки**: 2026-05-01
- **Серьёзность**: P2 (геймплей-фича, влияет на эндгейм без катастроф)
- **Статус**: **open** (возможно «упрощено сознательно» — нужен ADR)
- **Категория**: пропущенная сложность legacy

**Файлы**: [attack.go:823-877](../../projects/game-nova/backend/internal/fleet/attack.go#L823), [Assault.java:1280-1370](d:/Sources/oxsar2-java/assault/src/assault/Assault.java#L1280).

**Что не так**: Go-формула шанса луны — `min(20, debris/100000)`
([attack.go:825-828](../../projects/game-nova/backend/internal/fleet/attack.go#L825)). Java-формула композитная:
- `experienceMoonChance = round(min(atkExp, defExp)^0.8)` if
  `min(atkExp, defExp) >= MOON_EXP_START_CHANCE`;
- `debrisMoonChance = (debrisM + debrisS) / MOON_PERCENT_PER_RES`;
- `moonChance = min(experienceMoonChance, debrisMoonChance, min(atkLost, defLost))`;
- `guaranteedDebrisMoonChance` (нижний пол), `MOON_MAX_CHANCE` (верх);
- `moonAllowType` модификатор (вселенная исчерпала лимит лун);
- финальный roll vs random.

Эффект: у нас луны выпадают вне зависимости от опыта обоих сторон и
числа потерь. В legacy «гарантированно крупный бой» давал большой шанс,
у нас — только большой debris.

**Фикс**: либо порт композитной формулы (с `users.e_points` как опытом),
либо ADR «упрощённый шанс луны как осознанный trade-off с указанием
причины».

---

### BA-010: Building destroy / Moon destroy / DeathStar self-destroy не реализованы

- **Дата находки**: 2026-05-01
- **Серьёзность**: P1 (для прод-запуска отсутствует целая механика)
- **Статус**: **open**
- **Категория**: фича отсутствует целиком

**Файлы**: ничего в Go, [Assault.java:850-1000](d:/Sources/oxsar2-java/assault/src/assault/Assault.java#L850) (полный блок destroy).

**Что не так**: после боя в Java есть три отдельных пути:

1. **Уничтожение луны Death Star'ами** при `targetMoon==true && attackerWon`
   — formula `chance = 2 × (DS - minDS + 1)^0.45`, capped 20%; при
   успехе луна стирается, при провале — DS-флот атакующего взрывается с
   `attackerFleetsExplodeChance = 70%`.
2. **Разрушение здания** при `targetBuildingid != 0` — formula
   `chance = 5 × DS^0.3`, capped 25%; metal/silicon здания добавляются
   к `planetMetal/Silicon` (пополняют добычу атакующего).
3. **Само-уничтожение DS при попытке destroy** — formula
   `100 - clamp(targetDestroyChance × 4, 50, 90)`. Часть DS гибнут.

В Go всех трёх путей **нет**. `Report.MoonChance/MoonCreated`,
`HaulMetal/Silicon/Hydrogen` — просто болтаются в типе как
поля, но никогда не заполняются движком (только привязка через
`fleet/attack.go::tryCreateMoon`, см. BA-009).

Эффект: невозможно уничтожить луну, нельзя выпиливать постройки врага
DS-флотом, DS не взрываются при провале атаки на луну. Это **basic
OGame-механика**.

**Фикс**: отдельный план «port destroy-mechanics», требует решения
«что считать planetDiameter для лунной формулы», работы с `targetBuildingid`
в `Input` и записи в БД.

---

### BA-011: Опыт за ракетную атаку начисляется (legacy не начисляет)

- **Дата находки**: 2026-05-01
- **Серьёзность**: P3 (мелочь, IPM редко используются)
- **Статус**: **open**
- **Категория**: расхождение правила с legacy

**Файлы**: [engine.go:139-143](../../projects/game-nova/backend/internal/battle/engine.go#L139), [Assault.java:811](d:/Sources/oxsar2-java/assault/src/assault/Assault.java#L811).

**Что не так**: Java — `if (!isRocketAttack && ...)` блокирует расчёт
опыта при IPM-ударе. У Go нет ни поля `Input.IsRocketAttack`, ни такой
проверки. В `internal/rocket/` логика IPM есть отдельно, **возможно** не
вызывает `battle.Calculate` (проверить отдельно), но если будет
переписано через Calculate — игроки получат халявный опыт за IPM.

**Фикс**: добавить `Input.IsRocketAttack bool`, в `computeExperience`
рано выйти при `isRocket=true`. Пока IPM не идут через Calculate —
trivial-задача на будущее.

---

### BA-012: SimStats.AttackerExp — не опыт, а сумма ресурсов противника

- **Дата находки**: 2026-05-01
- **Серьёзность**: P1 (критичный мисдизайн UI симулятора)
- **Статус**: **open**
- **Категория**: данные UI расходятся с реальностью

**Файлы**: [simstats.go:60-65](../../projects/game-nova/backend/internal/battle/simstats.go#L60), [types.go:102-103](../../projects/game-nova/backend/internal/battle/types.go#L102).

**Что не так**: в `MultiRun` поле `SimStats.AttackerExp` заполняется
**не очками опыта**, а **сумарными потерями ресурсов противника**:

```go
atkExp += defLostM + defLostS + defLostH
defExp += atkLostM + atkLostS + atkLostH
```

Комментарий рядом честно говорит «приближённо как в legacy», но поле
называется `attacker_exp` и попадает в JSON ответ симулятора как
`AttackerExp/DefenderExp`. UI покажет «опыт атакующего: 1 234 567»
вместо реальных 5-10 очков — т.е. в 100 000× больше.

**Фикс**: либо переименовать поля в `SimStats` на `EstimatedAtkExp`
(понятная семантика), либо **вызывать в multi-run`computeExperience`**
для каждой итерации (правильно), либо собирать `Report.AttackerExp` через
`avg(reps[i].AttackerExp)` (тоже правильно — он уже считается).

---

### BA-013: Tech.Laser/Ion/Plasma — поля игнорируются

- **Дата находки**: 2026-05-01
- **Серьёзность**: P2 (фича отсутствует, но её и в legacy частично нет)
- **Статус**: **open**
- **Категория**: фича отсутствует / не нужна

**Файлы**: [types.go:48-52](../../projects/game-nova/backend/internal/battle/types.go#L48), нигде не читается в [engine.go](../../projects/game-nova/backend/internal/battle/engine.go).

**Что не так**: в `Tech` есть `Laser/Ion/Plasma int`, но `engine.go`
их **не использует**. В legacy эти tech модифицируют атаку конкретных
unit-ов (energy weapons). В config'е на 2026-05-01 урон зашит в
`unit.Attack` по умолчанию, тех-уровень не масштабирует.

**Фикс**: либо реализовать (добавить factor аналогично `gunFactor`),
либо удалить поля из `Tech` (YAGNI). См. [docs/balance/3channel-combat-idea.md](3channel-combat-idea.md)
— там обсуждается расширение боевой системы на 3 канала, тогда поля
понадобятся.

---

### BA-014: IsAliens поле не учитывается в Calculate

- **Дата находки**: 2026-05-01
- **Серьёзность**: P3 (low — пока ApplyBattleResult фильтрует)
- **Статус**: **open**
- **Категория**: фича частично реализована

**Файлы**: [engine.go](../../projects/game-nova/backend/internal/battle/engine.go) (нигде не читается), [types.go:27](../../projects/game-nova/backend/internal/battle/types.go#L27).

**Что не так**: `Side.IsAliens bool` — есть в типе, передаётся через
JSON, копируется в `SideResult` (engine.go:993). Но в **самой формуле
боя** (uron, опыт, потери) флаг **никак не используется**. Java для
пришельцев меняет несколько правил:
- skip building-exists check ([Assault.java:912](d:/Sources/oxsar2-java/assault/src/assault/Assault.java#L912));
- особый haul (только metal/silicon, hydrogen=0);
- defender'ы-aliens не получают `users.e_points` (наш `ApplyBattleResult`
  это уже учитывает по флагу).

**Фикс**: пока не нужен. Зафиксировать как known-limitation на случай
когда будем делать «alien empire» или сложнее текущей одноразовой
HOLDING-логики.

---

### BA-015: Validate пропускает malicious input (Damaged>Quantity, ShellPercent < 0/>100, Front >> 30, Rapidfire без лимита)

- **Дата находки**: 2026-05-01
- **Серьёзность**: P1 (для симулятор-handler где input открыт через JSON)
- **Статус**: **open**
- **Категория**: дыра в правилах / возможный exploit

**Файлы**: [engine.go:970-985](../../projects/game-nova/backend/internal/battle/engine.go#L970), [simulator/handler.go](../../projects/game-nova/backend/internal/simulator/handler.go).

**Что не так**: `validate` проверяет только `Quantity < 0`. Не отвергает:

- `Damaged > Quantity` (юнит «повреждён сильнее, чем существует»). У нас
  есть `clampDamaged` на этапе `newState`, но это инвариант приложения
  — лучше отвергать на входе.
- `ShellPercent < 0` или `> 100`. Та же ситуация — `clampPercent` после.
- `Front > 30` или `< 0`. Влияет на `unitWeight = 2^Front × Quantity`.
  Если кто-то подсунет Front=63 — `2^63` overflow в float64, weight
  становится огромным, вся пропорциональная дробёжка ломается. Сейчас
  `unitWeight` clamped на [0, 30], но это в самой функции — лучше
  отвергать на входе.
- `Rapidfire[i][j]` — нет лимита значения. Злонамеренный клиент через
  `simulator/handler.go` может прислать `{1: {2: 1_000_000_000}}`,
  получить `shots = quantity × 10^9` и хотя бы вызвать timeout (или
  переполнение int64 в `shots × attack`).
- `Tech.Gun/Shield/Shell > 99` — `gunFactor = 1 + 99×0.10 = 10.9×`,
  attack умножается на 10.9. Не дыра, но защиту хорошо бы.

Симулятор-handler уже залогинен (`auth.UserID(r.Context())`), но любой
залогиненный юзер может прислать malicious input.

**Фикс**: расширить `validate`:

```go
if u.Damaged < 0 || u.Damaged > u.Quantity { return ErrInvalidInput }
if u.ShellPercent < 0 || u.ShellPercent > 100 { return ErrInvalidInput }
if u.Front < 0 || u.Front > 30 { return ErrInvalidInput }
// Rapidfire: всем парам value <= 100 (legacy максимум).
// Tech: все поля 0..99.
```

---

### BA-016: Слабый шутер всегда стреляет минимум 1 раз (rawShots=0 → 1)

- **Дата находки**: 2026-05-01
- **Серьёзность**: P3 (мелкая дисперсия в крупных боях)
- **Статус**: **open**
- **Категория**: расхождение правила с legacy / численное

**Файлы**: [engine.go:758-760](../../projects/game-nova/backend/internal/battle/engine.go#L758).

**Что не так**: при распределении выстрелов по целям пропорционально
weight:

```go
rawShots := int64(math.Round(float64(shooter.quantity) * portion))
if rawShots <= 0 {
    rawShots = 1
}
```

Если `portion ≈ 0` (слабый юнит среди очень сильных целей), он всё
равно делает **минимум 1 выстрел в каждую цель**. Например, 1 Light
Fighter среди 1000 Lancer'ов противника — стреляет в каждый Lancer
по 1 разу, давая `1000 × attack_LF` урона по совокупности.

В Java [Units.processAttack:336](d:/Sources/oxsar2-java/assault/src/assault/Units.java#L336) формула пропорциональная без
гарантии минимума. Эффект — у нас слабые юниты в **больших боях**
имеют больше «номинального» вклада.

**Фикс**: либо убрать guard (тогда слабые юниты бьют только при
`portion > 1/quantity`), либо статистический подход — `rng.IntN`
roll'ом с вероятностью `portion`. Java-вариант проще: убрать guard.

---

### BA-017: Rapidfire применяется per-target, а не per-shooter (расхождение с Java)

- **Дата находки**: 2026-05-01
- **Серьёзность**: P2 (меняет числовые цифры боя в крупных RF-цепочках)
- **Статус**: **open**
- **Категория**: расхождение формулы с legacy

**Файлы**: [engine.go:761-762](../../projects/game-nova/backend/internal/battle/engine.go#L761), [Units.java:336-340](d:/Sources/oxsar2-java/assault/src/assault/Units.java#L336).

**Что не так**: в Go rapidfire применяется **после** распределения
выстрелов:

```go
rawShots := round(quantity × portion)  // распределили
shots := rawShots * rf                  // умножили на rapidfire
```

В Java: `total_shots = quantity × rf`, потом распределение по целям.
Разница: у нас RF не масштабирует число шотов **с других целей**.
Например, BS (RF=10 vs LF) против 50% LF и 50% Cruiser:

| Метод | Всего shots при 100 BS | По LF | По Cruiser |
|---|---|---|---|
| Java | 100×10 = 1000 (только LF засчитывают RF) | 500×10=5000 | 500×1=500 |
| Go | 50×10=500 + 50×1=50 = 550 | 50×10=500 | 50×1=50 |

В Java получается заметно больше выстрелов по «своей» жертве.

**Фикс**: пересмотреть распределение — сначала `rawShots = quantity × portion`,
затем для каждой цели `shots × rapidfire(this_target)`. Возможно текущая
семантика более «справедливая», но это **отклонение от паритета**.

---

### BA-018: Front contributes to weight (получает БОЛЬШЕ выстрелов), а не absorb (получает первым)

- **Дата находки**: 2026-05-01
- **Серьёзность**: P1 (фундаментальное расхождение боевой механики)
- **Статус**: **open**
- **Категория**: расхождение концепции с legacy

**Файлы**: [engine.go:955-968](../../projects/game-nova/backend/internal/battle/engine.go#L955), [Units.java getStartTurnWeight](d:/Sources/oxsar2-java/assault/src/assault/Units.java).

**Что не так**: В Go `unitWeight = 2^Front × Quantity` означает «больше
front → больше выстрелов по этому юниту получает». То есть **высокий
front = front-line absorber**, как в Java. Но в legacy тoже так? Нужно
сверить. Если в Java Front действительно контролирует «получение
выстрелов» (sense-check: `getStartTurnWeight` название), то всё верно.
Если же Front — это «приоритет, кто бьёт в текущем раунде» — то в Go
семантика обратна.

⚠️ **Open Question**: требуется уточнить семантику `getStartTurnWeight`
в Java (нужно ~30 мин на разбор Units.java и Assault.java
shootAtSides). На момент 2026-05-01 — флаг для повторной проверки.

---

### BA-019: int64 для LostMetal — теоретическое переполнение при 10^15+ потерях

- **Дата находки**: 2026-05-01
- **Серьёзность**: P3 (теоретический edge case)
- **Статус**: **open** (close-as-not-a-fix вероятно)
- **Категория**: численное

**Файлы**: [types.go:201-203](../../projects/game-nova/backend/internal/battle/types.go#L201), [engine.go:1008-1010](../../projects/game-nova/backend/internal/battle/engine.go#L1008).

**Что не так**: `LostMetal int64` = `Σ lost × Cost.Metal`. При
`Quantity = 10^9` (теоретически возможно через тилт-баг) и
`Cost.Metal = 10^7` получим `10^16`, что < `int64.MAX = 9.2×10^18`.
Безопасно, но тонкий margin. Защита: ограничить `Quantity`
на этапе validate (см. BA-015).

---

### BA-020: RNG не совместим с java.util.Random (cross-verification невозможен)

- **Дата находки**: 2026-05-01
- **Серьёзность**: P2 (блокирует §14.4 ТЗ — golden-сверку с Assault.jar)
- **Статус**: **open** (план — отдельный)
- **Категория**: тестируемость

**Файлы**: [pkg/rng/rng.go](../../projects/game-nova/backend/pkg/rng/rng.go) (xorshift64*),
[Java Random](https://docs.oracle.com/javase/8/docs/api/java/util/Random.html) (LCG).

**Что не так**: ADR-0002 декларирует «совместимость по семантике, не
bit-в-bit», но §14.4 ТЗ требует cross-verification против Assault.jar.
Это требует **JavaRandom-адаптера** (xorshift и Java.util.Random не
совпадают, у Java LCG `seed = (seed × 0x5DEECE66D + 0xB) & ((1L<<48)-1)`).
Адаптер сейчас отсутствует.

**Фикс**: реализовать `rng.NewJavaRandom(seed)` для cross-verification.
Не блокирует прод (наш xorshift статистически валиден), но блокирует
golden-снимки.

---

### BA-021: Multi-sim первая итерация имеет другой RNG-character (seed=0 → golden)

- **Дата находки**: 2026-05-01
- **Серьёзность**: P3 (артефакт)
- **Статус**: **open**
- **Категория**: численное

**Файлы**: [simstats.go:32](../../projects/game-nova/backend/internal/battle/simstats.go#L32), [rng.go:23-27](../../projects/game-nova/backend/pkg/rng/rng.go#L23).

**Что не так**: в `MultiRun(in, n)` `in.Seed = seed0 + i`. При
`seed0 == 0`:
- i=0 → `rng.New(0)` подменяет на golden-ratio константу.
- i=1 → `rng.New(1)` нормальный xorshift.

Первая симуляция из N имеет другой character RNG. Не сильно меняет
статистику, но создаёт неожиданный pattern — `seed=0` всегда даёт
тот же исход, не зависящий от N.

**Фикс**: либо убрать guard в `rng.New` (требует изменения семантики
RNG для всех клиентов), либо в `MultiRun` стартовать `i = 1` если
`seed0 == 0`.

---



Записи о том, что проверили и дыр не нашли — чтобы не перепроверять.

### NF-001: Рынок 1:2:4 — арбитраж невозможен

- **Проверено**: 2026-04-24
- **Файлы**: [backend/internal/market/service.go](../../backend/internal/market/service.go)
- **Расчёт**: round-trip metal → silicon → hydrogen → metal теряет 30.5%
  (комиссии × 3 обмена). Арбитраж в любую сторону невыгоден.

### NF-002: Metal Mine — окупаемость не эксплойт

- **Проверено**: 2026-04-24
- **Расчёт**: Metal Mine уровня 1 окупается за ~3 часа (45 мин при gamespeed=4).
  Быстро, но by design — стимулирует максимизацию production. Не эксплойт.

### NF-003: Боевой движок (ballistics/masking/ablation) — корректен

- **Проверено**: 2026-04-24
- **Update 2026-05-01**: ⚠️ **частично пересмотрено** — глубокий audit
  (план 87) подтвердил корректность *именно этих трёх блоков*
  (ballistics/masking, multi-channel attack, ablation order). Но
  движок в целом имеет **15 находок BA-007..BA-021**, см. секцию
  «Audit движка» выше. NF-003 НЕ покрывает: формулу опыта (BA-007),
  debris (BA-008), отсутствие destroy/moon-механик (BA-009/BA-010),
  validate-дыру (BA-015), front-семантику (BA-018), и др.
- **Файлы**: [backend/internal/battle/engine.go](../../backend/internal/battle/engine.go)
- **Проверено**: формула ballistics vs masking распределяет выстрелы
  правильно. Multi-channel attack раздаёт урон по каналам корректно.
  Partial damage (ablation) применяется в правильном порядке.

---

## Процедура обновления

При находке новой дыры:
1. Посчитать по методологии выше — с конкретными числами
2. Добавить запись `BA-NNN` с датой, серьёзностью, файлами, числами, почему
3. Если есть план исправления — статус `planned` + ссылка
4. Если нет — `open`, создаётся план или запись в существующий
5. После коммита фикса — статус `patched`

При пересмотре серьёзности (например, после плейтеста — оказалось не так
страшно) — дописать **Update YYYY-MM-DD** в запись, не удалять историю.
