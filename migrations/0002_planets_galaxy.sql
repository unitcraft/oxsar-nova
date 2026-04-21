-- +goose Up
CREATE TABLE planets (
    id                     uuid PRIMARY KEY,
    user_id                uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_moon                boolean NOT NULL DEFAULT false,
    name                   text NOT NULL,
    galaxy                 integer NOT NULL,
    system                 integer NOT NULL,
    position               integer NOT NULL,
    diameter               integer NOT NULL,
    used_fields            integer NOT NULL DEFAULT 0,
    temperature_min        integer NOT NULL,
    temperature_max        integer NOT NULL,
    metal                  numeric(20, 4) NOT NULL DEFAULT 500,
    silicon                numeric(20, 4) NOT NULL DEFAULT 500,
    hydrogen               numeric(20, 4) NOT NULL DEFAULT 0,
    last_res_update        timestamptz    NOT NULL DEFAULT now(),
    solar_satellite_prod   integer        NOT NULL DEFAULT 100,
    build_factor           real           NOT NULL DEFAULT 1,
    research_factor        real           NOT NULL DEFAULT 1,
    produce_factor         real           NOT NULL DEFAULT 1,
    energy_factor          real           NOT NULL DEFAULT 1,
    storage_factor         real           NOT NULL DEFAULT 1,
    created_at             timestamptz NOT NULL DEFAULT now(),
    destroyed_at           timestamptz,
    UNIQUE (galaxy, system, position, is_moon),
    CONSTRAINT coords_range CHECK (
        galaxy BETWEEN 1 AND 16
        AND system BETWEEN 1 AND 999
        AND position BETWEEN 1 AND 15
    )
);

CREATE INDEX ix_planets_user ON planets(user_id) WHERE destroyed_at IS NULL;

ALTER TABLE users
    ADD CONSTRAINT fk_users_cur_planet FOREIGN KEY (cur_planet_id) REFERENCES planets(id) ON DELETE SET NULL;

CREATE TABLE galaxy_cells (
    galaxy     integer NOT NULL,
    system     integer NOT NULL,
    position   integer NOT NULL,
    planet_id  uuid REFERENCES planets(id) ON DELETE SET NULL,
    debris_metal   numeric(20, 0) NOT NULL DEFAULT 0,
    debris_silicon numeric(20, 0) NOT NULL DEFAULT 0,
    PRIMARY KEY (galaxy, system, position)
);

-- +goose Down
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_cur_planet;
DROP TABLE IF EXISTS galaxy_cells;
DROP TABLE IF EXISTS planets;
