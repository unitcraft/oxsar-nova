-- +goose Up
-- espionage_reports — результат SPY-миссии (kind=11, event_spy в legacy).
--
-- По OGame-правилам видимая информация зависит от ratio = probes +
-- spy_self - spy_target:
--    >=1 — ресурсы
--    >=2 — ships
--    >=4 — defense
--    >=6 — buildings
--
-- report — JSONB c тем, что удалось увидеть; UI отображает только
-- непустые секции. Форма может расширяться без миграций (как в
-- battle_reports).

CREATE TABLE espionage_reports (
    id             uuid PRIMARY KEY,
    spy_user_id    uuid REFERENCES users(id) ON DELETE SET NULL,
    target_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    planet_id      uuid REFERENCES planets(id) ON DELETE SET NULL,
    ratio          integer NOT NULL,
    probes         integer NOT NULL,
    report         jsonb   NOT NULL,
    at             timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ix_espionage_reports_spy ON espionage_reports(spy_user_id, at DESC);

ALTER TABLE messages ADD COLUMN IF NOT EXISTS espionage_report_id uuid
    REFERENCES espionage_reports(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE messages DROP COLUMN IF EXISTS espionage_report_id;
DROP TABLE IF EXISTS espionage_reports;
