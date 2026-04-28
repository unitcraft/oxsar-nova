-- План 69 пост-фикс (R15 / defense-in-depth):
-- handler-level лимит MaxLength=50_000 уже есть в
-- internal/notepad/handler.go, но без DB-CHECK игрок может обойти
-- handler через bg-jobs / direct SQL / future endpoints.
-- Добавляем CHECK на 50_000 — защита уровня БД.
--
-- Лимит 50K соответствует существующему handler.MaxLength (план 69
-- упрощение: 50K вместо 16KB исходного плана — handler принят как
-- есть, см. simplifications.md [69-Ф.6]).

-- +goose Up
ALTER TABLE user_notepad
    ADD CONSTRAINT user_notepad_content_max_length
    CHECK (length(content) <= 50000);

-- +goose Down
ALTER TABLE user_notepad
    DROP CONSTRAINT IF EXISTS user_notepad_content_max_length;
