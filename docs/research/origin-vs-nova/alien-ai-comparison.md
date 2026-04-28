# Alien AI: origin vs nova — углублённое сравнение

**Дата сборки**: 2026-04-28
**Контекст**: артефакт плана 62, Категория 4 (event-loop). Самая
сложная подсистема legacy — 1127 строк AI-движка, которые в nova
реализованы лишь частично (план 15, Этап 3 — упрощён). Этот файл
становится спецификацией паритета: что именно и в каком объёме
нужно достроить в nova под legacy-вселенную.

---

## Часть I. Origin (legacy)

### Файлы

- `projects/game-origin-php/src/game/AlienAI.class.php` (1127 строк)
- `projects/game-origin-php/src/game/EventHandler.class.php:3532-3573`
  (обработчики EVENT_ALIEN_*)
- `projects/game-origin-php/config/consts.php:98-102, 440-445, 459-460,
  764-782` — константы

### Public/Protected методы

| Метод | file:line | Назначение |
|---|---|---|
| `isAttackTime($time = null)` | 15 | Проверяет, четверг ли (день атаки) — `date("w") == 4` |
| `isAlienPosition($galaxy, $system)` | 24 | Есть ли инопланетный флот в HOLDING на координатах |
| `checkAlientNeeds()` | 184 | Главный entry — генерирует миссии в очередь до лимита |
| `generateAttack($userid, $planetid, $fly_time, $alien_fleet)` | 161 | Кастомная атака (admin EVENT_ALIEN_ATTACK_CUSTOM) |
| `generateFleet($target_ships, $available_ships, $scale, $params)` | 405 | Подбор флота под целевую мощь |
| `setupFleet(array $fleet_ships)` | 372 | Нормализация структуры флота (qty/damaged/shell_percent) |
| `onFlyUnknownEvent($event)` | 652 | Обработчик прибытия — грабёж, подарок или атака |
| `onAttackEvent($event)` | 642 | Обработчик атаки — запуск Assault |
| `onGrabCreditEvent($event)` | 647 | → `onFlyUnknownEvent()` (с mode=GRAB_CREDIT) |
| `onHaltEvent($event)` | 827 | Создаёт EVENT_ALIEN_HOLDING |
| `onHoldingEvent($event)` | 856 | Просто ждёт + checkAlientNeeds() в конце |
| `onChangeMissionAIEvent($event)` | 864 | Усиливает миссию (power_scale += 1.5 за смену) |
| `onHoldingAIEvent($event)` | 924 | Случайное действие из 8 (с равными весами) |
| `findTarget()` | 336 | Поиск цели для атаки (≥1000 кораблей, активен 30 мин) |
| `findCreditTarget()` | 299 | Поиск цели грабежа (≥100k кредитов, ≥300k кораблей) |
| `generateMission($target, $params)` | 59 | Создание миссии целиком |
| `loadPlanetShips($planetid)` | 266 | Загрузка флота на планете |
| `loadUserResearches($userid)` | 278 | Уровни технологий цели |
| `shuffleKeyValues(&$r, $keys)` | 251 | Случайное ослабление техник |
| `onExtractAlientShipsAI($unload_resourses)` | 1025 | Извлечь 0-1% × times² кораблей |
| `onUnloadAlienResoursesAI()` | 1081 | Разгрузить ресурсы вместе с кораблями |
| `onRepairUserUnitsAI / onAddUserUnitsAI / onAddCreditsAI / onAddArtefactAI / onGenerateAsteroidAI / onFindPlanetAfterBattleAI` | 1086-1125 | **Заглушки** (пустые методы — 6 из 8 действий HOLDING_AI) |

### State machine (полная цепочка переходов)

```
checkAlientNeeds() [184]
├─ isAttackTime() [192] — четверг ⇒ FLEETS=250, scale=1.5..2.0; иначе FLEETS=50, scale=0.9..1.1
├─ findTarget() [69] или findCreditTarget() [65]
└─ generateMission() [59, 204]
   ├─ generateFleet() [99] — UNIT_A_CORVETTE/SCREEN/PALADIN/FRIGATE/TORPEDOCARIER
   ├─ mode = 90% EVENT_ALIEN_ATTACK / 10% EVENT_ALIEN_FLY_UNKNOWN [103]
   ├─ addEvent(...) [207] — fly_time = randRoundRange(15h, 24h) [126]
   └─ 60% вероятность: addEvent(EVENT_ALIEN_CHANGE_MISSION_AI) [217]
      └─ trigger: 8-10h или (parent_time-30 ... parent_time-10) [221, 226]

EVENT_ALIEN_FLY_UNKNOWN / EVENT_ALIEN_GRAB_CREDIT / EVENT_ALIEN_ATTACK_CUSTOM
  → EventHandler::alienFlyUnknown() [3532]
  → AlienAI::onFlyUnknownEvent() [652]:
    ├─ 10%: новая generateMission() → EVENT_ALIEN_CHANGE_MISSION_AI [774-814]
    ├─ if credit > 100k && (mode==GRAB_CREDIT || 10%):
    │   └─ grab = credit * 0.01 * rand(0.08, 0.10) [667]
    │   └─ 90% return [692]
    ├─ if !grab && 5%: подарок ресурсов (rand(0.7, 1.0) от данных события) [702]
    │   └─ return [733]
    ├─ if !grab && 5%: подарок кредитов min(500, credit*0.01*rand(5,10)) [739-740]
    │   └─ if mode==ATTACK_CUSTOM return [768-770]
    ├─ if (grab || isAttackTime ? 90% : 50%):
    │   └─ onAttackEvent() → EventHandler::attack() → Assault [821]
    └─ else:
        └─ onHaltEvent() → EVENT_ALIEN_HOLDING [824, 831]
            └─ duration = randRoundRange(12h, 24h) [127]
            └─ if admin: + EVENT_ALIEN_HOLDING_AI [836]

EVENT_ALIEN_HOLDING → onHoldingEvent() [856]
  └─ checkAlientNeeds() [858]

EVENT_ALIEN_HOLDING_AI → onHoldingAIEvent() [924]
  ├─ random выбор из 8 действий (по 10%):
  │   ├─ onExtractAlientShipsAI() [1025] — извлечь 0-1% × times² кораблей
  │   ├─ onUnloadAlienResoursesAI() [1081] — разгрузить ресурсы
  │   └─ 6 заглушек (no-op)
  ├─ duration_subphase = clamp(min(12h, 30s*times) ... max(24h, 60s*times)) [974]
  ├─ control_times++ [978]
  ├─ if paid_credit > 0: end_time += 2h * paid_credit/50 [993]
  └─ 1% probability: checkAlientNeeds() [1006-1008]

EVENT_ALIEN_CHANGE_MISSION_AI → onChangeMissionAIEvent() [864]
  ├─ if remaining >= 8h:
  │   └─ generateMission() with power_scale = 1 + control_times * 1.5 [884]
  │   └─ 50% mode = ATTACK / 50% FLY_UNKNOWN [885]
  │   └─ control_times++ [892]
  │   └─ update parent event time [895]
  └─ else (< 8h):
      └─ extend by random(10..50)s [910]
      └─ control_times++ [908]
```

### Алгоритм `generateFleet()` (подбор флота, 405-622)

**Вход**: `$target_ships` (флот цели), `$available_ships` (UNIT_A_*),
`$scale` (множитель мощи).

**Шаги**:
1. Целевая мощь: `target_power = max(100, sum(attack + shield) * scale)`.
   Учёт спец-кораблей (Death Star, Transplantator, Armored Terran):
   их вклад × 0.20.
2. Итеративно (`while`): выбираем случайный корабль из
   `$available_ships` и добавляем `quantity` штук, пока:
   - `power < target_power` И
   - `total_debris <= 1e9` (`ALIEN_FLEET_MAX_DERBIS`).
3. Лимиты:
   - UNIT_DEATH_STAR: max 30%-90% от Death Stars цели (90% веса).
   - UNIT_SHIP_TRANSPLANTATOR: max `1 + 2 * max_death_stars`.
   - UNIT_SHIP_ARMORED_TERRAN: только 1 штука + break.
   - Каждый корабль: `qty <= 20 000`.
4. Корабли пришельцев (`config/consts.php:98-102`):
   - 200 = UNIT_A_CORVETTE
   - 201 = UNIT_A_SCREEN
   - 202 = UNIT_A_PALADIN
   - 203 = UNIT_A_FRIGATE
   - 204 = UNIT_A_TORPEDOCARIER

### Грабёж кредитов (`onFlyUnknownEvent`, 667)

Условие: `user_credit > ALIEN_GRAB_MIN_CREDIT (100k)` И (`mode ==
GRAB_CREDIT` ИЛИ 10% случайно).

```
grab = round(user_credit * 0.01 * rand(0.08, 0.10), 2)
     = user_credit * [0.0008 .. 0.001]
```

После успешного грабежа — 90% вероятность улететь без атаки.

Защита: `ALIEN_GRAB_CREDIT_INTERVAL = 10 дней` — цель не атакуется
повторно этим режимом.

### Подарки

- **Ресурсы** (5%, при !grab_credit):
  ```
  scale = rand(0.7, 1.0)
  gift_metal = event["data"]["metal"] * scale
  gift_silicon = event["data"]["silicon"] * scale
  gift_hydrogen = event["data"]["hydrogen"] * scale
  ```
- **Кредиты** (5%, при !grab_credit):
  ```
  max_gift = 500 * rand(0.98, 1.02)
  gift_credit = min(max_gift, user_credit * 0.01 * rand(5, 10))
  ```

### Удержание (HOLDING)

- **Длительность**: `randRoundRange(12h, 24h)`, лимит реал-времени
  `ALIEN_HALTING_MAX_REAL_TIME = 15 дней`.
- **Каждый тик** (HOLDING_AI):
  - 1 из 8 действий с равными весами (10%):
    - `onExtractAlientShipsAI()` — изъять 0-1% × `times²` кораблей.
    - `onUnloadAlienResoursesAI()` — разгрузить ресурсы.
    - 6 заглушек (Repair / AddUnits / AddCredits / AddArtefact /
      GenerateAsteroid / FindPlanetAfterBattle).
  - Длительность подфазы `min(12h, 30s*times) ... max(24h, 60s*times)`.
  - `control_times++`.
- **Платный выкуп**: `end_time += 2h * paid_credit / 50`.

### Смена миссии в полёте (CHANGE_MISSION_AI)

Условие: 60% вероятность при `generateMission()`. Срабатывает за
`(parent_time - 30 .. parent_time - 10)` секунд до завершения,
либо за 8-10h до конца.

Эффект:
- При `>=8h` остаётся: новая `generateMission()` с
  `power_scale = 1 + control_times * 1.5` ⇒ мощь растёт
  квадратично через цепочку.
- При `<8h`: продление времени на `random(10..50)s`.

### Выбор цели

**`findTarget()` (336)**:
- активен (last < 30 мин)
- не в umode
- ≥ 1000 кораблей у игрока
- ≥ 100 кораблей на конкретной планете
- 10% — UNIT_SOLAR_SATELLITE как target; 90% — исключить
- не атакован за `ALIEN_ATTACK_INTERVAL = 6 дней`
- `ORDER BY rand() LIMIT 1`

**`findCreditTarget()` (299)**:
- активен
- не в umode
- ≥ 100 000 кредитов
- ≥ 300 000 кораблей у игрока
- ≥ 10 000 на планете
- не грабился `ALIEN_GRAB_CREDIT_INTERVAL = 10 дней`

### День атаки (четверг)

- `date("w") == 4` ⇒ `isAttackTime() = true`.
- Множители:
  - `ALIEN_NORMAL_FLEETS_NUMBER = 50` → `ALIEN_ATTACK_TIME_FLEETS_NUMBER = 250` (×5).
  - Power scale: `rand(0.9, 1.1)` → `rand(1.5, 2.0)` (×1.5..2.0).

### События EVENT_ALIEN_*

| Событие | ID | Метод EH | Делает | Следующее |
|---|---|---|---|---|
| EVENT_ALIEN_FLY_UNKNOWN | 33 | alienFlyUnknown (3532) | onFlyUnknownEvent — грабёж/подарок/атака | ATTACK / HOLDING / CHANGE_MISSION_AI |
| EVENT_ALIEN_HOLDING | 34 | alienHolding (3544) | onHoldingEvent — checkAlientNeeds в конце | HOLDING_AI (если admin) |
| EVENT_ALIEN_ATTACK | 35 | alienAttack (3562) | onAttackEvent → Assault | HALT (если выжили) |
| EVENT_ALIEN_HALT | 36 | alienHalt (3568) | onHaltEvent → EVENT_ALIEN_HOLDING | HOLDING |
| EVENT_ALIEN_GRAB_CREDIT | 37 | alienGrabCredit (3538) | → onFlyUnknownEvent (mode=GRAB) | как FLY_UNKNOWN |
| EVENT_ALIEN_ATTACK_CUSTOM | 38 | alienFlyUnknown (3532) | onFlyUnknownEvent (admin custom) | подарок/return |
| EVENT_ALIEN_HOLDING_AI | 80 | alienHoldingAI (3550) | onHoldingAIEvent — 1 из 8 | HOLDING_AI (рекурсивно) |
| EVENT_ALIEN_CHANGE_MISSION_AI | 81 | alienChangeMissionAI (3556) | onChangeMissionAIEvent — усилить parent | (обновляет parent ATTACK/FLY_UNKNOWN) |

### Ключевые константы (`config/consts.php`)

```php
ALIEN_ENABLED                   = !DEATHMATCH
ALIEN_NORMAL_FLEETS_NUMBER      = 50
ALIEN_ATTACK_TIME_FLEETS_NUMBER = 250
ALIEN_ATTACK_INTERVAL           = 6 * 24h
ALIEN_GRAB_CREDIT_INTERVAL      = 10 * 24h
ALIEN_FLY_MIN_TIME              = 15h
ALIEN_FLY_MAX_TIME              = 24h
ALIEN_HALTING_MIN_TIME          = 12h
ALIEN_HALTING_MAX_TIME          = 24h
ALIEN_HALTING_MAX_REAL_TIME     = 15 * 24h
ALIEN_CHANGE_MISSION_MIN_TIME   = 8h
ALIEN_CHANGE_MISSION_MAX_TIME   = 10h
ALIEN_GRAB_MIN_CREDIT           = 100_000
ALIEN_GRAB_CREDIT_MIN_PERCENT   = 0.08
ALIEN_GRAB_CREDIT_MAX_PERCENT   = 0.1
ALIEN_GIFT_CREDIT_MIN_PERCENT   = 5
ALIEN_GIFT_CREDIT_MAX_PERCENT   = 10
ALIEN_MAX_GIFT_CREDIT           = 500
ALIEN_FLEET_MAX_DERBIS          = 1_000_000_000
```

Технологии для случайного ослабления:
```
UNIT_EXPO_TECH=27, UNIT_GUN_TECH=15, UNIT_SHIELD_TECH=16,
UNIT_SHELL_TECH=17, UNIT_SPYWARE=13, UNIT_BALLISTICS_TECH=103,
UNIT_MASKING_TECH=104, UNIT_LASER_TECH=23, UNIT_ION_TECH=24,
UNIT_PLASMA_TECH=25
```

### Корабли пришельцев (`na_ship_datasheet`, `unitid` 200-204)

Параметры в БД (запрос `SELECT * FROM na_ship_datasheet WHERE
unitid IN (200,201,202,203,204);`):

| unitid | name | attack | shield | speed | cargo | fuel | ballistics | masking |
|---|---|---|---|---|---|---|---|---|
| 200 | Alien Corvette | 200 | 75 | 20 000 | 300 | 150 | 2 | 0 |
| 201 | Alien Screen | 22 | 5 000 | 10 000 | 800 | 75 | 0 | 0 |
| 202 | Alien Paladin | 75 | 50 | 8 000 | 50 | 20 | 5 | 1 |
| 203 | Alien Frigate | 1 250 | 150 | 10 000 | 2 000 | 300 | 1 | 0 |
| 204 | Alien Torpedocarrier | 350 | 100 | 13 000 | 200 | 100 | 4 | 0 |

(Источник: [docs/legacy/game-reference.md](../../legacy/game-reference.md)
§ «Параметры юнитов из БД легаси».)

---

## Часть II. Nova (current)

### Файлы (план 15)

- `projects/game-nova/backend/internal/event/kinds.go` — `KindAlien*`
  типы (Kind 50-58 ориентировочно — уточнить при чтении).
- `projects/game-nova/backend/internal/event/handlers.go` —
  обработчики.
- `projects/game-nova/backend/internal/alien/` (если выделен) или
  внутри `internal/event/`.
- План 15 (alien-holding-thursday) — реализован, но **Этап 3
  упрощён**: пропущены ChangeMission / GrabCredit / FlyUnknown
  без атаки.

### Что точно есть в nova (по плану 15)

- День атаки — четверг.
- Базовый цикл `Halt → Holding → Attack`.
- Простой `RaidWarning` за 10 мин до атаки (Kind=64).
- Простые корабли пришельцев (фиксированные 5 LF, по словам автора).

### Что точно НЕТ в nova (по плану 15 + memory)

| Функция | Статус в nova |
|---|---|
| Подбор флота под силу игрока (масштабируется) | ❌ (фиксированные 5 LF) |
| Корабли пришельцев UNIT_A_* (5 видов, id 200-204) | ❌ |
| День атаки — множитель ×5 флотов и ×1.5..2.0 мощь | ❌ (только триггер четверга) |
| EVENT_ALIEN_CHANGE_MISSION_AI (смена миссии в полёте) | ❌ |
| HOLDING с 8 действиями каждый тик | ❌ (только базовое HOLDING) |
| Грабёж кредитов | частично (`applyGrabCredit` есть, без state machine) |
| Подарки ресурсов / кредитов | ❌ |
| Платный выкуп удержания (paid_credit) | ❌ |
| Случайное ослабление техник (shuffleKeyValues) | ❌ |

> **Точная верификация nova-стороны** — выполнится при чтении
> `projects/game-nova/backend/internal/event/` в Ф.2. Этот раздел
> обновится по итогам.

---

## Часть III. Расхождения и план достройки

### Расхождения для журнала (записи D-NNN-A)

Каждая строка ниже превратится в одну или несколько записей в
`divergence-log.md` категории `event-loop`/`механика`:

| # | Что в origin | Что в nova | Цвет | Что нужно в nova |
|---|---|---|---|---|
| A1 | Подбор флота `generateFleet()` под мощь цели | Фиксированные 5 LF | 🟠 | Реализовать `internal/alien/fleet_generator.go` с алгоритмом из 405-622 |
| A2 | 5 видов кораблей UNIT_A_* | Нет | 🟠 | Добавить в `configs/units.yml` секцию `alien:`, в БД — 5 новых ship-id'ов под legacy-режимом |
| A3 | Четверг ×5 флотов, ×1.5..2.0 мощь | Только триггер четверга | 🟠 | Параметризовать `attack_time_fleets_number`, `attack_time_power_scale` в `configs/balance/legacy.yaml` |
| A4 | EVENT_ALIEN_CHANGE_MISSION_AI | Нет | 🟠 | Новый Kind в kinds.go + обработчик с power_scale |
| A5 | HOLDING с 8 действиями, control_times² | Простое HOLDING | 🟠 | Реализовать state в `internal/alien/holding.go` с теми же действиями |
| A6 | Грабёж кредитов с state machine | Частично (applyGrabCredit) | 🟡 | Достроить `EVENT_ALIEN_GRAB_CREDIT` отдельным Kind, переход FLY_UNKNOWN ↔ GRAB |
| A7 | Подарки ресурсов/кредитов с вероятностями | Нет | 🟠 | Реализовать в обработчике FlyUnknown |
| A8 | Платный выкуп `end_time += 2h * paid/50` | Нет | 🟡 | Endpoint `POST /api/alien/holding/{id}/pay` + событие продления |
| A9 | shuffleKeyValues — случайное ослабление техник | Нет | 🟢 | Опционально (балансовая особенность, можно вынести в legacy.yaml) |
| A10 | findCreditTarget с критериями ≥100k credit, ≥300k ships | Нет | 🟠 | Новый запрос в `internal/alien/target.go` |
| A11 | UNIT_A_* в `na_ship_datasheet` (id 200-204) | Нет в `configs/units.yml` | 🟠 | Добавить 5 юнитов под флагом `legacy_alien_units: true` |
| A12 | `ALIEN_FLEET_MAX_DERBIS = 1e9` | Нет ограничения | 🟢 | Параметр `alien.fleet_max_debris` в legacy.yaml |
| A13 | Все 12+ ALIEN_* констант (тайминги, проценты) | Хардкод в Go или нет | 🟡 | Вынести в `configs/balance/legacy.yaml` (секция `alien:`) |
| A14 | 6 заглушек в HOLDING_AI (Repair, AddUnits, ...) | Нет | 🟢 | **Не реализовывать** — origin сам их не вызывает (no-op). В журнал как «известное упрощение» |

### Спецификация nova-стороны для legacy-вселенной

После расширения nova-backend под legacy01:

```
internal/alien/
├── ai.go              — main entry (CheckAlienNeeds-аналог)
├── fleet_generator.go — generateFleet алгоритм из origin:405-622
├── target.go          — findTarget + findCreditTarget
├── flyunknown.go      — onFlyUnknownEvent (грабёж/подарок/атака)
├── holding.go         — state machine HOLDING + 8 действий
├── change_mission.go  — onChangeMissionAIEvent
├── config.go          — ALIEN_* константы из legacy.yaml
└── legacy_units.go    — 5 видов кораблей UNIT_A_*
```

Все ALIEN_*-параметры — из `configs/balance/legacy.yaml`,
в современных вселенных (uni01/uni02) этот файл не подгружается,
работает упрощённый план 15.

---

## References

- `projects/game-origin-php/src/game/AlienAI.class.php` (основной)
- `projects/game-origin-php/src/game/EventHandler.class.php:3532-3573`
- `projects/game-origin-php/config/consts.php:98-102, 440-445, 459-460,
  764-782`
- [docs/legacy/game-reference.md § «Система инопланетян»](../../legacy/game-reference.md)
- План 15 (alien-holding-thursday) — Этап 3 явно упрощён
- Memory: упоминание упрощений плана 15 относительно origin
