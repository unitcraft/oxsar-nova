-- +goose Up
CREATE TABLE messages (
    id           uuid PRIMARY KEY,
    to_user_id   uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    from_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    folder       integer NOT NULL,
    subject      text NOT NULL,
    body         text NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    read_at      timestamptz
);
CREATE INDEX ix_messages_inbox ON messages(to_user_id, created_at DESC);

CREATE TABLE chat_messages (
    id         uuid PRIMARY KEY,
    channel    text NOT NULL,          -- global | ally:<id> | pm:<u1>:<u2>
    author_id  uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body       text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ix_chat_channel_at ON chat_messages(channel, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS messages;
