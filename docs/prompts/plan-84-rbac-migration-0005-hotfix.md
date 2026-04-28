# Промпт: выполнить план 84 (hotfix миграции 0005_rbac_tables)

**Дата создания**: 2026-04-28
**План**: [docs/plans/84-rbac-migration-0005-hotfix.md](../plans/84-rbac-migration-0005-hotfix.md)
**Зависимости**: ✅ план 80 (Dockerfile.migrate fix выявил баг). HIGH
PRIORITY — identity-стек не поднимается с нуля.
**Объём**: ~5 строк SQL + smoke + docs. ~30 минут. 1 коммит.

---

```
Задача: hotfix миграции 0005_rbac_tables — добавить WHERE+ON CONFLICT
к INSERT для роли superadmin (строка 167 файла). Это **prod-баг**
выявленный планом 80 после починки Dockerfile.migrate.

КОНТЕКСТ:

В projects/identity/migrations/0005_rbac_tables.sql строка 167:

```sql
-- superadmin: всё (включая управление ролями + system config)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p;
```

Это CROSS JOIN для **всех** ролей (support/moderator/admin/
billing_admin/superadmin), включая те, кому в строках 130-163 уже
выданы permissions подмножествами. Конфликт PK
(role_id, permission_id) — `duplicate key value violates unique
constraint "role_permissions_pkey"` (SQLSTATE 23505).

Замысел был — выдать superadmin **все** permissions через CROSS JOIN,
но забыли WHERE r.name = 'superadmin'. До плана 80 миграция не
применялась (Dockerfile.migrate был сломан, identity-БД стартовала
пустой), баг был замаскирован.

Запись в simplifications.md как **P80.A** уже есть.

ПЕРЕД НАЧАЛОМ:

ПЕРВЫМ ДЕЙСТВИЕМ (до любого чтения плана):

1) git status --short. cat docs/active-sessions.md.

2) ОБЯЗАТЕЛЬНО добавь свою строку в раздел «Активные сессии»:
   | <N> | План 84 hotfix migration 0005 | projects/identity/migrations/0005_rbac_tables.sql, docs/plans/84-..., docs/simplifications.md, docs/project-creation.txt | <дата-время> | fix(identity): миграция 0005 — WHERE+ON CONFLICT для superadmin (план 84, P80.A) |

3) Если есть другие активные слоты — план 84 БЕЗОПАСЕН в параллели
   (точечная правка одной SQL-миграции, никто другой эту миграцию
   не трогает).

ТОЛЬКО ПОСЛЕ — переходи к чтению:

4) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/84-rbac-migration-0005-hotfix.md (твоё ТЗ)
   - docs/simplifications.md секцию P80.A
   - projects/identity/migrations/0005_rbac_tables.sql (саму миграцию,
     обрати внимание на строки 130-163 как эталон правильных INSERT'ов)

5) Опционально:
   - docs/plans/52-rbac-unification.md (контекст плана RBAC).
   - docs/plans/80-auth-leftovers-cleanup.md (где баг выявлен).

ЧТО НУЖНО СДЕЛАТЬ:

### Ф.1. Правка миграции

В projects/identity/migrations/0005_rbac_tables.sql найди строку 167:

```sql
-- superadmin: всё (включая управление ролями + system config)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p;
```

Замени на:

```sql
-- superadmin: всё (включая управление ролями + system config)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.name = 'superadmin'
ON CONFLICT (role_id, permission_id) DO NOTHING;
```

Изменения:
1. `FROM roles r, permissions p` → `FROM roles r CROSS JOIN permissions p`
   — стиль consistent с остальными INSERT'ами в файле (130-163).
2. Добавлен `WHERE r.name = 'superadmin'` — основной фикс,
   ограничивает roles только нужной.
3. Добавлен `ON CONFLICT (role_id, permission_id) DO NOTHING` —
   defense-in-depth, гарантирует идемпотентность повторного запуска.

### Ф.2. Smoke (ВАЖНО — нужен Docker)

ПЕРЕД smoke — спроси пользователя готов ли он к destructive
операции `docker compose down -v` (потеря локальных dev-данных в
identity-БД, она пересоздастся пустой).

Если пользователь подтвердил:

```bash
cd <проект>
docker compose -f deploy/docker-compose.yml down -v
docker compose -f deploy/docker-compose.yml up -d --build identity-db identity-migrate
docker compose -f deploy/docker-compose.yml logs identity-migrate
```

Ожидание:
- identity-migrate exits with status 0.
- В logs: `OK   0001_init.sql`, `OK   0002_*.sql`, ..., `OK   0005_rbac_tables.sql`,
  `OK   0006_*.sql` (если есть).
- БЕЗ строк `ERROR: duplicate key value`.

Затем psql проверка:
```sql
SELECT r.name, COUNT(rp.permission_id) AS perm_count
FROM roles r
LEFT JOIN role_permissions rp ON rp.role_id = r.id
GROUP BY r.name
ORDER BY perm_count DESC;
```

Ожидание:
- `superadmin` имеет permissions = `(SELECT count(*) FROM permissions)`
  (все permissions).
- Остальные роли имеют меньше — точные числа из явных списков
  в INSERT'ах:
  - `support`: 4 permissions.
  - `moderator`: 6.
  - `admin`: 17.
  - `billing_admin`: 5.
  - `player`: 0.

Если Docker недоступен в твоей среде — пропусти smoke, помести
ручное smoke-задание для пользователя в commit-message.

### Ф.3. Финализация

- Шапка плана 84 ✅.
- В docs/simplifications.md секция P80.A → пометить статус ✅
  закрыто планом 84.
- Запись итерации в docs/project-creation.txt («84 — hotfix
  миграции 0005, P80.A закрыт»).

══════════════════════════════════════════════════════════════════
ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА R0-R15 — см. блок rules-r0-r15.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/rules-r0-r15.md

Особо важно:
- R0: nova-баланс не меняем (это инфра-фикс).
- R7: backward compat — править саму миграцию 0005, НЕ создавать
  новую 0007. Это безопасно: на пустой БД — миграция применится
  чисто; на уже-применённой — ON CONFLICT защитит. Goose не
  перезапустит миграцию по hash-изменению (она уже зарегистрирована
  в schema_migrations). Для dev-БД где 0005 частично применилась —
  пользователь делает down -v && up.
- R15: без MVP-сокращений. Smoke обязателен (или явная пометка
  что отложен).
══════════════════════════════════════════════════════════════════

══════════════════════════════════════════════════════════════════
GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ — см. блок git-isolation.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/git-isolation.md

Свои пути для CC_AGENT_PATHS:
- projects/identity/migrations/0005_rbac_tables.sql
- docs/plans/84-rbac-migration-0005-hotfix.md
- docs/simplifications.md (только P80.A статус)
- docs/active-sessions.md
- docs/project-creation.txt
══════════════════════════════════════════════════════════════════

КОММИТЫ:

Один коммит:

fix(identity): миграция 0005 — WHERE+ON CONFLICT для superadmin (план 84, P80.A)

Trailer: Generated-with: Claude Code
ВСЕГДА с двойным тире:
git commit -m "..." -- $CC_AGENT_PATHS

ЧЕГО НЕ ДЕЛАТЬ:

- НЕ создавай новую миграцию 0007 — правь саму 0005 (см. R7
  обоснование выше).
- НЕ запускай docker compose down -v БЕЗ согласия пользователя.
- НЕ менять остальные INSERT'ы в 0005 (130-163) — они корректны.
- НЕ забывай про -- в git commit.

УСПЕШНЫЙ ИСХОД:

- Миграция 0005 применяется чисто на пустой БД (smoke OK или
  отложен).
- 0005 идемпотентна (ON CONFLICT защищает повторный запуск).
- superadmin имеет все permissions через WHERE-условие.
- Остальные роли имеют свои подмножества (без duplicate-error).
- P80.A в simplifications.md → ✅ закрыт.
- Шапка плана 84 ✅.
- Запись в docs/project-creation.txt.
- Удалена строка из docs/active-sessions.md.

Стартуй ТОЛЬКО когда slot 3 освободится (план 72 Ф.4 закроется).
```
