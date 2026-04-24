# План 15 — AlienAI: День четверга, HALT/HOLDING, платёж, защита планеты

## Контекст

В legacy oxsar2 пришельцы не просто атакуют — у них цельный цикл
«атака → удержание планеты → постепенный уход (или платёж игрока для
продления)». В nova базовый спавн (`EVENT_ALIEN_ATTACK`) уже работает
(план 01, раздел A.3). Не реализованы: день четверга с ×5 флотов,
переход в HALT/HOLDING после победы, HOLDING_AI с действиями,
«наёмная защита» (alien-флот защищает планету игрока от чужих атак),
платёж кредитами за продление HOLDING, блокировка новых alien-атак
на планету в HOLDING, `CHANGE_MISSION_AI`, `GRAB_CREDIT`,
`FLY_UNKNOWN`.

## Результат исследования legacy

### Источники
- Основная логика: `d:\Sources\oxsar2\www\game\AlienAI.class.php`
  (1127 строк) и `Assault.class.php`.
- Расширения: `d:\Sources\oxsar2\www\ext\ExtEventHandler.class.php`,
  `ext\consts.dm.local.php`, `ext\cronjobs\` — проверено, **не
  переопределяют ALIEN-логику**. `ExtEventHandler` просто
  делегирует в `AlienAI::on*Event()`. `consts.dm.local.php` содержит
  только платёжные ключи (PAY_2_*), ALIEN-константы не трогает.
  `cronjobs/database_check.php` — только housekeeping БД.
- Java-движок: `oxsar2-java/Participant.java`, `Assault.class.php:207-219`
  подтягивает alien-флот в defender через `EVENT_ALIEN_HOLDING` при
  `loadDefenders()`.

### Полный список event-типов пришельцев в legacy

| ID | Имя | Триггер | Эффект |
|----|-----|---------|--------|
| 33 | `EVENT_ALIEN_FLY_UNKNOWN` | `generateMission()` fallback (90% вероятность) | Неопределённая миссия — превращается в ATTACK / GRAB_CREDIT / тихо завершается |
| 34 | `EVENT_ALIEN_HOLDING` | Из `EVENT_ALIEN_HALT` через `duration` (12–24ч) | Пришельцы удерживают планету. Спавнит `HOLDING_AI` каждые 12–24ч |
| 35 | `EVENT_ALIEN_ATTACK` | Cron, каждые 6 дней; четверг = ×5 флотов | Бой против планеты. Флот масштабируется по `calcDefPower` |
| 36 | `EVENT_ALIEN_HALT` | Сразу после победы пришельцев в бою | Переходное состояние 12–24ч перед HOLDING |
| 37 | `EVENT_ALIEN_GRAB_CREDIT` | Альтернатива ATTACK для игроков с кредитом > 100k и флотом > 300k | Забирает 8–10% кредита; интервал 10 дней на игрока |
| 38 | `EVENT_ALIEN_ATTACK_CUSTOM` | Ручной спавн (отладка/админ-события) | Фиксированный флот |
| 80 | `EVENT_ALIEN_HOLDING_AI` | Из `HOLDING` через 5–10 сек, затем каждые 12–24ч | Случайное действие из 8 (в легаси реализовано 1–2): выгрузка ресурсов, извлечение alien-кораблей, и заглушки |
| 81 | `EVENT_ALIEN_CHANGE_MISSION_AI` | 60% вероятность перед атакой; задержка 8–10ч | Переоценка цели в полёте; может поменять ATTACK ↔ FLY_UNKNOWN, увеличивает силу флота ×1.5 |

### Константы (consts.php строки 752–770)

```
ALIEN_NORMAL_FLEETS_NUMBER       = 50
ALIEN_ATTACK_TIME_FLEETS_NUMBER  = 250  (= 50 × 5 в четверг)
ALIEN_ATTACK_INTERVAL            = 6 * 24 * 3600   (6 дней)
ALIEN_GRAB_CREDIT_INTERVAL       = 10 * 24 * 3600  (10 дней)
ALIEN_HALTING_MIN_TIME           = 12 * 3600       (12ч)
ALIEN_HALTING_MAX_TIME           = 24 * 3600       (24ч)
ALIEN_HALTING_MAX_REAL_TIME      = 15 * 24 * 3600  (15 дней)
ALIEN_CHANGE_MISSION_MIN_TIME    = 8 * 3600        (8ч)
ALIEN_CHANGE_MISSION_MAX_TIME    = 10 * 3600       (10ч)
ALIEN_GRAB_MIN_CREDIT            = 100_000
ALIEN_GRAB_CREDIT_MIN_PERCENT    = 0.08
ALIEN_GRAB_CREDIT_MAX_PERCENT    = 0.10
Power Scale Thursday             = randFloatRange(1.5, 2.0)
```

### Семантика HOLDING — гибрид «оккупация + наёмная защита»

Подтверждено кодом:

1. **Защита от других атак**. `Assault::loadDefenders` (строки 207–219)
   подтягивает alien-флот через активный `EVENT_ALIEN_HOLDING` как
   defender. Java-движок помечает `isAliens=true`
   (`Participant.java:20`). При атаке другого игрока пришельцы
   сражаются на стороне владельца планеты.

2. **Блокировка новых alien-атак**. `AlienAI::findTarget()` (строка
   361) исключает планеты с активным HOLDING из выбора целей в
   течение `ALIEN_ATTACK_INTERVAL`.

3. **Платёж за продление** (AlienAI.class.php строки 934–994):
   ```php
   $end_time = $parent_event["time"] + 60*60*2 * $paid_credit / 50.0;
   $end_time = min($end_time, $parent_event["start"] + ALIEN_HALTING_MAX_REAL_TIME);
   ```
   2 часа на каждые 50 кредитов, cap 15 дней от начала HOLDING.
   `paid_credit` приходит в `event.data` извне — т.е. точка входа
   платежа должна быть в нашем API.

4. **Уход без платежа**. `onHoldingEvent` возвращает `true` с
   комментарием "alien go away" — пришельцы тихо уходят по истечении
   duration, планета возвращается игроку без потерь сверх тех, что
   забрал HOLDING_AI.

## Скоуп плана

### Этап 1 — День четверга + HALT/HOLDING + защита — ЗАКРЫТО

Коммиты: f411e82, 755f678, 4c4920e, fd9c858.


- [ ] **Четверг-триггер** в спавне: `time.Now().Weekday() == time.Thursday`
  → `ALIEN_ATTACK_TIME_FLEETS_NUMBER=250` вместо `50`, сила
  `scaledAlienFleet` × `rand(1.5, 2.0)`.
  - Файл: `backend/internal/alien/alien.go` (`SpawnAttack`).
  - Константы: новый файл `backend/internal/alien/consts.go` с
    `AlienNormalFleetsNumber=50`, `AlienAttackTimeFleetsNumber=250`,
    `ThursdayPowerMin=1.5`, `ThursdayPowerMax=2.0`.

- [ ] **Kind-константы** в `backend/internal/event/kinds.go`:
  `KindAlienHalt=36`, `KindAlienHolding=34`, `KindAlienHoldingAI=80`.

- [ ] **Handler `KindAlienHalt`**: через 12–24ч после победы
  пришельцев создать `EVENT_ALIEN_HOLDING`. Сохранить в `event.data`
  флот пришельцев (корабли + количество).
  - Интеграция: в `fleet/attack.go` (или где фиксируется победа
    пришельцев) — спавн HALT-события.

- [ ] **Handler `KindAlienHolding`**: при `fire_at` создать первый
  `HOLDING_AI` с задержкой 5–10 сек. Запомнить `start_time`,
  `duration`, `alien_fleet`, `holding_eventid` в `data`.

- [ ] **Handler `KindAlienHoldingAI`**: тикает каждые 12–24ч. В
  рамках Этапа 1 — реализовать **только одно действие**:
  `onUnloadAlienResoursesAI` (выгрузка 7–10% захваченных ресурсов на
  склад планеты игрока). Остальные 7 — как заглушки / no-op с
  комментарием «reserved, plan 15 этап 2».

- [ ] **HOLDING-флот как defender в бою**. В `fleet/attack.go` (или
  `battle/engine.go` прелоадер) при атаке игрока X на планету Y
  найти активный `KindAlienHolding` с `destination=Y` и подтянуть
  alien-флот в `Participants` на defender-стороне (с флагом
  `is_aliens=true`). Если alien-флот погибает в бою — `HOLDING`
  закрывается (state=done), планета освобождается.

- [ ] **Расширить анти-спам в `Spawn`**. Сейчас
  `backend/internal/alien/alien.go` фильтрует события по `planet_id`
  и только `kind=35`. В легаси (`AlienAI::findTarget`, строка 358)
  фильтр работает **по `user_id`** и учитывает **все 4 типа**:
  `FLY_UNKNOWN | ATTACK | HOLDING | HALT`. Перевести на тот же
  принцип: пока у игрока есть любое активное alien-событие на
  любой его планете за последние 6 дней — новой атаки не спавним.
  Отдельной «блокировки HOLDING-планет» не требуется — это
  естественно получается из этого фильтра.

- [ ] **API платежа**: `POST /api/alien/holding/{event_id}/pay`
  body `{ "credit": N }`.
  - Списать `N` кредитов со счёта игрока (через кредитную систему
    из плана 06).
  - Обновить `event.fire_at += 2h * N / 50`, cap
    `start_time + 15 дней`.
  - Записать `paid_credit += N`, `paid_times += 1` в data.
  - Вернуть новое `fire_at` и остаток кредитов.
  - Ошибки: 404 если event не HOLDING / не активен, 402 если
    кредитов недостаточно, 409 если уже на cap.

- [ ] **Сообщения игроку**:
  - `MSG_ALIEN_HALTING` — пришельцы блокировали планету (при HALT →
    HOLDING).
  - `MSG_ALIEN_RESOURSES_GIFT` — выгрузка ресурсов.
  - `MSG_ALIEN_DEFENDED` (новый, в легаси явного нет) — пришельцы
    отбили атаку X; победа/поражение.
  - `MSG_ALIEN_DEPARTED` — пришельцы ушли, планета свободна.

- [ ] **Тесты**:
  - `alien_test.go`: четверг → ×5 флотов + power boost.
  - `event/handlers_alien_test.go`: HALT → HOLDING → HOLDING_AI
    цепочка, ресурсы зачисляются, event замыкается по таймеру.
  - `fleet/attack_test.go`: атака на планету с HOLDING подтягивает
    alien-defender, его смерть закрывает HOLDING.
  - `alien/pay_test.go`: платёж продлевает, cap 15 дней, 402 при
    нехватке кредитов.

### Этап 2 — Полный HOLDING_AI — ЗАКРЫТО частично (commit 1367332)

Проверка legacy (`AlienAI.class.php:1086–1126`) показала: из 8 действий
HOLDING_AI **6 имеют пустые тела** — в игре они никогда ничего не
делали. Реально работают только 2:

- [x] `onUnloadAlienResoursesAI` — ресурсы игроку (шаг 2, commit 755f678).
- [x] `onExtractAlientShipsAI` — убывание alien-флота, закрытие HOLDING
  при полном уходе (commit 1367332).
- [-] `onRepairUserUnitsAI`, `onAddUserUnitsAI`, `onAddCreditsAI`,
  `onAddArtefactAI`, `onGenerateAsteroidAI`, `onFindPlanetAfterBattleAI` —
  в legacy no-op. Портировать нечего. Зафиксировано в
  `docs/simplifications.md`.

### Этап 3 — Прочие alien-события — ПРОПУЩЕН

После детального анализа legacy принято решение не делать:

- **`KindAlienChangeMissionAI` (81)** — смена миссии **в полёте**.
  У нас в nova пришельцы не имеют промежуточного состояния «в полёте»:
  событие ATTACK создаётся сразу с `fire_at = now+flight`, а
  `AttackHandler` отрабатывает бой транзакционно за одну итерацию.
  Механика «поменять цель пока летят» нам не нужна.

- **`KindAlienGrabCredit` (37)** — отдельный event-тип для кражи
  кредитов у богатых игроков. У нас `applyGrabCredit` уже вызывается
  в `AttackHandler` при победе пришельцев (commit aedc65b в плане 01
  A.3). Отдельный event-тип = дублирование.

- **`KindAlienFlyUnknown` (33)** — переходная миссия, которая в legacy
  при прибытии превращается в ATTACK/GRAB_CREDIT/тихий уход.
  У нас эту роль выполняет сам `KindAlienAttack` с 30%-рандомизацией
  спавна в `Spawn`.

Три константы (`KindAlienFlyUnknown=33`, `KindAlienGrabCredit=37`,
`KindAlienChangeMissionAI=81`) добавлены в `event/kinds.go` как
«зарезервированные» — пригодятся, если в будущем понадобится более
точный порт.

### Этап 4 — UI (+1 день)

- [ ] Экран планеты: когда активен HOLDING — показывать alien-флот,
  таймер до ухода, кнопку «Заплатить за продление» с ползунком
  суммы.
- [ ] Индикатор в галактическом обзоре: планеты в HALT/HOLDING
  видны другим игрокам как «занято пришельцами».
- [ ] Сообщения в почте отрендерены с кнопками действий.

## Открытые вопросы (решить до старта)

1. **Валюта платежа**. В legacy — `credit` (отдельная валюта). В
   nova кредиты уже есть (план 06 — AI Advisor, план 07 — payments).
   Подтвердить: используем тот же баланс? Или вводим отдельный
   «alien-credit»?

2. **UI-кнопка платежа**: делаем в Этапе 1 или откладываем в Этап 4?
   Этап 1 только backend+API — игроки могут тестировать через curl.

3. **Cap 15 дней**: считать от `start_time` HOLDING или от момента
   захвата планеты (HALT)? Legacy считает от `parent_event["start"]`
   — это start HOLDING (parent = HOLDING для HOLDING_AI).

4. **Убивать alien-defender по rapidfire?** В legacy alien-флот
   стоит на орбите и участвует в бою как обычный defender —
   применяются те же формулы `battle.engine`. Подтвердить, что наша
   боевая модель обработает `is_aliens=true` без специальной логики.

5. **День четверга — по какому часовому поясу?** В legacy сервер
   работает в одной TZ. У нас — UTC в БД. Четверг в UTC или по
   Europe/Moscow (чтобы игроки-русские чувствовали «четверг» как
   четверг)?

## Критерии готовности Этапа 1

- В четверг UTC спавнится ×5 alien-атак с boost силой.
- После победы пришельцев на планете появляется таймер ухода
  12–24ч (HALT), затем HOLDING до 15 дней макс.
- Каждые 12–24ч на склад игрока капает 7–10% «добычи» пришельцев.
- Чужая атака на планету в HOLDING — бой идёт против
  игрок+пришельцы; если alien-флот выживает, HOLDING продолжается.
- `POST /api/alien/holding/{id}/pay` продлевает HOLDING по формуле
  `2h * credit/50`.
- Генерация новых alien-атак на планету с активным HOLDING
  заблокирована.
- Сообщения отправляются в 4 ключевых точках цикла.
- Все тесты (≥ 85% для `alien/` и `event/`) зелёные.

## Связи с другими планами

- [План 01 — backend combat](01-backend-combat.md) §A.3 — базовый
  AlienAI уже реализован (calcDefPower, scaledAlienFleet). Этот
  план — продолжение.
- [План 06 — credits-ai-advisor](06-credits-ai-advisor.md) —
  кредитная система как валюта платежа.
- [План 09 — event-system](09-event-system.md) — новые Kind'ы
  регистрируются в worker.
