-- План 72.1.1 «Зачисление опыта и потерь юзеру после реального боя»:
-- порт `oxsar2-java/Assault.jar` Participant.java:924-987.
--
-- Что делает legacy после реального боя (НЕ симулятор):
--   1. Инкрементит у каждого участника:
--        users.e_points  += battleExperience  (суммарный опыт)
--        users.be_points += battleExperience  (резервуар для апгрейда тех)
--        users.battles   += 1                 (счётчик боёв)
--   2. Декрементит:
--        users.points    -= lostPoints        (общий рейтинг)
--        users.u_points  -= lostPoints        (рейтинг по юнитам)
--        users.u_count   -= lostUnits         (общее число юнитов)
--   3. Пишет лог-запись в `na_user_experience (time, isatter, userid,
--      experience, assaultid)`.
--
-- Что добавляем этой миграцией:
--   - `users.u_count` — счётчик общего числа юнитов игрока (по легаси
--     присутствует, в game-nova отсутствовал).
--   - таблица `user_experience` — лог опыта (полностью новая).
--
-- Idempotency: UNIQUE (battle_id, user_id, is_atter) гарантирует, что
-- повторная обработка одного и того же события не зачислит опыт
-- дважды (`ON CONFLICT DO NOTHING` в ApplyBattleResult).
--
-- Подробности см. `docs/plans/72.1.1-battle-user-stats.md`.

-- +goose Up

ALTER TABLE users ADD COLUMN IF NOT EXISTS u_count integer NOT NULL DEFAULT 0;
ALTER TABLE users ADD CONSTRAINT users_u_count_nonneg CHECK (u_count >= 0);

CREATE TABLE IF NOT EXISTS user_experience (
    id          bigserial   PRIMARY KEY,
    at          timestamptz NOT NULL DEFAULT now(),
    is_atter    boolean     NOT NULL,
    user_id     uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    experience  integer     NOT NULL,
    battle_id   uuid        REFERENCES battle_reports(id) ON DELETE SET NULL,
    UNIQUE (battle_id, user_id, is_atter)
);

CREATE INDEX IF NOT EXISTS ix_user_experience_user_id
    ON user_experience (user_id, at DESC);
CREATE INDEX IF NOT EXISTS ix_user_experience_battle_id
    ON user_experience (battle_id);

-- +goose Down

DROP TABLE IF EXISTS user_experience;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_u_count_nonneg;
ALTER TABLE users DROP COLUMN IF EXISTS u_count;
