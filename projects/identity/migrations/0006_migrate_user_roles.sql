-- План 52 Ф.1: миграция данных из users.roles[] в user_roles table.
-- После миграции колонка users.roles остаётся (для backward-compat
-- старого кода); удаление в 0007 после полного перехода на user_roles.
--
-- Стратегия: для каждого юзера и каждой роли в его roles[] —
-- INSERT в user_roles, если такая роль есть в roles table. Несуществующие
-- роли игнорируются (с предупреждением через RAISE NOTICE).

-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    u_id UUID;
    role_name TEXT;
    r_id INTEGER;
    inserted_count INTEGER := 0;
    skipped_count INTEGER := 0;
BEGIN
    FOR u_id, role_name IN
        SELECT u.id, unnest(u.roles)::TEXT
        FROM users u
        WHERE u.deleted_at IS NULL
          AND u.roles IS NOT NULL
          AND array_length(u.roles, 1) > 0
    LOOP
        SELECT id INTO r_id FROM roles WHERE name = role_name;
        IF r_id IS NULL THEN
            RAISE NOTICE 'Skipping unknown role % for user %', role_name, u_id;
            skipped_count := skipped_count + 1;
            CONTINUE;
        END IF;
        -- Идемпотентность: ON CONFLICT DO NOTHING (повторный запуск не дублирует)
        INSERT INTO user_roles (user_id, role_id, granted_by, granted_at)
        VALUES (u_id, r_id, NULL, now())
        ON CONFLICT (user_id, role_id) DO NOTHING;
        inserted_count := inserted_count + 1;
    END LOOP;
    RAISE NOTICE 'Migrated user-role assignments: inserted=%, skipped=%',
        inserted_count, skipped_count;
END $$;

-- Audit запись о миграции (с system actor uuid '00000000-0000-0000-0000-000000000000')
INSERT INTO audit_role_changes (actor_id, target_id, role_name, action, reason)
SELECT
    '00000000-0000-0000-0000-000000000000'::UUID,
    ur.user_id,
    r.name,
    'grant',
    'auto-migration from users.roles[] (план 52 Ф.1)'
FROM user_roles ur
JOIN roles r ON r.id = ur.role_id
WHERE ur.granted_by IS NULL
  AND ur.granted_at > now() - INTERVAL '1 hour';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Откат удаляет ВСЕ user_roles assignments — пользоваться осторожно.
-- Audit-записи о migration не удаляются (immutable log).
DELETE FROM user_roles WHERE granted_by IS NULL;
-- +goose StatementEnd
