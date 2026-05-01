-- План 72.1.42: legacy `ArtefactMarket` использует ads.lifetime
-- (artefact_datasheet) для автоматического снятия лота. Origin не
-- имел этого механизма — офферы висели вечно.
--
-- Добавляем expire_at и event KindArtMarketExpire = 91 (новый kind).
-- Воркер при срабатывании события: state='listed' → 'held', DELETE offer.

-- +goose Up

ALTER TABLE artefact_offers
    ADD COLUMN IF NOT EXISTS expire_at timestamptz;

-- Дефолт +30 дней (legacy lifetime varies по типу артефакта; для MVP
-- единый TTL).
UPDATE artefact_offers
   SET expire_at = listed_at + INTERVAL '30 days'
 WHERE expire_at IS NULL;

ALTER TABLE artefact_offers
    ALTER COLUMN expire_at SET NOT NULL;

CREATE INDEX IF NOT EXISTS ix_artefact_offers_expire
    ON artefact_offers (expire_at);

-- +goose Down

DROP INDEX IF EXISTS ix_artefact_offers_expire;
ALTER TABLE artefact_offers DROP COLUMN IF EXISTS expire_at;
