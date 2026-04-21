-- +goose Up
ALTER TABLE messages ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX messages_to_user_not_deleted ON messages (to_user_id, created_at DESC)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS messages_to_user_not_deleted;
ALTER TABLE messages DROP COLUMN deleted_at;
