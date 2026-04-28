-- План 67 Ф.1: расширение enum alliance_relation до 5 значений
-- (D-014, B1-решение): friend / neutral / hostile_neutral / nap / war.
--
-- Текущий enum (миграция 0028): ('nap', 'war', 'ally').
-- Целевой:                       ('nap', 'war', 'ally', 'friend',
--                                  'neutral', 'hostile_neutral').
--
-- Стратегия: НЕ переименовываем 'ally' (это сломало бы code/тесты в
-- одной миграции — postgres до commit'а нового value не пускает его
-- в DML). Добавляем 3 новых значения этой миграцией; миграция данных
-- 'ally' → 'friend' выполняется отдельной миграцией 0077_*. Сервис в
-- Ф.2 будет принимать оба значения до полного перехода.
--
-- Down не удаляет enum-значения (postgres этого не умеет без
-- пересоздания типа), это намеренно — symmetric с 0013_artefact_market.

-- +goose Up
-- +goose StatementBegin
ALTER TYPE alliance_relation ADD VALUE IF NOT EXISTS 'friend';
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TYPE alliance_relation ADD VALUE IF NOT EXISTS 'neutral';
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TYPE alliance_relation ADD VALUE IF NOT EXISTS 'hostile_neutral';
-- +goose StatementEnd

-- +goose Down
-- enum-значения удалить нельзя без пересоздания типа — оставляем.
SELECT 1;
