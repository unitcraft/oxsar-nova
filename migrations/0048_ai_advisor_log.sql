-- +goose Up
CREATE TABLE ai_advisor_log (
    id          TEXT        PRIMARY KEY,
    user_id     TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    model       TEXT        NOT NULL,
    question    TEXT        NOT NULL,
    answer      TEXT        NOT NULL DEFAULT '',
    tokens_used INT         NOT NULL DEFAULT 0,
    credits     NUMERIC(10,2) NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_ai_advisor_log_user_day ON ai_advisor_log(user_id, created_at);

-- +goose Down
DROP TABLE IF EXISTS ai_advisor_log;
