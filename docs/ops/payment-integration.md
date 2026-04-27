# Подключение платёжных шлюзов

Документация по настройке приёма платежей в oxsar-nova.
Актуально для самозанятых и физлиц (без ИП).

---

## Архитектура

План 38 + 42: платежи и кошельки живут в **billing-service** (отдельный
микросервис, `projects/billing/backend/`). Все шлюзы реализуют интерфейс
`Gateway` (`internal/payment/gateway.go`). Выбор активного шлюза — через
переменную `BILLING_PRIMARY_PROVIDER` (или legacy `PAYMENT_PROVIDER`).

Точки входа billing-сервиса:
- `GET /billing/packages` — список пакетов кредитов (публичный).
- `POST /billing/orders` — игрок создаёт заказ, получает `pay_url`.
  Требует JWT и поддерживает `Idempotency-Key` (Stripe-style).
- `POST /billing/webhooks/{provider}` — callback от шлюза. Защита по подписи
  (mock) или IP-allowlist + re-fetch (yookassa).
- `GET /billing/wallet/balance` — текущий баланс OXC игрока.
- `GET /billing/wallet/history` — история транзакций.

После успешной оплаты игрок возвращается на `PAYMENT_RETURN_URL`. Webhook
прилетает раньше → кредиты уже зачислены к моменту, когда юзер вернулся.

Поддерживаемые провайдеры:
- **`yookassa`** — основной (план 42). См. §1.
- **`mock`** — dev/test (план 38 Ф.4). См. §2.
- **`robokassa`** — резервный, **код пока не реализован** (был в game-nova,
  будет портирован при наличии тестового аккаунта). См. §3.

---

## Шлюз 1: ЮKassa (основной)

**Статус:** реализован (`projects/billing/backend/internal/payment/yookassa.go`).
**Подходит:** самозанятые (НПД), физлица.
**Комиссия:** от 2.8% (карты) до 3.5% (СБП с эквайрингом).
**Выплаты:** в день регистрации платежа.
**Чек ФНС:** автоматический для НПД (через API `receipt`).

### Шаг 1 — Регистрация

1. Зайти на [yookassa.ru](https://yookassa.ru/) → «Подключить».
2. Заполнить анкету. Для самозанятого (НПД) выбрать тип «Физическое лицо
   как самозанятый».
3. Привязать статус НПД через приложение «Мой налог»: в нём разрешить
   ЮKassa регистрировать чеки от вашего имени.
4. Подписать договор (электронно через ЕСИА/Госуслуги).

### Шаг 2 — Получить креды

В личном кабинете ЮKassa: «Настройки → API → Отображать секретные
ключи» → получаем:
- `shopId` (числовой идентификатор магазина)
- `secretKey` (рекомендуется создать **отдельный** для prod, не reuse тестовый)

### Шаг 3 — Настроить webhook URL

В ЛК → «Уведомления → Webhooks»:
- URL: `https://billing.oxsar-nova.ru/billing/webhooks/yookassa`
  (или другой публичный домен биллинга).
- События: отметить **`payment.succeeded`** (минимум). Дополнительно
  можно `payment.canceled` и `refund.succeeded`.

ЮKassa не использует HMAC-подпись webhook — только IP-allowlist + re-fetch
(оба механизма реализованы в `yookassa.go`).

### Шаг 4 — ENV переменные

В `deploy/docker-compose.multiverse.yml` (prod):
```yaml
billing-service:
  environment:
    BILLING_PRIMARY_PROVIDER: yookassa
    YOOKASSA_SHOP_ID: <shopId>          # из Docker secret
    YOOKASSA_SECRET_KEY: <secretKey>    # из Docker secret
    YOOKASSA_API_URL: https://api.yookassa.ru/v3
    PAYMENT_RETURN_URL: https://oxsar-nova.ru/?payment=success
```

ЮKassa-секреты подкладываются через Docker secrets (`/run/secrets/`),
не env vars (security: env видны через `docker inspect`).

### Шаг 5 — Тестовый магазин

Прежде чем включать prod-магазин, проверить flow в **тестовом** магазине
ЮKassa (отдельный shopId/secretKey):
- `YOOKASSA_API_URL=https://api.yookassa.ru/v3` (ЮKassa определяет тест/prod
  по самому магазину, не по URL).
- Прогнать сценарий: createOrder → редирект на confirmation_url → выбрать
  «Тестовая карта» → webhook → проверить `wallet.balance` += amount.
- Проверить, что в реальном магазине через сутки чек появляется в
  приложении «Мой налог».

### Технические особенности реализации

`internal/payment/yookassa.go`:
- `BuildPayURL` делает `POST /v3/payments` с `Idempotence-Key=order_id`,
  `metadata={order_id}`. Получает `confirmation_url` для редиректа.
- `VerifyWebhook` проверяет (1) IP клиента в YooKassa-allowlist
  (`YooKassaTrustedNetworks`), (2) re-fetch платежа через
  `GET /v3/payments/{id}` для подтверждения статуса. Принимает только
  `event=payment.succeeded`.
- `kopToRubString`/`rubStringToKop` — точная конвертация копейки↔"500.00"
  (без float).

### Чек НПД через API (`receipt`)

В текущей реализации поле `receipt` в `BuildPayURL` **пока NIL** — для
полноценного НПД-чека нужен email/phone клиента, которого нет в текущей
сигнатуре `Gateway.BuildPayURL`. Расширение интерфейса (передача
customer-данных) — в план 42 §2 (TODO до публичного запуска).

После расширения:
```go
receipt: {
  customer: { email: "user@x.com" },
  items: [{
    description: "Игровые кредиты, пакет 500 OXC",
    quantity: "1",
    amount: { value: "500.00", currency: "RUB" },
    vat_code: 1,                  // НДС 0% для самозанятого
    payment_subject: "service",
    payment_mode: "full_payment"
  }]
}
```

ЮKassa автоматически отправит чек в ФНС через интеграцию с «Мой налог».

---

## Шлюз 2: Mock (dev/test)

**Статус:** реализован (`projects/billing/backend/internal/payment/mock.go`).
**Назначение:** локальная разработка, E2E-тесты, demo.

В отличие от тривиальной заглушки, Mock использует **настоящую HMAC-SHA256
подпись** и проверку timestamp ±5 мин (replay protection). Это значит:
- Тестовый код, который имитирует webhook, обязан правильно подписать
  payload (тем же secret-ом).
- E2E-тест прогоняет тот же verify-flow, что и prod-шлюз → меньше
  «работает на тесте, ломается на проде».

ENV:
- `BILLING_PRIMARY_PROVIDER=mock`
- `PAYMENT_MOCK_SECRET=<random-string>` (в `deploy/docker-compose.yml`
  для dev).
- `PAYMENT_MOCK_BASE_URL=http://billing-service:9100` (для построения
  pay-URL в docker-network).

В prod-окружении НЕ использовать.

---

## Шлюз 3: Робокасса (резервный, не реализован)

**Статус:** реализован (`backend/internal/payment/robokassa.go`).  
**Подходит:** самозанятые (НПД), физлица.  
**Комиссия:** ~3%.  
**Выплаты:** мгновенно.

### Шаг 1 — Регистрация

1. Зайти на [robokassa.com](https://robokassa.com)
2. Нажать «Стать партнёром» → «Физическое лицо» или «Самозанятый»
3. Заполнить анкету, прикрепить документы (паспорт, справку о постановке на учёт НПД)
4. Дождаться проверки (обычно 1–3 дня)

### Шаг 2 — Создание магазина

1. В личном кабинете: Магазины → Добавить магазин
2. Заполнить:
   - Название магазина (отображается покупателю)
   - URL сайта (домен игры)
   - Тип товаров: **Цифровые товары / Игровая валюта**
3. В разделе «Технические настройки» магазина:
   - Найти **Пароль #1** (для создания платежей)
   - Найти **Пароль #2** (для верификации уведомлений)
   - Скопировать **Идентификатор (логин)**

### Шаг 3 — Настройка уведомлений

В настройках магазина, раздел «Уведомления»:
- **Result URL**: `https://ВАШ_ДОМЕН/api/payment/webhook`
  - Метод: POST
  - Убедиться что галочка «Отправлять уведомление» включена
- **Success URL**: `https://ВАШ_ДОМЕН/?payment=success`
- **Fail URL**: `https://ВАШ_ДОМЕН/?payment=fail`

> Робокасса сначала стучит на Result URL (webhook), и только потом
> редиректит игрока на Success URL — кредиты будут зачислены до возврата.

### Шаг 4 — ENV переменные

```env
PAYMENT_PROVIDER=robokassa
ROBOKASSA_LOGIN=ВАШ_ЛОГИН
ROBOKASSA_PASS1=ПАРОЛЬ_1
ROBOKASSA_PASS2=ПАРОЛЬ_2
PAYMENT_RETURN_URL=https://ВАШ_ДОМЕН/?payment=success
```

### Шаг 5 — Проверка

1. Запустить сервер с заданными ENV
2. Зайти в игру → таб «Кредиты»
3. Нажать «Купить» на любом пакете — должен открыться сайт Робокассы
4. В тестовом режиме (личный кабинет → Тест) провести оплату
5. Убедиться что в логах появился `payment: webhook` и кредиты зачислились

### Тестовый режим

В личном кабинете Робокассы есть переключатель **Тест / Боевой режим**.
В тестовом режиме реальные деньги не списываются. Пароли в тестовом и
боевом режиме — **разные**, не перепутать при переключении.

### Алгоритм подписи (справочно)

- Создание платежа: `MD5(Login:OutSum:InvId:Pass1)`
- Верификация webhook: `MD5(OutSum:InvId:Pass2)` — сравнивается с `SignatureValue`
- Сравнение регистронезависимое
- Ответ на успешный webhook: plaintext `OK{InvId}` (например, `OK42`)

---

## Шлюз 2: Enot.io (резервный)

**Статус:** структура подготовлена в конфиге, файл `enot.go` не написан —
добавить когда понадобится второй шлюз.  
**Подходит:** самозанятые, физлица без ИП.  
**Комиссия:** 2–4%.

### Шаг 1 — Регистрация

1. Зайти на [enot.io](https://enot.io)
2. Зарегистрироваться, выбрать тип «Самозанятый»
3. Верификация обычно быстрая (до 24 часов)

### Шаг 2 — Получить ключи

В личном кабинете:
- Мои магазины → Создать магазин
- Скопировать **Shop ID** и **Secret Key** (он же API Key)

### Шаг 3 — Настройка уведомлений

В настройках магазина:
- **Notify URL**: `https://ВАШ_ДОМЕН/api/payment/webhook`
- **Success URL**: `https://ВАШ_ДОМЕН/?payment=success`
- **Fail URL**: `https://ВАШ_ДОМЕН/?payment=fail`

### Шаг 4 — ENV переменные

```env
PAYMENT_PROVIDER=enot
ENOT_API_KEY=ВАШ_SECRET_KEY
ENOT_SHOP_ID=ВАШ_SHOP_ID
PAYMENT_RETURN_URL=https://ВАШ_ДОМЕН/?payment=success
```

### Шаг 5 — Реализовать enot.go

Создать `backend/internal/payment/enot.go` реализующий `Gateway`:

```go
// BuildPayURL: POST https://api.enot.io/invoice/create
//   Headers: x-api-key: <apiKey>
//   Body: { shop_id, amount, order_id, currency: "RUB", description }
//   Response: { url }

// VerifyWebhook: проверить HMAC-SHA256 подпись
//   sig = HMAC-SHA256(payload, apiKey)
//   Header: x-sign

// SuccessResponse: HTTP 200 OK (Enot не требует специального тела)
```

После реализации поменять `switch cfg.Provider` в `service.go`:
```go
case "enot":
    svc.gateway = NewEnotGateway(cfg.EnotApiKey, cfg.EnotShopID)
```

---

## Общие замечания

### Лимит самозанятого

НПД позволяет получать до **2 400 000 руб/год**. При превышении —
смена налогового режима (ИП на УСН). Шлюзы не блокируют автоматически,
следить вручную.

### Тест до боевого режима

Всегда проверять полный цикл в тестовом режиме шлюза:
1. Создать заказ через игру
2. Провести тестовую оплату
3. Убедиться что webhook дошёл (лог сервера) и кредиты зачислились
4. Проверить таблицу `credit_purchases`: статус должен смениться `pending → paid`

### Переключение шлюза

Менять `PAYMENT_PROVIDER` можно без перекомпиляции — только рестарт сервера.
Незакрытые `pending`-заказы предыдущего шлюза останутся в БД без зачисления
(webhook не придёт от другого шлюза). При смене шлюза в бою — закрыть
старые заказы вручную или дождаться их истечения.

### Безопасность webhook

- Endpoint `/api/payment/webhook` открыт без JWT — доступен без авторизации
- Защита — только через проверку подписи (`VerifyWebhook`)
- Идемпотентность: повторный webhook с тем же `provider_id` игнорируется
  (поле `UNIQUE` в таблице `credit_purchases`)
- Логировать все входящие webhook с `remote_addr` для аудита
