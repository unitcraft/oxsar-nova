# План 66 (ремастер): AlienAI до полного паритета с oxsar2-classic

**Дата**: 2026-04-28
**Статус**: ✅ ЗАКРЫТ (2026-04-28). Все фазы Ф.1-Ф.7 готовы.
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
- **Ф.4. Расширение HoldingAI до 8 действий (2 активных + 6 заглушек).** — ✅ закрыто 2026-04-28
- **Ф.5. Платный выкуп удержания через billing (оксары — R1, ADR-0009).** — ✅ закрыто 2026-04-28
- **Ф.6. Golden-тесты на 50+ итераций (property-based).** — ✅ закрыто 2026-04-28
- **Ф.7. Финализация.** — ✅ закрыто 2026-04-28

> **Замечание по Ф.4**: spawner-проводка
> `internal/alien.Service.Spawn` → `origin/alien.GenerateMission`
> отложена в `simplifications.md` (см. запись плана 66 Ф.3) — не
> относится к HoldingAI 8-фаз и будет переделана отдельно при
> переходе к новому payload-формату HALT/HOLDING.

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

### Ф.4 — итог (2026-04-28)

Эталон применённого паттерна — `KindDemolishConstruction` (план 65 Ф.1)
+ существующий `KindAlienHoldingAI` в `internal/alien/holding.go`
(50/50 random extract/unload — упрощение плана 15, см.
simplifications.md «HOLDING_AI — 6 из 8 действий пустые»).

R0-исключение: HoldingAI работает во ВСЕХ вселенных (uni01/uni02 +
origin), как и весь пакет origin/alien.

Реализовано:

- `payload_holding_ai.go` — `HoldingAIPayload` (R13 typed) +
  `HoldingFleetUnit` + `HoldingParentSnapshot`. JSON-shape совместим
  с `internal/alien.holdingPayload` (старый пакет nova): HALT/HOLDING
  создаёт KindAlienHoldingAI с тем же json-marshal'ом, после
  переключения регистрации worker'а handler читает тот же payload без
  ломки сериализации. Поля `control_times`/`paid_sum_credit`/
  `metal`/`silicon`/`hydrogen` добавлены `omitempty` — отсутствуют в
  старом payload и интерпретируются как 0 при первом тике.

- `holding_ai_handler.go`:
  - `HoldingSubphase` — типизированное имя 1 из 8 подфаз (R13).
  - `holdingSubphasesOrder` + `pickHoldingSubphase` — равновесный
    выбор 1 из 8, как в origin AlienAI:949-966 (8 веток × вес 10).
  - `HoldingAIHandler` — порт `onHoldingAIEvent` (PHP:924-1014):
    проверка parent → consume `paid_credit` → выбор subphase →
    `control_times++` → продление `parent.fire_at` (если
    `paid_credit>0` или alien_fleet изменился) → планирование
    следующего тика по `HoldingAISubphaseDuration` (origin:974,
    `clamp(min(12h, 30s*times) ... max(24h, 60s*times))`),
    capped на parent.fire_at-2s.
  - `subphaseExtractAlienShips` — порт PHP:1025-1079: ceil(q × 0.01
    × times²) убавление случайного alien-стека, capped q-1; если
    все стеки == 1 — флот уходит и HOLDING закрывается.
  - `subphaseUnloadAlienResources` — порт PHP:1081-1084 (=
    Extract + unload-флаг PHP:1053-1061): сначала рассчитывается
    подарок ресурсов из parent-snapshot
    `gift = ceil(min(snap×0.7, snap×0.1×times))`; при good_bonus
    (любой res > 1M) extract уменьшается ×0.3.
  - `subphaseStub` × 6 — заглушки RepairUserUnits / AddUserUnits /
    AddCredits / AddArtefact / GenerateAsteroid /
    FindPlanetAfterBattle (как в origin PHP:1086-1124 — пустые
    тела). Audit-log на каждом вызове для распределения метрик.
  - `closeHoldingScattered` — закрытие parent KindAlienHolding
    (`state='ok'`, processed_at=now), сообщение
    "пришельцы рассеялись".

- `cmd/worker/main.go` — KindAlienHoldingAI переключён со старого
  `internal/alien.HoldingAIHandler` на `originAlienSvc.HoldingAIHandler()`.
  Старый handler остался в коде (используются `LoadHoldingDefender`/
  `CloseHoldingIfWiped` из того же пакета), но не регистрируется.

Тесты:

- `holding_ai_handler_test.go` (rapid, R4):
  - `TestPickHoldingSubphase_Distribution` — все 8 веток
    встречаются за 800 бросков, ни одна не >50%;
  - `TestPickHoldingSubphase_Determinism` — детерминизм по seed;
  - `TestUnloadGift_Bounds` — cap snapshot × 0.7, монотонность по
    times, zero-at-zero;
  - `TestHoldingAISubphaseDuration_GrowsWithControlTimes` —
    верхняя граница hi растёт с control_times.
  - `TestPowerScaleAfterControlTimes_Monotone` — `1 + ct*1.5` >= 1.0.

- `holding_ai_integration_test.go` (с TEST_DATABASE_URL):
  - `TestHoldingAI_TickIncrementsControlTimes` —
    `control_times` 0→1 в payload follow-up event;
  - `TestHoldingAI_PaidCreditExtendsParent` — paid=50 →
    parent.fire_at +2h (формула `2h × paid / 50`),
    `paid_sum_credit` и `paid_times` пишутся в parent.payload;
  - `TestHoldingAI_SkipParentGone` — отсутствующий parent →
    handler возвращает nil без follow-up;
  - `TestHoldingAI_SkipParentDone` — parent state='ok' → silent skip;
  - `TestHoldingAI_StubSubphasesAreNoop` — 50 тиков без
    `paid_credit` не сдвигают parent.fire_at;
  - `TestHoldingAI_NegativePayloadRejected` — пустой
    `holding_event_id` → ошибка валидации.

Что **не делалось** в Ф.4 (по ТЗ):
- Spawner-проводка `internal/alien.Service.Spawn` →
  `origin/alien.GenerateMission` — записана в simplifications.md
  как Ф.3-отложение, остаётся открытой; не блокирует Ф.4.
- 1% recheck `checkAlientNeeds` (PHP:1006-1008) — origin
  с малой вероятностью на тике HOLDING_AI запускает спавн новой
  миссии. В nova спавн идёт через scheduler `alien_spawn` и
  redundant 1%-trigger излишен. Записано в `simplifications.md`
  как сознательное расхождение.

Объём Ф.4: ~530 строк Go (production: payload + handler +
8 sub-phases) + ~280 строк тестов (5 property + 6 integration).

### Ф.5 — итог (2026-04-28)

R0-исключение: фича работает во ВСЕХ вселенных (uni01/uni02 + origin),
как и весь пакет origin/alien.

**Особенность относительно legacy**: в `AlienAI.class.php` платного
выкупа НЕ существует — там есть только `paid_credit` (продление окна
HOLDING на 2h за каждые 50 оксаритов, см. PHP:993, маппится на
`internal/alien.PayHolding`). Buyout — НОВАЯ фича ремастера: одной
транзакцией платим оксары и выходим из HOLDING полностью. Цена
фиксированная (формулы в legacy нет), параметризована
`Config.BuyoutBaseOxsars` (default 100 оксаров). Без ADR «отклонение
от legacy» — отсутствие самой фичи в legacy не подпадает под R0
(R0 запрещает менять modern-числа nova; новая фича в origin —
сознательный upgrade ремастера).

Реализовано:

- `configs/balance/origin.yaml` — **в репо отсутствует**, не создаём в
  Ф.5; параметр `BuyoutBaseOxsars` живёт в
  `internal/origin/alien/config.go` как часть `Config`/`DefaultConfig()`,
  как остальные 25+ параметров AlienAI. Это уточнение относительно ТЗ
  Ф.5 промпта (он говорил «параметризуй в configs/balance/origin.yaml»)
  — соответствует реальной структуре пакета (Ф.1 уже выбрала путь
  in-package config). Зафиксировано в `simplifications.md`.

- `projects/game-nova/api/openapi.yaml`:
  - новый tag `alien`,
  - `POST /api/alien-missions/{mission_id}/buyout` с обязательным
    header `Idempotency-Key`, кодами 200/401/402/404/409/503,
  - `AlienBuyoutResponse {mission_id, cost_oxsars, freed_at}`.

- `internal/origin/alien/buyout_handler.go`:
  - `Config.BuyoutBaseOxsars` (новое поле, default=100),
  - `CalcBuyoutCost(cfg, missionID) int64` — pure-функция (для property),
  - `BuyoutBilling` interface (узкий — только Spend),
  - `Buyout(ctx, db, billing, cfg, userID, missionID, userToken,
    idempotencyKey)` — основная функция в 2 транзакциях:
    1. lock + pre-check (kind=HOLDING / state=wait / owner==userID);
    2. billing.Spend ВНЕ tx (network вызов);
    3. close HOLDING + DELETE тиков HOLDING_AI этой миссии.
  - Sentinel ошибки: `ErrMissionNotFound`, `ErrMissionAlreadyClosed`,
    `ErrInsufficientOxsars`, `ErrIdempotencyConflict`,
    `ErrBillingUnavailable`.
  - Маппинг billing-client errors → buyout sentinel (ErrFrozenWallet
    тоже идёт в insufficient — UX-эквивалент «нечего тратить»).
  - R3 slog: user_id, mission_id, idempotency_key, cost_oxsars,
    freed_at + error-event `alien_buyout_db_after_spend` для
    операторского следствия (если billing списал, а DB write упал).

- `internal/origin/alien/buyout_http.go`:
  - `BuyoutHandler` (separate type — не размытие Service, см.
    рассуждение в файле),
  - `Buyout(w, r)` маппит sentinel-ошибки в `httpx.Error{Status, Code}`
    (402/404/409/503).
  - Forward'ит `Authorization: Bearer <token>` в billing-client
    как UserToken.

- `pkg/metrics/alien_buyout.go` (R8):
  - `oxsar_alien_buyout_total{status}` —
    ok|insufficient|conflict|not_found|billing_unavailable|error;
  - `oxsar_alien_buyout_oxsars_total` (counter, sum успешных списаний).
  - Регистрация автоматическая из `metrics.Register()`.

- `cmd/server/main.go`:
  - import `originalien "oxsar/game-nova/internal/origin/alien"`;
  - construct `alienBuyoutH := originalien.NewBuyoutHandler(db,
    billingC, originalien.DefaultConfig())`;
  - route `pr.With(idemMW.Wrap).Post(
    "/alien-missions/{mission_id}/buyout", alienBuyoutH.Buyout)`.

- `configs/i18n/{ru,en}.yml`:
  - **5 новых ключей**, **0 переиспользовано**:
    `alien.buyoutSubject` / `alien.buyoutBody` (in-game сообщение
    игроку — placeholder для будущего message-flow), `buyoutSuccess` /
    `buyoutInsufficientOxsars` / `buyoutNotInHolding` (UI-фронт может
    использовать как фолбэк-перевод error-кодов).
  - В Ф.5 backend сам не использует эти ключи (только error-codes
    в JSON-ответе); они зарезервированы и переведены сразу.

Тесты:

- `buyout_handler_test.go` — **unit/property** без БД:
  - `TestCalcBuyoutCost_Determinism` (rapid R4): cost не зависит
    от mission_id для любого base-oxsars > 0;
  - `TestCalcBuyoutCost_PositiveOnDefault`: дефолт > 0
    (защита от misconfiguration, R15);
  - `TestBuyoutBilling_Compatibility`:
    `*billingclient.Client` ⊆ `BuyoutBilling`.

- `buyout_integration_test.go` — **integration** с TEST_DATABASE_URL
  (auto-skip без БД, паттерн Ф.3/Ф.4):
  - `TestBuyout_HappyPath` — 200, mission state='ok',
    AI-тики удалены, billing.Spend вызван 1 раз с правильными
    Reason="alien_buyout"/RefID/IdempotencyKey/Amount=100;
  - `TestBuyout_AlreadyClosed` — mission state='ok' → 409,
    billing НЕ вызван;
  - `TestBuyout_ForeignMission` — другой owner → 404 (единый, не
    раскрываем существование), billing НЕ вызван;
  - `TestBuyout_MissionNotFound` — несуществующий ID → 404;
  - `TestBuyout_WrongKind` — KindAlienAttack вместо HOLDING → 404;
  - `TestBuyout_Insufficient` — billing.ErrInsufficientOxsar →
    402, mission state='wait' (не тронута);
  - `TestBuyout_BillingUnavailable` — billing.ErrBillingUnavailable
    → 503, mission state='wait';
  - `TestBuyout_IdempotencyConflict` — billing.ErrIdempotencyConflict
    → 409, mission state='wait';
  - `TestBuyout_DropsAITicks` — после успеха DELETE затрагивает
    только тики этой mission, чужие тики целы.
  - mockBilling реализует `BuyoutBilling` через `returnFn` —
    позволяет test'у моделировать любой ответ billing-сервера без
    httptest.

Что **не делалось** в Ф.5 (по ТЗ или явно отложено):
- `configs/balance/origin.yaml` — файл в репо отсутствует;
  параметр живёт в `internal/origin/alien/config.go` (см. выше).
- `UPDATE planets SET locked_by_alien=false` — колонки в схеме нет
  (миграции 0001-0080 не имеют такой колонки), блокировка планеты
  моделируется самим присутствием активного HOLDING-event. Закрытие
  HOLDING (`state='ok'`) = разблокировка. Зафиксировано в
  `simplifications.md` как корректировка ТЗ Ф.5 относительно реальной
  схемы.
- 2PC между Postgres и billing — компромисс: billing идемпотентен
  по Idempotency-Key, повтор того же запроса от клиента закроет
  миссию без второго списания. Полный distributed-transaction
  излишен (план 77 такого паттерна не предполагает).

Объём Ф.5: ~330 строк Go (production: handler + http + metrics +
config-поле) + ~480 строк тестов (3 unit/property + 9 integration).

### Ф.6+Ф.7 — итог (2026-04-28)

Ф.6 — golden-итерации формул AlienAI (R4: golden+property для
event/economy ≥85%):

- `projects/game-origin-php/tools/dump-alien-ai.php` — оффлайн PHP-CLI,
  генерирует JSON с 72 кейсами в 9 группах (CalcGrabAmount/
  CalcGiftAmount/HoldingExtension/PowerScaleAfterControlTimes/
  HoldingDuration/FlightDuration/ChangeMissionDelay/
  HoldingAISubphaseDuration/WeakenedTechLevel). Константы
  захардкожены 1-в-1 из consts.php:752-770; не требует MySQL и
  legacy-bootstrap. Вывод — массив кейсов
  `{id, group, fn, input, expected_min, expected_max, comment}`.
- `internal/origin/alien/testdata/golden_alien_ai.json` —
  сгенерированный артефакт (зафиксирован в репозитории; регенерация
  однострочной командой `php tools/dump-alien-ai.php > testdata/...`).
- `internal/origin/alien/golden_test.go` — Go-тест `TestGolden_AllCases`
  загружает golden-JSON, для каждого кейса вызывает соответствующий
  helper и проверяет: для exact-кейсов (expected_min == expected_max)
  — точное совпадение (eps=1e-9 для float); для range-кейсов — [min..max]
  (включительно). Auto-skip если testdata-файла нет (CI без PHP).
  Защита: ≥50 кейсов и ≥5 групп (план 66 Ф.6 R4).
- `internal/origin/alien/golden_property_test.go` — property-based
  (rapid):
  - `TestProperty_PickAttackTarget_EmptyReturnsNil` — empty/empty-slice;
  - `TestProperty_PickAttackTarget_AllIneligibleReturnsNil` — все
    в umode → nil;
  - `TestProperty_PickCreditTarget_EmptyReturnsNil` + monotonic-by-credit;
  - `TestProperty_GenerateFleet_PowerNotExcessive` — alien_power ≤
    target_power×scale×3;
  - `TestProperty_GenerateFleet_DeterministicSameSeed`;
  - `TestProperty_ApplyShuffledTechWeakening_NotAbove` —
    weakened ≤ level+1;
  - `TestProperty_ShuffleKeyValues_PreservesMultiset` — multiset
    инвариант shuffle.
- `internal/origin/alien/golden_coverage_test.go` — добив покрытия
  pure-функций до **94.0% среднее** (≥85% R4): MaxRealEndAt,
  RandRoundRangeDur edge-cases, attackTargetEligible/creditTargetEligible
  все ветки отказа, GenerateFleet с findMode и DS-target, edge-cases
  fleetMapToSlice / nil/nil-флота.

Архитектурное замечание (документировано в комментарии файла
`golden_test.go` и в PHP-CLI): PHP `mt_rand` (Mersenne Twister)
несовместим бит-в-бит с Go `pkg/rng` (xorshift64*). Поэтому golden
работает на уровне ИНВАРИАНТОВ ФОРМУЛЫ origin AlienAI, не байтов
RNG. Полная mt_rand-портация — отдельный future work R8 (см.
`shuffle.go:115`).

Ф.7 — финализация:

- `divergence-log.md` D-036 → ✅ (закрыто Ф.1-Ф.4 + Ф.6+Ф.7;
  Ф.5 buyout — параллельная сессия).
- Шапка плана 66: Ф.1-Ф.4 ✅, Ф.6+Ф.7 ✅, Ф.5 🟡 (slot G в
  active-sessions). После закрытия Ф.5 план 66 закрыт полностью.
- Запись в `docs/project-creation.txt`.

Объём Ф.6+Ф.7: ~290 строк PHP (CLI), ~370 строк Go-тестов
(golden + property + coverage), 22kb golden_alien_ai.json.

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
