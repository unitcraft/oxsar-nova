-- +goose Up
-- План 56: перенос user_reports из game-nova в portal-backend.
--
-- Жалоба — претензия к глобальному identity-аккаунту, не к конкретной
-- вселенной. Поэтому таблица живёт на портале, который собирает запросы
-- от всех вселенных (game-nova, game-origin, future).
--
-- target_type: 'user' | 'alliance' | 'chat_msg' | 'planet'.
-- target_id хранится как text — для разных типов это user_id (uuid),
-- alliance_id (uuid), chat_messages.id или planet_id.
-- Не FK по разным таблицам — оставляем text + индекс по
-- (target_type, target_id).
--
-- reporter_id / resolved_by — UUID юзеров из identity-DB. **Без FK**:
-- portal не дублирует таблицу users (портал использует identity как
-- источник истины через JWT, как и feedback_posts в 0001_init).
--
-- reason: короткий код категории нарушения; comment — свободный текст
-- от жалующегося, до 1000 символов (валидация на API).
--
-- status: 'new' (по умолчанию) → 'resolved' / 'rejected' (в админке).
-- resolution_note — пометка модератора при резолюции.
--
-- audit-log действий модератора пишется в admin_audit_log (план 14)
-- на стороне сервиса, который применяет санкцию (identity-admin для
-- бана и т. п.); здесь дублировать не нужно.

CREATE TABLE user_reports (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id     TEXT        NOT NULL,
    target_type     TEXT        NOT NULL CHECK (target_type IN ('user','alliance','chat_msg','planet')),
    target_id       TEXT        NOT NULL,
    reason          TEXT        NOT NULL,
    comment         TEXT        NOT NULL DEFAULT '',
    status          TEXT        NOT NULL DEFAULT 'new'
                       CHECK (status IN ('new','resolved','rejected')),
    resolved_by     TEXT,
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

-- +goose Down
DROP TABLE IF EXISTS user_reports;
