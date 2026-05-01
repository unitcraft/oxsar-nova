-- План 72.1.48: Formation — расширение acs_groups (см. 0021_acs.sql)
-- именем + лидером + опциональным leader_event_id, плюс инвитейшен-табличка.
--
-- Legacy схема (`Mission.class.php::formation`):
--   * attack_formation (eventid PK, name, time)
--   * formation_invitation (eventid, userid)
--
-- Nova:
--   * acs_groups уже создан 0021_acs.sql с target+arrive_at; добавляем
--     name, leader_user_id, leader_fleet_id, leader_event_id.
--   * acs_invitations (acs_group_id, user_id) — pending до accepted_at.

-- +goose Up

ALTER TABLE acs_groups
    ADD COLUMN IF NOT EXISTS name            text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS leader_user_id  uuid NULL REFERENCES users(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS leader_fleet_id uuid NULL,
    ADD COLUMN IF NOT EXISTS leader_event_id uuid NULL;

CREATE INDEX IF NOT EXISTS ix_acs_groups_leader ON acs_groups(leader_user_id);

CREATE TABLE IF NOT EXISTS acs_invitations (
    acs_group_id  uuid        NOT NULL REFERENCES acs_groups(id) ON DELETE CASCADE,
    user_id       uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invited_by    uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invited_at    timestamptz NOT NULL DEFAULT now(),
    accepted_at   timestamptz NULL,

    PRIMARY KEY (acs_group_id, user_id)
);

CREATE INDEX IF NOT EXISTS ix_acs_invitations_user
    ON acs_invitations(user_id) WHERE accepted_at IS NULL;

-- +goose Down

DROP INDEX IF EXISTS ix_acs_invitations_user;
DROP TABLE IF EXISTS acs_invitations;
DROP INDEX IF EXISTS ix_acs_groups_leader;
ALTER TABLE acs_groups
    DROP COLUMN IF EXISTS leader_event_id,
    DROP COLUMN IF EXISTS leader_fleet_id,
    DROP COLUMN IF EXISTS leader_user_id,
    DROP COLUMN IF EXISTS name;
