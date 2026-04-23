-- +goose Up
CREATE TABLE account_deletion_codes (
    user_id     uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    code_hash   text NOT NULL,
    issued_at   timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz NOT NULL,
    attempts    integer NOT NULL DEFAULT 0
);

-- +goose Down
DROP TABLE IF EXISTS account_deletion_codes;
