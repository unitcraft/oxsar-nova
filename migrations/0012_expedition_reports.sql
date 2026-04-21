-- +goose Up
-- expedition_reports — результаты миссии EXPEDITION (kind=15).
--
-- Один отчёт на одну экспедицию. outcome — тип события:
--   resources | artefact | pirates | loss | nothing
-- Полный контекст в report JSONB (конкретные ресурсы, id артефакта,
-- детали боя, процент потерь и т.п.). UI рисует секции по outcome.

CREATE TABLE expedition_reports (
    id         uuid PRIMARY KEY,
    user_id    uuid REFERENCES users(id) ON DELETE SET NULL,
    fleet_id   uuid,
    outcome    text NOT NULL,
    report     jsonb NOT NULL,
    at         timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ix_expedition_reports_user ON expedition_reports(user_id, at DESC);

ALTER TABLE messages ADD COLUMN IF NOT EXISTS expedition_report_id uuid
    REFERENCES expedition_reports(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE messages DROP COLUMN IF EXISTS expedition_report_id;
DROP TABLE IF EXISTS expedition_reports;
