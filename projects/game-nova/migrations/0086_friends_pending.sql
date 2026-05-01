-- План 72.1.14: двусторонний accept-flow для friendship.
--
-- Legacy `buddylist (relid, friend1, friend2, accepted)` реализует
-- двустороннюю дружбу с подтверждением. План 11 шаг 5 (2026-04-23)
-- упростил это до однонаправленной модели без `accepted`. План 72.1
-- §20.12 требует строгий функциональный паритет с legacy и закрывает
-- упрощение P72.1.13.FRIENDS_UNIDIRECTIONAL.
--
-- Семантика (legacy):
--   accepted=0 → pending: A отправил, B видит «входящий запрос».
--   accepted=1 → mutual: оба видят друг друга в общем списке друзей.
--
-- Backfill: все существующие записи считаем accepted=true (current
-- production model — без подтверждения; не превращаем живые friendships
-- в pending). Также создаём симметричную пару (B,A,accepted=true) для
-- каждой существующей (A,B), чтобы оба пользователя видели друг друга.

-- +goose Up

ALTER TABLE friends
    ADD COLUMN accepted boolean NOT NULL DEFAULT false;

-- Существующие записи — это «уже друзья» в текущей модели.
UPDATE friends SET accepted = true WHERE accepted = false;

-- Симметричная пара для каждого существующего friendship.
-- В новой модели обе стороны видят друг друга через свою запись
-- (user_id=me); legacy при accept создаёт логически симметричный view.
INSERT INTO friends (user_id, friend_id, created_at, accepted)
SELECT f.friend_id, f.user_id, f.created_at, true
FROM friends f
LEFT JOIN friends g ON g.user_id = f.friend_id AND g.friend_id = f.user_id
WHERE g.user_id IS NULL
ON CONFLICT DO NOTHING;

-- Partial-индекс для выборок «входящие запросы мне» (быстрый счётчик
-- pending в шапке UI и фильтр в /api/friends?pending=incoming).
CREATE INDEX IF NOT EXISTS friends_pending_idx
    ON friends (friend_id)
    WHERE NOT accepted;

-- +goose Down

DROP INDEX IF EXISTS friends_pending_idx;

-- Down НЕ удаляет симметричные пары (это безопасно — текущий handler
-- их игнорирует, в худшем случае дублируется список).
ALTER TABLE friends DROP COLUMN IF EXISTS accepted;
