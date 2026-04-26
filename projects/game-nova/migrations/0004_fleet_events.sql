-- +goose Up
CREATE TABLE fleets (
    id               uuid PRIMARY KEY,
    owner_user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    src_planet_id    uuid REFERENCES planets(id) ON DELETE SET NULL,
    dst_galaxy       integer NOT NULL,
    dst_system       integer NOT NULL,
    dst_position     integer NOT NULL,
    dst_is_moon      boolean NOT NULL DEFAULT false,
    mission          integer NOT NULL,
    state            text    NOT NULL DEFAULT 'outbound',  -- outbound|hold|returning|done
    depart_at        timestamptz NOT NULL,
    arrive_at        timestamptz NOT NULL,
    return_at        timestamptz,
    hold_seconds     integer NOT NULL DEFAULT 0,
    carried_metal    numeric(20, 0) NOT NULL DEFAULT 0,
    carried_silicon  numeric(20, 0) NOT NULL DEFAULT 0,
    carried_hydrogen numeric(20, 0) NOT NULL DEFAULT 0,
    speed_percent    integer NOT NULL DEFAULT 100,
    created_at       timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE fleet_ships (
    fleet_id       uuid    NOT NULL REFERENCES fleets(id) ON DELETE CASCADE,
    unit_id        integer NOT NULL,
    count          bigint  NOT NULL,
    damaged_count  bigint  NOT NULL DEFAULT 0,
    PRIMARY KEY (fleet_id, unit_id)
);

CREATE TABLE events (
    id          uuid PRIMARY KEY,
    user_id     uuid,
    planet_id   uuid,
    kind        integer NOT NULL,
    state       event_state NOT NULL DEFAULT 'wait',
    fire_at     timestamptz NOT NULL,
    payload     jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at  timestamptz NOT NULL DEFAULT now(),
    processed_at timestamptz
);

CREATE INDEX ix_events_fire_at ON events(fire_at) WHERE state = 'wait';

CREATE TABLE res_log (
    id              bigserial PRIMARY KEY,
    user_id         uuid NOT NULL,
    planet_id       uuid,
    reason          text NOT NULL,      -- build|research|fleet_cost|loot|debris|transfer|market|admin_gift|refund|tick
    delta_metal     numeric(20, 4) NOT NULL DEFAULT 0,
    delta_silicon   numeric(20, 4) NOT NULL DEFAULT 0,
    delta_hydrogen  numeric(20, 4) NOT NULL DEFAULT 0,
    at              timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ix_res_log_user_at ON res_log(user_id, at DESC);

-- +goose Down
DROP TABLE IF EXISTS res_log;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS fleet_ships;
DROP TABLE IF EXISTS fleets;
