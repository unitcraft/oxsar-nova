# План 56: Перенос reports из game-nova в portal-backend

**Дата**: 2026-04-27
**Статус**: Активный
**Зависимости**: план 46 закрыт (изначальная реализация reports в
game-nova). План 51 (identity-rename) и план 52 (RBAC) закрыты —
авторизация админских endpoint'ов опирается на permissions из identity.
**Связанные документы**: [46-age-rating-ugc.md](46-age-rating-ugc.md)
(оригинальная реализация reports в game-nova),
[50-game-origin-legal-fix.md](50-game-origin-legal-fix.md) (Ф.5 этого
плана зависит от 56), [52-rbac-unification.md](52-rbac-unification.md)
(RBAC permissions для admin-endpoint'ов).

---

## Цель

Перенести систему пользовательских жалоб (`user_reports`) из
`projects/game-nova/backend/internal/report/` в
`projects/portal/backend/internal/report/`. Обоснование архитектурное:

- Жалоба — это претензия к **аккаунту** (а не к конкретному
  игровому действию во вселенной). Аккаунт глобальный, поэтому
  и реестр жалоб должен быть глобальным.
- Portal — публичный фасад над всеми вселенными (game-nova,
  game-origin, будущие). Жалобы естественно собираются на этом
  уровне.
- Общепринятая практика игровых платформ (Steam, Riot, Blizzard) —
  жалобы централизованы на платформе, не в каждой игре.
- 149-ФЗ требует, чтобы оператор предоставлял данные жалоб по
  запросу — единый реестр на портале упрощает compliance.
- При добавлении новой вселенной она просто шлёт жалобы на
  тот же endpoint, без дублирования инфраструктуры.

После выполнения этого плана план 50 Ф.5 (кнопка «Пожаловаться» в
game-origin) реализуется тривиально: legacy-вселенная сразу шлёт на
правильный endpoint без необходимости знать про game-nova.

---

## Что меняем

### 1. `projects/portal/backend/internal/report/` (новый пакет)

Перенос Go-кода `service.go` и `handler.go` из game-nova:
- `Service.Create(ctx, reporterID, target...)` — создание жалобы.
- `Service.AdminList(ctx, status, limit)` — выборка для админки.
- `Service.AdminResolve(ctx, id, action, comment)` — резолюция.
- HTTP-эндпоинты: `POST /api/reports`, `GET /api/admin/reports`,
  `POST /api/admin/reports/{id}/resolve`.

Изменения относительно game-nova-версии:
- Авторизация через identity-JWT (как остальной portal-backend).
- RBAC-проверка для `/api/admin/reports*` — permission
  `moderation:reports:read` / `moderation:reports:resolve` (или те
  же, что в плане 52 RBAC, согласовать с плановой моделью).
- Логирование через slog с `trace_id` (стандарт portal).
- Связь с identity для получения username/avatar по reporter_id /
  target_user_id (если нужно для админки) — через identity client,
  как в других portal-сервисах.

### 2. Миграция `user_reports`

Вариант A (предпочтительный) — **новая таблица в БД portal**:

- `projects/portal/migrations/0NNN_user_reports.sql` — копия схемы
  из `projects/game-nova/migrations/0069_user_reports.sql`.
- В portal-БД создаётся таблица с нуля. Существующие жалобы из
  game-nova-БД (если есть в проде) — мигрировать отдельным шагом
  (см. Ф.4).

Вариант B (запасной) — **общая БД**: если portal и game-nova
используют один Postgres-кластер с разными схемами, можно сделать
схему `portal.user_reports` без физической миграции данных. Это
потребует уточнения деплой-конфигурации (см. `deploy/docker-compose.yml`).

В этом плане — идём по варианту A: чистое разделение, своя миграция,
свой ownership.

### 3. Удаление reports из game-nova

После переноса:
- Удалить `projects/game-nova/backend/internal/report/`.
- Удалить хендлеры из `projects/game-nova/backend/cmd/server/main.go`.
- Миграция `projects/game-nova/migrations/0069_user_reports.sql` —
  **не удаляем** (миграции не реверсируем), но добавляем
  следующую: `0NNN_drop_user_reports.sql` с `DROP TABLE user_reports`,
  чтобы приложение перестало использовать локальную таблицу.

### 4. Перенос данных (если есть в проде)

Если в game-nova-БД есть существующие записи жалоб — отдельный
скрипт-миграция:
1. Снять snapshot (`pg_dump --table=user_reports`).
2. Восстановить в portal-БД до выполнения `DROP` в game-nova.
3. После проверки целостности — выполнить `0NNN_drop_user_reports.sql`
   в game-nova-БД.

В dev-окружении этот шаг можно пропустить — таблица пуста или
содержит тестовые данные.

### 5. Frontend — game-nova

Текущая реализация: `projects/game-nova/frontend/src/components/ReportButton.tsx`
шлёт POST на game-nova-API. Изменить:
- Endpoint → portal-URL (например, `https://oxsar-nova.ru/api/reports`
  или относительный `/api/reports` через portal-roxy если есть).
- Базовый URL для portal — обычно уже есть в config'е frontend
  (там, где живёт связь с portal-backend).
- Auth-токен — тот же identity-JWT, который используется для других
  portal-вызовов.

Аналогично для админки: `projects/game-nova/frontend/src/features/admin/AdminReportsTab.tsx` —
**перенести в portal-frontend** (см. Ф.6).

### 6. Frontend — portal (новая админка)

В `projects/portal/frontend/src/pages/admin/` (или эквивалент)
создать:
- `AdminReportsPage.tsx` — список жалоб с фильтром по статусу,
  кнопками «Resolve» / «Reject» с комментарием.
- Перенести компонент из game-nova-frontend, адаптировать под
  portal-стек (если отличается).
- Маршрут — `/admin/reports`, защищён RBAC permission
  `moderation:reports:read`.

### 7. game-origin (план 50 Ф.5 — будущая работа)

В рамках этого плана **не трогаем**. После закрытия плана 56 — план
50 Ф.5 знает endpoint `https://oxsar-nova.ru/api/reports` и шлёт
туда напрямую.

---

## Чего НЕ делаем в этом плане

- Не меняем формат жалоб (поля таблицы, причины) — миграция
  «как есть», только переезд места.
- Не вводим новые типы жалоб или новые причины — оставляем 7 категорий
  плана 46.
- Не меняем UI существенно — только переехал endpoint и место кнопки в
  navigate-структуре админки.
- Не делаем **миграцию данных из game-nova-БД в portal-БД** автоматически
  для dev-окружения. Это шаг для прод-деплоя — описан, но не
  автоматизирован.

---

## Этапы

### Ф.1. Скелет в portal-backend

- Создать `projects/portal/backend/internal/report/` с пустым `service.go`
  и `handler.go`.
- Подключить к роутеру в `projects/portal/backend/cmd/server/main.go`.
- Добавить миграцию `projects/portal/migrations/0NNN_user_reports.sql`
  (копия 0069 из game-nova).
- Прогнать миграцию в dev-БД.

### Ф.2. Перенос Service.Create + POST /api/reports

- Скопировать логику Create из game-nova/internal/report/service.go.
- Адаптировать под portal-style (logger, errors, JWT-context).
- Переменная окружения для identity-URL (если нужно валидировать
  reporter_id или target_user_id через identity).
- Smoke-тест: curl-запрос к новому endpoint'у — запись появилась
  в portal-БД.

### Ф.3. Перенос Admin-endpoint'ов

- Скопировать AdminList и AdminResolve.
- RBAC через identity-JWT (по плану 52 — middleware проверяет
  permission `moderation:reports:read` / `:resolve`).
- Smoke-тест: curl с админ-JWT возвращает список / резолвит запись.

### Ф.4. Миграция данных (опционально для dev, обязательно для prod)

- Скрипт `scripts/migrate-reports-game-nova-to-portal.sh`:
  - `pg_dump --table=user_reports` из game-nova-БД;
  - `pg_restore` в portal-БД с таблицей `user_reports`;
  - проверка `SELECT count(*)` в обеих БД.
- В dev — пропустить (data-loss безопасен).

### Ф.5. Удаление reports из game-nova

- Удалить `projects/game-nova/backend/internal/report/`.
- Удалить регистрацию роутов в game-nova `cmd/server/main.go`.
- Добавить миграцию `0NNN_drop_user_reports.sql` в
  game-nova/migrations/.
- Прогнать миграцию в dev-БД.
- Проверить, что приложение собирается и тесты проходят.

### Ф.6. Frontend — обновление endpoint'а в game-nova

- `projects/game-nova/frontend/src/components/ReportButton.tsx` —
  изменить базовый URL на portal.
- Перенести `AdminReportsTab` из `game-nova/frontend/src/features/admin/`
  в `portal/frontend/src/pages/admin/AdminReportsPage.tsx`.
- В game-nova-admin удалить вкладку «Жалобы» (или сделать
  re-direct/ссылку на portal-admin).

### Ф.7. Документация

- Обновить `docs/plans/46-age-rating-ugc.md` — пометка, что reports
  перенесены в portal по плану 56 (раздел «История изменений»).
- Обновить `docs/plans/50-game-origin-legal-fix.md` Ф.5 — endpoint
  на portal (уже сделано в подготовке плана 56).
- Обновить `docs/ops/ugc-moderation.md` — путь к админке жалоб.
- Запись в `docs/project-creation.txt` — итерация 56.

### Ф.8. Финализация

- Smoke-тест: подача жалобы через game-nova UI → запись в portal-БД,
  видна в portal admin → резолвится с комментарием.
- `git status --short` → коммит только своими файлами поимённо.
- Один или несколько коммитов (на усмотрение исполнителя):
  - `feat(portal): перенос reports из game-nova (план 56)`
  - `chore(game-nova): удаление reports после переноса в portal`
  - `feat(portal): admin reports UI`
  Минимум — 1 коммит, максимум — 3 (по логическим этапам).

---

## Тестирование

- Подача жалобы через game-nova `ReportButton` — запись в portal-БД.
- Подача жалобы из (будущего) game-origin — то же самое (это план 50
  Ф.5, не тестируется в плане 56).
- Админ с permission `moderation:reports:read` видит список;
  без permission — 403.
- Резолюция жалобы — статус меняется на `resolved`/`rejected`,
  комментарий сохраняется.
- Регрессия в game-nova: попытка POST на старый endpoint —
  должно быть 404 (после удаления роутов).
- Cross-universe: жалоба от игрока, играющего в game-nova, на
  игрока, играющего в game-origin — обе записи одного и того же
  identity-аккаунта, видны вместе.

---

## Известные ограничения

- **Cross-universe identification.** Пользователь, играющий в game-nova
  и game-origin под одним identity-аккаунтом, должен идентифицироваться
  одинаково в обеих жалобах. Это **гарантировано** identity-сервисом
  (план 51): user_id в JWT — глобальный. Дополнительной работы не
  требуется.

- **Связь с банами.** Если бан-санкция применяется через identity
  (план 52 RBAC, бан-флаг в users), portal-reports должен иметь
  возможность инициировать бан (через identity API). В этом плане —
  не реализуем, оставляем ручной flow («админ резолвит → идёт в
  identity-admin и банит»). Автоматизация — отдельный план.

- **Анонимные жалобы** — в этом плане не вводим. Все жалобы привязаны
  к reporter_id (из JWT).

---

## Итог

3–6 коммитов, ~500–800 строк правок (в основном перенос). Закрывает
архитектурную проблему «reports живут в game-nova, хотя должны быть
платформенными». Разблокирует план 50 Ф.5 (game-origin сможет слать
жалобы в правильное место).

После выполнения единый реестр жалоб на portal-backend становится
точкой соответствия 149-ФЗ для всех вселенных проекта.
