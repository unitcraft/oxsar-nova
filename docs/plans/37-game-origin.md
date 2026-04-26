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
