# Доступ к запущенному game-origin

Clean-room PHP-порт legacy oxsar2 (план 37, 43) запущен в Docker
рядом с oxsar-nova. Это **не** legacy oxsar2 — это новый PHP-проект
в репозитории oxsar-nova (`projects/game-origin/`). Использовать
для сверки UI/механик при работе по плану 62 и далее.

## Параметры доступа

- **URL**: http://localhost:8092/
- **Dev-логин**: http://localhost:8092/dev-login.php — мгновенно
  ставит JWT-cookie и редиректит на `?go=Main`. Привязан к
  `na_user.userid=1` (`username=test`, 9 планет, 4 луны, ~36M очков,
  роль admin).
- **Стек**: PHP 8.3 + nginx + MySQL 5.7 + memcached. Docker Compose
  в `projects/game-origin/docker/docker-compose.yml`.
- **JWT**: `alg=none` в dev-режиме (когда `IDENTITY_JWKS_URL` пустой).
  Куки `oxsar-jwt`, SameSite=Strict, httponly.

## Доступ через curl (для анализа без браузера)

```bash
# 1. Логин (сохраняем JWT-cookie)
curl -s -c /tmp/origin-jar.txt -L "http://localhost:8092/dev-login.php" \
  -o /dev/null -w "HTTP %{http_code}\n"

# 2. Любой экран через ?go=Page (PATH_INFO в game-origin тоже работает,
#    но ?go= — основной)
curl -s -b /tmp/origin-jar.txt "http://localhost:8092/game.php?go=Main" \
  -o /tmp/main.html -w "%{http_code} %{size_download}\n"
```

### Проверенные экраны (на 2026-04-28, под dev-user)

| Page | Размер | Замечание |
|---|---|---|
| Main | 21 390 | главный |
| Empire | 63 751 | обзор империи |
| Constructions | 40 020 | здания |
| Shipyard | 43 480 | верфь |
| Research | 43 861 | исследования |
| Galaxy | 27 074 | галактика |
| Mission | 18 919 | миссии флота |
| Alliance | 18 298 | альянс |
| MSG | 18 368 | сообщения |
| Simulator | 80 786 | симулятор боя |
| Repair | 20 061 | ремонт |
| Buddy | 21 923 | друзья |
| Artefact | 21 964 | артефакты |
| Exchange | 23 210 | биржа |

Страницы, возвращающие размер ~21 618 (Settings/Statistics/Achievement/
Tournament), вероятно отдают маску Main page — нужен другой
параметр или они скрыты. Проверять отдельно при необходимости.

## База данных game-origin

| Параметр | Значение |
|---|---|
| Host | `localhost` (снаружи Docker) или `mysql` (внутри сети) |
| Port | **3307** (снаружи) / 3306 (внутри сети Docker) |
| Database | `oxsar_db` |
| User | `oxsar_user` |
| Password | `oxsar_pass` |
| Root password | `root_pass` |
| Container | `docker-mysql-1` |
| Prefix | `na_` |

### SQL через docker exec

```bash
docker exec docker-mysql-1 mysql -uoxsar_user -poxsar_pass oxsar_db \
  -e "SELECT userid, username, points FROM na_user ORDER BY points DESC LIMIT 5;"
```

### Состояние БД (на 2026-04-28)

- 200+ таблиц (`SHOW TABLES`).
- 66 пользователей в `na_user` (legacy-сид перенесён через
  `legacy_dump.sql`).
- Балансовые формулы — в `na_construction` (поля `prod_*`, `cons_*`,
  `charge_*` как DSL-строки).
- Пароли — в **отдельной** таблице `na_password` (поля `password`
  MD5 + `password_sha1`). Для прохода через `/login` (не dev-login)
  потребуется знать пароль конкретного юзера или сбросить хэш.

## Связь с legacy oxsar2

- **legacy oxsar2** (отдельный проект `d:\Sources\oxsar2\`) запущен
  на http://localhost:8080 — описан в [game-reference.md](game-reference.md).
- **game-origin** = clean-room PHP-порт legacy в репозитории oxsar-nova
  (план 43). Источник данных и формул — копия `legacy_dump.sql`,
  затем правки/миграции в `projects/game-origin/migrations/`.

При сверке для плана 62:
- **game-origin** — основной источник «как это работает в legacy
  при текущих миграциях/правках clean-room» (`localhost:8092`).
- **legacy oxsar2** + `www/ext/` — источник для сверки оригинального
  PHP-кода (если в game-origin что-то могло уйти при clean-room).

## Что НЕ работает (известные ограничения dev-режима)

- Реальный логин через `/login` редиректит на `/login?error=NO_ACCESS`,
  потому что Identity-сервис не подключён в этой инсталляции —
  используй `/dev-login.php`.
- `?as_uid=N` параметр в dev-login пока не поддержан — заходим
  всегда как `test` (userid=1). Чтобы зайти под другим — менять
  `payload['sub']` в `public/dev-login.php`.
