-- +goose Up
-- Поле users.exchange_rate (default 1.2) хранит персональный
-- модификатор курса биржи от активного ARTEFACT_MERCHANTS_MARK (300).
-- Это НЕ курс биржи как таковой — базовый курс берётся из
-- configs/market.yml. См. §5.10.1 oxsar-spec.txt.
--
-- Забыли добавить в 0001_init.sql в предыдущей итерации. Append-only
-- миграции (§17.5) требуют отдельного файла, а не правки 0001.
ALTER TABLE users
    ADD COLUMN exchange_rate numeric(10, 4) NOT NULL DEFAULT 1.2;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS exchange_rate;
