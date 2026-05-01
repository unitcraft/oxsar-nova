-- План 72.1.45 §8-9: legacy `exchange` таблица — настройки брокера.
-- Каждый user является потенциальным брокером (legacy создаёт запись
-- при первом обращении). Поля: title (название биржи), comission (=fee%),
-- created_at.
--
-- Используется:
--   1) /p2p-exchange — fee для расчёта Profit (broker_stats.go).
--   2) ?go=ExchangeOpts — admin страница для настройки брокером своей
--      комиссии и заголовка биржи.

-- +goose Up

CREATE TABLE IF NOT EXISTS exchange_settings (
    user_id      uuid           PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    title        text           NOT NULL DEFAULT 'My exchange',
    fee_percent  numeric(5,2)   NOT NULL DEFAULT 5.00 CHECK (fee_percent >= 0 AND fee_percent <= 50),
    created_at   timestamptz    NOT NULL DEFAULT now(),
    updated_at   timestamptz    NOT NULL DEFAULT now()
);

-- +goose Down

DROP TABLE IF EXISTS exchange_settings;
