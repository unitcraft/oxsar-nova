# Balance audit — реестр игровых дыр и эксплойтов

Живой документ. Каждая найденная дыра — запись с id, датой, расчётами,
серьёзностью, статусом и ссылкой на план исправления.

**Связанные документы**:
- [docs/balance-analysis.md](balance-analysis.md) — разбор балансных формул
- [docs/simplifications.md](simplifications.md) — реестр упрощений при порте
- [docs/plans/](plans/) — планы исправлений

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
- **Серьёзность**: P0
- **Статус**: planned → [план 18 §2.8](plans/18-unit-rebalance.md)

**Файлы**: [configs/ships.yml:126-133](../configs/ships.yml), [configs/rapidfire.yml](../configs/rapidfire.yml)

**Числа**:
- Death Star (id=42): attack=200,000, shell=9,000,000, shield=50,000
- Стоимость: 5M metal + 4M silicon + 1M hydrogen = **14M metal-eq**
- Rapidfire: x30 BS, x33 Cruiser, x100 Strong Fighter, x250 транспортёры
- Единственная контра — Lancer (id=102) с rapidfire x3 vs DS
- Чтобы убить 1 DS (14M) нужно ~1636× Lancer — **~41M ресурсов**
- BS vs DS: rapidfire **x0** → нет контры тяжёлыми кораблями

**Почему это дыра**: ассиметрия 1:3 не в пользу атакующего DS. Единственный
эндгейм-сценарий — DS vs DS. Нет интересного контрплея.

**Фикс** (план 18): shield 50k→30k (снижает порог игнора с 500 до 300 →
Battleship attack=1000 теперь пробивает броню), добавить rapidfire BS×2,
SD×3 vs DS.

---

### BA-002: Lancer Ship — аномальное attack-per-resource

- **Дата находки**: 2026-04-24
- **Серьёзность**: P0
- **Статус**: planned → [план 21 Блок A](plans/21-gameplay-hardening.md)

**Файлы**: [configs/ships.yml:150-157](../configs/ships.yml)

**Числа (attack per 1k metal-equiv)**:
| Юнит | Attack | Cost (metal-eq) | Attack/1k |
|---|---|---|---|
| Light Fighter | 50 | 4k | 12.5 |
| Cruiser | 400 | ~30k | 13.3 |
| Battleship | 1000 | ~50k | 20.0 |
| **Lancer** | **5500** | **47.5k** | **115.8** |

Нормализованно с учётом shell: Lancer ≈ 61.1 attack/1k — **в 3-5× эффективнее** любого другого юнита.

**Почему это дыра**: оптимальная мета = строить только Lancer (+DS endgame).
Frigate, HF, Cruiser, Bomber экономически бессмысленны. План 18 удешевляет
альтернативы, но Lancer сам остаётся супер-юнитом.

**Фикс** (план 21 A1): повысить стоимость Lancer до 50k/10k/20k
(attack/1k падает до ~30 — уровень Cruiser/Frigate). Lancer остаётся
сильным, но элитным, а не spam-юнитом.

**Альтернатива A2** (не выбрана): снизить attack 5500→3500. Ломает
golden-тесты паритета с Java JAR.

---

### BA-003: Экспедиции — фарм ресурсов минимальным флотом

- **Дата находки**: 2026-04-24
- **Серьёзность**: P0
- **Статус**: planned → [план 17 B1](plans/17-gameplay-improvements.md) + [план 20 Ф.7](plans/20-legacy-port.md) + [план 21 Блок B](plans/21-gameplay-hardening.md)

**Файлы**: [backend/internal/fleet/expedition.go:676-725](../backend/internal/fleet/expedition.go)

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

### BA-004: Rapidfire-граф — нет контры DS тяжёлыми кораблями

- **Дата находки**: 2026-04-24
- **Серьёзность**: P1
- **Статус**: planned → [план 18 §2.8](plans/18-unit-rebalance.md)

**Файлы**: [configs/rapidfire.yml:50-68](../configs/rapidfire.yml)

**Числа**:
- Battleship vs DS: rapidfire **x0**
- Star Destroyer vs DS: x5 (мало для корабля стоимостью 150k)
- Bomber vs DS: x0
- Только Lancer x3 и Plasma Turret x2

**Почему это дыра**: связана с BA-001. Игрок с DS неуязвим для любого
флота кроме Lancer-роя.

**Фикс** (план 18): BS→DS x2, SD→DS x3 (вместо x5). Battleship-флот
становится жизнеспособной анти-DS-стратегией.

---

### BA-005: Щиты близки к неуязвимости при high-tech

- **Дата находки**: 2026-04-24
- **Серьёзность**: P2 (не подтверждено симуляцией)
- **Статус**: planned → [план 21 Блок C](plans/21-gameplay-hardening.md)

**Файлы**: [backend/internal/battle/engine.go:481-491](../backend/internal/battle/engine.go)

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
- **Статус**: planned → [план 17 A1](plans/17-gameplay-improvements.md)

**Файлы**: [backend/internal/fleet/attack.go](../backend/internal/fleet/attack.go) — нет проверки

**Механика в legacy**:
- `consts.dm.local.php`: BASHING_PERIOD=18000 (5ч), BASHING_MAX_ATTACKS=4
- `NS.class.php:2285`: считает все атаки attacker → все планеты defender за 5ч
  (pending + finished). Блок если ≥4

**Почему это дыра**: в nova один игрок может атаковать другого бесконечно.
Для legacy это было исправлено ещё в oxsar2, у нас — не портировано.

**Фикс** (план 17 A1): порт legacy-значений 4/5h, два COUNT по `events`
(pending wait + finished ≤5h), `ErrAttackBashingLimit`.

---

## Не-дыры (проверено)

Записи о том, что проверили и дыр не нашли — чтобы не перепроверять.

### NF-001: Рынок 1:2:4 — арбитраж невозможен

- **Проверено**: 2026-04-24
- **Файлы**: [backend/internal/market/service.go](../backend/internal/market/service.go)
- **Расчёт**: round-trip metal → silicon → hydrogen → metal теряет 30.5%
  (комиссии × 3 обмена). Арбитраж в любую сторону невыгоден.

### NF-002: Metal Mine — окупаемость не эксплойт

- **Проверено**: 2026-04-24
- **Расчёт**: Metal Mine уровня 1 окупается за ~3 часа (45 мин при gamespeed=4).
  Быстро, но by design — стимулирует максимизацию production. Не эксплойт.

### NF-003: Боевой движок (ballistics/masking/ablation) — корректен

- **Проверено**: 2026-04-24
- **Файлы**: [backend/internal/battle/engine.go](../backend/internal/battle/engine.go)
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
