-- +goose Up
CREATE TABLE credit_purchases (
    id             TEXT PRIMARY KEY,
    user_id        uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    package_key    TEXT NOT NULL,
    amount_credits INT  NOT NULL,
    price_rub      NUMERIC(10,2) NOT NULL,
    provider       TEXT NOT NULL DEFAULT 'robokassa',
    provider_id    TEXT UNIQUE,
    status         TEXT NOT NULL DEFAULT 'pending',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    paid_at        TIMESTAMPTZ
);
CREATE INDEX idx_credit_purchases_user   ON credit_purchases(user_id);
CREATE INDEX idx_credit_purchases_status ON credit_purchases(status) WHERE status = 'pending';

-- +goose Down
DROP TABLE credit_purchases;
