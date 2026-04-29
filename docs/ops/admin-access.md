# Доступ к admin-консоли

Как получить доступ к `admin.oxsar-nova.ru` и что делать в типовых
ситуациях. Архитектурный обзор — в
[architecture/admin-frontend.md](../architecture/admin-frontend.md).

---

## Bootstrap: первый superadmin

Admin-консоль защищена permission-чек'ами: чтобы кто-то мог управлять
ролями (`roles:grant` / `roles:revoke`), у него уже должна быть роль
`superadmin`. Замкнутый круг разрывается вручную через SQL в БД identity
(план 53 §Bootstrap, CLI `identity-cli grant-role` пока не реализован —
sub-план 53e-identity-cli).

**До выполнения этого шага admin-консоль недоступна для всех** —
никто не имеет permission `roles:grant`, чтобы выдать роль кому ещё.
Это by design.

### Процедура

1. Юзер регистрируется через публичный портал (или `POST /auth/register`)
   и получает дефолтную роль `player`.
2. Достаём его `id` из `users` и выдаём роль `superadmin`:

   ```bash
   # На сервере, где запущен identity-db (для dev — deploy-identity-db-1):
   docker exec -i deploy-identity-db-1 psql -U identitysvc -d identitysvc <<'SQL'
   INSERT INTO user_roles (user_id, role_id, granted_at)
   SELECT u.id, r.id, now()
   FROM users u, roles r
   WHERE u.username = '<USERNAME>' AND r.name = 'superadmin'
   ON CONFLICT (user_id, role_id) DO NOTHING;
   SQL
   ```

   На проде имя контейнера БД и креды отличаются — см. `docker-compose.prod.yml`.

3. Юзер должен **перелогиниться** — его старый JWT содержит только `player`,
   новый возьмёт актуальные роли из `user_roles`.

После этого superadmin может выдавать остальные роли (`admin`, `support`,
`moderator`, `billing_admin`) штатно через UI админки (`/users/<uuid>` →
trash/grant icon) — этот же SQL не нужен для последующих ролей, только
для первого разрыва круга.

Для **локального dev** уже готов набор тестовых юзеров со всеми ролями —
см. [dev-test-users.md](dev-test-users.md). Запускать SQL вручную там не
нужно, юзеры созданы.

## IP-allowlist

Admin-консоль защищена nginx-уровневым IP-whitelist'ом. Без
прописанного IP юзер получит 403 даже на `/login`.

Файл whitelist'а на сервере: `/etc/nginx/admin-ips.conf` (монтируется
через docker-compose volume).

Формат:
```
1.2.3.4         1;     # моя VPN-нода
10.0.0.0/24     1;     # corporate range
```

После изменения файла:
```bash
docker compose -f deploy/docker-compose.admin.yml restart admin-frontend
```

## Login flow

1. Открой `https://admin.oxsar-nova.ru` (если IP в allowlist).
2. Форма логина: username + password (те же что в основной игре).
3. После 2FA (план 53 Ф.8 — WebAuthn / TOTP, отложено) →
   попадаешь в Dashboard.
4. Сессия живёт 30 минут idle (sliding TTL). При activity —
   продлевается.

## Что делать если

### «403 ip not allowed»

Твой IP не в whitelist'е. Проверь свой публичный IP
(`curl ifconfig.me`) и обратись к superadmin'у для добавления.

### «401 session expired»

Прошло > 30 мин без активности. Просто логинься заново.

### «refresh_invalid» при работе

Identity-сервис отозвал refresh-token (например, ты залогинился
на другом устройстве и `users.session_jti` обновился). Логинься
заново.

### Logout не работает

Logout best-effort: даже при ошибке identity локальная сессия
будет почищена. Если упорно не выходит — очисти cookies вручную в
браузере.

## Отзыв доступа у админа

1. Через UI (если у тебя есть `roles:revoke`):
   - Перейти в `/users/<uuid>`.
   - Найти роли `admin` / `superadmin` → нажать trash icon.
   - Указать reason (попадёт в audit log).
2. Через CLI на сервере (если UI недоступен, например compromised
   admin'а):
   ```bash
   identity-cli revoke-role admin <user-uuid> --reason "..."
   ```

После revoke максимум через 15 мин (срок действия access-токена)
у юзера перестанут работать его permissions. Для немедленного
отзыва — JTI-blacklist (план 53 Ф.8, отложен) или shutdown сессии
через `Del admin:sess:*` в Redis (это инвалидирует cookie).

## Audit log

Все role grants/revokes неизменно записаны в
`audit_role_changes` (identity-БД). Просмотр:
- В UI: `/audit` страница с фильтрами actor/target/action/since/
  until. Permission `audit:read`.
- В БД напрямую (для расследования инцидентов):
  ```sql
  SELECT created_at, action, role_name, actor_id, target_id,
         reason, ip_address
  FROM audit_role_changes
  WHERE target_id = '<uuid>'
  ORDER BY created_at DESC;
  ```

## Production deploy

Архитектура: docker-compose, nginx + admin-frontend SPA + admin-bff
Go-binary + Redis. См.
[deploy/docker-compose.admin.yml](../../deploy/docker-compose.admin.yml).

Запуск:
```bash
# Установить SESSION_SECRET (32+ байт hex):
export SESSION_SECRET=$(openssl rand -hex 32)
# Или в .env.

docker compose \
  -f deploy/docker-compose.yml \
  -f deploy/docker-compose.prod.yml \
  -f deploy/docker-compose.admin.yml \
  up -d --build
```

Healthchecks:
- admin-bff: `wget http://admin-bff:9200/healthz`.
- admin-frontend: `wget http://admin-frontend:80/healthz`.

## Локальный dev-доступ

Админка не имеет ссылок из публичного портала (by design — отдельный
поддомен в проде, IP-allowlist, отдельная сессия). Доступ — только по
прямому URL.

**URL:** http://localhost:8086

Запуск поверх dev-стека:

```bash
SESSION_SECRET=$(openssl rand -hex 32) \
BFF_COOKIE_SECURE=false \
IDENTITY_URL=http://identity-service:9000 \
docker compose \
  -f deploy/docker-compose.yml \
  -f deploy/docker-compose.admin.yml \
  -p deploy up -d --build admin-bff admin-frontend
```

- `BFF_COOKIE_SECURE=false` обязательно — без него cookie с `Secure`-флагом
  не доедет по `http://localhost`, юзер не залогинится.
- `IDENTITY_URL` явно — в compose дефолт указывает на `:9000`, переопределение
  не требуется, но явное значение страхует от расхождения если в env что-то
  залипло.

Допуск по ролям (login form принимает username или email, см.
[dev-test-users.md](dev-test-users.md), пароль у всех `DevPass123`):

| Юзер | Роль | Что увидит |
|------|------|-----------|
| `super1` | superadmin | всё — управление ролями + игровая админка + биллинг + модерация + саппорт |
| `admin1` | admin | игровая админка (planet ops, fleet recall, грант ресурсов, баны) |
| `billing1` | billing_admin | биллинг (отчёты, возвраты, audit) |
| `mod1` | moderator | модерация UGC, баны |
| `support1` | support | чтение тикетов / аудита / профилей |

IP-allowlist в dev открыт настежь — [deploy/admin-ips.conf](../../deploy/admin-ips.conf)
содержит `0.0.0.0/0 1;`. Файл в `.gitignore`, в репо лежит только `.example`.

## Закладка в браузере

Чтобы не светить URL админки в публичном UI и не запоминать его руками,
админ кладёт ссылку себе в bookmark bar. Это работает за всех условных
«Перейти в админку» в шапке сразу: один раз, один клик, без любого
вмешательства в код портала.

**Prod**

| Поле | Значение |
|------|----------|
| Name | `oxsar admin` |
| URL  | `https://admin.oxsar-nova.ru` |

**Dev**

| Поле | Значение |
|------|----------|
| Name | `oxsar admin (dev)` |
| URL  | `http://localhost:8086` |

Если хочется попасть сразу на конкретную страницу — добавляется путь:
`https://admin.oxsar-nova.ru/users` и т.п. Полезно админам с одной
ролью: `support` → `/tickets`, `billing_admin` → `/billing/refunds`,
`moderator` → `/ugc`.

Альтернативные браузерные варианты:
- **Chrome / Edge / Brave** — `Ctrl+D` на открытой странице админки →
  ввести имя → выбрать `Bookmarks bar` как папку.
- **Firefox** — то же `Ctrl+D`.
- **Safari** — `Cmd+D` → `Add this Page to → Favorites`.

Это **не** добавляется автоматически: каждый админ делает это руками
один раз. Намеренно — bookmark live в личном профиле браузера, не в
shared-конфиге, и не утекает с компрометацией одного устройства.

## Развалилось — что смотреть в логах

```bash
docker compose -f deploy/docker-compose.admin.yml logs admin-bff --tail=100
docker compose -f deploy/docker-compose.admin.yml logs admin-frontend --tail=100
```

admin-bff пишет JSON-logs (slog). Полезные поля: `msg`, `err`,
`upstream` (для proxy errors).

Типовые проблемы:
- `dial tcp ...:6379: connection refused` — Redis недоступен,
  admin-bff падает на старте.
- `identity_unavailable` (502 в логах) — identity-service лежит,
  логин не работает.
- `csrf_mismatch` — frontend не послал X-CSRF-Token (clear local
  storage и повторить).
