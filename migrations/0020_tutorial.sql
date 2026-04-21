-- +goose Up
-- tutorial_state уже есть в users (0001_init.sql, колонка DEFAULT 0).
-- Добавляем таблицу tutorial_rewards для идемпотентной выдачи наград.
CREATE TABLE IF NOT EXISTS tutorial_rewards (
    user_id    uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    step       integer     NOT NULL,
    rewarded_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, step)
);

-- +goose Down
DROP TABLE IF EXISTS tutorial_rewards;
