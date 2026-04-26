-- +goose Up
ALTER TABLE users ADD COLUMN last_seen_at timestamptz NOT NULL DEFAULT now();

-- Индекс для воркера — выбираем пользователей по last_seen_at.
CREATE INDEX IF NOT EXISTS ix_users_last_seen ON users(last_seen_at);

-- +goose Down
DROP INDEX IF EXISTS ix_users_last_seen;
ALTER TABLE users DROP COLUMN IF EXISTS last_seen_at;
