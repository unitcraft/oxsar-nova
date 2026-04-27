-- +goose Up
-- План 56: user_reports перенесена в portal-backend (миграция
-- portal/migrations/0002_user_reports.sql). game-nova больше не
-- владеет таблицей жалоб — это глобальный платформенный реестр,
-- собирающий запросы от всех вселенных через POST /api/reports
-- portal'а.
--
-- В dev/prod: данные из game-nova-БД при необходимости перенести
-- скриптом scripts/migrate-reports-game-nova-to-portal.sh ДО
-- выполнения этой миграции (план 56 Ф.4). В dev обычно пусто —
-- миграция выполняется без потерь.

DROP TABLE IF EXISTS user_reports;

-- +goose Down
-- Восстановление точно по схеме 0069_user_reports.sql.
CREATE TABLE user_reports (
    id              UUID PRIMARY KEY,
    reporter_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type     TEXT NOT NULL CHECK (target_type IN ('user','alliance','chat_msg','planet')),
    target_id       TEXT NOT NULL,
    reason          TEXT NOT NULL,
    comment         TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'new'
                       CHECK (status IN ('new','resolved','rejected')),
    resolved_by     UUID REFERENCES users(id) ON DELETE SET NULL,
    resolution_note TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at     TIMESTAMPTZ
);

CREATE INDEX idx_user_reports_status_created
    ON user_reports (status, created_at DESC);
CREATE INDEX idx_user_reports_target
    ON user_reports (target_type, target_id);
CREATE INDEX idx_user_reports_reporter
    ON user_reports (reporter_id, created_at DESC);
