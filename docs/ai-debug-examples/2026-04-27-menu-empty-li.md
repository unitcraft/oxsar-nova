# Пример отладки: пустое меню в game-origin

**Дата**: 2026-04-27
**Контекст**: после слияния ext/ → game/ в game-origin (план 37.5b) пользователь
заметил скриншот с почти пустым левым меню (только "Oxsar 2.14.2" внизу).
**Симптом**: в HTML видно `<ul><li></li><li></li><li></li><li></li><li></li><li>Oxsar 2.14.2</li></ul>`
**Корневая причина**: pre-existing баг приоритета операторов в `isMobileSkin()`
из legacy oxsar2, который раньше не активировался, но обнажился после слияния.
**Фикс**: одна пара скобок в двух функциях.

Этот документ — пример *моего хода рассуждения* при поиске бага в чужом коде:
от первого взгляда на симптом до точечного фикса. Сохранён по просьбе автора
проекта как образец для будущих таких задач.

---

## Шаг 1. Симптом и первая гипотеза

Вижу скриншот: header с ресурсами (Металл/Кремний/Водород/Энергия/Кредиты),
сверху наезжает блок "Пополнить кредиты". Меню слева пустое.

Сразу две проблемы — наезд верстки и пустое меню. Пользователь обратил
внимание именно на меню — **это важно**, потому что наезд может быть следствием
того же бага (иногда пустые контейнеры нарушают сеточную раскладку).

Первая гипотеза: после слияния `ExtMenu` → `Menu` я что-то сломал.

## Шаг 2. Воспроизводимость и грубая локация

```bash
curl -sL -b /tmp/cookies.txt -c /tmp/cookies.txt -o /tmp/main.html \
  "http://localhost:8092/?go=Main"
grep -nE "<ul|navigation|<li|</ul>" /tmp/main.html | head
```

Вижу:
```html
<div id="leftMenu"> <ul>
  <li></li><li></li><li></li><li></li><li></li>
  <li><nobr>Oxsar 2.14.2</nobr></li>
</ul></div>
```

5 пустых `<li>` и один с версией. Версия в шаблоне отдельная (`<li><nobr>Oxsar
{const=OXSAR_VERSION}</nobr></li>`) — то есть из навигации генерируется ровно
5 пустых элементов.

**Что важно**: 5 — не 0 и не 30+. Это значит цикл foreach отработал, массив
не пустой, но элементы выглядят как `<li></li>`. То есть в массиве **есть**
объекты, но их свойства `inner` и `content` пусты или отсутствуют.

## Шаг 3. Шаблон: что он рендерит

```php
{foreach[navigation]}<li{loop}inner{/loop}>{loop}content{/loop}</li>...{/foreach}
```

Шаблон ожидает массив с полями `inner` и `content`. Если оба пустые — выходит
`<li></li>`. То есть вопрос: **что попадает в массив `navigation`?**

## Шаг 4. Источник массива

`NS::class.php:266-275`:
```php
if (isFacebookSkin() || isMobileSkin()) {
    $menu = new Menu("Menu");
    Core::getTPL()->addLoop("menu_headers", $menu->getMenuTitles($planets));
    Core::getTPL()->addLoop("navigation", $menu);
} else {
    Core::getTPL()->addLoop("navigation", new Menu("Menu"));
}
```

Обе ветки добавляют `new Menu("Menu")`. Я уже знаю (из прошлой работы),
что `Menu::generateMenu()` имеет диспатч на `generateMenuMobile()` через
`isMobileSkin()` — это я добавил при слиянии.

## Шаг 5. Проверка XML-источника

Может быть, файл XML повреждён? Тестирую напрямую в контейнере:

```bash
docker exec docker-php-1 php -r "
\$xml = simplexml_load_file('/var/www/src/game/xml/Menu.xml');
foreach (\$xml as \$first) {
    echo 'group: ' . \$first->getName() . ', children: ' . count(\$first->children()) . PHP_EOL;
}
"
```

Вывод: `info: 12 children, info2: 7, info3: 5, info4: 5, info5: 3, info6: 2`.
**Итого 8 групп с детьми**. То есть XML парсится правильно, а `Menu`
почему-то из этого делает 5 пустых элементов.

## Шаг 6. Гипотеза: PHP 8 Deprecated → null

В логах php-fpm раньше видел:
```
Deprecated: Return type of XMLObj::getChildren() should either be compatible
with SimpleXMLElement::getChildren(): ?SimpleXMLElement, or the
#[\ReturnTypeWillChange] attribute should be used to temporarily suppress the notice
```

Может быть, в PHP 8.3 это уже не Deprecated, а fatal? Тогда `getChildren()`
возвращает null, и `foreach` ничего не делает.

Добавил `#[\ReturnTypeWillChange]` в `XMLObj.util.class.php`. **Не помогло** —
по-прежнему 5 пустых li. Гипотеза неверна.

## Шаг 7. Прямой замер: добавляю debug-логи в Menu

Когда логические рассуждения не дают ответа, эффективнее добавить временный
`error_log()` и посмотреть что происходит реально.

```php
protected function generateMenu()
{
    error_log("[DEBUG] generateMenu called, isMobile=" . (isMobileSkin() ? "1" : "0")
              . ", xml type=" . get_class($this->xml)
              . ", xml count=" . count($this->xml->children()));
    if (isMobileSkin()) {
        return $this->generateMenuMobile();
    }
    // ...
}
```

Триггерю запрос, читаю `docker compose logs php --tail=20`:

```
[DEBUG] generateMenu called, isMobile=1, xml type=XMLObj, xml count=8
```

**`isMobile=1`!** Хотя я тестирую через обычный curl на десктоп-эндпоинт.
Это означает, что `isMobileSkin()` возвращает true там, где не должна, и
вызывается mobile-вариант меню. А mobile-вариант формирует совсем другую
структуру массива (`output_menu[i]['items']`), несовместимую с шаблоном
`{loop}content{/loop}` из десктоп-ветки.

## Шаг 8. Корневая причина — `isMobileSkin()`

```php
function isMobileSkin()
{
    if (!defined('SN_FULLSCREEN')) {
        if (defined('SN'))   return false;
        if (defined('MOBI')) return true;
    }
    return !empty($_SESSION["userid"]) && $_SESSION["skin_type"] ?? "standard" == SKIN_TYPE_MOBI;
}
```

Внимательно смотрю на последнюю строку. Из-за приоритета операторов в PHP
(`==` имеет приоритет выше чем `??`), это парсится как:

```php
return !empty($_SESSION["userid"]) && ($_SESSION["skin_type"] ?? ("standard" == SKIN_TYPE_MOBI));
```

Разбор:
1. `"standard" == SKIN_TYPE_MOBI` — обычно `false`.
2. `$_SESSION["skin_type"] ?? false` — если `$_SESSION["skin_type"]` определён
   (например, "standard"), вернёт его → строку "standard" → **truthy**.
3. `!empty(userid) && "standard"` → **`true`**.

Получаем `isMobileSkin() === true` для любого авторизованного юзера, у кого
`$_SESSION["skin_type"]` хоть как-то выставлен.

То же самое в `isFacebookSkin()` — копи-пейст той же ошибки.

## Шаг 9. Почему этот баг не проявлялся в legacy

В legacy oxsar2 эта же ошибка была — но у них десктоп-меню работал. Почему?

Гипотеза: в legacy `NS::factory()` искал `ext/page/ExtMain.class.php` сначала,
а ExtMain через ExtMenu использовал mobile-вариант. То есть в legacy ВСЕГДА
работал mobile-ветка, а десктоп не использовался. Шаблон `layout.tpl` тоже
имел соответствующую mobile-ветку для navigation, и она работала корректно.

После моего слияния (план 37.5b) шаблонная десктоп-ветка осталась
(`{if[ !isFacebookSkin() && !isMobiSkin() ]}`), а из-за бага приоритета
эта ветка никогда не активна → всегда идёт `else if[ isFacebookSkin() ]`
для FB или mobile-структура для mobile, но вышло что `Menu::generateMenu()`
тоже шёл в mobile-ветку (мой диспатч), а шаблон рендерил его как десктоп.

То есть это **взаимодействие** двух багов: pre-existing в isMobileSkin +
мой новый диспатч в Menu, который начал ловить тот старый баг.

## Шаг 10. Фикс

Одна пара скобок в двух местах:

```diff
- return !empty($_SESSION["userid"]) && $_SESSION["skin_type"] ?? "standard" == SKIN_TYPE_FB;
+ return !empty($_SESSION["userid"]) && (($_SESSION["skin_type"] ?? "standard") == SKIN_TYPE_FB);

- return !empty($_SESSION["userid"]) && $_SESSION["skin_type"] ?? "standard" == SKIN_TYPE_MOBI;
+ return !empty($_SESSION["userid"]) && (($_SESSION["skin_type"] ?? "standard") == SKIN_TYPE_MOBI);
```

После фикса: `isMobile=0` → десктоп-ветка `Menu::generateMenu()` → меню
заполнено. HTML вырос с 6931 до 11216 байт. Видны Обзор, Сырьё, Постройки,
Исследования, Верфь, Оборона, Флот, и т.д.

## Шаг 11. Удаление debug-логов

После подтверждения убрал `error_log("[DEBUG] ...")` из `generateMenu()`.

## Уроки и метод

1. **Доверяй симптому, не первой гипотезе**. Я начал с предположения "я
   что-то сломал в Menu при слиянии". На самом деле я ничего не сломал в
   логике — но мой новый диспатч `if(isMobileSkin())` оказался триггером
   старого бага, который раньше дремал.

2. **Считай артефакты**. 5 пустых li — это не 0 и не 30+. Конкретное число
   подсказывает: цикл foreach отрабатывает (5 раз = 5 групп info-info5 без
   info6), но что-то происходит на втором уровне. Если бы было 0 — была бы
   гипотеза про сломанный шаблон.

3. **Когда логика не помогает — ставь print**. После двух неудачных гипотез
   (XML повреждён, PHP 8 Deprecated) пора просто посмотреть, что
   реально происходит. `error_log()` в горячую функцию + `docker logs` — за
   30 секунд получаю достоверный факт `isMobile=1`, который сразу указывает
   на корневую причину.

4. **Читай PHP-код вслух с приоритетами в голове**. Запись `A && B ?? C == D`
   читается слева направо как «A и B, либо если null то C==D». А PHP читает
   как `A && (B ?? (C == D))`. Эта ловушка существует в любом PHP-коде, где
   `??` встречается рядом с `&&`/`||`/`==`. Полезное правило: если в строке
   несколько операторов разной арности — **ставь скобки явно**, даже когда
   "и так понятно".

5. **Бесплатный фикс, спасибо багу**. Сам факт что в legacy десктоп-ветка
   шаблона никогда не использовалась — это скрытый mёртвый код. Теперь, когда
   она работает, можно удалить fallback-ветки. Но это уже отдельная задача.

## Связанные коммиты

- Коммит фикса: `ec013e817 fix(game-origin): pre-existing баг приоритета
  операторов в isMobileSkin/isFacebookSkin`
- Слияние ext/ (предыдущий, обнаживший баг): `3d172a1d9 refactor(game-origin):
  план 37.5b — слияние ext/ → game/`

## Файлы

- [projects/game-origin-php/src/game/Functions.inc.php](../../projects/game-origin-php/src/game/Functions.inc.php) — функции `isMobileSkin()`, `isFacebookSkin()`
- [projects/game-origin-php/src/game/Menu.class.php](../../projects/game-origin-php/src/game/Menu.class.php) — `generateMenu()` (десктоп) и `generateMenuMobile()`
- [projects/game-origin-php/src/templates/standard/layout.tpl](../../projects/game-origin-php/src/templates/standard/layout.tpl) — рендер меню
- [projects/game-origin-php/src/game/xml/Menu.xml](../../projects/game-origin-php/src/game/xml/Menu.xml) — структура меню
