-- +goose Up
CREATE TABLE alliances (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tag         text NOT NULL UNIQUE,           -- короткое имя [3..5 символов]
    name        text NOT NULL UNIQUE,
    description text NOT NULL DEFAULT '',
    owner_id    uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ix_alliances_owner ON alliances(owner_id);

CREATE TABLE alliance_members (
    alliance_id uuid NOT NULL REFERENCES alliances(id) ON DELETE CASCADE,
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rank        text NOT NULL DEFAULT 'member', -- owner | member
    joined_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id)                       -- один игрок — один альянс
);
CREATE INDEX ix_alliance_members_ally ON alliance_members(alliance_id);

-- alliance_id на users (денормализация для быстрого поиска).
ALTER TABLE users ADD COLUMN alliance_id uuid REFERENCES alliances(id) ON DELETE SET NULL;
CREATE INDEX ix_users_alliance ON users(alliance_id) WHERE alliance_id IS NOT NULL;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS alliance_id;
DROP TABLE IF EXISTS alliance_members;
DROP TABLE IF EXISTS alliances;
