# План 68 (ремастер): Биржа артефактов (Exchange)

**Дата**: 2026-04-28
**Статус**: ✅ Backend готов (Ф.1-Ф.4 + Ф.6 + Ф.7, 2026-04-28).
UI Ф.5 отложена в план 76 (для nova-фронта) и спринт 5 плана 72
(для origin-фронта).
**Зависимости**: нет критичных (можно параллелить).
**Связанные документы**:
- [62-origin-on-nova-feasibility.md](62-origin-on-nova-feasibility.md)
- [docs/research/origin-vs-nova/divergence-log.md](../research/origin-vs-nova/divergence-log.md) D-039
- [docs/research/origin-vs-nova/nova-ui-backlog.md](../research/origin-vs-nova/nova-ui-backlog.md) —
  U-001
- [docs/research/origin-vs-nova/roadmap-report.md](../research/origin-vs-nova/roadmap-report.md) —
  R1-R5 + раздел плана 68
- [ADR-0009](../adr/0009-currency-rebranding.md) — оплата лотов (R1
  «Особый случай: валюта»)

---

## Цель

Реализовать новую cross-universe фичу — player-to-player биржу
артефактов, **общий знаменатель** для всех вселенных
(uni01/uni02/origin). В origin есть в виде 3 контроллеров
(Exchange + Stock + ExchangeOpts, ~2800 строк PHP), в nova нет.

---

## Что делаем

### Backend

- Модуль `internal/exchange/` (~2000 строк Go).
- Endpoint'ы:
  - `GET /api/exchange/lots` — список с фильтрами (тип артефакта,
    диапазон цен, владелец).
  - `POST /api/exchange/lots` — создать лот.
  - `GET /api/exchange/lots/{id}` — детали.
  - `POST /api/exchange/lots/{id}/buy` — купить.
  - `DELETE /api/exchange/lots/{id}` — отозвать.
  - `GET /api/exchange/stats` — статистика (опц.).
- БД-схема (по R1):
  - `exchange_lots` — id, user_id, artifact_type, quantity,
    price_oxsarit (см. R1 валюта), created_at, expires_at,
    status (enum: active, sold, cancelled, expired).
  - `exchange_history` — журнал сделок.
- Event-loop (план 65 совместимо):
  - `KindExchangeExpire` (snake_case, не `KindExchExpire` —
    полные слова, R1).
  - `KindExchangeBan` — служебный (как EVENT_EXCH_BAN в origin).
- Премиум-механика: «Знак торговца» (артефакт-permit на торговлю,
  как в origin) — расширение существующего artifact-домена.
- Антифрод: cap на цену (EXCH_SELLER_MAX_PROFIT 1000% из origin).

### Frontend (game-nova)

- 3 экрана в `frontend/src/features/exchange/`:
  - Список лотов (с фильтрами).
  - Детали лота + кнопка «Купить».
  - Создание лота.
- Применимо для всех вселенных (origin-фронт получит автоматически
  через nova-API в плане 72).

---

## Что НЕ делаем

- Не реализуем валютные обменники (ресурсы ↔ оксариты) — отдельная
  задача, не входит в биржу артефактов.
- Не делаем аукцион (только fixed-price лоты на старте).
- Не вводим cross-universe торговлю в первой итерации (лоты
  per-universe, как и игроки).

## Этапы (статус)

- ✅ Ф.1. Миграции БД (exchange_lots, exchange_lot_items, exchange_history) — 0081_exchange.sql.
- ✅ Ф.2. OpenAPI-схемы (R2) — секция `/api/exchange/*` + ExchangeLot.
- ✅ Ф.3. Backend (service + handler + repo + errors + payload + metrics).
- ✅ Ф.4. Event-handler'ы expire/ban + wiring в worker.
- ⏸ Ф.5. Frontend (3 экрана) → план 76 (nova) + спринт 5 плана 72 (origin).
- ✅ Ф.6. Антифрод (price cap по rolling-30d AVG, max active lots, max quantity).
- ✅ Ф.7. Финализация: тесты, property-based (escrow + oxsarit invariants), i18n.

## Архитектурные адаптации (при реализации)

Принятые при реализации Ф.1-Ф.7 решения, отличающиеся от исходного
текста плана и зафиксированные в simplifications.md:

- **Currency**: используется существующая колонка `users.credit bigint`
  как backing-storage для оксаритов (по ADR-0009 семантически
  идентично; переименование колонки credit→oxsarit — отдельный план).
- **Schema**: `artifact_unit_id INT` (а не TEXT) + row-per-item модель
  через `exchange_lot_items` (link-таблица), а не quantity-based как
  в legacy oxsar2 — это соответствует существующей схеме `artefacts_user`
  (миграция 0007).
- **Эскроу**: `artefact_state='listed'` (введено миграцией 0013 для
  artmarket) переиспользуется — биржа и artmarket сосуществуют, имея
  одинаковую семантику «escrow в продаже».
- **universe_id не добавляется**: nova однобазная (universe = отдельный
  инстанс БД, см. комментарий в миграции 0075). R10 неприменим в этом
  плане.
- **Permit «Знак торговца»**: DI-интерфейс `PermitChecker` с MVP-stub
  `AlwaysAllowPermit` (gating отключён, активация — отдельный план под
  премиум-вселенные).
- **Балансовый конфиг**: `configs/balance/{default,origin}.yaml` создан
  с секцией `exchange`, но loader пока не реализован — service.go
  использует `DefaultConfig()` с теми же значениями. Подключение
  loader'а — отдельный пост-фикс.

## Конвенции (R1-R5)

- **Валюта оплаты лотов** — оксариты (`oxsarit`, см. R1 «Особый
  случай»). Это игровой ресурс, юр-чисто (ст. 1062 ГК), без 161-ФЗ.
- Колонки `_at` для timestamps.
- Названия Kind'ов — полные слова (`KindExchangeExpire`).
- ENUM для status — TEXT с CHECK по R1 (не магические числа).

## Объём

3-4 недели. ~2000 строк Go + ~600-800 строк frontend.

## References

- D-039 в divergence-log.md.
- U-001 в nova-ui-backlog.md.
- `projects/game-origin-php/src/game/Exchange.class.php` (1220 строк) +
  Stock.class.php (757) + StockNew.class.php (850) — origin-референс.
