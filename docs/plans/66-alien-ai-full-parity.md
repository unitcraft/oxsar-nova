# План 66: AlienAI до полного паритета с legacy

**Дата**: 2026-04-28
**Статус**: Скелет (детали допишет агент-реализатор при старте)
**Зависимости**: блокируется планом 64 (alien-юниты в `configs/balance/origin.yaml`).
**Связанные документы**:
- [15-alien-holding-thursday.md](15-alien-holding-thursday.md) —
  предыдущий этап AlienAI в nova (Этапы 1-2 закрыты, Этап 3
  пропущен — закрывается этим планом)
- [docs/research/origin-vs-nova/alien-ai-comparison.md](../research/origin-vs-nova/alien-ai-comparison.md) —
  state machine + переходы + параметры (A1-A14 расхождения)
- [docs/research/origin-vs-nova/divergence-log.md](../research/origin-vs-nova/divergence-log.md) D-036
- [docs/research/origin-vs-nova/roadmap-report.md](../research/origin-vs-nova/roadmap-report.md) —
  R1-R5 + раздел плана 66

---

## Цель

Достроить AlienAI в game-nova до полного паритета с origin
(вселенная origin): реализовать оставшиеся EVENT_ALIEN_*, перенести
полный AI-движок (`AlienAI.class.php` 1127 строк → ~800 строк Go).

Применимо для вселенной **origin** под флагом, в modern-вселенных
(uni01/uni02) AlienAI остаётся в текущем урезанном виде плана 15.

---

## Что делаем (по A-NNN из alien-ai-comparison.md)

- **`KindAlienFlyUnknown`** handler — грабёж / подарок / атака как
  альтернативы.
- **`KindAlienGrabCredit`** — отдельный сценарий кражи кредитов
  (теперь — оксаритов по ADR-0009; см. R1 «Особый случай: валюта»).
- **`KindAlienChangeMissionAI`** — control_times, power_scale.
- Расширение **`KindAlienHoldingAI`** до 8 действий (с заглушками
  для 6 неактивных, как в origin).
- Алгоритм **`generateFleet()`** — target_power, итеративное
  добавление кораблей.
- 5 алиен-кораблей `alien_unit_1..5` в `configs/balance/origin.yaml`
  (план 64 уже добавил).
- Множитель «четверг» ×5 / ×1.5..2.0 — вынести в
  `configs/balance/origin.yaml` как параметр.
- **`findTarget`** / **`findCreditTarget`** с критериями выбора цели.
- **`shuffleKeyValues`** — случайное ослабление техник.
- Платный выкуп удержания (через billing-API в оксарах? или
  оксаритах? — см. R1).

---

## Что НЕ делаем

- Не дублируем в modern-вселенных. Все новые механики применяются
  только к вселенной **origin** (по `universes.code = 'origin'`)
  ИЛИ через отдельное поле в override-файле
  `configs/balance/origin.yaml` (например, `alien_ai.full_parity: true`).
- Не реализуем 6 заглушек HOLDING_AI как полные действия — это
  no-op в самом origin.

## Этапы (детали — при старте)

- Ф.1. Расширение state machine + переходы.
- Ф.2. generateFleet + findTarget + shuffleKeyValues (helper-логика).
- Ф.3. Реализация Kind'ов FlyUnknown, GrabCredit, ChangeMissionAI.
- Ф.4. Расширение HoldingAI до 8 действий (2 активных + 6 заглушек).
- Ф.5. Платный выкуп удержания через billing.
- Ф.6. Golden-тесты на 50+ итераций (property-based).
- Ф.7. Финализация.

## Конвенции (R1-R5)

- Алиен-юниты в `configs/balance/origin.yaml`: `alien_unit_1`..`_5`
  (snake_case), не `UNIT_A_*`.
- Поля в БД для alien-state: `holds_until_at` (TIMESTAMPTZ, по R1).
- Валюта при грабеже / выкупе — по ADR-0009: оксариты для
  игровых эффектов, оксары для реальных платежей.

## Объём

3 недели. ~800-1000 строк Go (новый internal/legacy/alien/) +
golden-тесты на 50+ итераций.

## References

- alien-ai-comparison.md A1-A14 — state machine, формулы.
- План 15 — что уже сделано в nova (Этапы 1-2).
- `projects/game-origin-php/src/game/AlienAI.class.php` — referenc.
