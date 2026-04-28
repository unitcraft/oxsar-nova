# План 77 (инфра): billing-client integration в game-nova

**Дата**: 2026-04-28
**Статус**: ✅ Завершён (2026-04-28)
**Зависимости**: нет блокирующих. **Разблокирует** план 65 Ф.6
(KindTeleportPlanet) + план 66 Ф.5 (платный выкуп удержания) +
потенциально план 68 (биржа артефактов, если понадобится продажа
лотов за оксары).
**Связанные документы**:
- [docs/plans/38-billing-service.md](38-billing-service.md) —
  billing-сервис (закрыт), источник API.
- [docs/adr/0009-currency-rebranding.md](../adr/0009-currency-rebranding.md) —
  оксары (hard) vs оксариты (soft), что списывать через billing.
- [projects/portal/backend/internal/portalsvc/credits.go](../../projects/portal/backend/internal/portalsvc/credits.go) —
  существующий billing-client в portal-backend (паттерн).
- [docs/research/origin-vs-nova/roadmap-report.md](../research/origin-vs-nova/roadmap-report.md) —
  правила R0-R15 серии ремастера.

---

## Цель

Добавить в **game-nova-backend** billing-client для списания
**оксаров** (hard, ст. 437 ГК) с пользовательских кошельков через
billing-сервис.

Без этого блокированы планы 65 Ф.6 (KindTeleportPlanet — премиум-
телепорт планеты) и 66 Ф.5 (платный выкуп удержания инопланетянами).

Текущее состояние game-nova: **нет интеграции** с billing-сервисом
(только локальные `users.credits`, не `wallets.oxsar`). Аналогичный
client в portal-backend (`portalsvc/credits.go`) есть и работает —
будет образцом.

---

## Контекст

### Зачем нужен

Сейчас в game-nova **нет** способа списать **оксары** (реальные
деньги, hard-валюта по ADR-0009). Только локальные `users.credits`
для оксаритов (soft).

Несколько фич ремастера требуют списания оксаров:

1. **План 65 Ф.6 KindTeleportPlanet** — премиум-фича через оксары
   (ADR-0009 §«Что покупается за оксары»: премиум-эффекты).
2. **План 66 Ф.5 платный выкуп удержания** — игрок может
   досрочно выкупить свою планету у инопланетян за оксары
   (классическая premium-механика).
3. **План 68 биржа артефактов** — лоты продаются за **оксариты**
   (soft, юр-чисто, ст. 1062 ГК). НЕ требует billing-client. Но
   **покупка premium-permit «Знак торговца»** может потребовать
   оксары — это решит план 68 при реализации.
4. **Будущие премиум-фичи** — AI-советник, ускорение строительства,
   защита планеты, и т.д.

### Что уже работает (источник истины)

В **portal-backend** есть `internal/portalsvc/credits.go`
`BillingClient.Spend(ctx, SpendInput)`:

- HTTP POST `/billing/wallet/spend` с body `{amount, reason, ref_id,
  to_account}`.
- JWT-forwarding через `Authorization: Bearer <UserToken>`.
- Idempotency-Key header.
- Обработка статусов: 200 OK / 402 Insufficient / прочее → error.

Этот образец **переиспользуем** в game-nova, расширим под нужды
ремастера.

### Чего нет в game-nova

- HTTP-клиента к billing.
- Idempotency-Key middleware в API-роутере (есть только в
  billing-сервисе).
- Конфигурации `BILLING_URL` env-переменной.
- Тестовой инфраструктуры для мока billing.

---

## Что меняем

### 1. Новый пакет `internal/billing/client/`

```go
// projects/game-nova/backend/internal/billing/client/client.go

package client

import (
    "context"
    "errors"
    // ...
)

// Client — HTTP-клиент к billing-сервису для списаний оксаров.
type Client struct {
    billingURL string
    httpClient *http.Client
}

func New(billingURL string) *Client { ... }

// SpendInput — параметры списания.
type SpendInput struct {
    UserToken      string // JWT юзера (forward из request header)
    Amount         int64  // сумма в копейках оксаров (R1: целые, не float)
    Reason         string // символическое описание ("teleport_planet",
                          // "alien_buyout", "marketplace_permit")
    RefID          string // ID связанной сущности (planet_id, holding_id, lot_id)
    ToAccount      string // куда зачислится ("system:teleport",
                          // "system:alien_buyout")
    IdempotencyKey string // обязательно для R9
}

// Spend списывает amount оксаров с user-кошелька через
// billing-сервис POST /billing/wallet/spend.
//
// Возвращает:
//   - nil при успехе (HTTP 200)
//   - ErrInsufficientOxsar при HTTP 402 (недостаточно оксаров)
//   - ErrBillingUnavailable при сетевой ошибке (timeout, refuse)
//   - ErrIdempotencyConflict при HTTP 409 (тот же Idempotency-Key,
//     но другие параметры — клиент-баг)
//   - error прочее
func (c *Client) Spend(ctx context.Context, in SpendInput) error { ... }

// Refund возвращает amount оксаров на user-кошелёк (для отмены
// telepor'а, отказа от buyout и т.п.). Использует тот же
// Idempotency-Key с суффиксом ":refund".
func (c *Client) Refund(ctx context.Context, in SpendInput) error { ... }
```

Sentinel errors:

```go
// ErrInsufficientOxsar — у юзера не хватает оксаров (HTTP 402).
// Доменная ошибка, проксируется клиенту как 402.
var ErrInsufficientOxsar = errors.New("billing: insufficient oxsar")

// ErrBillingUnavailable — billing-сервис недоступен (timeout/refuse).
// Транзиентная, можно retry.
var ErrBillingUnavailable = errors.New("billing: service unavailable")

// ErrIdempotencyConflict — клиентский баг (тот же ключ, другие данные).
var ErrIdempotencyConflict = errors.New("billing: idempotency conflict")
```

### 2. Idempotency-middleware в API-роутер game-nova

`pkg/idempotency/middleware.go` — Chi-middleware для проверки
`Idempotency-Key` header на мутирующих endpoint'ах.

Логика:
- Извлечь header.
- Если есть — проверить в Redis: был ли уже запрос с этим ключом
  и тем же body-hash?
  - Да + body совпадает → вернуть кешированный ответ (идемпотентно).
  - Да + body отличается → 409 (клиент-баг).
  - Нет → выполнить handler, сохранить ответ в Redis с TTL 24h.
- Если нет header (опционально) → выполнить как обычно (idempotency
  отключена для этого вызова).

```go
// pkg/idempotency/middleware.go
package idempotency

import (
    "github.com/redis/go-redis/v9"
    "net/http"
)

type Middleware struct {
    rdb *redis.Client
    ttl time.Duration
}

func New(rdb *redis.Client) *Middleware { ... }

// Wrap оборачивает handler в idempotency-проверку.
// Если Idempotency-Key отсутствует — handler вызывается как есть.
func (m *Middleware) Wrap(next http.HandlerFunc) http.HandlerFunc { ... }
```

**Зависимость от Redis**: в game-nova Redis уже подключён (план 32
multi-instance scheduler). Переиспользуем тот же клиент.

### 3. Конфигурация

В `cmd/server/main.go`:
- Новая env-переменная `BILLING_URL` (например,
  `http://billing-service:8080`).
- Инициализация `billing/client.New(billingURL)`.
- Регистрация idempotency-middleware с redis-клиентом.
- Прокидывание клиента в handler'ы плана 65 Ф.6 / 66 Ф.5.

В `deploy/docker-compose.multiverse.yml`:
- Добавить env-переменную `BILLING_URL` для game-nova-backend
  сервиса.

### 4. Тесты

#### 4.1. Unit-тесты client'а

С `httptest.NewServer` мокаем billing:
- 200 OK → `Spend` возвращает nil.
- 402 → `ErrInsufficientOxsar`.
- 409 → `ErrIdempotencyConflict`.
- 500 → generic error.
- timeout → `ErrBillingUnavailable`.
- Retry-логика: один retry на 503/504.

#### 4.2. Unit-тесты idempotency-middleware

С mock Redis:
- Первый запрос → handler выполняется, ответ сохранён.
- Повторный запрос с тем же ключом + body → возврат кешированного
  ответа.
- Повторный запрос с тем же ключом + другим body → 409.
- Запрос без header → handler выполняется, ничего не сохраняется.
- TTL истёк → следующий запрос с тем же ключом выполняется заново.

#### 4.3. Integration-тест end-to-end

Поднять billing-сервис (или mock) + game-nova:
- Создать тестового пользователя с кошельком 1000 оксаров.
- Сделать запрос на «гипотетический» premium-endpoint с Idempotency-Key.
- Проверить: оксары списались, ответ корректный.
- Сделать тот же запрос → возврат кешированного ответа, оксары
  не списываются повторно.

### 5. Документация

- `internal/billing/client/doc.go` — краткое описание пакета.
- В `docs/release-roadmap.md` обновить: «План 77 — billing-client
  для game-nova, разблокирует premium-фичи ремастера».

---

## Чего НЕ делаем

- **Не вводим** wallet management в game-nova (создание кошельков,
  чтение баланса). Это домен billing-сервиса. game-nova **только
  списывает** через client.
- **Не реализуем** оксариты (soft) через billing — оксариты живут
  в `game-nova.users.oxsarit` и `oxsarit_transactions` (план 58).
  Это локальный domain game-nova, billing не нужен.
- **Не переписываем** существующий portal `BillingClient` — он
  работает, не трогаем (можно потом унифицировать в общий пакет
  `pkg/billingclient`, если хотим, но это отдельная задача).
- **Не изменяем** billing-сервис (план 38 закрыт). Только используем
  его API.
- **Не вводим** webhook-callback от billing (например, при асинхронном
  возврате) — billing /wallet/spend синхронный.

---

## Этапы

### Ф.1. Скаффолд клиента + sentinel errors

- Создать `internal/billing/client/` с `client.go` (тип `Client`,
  методы `Spend`/`Refund`) и `errors.go` (sentinel errors).
- Скопировать структуру из `portal/internal/portalsvc/credits.go`
  с расширением (Refund, ErrIdempotencyConflict).
- Никаких handler'ов пока — только пакет.
- Unit-тесты с httptest мок billing.
- `go build` + `go test ./internal/billing/client/...` — зелёные.

### Ф.2. Idempotency-middleware

- `pkg/idempotency/middleware.go` — Chi-middleware с Redis-cache.
- Конфиг TTL (по умолчанию 24h).
- Body-hash через SHA-256.
- Unit-тесты с mock Redis (через `redismock` или поднятый
  redis-test).

### Ф.3. Интеграция в роутер

- В `cmd/server/main.go`:
  - Чтение env `BILLING_URL`.
  - Инициализация `billing.NewClient(billingURL)`.
  - Регистрация idempotency-middleware (пока не привязан к
    конкретным роутам — это сделают планы 65 Ф.6 / 66 Ф.5).
  - Прокидывание client в handler factory для planet/alien.
- `deploy/docker-compose.multiverse.yml` — env `BILLING_URL`.

### Ф.4. Integration-тест end-to-end

- Тестовый сценарий: оксары списываются через client, idempotency
  работает (повторный запрос не списывает дважды).
- Использовать `cmd/yookassa-mock` или поднимать billing-сервис в
  test-стеке.

### Ф.5. Документация и финал

- `internal/billing/client/doc.go`.
- Обновить шапку плана 77 → ✅.
- Запись в `docs/project-creation.txt` — итерация 77.
- Обновить release-roadmap «Пост-запуск v3» — отметить что 77
  разблокировал 65 Ф.6 и 66 Ф.5.

---

## Тестирование

- Ф.1: unit-тесты client'а ≥ 85% покрытие (по R4 для billing-логики).
- Ф.2: unit-тесты middleware ≥ 85%.
- Ф.4: integration end-to-end зелёный.
- Все existing nova-тесты остаются зелёными (R0).

---

## Объём

- ~400-600 строк Go (client + middleware + тесты).
- 1-2 коммита.
- 1-2 недели работы агента в активном темпе.

**Время выполнения**: ~1-2 недели.

---

## Когда запускать

**Сейчас**, потому что разблокирует:
- 65 Ф.6 KindTeleportPlanet
- 66 Ф.5 платный выкуп удержания

Можно делать **параллельно** с:
- 67 Ф.5 frontend (alliance UI) — разные пакеты.
- 66 Ф.6 golden-итерации — разные пакеты.
- 71 UX-микрологика — другой проект (frontend).

---

## КОНВЕНЦИИ ИМЕНОВАНИЯ (R1-R15)

- R0: не трогаем nova-баланс/механики, только инфра-код.
- R1: пакет `internal/billing/client` (snake_case структурно,
  английский), типы `Client`, `SpendInput` (PascalCase).
- R6: REST-style HTTP вызовы, JSON body (POST /billing/wallet/spend
  уже определён в billing-сервисе, переиспользуем).
- R8: **обязательно** Prometheus метрики
  (`oxsar_billing_client_spend_total{status}` +
  `oxsar_billing_client_duration_seconds`). Не пропускать.
- R9: Idempotency-Key — это **ключевой** функционал плана.
- R12: i18n не применимо (нет user-facing строк, только error
  messages в логи).
- R13: typed payload (SpendInput) ✅.
- R15: без упрощений — обработка всех edge-кейсов (timeout,
  402, 409, 500, transient errors с retry).

---

## Известные риски

| Риск | Митигация |
|---|---|
| billing-сервис недоступен на момент списания (network failure) | Sentinel `ErrBillingUnavailable` + явная политика на стороне handler'а: telepor план 65 — отказать игроку с retry; alien-buyout план 66 — разрешить retry в течение 5 минут |
| Idempotency-Key collision между разными игроками | Ключ генерируется как `<user_id>:<operation_type>:<ref_id>` — uniqueness гарантирована |
| Race condition при параллельных запросах с тем же ключом | Redis SET NX в middleware — атомарная блокировка ключа |
| Timeout слишком долгий → игрок ждёт | Default 10s, конфигурируется. Метрика histogram покажет распределение |
| Refund не доходит до billing → оксары «застряли» | Логирование + manual reconciliation tool через admin-frontend (отдельная задача — не в плане 77) |

---

## Что после плана 77

- Game-nova может **списывать оксары** через billing-сервис.
- **План 65 Ф.6 KindTeleportPlanet** разблокирован — можно
  реализовывать.
- **План 66 Ф.5 платный выкуп** разблокирован.
- Идемпотентность работает на любых мутирующих endpoint'ах
  game-nova через единый middleware.
- Будущие премиум-фичи (AI-советник, ускорения, защита) — могут
  переиспользовать client + middleware.

---

## References

- ADR-0009 (currency rebranding) — оксары vs оксариты.
- План 38 (billing-service) — закрыт, источник API.
- `projects/portal/backend/internal/portalsvc/credits.go` —
  образец billing-клиента в portal.
- `projects/billing/backend/internal/billing/wallet.go` —
  серверная сторона `/wallet/spend`.
