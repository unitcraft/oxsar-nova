-- +goose Up
--
-- Артефакты (§5.10 oxsar-spec.txt).
--
-- artefacts_user — инвентарь. Один артефакт = одна строка.
-- При активации артефактов, модифицирующих факторы (§5.10.1),
-- мы меняем поля в users/planets напрямую — так делает oxsar2
-- (Artefact.class.php). Трек активных артефактов нужен для:
--  1) expire-событий (EVENT_ARTEFACT_EXPIRE);
--  2) resyncUser — пересчёта факторов с нуля.

-- +goose StatementBegin
CREATE TYPE artefact_state AS ENUM ('held', 'delayed', 'active', 'expired', 'consumed');
-- +goose StatementEnd

CREATE TABLE artefacts_user (
    id              uuid PRIMARY KEY,
    user_id         uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    planet_id       uuid REFERENCES planets(id) ON DELETE SET NULL,
    unit_id         integer NOT NULL,         -- ARTEFACT_* код из каталога
    state           artefact_state NOT NULL DEFAULT 'held',
    acquired_at     timestamptz NOT NULL DEFAULT now(),
    activated_at    timestamptz,
    expire_at       timestamptz,
    payload         jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX ix_artefacts_user_user   ON artefacts_user(user_id);
CREATE INDEX ix_artefacts_user_active ON artefacts_user(user_id) WHERE state = 'active';
CREATE INDEX ix_artefacts_user_expire ON artefacts_user(expire_at) WHERE state = 'active';

-- +goose Down
DROP TABLE IF EXISTS artefacts_user;
-- +goose StatementBegin
DROP TYPE IF EXISTS artefact_state;
-- +goose StatementEnd
