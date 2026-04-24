-- +goose Up
-- Журнал всех деструктивных админских действий (ban/unban/credit/role/
-- automsg/event-retry и т.д.). Не-read-only операции пишут запись через
-- middleware (backend/internal/admin/audit.go).
--
-- Запросы с фильтром: admin_id, action, target_id, from/to created_at.

CREATE TABLE admin_audit_log (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id    uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action      text NOT NULL,            -- 'ban' | 'unban' | 'credit' | 'role' | 'automsg.update' | 'event.retry' | 'event.cancel' | …
    target_kind text NOT NULL,            -- 'user' | 'automsg' | 'event' | '' (если не про объект)
    target_id   text NOT NULL DEFAULT '', -- id цели (uuid / string-key) или пусто
    payload     jsonb NOT NULL DEFAULT '{}'::jsonb,   -- тело запроса, без секретов
    status      smallint NOT NULL DEFAULT 200,        -- HTTP-код ответа
    ip          inet,
    user_agent  text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ix_admin_audit_log_created_at ON admin_audit_log (created_at DESC);
CREATE INDEX ix_admin_audit_log_admin_id   ON admin_audit_log (admin_id, created_at DESC);
CREATE INDEX ix_admin_audit_log_action     ON admin_audit_log (action, created_at DESC);
CREATE INDEX ix_admin_audit_log_target     ON admin_audit_log (target_kind, target_id) WHERE target_id <> '';

-- +goose Down
DROP TABLE IF EXISTS admin_audit_log;
