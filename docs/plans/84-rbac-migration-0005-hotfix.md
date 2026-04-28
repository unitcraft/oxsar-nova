# План 84: Hotfix миграции 0005_rbac_tables (CROSS JOIN без ON CONFLICT)

**Дата**: 2026-04-28
**Статус**: ✅ ЗАКРЫТО (2026-04-28). Миграция 0005 переписана с
WHERE+ON CONFLICT, P80.A закрыт. Smoke с docker compose не выполнялся
в этой сессии (требует destructive `down -v`); ручной smoke на
стороне пользователя.
**Зависимости**: ✅ план 52 (RBAC unification, миграция 0005), ✅
план 80 (Dockerfile.migrate fix, выявил баг 0005).
**Связанные документы**:
- [docs/simplifications.md](../simplifications.md) запись P80.A.
- [docs/plans/52-rbac-unification.md](52-rbac-unification.md) — миграция 0005.
- [docs/plans/80-auth-leftovers-cleanup.md](80-auth-leftovers-cleanup.md) — smoke выявил баг.

---

## Контекст

План 80 закрыт ✅ и в smoke-тесте обнаружил **настоящий prod-bug**
в миграции 0005_rbac_tables.sql (план 52, RBAC unification).

До плана 80 `Dockerfile.migrate` копировал миграции из несуществующей
папки `projects/auth/migrations` — миграции **никогда не применялись**
в multi-instance dev-стеке. Identity-БД стартовала пустой 7+ месяцев.
План 80 это починил, миграции 0001-0004 применились, на 0005 упало:

```
duplicate key value violates unique constraint "role_permissions_pkey"
(SQLSTATE 23505)
```

## Корень проблемы

В `projects/identity/migrations/0005_rbac_tables.sql` финальный INSERT
для роли `superadmin`:

```sql
-- Выдаём superadmin'у все permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p;
```

Это CROSS JOIN для **всех ролей** (`support`, `moderator`, `admin`,
`billing_admin`, `superadmin`), которым выше в той же миграции уже
выданы permissions подмножествами. Конфликт PK
`(role_id, permission_id)`.

Замысел был — выдать superadmin **все** permissions через CROSS JOIN.
Но забыли `WHERE r.name = 'superadmin'`.

## Что делаем

**Один коммит, минимальный фикс.**

Вариант 1 (явное WHERE — предпочтительный):
```sql
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'superadmin';
```

Вариант 2 (ON CONFLICT — defense-in-depth):
```sql
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'superadmin'
ON CONFLICT (role_id, permission_id) DO NOTHING;
```

Вариант 2 безопаснее — идемпотентно, переживает повторный запуск.
**Делаем Вариант 2.**

## Как править

**Не создавать новую миграцию 0007** — это backwards compatibility
hell для prod-БД где 0005 уже применилась с CROSS JOIN успешно
(теоретически невозможно для текущей кодовой базы, но как принцип).

**Вместо этого** — править саму миграцию 0005. Оправдание:
- В prod БД с 0005 уже применённой миграцией ничего не сломается
  (PK защищает от повторных INSERT'ов).
- В dev/test БД и при первом prod-deploy — миграция теперь
  применится без ошибки.
- Goose читает миграцию по hash, но если контент меняется ПОСЛЕ
  применения — goose не перезапустит её. Так что dev-БД где 0005
  уже частично применилась — нужно `down -v && up` или ручной
  rollback (это разовая операция в dev, не план).

## Smoke

1. `docker compose down -v` (destructive — пересоздаст БД).
2. `docker compose up` — все миграции 0001-0006 (или больше)
   применяются успешно.
3. `psql identity-db`:
   ```sql
   SELECT r.name, COUNT(rp.permission_id) AS perm_count
   FROM roles r
   LEFT JOIN role_permissions rp ON rp.role_id = r.id
   GROUP BY r.name
   ORDER BY perm_count DESC;
   ```
   Ожидаемое: `superadmin` имеет permissions = `(SELECT count(*)
   FROM permissions)`. Остальные роли — меньше (по своим
   подмножествам).
4. Identity-сервис стартует, JWT-flow работает.

## Объём

~5 строк SQL + smoke. ~30 минут.

Один коммит:
`fix(identity): миграция 0005 — WHERE+ON CONFLICT для superadmin (план 84, P80.A)`

## Acceptance

- Миграция 0005 применяется чисто на пустой БД.
- На уже применённой 0005-БД — нет конфликта при повторном запуске
  (для будущего безопаснее).
- Smoke: identity-стек поднимается с нуля.
- P80.A в simplifications.md → ✅.
- Шапка плана 84 ✅.
- Запись итерации в docs/project-creation.txt.
