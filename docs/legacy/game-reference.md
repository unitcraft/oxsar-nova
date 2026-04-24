Старая игра (oxsar2, PHP/Yii 1.1) запущена локально:

- **URL**: http://localhost:8080/game.php/Main
- **Логин**: test (или любой другой пользователь)
- **Пароль**: quoYaMe1wHo4xaci (универсальный, подходит для любого пользователя)

## Доступ через curl (для анализа без браузера)

Логин делается через `/login.php` (не через Yii-роут `/game.php/Login`):

```bash
# 1. Логин — сохраняет сессионные куки в файл
curl -s -c /tmp/oxsar_cookies.txt -b /tmp/oxsar_cookies.txt \
  -X POST "http://localhost:8080/login.php" \
  -d "username=test&password=quoYaMe1wHo4xaci&login=OK" \
  -L -o /dev/null -w "HTTP %{http_code}\n"

# 2. Любой последующий запрос — передаём куки
curl -s -c /tmp/oxsar_cookies.txt -b /tmp/oxsar_cookies.txt \
  "http://localhost:8080/game.php/Main" -L

# Примеры других страниц:
# /game.php/Constructions  — здания
# /game.php/Research       — исследования
# /game.php/Shipyard       — верфь
# /game.php/Defense        — оборона
# /game.php/Fleet          — флот
# /game.php/Galaxy         — галактика
# /game.php/Ranking        — рейтинг
```

После успешного логина сервер вернёт редирект на `/game.php` и установит два куки: `PHPSESSID` и хэш-куку с данными пользователя.

**Легаси код**: `d:\Sources\oxsar2` — PHP/Yii 1.1. UI строится на шаблонизаторе, шаблоны в файлах с расширением `.tpl`.

**Расширение игры**: `d:\Sources\oxsar2\www\ext` — обязательно смотреть при изучении логики легаси. Содержит расширения/переопределения базовой игры. Без него картина неполная — часть UI и бизнес-логики реализована именно здесь.

**Зачем**: сравнивать дизайн и функционал старой игры с новой (oxsar-nova), чтобы убедиться что всё реализовано. Перед реализацией новой фичи — сначала смотреть как оно работает в старой игре (живой сайт или `.tpl`-шаблоны).

## Конфиги и параметры игры

Основные параметры находятся в `d:\Sources\oxsar2\www\new_game\protected\config\`:

- `consts.php` — главный файл констант (значения по умолчанию)
- `consts.local.php` → `consts.dominator.local.php` — **активный локальный override**, переопределяет дефолты. Именно здесь заданы реальные значения запущенного инстанса.
- `params.php` — дополнительные параметры приложения (стартовые ресурсы, лимиты, ссылки)

Механизм: `consts.php` проверяет `defined('X')` перед каждым `define`, поэтому `consts.local.php` грузится первым и имеет приоритет.

### Ключевые параметры (конфиг активного инстанса — Dominator)

| Параметр | Значение | Описание |
|----------|----------|----------|
| `UNIVERSE_NAME` | `'Dominator'` | Название вселенной |
| `NUM_GALAXYS` | 3 | Количество галактик (дефолт 8) |
| `NUM_SYSTEMS` | 300 | Систем в галактике (дефолт 600) |
| `GAMESPEED` | `0.75 / 8 = 0.09375` | Скорость игры (очень медленная: `GAMESPEED_SCALE=8`) |
| `FLY_SPEED_SCALE` | 1 | Множитель скорости полёта |
| `ROCKET_SPEED_FACTOR` | 8 | Скорость ракет (`GAMESPEED_SCALE × FLY_SPEED_SCALE`) |
| `MAX_PLANETS` | 12 | Максимум планет на игрока (дефолт 10) |
| `ADDITIONAL_ARTEFACT_PLANETS_NUMBER` | 5 | Доп. планеты за артефакты (дефолт 3) |
| `TEMP_PLANETS_NUMBER` | 5 | Максимум временных планет |
| `BASHING_PERIOD` | `60×60×5 = 5ч` | Период защиты от башинга (дефолт 0 — выключено) |
| `BASHING_MAX_ATTACKS` | 4 | Макс. атак одного игрока за период башинга |
| `PROTECTION_PERIOD` | `60×60×24 = 24ч` | Защита нового игрока от атак |
| `NEW_USER_OBSERVER` | 0 | Режим наблюдателя для новых игроков |
| `BLOCK_TARGET_BY_ALLY_ATTACK_USED_PERIOD` | `60×60×24×1` | Блокировка цели после альянсовой атаки |
| `EXCH_INVIOLABLE` | 0 | ID неприкосновенного биржевого лота (0 = нет) |
| `ACHIEVEMENTS_ENABLED` | false | Достижения отключены |
| `TUTORIAL_ENABLED` | false | Туториал отключён |
| `SHOW_DM_POINTS` | true | Показывать DM-очки |

### Параметры начального старта игрока (Dominator)

Начальные здания (`INITIAL_BUILDINGS`): Metal Mine 2, Silicon Lab 2, Hydrogen Lab 2, Solar Plant 4, Robotic Factory 2, Shipyard 2, Research Lab 2, Defense Factory 1, Repair Factory 1.

Начальные исследования (`INITIAL_RESEARCHES`): Computer Tech 1, Energy Tech 1, Combustion Engine 2.

Начальный флот (`INITIAL_UNITS`): 20 Small Transporter, 10 Light Fighter, 10 Recycler, 3 Colony Ship, 10 Espionage Probe.

Стартовые ресурсы (из `params.php`): Metal 1000, Silicon 500, Hydrogen 0. Домашняя планета: 18 800 полей (`HOME_PLANET_SIZE`).

### Параметры галактик (GALAXY_SPEC_PARAMS — Dominator)

Для каждой из 3 галактик одинаковые множители:

| Параметр | Значение | Описание |
|----------|----------|----------|
| `RESOURCES_PRODUCTION_FACTOR` | **5** | Производство ресурсов ×5 |
| `STORAGE_FACTOR` | **5** | Ёмкость хранилищ ×5 |
| `RESEARCH_SPEED_FACTOR` | **2** | Скорость исследований ×2 |
| `MOON_CONSTRUCTION_SPEED_FACTOR` | 2 | Скорость стройки на луне ×2 |
| `TEMP_MOON_CONSTRUCTION_SPEED_FACTOR` | 4 | Скорость стройки на вр. луне ×4 |
| `ENEGRY_PRODUCTION_FACTOR` | 0.8 | Производство энергии ×0.8 (снижено!) |
| `FLEET_SPEED_FACTOR` | 1 | Скорость флота (FLY_SPEED_SCALE) |
| `ALLOW_DESTROY_MOON` | 1 | Разрешено уничтожение луны |
| `ALLOW_STARGATE_TRANSPORT` | 1 | Разрешён транспорт через Stargate |
| `ADVANCED_BATTLE` | 0 | Продвинутый бой отключён |

### Параметры очков (Dominator — отличаются от дефолта)

| Параметр | Значение | Описание |
|----------|----------|----------|
| `RES_TO_UNIT_POINTS` | `2.0/1000` | Очки за юниты (флот/оборону) |
| `RES_TO_RESEARCH_POINTS` | `0.5/1000` | Очки за исследования (дефолт 1.0/1000) |
| `RES_TO_BUILD_POINTS` | `0.05/1000` | Очки за постройки (дефолт 0.5/1000) |

### Параметры Биржи (Exchange)

| Параметр | Значение | Описание |
|----------|----------|----------|
| `EXCH_MERCHANT_PREMIUM_COMMISSION` | 10% | Комиссия владельца биржи с премиум-игрока |
| `EXCH_MERCHANT_COMMISSION` | 13% | Комиссия владельца биржи с обычного игрока |
| `EXCH_NO_MERCHANT_PREMIUM_COMMISSION` | 16% | Комиссия без биржи (премиум) |
| `EXCH_NO_MERCHANT_COMMISSION` | 19% | Комиссия без биржи (обычный) |
| `EXCH_MAX_TTL` | 7 дней | Максимальный TTL лота |
| `EXCH_MIN_TTL` | 3 дня | Минимальный TTL |
| `EXCH_LEVEL_SLOTS` | 15 | Слотов на уровень биржи |
| `EXCH_SELLER_MAX_PROFIT` | 1000% | Макс. наценка продавца |
| `EXCH_RADIUS_SYSTEMS_PER_GALAXY` | 300 | Радиус действия биржи (систем) |

### Другие важные константы (consts.php)

| Параметр | Значение | Описание |
|----------|----------|----------|
| `MAX_BUILDING_LEVEL` | 40 | Макс. уровень зданий |
| `MAX_RESEARCH_LEVEL` | 40 | Макс. уровень исследований |
| `FLEET_BULK_INTO_DEBRIS` | 0.50 | 50% флота уходит в обломки |
| `DEFENCE_BULK_INTO_DEBRIS` | 0.01 | 1% обороны уходит в обломки |
| `EV_ABORT_MAX_BUILD_PERCENT` | 95% | Возврат ресурсов при отмене стройки |
| `EV_ABORT_SAVE_TIME` | 15 сек | Полный возврат ресурсов в первые 15 сек |
| `EV_ABORT_MAX_SHIPYARD_PERCENT` | 70% | Возврат при отмене верфи |
| `EV_ABORT_MAX_REPAIR_PERCENT` | 70% | Возврат при отмене ремонта |
| `EV_ABORT_MAX_FLY_PERCENT` | 90% | Возврат топлива при отзыве флота |
| `FLEET_FUEL_CONSUMPTION` | 0.5 | Базовый множитель расхода топлива |
| `VACATION_DISABLE_TIME` | 30 дней | Через сколько включится защита отпуска |
| `LAST_TIME_ON_VACATION_DISABLE` | 20 дней | Мин. интервал между отпусками |
| `MOON_CREATION_USER_INTERVAL` | 7 дней | Интервал создания луны для пользователя |
| `MOON_CREATION_SYSTEM_INTERVAL` | 7 дней | Интервал создания луны в системе |
| `ALIEN_ATTACK_INTERVAL` | 6 дней | Интервал атак пришельцев |
| `ALIEN_NORMAL_FLEETS_NUMBER` | 50 | Нормальное число флотов пришельцев |
| `STARGATE_TRANSPORT_SPEED` | 1ч (3600 сек) | Базовая скорость Stargate-транспорта |
| `PRODUCTION_FACTOR` (params) | 1.5 | Базовый множитель производства |
| `STORAGE_SAVE_FACTOR` | 0.01 | Коэффициент сохранения хранилища |
| `UNITS_GROUP_CONSUMTION_POWER_BASE` | 1.000004 | База потребления группы юнитов |

### Система профессий (PROFESSIONS)

В легаси есть 4 профессии игрока, каждая даёт бонусы/штрафы к уровням зданий и исследований:

| Профессия | Бонусы | Штрафы |
|-----------|--------|--------|
| **MINER** | Metal Mine+1, Silicon Lab+1, Solar Plant+2 | Shipyard−2, Gun/Shield/Shell−2, Ballistics/Computer−1 |
| **ATTACKER** | Gun/Shield/Shell+1, Ballistics+1, Shipyard+1 | Metal Mine−1, Silicon Lab−1, IGN−1, Defense Factory−3 |
| **DEFENDER** | Masking+1, Shield/Shell+1, Defense Factory+1, Rocket Station+1 | Computer−1, Gun−1, Shipyard−3 |
| **TANK** | Gun/Shield/Shell+2 | Gravi−2, все двигатели−2 |

Смена профессии: `PROFESSION_CHANGE_COST = 1000 кредитов`, мин. интервал 14 дней.

### Специальные лимиты уровней ($GLOBALS["MAX_UNIT_LEVELS"])

| Здание | Макс. уровень |
|--------|---------------|
| Moon Hydrogen Lab | 10 |
| Moon Repair Factory | 9 |
| Moon Lab | 5 |
| Nano Factory | 12 |
| Star Gate | 15 |
| Gravi | 10 |

### Система экспедиций (типы исходов)

Исходный файл логики: `d:\Sources\oxsar2\www\game\Expedition.class.php`.

В легаси **13 типов** исходов (в nova реализовано 6):

| Тип | Константа | В nova | Описание |
|-----|-----------|--------|----------|
| Артефакт | `EXPED_TYPE_ARTEFACT` (1) | ✅ | Случайный артефакт, зависит от expo_tech |
| Астероид | `EXPED_TYPE_ASTEROID` (2) | частично | M+Si, без водорода |
| Поле битвы | `EXPED_TYPE_BATTLEFIELD` (3) | ❌ | Бой с повреждённым флотом, можно выиграть корабли |
| Чёрная дыра | `EXPED_TYPE_BLACK_HOLE` (4) | ❌ | Не реализовано даже в legacy |
| Кредиты | `EXPED_TYPE_CREDIT` (5) | ❌ | Находка кредитов, зависит от покупок игрока |
| Задержка | `EXPED_TYPE_DELAY` (6) | ❌ | Флот возвращается на 10–30% позже |
| Потери | `EXPED_TYPE_LOST` (7) | ✅ | Флот не возвращается (`EXPED_LOST_ENABLED`) |
| Быстрый возврат | `EXPED_TYPE_FAST` (8) | ❌ | Флот возвращается на 10–60% быстрее |
| Ничего | `EXPED_TYPE_NOTHING` (9) | ✅ | |
| Пираты / Стычка | `EXPED_TYPE_PIRATES` / `xSkirmish` (10) | ✅ частично | В legacy — бой с 1–3 флотами чужих, возможна нейтральная планета с добычей |
| Ресурсы | `EXPED_TYPE_RESOURCE` (11) | ✅ | |
| Корабли | `EXPED_TYPE_SHIP` (12) | ❌ | Флот противника присоединяется к игроку |
| Неизвестное | `EXPED_TYPE_UNKNOWN` (13) | ❌ | Зарезервировано |

Временная планета из экспедиции: живёт `12–24ч` (`EXPED_PLANET_LIFETIME_MIN/MAX`). Постоянная временная планета (`TEMP_PLANET_LIFETIME`) — 21 день.

#### Мощь и вероятности экспедиции (формулы legacy)

```
exp_power = expo_tech_level
          + expedition_hours × 2
          + spy_tech / 10 × pow(spy_probes, 0.4)
```

Штраф за повторные посещения одной системы (`visited_scale`):
- 3–4 раза → `× visits^(−0.2)`
- 5–9 раз → `× visits^(−0.3)`
- 10–19 раз → `× visits^(−0.5)`
- 20+ раз → `× visits^(−0.7)`
- Дополнительно × `((hours + 1) / 6)^1.5`

Базовые вероятности (растут с `exp_power`, нормализуются):
```
resourceDiscovery   = ceil(100 × 1.22^exp_power)
asteroidDiscovery   = ceil(70  × 1.22^exp_power)  [+30 если 0 часов]
shipsDiscovery      = ceil(20  × 1.25^exp_power)  [только hours >= 2]
battlefieldDiscovery= ceil(20  × 1.25^exp_power)  [только hours >= 1]
xSkirmish           = ceil(10  × 1.26^exp_power)  [только hours >= 3]
artefactDiscovery   = ceil(3   × 1.28^exp_power)  [только hours >= 4, cap = xSkirmish/2]
creditDiscovery     = ceil(4   × 1.28^exp_power)  [только hours >= 4, cap = xSkirmish/2]
delayReturn         = ceil(30  × 1.25^(visits + power/4)) [hours >= 1]
fastReturn          = ceil(60  × 1.25^(visits + power/4))
nothing             = ceil(40  × 1.25^(visits + power/4))
expeditionLost      = 10 (фиксированно)
```

Штрафы за крупный флот: >10 000 кораблей на планете → xSkirmish/ships/battlefield × 0.5; >100 Deathstar → xSkirmish × 0.1.

Каждый тип получает случайный множитель `×(1 ± 5%)`. С вероятностью 0.01% тип усиливается в 10 раз; с вероятностью 0.01% — исключается.

#### Ресурсный исход (RESOURCE/ASTEROID)
```
res_scale = max(0.5, (1 + pow(hours, 1.1)) × exp_power / 40 × visited_scale) × (1 ± 5%)
res_k     = random(500k, 1M) × res_scale × 2
            [2% шанс ×100 джекпот, cap = 10M × res_scale]
metal   = ceil(res_k)
silicon = ceil(res_k / 2 × (1 ± 10%))
hydrogen= ceil(res_k / 3 × (1 ± 10%))   [только RESOURCE, не ASTEROID]
```

#### Кредитный исход (CREDIT)
```
buy_credit = сумма покупок игрока за последние 3 дня
credit = random(
    10 + min(100, buy_credit/10) + exp_power/2,
    29 + min(300, buy_credit)    + exp_power×2
) × visited_scale × 0.7
credit = ceil(credit/10)×10 + random(5,9)   [округление до десятков]
```

#### Задержка / быстрый возврат
```
delay: время полёта += round(flight_time × random(0.10, 0.30))
fast:  время полёта -= round(flight_time × random(0.10, 0.60))
```

### Реферальная система

| Параметр | Значение |
|----------|----------|
| `REFERRAL_CREDIT_PERCENT` | 20% от покупок реферала |
| `REFERRAL_BONUS_POINTS` | 3 000 очков за реферала |
| `REFERRAL_MAX_BONUS_POINTS` | 500 000 очков |
| `REFERRAL_METAL_BONUS` | 10 металла |
| `REFERRAL_SILICON_BONUS` | 5 кремния |
| `REFERRAL_HYDROGEN_BONUS` | 2 водорода |

## Система инопланетян (Alien AI)

Исходный файл: `d:\Sources\oxsar2\www\game\AlienAI.class.php`, события обрабатываются в `ExtEventHandler.class.php`.

### State machine пришельцев

Полный цикл жизни одного флота пришельцев:

```
checkAlienNeeds()
  → generateMission()            — выбор цели и генерация флота
    → EVENT_ALIEN_FLY_UNKNOWN    — полёт к цели (15–24 часа)
      → onFlyUnknownEvent()
        ├─ попытка грабежа кредитов (10%)
        ├─ попытка подарка ресурсов (5%)
        ├─ попытка подарка кредитов (5%)
        ├─ 90% → EVENT_ALIEN_ATTACK → бой
        └─ 10% → EVENT_ALIEN_HALT
             → EVENT_ALIEN_HOLDING  (удержание 12–24 часа)
               → EVENT_ALIEN_HOLDING_AI  (цикл действий каждые N минут)
```

60% летящих флотов получают параллельный `EVENT_ALIEN_CHANGE_MISSION_AI` — смена цели в пути, сила флота растёт: `1 + control_times × 1.5`.

### Параметры флотов пришельцев

| Параметр | Нормально | День атаки (четверг) |
|----------|-----------|----------------------|
| Количество флотов | 50 (`ALIEN_NORMAL_FLEETS_NUMBER`) | 250 (`×5`) |
| Множитель силы | `random(0.9, 1.1)` — 90–110% от игрока | `random(1.5, 2.0)` — 150–200% |
| Время полёта | 15–24 часа | то же |
| Время удержания | 12–24 часа (до 15 дней реального времени) | то же |

Алгоритм подбора флота: итеративно добавляет корабли пришельцев (`UNIT_A_CORVETTE`…`UNIT_A_TORPEDOCARIER`) пока суммарная боевая сила не достигнет `target_power × power_scale`. Макс. обломков флота = `ALIEN_FLEET_MAX_DERBIS = 1 млрд`.

Корабли пришельцев (отдельные от кораблей игрока, id 200–204):
- `UNIT_A_CORVETTE` (200)
- `UNIT_A_SCREEN` (201)
- `UNIT_A_PALADIN` (202)
- `UNIT_A_FRIGATE` (203)
- `UNIT_A_TORPEDOCARIER` (204)

### Выбор цели

**Обычная атака** (`findTarget`): игрок активен в последние 30 мин, не в отпуске, >1 000 кораблей суммарно, >100 кораблей на конкретной планете, не атаковался за последние `ALIEN_ATTACK_INTERVAL = 6 дней`.

**Грабёж кредитов** (`findCreditTarget`): игрок активен 30 мин, >100 000 кредитов, >300 000 кораблей суммарно, >10 000 на планете, не грабился за 10 дней.

### Механика грабежа кредитов

При прилёте `EVENT_ALIEN_FLY_UNKNOWN`:
```
grab = round(user_credit × 0.0001 × random(0.08, 0.10))
       [фактически 0.0008–0.001% от кредитов]
```
После успешного грабежа — 90% вероятность улететь без атаки, игрок получает сообщение `MSG_CREDIT_ALIEN_GRAB`.

### Механика подарков (при прилёте без грабежа)

- **Ресурсы** (5%): дарит часть найденного на планете × `random(0.7, 1.0)`
- **Кредиты** (5%): `min(500, user_credit × 0.01 × random(5%, 10%))`

### Режим удержания (HOLDING)

Пришельцы занимают планету на 12–24 часа. В это время `EVENT_ALIEN_HOLDING_AI` циклически совершает одно из 8 действий (с равными весами 10:10:…):
1. Извлечь часть своих кораблей из флота
2. Выгрузить ресурсы игрока
3–8. Починить юниты / добавить юниты / добавить кредиты / добавить артефакт / создать астероид / найти планету после боя — **не реализованы** (пустые методы)

Количество изымаемых кораблей растёт квадратично по итерациям:
```
quantity = ceil(alien_count × 0.01 × pow(iteration, 2) × bonus_factor)
```

Если игрок платит кредиты за снятие удержания — время продлевается:
```
end_time = parent_event_time + 2ч × paid_credits / 50
```

### Что не реализовано в nova vs legacy

| Функция | Статус в nova |
|---------|---------------|
| Флот пришельцев масштабируется по силе игрока | ❌ (фиксированные 5 LF) |
| Корабли пришельцев (UNIT_A_*, id 200–204) | ❌ |
| День усиленных атак (четверг, ×5 флотов) | ❌ |
| Смена миссии в полёте (EVENT_ALIEN_CHANGE_MISSION_AI) | ❌ |
| Режим удержания планеты (HOLDING) | ❌ |
| Грабёж кредитов | частично (applyGrabCredit есть, но без state machine) |
| Подарки ресурсов и кредитов | ❌ |
| Корабли пришельцев как отдельная раса | ❌ |

## Параметры юнитов из БД легаси

**Источник истины для боевых характеристик юнитов — таблицы MySQL**, а не конфиги oxsar-nova. При расхождении между `configs/ships.yml` / `configs/defense.yml` и БД — доверять БД.

### Таблица `na_ship_datasheet` — боевые и ходовые параметры

Колонки: `unitid`, `capicity` (грузоподъёмность), `speed`, `consume` (расход топлива), `attack`, `shield`, `front`, `ballistics`, `masking`, `attacker_attack`, `attacker_shield`, `attacker_front`, `attacker_ballistics`, `attacker_masking`.

Колонки `attacker_*` — параметры юнита когда он выступает **атакующим** (могут отличаться от защитника). Пример: Deathstar `front=10` как защитник, `attacker_front=9` как атакующий.

Запрос для получения всех данных:
```sql
SELECT * FROM na_ship_datasheet ORDER BY unitid;
```

Актуальные данные (все юниты из БД):

| unitid | name | cargo | speed | fuel | attack | shield | front | ballistics | masking |
|--------|------|-------|-------|------|--------|--------|-------|------------|---------|
| 29 | Small Transporter | 5 000 | 5 000 | 10 | 5 | 10 | 10 | 0 | 0 |
| 30 | Large Transporter | 25 000 | 7 500 | 50 | 5 | 25 | 10 | 0 | 0 |
| 31 | Light Fighter | 50 | 12 500 | 20 | 50 | 10 | 10 | 0 | 0 |
| 32 | Heavy Fighter | 100 | 10 000 | 75 | 150 | 25 | 10 | 0 | 0 |
| 33 | Cruiser | 800 | 15 000 | 300 | 400 | 50 | 10 | 0 | 0 |
| 34 | Battleship | 1 500 | 10 000 | 500 | 1 000 | 200 | 10 | 0 | 0 |
| 35 | Frigate | 750 | 10 000 | 250 | 700 | 400 | 10 | 0 | 0 |
| 36 | Colony Ship | 7 500 | 2 500 | 1 000 | 50 | 100 | 10 | 0 | 0 |
| 37 | Recycler | 20 000 | 2 000 | 300 | 1 | 10 | 10 | 0 | 0 |
| 38 | Espionage Probe | 5 | 100 000 000 | 1 | 0 | 0 | 10 | 0 | 0 |
| 39 | Solar Satellite | 0 | 0 | 0 | 0 | 2 | 10 | 0 | 0 |
| 40 | Bomber | 500 | 4 000 | 1 000 | 900 | 550 | 10 | 0 | 0 |
| 41 | Star Destroyer | 2 000 | 5 000 | 1 000 | 2 000 | 500 | 10 | 0 | 0 |
| 42 | Deathstar | 1 000 000 | 100 | 1 | 200 000 | 50 000 | 10 (atk: 9) | 4 | 0 |
| 43 | Rocket Launcher | — | — | — | 80 | 20 | 10 | 0 | 0 |
| 44 | Light Laser | — | — | — | 100 | 25 | 10 | 0 | 0 |
| 45 | Heavy Laser | — | — | — | 250 | 100 | 10 | 0 | 1 |
| 46 | Ion Cannon | — | — | — | 150 | 500 | 10 | 1 | 1 |
| 47 | Gauss Cannon | — | — | — | 1 100 | 200 | 10 | 1 | 2 |
| 48 | Plasma Turret | — | — | — | 3 000 | 300 | 10 | 2 | 2 |
| 49 | Small Shield Dome | — | — | — | 1 | 10 000 | 15 | 0 | 0 |
| 50 | Large Shield Dome | — | — | — | 1 | 50 000 | 16 | 0 | 0 |
| 51 | Interceptor Missile | — | — | — | 1 | 1 | 10 | 0 | 0 |
| 52 | Interplanetary Missile | — | — | — | 32 000 | 1 | 10 | 10 | 0 |
| 102 | Lancer Ship | 200 | 8 000 | 100 | 5 500 | 200 | 8 (atk: 8) | 3 | 0 |
| 200 | Alien Corvette | 300 | 20 000 | 150 | 200 | 75 | 10 | 2 | 0 |
| 201 | Alien Screen | 800 | 10 000 | 75 | 22 | 5 000 | 15 (atk: 16) | 0 | 0 |
| 202 | Alien Paladin | 50 | 8 000 | 20 | 75 | 50 | 10 | 5 | 1 |
| 203 | Alien Frigate | 2 000 | 10 000 | 300 | 1 250 | 150 | 10 | 1 | 0 |
| 204 | Alien Torpedocarrier | 200 | 13 000 | 100 | 350 | 100 | 10 | 4 | 0 |
| 325 | Shadow Ship | 75 | 13 000 | 35 | 30 | 30 | 9 | 3 | 4 |
| 352 | Ship Transplantator | 2 000 000 | 4 000 | 400 000 | 500 | 20 000 | 12 | 0 | 0 |
| 353 | Ship Collector | 2 000 | 1 700 | 1 | 0 | 5 | 10 | 0 | 0 |
| 354 | Small Planet Shield | — | — | — | 100 | 300 000 | 22 | 0 | 0 |
| 355 | Large Planet Shield | — | — | — | 100 000 | 1 000 000 | 30 | 10 | 0 |
| 358 | Armored Terran | 10 000 | 3 250 | 100 000 | 330 | 15 000 | 24 | 0 | 0 |

> **Важно:** `shell` (броня) хранится **не** в `na_ship_datasheet`, а вычисляется через формулу из `na_construction` (`charge_*` поля). Броня юнита = суммарная стоимость (metal + silicon) поделённая на коэффициент.

### Таблица `na_rapidfire` — таблица rapidfire (полная)

Полная таблица из БД (существенно богаче, чем `configs/rapidfire.yml`):

```sql
SELECT unitid, target, value FROM na_rapidfire ORDER BY unitid, target;
```

Ключевые отличия от текущего `rapidfire.yml` в nova:
- **Frigate (35)**: имеет RF против Small/Large Transporter (×3), Heavy Fighter (×7), Cruiser (×4), Battleship (×7) — полностью отсутствует в nova
- **Bomber (40)**: RF против Rocket Launcher (×20), Light Laser (×20), Heavy Laser (×10), Ion Cannon (×10), Lancer (×5)
- **Star Destroyer (41)**: RF против Frigate (×2), Light/Heavy Laser и Lancer
- **Lancer (102)**: RF против Deathstar (×3) — отсутствует в nova
- **Plasma Turret (48)**: RF против Deathstar (×2) — оборона может уничтожать DS!
- **Alien корабли (200–204)**: своя таблица RF между собой и против игрока
- **Armored Terran (358)**: RF ×900 против всего — суперюнит

### Таблица `na_construction` — стоимости и формулы

Хранит формулы `prod_*`, `cons_*`, `charge_*` для всех зданий/кораблей/обороны в виде строк DSL (например: `{basic} * pow(2, ({level} - 1))`). Это **первичный источник** для `configs/construction.yml`.

```sql
SELECT buildingid, name, mode, basic_metal, basic_silicon, basic_hydrogen,
       charge_metal, charge_silicon, charge_hydrogen
FROM na_construction
WHERE mode IN (1,2,3,4)  -- 1=здания, 2=исследования, 3=флот, 4=оборона
ORDER BY mode, buildingid;
```

## База данных oxsar2 (MySQL)

| Параметр | Значение |
|----------|----------|
| Host | `mysql` (внутри Docker) / `localhost` (снаружи) |
| Port | 3306 |
| Database | oxsar_db |
| User | oxsar_user |
| Password | oxsar_pass |
| Root password | root |
