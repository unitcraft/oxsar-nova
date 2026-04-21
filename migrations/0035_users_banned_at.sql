-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS banned_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS banned_at;
