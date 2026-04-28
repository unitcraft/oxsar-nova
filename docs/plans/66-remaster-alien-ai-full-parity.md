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
- **Ф.3. Реализация Kind'ов FlyUnknown, GrabCredit, ChangeMissionAI.** — ✅ закрыто 2026-04-28
- Ф.4. Расширение HoldingAI до 8 действий (2 активных + 6 заглушек).
  Spawner-проводка `internal/alien/Spawn` → `origin/alien.GenerateMission` (использует pgx-Loader).
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

### Ф.3 — итог (2026-04-28)

Эталон применённого паттерна — `KindDemolishConstruction` из плана 65 Ф.1
(commit 9a3992a384): typed payload, slog audit, R8 метрики автоматом
на уровне worker, idempotency через сравнение состояний / state-machine.

Реализовано:

- `payload.go` — `MissionPayload` + `ChangeMissionPayload` (R13 typed).
- `service.go` — `Service` с зависимостями (catalog, bundle, loader,
  cfg). Конструктор `NewService(cat, loader)`, `WithBundle`,
  `WithConfig`. Stateless между событиями.
- `handlers.go` — три `event.Handler`-метода:
  - `FlyUnknownHandler()` — порт onFlyUnknownEvent (PHP:652-826):
    грабёж → подарок ресурсов → подарок оксаритов → атака → halt.
  - `GrabCreditHandler()` — тонкий wrapper, форсирует
    `mode=GrabCredit` и делегирует в FlyUnknown (PHP:647-650).
  - `ChangeMissionAIHandler()` — порт onChangeMissionAIEvent
    (PHP:864-921): при remaining ≥ 8h обновляет parent payload с
    новым `power_scale = 1 + control_times*1.5` и random mode
    (Attack/FlyUnknown); при remaining < 8h продлевает parent
    fire_at на rand(10..50)s. control_times++ всегда.
  - Helpers: `spawnAttackFromMission`, `spawnHaltFromMission`,
    `sendMessage`. Halt-payload совместим с
    `internal/alien.Service.HaltHandler` через identical JSON-shape.
- `loader_pgx.go` — pgx-реализация `Loader`:
  - `LoadAttackCandidates` (порт findTarget PHP:336-369);
  - `LoadCreditCandidates` (порт findCreditTarget PHP:299-334);
  - `LoadPlanetShips` (PHP:266-276);
  - `LoadUserResearches` (PHP:278-297);
  - `LoadActiveAlienMissionsCount` (PHP:191).
  - R10: фильтр по `users.universe_id` во всех queries.
- `cmd/worker/main.go` — регистрация трёх handlers рядом с
  существующими `internal/alien.Service` handlers. loader=nil для
  handlers Ф.3 (они работают через event.UserID/PlanetID без поиска
  целей); loader потребуется в Ф.4 для Spawner / generateMission
  replan-mode.

Тесты:

- `payload_test.go` — round-trip JSON для `MissionPayload`
  и `ChangeMissionPayload` (защита от случайного rename JSON-тэга).
- `handlers_property_test.go` — property-based (rapid, R4):
  - `CalcGrabAmount` детерминизм + bounds (0.0008..0.001 от credit);
  - `CalcGiftAmount` детерминизм + cap MaxGiftCredit*1.02;
  - `HoldingExtension` монотонность + cap.
- `handlers_integration_test.go` — golden-тесты с TEST_DATABASE_URL
  (auto-skip без БД, как demolish_test.go):
  - `TestFlyUnknown_GrabBranch` — credit=1M, mode=GrabCredit →
    оксариты списаны (~0.08-0.10%), сообщение в инбоксе;
  - `TestFlyUnknown_HaltOrAttackBranch` — бедный игрок → ровно 1
    follow-up event (Attack или Halt);
  - `TestChangeMissionAI_ExtendsParent` — remaining=1h → fire_at
    parent +10..50s, control_times++;
  - `TestChangeMissionAI_ReplansWithPowerScale` — remaining=20h →
    `power_scale = 1 + 2*1.5 = 4.0`, mode random Attack/FlyUnknown;
  - `TestChangeMissionAI_SkipParentGone` — отсутствующий parent →
    handler возвращает nil;
  - `TestFlyUnknown_NegativePayloadRejected` — пустой
    user_id/planet_id в payload → ошибка валидации (без БД).

Что **не делалось** в Ф.3 (отложено в Ф.4):
- Spawner-проводка `internal/alien.Service.Spawn` →
  `origin/alien.GenerateMission` (не трогаю существующий spawner,
  чтобы не сломать 4 существующих handler'а; будет переделано в
  Ф.4 одновременно с переписыванием HoldingAI на typed payloads).
- ChangeMissionAI replan-mode не вызывает новый `GenerateFleet`
  для подмены alien-флота parent'а — только обновляет mode/
  control_times/power_scale. Полный replan с новым флотом —
  Ф.4 (требует loader для loadPlanetShips/loadUserResearches).
  Зафиксировано в `simplifications.md` как сознательное отложение.

Объём Ф.3: ~700 строк Go (production: handlers+payload+service+loader_pgx)
+ ~600 строк тестов.

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
