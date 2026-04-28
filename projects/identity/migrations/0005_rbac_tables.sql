-- План 52 Ф.1: RBAC unification — identity-service как единственный источник
-- ролей и permissions. До этой миграции роли хранились в users.roles TEXT[]
-- (плоский массив строк) + дублировались в game-nova/users.role (ENUM).
--
-- После миграции:
--   - roles                 — динамический справочник ролей.
--   - permissions           — справочник гранулярных permissions.
--   - role_permissions      — many-to-many mapping роль → permissions.
--   - user_roles            — assignments пользователь → роли с TTL.
--   - audit_role_changes    — immutable log изменений ролей.
--
-- users.roles[] остаётся в этой миграции (миграция данных в 0006 после
-- проверки seed). Удаление колонки — отдельной миграцией 0007.

-- +goose Up
-- +goose StatementBegin

-- 1. Справочник ролей.
CREATE TABLE roles (
    id           SERIAL PRIMARY KEY,
    name         VARCHAR(64) UNIQUE NOT NULL,
    description  TEXT,
    is_system    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 2. Справочник permissions (action-based, format: domain:action).
CREATE TABLE permissions (
    id           SERIAL PRIMARY KEY,
    name         VARCHAR(128) UNIQUE NOT NULL,
    description  TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 3. Mapping роль → permissions (many-to-many).
CREATE TABLE role_permissions (
    role_id        INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id  INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);
CREATE INDEX ON role_permissions(permission_id);

-- 4. Assignments пользователь → роль с опциональным TTL.
CREATE TABLE user_roles (
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id      INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_by   UUID REFERENCES users(id),
    granted_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ,
    PRIMARY KEY (user_id, role_id)
);
CREATE INDEX ON user_roles(role_id);
CREATE INDEX ON user_roles(expires_at) WHERE expires_at IS NOT NULL;

-- 5. Audit log (immutable: только INSERT, без UPDATE/DELETE).
CREATE TABLE audit_role_changes (
    id           BIGSERIAL PRIMARY KEY,
    actor_id     UUID NOT NULL,
    target_id    UUID NOT NULL,
    role_name    VARCHAR(64) NOT NULL,
    action       VARCHAR(16) NOT NULL CHECK (action IN ('grant', 'revoke')),
    reason       TEXT,
    ip_address   INET,
    user_agent   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON audit_role_changes(actor_id);
CREATE INDEX ON audit_role_changes(target_id);
CREATE INDEX ON audit_role_changes(created_at DESC);

-- 6. Системные роли (is_system=true — нельзя удалить из админки).
INSERT INTO roles (name, description, is_system) VALUES
    ('player',         'Обычный игрок (default)', TRUE),
    ('support',        'Поддержка: чтение тикетов, аудита, профилей юзеров', TRUE),
    ('moderator',      'Модерация UGC: чат, ник-неймы, имена планет, баны', TRUE),
    ('admin',          'Игровая админка: planet ops, fleet recall, events, грант ресурсов', TRUE),
    ('billing_admin',  'Биллинг: отчёты, возвраты, отключение пополнения, audit', TRUE),
    ('superadmin',     'Управление ролями других юзеров, системные настройки', TRUE);

-- 7. Permissions — гранулярные действия в формате domain:action.
INSERT INTO permissions (name, description) VALUES
    -- Users (system-level operations через identity-service)
    ('users:read',           'Просмотр списка юзеров и их профилей'),
    ('users:warn',           'Предупреждение юзера (без блокировки)'),
    ('users:mute',           'Mute юзера (chat/ugc)'),
    ('users:ban',            'Бан юзера (полная блокировка доступа)'),
    ('users:delete',         'Soft-delete юзера'),
    ('users:create',         'Создание юзера через админку'),

    -- Roles management (только superadmin)
    ('roles:read',           'Просмотр ролей и assignments'),
    ('roles:grant',          'Выдача роли юзеру'),
    ('roles:revoke',         'Снятие роли с юзера'),

    -- Audit
    ('audit:read',           'Просмотр audit log изменений ролей и admin-действий'),

    -- Tickets (support)
    ('tickets:read',         'Просмотр support-тикетов'),
    ('tickets:write',        'Ответы на тикеты, изменение статуса'),

    -- UGC moderation (план 48)
    ('ugc:read',             'Просмотр UGC-репортов от пользователей'),
    ('ugc:moderate',         'Принятие решений по UGC-репортам'),

    -- Game operations (game-nova)
    ('game:events:retry',     'Повтор event-обработки'),
    ('game:events:cancel',    'Отмена event'),
    ('game:planets:transfer', 'Передача планеты другому юзеру'),
    ('game:planets:rename',   'Переименование планеты'),
    ('game:planets:delete',   'Удаление планеты'),
    ('game:fleets:recall',    'Принудительный отзыв флота'),
    ('game:resources:grant',  'Выдача ресурсов юзеру'),
    ('game:credits:grant',    'Выдача игровых кредитов юзеру'),
    ('game:artefacts:grant',  'Выдача артефактов юзеру'),

    -- Billing (план 54)
    ('billing:read',         'Просмотр платежей, отчётов, лимитов'),
    ('billing:refund',       'Возврат платежа'),
    ('billing:limits',       'Включение/выключение лимита самозанятого'),
    ('billing:reports',      'Экспорт отчётов CSV/PDF'),

    -- System (только superadmin)
    ('system:config',        'Изменение системных настроек');

-- 8. Mapping ролей на permissions.
-- player — без специальных permissions (default).

-- support: read-only по юзерам и тикетам, audit
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.name = 'support' AND p.name IN (
    'users:read', 'audit:read', 'tickets:read', 'tickets:write'
);

-- moderator: UGC + базовые санкции
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.name = 'moderator' AND p.name IN (
    'users:read', 'users:warn', 'users:mute', 'audit:read',
    'ugc:read', 'ugc:moderate'
);

-- admin: всё игровое + расширенные санкции (без управления ролями)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.name = 'admin' AND p.name IN (
    'users:read', 'users:warn', 'users:mute', 'users:ban', 'users:delete',
    'audit:read', 'tickets:read', 'tickets:write',
    'ugc:read', 'ugc:moderate',
    'game:events:retry', 'game:events:cancel',
    'game:planets:transfer', 'game:planets:rename', 'game:planets:delete',
    'game:fleets:recall',
    'game:resources:grant', 'game:credits:grant', 'game:artefacts:grant'
);

-- billing_admin: только биллинг
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.name = 'billing_admin' AND p.name IN (
    'audit:read',
    'billing:read', 'billing:refund', 'billing:limits', 'billing:reports'
);

-- superadmin: всё (включая управление ролями + system config)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.name = 'superadmin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS audit_role_changes;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
-- +goose StatementEnd
