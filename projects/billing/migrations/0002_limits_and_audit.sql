-- План 54 Ф.1: лимит самозанятого (ФЗ-422 НПД, 2.4 млн ₽/год) +
-- audit-log + per-year alert state.
--
-- Принципы:
-- * billing_system_state — singleton-row с глобальным kill-switch
--   payments_active. Меняется автоматически reconciler-job'ом
--   (auto-disabled при HARD_STOP) или вручную через admin API
--   (manual override). После manual enable — снова автоматический
--   контроль; refund'ы не сбрасывают флаг авто (чтобы не было
--   yo-yo) — сбрасывает только admin override.
-- * billing_alert_state per-year: какие пороги (80/90/95/hard) уже
--   сработали в этом году. Новый год (1 января МСК) — новая строка
--   с NULL по всем порогам, alerts шлются заново.
-- * billing_audit_log — INSERT-only журнал админских действий.

-- +goose Up

-- Глобальный state: payments_active singleton.
-- Альтернатива key-value (TEXT key, JSONB value) была отвергнута: для
-- ровно двух булевых полей и нескольких метаданных типизированные
-- колонки удобнее (атомарный SELECT без JSON-парсинга, явные индексы
-- и FK к UUID).
CREATE TABLE billing_system_state (
    id                  SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    payments_active     BOOLEAN NOT NULL DEFAULT TRUE,
    last_changed_by     UUID,                  -- NULL при system-init
    last_changed_at     TIMESTAMPTZ,
    last_change_reason  TEXT,
    auto_disabled_at    TIMESTAMPTZ            -- когда reconciler выключил
);
INSERT INTO billing_system_state (id, payments_active) VALUES (1, true)
ON CONFLICT DO NOTHING;

-- Per-year alert state: какие пороги уже сработали в этом году.
-- Reconciler INSERT'ит строку при первом alert'е года.
-- Используется для anti-spam: один email/log per (year, threshold).
CREATE TABLE billing_alert_state (
    year                INTEGER PRIMARY KEY,
    threshold_80_sent   TIMESTAMPTZ,
    threshold_90_sent   TIMESTAMPTZ,
    threshold_95_sent   TIMESTAMPTZ,
    threshold_hard_sent TIMESTAMPTZ
);

-- Audit log админских действий в billing-сервисе. INSERT-only.
CREATE TABLE billing_audit_log (
    id           BIGSERIAL PRIMARY KEY,
    actor_id     UUID NOT NULL,
    action       VARCHAR(64) NOT NULL,   -- 'limit:enable' | 'limit:disable' | …
    target_type  VARCHAR(32),            -- 'system' | 'payment' | NULL
    target_id    TEXT,                   -- 'limit' / order_id / NULL
    payload      JSONB,                  -- произвольные детали
    reason       TEXT NOT NULL,          -- обязателен для всех действий
    ip_address   INET,
    user_agent   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ix_billing_audit_log_created ON billing_audit_log(created_at DESC);
CREATE INDEX ix_billing_audit_log_actor   ON billing_audit_log(actor_id);
CREATE INDEX ix_billing_audit_log_action  ON billing_audit_log(action);
CREATE INDEX ix_billing_audit_log_target  ON billing_audit_log(target_type, target_id)
    WHERE target_id IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS billing_audit_log;
DROP TABLE IF EXISTS billing_alert_state;
DROP TABLE IF EXISTS billing_system_state;
