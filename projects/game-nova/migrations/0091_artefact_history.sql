-- План 72.1.45 §2: legacy `ArtefactInfo.class.php` (L.66) показывает
-- хронологию приобретений артефакта в боях. Источник — таблица
-- `artefact_history` (legacy schema), вставка в `Artefact.class.php`
-- L.296 при `awardArtefact` для unique-артефактов.
--
-- В origin/nova все артефакты-трофеи являются «персональными»; история
-- здесь — лог `who got which artefact when (and as a battle trophy or
-- otherwise)`. Поле `battle_report_id` опционально: NULL для не-боевого
-- источника (квест/экспедиция/админ).

-- +goose Up

CREATE TABLE IF NOT EXISTS artefact_history (
    id               uuid           PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_id          int            NOT NULL,
    user_id          uuid           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    battle_report_id uuid           NULL,
    acquired_at      timestamptz    NOT NULL DEFAULT now(),
    source           text           NOT NULL DEFAULT 'battle' CHECK (source IN ('battle','expedition','quest','admin','market'))
);

CREATE INDEX IF NOT EXISTS ix_artefact_history_unit
    ON artefact_history (unit_id, acquired_at DESC);

CREATE INDEX IF NOT EXISTS ix_artefact_history_user
    ON artefact_history (user_id, acquired_at DESC);

-- +goose Down

DROP INDEX IF EXISTS ix_artefact_history_user;
DROP INDEX IF EXISTS ix_artefact_history_unit;
DROP TABLE IF EXISTS artefact_history;
