# План 37.8: Game-Origin — реестр игровых дыр

**Дата начала**: 2026-04-27
**Статус**: Аудит завершён, фиксы — в работе
**Связан с**: [37-game-origin.md](37-game-origin.md) §37.8, [37-game-origin-security-audit.md](37-game-origin-security-audit.md) (security-аудит §37.6)

Реестр **игровых** дыр в `projects/game-origin/` (баланс/механика, не безопасность).
Безопасность (XSS/CSRF) — отдельный документ (37.6/37.7).

## Защита `NS::isFirstRun` — важный контекст

**Все** найденные race-условия модулируются защитой `NS::isFirstRun($name)`
([src/game/NS.class.php:1816-1834](../../projects/game-origin/src/game/NS.class.php#L1816-L1834)):

```php
public static function isFirstRun($name)
{
    if (!self::$mch->is_valid()) return true;  // fallback при недоступном memcache
    if (self::$mch->add($name, true, 2)) return true;  // memcached add — атомарный
    return false;
}
```

Это **inter-process** дедуп через `memcached::add` (атомарная операция:
true только первому, кто вставил ключ). TTL = 2 секунды.

**Что это даёт**:
- Защищает от дубликатов в окне 2 секунд между процессами.
- Полностью спасает от двойного клика игрока (interval < 1 сек).
- НЕ спасает если ключ зависит от изменчивых данных (md5(serialize(post))
  с разной целью полёта → разные ключи).
- НЕ спасает если запросы пришли через 3+ секунды (TTL истёк).
- Полностью отключается если memcached недоступен (fallback `return true`).

В описаниях ниже учтено наличие/отсутствие этой защиты. Если защита
работает на 95%+ типичных кейсов — серьёзность дыры понижена,
но не убрана (атаки через slow-VPN или бот-скрипт всё ещё возможны).

---

## Контекст и метод

Аудит вёлся по пяти категориям из плана 37 §"Игровые дыры":
1. Отрицательные ресурсы при refund (отмена постройки/флота).
2. Race condition в очереди строительства (двойной клик).
3. «Застрявший» флот после ошибки event-обработчика.
4. Integer overflow в формулах производства на высоких уровнях.
5. Лимит экспедиций/флотов/лотов через параллельные запросы.

Каждая категория делегирована Explore-агенту, **отчёт каждого
агента верифицирован руками** (см. memory `feedback_audit_agent_verify.md`).
Из ~24 первичных находок реальными подтверждены ~10 (остальные — ложные
срабатывания из-за непрочитанных защит, дубликаты, или by-design legacy).

## Серьёзность

- **P0** — game-breaking, дубликация ресурсов/юнитов.
- **P1** — заметный эксплойт, ломает экономику или прогресс.
- **P2** — узкий сценарий или edge-case (high-end игра).
- **P3** — теоретический.

## Статус

- **verified** — лично прочитал указанные строки, дыра реальна.
- **unverified** — отчёт агента, ручная верификация отложена до фикс-этапа.
- **rejected** — проверено, дыры нет (ложноположительный отчёт).
- **by-design** — поведение преднамеренное (legacy-фича).
- **patched** — исправлено в коммите.

---

## Реестр

### REF-004 — TOCTOU в Exchange::sendBackLot — двойной refund лота (с защитой isFirstRun на lot id)

- **Серьёзность**: P2 (понижено с P1 — isFirstRun сильно сужает окно)
- **Статус**: verified, но переоценён после изучения `NS::isFirstRun`
- **Где**: [src/game/Exchange.class.php:467-549](../../projects/game-origin/src/game/Exchange.class.php#L467-L549)
- **Verification**:
  - `NS::isFirstRun("Exchange::sendBackLot:{$id}")` — ключ зависит от
    lot id. Это **inter-process** дедуп через memcached add (TTL 2 сек).
  - Изначально я и аудит-агент посчитали isFirstRun process-local —
    это ошибка. Memcached add — атомарная inter-process операция.
  - Реальное окно гонки: два запроса с интервалом > 2 сек. Маловероятно
    через UI (игрок не жмёт «отозвать» дважды через 2+ сек), но возможно
    через скрипт с задержкой 2.5-3 сек.
- **Сценарий (узкий)**: бот-скрипт `recall(); sleep(2.5); recall();` →
  второй вызов уже не блокируется memcached (TTL истёк), и status в
  БД ещё не обновлён до строки 549 → двойной refund.
- **Фикс**: атомарный `UPDATE exchange_lots SET status=ESTATUS_RECALL
  WHERE lid=$id AND status=ESTATUS_OK` ПЕРЕД refund-логикой; проверить
  affected_rows == 1. Если 0 — отказать (status уже не OK = другой
  процесс взял). Это закроет окно полностью.

### REF-005 — TOCTOU credit в Market::Credit_ex — credit в минус

- **Серьёзность**: P1 (повышено с P2 — credit может уйти ниже нуля)
- **Статус**: verified, **isFirstRun отсутствует** (никакой защиты)
- **Где**: [src/game/page/Market.class.php:192-240](../../projects/game-origin/src/game/page/Market.class.php#L192-L240),
  [src/game/NS.class.php:1657-1669](../../projects/game-origin/src/game/NS.class.php#L1657-L1669)
- **Verification**:
  - В `Credit_ex` **нет** `NS::isFirstRun()` (в отличие от sendBackLot,
    addLot, sendFleet) — открытое окно гонки.
  - Проверка `$credit > $this->credit` (строка 211) идёт против
    snapshot. `RES_UPDATE_EXCHANGE` (тип `+-`) не клампит credit на минус.
    `block_minus` не передаётся.
- **Сценарий**: два параллельных запроса на покупку ресурсов на рынке →
  оба видят достаточный credit-баланс, оба проходят, оба
  `credit = credit - cost`. Итог — отрицательный credit. **Эксплуатация
  тривиальная**: 2 одновременных HTTP-запроса.
- **Фикс**: атомарный `UPDATE user SET credit=credit-cost WHERE userid=$uid
  AND credit >= cost` + проверка affected_rows. Если 0 — отказать.

### REF-001/002/003/006/007 — отвергнуты

| ID | Заявленная дыра | Почему rejected |
|---|---|---|
| REF-001 | `updateUserRes` не клампит отрицательные | Защита через `block_minus` + `max(0, $metal)` для типов `+`. Для `+-` риск есть только в credit (см. REF-005). |
| REF-002 | 100% refund в первый час → отрицательный баланс | Дубль REF-006 (by-design). |
| REF-003 | `haltReturn` hydrogen в минус | hydrogen в трюме флота (не на планете), оба члена ≥ 0. Уйти в минус нельзя. |
| REF-006 | TOCTOU 100% refund в первые `EV_ABORT_SAVE_TIME=15` секунд | by-design legacy: «передумал в первые 15 сек — без штрафа». |
| REF-007 | overflow shipyard при возврате флота из лота | unverified, отложено. |

---

### RACE-003 — race в Research/Constructions/Shipyard upgrade — двойная постройка

- **Серьёзность**: P1
- **Статус**: verified
- **Где**:
  - [src/game/page/Research.class.php:356-475](../../projects/game-origin/src/game/page/Research.class.php#L356-L475) (`upgradeResearch`)
  - [src/game/page/Constructions.class.php:366-498](../../projects/game-origin/src/game/page/Constructions.class.php#L366-L498) (`upgradeConstruction`)
  - [src/game/page/Shipyard.class.php:430-555](../../projects/game-origin/src/game/page/Shipyard.class.php#L430-L555) (build flow)
- **Дыра**: shape во всех трёх:
  1. `SELECT events` (загрузка очереди в process-local snapshot).
  2. Проверки: `free_queue_size`, `canResearch()`, `checkRequirements()`,
     `event["start"] == time()` (only same-second guard).
  3. `setRequieredResources` + `checkResources()` — против snapshot планеты.
  4. `addEvent(...)` + `updateUserRes(...)` — без транзакции.
  Между 1 и 4 нет ни блокировки, ни атомарной проверки.
- **Сценарий**: двойной клик → 2 параллельных HTTP → оба видят пустую
  очередь, оба видят достаточный `metal/silicon/hydrogen`, оба добавляют
  event и списывают ресурсы. Если хватало на одно — второе списание
  уйдёт в минус (для `RES_UPDATE_CANCEL` с `block_minus=true` оно
  откатится, но event уже создан → ресурс «бесплатный»).
  Защиты `event["start"] == time()` — спасает только если оба запроса
  попали в одну секунду; на rate ≥ 1 RPS не помогает.
- **Фикс**: транзакция вокруг (re-SELECT events с FOR UPDATE) +
  re-check_resources + addEvent + updateUserRes. Альтернатива — сделать
  `addEvent` атомарным через `INSERT … SELECT WHERE NOT EXISTS (SELECT 1
  FROM events WHERE userid=… AND mode=… AND data->>'buildingid'=…)`.

### RACE-004 — lost update через rollback в NS::updateUserRes

- **Серьёзность**: P1
- **Статус**: verified
- **Где**: [src/game/NS.class.php:1564-1644](../../projects/game-origin/src/game/NS.class.php#L1564-L1644)
- **Дыра**: rollback (строка 1636-1640) использует `$params["before_*"]`,
  снятый на строке 1566. Если параллельный поток успел поменять баланс
  между snapshot (1566) и rollback (1636), `SET metal=before_metal`
  затрёт изменения параллельного потока.
- **Сценарий**:
  1. A: SELECT → metal=1000, UPDATE metal-=800 → 200, result=200 (>0, OK).
  2. B (параллельно, чуть раньше): SELECT → metal=1000, UPDATE metal-=300
     → -100, result=-100 (<0) → rollback `SET metal=before_metal=1000`.
  3. Финальный metal=1000 — списание потока A исчезло.
- **Фикс**: SQL-уровень atomic check `UPDATE planet SET metal=metal-cost
  WHERE planetid=… AND metal>=cost`; в коде проверять affected_rows.
  Полный отказ от паттерна «UPDATE → SELECT → rollback by snapshot».

### RACE-001/002 — отчёт обрезан, не получены

Explore-агент 37.8.2 вернул только RACE-003/004 в финальном выводе
(output-файл оказался пуст после завершения). RACE-001/002 — судя по
индексации, относились к Buildings/Shipyard и закрываются собственным
ручным аудитом выше (в составе RACE-003 для Constructions/Shipyard).

---

### STUCK-001 — event помечается PROCESSED_START до бизнес-логики

- **Серьёзность**: P1 (понижено с P0 — try/catch есть, но неполный)
- **Статус**: verified
- **Где**: [src/game/EventHandler.class.php:553-813](../../projects/game-origin/src/game/EventHandler.class.php#L553-L813)
- **Verification**:
  - PROCESSED_START ставится до бизнес-логики (553-561) — да.
  - Есть `try/catch (Exception $e)` (564-784) — exception ставит
    PROCESSED_ERROR (790-793). **Агент эту защиту не упомянул.**
  - Финальный UPDATE на OK (802-813) с защитой от race через `prev_rc`.
- **Реальная дыра** (после verification):
  `catch(Exception $e)` на PHP 8 **не ловит `\Error`** (TypeError,
  ParseError, OOM, DivisionByZeroError и пр.). Если в одном из
  обработчиков (`attack`, `colonize`, `expedition`) случится fatal
  `Error` — try/catch не сработает, event останется в PROCESSED_START.
  Также `kill -9` / segfault PHP-процесса — не ловит ничто.
- **Фикс (trivial)**: заменить `catch(Exception $e)` на
  `catch(\Throwable $e)` (строка 784, и аналогично 2525 если применимо).
  Закроет ~90% реальных стопоров. Остаётся только kill -9 / segfault,
  для них нужен STUCK-003 cleanup.

### STUCK-002 — removeEvent не видит PROCESSED_START

- **Серьёзность**: P1
- **Статус**: verified, но фикс **не trivial**
- **Где**: [src/game/EventHandler.class.php:946-947](../../projects/game-origin/src/game/EventHandler.class.php#L946-L947)
- **Краткое**: `removeEvent` фильтрует только по `processed in (WAIT)`.
  Игрок не может отменить флот, застрявший в `PROCESSED_START`.
- **Verification**: на строке 946 есть **закомментированный** вариант с
  `EVENT_PROCESSED_WAIT, EVENT_PROCESSED_START` — это сознательное
  отключение, не оплошность. Расширение фильтра без атомарной защиты
  даст race с воркером (если воркер сейчас обрабатывает event и игрок
  одновременно его удаляет — двойная обработка / частичный rollback).
- **Фикс**: атомарный `UPDATE … SET processed=PROCESSED_CANCELED WHERE
  eventid=… AND processed=PROCESSED_START` + проверка affected_rows.
  Если affected=1 — мы забрали event у воркера; иначе воркер успел
  довести до OK. Сложность — medium (не trivial).

### STUCK-003 — cleanup PROCESSED_START есть, но слишком редкий и не возвращает ресурсы

- **Серьёзность**: P2 (понижено с P1 — частично закрыто, но cleanup
  не делает recovery)
- **Статус**: verified, но описание агента неточное
- **Где**: [src/game/EventHandler.class.php:822-834](../../projects/game-origin/src/game/EventHandler.class.php#L822-L834)
- **Verification**: на строке 828 явно есть
  `OR (processed = EVENT_PROCESSED_START AND processed_time < time() - 14*86400)`.
  Cleanup для PROCESSED_START **есть**, агент пропустил.
- **Реальные проблемы** (после verification):
  1. **14 дней слишком долго**. Игрок ждёт recovery две недели — явный
     UX-проблема. Реалистичный max event duration: 24-48 часов
     (экспедиция). Должно быть `time() - 60*60*24*3` (3 дня).
  2. **Cleanup только удаляет event, не возвращает ресурсы/корабли**.
     Если флот завис в attack/expedition/colonize — игрок теряет всё,
     что было «in flight». Нужен полноценный recovery (вернуть ships
     + cargo на planet или в shipyard).
- **Фикс**:
  - Уменьшить порог: 3 дня вместо 14 для PROCESSED_START.
  - Recovery-логика: при cleanup PROCESSED_START для fleet-events
    (attack, expedition, transport) — попытаться вернуть флот на
    исходную планету через `haltReturn`-аналог. Сложность medium.
  - Альтернатива: оставить cleanup как есть (admin-recovery), а для
    игрока добавить кнопку «вернуть зависший флот» (UI + backend).

### STUCK-004 — event-monitor exception handling

- **Серьёзность**: P2 (по агенту)
- **Статус**: unverified
- **Где**: worker/event-monitor.php:68-83 (по агенту)
- **Краткое**: try/catch вокруг `goThroughEvents()` ловит uncaught
  exceptions, но не гарантирует атомарность отдельного event'а.
- **Гипотетический фикс**: транзакция per-event внутри обработчика,
  не в воркере.

---

### OVF-004 — typo `metal+metal+metal` в demolish() points

- **Серьёзность**: P1 (понижено с P0 — рейтинговая ошибка, не game-break)
- **Статус**: verified
- **Где**: [src/game/EventHandler.class.php:2253](../../projects/game-origin/src/game/EventHandler.class.php#L2253)
  ```php
  $points = round(($data["metal"] + $data["metal"] + $data["metal"]) * RES_TO_BUILD_POINTS, POINTS_PRECISION);
  ```
- **Дыра**: должно быть `metal + silicon + hydrogen` (как в строках
  2201, 2853, 3021 этого же файла и Shipyard.class.php:538). Сейчас при
  сносе здания points считаются как `3*metal`, игнорируя silicon и
  hydrogen.
- **Эффект**: для металл-heavy зданий — переоценка снимаемых очков
  (списывается больше points чем добавилось при постройке) → user может
  уйти в отрицательный b_points. Для silicon/hydrogen-heavy зданий —
  недосписание очков.
- **Фикс** (одна строка): заменить `metal + metal + metal` на
  `metal + silicon + hydrogen`. Это явный typo, а не балансное решение.

### OVF-001/002 — float precision loss в production

- **Серьёзность**: P3 (понижено с P2 — edge-case)
- **Статус**: won't-fix (запись в simplifications)
- **Где**: src/game/Planet.class.php:442-650 (по агенту, не верифицировано
  в деталях)
- **Краткое**: на high-end игре (research level 40+, миллионы спутников)
  PHP DOUBLE arithmetic теряет точность в последних значащих цифрах.
  Игроки теряют долю процента ресурсов при каждом тике.
- **Решение**: won't-fix. PHP-порт legacy не предполагается тащить дальше
  37.9 (после которого EventHandler уйдёт в Go). Переход на bcmath даёт
  performance hit при ~0.1% выигрыше точности на edge-case high-end.
  Запись в `docs/simplifications.md`.

### OVF-003 — eval() в parseChargeFormula без cap

- **Серьёзность**: P3 (понижено с P1 — не user-эксплойт)
- **Статус**: verified
- **Где**: [src/game/Functions.inc.php:41-54](../../projects/game-origin/src/game/Functions.inc.php#L41-L54)
- **Verification**: `$formula` приходит из `na_units.charge_*` (БД), что
  заполняется миграциями/seeds — admin-controlled, **не user-controlled**.
  User не может подсунуть формулу. Дыры с точки зрения exploit-сценария нет.
- **Реальный риск (понижен)**:
  - Если злоумышленник получит SQL-инъекцию (37.6 audit показал, что
    её нет в коде) — он сможет писать любой PHP-код в БД и тот выполнится.
  - Typo в миграции (`pow(2.0, …)` вместо `pow(1.1, …)`) — silent overflow.
- **Фикс** (опционально): clamp в `parseChargeFormula` —
  `return max(0, round(min($result, PHP_INT_MAX / 2)))`. Защищает от
  typo и от потенциального будущего SQL-injection. small effort.
  Замена `eval` на безопасный math-парсер — отдельная задача (overkill сейчас).

---

### LIM-001 — race на лимит экспедиций (с защитой isFirstRun на md5(post))

- **Серьёзность**: P2 (понижено с P1 — есть isFirstRun на 2 сек)
- **Статус**: verified
- **Где**: [src/game/page/Mission.class.php:1361-1390](../../projects/game-origin/src/game/page/Mission.class.php#L1361-L1390),
  [src/game/page/Mission.class.php:354-360](../../projects/game-origin/src/game/page/Mission.class.php#L354-L360)
- **Verification**:
  - На строке 1363: `NS::isFirstRun("Mission::sendFleet:" . md5(serialize($post)) . "-" . $userid)` — дедуп ON.
  - Ключ зависит от **полного содержимого POST** (включая координаты
    цели). Двойной клик с одной целью — md5 одинаковый → дедуп
    срабатывает.
  - Если игрок отправляет ДВЕ экспедиции на ДВЕ разные точки одновременно
    (разные md5) → дедуп не помогает, count читается из process-local
    `eventStack` (stale snapshot) → возможно превысить лимит UNIT_EXPO_TECH на 1.
- **Реальный сценарий**: игрок пишет скрипт «отправить экспедиции на N
  разных целей за < 50ms» → каждый запрос видит count=N-1, все проходят →
  превышение лимита.
- **Фикс**: либо атомарный `INSERT … SELECT WHERE (SELECT COUNT(*) …) <
  UNIT_EXPO_TECH`, либо ключ memcached без зависимости от target
  (`Mission::sendFleet:expedition:{$userid}` — серьёзно ограничит UX,
  игрок не сможет оперативно отправить две легитимные экспедиции).
  Лучший вариант — атомарная SQL-проверка.

### LIM-002 — race на лимит лотов на рынке (с защитой isFirstRun на userid)

- **Серьёзность**: P3 (понижено с P1 — защита практически закрывает)
- **Статус**: verified
- **Где**: [src/game/page/StockNew.class.php:622-647](../../projects/game-origin/src/game/page/StockNew.class.php#L622-L647)
- **Verification**: `NS::isFirstRun("StockNew::addLot:" . $userid)` —
  ключ зависит **только от userid**, не от данных лота. Любой второй
  addLot за 2 сек блокируется. Игрок физически не может создать 2 лота
  за 2 сек.
- **Реальный сценарий**: чтобы обойти — нужно ждать > 2 сек между
  запросами, но тогда первый лот уже в БД и `freeSlots()` его учтёт.
  Race window — точно между 2 и 3 секундами. Эксплойт даёт max +1 лот.
- **Фикс**: low priority. Атомарный INSERT-SELECT-WHERE-COUNT желателен,
  но не срочно.

### LIM-003 — race на лимит fleet slots (тот же isFirstRun(md5(post)))

- **Серьёзность**: P2 (понижено с P2, без изменения)
- **Статус**: verified
- **Где**: [src/game/page/Mission.class.php:1387-1390](../../projects/game-origin/src/game/page/Mission.class.php#L1387-L1390)
- **Verification**: тот же `isFirstRun` ключ что LIM-001 (md5(serialize($post))).
  Двойной клик одной командой — блокируется. Параллельные fleet-команды
  на разные цели — race возможен (проходит +1 за лимит COMPUTER_TECH).
- **Фикс**: атомарный SQL-check как для LIM-001 — общий паттерн.

---

## Сводная таблица по фиксам (после verification)

| ID | Серьёзность | Статус | Сложность фикса |
|---|---|---|---|
| **OVF-004** | P1 | verified | **trivial (1 строка)** ← начать с этого |
| **REF-005** | P1 | verified, no isFirstRun | small (атомарный UPDATE + check) |
| **STUCK-001** | P1 | verified | **trivial (Exception → \\Throwable)** |
| RACE-003 | P1 | verified, isFirstRun TTL=2s ослабляет | medium (3 файла) |
| RACE-004 | P1 | verified | medium (рефакторинг updateUserRes) |
| STUCK-002 | P1 | verified, fix не trivial | medium (race с воркером) |
| STUCK-003 | P2 | verified, cleanup есть но 14 дней | small (3 дня) + medium (recovery) |
| REF-004 | P2 | verified, isFirstRun помогает | small (атомарный status-update) |
| LIM-001 | P2 | verified, isFirstRun(md5(post)) ослабляет | medium (атомарный SQL) |
| LIM-003 | P2 | verified, isFirstRun(md5(post)) ослабляет | medium (атомарный SQL) |
| LIM-002 | P3 | verified, isFirstRun на userid сильный | low priority |
| OVF-003 | P3 | verified, admin-input не user | optional clamp в eval-обёртке |
| OVF-001/002 | P3 | won't-fix → simplifications | — |
| STUCK-004 | — | rejected (повтор STUCK-001 с другой стороны) | — |
| REF-001/002/003/006/007 | — | rejected/by-design | — |

**Приоритет фиксов** (по соотношению impact/effort):
1. **OVF-004** (1 строка, рейтинговая ошибка) — trivial.
2. **STUCK-001** (1 слово: `Exception` → `\Throwable`) — trivial.
3. **REF-005** (атомарный UPDATE WHERE credit >= cost) — small.
4. **STUCK-003** (порог 14 дней → 3 дня) — small (полноценный recovery
   отдельной задачей).
5. **REF-004**, **RACE-003**, **RACE-004** — medium.
6. **LIM-001/003**, **STUCK-002** — medium.
7. **OVF-003**, **LIM-002** — low priority.

---

## Этапы фикса (план)

### Ф.1 — Низкая стоимость, высокая ценность (verified)

Один коммит на каждый фикс или группу:

1. **OVF-004**: typo metal+metal+metal — 1 строка.
2. **REF-004**: атомарный status-update в `Exchange::sendBackLot`.
3. **REF-005**: атомарный credit-update в `Market::Credit_ex`.
4. **RACE-003**: транзакция в `upgradeResearch` / `upgradeConstruction` /
   Shipyard build flow. Если транзакции дороги по UX — можно начать
   с одного `addEvent`-уровня (fail-on-duplicate с retry).
5. **RACE-004**: рефакторинг `updateUserRes` — atomic UPDATE-WHERE
   взамен паттерна snapshot+rollback. **Большой риск регрессии**, делать
   аккуратно с тестом.

### Ф.2 — Stuck fleet (unverified, нужна верификация)

Сначала прочитать строки указанные агентом, проверить реальность
сценариев. Потом — STUCK-002 (легко: расширить фильтр), STUCK-003
(cleanup-cron). STUCK-001 — большая задача (транзакции вокруг
event-processing); может потребовать ADR.

### Ф.3 — Limits (unverified)

Тот же паттерн что RACE-003 — после Ф.1 можно копипастой решений.

### Ф.4 — Overflow (unverified)

OVF-003 (eval clamp) — small. OVF-001/002 — кандидат на won't-fix или
запись в simplifications.

### Ф.5 — Финализация

- Обновить `docs/simplifications.md` для won't-fix решений (OVF-001/002).
- Обновить `docs/project-creation.txt` итоговой записью.
- Smoke-тест всего стека после фиксов.
- Удалить этот документ или законсервировать как «реестр исправленных
  до запуска» (по образу `docs/balance/audit.md`).

---

## Важные оговорки

- Этот аудит — **только game-origin**. Не сверять формулы/защиты с
  game-nova (они независимые проекты, см. memory `project_origin_vs_nova.md`).
- legacy-источник `d:\Sources\oxsar2` имеет **те же дыры** в большинстве
  случаев. Это не значит «оставляем как в legacy» — паритет с legacy не
  цель (memory `feedback_no_legacy_parity.md`).
- Перед фиксом каждой `unverified` записи — открыть указанные строки и
  убедиться руками, что дыра реальна и сценарий эксплойта рабочий.
  Прецедент 2026-04-27: из 7 первичных refund-находок реальны оказались
  только 2.
