-- +goose Up
-- Расширяем market_lots для лотов с кораблями:
-- kind='resource' — обычный лот ресурса (старое поведение, обратная совместимость).
-- kind='fleet'    — лот с пакетом кораблей (sell_fleet jsonb: {"unit_id": count}).

-- Снимаем старый CHECK sell_resource — теперь поле nullable для fleet-лотов.
ALTER TABLE market_lots DROP CONSTRAINT IF EXISTS market_lots_sell_resource_check;
ALTER TABLE market_lots DROP CONSTRAINT IF EXISTS market_lots_diff_resource;
ALTER TABLE market_lots ALTER COLUMN sell_resource DROP NOT NULL;
ALTER TABLE market_lots ALTER COLUMN sell_amount DROP NOT NULL;

ALTER TABLE market_lots
    ADD COLUMN IF NOT EXISTS kind text NOT NULL DEFAULT 'resource'
        CHECK (kind IN ('resource', 'fleet'));
ALTER TABLE market_lots
    ADD COLUMN IF NOT EXISTS sell_fleet jsonb;

-- Согласованность: для resource нужны sell_resource/sell_amount;
-- для fleet нужен sell_fleet, но не sell_resource.
ALTER TABLE market_lots ADD CONSTRAINT market_lots_kind_shape CHECK (
    (kind = 'resource' AND sell_resource IS NOT NULL AND sell_amount IS NOT NULL AND sell_fleet IS NULL
        AND sell_resource IN ('metal','silicon','hydrogen')
        AND (sell_resource <> buy_resource))
    OR
    (kind = 'fleet' AND sell_fleet IS NOT NULL
        AND sell_resource IS NULL AND sell_amount IS NULL)
);

CREATE INDEX IF NOT EXISTS market_lots_kind_idx ON market_lots(kind, state, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS market_lots_kind_idx;
ALTER TABLE market_lots DROP CONSTRAINT IF EXISTS market_lots_kind_shape;
ALTER TABLE market_lots DROP COLUMN IF EXISTS sell_fleet;
ALTER TABLE market_lots DROP COLUMN IF EXISTS kind;
ALTER TABLE market_lots ALTER COLUMN sell_amount SET NOT NULL;
ALTER TABLE market_lots ALTER COLUMN sell_resource SET NOT NULL;
