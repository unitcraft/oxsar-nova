-- План 72.1.10 ч.A1: добавление колонок-флагов в battle_reports для
-- фильтров /battlestats (has_aliens / moon_created / is_moon).
--
-- Контекст: legacy `Battlestats.class.php::showBattles` принимает
-- 12 параметров фильтрации, в том числе show_aliens / new_moon /
-- moon_battle. У нас в `battle_reports` нет этих признаков на
-- уровне колонок — только в JSON `report`. Для эффективной фильтрации
-- (партициальные индексы) выносим их в отдельные колонки.
--
-- Backfill: для существующих записей колонки = false (default).
-- Это допустимо: старые записи просто не попадут под фильтр по этим
-- условиям (которое и без того опционально, дефолт false).
--
-- Подробности — `docs/plans/72.1.10-battlestats-filters.md`.

-- +goose Up

ALTER TABLE battle_reports ADD COLUMN IF NOT EXISTS
    has_aliens boolean NOT NULL DEFAULT false;
ALTER TABLE battle_reports ADD COLUMN IF NOT EXISTS
    moon_created boolean NOT NULL DEFAULT false;
ALTER TABLE battle_reports ADD COLUMN IF NOT EXISTS
    is_moon boolean NOT NULL DEFAULT false;

-- Partial-индексы — только для true-записей (редкие события).
CREATE INDEX IF NOT EXISTS ix_battle_reports_has_aliens
    ON battle_reports (has_aliens) WHERE has_aliens = true;
CREATE INDEX IF NOT EXISTS ix_battle_reports_moon_created
    ON battle_reports (moon_created) WHERE moon_created = true;
CREATE INDEX IF NOT EXISTS ix_battle_reports_is_moon
    ON battle_reports (is_moon) WHERE is_moon = true;

-- +goose Down

DROP INDEX IF EXISTS ix_battle_reports_is_moon;
DROP INDEX IF EXISTS ix_battle_reports_moon_created;
DROP INDEX IF EXISTS ix_battle_reports_has_aliens;
ALTER TABLE battle_reports DROP COLUMN IF EXISTS is_moon;
ALTER TABLE battle_reports DROP COLUMN IF EXISTS moon_created;
ALTER TABLE battle_reports DROP COLUMN IF EXISTS has_aliens;
