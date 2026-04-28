# Промпт: выполнить план 76 (nova-frontend exchange UI)

**Дата создания**: 2026-04-28
**План**: [docs/plans/76-remaster-nova-frontend-exchange-ui.md](../plans/76-remaster-nova-frontend-exchange-ui.md)
**Зависимости**: блокируется планом 68 (биржа в nova-backend).
**Объём**: 1-2 нед, ~600-800 строк frontend.

---

```
Задача: выполнить план 76 (ремастер) — UI биржи артефактов в
nova-frontend (uni01/uni02), чтобы игроки modern-вселенных
получили ту же фичу, что и origin.

ВАЖНОЕ:
- Зависит от плана 68 — backend биржи должен быть готов.
- Биржа — общий знаменатель (план 68 backend), для origin-фронта
  3 экрана уже описаны в плане 72. Здесь — те же экраны для
  nova-frontend в современном дизайне.

ПЕРЕД НАЧАЛОМ:

1) git status --short.

2) Прочитай:
   - docs/plans/76-remaster-nova-frontend-exchange-ui.md
   - docs/plans/68-remaster-exchange-artifacts.md (backend, OpenAPI).
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5»
     (R0-R15).
   - docs/adr/0009-currency-rebranding.md (валюта оплаты).

3) Выборочно:
   - projects/game-nova/api/openapi.yaml — endpoints биржи (после
     плана 68).
   - projects/game-nova/frontend/src/features/ — стиль компонентов.

ЧТО НУЖНО СДЕЛАТЬ:

3 экрана в projects/game-nova/frontend/src/features/exchange/:

1. Список лотов (ExchangeListPage):
   - TanStack Query infinite scroll (cursor-pagination, R6).
   - Фильтры: тип артефакта, диапазон цен (oxsarit), владелец.
   - Cards/Table с фильтрами и сортировкой.

2. Детали лота (ExchangeLotPage):
   - Полная информация о лоте.
   - Кнопка «Купить» — POST с Idempotency-Key (R9).
   - Списание оксаритов через billing-API (нет, через game-nova
     /api/exchange/lots/{id}/buy — он внутренне списывает).

3. Создание лота (ExchangeCreatePage):
   - Форма создания.
   - Premium-проверка (Знак торговца — артефакт-permit).
   - Cap на цену (EXCH_SELLER_MAX_PROFIT 1000%).

ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА:

R0: общий знаменатель (биржа в nova-frontend) — не трогает
существующие nova-механики.
R1: API JSON-поля snake_case (price_oxsarit, expires_at,
artifact_type).
R6: REST по nova-конвенциям (используется openapi.yaml плана 68).
R9: Idempotency-Key на покупке/создании.
R12: i18n — обязательный grep nova-bundle перед созданием новых
ключей. exchange.* namespace, переиспользовать общие
error.common.* и т.д.
R15: без упрощений — все edge-кейсы (пустой список, ошибка
покупки, недостаточно оксаритов, истёкший лот, etc).

R15 УТОЧНЕНО (см. roadmap-report.md "Часть I.5 / R15"):
🚫 Пропуск Idempotency-Key на покупке — нет, обязательно.
🚫 Хардкод строки вместо Tr() — нет, обязательно через i18n.
✅ Trade-off: cursor-pagination отложить если бэкенд (план 68) ещё
   не поддерживает — ок, явный TODO с Приоритет M.

GIT-ИЗОЛЯЦИЯ:
- Свои пути: frontend/src/features/exchange/, frontend/src/api/
  (если генерация OpenAPI требует обновления), nova-bundle i18n,
  docs/plans/76-..., тесты.

КОММИТЫ:

1-2 коммита:
1. feat(frontend): exchange UI 3 экрана (план 76).
2. (опц.) test(frontend): exchange E2E тесты + i18n финализация.

ЧЕГО НЕ ДЕЛАТЬ:
- НЕ дублировать UI origin-фронта — у origin pixel-perfect клон
  (план 72), у nova современный дизайн.
- НЕ принимать оплату в оксарах — только оксариты (R1, ADR-0009).
- НЕ делать cross-universe торговлю.

УСПЕШНЫЙ ИСХОД:
- 3 экрана в nova-frontend работают.
- U-001 (биржа в UI) и часть X-017/X-020 закрыты для nova-стороны.
- В коммитах i18n-метрика «переиспользовано/новых».

Стартуй.
```
