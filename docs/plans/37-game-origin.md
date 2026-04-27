# План 37: game-origin — legacy-вселенная (клон oxsar2)

**Дата**: 2026-04-26  
**Статус**: Черновик (ADR-решения зафиксированы)
**Домен**: `projects/game-origin/`  
**Контекст**: plan-36 описывает портал с несколькими вселенными. game-origin — первая из
них: оригинальная игра oxsar2 без переделки геймплея и дизайна, встроенная в портал
как одна из вселенных. Индексная страница не нужна — точка входа идёт с портала.

## Подключение к legacy для сверки и копирования данных

Legacy oxsar2 запущена в Docker (`d:\Sources\oxsar2`). Использовать как
**эталон поведения и источник данных** при отладке game-origin.

### Доступ к UI legacy

- URL: http://localhost:8080/game.php/Main
- Логин: `test`, пароль: `quoYaMe1wHo4xaci` (универсальный)
- Логин делается через `/login.php` (не Yii-роут):

```bash
curl -s -c /tmp/legacy.cookies -X POST "http://localhost:8080/login.php" \
  -d "username=test&password=quoYaMe1wHo4xaci&login=OK" \
  -L -o /dev/null -w "HTTP %{http_code}\n"
curl -s -b /tmp/legacy.cookies "http://localhost:8080/game.php/Main"
```

### Доступ к БД legacy

```bash
# Контейнер MySQL 5.7
docker exec oxsar2-mysql-1 mysql -uroot -proot oxsar_db -e "<query>"

# Дамп таблицы (только структура)
docker exec oxsar2-mysql-1 mysqldump -uroot -proot oxsar_db <table> --no-data 2>/dev/null

# Дамп данных по условию (например, всё для test-юзера userid=1)
docker exec oxsar2-mysql-1 mysqldump -uroot -proot oxsar_db \
  --no-create-info --skip-extended-insert --skip-comments --hex-blob \
  --where="userid=1" \
  na_user na_planet na_research2user 2>/dev/null
```

**Важно**:
- `2>/dev/null` обязателен — иначе stderr (`Using a password...` warning)
  попадает в SQL-файл и ломает применение.
- Никогда не подключать game-origin к БД legacy для запуска — только для
  чтения/копирования данных в `migrations/` нашего проекта.
- Параллельные сессии могут изменять данные test-юзера; перед копированием —
  убедиться что никто не играет.

### Что брать из legacy

- **Структура таблиц**: `mysqldump --no-data` (для актуальной DM-схемы)
- **Справочники**: `na_phrases`, `na_construction`, `na_ship_datasheet`,
  `na_artefact_datasheet`, `na_requirements`, `na_languages`, `na_config`,
  `na_tutorial_states`, `na_achievement_datasheet`
- **Тестовые данные**: данные `test`-юзера (userid=1) с планетами/постройками/
  исследованиями для dev-окружения oxsar-nova

### Различия схемы game-origin vs legacy

**Принцип**: схема legacy — источник истины, мы её НЕ модифицируем
(не меняем типы, не добавляем DEFAULT, не делаем nullable). Единственное
расхождение — наша надстройка для JWT auth:

- `na_user.global_user_id VARCHAR(36) UNIQUE` — добавлено в нашу схему
  (см. `migrations/003_add_global_user_id.sql`).

При импорте дампа `na_user` из legacy: указывать колонки явно или
дамп с пропуском `global_user_id` (NULL по умолчанию).

Если код требует «удобных» дефолтов которых нет в legacy — добавлять
их в коде (INSERT с явными значениями), а НЕ через миграции `ALTER TABLE`.

### Сверка страниц game-origin vs legacy

Чтобы проверить что страница в game-origin рендерится так же как в legacy
(контент, баланс, формулы, права) — импортировать полный дамп legacy:

```bash
bash projects/game-origin/tools/import-legacy-dump.sh
```

Дамп `legacy_dump.sql` (~1.5 GB) **в .gitignore**, не коммитится.
Это **отладочный инструмент**, а не часть штатного запуска. Штатно БД
наполняется через миграции `001+002+003` (схема + справочники + JWT-колонка).

---

## Зафиксированные решения

| Вопрос | Решение |
|---|---|
| Порядок | Сначала PHP-клон «как есть» (37.1–37.5), потом решение про Go-порт |
| База данных | **MySQL 5.7** (как в legacy oxsar2); MySQL 8 ломает SQL без backticks (`system` стало reserved); при полном Go-порте — PostgreSQL |
| Референс для отладки | Legacy `d:\Sources\oxsar2` запускается в Docker (`oxsar2-mysql-1`, `oxsar2-php-1`, `oxsar2-nginx-1`, порт 8080) — сверять SQL/UI/поведение страниц |
| Конвертация схемы | Дамп 2022 — эталон схемы для сверки; конвертация MySQL→PostgreSQL только при полном Go-порте |
| Аутентификация | Сразу auth-service из plan-36 (JWT, единый аккаунт); `session_start()` / собственный логин убирается; PHP проверяет JWT через `firebase/php-jwt` + JWKS |
| Баги и дыры | Сначала запускаем, потом патчим; SQL-инъекции — критичны до открытия игрокам |
| Несколько вселенных | Возможно несколько изолированных инстансов с разными `consts*.php`; общего рейтинга между инстансами нет |
| UI/шаблоны | `.tpl` остаются; правки и багфиксы допустимы; при Go-порте — `pongo2` (Smarty-compatible) |
| Universe Switcher | Vanilla JS виджет (~100 строк) в `.tpl` шапке: список вселенных + баланс кредитов из auth-service |
| Домен | `origin.oxsar-nova.ru` |
| Чат | Legacy PHP-чат остаётся как есть |
| Платежи | Отключаем на старте; позже подключим систему платежей из game-nova |
| Начальные данные | Чистая БД + справочные данные из `new-for-dm/data.sql`; реальные игроки из дампа — позже, вне репо |
| Yii | Полностью выкидываем; игра работала без него через `game.php` + `game/` + `ext/` |
| EventHandler | PHP-крон работает на старте; Go-воркер пишем параллельно, переключаем после golden-тестов |
| Полный Go-порт | Решение после этапа 37.2 (ревизия читаемости PHP-кода); приоритет — Go EventHandler |

---

## Цели

1. Сохранить игру **один-в-один** (PHP, MySQL, `.tpl`-шаблоны, логика, баланс).
2. Убрать только то, что **гарантированно не используется**: Yii, соцсети-OAuth, платежи (временно), `index.php`.
3. Интегрировать auth-service (plan-36): JWT вместо собственного логина, Universe Switcher в шапке.
4. Заменить EventHandler на Go-воркер **без изменения логики** (параллельно с PHP-кроном).
5. Исправить **явные баги и дыры** после первого запуска.
6. Привести структуру файлов к чистому виду внутри `projects/game-origin/`.

---

## Стратегия копирования: «всё скопировать, потом убрать лишнее»

Причина: безопаснее удалять ненужное, чем угадывать нужное. Удаление проверяется
тестами и ручным прогоном; пропуск нужного файла ломает игру незаметно.

### Источники схемы и данных

Есть два дампа и отдельные SQL-файлы:

| Файл | Описание |
|---|---|
| `d:\Sources\backups-oxsar2\oxsar_db_2019-01-25_03h31m.Friday.sql.gz` | дамп 2019 — реальная игровая БД с данными |
| `d:\Sources\PX92-data\maria-backup\latest\oxsar_db_2022-03-06_04h01m.Sunday.sql.gz` | дамп 2022 — более свежая схема (приоритет) |
| `d:\Sources\oxsar2\sql\new-for-dm\schema.sql` | чистая схема для DM-вселенной |
| `d:\Sources\oxsar2\sql\new-for-dm\data.sql` | справочные данные (юниты, постройки) |
| `d:\Sources\oxsar2\sql\table_dump\` | отдельные таблицы (ачивки, артефакты, фразы) |

**Использование дампов:**
- Дамп 2022 — источник истины для схемы PostgreSQL-миграции
- Дамп 2019 + данные из `table_dump/` — наполнение тестовой БД для golden-тестов
- `schema.sql` / `data.sql` — перекрёстная проверка при конвертации типов

**Конвертация MySQL → PostgreSQL** выполняется на этапе 37.4 (до первого запуска):
- `AUTO_INCREMENT` → `SERIAL` / `BIGSERIAL`
- `TINYINT(1)` → `BOOLEAN`
- `INT UNSIGNED` → `BIGINT`
- `ENUM(...)` → `TEXT` с CHECK constraint
- `DATETIME` → `TIMESTAMPTZ`
- MySQL-специфичные функции в хранимках/триггерах → PostgreSQL-эквиваленты
- Сверка с `d:\Sources\oxsar2\sql\new-for-dm\schema.sql` как эталоном чистой схемы

---

### Этап A — слепое копирование

Скопировать в `projects/game-origin/` всё из `d:\Sources\oxsar2\www\`, включая:

| Источник | Назначение |
|---|---|
| `game/` | `src/game/` |
| `ext/` | `src/ext/` |
| `ext/templates/` | `templates/` |
| `new_game/protected/config/` | `config/` |
| `new_game/protected/` (остальное) | `src/yii-protected/` (изолировано) |
| `css/`, `js/`, `images/`, `fonts/` | `public/` |
| `game.php`, `common.inc.php`, `global.inc.php` | корень `src/` |
| `PEAR/` | `src/PEAR/` |
| `game/Assault.jar` | `tools/Assault.jar` (архив, замена — plan B) |

Что **не копировать** сразу:
- `new_game/` (Yii-фреймворк целиком) — копируем только `protected/`
- `cache/`, `assets/` runtime-директории
- `config.inc.php` с реальными паролями — только `config.inc.example.php`
- `bd_connect_info.php` — заменить на ENV-based конфиг
- Маркетинговые/служебные файлы: `flippa_*.txt`, `WMT-*.HTML`, `googlehostedservice.html`
- `index.php` — точка входа с портала, не нужна

### Этап B — ревизия и удаление

После копирования пройтись по каждой категории:

**Удаляем безопасно:**
- Yii-фреймворк (`yii/` директория) — заменяем его роутинг/autoload минимальным
  bootstrap-файлом
- OAuth-адаптеры соцсетей (`SocialAPI_*.php`, `network_*.php`, конфиги mailru/vk/ok)
  — вход только через портальный auth-service (plan-36)
- `PEAR/` — используется только для email; заменяем одной функцией через `mail()` или
  портальный notification-сервис
- `chat.php`, `chatPro.php` — чат будет отдельным сервисом
- `new_game/protected/commands/` — консольные команды Yii; крон-джобы переезжают на Go

**Оставляем нетронутым:**
- Весь `game/*.class.php` — логика игры
- Весь `ext/` — переопределения и расширения (обязательно! см. memory)
- Все `.tpl` шаблоны — UI не трогаем
- `new_game/protected/config/consts*.php` — игровые константы
- `new_game/protected/config/params.php`, `base.php` — настройки игры
- `new_game/protected/models/` — модели данных
- `new_game/protected/controllers/` — контроллеры страниц

### Этап C — безопасный рефактор структуры

Только переименования/перекладывания, **без изменения логики**:

- Все `require`/`include` пути обновляются под новую структуру `projects/game-origin/`
- `config.inc.php` → читает из ENV (DB_HOST, DB_USER, DB_PASS, DB_NAME) — единственное
  изменение поведения, всё остальное через переменные
- Autoload: минимальный PSR-0-compatible загрузчик вместо Yii autoload (~30 строк)
- Крон: `ext/cronjobs/` → запускаются через Go-воркер (см. §EventHandler ниже)

---

## Замена EventHandler → Go

`game/EventHandler.class.php` (2769 строк) + `ext/ExtEventHandler.class.php` (927 строк)
составляют ядро обработки событий. Это самое сложное место в legacy.

**Подход**: порт логики на Go с сохранением 100% семантики.

### Принципы порта

- PHP-код остаётся как **эталонная спецификация** — каждый case в EventHandler
  покрывается тестом, который сравнивает результат Go с результатом PHP (через
  `php -r` в тестах или golden-файлы).
- `ext/ExtEventHandler` переопределяет часть методов базового — порт учитывает оба
  файла; итоговое поведение = ext-версия там, где она есть.
- Состояние БД — та же MySQL (не меняем схему); Go-воркер использует `database/sql` +
  сгенерированные запросы (sqlc или ручные — по объёму).
- Новый воркер: `projects/game-origin/worker/` — отдельный бинарь, запускается рядом
  с PHP-приложением.

### Этапы

1. Составить полный список event-типов из PHP (grep по `case`, `switch`, `$event->type`).
2. Написать Go-структуры событий и интерфейс `EventProcessor`.
3. Портировать event по event, покрывая каждый golden-тестом.
4. Параллельный прогон: Go-воркер обрабатывает события, PHP-крон работает рядом —
   результаты сравниваются в staging до переключения.
5. После валидации — отключить PHP-крон, оставить только Go-воркер.

---

## Замена Assault.jar → Go

`game/Assault.jar` (и его PHP-обёртка `Assault.class.php`) — боевой движок.

**Подход**: реиспользовать уже написанный `projects/game-nova/backend/internal/battle/`
как основу, адаптировав под схему данных oxsar2 (MySQL таблицы assault/units).

- Исходник Java: `d:\Sources\oxsar2-java\assault\` — полный Java-проект (ant build).
  Это источник истины для портирования логики боя на Go; не нужно реверсить JAR.
- Интерфейс вызова — HTTP или Unix-socket; PHP вызывает Go-сервис вместо `java -jar`.
- Формулы — из Java-исходников `oxsar2-java/assault/src/` (приоритет над `Assault.class.php`).
- Тесты: golden-файлы генерируем запуском `Assault.jar` с тестовыми входами, Go-сервис сравниваем побайтово.

---

## Исправление явных багов и дыр

Перед запуском провести аудит по следующим категориям. Каждая найденная дыра → запись
в `docs/balance/audit.md` (общий реестр) → отдельный тикет/PR.

### Безопасность (обязательно до запуска)
- [ ] SQL-инъекции: прогнать `grep -r "mysql_query\|\"SELECT.*\$_" src/` — все конкаты
      переменных в SQL заменить на prepared statements или `mysqli_real_escape_string`
- [ ] XSS: все `echo $_GET/POST` без `htmlspecialchars` — зафиксировать и исправить
- [ ] CSRF: формы без токена — добавить минимальный токен-механизм
- [ ] Прямой доступ к `game/*.class.php` без авторизации — проверить middleware
- [ ] `config.inc.php` не должен быть доступен по HTTP — `.htaccess` / nginx rule

### Игровые дыры (геймплей не меняем, только читы закрываем)
- [ ] Отрицательные ресурсы при отмене постройки / атаке — проверить все refund-пути
- [ ] Race condition в очереди строительства (двойной клик → двойное списание)
- [ ] Флот «застрявший» после ошибки обработчика событий — механизм recovery
- [ ] Переполнение integer в формулах производства на высоких уровнях шахт
- [ ] Экспедиция: возможность отправить больше флотов чем лимит через параллельные запросы

### Качество (не критично, но дёшево исправить при копировании)
- [ ] PHP notices/warnings в error_log — убрать `@` заглушки, исправить причины
- [ ] `ext/cronjobs/database_check.php` и `ext/payment.inc.php` используют `mysql_*` (deprecated) — первый чиним, второй отключаем вместе с платежами
- [ ] Неиспользуемые файлы после удаления Yii/PEAR/соцсетей — финальная чистка

---

## PHP-версия и DB-слой

**PHP**: legacy уже работает на PHP 8.3 (проверено локально) — поднимать версию не нужно,
она уже современная.

**DB-слой**: legacy использует PDO, но через Yii-обёртку:
`DB_MYSQL_PDO` → `Yii::app()->db` → PDO MySQL.

При удалении Yii (этап 37.3) `includes/database/DB_MYSQL_PDO.class.php` нужно
переписать — заменить `Yii::app()->db` на прямой `new PDO("mysql:...")`. Это ~20 строк
и даёт чистый PDO без Yii. Остальной код (`game/`, `ext/`) не меняется — он работает
через абстракцию `Database`.

Два файла с устаревшими `mysql_*` функциями:
- `ext/payment.inc.php` — отключается вместе с платежами
- `ext/cronjobs/database_check.php` — заменяем на PDO при этапе 37.3

---

## Конфиги игры

`new_game/protected/config/` содержит критические игровые настройки:

| Файл | Содержит |
|---|---|
| `consts.php` | базовые игровые константы (скорость, лимиты, множители) |
| `consts.dm.local.php` | переопределения для вселенной DM (наш случай) |
| `consts.dominator.local.php` | переопределения для dominator-режима |
| `params.php` | параметры приложения |
| `base.php` | Yii application config (роутинг, компоненты) |
| `main.php` | production конфиг |

**Правило**: все `consts*.php` копируются без изменений. Выбор нужного файла
(`dm.local` vs `dominator.local`) — через ENV-переменную `GAME_UNIVERSE=dm`.

---

## Структура `projects/game-origin/`

Файлы перекомпонованы логически относительно оригинала. Расположение можно менять
если это делает структуру чище — при условии что `require`/`include` пути обновляются
и логика кода не меняется.

```
projects/game-origin/
├── public/                  # веб-корень (nginx document_root)
│   ├── game.php             # единая точка входа
│   ├── css/, js/, images/, fonts/
│   └── universe-switcher.js # vanilla JS виджет (~100 строк)
├── src/
│   ├── core/                # "Recipe" framework: Core, autoloader, TPL, http, DB-абстракция
│   │   ├── database/        # Database.abstract + DB_MYSQL_PDO (PDO без Yii)
│   │   ├── template/        # шаблонизатор .tpl
│   │   └── ...              # остальные includes/
│   ├── game/                # game/*.class.php — бизнес-логика
│   │   └── page/            # game/page/*.class.php — контроллеры страниц
│   ├── ext/                 # ext/*.class.php + ext/page/ — переопределения
│   │   └── cronjobs/        # ext/cronjobs/ — PHP-крон (работает до Go-воркера)
│   └── bootstrap.php        # autoload + config loader (замена Yii entry)
├── templates/               # ext/templates/standard/*.tpl
├── config/                  # new_game/protected/config/consts*.php, params.php
├── worker/                  # Go EventHandler
│   ├── cmd/worker/
│   └── internal/events/
├── battle-service/          # Go Assault (адаптер поверх game-nova/battle)
├── tools/
│   └── Assault.jar          # архив оригинала для сравнения в тестах
├── migrations/              # SQL миграции (MySQL, только для багфиксов схемы)
├── docker/
│   └── docker-compose.yml   # mysql + php-fpm + go-worker
└── docs/
    └── bugfix-log.md
```

**Принцип перекомпоновки**: перемещать можно свободно, если:
1. Все `require`/`include`/`define` пути обновлены
2. Логика файла не изменена ни на строку
3. Перемещение проверено запуском игры

---

## Опциональный путь: полный порт на Go

Если в процессе анализа PHP-кода окажется, что логика хорошо читаема и покрываема
тестами — рассматриваем полный порт всего бэкенда на Go, убирая PHP полностью.

### Когда это имеет смысл

Объём кода legacy (без Yii, PEAR, соцсетей, чата):

| Категория | Файлы | Строк |
|---|---|---|
| EventHandler + ext | 2 | ~3700 |
| Контроллеры страниц (`game/page/`, `ext/page/`) | ~45 | ~18000 |
| Бизнес-логика (`game/*.class.php`) | ~25 | ~18000 |
| Конфиги и константы | ~10 | ~2000 |
| **Итого** | **~82** | **~42000** |

42 000 строк PHP — это сопоставимо с тем, что уже написано в game-nova (~40k строк Go).
Порт реален за 2–3 месяца при систематическом подходе (файл-за-файлом с тестами).

### Условия для принятия решения

Решение «портировать всё» принимаем **после этапа 37.2** (ревизия кода), когда будет
понятно:

1. **Читаемость**: нет ли в PHP-коде магии, динамических include, eval, runtime-генерации
   классов — то, что делает порт непредсказуемым.
2. **Связность**: насколько контроллеры страниц зависят от Yii-специфики (если сильно —
   порт дороже, если слабо — порт чист).
3. **UI**: если `.tpl` шаблоны плотно связаны с PHP-слоем через `$this->assign()` —
   придётся решить, оставить шаблоны с Go-бэкендом через `html/template` или переписать
   UI отдельно.

### Стратегия полного порта (если решим делать)

**Принцип**: PHP-файл = спецификация. Go-файл = реализация + тесты.

- Каждый контроллер страницы (`game/page/X.class.php`) → Go handler в `internal/page/`
- Бизнес-логика (`game/Planet.class.php`, `game/NS.class.php` и др.) → Go service/domain
- `.tpl` шаблоны → либо `html/template` (Go) с минимальными правками, либо оставить
  PHP-шаблонизатор как отдельный слой рендеринга (hybrid: Go HTTP + PHP render)
- MySQL схема не меняется; Go использует `database/sql` + sqlc
- Тесты: golden-файлы для каждого экрана (HTML diff) + unit-тесты бизнес-логики

**Hybrid-вариант** (менее рискованный): Go обрабатывает HTTP-запросы и бизнес-логику,
`.tpl` рендерит Go-data через легковесный Smarty-compatible движок на Go
(`flosch/pongo2` или аналог). Тогда шаблоны не трогаем совсем.

**Полный порт UI** (более рискованный): `.tpl` → React/HTMX/plain HTML. Это уже
фактически game-nova с legacy-дизайном — обсуждать отдельно.

### Порядок порта по приоритету (если выберем полный Go)

1. `EventHandler` + `Assault` — уже запланированы в 37.9/37.10
2. `Planet`, `NS`, `Functions` — ядро данных, используется везде
3. `AccountCreator`, `GameLogin` — авторизация (заменяется auth-service из plan-36)
4. Контроллеры читающие данные: `Main`, `Empire`, `Galaxy`, `Research`, `Shipyard`
5. Контроллеры изменяющие данные: `Mission`, `Construction`, `Constructions`
6. Периферия: `Alliance`, `Exchange`, `Artefact`, `Market`, `Achievements`
7. Вспомогательные: `Chat`, `MSG`, `Ranking`, `Records`

---

## Чего НЕ делаем

- Не меняем баланс, формулы, тайминги — без ADR и согласования
- Не переделываем UI/шаблоны — `.tpl` остаются как есть (если не выбрали полный порт)
- Не добавляем новые фичи — только перенос + bugfix
- Не мигрируем на PostgreSQL — MySQL остаётся (отдельная БД от game-nova)
- Не меняем схему БД — только индексы/constraints если нужны для bugfix

---

## Связь с планом 36 (портал)

- Аутентификация: игра принимает JWT от auth-service (plan-36), свой логин убирается
- Переключение вселенных: кнопка в игровом меню → редирект на портал
- `index.php` не нужен — пользователь приходит уже авторизованным через портал

---

## Этапы выполнения

| # | Задача | Риск |
|---|---|---|
| 37.1 | Слепое копирование всех файлов в `projects/game-origin/` | низкий |
| 37.2 | Ревизия: составить список к удалению + оценить читаемость кода для полного Go-порта | низкий |
| 37.3 | Удаление Yii, PEAR, OAuth, соцсетей + минимальный bootstrap | средний |
| 37.4 | Перекомпоновка файлов в новую структуру + ENV-based конфиг + обновление путей include/require + замена DB_MYSQL_PDO на чистый PDO (без Yii) | средний |
| 37.5 | Docker-compose: mysql + php-fpm, первый запуск; наполнение схемой из `new-for-dm/schema.sql` + `data.sql` | средний |
| 37.5b | **Слияние ext/ → game/** (отступление от §198 «оставляем нетронутым»): вместо отдельного слоя ext-классы вмерживаются в базовые. Причина — упрощение архитектуры и забывчивость о том, что ext имеет приоритет; стало источником багов. См. подробности ниже. | средний |
| 37.5c | **Event-monitor воркер + стартовая планета**: PHP-аналог `NewEHMonitorCommand` запускает `EventHandler::goThroughEvents()` в цикле; при первом логине вставляется event `EVENT_COLONIZE_NEW_USER_PLANET` (как в legacy `BaseWebUser`). См. подробности ниже. | средний |
| 37.5d | **UI-багфиксы первого запуска**: «Пополнить кредиты» наезжает на ресурсы, прочие layout-проблемы. См. подробности ниже. | низкий |
| 37.6 | Аудит багов и дыр (SQL-инъекции, XSS, CSRF, race cond.) | высокий |
| 37.7 | Исправление критических уязвимостей (безопасность) | высокий |
| 37.8 | Исправление игровых дыр (не меняя баланс) | высокий |
| 37.9 | Порт EventHandler → Go + golden-тесты | очень высокий |
| 37.10 | Порт Assault → Go (адаптер battle-service) | очень высокий |
| 37.11 | Параллельный прогон PHP vs Go, валидация | высокий |
| 37.12 | Интеграция с auth-service (JWT вместо own session) | средний |
| 37.13 | Финальная чистка неиспользуемых файлов | низкий |

---

## 37.5b — Слияние ext/ в game/ (отступление от §198)

**Контекст**: §198 этого плана говорил «Оставляем нетронутым: весь `ext/`».
В legacy схеме `NS::factory()` сначала искал `ext/<path>/Ext<Class>.class.php`,
потом `game/<path>/<Class>.class.php`. Ext-версия имела приоритет — по сути,
решение работало через override-слой.

**Что пошло не так**: при работе над game-origin постоянно забывалось, что ext
имеет приоритет — правились методы базового класса в `game/`, а ext-версия
тихо перекрывала правки. Это ловушка, которой нет смысла оставлять для PHP-клона
(особенно учитывая, что цель этапа 37.9 — порт всего на Go, где никаких
override-слоёв не будет).

**Принятое решение** (отклонение от плана, согласовано):
- Все классы из `ext/` мерджатся в одноимённый базовый класс в `game/`.
- Удалены: `src/ext/` целиком (53 файла).
- `ext/templates/standard/*.tpl` → `src/templates/standard/` (один уровень).
- `NS::factory()` упрощён — теперь только `game/<path>/<Class>.class.php`.
- `AUTOLOAD_PATH_APP_EXT` удалён, autoload теперь только `game/,game/page/,game/models/`.

**Стратегии слияния** (определены по содержимому каждого Ext-файла):

| Ext-класс | Стратегия | Куда |
|---|---|---|
| `ExtMission` (1062 стр.) | Базовые stub'ы (`controlFleet`, `executeJump`, `starGateJump`) заменены реальными реализациями + добавлены новые методы (`starGateDefenseJump`, `holdingSelectCoords`, `holdingSendFleet`, etc.) | `game/page/Mission.class.php` |
| `ExtShipyard` (33 стр.) | Decorator: добавлен countdown-блок в начало `index()`, дальше идёт оригинал | `game/page/Shipyard.class.php` |
| `ExtMenu` (137 стр.) | Сохранены **обе** реализации `generateMenu`: базовая (десктоп) + новый `generateMenuMobile` (mobile skin). Диспатч в `generateMenu()` через `isMobileSkin()` | `game/Menu.class.php` |
| `ExtEventHandler` (927 стр.) | Заменены реализации в base: `queryExpiredEvents`, `removePlanetEvents`, `getFormationAllyFleets`, `startConstructionEventVIP`, `abortConstructionEvent`, `repair`, `disassemble`, `teleportPlanet`, `allianceAttack`, `rocketAttack`. В `fReturn` — Ext-логика добавлена в начало (приоритет — Ext, потом базовая логика). Новые методы (`disassembleOld`, `haltPosition`, `haltReturn`, `alien*`) добавлены в конец класса | `game/EventHandler.class.php` |
| `ExpedPlanetCreator` (149 стр.) | Самостоятельный класс (не override базового), используется явно в `Expedition.class.php`. **Оставлен как есть** — это extension в смысле "Expedition Planet Creator", не Ext-в-смысле-override | `game/ExpedPlanetCreator.class.php` |
| `ExtArtefactMarket`, `ExtArtefacts`, `ExtNotepad`, `ExtPayment`, `ExtRepair`, `ExtSimulator`, `ExtSupport`, `ExtTutorial`, `ExtUserAgreement`, `ExtWidgets` | Все наследуются от `Page` напрямую (не от существующего базового класса). **Просто переименованы** `ExtX → X` (файл и класс). Доп. правки: `NS::isFirstRun("ExtX::...")` → `"X::..."`. В `Payment.class.php` починен сломанный `paymentVkontakte()` (после грубого комментирования VK-platежей остался невалидный синтаксис — заменён на pass-through redirect, т.к. соцплатежи отключены планом 37). | `game/page/<X>.class.php` |
| `ExtAchievements` | **Конфликт имени**: в `game/` есть бизнес-сервис `Achievements`, и ExtAchievements (page) после переименования стал бы вторым `class Achievements`. Решение: переименовать бизнес-класс `Achievements` → `AchievementsService` (он чистый static service), 14 вызовов `Achievements::` обновлены в Page, NS, EventHandler, PlanetCreator, Assault, Functions, achievements.tpl. После этого `ExtAchievements` → `Achievements` (page) без конфликта | `game/AchievementsService.class.php` (бизнес) + `game/page/Achievements.class.php` (page) |

**Проверка после слияния**:
- `php -l` всех изменённых классов — ✅ без syntax errors.
- ext/ полностью удалён.
- Запуск в docker и smoke-тест Main + ключевых страниц.

**Памятка для будущего AI/разработчика**: записи о приоритете ext в memory
актуальны только для **legacy-кода** (`d:\Sources\oxsar2`), но НЕ для
game-origin — здесь ext-слой убран.

---

## 37.5c — Event-monitor воркер + стартовая планета

### Контекст

Сейчас новый юзер видит «колонисты ищут планету», но планета никогда не
появляется. Причина: в legacy планета создаётся не синхронно при регистрации,
а через **асинхронное событие** `EVENT_COLONIZE_NEW_USER_PLANET`, которое
обрабатывается отдельным воркером:

1. **Триггер** — `BaseWebUser::checkAndCreateHomePlanet()` (Yii-фильтр на
   каждый HTTP-запрос). Проверяет `na_user.hp IS NULL`, если нет уже
   запланированного события — вставляет `EVENT_COLONIZE_NEW_USER_PLANET` с
   `time = now() + COLONIZE_NEW_USER_PLANET_TIME` (3 сек по дефолту).
2. **Обработчик** — `cron-event-monitor.sh` через системный cron раз в
   минуту запускает `console.php NewEHMonitor` (Yii-команда). Команда
   живёт ~125 сек, в цикле вызывает `EventHandler::goThroughEvents(100)`
   каждые 50ms. При обработке `EVENT_COLONIZE_NEW_USER_PLANET` →
   `EventHandler::colonize()` → `new PlanetCreator($userid)` → планета в БД +
   `na_user.curplanet/hp` обновляются.

У нас:
- Yii выкинут → `BaseWebUser` нет → событие никто не вставляет.
- Воркера нет → даже если вставить событие вручную, оно не обработается.
- `NS::startEvents()` вызывается на каждом HTTP-запросе, но это **не главный
  обработчик** — это только обновление event-stack для текущего юзера, и оно
  rate-limited через memcache (раз в секунду на юзера). Memcache в
  контейнере **не установлен** (`Memcached` class not found), что отдельно
  ломает rate-limit.

### Подход

Сделать **два независимых компонента**, повторяющих legacy-механизм:

#### 1. Триггер `OnboardingService::ensureColonizationScheduled($userid)`

- Новый файл `projects/game-origin/src/game/OnboardingService.class.php`
- Один статический метод — копия логики `BaseWebUser:65-114` без Yii:
  - Если `na_user.hp IS NULL` И нет события `EVENT_COLONIZE_NEW_USER_PLANET`
    в `processed = WAIT` для этого юзера → `INSERT events`.
  - Защита от race condition: `INSERT ... ON DUPLICATE KEY` нельзя (нет
    уникального индекса), используем `SELECT FOR UPDATE` в транзакции
    или короткий advisory-lock через `GET_LOCK("colonize:$userid", 0)`
    (legacy использовал Yii cache.add — у нас аналог через MySQL lock).
- Вызывать из `Core::setUser()` (после успешной JWT-аутентификации, перед
  `setUser` финализацией).

#### 2. Воркер `projects/game-origin/worker/event-monitor.php`

**Стратегия (выбран Вариант 2 из 4 рассмотренных)**: PHP-скрипт живёт
ограниченное время (~125 сек) и сам выходит, docker-сервис с
`restart: unless-stopped` поднимает его заново.

Контекст выбора:
- В legacy этот скрипт периодически падал по непонятным причинам
  (memory leaks / stale connections / etc.) — поэтому короткий life-cycle
  и проверка `oxsar-monitor-manager.txt` для предотвращения дублирования.
- Docker даёт **более простую** реализацию того же подхода:
  - Не нужен системный cron внутри контейнера.
  - Не нужен token-файл — docker гарантирует один экземпляр сервиса
    (если не делать `--scale=2`).
  - Не нужна `ps grep` защита от параллельного запуска — гарантирует docker.
  - При краше PHP (`exit 1` или сегфолт) docker auto-restart через ~1-2 сек.
  - При нескольких крашах подряд docker уходит в backoff — видно в
    `docker compose ps`, тривиально диагностировать.

Рассмотренные альтернативы:
- **Внешний cron каждую минуту** (как в legacy) — нужен системный cron в
  контейнере, антипаттерн docker «один контейнер — один процесс».
- **`pcntl_fork` с auto-restart внутри одного процесса** — больше кода,
  ещё один failure mode (родительский watchdog тоже может упасть).
- **Supervisor внутри контейнера** — индустриальный стандарт, но
  избыточен когда сам docker умеет restart-policy.

Реализация:
- CLI-скрипт, копия `NewEHMonitorCommand::run()` без Yii:
  - Bootstrap через `bd_connect_info.php` + `global.inc.php` + `Core` + `NS`
    (как `public/game.php`, но с CLI-режимом — без HTTP-роутинга).
  - Цикл `while(time() - $start < 125)`: `EventHandler::goThroughEvents(100)`
    → sleep 50ms если 0 обработано.
  - **Безопасность каждой итерации**: `try/catch` вокруг
    `goThroughEvents()` — одно битое событие не убивает цикл, exception
    логируется в stdout, цикл продолжается.
  - **Защита от мёртвого MySQL connection**: перед каждой итерацией
    короткий `SELECT 1` — при разрыве выходим с `exit 1`, docker
    перезапустит со свежим коннектом (вместо «висящего» процесса).
  - Logging: stdout (docker compose logs подберёт).

В docker-compose.yml добавить сервис `event-monitor`:
```yaml
event-monitor:
  build: { context: .., dockerfile: docker/Dockerfile.php }
  command: php /var/www/worker/event-monitor.php
  restart: unless-stopped  # перезапускает после 125-сек выхода
  depends_on: [mysql]
  volumes: [ ../:/var/www ]
```

#### 3. Memcache

**Решение**: ставим **настоящий Memcached** (как в проде), не заглушку.

Контекст: класс `MemCacheHandler` уже устроен как graceful no-op — при
отсутствии расширения `Memcache` всё работает в degraded-режиме. Но:
- Использование (8 точек): rate-limit `EventHandler::startEvents` (раз в
  секунду на юзера), кеш кредитов в `UserList`, кеш счётчиков `AlienAI`,
  и др. Без кеша каждый HTTP-запрос делает лишний SQL — для прод-нагрузки
  неприемлемо.
- Лучше один раз поднять как в проде, чем потом ловить деградацию.

Реализация:
- Сервис `memcached:alpine` в `docker-compose.yml`.
- В `Dockerfile.php` добавить расширение `memcache` (legacy использует
  старое API `class Memcache`, не `Memcached` — см. `MemCacheHandler:24`):
  ```dockerfile
  RUN apk add --no-cache --virtual .build-deps $PHPIZE_DEPS zlib-dev \
      && pecl install memcache-8.2 \
      && docker-php-ext-enable memcache \
      && apk del .build-deps
  ```
- В `bd_connect_info.php` константы `MC_SERVER=memcached`, `MC_PORT=11211`.

### Приёмочный тест

1. `dev-login.php` → JWT cookie → первый GET `/?go=Main`.
2. В БД сразу появляется event `EVENT_COLONIZE_NEW_USER_PLANET` для юзера.
3. Через ~5 сек воркер его обрабатывает — в `na_planet` появляется запись,
   `na_user.curplanet` и `hp` заполнены.
4. Повторный GET `/?go=Main` показывает страницу планеты с ресурсами,
   а не «колонисты ищут».
5. Логи воркера в `docker compose logs event-monitor` — без fatal/warnings.

### Риски

- **EventHandler после слияния не тестировался end-to-end** — это первый
  реальный прогон обработки событий. Возможны баги от слияния, которые
  всплывут только сейчас. Готов чинить по мере выявления.
- **Race condition при первом логине из нескольких вкладок** одновременно —
  triггер должен быть идемпотентен. Отсюда необходимость advisory-lock.
- **Воркер уходит в бесконечный цикл при ошибке** — если `goThroughEvents`
  бросает exception, не катит. Обернуть в `try/catch + exit 1` — restart
  поднимет заново.

---

## 37.5d — UI-багфиксы через сравнение со снимком legacy

### Концепция

Юзер обнаружил баг — «Пополнить кредиты» наезжает на строку ресурсов.
Без эталона неясно, это наша регрессия или так в legacy. Решение:
взять **снимок реального test-юзера** из боевой legacy БД, импортировать
к себе как fixture, рендерить **те же** страницы с тем же `userid` →
сравнивать с legacy.

Альтернативы (отклонены):
- **Полный дамп legacy** (~1.5GB) — избыточно для UI-сравнения, тяжело
  коммитить, протухает.
- **Поднимать второй legacy-инстанс** с нашим seed — нужно отключать
  event-loop у legacy (cron + `NS::startEvents` + `oxsar-monitor-manager.txt`),
  схема тоже синхронизировать. 80% работы — подгонка, 20% — само
  сравнение.
- **Side-by-side без синхронизации БД** — ловит только структурные баги
  (#1 в типологии ниже), не контентные.

### Типология UI-багов

| Тип | Зависит от данных? | Где видно |
|---|---|---|
| Структурные (DOM/CSS): наезд, сломанный grid, отсутствующий контейнер | Нет | Сравнение даже без идентичных данных |
| Контентные: неправильное число, пустой блок при 0 ресурсов | Да | Только при идентичных данных |
| Шаблонные: `MENU_X` вместо «Главная», не подставлена переменная | Слабо | Видно даже на пустом юзере |

«Пополнить кредиты» наезжает — структурный (#1). Но раз уж готовим
fixture — заодно поймаем контентные.

### Подход: snapshot test-юзера → fixture → сравнение

#### 1. Снять снимок test-юзера из legacy

Legacy БД доступна в контейнере `oxsar2-mysql-1`. Test-юзер: `userid=1`.

```bash
# Список таблиц где у юзера есть данные (предварительно — найти все
# таблицы с колонкой userid или связанной FK):
docker exec oxsar2-mysql-1 mysql -uroot -proot oxsar_db -N -B -e "
  SELECT TABLE_NAME FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA='oxsar_db' AND COLUMN_NAME='userid'
" 2>/dev/null

# Также таблицы по planetid (FK к user через na_planet):
docker exec oxsar2-mysql-1 mysql -uroot -proot oxsar_db -N -B -e "
  SELECT TABLE_NAME FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA='oxsar_db' AND COLUMN_NAME='planetid'
" 2>/dev/null
```

Снять данные через `mysqldump --where`:
```bash
mkdir -p projects/game-origin/migrations/fixtures
docker exec oxsar2-mysql-1 mysqldump -uroot -proot oxsar_db \
  --no-create-info --skip-extended-insert --skip-comments --hex-blob \
  --where="userid=1" \
  na_user na_planet na_research2user na_unit2shipyard \
  na_building2planet na_user2group na_buddylist na_message \
  na_alliance na_user2alliance na_artefact2user \
  2>/dev/null > projects/game-origin/migrations/fixtures/test-user-snapshot.sql

# Дополнительные таблицы где юзер может быть как destuser/sender:
docker exec oxsar2-mysql-1 mysqldump -uroot -proot oxsar_db \
  --no-create-info --skip-extended-insert --skip-comments --hex-blob \
  --where="planetid IN (SELECT planetid FROM na_planet WHERE userid=1)" \
  na_galaxy \
  2>/dev/null >> projects/game-origin/migrations/fixtures/test-user-snapshot.sql
```

**Важно**:
- `2>/dev/null` обязателен — иначе stderr-warning про пароль попадает в SQL-файл и ломает применение.
- Не забыть про `events` (текущие события юзера) — без них некоторые страницы (Mission, Construction) показывают неправильное состояние.
- Размер ожидается ~100-500KB. Коммитим в репо как dev-fixture.

#### 2. Применить fixture к нашей БД

Скрипт `projects/game-origin/tools/apply-test-user-fixture.sh`:
```bash
#!/bin/bash
# Заменяет dev-юзера на снимок test-юзера из legacy.
# Сохраняет global_user_id (наша надстройка для JWT).

GUID=$(docker exec docker-mysql-1 mysql -uroot -proot_pass oxsar_db -N -B \
  -e "SELECT global_user_id FROM na_user WHERE userid=1" 2>/dev/null)

# Очистить нашего dev-юзера и связанные таблицы
docker exec docker-mysql-1 mysql -uroot -proot_pass oxsar_db <<SQL
DELETE FROM na_user WHERE userid=1;
DELETE FROM na_planet WHERE userid=1;
-- ... (все связанные таблицы)
SQL

# Применить snapshot
docker exec -i docker-mysql-1 mysql -uroot -proot_pass oxsar_db \
  < projects/game-origin/migrations/fixtures/test-user-snapshot.sql

# Восстановить global_user_id для JWT lazy-join
docker exec docker-mysql-1 mysql -uroot -proot_pass oxsar_db \
  -e "UPDATE na_user SET global_user_id='${GUID}' WHERE userid=1"
```

#### 3. Залогиниться как тот же юзер

Наш `dev-login.php` ставит JWT с `sub: dev-user-001` → JwtAuth lazy-join
находит юзера по `global_user_id`. После применения fixture **userid=1**
получает наш `global_user_id` → залогинится как `test`.

#### 4. Сравнение всех страниц

Список страниц для сравнения (из `src/game/page/`):

**Основные** (приоритет 1, проверять обязательно):
Main, Resource, Constructions, Research, Shipyard, Defense, Mission,
Galaxy, Empire, Stock, ExchangeOpts, Repair, Disassemble.

**Коммуникации** (приоритет 2):
Chat, MSG, Notepad, Alliance, Friends, Search.

**Геймплей** (приоритет 2):
Artefacts, ArtefactMarket, Market, Achievements, Profession, Tutorial.

**Статистика и инфо** (приоритет 3):
Ranking, Records, Battlestats, ResTransferStats, Techtree,
AdvTechCalculator, Simulator, BuildingInfo, UnitInfo, ArtefactInfo.

**Прочее** (приоритет 3):
Preferences, UserAgreement, Support, Widgets, Referral, Changelog.

**Намеренно НЕ сравниваем** (выкинуты в game-origin):
Payment (соцплатежи отключены), Logout (через auth-service), Officer,
Moderator, RocketAttack, FleetAjax, MonitorPlanet, ChatPro, ChatAlly,
ArtefactMarketOld, EditConstruction, EditUnit, ExchangeOpts, Stock,
StockNew, TestAlienAI, ResTransferStats, Page (базовый), Construction
(базовый абстрактный).

#### 5. Diff-tool

Скрипт `projects/game-origin/tools/compare-with-legacy.sh`:
```bash
#!/bin/bash
# Для каждой страницы из списка:
# 1. curl нашу версию (port 8092) → /tmp/ours/$page.html
# 2. curl legacy версию (port 8080) → /tmp/legacy/$page.html
# 3. Нормализовать (убрать таймстампы, ID, value=, randomized id-event-N)
# 4. diff → отчёт что отличается

PAGES=(Main Resource Constructions Research Shipyard Defense Mission \
       Galaxy Empire ...)

# Логин в legacy (cookie-based)
curl -s -c /tmp/legacy.cookies -X POST \
  "http://localhost:8080/login.php" \
  -d "username=test&password=quoYaMe1wHo4xaci&login=OK" -L -o /dev/null

# Логин у нас
curl -s -c /tmp/ours.cookies "http://localhost:8092/dev-login.php" -o /dev/null

normalize() {
  sed -e 's/[0-9]\{2\}\.[0-9]\{2\}\.[0-9]\{4\} [0-9:]\+//g' \
      -e 's/id="[a-z_]*[0-9]\{4,\}"//g' \
      -e 's/value="[0-9]\+"//g' \
      -e 's/sid=[a-zA-Z0-9]\+//g' \
      -e 's/[0-9]\+\.\?[0-9]*//g' \
      "$1"
}

for page in "${PAGES[@]}"; do
  curl -sb /tmp/ours.cookies "http://localhost:8092/?go=$page" \
       > /tmp/ours/$page.html
  curl -sb /tmp/legacy.cookies "http://localhost:8080/game.php/$page" \
       > /tmp/legacy/$page.html

  diff <(normalize /tmp/ours/$page.html) <(normalize /tmp/legacy/$page.html) \
       > /tmp/diff/$page.diff

  size=$(wc -l < /tmp/diff/$page.diff)
  echo "$page: $size diff lines"
done
```

#### 6. Триаж и фикс

По каждой странице с ненулевым diff:
1. Открыть оба HTML в браузере (или просто диффы).
2. Классифицировать различия:
   - **Тип A** (структурный): чинить в шаблоне или CSS.
   - **Тип B** (контентный из-за разной БД): не баг, игнорировать.
   - **Тип C** (шаблонный, не подставлено что-то): чинить.
3. Каждый фикс — отдельный коммит, conventional commit
   `fix(game-origin/ui): <страница> — <короткий описание>`.

### Замораживание состояния на время сравнения

Чтобы diff на повторных запусках был стабильным:
- **У нас**: `docker compose stop event-monitor` перед началом.
- **В legacy**: создать stop-token: `docker exec oxsar2-php-1 sh -c
  "echo stop > /var/www/oxsar-monitor-manager.txt"` — бегущий
  `NewEHMonitor` процесс умрёт при следующем чек-цикле, новый не
  стартует пока файл существует.
- **NS::startEvents на HTTP-запросах**: добавить временный
  `define('SKIP_EVENT_LOOP', true)` в `bd_connect_info.php` обоих
  окружений (или просто игнорировать — за 30 сек curl-ов состояние
  обычно не успевает измениться).

### Что НЕ делаем в 37.5d

- Не редизайним. UI остаётся legacy-style (`.tpl` шаблоны).
- Не меняем балансные числа на странице (это 37.6+).
- Не добавляем новые UI-фичи — только фиксы существующих.
- Не пытаемся добиться pixel-perfect diff — рассинхрон по таймстампам и
  random-id неизбежен.

### Этапы 37.5d

| # | Подзадача | Артефакт |
|---|---|---|
| 37.5d.1 | Скрипт snapshot test-юзера из legacy → fixture SQL | `tools/snapshot-legacy-user.sh` + `migrations/fixtures/test-user-snapshot.sql` |
| 37.5d.2 | Скрипт apply fixture к нашей БД (с сохранением global_user_id) | `tools/apply-test-user-fixture.sh` |
| 37.5d.3 | Compare-tool: curl всех страниц у нас и в legacy + нормализованный diff | `tools/compare-with-legacy.sh` |
| 37.5d.4 | Триаж diff-отчёта: список страниц с реальными расхождениями | Запись в dev-log |
| 37.5d.5+ | Поштучные фиксы по каждой проблемной странице | По коммиту на каждый фикс |
