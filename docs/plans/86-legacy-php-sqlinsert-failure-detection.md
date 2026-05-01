# План 86: legacy-PHP — `sqlInsert()` тихо игнорирует ошибки INSERT

**Дата**: 2026-05-01
**Статус**: 🔵 Открыт
**Зависимости**: нет (точечная задача внутри `projects/game-legacy-php/`).

**Связанные документы**:
- [docs/project-creation.txt](../project-creation.txt) — итерация 2026-05-01,
  где обнаружен баг (последствия для симулятора боя).
- [docs/plans/75-rename-game-origin-to-php.md](75-rename-game-origin-to-php.md)
  — папка `projects/game-legacy-php/`.

---

## Контекст

В `projects/game-legacy-php/src/core/Functions.php:215-223` глобальный
helper `sqlInsert($table, $values)`:

```php
function sqlInsert($table_name, $values)
{
    Core::getQuery()->insert(  // ← возврат игнорируется
        $table_name,
        array_keys($values),
        array_values($values)
    );
    return Core::getDB()->insert_id();
}
```

`Core::getQuery()->insert()` (`QueryParser.class.php:26-41`) возвращает
`PDOStatement` при успехе и `false` при ошибке (см. `DB_MYSQL_PDO::query`
`projects/game-legacy-php/src/core/database/DB_MYSQL_PDO.class.php:46-63`
— PDOException пишется в error_log и возвращается `false`).

`sqlInsert()` **никогда не смотрит** на этот возврат. Если INSERT упал
(FK, UNIQUE, NOT NULL, etc.), `sqlInsert()` спокойно идёт дальше и
возвращает результат `Core::getDB()->insert_id()`. PDO `lastInsertId()`
**возвращает PK от ПОСЛЕДНЕГО успешного INSERT в этой connection**, а
не от только что упавшего. Вызывающий код получает ID **другой**
строки и использует его дальше — обычно как FK на следующий INSERT,
который тоже падает (или, что хуже, проходит и пишет данные с
неправильным reference).

## Где это уже выстрелило

В сессии 2026-05-01 баг привёл к **пустым отчётам в симуляторе боя**
(см. дневник за эту дату):

1. `Simulator.class.php:711` вызывает `addParticipant()` →
   `Core::getQuery()->insert("sim_assaultparticipant", ...)`.
2. INSERT падает с FK violation (`planetid=1` отсутствовал в
   `na_sim_planet`, исправлено отдельно миграцией
   [006_seed_sim_planet.sql](../../projects/game-legacy-php/migrations/006_seed_sim_planet.sql)).
3. `Core::getDB()->insert_id()` возвращает `6` — PK от **последнего
   успешного** INSERT в этой connection (это был `sim_assault.assaultid=6`
   несколькими строками выше).
4. `participantid=6` записывается в `sim_fleet2assault.participantid`,
   FK ibfk_2 (на `sim_assaultparticipant.participantid`) тоже падает.
5. Java стартует, не видит участников боя, пишет пустой отчёт
   («атакующий выиграл, 0 потерь, 0 раундов»).
6. PHP молча проглатывает (`try/catch` без логирования вокруг exec),
   пользователь видит пустой бой.

Это конкретное проявление. Аналогичные сценарии возможны во **всех**
58 вызовах `sqlInsert()` в кодовой базе. Также паттерн повторяется
в 9 прямых вызовах `Core::getDB()->insert_id()` (без обёртки
`sqlInsert`):

- `src/core/JwtAuth.php:250`
- `src/game/page/ArtefactMarketOld.class.php:138, :185`
- `src/game/page/MSG.class.php:179`
- `src/game/page/Alliance.class.php:208`
- `src/game/page/Simulator.class.php:716`
- `src/game/PlanetCreator.class.php:438, :459`

## Цель

Сделать так, чтобы провалившийся INSERT гарантированно **не давал
вызывающему коду «фейковый» insert_id** от другой строки. Вместо
этого:
- залогировать факт ошибки (с SQL-контекстом, чтобы было видно в
  `docker logs`);
- вернуть значение, которое вызывающий код может явно проверить и не
  сможет случайно использовать как валидный FK (`false` / `null` / `0`).

## Не цель

- **Не** исправлять каждый из 58 + 9 call sites индивидуально (внутри
  Simulator, Alliance, MSG и др.). Защита делается **в одном месте** —
  в `sqlInsert()` (центральная обёртка) и в `QueryParser::insert()`
  (для прямых пользователей `Core::getQuery()->insert(...)`).
- **Не** менять семантику successful-вызовов — они продолжают
  возвращать тот же `lastInsertId()`. Меняется только behaviour при
  ошибке.
- **Не** переделывать `DB_MYSQL_PDO::query()` — он уже корректно
  возвращает `false` при PDOException и пишет error_log.
- **Не** трогать симулятор-баг с `na_sim_planet` (закрыт отдельной
  миграцией 006). Этот план — про корневую защиту, чтобы аналогичные
  тихие провалы не повторились в других местах.

## Что делаем

### 1. Поправить `sqlInsert()` — основная точка фикса

Файл: `projects/game-legacy-php/src/core/Functions.php:215-223`.

Было:
```php
function sqlInsert($table_name, $values)
{
    Core::getQuery()->insert(
        $table_name,
        array_keys($values),
        array_values($values)
    );
    return Core::getDB()->insert_id();
}
```

Стало:
```php
function sqlInsert($table_name, $values)
{
    $rc = Core::getQuery()->insert(
        $table_name,
        array_keys($values),
        array_values($values)
    );
    // QueryParser::insert() возвращает false при PDOException
    // (DB_MYSQL_PDO::query пишет error_log и проглатывает throw).
    // Без этой проверки sqlInsert() возвращал бы lastInsertId() от
    // ПРЕДЫДУЩЕГО успешного INSERT — вызывающий код получил бы
    // фейковый PK другой строки и каскадно ломал FK на следующих
    // запросах. Подробности — план 86.
    if ($rc === false) {
        return false;
    }
    return Core::getDB()->insert_id();
}
```

Семантика для вызывающего кода:
- На успехе — как раньше: PK только что вставленной строки (string,
  PDO так возвращает).
- На ошибке — `false` вместо случайного `int|string`.

Большинство существующих 58 call sites сравнивают результат
`sqlInsert()` через truthy check (`if ($id)` / `if ($x = sqlInsert(...))`)
— `false` корректно отвалится. Несколько мест делают `(int)$x` /
прямой return — там `false` приведётся к `0` (в worst-case следующий
INSERT упадёт на FK с `0`, что **гораздо лучше** чем фейковый ID и
повредившиеся данные).

### 2. Поправить `QueryParser::insert()` — для прямых пользователей

Файл: `projects/game-legacy-php/src/core/QueryParser.class.php:26-41`.

Возврат уже корректный (`false` при ошибке, `PDOStatement` при
успехе) — изменения не нужны. Но **9 call sites** обращаются к нему
напрямую (`Core::getQuery()->insert(...)` + `Core::getDB()->insert_id()`),
не через обёртку `sqlInsert`. Для них фикс через `sqlInsert` не
действует.

Стратегия: оставить `QueryParser::insert()` как есть, починить эти
9 точек точечно (см. п.3) — они в специфических местах, у каждого
своя логика обработки ошибки.

### 3. Защита 9 прямых call sites `insert() + insert_id()`

Все 9 мест приведены в шапке плана. Паттерн правки везде один:

```php
// Было:
Core::getQuery()->insert("table", [...], [...]);
$pk = Core::getDB()->insert_id();
// ... использовать $pk

// Стало:
$rc = Core::getQuery()->insert("table", [...], [...]);
if ($rc === false) {
    // конкретное поведение зависит от call site, см. ниже
    return; // или Logger::dieMessage('DB_ERROR'); или throw new Exception(...);
}
$pk = Core::getDB()->insert_id();
// ... использовать $pk
```

Конкретные стратегии по call site:

| Файл:Строка | Контекст | Стратегия при провале INSERT |
|---|---|---|
| `Simulator.class.php:716` | `addParticipant` в симуляторе | early-return из `addParticipant`; вызывающий цикл `simulate()` пропустит `for($i=0;...)` потому что `$participantid===false` → пустые данные не пишем |
| `Alliance.class.php:208` | создание альянса | `Logger::dieMessage('DB_ERROR_CREATE_ALLIANCE')` (форма не должна продолжаться без валидного ID) |
| `MSG.class.php:179` | создание события `eventid` для сообщения | early-return; сообщение всё равно отправлено через предыдущий INSERT |
| `ArtefactMarketOld.class.php:138, :185` | артефакт с истечением | `Logger::dieMessage('DB_ERROR_ARTEFACT_DISAPPEAR')` (без ID событие удаления не повесить, артефакт повиснет вечно) |
| `JwtAuth.php:250` | создание JWT-сессии | `throw new Exception('Failed to create user record')` — вызывающий код в auth-layer обработает |
| `PlanetCreator.class.php:438, :459` | создание планеты при регистрации | `throw new Exception('Failed to create planet')` (PlanetCreator уже в try/catch выше по стеку) |

### 4. Тесты / проверка

Нет существующих unit-тестов на `sqlInsert()` (legacy-PHP вообще без
PHPUnit-инфры). Стратегия проверки — manual smoke:

1. Применить фикс п.1 (`sqlInsert`).
2. Откатить миграцию `006_seed_sim_planet.sql` (`DELETE FROM na_sim_planet
   WHERE planetid=1`) — это вернёт состояние с FK-violation в симуляторе.
3. Зайти в `/game.php?go=Simulator`, заполнить флоты, нажать
   «Симулировать».
4. Проверить `docker logs legacy-php-php-1`:
   - **до фикса** (для сравнения): пользователь видит пустой отчёт,
     в логе цепочка `Cannot add or update a child row` × N штук без
     явной точки остановки.
   - **после фикса**: первый FK-fail на `sim_assaultparticipant`
     останавливает всю цепочку. В логе только одна ошибка
     `Cannot add or update a child row` с понятным контекстом, потом
     бой не запускается (или запускается с пустыми данными — но не
     создаёт каскадный мусор в `sim_fleet2assault`).
5. Восстановить `006_seed_sim_planet.sql` (`mysql < 006_*.sql`).
6. Повторить симуляцию — должен работать как и до плана 86 (smoke
   проходит, отчёт не пустой). Это подтверждает, что фикс не сломал
   успешный путь.

После п.5 — точечные правки 9 call sites (п.3) проверять каждый
руками невозможно (некоторые крайне сложно триггерить). Достаточно
синтаксической проверки `php -l` каждого файла + `grep` подтверждения,
что ни одного `Core::getDB()->insert_id()` не осталось без guard'а.

## Файлы

**Правим**:
- `projects/game-legacy-php/src/core/Functions.php` — `sqlInsert()` +5 строк.
- `projects/game-legacy-php/src/game/page/Simulator.class.php` — `addParticipant()` +3 строки.
- `projects/game-legacy-php/src/game/page/Alliance.class.php` — около строки 208.
- `projects/game-legacy-php/src/game/page/MSG.class.php` — около строки 179.
- `projects/game-legacy-php/src/game/page/ArtefactMarketOld.class.php` — строки 138, 185.
- `projects/game-legacy-php/src/core/JwtAuth.php` — около строки 250.
- `projects/game-legacy-php/src/game/PlanetCreator.class.php` — строки 438, 459.

**Не трогаем**:
- `projects/game-legacy-php/src/core/QueryParser.class.php` (возврат уже корректный).
- `projects/game-legacy-php/src/core/database/DB_MYSQL_PDO.class.php` (error_log + return false уже на месте).
- `projects/game-legacy-php/src/core/Functions.php` функции кроме `sqlInsert` (`sqlUpdate`, `sqlSelect*` — у них другая семантика возврата, отдельная задача если потребуется).

## Smoke

См. п.4 «Тесты / проверка».

## Финализация

1. Обновить шапку плана 86 — статус «✅ Закрыт <дата>».
2. Запись в `docs/project-creation.txt` — итерация 86 (хронология).
3. Коммит: `fix(legacy-php): sqlInsert() и 9 прямых insert_id-вызовов проверяют rc (план 86)`.
4. Опционально — заметка в `docs/simplifications.md` если в каком-то
   call site пришлось вместо graceful handling сделать
   `Logger::dieMessage` (по сути hard-fail) — это не упрощение,
   но trade-off «надёжность vs UX», стоит зафиксировать.

## Риски и рассуждения

**Риск**: возврат `false` из `sqlInsert()` при ошибке может быть
некорректно интерпретирован существующим кодом, который ожидал
числовой ID. Например, `(int) sqlInsert(...)` даст `0`, а если этот
ноль не проверяется — следующий INSERT с `parent_id=0` возможно
создаст «висячую» запись (если FK на `parent_id` не настроен или
nullable).

**Митigation**: даже в худшем сценарии «висячая запись с parent_id=0»
лучше чем «фейковый parent_id, указывающий на стороннюю запись».
И первое заметнее в данных (сразу выявляется при выборках), второе
— тихая порча. Принимаем risk-trade.

**Альтернатива**: бросать `Exception` из `sqlInsert()` при ошибке.
Чище, но требует добавления `try/catch` во всех 58 call sites —
несоразмерно вложениям в задачу. Откладываем; если в будущем
обнаружится конкретная регрессия от `false`-возврата — рассмотрим
повторно.

**Не делаем сейчас**:
- audit оставшихся 49 call sites `sqlInsert()` (помимо 9 прямых) —
  они защищены через п.1 автоматически. Если какие-то из них зависят
  от строгого числового возврата, это всплывёт в smoke на конкретной
  странице — фиксим по факту.
- transactional-обёртка вокруг multi-INSERT-блоков (`addParticipant +
  fleet2assault inserts` могли бы быть в одной транзакции) — это
  отдельный рефакторинг, не входит в этот план.
