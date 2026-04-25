---
title: 27 — Глубокая ребалансировка юнитов
date: 2026-04-25
status: in-progress
---

> **Статус 2026-04-25 (итерация 1)**:
> - 27-A (Lancer cost) — **не нужен**. ADR-0004 уже применён, симуляция
>   подтверждает Lancer невыгоден атакующему (exchange 0.05/0.21).
> - 27-B (Shadow anti-DS) — **применён, но НЕ работает в движке**. См.
>   §10 ниже — attack=200 ≤ DS ignoreAttack=500, выстрелы Shadow
>   полностью поглощаются щитом DS.
> - 27-C / 27-D / 27-E — отложены до следующей итерации.
>
> **Статус 2026-04-25 (итерация 2 — глубокая)**:
> - Расширен battle-sim: 35+ сценариев (см. §9).
> - Найдены новые перекосы: DS god-tier, Bomber-роль не работает,
>   Strong Fighter без ниши, Shadow ADR-0007 нерабочий, defense
>   слаба для endgame, Lancer-spam vs lite-flот всё ещё работает.
> - Конкретные предложения — §10–§13.
>
> **Статус 2026-04-25 (итерация 2 — ✅ ВНЕДРЕНА)**:
> - 27-F (Shadow attack 520) — ✅ done. shadow-mass-vs-ds: 0.00 → 6.67.
> - 27-G (Bomber RF vs Gauss/Plasma) — ✅ done. bomber-vs-plasma: 0.64 → 1.44.
> - 27-H (DS ignoreAttack /150) — ❌ отвергнуто (слишком инвазивно).
> - 27-I (SF rebalance) — ✅ done. sf-vs-lf: 1.57 → 2.05.
> - 27-J (Lancer fuel/speed) — ✅ done (фуэль/скорость = миссионные параметры).
> - 27-K (Defense shell ×1.5) — ✅ done.
> - 27-L (SSat shell+cost) — ✅ done.
> - 27-N..T (front-tweaks) — ✅ done. lancer-bs-combo: 0.22 → 0.79.
> - 27-U (удалить per-unit ballistics/masking) — ✅ done.
> - **ADR-0008** — единый ADR на всю итерацию.

# План 27: Глубокая ребалансировка юнитов

**Контекст**: после плана 18 (порт rapidfire из legacy) и плана 21 (фиксы)
+ упрощения боевого движка до scalar Attack/Shield (2026-04-25) изменился
характер боя. Нужна новая итерация анализа: какие юниты сейчас сверхсильны,
какие лишние, и какие интересные изменения параметров (включая front,
ballistics, masking, fuel) могут оживить геймплей.

**Принципы**:
- Паритет с legacy oxsar2/Java JAR **больше не является целью**
  (отменено 2026-04-25). Балансные изменения принимаются по симуляции
  и геймдизайну, ADR «отклонение от legacy» не нужен.
- Legacy остаётся **референсом** (можно сравнивать), не источником истины.
- Симулятор `backend/cmd/tools/battle-sim/` — обязательное
  подтверждение каждого предложения.

---

## 1. Текущее состояние (baseline 2026-04-25)

### 1.1 Симуляция всех сценариев

```
=== lancer-vs-cruiser ===     1000 Lancer (110M) vs 1000 Cruiser (29M)
  defender wins 100%, exchange 0.05 — Cruiser рвёт Lancer (rapidfire×35 работает)

=== lancer-vs-mixed ===       500 Lancer (55M) vs 200 LF+100 Cru+50 BS (6.7M)
  attacker wins 100%, exchange 0.21 — Lancer побеждает дёшево

=== ds-vs-lancer ===          1 DS (10M) vs 300 Lancer (33M)
  draw 100%, exchange 3.30 — взаимоистребление, Lancer теряет всё

=== ds-vs-bs-fleet ===        1 DS vs 200 BS (12M)
  draw 100%, def loss 90% — DS не теряет ничего, выживает

=== ds-vs-bs-sd-fleet ===     1 DS vs 100 BS+50 SD (12.25M)
  draw 100%, def loss 80% — DS неуязвим даже к SD

=== bomber-vs-rl ===          100 Bomber (9M) vs 3000 RL (6M)
  attacker wins 100%, exchange 16.7 — Bomber как anti-defense

=== cruiser-vs-rl ===         200 Cruiser (5.8M) vs 2000 RL (4M)
  attacker wins 100%, exchange 8.6 — Cruiser хорош vs defense

=== mixed-vs-mixed ===        500 LF+200 Cru+50 BS vs 1000 RL+300 LL+50 Plasma
  draw 100%, exchange 2.12 — атакующий проседает
```

### 1.2 Главные выводы

**Что работает:**
- Bomber — anti-defense специалист, exchange 16.7 ✅
- Cruiser × Lancer — rapidfire×35 убивает Lancer-spam в чистом виде ✅
- DS × Lancer — взаимоистребление, не доминирует ✅
- DS contrably через BS+SD — DS застревает на 80% потерь защитника ✅

**Что не работает:**
- **DS неубиваем мирным флотом** — даже 200 BS не наносят DS урона (он лечится регеном щита). **DS всё ещё топ-юнит endgame**.
- **Lancer vs mixed flotilla** — exchange 0.21 — Lancer стоит 110k каждый, но эффективен против лёгкого флота.
- **Solar Satellite (39)** — пассивный юнит без атаки, 2 щита. Крайне дешёвая «ловушка» для DS×1250 — потенциальный эксплойт защиты.
- **Frigate** — atk/1k=8.2, ниже медианы; legacy-роль (anti-cruiser+anti-BS) работает только с rapidfire.
- **Shadow Ship** — даже с rapidfire-портом остаётся слабым (attack=30) — нужна функциональная роль.
- **Solar Satellite vs DS×1250** — 1250 SSat могут «занять» урон DS на раунд, теоретически.

### 1.3 Аномальные метрики

| Юнит | atk/1k | shell/1k | atk/Shield | Заметки |
|---|---:|---:|---:|---|
| Lancer | **220** | 400 | 27.5 | atk слишком высокий относительно щита |
| Bomber | 10 | 833 | 1.6 | сбалансирован за счёт rapidfire |
| Battleship | 16.7 | 1000 | 5 | флагман флота |
| Frigate | 8.2 | 823 | 1.75 | низкая атака, но высокий щит |
| Death Star | 20 | 900 | 4 | endgame-каноничный |
| Shadow | 6 | 800 | 1 | декоративный |

---

## 2. Анализ параметров — что можно крутить

### 2.1 Front (приоритет цели)

Сейчас почти у всех юнитов `front=10`. Исключения:
- Lancer: 8 — реже выбирается как цель (`weight = 2^front × N`)
- Star Destroyer attacker_front: 9 — атакует раньше при нападении

**Возможности:**
- **Frigate front=11** — становится «приоритетной целью», т.е. весь огонь идёт на него первым. Ролевая идея: «корвет-щит флота», поглощает урон.
- **Bomber front=7** — становится сложнее попасть, как «глубокий бомбер». Но это перегиб: и так силён против defense.
- **Shadow Ship front=6** — сильно реже попадается, синергия с masking. Stealth-разведка.

**Рекомендация**: оставить как в legacy. Манипуляции с front могут давать неинтуитивные результаты (8 вместо 10 — это ×4 меньше, не «чуть-чуть»).

### 2.2 Ballistics / Masking

Сейчас активны только:
- Death Star: ballistics=4
- Shadow Ship: masking=4, ballistics=3
- Strong Laser (45): masking=2
- Bomber: ballistics=2

**Возможности:**
- **Star Destroyer ballistics=2** — symmetria с DS, представляет «средний капитал-флот».
- **Frigate masking=2** — даёт ему уклонение, оправдывает низкий attack/1k.
- **Recycler ballistics=1** — снижает потери при сборе обломков с боем.

**Рекомендация**: дать Frigate `masking=2` — экспериментально. Это
оправдает его низкий attack и даст роль «выживальщика». ADR.

### 2.3 Fuel / Speed

- **Lancer speed=8000, fuel=100** — **аномально дёшев в полёте** для своей силы. Можно поднять fuel до 250 (как у Cruiser).
- **Death Star fuel=1** — символический; speed=100 ограничивает применение, это уже балансир.
- **Frigate fuel=250, speed=10000** — нормально.
- **Bomber fuel=1000** — самый дорогой по fuel, защита от спама.

**Рекомендация**: Lancer fuel 100 → 250. Это незаметный нерф, но
ограничивает возможность отправлять Lancer-волны на дальние цели.

### 2.4 Shell / Shield

- **DS shield=50000** → ignoreAttack=500. Сейчас все юниты с attack>500 пробивают (BS, SD, Lancer, Bomber, Cruiser >400 не пробивает).
- Понизить DS shield 50k → 30k включит Cruiser в список пробивающих. Сейчас в плане 18 это отброшено как «не нужно».

**Рекомендация**: оставить. DS уже балансируется через rapidfire 102→42×3.

### 2.5 Cargo

Чисто экономический параметр. Lancer cargo=200 — символический (это файтер).
DS cargo=1000000 — позволяет ему быть транспортом-флагманом.

**Рекомендация**: не трогаем.

---

## 3. Юниты под вопросом

### 3.1 Solar Satellite (39) — дисбаланс защиты

**Проблема**: 2 щита, 2000 shell, цена 2000 Si + 500 H = 2.5k metal-eq.
DS имеет rapidfire ×1250 vs SSat — это значит, что 1 DS убивает 1250 SSat
за раунд. Но SSat **на планете**, и если поставить 100k SSat, это обойдётся
в 250M ресурсов. Тогда DS убивает 1250×6=7500 SSat за бой = 18.75M.
Защитник «крадёт» 230M ресурсов из DS-атаки (DS делает выстрелы по SSat
вместо настоящей обороны).

**Симуляция**: запустить `1 DS vs 50000 SSat` и проверить.

**Решение**: либо повысить shell SSat (4000+), либо снизить DS×1250 до
×500 (legacy число другое — проверить).

### 3.2 Lancer (102) — всё ещё проблемный

**Симуляция (lancer-vs-mixed)**: exchange 0.21 — атакующий проигрывает,
но количественно: 500 Lancer (55M) теряют 32.5M, защитник теряет 6.7M.
Lancer как атакующий — выгоден.

**Решение**: план 21 блок A — нерф стоимости. Поднять цену с
60k+15k+35k=110k metal-eq до 80k+20k+50k=150k. Это вернёт Lancer в
один диапазон с другими капитал-юнитами.

ADR-кандидат: **поднять цену Lancer на 40%** до баланса.

### 3.3 Shadow Ship (325) — мало пользы

**Сейчас**: attack=30, shield=30, shell=4000, masking=4. Дешёвый разведчик
с rapidfire против большого флота.

**Проблема**: при attack=30 даже массовая Shadow-армия не наносит ощутимого
урона. Шпионаж лучше делать через Espionage Sensor (probe).

**Идея**: дать Shadow роль anti-DS-стелса. Attack 30 → 200, masking 4 → 5.
Тогда 100 Shadow (500k metal-eq) с rapidfire×70 vs DS = 7000 виртуальных
выстрелов; даже с masking-промахами DS будет получать урон. Это
**альтернативный анти-DS** к BS+SD.

ADR: **Shadow attack 30 → 200**.

### 3.4 Star Destroyer (41) — есть, но без роли

**Сейчас**: attack 2000, shell 110k, cost 60+50+15=125k metal-eq.
По метрикам — норм (atk/1k=16). По gameplay — дублирует Battleship,
просто крупнее.

**Идея**: дать SD `attacker_ballistics=2` (есть в legacy через
`attacker_ballistics`-поле — проверить). Это делает SD сильнее в
атаке против stealth-целей.

### 3.5 Frigate (35) — без чёткой роли

**Сейчас**: attack 700, shield 400, shell 70k, rapidfire ×3-7 против
лёгкого флота. Дорогой (85k metal-eq) и слабый (atk/1k=8.2).

**Идея**: добавить `masking=2`. Frigate становится «уклонист-средний
корабль», который сложнее попасть. Atk/1k=8 + ускользание = ниша.

---

## 4. Геймплейные предложения

### 4.1 Иерархия флота (после изменений)

```
Tier 1 (early):   LF, SF, Cru        — массовый флот
Tier 2 (mid):     BS, Frigate, Bmb   — капитал, anti-defense, anti-fleet
Tier 3 (late):    SD, Lancer         — флагманы
Tier 4 (endgame): DS, Shadow(buff)   — стратегические юниты
```

### 4.2 Stealth-мета через Shadow buff

Если Shadow получит attack=200 + masking=5:
- **Stealth-флот**: 50 Shadow (250k metal-eq) — невидимый ударный кулак
- **Контр-DS**: Shadow×70 vs DS — оживает
- **Шпионаж + удар**: один тип юнита для двух ролей

Это **интересная** механика — даёт второй путь в endgame.

### 4.3 Frigate с masking — выживальщик

Frigate masking=2 + low attack/1k = сильный «средний» корабль для тех,
кто не хочет рисковать. Защитный флот.

### 4.4 SD с attacker_ballistics

SD на атаке имеет +2 ballistics → пробивает masking-щиты Shadow и DS.
Дает роль «штурмового флагмана». Совпадает с тем, что в legacy у DS
есть `attacker_*` поля — мы их частично игнорируем.

---

## 5. План реализации

### Фаза 1: Quick fixes

| Задача | Изменение | Файл |
|---|---|---|
| 1.1 | Пересмотреть DS → SSat RF×1250 (возможно нужен меньший множитель) | `configs/rapidfire.yml` |
| 1.2 | Включить `attacker_*` поля в боевой движок (если нужно) | `backend/internal/battle/` |
| 1.3 | Симуляция: SSat-эксплойт (50000 SSat vs 1 DS) | sim |

### Фаза 2: ADR-предложения

| ADR | Предложение | Обоснование | Симуляция |
|---|---|---|---|
| 27-A | Lancer cost ×1.4 (60/15/35 → 84/21/49) | Нерф Lancer-spam (BA-002) | lancer-vs-mixed exchange 0.21 → ~0.5 |
| 27-B | Shadow attack 30 → 200, masking 4 → 5 | Дать Shadow роль anti-DS-стелса | Shadow×70 vs DS должен наносить урон |
| 27-C | Frigate masking +2 | Защитная ниша, оправдание low atk/1k | frigate-vs-cruiser exchange улучшится |
| 27-D | Lancer fuel 100 → 250 | Ограничение spam-волн | Не для боя — для миссий |
| 27-E | DS shield 50k → 40k (мягкая версия плана 18 п.4.4) | Расширить anti-DS-мету, включить Cruiser | ds-vs-cruiser-mass |

### Фаза 3: Симуляция всех ADR вместе

Прогнать все сценарии после применения каждого ADR и совместно. Цель:
exchange ratio в [0.8, 2.5] для основных pvp-боёв.

### Фаза 4: Wiki + dev-log

- Обновить `docs/wiki/ru/combat/index.md` — секция «Тактика» с новой
  ролью Shadow и Frigate.
- Обновить `docs/wiki/ru/ships/index.md` — статы и роли.
- Записать в `docs/ui/dev-log.md` новую итерацию.

---

## 6. Что НЕ делаем

- Не возвращаем 3-канальную систему (план 27 — про числа, не про
  механику; 3 канала — отдельная инициатива в memory).
- Не меняем стоимости основных корпусов (BS, Cruiser, SD) — там legacy
  проверен.
- Не меняем формулу боя — только параметры.
- Не трогаем экономику ресурсов.

---

## 7. Геймплейная оценка

**Насколько это интересно:**

1. **Lancer cost nerf (27-A)** — обязательно, иначе мета сводится к
   Lancer-spam. Польза: высокая.
2. **Shadow buff (27-B)** — даёт **второй путь** в endgame (stealth vs
   массовый флот). Польза: высокая, интересно.
3. **Frigate masking (27-C)** — даёт middle-tier корабль с уникальной
   ролью. Польза: средняя — не критично, но углубляет.
4. **Lancer fuel (27-D)** — мелкий нерф. Польза: низкая.
5. **DS shield (27-E)** — ослабляет endgame-доминирование, но рискует
   ослабить DS как «угрозу всему живому». Польза: спорная.

**Рекомендуемый минимальный набор**: 27-A + 27-B. Это закроет дыру с
Lancer и оживит Shadow как полноценный юнит.

**Расширенный набор**: + 27-C (Frigate masking).

**Оптимистичный**: всё кроме 27-E (риск нарушить ощущение endgame).

---

## 8. Связь с другими планами

- **План 18** (rapidfire) — закрыт, на нём базируемся.
- **План 21 блок A** (Lancer cost) — поглощается этим планом (ADR 27-A).
- **План 24** (AI-боты) — после ребаланса нужно перетренировать ботов
  на новых статах.
- **Memory: 3-канальная система** — параллельный путь, не пересекается
  с этим планом.

---

## 9. Риски

1. **Тесты battle/golden** — обновить после каждого изменения.
2. **Live игроки** — если уже играют на текущем балансе, ребаланс
   потребует объявления и компенсации.
3. **Symptom-fix vs root cause**: некоторые проблемы (Lancer-spam) —
   следствие исходной идеи юнита («премиум корабль»). Нерф цены — это
   симптом-фикс. Корень: убрать Lancer вовсе или превратить в
   special-юнит (требование исследования + лимит на штуку).

---

## 10. Итерация 2 — глубокая ребалансировка (2026-04-25)

Расширил battle-sim до 35+ сценариев, прокачал передачу `front` из
construction.yml в movement (в исходной версии у всех юнитов был
front=0, веса целей считались равными). Прогон по `--all --runs=20`
вскрыл ряд проблем, не видимых в первой итерации.

### 10.1 Расширенный battle-sim — что добавлено

Новые сценарии (`backend/cmd/tools/battle-sim/main.go`):

- **SSat-эксплойт**: `ssat-trap-vs-ds`, `ssat-trap-vs-ds-fleet`
- **Frigate-роль**: `frigate-vs-cruiser`, `frigate-vs-bs`, `frigate-vs-sf`
- **Shadow Ship**: `shadow-vs-ds`, `shadow-mass-vs-ds`, `shadow-vs-mixed`
- **Star Destroyer**: `sd-vs-bs`, `sd-vs-frigate`
- **Mass-fleet vs DS**: `ds-vs-bs-mass`, `ds-vs-bomber-mass`,
  `ds-vs-plasma-mass`, `ds-vs-gauss-wall`
- **Defense scaling**: `lf-mass-vs-rl-ll`, `bs-vs-plasma`,
  `bomber-vs-plasma`, `bomber-vs-gauss`, `plasma-wall-vs-bs-mass`
- **Strong Fighter роль**: `sf-vs-rl`, `sf-vs-lf`
- **Lancer**: `lancer-vs-bs`, `lancer-vs-sf`, `lancer-vs-plasma`
- **Endgame mirrors**: `huge-fleet-vs-huge-fleet`, `ds-fleet-vs-ds-fleet`,
  `ds-fleet-vs-defended-planet`, `ds-as-defense-vs-fleet`
- **Front/shield-coverage**: `small-shield-coverage`, `large-shield-coverage`
- **Soft targets**: `recycler-in-fleet`, `transport-in-fleet`,
  `esensor-in-fleet`, `esensor-stealth-vs-mixed`, `colony-ship-escort`
- **Combo**: `lancer-bs-combo-vs-mixed`, `trio-vs-defense`,
  `lf-swarm-vs-cruiser`

### 10.2 Сводка результатов (--runs=20)

| Сценарий | Atk | Def | Atk wins | Exchange | Заметка |
|---|---:|---:|---:|---:|---|
| lancer-vs-mixed | 55M | 6.7M | 100% | **0.21** | Lancer-spam vs lite — атакующий пробивает |
| frigate-vs-cruiser | 17M | 17.4M | 0% | 3.27 | Frigate имеет роль |
| frigate-vs-bs | 17M | 12M | 0% | **8.21** | Frigate **сильнее**, чем ожидалось |
| frigate-vs-sf | 8.5M | 10M | 0% | 4.03 | Frigate как anti-fighter |
| shadow-vs-ds | 0.5M | 10M | 0% | **0.00** | ADR-0007 не работает (см. §10.4) |
| shadow-mass-vs-ds | 5M | 10M | 0% | **0.00** | То же |
| shadow-vs-mixed | 2.5M | 8.25M | 0% | 0.11 | Shadow умирает за 1 раунд |
| sd-vs-bs | 6.25M | 6M | 0% | 1.28 | SD нормально |
| sd-vs-frigate | 6.25M | 8.5M | 0% | 4.42 | SD анти-Frigate (RF×2) |
| ds-vs-bs-mass | 10M | 60M | 0% | n/a (DS=0) | DS не теряет при 6× ресурсов оппонента |
| ds-vs-bomber-mass | 18M | 10M | 0% | **0.00** | Bomber бесполезен против DS |
| ds-vs-plasma-mass | 10M | 6.5M | 0% | n/a | Plasma×2 RF слабо работает |
| ds-vs-gauss-wall | 10M | 7.4M | 100% | n/a | DS пробивает Gauss-wall (нет RF у Gauss vs DS) |
| bs-vs-plasma | 12M | 6.5M | 0% | 1.81 | BS эффективнее против Plasma, чем... |
| bomber-vs-plasma | 9M | 6.5M | 0% | **0.64** | ...Bomber. Bomber против Plasma теряет ресурсы! |
| bomber-vs-gauss | 9M | 3.7M | 0% | 1.13 | Bomber бесполезен и против Gauss |
| sf-vs-rl | 5M | 4M | 0% | **0.43** | SF проигрывает дёшевой defense |
| sf-vs-lf | 2.5M | 2.4M | 0% | 1.57 | SF чуть лучше LF при равной cost |
| lancer-vs-bs | 55M | 60M | 0% | 0.16 | Lancer слабее BS при равной цене |
| lancer-vs-sf | 11M | 10M | 0% | **0.05** | Lancer катастрофически плох vs lite |
| lancer-vs-plasma | 22M | 13M | 0% | 0.39 | Lancer не пробивает Plasma stack |
| recycler-in-fleet | 7.8M | 5.8M | 0% | 2.07 | Recycler — soft target, рискованно |
| transport-in-fleet | 3.2M | 5.8M | 0% | **0.01** | Транспорты в боевом флоте — самоубийство |
| huge-fleet-vs-huge-fleet | 18M | 18M | 0% | 1.00 | Mirror корректен |
| ds-fleet-vs-ds-fleet | 56M | 56M | 0% | 1.00 | Mirror корректен |
| ds-fleet-vs-defended-planet | 106M | 30M | 0% | **3.51** | Defense слаба — атакующий теряет 3% при 36% защитника |
| ds-as-defense-vs-fleet | 55M | 50M | 0% | **0.00** | DS **неубиваемая защита** — флот теряет 68%, DS теряет 0% |
| lf-mass-vs-rl-ll | 20M | 12M | 0% | **0.32** | Cheap defense эффективнее cheap fleet |
| lf-swarm-vs-cruiser | 20M | 29M | 0% | 0.06 | LF-swarm бесполезен против Cruiser-RF×6 |
| ssat-trap-vs-ds | 10M | 125M | 0% | n/a | DS не пробивает 50k SSat (потолок) |
| esensor-in-fleet | 6.05M | 5.8M | 0% | 1.66 | ESensor отвлекает, но переживает |
| esensor-stealth-vs-mixed | 1M | 2.9M | 0% | 0.00 | Probe-spam в бою бесполезен |
| trio-vs-defense | 7.9M | 8.9M | 0% | 1.67 | Mix атакующего против mix-defense — нормально |
| lancer-bs-combo-vs-mixed | 23M | 18M | 0% | 0.22 | Lancer-в-комбо тащит вниз весь флот |
| plasma-wall-vs-bs-mass | 60M | 26M | 0% | **2.43** | Plasma-wall эффективна против BS |

**Жирным** — критические перекосы.

### 10.3 Ключевые проблемы

#### A. Death Star — god-tier endgame

- `ds-as-defense-vs-fleet` (5 DS vs 500 BS + 200 SD): **DS теряет 0%**
  при 68% потерь у атакующего. Реалистичный endgame-mirror (`ds-fleet`)
  показывает 10.7% — это норма, но **только при равных DS**.
- `ds-fleet-vs-defended-planet` (10 DS vs 200 Plasma + 100 Gauss + 5 LS):
  атакующий теряет 3% (3.12M из 106M), защитник — 36%. Реалистичная
  планетарная защита **не сдерживает** DS-флот.
- `ds-vs-bs-mass` (1 DS vs 1000 BS = 60M ресурсов): DS **не теряет
  ничего**. То есть если у защитника есть 1 DS, для его убийства
  атакующему нужно вдвое-втрое больше ресурсов в BS, **и всё равно DS
  не падает**.
- Корень — DS shield=50000 регенится каждый раунд, и BS attack=1000
  > ignoreAttack=500, но shieldDamageFactor никогда не становится 0
  до конца раунда (regen).

**Контр существующий**:
- Lancer (RF×3) — покрыт ADR-0004, теперь невыгоден.
- BS+SD-mix без RF vs DS — eat 80% потерь и не убивают (см. baseline).
- Bomber — RF нет vs DS, attack=900 не хватает.
- Cruiser — attack=400 < ignoreAttack=500.
- Plasma — RF×2, attack=3000 пробивает, но нужно много (50 Plasma →
  exchange 0.12 потери защитника).

#### B. Bomber-роль не работает

| Сценарий | Exchange | Вывод |
|---|---:|---|
| bomber-vs-rl (legacy) | **16.67** | Anti-RL ✅ |
| bomber-vs-plasma | **0.64** | Bomber **проигрывает** Plasma |
| bomber-vs-gauss | 1.13 | Чуть лучше, чем размен 1:1 |
| ds-vs-bomber-mass | **0.00** | Bomber бесполезен против DS |

Bomber имеет RF только против RL (43), LL (44), SL (45), IG (46). Но
**не имеет RF против Gauss, Plasma и DS**. Его attack=900 уверенно
пробивает все defense (Plasma shield=300, ignoreAttack=3), shell=75k
держит много, но **низкая скорость 4000 + cost 90k metal-eq делают
его дороже-неэффективнее BS** для anti-defense.

Корень: дизайн Bomber — «штурм первого этапа» (RL+LL), а в endgame-
defense (Gauss+Plasma) Bomber проигрывает обычным капитал-кораблям.

#### C. Strong Fighter — без ниши

| Сценарий | Exchange | Вывод |
|---|---:|---|
| sf-vs-lf | 1.57 | SF чуть лучше LF при равной metal-eq |
| sf-vs-rl | **0.43** | SF проигрывает RL |
| frigate-vs-sf | 4.03 | Frigate рвёт SF |
| lancer-vs-sf | 20.0 (def loss / atk loss) | SF дёшево умирает от Lancer |

Strong Fighter — middle-tier между LF и Cruiser. По цене 10k metal-eq
он 2× LF (5k), но эффективность только 1.57× — фактически **trap-юнит**.
Нет уникальной роли: вместо SF лучше копить на Cruiser.

#### D. Shadow ADR-0007 не работает

Ключевая ошибка моделирования в плане 27-B:

```
DS shield = 50000
DS ignoreAttack = baseShield / 100 = 500
Shadow attack = 200 (после ADR-0007)
200 ≤ 500 → Shadow attack полностью поглощается щитом DS
```

В коде `applyShots` (`backend/internal/battle/engine.go:476`):
```
if attack > 0 && attack <= ignoreAttack {
    pool := attack * float64(shots)
    if pool > target.turnShield {
        pool = target.turnShield
    }
    target.turnShield -= pool
    return  // Урон в shell НЕ доходит
}
```

Подтверждено симуляцией:
- Shadow attack=200: 1000 Shadow vs 1 DS — **DS не теряет ничего**.
- Shadow attack=350: то же.
- Shadow attack=600 (>500): 1000 Shadow vs 1 DS — exchange 6.67 ✅,
  100 Shadow vs 1 DS — всё ещё проигрыш (мало массы).

**Решение**: либо attack 200→**600+** (полная переработка ADR-0007),
либо отказаться от anti-DS-роли Shadow.

Обоснование ratio Shadow:
- Cost 5k metal-eq (1k+3k+1k) = очень дёшево.
- 1000 Shadow = 5M-eq vs 1 DS (10M-eq) → атакующий должен платить
  ~50% от DS, чтобы убить DS. Это ниже Lancer-cost-ratio (RF×3 = 33k
  / 10M = 0.33%).
- Если attack=600, 1000 Shadow убивают DS за 4 раунда с 30% потерь
  атакующего → exchange 6.67. **Чуть слишком сильно** (Shadow becomes
  god-anti-DS).
- Целевой exchange — 1.5–3.0 (как в ADR-0007). Это значит attack
  ~500–550, либо повысить cost Shadow, либо снизить RF×70 → ×30–50.

#### E. Lancer-spam vs lite — сохранилась дыра

`lancer-vs-mixed`: 500 Lancer (55M) vs 200 LF + 100 Cru + 50 BS (6.7M):
- Атакующий побеждает 100% времени.
- Защитник теряет всё (6.7M).
- Атакующий теряет 32.5M из 55M (59%).
- Exchange 0.21 — **атакующий невыгодно тратит ресурсы**, но
  **флот защитника уничтожен** → flot уничтожен + raid loot >
  потерь Lancer'а.

Это значит: Lancer **используется как рейдер**, а не как боевой юнит.
Размер Lancer-флота 55M — это endgame-цена, и атаковать на нём 6.7M
лёгкого защитника — всё равно невыгодно по чистым потерям, **но**:
- Lancer — single-юнит, легко прицелиться.
- Защитник теряет timer на восстановление (1-2 дня production).
- При raid'е > 5M атакующий получает >5M loot → суммарно в плюсе.

Корень: **Lancer привлекателен из-за speed=8000 + fuel=100 + RF×3 vs
DS — это премиум-юнит для быстрых рейдов**. Нерфить нужно либо speed,
либо fuel, либо роль (например, ограничить max количество).

#### F. Defense слаба

- `ds-fleet-vs-defended-planet`: атакующий теряет 3%, защитник 36%.
  Реалистичная planet-defense **не сдерживает** атакующий флот при
  паритете 3:1 (atkM/defM = 106M/30M = 3.5×).
- `lf-mass-vs-rl-ll`: единственный сценарий, где защита эффективнее
  атаки (exchange 0.32).
- `plasma-wall-vs-bs-mass`: защита на 26M удерживает 60M-флот с
  exchange 2.43 (защитник теряет 7M, атакующий 3M).

Defense balance OK для **сопоставимых** atk/def ratio, но для
**превосходящего** атакующего флота defense служит лишь налогом, не
барьером.

#### G. Shield/Front-юниты как «ловушки»

- `small-shield-coverage` (front=16): 1 SS перетягивает почти весь
  огонь (weight = 2^16 = 65536, vs 500 RL = 500 × 2^10 = 512k).
  Доля SS в выстрелах ≈ 11%, но при shell=20k он эпически быстро
  погибает (200 Cruiser × 400 attack × 6 раундов = 480k power, SS
  shell=20k → SS убивается за 1 раунд) → **front=16 у SS не оправдан**:
  он умирает мгновенно и не успевает «принять урон».
- `large-shield-coverage` (front=17): то же. LS shell=100k держится
  один раунд.

Решение: shield-юниты должны иметь **либо больше shell**, либо быть
**полу-неуязвимыми** через ignoreAttack=0 (что и реализовано через
exception для UnitID 49/50). Но это значит, что shield-юниты — целеуказатели,
а не tank'ы. **Это известный паттерн OGame**, оставляем.

#### H. Cargo / Speed / Fuel — не учитываются в бою

В battle-движке используются: attack, shield, shell, front, ballistics
(Tech), masking (Tech), rapidfire. **Не используются**:
- `cargo` — экономика рейдов (loot capacity).
- `speed` — навигация миссий (mission travel time).
- `fuel` — расход на миссию.

Это корректно. Но balance-эффекты этих характеристик есть:
- Lancer: speed=8000 + fuel=100 → **дёшево гонять на дальние цели**.
- Bomber: speed=4000 + fuel=1000 → **медленный и дорогой**, оправдан
  ролью.
- DS: speed=100 + fuel=1 → защитник, не рейдер.
- Recycler: speed=2000 + fuel=300 → собиратель обломков.
- Shadow: speed=13000 + fuel=35 → **быстрейший корабль с RF против
  всего** + дёшев → потенциальный raider.

### 10.4 Подтверждение через рекомендуемое изменение Shadow

Прогнал sed-override `ships.yml` Shadow attack: 200 → 350 → 600:

| attack | shadow-mass-vs-ds | shadow-vs-mixed |
|---:|---|---|
| 200 (сейчас) | exchange 0.00, def loss 0% | exchange 0.11 |
| 350 (≤ ignoreAttack DS) | exchange 0.00, def loss 0% | exchange 0.20 |
| 600 (> ignoreAttack DS) | **exchange 6.67**, def loss 100% | exchange 0.36 |

То есть Shadow ADR-0007 **физически требует** attack ≥ 501 для работы.
При attack=600 anti-DS-роль работает (даже **слишком сильно** — 5M
Shadow убивает 10M DS), но vs обычный флот Shadow всё равно хрупкий.

---

## 11. Предлагаемые изменения (итерация 2)

### 11.1 ADR-кандидаты (по приоритету)

| ADR | Юнит | Изменение | Цель | Статус |
|---|---|---|---|---|
| **27-F** | Shadow Ship | attack 200 → **520** | Починить ADR-0007 (включить anti-DS) | high prio |
| **27-G** | Bomber | RF добавить vs Gauss (47) ×5, Plasma (48) ×3 | Восстановить anti-defense-роль | high prio |
| **27-H** | Death Star | ignoreAttack: shield/100 → shield/**150** | Снизить порог пробивания (Cruiser attack=400 не пробивает, но если ÷150, то порог=333 → пробивает) — даёт BS+Cruiser-ammo против DS | medium prio |
| **27-I** | Strong Fighter | cost ↓ 6k+4k → 4k+2k = 6k metal-eq, attack 150 → 120 | Дешевле LF×2, нишевая роль anti-LF без overkill | medium prio |
| **27-J** | Lancer Ship | fuel 100 → **400**, speed 8000 → **6000** | Нерф рейдинг-возможностей (атаки на дальние цели становятся дорогими по fuel) | medium prio |
| **27-K** | Defense | Все defense shell ×1.5 (RL 2k→3k, LL 2k→3k, SL 8k→12k, IG 8k→12k) | Усилить planet defense — не должно быть 3% потерь у 3:1 атакующего | medium prio |
| **27-L** | Solar Satellite | shell 2000 → **5000**, cost 2k+0.5k → **3k+1k** (+50% cost) | Закрыть SSat-эксплойт в качестве «poison pill» против DS | low prio |
| **27-M** | Frigate | без изменений | **Подтверждено сильным** (exchange 8.21 vs BS) — не трогаем | n/a |

### 11.2 Симуляция целей после ADR

| Сценарий | Текущ. exchange | Цель |
|---|---:|---:|
| shadow-mass-vs-ds | 0.00 | 1.5–3.0 |
| bomber-vs-plasma | 0.64 | 1.5–2.5 |
| ds-fleet-vs-defended-planet | atk loss 3% | atk loss 8–15% |
| lf-mass-vs-rl-ll | 0.32 | 0.4–0.6 |
| ssat-trap-vs-ds | def loss 37% (15 раундов) | def loss > 60% |

### 11.3 Что **не** трогаем

- **Frigate** — exchange 8.21 vs BS, 4.03 vs SF — **сильный нишевый
  юнит**. Возможный нерф shell 70k → 50k обсуждается, но не критично.
- **Battleship** — флагман, exchange 1.81 vs Plasma, 1.0 mirror — норм.
- **Cruiser** — anti-LF (RF×6), exchange 8.62 vs RL, 0.06 mirror против
  LF-swarm — норм.
- **Light Fighter** — массовый, mirror и exchange OK.
- **Star Destroyer** — exchange 1.28 vs BS, 4.42 vs Frigate (RF×2) —
  норм.
- **Recycler / Transport** — soft targets, **по дизайну**.

### 11.4 Дополнительные идеи (не для текущей итерации)

1. **Cap Lancer per planet** — лимит 50–100 Lancer/планета. Решает
   проблему рейдов (не блокирует прокачку).
2. **Bomber speed 4000 → 6000** — сделать конкурентоспособным с BS
   (10000) для anti-defense рейдов.
3. **DS speed 100 → 50** — ещё больше зафиксировать DS как защитник.
4. **Star Destroyer attacker_ballistics** — добавить поле в движок,
   SD получит ballistics=2 в атаке → пробивает masking.
5. **Frigate masking +1** — обсуждалось в 27-C, но Frigate сильный
   и без этого. Отвергаем.
6. **DS cost ×1.5** — как «нерф через стоимость», но это меняет
   накопление endgame-юнитов и требует большего ребаланса экономики.

---

## 12. План внедрения (итерация 2)

### Фаза 1: Quick wins (без ADR, чисто симуляции)

- [ ] Добавить рекомендуемые сценарии в `cmd/tools/battle-sim/main.go`
  как permanent (закоммитить расширение).
- [ ] Прогнать `--all --runs=50` и зафиксировать baseline в
  `docs/balance/audit.md`.

### Фаза 2: ADR-кандидаты с симуляцией

Каждое изменение — отдельный ADR, проверяется симулятором.

- [ ] **ADR-0008** — 27-F (Shadow attack=520). Critical fix для
  ADR-0007.
- [ ] **ADR-0009** — 27-G (Bomber RF vs Gauss/Plasma). Восстанавливает
  anti-defense-роль.
- [ ] **ADR-0010** — 27-H (DS ignoreAttack /150). Открывает anti-DS-
  путь через Cruiser-mass.
- [ ] **ADR-0011** — 27-I (SF rebalance).
- [ ] **ADR-0012** — 27-J (Lancer fuel/speed nerf).
- [ ] **ADR-0013** — 27-K (Defense shell +50%).
- [ ] **ADR-0014** — 27-L (SSat shell+cost). Опционально.

### Фаза 3: Тесты

- [ ] Обновить `backend/internal/battle/testdata/*.json` golden-файлы.
- [ ] Запустить property-based (rapid) tests.
- [ ] Прогнать full battle-sim — exchange-ratio в [0.8, 2.5] для
  основных сценариев.

### Фаза 4: Wiki + dev-log + audit

- [ ] Обновить `docs/wiki/ru/ships/index.md` с новыми статами.
- [ ] Обновить `docs/wiki/ru/combat/index.md` — раздел «Тактика».
- [ ] Записать в `docs/balance/audit.md` найденные дыры.
- [ ] Записать в `docs/ui/dev-log.md` итерацию 2 балансировки.

---

## 13. Геймплейная оценка

**Что станет лучше после внедрения 27-F..L:**

1. **Anti-DS меню** будет включать:
   - Shadow-mass (новое — после 27-F)
   - Cruiser-mass (новое — после 27-H)
   - Plasma-defense (уже работает, RF×2)
   - Lancer (RF×3, ADR-0004 сбалансировал)
   - BS+SD-mix (тяжёлая, дорогая опция)
2. **Bomber** возвращается как универсальный anti-defense (после 27-G).
3. **Strong Fighter** становится дешёвым anti-LF (после 27-I).
4. **Lancer** теряет рейдинг-преимущество (после 27-J), но остаётся
   капитал-юнитом.
5. **Defense** повышается до уровня барьера, не налога (после 27-K).
6. **SSat-эксплойт** закрывается (после 27-L).

**Что не починится:**

- DS-mirror всё равно решается DS-флотом большего размера. Это by-
  design — endgame-юнит.
- Strong Fighter всё равно остаётся «промежуточным» юнитом. Полное
  его удаление — отдельный (более радикальный) ребаланс.
- Recycler/Transport мягкие — дизайн.

**Риски:**

1. **27-H (DS ignoreAttack)** — самое спорное. Открывает Cruiser-mass
   как anti-DS, что меняет endgame-стратегию. Нужна детальная
   симуляция.
2. **27-G (Bomber RF vs Plasma/Gauss)** — Bomber становится
   «универсальный anti-defense», что может сделать его доминирующим
   anti-defense-юнитом. Нужно проверить, не делает ли это RL/LL-стену
   бессмысленной.
3. **27-J (Lancer fuel)** — сильный нерф для существующих игроков с
   Lancer-флотом.

**Минимально-достаточный набор для итерации 2**: 27-F (Shadow fix)
+ 27-G (Bomber RF) + 27-K (Defense buff). Они вместе чинят 3 главных
проблемы (Shadow, Bomber, Defense), не ломая endgame-баланс.

---

## 14. Связь с другими планами и memory

- **План 24** (AI-боты) — после 27-* нужно перетренировать тренировку
  ботов (или хотя бы пересчитать профайл оборонительных стек'ов).
- **3-канальная боевая система** (memory `3channel-combat-idea.md`) —
  параллельный путь, не пересекается. Если 3-канальная будет принята,
  большинство 27-* станет неактуальным.
- **План 18** (rapidfire) — основа для 27-G (добавить RF Bomber→Gauss/Plasma).
- **ADR-0004** (Lancer cost) — сохраняется. 27-J (Lancer fuel/speed)
  идёт в дополнение.
- **ADR-0007** (Shadow anti-DS) — **переписывается** в 27-F (изменение
  параметра attack 200→520).

---

## 15. Открытые вопросы

1. **Должен ли DS быть «непробиваемым endgame»?** Ответ влияет на 27-H.
   Если да — оставить ignoreAttack=shield/100 (текущее), и тогда anti-DS
   только через RF (Lancer/Plasma/Shadow). Если нет — снижать порог.
2. **Strong Fighter — оставить или удалить?** При текущих статах — trap-
   юнит. Если оставлять, нужна сильная роль. Кандидат: дать SF role «anti-
   bomber» (RF×3 vs Bomber, attack 150 → 120, cost ↓).
3. **Solar Satellite — атакующий юнит или энергоисточник?** Сейчас
   гибрид (production + battle). Можно вообще убрать его из боя
   (через `mode=defense_passive` в movement) — radikal но решает.
4. **Bomber speed 4000 vs 6000** — отдельная дискуссия. В legacy
   speed=4000.

---

## 16. Анализ front / ballistics / masking (2026-04-25)

### 16.1 Текущая работа в движке

| Поле | Источник | Использует движок? |
|---|---|---|
| `front` | `construction.yml` per-unit | **Да** (через `unit.Front`, weight = 2^front × N) |
| `ballistics` | `construction.yml` per-unit | **Нет** — игнорируется. Движок берёт только `Side.Tech.Ballistics` |
| `masking` | `construction.yml` per-unit | **Нет** — игнорируется. Движок берёт только `Side.Tech.Masking` |

Это значит:
- DS `ballistics=4`, Lancer `ballistics=3`, Shadow `ballistics=5`/`masking=5`,
  Paladin `ballistics=5`/`masking=1`, Interplanetary Rocket `ballistics=10` —
  **все per-unit значения игнорируются**.
- `Tech.Ballistics`/`Tech.Masking` — это research игрока (на всю сторону),
  работает корректно, но не различает юнитов.

### 16.2 Текущие non-default значения

| Юнит | front | ballistics (per-unit) | masking (per-unit) | Заметка |
|---|---:|---:|---:|---|
| Death Star | 9 | 4 | — | front−1 → ×2 реже целью |
| Lancer Ship | 8 | 3 | — | front−2 → ×4 реже |
| Shadow Ship | 7 | 5 | 5 | stealth-семантика, но per-unit мёртв |
| Star Destroyer | 10 | — | — | без ниши (front=10 как у всех) |
| Bomber | 10 | — | — | без ниши |
| Frigate | 10 | — | — | без ниши |
| Small Shield | 16 | — | — | weight ×64 — гипер-приоритет |
| Large Shield | 17 | — | — | weight ×128 |
| Alien Screen (201) | 15 | — | — | weight ×32 |
| Paladin (alien 202) | 10 | 5 | 1 | per-unit мёртв |
| Corvette (alien 200) | 10 | 2 | — | per-unit мёртв |
| Torpedocarrier (alien 204) | 10 | 4 | — | per-unit мёртв |
| Frigate (alien 203) | 10 | 1 | — | per-unit мёртв |
| Interplanetary Rocket | 10 | 10 | — | per-unit мёртв (но юнит для kind=16, не обычный бой) |
| UNIT_EXCH_SUPPORT_SLOT (106) | 5 | — | — | mode=4 defense, реже целью |

### 16.3 Симуляция чувствительности (battle-sim --front=...)

**Lancer в смешанной атаке (lancer-bs-combo-vs-mixed)**:

| Lancer front | Atk loss % | Exchange |
|---:|---:|---:|
| 6 | 21% | **0.78** |
| 8 (текущий) | 48% | 0.22 |
| 10 (как все) | 53% | 0.12 |

Lancer front=6 **драматически меняет** баланс смешанного флота: Lancer
становится почти невидимым (1/16 целей), огонь идёт на BS/Cru, Lancer
переживает больше → атакующий теряет в 2.5× меньше. **front — мощный
рычаг балансировки**.

**Shadow в смешанной атаке (shadow-bs-mix-vs-cruiser)**:

| Shadow front | Atk loss % | Exchange |
|---:|---:|---:|
| 4 | 4.5% | **2.58** |
| 7 (текущий) | 4.9% | 2.39 |
| 10 | 5.6% | 2.01 |

Эффект меньше, но Shadow front=4 — заметное улучшение анти-цены.
front=7 (текущее) — компромисс.

**Plasma в защите (ds-fleet-vs-defended-planet)**:

| Plasma front | Atk loss % | Exchange |
|---:|---:|---:|
| 10 (текущий) | 2.9% | 3.51 |
| 11 | 2.9% | 3.71 |
| Plasma+Gauss=11 | 2.8% | 3.78 |

Повышать front Plasma/Gauss **анти-улучшение** для защитника:
приоритетные цели умирают первыми, защита теряет огневую мощь раньше.

**Large Shield (large-shield-coverage)**:

| LS front | Atk loss % | Exchange |
|---:|---:|---:|
| 17 (текущий) | 2.2% | 5.20 |
| 12 | 2.1% | 5.86 |
| 10 | 2.1% | 5.88 |

LS front=17 **немного помогает** защитнику (LS тащит часть выстрелов
в первые раунды), но shell=100k всё равно мал — эффект слабый.

### 16.4 Выводы

1. **front работает только в mixed-fleet** — у атакующего из 1 типа
   юнитов front не влияет (защитник всё равно стреляет в единственный
   тип).
2. **Низкий front полезнее высокого** в роли «спрятать дорогой юнит за
   щитом массы». Lancer front=6 даёт +130% эффективности в
   смешанной атаке.
3. **Высокий front «приоритетной цели»** парадоксально вреден
   защитнику: cher-bait умирает первым, лишая защиту огневой мощи.
4. **Per-unit ballistics/masking — мёртвый код в `construction.yml`**.
   Либо включить в движок, либо удалить из конфига, либо переехать
   в ships.yml с явной семантикой.

### 16.5 ADR-кандидаты по front

| ADR | Юнит | Изменение | Цель | Приоритет |
|---|---|---|---|---|
| 27-N | Lancer | front 8 → **6** | Дополнительный нерф (Lancer прячется за BS, ослабляет combo-атаку Lancer+BS) | high |
| 27-O | Bomber | front 10 → **8** | Bomber-уклонист, меньше теряется в смешанной атаке | medium |
| 27-P | Shadow | front 7 → **5** | Усиление stealth-семантики в смешанной атаке | medium |
| 27-Q | Recycler/Transport | front 10 → **6** | Естественная защита для логистики (logistic units не стреляют → не должны быть приоритетной целью) | medium |
| 27-R | Espionage Sensor | front 10 → **5** | Probe не должен «принимать» обычные выстрелы — он шпион | low |
| 27-S | Small/Large Shield | front 16/17 → **13/14** | Менее агрессивный приоритет, чтобы shield не «сгорал» в раунд 1 | low |
| 27-T | Star Destroyer | front 10 → **9** | Дать SD нишу «капитал-средняя цель» (между BS=10 и Lancer=6/8) | low |

### 16.6 ADR-кандидаты по ballistics/masking

| ADR | Действие | Обоснование |
|---|---|---|
| 27-U | Удалить per-unit `ballistics` и `masking` из `construction.yml` | Они не работают, висят как dead code и вводят в заблуждение |
| 27-V (alt) | Включить per-unit ballistics/masking в движок | Расширение `battle.Unit` + `applyMasking` берёт max(per-unit, side-Tech). Усложнение модели, требует тестов |

**Рекомендация**: 27-U (удалить). Если когда-то понадобится stealth-
дифференциация юнитов — добавить отдельным ADR с осмысленной семантикой
(например, через подкласс юнитов «scout/stealth»).

### 16.7 Расширение battle-sim

Добавлен флаг `--front=ship_key=N` для override front без изменения
конфига:

```bash
battle-sim --scenario=lancer-bs-combo-vs-mixed --front=lancer_ship=6
```

Также добавлены сценарии `shadow-bs-mix-vs-cruiser` и
`recycler-bs-mix` для тестирования front в смешанных флотах.

---

## 17. Аудит ролей: матрица 1v1 + сравнение с OGame (2026-04-25)

После применения 27-F..27-U прогнал **матрицу 1v1** для всех combat-юнитов
при равной metal-eq budget (1M, 10M, 100M — стабильны), плюс
**матрицы групп vs групп** (lite/mid/capital/endgame/shadow/lancer
+ defense lite/heavy/mixed). Расширил battle-sim: режимы
`--matrix --matrix-budget=N` и `--groups`.

### 17.1 Матрица 1v1 при 10M metal-eq (exchange = def_loss / atk_loss)

```
                | LF    SF    Cru   BS    Frig  Bomb  SD    Lan   Sha   DS    RL    LL    SL    IG    Gauss Plas
light_fighter   | 1.00  0.26  0.11  0.50  1.08  0.62  0.53  27.5  0.01  def   0.12  0.08  def   def   0.15  0.29
strong_fighter  | 3.84  1.00  2.58  1.72  0.49  2.95  1.83  47.1  12.1  def   0.50  0.37  0.27  def   0.58  1.00
cruiser         | 9.05  0.39  1.00  0.55  0.26  0.93  0.65  26.3  0.11  def   2.45  0.15  0.20  def   0.16  0.37
battle_ship     | 2.01  0.58  1.83  1.00  0.22  1.71  0.96  7.88  10.4  def   0.28  0.21  0.34  0.70  0.35  0.54
frigate         | 0.93  2.03  3.81  4.59  1.00  0.26  0.06  1.97  0.01  def   0.11  0.08  0.14  0.17  0.12  0.19
bomber          | 1.62  0.34  1.07  0.58  3.78  1.00  0.46  18.3  0.01  def   4.70  3.37  2.31  6.58  0.59  1.03
star_destroyer  | 1.89  0.55  1.53  1.04  15.6  2.16  1.00  19.8  0.02  def   0.24  2.50  0.32  0.97  0.34  0.52
lancer_ship     | 0.04  0.02  0.04  0.13  0.51  0.05  0.05  1.00  0.04  def   0.03  0.03  0.05  0.14  0.05  0.07
shadow_ship     | 90.9  0.08  9.18  0.10  104.7 83.2  51.3  22.2  1.00  6.67  1.70  1.25  1.94  4.19  2.15  3.71
death_star      | ATK   ATK   ATK   ATK   ATK   ATK   ATK   ATK   0.15  —     ATK   ATK   ATK   ATK   ATK   ATK
```

(`def` = атакующий проигрывает; `ATK` = атакующий не теряет; `—` = оба сохраняют флот)

### 17.2 Матрица групп при 50M metal-eq

```
atk\def         | lite  mid   cap   end   sha   lan   d-lt  d-hv  d-mx
lite-fleet      | 1.01  0.37  0.85  0.20  0.55  6.27  0.36  0.19  0.48
mid-fleet       | 2.75  1.00  0.88  0.14  2.00  4.34  0.68  0.22  0.86
capital-fleet   | 1.18  1.14  1.00  0.11  0.67  1.31  1.33  0.62  1.50
endgame-fleet   | 5.03  6.94  8.94  1.00  0.80  17.85 1.37  3.32  2.62
shadow-fleet    | 1.81  0.50  1.50  1.25  1.00  0.32  1.40  2.33  1.75
lancer-fleet    | 0.16  0.23  0.76  0.06  3.12  1.00  0.31  0.59  0.38
```

### 17.3 Анализ ролей (нет/есть/доминанта)

| Юнит | Уникальная роль? | Доминирует? | Trap? | Вердикт |
|---|---|---|---|---|
| **Light Fighter** | Mass anti-defense early, дёшев | Нет | Нет | ✅ ОК |
| **Strong Fighter** | Anti-LF/Cru/Bomber/BS, универсал mid-tier | Нет | Нет | ✅ ОК (после 27-I) |
| **Cruiser** | **Anti-LF (9.05)** + anti-RL (2.45) | Нет | Нет | ✅ ОК |
| **Battleship** | Универсал mid-tier capital | Нет | Нет | ✅ ОК |
| **Frigate** | **Anti-Cruiser (3.81), anti-BS (4.59)** — «king of capital» | Нет | Нет | ✅ ОК |
| **Bomber** | **Anti-defense (RL 4.70, LL 3.37, IG 6.58, Plasma 1.03)** | Нет | Нет | ✅ ОК (после 27-G) |
| **Star Destroyer** | **Anti-Frigate (15.64)**, anti-LL (2.50), anti-Bomber (2.16) | Нет | Нет | ✅ ОК |
| **Death Star** | Endgame godfist, контр всех 1v1 | **Да** (ATK clean vs всё кроме Shadow) | Нет | ⚠ См. 17.5 |
| **Shadow Ship** | **Anti-DS (6.67), anti-mass (90+)** | **Да чрезмерно** | Нет | ⚠ См. 17.5 |
| **Lancer Ship** | Был anti-DS, после 27-J **проигрывает всем 1v1** | Нет | **Да!** | ⚠ См. 17.5 |
| **Recycler** | Logistics (collect debris), не combat | — | — | ✅ ОК (logistics) |
| **Espionage Sensor** | Spy probe, не combat | — | — | ✅ ОК (logistics) |
| **Solar Satellite** | Production (energy), не combat | — | — | ✅ ОК (production) |
| **Colony Ship** | Logistics (новые планеты) | — | — | ✅ ОК (logistics) |
| **Small/Large Transport** | Logistics (cargo) | — | — | ✅ ОК (logistics) |

### 17.4 Сравнение с OGame паттернами

| Паттерн | OGame | oxsar-nova (после ребаланса) | Совпадает? |
|---|---|---|---|
| Cruiser RF×6 vs LF | ✅ | ✅ | ✅ |
| Battlecruiser/Frigate как «king of fleet» | ✅ (RF vs Cru/BS) | ✅ (Frigate exchange 3.81/4.59) | ✅ |
| Bomber как anti-defense | ✅ (RF vs RL/LL/HL/Gauss/Plasma/Ion) | ✅ (после 27-G) | ✅ |
| DS как «белый слон» в чистом PvP | ✅ (community: контрится Bomber+Destroyer+BC) | ⚠ DS в защите неуязвим без Shadow/Lancer | Частично |
| Heavy Fighter (Strong Fighter) — бесполезный | ⚠ (community: trap) | ✅ После 27-I получил роль (anti-LF mass-cheap) | **Лучше OGame** |
| Plasma Turret + Large Shield = «turtle gold standard» | ✅ | ✅ | ✅ |
| Shield 1% threshold (ignoreAttack) | ✅ shield × 0.01 | ✅ baseShield / 100 | ✅ |
| Shield регенерируется между раундами | ✅ | ✅ | ✅ |
| 6 раундов | ✅ | ✅ (default) | ✅ |
| Уникальный stealth-юнит | ❌ нет в classic | ✅ Shadow (oxsar-флавор) | **oxsar-уникум** |

**Вывод**: oxsar-nova **в основном совпадает** с OGame паттернами. Главные расхождения:
1. **SF лучше OGame** — после 27-I получил нишу.
2. **Shadow** — уникальная фича oxsar-nova, нет в OGame.
3. **Lancer** — был унаследован из oxsar2, в OGame нет аналога.

### 17.5 Найденные проблемы (после ребаланса)

#### 🔴 Проблема 1: Shadow Ship доминирует чрезмерно

При равной metal-eq **vs lite/mid флот**:
- vs LF: **90.91** (атакующий тратит 1k Shadow, защитник теряет 90× ресурсов)
- vs Frigate: **104.68**
- vs Bomber: **83.25**
- vs SD: **51.28**

**Почему**: Shadow стоит 5k metal-eq (1+3+1) — **самый дешёвый юнит с attack=520** в игре. После 27-F любые 1k+ Shadow убивают пугный флот без потерь.

**Контр**: только Strong Fighter (RF×25 vs Shadow) и Battleship (RF×70 vs Shadow), но даже они теряют:
- Shadow vs SF: 0.08 (SF уверенно убивает Shadow при равной cost — RF×25 работает)
- Shadow vs BS: 0.10 (BS RF×70 vs Shadow эффективен)

Но Shadow vs «мирный флот без RF» — катастрофа.

**Рекомендация — 27-V**: поднять cost Shadow с 5k до **15k** (3k+9k+3k). Тогда:
- 10M / 15k = 667 Shadow (вместо 2000 при 5k cost).
- Соотношение «ресурсов в бой» снизится в 3×.
- Anti-DS-роль сохранится (нужно меньше юнитов, но они дороже).

Альтернатива — снизить attack Shadow с 520 до граничных 501 (минимум для пробоя DS shield=50000 → ignoreAttack=500). Это **не поможет** — отношения не меняются, только absolute damage.

**Лучшее решение**: cost 15k + attack 520 (оставить).

#### 🔴 Проблема 2: Lancer стал бесполезен

После 27-J (attack 5500→4000) Lancer проигрывает всем 1v1:
- vs LF: **0.04**
- vs Cruiser: **0.04**
- vs SF: **0.02**
- vs Frigate: 0.51
- vs DS: только def-clean (Lancer теряет всё, DS не теряет)

**Lancer-fleet vs всё** в группах: побеждает только vs Shadow (3.12) — потому что
у Lancer RF×3 vs DS, но Shadow его убивает быстрее.

**Lancer теперь — trap-юнит**. Нужно либо:
- **27-W: вернуть attack Lancer обратно** в 5500, но сделать другой anti-spam-фикс
  (например, кэп на количество Lancer per planet — game-механика).
- **Удалить Lancer полностью** (radikal, но честно — у него нет роли после 27-F Shadow).
- **27-X: Дать Lancer уникальную роль** — RF против alien-юнитов (anti-AI), либо
  shell ↑ до 30k (чтобы переживал больше).

#### 🟡 Проблема 3: Death Star в защите неуязвим

DS exchange vs всё 1v1: ATK clean (атакующий не теряет, но **только если атакующий cost ≥ DS cost ≥ 10M**).

Mirror DS vs DS — draw (1.00 в groups, endgame-fleet vs endgame-fleet).
Vs Shadow — DS теряет (ratio 0.15).

**Это OGame-паттерн** (DS — «белый слон»). Не проблема, **by design**.

#### 🟢 Не-проблемы (после анализа)

- **Light Fighter mirror 1.00, проигрывает всем кроме SF/RL** — нормально, LF это
  раннее mass-юнит, не для late-game.
- **Cruiser mirror 1.00, vs equal-tier проигрывает** — нормально, Cruiser anti-LF.
- **Battleship mid-tier универсал** — норма OGame.
- **Death Star vs все ATK clean** — endgame.

### 17.6 Дубликаты ролей (после ребаланса)

После анализа **дубликатов нет**. Все combat-юниты имеют уникальные ниши:

| Tier | Юнит | Уникальная роль |
|---|---|---|
| 0 (mass) | Light Fighter | Cheap mass, RF anti-defense (early) |
| 1 (mid-light) | Strong Fighter | Anti-LF/Cru/Bomber, дёшев |
| 1 (mid-light) | Cruiser | Anti-LF (RF×6), anti-RL (RF×10) |
| 2 (mid-capital) | Battleship | Универсал capital |
| 2 (mid-capital) | Frigate | Anti-Cruiser/BS (RF), king of fleet |
| 2 (mid-capital) | Bomber | Anti-defense (RF vs RL/LL/SL/IG/Gauss/Plasma) |
| 3 (capital) | Star Destroyer | Anti-Frigate (RF×2) — единственный hard-counter |
| 3 (capital) | Lancer Ship | ⚠ Anti-DS (RF×3) — **роль ослаблена после 27-J** |
| 3 (stealth) | Shadow Ship | Anti-DS (RF×70), anti-mass — **доминирует** |
| 4 (endgame) | Death Star | God-tier, RF почти всё |

### 17.7 Защита (defense) — анализ ролей

| Юнит | Роль | Дубликат? |
|---|---|---|
| Rocket Launcher | Cheap mass-defense | Нет |
| Light Laser | Чуть лучше RL по cost-to-shield | ⚠ полу-дубль RL |
| Strong Laser | Mid-tier, anti-LF | Нет |
| Ion Gun | Anti-shield (high shield, low atk) | Нет |
| Gauss Gun | Heavy defense, attack=1100 (пробивает >shield/100 многих) | Нет |
| Plasma Gun | Endgame defense, attack=3000 | Нет |
| Small Shield | Front=13, абсорб первого огня | Нет |
| Large Shield | Front=14, абсорб ультра | ⚠ полу-дубль SS |

**Light Laser ≈ Rocket Launcher**: оба shell=3000, attack 80 vs 100, shield 20 vs 25, cost 2k vs 2k. **Light Laser ≈ почти дубль RL** (чуть лучше за +25% cost). Но roles они разные:
- RL: anti-LF (LF проигрывает RL экономически)
- LL: anti-LF чуть лучше + anti-Heavy Fighter

**Не критично**, но **27-Y candidate**: упростить — слить LL и RL в один или дать LL чёткую роль (например, anti-SF RF×3).

**Small/Large Shield**: дубликаты по роли (front-bait), различаются только масштабом. В OGame так же. **Не дубликат**.

### 17.8 Final-рекомендации (что ещё сделать)

| ADR | Изменение | Priority | Статус |
|---|---|---|---|
| **27-V** | Shadow cost 5k → **15k** (3k+9k+3k) | **HIGH** — фикс доминанты | ✅ done |
| **27-W** | Lancer attack 4000 → **5000** + кэп per planet=50 (game-механика) | **HIGH** — починка trap-юнита | ✅ done |
| **27-X** (alt) | Удалить Lancer полностью | medium — radikal | ❌ отвергнуто (выбран Variant Б = 27-V + 27-W) |
| **27-Y** | Light Laser → дать чёткую роль (например, RF×3 vs SF) | low | pending |

### 17.9 Что **НЕ нужно** трогать (work as designed)

- Death Star endgame-доминирование — by design, OGame-паттерн.
- Frigate как «king of capital» — by design.
- LF как cheap mass — by design.
- Defense vs DS-флот — by design (defense — налог при 3:1 превосходстве).
- Shadow anti-DS-роль (после фикса cost) — by design.

---

## 18. Итоговый вердикт по плану 27 итерация 2

**Общая оценка**: ребаланс **в основном успешен**, паттерны совпадают с OGame.

**Зелёное** (10 из 13):
- Cruiser, BS, Frigate, Bomber, SD, SF, LF, DS — все имеют уникальные роли.
- Defense scaling (RL → Plasma) корректен.
- Logistics-юниты (LT, ST, Recycler, Probe, Colony) не пересекаются с combat.

**Жёлтое (требует доделки)**:
- **Shadow** — слишком сильный, нужен 27-V (cost ↑).
- **Lancer** — стал trap-юнитом, нужен 27-W (attack ↑ + cap) или 27-X (удалить).

**Красное**: нет.

**Финальный recommended набор (после 27-V/W/X)**:
- Каждый combat-юнит имеет уникальную нишу, без дубликатов.
- Endgame-меню (anti-DS): Shadow-mass, Plasma-defense (RF×2 vs DS), Bomber против DS (через 27-G — частично), Lancer (после 27-W/X).
- Mid-tier меню: Frigate vs BS/Cru, SD vs Frigate, Cruiser vs LF, Bomber vs defense.
- Early: LF mass, SF anti-LF.

---

## 19. 27-V/W финальные результаты (Variant Б, ✅ внедрён)

### 19.1 Что применено

- **27-V**: Shadow cost 5k → **15k** (3k+9k+3k).
- **27-W**: Lancer attack 4000 → **5000** + `max_per_planet: 50`
  (новое поле в `ShipSpec`, runtime-проверка в
  `shipyard.Service.Enqueue`).

### 19.2 Эффект на матрицу 1v1 при 10M metal-eq

| Юнит | Доминанты до 27-V/W | После 27-V/W |
|---|---|---|
| Shadow vs мирный флот | exchange 50-100 (катастрофа) | **7-30** (норма для anti-fleet) |
| Shadow vs DS | 6.67 | **2.22** (целевой коридор) |
| Shadow vs defense | 1.25-3.71 (Shadow всех бьёт) | **0.07-0.27** (defense держит Shadow) |
| Lancer vs всё | 0.02-0.13 (trap) | **0.04-0.66** (слабый, но боеспособный) |
| Lancer-spam vs lite | exchange 0.21 (raid выгоден) | exchange 0.19 теоретически, но **cap=50/планета** делает Lancer-spam невозможным |

### 19.3 Финальные роли (после 27-V/W)

```
Tier 0 (mass):     LF              ← cheap mass-defense early
Tier 1 (mid-light): SF, Cruiser    ← anti-LF (Cruiser RF×6, SF anti-LF/Cru)
Tier 2 (capital):   BS, Frigate, Bomber
                    ↑ BS — универсал
                    ↑ Frigate — anti-Cru/BS (king of fleet)
                    ↑ Bomber — anti-defense (после 27-G)
Tier 3 (specialist): SD, Lancer, Shadow
                    ↑ SD — anti-Frigate hard-counter
                    ↑ Lancer — premium-комбо-юнит, max 50/планета
                    ↑ Shadow — anti-DS specialist (cost 15k, attack 520)
Tier 4 (endgame):   DS              ← god-tier (OGame-паттерн)
```

### 19.4 Дубликаты ролей

**Нет.** Каждый combat-юнит имеет уникальную нишу:
- LF, SF — разные tier'ы (mass vs anti-fighter)
- Cruiser — anti-LF
- BS, Frigate, Bomber — разные роли в capital
- SD, Lancer, Shadow — разные специалисты
- DS — endgame godfist

### 19.5 Trap-юнитов нет

- LF: **есть роль** (cheap mass-fodder + anti-RL early)
- SF: **есть роль** после 27-I (anti-LF/Cru/BS)
- Cruiser: **есть роль** (anti-LF, RF×6)
- BS: универсал capital
- Frigate: **king of fleet**
- Bomber: **anti-defense**
- SD: **anti-Frigate**
- Lancer: **premium-комбо** (теперь капкой защищён от spam)
- Shadow: **anti-DS specialist** (после 27-V — не «wunderwaffe»)
- DS: endgame
- LT/ST/Recycler/Probe/Colony/SSat: logistics/production (по дизайну
  не combat)

### 19.6 Light Laser ≈ Rocket Launcher (полу-дубль, отложено)

LL и RL имеют близкие стат-профили (shell=3000, attack 80 vs 100,
cost ~2k). Это **не критично** — работают как разные mass-defense
юниты; в OGame так же. Можно дать LL уникальную роль через RF×3 vs
SF (27-Y), но это **низкий приоритет**. Откладывается.
