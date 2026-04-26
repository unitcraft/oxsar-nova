-- +goose Up
ALTER TABLE planets ADD COLUMN expires_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE planets DROP COLUMN IF EXISTS expires_at;
