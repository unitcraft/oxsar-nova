# Dev test users

Тестовые учётки для локального dev-стека (`make dev-up`, http://localhost:5174,
http://localhost:5173, http://localhost:5175). Не использовать в проде —
пароли публично известны и слабы.

Создано: 2026-04-29 ручной регистрацией через `POST /auth/register` +
SQL-grant ролей в `identity-db.user_roles`. Канонический описанный
механизм первичного grant'а (включая выдачу `superadmin`) — в
[admin-access.md § Bootstrap](admin-access.md#bootstrap-первый-superadmin).

Поле «логин» в форме принимает username **или** email. Пароль у всех
одинаковый: `DevPass123`.

| Логин | Email | Роль | Назначение |
|-------|-------|------|------------|
| `player1` | player1@test.local | player | обычный игрок (default) |
| `support1` | support1@test.local | support | чтение тикетов, аудита, профилей |
| `mod1` | mod1@test.local | moderator | модерация UGC, баны |
| `admin1` | admin1@test.local | admin | игровая админка, planet ops |
| `billing1` | billing1@test.local | billing_admin | биллинг, возвраты, audit |
| `super1` | super1@test.local | superadmin | управление ролями + всё остальное |

После grant'а ролей нужно перелогиниться — JWT, выданный при `register`,
содержит только `player`. Новый токен после login получит актуальный
список ролей из `user_roles`.

## Воссоздание после wipe identity-db

```bash
# 1. Зарегистрировать всех 6 юзеров
for u in player1 support1 mod1 admin1 billing1 super1; do
  curl -s -X POST http://localhost:9000/auth/register \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$u\",\"email\":\"$u@test.local\",\"password\":\"DevPass123\",\"consent_accepted\":true,\"terms_accepted\":true}"
done

# 2. Раздать роли
docker exec deploy-identity-db-1 psql -U identitysvc -d identitysvc -c "
INSERT INTO user_roles (user_id, role_id, granted_at)
SELECT u.id, r.id, now()
FROM users u JOIN roles r ON
  (u.username='support1'  AND r.name='support')      OR
  (u.username='mod1'      AND r.name='moderator')    OR
  (u.username='admin1'    AND r.name='admin')        OR
  (u.username='billing1'  AND r.name='billing_admin')OR
  (u.username='super1'    AND r.name='superadmin')
ON CONFLICT (user_id, role_id) DO NOTHING;"
```
