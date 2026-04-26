-- +goose Up
-- Artefact Market: продажа артефактов за credit.
--
-- credit — отдельная валюта (legacy: EXT_MODE=true для market). Не
-- путать с metal/silicon/hydrogen — на них артефакты не покупаются.
-- Стартовый баланс 0; зарабатывается продажей артефактов.
--
-- artefact_offers — активные листинги. При продаже артефакт переходит
-- в state='listed' и НЕ доступен для активации. При покупке —
-- удаление оффера + перевод владельца + state='held'.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS credit bigint NOT NULL DEFAULT 0;

-- +goose StatementBegin
ALTER TYPE artefact_state ADD VALUE IF NOT EXISTS 'listed';
-- +goose StatementEnd

CREATE TABLE artefact_offers (
    id              uuid PRIMARY KEY,
    artefact_id     uuid NOT NULL REFERENCES artefacts_user(id) ON DELETE CASCADE,
    seller_user_id  uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    unit_id         integer NOT NULL,
    price_credit    bigint  NOT NULL CHECK (price_credit > 0),
    listed_at       timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ix_artefact_offers_unit ON artefact_offers(unit_id, price_credit);
CREATE INDEX ix_artefact_offers_seller ON artefact_offers(seller_user_id);

-- +goose Down
DROP TABLE IF EXISTS artefact_offers;
-- artefact_state enum: добавленное значение 'listed' оставляем —
-- в postgres нельзя удалить enum-значение без пересоздания типа.
ALTER TABLE users DROP COLUMN IF EXISTS credit;
