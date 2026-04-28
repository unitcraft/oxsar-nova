-- План 68: биржа артефактов player-to-player.
--
-- Семантика (см. docs/plans/68-remaster-exchange-artifacts.md):
--   - Лот = N штук артефактов одного unit_id, выставленных на продажу
--     за фиксированную цену в оксаритах (users.credit; см. ADR-0009 +
--     simplifications.md «План 68: credit-as-oxsarit»).
--   - Escrow: при создании лота N конкретных artefacts_user.id переводятся
--     в state='listed' и связываются с лотом через exchange_lot_items.
--     state='listed' уже введён миграцией 0013 (artefact_market) — биржа
--     повторно использует то же значение enum.
--   - Покупка: переводит owner всех artefacts_user из lot_items на buyer
--     (planet_id = home-планета buyer'а), state→'held', oxsarit-перевод
--     seller→buyer. UPDATE lot status='sold'.
--   - Отзыв (cancel) и истечение (expire) — артефакты возвращаются seller'у
--     state='held', user_id и planet_id у них не менялись с момента
--     listing, так что просто UPDATE state.
--   - Истечение реализовано через event-loop (KindExchangeExpire=66).
--     При создании лота вставляется event с fire_at=expires_at; при покупке
--     или cancel этот event переводится в state='cancelled'
--     (UPDATE events SET state='ok', чтобы worker его не подобрал — см.
--     event_state enum в 0001_init.sql, валидные wait/start/ok/error;
--     помечаем 'ok' с пометкой в last_error «cancelled by buy/cancel»).
--
-- exchange_history — типизированный audit-журнал по лоту. payload — JSONB
-- с фиксированной структурой (R13), парсится через payload.go в Go.
--
-- R10 (universe_id) НЕ применим: nova однобазная (universe = отдельный
-- инстанс БД, см. комментарий в 0075_alliance_audit_log.sql).
--
-- artifact_unit_id — int (ARTEFACT_* код из catalog), а не TEXT. План
-- упоминал TEXT по неточности: реальная nova-схема (artefacts_user,
-- artefact_offers) использует unit_id integer.

-- +goose Up

CREATE TABLE exchange_lots (
    id                uuid PRIMARY KEY,
    seller_user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    artifact_unit_id  integer NOT NULL,
    quantity          integer NOT NULL CHECK (quantity > 0 AND quantity <= 100),
    price_oxsarit     bigint  NOT NULL CHECK (price_oxsarit > 0),
    created_at        timestamptz NOT NULL DEFAULT now(),
    expires_at        timestamptz NOT NULL,
    status            text NOT NULL DEFAULT 'active'
                         CHECK (status IN ('active','sold','cancelled','expired')),
    buyer_user_id     uuid REFERENCES users(id) ON DELETE SET NULL,
    sold_at           timestamptz,
    expire_event_id   uuid -- ссылка на events.id для KindExchangeExpire (см. cancel-логику).
);

CREATE INDEX ix_exchange_lots_browse
    ON exchange_lots(status, artifact_unit_id, price_oxsarit)
    WHERE status = 'active';
CREATE INDEX ix_exchange_lots_expire
    ON exchange_lots(status, expires_at)
    WHERE status = 'active';
CREATE INDEX ix_exchange_lots_seller
    ON exchange_lots(seller_user_id, status);

-- exchange_lot_items: какие конкретно артефакты в каком лоте.
-- PRIMARY KEY (lot_id, artefact_id) допускает один artefact_id в нескольких
-- разных lot_id (разные лоты в разное время). Уникальность «артефакт в одном
-- активном лоте» обеспечивается логикой service (SELECT artefact_id FROM
-- artefacts_user WHERE state='held' исключает уже listed-артефакты).
-- Записи остаются после sold/cancelled/expired как audit-trail.
CREATE TABLE exchange_lot_items (
    lot_id        uuid NOT NULL REFERENCES exchange_lots(id) ON DELETE CASCADE,
    artefact_id   uuid NOT NULL REFERENCES artefacts_user(id) ON DELETE RESTRICT,
    PRIMARY KEY (lot_id, artefact_id)
);
CREATE INDEX ix_exchange_lot_items_artefact
    ON exchange_lot_items(artefact_id);

-- exchange_history: журнал событий по лотам (R13 typed payload).
-- event_kind: created|bought|cancelled|expired|banned.
-- actor_user_id: NULL для expired (системное), для banned — модератор/admin.
CREATE TABLE exchange_history (
    id             uuid PRIMARY KEY,
    lot_id         uuid NOT NULL REFERENCES exchange_lots(id) ON DELETE CASCADE,
    event_kind     text NOT NULL
                      CHECK (event_kind IN ('created','bought','cancelled','expired','banned')),
    actor_user_id  uuid REFERENCES users(id) ON DELETE SET NULL,
    payload        jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ix_exchange_history_lot
    ON exchange_history(lot_id, created_at);
CREATE INDEX ix_exchange_history_actor
    ON exchange_history(actor_user_id, created_at);

-- Антифрод-источник для price-cap: AVG(price/quantity) по 'bought' за
-- последние 30 дней по unit_id. Отдельный индекс для быстрой агрегации.
CREATE INDEX ix_exchange_history_pricing
    ON exchange_history(event_kind, created_at)
    WHERE event_kind = 'bought';

-- +goose Down
DROP TABLE IF EXISTS exchange_history;
DROP TABLE IF EXISTS exchange_lot_items;
DROP TABLE IF EXISTS exchange_lots;
