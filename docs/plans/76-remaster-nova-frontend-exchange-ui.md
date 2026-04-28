# План 76 (ремастер): nova-frontend UI для биржи артефактов

**Дата**: 2026-04-28
**Статус**: Скелет (детали допишет агент-реализатор при старте)
**Зависимости**: блокируется планом 68 (биржа артефактов в
nova-backend).
**Связанные документы**:
- [68-remaster-exchange-artifacts.md](68-remaster-exchange-artifacts.md) —
  backend биржи + 3 экрана в origin-фронте
- [docs/research/origin-vs-nova/roadmap-report.md](../research/origin-vs-nova/roadmap-report.md) —
  R0-R15 + раздел плана 76

---

## Цель

Добавить UI биржи артефактов в **nova-frontend** (uni01/uni02),
чтобы игроки modern-вселенных получили ту же фичу, что и origin.

Биржа — общий знаменатель (план 68 backend), для origin-фронта
3 экрана уже описаны в плане 72. Здесь — те же экраны для
nova-frontend, в современном дизайне.

---

## Что делаем

3 экрана в `projects/game-nova/frontends/nova/src/features/exchange/`:

- **Список лотов** с фильтрами (тип артефакта, диапазон цен,
  владелец).
- **Детали лота** + кнопка «Купить» (списание оксаритов через
  `Idempotency-Key`, R9).
- **Создание лота** (форма с premium-маркером «Знак торговца»,
  cap на цену).

Все UI-компоненты на nova-конвенциях (TanStack Query, Zustand,
TS strict, OpenAPI-клиент к `/api/exchange/*`).

---

## Что НЕ делаем

- Не дублируем UI origin-фронта (план 72) — у origin свой
  pixel-perfect-клон, у nova — современный дизайн.
- Не вводим новые backend-эндпоинты — используем endpoints плана 68.
- Не делаем cross-universe торговлю — лоты per-universe (как и
  игроки).

## Этапы (детали — при старте)

- Ф.1. UI-каркас (route, layout, navigation entry).
- Ф.2. Список лотов + фильтры (TanStack Query infinite scroll).
- Ф.3. Детали лота + покупка.
- Ф.4. Создание лота + premium-проверка.
- Ф.5. i18n (русский, R12).
- Ф.6. Финализация.

## Конвенции (R0-R15)

- R0: nova-механика биржи **общий знаменатель** — добавляется в
  обе вселенные, не нарушает R0 (это новая фича во всех
  вселенных, не upgrade одной).
- R1: API JSON-поля snake_case (`price_oxsarit`, `expires_at`).
- R6: REST по nova-конвенциям (используется openapi.yaml плана 68).
- R9: Idempotency-Key на покупке/создании.
- R12: i18n с самого начала.

## Объём

1-2 недели. ~600-800 строк frontend.

## References

- План 68 — backend биржи + endpoints.
- План 72 — UI биржи в origin-фронте (для сравнения дизайна).
- U-001 (биржа в UI) и X-017/X-020 (UX-микрологика) — частично
  закроется этим планом для nova-стороны.
