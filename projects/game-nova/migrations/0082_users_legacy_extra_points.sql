-- План 72.1 ч.17: восстановление полей `be_points`, `of_level`, `of_points`,
-- `dm_points`, ранее отвергнутых планом 69 Ф.0 (D-001 расширенное) как YAGNI.
--
-- Решение об отмене D-001-отказа принято пользователем 2026-04-29 в рамках
-- pixel-perfect-серии (главный экран legacy-PHP `main.tpl` показывает все
-- эти поля). Detail в `docs/plans/72.1-post-remaster-stabilization.md` ч.17.
--
-- Семантика (legacy):
--   - `be_points` — резервуар «активных уровней технологий» для боя.
--     Тратится при отправке атаки, возвращается при завершении события.
--     Источник: `Mission.class.php:1675`, `EventHandler:363/1156`.
--   - `of_points` + `of_level` — система уровня профессии Шахтёр.
--     `of_points` накапливается при добыче ресурсов; `of_level`
--     повышается автоматически при достижении порога
--     `need_points = round(pow(1.5, level-1) * 200)` (legacy:
--     `Functions.inc.php:1642-1665`). Эффект: бонус к скорости добычи.
--   - `dm_points` — derived-метрика для альтернативного рейтинга
--     (legacy: `Functions.inc.php:1432-1434`). Считается воркером в
--     `score.RecalcUser` по формуле:
--     `POW(GREATEST(LEAST(e_points,100), LEAST(e_points,
--      POW(points/4000,1.1)+e_points/100)) * points /
--      POW(GREATEST(1,max_points),0.9), 0.5) * 100`.
--
-- Все поля nullable-safe (NOT NULL DEFAULT 0). Backfill не требуется —
-- значения начнут накапливаться от текущей БД-state.

-- +goose Up

ALTER TABLE users ADD COLUMN IF NOT EXISTS be_points numeric(20, 4) NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS of_points numeric(20, 4) NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS of_level  integer        NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS dm_points numeric(20, 4) NOT NULL DEFAULT 0;

-- be_points/of_level не должны уйти в минус (sanity check).
ALTER TABLE users ADD CONSTRAINT users_be_points_nonneg CHECK (be_points >= 0);
ALTER TABLE users ADD CONSTRAINT users_of_points_nonneg CHECK (of_points >= 0);
ALTER TABLE users ADD CONSTRAINT users_of_level_nonneg  CHECK (of_level  >= 0);
ALTER TABLE users ADD CONSTRAINT users_dm_points_nonneg CHECK (dm_points >= 0);

-- Индекс для альтернативного рейтинга по dm_points (если будем выводить
-- топ по dm). Partial — только активные игроки в highscore.
CREATE INDEX IF NOT EXISTS users_dm_points_idx
  ON users (dm_points DESC)
  WHERE umode = false AND is_observer = false;

-- +goose Down

DROP INDEX IF EXISTS users_dm_points_idx;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_dm_points_nonneg;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_of_level_nonneg;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_of_points_nonneg;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_be_points_nonneg;
ALTER TABLE users DROP COLUMN IF EXISTS dm_points;
ALTER TABLE users DROP COLUMN IF EXISTS of_level;
ALTER TABLE users DROP COLUMN IF EXISTS of_points;
ALTER TABLE users DROP COLUMN IF EXISTS be_points;
