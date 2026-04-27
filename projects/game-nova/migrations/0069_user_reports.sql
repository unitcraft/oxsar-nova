-- План 46 Ф.3 (149-ФЗ): жалобы пользователей на UGC-нарушения.
--
-- target_type: 'user' | 'alliance' | 'chat_msg' | 'planet'.
-- target_id хранится как text — для разных типов это user_id (uuid),
-- alliance_id (uuid), chat_messages.id (uuid) или planet_id (uuid).
-- Не FK по разным таблицам — оставляем text + индекс по (target_type,target_id).
--
-- reason: короткий код категории нарушения; comment — свободный текст
-- от жалующегося, до 1000 символов (валидация на API).
--
-- status: 'new' (по умолчанию) → 'resolved' / 'rejected' (в админке).
-- resolved_by — id модератора (users.id), resolution_note — что решили.
--
-- audit-log действий модератора уже есть (admin_audit_log, план 14);
-- здесь дублировать не нужно.

-- +goose Up
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

-- +goose Down
DROP TABLE IF EXISTS user_reports;
