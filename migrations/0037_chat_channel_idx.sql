-- +goose Up
CREATE INDEX IF NOT EXISTS chat_messages_channel_created_idx
    ON chat_messages(channel, created_at DESC)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS chat_messages_channel_created_idx;
