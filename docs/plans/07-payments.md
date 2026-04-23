# План G: Платёжная система

---

## Выбор платёжного шлюза

### Сравнение для физлица/ИП (Россия, 2025)

| Шлюз | Физлицо | Комиссия | Выплаты | API | Особенности |
|------|---------|----------|---------|-----|-------------|
| **Робокасса** | ✅ ИП, самозанятый | ~3% | Мгновенно | REST + webhook | 130+ методов оплаты, специализация на цифровых товарах, СБП |
| **ЮКасса** | ✅ ИП | 2–3% | На счёт/СБП | REST SDK (Go, Python, PHP) | Самая известная, хорошая документация |
| **Enot.io** | ✅ ИП, физлицо | 2–4% | Авто-выплаты | REST + webhook | Простая регистрация, быстрый старт |
| **CloudPayments** | ✅ ИП | 2–3% | По расписанию | REST | Хорош для подписок, рекуррентные платежи |
| **Unitpay** | ✅ ИП | 2–4% | Ежедневно | REST | Агрегатор, много методов оплаты |

**Рекомендация: Робокасса** — прямая поддержка самозанятых, мгновенные выплаты, специализация на цифровых товарах (игровые кредиты), подробная документация на русском.

**Запасной вариант: ЮКасса** — Go SDK, более строгая проверка документов, но надёжнее.

---

## Пакеты кредитов (из плана F.1)

| Пакет | Кредиты | Цена | Бонус |
|-------|---------|------|-------|
| Пробный | 400 | 49 руб | — |
| Стартовый | 1 000 | 100 руб | — |
| Средний | 3 000 | 250 руб | +200 кр |
| Большой | 7 000 | 500 руб | +500 кр |
| Максимальный | 15 000 | 1 000 руб | +2 000 кр |

---

## Открытые задачи

### G.1 Миграция и модели данных (приоритет: HIGH)

**Шаг 1** — `migrations/0049_credit_purchases.sql`:
```sql
CREATE TABLE credit_purchases (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    package_key  TEXT NOT NULL,        -- "trial", "starter", "medium", "big", "max"
    amount_credits INT  NOT NULL,      -- кредиты с бонусом
    price_rub    NUMERIC(10,2) NOT NULL,
    provider     TEXT NOT NULL,        -- "robokassa" | "yookassa"
    provider_id  TEXT,                 -- ID транзакции от шлюза
    status       TEXT NOT NULL DEFAULT 'pending',  -- pending | paid | failed | refunded
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    paid_at      TIMESTAMPTZ
);
CREATE INDEX idx_credit_purchases_user ON credit_purchases(user_id);
CREATE INDEX idx_credit_purchases_provider ON credit_purchases(provider, provider_id);
```

**Проверка готовности:**
- [ ] Миграция применяется без ошибок
- [ ] `expCredit` в `fleet/expedition.go` читает из `credit_purchases` (исправляет B.2)

---

### G.2 Конфиг платёжного шлюза (приоритет: HIGH)

**Шаг 1** — `backend/internal/config/config.go`, добавить `PaymentConfig`:
```go
type PaymentConfig struct {
    Provider      string  // PAYMENT_PROVIDER: "robokassa" | "yookassa" | ""
    RobokassaLogin  string  // ROBOKASSA_LOGIN
    RobokassaPass1  string  // ROBOKASSA_PASS1 (для создания платежей)
    RobokassaPass2  string  // ROBOKASSA_PASS2 (для верификации webhook)
    YookassaShopID  string  // YOOKASSA_SHOP_ID
    YookassaKey     string  // YOOKASSA_SECRET_KEY
    ReturnURL       string  // PAYMENT_RETURN_URL (куда вернуть игрока после оплаты)
    WebhookSecret   string  // PAYMENT_WEBHOOK_SECRET (для HMAC-верификации)
}
```

Добавить `Payment PaymentConfig` в `Config` и читать в `Load()`.

**Проверка готовности:**
- [ ] `PaymentConfig` в config.go
- [ ] Все поля читаются из ENV

---

### G.3 Пакеты кредитов в конфиге (приоритет: HIGH)

**Шаг 1** — `backend/internal/payment/packages.go`:
```go
type CreditPackage struct {
    Key          string
    Label        string
    Credits      int
    BonusCredits int
    PriceRub     int  // в копейках (4900 = 49.00 руб)
}

var Packages = []CreditPackage{
    {Key: "trial",   Label: "Пробный",      Credits: 400,   BonusCredits: 0,    PriceRub: 4900},
    {Key: "starter", Label: "Стартовый",    Credits: 1000,  BonusCredits: 0,    PriceRub: 10000},
    {Key: "medium",  Label: "Средний",      Credits: 3000,  BonusCredits: 200,  PriceRub: 25000},
    {Key: "big",     Label: "Большой",      Credits: 7000,  BonusCredits: 500,  PriceRub: 50000},
    {Key: "max",     Label: "Максимальный", Credits: 15000, BonusCredits: 2000, PriceRub: 100000},
}
```

---

### G.4 Backend: сервис и шлюзы (приоритет: HIGH)

**`backend/internal/payment/` — структура:**
```
payment/
  packages.go     — пакеты кредитов (G.3)
  service.go      — CreateOrder, ConfirmPayment, ListPurchases
  robokassa.go    — реализация для Робокассы
  yookassa.go     — реализация для ЮКассы
  handler.go      — HTTP endpoints
```

**`payment/service.go`:**
```go
type Service struct {
    db      repo.Exec
    cfg     config.PaymentConfig
    gateway Gateway
    referral *referral.Service  // для ProcessPurchase
}

type Gateway interface {
    BuildPayURL(orderID, description string, amountKop int, returnURL string) (string, error)
    VerifyWebhook(r *http.Request, pass2 string) (orderID string, amountKop int, ok bool)
}

// CreateOrder — создаёт pending-запись и возвращает URL оплаты.
func (s *Service) CreateOrder(ctx, userID, packageKey string) (payURL string, err error)

// ConfirmPayment — вызывается из webhook, зачисляет кредиты.
func (s *Service) ConfirmPayment(ctx, orderID string) error

// ListPurchases — история покупок игрока.
func (s *Service) ListPurchases(ctx, userID string) ([]Purchase, error)
```

**`payment/robokassa.go`:** подпись через MD5(login:amount:inv_id:pass1).

**`payment/yookassa.go`:** Basic Auth (shopID:secretKey), JSON API.

**`payment/handler.go`:**
```
POST /api/payment/order       — создать заказ, вернуть {pay_url}
POST /api/payment/webhook     — callback от шлюза (не требует авторизации)
GET  /api/payment/packages    — список пакетов (публичный)
GET  /api/payment/history     — история покупок игрока
```

**Webhook-безопасность:**
- Робокасса: MD5(amount:inv_id:pass2) — verifyWebhook
- ЮКасса: заголовок `X-YooMoney-Signature` (HMAC-SHA256)
- Идемпотентность: повторный webhook с тем же `provider_id` → ранний выход

**Проверка готовности:**
- [ ] `Gateway` интерфейс + RobokassaGateway + YookassaGateway
- [ ] `CreateOrder`: создаёт pending в БД, возвращает pay_url
- [ ] `ConfirmPayment`: идемпотентен, зачисляет кредиты, вызывает referral.ProcessPurchase
- [ ] Webhook верифицируется перед обработкой
- [ ] `make test` зелёный

---

### G.5 Frontend: экран пополнения (приоритет: HIGH)

**`frontend/src/features/payment/CreditsScreen.tsx`:**

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
│  [Купить] → redirect на страницу оплаты шлюза               │
│                                                              │
│  История покупок                                             │
│  04.05.2026  Стартовый  +1000 кр  ✅ оплачен                │
└──────────────────────────────────────────────────────────────┘
```

**Шаги:**
- `GET /api/payment/packages` → список пакетов с ценами
- `POST /api/payment/order` → `{pay_url}` → `window.open(pay_url)`
- `GET /api/payment/history` → таблица покупок
- После возврата игрока (ReturnURL) — обновить баланс кредитов

**Проверка готовности:**
- [ ] `CreditsScreen.tsx` показывает пакеты и историю
- [ ] Кнопка "Купить" открывает страницу оплаты
- [ ] После оплаты баланс обновляется

---

### G.6 Возврат игрока после оплаты (приоритет: MEDIUM)

**ReturnURL** = `https://game.example.com/?payment=success` (или `?payment=fail`).

В App.tsx при старте: если `?payment=success` в URL — показать toast "Оплата прошла успешно,
кредиты зачислены" и обновить баланс через `queryClient.invalidateQueries(['me'])`.

Webhook срабатывает раньше, чем игрок вернётся — кредиты уже будут на счету.

**Проверка готовности:**
- [ ] App.tsx обрабатывает `?payment=success/fail` параметр
- [ ] Toast отображается при возврате
- [ ] Баланс обновляется автоматически

---

## Порядок реализации

1. **G.1** — миграция (устраняет B.2 из плана 02)
2. **G.2 + G.3** — конфиг + пакеты
3. **G.4** — backend сервис (начать с Робокассой)
4. **G.5** — frontend экран
5. **G.6** — обработка возврата

**Общая оценка:** ~2–3 итерации. Начать с Робокассой (быстрее подключить для ИП/самозанятого).

---

## Что нужно до начала

- [ ] Зарегистрироваться на robokassa.com (или yookassa.ru) как ИП/самозанятый
- [ ] Получить: `login`, `pass1`, `pass2` (Робокасса) или `shop_id` + `secret_key` (ЮКасса)
- [ ] Задать `PAYMENT_RETURN_URL` (домен игры должен быть готов)
