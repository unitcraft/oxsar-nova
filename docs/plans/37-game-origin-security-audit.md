# Security Audit game-origin (план 37.6)

**Дата**: 2026-04-27
**Аудитор**: Claude (statических анализ через grep + анализ кода)
**Scope**: `projects/game-origin-php/src/` + `templates/` + `public/`

## Резюме

| Категория | Статус | Комментарий |
|---|---|---|
| SQL-injection | ✅ OK | Найденные потенциалы (JwtAuth) защищены через `quote_db_value()`. Все остальные SQL — через `sqlVal()/sqlArray()`. |
| Direct file access | ✅ OK | nginx deny `*.class.php` + `*.inc.php`. `config.inc.php` отсутствует (env-based). 73 из 89 game/*.class.php имеют die-guard, остальные защищены nginx. |
| XSS | 🔴 CRITICAL | Template-engine `{@var}` НЕ escapes. ~50+ мест с user-controlled (username, planetname, message, alliance name, notes, и т.д.) выводят raw HTML. |
| CSRF | 🔴 CRITICAL | 91 POST-форма, **0** с токеном. **0** проверок CSRF / Referer / Origin в обработчиках. Любая внешняя страница может POST'нуть `/game.php/SaveNotes`, `/game.php/StockNew`, `/game.php/Mission` и т.д. |
| Other | 🟡 minor | `error_log()` без sanitize в нескольких местах (log injection low risk). |

## 37.6.1 — SQL-injection

### Метод

```bash
grep -rEn 'mysql_query\b' src/                  # deprecated API
grep -rEn '"(SELECT|INSERT|UPDATE|DELETE)[^"]*\$_(GET|POST|COOKIE|REQUEST)' src/
grep -rEn 'sqlQuery\([^)]*\$_(GET|POST|COOKIE|REQUEST)' src/
grep -rEn 'sqlQuery\(.*\.\s*\$[a-z]' src/game/ | grep -v sqlVal\|sqlUser\|intval
```

### Находки

**Прямых $_GET/POST в SQL — НЕТ.**

`mysql_query()` (deprecated API):
- `core/database/DB_MySQL.class.php:100` — альтернативный класс (не используется, default `DB_MySQL_PDO`).
- `core/cronjobs/database_check.php:9-52` — статичные `SHOW/CHECK/OPTIMIZE TABLE` без юзер-данных, безопасно.
- `core/legacy_payment/payment.inc.php:64-121` — отключено (37.3, платежи через portal).

Подозрительные SELECT с `$variable`:
- `core/JwtAuth.php:186, 213` — `WHERE global_user_id = $gid` без quotes.
  **OK**: `$gid = $db->quote_db_value(...)` ⇒ возвращает уже quoted+escaped.
- `core/User.class.php:162` — статический SQL без переменных.
- `core/Options.class.php:124` — статический SQL.

### Вывод

Legacy-код oxsar2 **дисциплинированно** использует `sqlVal()`, `sqlArray()`,
`sqlUser()`, `sqlPlanet()` для всех user-input в SQL. Это типичный
анти-инъекционный паттерн.

JwtAuth (наш новый код 37.5c) тоже escapes. **No action needed**.

---

## 37.6.2 — XSS 🔴 CRITICAL

### Метод

1. Прямой echo: `grep -rEn 'echo\s+[^;]*\$_(GET|POST|...)' src/` — **0 находок**.
2. Template engine: `Template::get()` line 559-563 — `return $this->templateVars[$var]`
   **без htmlspecialchars**. Шаблонизатор compile'ит `{@var}` в `<?php echo $this->get("var"); ?>`.

### Уязвимые поверхности

User-controlled поля выводятся через `{@var}` без escape во многих
шаблонах. Найдено grep'ом по часто-юзер-управляемым именам:

```
{@username}, {@planetname}, {@alliance_name}, {@tag}, {@message},
{@notes}, {@text}, {@email}, {@content}, {@title}, {@name},
{@signature}, {@description}
```

Примеры:
- `templates/standard/cms_page.tpl: {@content}` — CMS HTML, может содержать `<script>`.
- `templates/standard/allypage_own.tpl: <td>{@tag}</td>, <td>{@name}</td>` — alliance name/tag.
- `templates/standard/edit_unit.tpl: <input value="{@name}">` — XSS через value-quote-escape.
- `templates/standard/moderate_user.tpl: <input value="{@username}">` — то же.

### Атака

1. Юзер регистрируется с username `<script>alert(1)</script>` (legacy validation
   ограничивает символы, но проверить надо отдельно).
2. Любой кто видит этот username в рейтинге, чате, списке alliance —
   получает execute JS в своём браузере → cookie theft, XSRF, и т.д.

### Митигация

**Опции** (выбрать в 37.7):

1. **Глобально escape в `Template::get()`**: добавить `htmlspecialchars()`
   по умолчанию. Risk: где-то в legacy шаблонах специально кладётся
   pre-rendered HTML (например, `Link::get(...)` → `<a href>`). Эти места
   ломаются → нужен `{@var:raw}` или `{@@var}` синтаксис для opt-out.
2. **Escape на уровне `Core::getUser()->get()` для строковых полей**:
   точечно для `username`, `email`, и т.п. на этапе чтения из БД. Не
   ломает HTML-genератора (Link/Image), потому что там escape не нужен.
3. **Точечно**: пройтись по всем `{@var}` где var = user-string (из grep
   списка), заменить на `{@var:html}` (если есть такой helper) или
   `<?php echo htmlspecialchars($this->get("var")); ?>` напрямую.

**Рекомендация**: подход #2 — самый безопасный для legacy кода. На уровне
DAO добавить:
```php
$row['username'] = htmlspecialchars($row['username'], ENT_QUOTES, 'UTF-8');
```
для всех строковых user-полей.

---

## 37.6.3 — CSRF 🔴 CRITICAL

### Находки

```bash
grep -rEn '<form\s+method=["'"'"']post' templates/ | wc -l        # 91 формы
grep -rEn 'csrf|_token|nonce' templates/                          # 0 совпадений
grep -rEn 'csrf|verify.+token|http_referer' src/game/             # 0 совпадений
```

**Любая POST-форма уязвима к CSRF**. Атака:
1. Игрок залогинен (JWT cookie действителен).
2. Заходит на сайт злоумышленника.
3. Тот делает скрытую форму:
   ```html
   <form action="http://origin.oxsar-nova.ru/game.php/Mission" method="POST">
     <input name="action" value="attack">
     <input name="target" value="...">
   </form>
   <script>document.forms[0].submit()</script>
   ```
4. Браузер игрока отправляет запрос со своим JWT cookie → атака идёт
   от его имени.

### Уязвимые операции

Любая state-изменяющая POST-операция:
- `/game.php/Mission` — отправка флота (атаки!).
- `/game.php/StockNew`, `/StockBan`, `/StockLotRecall` — биржа.
- `/game.php/SaveNotes` — заметки.
- `/game.php/Friends` — друзья / banlist.
- `/game.php/Alliance/*` — alliance actions.
- `/game.php/Constructions`, `/Research`, `/Shipyard` — постройки.
- `/game.php/Logout` — discard сессии.

### Митигация

**Опции** (выбрать в 37.7):

1. **CSRF-токен в форме** — стандартный подход:
   - Генерировать токен в сессии при логине (random 32-bytes hex).
   - Добавить `<input type="hidden" name="_csrf" value="{config}csrf_token{/config}">`
     в каждую форму (или в общий include).
   - На каждый POST в `Page::__construct` или middleware проверять
     `$_POST['_csrf'] === $_SESSION['csrf_token']`.
2. **SameSite=Strict cookies** — современный подход, не требует
   правок в формах:
   - Установить `SameSite=Strict` на JWT-cookie. Браузер не отправит
     cookie на cross-site POST → CSRF блокируется на уровне браузера.
   - **Risk**: ломается deep-linking из portal (если portal на другом
     subdomain). Проверить.
3. **Origin/Referer header check** — слабее, но прост:
   - Добавить middleware который для всех POST проверяет
     `$_SERVER['HTTP_ORIGIN']` или `$_SERVER['HTTP_REFERER']`
     совпадает с `BASE_FULL_URL`.

**Рекомендация**: SameSite=Strict (1 строка кода в JwtAuth) + Origin check
(20 строк в Page) как defense-in-depth. Token-based — overkill для текущего
этапа, но добавить на 37.13 финальной чистке.

---

## 37.6.4 — Direct file access ✅ OK

### Защита уже на месте

1. **nginx** (`docker/nginx.conf:13`):
   ```nginx
   location ~* \.(class|inc)\.php$ {
       deny all;
   }
   ```
2. **Структура папок**: `nginx root = /var/www/public/`. `src/`, `config/`,
   `migrations/` вне document_root → недоступны через HTTP.
3. **die-guard в class-файлах**: 73 из 89 файлов в `game/*.class.php`
   имеют `if(!defined("APP_ROOT_DIR")) die("Hacking attempt detected.");`.
   Остальные 16 — защищены nginx + структурой папок.
4. **`config.inc.php` отсутствует** — мы используем env-based конфиг
   (`bd_connect_info.php` читает из ENV).

### Файлы без die-guard (защищены только nginx)

- Alliance, BuildingInfo, Moderator, AdvTechCalculator, ResTransferStats,
  MonitorPlanet, ChatPro, Profession, UnitInfo, Records (10+ файлов в
  `game/page/`).

**Risk**: если nginx config сломается / direct PHP запуск → уязвимы.
**Action**: добавить die-guard всем (5 минут sed). Для defense-in-depth.

---

## 37.6.5 — Прочие находки

### `error_log()` без sanitize

В нескольких местах `error_log($_GET['x'])` или подобное может позволить
log-injection (вставка \n в логи, затрудняет анализ). Не critical.

### PHP `display_errors`

В prod сейчас `display_errors=Off` (default php-fpm) — error pages
не выдают пути файлов. **OK**.

### Файлы с `@`-suppress

```bash
grep -rEn '@\$_' src/ | head
```
- `core/Functions.php` — несколько `@unlink`, `@file_get_contents`. Не
  блокеры.

---

## Итог 37.6

| Категория | Severity | Action |
|---|---|---|
| SQL-injection | ✅ OK | — |
| Direct access | ✅ OK | (опционально) добавить die-guard в 16 файлов |
| **XSS** | 🔴 CRITICAL | Этап 37.7.1 — escape в Template::get() или DAO-level |
| **CSRF** | 🔴 CRITICAL | Этап 37.7.2 — SameSite=Strict + Origin check |
| log-injection | 🟡 minor | 37.13 |

**Готовность к открытию игрокам**: 🔴 НЕТ. Нужны 37.7.1 (XSS) и 37.7.2 (CSRF) до того.
