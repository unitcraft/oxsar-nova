# Промпт: выполнить план 68 (биржа артефактов — backend)

**Дата создания**: 2026-04-28 (перезапись после планов 64/65/66/67/77/78)
**План**: [docs/plans/68-remaster-exchange-artifacts.md](../plans/68-remaster-exchange-artifacts.md)
**Зависимости**: ✅ план 64 (override-схема), ✅ план 65 Ф.1-Ф.6
(event-loop эталоны), ✅ план 77 (billing-client), ✅ план 78
(переименование). Не блокирует план 72 — origin-фронт получит
биржу через nova-API в спринте 5 плана 72 (когда биржа будет готова).
Параллелится с планом 72 Ф.2 (разные файлы, разные папки).
**Объём**: ~2000 строк Go + миграция + тесты + i18n,
3-4 коммита по фазам.

---

```
Задача: выполнить план 68 — биржа артефактов player-to-player в
game-nova-backend. ТОЛЬКО backend этой сессии (Ф.1-Ф.4 + Ф.6); UI
(Ф.5) — отдельный план 76 (для nova-фронта) и спринт 5 плана 72
(для origin-фронта).

КОНТЕКСТ:

Биржа артефактов есть в legacy oxsar2 / projects/game-legacy-php/
как 3 контроллера (Exchange + Stock + ExchangeOpts, ~2800 строк PHP).
В game-nova биржи нет. Это cross-universe фича — общий знаменатель
для всех вселенных (uni01/uni02/origin), R0-исключение для origin
не требуется (фича применима ко всем).

Валюта лотов — **оксариты** (soft, ст. 1062 ГК, R1/ADR-0009).
Не оксары. Это P2P-обмен внутриигрового артефакта на soft-currency,
которая добывается. Никаких реальных денег не задействовано.
billing-client (план 77) НЕ используется — он только для оксаров.
Оксариты управляются game-nova-БД напрямую через UPDATE users SET
oxsarit=oxsarit-X в одной транзакции с покупкой/возвратом.

Эталонные паттерны:
- handler/service: internal/alliance/handler.go (план 67),
  internal/origin/alien/buyout_handler.go (план 66 Ф.5).
- Idempotency-middleware: pkg/idempotency/middleware.go (план 77).
- Event-handler с DI: internal/event/teleport_handler.go (план 65 Ф.6).
- Property-based + golden: internal/origin/alien/golden_test.go.

ПЕРЕД НАЧАЛОМ:

1) git status --short. cat docs/active-sessions.md. Если активные
   слоты есть — сверь файлы, при пересечении спроси пользователя.
   План 68 безопасен параллельно с планом 72 Ф.2 (разные папки).

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/68-remaster-exchange-artifacts.md (твоё ТЗ)
   - docs/research/origin-vs-nova/divergence-log.md D-039
   - docs/research/origin-vs-nova/nova-ui-backlog.md U-001
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5» R0-R15
   - docs/adr/0009-currency-rebranding.md (валюта оплаты лотов)
   - projects/game-legacy-php/src/game/Exchange.class.php (~1220 строк)
   - projects/game-legacy-php/src/game/Stock.class.php (~800 строк)
   - projects/game-legacy-php/src/game/ExchangeOpts.class.php (~800 строк)

3) Прочитай выборочно (эталоны кода):
   - projects/game-nova/backend/internal/alliance/handler.go
   - projects/game-nova/backend/internal/origin/alien/buyout_handler.go
   - projects/game-nova/backend/internal/event/teleport_handler.go
   - projects/game-nova/backend/internal/event/kinds.go (свободные
     номера для KindExchangeExpire/Ban)
   - projects/game-nova/backend/pkg/idempotency/middleware.go
   - projects/game-nova/backend/pkg/metrics/billing.go (R8 эталон)
   - projects/game-nova/migrations/0080_*.sql (последний номер
     миграций — твой будет 0081 или старше)

4) Добавь свою строку в docs/active-sessions.md:
   | <N> | План 68 биржа артефактов backend | projects/game-nova/backend/internal/exchange/, migrations/, openapi.yaml, cmd/server, cmd/worker, configs/i18n/ | <дата-время> | feat(exchange): биржа артефактов backend (план 68) |

ЧТО НУЖНО СДЕЛАТЬ:

### Ф.1. Миграция БД

Новая миграция `00NN_exchange.sql`. Номер — следующий после
актуального (`ls projects/game-nova/migrations/`).

Таблицы (R1 snake_case, R10 universe_id):
- `exchange_lots`:
  - `id UUID PRIMARY KEY` (UUIDv7)
  - `seller_user_id UUID NOT NULL`
  - `universe_id INT NOT NULL`
  - `artifact_type TEXT NOT NULL`
  - `quantity INT NOT NULL CHECK (quantity > 0 AND quantity <= 100)`
  - `price_oxsarit BIGINT NOT NULL CHECK (price_oxsarit > 0)`
  - `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
  - `expires_at TIMESTAMPTZ NOT NULL`
  - `status TEXT NOT NULL DEFAULT 'active' CHECK (status IN
    ('active','sold','cancelled','expired'))`
  - `buyer_user_id UUID NULL`
  - `sold_at TIMESTAMPTZ NULL`
  - FK seller_user_id, buyer_user_id → users(id)
  - INDEX (universe_id, status, artifact_type)
  - INDEX (universe_id, status, expires_at)
  - INDEX (seller_user_id, status)
- `exchange_history`:
  - `id UUID PRIMARY KEY`
  - `lot_id UUID NOT NULL`
  - `event_kind TEXT NOT NULL CHECK (event_kind IN
    ('created','bought','cancelled','expired','banned'))`
  - `actor_user_id UUID NULL`
  - `universe_id INT NOT NULL`
  - `payload JSONB NOT NULL` (R13 typed payload)
  - `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
  - INDEX (universe_id, lot_id, created_at)
  - INDEX (universe_id, actor_user_id, created_at)

### Ф.2. OpenAPI (R2 — первым)

В projects/game-nova/api/openapi.yaml — секция /api/exchange/*:

- `GET /api/exchange/lots` с filters: `?artifact_type=&min_price=
  &max_price=&seller_id=&status=&cursor=&limit=50` (cursor-based).
- `POST /api/exchange/lots` — создать. Body: `{artifact_type,
  quantity, price_oxsarit, expires_in_hours}`. Header
  Idempotency-Key обязателен (R9).
- `GET /api/exchange/lots/{id}` — детали.
- `POST /api/exchange/lots/{id}/buy` — купить. Idempotency-Key.
- `DELETE /api/exchange/lots/{id}` — отозвать (только seller).
- `GET /api/exchange/stats` — опц., агрегаты по типам.

Коды: 200/201/400/401/402 (insufficient oxsarits)/403/404/409
(sold/cancelled/idempotency-conflict)/422 (price cap exceeded или
quantity > max)/503.

### Ф.3. Backend (service + handler)

Пакет `internal/exchange/`:
- `service.go` — ListLots/CreateLot/GetLot/BuyLot/CancelLot/Stats.
- `handler.go` — REST-адаптер. Idempotency.Wrap на CreateLot/BuyLot.
- `errors.go` — sentinel:
  ErrLotNotFound, ErrNotASeller, ErrLotNotActive,
  ErrInsufficientOxsarits, ErrPriceCapExceeded, ErrPermitRequired,
  ErrMaxActiveLots, ErrMaxQuantity.
- `repo.go` (interface) + `repo_pgx.go` (pgx-реализация). R10 в
  каждом query.
- `payload.go` — typed structs для exchange_history.payload (R13).

**Ключевая логика:**

- **Создание лота (escrow!)**: в одной tx —
  (а) UPDATE user_artefacts SET quantity=quantity-X WHERE
      user_id=seller AND artifact_type=Y AND universe_id=Z AND
      quantity>=X (если 0 строк → ErrInsufficient — нет такого
      артефакта/мало);
  (б) INSERT exchange_lots ... RETURNING id;
  (в) INSERT exchange_history (event_kind='created', actor=seller, payload);
  (г) INSERT events для KindExchangeExpire с fire_at=expires_at.
  
  Артефакт списывается **при выставлении** (escrow), не при покупке.

- **Покупка**: в одной tx —
  (а) lock+check лота (status='active', expires_at>now, seller≠buyer);
  (б) UPDATE users SET oxsarit=oxsarit-price WHERE id=buyer AND
      oxsarit>=price AND universe_id=Z (если 0 строк → ErrInsufficientOxsarits);
  (в) UPDATE users SET oxsarit=oxsarit+price WHERE id=seller AND universe_id=Z;
  (г) INSERT/UPDATE user_artefacts buyer +X;
  (д) UPDATE exchange_lots SET status='sold', buyer, sold_at;
  (е) INSERT exchange_history (event_kind='bought');
  (ж) UPDATE events SET state='cancelled' (отмена ExpireEvent).
  
  Всё атомарно. R10 — universe_id в каждом WHERE.

- **Отзыв (CancelLot)**: проверка seller, return артефакта,
  UPDATE lot SET status='cancelled', exchange_history,
  отмена ExpireEvent.

R3 slog: trace_id, user_id, universe_id, lot_id, artifact_type,
price_oxsarit.

R8 Prometheus в pkg/metrics/exchange.go:
- counter `oxsar_exchange_lots_total{action,status}` (action=create/
  buy/cancel)
- counter `oxsar_exchange_oxsarits_volume_total` (sum)
- histogram `oxsar_exchange_action_duration_seconds{action}`

### Ф.4. Event-handler'ы expire/ban

В internal/event/exchange_handlers.go:
- **KindExchangeExpire** (snake_case полные слова, R1): fire_at =
  lot.expires_at. При срабатывании — return artefact в
  user_artefacts, UPDATE lot SET status='expired', INSERT
  exchange_history (event_kind='expired'). Idempotent через
  WHERE status='active'.
- **KindExchangeBan**: служебный (бан seller'а от
  модератора → отзыв всех его активных лотов).
  payload {seller_user_id, reason}. SELECT all active lots WHERE
  seller=X → for each: return artefact + UPDATE status='cancelled'
  + history (event_kind='banned'). Используется
  автомодерацией / админ-tool'ами (расширение admin-bff не
  требуется в этой фазе — endpoint POST /api/admin/exchange/ban
  опционально, можно отложить).

Регистрация Kind'ов в `internal/event/kinds.go` (KindExchangeExpire
= следующий свободный номер). Wiring в `cmd/worker/main.go`.

R10: WHERE universe_id во всех queries.
R8: pkg/metrics/exchange.go — counter
oxsar_exchange_event_total{kind,status}.

### Ф.6. Антифрод (configs/balance/)

Параметры в `projects/game-nova/configs/balance/default.yaml` (общий)
+ override в `origin.yaml` если значения отличаются:
```yaml
exchange:
  max_quantity_per_lot: 100
  max_active_lots_per_user: 10
  price_cap_multiplier: 10.0  # 1000% от reference
  reference_window_days: 30
  expires_in_hours_min: 1
  expires_in_hours_max: 168   # 7 дней
```

- **Cap на цену**: при CreateLot — SELECT AVG(price_oxsarit/quantity)
  FROM exchange_history WHERE event_kind='bought' AND artifact_type=Y
  AND universe_id=Z AND created_at > now() - 30 days. Если
  AVG > 0: reference = AVG. Если NULL (нет истории): fallback
  reference из конфига `exchange.fallback_reference_price.<artifact_type>`
  (если есть) или unlimited (если в legacy игрок мог поставить
  любую цену для нового артефакта).
  При price_oxsarit > reference * price_cap_multiplier → 422
  ErrPriceCapExceeded.
- **Лимит активных лотов**: COUNT WHERE seller=X AND status='active'
  AND universe_id=Z. Если >= max_active_lots_per_user → 422
  ErrMaxActiveLots.
- **Quantity cap**: CHECK в миграции (`<= 100`) + проверка в
  handler с понятной ошибкой 422 ErrMaxQuantity.
- **Permit «Знак торговца»**: УТОЧНИ в legacy
  Exchange.class.php — обязателен permit ВСЕМ или только в премиум-
  вселенных. Если всем — реализуй (artefact_type='merchant_permit'
  должен быть у seller'а в активном состоянии). Если только премиум —
  отметь trade-off в simplifications.md (отложить до решения по
  премиум-вселенным; ввести stub-проверку которая всегда возвращает
  true).

### Тесты

- `service_test.go` (mock repo, без БД):
  - happy-path Create/Buy/Cancel.
  - все error-paths: ErrLotNotFound, ErrNotASeller, ErrLotNotActive,
    ErrInsufficientOxsarits, ErrPriceCapExceeded, ErrPermitRequired,
    ErrMaxActiveLots, ErrMaxQuantity.
  - escrow-инвариант через моки.
- `repo_pgx_test.go` (TEST_DATABASE_URL, auto-skip без БД):
  - CRUD всех методов.
  - R10 isolation: лот из universe=1 невидим в universe=2.
  - SELECT FOR UPDATE на lock+check работает (concurrent buy).
- `handler_test.go` (httptest):
  - table-driven по кодам ответов.
  - Idempotency-Key: повторный POST → 200 с тем же id.
- `event_handlers_test.go`:
  - KindExchangeExpire: возврат артефакта.
  - KindExchangeBan: отзыв всех лотов seller'а.
- Property-based (rapid, R4):
  - **Escrow-инвариант**: total quantity артефакта в системе =
    constant (sum user_artefacts + sum exchange_lots.active).
  - **Oxsarit-инвариант**: при покупке сумма oxsarit участников
    constant.
  - Pricing: price_cap detection детерминирован.

Покрытие изменённых строк ≥ 85% (R4).

### i18n (R12)

Grep `projects/game-nova/configs/i18n/{ru,en}.yml` на
`exchange|биржа|лот|trade|sell|buy_artefact`.

Новые ключи (минимум):
- `exchange.lotCreated`
- `exchange.lotBought`
- `exchange.lotCancelled`
- `exchange.lotExpired`
- `exchange.errors.insufficientOxsarits`
- `exchange.errors.priceCapExceeded`
- `exchange.errors.permitRequired`
- `exchange.errors.lotNotActive`
- `exchange.errors.maxActiveLots`
- `exchange.errors.maxQuantity`
- `event.exchange.expire.subject`
- `event.exchange.ban.subject`

В коммите указать соотношение **переиспользовано/новых**.

══════════════════════════════════════════════════════════════════
ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА R0-R15 — см. блок rules-r0-r15.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/rules-r0-r15.md

Особо важно для этого плана:
- R0: НЕ менять modern-баланс. Cap, max-quantity, expiry — параметры
  в configs/balance/default.yaml + override origin.yaml если
  значения отличаются.
- R8 (Prometheus) и R9 (Idempotency-Key) — ПРОПУСК, не trade-off.
- R10 (universe_id во всех queries) — обязательно.
- R12 (i18n) — grep сначала.
- R13 (typed payload) — exchange_history.payload через Go-struct.
══════════════════════════════════════════════════════════════════

══════════════════════════════════════════════════════════════════
GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ — см. блок git-isolation.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/git-isolation.md

Свои пути для CC_AGENT_PATHS:
- projects/game-nova/backend/internal/exchange/
- projects/game-nova/backend/internal/event/exchange_handlers.go
- projects/game-nova/backend/internal/event/exchange_handlers_test.go
- projects/game-nova/backend/internal/event/kinds.go (только новые Kind'ы)
- projects/game-nova/backend/cmd/server/main.go (только новые routes)
- projects/game-nova/backend/cmd/worker/main.go (только Kind-регистрации)
- projects/game-nova/backend/pkg/metrics/exchange.go
- projects/game-nova/backend/pkg/metrics/metrics.go (только RegisterExchange call)
- projects/game-nova/api/openapi.yaml (только секция /api/exchange/*)
- projects/game-nova/migrations/00NN_exchange.sql
- projects/game-nova/configs/i18n/ru.yml (только exchange.* ключи)
- projects/game-nova/configs/i18n/en.yml (только exchange.* ключи)
- projects/game-nova/configs/balance/default.yaml (только exchange:)
- projects/game-nova/configs/balance/origin.yaml (только exchange: override)
- docs/plans/68-remaster-exchange-artifacts.md
- docs/research/origin-vs-nova/divergence-log.md (закрытие D-039)
- docs/research/origin-vs-nova/nova-ui-backlog.md (U-001 backend ✅)
- docs/active-sessions.md
- docs/project-creation.txt (запись итерации)
══════════════════════════════════════════════════════════════════

КОММИТЫ:

3-4 коммита для blame-изоляции:

1) feat(exchange): миграция + OpenAPI (план 68 Ф.1+Ф.2)
2) feat(exchange): backend service + handler + idempotency (Ф.3)
3) feat(exchange): event-handlers expire/ban + антифрод (Ф.4+Ф.6)
4) test(exchange): integration + property-based + финализация (Ф.7)

Trailer во всех: Generated-with: Claude Code
ВСЕГДА с двойным тире:
git commit -m "..." -- $CC_AGENT_PATHS

ЧЕГО НЕ ДЕЛАТЬ:

- НЕ списывать оксары через billing-client. Биржа = оксариты.
- НЕ забывать про escrow (артефакт списывается с игрока ПРИ создании
  лота, не при покупке).
- НЕ забывать R8/R9/R10/R12 (это пропуски по R15, не trade-off).
- НЕ делать UI — это план 76 (nova) и спринт 5 плана 72 (origin).
- НЕ делать аукцион — только fixed-price лоты.
- НЕ делать cross-universe торговлю — лоты per-universe.
- НЕ забывай про -- в git commit.

УСПЕШНЫЙ ИСХОД:

- Миграция exchange_lots + exchange_history применяется чисто.
- 6 endpoint'ов работают, Idempotency-Key корректно.
- KindExchangeExpire / KindExchangeBan зарегистрированы в worker.
- Антифрод: price cap, permit (или trade-off), лимиты.
- Покрытие ≥ 85%.
- D-039 в divergence-log → backend ✅.
- U-001 в nova-ui-backlog → backend ✅, UI в плане 76.
- Шапка плана 68: backend ✅ (Ф.1-Ф.4 + Ф.6 + Ф.7), UI Ф.5
  откладывается на план 76 / спринт 5 плана 72.
- Запись в docs/project-creation.txt — итерация 68.
- Удалена строка из docs/active-sessions.md.

Стартуй.
```
