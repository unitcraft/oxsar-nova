# План 66 (ремастер): AlienAI до полного паритета с oxsar2-classic

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

**Применимо ко всем вселенным** (uni01/uni02 + origin) — решение
пользователя 2026-04-28 (явное исключение R0, см. roadmap-report
«Часть I.5» / R0). План 15 этап 3 был пропущен в nova; теперь
полная AI применяется ко всем modern-вселенным одинаково. Это
сознательный upgrade игрового опыта modern, не нарушение R0.

---

## Что делаем (по A-NNN из alien-ai-comparison.md)

- **`KindAlienFlyUnknown`** handler — грабёж / подарок / атака как
  альтернативы.
- **`KindAlienGrabCredit`** — отдельный сценарий кражи **оксаритов**
  (название Kind осталось `GrabCredit` исторически, фактическая
  валюта — оксариты по ADR-0009)
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

- Не вводим feature-флаги по вселенным — AlienAI работает
  одинаково для всех (uni01/uni02/origin). Это явное исключение R0
  по решению пользователя.
- Не реализуем 6 заглушек HOLDING_AI как полные действия — это
  no-op в самом origin.

## Этапы (детали — при старте)

- **Ф.1. Расширение state machine + переходы.** — ✅ закрыто 2026-04-28
- **Ф.2. generateFleet + findTarget + shuffleKeyValues (helper-логика).** — ✅ закрыто 2026-04-28
- Ф.3. Реализация Kind'ов FlyUnknown, GrabCredit, ChangeMissionAI.
  Зависит от плана 65 (typed payload R13, Idempotency R9, метрики R8).
- Ф.4. Расширение HoldingAI до 8 действий (2 активных + 6 заглушек).
- Ф.5. Платный выкуп удержания через billing (оксары — R1, ADR-0009).
- Ф.6. Golden-тесты на 50+ итераций (property-based).
- Ф.7. Финализация.

### Ф.1+Ф.2 — итог (2026-04-28)

Создан пакет `projects/game-nova/backend/internal/origin/alien/`
(R0-исключение: пакет применяется во ВСЕХ вселенных, не только origin):

- `doc.go` — комментарий о R0-исключении и составе пакета.
- `config.go` — `Config` + `DefaultConfig()` (25+ параметров,
  1-в-1 с `consts.php:752-770`). Защита R15: значения семантически
  идентичны origin.
- `state.go` — типизированные структуры `Mission`, `Fleet`,
  `FleetUnit`, `HoldingState`, `PlanetSnapshot`, `TechProfile`,
  `MissionMode` (R13).
- `fleet_generator.go` — `GenerateFleet(target, available, scale,
  cfg, r, opts...) Fleet` — порт PHP:405-622. Поддерживает
  спец-юниты (Death Star, Transplantator, Armored Terran,
  Espionage Sensor, Alien Screen).
- `target.go` — `PickAttackTarget`, `PickCreditTarget` (порт
  PHP:299-370). Pure-функции; loader отделён.
- `shuffle.go` — `ShuffleKeyValues`, `ShuffleAllAlienTechGroups`,
  `ApplyShuffledTechWeakening` (PHP:251-264, 138).
- `helpers.go` — `IsAttackTime`, `RandRoundRange*`, `FlightDuration`,
  `HoldingDuration`, `ChangeMissionDelay`, `HoldingExtension`,
  `HoldingAISubphaseDuration`, `PowerScale*`, `CalcGrabAmount`,
  `CalcGiftAmount`. Все pure-функции с детерминированным `*rng.R`.
- `repo.go` — интерфейс `Loader` (4 метода: LoadAttackCandidates,
  LoadCreditCandidates, LoadPlanetShips, LoadUserResearches,
  LoadActiveAlienMissionsCount). Pgx-реализация — Ф.3.

Тесты: `config_test.go`, `helpers_test.go`, `shuffle_test.go`,
`target_test.go`, `fleet_generator_test.go` — все зелёные.

Что **не делается** в Ф.1+Ф.2:
- Kind handlers — Ф.3.
- Spawner-проводка `internal/alien/Spawn` под `origin/alien` — Ф.3.
- pgx-реализация Loader — Ф.3.
- Prometheus-метрики (R8) — Ф.3 (после плана 65).
- Idempotency-Key (R9) — Ф.3.
- Audit-log — Ф.3.
- 50+ golden-тестов — Ф.6.

Объём Ф.1+Ф.2: ~700 строк Go (production) + ~500 строк тестов.

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
