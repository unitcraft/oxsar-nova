# Промпт: выполнить план 77 (billing-client integration в game-nova)

**Дата создания**: 2026-04-28
**План**: [docs/plans/77-game-nova-billing-client.md](../plans/77-game-nova-billing-client.md)
**Зависимости**: нет блокирующих. Разблокирует план 65 Ф.6
(KindTeleportPlanet) + план 66 Ф.5 (платный выкуп удержания).
**Объём**: 1-2 нед, ~400-600 строк Go + тесты, 1-2 коммита.

---

```
Задача: выполнить план 77 — billing-client integration в game-nova.

Добавить в game-nova-backend HTTP-клиент для списания оксаров
(hard, ст. 437 ГК) с пользовательских кошельков через billing-сервис.
Без этого блокированы планы 65 Ф.6 и 66 Ф.5 (премиум-фичи через
оксары).

ВАЖНОЕ:
- Это инфра-задача, не game-mechanic. R0 (геймплей nova заморожен)
  не нарушается — только добавляем способ списывать оксары.
- Параллельно могут идти агенты по другим планам (66 Ф.6 golden,
  67 Ф.5 frontend, 71 UX). Не пересекаются по файлам.
- ВНИМАНИЕ ПО GIT (4 раза граблей): ВСЕГДА `git commit -m "..." --
  path1 path2 ...` с двойным тире. Без `--` риск захватить чужие
  staged-файлы из параллельных сессий.

ПЕРЕД НАЧАЛОМ:

1) git status --short — если есть чужие правки от параллельных
   сессий, бери только свои файлы. Поимённый git add.

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/77-game-nova-billing-client.md (твоё ТЗ)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5»
     (R0-R15, особенно R6/R8/R9/R15)
   - CLAUDE.md (правила: pgx без sqlc, поимённый git add,
     conventional commits)

3) Прочитай выборочно как образцы:
   - projects/portal/backend/internal/portalsvc/credits.go —
     EXISTING billing-client в portal (~80 строк). Это ТВОЙ образец,
     расширишь под game-nova.
   - projects/billing/backend/internal/billing/wallet.go — серверная
     сторона `/wallet/spend` (понимать что приходит/возвращает).
   - projects/game-nova/backend/cmd/server/main.go — как роутер
     устроен, куда вставить idempotency-middleware.
   - projects/game-nova/backend/pkg/metrics/balance.go или
     alliance.go — паттерн Prometheus метрик (R8).
   - docs/adr/0009-currency-rebranding.md — что списывается как
     оксар vs оксарит (R1).

ЧТО НУЖНО СДЕЛАТЬ:

Ф.1. Скаффолд клиента internal/billing/client/

- client.go: тип Client с методами Spend(ctx, in) и
  Refund(ctx, in). Образец — portal/internal/portalsvc/credits.go,
  но добавь Refund (для отмены telepor'а) и расширенный набор
  sentinel-ошибок.
- errors.go: ErrInsufficientOxsar (HTTP 402), ErrBillingUnavailable
  (timeout/network), ErrIdempotencyConflict (HTTP 409).
- HTTP POST /billing/wallet/spend (и /refund если нужен — иначе
  через тот же endpoint с отрицательным amount? — посмотри
  billing-сервис).
- JWT-forwarding через Authorization: Bearer header.
- Idempotency-Key header — обязательный параметр (R9).
- Prometheus метрики (R8):
  · oxsar_billing_client_spend_total{status} (counter:
    ok/insufficient/conflict/unavailable/error)
  · oxsar_billing_client_duration_seconds{operation} (histogram:
    spend/refund)
  Регистрация через RegisterBilling() в pkg/metrics/billing.go,
  подключить к metrics.Register().
- Unit-тесты с httptest.NewServer:
  · 200 OK
  · 402 → ErrInsufficientOxsar
  · 409 → ErrIdempotencyConflict
  · 500 → generic error
  · timeout → ErrBillingUnavailable
  · retry-логика на 503/504 (если решишь добавить — обоснуй или
    отложи; R15 говорит без упрощений, retry полезен)
- go build + go test ./internal/billing/client/... — зелёные.

Ф.2. Idempotency-middleware pkg/idempotency/middleware.go

- Chi-middleware: проверка Idempotency-Key header на мутирующих
  endpoint'ах.
- Redis-кеш через github.com/redis/go-redis/v9 (используется в
  game-nova, см. план 32).
- TTL 24h (или конфиг).
- Body-hash через SHA-256 для проверки совпадения.
- Логика:
  · header нет → handler вызывается без кеша (idempotency опц.)
  · header есть, ключ в Redis + body совпадает → вернуть
    кешированный ответ
  · header есть, ключ в Redis + body другой → 409
  · header есть, ключа нет → выполнить + сохранить
- Атомарность через Redis SET NX.
- Unit-тесты с redismock или поднятым redis-test.

Ф.3. Интеграция в роутер

- cmd/server/main.go:
  · env BILLING_URL (по умолчанию пусто = client отключён, как у
    portal'а)
  · billing.NewClient(billingURL) инициализация
  · idempotency-middleware регистрация (пока не привязан к
    конкретным роутам — это сделают планы 65 Ф.6 / 66 Ф.5)
  · прокидывание client в HandlerDeps или подобное
- deploy/docker-compose.multiverse.yml — env BILLING_URL для
  game-nova-backend (как уже у portal-backend).

Ф.4. Integration-тест end-to-end

- httptest или docker-compose с реальным billing.
- Сценарий: создать тестового пользователя с кошельком 1000 оксаров;
  Spend через client; проверить что списалось; повторный Spend с
  тем же Idempotency-Key — кешированный ответ, не списывает повторно.

Ф.5. Финализация

- Шапка плана 77 → ✅.
- Запись в docs/project-creation.txt — итерация 77.
- Обновить release-roadmap «Пост-запуск v3» — план 77 разблокировал
  65 Ф.6 + 66 Ф.5.

ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА (R0-R15):

R0: не правь modern-числа/механики (это инфра, не game-mechanic).
R1: snake_case JSON, английский, полные слова.
R2: OpenAPI первым (для нового /api/* endpoint'а в game-nova —
    хотя план 77 не вводит публичных endpoint'ов сам, только
    инфраструктуру; openapi.yaml здесь не нужен).
R6: REST-стиль вызовов billing-сервиса (POST /wallet/spend).
R8: Prometheus метрики обязательно (counter+histogram).
R9: Idempotency-Key — это центральный функционал плана.
R10: per-universe изоляция не применима (billing — cross-universe).
R12: i18n — нет user-facing строк, только error messages для логов.
R13: typed payload (SpendInput / RefundInput) ✅.
R15: без упрощений — все edge-кейсы покрыты тестами, retry-логика
     для transient errors, явный TTL idempotency.

R15 УТОЧНЕНО (см. roadmap-report.md "Часть I.5 / R15"):

🚫 НЕ КЛАССИФИЦИРУЙ КАК TRADE-OFF:
- Пропуск R8 Prometheus метрик — обязательно для billing-операций
  (financial flow должен быть наблюдаемым).
- Пропуск Idempotency-Key middleware — это ЦЕЛЬ плана 77.
- Skipped retry на transient errors — financial операции должны
  иметь retry с backoff, иначе пользователь теряет оксары при
  network glitch.
- Skipped тесты на edge-кейсы (timeout, 409, 500) — обязательны.

✅ TRADE-OFF (можно с обоснованием):
- Использование redismock vs поднятый redis-контейнер для тестов.
- Базовая retry-политика без exponential backoff (с обоснованием
  «MVP, доработаем при росте traffic»).

GIT-ИЗОЛЯЦИЯ:
- Свои пути:
  · projects/game-nova/backend/internal/billing/client/
  · projects/game-nova/backend/pkg/idempotency/
  · projects/game-nova/backend/pkg/metrics/billing.go (новый файл
    для метрик)
  · projects/game-nova/backend/pkg/metrics/metrics.go (только
    регистрация RegisterBilling в Register())
  · projects/game-nova/backend/cmd/server/main.go (только инициализация)
  · deploy/docker-compose.multiverse.yml (только env BILLING_URL)
  · docs/plans/77-..., docs/project-creation.txt,
    docs/release-roadmap.md
- ВСЕГДА `git commit -m "..." -- path1 path2 path3` с двойным
  тире (см. memory feedback_parallel_session_check.md, 4 прецедента).
- git status --short ПЕРЕД commit, git diff --cached --name-only
  для проверки.

КОММИТЫ:

1-2 коммита:
1. feat(billing/client): client + idempotency + интеграция (план 77 Ф.1-Ф.3).
2. (опц.) test(billing/client): integration end-to-end + финализация (план 77 Ф.4-Ф.5).

ИЛИ один большой:
feat(billing/client): integration в game-nova (план 77).

ЧЕГО НЕ ДЕЛАТЬ:
- НЕ менять billing-сервис (план 38 закрыт).
- НЕ дублировать portal/internal/portalsvc/credits.go в общий пакет
  pkg/billingclient — это отдельная задача унификации, не сейчас.
- НЕ вводить wallet management в game-nova (создание/чтение балансов).
  Только списание через client.
- НЕ реализовывать оксариты через billing — оксариты live в
  game-nova локально (план 58).
- НЕ вводить webhook-callback от billing — synchronous spend
  достаточно.

УСПЕШНЫЙ ИСХОД:
- internal/billing/client/ работает (Spend + Refund).
- pkg/idempotency/middleware.go активен в роутере.
- Prometheus метрики регистрируются.
- Unit + integration тесты зелёные.
- Все existing nova-тесты зелёные (R0).
- План 77 ✅, разблокированы 65 Ф.6 и 66 Ф.5.

Стартуй.
```
