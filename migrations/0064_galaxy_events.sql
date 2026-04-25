-- +goose Up
-- План 17 F: Галактические события.
-- MVP: одно событие активно за раз. Создаётся админом через /api/admin
-- (авто-планирование — отложено в v1.x).
--
-- kind влияет на расчёты в коде:
--   'meteor_storm'    — +30% metal производство
--   'solar_flare'     — -20% energy (опционально, не реализовано в MVP)
--   'trade_forum'     — изменение рыночных курсов (опционально)
--   'star_nebula'     — +15% к exp_power (опционально)
CREATE TABLE galaxy_events (
    id         SERIAL PRIMARY KEY,
    kind       TEXT        NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ends_at    TIMESTAMPTZ NOT NULL,
    params     JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Активное событие = ends_at > now(). Индекс на быстрый запрос.
CREATE INDEX ix_galaxy_events_active ON galaxy_events(ends_at) WHERE ends_at > now();

-- +goose Down
DROP TABLE galaxy_events;
