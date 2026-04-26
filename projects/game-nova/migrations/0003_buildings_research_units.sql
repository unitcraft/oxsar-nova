-- +goose Up
CREATE TABLE buildings (
    planet_id uuid    NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    unit_id   integer NOT NULL,
    level     integer NOT NULL DEFAULT 0,
    PRIMARY KEY (planet_id, unit_id)
);

CREATE TABLE research (
    user_id uuid    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    unit_id integer NOT NULL,
    level   integer NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, unit_id)
);

CREATE TABLE ships (
    planet_id      uuid    NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    unit_id        integer NOT NULL,
    count          bigint  NOT NULL DEFAULT 0,
    damaged_count  bigint  NOT NULL DEFAULT 0,
    shell_percent  real    NOT NULL DEFAULT 0,
    PRIMARY KEY (planet_id, unit_id)
);

CREATE TABLE defense (
    planet_id uuid    NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    unit_id   integer NOT NULL,
    count     bigint  NOT NULL DEFAULT 0,
    PRIMARY KEY (planet_id, unit_id)
);

CREATE TABLE construction_queue (
    id              uuid PRIMARY KEY,
    planet_id       uuid NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    unit_id         integer NOT NULL,
    unit_type       text    NOT NULL,       -- building | moon_building | research
    target_level    integer NOT NULL,
    start_at        timestamptz NOT NULL,
    end_at          timestamptz NOT NULL,
    cost_metal      numeric(20, 0) NOT NULL,
    cost_silicon    numeric(20, 0) NOT NULL,
    cost_hydrogen   numeric(20, 0) NOT NULL,
    status          queue_status NOT NULL DEFAULT 'queued',
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ix_construction_queue_planet ON construction_queue(planet_id) WHERE status IN ('queued','running');

CREATE TABLE shipyard_queue (
    id                 uuid PRIMARY KEY,
    planet_id          uuid NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    unit_id            integer NOT NULL,
    count              bigint  NOT NULL,
    per_unit_seconds   integer NOT NULL,
    start_at           timestamptz NOT NULL,
    end_at             timestamptz NOT NULL,
    status             queue_status NOT NULL DEFAULT 'queued',
    created_at         timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ix_shipyard_queue_planet ON shipyard_queue(planet_id) WHERE status IN ('queued','running');

-- +goose Down
DROP TABLE IF EXISTS shipyard_queue;
DROP TABLE IF EXISTS construction_queue;
DROP TABLE IF EXISTS defense;
DROP TABLE IF EXISTS ships;
DROP TABLE IF EXISTS research;
DROP TABLE IF EXISTS buildings;
