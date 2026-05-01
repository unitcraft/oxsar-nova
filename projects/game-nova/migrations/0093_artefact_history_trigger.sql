-- План 72.1.46 P1#2: artefact_history trigger.
--
-- В origin/nova существует 8+ мест где артефакты создаются прямым
-- INSERT в artefacts_user (expedition, alien, pack, admin grant,
-- import-legacy, и т.д.). Service.Grant() с записью в artefact_history
-- вызывается мало где, поэтому таблица истории остаётся пустой.
--
-- Решение: trigger AFTER INSERT на artefacts_user пишет запись в
-- artefact_history c source='admin' по умолчанию. Места где известен
-- более точный source (battle/expedition/alien) UPDATE'ят source
-- после INSERT.
--
-- payload JSON может содержать поле `acquisition_source` — триггер
-- использует его как source если указан. Это позволяет вызывающему
-- задать source без двух запросов.

-- +goose Up

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION artefact_history_record() RETURNS trigger AS $$
DECLARE
    src text := 'admin';
BEGIN
    IF NEW.payload IS NOT NULL
       AND NEW.payload ? 'acquisition_source'
    THEN
        src := NEW.payload ->> 'acquisition_source';
        IF src NOT IN ('battle','expedition','quest','admin','market') THEN
            src := 'admin';
        END IF;
    END IF;
    INSERT INTO artefact_history (unit_id, user_id, source, acquired_at, battle_report_id)
    VALUES (
        NEW.unit_id,
        NEW.user_id,
        src,
        COALESCE(NEW.acquired_at, now()),
        NULLIF(NEW.payload ->> 'battle_report_id', '')::uuid
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_artefacts_user_history ON artefacts_user;
CREATE TRIGGER trg_artefacts_user_history
    AFTER INSERT ON artefacts_user
    FOR EACH ROW
    EXECUTE FUNCTION artefact_history_record();

-- +goose Down

DROP TRIGGER IF EXISTS trg_artefacts_user_history ON artefacts_user;
DROP FUNCTION IF EXISTS artefact_history_record();
