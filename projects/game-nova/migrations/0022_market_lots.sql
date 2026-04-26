-- +goose Up
-- Лоты обмена ресурсов (ордерная книга биржи).
-- Продавец создаёт лот: хочу продать sell_amount ресурса sell_resource
-- и получить buy_amount ресурса buy_resource.
-- Покупатель принимает лот целиком.

CREATE TABLE IF NOT EXISTS market_lots (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id     uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    planet_id     uuid        NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    sell_resource text        NOT NULL CHECK (sell_resource IN ('metal', 'silicon', 'hydrogen')),
    sell_amount   bigint      NOT NULL CHECK (sell_amount > 0),
    buy_resource  text        NOT NULL CHECK (buy_resource IN ('metal', 'silicon', 'hydrogen')),
    buy_amount    bigint      NOT NULL CHECK (buy_amount > 0),
    state         text        NOT NULL DEFAULT 'open' CHECK (state IN ('open', 'accepted', 'cancelled')),
    buyer_id      uuid        REFERENCES users(id) ON DELETE SET NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT market_lots_diff_resource CHECK (sell_resource <> buy_resource)
);

CREATE INDEX IF NOT EXISTS market_lots_open_idx ON market_lots(state, sell_resource, created_at DESC)
    WHERE state = 'open';

-- +goose Down
DROP TABLE IF EXISTS market_lots;
