-- +goose Up
-- План 06.1: Автопилот.
--
-- Расширяет ai_advisor_log полями для rule-based автопилота.
-- LLM-записи (source='llm') продолжают писаться как раньше.
--
-- source        : 'llm' | 'autopilot'
-- strategy      : economy|military|defense|expansion|auto (только для autopilot)
-- action_params : JSONB с топ-3 рекомендациями (структура AutopilotResult в Go)
-- event_id      : uuid игрового события, созданного при Execute (build/research/...)
-- status        : pending|ready|executed|expired
ALTER TABLE ai_advisor_log
    ADD COLUMN IF NOT EXISTS source        TEXT NOT NULL DEFAULT 'llm',
    ADD COLUMN IF NOT EXISTS strategy      TEXT,
    ADD COLUMN IF NOT EXISTS action_params JSONB,
    ADD COLUMN IF NOT EXISTS event_id      uuid REFERENCES events(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS status        TEXT NOT NULL DEFAULT 'ready';

-- LLM-записи финализированы при создании, поэтому status='ready' для них нейтрален.
-- Для autopilot-записей сервис явно ставит 'pending' при Enqueue и
-- двигает по жизненному циклу (pending → ready → executed | expired).

CREATE INDEX IF NOT EXISTS idx_ai_advisor_log_user_status
    ON ai_advisor_log(user_id, status)
    WHERE source = 'autopilot';

-- +goose Down
DROP INDEX IF EXISTS idx_ai_advisor_log_user_status;
ALTER TABLE ai_advisor_log
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS event_id,
    DROP COLUMN IF EXISTS action_params,
    DROP COLUMN IF EXISTS strategy,
    DROP COLUMN IF EXISTS source;
