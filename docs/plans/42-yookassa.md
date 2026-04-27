# План 42: ЮKassa как основной платёжный провайдер

**Дата**: 2026-04-26
**Статус**: Активный
**Зависимости**: план 38 (billing-service, Gateway-интерфейс) — фазы Ф.1–Ф.4
завершены, mock-провайдер работает.
**Связанные планы**: [07-payments.md](07-payments.md) — корректируется
(Robokassa → ЮKassa как основной), [38-billing-service.md](38-billing-service.md) —
Ф.7 (frontend) подхватит новый провайдер.

---

## Цель

Подключить ЮKassa как основной платёжный шлюз billing-сервиса:

- автоматическая регистрация чека в "Мой налог" через API ЮKassa для
  самозанятого (НПД) — без ручной выписки чеков;
- создание платежа через ЮKassa API, обработка webhook'ов
  (`payment.succeeded`, `payment.canceled`);
- поддержка СБП, банковских карт, кошельков;
- сохранение совместимости с Mock-провайдером для тестов.

Robokassa и Enot.io остаются опциональными резервными провайдерами (через
тот же `Gateway` интерфейс), но не реализуются в рамках этого плана.

---

## Что меняем

### 1. `projects/billing/backend/internal/payment/yookassa.go` (новый)

Реализация интерфейса `Gateway` для ЮKassa:

- `CreatePayment(ctx, order)` — POST `/v3/payments` к ЮKassa API с указанием
  суммы, описания, return_url, метаданных (order_id), флагом
  `capture: true`. Идемпотентность через `Idempotence-Key` header (UUID
  заказа).
- `HandleWebhook(ctx, body, headers)` — проверка подписи (если используется),
  парсинг события (`payment.succeeded` / `payment.canceled` /
  `payment.waiting_for_capture`), извлечение `payment.id` и `metadata.order_id`.
- `Refund(ctx, paymentID, amount)` — POST `/v3/refunds` для возвратов.

Авторизация — Basic Auth (shopId + secretKey) через ENV-переменные
`YOOKASSA_SHOP_ID`, `YOOKASSA_SECRET_KEY`.

### 2. Чек для самозанятого (НПД) через API ЮKassa

В запросе `CreatePayment` передавать объект `receipt` с:
- `customer.email` (или `customer.phone`) — обязательно для НПД;
- `items[]` с описанием "Игровые кредиты, пакет N", `quantity`,
  `amount.value`, `vat_code: 1` (без НДС для самозанятого),
  `payment_subject: service`, `payment_mode: full_payment`.

ЮKassa автоматически передаёт чек в ФНС для самозанятых при правильно
заполненном `receipt`. Проверка после первой реальной транзакции в режиме
тест-магазина: чек должен появиться в приложении "Мой налог" в течение
суток.

### 3. ENV-конфиг

Добавить в `projects/billing/backend/cmd/server/main.go` чтение:
- `YOOKASSA_SHOP_ID`
- `YOOKASSA_SECRET_KEY`
- `YOOKASSA_RETURN_URL` — куда редиректить пользователя после оплаты
- `BILLING_PRIMARY_PROVIDER` — `yookassa` | `mock` | `robokassa`

Регистрация провайдера в фабрике `payment.NewGateway(name string)` —
`case "yookassa": return NewYooKassaGateway(...)`.

### 4. Webhook endpoint

Существующий `POST /billing/webhooks/{provider}` поддержит `provider=yookassa`
без изменений — роутинг по имени уже есть. Реализация
`HandleWebhook` парсит формат ЮKassa (отличается от Mock).

### 5. Frontend (минимально)

В рамках плана 38 Ф.7 (frontend для billing) — добавить кнопку "Оплатить",
которая вызывает backend `POST /billing/orders` и редиректит пользователя
на `confirmation_url` из ответа ЮKassa.

В этом плане frontend **не делается** — он часть 38 Ф.7. Здесь только
backend готов отдать `confirmation_url` в ответе на создание заказа.

### 6. Корректировка плана 07-payments.md

Заменить в таблице "Сравнение шлюзов" и в строке "Выбор":
- основной: ЮKassa (вместо Robokassa);
- резервные: Robokassa, Enot.io (по необходимости).

Указать, что выбор ЮKassa обусловлен:
- автоматический чек в "Мой налог" для самозанятого через API;
- широкая поддержка методов оплаты (карты, СБП, кошельки);
- зрелая документация и SDK.

### 7. Документация ops

В `docs/ops/payment-integration.md` добавить раздел "ЮKassa":
- регистрация магазина в личном кабинете ЮKassa;
- получение `shopId` и `secretKey`;
- настройка webhook URL в ЛК ЮKassa;
- привязка к статусу самозанятого (через ЛК "Мой налог" → разрешение
  ЮKassa регистрировать чеки от имени НПД-плательщика);
- тестовый магазин для отладки.

---

## Этапы

### Ф.1. Backend-провайдер ЮKassa

- Реализовать `yookassa.go` (CreatePayment, HandleWebhook, Refund).
- Добавить ENV-переменные и регистрацию провайдера в фабрике.
- Unit-тесты: парсинг webhook, формирование запроса CreatePayment с
  корректным `receipt` для самозанятого.

### Ф.2. Чек для НПД — проверка end-to-end

- Развернуть тестовый магазин ЮKassa (без реальных платежей).
- Прогнать сценарий: создание заказа → редирект на ЮKassa → тестовая
  оплата → webhook → запись в `payment_orders` → выдача кредитов.
- Проверить, что в режиме реального магазина чек появляется в "Мой налог".

### Ф.3. Документация

- Обновить `docs/plans/07-payments.md` (Robokassa → ЮKassa).
- Добавить раздел "ЮKassa" в `docs/ops/payment-integration.md` со
  списком шагов настройки.
- Обновить README billing-service (если есть) с примером ENV.

### Ф.4. Финализация

- `git status` — в индексе только файлы плана 42, чужого не зацеплять.
- Обновить `docs/project-creation.txt`: "план 42: ЮKassa как основной
  платёжный шлюз с автоматическим чеком в Мой налог".
- Коммит: `feat(billing): подключить ЮKassa с автоматическим чеком НПД`.

### Ф.5. Не утекать имя провайдера в публичные URL

**Принцип:** имя платёжного провайдера — это backend-деталь. Frontend
оперирует только обобщённым `?payment=success` / `?payment=fail`,
не знает про yookassa/robokassa/mock.

**Что нельзя:**
- redirect на `?yookassa_status=...` или `?robokassa_status=...` после
  оплаты;
- параметры с именем провайдера в `PAYMENT_RETURN_URL`;
- упоминания `yookassa` в `confirmation_url` (мы не выбираем confirmation_url
  у YooKassa — она его сама строит, эта часть unavoidable).

**Что нужно:**
- yookassa-mock-сервер при redirect передаёт `ReturnURL` **as-is**, без
  добавочных query-params (как делает реальная YooKassa).
- Если frontend хочет знать «оплата только что прошла» — он узнаёт это:
  - либо через `?payment=success` (если реальная YooKassa добавит — нужно
    сконфигурировать `PAYMENT_RETURN_URL=...?payment=success`);
  - либо через polling баланса после возврата на /shop (TanStack Query
    refetchInterval).
- В UI после редиректа: invalidate `['billing','balance']` query, обновить
  badge в шапке.

**Реализовано (2026-04-27):**
- yookassa-mock `handleCheckoutPay` / `handleCheckoutCancel` редиректят
  на `ReturnURL` без `?yookassa_status=...`.
- Frontend (план 38 Ф.7): App.tsx уже читает `?payment=success` и
  показывает toast — этого достаточно. yookassa_status больше не приходит.

---

## Тестирование

- Unit-тесты на `yookassa.go`: формирование запроса, разбор webhook,
  валидация подписи.
- Integration-тест с тестовым магазином ЮKassa (один прогон сценария
  оплаты).
- Проверить, что Mock-провайдер по-прежнему работает (регрессия).
- Проверить, что переключение `BILLING_PRIMARY_PROVIDER` действительно
  меняет провайдера без изменений в остальном коде.

---

## Итог

Один новый Go-файл (~200–300 строк) + ENV-конфиг + правки документации.
Robokassa и Enot.io остаются возможными провайдерами через `Gateway`
интерфейс — могут быть добавлены позже без переделок. Frontend-часть —
в рамках плана 38 Ф.7.
