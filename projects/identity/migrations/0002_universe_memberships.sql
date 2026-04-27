-- +goose Up
-- +goose StatementBegin

-- universe_memberships: игровые серверы регистрируют здесь
-- при lazy join (первый запрос игрока к вселенной).
CREATE TABLE universe_memberships (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    universe_id TEXT NOT NULL,
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, universe_id)
);
CREATE INDEX ON universe_memberships (user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS universe_memberships;
-- +goose StatementEnd
