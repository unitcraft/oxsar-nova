-- План 38 Ф.2: billing-db schema.
-- Архитектура: см. docs/plans/38-billing-service.md.
--
-- Принципы:
-- * Все суммы — BIGINT в минимальных единицах валюты (копейки/центы).
--   NUMERIC и FLOAT не используем — медленно и опасно для денег.
-- * transactions immutable (INSERT-only), бухгалтерская книга.
-- * wallets.balance — производная от SUM(transactions.delta), но
--   материализована для скорости (с SELECT FOR UPDATE на UPDATE).
-- * webhook_log хранит сырые webhook-запросы навсегда (для disputes).
-- * idempotency_keys — Stripe-style, TTL 24h.

-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- +goose StatementEnd

CREATE TABLE wallets (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL,
    currency_code TEXT NOT NULL DEFAULT 'OXC',
    balance       BIGINT NOT NULL DEFAULT 0,
    -- Заморозка кошелька при расхождении reconcile (см. план 38 §Reconciliation).
    -- При frozen=true все списания/пополнения возвращают 423 Locked.
    frozen        BOOLEAN NOT NULL DEFAULT false,
    frozen_reason TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, currency_code)
);
CREATE INDEX ix_wallets_user ON wallets(user_id);

-- Immutable бухгалтерская книга. INSERT-only, никаких UPDATE/DELETE.
-- Двойная запись: from_account → to_account.
-- Примеры:
--   top_up:        from='payment:robokassa:order_<uuid>', to='wallet:user_<uuid>:OXC'
--   feedback_vote: from='wallet:user_<uuid>:OXC',         to='vote:feedback:<feedback_id>'
--   refund:        from='wallet:user_<uuid>:OXC',         to='refund:order_<uuid>'
CREATE TABLE transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id       UUID NOT NULL REFERENCES wallets(id),
    delta           BIGINT NOT NULL,           -- + при пополнении, − при списании
    balance_after   BIGINT NOT NULL,           -- snapshot для аудита
    from_account    TEXT NOT NULL,
    to_account      TEXT NOT NULL,
    reason          TEXT NOT NULL,             -- 'top_up' | 'feedback_vote' | 'refund' | ...
    ref_id          TEXT,                      -- order_id / feedback_id / etc
    idempotency_key TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ix_transactions_wallet     ON transactions(wallet_id, created_at DESC);
CREATE INDEX ix_transactions_idempotent ON transactions(idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX ix_transactions_ref        ON transactions(ref_id) WHERE ref_id IS NOT NULL;

-- Платёжные заказы.
CREATE TABLE payment_orders (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL,
    provider     TEXT NOT NULL,                  -- 'robokassa' | 'enot' | 'mock'
    package_id   TEXT NOT NULL,                  -- 'pack_500' | 'pack_2000' | ...
    amount_kop   BIGINT NOT NULL,                -- сумма к оплате (RUB в копейках)
    credits      BIGINT NOT NULL,                -- сколько OXC получит юзер
    status       TEXT NOT NULL,                  -- 'pending' | 'paid' | 'failed' | 'expired'
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    paid_at      TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ NOT NULL DEFAULT now() + interval '1 hour'
);
CREATE INDEX ix_payment_orders_user   ON payment_orders(user_id, created_at DESC);
CREATE INDEX ix_payment_orders_status ON payment_orders(status, expires_at) WHERE status = 'pending';

-- Сырой лог webhook'ов от платёжных шлюзов. Никогда не удаляется
-- (или архивируется в холодное хранилище через несколько лет).
-- Нужно для разбора disputes/chargebacks.
CREATE TABLE webhook_log (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider     TEXT NOT NULL,
    received_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    headers      JSONB NOT NULL,
    body         BYTEA NOT NULL,
    signature_ok BOOLEAN,                       -- NULL до verify, true/false после
    order_id     UUID,                          -- если удалось распарсить
    processed_at TIMESTAMPTZ,                   -- когда обработан end-to-end
    error        TEXT                           -- если обработка упала — текст ошибки
);
CREATE INDEX ix_webhook_log_received ON webhook_log(received_at DESC);
CREATE INDEX ix_webhook_log_order    ON webhook_log(order_id) WHERE order_id IS NOT NULL;

-- Stripe-style idempotency keys.
-- TTL 24 часа: записи старше — удаляются background-job (или по факту запроса).
CREATE TABLE idempotency_keys (
    key             TEXT PRIMARY KEY,
    user_id         UUID NOT NULL,              -- ключ scoped к юзеру (защита от cross-user replay)
    request_hash    TEXT NOT NULL,              -- sha256(method+path+body) — для defence in depth
    response_body   JSONB NOT NULL,
    response_status INTEGER NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ NOT NULL DEFAULT now() + interval '24 hours'
);
CREATE INDEX ix_idempotency_expires ON idempotency_keys(expires_at);
CREATE INDEX ix_idempotency_user    ON idempotency_keys(user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS webhook_log;
DROP TABLE IF EXISTS payment_orders;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS wallets;
