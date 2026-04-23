-- +goose Up
CREATE TABLE resource_transfers (
    id           bigserial PRIMARY KEY,
    from_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    to_user_id   uuid REFERENCES users(id) ON DELETE SET NULL,
    metal        numeric(20, 0) NOT NULL DEFAULT 0,
    silicon      numeric(20, 0) NOT NULL DEFAULT 0,
    hydrogen     numeric(20, 0) NOT NULL DEFAULT 0,
    at           timestamptz NOT NULL DEFAULT now(),
    CHECK (from_user_id IS NULL OR to_user_id IS NULL OR from_user_id <> to_user_id)
);

CREATE INDEX ix_rt_from ON resource_transfers(from_user_id, at DESC);
CREATE INDEX ix_rt_to   ON resource_transfers(to_user_id,   at DESC);

-- +goose Down
DROP TABLE IF EXISTS resource_transfers;
