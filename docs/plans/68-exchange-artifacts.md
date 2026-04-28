# План 68: Биржа артефактов (Exchange)

**Дата**: 2026-04-28
**Статус**: Скелет (детали допишет агент-реализатор при старте)
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

## Этапы (детали — при старте)

- Ф.1. Миграции БД (exchange_lots, exchange_history).
- Ф.2. OpenAPI-схемы (R2).
- Ф.3. Backend (service + handler).
- Ф.4. Event-handler'ы для expire/ban.
- Ф.5. Frontend (3 экрана).
- Ф.6. Антифрод (price cap, премиум-permit).
- Ф.7. Финализация.

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
