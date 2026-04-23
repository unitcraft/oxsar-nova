-- +goose Up
CREATE TABLE user_notepad (
    user_id    uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    content    text NOT NULL DEFAULT '',
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS user_notepad;
