# Доступ к admin-консоли

Как получить доступ к `admin.oxsar-nova.ru` и что делать в типовых
ситуациях. Архитектурный обзор — в
[architecture/admin-frontend.md](../architecture/admin-frontend.md).

---

## Bootstrap: первый superadmin

Admin-консоль защищена permission-чек'ами: чтобы кто-то мог управлять
ролями (`roles:grant` / `roles:revoke`), у него уже должна быть роль
`superadmin`. Замкнутый круг разрывается через CLI на сервере (план
53 §Bootstrap):

```bash
# На VPS, где запущен identity-service:
identity-cli grant-role superadmin <user-uuid> --reason "bootstrap superadmin"
```

**До выполнения этой команды admin-консоль недоступна для всех** —
никто не имеет permission `roles:grant`, чтобы выдать роль кому
ещё. Это by design.

(Команда `identity-cli` будет реализована в sub-плане
53e-identity-cli — пока вручную через SQL миграцию seed-роли в
БД identity.)

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
