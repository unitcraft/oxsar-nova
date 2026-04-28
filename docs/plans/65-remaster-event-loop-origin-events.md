# План 65 (ремастер): Расширение event-loop — события origin

**Дата**: 2026-04-28
**Статус**:
- Ф.1 ✅ (KindDemolishConstruction — эталонный handler).
- Ф.2 ✅ (KindDeliveryArtefacts — реальный handler по эталону, 2026-04-28).
- Ф.3 ✅ KindAttackDestroyBuilding (2026-04-28, см. Ф.3 ниже).
- Ф.4 ✅ KindAttackAllianceDestroyBuilding ACS (2026-04-28, см. Ф.4 ниже).
- Ф.5 ✅ KindAllianceAttackAdditional no-op referrer (2026-04-28, см. Ф.5 ниже).
- Ф.6 KindTeleportPlanet — TODO (отдельная сессия). **Препятствия**:
  (а) в `projects/game-nova/backend/` нет HTTP-клиента к billing-сервису
  (oxsar credits) — есть только локальный `users.credits`, который не
  годится для премиум-фичи;
  (б) Idempotency-Key middleware есть в `projects/billing/`, но не
  подключён к API-роутеру game-nova ([cmd/server/main.go]
  paths /api/planets/...);
  (в) телепорт требует REST + OpenAPI с нуля + cooldown handler +
  artefact-проверки (ARTEFACT_PLANET_TELEPORTER в legacy) — связка
  ~600-800 строк в 3 новых подсистемах.
  Запланировано: ввести billing-client в game-nova → подключить idempotency
  middleware → POST /api/planets/{id}/teleport + KindTeleportPlanet handler.
- Kind'ы EXCHANGE_* (Expire/Ban) **перенесены в план 68** — биржа артефактов
  реализует их в рамках `internal/exchange/`. Обоснование (2026-04-28):
  stub-handler с `ErrSkip` нарушал бы R15 (без TODO/MVP-сокращений), а
  концептуально оба Kind'а — биржевые и должны жить рядом со своим
  service'ом, не в общем `internal/event/handlers.go`. Снижение scope
  плана 65 с 6 Kind'ов до 5.
**Зависимости**: блокируется планом 64 (`configs/balance/origin.yaml` — для balance numbers).
**Связанные документы**:
- [62-origin-on-nova-feasibility.md](62-origin-on-nova-feasibility.md)
- [docs/research/origin-vs-nova/divergence-log.md](../research/origin-vs-nova/divergence-log.md) —
  записи D-031..D-037
- [docs/research/origin-vs-nova/roadmap-report.md](../research/origin-vs-nova/roadmap-report.md) —
  Часть I.5 (R1-R5) + раздел плана 65

---

## Цель

Реализовать недостающие event-Kind'ы в game-nova event-loop'е,
которые есть в game-origin-php (75 типов) и нужны для вселенной
origin. Расширить существующие хендлеры под origin-сценарии.

---

## Что делаем (по D-NNN)

| D-NNN | Kind | Что добавляем |
|---|---|---|
| D-031 | `KindDemolishConstruction` | Объявлен в kinds.go, но handler пустой |
| D-035 | `KindDeliveryUnits`, `KindDeliveryResources`, `KindDeliveryArtefacts` | Доставка флотом разных payload |
| план 20 Ф.5 | `KindStargateTransport`, `KindStargateJump` | Уже частично |
| D-037 | `KindAttackDestroyBuilding`, `KindAttackAllianceDestroyBuilding` | Атака с целью разрушения постройки |
| D-032 + U-009 | `KindTeleportPlanet` | Телепорт планеты на новые координаты — премиум-фича через оксары (общий знаменатель для всех вселенных, по решению 2026-04-28) |
| — | `KindArtefactDisappear` | Артефакт исчезает |
| D-034 (опц.) | `KindRunSimAssault` | Отложенный запуск симулятора боя |

Для каждого: handler в `internal/event/handlers.go`, payload-схема
JSON, идемпотентность через advisory locks (план 32), запись в
audit_log, тесты.

---

## Что НЕ делаем

- Не вводим **турниры** (D-038, EVENT_TOURNAMENT_*) — отдельный
  план после плана 74 (см. roadmap §«Что НЕ делать»).
- Не реализуем 6 заглушек HOLDING_AI (Repair, AddUnits, ...) —
  в origin они тоже no-op.

---

## Этапы

### Ф.1. Эталонный handler — KindDemolishConstruction ✅ (2026-04-28)

Реализован как «единичный end-to-end», задающий паттерн для остальных
6 Kind'ов. См. [HandleDemolishConstruction](../../projects/game-nova/backend/internal/event/handlers.go)
+ [demolish_test.go](../../projects/game-nova/backend/internal/event/demolish_test.go).

**Установленный паттерн** (для следующих Kind'ов следовать):

| Аспект | Решение | Откуда |
|---|---|---|
| **Payload (R13)** | typed Go-struct `BuildingPayload` (переиспользован, идентичная форма с `BuildConstruction`); JSON-тэги snake_case | R1, R13 |
| **Идемпотентность** | сравнение текущего состояния с целевым (`cur <= target_level`) → no-op + закрыть очередь. Нет нужды в advisory locks: worker уже даёт `FOR UPDATE SKIP LOCKED` per-event | план 09 |
| **Prometheus (R8)** | автоматически на уровне worker'а: `oxsar_events_processed{kind,status}` counter + `oxsar_event_handler_seconds{kind}` histogram. Handler ничего сам не пишет | worker.go:278-345 |
| **Audit** | структурированный `slog.InfoContext` с полями `event_id, planet_id, unit_id, level_from, level_to`. Отдельной таблицы для player-action audit в nova нет (`admin_audit_log` — admin-only). Slog уезжает в централизованный лог-агрегатор. Если понадобится SQL-доступ — payload остаётся в `events` / `events_dead` | R3 |
| **Очки** | НЕ инкрементить в handler'е (отличие от legacy oxsar2). Очки derived state, пересчитываются `ScoreRecalcAll` (батч) или decorator `withScore` (per-user, после handler'а) | score/service.go |
| **used_fields** | зеркало `HandleBuildConstruction`: при demolish до 0 → `used_fields - 1` через `GREATEST(...,0)` (защита от рассинхрона) | план 23 |
| **Тесты** | (1) pure round-trip JSON payload, (2) property-based rapid (R4) на детерминизм skip-decision, (3) golden 3 сценария через `TEST_DATABASE_URL` (level 5→4, 1→0 с освобождением поля, idempotent replay), (4) валидация negative target_level | R4, helpers_test.go-стиль |
| **Регистрация** | в `cmd/worker/main.go` рядом с BuildConstruction; декораторы `withAchievement(withScore(...))` (без `withDailyQuest`: квеста «снеси здание» в дизайне нет) | worker/main.go:213 |
| **R10 (per-universe)** | соблюдено — `events.user_id/planet_id` уже фильтруются вселенной через FK на `users/planets` | план 36 |
| **R12 (i18n)** | не применимо — handler не возвращает user-facing строк | — |
| **R15** | без TODO/MVP-сокращений | R15 |

**Сознательное упрощение** (зафиксировать в simplifications.md):
NЕТ публичного API `POST /api/planets/{id}/demolish` и `building.Demolish()`
service-метода — добавляется отдельным планом, когда дойдёт UI. Handler
готов «принимать» события, кто бы их ни вставил. Альтернатива (полный
service+API) расширила бы scope с эталонного Kind'а до полноценной
фичи (~+400 строк, требует i18n, OpenAPI, FE), что нарушило бы
«один Kind за сессию».

**Не закрытый D-NNN**: записи D-031..D-037 в `divergence-log.md`
относятся к разным событиям (D-031 = TOURNAMENT_*, не demolish).
Проблема «handler пуст» не имела отдельного D-NNN — это inventory-bug,
зафиксирован в [divergence-log.md D-031b](../research/origin-vs-nova/divergence-log.md#d-031b).

### Ф.2. KindDeliveryArtefacts ✅ (2026-04-28)

Реальный handler по эталону Ф.1 — доставка артефактов флотом-курьером
(источник: биржа артефактов плана 68 либо premium-механика подарков).
Закрывает D-035.

**Что сделано**:

- `KindDeliveryArtefacts Kind = 23` в [kinds.go](../../projects/game-nova/backend/internal/event/kinds.go)
  (свободный номер рядом с DeliveryUnits=21 и DeliveryResources=22; в legacy
  origin EVENT_DELIVERY_ARTEFACTS=29).
- [HandleDeliveryArtefacts](../../projects/game-nova/backend/internal/event/handlers.go)
  + typed payload `DeliveryArtefactsPayload{FleetID, ArtefactIDs[]}`
  (R13).
- [delivery_artefacts_test.go](../../projects/game-nova/backend/internal/event/delivery_artefacts_test.go):
  pure round-trip + property-based (rapid, R4) + 4 golden-сценария +
  payload-validation (5 негативных кейсов).
- Регистрация в [cmd/worker/main.go](../../projects/game-nova/backend/cmd/worker/main.go):
  `withAchievement(event.HandleDeliveryArtefacts)` — без `withScore`
  (артефакты не входят в очки), без `withDailyQuest` (нет такого квеста).

**Семантика handler'а** (порт от EventHandler::transport ветка
EVENT_DELIVERY_ARTEFACTS, EventHandler.class.php:2718-2754 + Artefact::onOwnerChange):

1. Флот в state ≠ `outbound` → no-op (ArriveHandler-паттерн).
2. Для каждого артефакта в payload:
   - `artefacts_user.user_id` ← `e.UserID`, `planet_id` ← `e.PlanetID`;
   - active → held (см. ниже про revert);
   - per-universe (R10): обе стороны в одной вселенной, иначе ошибка.
3. Флот → `returning`.
4. Идемпотентность: артефакт уже у получателя → skip; флот returning → no-op.

**Сознательное упрощение** (зафиксировано в [simplifications.md](../simplifications.md)):
не вызываем `applyChange(revert)` синхронно при `active → held` —
полагаемся на то, что nova вычисляет effect-стек по списку активных
артефактов на каждом чтении (`ActiveBattleModifiers`, `service.go:349`).
Биржевая операция (план 68) обязана ставить артефакт в `held` ДО полёта
— тогда delivery просто переписывает владельца. Если в проде поймаем
`active`-артефакт в delivery — добавим явный revert-вызов отдельным
планом.

**Не закрытый D-NNN**: D-035 в [divergence-log.md](../research/origin-vs-nova/divergence-log.md#d-035-event_delivery_artefacts-доставка-артефактов-флотом).

### Ф.3. KindAttackDestroyBuilding ✅ (2026-04-28)

Атака с целью разрушить постройку — Kind=26. Обработчик переиспользует
`TransportService.AttackHandler()` ([fleet/attack.go](../../projects/game-nova/backend/internal/fleet/attack.go))
с новой веткой destroy-building, аналогично существующей destroy-moon
(Kind=25). Реализация общей логики разрушения здания вынесена в
[fleet/destroy_building.go](../../projects/game-nova/backend/internal/fleet/destroy_building.go).

**Что сделано**:

- `KindAttackDestroyBuilding Kind = 26` в [event/kinds.go](../../projects/game-nova/backend/internal/event/kinds.go)
  (legacy origin EVENT_ATTACK_DESTROY_BUILDING=23, но 23 уже занят
  KindDeliveryArtefacts из Ф.2 — берём свободный 26).
- Расширен `transportPayload` опциональным `TargetBuildingID int`
  (R13 typed payload). Поле omitempty — обратная совместимость с
  существующими событиями.
- Ветка в `AttackHandler`: после `applyDefenderLosses` и до
  `finalizeAttack`, при `e.Kind == KindAttackDestroyBuilding`,
  вызывается `tryDestroyBuilding(ctx, tx, planetID, isMoon, winner,
  targetUnitID, seed)` — общая функция destroy_building.go.
- Регистрация в [worker/main.go](../../projects/game-nova/backend/cmd/worker/main.go)
  с `withAchievement` (без withScore — score derived state, batch-пересчёт).

**Семантика** (порт от Assault.class.php:599-651):

1. Срабатывает только при `winner=="attackers"` и `!isMoon` (для лун —
   Kind=25, отдельная ветка).
2. `target_building_id` берётся из payload (выбор атакующего на момент
   запуска миссии) либо случайно из buildings планеты, кроме
   UNIT_EXCHANGE=107 и UNIT_NANO_FACTORY=7 (origin-фильтр, consts.php:317-327).
3. Уровень здания понижается на 1 (или удаляется, если level=1→0;
   при удалении освобождается 1 used_field планеты — зеркало
   HandleDemolishConstruction).
4. Сообщения (i18n): защитнику — `assaultReport.buildingDestroyed*`,
   атакующему — `assaultReport.enemyBuildingDestroyed*` (R12).
5. Audit (R3): структурированный slog с полями event_id, planet_id,
   unit_id, level_from, level_to, attacker/defender_user_id.
6. Метрики (R8): автоматически на уровне worker'а
   (`oxsar_events_processed{kind="26"}`).

**Сознательное упрощение** (зафиксировано в
[simplifications.md](../simplifications.md#план-65-фф3-ф4-разрушение-зданий-без-эвристики-сравнимого-уровня)):
не реализована legacy-эвристика «у атакующего должно быть здание
сравнимого уровня» (Assault.class.php:253-272, константа
`DESTROY_BUILD_RESULT_MIN_OFFS_LEVEL`). В nova миссия более прямолинейна
— random-выбор из всех eligible зданий, без тонкой балансировки.
Возвратиться при балансовой настройке, если выяснится дисбаланс.

### Ф.4. KindAttackAllianceDestroyBuilding ✅ (2026-04-28)

ACS-вариант Ф.3 — Kind=29. Обработчик переиспользует
`TransportService.ACSAttackHandler()` ([fleet/acs_attack.go](../../projects/game-nova/backend/internal/fleet/acs_attack.go))
с веткой destroy-building после ACS Moon Destruction.

**Что сделано**:

- `KindAttackAllianceDestroyBuilding Kind = 29` в kinds.go
  (legacy EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING=24, но 24 свободен —
  берём 29 для группировки рядом с 25/27/29 destroy-вариантами).
- Расширен `acsPayload` опциональным `TargetBuildingID int`. У всех
  флотов группы должен быть одинаковый TargetBuildingID — валидация
  на стороне инициатора миссии (тот же контракт что у moon-destroy).
- Та же `tryDestroyBuilding` что у Ф.3 — единая логика для single и ACS.
  Атрибуция атакующего — `lead.ownerUserID` (leader группы).
- Регистрация в worker/main.go.

**Не закрытый D-NNN**: D-037 ✅ (closing comment в divergence-log).

### Ф.5. KindAllianceAttackAdditional ✅ (2026-04-28, no-op handler)

Служебный referrer для ACS — Kind=30. В legacy origin это no-op в
event-loop (EventHandler.class.php:707-708: `case EVENT_ALLIANCE_ATTACK_ADDITIONAL: break`),
сам тип события используется только как маркер «дополнительный флот,
примыкающий к ACS-атаке».

**В nova ACS архитектурно иной** — все флоты группы получают одно
KindAttackAlliance с общим acs_group_id, и leader выполняет всю
работу за группу (см. [fleet/acs_attack.go](../../projects/game-nova/backend/internal/fleet/acs_attack.go)).
KindAllianceAttackAdditional концептуально излишен — но регистрируем
его как явный no-op для:
1. совместимости с возможной репликацией events из game-origin-php
   (если когда-нибудь введём общую events-таблицу для legacy/nova);
2. не давать событиям этого Kind'а уезжать в `StateError` при импорте
   архива origin.

**Что сделано**:

- `KindAllianceAttackAdditional Kind = 30` в kinds.go.
- `HandleAllianceAttackAdditional` в [event/handlers.go](../../projects/game-nova/backend/internal/event/handlers.go) —
  тривиальный no-op handler с info-slog для отладки (R3).
- Регистрация в worker/main.go (без декораторов).
- Pure-тесты в [event/alliance_attack_additional_test.go](../../projects/game-nova/backend/internal/event/alliance_attack_additional_test.go) —
  no-op для любого payload (включая невалидный JSON, nil tx).

**R15-уточнение**: НЕ trade-off в simplifications.md — no-op handler
адекватно отражает no-op-семантику legacy. R8/R9/R12 неприменимы (нет
мутации, нет user-facing вывода).

### Тесты Ф.3-Ф.5

- [event/alliance_attack_additional_test.go](../../projects/game-nova/backend/internal/event/alliance_attack_additional_test.go):
  pure-тесты Ф.5 no-op (5 кейсов: nil/empty/object/foreign/malformed payload).
- [fleet/destroy_building_test.go](../../projects/game-nova/backend/internal/fleet/destroy_building_test.go):
  payload round-trip Ф.3+Ф.4 (transportPayload + acsPayload), property-based
  rapid (R4) на детерминизм no-op-decision, golden-сценарии для
  `tryDestroyBuilding` через TEST_DATABASE_URL (7 сценариев: explicit
  level 5→4, level 1→0 + used_fields-1, defenders-win no-op, moon
  no-op, random skip excluded units, random pick eligible only,
  idempotent explicit).

### Ф.6. KindTeleportPlanet — отдельная сессия

Премиум-фича через оксары. Зависит от плана 69 (миграция
`users.last_planet_teleport_at` — есть) и интеграции с billing-сервисом.
POST /api/planets/{id}/teleport с Idempotency-Key. ~500+ строк (билинг
+ REST + OpenAPI + cooldown), заслуживает изоляции в отдельной сессии.

### Ф.7. Финализация (после Ф.2-Ф.6)

Smoke с тестовой вселенной origin, e2e-проверка, финализация плана.

## Конвенции (R1-R5)

- Имена Kind'ов в Go — `KindXxx Kind = NN` (см. существующий `kinds.go`).
  Для origin-only — добавить комментарий «// origin-only».
- payload-поля в JSON — snake_case.
- Тесты — golden + property-based (R4).

## Объём

3-4 недели. ~1000-2000 строк Go + тесты.

## References

- D-031..D-037 в `divergence-log.md`.
- Существующий `internal/event/kinds.go` — формат добавления Kind'ов.
- План 09 (event-system) — паттерны handler'ов, надёжность.
- План 32 (multi-instance) — Postgres advisory locks для идемпотентности.
