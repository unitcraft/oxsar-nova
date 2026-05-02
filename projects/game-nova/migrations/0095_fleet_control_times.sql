-- План 72.1.48 (доделка): control_times rate-limit + back_consumption
-- check для load/unload (legacy `getRemainFleetControls` +
-- `back_consumption` в `Mission.class.php`).
--
-- Хранение в fleets-таблице (а не в events.payload), потому что:
--   1. nova events.payload immutable — мутация сломала бы audit trail.
--   2. Это атрибут флота, не события — каждый flow load/unload
--      инкрементит счётчик per-fleet.

-- +goose Up

ALTER TABLE fleets
    ADD COLUMN IF NOT EXISTS control_times       integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS max_control_times   integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS back_consumption    bigint  NOT NULL DEFAULT 0;

COMMENT ON COLUMN fleets.control_times IS
    '72.1.48: количество совершённых load/unload операций (legacy control_times)';
COMMENT ON COLUMN fleets.max_control_times IS
    '72.1.48: лимит = 1 + floor(comp_tech_owner/6) [+ floor(comp_tech_loc/6) если на чужой]';
COMMENT ON COLUMN fleets.back_consumption IS
    '72.1.48: H зарезервированный на возврат, нельзя выгрузить ниже этой границы';

-- +goose Down

ALTER TABLE fleets
    DROP COLUMN IF EXISTS back_consumption,
    DROP COLUMN IF EXISTS max_control_times,
    DROP COLUMN IF EXISTS control_times;
