# План 18: Ребалансировка юнитов

**Дата**: 2026-04-24  
**Статус**: ЧЕРНОВИК — требует ADR и согласования с геймдизайном  
**Затрагивает**: `configs/ships.yml`, `configs/defense.yml`, `configs/rapidfire.yml`, `configs/construction.yml`

> **Принцип**: формулы боевого движка не трогаем. Меняем только входные данные —
> характеристики юнитов и стоимости. Паритет с Java JAR сохраняется (движок те же
> числа считает, меняются только статы в конфигах).

---

## 1. Диагностика проблем баланса

### 1.1 Юниты-пустышки (почти никогда не строятся)

| Юнит | Почему не строится |
|---|---|
| **Heavy Fighter (32)** | 6,000M/4,000Si. Cruiser стоит 20k/7k и имеет ×6 rapidfire vs Light Fighter плюс ×10 vs Rocket Launcher. HF ничем не отличается от LF, просто дороже. |
| **Star Destroyer (41)** | 60k/50k/15k — дороже Bomber (50k/25k/15k), медленнее (5000 vs 4000), у Bomber выше DPS/стоимость против защит. SD привлекателен только против Frigates (×2 rapidfire) и нескольких лазеров. |
| **Ion Cannon (46)** | ~~Отсутствует в `defense.yml` — нет статов, фактически сломан.~~ ⚠ **УСТАРЕЛО:** фикс уже в main (commit `8820826` от 2026-04-23) — ion_gun добавлен со статами `attack=150, shield=500, shell=8000`. Пункт 2.6 ниже — no-op, ничего добавлять не нужно. |
| **Shadow Ship (325)** | 1,000M/3,000Si/1,000H, masking=5, attack=30. Нет rapidfire против чего-либо. Полезен только для шпионажа (masking), но шпионажная механика отдельна. |

### 1.2 Доминантные стратегии (нет контрплея)

**Проблема 1 — Cruiser spam**: Cruiser имеет ×6 rapidfire vs LF и ×10 vs Rocket Launcher.
Любая защита из Rocket Launcher'ов моментально уничтожается. LF как контр-юниты
не работают, потому что у Cruiser 400 атаки vs 10 щита LF — порог щита 400/100=4,
LF имеет shield=10, значит атака Cruiser (400) > 10/100=0.1, Cruiser бьёт по броне.
У LF нет rapidfire против Cruiser. Результат: stack Cruisers = win.

**Проблема 2 — Deathstar endgame deadlock**: Deathstar (DS) имеет attack=200,000 и
shell=9,000,000. Против DS есть только Lancer (×3 rapidfire, attack=5,500) и
Plasma Turret (×2 rapidfire). Lancer attack 5,500 vs DS shield 50,000 — порог
50,000/100=500, Lancer (5,500) > 500, бьёт по броне. Но one DS = 9M брони.
Для убийства одного DS нужно ~9,000,000/5,500 ≈ 1,636 выстрелов Lancer, т.е.
1,636 Lancer'ов на 1 DS. 1 DS стоит ~10M ресурсов. 1,636 Lancer = 1,636 × 25k =
~41M ресурсов. DS выгоден только при числе ≥3 (тогда эффект rapidfire множится).
Нет интересного контрплея, просто гонка числа DS.

**Проблема 3 — Frigate недоступен**: Frigate (35) стоит 30k/40k/15k. Это ~150k
ресурсов в эквиваленте. При этом у него rapidfire против HF (×7), Cruiser (×4),
Battleship (×7). Frigate — идеальный counter vs Cruiser spam, но его стоимость
Silicon-heavy (40k Si vs 30k M) делает его недоступным на ранней игре. Он должен
появляться раньше, как counter vs Cruiser spam.

**Проблема 4 — Bomber один делает всё**: Bomber имеет ×20 rapidfire vs Rocket
Launcher, ×20 vs Light Laser, ×10 vs Heavy Laser, ×10 vs Ion Cannon. Один юнит
покрывает всю защитную линейку. Нет смысла строить микс из кораблей для атаки:
«возьми Bombers — и победишь любую оборону». Bomber стоит 50k/25k/15k — дорого,
но эффективнее любого другого анти-оборонного юнита.

### 1.3 Защита vs Флот — структурный дисбаланс

Из balance-analysis.md: атака/1000 ресурсов у обороны 2–3× выше флота. Это
намеренно (OGame-дизайн: оборона должна быть выгодна). Но текущий дисбаланс слишком
резкий: Bomber нивелирует преимущество через rapidfire ×20.

Целевой баланс: оборона выгоднее флота в DPS/ресурс на ~1.5×, не 2–3×.

### 1.4 Структурные проблемы прогрессии

- **LF → HF → Cruiser**: шаг HF избыточен. Игрок строит LF → сразу Cruiser.
- **Battleship vs Frigate**: нет четкой роли. Battleship дешевле (40k vs 70k),
  но у Frigate больше щит и rapidfire против Cruiser/HF/Battleship.
  Battleship должен быть «доступный вариант», Frigate — «специализированный контр».

---

## 2. Предлагаемые изменения

> Все числа — **предложение для ADR**. До согласования ничего не меняется в коде.

### 2.1 Heavy Fighter — переосмысление роли

**Текущие**: Attack=150, Shield=25, Shell=10,000, Speed=10,000, Cargo=100, Fuel=75  
**Стоимость**: 6,000M/4,000Si

**Проблема**: дорогой LF без identity.

**Изменение**: превратить HF в специализированного Cruiser-counter через rapidfire.

```yaml
# ships.yml — Heavy Fighter (32)
attack: 200         # +50 (был 150)
shield: 50          # +25 (был 25)
shell: 10000        # без изменений
speed: 10000        # без изменений
cargo: 100          # без изменений
fuel: 75            # без изменений

# стоимость — снизить Silicon barrier
cost:
  metal: 6000       # без изменений
  silicon: 2000     # -2000 (был 4000) — сделать доступнее на Si
  hydrogen: 0
```

```yaml
# rapidfire.yml — добавить новые записи:
- from: 32  # Heavy Fighter
  to: 31    # Light Fighter
  factor: 3             # HF эффективен против LF-роёв

- from: 32  # Heavy Fighter
  to: 29    # Small Transporter
  factor: 2             # HF полезен для рейдов
```

**Геймплейный эффект**: HF становится «анти-рой» контрюнитом. Против армии LF выгоден.
Cruiser остаётся лучше против защит через rapidfire ×10/×6. Теперь есть выбор.

---

### 2.2 Star Destroyer — узкая роль за нормальную цену

**Текущие**: Attack=2,000, Shield=500, Shell=110,000, Speed=5,000, Cargo=2,000, Fuel=1,000  
**Стоимость**: 60,000M/50,000Si/15,000H

**Проблема**: стоит как DS-половинка, но роль нечёткая.

**Изменение**: снизить стоимость, добавить rapidfire против Battleship — сделать
чётким «убийцей тяжёлых кораблей».

```yaml
# ships.yml — Star Destroyer (41)
attack: 2000        # без изменений
shield: 600         # +100 (был 500)
shell: 110000       # без изменений
speed: 6000         # +1000 (был 5000) — чуть подвижнее
fuel: 800           # -200 (был 1000) — чуть дешевле в эксплуатации

# стоимость — снизить Silicon barrier
cost:
  metal: 50000      # -10000 (был 60000)
  silicon: 35000    # -15000 (был 50000)
  hydrogen: 15000   # без изменений
```

```yaml
# rapidfire.yml — Star Destroyer (41):
- from: 41
  to: 34    # Battleship
  factor: 3             # SD = Battleship killer (новое)
# Существующие: vs Frigate ×2, vs LightLaser ×2, vs HeavyLaser ×2, vs Lancer ×2
```

**Геймплейный эффект**: SD появляется как «mid-game heavy killer» после Bomber.
Battleship теперь уязвим к SD → нужна защита от SD → строй Frigates. Новый цикл.

---

### 2.3 Frigate — снижение Silicon-барьера

**Текущие**: Attack=700, Shield=400, Shell=70,000, Speed=10,000, Cargo=750, Fuel=250  
**Стоимость**: 30,000M/40,000Si/15,000H

**Проблема**: 40k Silicon слишком высокий барьер входа.

**Изменение**: только стоимость.

```yaml
# ships.yml — Frigate (35)
cost:
  metal: 35000      # +5000 (был 30000) — небольшой рост M
  silicon: 28000    # -12000 (был 40000) — снижаем Si-барьер
  hydrogen: 10000   # -5000 (был 15000) — снижаем H-барьер
```

**Геймплейный эффект**: Frigate становится доступен на среднем этапе. Контрплей
против Cruiser spam (rapidfire ×4 vs Cruiser) становится реальным вариантом, а не
роскошью.

---

### 2.4 Bomber — нерф rapidfire против лёгких защит

**Текущие**: rapidfire ×20 vs Rocket Launcher, ×20 vs Light Laser

**Проблема**: Bomber слишком универсален. Любая оборона бесполезна против Bombers.

**Изменение**: только rapidfire.yml, стат-блок не меняем.

```yaml
# rapidfire.yml — Bomber (40):
# Было:
#   to: 43 (Rocket Launcher) factor: 20
#   to: 44 (Light Laser)     factor: 20
#   to: 45 (Heavy Laser)     factor: 10
#   to: 46 (Ion Cannon)      factor: 10
# Становится:
- from: 40
  to: 43    # Rocket Launcher
  factor: 12    # -8 (был 20) — Rocket Launcher дешевле 2x против Bomber
- from: 40
  to: 44    # Light Laser
  factor: 12    # -8 (был 20)
- from: 40
  to: 45    # Heavy Laser
  factor: 8     # -2 (был 10)
- from: 40
  to: 46    # Ion Cannon
  factor: 8     # -2 (был 10)
```

**Геймплейный эффект**: Rocket Launcher и Light Laser снова имеют смысл как
«первичная защита от рейдов». Нужно строить больше Bombers или добавлять другие
корабли для атаки обороны. Оборонительные постройки дешевле относительно Bomber.

---

### 2.5 Cruiser — лёгкий нерф против защит

**Текущие**: rapidfire ×10 vs Rocket Launcher

**Изменение**: только rapidfire.yml.

```yaml
# rapidfire.yml — Cruiser (33):
# Было: to: 43 (Rocket Launcher) factor: 10
# Становится:
- from: 33
  to: 43    # Rocket Launcher
  factor: 6     # -4 (был 10)
- from: 33
  to: 31    # Light Fighter
  factor: 6     # без изменений
```

**Геймплейный эффект**: Cruiser остаётся лучшим универсальным кораблём, но
Rocket Launcher как «anti-cruiser spam» снова имеет смысл на ранней игре.

---

### 2.6 Ion Cannon — фикс и переработка

> ⚠ **Уже сделано.** На момент написания плана 18 казалось, что `ion_gun`
> отсутствует в `defense.yml`. Проверка показала: юнит добавлен в коммите
> `8820826` (2026-04-23) со статами, ровно совпадающими с теми, что здесь
> предлагалось «добавить»: `attack: 150, shield: 500, shell: 8000`.
> Этот пункт плана — no-op, переходите к следующему.

**Изначально было в плане**: статы отсутствуют в `defense.yml` — юнит сломан.

```yaml
# defense.yml — Ion Cannon (46) — ДОБАВИТЬ:
# (уже есть, см. configs/defense.yml::ion_gun)
- id: 46
  name: ion_cannon
  attack: 150
  shield: 500       # очень высокий щит — специализация «анти-rapidfire»
  shell: 8000
```

**Геймплейный эффект**: Ion Cannon становится «защитой от Cruiser-спама» — высокий
щит снижает количество попаданий по броне от слабых атак. Оборонительная ниша.
При attack=150 и shield=500 — Cruiser (attack=400) бьёт по броне, но высокий щит
частично поглощает урон каждый раунд.

Добавить также в rapidfire:

```yaml
# rapidfire.yml — добавить:
- from: 40  # Bomber
  to: 46    # Ion Cannon
  factor: 6             # -4 (был 10) — согласовано с нерфом Bomber
```

---

### 2.7 Shadow Ship — добавить геймплейную роль

**Текущее**: Attack=30, Shield=30, Shell=4,000, masking=5, ballistics=5

**Проблема**: нет боевой роли. Masking снижает попадания по нему, но он слишком
слаб, чтобы иметь значение в бою.

**Изменение**: добавить Shadow Ship rapidfire против Espionage Probe и экономичный
Scout-raid профиль.

```yaml
# ships.yml — Shadow Ship (325)
attack: 50          # +20 (был 30) — слабо, но не ноль
shield: 50          # +20 (был 30)
shell: 5000         # +1000 (был 4000)
cargo: 150          # +75 (был 75) — лучше для мелких рейдов
fuel: 25            # -10 (был 35) — дешевле содержать
```

```yaml
# rapidfire.yml — Shadow Ship (325):
- from: 325
  to: 38    # Espionage Probe
  factor: 5             # Scout killer
```

**Геймплейный эффект**: Shadow Ship становится «Anti-spy scout» — высокий masking
делает его труднодостижимым, rapidfire против зондов. Нишевый, но понятный.

---

### 2.8 Deathstar — разнообразие контрплея

**Проблема**: DS-vs-DS — единственный эндгейм сценарий. Нет интересных решений.

**Изменение**: добавить rapidfire против DS для нескольких юнитов, снизить порог
щита чтобы тяжёлые корабли могли бить по броне DS.

```yaml
# ships.yml — Deathstar (42)
shield: 30000       # -20000 (был 50000) — снизить щит для лучшего контрплея
                    # порог игнора теперь 300 (был 500)
                    # Lancer (attack=5500) > 300 — бьёт по броне (уже работало)
                    # Star Destroyer (attack=2000) > 300 — теперь тоже бьёт!
                    # Battleship (attack=1000) > 300 — теперь тоже бьёт!
```

```yaml
# rapidfire.yml — добавить новые записи:
- from: 34  # Battleship
  to: 42    # Deathstar
  factor: 2             # Battleship fleet как DS counter

- from: 41  # Star Destroyer
  to: 42    # Deathstar
  factor: 3             # SD теперь лучший DS hunter (не Lancer)
```

**Геймплейный эффект**:
- Battleship + SD fleet = жизнеспособный anti-DS контр, а не только Lancer
- Lancer (премиум) по-прежнему лучший, но теперь не единственный вариант
- DS против DS всё ещё работает, но не доминирует так жёстко

---

## 3. Итоговая матрица изменений

| Юнит/Защита | Изменение | Тип |
|---|---|---|
| Heavy Fighter (32) | Снизить Si-стоимость, добавить rapidfire vs LF ×3 | Buff |
| Star Destroyer (41) | Снизить стоимость, +rapidfire vs Battleship ×3 | Buff |
| Frigate (35) | Снизить Si и H стоимость | Buff |
| Shadow Ship (325) | +attack, +shield, +cargo, -fuel, +rapidfire vs Probe ×5 | Buff |
| Ion Cannon (46) | Добавить статы (был сломан), shield=500 | Fix + Design |
| Bomber (40) | rapidfire vs RocketLauncher/LightLaser: 20→12 | Nerf |
| Cruiser (33) | rapidfire vs RocketLauncher: 10→6 | Nerf |
| Deathstar (42) | shield: 50000→30000; +rapidfire vs DS от Battleship ×2 и SD ×3 | Rebalance |

---

## 4. Ожидаемый эффект на геймплей

| Геймплейная цель | До | После |
|---|---|---|
| Контрплей против Cruiser spam | Только Frigate (недоступен) | Frigate + HF + Ion Cannon |
| Контрплей против DS endgame | Только Lancer (премиум) | Battleship fleet + SD + Lancer |
| Роль Heavy Fighter | Нет (ignore) | Anti-LF swarm specialist |
| Роль Star Destroyer | Нечёткая | Battleship hunter |
| Оборона vs Bomber | Бесполезна (rapidfire ×20) | Rocket Launcher стоит строить |
| Ion Cannon | Сломан | Anti-rapidfire shield wall |
| Shadow Ship | Декорация | Anti-scout + стелс-рейдер |

---

## 5. Риски и ограничения

1. **Java JAR paritry**: боевой движок не меняем. Изменения — только в YAML-конфигах.
   Java симулятор использует те же конфиги — проверить, что `oxsar2-java.jar`
   читает конфиги из `configs/` или нужна синхронизация.

2. **Legacy oxsar2**: изменения ломают совместимость с legacy-балансом. Это **ADR-вопрос** —
   принимаем осознанно. Записать в `docs/simplifications.md`.

3. **Playtesting**: без реального плейтестинга любые числа — гипотезы.
   Рекомендуется тестировать в dev-среде с командой перед деплоем.

4. **Экономические последствия**: снижение стоимостей Frigate/SD/HF изменяет
   resource sink — игроки тратят меньше на те же силы, производство «чувствует» себя
   богаче. Нужно проверить, не нарушает ли это economy loop.

5. **Абузы**:
   - Battleship×2 rapidfire vs DS — нужно проверить, что 200 Battleships не убивают
     DS слишком дёшево (симулируем перед деплоем)
   - Ion Cannon shield=500 + rapidfire от Bomber 6 (не 10): проверить, что это не
     делает Ion Cannon слишком дешёвой защитой

---

## 6. Порядок реализации (рекомендация)

| Фаза | Задача | Приоритет |
|---|---|---|
| 18.1 | ADR: согласовать все числа с геймдизайном | **Обязательно** |
| 18.2 | ~~Фикс Ion Cannon (добавить статы в defense.yml)~~ — **уже сделано в commit 8820826, пропустить** | ✅ done |
| 18.3 | Нерф Bomber rapidfire (20→12 vs лёгкие защиты) | H |
| 18.4 | Нерф Cruiser rapidfire vs Rocket Launcher (10→6) | H |
| 18.5 | Buff Frigate (снижение стоимости) | M |
| 18.6 | Rebalance DS (shield 50k→30k, новые rapidfire контры) | M |
| 18.7 | Buff HF (стоимость + rapidfire) | M |
| 18.8 | Buff Star Destroyer (стоимость + rapidfire vs BS) | L |
| 18.9 | Buff Shadow Ship (статы + rapidfire) | L |

---

## 7. ADR-требования (обязательно перед реализацией)

- **18-A**: Принять/отклонить нерф Bomber (rapidfire 20→12). Обоснование:
  «восстановление ценности Rocket Launcher как starting defense».
- **18-B**: Принять/отклонить снижение щита DS (50k→30k). Обоснование:
  «расширение anti-DS опций за пределы Lancer».
- **18-C**: Принять дизайн Ion Cannon (shield=500). Это новая фича, не фикс.
- **18-D**: Принять rapidfire Battleship→DS ×2 и SD→DS ×3.
- **18-E**: Принять совместимость с oxsar2-java JAR (конфиги синхронизированы?).

---

## Что НЕ делаем в этом плане

- Не трогаем боевой движок (`battle/engine.go`) — только YAML
- Не меняем экономические формулы (`economy/formulas.go`)
- Не добавляем новые типы юнитов
- Не меняем Alien-юниты (AI-баланс отдельно)
- Не меняем балансовые числа без ADR

---

## Связанные механики из плана 17 (блок H)

После реализации плана 18 имеет смысл реализовать из [плана 17](17-gameplay-improvements.md):

- **H3. Attacker-specific stats** — Deathstar и Alien Screen имеют разные `front` в атаке и защите. После ребаланса DS (п. 2.8) эта механика станет ощутимее: `front` 10→9 у DS в атаке немного снижает вес в targeting.
- **H4. Multi-channel attack** — Deathstar и Alien Screen имеют несколько ненулевых каналов атаки. После ребаланса их щита (DS: 50k→30k) точность multi-channel расчёта важнее.

Эти механики не меняют YAML-числа, но делают их более точно отражёнными в бою.
