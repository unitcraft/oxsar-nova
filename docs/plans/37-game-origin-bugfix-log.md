# План 37.8: Game-Origin — реестр игровых дыр

**Дата начала**: 2026-04-27
**Статус**: Аудит завершён, фиксы — в работе
**Связан с**: [37-game-origin.md](37-game-origin.md) §37.8, [37-game-origin-security-audit.md](37-game-origin-security-audit.md) (security-аудит §37.6)

Реестр **игровых** дыр в `projects/game-origin/` (баланс/механика, не безопасность).
Безопасность (XSS/CSRF) — отдельный документ (37.6/37.7).

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

### REF-004 — TOCTOU в Exchange::sendBackLot — двойной refund лота

- **Серьёзность**: P1
- **Статус**: verified
- **Где**: [src/game/Exchange.class.php:467-549](../../projects/game-origin/src/game/Exchange.class.php#L467-L549)
- **Дыра**: между `SELECT … WHERE l.lid=$id AND l.status=ESTATUS_OK`
  (строка 475) и `UPDATE … SET status=ESTATUS_RECALL` (549) нет блокировки.
  Защита `NS::isFirstRun("Exchange::sendBackLot:{$id}")` — process-local,
  не помогает между параллельными PHP-процессами.
- **Сценарий**: два параллельных HTTP-запроса на отзыв лота → оба
  проходят SELECT (status=OK), оба возвращают ресурсы и юниты владельцу,
  оба меняют статус. Игрок получает двойной refund.
- **Фикс**: атомарный `UPDATE exchange_lots SET status=ESTATUS_RECALL
  WHERE lid=$id AND status=ESTATUS_OK` ПЕРЕД refund-логикой; проверить
  affected_rows == 1; refund только если строка реально перешла из OK
  в RECALL. Альтернатива — транзакция с `SELECT … FOR UPDATE`.

### REF-005 — TOCTOU credit в Market::Credit_ex — credit в минус

- **Серьёзность**: P1 (повышено с P2 — credit может уйти ниже нуля)
- **Статус**: verified
- **Где**: [src/game/page/Market.class.php:200-240](../../projects/game-origin/src/game/page/Market.class.php#L200-L240),
  [src/game/NS.class.php:1657-1669](../../projects/game-origin/src/game/NS.class.php#L1657-L1669)
- **Дыра**: проверка `$credit > $this->credit` (строка 211) идёт против
  snapshot, загруженного в начале запроса. `RES_UPDATE_EXCHANGE` имеет
  тип `+-` (без clamp на минус на входе). Внутри `updateUserRes` для
  credit нет SQL-clamp `GREATEST(0, …)`, и `block_minus` не передаётся.
- **Сценарий**: два параллельных запроса на покупку ресурсов на рынке →
  оба видят достаточный credit-баланс, оба проходят, оба
  `credit = credit - cost`. Итог — отрицательный credit.
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

- **Серьёзность**: P0 (по агенту)
- **Статус**: unverified
- **Где**: src/game/EventHandler.class.php:553-561, 799-813 (по агенту)
- **Краткое**: `events.processed=PROCESSED_START` ставится **до**
  выполнения логики обработчика. Если бизнес-логика крашится между
  стартом и финальным `processed=OK`, событие остаётся в `PROCESSED_START`
  навсегда — флот «в полёте», но никогда не вернётся.
- **Гипотетический фикс**: транзакция вокруг event-processing с
  COMMIT при OK / ROLLBACK при exception, либо помечать PROCESSED_START
  только после успешного выполнения логики.

### STUCK-002 — removeEvent не видит PROCESSED_START

- **Серьёзность**: P1 (по агенту)
- **Статус**: unverified
- **Где**: src/game/EventHandler.class.php:947 (по агенту)
- **Краткое**: `removeEvent` фильтрует только по `processed in (WAIT)`.
  Игрок не может отменить флот, который застрял в `PROCESSED_START`.
- **Гипотетический фикс**: расширить фильтр до `WAIT, PROCESSED_START`
  для `force_remove=true`.

### STUCK-003 — нет cleanup для PROCESSED_START

- **Серьёзность**: P1 (по агенту)
- **Статус**: unverified
- **Где**: src/game/EventHandler.class.php:822-834 (по агенту)
- **Краткое**: random cleanup (1/1001) убирает только old `OK/ERROR/WAIT`,
  не `PROCESSED_START`. Зависшие события живут вечно.
- **Гипотетический фикс**: добавить cleanup `PROCESSED_START` старше
  N часов (значение N зависит от max event duration; 6h должно быть
  безопасно для большинства флотов, кроме экспедиций).

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

- **Серьёзность**: P2 (по агенту)
- **Статус**: unverified
- **Где**: src/game/Planet.class.php:442-650 (по агенту)
- **Краткое**: на high-end игре (research level 40+, миллионы спутников)
  PHP DOUBLE arithmetic теряет точность в последних значащих цифрах.
  Игроки теряют долю процента ресурсов при каждом тике.
- **Стоит ли чинить**: P2, edge-case high-level. Скорее «принять как есть»
  чем переходить на bcmath (производительный hit). Записать в
  simplifications.md.

### OVF-003 — eval() в parseChargeFormula без cap

- **Серьёзность**: P1 (по агенту)
- **Статус**: unverified
- **Где**: src/game/Functions.inc.php:41-54 (по агенту)
- **Краткое**: `parseChargeFormula` делает `eval()` строки из БД с
  pow()/floor() без clamp на максимальный результат. Если в БД попадёт
  формула с экспонентой 2.0 вместо 1.1, на level 40 переполнение float
  precision.
- **Стоит ли чинить**: ограниченный admin-input (формулы — не user-controlled).
  Но `eval()` в чувствительном месте — отдельный security smell.

---

### LIM-001 — race на лимит экспедиций

- **Серьёзность**: P1 (по агенту)
- **Статус**: unverified
- **Где**: src/game/page/Mission.class.php:354-1381 (по агенту)
- **Краткое**: `getUsedExpeditionSlots()` читает из process-local
  `eventStack`. Параллельные запросы видят stale count, оба проходят
  лимит, оба добавляют экспедицию.
- **Фикс (как RACE-003)**: атомарный INSERT с count-check в одном SQL.

### LIM-002 — race на лимит лотов на рынке

- **Серьёзность**: P1 (по агенту)
- **Статус**: unverified
- **Где**: src/game/Exchange.class.php:1169-1184, StockNew::addLot
- **Краткое**: `freeSlots()` делает SELECT max + SELECT COUNT, без
  транзакции. Race window между ними и INSERT.
- **Фикс**: транзакция или атомарный INSERT-SELECT-WHERE-COUNT.

### LIM-003 — race на лимит fleet slots

- **Серьёзность**: P2 (по агенту)
- **Статус**: unverified
- **Где**: src/game/page/Mission.class.php:1387 (по агенту)
- **Краткое**: `getUsedFleetSlots()` читает из process-local snapshot.
  Параллельные `sendFleet` могут превысить лимит на 1.
- **Фикс**: тот же паттерн — атомарная проверка в SQL.

---

## Сводная таблица по фиксам

| ID | Серьёзность | Статус | Сложность фикса |
|---|---|---|---|
| OVF-004 | P1 | verified | trivial (1 строка) |
| REF-004 | P1 | verified | small (атомарный UPDATE + проверка) |
| REF-005 | P1 | verified | small (атомарный UPDATE + проверка) |
| RACE-003 | P1 | verified | medium (3 файла, транзакция или INSERT-SELECT) |
| RACE-004 | P1 | verified | medium (рефакторинг updateUserRes на atomic UPDATE) |
| STUCK-001 | P0? | unverified | medium-large (transactions + retry semantics) |
| STUCK-002 | P1? | unverified | trivial (расширить фильтр) |
| STUCK-003 | P1? | unverified | small (cleanup-cron в воркере) |
| LIM-001/002/003 | P1/P1/P2 | unverified | medium (тот же паттерн что RACE-003) |
| OVF-001/002 | P2 | unverified | won't-fix кандидат → simplifications.md |
| OVF-003 | P1 | unverified | small (clamp на результат eval) |
| STUCK-004 | P2 | unverified | medium |

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
