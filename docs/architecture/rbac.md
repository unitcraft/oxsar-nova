# RBAC: Role-Based Access Control в oxsar-nova

**Введён**: план 52 (2026-04-27)
**Owner**: identity-service.
**Связано**: [51-rename-auth-to-identity.md](../plans/51-rename-auth-to-identity.md),
[52-rbac-unification.md](../plans/52-rbac-unification.md),
[53-admin-frontend.md](../plans/53-admin-frontend.md),
[54-billing-limits-reports.md](../plans/54-billing-limits-reports.md).

---

## Принципы

1. **Single source of truth — identity-service.** Все остальные сервисы
   (game-nova, billing, portal, game-origin) читают роли и permissions
   ТОЛЬКО из JWT-claims (локальная валидация через JWKS, без походов в
   identity на каждый запрос).
2. **Roles — высокоуровневая абстракция, permissions — гранулярные.**
   Один админ-юзер имеет роль `admin`, а доступ к конкретному действию
   проверяется по permission `users:delete`. Mapping роль → permissions
   описан в `role_permissions` таблице identity-БД.
3. **Audit log immutable.** Каждое назначение/снятие роли пишется в
   `audit_role_changes` (только INSERT, без UPDATE/DELETE).
4. **Динамические таблицы, не enum.** Таблица `roles` — справочник;
   новые роли добавляются INSERT'ом, без миграций.

## Модель данных (в identity-БД)

```
roles (справочник)
  id, name (unique), description, is_system, created_at

permissions (справочник)
  id, name (domain:action, unique), description, created_at

role_permissions (mapping)
  role_id, permission_id (PK = пара)

user_roles (assignments)
  user_id, role_id (PK = пара),
  granted_by, granted_at, expires_at (опц.)

audit_role_changes (immutable log)
  id (BIGSERIAL), actor_id, target_id, role_name, action ('grant'|'revoke'),
  reason, ip_address, user_agent, created_at
```

## Системные роли

`is_system=true` означает «нельзя удалить из админки», но permissions
mapping можно менять.

| Role | Permissions |
|---|---|
| `player` | (default, без доп. permissions) |
| `support` | `users:read`, `audit:read`, `tickets:read`, `tickets:write` |
| `moderator` | + `users:warn`, `users:mute`, `ugc:read`, `ugc:moderate` |
| `admin` | + `users:ban`, `users:delete`, `game:events:retry/cancel`, `game:planets:transfer/rename/delete`, `game:fleets:recall`, `game:resources:grant`, `game:credits:grant`, `game:artefacts:grant` |
| `billing_admin` | `audit:read`, `billing:read`, `billing:refund`, `billing:limits`, `billing:reports` |
| `superadmin` | все permissions, включая `roles:grant`, `roles:revoke`, `system:config` |

## JWT claims

После плана 52 JWT-токен (access) содержит:

```json
{
  "sub": "uuid-юзера",
  "username": "name",
  "active_universes": ["uni01"],
  "roles": ["admin", "billing_admin"],
  "permissions": [
    "users:read", "users:warn", "users:ban",
    "billing:read", "billing:refund", "audit:read"
  ],
  "exp": 1700000900,
  "iat": 1700000000,
  "jti": "access:uuid"
}
```

`permissions` — flatten-список из всех ролей юзера, дедуплицирован
identity-сервисом перед выпуском токена (через
`RBACService.GetUserPermissions(userID)` → SQL DISTINCT).

Размер ~1-2KB при 6 ролях × 5-10 permissions — приемлемо для cookie/header.

## API в identity-service

```
# Read (требует "roles:read")
GET    /api/admin/roles                       # все роли
GET    /api/admin/roles/{id}/permissions      # permissions роли
GET    /api/admin/users/{id}/roles            # роли юзера

# Mutate (требует "roles:grant" / "roles:revoke")
POST   /api/admin/users/{id}/roles            # body: {role, expires_at?, reason}
DELETE /api/admin/users/{id}/roles/{role}     # ?reason=...

# Audit (требует "audit:read")
GET    /api/admin/audit/role-changes          # фильтры actor/target/action/since/until
```

## Использование в клиентах

### Go-сервисы (game-nova, billing, portal)

JWT валидируется через JWKS-loader (`internal/auth/jwksloader.go`),
permissions из claims кладутся в request context middleware'ом.

**game-nova** (`internal/admin/rbac.go`):

```go
// Legacy API (обратная совместимость): RequireRole(db, RoleAdmin)
// Сейчас читает permissions из JWT, db-параметр игнорируется.
ar.With(admin.RequireRole(db, admin.RoleAdmin)).Post("/credit", h.Credit)

// Новый идиоматичный API:
ar.With(admin.RequirePermission("game:credits:grant")).Post("/credit", h.Credit)
```

**billing** (`internal/billing/middleware.go`):

```go
r.With(billing.AuthMiddleware(ver)).
  With(billing.RequirePermission("billing:read")).
  Get("/api/admin/billing/payments", h.ListPayments)
```

### PHP (game-origin)

`JwtAuth::authenticate()` кладёт `$_SESSION['permissions']`. Проверка:

```php
if (!JwtAuth::hasPermission('billing:refund')) {
    Logger::dieMessage('PERMISSION_DENIED');
}
```

Legacy `JwtAuth::hasRole($role)` оставлен для backward-compat, но
для новых проверок предпочтительнее `hasPermission()`.

### Frontend

JWT в memory store (Zustand). Permissions — отдельный селектор:

```ts
const canRefund = useUser((s) => s.permissions.includes('billing:refund'))
```

UI скрывает кнопки на основе permissions; backend всё равно перепроверяет
(defence in depth).

## Token revocation

Короткий access-TTL (15 мин) — основной механизм. При смене роли
максимум через 15 мин юзер получит новый токен с обновлёнными permissions
(автоматически, через refresh-flow).

Для критических действий (ban, delete user) — дополнительная проверка
через identity API (`/api/admin/users/{id}` → `is_banned`), не из JWT.

JTI-blacklist в Redis (план 52 partial, расширение в плане 53):
позволяет немедленно отозвать конкретный токен. Каждый клиент проверяет
JTI до обработки запроса.

## Bootstrap superadmin

Первый superadmin выдаётся через CLI (план 53 Ф.8 описывает
`identity-cli grant-role superadmin <user-uuid>`). До его выполнения
admin-frontend недоступен (ни у кого нет permission `roles:grant`).

## Миграция данных при включении RBAC

План 52 Ф.1 миграция `0006_migrate_user_roles.sql` переносит
существующие `users.roles[]` (плоский array из identity-БД) в
`user_roles` table. После проверки старая колонка `users.roles` удаляется
отдельной миграцией (план 52 finalize).

В game-nova колонка `users.role` (ENUM) удалена миграцией
`0070_drop_users_role.sql` без переноса данных — все роли уже в
identity user_roles.

## Расширение в будущем

- **ABAC** (атрибут-based, на основе свойств ресурса): не сейчас. RBAC
  достаточно при текущем числе ролей.
- **Multi-tenancy** (роли per-universe): через permissions-namespace
  (`uni01:billing:read`). Структура поддерживает, нужно только заполнить.
- **OAuth2 introspection** (`/oauth/introspect`, RFC 7662): добавляется
  отдельным планом для machine-to-machine auth.
- **SAML federation** (Google sign-in, и т.п.): отдельный план.
