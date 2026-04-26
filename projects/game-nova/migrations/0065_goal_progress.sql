-- +goose Up
-- План 30 Ф.1: Goal Engine — единый движок целей вместо двух систем
-- (achievement + dailyquest).
--
-- Определения целей — в configs/goals.yml (как остальной content проекта;
-- см. план 28). БД хранит только пользовательский state.
--
-- goal_key — string-ссылка на YAML (UUID/serial-id не нужен, ключ из
-- YAML стабилен). Если goal удалена из YAML — записи в goal_progress
-- остаются как «история» (load YAML логирует sirop-ключи как WARN).
--
-- period_key (определяется по lifecycle цели):
--   ''           — permanent / one-time / seasonal
--   'YYYY-MM-DD' — daily   (UTC)
--   'YYYY-Www'   — weekly  (ISO week, UTC)

-- ───────────────────────────── goal_progress ─────────────────────────

CREATE TABLE goal_progress (
    user_id      UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    goal_key     TEXT         NOT NULL,
    period_key   TEXT         NOT NULL DEFAULT '',
    progress     INTEGER      NOT NULL DEFAULT 0 CHECK (progress >= 0),
    completed_at TIMESTAMPTZ  NULL,
    claimed_at   TIMESTAMPTZ  NULL,
    seen_at      TIMESTAMPTZ  NULL,

    PRIMARY KEY (user_id, goal_key, period_key),
    -- claimed_at без completed_at невозможен.
    CHECK (claimed_at IS NULL OR completed_at IS NOT NULL)
);

-- Для UI List: список целей пользователя за период.
CREATE INDEX ix_goal_progress_user ON goal_progress(user_id, period_key);
-- Для toast/badge: новые завершённые но непросмотренные.
CREATE INDEX ix_goal_progress_unseen
    ON goal_progress(user_id)
    WHERE completed_at IS NOT NULL AND seen_at IS NULL;

-- ───────────────────────────── goal_rewards_log ──────────────────────
-- Аудит-лог выданных наград (для UI «история наград» и debug).
-- reward — JSONB-snapshot выданного на момент claim (на случай если
-- YAML потом изменится, история игрока не теряется).

CREATE TABLE goal_rewards_log (
    id          UUID         PRIMARY KEY,
    user_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    goal_key    TEXT         NOT NULL,
    period_key  TEXT         NOT NULL DEFAULT '',
    granted_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    reward      JSONB        NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX ix_goal_rewards_log_user
    ON goal_rewards_log(user_id, granted_at DESC);

-- +goose Down
DROP TABLE IF EXISTS goal_rewards_log;
DROP TABLE IF EXISTS goal_progress;
