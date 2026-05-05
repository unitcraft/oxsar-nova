# Пример исследования: тик ресурсов в шапке (1:1 с legacy)

**Дата**: 2026-05-05
**Контекст**: при доводке origin-фронта (план 72.1.59) пользователь спросил
«показ ресурсов в хидере тикает?». Я ответил «нет, snapshot, как в legacy».
Пользователь возразил: «в легаси пхп тикает».
**Симптом**: я ошибся, не проверив legacy. Нужно было найти реальный механизм
тика в `oxsar2` и воспроизвести в origin.
**Корневая причина**: в legacy сервер инжектит inline-`<script>` с
`jquery.iterator.js` через `Page::proceedRequest()`, поэтому тика нет в
`layout.tpl` или в `main.js` — он рождается в php-классе и попадает в HTML
как часть `{@real_metal}`.
**Фикс**: добавить клиентскую интерполяцию в `TopHeader.tsx` (setInterval
1с + linear extrapolation `metal + per_sec * dt`) с теми же граничными
условиями, что у legacy iterator: cap при росте, 0 при потреблении, тик
выключается когда cap уже достигнут и нет потребления.

Этот документ — пример *моего хода рассуждения* при сверке поведения с
legacy, когда первое впечатление оказалось неверным. Сохранён по просьбе
автора проекта как образец для будущих таких задач.

---

## Шаг 0. Ошибка первого ответа

Пользователь: «показ ресурсов в хидере тикает?». Я заглянул в свой код
([TopHeader.tsx](../../projects/game-nova/frontends/origin/src/layout/TopHeader.tsx))
и в `useResolvedPlanet` (`staleTime: 30_000`, без `refetchInterval`) —
увидел статичный snapshot, ответил «не тикает, как в legacy».

**Что я сделал не так**: предположил поведение legacy, не открыв ни одного
файла из `d:\Sources\oxsar2`. Ответ был основан на здравом смысле
(«PHP — server-side рендер, F5 для обновления») вместо проверки.

Когда пользователь возразил «в легаси пхп тикает», я не стал спорить, а
сразу пошёл искать механизм. Это правильный режим — пользователь играл в
legacy, его слово первично.

## Шаг 1. От ID элемента к месту инициализации

В `TopHeader.tsx` я уже копировал DOM legacy — там стоят узнаваемые id
`header_layout_metal/silicon/hydrogen`. Это удобный якорь: ищу их в legacy.

```text
Grep header_layout_metal|header_layout_silicon|header_layout_hydrogen
  → d:\Sources\oxsar2\www\templates\standard\layout.tpl:285-301
```

Открываю layout.tpl. Нахожу:

```html
<span id='header_layout_metal' class="{@metalClass}">{@real_metal}</span>
```

`{@real_metal}` — переменная TPL-движка. Значит её куда-то assign'ит php.
Inline-JS в этом блоке нет. Вокруг — ни setInterval, ни ajax. Если бы тик
был чисто template-side, я бы его уже видел.

**Гипотеза 1**: тик прячется в `main.js` или другом подключаемом скрипте.

```text
Grep setInterval|header_layout|prod_per
  → d:\Sources\oxsar2\www\js\main.js — ничего по тику
  → ext/ — пусто
```

Пусто. Гипотеза 1 не подтвердилась.

## Шаг 2. Кто assign'ит `real_metal`

Раз JS не находит, значит магия в php. Шаблон проксирует имена через
`Template::get()` — обычная hash-map, без `__get`. Значит где-то есть
явный `assign("real_metal", ...)`.

```text
grep -rn 'real_metal\|"real_' d:/Sources/oxsar2/www --include='*.php' | grep -v cache
  → game/page/Page.class.php:93:
      Core::getTPL()->assign("real_" . $key, $span_str);
```

**Это корень**. Базовый класс Page (от которого наследуется Main, Empire,
Profession и пр.) на каждом запросе формирует `real_<key>` для метала,
кремния, водорода. Открываю окружающий код:

```php
foreach(array("metal", "silicon", "hydrogen") as $key) {
    $current[$key]    = NS::getPlanet()->getData($key);
    $storage[$key]    = NS::getPlanet()->getStorage($key);
    $production[$key] = NS::getPlanet()->getProd($key);

    if($storage[$key] > $current[$key] || $production[$key] < 0) {
        $span_str = "<script type='text/javascript'>
            $(function($) {
                var options = {
                    startNum: {$current[$key]},
                    stopNum:  ".($production[$key] < 0 ? 0 : $storage[$key]).",
                    step:     {$production[$key]}/3600.0
                }
                $('.iter_{$key}').iterator(options);
            });
            </script>
            <span class='iter_{$key}".($production[$key] < 0 ? " false" : "")."'>"
            .fNumber($current[$key])."</span>";
    } else {
        $span_str = fNumber($current[$key]);
    }
    Core::getTPL()->assign("real_" . $key, $span_str);
}
```

Вот она, вся правда о тике в одной функции. Расшифровываю:

1. **Условие тика**: `storage > current` (склад не полон) **или**
   `production < 0` (потребление). Иначе — статичная цифра.
2. **Шаг**: `production / 3600` единиц/секунду (per_sec).
3. **Начало**: текущее значение из БД на момент запроса.
4. **Конец**: `storage` при росте, `0` при потреблении.
5. **Цвет**: `class="false"` (красный) если потребление.
6. **Реализация**: jQuery-плагин `jquery.iterator.js` — клиентский timer,
   который инкрементит число в DOM шагами `step` каждый тик.

Это гибрид: первый snapshot — server-side, дальше клиент сам
экстраполирует пока пользователь не уйдёт со страницы.

## Шаг 3. Проектирование клиентского эквивалента

В origin DTO планеты ([api/types.ts](../../projects/game-nova/frontends/origin/src/api/types.ts))
уже есть всё что нужно:

```ts
metal: number;
metal_per_sec: number;   // = production / 3600 в legacy
metal_cap: number;       // = storage
last_res_update: string; // ISO timestamp
```

Наш `metal_per_sec` уже делён на 3600 (legacy step) и backend пересчитывает
БД-значение в `metal` на момент `last_res_update`. Значит формула на
клиенте:

```ts
val = base + per_sec * (now - last_res_update_ms) / 1000
```

с теми же граничными условиями:
- `per_sec >= 0`: cap'аем `metal_cap` (legacy `stopNum: storage`).
- `per_sec < 0`: cap'аем 0 (legacy `stopNum: 0`).

Тик нужен только когда показывает смысл (legacy подключает iterator
условно). У нас можно тикать всегда — если cap уже достигнут,
`val + per_sec*dt > cap` сразу cap'ается, число не меняется. Лишний
re-render раз в секунду — терпимо за простоту.

## Шаг 4. Реализация и проверка

Обернул логику в `tick(base, perSec, cap, lastUpdateIso)` внутри
[TopHeader.tsx](../../projects/game-nova/frontends/origin/src/layout/TopHeader.tsx),
завёл `useEffect` с `setInterval(1000)` для force-rerender.

`useState(() => Date.now())` для `now`, чтобы избежать пересчёта `Date.now()`
на mount (мелочь, но React это любит).

Проверка:
- `tsc --noEmit` — чисто.
- В браузере: смотрю шапку на `/profession`, цифра металла увеличивается на
  ~`per_sec` каждую секунду; если открыть planet с пустым складом и
  отрицательным энергобалансом — водород убывает (если бы он потреблялся).

## Шаг 5. Ограничения относительно legacy

Я не воспроизвёл одну деталь: legacy скрывает `<script>` целиком если
`storage <= current` и `production >= 0`. У нас тик включён всегда.
Визуально не отличается (cap'ается сразу), но это +1 setInterval впустую.

Если станет узким местом — можно добавить условие в `useEffect` (запускать
интервал только если хотя бы один ресурс «живой»). Сейчас оптимизировать
не стал — `setInterval(1000)` с одним setState стоит наносекунды.

Энергия и кредиты не тикают — в legacy `Page::proceedRequest` тоже не
оборачивает их в `iter_*` (энергия — функция от зданий, не от времени;
кредиты — событийная сущность).

## Что я уношу из этого случая

1. **Не предполагать поведение legacy без проверки**, особенно когда
   пользователь играл в неё годами. «Здравый смысл про PHP» проиграл
   фактическому коду. Если бы я открыл `Page.class.php` *до* первого
   ответа, я бы не дал неверный ответ и не пришлось бы делать второй
   заход.
2. **ID элементов в DOM — лучший якорь** для cross-stack сверки. Я
   копировал id `header_layout_metal` буквально, и это позволило за один
   grep найти legacy-владельца поведения. Если бы был свой id — нужно
   было бы рыться по семантике.
3. **php → tpl assign — ищется одним rg**: `assign\(.real_` достаточно
   узкий шаблон, чтобы вычистить шум кэша и тестов. Cache-файлы (
   `www/cache/templates/*.cache.php`) фильтрую через `grep -v cache` —
   они дублируют исходный tpl и засоряют поиск.
4. **legacy js нужно искать в php**: php-классы инжектят скрипты в HTML
   через `addHTMLHeaderFile` и через inline-`<script>` в строке assign'а.
   Просто `grep setInterval *.js` пропустит эти случаи. Поэтому мой
   первый поиск в `main.js` ничего не дал — а реальный механизм лежал
   в `Page.class.php`.
5. **Перенос «тика» из jQuery iterator в React** — однострочная формула
   плюс setInterval. Но граничные условия (cap при росте, 0 при
   потреблении) — критичны: без них при cap-полном складе цифра
   убегала бы в бесконечность.
