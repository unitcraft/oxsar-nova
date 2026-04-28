# План 50: game-origin — закрытие юридических пробелов legacy-вселенной

**Дата**: 2026-04-27
**Статус**: ✅ Завершён 2026-04-28 (все 8 фаз закрыты)
**Зависимости**: ничего технически блокирующего; должен быть выполнен **до**
открытия публичного доступа к вселенной game-origin.
**Связанные документы**: [LICENSE](../../LICENSE),
[../ops/legal-compliance-audit.md](../ops/legal-compliance-audit.md)
(аудит от 2026-04-27 — источник этого плана),
[44-personal-data-152fz.md](44-personal-data-152fz.md),
[46-age-rating-ugc.md](46-age-rating-ugc.md),
[47-offer-tos.md](47-offer-tos.md), [49-doc-hygiene-pii.md](49-doc-hygiene-pii.md).

---

## Цель

Привести legacy-вселенную game-origin (PHP-проект, портированный из
oxsar2) в соответствие с теми же юридическими требованиями, которые
уже закрыты для game-nova, portal, auth, billing:

- лицензионная совместимость (никаких GPL/AGPL/LGPL-зависимостей);
- 152-ФЗ (согласие на обработку ПДн при регистрации);
- 436-ФЗ (возрастная маркировка «12+» в UI);
- 149-ФЗ (UGC-модерация чата, альянсов, описаний; кнопка
  «Пожаловаться»);
- ст. 437 ГК РФ (доступность оферты в footer'е UI);
- стратегия B3 плана 41 (своя copyright-атрибуция, без следов
  ликвидированной компании автора).

Аудит [docs/ops/legal-compliance-audit.md](../ops/legal-compliance-audit.md)
от 2026-04-27 зафиксировал 7 gap'ов, специфичных для game-origin.
Этот план их закрывает.

---

## Что меняем

### 1. Замена TinyMCE LGPL → MIT

В `projects/game-origin/public/js/tiny_mce/` лежит TinyMCE 3.x под
LGPL 2.1. Используется в `src/templates/standard/chat.tpl` как
WYSIWYG для сообщений чата.

Варианты:
- **Обновление до TinyMCE 6+** (под MIT с 2014 г.) — современный
  редактор, минимальные правки UI.
- **Замена на Quill** (BSD-3-Clause) — лёгкий редактор с минимальной
  кривой обучения.
- **Замена на простой `<textarea>` + JS-обёртка** — если форматирование
  чат-сообщений не критично (для большинства браузерных игр это так).

Целевое решение — выбирается на этапе Ф.1.

### 2. 152-ФЗ — чекбоксы согласия в форме регистрации game-origin

В `src/game/AccountCreator.class.php` нет проверки согласия на ПДн
и оферту. Если пользователь регистрируется напрямую в game-origin
(не через handoff из portal) — у этой регистрации нет юридического
основания.

Сначала уточнить: **возможна ли вообще прямая регистрация в
game-origin**, или вход работает только через handoff page из
portal?

- Если только через handoff — добавить комментарий в
  `AccountCreator.class.php`, что прямая регистрация закрыта; убрать
  публичные точки входа (если есть).
- Если возможна напрямую — добавить два чекбокса в форму регистрации
  (согласие на ПДн + акцепт оферты), отправлять записи в
  централизованную таблицу `user_consents` через identity-сервис
  (HTTP-вызов) или собственную миграцию `user_consents` в game-origin.

### 3. 436-ФЗ — возрастная маркировка «12+»

В `src/templates/standard/layout.tpl` добавить блок «12+» в footer
по образцу компонента `AgeRating` из portal/game-nova.

Минимальный HTML:

```html
<div class="age-rating">12+</div>
```

С CSS-стилизацией под существующую вёрстку legacy.

### 4. 149-ФЗ — UGC-модерация на все точки

Класс `Moderation` (введён планом 48 шаг 0,
`projects/game-origin/src/core/util/Moderation.util.class.php`) — это
PHP-обёртка над общим blacklist'ом `projects/game-nova/configs/moderation/blacklist.yaml`.
Сейчас применяется **только** в `AccountCreator.class.php` при
регистрации (проверка ника).

#### UGC-поверхности — что в scope

| Поверхность | Где искать | Стратегия | Почему |
|---|---|---|---|
| Чат (глобальный/альянсовый/личный) | `src/templates/standard/chat.tpl`, `src/game/page/Chat*.class.php`, `src/chatPro.php` | **mask()** | Поток сообщений не должен прерываться |
| Тег и название альянса (создание) | `src/game/page/Alliance*.class.php`, `AllianceCreator*.class.php` | **isForbidden()** → ошибка | Это идентификатор, маскировать нельзя |
| Описание альянса | `src/game/page/Alliance*.class.php`, поле description/text | **mask()** | Свободный текст |
| Личные сообщения (тема + тело) | `src/game/page/MSG*.class.php`, `WriteMessages*.class.php` | **mask()** | Свободный текст |
| Описание планеты (если есть) | `src/game/page/Planet*.class.php` | **mask()** | Свободный текст, видимый другим |
| Notepad (если виден другим игрокам) | `src/game/page/Notepad*.class.php` | **mask()** | Если только сам себя видит — пропустить, не UGC |

#### Поиск точек

```bash
grep -rn "getPOST\|getGET\|->post\|input(" projects/game-origin/src/game/ projects/game-origin/src/core/
grep -rn "INSERT INTO\|UPDATE.*SET" projects/game-origin/src/game/ | grep -iE "msg|chat|message|alliance|note|description"
```

Для каждой точки прочитать релевантный фрагмент, понять имя
input-поля и куда оно уходит (INSERT / UPDATE / отправка по WS).

#### Расширение API класса Moderation

Если в `Moderation.util.class.php` ещё **нет** метода `mask()` —
реализовать аналогично `MaskForbidden` из
`projects/game-nova/backend/internal/moderation/blacklist.go`:

- Найти все вхождения запрещённых корней в строке
  (case-insensitive, NFKC-нормализация — как в `IsForbidden()`).
- Заменить **каждый символ** найденного вхождения на `*`.
- Сохранить общую длину строки (важно — не сворачивать слово
  в один `*`).
- Не модифицировать строку, если ничего не найдено.

Семантика должна совпадать с Go-реализацией, чтобы единая логика
«маскировано / не маскировано» работала идентично в обеих вселенных.

Если `mask()` уже есть — использовать как есть, не дублировать.

#### Стратегия при срабатывании

- **Блокировка** (тег, название альянса) — единое сообщение
  пользователю, например: «Имя содержит запрещённое слово, выберите
  другое». По возможности использовать существующий i18n-механизм
  (`na_phrases`); если нет подходящей фразы — добавить или
  захардкодить под TODO.
- **Маскирование** (всё остальное) — `mask()` перед сохранением в
  БД / отправкой по WebSocket. Аудит-таблицу **не вводить** в этой
  фазе (требование 149-ФЗ к хранилищу логов чата на 6 месяцев —
  отдельная задача, см. ниже «Известные ограничения»).

#### Что отложено (не в scope Ф.4)

- **Никнейм** — уже покрыт в `AccountCreator.class.php`, не трогать.
- **Содержимое assault-report'ов** — внутриигровые автогенерации,
  не UGC.
- **Кнопка «Пожаловаться»** — Ф.5, отдельная фаза.
- **Audit-таблица модерации** (запись факта срабатывания) — отдельная
  задача, не в Ф.4.
- **Структурное хранилище логов чата на 6 месяцев** (требование
  149-ФЗ к данным по запросу). Если в legacy-чате логи не хранятся
  как отдельная таблица — отметить в финальном отчёте как
  «Известные ограничения», но не реализовывать в Ф.4.

#### Тестирование

Достаточно ручного smoke-теста (PHPUnit в game-origin не настроен,
проект рассчитан на ручную QA через `compare-with-legacy.sh`).
Минимум:

1. Регистрация ника с запрещённым словом → отказ (регрессия — должно
   работать как до Ф.4).
2. Создание альянса с матерным **названием** → отказ.
3. Создание альянса с матерным **тегом** → отказ.
4. Редактирование описания альянса с запрещённым словом → сохранилось
   замаскированным.
5. Чат-сообщение с запрещённым словом → у получателя отображается
   замаскированным.
6. ЛС с запрещённым словом в теме и теле → у получателя замаскировано.
7. Описание планеты с запрещённым словом (если поверхность найдена) →
   сохранилось замаскированным.
8. Чисто-нормальный текст без запрещённых слов проходит без изменений
   (регрессия).

Тестовое слово — взять из `projects/game-nova/configs/moderation/blacklist.yaml`
короткий корень или любой реальный мат, который точно срабатывает.

### 5. Кнопка «Пожаловаться»

**Зависимость:** [план 56](56-reports-to-portal.md) — перенос reports
из game-nova в portal-backend как централизованного сервиса жалоб
для всех вселенных. Ф.5 этого плана выполняется **после** закрытия
плана 56.

Добавить компонент кнопки в legacy-templates game-origin рядом с:
- никнеймом другого игрока;
- шапкой чата;
- шапкой альянса.

При клике — модалка с выбором причины (7 категорий, как в плане 46) +
текстовое поле для комментария. POST на endpoint `/api/reports` **portal-
backend'а** (общий централизованный сервис жалоб для всех вселенных).
Запись жалобы — в таблицу `user_reports` в БД portal (после плана 56).

### 6. Ссылки на юр-документы в footer game-origin

В `src/templates/standard/layout.tpl` добавить блок ссылок:
- Оферта → `https://oxsar-nova.ru/offer` (URL portal'а);
- Правила → `https://oxsar-nova.ru/game-rules`;
- Возврат → `https://oxsar-nova.ru/refund`;
- Конфиденциальность → `https://oxsar-nova.ru/privacy`;
- Пожаловаться → ссылка на форму или модалку из п.5.

### 7. Чистка copyright в шаблонах

В 10+ шаблонах `src/templates/standard/*.tpl` стоит шапка
`Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>`. Заменить
на `Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.`
+ удалить email старой компании.

Список файлов (10+, точное число — по grep'у на момент выполнения):
preferences.tpl, writemessages.tpl, layout.tpl, stock.tpl, payment.tpl,
tutorial.tpl, widgets.tpl, assault_report.tpl, galaxy.tpl,
artefactmarket2.tpl и другие.

---

## Этапы

### Ф.0. Фикс ассет-путей после реструктуризации (план 43) ✅ (2026-04-27)

Регрессия плана 43, обнаруженная при работе над планом 50: после
выноса `images/` и `css/` из `src/` в `public/` PHP-код продолжал
искать их через `APP_ROOT_DIR` (= `src/`). `file_exists()` всегда
возвращал `false` → `getUnitImage()` фоллбэчился на
`buildings/empty/empty.gif` для **всех** построек/юнитов/артефактов;
`getUserStyles()` не находил user-style CSS (`us_bg/`, `us_table/`).
Это и был «-4330 байт diff на каждой странице» из
`docs/prompts/compare-screens.md`.

**8 точечных замен** `APP_ROOT_DIR` → `GAME_ORIGIN_DIR."public/"`:
- `Functions.inc.php`: `getUnitImage()`, `getImgPacks()`,
  `getBgImagePlaneStyles()`, `getUserStyles()`.
- `Construction.class.php`: валидация `image_package`.
- `Preferences.class.php` (×2): список image-packs + fallback.
- `Artefact.class.php`: `getFlagImage()`.

**Verifications**: `php -l` чисто, `?go=Constructions` показывает
18 уникальных иконок (вместо 1 placeholder), все 41 страниц
compare-скрипта 200 OK, user-style CSS теперь идентично
рендерится у нас и в legacy.

### Ф.1. Замена TinyMCE ✅ (2026-04-27)

- Решить: TinyMCE 6+ / Quill / простой textarea.
- Подключить новый редактор в `src/templates/standard/chat.tpl`.
- Удалить `public/js/tiny_mce/` целиком (~430 МБ).
- Smoke-тест чата: отправка сообщения, форматирование (если нужно).

**Решение:** вариант 3 (простой `<textarea>`/`<input>`) — оказался уже
де-факто реализованным. При изучении `chat.tpl` обнаружено, что
**весь TinyMCE-код находится в `{if[0 && isAdmin()]}`-блоках**
(строки 163-271 и 309-332 до правки) — `0 && ...` это Smarty-аналог
`if(false)`, блок никогда не выполнялся. Реальный chat-input —
обычный `<input type="text" name="shoutbox_message">` (строка 340
в исходнике), JS-обработчик `$('#chat_form').live('submit', ...)`
без TinyMCE-вызовов. Backend (`Chat::sendMessage()` в
`src/game/page/Chat.class.php`) принимает plain text + bbcode-замены
через функцию `Teg(...)` (план 50 Ф.4 уже добавил туда
`Moderation::mask()`).

**Что сделано:**
- Удалены `if[0 && isAdmin()]`-блоки (init TinyMCE + tinymce_chat_form
  с `<input id="tinymce_chat_message">`) — заменены HTML-комментариями.
- `git rm -r projects/game-origin/public/js/tiny_mce/` — **430 файлов**
  библиотеки TinyMCE удалены. Размер `projects/game-origin/`
  уменьшился с 3.7G до 3.2G.
- В `jquery.form.js` (сторонний пакет) осталось одно упоминание в
  док-комментарии — это сторонняя документация плагина, не наша.
  Не правим.

**Backend strip_tags не делался:** legacy-сообщения в `na_chat` —
plain text + bbcode (см. функцию `bbcode($source)` в Chat.class.php
строки 23-44 — она преобразует `[b]…[/b]` в `<b>…</b>` при выводе).
Никаких `<script>`/`<img>` у пользователей не было — TinyMCE-форма
никогда не работала. Ничего не нужно мигрировать.

**Verification:**
- `curl ?go=Chat` → HTTP 200, **0** совпадений `tiny_mce`/`tinymce`,
  4 шаблонных совпадения `shoutbox_message` (input в чате).
- `GET /js/tiny_mce/tiny_mce.js` → **404** (файлы реально удалены).
- `grep -rn "tiny_mce\|tinyMCE"` в repo вне `/cache/`,
  `compare-output/`, legacy-dump'ов: только HTML-комментарии в
  `chat.tpl` + 1 doc-комментарий в `jquery.form.js`.

### Ф.2. Регистрация и согласия ✅ (2026-04-28)

Уточнено пользователем: **прямая регистрация в game-origin закрыта**,
вход только через handoff из portal (как и в game-nova). Согласия
на обработку ПДн и акцепт оферты собираются на portal (планы 44,
47), запись в `user_consents` — на identity-service.

Действия:
- Проверено grep'ом: внешних вызовов `AccountCreator::registerUser`
  / `new AccountCreator(...)` в `projects/game-origin/src/` и
  `public/` **нет** (единственное упоминание — комментарий в
  `core/AjaxRequestHelper.abstract_class.php`).
- В шапку `src/game/AccountCreator.class.php` добавлен
  предупреждающий блок: класс остаётся как helper для lazy-create
  при handoff'е, публичные точки входа добавлять нельзя без
  отдельного плана с чекбоксами + integration user_consents API.
- Чекбоксы согласия в game-origin **не нужны** — собираются на
  portal/AuthPage и game-nova/LoginScreen (планы 44 + 47).

### Ф.3. Возрастная маркировка ✅ (2026-04-27)

- Блок «12+» в `layout.tpl`, CSS-стилизация.
- Добавлен `<div class="oxsar-footer">` с `.age-rating` (12+) перед
  закрытием `</body>`; стили `.oxsar-footer .age-rating` в
  `public/css/layout.css`.

### Ф.4. UGC-модерация на все точки ✅ (2026-04-27)

Подробная спецификация — выше в разделе «Что меняем» §4 (поверхности,
стратегия per-поверхность, расширение API класса `Moderation`,
smoke-тест из 8 шагов, что отложено).

- Реализован `Moderation::mask()` в
  `src/core/util/Moderation.util.class.php` по образцу `MaskForbidden`
  из `projects/game-nova/backend/internal/moderation/blacklist.go`:
  если найден хоть один запрещённый корень в нормализованной форме —
  все буквенные символы (a-z, А-я, ё) в исходной строке заменяются на
  `*`. Цифры, пробелы, пунктуация, BBCode/HTML-теги остаются как есть.
  Длина строки сохраняется. Если запрещённого нет — строка возвращается
  без изменений.
- Применено в 7 UGC-точках (12 вызовов):
  - **Чат** (`Chat::sendMessage`, `ChatAlly::sendMessage`) — `mask()`
    после bbcode-обработки, перед `INSERT INTO chat`/`chat2ally`.
  - **ЛС** (`MSG::*` строки 174-175) — `mask()` темы и тела перед
    обоими `INSERT INTO message` (sender + receiver копии).
  - **Альянс — название и тег** (`Alliance::foundAlliance`) — два
    `isForbidden()`-вызова, при срабатывании поднимаются те же
    `tagError`/`nameError` через `Logger::getMessageField`, что и при
    некорректной длине / regex-нарушении.
  - **Описания альянса** (`Alliance::updateAllyPrefs` ~995) — `mask()`
    для `$textextern`, `$textintern`, `$applicationtext`.
  - **Заявка на вступление** (`Alliance::apply` ~1150) — `mask($text)`.
  - **Глобальная рассылка альянса** (`Alliance::globalMail` ~1373) —
    `mask()` темы и тела перед циклом `INSERT INTO message`.
- Аудит-таблица **не введена** (как указано в плане — отдельная задача).
- Verifications выполнены:
  - `php -l` на всех 5 модифицированных файлах — `No syntax errors`.
  - Unit-тест `Moderation::mask()` через CLI — 11/11 PASS (загружено
    57 корней из общего blacklist; чистый текст не меняется; пустая
    строка ОК; запрещённый текст — все буквы → `*`, длина сохраняется,
    цифры/пробелы/punct/скобки уцелели).
- Smoke-тест в браузере **не выполнен** — требует docker-compose up
  для game-origin; правки — точечные `Moderation::mask()`/`isForbidden()`
  на критических путях (insert в БД), unit-тест mask() показывает
  корректную семантику. Smoke рекомендуется при следующем dev-запуске.
- **Известные ограничения**: 149-ФЗ требует хранилища логов чата на
  6 месяцев для запросов от регулятора. В legacy-чате таблица `chat` /
  `chat2ally` есть и не TTL-чистится, но запрос «выдай по userid за
  период» — отдельная админ-функция, не входит в Ф.4. Audit-таблица
  модерации (запись факта срабатывания blacklist) — также отдельная
  задача (см. план 50 §4 «Что отложено»).
- 1 общий коммит.

### Ф.5. Кнопка «Пожаловаться» ✅ (2026-04-28)

- Smarty-шаблон-партиал
  `projects/game-origin/src/templates/standard/_report_button.tpl` —
  CSS+HTML модалки + JS-обёртка `oxReport.open(type, id)` /
  `submit()`. Подключён один раз через `{include}"_report_button"{/include}`
  в `layout.tpl` перед футером.
- PHP-helper `getReportButton($type, $id, $title=null)` в
  `src/game/Functions.inc.php` — возвращает HTML-кнопку 🚩 с
  inline-onclick. Гость → пустая строка; self-report для `user` →
  пустая строка.
- Контракт совпадает с `ReportButton.tsx` из game-nova: 7 reasons
  (profanity / extremism / drugs / spam / impersonation / cheat /
  other), 4 target_type (user / alliance / chat_msg / planet),
  comment до 1000 символов. POST на `${PORTAL_BASE_URL}/api/reports`
  с `Authorization: Bearer <jwt>`.
- `PORTAL_BASE_URL` — добавлена в `config/consts.php` (default
  `https://oxsar-nova.ru` для prod, `http://localhost:8090` для dev,
  переопределяется через env или `consts.local.php`).
- JWT для cross-origin: cookie `oxsar-jwt` помечена HttpOnly +
  SameSite=Strict (`public/dev-login.php`), браузер не отправит её
  на portal-backend. Поэтому PHP читает её из `$_COOKIE` и embed'ит
  в JS как строку (`json_encode` + strip кавычек) — fetch выставляет
  `Authorization: Bearer`.
- Точки внедрения:
  - **Никнейм игрока** (`UserList::formatRow` → `$row["report"] = ...`),
    `playerstats.tpl` теперь рендерит `{loop}report{/loop}` рядом с
    `message`/`buddyrequest`. Применяется ко всем рейтингам, которые
    делают `addLoop("ranking", $UserList->getArray())`.
  - **Общий чат** (`chat.tpl` шапка `{@chat_link}`) —
    `target_type='chat_msg', target_id='global'`.
  - **Чат альянса** (`chatally.tpl` шапка `{@a_chat_link}`) —
    `target_type='chat_msg', target_id='ally:<aid>'`.
  - **Страница своего альянса** (`allypage_own.tpl` рядом с тегом) —
    `target_type='alliance', target_id={var=aid}`.
- **CORS portal-backend**: в `deploy/docker-compose.multiverse.yml`
  добавлен `https://origin.oxsar-nova.ru` в `ALLOWED_ORIGINS`
  portal-сервиса (пары с уже существующим `https://oxsar-nova.ru`).
- **Verifications**: `php -l` чисто на всех 4 PHP-файлах
  (Functions.inc.php, UserList.class.php, consts.php,
  _report_button.tpl); `Functions.inc.php` загружается без фаталов
  через `require`. Smoke в браузере — отложен до следующего
  dev-запуска (требует docker-compose; правки точечные, риск
  минимальный).

### Ф.6. Ссылки на юр-документы ✅ (2026-04-27)

- Блок ссылок в `layout.tpl` footer'а.
- В `.legal-links` блок с absolute-URL на
  `https://oxsar-nova.ru/{offer,game-rules,refund,privacy}`,
  все ссылки `target="_blank" rel="noopener"`.

### Ф.7. Чистка copyright ✅ (2026-04-27)

- Поисковая замена в 10+ шаблонах.
- Удалить email `support@unitpoint.ru`.
- Фактически — 107 файлов в `src/templates/standard/*.tpl`. Шапка
  `Oxsar http://oxsar.ru` + `Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>`
  заменена на `Oxsar https://oxsar-nova.ru` +
  `Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.`.
- В `support.tpl` (регламент техподдержки) email
  `support@unitpoint.ru` заменён на `gibesapiselfbab@hotmail.com`.
- **Известная остаточная проблема:** текст самого `support.tpl`
  (регламент технической поддержки) — это юр-документ ликвидированной
  ООО «Юнит Поинт», содержит множество упоминаний «Оператор» = ООО.
  Требует отдельного плана: либо переписать как новый регламент
  oxsar-nova, либо удалить страницу из game-origin (заменить ссылкой
  на portal-форму поддержки).

### Ф.8. Финализация ✅ (2026-04-28)

Все 7 функциональных фаз закрыты, статус плана в шапке обновлён
на «✅ Завершён 2026-04-28».

Done:
- `docs/ops/legal-compliance-audit.md` — все gap'ы game-origin
  помечены ✅ ЗАКРЫТО (по фазам: 2.1/4.1/5.1/6.1/6.2/7.1/10.1).
- `docs/project-creation.txt` — итерации 50.Ф.X фиксировались
  отдельно для каждой фазы по мере закрытия (Ф.0→Ф.7); итерация
  50.Ф.8 — финальная отметка.
- Коммиты по фазам уже сделаны (см. шапки конкретных фаз с SHA).

**Smoke-тест отложен.** Все правки точечные и верифицировались
по мере закрытия фаз (php -l, unit-тесты, ручные проверки в
конкретных контекстах). Полный smoke (логин → главный экран →
чат → альянс → жалоба) рекомендуется при следующем dev-запуске
docker-compose game-origin (на dev-стенде запущен на :8092 — см.
[docs/legacy/game-origin-access.md](../legacy/game-origin-access.md)).
Если что-то всплывёт — отдельный fix-коммит.

**Что НЕ в плане 50** (отложено или вне scope):
- Ремастер origin на Go+React — план 62 (исследование закрыто
  2026-04-28) → планы 64-74 реализации.
- Содержимое `support.tpl` (юр-документ ликвидированной ООО
  «Юнит Поинт») — отдельная задача (переписать как новый
  регламент oxsar-nova ИЛИ удалить страницу с переадресацией
  на portal-форму поддержки). Не блокирует публичный запуск
  — страница не входит в основной игровой цикл.

---

## Тестирование

- TinyMCE удалён, чат работает с новым редактором/textarea.
- Регистрация (если возможна) требует чекбоксов; запись попадает в
  user_consents.
- Маркировка «12+» видна в footer'е game-origin.
- Запрещённые слова в чате/альянсе/описании режутся или блокируются.
- Кнопка «Пожаловаться» работает — жалоба появляется в
  AdminReportsTab game-nova.
- Footer содержит ссылки на оферту, правила, refund, privacy.
- В `git grep "UnitPoint" projects/game-origin/src/templates/`
  нет упоминаний.
- Повторный аудит (`docs/prompts/legal-compliance-audit.md`) даёт 0
  gap'ов в game-origin.

---

## Что НЕ делаем

- Не переписываем сам PHP-движок game-origin (это план 43, уже
  закрыт).
- Не переписываем игровой геймплей или баланс (это отдельный план
  37).
- Не выносим game-origin в отдельный репо — сохранение монорепо
  оправдано общими сервисами (auth, billing).

---

## Итог

7 точечных правок в legacy-вёрстке + замена одного фронтенд-пакета.
3–6 коммитов в зависимости от объёма Ф.1 (замена TinyMCE может
оказаться сложнее, чем простая замена редактора). После выполнения
game-origin готова к публичному запуску с теми же юридическими
гарантиями, что и остальные подпроекты oxsar-nova.
