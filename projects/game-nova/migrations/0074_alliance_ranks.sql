-- План 67 Ф.1: гранулярные права рангов (D-014, U-005).
--
-- Текущая модель: alliance_members.rank TEXT ('owner'|'member') +
-- rank_name TEXT (свободный текст из 0034). Прав нет.
--
-- Целевая модель: таблица alliance_ranks с permissions JSONB
-- (snake_case ключи). alliance_members.rank_id FK (nullable —
-- fallback на builtin alliance_members.rank до настройки рангов
-- лидером альянса).
--
-- Owner всегда имеет все права независимо от rank_id (проверка в
-- middleware Ф.2).

-- +goose Up
CREATE TABLE alliance_ranks (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    alliance_id uuid NOT NULL REFERENCES alliances(id) ON DELETE CASCADE,
    name        text NOT NULL,
    -- position: порядок в UI; меньше = выше. 0 зарезервирован под
    -- builtin "owner-row" если когда-нибудь решим материализовать.
    position    integer NOT NULL DEFAULT 100,
    -- permissions JSONB с булевыми ключами (snake_case по R1):
    --   can_invite, can_kick, can_send_global_mail,
    --   can_manage_diplomacy, can_change_description,
    --   can_propose_relations, can_manage_ranks
    -- Отсутствующий ключ интерпретируется как false (Ф.2).
    permissions jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (alliance_id, name)
);
CREATE INDEX ix_alliance_ranks_alliance ON alliance_ranks(alliance_id, position);

-- FK на ранг для членов. Nullable — пока лидер не создал рангов,
-- работает старая логика (rank='owner'|'member' + rank_name).
ALTER TABLE alliance_members
    ADD COLUMN IF NOT EXISTS rank_id uuid
        REFERENCES alliance_ranks(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS ix_alliance_members_rank ON alliance_members(rank_id)
    WHERE rank_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS ix_alliance_members_rank;
ALTER TABLE alliance_members DROP COLUMN IF EXISTS rank_id;
DROP TABLE IF EXISTS alliance_ranks;
