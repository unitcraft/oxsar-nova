# План 43: game-origin — заменить Recipe-фреймворк на Composer-зависимости

**Дата**: 2026-04-26
**Статус**: ✅ Closed 2026-04-27 — 0 GPL-файлов в src/core/, см. project-creation.txt итерация 43.
**Зависимости**: ничего блокирующего; план 40 (аудит лицензий) лучше выполнять
**после** этого плана — в текущем состоянии game-origin содержит 54 GPL-файла,
которые CI-проверка пометит как несовместимые с PolyForm Noncommercial.
**Связанные документы**: [37-game-origin.md](37-game-origin.md) (исходный план
запуска game-origin), [40-license-audit.md](40-license-audit.md),
[41-origin-rights.md](41-origin-rights.md), [LICENSE](../../LICENSE).

---

## Цель

В `projects/game-origin-php/src/core/` сейчас 54 PHP-файла внешнего фреймворка
(шапка `@package Recipe 1.1`, `@author Sebastian Noll`, `@license GPL`).
GPL-лицензия фреймворка **несовместима** с PolyForm Noncommercial,
которой защищён весь oxsar-nova: пользователь, скачивая oxsar-nova,
получает GPL-код вместе с PolyForm-кодом, и юридически любой может
требовать, чтобы весь oxsar-nova распространялся под GPL.

Цель плана: убрать 54 файла Recipe из репозитория, заменив их
функциональность на установленные через Composer пакеты с лицензиями,
совместимыми с PolyForm (MIT, Apache-2.0, BSD-2/3-Clause, ISC, MPL-2.0).
Сама игровая логика (`src/game/`) и шаблоны (`src/templates/`) в этом
плане **не трогаются** — они остаются как есть. Их рефакторинг — отдельная
тема (план 37 / будущий rewrite).

---

## Что меняем

### 1. Подключить Composer

В `projects/game-origin-php/` создать `composer.json` с зависимостями:

| Recipe-класс | Замена | Лицензия |
|---|---|---|
| `Cache::class` | `symfony/cache` | MIT |
| `Database::*`, `DB_MySQL::*`, `DB_MySQLi::*`, `DB_MYSQL_PDO::*` | `doctrine/dbal` (или прямой PDO + тонкая обёртка) | MIT |
| `Template::*`, `TemplateCompiler::*` (Smarty-based) | **не заменяем в этом плане** — оставляем `Smarty` через Composer (`smarty/smarty`, BSD-3-Clause) для совместимости с существующими 125 `.tpl`. Smarty-3+ под BSD-3-Clause — совместима с PolyForm. | BSD-3-Clause |
| `Logger::class` | `monolog/monolog` | MIT |
| `Curl::class`, `Fsock::class`, `HTTP_Request::class`, `Request_Adapter::class` | `guzzlehttp/guzzle` | MIT |
| `Cron::class` | `dragonmantank/cron-expression` (парсер) или удалить, если cronjobs запускаются через системный cron | MIT |
| `XML::util`, `XMLObj::util` | стандартная `SimpleXML` PHP (встроена в core) | PHP License (совместима) |
| `Email::util` | `symfony/mailer` | MIT |
| `Image::util` | `intervention/image` (если активно используется) или GD/Imagick напрямую | MIT |
| `File::util`, `Upload::util` | `symfony/filesystem` + нативные функции | MIT |
| `util/Arr`, `util/Boolean`, `util/Integer`, `util/OxsarFloat`, `util/OxsarString`, `util/Str`, `util/Type`, `util/Date`, `util/Map`, `util/Link`, `util/Login` | стандартная PHP + точечные пакеты `symfony/string`, `nesbot/carbon` | MIT |
| `Hook::class`, `Plugin::abstract_class` | `symfony/event-dispatcher` или удалить, если не используется активно | MIT |
| `Language::class`, `LanguageCompiler::class`, `LanguageImporter`, `LanguageExporter`, `XML_LanguageImporter`, `SortPhrases` | `symfony/translation` (если останется i18n) или удалить, если переходим на одиночный язык | MIT |
| `QueryParser::class` | удалить, если не используется напрямую (часто внутренний хелпер) | — |
| `Timer::class` | удалить, заменить нативным `microtime()` | — |
| `Request::class` | `symfony/http-foundation` | MIT |
| `Options::class`, `Collection::class` | пересмотреть — часто это тонкие обёртки, можно заменить нативными типами | — |
| `AutoLoader::class` | **удаляется целиком** — Composer даёт PSR-4 autoload | — |
| `Core::class` | переписать как тонкий wire-up композитор Composer-сервисов | (новый код, наш copyright) |
| `User::class` | пересмотреть (часть бизнес-логики, не фреймворк) — оставить или переписать в `src/game/` | — |
| `AjaxRequestHelper::abstract_class` | переписать поверх `symfony/http-foundation` | — |
| `Exception::interface`, `GenericException::class`, `IssueException::class` | стандартные `\Exception` и кастомные классы (новый код) | — |
| `init.php` | переписать как Composer-bootstrap | — |
| `Functions.php` | разобрать функции по namespaces или Composer files autoload | — |
| `maintenance/TableCleaner`, `maintenance/banIPClasses`, `maintenance/ClearSessions` | переписать как Symfony Console-команды | — |

После замены — все 54 файла Recipe-фреймворка **удаляются** из репозитория.

### 2. Что **не** делаем в этом плане

- **Не трогаем `src/templates/*.tpl`** (125 файлов). Smarty остаётся как
  Composer-зависимость (`smarty/smarty`, BSD-3-Clause), синтаксис шаблонов
  не меняется. Это сильно сокращает scope: переписать сотни шаблонов под
  другой движок — отдельная большая работа.
- **Не трогаем `src/game/`** — игровая логика. Она использует Recipe-API
  (`Core::getUser()`, `Cache::get()`, и т.п.); точки вызова обновляются
  под новые API в шаге 3, но семантика логики не меняется.
- **Не делаем clean-room rewrite** игровой логики. Это вопрос отдельный
  (стратегия B3 плана 41); сейчас закрываем только GPL-конфликт фреймворка.
- **Не переходим на Go**. Game-origin остаётся на PHP — это сознательный
  выбор пользователя для сохранения функционала и UI.

### 3. Точки интеграции в игровой логике

После замены `Cache`, `DB`, `Logger`, `Request` и т.п. на Composer-пакеты —
обновить вызовы в `src/game/` и `src/core/legacy_payment/` (это уже свой
код, не Recipe).

Стратегия — **минимально инвазивные обёртки**: если в `Core.class.php`
сейчас `Core::getCache()` возвращает старый `Cache`-объект, в новой версии
возвращать `Symfony\Component\Cache\Adapter\AdapterInterface` с тем же
именем метода. Тогда массовых правок в игровой логике не требуется.

Если совместимость API сохранить нельзя — точечные правки call-сайтов.

### 4. .gitignore + Dockerfile

- В `.gitignore`: `projects/game-origin-php/vendor/`.
- В `projects/game-origin-php/docker/Dockerfile.php`: добавить шаг
  `composer install --no-dev --optimize-autoloader` при сборке образа.

### 5. CI

В `.github/workflows/ci.yml` (когда план 40 будет выполнен):
- `composer validate` для `game-origin`;
- `composer install` перед PHP-тестами;
- `license-checker` для composer-зависимостей (см. план 40).

### 6. Документация

В `docs/origin-rights.md` (создан в плане 41) — короткая ремарка в раздел
"Что не заимствуется": "Фреймворк Recipe (Sebastian Noll, GPL) исходно
присутствовал в legacy-коде game-origin; в плане 43 заменён на
Composer-пакеты под совместимыми с PolyForm лицензиями (MIT, BSD,
Apache-2.0); файлы Recipe удалены из репозитория."

---

## Этапы

### Ф.1. Бутстрап Composer

- Установить Composer в Docker-окружение game-origin (если ещё не).
- Создать `projects/game-origin-php/composer.json` с минимальным набором
  зависимостей (Symfony Cache, Doctrine DBAL, Smarty, Monolog, Guzzle).
- `composer install` — убедиться, что vendor/ создаётся, autoload работает.
- Добавить `vendor/` в `.gitignore` для проектной папки.

### Ф.2. Замена утилит (низкий риск)

Начать с самых простых классов, у которых мало callers:
- `util/Arr`, `util/Boolean`, `util/Integer`, `util/OxsarFloat`,
  `util/OxsarString`, `util/Str`, `util/Type`, `util/Date`, `util/Map` →
  заменить вызовы на нативный PHP / `symfony/string` / `nesbot/carbon`.
- `Timer::class` → `microtime(true)`.
- `XML::util`, `XMLObj::util` → `SimpleXMLElement`.
- `Email::util` → `symfony/mailer`.
- `File::util`, `Upload::util` → `symfony/filesystem`.

После каждой замены: запустить smoke-тест game-origin (login → главный
экран без ошибок) и удалить заменённые файлы.

### Ф.3. Замена ядра (средний риск)

- `Logger::class` → `monolog/monolog`.
- `Curl::class`, `Fsock::class`, `HTTP_Request::class`,
  `Request_Adapter::class` → `guzzlehttp/guzzle`.
- `Cache::class` → `symfony/cache` (на FilesystemAdapter для совместимости с
  существующим cache-каталогом).
- `Request::class` → `symfony/http-foundation`.
- `Hook::class`, `Plugin::abstract_class` → `symfony/event-dispatcher` или
  удалить, если не используются.

### Ф.4. Замена БД (средний риск)

- `Database::abstract_class`, `DB_MySQL`, `DB_MySQLi`, `DB_MYSQL_PDO` →
  `doctrine/dbal`. Создать тонкий адаптер `LegacyDB` с теми же методами
  (`query`, `select`, `insert`, ...), реализованный поверх DBAL — это
  избавляет от массовых правок в `src/game/`.
- Прогнать ручной smoke-тест на тестовой БД (логин, переход по экранам,
  одна транзакция: построить здание / отправить флот).

### Ф.5. Замена шаблонизатора (минимально)

- `Template::class`, `TemplateCompiler::class` → `smarty/smarty` (Composer).
  Smarty 4 поддерживает Smarty 2/3-синтаксис, должен подхватить существующие
  `.tpl` без правок (или с минимальными).
- Проверить, что 125 `.tpl` рендерятся корректно (визуальный смок основных
  экранов: Main, Galaxy, Empire, Fleet, Tech).

### Ф.6. Замена ядра-композитора и инфраструктуры

- `Core::class` → переписать как тонкий wire-up Composer-сервисов
  (DI-контейнер `symfony/dependency-injection` опционально).
- `init.php`, `Functions.php`, `AutoLoader::class` — заменить на
  Composer-bootstrap + namespaces в `composer.json`.
- `Language::*`, `LanguageCompiler::*`, импортёры/экспортёры —
  `symfony/translation`, либо упростить до single-locale если активно
  используется только русский.
- `Options::class`, `Collection::class` — пересмотреть; если это тонкие
  обёртки над массивами — заменить нативными типами/`ArrayObject`.
- `AjaxRequestHelper::abstract_class` → `symfony/http-foundation` Response.
- `User::class` — оценить; если это бизнес-логика — переместить в `src/game/`,
  пометить как наш новый код (свой copyright). Если это framework-уровень —
  отказаться от него.
- `Exception::interface`, `GenericException`, `IssueException` — стандартные
  PHP-исключения и кастомные классы под нашим copyright.

### Ф.7. Maintenance-скрипты

- `maintenance/TableCleaner`, `banIPClasses`, `ClearSessions`,
  `LanguageImporter/Exporter`, `XML_LanguageImporter`, `SortPhrases` →
  переписать как Symfony Console-команды или простые PHP-CLI скрипты под
  нашим copyright.

### Ф.8. Удаление Recipe и финализация

- Когда все callers переключены — `git rm -r src/core/{Cache.class.php,
  Logger.class.php, ...}` (все 54 файла).
- Прогон полного smoke-теста: login + 5 ключевых экранов + одна транзакция.
- Обновить `docs/origin-rights.md` (короткая ремарка про замену Recipe).
- Обновить `docs/plans/37-game-origin.md` — пометить, что core-фреймворк
  заменён на Composer-зависимости (если этот план ещё активен).
- Обновить `docs/project-creation.txt`: итерация 43.
- Финальный коммит. После этого можно безопасно прогонять план 40
  (license-audit) — он не найдёт GPL-файлов.

---

## Тестирование

- После Ф.2/Ф.3/Ф.4/Ф.5/Ф.6/Ф.7 — каждый раз smoke-тест: login →
  главный экран без ошибок в логах. Без этого следующая фаза не стартует.
- В конце Ф.8: полный сценарий — регистрация (или логин test/quoYaMe1wHo4xaci),
  Main, Galaxy, Empire, Fleet (отправить флот), Tech (запустить
  исследование), Build (построить здание). Без регрессий относительно
  current PHP-версии.
- Нет требования автотестов — game-origin сейчас без unit-тестов.
  Регрессии ловятся вручную.

---

## Итог

Замена 54 GPL-файлов на ~10 Composer-зависимостей под лицензиями MIT/BSD,
без переписывания шаблонов и игровой логики. Объём — 8 фаз, каждая с
точечным smoke-тестом. После выполнения GPL-конфликт фреймворка снят,
PolyForm Noncommercial валидна для всего oxsar-nova, план 40 (CI license
check) можно выполнять без блокировки.

---

## Постфактум: регрессия Ф.5 — XMLObj не Iterable (2026-04-27)

**Симптом:** левое меню в game-origin рендерится пустым у авторизованного
игрока — `<div id="leftMenu"><ul><li><nobr>Oxsar 2.14.2</nobr></li></ul>`,
без 46 элементов навигации. Найдено визуально на скриншоте главной страницы
во время работ по плану 50 Ф.4.

**Причина:** clean-room rewrite `XMLObj`
(`projects/game-origin-php/src/core/util/XMLObj.util.class.php`) — обёртка
над `SimpleXMLElement` — был сделан как «тонкая обёртка» с методами
`getAttribute/getName/getString/getChildren/getNode`, но **не реализовывал
`IteratorAggregate`**. Legacy-Recipe `XMLObj` был Iterable.

В `Menu::generateMenu()` (`src/game/Menu.class.php:84`) код:

```php
foreach($this->xml as $first) { ... }
```

где `$this->xml` — объект `XMLObj`. Без `IteratorAggregate` PHP перебирал
публичные свойства объекта (которых у XMLObj нет, есть только `private $node`),
foreach становился no-op'ом, `$this->menu` оставался пустым массивом. То же
самое касалось всех потребителей XMLObj через foreach
(`PlanetCreator`, `Options`).

**Smoke-тест плана 43 не поймал**, потому что Ф.8 говорит про «main экран
без ошибок в логах». PHP действительно не падал (foreach по пустому ничего
не выводит), но навигация молча пропадала — это видно только глазами в
браузере.

**Fix (1 коммит, 1 файл):**

```php
class XMLObj implements \IteratorAggregate
{
    // ...
    public function getIterator(): \Iterator
    {
        return new \ArrayIterator($this->getChildren());
    }
}
```

**Verification:** `Menu count` через CLI — было 0, стало 46. После очистки
Smarty-кэша (`src/cache/templates/standard/*.cache.php`) HTTP-ответ
`/?go=Main` содержит 54 `<li>` в `<div id="leftMenu">` (вместо 1).

**Урок:** для clean-room rewrite'а Recipe `XMLObj`/`Menu`/`Options` нужно
было сравнить публичный API не только по списку методов, но и по
implements-интерфейсам (Iterable, Countable, ArrayAccess и т.п.). В план 43
Ф.5 это явно не вписано — дефицит спецификации, не реализации.
