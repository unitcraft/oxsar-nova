-- +goose Up
CREATE TABLE friends (
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id   uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, friend_id),
    CHECK (user_id <> friend_id)
);

CREATE INDEX ix_friends_friend_id ON friends(friend_id);

-- +goose Down
DROP TABLE IF EXISTS friends;
