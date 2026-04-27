# План 38: Billing Service — отдельный микросервис для платежей и кошельков

**Дата**: 2026-04-27
**Статус**: ✅ Закрыт 2026-04-27 (все 8 фаз реализованы: d94a7ad1fd, afaf9b7fa4,
e3ee497af4, a712c8f5f8, e75fa68d8e). YooKassa как основной провайдер —
план 42 (0cda99602f, 39ced883f3, 70c04e2f67).
**Зависимости**: план 36 (auth-service, JWKS) — Ф.10 завершена.
**Отношение к плану 36**: заменяет «Ф.7 — Global Credits + платёжные webhook'и в auth-service».
План 36 §Auth Service частично пересмотрен: `/auth/credits/*` endpoints
переезжают в billing.

---

## Зачем отдельный сервис

Auth-service отвечает за **identity**, billing — за **money**. Это два разных
bounded context'а в DDD-смысле. Объединять их — нарушение Single Responsibility
Principle на уровне сервиса:

- Auth меняется при изменениях OAuth-провайдеров, JWT-схем, MFA.
- Billing меняется при добавлении платёжных шлюзов, тарифов, промокодов.

Отдельный сервис даёт:

- Изоляцию compliance-зон (PCI-related audit-логи отдельно от identity-логов).
- Независимый scale (платёжный шлюз тормозит — auth не страдает).
- Изоляцию blast radius при компрометации (украли auth-service ≠ украли деньги).
- Чистый CI/audit: billing меняется реже, изменения требуют большего скрутиния.

Индустриальный canon: Stripe/PayPal/EVE-PLEX/Steam Wallet — billing **всегда**
отдельный сервис. Identity и Payments никогда не смешиваются.

---

## Архитектура

```
                 ┌─────────────────┐
                 │  billing-service │ :9100
                 │ (Go, oxsar/billing)│
                 └────────┬────────┘
                          │
       ┌──────────────────┼──────────────────┐
       │                  │                  │
       ▼                  ▼                  ▼
 ┌──────────┐      ┌────────────┐    ┌──────────────┐
 │billing-db│      │   redis    │    │ payment      │
 │(Postgres)│      │(idempotency│    │ gateways     │
 │          │      │ + caches)  │    │(Robo/Enot/   │
 │          │      │            │    │ Mock)        │
 └──────────┘      └────────────┘    └──────────────┘
```

Frontend и backend-сервисы (game-nova, portal) дёргают billing напрямую с тем
же RSA-JWT, выпущенным auth-service. Billing верифицирует JWT через JWKS
(аналогично portal/game-nova).

---

## Прод-стандарты для billing (это критично — не упрощать)

### 1. Идемпотентность

Все мутирующие endpoints (`spend`, `credit`, `webhook`) принимают
`Idempotency-Key` header. Реализация Stripe-style:

- Уникальный индекс `idempotency_keys (key) UNIQUE`.
- При первом запросе — выполняем, сохраняем `(key → response_body, response_status)`.
- При повторе с тем же ключом — возвращаем сохранённый ответ без побочных эффектов.
- TTL ключа — 24 часа. После — INSERT может удалить старую запись.

Это закрывает: повторные нажатия «голосовать», retry платёжного шлюза, network glitches.

### 2. Транзакции

Каждое списание/пополнение — внутри `BEGIN; UPDATE wallets; INSERT transactions; COMMIT`.
Никаких «сначала UPDATE, потом INSERT, авось не упадёт».

`SELECT ... FOR UPDATE` на `wallets.balance` — блокировка строки до конца
транзакции (защита от race condition при параллельных списаниях).

### 3. Audit log

Каждая транзакция — INSERT в `transactions` (immutable, INSERT-only — никаких
UPDATE/DELETE). Это бухгалтерская книга. Баланс кошелька — производная
(можно пересчитать как `SUM(delta)`).

`webhook_log` хранит **сырые** запросы от платёжных шлюзов (полный body, headers,
timestamp). Никогда не удаляется (или архивируется в холодное хранилище через
N лет). Нужно для разбора disputes/chargebacks/споров с банком.

### 4. Webhook security

Webhook от платёжного шлюза:
- Проверка подписи (HMAC) — каждый шлюз имеет secret.
- Проверка timestamp — отвергаем запросы старше 5 минут (replay protection).
- IP allowlist (опционально, но полезно).
- Rate-limit на endpoint (защита от flood'а).

### 5. Двойная запись (double-entry bookkeeping)

Каждая транзакция имеет **две стороны**:
- Откуда (debit account) — например, `payment_orders.id` для пополнения, `users.wallet` для списания.
- Куда (credit account) — например, `users.wallet` при пополнении, `feedback.id` при голосовании.

В `transactions` фиксируем `from_account, to_account, amount`. Это позволяет:
- Точно знать, на какие конкретно цели списаны деньги юзера.
- Легко считать revenue (`SUM(amount) WHERE to_account LIKE 'platform:%'`).
- Делать reversals (chargeback) по конкретному `payment_order_id`.

### 6. Currency

Один `wallet` — одна валюта. Сейчас одна валюта (RUB или OXC — внутренняя),
но структура поддерживает мульти-валютность (`wallets.currency_code`).

Все суммы — в **минимальных единицах** (копейки/центы/satoshi), `BIGINT`.
`NUMERIC` НЕ используем — медленнее, плавающая точка опасна.

### 7. Reconciliation

Ежедневный (или почасовой) cron-job сверяет:
- `SUM(transactions.delta) WHERE user_id = X` ?= `wallets.balance WHERE user_id = X`.

Если расхождение — алерт и блокировка операций по этому кошельку
до ручного разбора.

### 8. Безопасность секретов

API-ключи платёжных шлюзов:
- В dev — `.env` (gitignored).
- В prod — Docker secrets (`/run/secrets/billing_robokassa_password`,
  `/run/secrets/billing_enot_secret`, и т.п.) — НЕ env vars (не должны попадать
  в `docker inspect`).
- Никогда не логируем (даже на DEBUG-уровне).

### 9. Observability

- Prometheus metrics: `billing_transactions_total{type, status}`,
  `billing_balance_total{currency}`, `billing_webhook_duration_seconds{provider}`.
- Structured logs: `user_id`, `transaction_id`, `payment_order_id`, `idempotency_key`.
- Distributed tracing: `trace_id` пробрасывается из вызывающего сервиса.

### 10. Health checks

- `GET /healthz` — alive (200).
- `GET /api/ready` — ready (200) если БД ОК + Redis ОК. 503 при проблемах
  (drain mode).

---

## Структура

```
projects/billing/backend/
  cmd/server/main.go              — HTTP :9100
  internal/
    billing/
      wallet.go                   — Spend, Credit, Balance, History
      service.go                  — основной Service, оркестрация
      handler.go                  — HTTP-адаптер
      idempotency.go              — Idempotency-Key store (Redis или billing-db)
      reconcile.go                — daily cron-job
    payment/
      gateway.go                  — interface PaymentGateway
      mock.go                     — MockGateway (dev)
      robokassa.go                — RobokassaGateway
      enot.go                     — EnotGateway
      webhook.go                  — webhook handler с verify
      packages.go                 — каталог пакетов кредитов
    auth/
      jwksloader.go               — DUPLICATE
      middleware.go               — DUPLICATE
    httpx/                        — DUPLICATE
    storage/                      — DUPLICATE
    repo/tx.go                    — DUPLICATE
  pkg/
    ids/                          — DUPLICATE
    jwtrs/                        — DUPLICATE (без Email, без global_credits)
    metrics/                      — DUPLICATE
    trace/                        — DUPLICATE
  go.mod                          — module oxsar/billing
  Dockerfile

projects/billing/migrations/
  0001_init.sql                   — wallets, transactions, payment_orders,
                                    webhook_log, idempotency_keys
```

---

## Схема БД `billing-db`

```sql
-- Кошельки. Один user → много wallet (по валюте).
CREATE TABLE wallets (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL,
    currency_code TEXT NOT NULL DEFAULT 'OXC',  -- внутренняя валюта oxsar
    balance       BIGINT NOT NULL DEFAULT 0,    -- в минимальных единицах
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, currency_code)
);

-- Транзакции (immutable, INSERT-only).
CREATE TABLE transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id       UUID NOT NULL REFERENCES wallets(id),
    delta           BIGINT NOT NULL,            -- + при пополнении, − при списании
    balance_after   BIGINT NOT NULL,            -- snapshot для аудита
    from_account    TEXT NOT NULL,              -- 'payment:robokassa:order_id' / 'wallet:user_id'
    to_account      TEXT NOT NULL,              -- 'wallet:user_id' / 'feedback:vote:id'
    reason          TEXT NOT NULL,              -- enum: 'top_up' | 'feedback_vote' | 'refund' | ...
    ref_id          TEXT,                       -- order_id / feedback_id / etc
    idempotency_key TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ix_transactions_wallet ON transactions(wallet_id, created_at DESC);
CREATE INDEX ix_transactions_idempotency ON transactions(idempotency_key) WHERE idempotency_key IS NOT NULL;

-- Платёжные заказы (создаются при покупке пакета кредитов).
CREATE TABLE payment_orders (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL,
    provider     TEXT NOT NULL,                  -- 'robokassa' | 'enot' | 'mock'
    package_id   TEXT NOT NULL,                  -- 'pack_500' | 'pack_2000' | ...
    amount_kop   BIGINT NOT NULL,                -- сумма к оплате (RUB в копейках)
    credits      BIGINT NOT NULL,                -- сколько OXC получит юзер
    status       TEXT NOT NULL,                  -- 'pending' | 'paid' | 'failed' | 'expired'
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    paid_at      TIMESTAMPTZ
);
CREATE INDEX ix_payment_orders_user ON payment_orders(user_id, created_at DESC);

-- Webhook log (сырой).
CREATE TABLE webhook_log (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider     TEXT NOT NULL,
    received_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    headers      JSONB NOT NULL,
    body         BYTEA NOT NULL,
    signature_ok BOOLEAN,
    order_id     UUID,                           -- если удалось распарсить
    processed_at TIMESTAMPTZ                     -- когда обработан
);
CREATE INDEX ix_webhook_log_received ON webhook_log(received_at DESC);

-- Idempotency keys.
CREATE TABLE idempotency_keys (
    key            TEXT PRIMARY KEY,
    response_body  JSONB NOT NULL,
    response_status INTEGER NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at     TIMESTAMPTZ NOT NULL          -- created_at + 24h
);
CREATE INDEX ix_idempotency_expires ON idempotency_keys(expires_at);
```

---

## API

```
# Внутренние (требуют RSA-JWT, дёргаются из game-nova/portal)
POST /billing/wallet/spend             — списать кредиты (с Idempotency-Key)
POST /billing/wallet/credit            — пополнить (внутреннее, для admin/grant)
GET  /billing/wallet/balance           — текущий баланс
GET  /billing/wallet/history?limit=50  — история транзакций

# Покупка кредитов (требует JWT)
GET  /billing/packages                 — каталог пакетов
POST /billing/orders                   — создать заказ → payURL

# Webhook'и (публичные, проверяют подпись)
POST /billing/webhooks/robokassa
POST /billing/webhooks/enot
POST /billing/webhooks/mock            — для dev

# Health
GET  /healthz
GET  /api/ready
GET  /metrics
```

---

## Изменения в auth-service

- Удалить `GET /auth/credits/balance`, `POST /auth/credits/spend`,
  `GET /auth/credits/history` из роутов.
- Удалить `Service.SpendCredits/AddCredits/CreditBalance/CreditHistory`.
- Удалить колонку `users.global_credits` из auth-db (миграция `0003_drop_global_credits.sql`).
  Баланс теперь в billing-db.
- Удалить `global_credits` из JWT claims (его уже не из чего брать).
- Удалить таблицу `credit_transactions` (если она в auth-db) — она в billing.

---

## Изменения в game-nova

- Удалить `internal/payment/` целиком (handler, gateway, robokassa, enot, mock).
- Удалить `users.credit` из миграций? **Нет** — `users.credit` это *внутриигровой
  credit* (стартовый бонус 5 OXC при регистрации, экспедиционные находки, etc.).
  Это не платёжный кредит — он не покупается, а зарабатывается / даётся механиками.
  Остаётся как есть.
- НО: action'ы покупки **в магазине игры за реальные деньги** (если такие есть)
  переезжают на billing.

---

## Изменения в portal-backend

- `feedback.vote` дёргает `billing /billing/wallet/spend` с `Idempotency-Key`
  = `vote:user_id:feedback_id`. Если 402 (insufficient funds) — возвращаем
  ошибку, голос не записываем.

---

## Frontend

- В шапке игры/портала balance берётся через `GET /billing/wallet/balance`
  (с RSA-JWT). Кэшировать через TanStack Query, инвалидировать после spend.
- Settings → «Купить кредиты» → `GET /billing/packages` → список пакетов →
  `POST /billing/orders` → редирект на payURL.
- Vite proxy `/billing/*` → `billing-service:9100`.

---

## Compose

```yaml
billing-db:
  image: postgres:16-alpine
  ports: ["5436:5432"]                # 5435 уже у auth-db
  volumes: [billing-db-data:/var/lib/postgresql/data]

billing-migrate:
  build: deploy/Dockerfile.migrate
  environment:
    DB_URL: postgres://billing:billing@billing-db:5432/billing?sslmode=disable
    MIGRATE_DIR: /migrations/billing

billing-service:
  build: projects/billing/backend/Dockerfile
  ports: ["9100:9100"]
  environment:
    BILLING_ADDR: ":9100"
    BILLING_DB_URL: postgres://billing:billing@billing-db:5432/billing?sslmode=disable
    REDIS_URL: redis://redis:6379/3   # 0/1/2 заняты game/portal/auth
    AUTH_JWKS_URL: http://auth-service:9000
    PAYMENT_PROVIDER: mock            # dev: mock; prod: robokassa
    PAYMENT_MOCK_BASE_URL: http://billing-service:9100
    PAYMENT_RETURN_URL: http://localhost:5173/
    LOG_LEVEL: info
  depends_on:
    billing-db: { condition: service_healthy }
    redis: { condition: service_healthy }
    auth-service: { condition: service_healthy }
    billing-migrate: { condition: service_completed_successfully }
```

В `Dockerfile.migrate` добавить копирование billing-миграций.

---

## Acceptance

- [ ] `docker compose up` поднимает billing-service вместе со стеком, healthy.
- [ ] `POST /auth/register` → JWT. `GET /billing/wallet/balance` → `{balance: 0}`.
- [ ] `POST /billing/orders {package_id: "pack_500"}` → `{pay_url: ".../mock/pay?..."}`.
- [ ] Кликнуть pay_url → mock-handler → webhook `POST /billing/webhooks/mock`
  → `wallets.balance += 50000` (500 OXC = 50000 копеек), запись в `transactions`.
- [ ] `POST /billing/wallet/spend {amount: 100, reason: "feedback_vote",
  ref_id: "..."}` с `Idempotency-Key: ...` → 204. Баланс = 49900.
- [ ] Повтор того же запроса с тем же Idempotency-Key → 204 без изменения баланса.
- [ ] Spend на 100000 (больше баланса) → 402 «insufficient funds».
- [ ] `feedback.vote` в portal списывает 100 кредитов через billing.
- [ ] `SUM(transactions.delta WHERE wallet_id=X) == wallets.balance WHERE id=X`
  (reconcile).
- [ ] Webhook без подписи / с неверной подписью → 400, `webhook_log.signature_ok=false`.
- [ ] Replay webhook (тот же body через 10 минут) → 400 (timestamp check).
- [ ] `/auth/credits/*` endpoints в auth-service возвращают 404 (удалены).
- [ ] `users.global_credits` в auth-db удалён (миграция 0003).
- [ ] `internal/payment/*` в game-nova удалён.

---

## Фазы

### Ф.1 — Скелет billing-service (½ дня)

- `projects/billing/backend/` — go.mod, cmd/server, Dockerfile.
- DUPLICATE shared-пакеты: `httpx`, `storage`, `repo/tx`, `pkg/jwtrs`,
  `pkg/ids`, `pkg/metrics`, `pkg/trace`, `internal/auth/{jwksloader,middleware}`.
- `Dockerfile.migrate` — добавить COPY billing-миграций.
- Compose добавляет `billing-db`, `billing-migrate`, `billing-service`.
- `GET /healthz` отвечает 200.

### Ф.2 — Миграция БД + wallet API (1 день)

- `0001_init.sql` со всеми таблицами выше.
- `internal/billing/wallet.go`: `Spend`, `Credit`, `Balance`, `History`
  с `SELECT ... FOR UPDATE` + INSERT транзакций.
- `internal/billing/idempotency.go`: middleware читает `Idempotency-Key`,
  ищет в `idempotency_keys`. Если есть — возвращает кэшированный ответ.
  Если нет — выполняет, сохраняет.
- HTTP handler: `POST /billing/wallet/spend|credit`, `GET /billing/wallet/balance|history`.
- RSA middleware читает sub из JWT, авторизует операцию.
- Тесты: spend успех, spend insufficient, double-spend (idempotency),
  параллельный spend (race condition).

### Ф.3 — Payment gateways + orders (1 день)

- Перенести `mock.go`, `robokassa.go`, `enot.go` из game-nova в billing.
- `internal/billing/payment.go`: создание order, BuildPayURL.
- Handler: `GET /billing/packages`, `POST /billing/orders`.
- `packages.go`: каталог пакетов кредитов (mock-данные).
- Тесты на каждый gateway.

### Ф.4 — Webhook'и (½ дня)

- `POST /billing/webhooks/{provider}` — verify signature + timestamp.
- INSERT в `webhook_log` (raw body) **до** verify (audit trail независимо
  от валидности).
- При успехе: UPDATE order.status='paid', credit wallet, INSERT transaction.
  Всё в одной транзакции.
- Идемпотентность по `order_id` (повторный webhook → 200, no-op).
- Тесты: верный signature → balance обновлён; неверный → 400, log записан;
  replay → 400.

### Ф.5 — Migration: убрать payments из game-nova и auth-service (½ дня)

- Удалить `projects/game-nova/backend/internal/payment/` целиком.
- Удалить `users.credit` payments routes (но оставить колонку `credit`
  для внутриигровой механики).
- Удалить из auth-service: `Service.{Spend,Add}Credits`,
  `Service.Credit{Balance,History}`, соответствующие handlers и routes.
- Миграция `auth-db/0003_drop_global_credits.sql`: `ALTER TABLE users
  DROP COLUMN global_credits;` + `DROP TABLE IF EXISTS credit_transactions;`.
- Убрать `GlobalCredits` из JWT Claims (DUPLICATE — три модуля).
- Обновить план 36.

### Ф.6 — Portal: feedback.vote через billing (½ дня)

- `portalsvc.VoteFeedback` дёргает `POST billing/wallet/spend` с
  `Idempotency-Key = "vote:" + userID + ":" + feedbackID`.
- При 402 (insufficient) → 402 клиенту, голос не записывается.
- Только при 204 → INSERT `feedback_votes`.
- Транзакция portal-db: вставка vote и инкремент `vote_count` атомарны
  (если vote_count уже денормализован, иначе SUM по запросу).
- Тесты: spend ОК → vote записан; spend 402 → vote не записан; повторный
  vote с тем же ключом → idempotent.

### Ф.7 — Frontend интеграция (½ дня)

- Vite proxy `/billing/*` в game-nova и portal frontends.
- Компонент в шапке: `GET /billing/wallet/balance` через TanStack Query.
- Settings page → buy credits flow (packages → order → redirect).
- Inv

### Ф.8 — Reconcile + observability (½ дня)

- `internal/billing/reconcile.go`: cron-job (каждый час) сверяет
  `SUM(transactions.delta) GROUP BY wallet_id` с `wallets.balance`.
  При расхождении — алерт + блокировка кошелька (флаг `wallets.frozen`).
- Prometheus metrics: `billing_transactions_total`, `billing_webhook_total`,
  `billing_balance_total`.
- Structured logs.

**Итого**: ~5 рабочих дней.

---

## Свобода ошибиться

Если в процессе обнаружим, что что-то можно сделать проще без потери качества —
обновляем план. Любое отклонение от прод-стандартов — записываем в
[simplifications.md](../simplifications.md) с приоритетом и планом возврата.
