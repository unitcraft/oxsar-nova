# Промпт: выполнить план 68 (биржа артефактов)

**Дата создания**: 2026-04-28
**План**: [docs/plans/68-remaster-exchange-artifacts.md](../plans/68-remaster-exchange-artifacts.md)
**Зависимости**: нет критичных. Параллельно с 64.
**Объём**: 3-4 нед, ~2000 строк Go + ~600-800 строк frontend
(origin-фронт, для nova-фронта — план 76).

---

```
Задача: выполнить план 68 (ремастер) — биржа артефактов
(player-to-player) как общий знаменатель для всех вселенных.

ПЕРЕД НАЧАЛОМ:

1) git status --short.

2) Прочитай:
   - docs/plans/68-remaster-exchange-artifacts.md
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5»
     (R0-R15)
   - docs/research/origin-vs-nova/divergence-log.md D-039
   - docs/research/origin-vs-nova/nova-ui-backlog.md U-001
   - docs/adr/0009-currency-rebranding.md (валюта оплаты лотов)

3) Выборочно:
   - projects/game-origin-php/src/game/Exchange.class.php (1220) +
     Stock.class.php (757) + StockNew.class.php (850) — referenc
     (legacy-PHP).
   - projects/game-nova/backend/internal/billing/ — образец паттерна
     idempotency.

ЧТО НУЖНО СДЕЛАТЬ:

Backend (internal/exchange/):
- Миграции (R1, R10):
  · exchange_lots: id, user_id, universe_id, artifact_type,
    quantity, price_oxsarit, created_at, expires_at, status
    (TEXT enum: active/sold/cancelled/expired через CHECK).
  · exchange_history: журнал сделок.
  · per-universe изоляция: universe_id во всех таблицах.

- 5+ endpoints (R6 REST + R2 OpenAPI первым):
  · GET /api/exchange/lots — список с фильтрами (cursor pagination
    R6).
  · POST /api/exchange/lots — создать (Idempotency-Key R9).
  · GET /api/exchange/lots/{id} — детали.
  · POST /api/exchange/lots/{id}/buy — покупка (Idempotency-Key R9).
  · DELETE /api/exchange/lots/{id} — отозвать.
  · GET /api/exchange/stats — статистика (опц.).

- Event-handlers (зависит от плана 65):
  · KindExchangeExpire (snake_case полные слова).
  · KindExchangeBan (служебный).

- Премиум-механика: «Знак торговца» (артефакт-permit) — расширение
  существующего artifact-домена.

- Anti-fraud:
  · Cap на цену (EXCH_SELLER_MAX_PROFIT 1000% из legacy-PHP).
  · Rate-limit (R11): max 10 лотов/час/игрок.
  · Anti-fraud cap fingerprint, IP-pattern (план 59 опытный).

- Валюта оплаты (R1 особый случай / ADR-0009):
  · Лоты продаются за **оксариты** (soft, ст. 1062 ГК — юр-чисто).
  · НЕ оксары (hard, ст. 437 ГК — реальные деньги).
  · Имя поля БД: price_oxsarit (snake_case, не price_amount).

Frontend для origin-фронта (план 72 Spring 5):
- 3 экрана (список / детали / создание) — это часть плана 72,
  здесь пишется код origin-фронта в projects/game-origin/frontend/.
- Для nova-frontend — отдельный план 76 (после плана 68).

ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА:

R0: общий знаменатель — новая функциональность, не правка
существующих механик nova.
R1: snake_case + полные слова + valuta по ADR-0009 (oxsarit).
R6: REST API с нуля (POST для buy/cancel, не GET с эффектами).
R8: Prometheus метрики (создание/покупка/expire/cancel).
R9: Idempotency-Key для всех мутирующих endpoint'ов.
R10: per-universe изоляция (universe_id обязателен).
R11: rate-limiting на public endpoints.
R12: i18n — grep nova-bundle перед созданием новых ключей.
R13: typed payload для Kind'ов.
R15: без упрощений (anti-fraud cap, rate-limit, full tests со старта).

R15 УТОЧНЕНО (обязательно прочитай — roadmap-report.md "Часть I.5 / R15"):

🚫 НЕ КЛАССИФИЦИРУЙ КАК TRADE-OFF В simplifications.md:
- R8 Prometheus метрики (counter+histogram) — для каждого endpoint.
  Особенно критично для биржи: создание/покупка/expire/cancel —
  всё с метриками со старта, financial flow должен быть наблюдаемым.
- R9 Idempotency-Key для buy/create — критично, без неё двойное
  списание оксаритов.
- R12 i18n — все user-facing strings через Tr().
- R10 universe_id в exchange_lots / exchange_history.
- DB CHECK для status enum (active/sold/cancelled/expired).
- FK constraint + index на user_id, artifact_type, universe_id.

✅ TRADE-OFF (можно с обоснованием в simplifications.md):
- Полнотекстовый поиск в Ф.5 если нужно отложить — ок, явный
  «Приоритет M, Ф.X».
- Премиум-permit «Знак торговца» — если интеграция с artifact-доменом
  займёт много времени, можно отложить cap-проверку с явным TODO.

GIT-ИЗОЛЯЦИЯ:
- Свои пути: internal/exchange/, migrations/, openapi.yaml,
  internal/event/handlers.go (для регистрации Kind'ов exchange),
  тесты, docs/plans/68-..., divergence-log.md (D-039),
  nova-ui-backlog.md (U-001).

КОММИТЫ:

3 коммита:
1. feat(exchange): миграции + OpenAPI + service.
2. feat(exchange): event-handlers (Kind*) + premium-permit.
3. feat(exchange): anti-fraud + rate-limit + golden-тесты.

ЧЕГО НЕ ДЕЛАТЬ:
- НЕ принимать оплату в оксарах (hard). Только оксариты.
- НЕ делать аукцион — только fixed-price лоты на старте.
- НЕ делать cross-universe торговлю — лоты per-universe.
- НЕ дублировать i18n-ключи.

УСПЕШНЫЙ ИСХОД:
- 5+ endpoints работают, golden-тесты + integration-тесты зелёные.
- D-039 закрыт, U-001 зафиксирован.
- Anti-fraud cap + rate-limit + Prometheus метрики со старта.
- 2 Kind'а в event-loop'е работают (expire/ban).
- Все existing nova-тесты зелёные (R0).

Стартуй.
```
