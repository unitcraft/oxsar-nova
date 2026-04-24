# План 14: Расширение админ-панели

Цель: превратить текущую базовую админку (~6 эндпоинтов, 1 экран) в рабочий
инструмент для GM/техподдержки — модерация, дебаг, гейм-мастер-команды,
мониторинг, CMS.

**Scope:** только сервер + админ-UI. Не трогаем публичные правила игры,
балансы (§17 ТЗ — через ADR).

---

## Что уже есть (baseline, 2026-04-24)

| Фича | Эндпоинт | Статус |
|------|----------|--------|
| Статы | `GET /api/admin/stats` | ✅ 4 числа: users/planets/fleets/events |
| Список игроков | `GET /api/admin/users` | ✅ с пагинацией/поиском базово |
| Ban/unban | `POST /api/admin/users/{id}/ban\|unban` | ✅ |
| Grant credits | `POST /api/admin/users/{id}/credit` | ✅ |
| Role | `POST /api/admin/users/{id}/role` | ✅ |
| AutoMsg CMS | `GET/PUT /api/admin/automsgs` | ✅ редактирование шаблонов |

Текущий AdminScreen — одна длинная простыня. TZ: нужен дашборд + табы по разделам.

---

## Ф.1 UX-каркас (приоритет: HIGH)

Без структуры дальше нельзя — всё сольётся в кашу.

### Ф.1.1 Разделить AdminScreen на табы
- `dashboard` — метрики + алерты
- `users` — список с фильтрами, профиль отдельной карточкой
- `economy` — ресурсы, лоты, market health
- `events` — worker queue, retry, dead-letter
- `cms` — AutoMsg, i18n-ключи
- `combat` — последние бои, аномалии
- `audit` — лог админских действий

### Ф.1.2 Audit log для самих админов
Все POST/DELETE под `/api/admin/*` должны писать в новую таблицу
`admin_audit_log`:
- migration `0059_admin_audit_log.sql` — id, admin_id, action, target_id,
  target_kind, payload (JSONB), ip, user_agent, created_at
- middleware `adminAudit` вешает запись после успешного handler'а
- эндпоинт `GET /api/admin/audit?from=&to=&admin_id=&action=` с пагинацией
- UI: таб `audit` — хронология, фильтры

**Почему без этого плохо:** любое действие через ban/grant/role не оставляет
следов. При споре с игроком не найдём, кто и когда дал 50k кредитов.

---

## Ф.2 Users — глубокая карточка игрока (HIGH)

Один клик → увидеть всё про пользователя.

### Ф.2.1 `GET /api/admin/users/{id}` — профиль игрока
Возвращает агрегат:
- базовые поля (username, email, role, banned_at, last_seen, regtime)
- список планет (id, name, coords, points)
- флот в полёте (миссии × планеты)
- открытые лоты на рынке и арт-рынке
- активные офицеры и их до
- последние 20 battle/espionage/expedition reports
- последние 20 транзакций res_log (res_log JOIN)
- последние 20 credit_purchases
- последние 20 сообщений inbox
- IP-история (если собирается; если нет — пропустить)

**UI:** drawer сбоку или отдельная страница `#admin/user/<id>`.

### Ф.2.2 Ресурсы — read/write
- `POST /api/admin/users/{id}/resources` — body: {planet_id, metal, silicon, hydrogen}
  (добавить к существующим, может быть отрицательным — списание)
- аудируется в `admin_audit_log`

### Ф.2.3 Артефакты игрока
- `GET /api/admin/users/{id}/artefacts`
- `POST /api/admin/users/{id}/artefacts/grant {unit_id, count}`
- `DELETE /api/admin/users/{id}/artefacts/{artefact_id}` — для жалоб

### Ф.2.4 Флот — read-only + force-recall — ✅ ЗАКРЫТО 2026-04-25
- `POST /api/admin/fleets/{fleet_id}/recall` — `admin.FleetAdminHandler`.
  Вычитывает owner_user_id, затем `transport.Recall` от его имени.
  Audit-мидлварь пишет лог автоматически. Только admin+.

### Ф.2.5 Планеты — ✅ ЗАКРЫТО 2026-04-25
- `POST /api/admin/planets/{id}/rename {name}` — UPDATE planets.name
  (validate 1..40 chars).
- `POST /api/admin/planets/{id}/transfer {new_user_id}` — проверяет,
  что target user не удалён. UPDATE planets.user_id.
- `DELETE /api/admin/planets/{id}` — soft-delete через `destroyed_at=now`.
  Отказ, если это последняя живая планета игрока (409 Conflict).

### Ф.2.6 Merge/delete-account — ✅ ЗАКРЫТО 2026-04-25
- `DELETE /api/admin/users/{id}` — soft-delete в транзакции:
  users.deleted_at=now, alliance_id=NULL, DELETE FROM alliance_members.
  Планеты не трогаем — они заброшены.
- `POST /api/admin/users/{id}/restore` — снимает deleted_at. Восстановление
  в альянс — ручное (join).

---

## Ф.3 Events / Worker — наблюдение и ручное вмешательство (HIGH)

Worker падает → нужно разбираться. Сейчас — только psql.

### Ф.3.1 `GET /api/admin/events` (list с фильтрами)
- kind, status (pending|processing|dead|done), trace_id, from/to
- пагинация, сортировка по created_at DESC

### Ф.3.2 `GET /api/admin/events/dead-letter`
- список из events_dead_letter с причиной, retry_count, payload
- UI: таблица с кнопкой «👁 детали»

### Ф.3.3 Ручное управление
- `POST /api/admin/events/{id}/retry` — вытащить из dead-letter и положить обратно в events с retry_count=0
- `POST /api/admin/events/{id}/cancel` — помечает status=cancelled
- оба auditable

### Ф.3.4 Метрики worker'а
`GET /api/admin/events/metrics` — за последний час:
- processed, failed, avg_processing_ms, oldest_pending_age_sec
UI: карточки на Dashboard (Ф.1.1).

### Ф.3.5 Trace explorer
Один клик по `trace_id` в таблице events → открыть все связанные строки
(одна миссия может породить 3-4 события: Transport → Arrive → Return).

---

## Ф.4 Economy — здоровье экономики (MEDIUM)

### Ф.4.1 Дашборд-виджеты
`GET /api/admin/economy/overview`:
- общий объём ресурсов на руках (sum metal/silicon/hydrogen)
- общий объём кредитов
- объём сделок на рынке за 24ч
- топ-5 богатых игроков (points)
- 10 самых дорогих лотов арт-рынка

### Ф.4.2 Market moderation
- `GET /api/admin/market/lots` + `DELETE /api/admin/market/lots/{id}` — снять
  подозрительный лот (например, продажа 1 металла за 100k кредитов — мусор)
- auditable

### Ф.4.3 Credit purchases explorer
- `GET /api/admin/purchases?status=&user=&from=&to=`
- `POST /api/admin/purchases/{id}/refund` — пометить refunded + списать
  кредиты с баланса (без возврата денег; реальный refund — через ЛК Робокассы)

### Ф.4.4 Referral overview
- `GET /api/admin/referrals` — кто кого привёл, сумма бонусов
- защита от самореферирования (sanity-check уже есть, но админу полезно
  видеть подозрительные цепочки — 1 IP + 5 регистраций)

---

## Ф.5 Combat — дебаг боёв (MEDIUM)

### Ф.5.1 Browse battle reports
- `GET /api/admin/battle-reports?user=&from=&to=&kind=`
- фильтр по kind (ATTACK, ALLIANCE, ALIEN, PIRATE)
- открытие полного отчёта со всеми участниками

### Ф.5.2 Replay via simulator
Кнопка в отчёте: «Перепроверить через симулятор» — берёт исходные фракции
и прогоняет BattleSim 10 раз, показывает распределение. Помогает на
жалобы «меня неправильно убило».

### Ф.5.3 Anomaly detector (nice-to-have)
`GET /api/admin/combat/anomalies`:
- бои с потерями 0:0 (возможен баг)
- очень короткие выигрыши игроков с низким флотом
- UI: таблица-алерт

---

## Ф.6 CMS-улучшения (MEDIUM)

### Ф.6.1 AutoMsg editor — улучшения
- preview с подстановкой переменных
- история версий (кто и когда менял шаблон)
- rollback к предыдущей версии

### Ф.6.2 i18n CMS
- `GET/PUT /api/admin/i18n/{lang}/{group}/{key}` — редактировать ключ
  без пересборки фронта (сейчас — только через yaml + restart)
- UI: дерево групп → ключи, diff до/после
- осторожно: не все ключи safe к перезаписи (template-переменные)

### Ф.6.3 Tutorial/Achievements editor
- `GET /api/admin/tutorial/steps` + `PUT /api/admin/tutorial/steps/{id}`
- та же схема для `/api/admin/achievements/{id}` — порог/название/описание

---

## Ф.7 Мониторинг и алерты (LOW)

### Ф.7.1 Server health endpoint
`GET /api/admin/health/detailed`:
- pg conn pool (in use/idle/max)
- redis ping_ms
- worker heartbeat age
- disk usage (если доступен через os)
- memory (runtime.MemStats)

### Ф.7.2 Slow queries
`GET /api/admin/slow-queries` — из pg_stat_statements топ-20 по total_time.
Требует extension в миграциях (`CREATE EXTENSION IF NOT EXISTS pg_stat_statements`).

### Ф.7.3 Alerts panel
Правила: «>100 events pending», «worker heartbeat >60s назад»,
«dead-letter растёт». Не отправляем никуда — просто красные карточки
на Dashboard.

---

## Ф.8 Tech debt / safety (HIGH — фоном)

### Ф.8.1 Proper RBAC
Сейчас всё — под `superadmin`. Добавить промежуточную роль `moderator`:
- moderator: ban/unban, снять лот, view-only
- superadmin: credit grant, role change, delete user

Миграция: `ALTER TYPE user_role ADD VALUE 'moderator'`.
Middleware `RequireRole(min)` вместо `AdminOnly`.

### Ф.8.2 Rate-limit админских действий — ✅ ЗАКРЫТО 2026-04-25
- `admin.RateLimitMiddleware()` подключён к `/api/admin` scope (после RBAC
  и Audit). 100 write-действий (POST/PUT/PATCH/DELETE) в час на админа.
  429 + `Retry-After: 3600` секунд. GET не ограничивается.
- Счётчик in-memory (`sync.Mutex` + map). Перезапуск сервера обнуляет;
  этого достаточно, так как админов мало и это soft-safeguard.

### Ф.8.3 Подтверждения для деструктивных операций
- delete user / planet → модалка с вводом `username` для подтверждения
- force-recall flot → confirm
- refund purchase → confirm + обоснование (поле reason обязательно)

### Ф.8.4 2FA для superadmin
TOTP / email-code на вход в админку. Не в базовом auth — отдельный
шаг перед открытием `#admin`. Приоритет низкий, пока игроков мало.

---

## Порядок реализации

1. **Ф.1** (табы + audit log) — фундамент, 1 итерация
2. **Ф.2** (карточка игрока) — самое востребованное, 2 итерации
3. **Ф.3** (events) — важно для стабильности, 1-2 итерации
4. **Ф.8.1 + Ф.8.3** (RBAC + confirm) — параллельно с Ф.2/Ф.3, 1 итерация
5. **Ф.4** (economy) — 1 итерация
6. **Ф.5** (combat) — 1 итерация
7. **Ф.6** (CMS-улучшения) — 1-2 итерации
8. **Ф.7** (мониторинг) — 1 итерация
9. **Ф.8.2, 8.4** (rate-limit, 2FA) — по потребности

**Оценка:** 10-12 итераций (~3-4 недели при 1 фокус-сессия/день).

---

## Что НЕ делаем в этом плане

- **Web-UI для редактирования игровых формул** — это ADR'ы через git,
  не CMS. Админка случайно поменяет баланс — восстанавливать неделю.
- **Real-time чат между админами** — если понадобится, Slack/Telegram.
- **Отчёты в PDF/Excel** — пока нет запроса.
- **Плейбуки / runbooks внутри UI** — держим их в `docs/ops/runbooks/` (git).

---

## Риски

- **Авария через админку быстрее, чем через код.** Каждая деструктивная
  операция — auditable + confirm + RBAC + rate-limit. Иначе у нас
  human error > bug rate.
- **Админский трафик в общей БД.** Тяжёлый GET /admin/events с фильтрами
  может положить pg. Нужны индексы (уже есть на events, но проверить
  на composite фильтры).
- **Утечка PII.** `GET /api/admin/users/{id}` возвращает email, ip,
  payments. Доступ только superadmin. Логировать открытия профилей.
- **Права moderator vs superadmin** — ошибка в RBAC = эскалация
  привилегий. Тесты на middleware (Ф.8.1) обязательны.

---

## Связанные планы

- [13-ui-testing.md](13-ui-testing.md) — E2E-тесты должны покрыть новые
  админ-экраны (добавить в matrix при реализации Ф.2/Ф.3)
- [07-payments.md](07-payments.md) — refund-флоу частично описан там;
  Ф.4.3 доделывает UI к нему
- [09-event-system.md](09-event-system.md) — worker metrics в Ф.3.4
  берутся из ф.5 того плана
