-- +goose Up
ALTER TABLE chat_messages
    ADD COLUMN IF NOT EXISTS edited_at  timestamptz,
    ADD COLUMN IF NOT EXISTS deleted_at timestamptz;

-- +goose Down
ALTER TABLE chat_messages
    DROP COLUMN IF EXISTS edited_at,
    DROP COLUMN IF EXISTS deleted_at;
