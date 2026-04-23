-- +goose Up
ALTER TABLE users ADD COLUMN profession TEXT NOT NULL DEFAULT 'none';
ALTER TABLE users ADD COLUMN profession_changed_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE users DROP COLUMN profession;
ALTER TABLE users DROP COLUMN profession_changed_at;
