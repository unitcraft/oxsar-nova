-- +goose Up
-- Отношения между альянсами: NAP (ненападение), WAR (война), ALLY (союз).
CREATE TYPE alliance_relation AS ENUM ('nap', 'war', 'ally');

CREATE TABLE alliance_relationships (
    alliance_id        uuid NOT NULL REFERENCES alliances(id) ON DELETE CASCADE,
    target_alliance_id uuid NOT NULL REFERENCES alliances(id) ON DELETE CASCADE,
    relation           alliance_relation NOT NULL,
    set_at             timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (alliance_id, target_alliance_id),
    CHECK (alliance_id <> target_alliance_id)
);

-- +goose Down
DROP TABLE IF EXISTS alliance_relationships;
DROP TYPE IF EXISTS alliance_relation;
