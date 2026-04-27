# План 52: RBAC unification — identity-service как единственный источник ролей

**Дата**: 2026-04-27
**Статус**: Активный
**Зависимости**: **План 51** (переименование auth → identity) должен быть выполнен.
**Связанные документы**: [51-rename-auth-to-identity.md](51-rename-auth-to-identity.md),
[53-admin-frontend.md](53-admin-frontend.md), [54-billing-limits.md](54-billing-limits.md).

---

## Зачем

Сейчас в монорепо **две разные системы ролей**:

| Источник | Поле | Тип |
|---|---|---|
| `identity/users.roles` | `TEXT[] DEFAULT '{player}'` | массив строк |
| `game-nova/users.role` | `user_role ENUM('player','support','admin','superadmin')` | единичный enum |

Это приводит к:
- Двойной источник правды → consistency-проблемы (юзер банится в
  identity, но в game-nova по-прежнему `admin`).
- Дублирование логики проверки ролей в каждом сервисе.
- Существующий `POST /api/admin/users/{id}/role` в game-nova пишет
  не туда (в локальную копию вместо идентити).
- Невозможно дать роль `billing_admin` (нужна в плане 54) — её нет
  в game-nova ENUM.

Современная архитектура микросервисов (Auth0, Keycloak, ORY Kratos):
**один Identity Provider владеет ролями + permissions, остальные
сервисы читают их из JWT** (локальная валидация через JWKS, без
сетевых вызовов на каждый запрос).

## Архитектура (Production-grade)

### Модель данных в identity

**Таблица `roles`** (динамическая, не enum):
```sql
CREATE TABLE roles (
    id           SERIAL PRIMARY KEY,
    name         VARCHAR(64) UNIQUE NOT NULL,  -- 'admin', 'billing_admin', ...
    description  TEXT,
    is_system    BOOLEAN DEFAULT FALSE,        -- системная (нельзя удалить)
    created_at   TIMESTAMPTZ DEFAULT now()
);
```

**Таблица `permissions`** (action-based, гранулярные):
```sql
CREATE TABLE permissions (
    id           SERIAL PRIMARY KEY,
    name         VARCHAR(128) UNIQUE NOT NULL, -- 'billing:read', 'users:delete'
    description  TEXT,
    created_at   TIMESTAMPTZ DEFAULT now()
);
```

**Таблица `role_permissions`** (mapping):
```sql
CREATE TABLE role_permissions (
    role_id        INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id  INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);
```

**Таблица `user_roles`** (assignments):
```sql
CREATE TABLE user_roles (
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id      INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_by   UUID REFERENCES users(id),         -- кто назначил
    granted_at   TIMESTAMPTZ DEFAULT now(),
    expires_at   TIMESTAMPTZ,                       -- опциональный TTL
    PRIMARY KEY (user_id, role_id)
);
```

**Таблица `audit_role_changes`** (immutable log):
```sql
CREATE TABLE audit_role_changes (
    id           BIGSERIAL PRIMARY KEY,
    actor_id     UUID NOT NULL,           -- кто делал изменение
    target_id    UUID NOT NULL,           -- кому
    role_name    VARCHAR(64) NOT NULL,
    action       VARCHAR(16) NOT NULL,    -- 'grant' | 'revoke'
    reason       TEXT,                     -- комментарий админа
    ip_address   INET,
    user_agent   TEXT,
    created_at   TIMESTAMPTZ DEFAULT now()
);
```

### Стартовый набор ролей

| Role name | System | Permissions |
|---|---|---|
| `player` | yes | (default, без специальных permissions) |
| `support` | yes | `users:read`, `tickets:write`, `audit:read` |
| `moderator` | yes | `ugc:read`, `ugc:moderate`, `users:warn`, `users:mute` |
| `admin` | yes | все support+moderator + `users:delete`, `events:retry`, `planets:transfer`, `fleets:recall` |
| `billing_admin` | yes | `billing:read`, `billing:refund`, `billing:reports`, `billing:limits` |
| `superadmin` | yes | `roles:grant`, `roles:revoke`, `users:create`, `system:config` (управление ролями других) |

`is_system=true` означает что роль нельзя удалить из админки — только
переименовать description / поменять permissions mapping.

### JWT claims

**Что было**:
```json
{ "sub": "uuid", "username": "name", "email": "x@y", "exp": ... }
```

**Что станет**:
```json
{
  "sub": "uuid",
  "iss": "https://identity.oxsar-nova.ru",
  "aud": ["api"],
  "exp": 1700000900,
  "iat": 1700000000,
  "jti": "uuid-токена",          // для revocation list
  "username": "name",
  "email": "x@y",
  "roles": ["admin", "billing_admin"],
  "permissions": ["billing:read", "billing:write", "users:delete", ...]
}
```

`permissions` — flatten из всех ролей юзера, дедуплицированный.
Размер JWT ≈ 1-2KB при 6 ролях × 5 permissions — приемлемо.

**TTL**:
- Access token: **15 минут** (баланс между security и UX).
- Refresh token: **7 дней** с **rotation на каждом use** (детектирование
  украденных токенов: использование старого refresh после ротации →
  немедленно отзываем всю цепочку).

### JWKS endpoint

`GET /identity/.well-known/jwks.json` — стандартный OIDC endpoint.
Все клиенты (game-nova, billing, portal, game-origin) валидируют JWT
**только локально** через cached JWKS (cache TTL 24h, refresh on key
rotation).

### Token revocation

При смене роли / бане / явном logout — токен может быть в браузере
ещё до 15 минут (TTL). Решение:

1. **Короткий TTL** (15 мин) делает revocation естественной — даже
   без active blacklist старый токен умрёт сам.
2. **Critical actions** (delete user, ban user) проверяют не только
   JWT, но и текущий `users.is_banned` через identity — этот вызов
   делает только game-nova/admin handlers, не на каждый request.
3. **JTI blacklist** в Redis с TTL 15 мин — для случаев когда нужно
   forced revocation. Каждый клиент проверяет JTI в blacklist
   перед обработкой запроса.

## API в identity-service

```
# Roles management (требует superadmin)
GET    /api/admin/roles                       # список всех ролей
POST   /api/admin/roles                       # создать (для будущих кастомных)
GET    /api/admin/roles/{id}                  # детали + permissions
PUT    /api/admin/roles/{id}                  # переименовать description
DELETE /api/admin/roles/{id}                  # только не is_system

# Permissions
GET    /api/admin/permissions                 # список всех
POST   /api/admin/roles/{id}/permissions      # добавить permission к роли
DELETE /api/admin/roles/{id}/permissions/{p}  # снять

# User-role assignments (требует superadmin)
GET    /api/admin/users                       # список юзеров с фильтрами
GET    /api/admin/users/{id}                  # детали + текущие роли
GET    /api/admin/users/{id}/roles
POST   /api/admin/users/{id}/roles            # body: {"role": "admin", "expires_at": "...", "reason": "..."}
DELETE /api/admin/users/{id}/roles/{role}     # snять; обязательно reason

# Audit
GET    /api/admin/audit/role-changes          # с пагинацией и фильтрами
GET    /api/admin/audit/role-changes/{id}     # детали записи

# JWT introspection (для admin'ов и сервисов)
POST   /identity/oauth/introspect             # RFC 7662; возвращает active/scope/sub
POST   /identity/oauth/revoke                 # RFC 7009; помещает JTI в blacklist
```

Все `/api/admin/*` endpoints:
- Проверяют JWT (через middleware).
- Проверяют permission (`roles:grant` / `roles:read` / etc).
- Логируют запрос в audit (action, actor, target, ip, ua).
- Rate limit: 60 req/min per admin.

## Что меняется в каждом сервисе

### identity (бывший auth)

- Новые таблицы: roles, permissions, role_permissions, user_roles,
  audit_role_changes.
- Удалить старое поле `users.roles TEXT[]` через миграцию (заменяется
  на user_roles table).
- Новая модель в Go: `internal/identitysvc/rbac.go` (Role, Permission,
  Assignment структуры + service-методы).
- Новые HTTP-handlers: `/api/admin/roles/*`, `/api/admin/users/*/roles`.
- Расширение JWT issuance: при выпуске токена → join user_roles +
  role_permissions, заполнить claims `roles[]` и `permissions[]`.
- Stub для `/oauth/introspect` (RFC 7662) — пока не используется
  внешними клиентами, но готов к добавлению.

### game-nova/backend

- **Снести** `users.role user_role` ENUM через миграцию (data
  migration: текущие role-значения трансформируем в записи user_roles
  identity-service через одноразовый job).
- **Снести** `internal/admin/rbac.go::loadRole` — больше не нужен.
- **Переписать** `internal/admin/rbac.go::RequireRole` чтобы читать
  permissions из JWT-claims вместо БД-запроса.
- **Снести** endpoint `POST /api/admin/users/{id}/role` (теперь это в
  identity, game-nova через UI просто redirect/proxy).
- **Оставить** game-domain ролей если есть (например, observer-mode
  как `users.is_observer BOOLEAN` отдельно от системных ролей).

### billing/backend

- Уже использует JWT для аутентификации, но не проверяет роли.
- Добавить middleware `RequirePermission("billing:read")` для
  `/api/admin/billing/*` endpoints (готовится в плане 54).
- В тестах: helper для генерации test JWT с произвольными
  permissions.

### portal/backend, game-origin

- Permission-checks в этих сервисах **не нужны на старте** (нет
  admin-функциональности).
- Когда понадобится — добавляем тот же middleware.

### Все клиенты (frontend)

- В TanStack Query / Axios interceptor: парсить JWT → выставлять в
  Zustand store список permissions. UI скрывает кнопки на основе
  permissions (`if (permissions.includes('billing:refund')) <Button />`).

## Этапы

### Ф.1. Миграции БД

- Создать миграцию `identity/migrations/N_rbac_tables.sql`:
  roles, permissions, role_permissions, user_roles,
  audit_role_changes.
- Seed-миграция: вставить 6 системных ролей и их permissions.
- Data-migration: для каждого юзера в `users.roles[]` создать запись
  в `user_roles` с granted_by=NULL, reason='migration'.
- Удалить колонку `users.roles` — отдельной миграцией после успешного
  переноса.
- В `game-nova/migrations/`: миграция «before» (один раз делаем
  data-export role→user_role mapping), потом DROP TABLE/COLUMN
  `users.role` + DROP TYPE user_role.

### Ф.2. Identity-service: backend

- Реализовать `internal/identitysvc/rbac.go`:
  - `RoleRepo`, `PermissionRepo`, `UserRolesRepo`.
  - `RBACService` с методами `GrantRole`, `RevokeRole`,
    `ListUserRoles`, `ListAllRoles`, `CheckPermission`.
- HTTP handlers: `internal/identitysvc/rbac_handler.go`.
- Audit log: middleware пишет в audit_role_changes на каждый
  POST/PUT/DELETE.
- JWT issuance: расширить `pkg/jwtrs` для включения `roles` и
  `permissions` в claims.
- Tests: integration tests для каждого endpoint (с реальной test-DB,
  как `consent_test.go` сейчас).

### Ф.3. Game-nova: миграция админки

- Удалить `users.role` ENUM (миграция).
- Переписать `internal/admin/rbac.go`:
  - `RequireRole(role string)` → `RequirePermission(perm string)`.
  - Чтение из JWT claims, не из БД.
- Все endpoints `ar.With(admin.RequireRole(db, RoleAdmin))` →
  `ar.With(admin.RequirePermission("scope:action"))`.
- Удалить endpoint `POST /api/admin/users/{id}/role` (теперь в
  identity).
- В `internal/admin/handler.go::Audit*` — продолжаем писать в
  game-nova local audit для game-domain действий, но system-level
  audit (роли) идёт в identity.

### Ф.4. Billing: middleware

- Добавить `internal/auth/permission.go` с `RequirePermission(perm)`
  middleware.
- В тестах — helper `signTestJWT(permissions []string)`.

### Ф.5. Game-origin (PHP)

- В `core/JwtAuth.php` (если есть) расширить: при resolve user'а
  читать `roles` и `permissions` из JWT, класть в `$_SESSION['roles']`,
  `$_SESSION['permissions']`.
- В `User::ifPermissions($perm)` — проверять список из JWT, не из
  na_permissions/na_group2permission (legacy таблицы остаются для
  in-game permissions, но system-level admin-permissions — из JWT).

### Ф.6. Тестирование

- Integration tests:
  - Юзер с ролью `admin` может вызвать DELETE /users/{id}.
  - Юзер с `billing_admin` может вызвать GET /billing/reports.
  - Юзер с `player` (только) НЕ может вызвать ни то ни другое (403).
  - Granting роли пишет audit-запись с правильным actor/target.
  - Revoking активной роли блокирует доступ к её endpoints
    (после refresh JWT).
- E2E (Playwright):
  - Логин как admin → доступ к admin-роутам.
  - SetRole(billing_admin) → новый JWT с этой ролью.
  - Logout → cleanup.

### Ф.7. Документация

- `docs/architecture/rbac.md` (новый файл): полная модель — роли,
  permissions, JWT claims, flow refresh, audit.
- `docs/ops/admin-access.md` (создаётся в плане 53, но базовая часть
  закладывается здесь): как админ получает свою первую роль (через
  CLI `identity-cli grant-role superadmin <user-uuid>` для bootstrap).
- Обновить `docs/plans/52-followup-checklist.md` со списком всех
  endpoints, которые нужно проверить.

### Ф.8. Финализация

- Smoke-test: новый юзер регистрируется → получает default роль
  `player` → JWT содержит `roles:["player"]`, `permissions:[]`.
- Существующие admin-юзеры: их роли мигрировали из game-nova users.role
  → identity user_roles, JWT при следующем refresh содержит правильный
  набор.
- Bootstrap superadmin: один из существующих юзеров получает
  `superadmin` через CLI (документировать процесс).
- Запись в `docs/project-creation.txt` итерация 52.

## Тестирование

- **Unit tests** в каждом сервисе: проверка middleware permission на
  моках JWT.
- **Integration tests**: реальная test-DB, реальные JWT (signed with
  test private key).
- **E2E**: Playwright-сценарий full RBAC flow.
- **Migration tests**: одноразовый job переноса роль из game-nova в
  identity — проверить идемпотентность (повторный запуск не дублирует).
- **Backward compat**: в течение grace-period 1-2 недели — юзеры со
  старыми JWT (без `permissions` claim) принимаются с fallback
  permissions = из роли.

## Риски

1. **Игроки во время миграции** теряют свои роли на 5-15 минут (TTL
   старого JWT). Митигация: data-migration ДО deploy'я; новый JWT
   issuance уже работает, старые токены grace-период 15 мин.
2. **Размер JWT**: 6 ролей × 5-10 permissions = 30-60 entries в claims.
   ~2-3KB JWT — нормально, но не extreme. Митигация: cap на 100
   permissions; если больше — refactoring (вместо `permissions[]`
   передаём только `roles[]`, клиенты резолвят через
   `/oauth/introspect`).
3. **Rollback**: если что-то ломается, откат двойной (identity migration
   + game-nova migration). Feature-flag `RBAC_FROM_IDENTITY=true`
   позволяет ходить и в старую логику тоже.
4. **Race condition**: между revoke роли и истечением старого JWT
   юзер ещё имеет доступ. Митигация: для critical actions (delete user)
   проверять `users.is_banned` свежим запросом в identity, а не из JWT.

## Альтернативы (отвергнуты)

- **Единый ENUM в game-nova**: оставить как есть, не добавлять
  `billing_admin` в game-nova — невозможно, биллинг админка не привязана
  к game-nova.
- **Federated identity (multi-tenant)**: identity-service выдаёт
  токены для нескольких независимых вселенных. Сейчас вселенная одна,
  оверкилл.
- **OAuth2 client_credentials для service-to-service**: не нужно
  пока. Когда понадобится machine-to-machine auth — добавим.

## Out of scope (отдельные планы)

- Поддержка SAML / external IdP federation (Google sign-in, и т.п.) —
  отдельный план, когда понадобится.
- Multi-factor authentication для юзеров (не админов; для админов
  WebAuthn в плане 53) — отдельный план.
- ABAC (attribute-based, на основе атрибутов ресурсов) — для нашего
  размера системы оверкилл, RBAC хватает.

## Итог

Single source of truth для ролей в identity-service. Чистая модель
(roles + permissions, динамические таблицы), JWT claims содержат всё
нужное для local validation, audit log immutable. Готовая база для
плана 53 (admin frontend) и 54 (billing limits).
