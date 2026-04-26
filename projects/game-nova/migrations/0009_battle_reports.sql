-- +goose Up
-- battle_reports — результат боя (§5.7 ТЗ, §12.1 боевые отчёты).
--
-- Хранится как JSONB для удобства версионирования structure'ы:
-- новые поля report'а (ballistics-trace, ablation-events и т.п.) не
-- требуют миграции, UI просто читает отсутствующие поля как null.
--
-- Ссылка двусторонняя: messages.battle_report_id → battle_reports.id.
-- При удалении отчёта (редкий случай — админом) соответствующие
-- сообщения теряют link, но остаются читаемыми.

CREATE TABLE battle_reports (
    id              uuid PRIMARY KEY,
    attacker_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    defender_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    planet_id       uuid REFERENCES planets(id) ON DELETE SET NULL,
    seed            bigint NOT NULL,
    winner          text   NOT NULL,     -- attackers | defenders | draw
    rounds          integer NOT NULL,
    debris_metal    numeric(20, 0) NOT NULL DEFAULT 0,
    debris_silicon  numeric(20, 0) NOT NULL DEFAULT 0,
    loot_metal      numeric(20, 0) NOT NULL DEFAULT 0,
    loot_silicon    numeric(20, 0) NOT NULL DEFAULT 0,
    loot_hydrogen   numeric(20, 0) NOT NULL DEFAULT 0,
    report          jsonb  NOT NULL,     -- полный battle.Report
    at              timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ix_battle_reports_attacker ON battle_reports(attacker_user_id, at DESC);
CREATE INDEX ix_battle_reports_defender ON battle_reports(defender_user_id, at DESC);

ALTER TABLE messages ADD COLUMN IF NOT EXISTS battle_report_id uuid
    REFERENCES battle_reports(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE messages DROP COLUMN IF EXISTS battle_report_id;
DROP TABLE IF EXISTS battle_reports;
