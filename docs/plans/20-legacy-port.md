# План 20: Порт механик из legacy oxsar2

**Дата**: 2026-04-24  
**Статус**: ЧЕРНОВИК  
**Источники**: `d:\Sources\oxsar2\www\game\`, `d:\Sources\oxsar2\www\ext\` (приоритет над base)  
**Затрагивает**: backend (Go), frontend (React/TS), configs/buildings.yml, migrations

> Все механики этого плана **реализованы в legacy oxsar2** (включая `ext/`).
> Задача — портировать, а не изобретать. Числа и формулы берутся из кода legacy
> без изменений; любое отклонение требует ADR.

---

## Что портируем и что нет

| Механика | Статус в legacy | Статус в nova | Приоритет |
|---|---|---|---|
| Vacation mode (umode) | ✅ полная реализация | Поля есть, логика не подключена | M |
| Fleet slots (computer_tech) | ✅ полная реализация | Исследование есть, лимит не проверяется | M |
| Миссия POSITION (kind=6) | ✅ полная реализация | enum есть, handler отсутствует | M |
| Сенсорная Фаланга | ✅ полная реализация | folder=11 есть, сканер не реализован | M |
| Stargate Jump (kind=32) | ✅ полная реализация (ext/) | enum есть, handler отсутствует | L |
| Moon Destruction (kind=25/27) | ✅ через attack (Assault.class.php) | enum есть, ветка в attack.go отсутствует | L |
| Astrophysics (ASTRO_TECH) | ⬜ нет в oxsar2 — новая фича из spec §12.5 | unit_id=112 зарезервирован | L |
| IGR network | ⬜ нет в oxsar2 — новая фича из spec §12.6 | unit_id=113 зарезервирован | L |

---

## Ф.1: Vacation Mode (umode)

**Legacy**: `Preferences.class.php`, `Functions.inc.php:548`  
**Ext-override**: нет

**Константы из legacy**:
- `VACATION_DISABLE_TIME = 60*60*24*30` — автоотключение через 30 дней (vacation не может висеть вечно)
- `LAST_TIME_ON_VACATION_DISABLE = 60*60*24*20` — после выхода нельзя повторно войти 20 дней
- `VACATION_BLOCKING_EVENTS` — полный список event kinds, блокирующих включение vacation:
  `BUILD_FLEET, BUILD_DEFENSE, TRANSPORT, POSITION, ATTACK_SINGLE, ATTACK_ALLIANCE, MOON_DESTRUCTION, ROCKET_ATTACK, REPAIR, DISASSEMBLE, RECYCLING, COLONIZE, SPY, HALT, EXPEDITION, HOLDING, STARGATE_TRANSPORT, STARGATE_JUMP, BUILD_CONSTRUCTION, DEMOLISH_CONSTRUCTION, RESEARCH`
- При включении: `setProdOfUser()` ставит `prod_factor = 0` на всех планетах пользователя
- Spec §18.8: минимум 48ч (`umode_min = now + 172800`)

**Что в коде**: `users.umode bool`, `users.umode_min timestamptz` — поля есть (migration 0001), но нигде не читаются.

**Реализация**:
- `POST /api/me/vacation` (включить): проверяем `COUNT events WHERE user_id=? AND state='wait' AND kind IN (BLOCKING_EVENTS)` = 0; выставляем `umode=true, umode_min=now+48h`; `evalProd` немедленно обнуляет производство
- `POST /api/me/vacation` (выключить): проверяем `umode_min < now`; выставляем `umode=false`, `last_vacation_end=now` (для cooldown 20 дней)
- `planet/service.go::evalProd`: добавить `if user.UMode { return 0 }`
- `fleet/attack.go::validateAttack`: добавить `if defender.UMode { return ErrVacationShield }`
- Worker `inactivity-checker`: если `umode=true AND last_seen_at < now-30d` → `umode=false` (автовыключение)
- UI: кнопка «Режим отпуска» в настройках/overview, индикатор + таймер до `umode_min`

**Связь**: блок A2 плана 17 (щит неактивности) реализуется поверх vacation — worker выставляет `umode=true` автоматически при 3+ днях отсутствия.

---

## Ф.2: Fleet Slots через Computer Tech

**Legacy**: `NS.class.php:1871`, `EventHandler.class.php:1442`  
**Ext-override**: нет

**Формула из legacy**:
```
maxSlots = 1 + floor(computer_tech / 6)
```
Если флот летит на планету союзника — суммируются `computer_tech` **обоих** игроков.

**Что не считается в слот** (из `isFleetSlotUsed()` в legacy):
- Expedition (kind=15)
- Artefact delivery (kind=29)
- Return events (kind=20, 21, 22)

**Что в коде**: `computer_tech` (research id=14) исследуется и хранится; слот-лимит нигде не проверяется.

**Реализация**:
- `fleet/service.go::validateSend`: добавить подсчёт `COUNT fleets WHERE user_id=? AND state='outbound' AND kind NOT IN (15, 29, 20, 21, 22)` ≥ `1 + floor(researchLevel(14)/6)` → `ErrFleetSlotsExceeded`
- UI: индикатор «Слоты флота: N / M» в FleetScreen

---

## Ф.3: Миссия POSITION (Перебазирование, kind=6)

**Legacy**: `EventHandler.class.php:2067`  
**Ext-override**: нет

**Логика `arrive` из legacy**:
1. Если есть `pathstack` (halt-цепочка) — выполнить halt, не продолжать
2. Разгрузить груз флота (metal/silicon/hydrogen) на планету назначения
3. Перенести корабли из `fleet_ships` в `shipyard` назначения (INSERT/UPDATE — суммировать со стеком)
4. Отцепить арtefакты от флота
5. Отправить сообщение `MSG_POSITION_REPORT` владельцу

**Ограничения**:
- Только на свои планеты/луны **или** планеты союзника (relations ALLY/NAP)
- Флот не возвращается (нет `KindFleetReturn` после arrive)

**Что в коде**: `kind=6` в enum событий, но в `worker/main.go` нет кейса для него; нет `fleet/position.go`.

**Реализация**:
- `fleet/position.go::ArriveHandler`: реализовать шаги 1–5 выше; корабли — `INSERT INTO ships ... ON CONFLICT (planet_id, unit_id) DO UPDATE SET quantity = quantity + excluded.quantity`
- Проверка при отправке: целевая планета принадлежит игроку или союзнику
- Зарегистрировать handler в `worker/main.go`
- UI: добавить опцию «Перебазирование» (POSITION) в FleetScreen — только для своих/союзных координат

---

## Ф.4: Сенсорная Фаланга (Sensor Phalanx / Star Surveillance)

**Legacy**: `MonitorPlanet.class.php`, `NS.class.php:1109`  
**Ext-override**: нет

**Константы из legacy**:
- `UNIT_STAR_SURVEILLANCE = 55` — id здания
- `STAR_SURVEILLANCE_CONSUMPTION = 5000` — водород за скан

**Формула радиуса** (`NS.class.php:1109`):
```
range = round((level^2 - 1) * (1 + hyperspace_tech / 10))
// уровень 1, hyperspace 0: range = 0 (только своя система)
// уровень 3, hyperspace 5: range = round(8 * 1.5) = 12 систем
```

**Что возвращает скан**: все fleet-события в целевой системе — тип миссии, состав флота, количество кораблей, ETA, перевозимые ресурсы.

**Что в коде**: `folder=11` (Phalanx) в messages есть. `UNIT_STAR_SURVEILLANCE` (id=55) отсутствует в `buildings.yml`. Это отмечено в `simplifications.md` как приоритет M.

**Реализация**:
- Добавить `star_surveillance` (id=55, `moon_only: true`) в `buildings.yml` с формулой стоимости из legacy
- `GET /api/phalanx?target_g=X&target_s=Y&source_planet_id=Z`:
  - Проверить: source_planet — луна текущего игрока с `star_surveillance >= 1`
  - Проверить: та же галактика, `|source_system - target_system|` ≤ range
  - Списать 5000H с планеты-источника
  - Вернуть `[]FleetScan` — fleet events WHERE `dst_galaxy=X AND dst_system=Y OR src_galaxy=X AND src_system=Y`
- Отправить сообщение в folder=11 каждому владельцу просканированного флота
- UI: кнопка «🔭 Скан» в GalaxyScreen на строке системы (видна только при наличии луны с phalanx в этой галактике)

---

## Ф.5: Stargate Jump (kind=32)

**Legacy**: `ExtMission.class.php:185`, `ExtEventHandler.class.php:630`, `Functions.inc.php:1693`  
**Ext-override**: **есть** (`ExtMission` + `ExtEventHandler`)

**Формула cooldown** (`Functions.inc.php:1693`):
```
cooldown_sec = 3600 * 0.7^(max(0, jump_gate_level - 1))
// уровень 1: 3600с (1ч) | уровень 2: 2520с (42м) | уровень 3: 1764с (29м)
```

**Запрещённые юниты** (из `consts.php:315`): щиты (small/large/planet shield), ракеты (interceptor, interplanetary), exchange-слоты.

**Ограничения по уровню планеты** (`ExtMission.class.php:235`):
- Планета позиции 1–2: прыжок невозможен
- Позиция 3: только на луны
- Позиция 4+: на любые stargate

**Логика прыжка**: Fleet прилетает мгновенно (`fire_at = now`), корабли разгружаются как POSITION. Cooldown записывается в `stargate_jump` таблицу.

**Что в коде**: kinds 28/32 в enum, handlers отсутствуют. `jump_gate` не в `buildings.yml`.

**Реализация**:
- Добавить `jump_gate` (id=56, `moon_only: true`) в `buildings.yml`
- Миграция: `stargate_cooldowns(planet_id PRIMARY KEY, last_jump_at TIMESTAMPTZ)`
- `fleet/stargate.go::SendHandler`: validate (обе луны `jump_gate >= 1`, cooldown, запрещённые юниты, ограничение позиции); создать event kind=32 с `fire_at = now` (мгновенно)
- `fleet/stargate.go::ArriveHandler`: как POSITION arrive + записать cooldown
- UI: вкладка «Старгейт» в FleetScreen или отдельный экран `/stargate`

---

## Ф.6: Уничтожение Луны (Moon Destruction, kind=25/27)

**Legacy**: `Assault.class.php:505–642`  
**Ext-override**: **есть** для alliance-варианта (`ExtEventHandler.class.php:704`)

**Как работает в legacy**: `EVENT_MOON_DESTRUCTION` (kind=14) — мёртвый event, handler пустой. Реальная механика — через attack: kinds `EVENT_ATTACK_DESTROY_MOON=25` и `EVENT_ATTACK_ALLIANCE_DESTROY_MOON=27` маршрутизируются в `Assault.class.php`, который после боя применяет moon-logic.

**Логика** (`Assault.class.php:628–642`):
```
if target_moon:
    UPDATE planet SET userid=NULL WHERE id=moon_id
    UPDATE galaxy SET moon_id=NULL WHERE moon_id=moon_id
    AutoMsg(MSG_MOON_DESTROYED=56, owner_id)
```

**Формула уничтожения** (OGame-стандарт, из spec §12):
```
P_destroy_moon  = clamp((100 - sqrt(diameter)) * sqrt(rip_count), 0, 100)  %
P_destroy_fleet = clamp((100 - sqrt(rip_count)) * sqrt(diameter) / 200, 0, 100)  %
```
Точное место формулы в `Assault.class.php` — уточнить при реализации.

**Ограничение**: только корабли с `rip=true` (Deathstar, id=42) участвуют в moon destruction roll. Обычный бой проходит нормально.

**Что в коде**: kinds 25/27 в enum; `attack.go` обрабатывает `kind=10` и `kind=12`; `planets.is_moon` есть.

**Реализация**:
- В `fleet/attack.go::AttackHandler`: добавить ветку `if kind IN (25, 27)`:
  - После обычного боя: посчитать `rip_count` (DS выживших атакующих), взять `moon.diameter`
  - Roll `P_destroy_moon` через `rng.New(seed)`, roll `P_destroy_fleet` через тот же rng
  - Если луна уничтожена: `UPDATE planets SET deleted_at=now() WHERE id=? AND is_moon=true`, `UPDATE galaxy SET moon_id=NULL`; debris = сумма стоимостей построек луны × 50%
  - Если флот уничтожен: удалить fleet_ships DS-кораблей
- ACS-вариант (kind=27): alliance-участники объединяются, `rip_count` суммируется по всем флотам
- Сообщение `MSG_MOON_DESTROYED` обоим игрокам

---

## Ф.7: Astrophysics (ASTRO_TECH, id=112) — новая фича, не порт

**Legacy**: аналога нет в oxsar2. Spec §12.5 описывает как OGame-фичу.

**Формулы из spec**:
```
expedition_slots = max(1, floor(sqrt(astro_level)))
colony_limit     = floor(astro_level / 2) + 1
```

**Breaking change**: без ASTRO_TECH игроки не смогут колонизировать после введения лимита.  
**Решение**: дать всем игрокам стартовый `astro_level=2` при миграции (даёт 2 колонии, 1 экспедицию).

**Реализация**:
- Добавить `ASTRO_TECH` (id=112) в `research.yml`, требование `expo_tech >= 3`
- Миграция: `UPDATE research SET level=2 WHERE unit_id=112` для всех игроков без ASTRO_TECH
- `fleet/expedition.go::Validate`: `COUNT outbound expeditions WHERE kind=15` ≥ `floor(sqrt(astro_level))` → `ErrExpeditionSlotsFull`
- `fleet/colonize.go::Validate`: `COUNT planets WHERE user_id=?` ≥ `floor(astro_level/2)+1` → `ErrColonyLimit`
- **Требует ADR** — breaking change для текущих игроков

---

## Ф.8: Интергалактическая исследовательская сеть (IGR_TECH, id=113) — новая фича

**Legacy**: частично через `research_factor` artефакта. Spec §12.6: `research_network_level` (id=113) зарезервирован.

**Формула**:
```
effective_lab_level = sum(level ORDER BY level DESC LIMIT igr_level)
// из всех планет пользователя, топ-N по уровню research_lab
```

**Реализация**:
- Добавить `IGR_TECH` (id=113) в `research.yml`, требование `research_lab >= 10, expo_tech >= 5`
- `research/service.go::calcDuration`: при `igr_level > 1` — SELECT топ-N `research_lab` уровней WHERE `user_id=?`, сумма вместо одного уровня
- Стимул развивать лаборатории на нескольких планетах

---

## Порядок реализации (рекомендация)

| Фаза | Задача | Приоритет | Зависимость |
|---|---|---|---|
| 19.1 | Ф.2 Fleet slots (computer_tech) | M | — |
| 19.2 | Ф.1 Vacation mode | M | — |
| 19.3 | Ф.3 Миссия POSITION | M | — |
| 19.4 | Ф.4 Фаланга | M | Нужна луна + здание id=55 |
| 19.5 | Ф.6 Moon Destruction | L | Нужен Ф.3 (позиция луны) |
| 19.6 | Ф.5 Stargate | L | Нужен Ф.3 (POSITION arrive как основа) |
| 19.7 | Ф.7 Astrophysics | L | ADR обязателен |
| 19.8 | Ф.8 IGR network | L | После Ф.7 |

---

## ADR-требования

- **Ф.1**: минимальный порог vacation 48ч и cooldown 20 дней — из spec §18.8, не менять без ADR
- **Ф.7**: breaking change — лимит колоний для текущих игроков; стартовый `astro_level=2` как решение
- **Ф.6**: debris от луны (50% стоимости построек) — проверить формулу в oxsar2 перед реализацией

---

## Что НЕ делаем в этом плане

- Не меняем балансовые формулы
- Не добавляем новые типы юнитов
- Lifeforms — отдельно (v2, флаг в конфиге)
- Платёжная система — план 07
