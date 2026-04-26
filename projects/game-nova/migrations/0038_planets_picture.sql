-- +goose Up
ALTER TABLE planets ADD COLUMN IF NOT EXISTS picture TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE planets DROP COLUMN IF EXISTS picture;
