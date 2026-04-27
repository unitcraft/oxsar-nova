-- План 38 Ф.5: кошельки и платежи переехали в billing-service.
-- В auth-db не нужны users.global_credits и credit_transactions.

-- +goose Up
DROP TABLE IF EXISTS credit_transactions;
ALTER TABLE users DROP COLUMN IF EXISTS global_credits;

-- +goose Down
ALTER TABLE users ADD COLUMN global_credits BIGINT NOT NULL DEFAULT 0;
CREATE TABLE credit_transactions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id),
    delta      BIGINT NOT NULL,
    reason     TEXT NOT NULL,
    ref_id     TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ix_credit_transactions_user ON credit_transactions(user_id, created_at DESC);
