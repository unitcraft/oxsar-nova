# Пример рассуждения: слияние ext/ → game/ в game-origin

**Дата**: 2026-04-27
**Контекст**: план 37.5b. В legacy oxsar2 был отдельный override-слой
`ext/`, который через `NS::factory()` имел приоритет над базовыми
классами в `game/`. Юзер сказал «сливай ext в game, ты постоянно
забываешь про приоритет ext» — нужно вмерджить 16 Ext-классов в базовые
и удалить `src/ext/`.

**Результат**: 78 файлов, +2475/-2830 строк. Всё собралось без syntax
errors, smoke test 13 страниц возвращает HTTP 200 (хотя позже всплыл
другой баг — см. [меню-пример](2026-04-27-menu-empty-li.md)).

Этот документ — пример *моего хода рассуждения* при систематическом
слиянии множества файлов с разной структурой. Сохранён по просьбе
автора как образец.

---

## Шаг 1. Понять текущее состояние перед стартом

Окно с предыдущей сессией закрылось, юзер сказал «продолжи мерж файлов
по плану 37». Первое — прочитать план и сверить с реальным `git status`:

```bash
git status --short | head -30   # 53 файла удалены из ext/, ~17 untracked в game/
git log --oneline -5             # последний коммит — "первый запуск PHP-клона"
```

Уже видно: предыдущая сессия начала переносить файлы из `ext/` в `game/`
(удалила старые, создала новые), но не закончила слияние и не коммитила.

## Шаг 2. Не доверять «продолжай» — уточнить стратегию

В плане 37 §198 было прямо написано «оставляем ext/ нетронутым». Но
рабочее дерево показывает противоположное — `ext/` удалён. Это
противоречие.

**Стоп — не работаю молча, спрашиваю**:

> «План 37 говорит оставить ext/, но в working tree он удалён. Подтверди:
> мы сознательно отходим от плана? Какие классы уже трогал?»

Юзер: «да, сливай. Не помню что трогал.»

→ Зафиксировал отступление от плана как этап **37.5b** в plan-файле.
Перед началом работы.

## Шаг 3. Инвентаризация: сколько и каких файлов

```bash
for f in projects/game-origin-php/src/game/page/Ext*.class.php \
         projects/game-origin-php/src/game/Ext*.class.php; do
  base=$(basename "$f" .class.php)
  parent=$(grep -oP "class \w+ extends \K\w+" "$f" | head -1)
  printf "%-30s extends %-20s lines=%s\n" \
    "$base" "$parent" "$(wc -l < "$f")"
done
```

Вывод:
```
ExtMission           extends Mission        lines=1062
ExtEventHandler      extends EventHandler   lines=927
ExtSimulator         extends Page           lines=749
ExtRepair            extends Construction   lines=688
ExtPayment           extends Page           lines=542
ExtArtefacts         extends Page           lines=392
ExtArtefactMarket    extends Construction   lines=322
ExtAchievements      extends Page           lines=207
ExtMenu              extends Menu           lines=137
ExpedPlanetCreator   extends PlanetCreator  lines=149
ExtUserAgreement     extends Page           lines=99
ExtNotepad           extends Page           lines=59
ExtTutorial          extends Page           lines=61
ExtShipyard          extends Shipyard       lines=33
ExtSupport           extends Page           lines=31
ExtWidgets           extends Page           lines=31
```

**Что важно**: глядя на колонку `extends`, видно что **не все классы —
override базового**. Часть наследует прямо `Page` (страничный
контроллер), часть — реальные базовые (`Mission`, `EventHandler`).

→ Это значит **не одна стратегия слияния, а несколько**. Пытаться
слить всё одним подходом — ошибка.

## Шаг 4. Сверить с тем, есть ли базовый класс в game/

```bash
for cls in Achievements Construction ArtefactMarket Artefacts Mission \
           Notepad Payment Repair Shipyard Simulator Support Tutorial \
           UserAgreement Widgets EventHandler Menu PlanetCreator; do
  pages=$(ls projects/game-origin-php/src/game/page/${cls}.class.php 2>/dev/null)
  game=$(ls projects/game-origin-php/src/game/${cls}.class.php 2>/dev/null)
  printf "%-20s page:%s  game:%s\n" \
    "$cls" "${pages:+YES}" "${game:+YES}"
done
```

Вывод (фрагмент):
```
Achievements         page:  game:YES   ← ВАЖНО: бизнес-сервис в game/, нет page-контроллера
Mission              page:YES  game:   ← page есть, бизнес-класс не нужен
Shipyard             page:YES  game:
Notepad              page:  game:      ← нет ни там ни там — просто переименовать
Menu                 page:  game:YES   ← override базового
EventHandler         page:  game:YES   ← override базового
PlanetCreator        page:  game:YES
```

Сразу проявляется **конфликт имени**: `Achievements` есть и как
бизнес-класс в `game/`, и как страничный контроллер `ExtAchievements`
в `page/`. После переименования получится два `class Achievements` —
fatal error.

→ Помечаю себе: для Achievements нужно отдельное решение.

## Шаг 5. Понять как роутер находит классы

Если URL `/Achievements` ищется через autoloader — то будет загружен
**первый найденный** в `AUTOLOAD_PATH_APP="game/,game/page/"`. То есть
бизнес-класс перекроет страничный → URL `/Achievements` сломается.

Проверяю:

```bash
grep -rn "ext.*Ext\|new Ext\|class_exists\|page_class" \
  projects/game-origin-php/src/core/ projects/game-origin-php/src/game/NS.class.php \
  | head
```

Нашёл `NS::factory()` (строка 1395) — отдельная функция, которая ищет
сначала `ext/<path>/Ext<Class>.class.php`, **потом** `game/<path>/<Class>.class.php`.

Это значит: после удаления `ext/` фабрика возвращает базовый класс из
`game/<path>/`. То есть **имя класса в файле должно совпадать с URL** —
для `/Achievements` нужен `class Achievements` в `game/page/Achievements.class.php`.

→ Решение для Achievements: переименовать **бизнес-сервис** `Achievements
→ AchievementsService` (он чистый static), а `ExtAchievements` (page) →
`Achievements`. 14 вызовов `Achievements::` обновить через sed.

## Шаг 6. Прочитать «маленькие» классы, чтобы выделить категории

Для классов где `extends Page` (Notepad, Tutorial, Support и т.п.) —
прочитал по 1-2 файла:

```php
class ExtNotepad extends Page    // ← не override Notepad, а самостоятельный page-контроллер
class ExtSupport extends Page    // ← аналогично
class ExtTutorial extends Page   // ← аналогично
```

Вывод: эти классы **не накладываются** на одноимённые базовые (их нет).
Они — **самостоятельные контроллеры**, исторически с префиксом `Ext`.
Стратегия для них тривиальна: переименовать `ExtX → X` (файл и класс),
никакого слияния методов.

## Шаг 7. Классификация всех файлов

После шагов 3-6 у меня сформировалась таблица:

| Тип | Файлы | Стратегия |
|---|---|---|
| Override базового метода | Mission, Shipyard, Menu, EventHandler | Слить методы (stub'ы → реализация) |
| Самостоятельный extends Page | 10 классов | Переименовать ExtX → X |
| Не override (несмотря на имя) | ExpedPlanetCreator | Оставить как есть |
| Конфликт имени | ExtAchievements | Переименовать бизнес → Service |

Это сразу убрало риск «сделать одну стратегию для всех» (что бы дало
fatal error на Achievements и потерю override-логики на Menu/EventHandler).

## Шаг 8. От самого опасного к самому простому

Порядок работы:
1. **Mission** — самое большое слияние с реальным override (~1050 строк
   методов из ExtMission). Прочитал stub'ы в base — они уже были
   помечены комментариями `// controlFleet — реализация ниже из
   ExtMission` (предыдущая сессия начала работать). Дочистил: убрал
   плейсхолдеры, вставил реальные реализации перед закрывающей `}`
   класса. Удалил ExtMission.class.php. `php -l` — ✅.
2. **Shipyard** — простой decorator. Прочитал ExtShipyard (33 строки):
   просто добавляет TPL-переменную в начало `index()`, потом
   `parent::index()`. Стратегия: вставить дополнительный код в начало
   базового `Shipyard::index()`, удалить ExtShipyard.
3. **Menu** — сложнее. ExtMenu полностью переопределяет `generateMenu()`
   (137 строк), но вызывается только для `isMobileSkin()`
   (`NS::class.php:268`). Если просто заменить базовый метод — потеряем
   десктопный вариант. Решение: сохранить **обе** реализации, в базовый
   `generateMenu()` добавить диспатч `if (isMobileSkin()) return
   $this->generateMenuMobile();`. Десктоп остаётся, mobile получает
   спецлогику.
4. **EventHandler** — самое массивное (927 строк override на 2769
   базовых). Использовал systematic подход:
   - `awk` по сигнатурам методов в обоих файлах → таблица «base строка
     N, ext строка M, какой подход».
   - Базовые stub'ы (`removeEvent + return $this`, ~5 строк каждый)
     **полностью заменяются** ext-версией.
   - **Кроме `fReturn`**: в base нетривиальная логика (~70 строк), а
     ext добавляет prequel-проверку и в конце вызывает
     `parent::fReturn()`. Решение: вставить ext-логику в **начало**
     базового `fReturn`, остальная база сохраняется.
   - Новые методы из ext (которых в base нет — `disassembleOld`,
     `haltPosition`, `haltReturn`, 7 alien*) — добавлены в **конец**
     класса.
5. **ExpedPlanetCreator** — посмотрел использование (`grep -rn`):
   `Expedition.class.php:1060: new ExpedPlanetCreator(...)`. Это **не
   override**, а самостоятельный класс с собственным конструктором.
   Стратегия: оставить как есть, просто переместить в `game/`.
6. **10 самостоятельных** ExtX — массовый sed:
   ```bash
   for ext_file in ExtArtefactMarket ExtArtefacts ExtNotepad ExtPayment \
                   ExtRepair ExtSimulator ExtSupport ExtTutorial \
                   ExtUserAgreement ExtWidgets; do
     base="${ext_file#Ext}"
     sed -i "s/^class ${ext_file}\b/class ${base}/; \
             s/\b${ext_file}::/${base}::/g" \
       "${ext_file}.class.php"
     mv "${ext_file}.class.php" "${base}.class.php"
   done
   ```
7. **Achievements** — переименование бизнес-сервиса:
   - `sed -i 's/\bAchievements::/AchievementsService::/g'` по 7 файлам,
     где видно `Achievements::` (нашёл через `grep -rln`).
   - `mv game/Achievements.class.php game/AchievementsService.class.php`.
   - Внутри файла: `s/^class Achievements/class AchievementsService/`.
   - `mv game/page/ExtAchievements.class.php game/page/Achievements.class.php`,
     `s/class ExtAchievements/class Achievements/`.
   - Также `Achievements::loadAchievementsTemplateData(...)` внутри
     самого нового page-класса заменены тем же sed.

## Шаг 9. Очистка инфраструктуры

После того как все файлы перенесены, упростил то, что искало в `ext/`:

- `NS::factory()` — убрал ветку поиска в `ext/<path>/Ext<Class>.class.php`,
  оставил только `game/<path>/<Class>.class.php`.
- `AutoLoader.php` — убрал блок `AUTOLOAD_PATH_APP_EXT` (константа уже
  была удалена в 37.5).
- `NS.class.php:268` — `new ExtMenu` → `new Menu` (Menu сам диспатчит
  через `isMobileSkin()`).

## Шаг 10. Проверка после слияния

1. **php -l каждого изменённого файла**:
   ```bash
   find projects/game-origin-php/src -name "*.php" -exec php -l {} \; \
     2>&1 | grep -v "^No syntax errors" | head
   ```
   Нашёл pre-existing ошибки (`Upload.util.class.php`,
   `Float.util.class.php`, `Officer.class.php`, `Sql.inc.php`) — не
   связаны со слиянием, отдельно.

2. **Поиск осиротевших ссылок**:
   ```bash
   grep -rEn "\bExt(Mission|Shipyard|Menu|...)\b" projects/game-origin-php/src/
   ```
   Нашёл только комментарии (например `// Из ExtShipyard:` в Shipyard
   после слияния — это документирующие маркеры, оставил). Реальных
   `new ExtX()` или `class_exists("ExtX")` в коде нет.

3. **Smoke test в docker**:
   - `dev-login.php` → JWT cookie → `/?go=Main` → HTTP 200, "Мир Oxsar".
   - 13 ключевых страниц (Mission, Shipyard, Achievements, Notepad,
     Tutorial, Widgets, Support, UserAgreement, Repair, Simulator,
     Payment, Artefacts, ArtefactMarket) — все HTTP 200, 0 fatal в
     `docker compose logs php`.

## Шаг 11. Документация решения

После завершения — записал в plan **подробный** раздел 37.5b:
- Контекст (почему отступаем от §198)
- Принятое решение и причины
- Таблица стратегий по каждому Ext-классу
- Памятка: правило про обязательный ext/ актуально только для legacy
  (не для game-origin)

Также обновил memory `reference_legacy_ext.md` — добавил явный блок
«ВАЖНО: только legacy, не game-origin».

## Уроки и метод

1. **Нельзя «сливать всё одним подходом»**, когда исходный набор
   неоднороден. Шаг инвентаризации (extends X, lines, есть ли базовый)
   дал разделение на 4 разных стратегии — это сразу убрало риск
   fatal-error на Achievements и потери override-логики на Menu.

2. **Конфликт имён в PHP-классах** не виден до runtime. Проверять
   надо **до** переименования: «если файл назван X.class.php и в нём
   `class X`, есть ли где-то ещё `class X`?». Если да — нужно
   переименовать **другую** сторону.

3. **Самое опасное — первым**. Mission и EventHandler — массивные
   слияния, где легко ошибиться. Делал их когда мозг свежий, оставил
   тривиальное переименование (10 классов) на конец, когда выработался
   автоматизм.

4. **systematic — через awk/sed**:
   - `awk` по сигнатурам методов = таблица соответствий base ↔ ext.
   - `sed` для массового переименования с одним паттерном — лучше
     ручных правок (меньше шанс пропустить).
   - Цикл `for ext_file in ...; do sed ...; mv ...; done` для
     повторяющегося действия — безопаснее, чем 10 ручных копий.

5. **Stub vs реальный override — это разные слияния**.
   - Stub в base (`return $this`) → просто **заменить** на ext-версию.
   - Реальная логика в base + ext дополняет → **либо** объединить (ext
     в начало, base в конец), **либо** превратить base в `xxxBase` и
     дать ext-версию имя `xxx` (грязнее, но иногда необходимо).
   - Для `fReturn` выбрал первое — ext-логика срабатывает только при
     `!isPlanetOccupied`, базовая логика идёт после.

6. **Документировать почему**, не только что. В plan-файле подробно
   расписал не «слили ExtMission в Mission», а «причина — забывалось
   что ext имеет приоритет, привело к багам». Это нужно будущему
   разработчику (и мне в следующей сессии), чтобы не пытаться
   восстановить старую структуру «потому что план так говорил».

7. **После большого слияния — сразу запускать end-to-end smoke test**.
   `php -l` ловит синтаксис, но не runtime-ошибки. 13 HTTP-запросов с
   проверкой `Fatal error` в логах — занимает 30 сек, ловит большинство
   проблем. (Хотя один баг всё равно проскочил — см.
   [меню-пример](2026-04-27-menu-empty-li.md), `isMobileSkin()` приоритет
   операторов.)

## Связанные коммиты и файлы

- `3d172a1d9 refactor(game-origin): план 37.5b — слияние ext/ → game/`
  (78 файлов, +2475/-2830)
- `ec013e817 fix(game-origin): pre-existing баг приоритета операторов
  в isMobileSkin/isFacebookSkin` (последствие слияния — обнажился
  старый баг)

Файлы:
- [docs/plans/37-game-origin.md](../plans/37-game-origin.md) — раздел 37.5b
- [docs/project-creation.txt](../project-creation.txt) — итерация 37.5b
- [projects/game-origin-php/src/game/](../../projects/game-origin-php/src/game/) —
  результат слияния
- [Memory: reference_legacy_ext](../../../../../../Users/Евгений/.claude/projects/d--Sources-oxsar-nova/memory/reference_legacy_ext.md) —
  обновлено («только для legacy, не game-origin»)
