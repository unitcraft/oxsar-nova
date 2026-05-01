-- План 72.1.27: Premium-лоты + Ban (legacy `Stock.class.php::premiumLot/ban`).
--
-- Premium: любой игрок за credit (max(10, price × 0.5%)) ставит лот
-- featured на 2 часа (legacy EXCH_PREMIUM_LOT_EXPIRY_TIME). Featured-
-- лоты идут сверху списка (max 5 одновременно).
--
-- Ban: admin-only действие — лот возвращается продавцу (escrow refund),
-- статус 'banned'. Legacy EXCH_BAN_TIME не определён в config — у нас
-- семантика только status='banned' без блокировки продавца.

-- +goose Up

ALTER TABLE exchange_lots
    ADD COLUMN IF NOT EXISTS featured_at timestamptz,
    ADD COLUMN IF NOT EXISTS banned_at   timestamptz;

-- Расширяем CHECK status: добавляем 'banned' (legacy ESTATUS_BANNED).
ALTER TABLE exchange_lots
    DROP CONSTRAINT IF EXISTS exchange_lots_status_check;

ALTER TABLE exchange_lots
    ADD CONSTRAINT exchange_lots_status_check
    CHECK (status IN ('active', 'sold', 'cancelled', 'expired', 'banned'));

-- Partial-индекс для list query: featured-лоты сверху, expiry 2 часа.
-- WHERE NOT banned гарантирует что забаненные не выпрыгивают.
CREATE INDEX IF NOT EXISTS exchange_lots_featured_idx
    ON exchange_lots (featured_at DESC)
    WHERE banned_at IS NULL AND featured_at IS NOT NULL AND status = 'active';

-- +goose Down

DROP INDEX IF EXISTS exchange_lots_featured_idx;

ALTER TABLE exchange_lots DROP CONSTRAINT IF EXISTS exchange_lots_status_check;
ALTER TABLE exchange_lots
    ADD CONSTRAINT exchange_lots_status_check
    CHECK (status IN ('active', 'sold', 'cancelled', 'expired'));

ALTER TABLE exchange_lots
    DROP COLUMN IF EXISTS banned_at,
    DROP COLUMN IF EXISTS featured_at;
