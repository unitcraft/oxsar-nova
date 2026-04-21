-- +goose Up
-- Alliance join approval: is_open flag + applications table.
-- is_open=true  → direct join (legacy behaviour, default).
-- is_open=false → Join() creates application; owner approves/rejects.

ALTER TABLE alliances ADD COLUMN is_open boolean NOT NULL DEFAULT true;

CREATE TABLE alliance_applications (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    alliance_id  uuid        NOT NULL REFERENCES alliances(id) ON DELETE CASCADE,
    user_id      uuid        NOT NULL REFERENCES users(id)     ON DELETE CASCADE,
    message      text        NOT NULL DEFAULT '',
    created_at   timestamptz NOT NULL DEFAULT now(),
    UNIQUE (alliance_id, user_id)
);
CREATE INDEX ix_alliance_apps_alliance ON alliance_applications(alliance_id);

-- +goose Down
DROP TABLE IF EXISTS alliance_applications;
ALTER TABLE alliances DROP COLUMN IF EXISTS is_open;
