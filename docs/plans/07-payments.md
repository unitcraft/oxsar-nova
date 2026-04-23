# План G: Платёжная система

Реализуется с нуля. Нет легаси-кода для переноса.

---

## Выбор платёжного шлюза

### Сравнение для самозанятого/физлица (без ИП, Россия, 2025)

| Шлюз | Самозанятый/физлицо | Комиссия | Выплаты | Особенности |
|------|---------------------|----------|---------|-------------|
| **Робокасса** | ✅ Самозанятый (НПД) | ~3% | Мгновенно | 130+ методов, СБП, специализация на цифровых товарах |
| **Enot.io** | ✅ Самозанятый, физлицо | 2–4% | Авто-выплаты | Простой API, быстрая регистрация |
| **Prodamus** | ✅ Самозанятый (НПД) | 3–5% | По расписанию | Встроенный чек ОФД |

> ⚠️ Лимит НПД: 2.4 млн руб/год. При превышении — смена налогового режима.

**Выбор: Робокасса** — основной. Enot.io — резервный (подключается позже при необходимости).
Оба поддерживаются через `Gateway` интерфейс — смена шлюза не ломает сервисный слой.

---

## Пакеты кредитов

| Пакет | Ключ | Кредиты | Бонус | Цена |
|-------|------|---------|-------|------|
| Пробный | `trial` | 400 | — | 49 руб |
| Стартовый | `starter` | 1 000 | — | 100 руб |
| Средний | `medium` | 3 000 | +200 кр | 250 руб |
| Большой | `big` | 7 000 | +500 кр | 500 руб |
| Максимальный | `max` | 15 000 | +2 000 кр | 1 000 руб |

Кредиты хранятся в `users.credit` (уже существует, `numeric(15,2)`).

---

## Контекст проекта

- Миграции: следующая — `0049_credit_purchases.sql`
- Inline SQL в service.go (нет отдельного `queries/`)
- Транзакции через `s.db.InTx(ctx, fn)`
- Конфиг из ENV через `config.Load()`
- Маршруты регистрируются в `backend/cmd/server/main.go`
- OpenAPI-контракт: `api/openapi.yaml`

---

## Открытые задачи

### G.1 Миграция (приоритет: HIGH)

Файл `migrations/0049_credit_purchases.sql`:

```sql
-- +goose Up
CREATE TABLE credit_purchases (
    id             TEXT PRIMARY KEY,
    user_id        TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    package_key    TEXT NOT NULL,
    amount_credits INT  NOT NULL,
    price_rub      NUMERIC(10,2) NOT NULL,
    provider       TEXT NOT NULL DEFAULT 'robokassa',
    provider_id    TEXT UNIQUE,                        -- ID транзакции от шлюза, NULL до оплаты
    status         TEXT NOT NULL DEFAULT 'pending',   -- pending | paid | failed | refunded
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    paid_at        TIMESTAMPTZ
);
CREATE INDEX idx_credit_purchases_user   ON credit_purchases(user_id);
CREATE INDEX idx_credit_purchases_status ON credit_purchases(status) WHERE status = 'pending';

-- +goose Down
DROP TABLE credit_purchases;
```

Проверка:
- [ ] `make dev-up` применяет миграцию без ошибок

---

### G.2 Конфиг (приоритет: HIGH)

В `backend/internal/config/config.go`:

```go
type PaymentConfig struct {
    Provider       string // PAYMENT_PROVIDER: "robokassa" | "enot" | "" (отключено)
    RobokassaLogin string // ROBOKASSA_LOGIN
    RobokassaPass1 string // ROBOKASSA_PASS1  — подпись при создании платежа
    RobokassaPass2 string // ROBOKASSA_PASS2  — подпись при верификации webhook
    EnotApiKey     string // ENOT_API_KEY
    EnotShopID     string // ENOT_SHOP_ID
    ReturnURL      string // PAYMENT_RETURN_URL  — куда вернуть игрока
}
```

Добавить `Payment PaymentConfig` в `Config`, читать в `Load()`.

Проверка:
- [ ] `PaymentConfig` в `Config`
- [ ] Все поля читаются из ENV

---

### G.3 Пакеты кредитов (приоритет: HIGH)

Файл `backend/internal/payment/packages.go`:

```go
package payment

type CreditPackage struct {
    Key          string
    Label        string
    Credits      int
    BonusCredits int
    PriceKop     int // цена в копейках: 4900 = 49.00 руб
}

func (p CreditPackage) TotalCredits() int { return p.Credits + p.BonusCredits }
func (p CreditPackage) PriceRub() float64 { return float64(p.PriceKop) / 100 }

var Packages = []CreditPackage{
    {Key: "trial",   Label: "Пробный",      Credits: 400,   BonusCredits: 0,    PriceKop: 4900},
    {Key: "starter", Label: "Стартовый",    Credits: 1000,  BonusCredits: 0,    PriceKop: 10000},
    {Key: "medium",  Label: "Средний",      Credits: 3000,  BonusCredits: 200,  PriceKop: 25000},
    {Key: "big",     Label: "Большой",      Credits: 7000,  BonusCredits: 500,  PriceKop: 50000},
    {Key: "max",     Label: "Максимальный", Credits: 15000, BonusCredits: 2000, PriceKop: 100000},
}

func PackageByKey(key string) (CreditPackage, bool) {
    for _, p := range Packages {
        if p.Key == key { return p, true }
    }
    return CreditPackage{}, false
}
```

---

### G.4 Backend: сервис (приоритет: HIGH)

#### Файлы

```
backend/internal/payment/
  packages.go      — G.3
  gateway.go       — Gateway интерфейс + sentinel errors
  robokassa.go     — RobokassaGateway
  service.go       — Service: CreateOrder, ConfirmPayment, ListPurchases
  handler.go       — HTTP handlers
```

#### `gateway.go`

```go
package payment

import (
    "context"
    "errors"
    "net/http"
)

var (
    ErrPackageNotFound  = errors.New("payment: unknown package key")
    ErrOrderNotFound    = errors.New("payment: order not found")
    ErrAlreadyPaid      = errors.New("payment: already paid")
    ErrWebhookInvalid   = errors.New("payment: webhook signature invalid")
    ErrGatewayDisabled  = errors.New("payment: no payment provider configured")
)

type Gateway interface {
    // BuildPayURL формирует ссылку для оплаты.
    BuildPayURL(ctx context.Context, orderID, description string, amountKop int, returnURL string) (string, error)
    // VerifyWebhook разбирает и верифицирует callback от шлюза.
    // Возвращает orderID и уплаченную сумму в копейках.
    VerifyWebhook(r *http.Request) (orderID string, amountKop int, err error)
}
```

#### `robokassa.go` — подпись MD5

```go
// BuildPayURL: MD5(login:amountRub.xx:invID:pass1) → URL merchant.robo.ru
// VerifyWebhook: POST-поля OutSum, InvId, SignatureValue
//   MD5(OutSum:InvId:pass2) == SignatureValue (case-insensitive)
```

#### `service.go`

```go
type ReferralProcessor interface {
    ProcessPurchase(ctx context.Context, buyerID string, amountRub float64) error
}

type Service struct {
    db       repo.Exec
    cfg      config.PaymentConfig
    gateway  Gateway               // nil если Provider == ""
    referral ReferralProcessor     // nil-safe
}

func NewService(db repo.Exec, cfg config.PaymentConfig) *Service

func (s *Service) WithReferral(r ReferralProcessor) *Service

// CreateOrder создаёт pending-запись в credit_purchases и возвращает URL оплаты.
// Если Provider не настроен — ErrGatewayDisabled.
func (s *Service) CreateOrder(ctx context.Context, userID, packageKey string) (payURL string, err error)

// ConfirmPayment вызывается из webhook: зачисляет кредиты, вызывает referral.
// Идемпотентен: повторный вызов с тем же orderID возвращает nil без двойного зачисления.
func (s *Service) ConfirmPayment(ctx context.Context, orderID string) error

// HandleWebhook — точка входа из handler.go: VerifyWebhook → ConfirmPayment → ответ шлюзу.
func (s *Service) HandleWebhook(w http.ResponseWriter, r *http.Request)

// ListPurchases возвращает историю покупок игрока (paid + pending).
func (s *Service) ListPurchases(ctx context.Context, userID string) ([]Purchase, error)

type Purchase struct {
    ID           string
    PackageKey   string
    PackageLabel string
    Credits      int
    PriceRub     float64
    Status       string
    CreatedAt    time.Time
    PaidAt       *time.Time
}
```

**Логика `ConfirmPayment` внутри транзакции:**
1. `SELECT status FROM credit_purchases WHERE id=$1 FOR UPDATE`
2. Если `status != 'pending'` → return nil (идемпотентность)
3. `UPDATE credit_purchases SET status='paid', paid_at=now(), provider_id=$2 WHERE id=$1`
4. `UPDATE users SET credit = credit + $amount WHERE id=$userID`
5. `s.referral.ProcessPurchase(...)` (вне транзакции, ошибки игнорируются)

#### `handler.go` — HTTP endpoints

```
POST /api/payment/order      — тело: {package_key}, ответ: {pay_url, order_id}
POST /api/payment/webhook    — callback Робокассы (без авторизации)
GET  /api/payment/packages   — публичный список пакетов
GET  /api/payment/history    — история покупок (требует авторизации)
```

Webhook не требует JWT, но проверяет подпись через `VerifyWebhook`.
При успехе Робокасса ожидает тело `OK{InvId}` (plaintext).

Проверка:
- [ ] `Gateway` интерфейс + `RobokassaGateway`
- [ ] `CreateOrder` создаёт pending в БД, возвращает pay_url
- [ ] `ConfirmPayment` идемпотентен, зачисляет кредиты
- [ ] Webhook верифицируется перед обработкой
- [ ] `make test` зелёный (тесты на `ConfirmPayment` + `RobokassaGateway.VerifyWebhook`)

---

### G.5 OpenAPI (приоритет: HIGH)

Добавить тег `payment` и схемы в `api/openapi.yaml`:

```yaml
# Schemas
CreditPackage:
  type: object
  properties:
    key:           { type: string }
    label:         { type: string }
    credits:       { type: integer }
    bonus_credits: { type: integer }
    total_credits: { type: integer }
    price_rub:     { type: number }

CreateOrderRequest:
  type: object
  required: [package_key]
  properties:
    package_key: { type: string }

CreateOrderResponse:
  type: object
  properties:
    order_id: { type: string }
    pay_url:  { type: string }

Purchase:
  type: object
  properties:
    id:            { type: string }
    package_key:   { type: string }
    package_label: { type: string }
    credits:       { type: integer }
    price_rub:     { type: number }
    status:        { type: string, enum: [pending, paid, failed, refunded] }
    created_at:    { type: string, format: date-time }
    paid_at:       { type: string, format: date-time, nullable: true }
```

Проверка:
- [ ] `make lint` (openapi-check) зелёный

---

### G.6 Frontend: экран пополнения (приоритет: HIGH)

Файл `frontend/src/features/payment/CreditsScreen.tsx`.

```
┌─ Пополнение кредитов ──────────────────────────────────────┐
│  Баланс: 💳 1 250 кр                                        │
│                                                              │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐       │
│  │ Пробный  │ │Стартовый │ │  Средний │ │  Большой │ ...   │
│  │  400 кр  │ │ 1 000 кр │ │ 3 200 кр │ │ 7 500 кр │       │
│  │  49 руб  │ │ 100 руб  │ │ 250 руб  │ │ 500 руб  │       │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘       │
│                                                              │
│  [Купить] → window.open(pay_url)                             │
│                                                              │
│  История покупок                                             │
│  04.05.2026  Стартовый  +1000 кр  ✅ оплачен                │
└──────────────────────────────────────────────────────────────┘
```

Запросы:
- `GET /api/payment/packages` → пакеты (публичный, кешируется TanStack Query staleTime=Infinity)
- `POST /api/payment/order` → `{pay_url}` → `window.open(pay_url, '_blank')`
- `GET /api/payment/history` → таблица покупок

Проверка:
- [ ] Карточки пакетов с ценой и кредитами
- [ ] Кнопка "Купить" открывает страницу оплаты в новой вкладке
- [ ] История покупок отображается

---

### G.7 Возврат игрока после оплаты (приоритет: MEDIUM)

ReturnURL = `https://game.example.com/?payment=success` (или `?payment=fail`).

В `App.tsx` при монтировании:
- `?payment=success` → toast "Оплата прошла, кредиты зачислены" + `invalidateQueries(['me'])`
- `?payment=fail` → toast "Оплата не прошла, попробуйте снова" (warning)
- После показа toast — убрать параметр из URL (`history.replaceState`)

Webhook обрабатывается раньше, чем игрок возвращается — кредиты уже на счету.

Проверка:
- [ ] `App.tsx` обрабатывает `?payment=success/fail`
- [ ] Toast отображается, параметр очищается из URL
- [ ] Баланс обновляется автоматически

---

## Порядок реализации

1. **G.1** — миграция
2. **G.2** — конфиг
3. **G.3 + G.4** — packages.go + gateway.go + robokassa.go + service.go + handler.go
4. **G.5** — OpenAPI схемы и пути
5. **G.6** — frontend CreditsScreen
6. **G.7** — обработка возврата в App.tsx

**Оценка:** ~2 итерации (backend отдельно, frontend отдельно).

---

## Что нужно до начала

- [ ] Зарегистрироваться на robokassa.com как самозанятый (НПД)
- [ ] Получить: `login`, `pass1`, `pass2` (раздел "Технические настройки" → пароли)
- [ ] Задать `PAYMENT_RETURN_URL` (нужен рабочий домен)
- [ ] (Позже) Enot.io как резервный шлюз — структура уже готова под него
