# Промпт: выполнить план 65 Ф.6 (KindTeleportPlanet)

**Дата создания**: 2026-04-28
**План**: [docs/plans/65-remaster-event-loop-origin-events.md](../../plans/65-remaster-event-loop-origin-events.md)
**Зависимости**: ✅ план 64 (origin.yaml), ✅ план 77 (billing-client + idempotency-middleware).
**Объём**: ~600-800 строк Go + тесты, 1-2 коммита.

---

```
Задача: выполнить план 65 Ф.6 — реализовать KindTeleportPlanet
(премиум-фича телепорта планеты на новые координаты, через
оксары/billing-сервис).

КОНТЕКСТ:

План 65 Ф.1-Ф.5 закрыт. Эталонный паттерн handler'а — Ф.1
KindDemolishConstruction (commit 9a3992a384) и Ф.3-Ф.5
attack-destroy-building (commit 1fec2edb64). EXCHANGE_*-Kind'ы
вынесены в план 68. Осталась только Ф.6 KindTeleportPlanet.

План 77 закрыт коммитом 70d448a601 — добавил
`internal/billing/client/` (Spend/Refund + sentinel errors
ErrInsufficientOxsar/ErrBillingUnavailable/ErrIdempotencyConflict),
`pkg/idempotency/middleware.go` (Chi-middleware с Redis),
`pkg/metrics/billing.go`, env BILLING_URL. Это снимает все три
препятствия из шапки плана 65 (отсутствие billing-client, отсутствие
idempotency-middleware, отсутствие схемы для премиум-телепорта).

Особенность R0-исключения: команды Kind'а EVENT_TELEPORT_PLANET в
origin/event_kinds.go нет — реализуем как новый Kind в game-nova
(применимо ко всем вселенным как премиум-фича через оксары —
общий знаменатель, по решению пользователя 2026-04-28).

Артефакт ARTEFACT_PLANET_TELEPORTER в legacy: в origin телепорт
гейтился артефактом + cooldown'ом. В nova артефакта нет в текущем
каталоге — гейтинг **только через оплату оксарами** (без артефакта,
по решению пользователя; artefact-проверку заложить интерфейсом
для будущего расширения, но не реализовывать).

ПЕРЕД НАЧАЛОМ:

1) git status --short. cat docs/active-sessions.md. Если чужие
   файлы пересекаются — спроси пользователя.

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/65-remaster-event-loop-origin-events.md (твоё ТЗ)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5»
     (R0-R15)
   - projects/game-nova/backend/internal/billing/client/client.go
     (Spend signature, errors)
   - projects/game-nova/backend/pkg/idempotency/middleware.go
     (как подключать в Chi-router)
   - projects/game-nova/backend/internal/event/handlers.go (как
     зарегистрирован эталонный KindDemolishConstruction)

3) Прочитай выборочно:
   - commit 9a3992a384 (эталон Ф.1 demolish — handler+payload+test
     pattern)
   - commit 1fec2edb64 (эталон Ф.3-Ф.5 attack — fleet/destroy_building.go
     общая логика, 7 golden-сценариев)
   - commit 70d448a601 (как использовать billing-client и
     idempotency-middleware в API-роутере)
   - projects/game-nova/backend/internal/planet/handler.go
     (паттерн POST /api/planets/{id}/...)
   - configs/balance/origin.yaml — есть ли уже `teleport_cost_oxsars`
     и `teleport_cooldown_hours`. Если нет — добавь как параметры.

4) Добавь свою строку в docs/active-sessions.md:
   | <slot> | План 65 Ф.6 KindTeleportPlanet | projects/game-nova/backend/internal/{event,planet}/ projects/game-nova/api/openapi.yaml docs/plans/65-... | <дата-время> | feat(event-loop): KindTeleportPlanet (план 65 Ф.6) |

ЧТО НУЖНО СДЕЛАТЬ:

1. **OpenAPI первым (R2)**:
   - POST /api/planets/{planet_id}/teleport — body
     `{target_galaxy:int, target_system:int, target_position:int}`,
     header `Idempotency-Key` обязателен (R9).
   - 200 `{event_id, fire_at, cost_oxsars}` (запланировано),
     400 invalid coords / cooldown still active,
     402 insufficient oxsars,
     404 planet not found,
     409 target slot occupied / idempotency conflict,
     503 billing unavailable.

2. **Backend handler POST /api/planets/{id}/teleport**:
   - Валидация координат (галактика 1..N_GALAXIES, система 1..N_SYSTEMS,
     позиция 1..N_POSITIONS из configs/balance/<universe>.yaml).
   - Проверка cooldown (`users.last_planet_teleport_at` — закрыто
     планом 69 коммитом 32d24a1f2b, поле есть).
   - Проверка occupied slot (SELECT FROM planets WHERE galaxy=? AND
     system=? AND position=? — должно быть пусто).
   - Списание оксаров через billing-client `Spend(ctx, userID,
     costOxsars, idempotencyKey, "planet_teleport:"+planetID)`.
     - При ErrInsufficientOxsar → 402.
     - При ErrBillingUnavailable → 503.
     - При ErrIdempotencyConflict → 409.
   - Запись KindTeleportPlanet в events с typed payload (R13)
     `TeleportPlanetPayload{TargetGalaxy, TargetSystem, TargetPosition,
     CostOxsars, IdempotencyKey}` и fire_at = now + duration
     (`teleport_duration_minutes` из configs/balance/<universe>.yaml,
     по умолчанию 0 = мгновенно или N мин по решению; уточни в
     legacy origin: cronjobs/EventHandler.php если есть EVENT_TELEPORT_PLANET).
   - Idempotency-middleware подключи к этому endpoint'у (как в
     billing/portal — кеш 24h в Redis по ключу пользователя+ключа).
   - R3 slog: trace_id, user_id, planet_id, event_id.
   - R8 Prometheus: oxsar_planet_teleport_total{status}, histogram
     длительности handler'а.

3. **Event-handler KindTeleportPlanet в `internal/event/handlers.go`**:
   - Эталон — KindDemolishConstruction.
   - При срабатывании: UPDATE planets SET galaxy=?, system=?, position=?
     WHERE id=? AND user_id=?.
   - UPDATE users SET last_planet_teleport_at=now WHERE id=?.
   - При ошибке (planet удалена / target теперь занят) — Refund
     через billing-client.Refund(ctx, ...) и audit-запись.
   - audit_log: event_planet_teleported.
   - R10: WHERE universe_id во всех SELECT (cross-universe телепорт
     запрещён — нет смысла в gameplay).

4. **Тесты**:
   - payload_test.go — round-trip JSON для TeleportPlanetPayload.
   - handler_test.go — mock billing-client:
     - happy-path (200, event запланирован),
     - cooldown active (400),
     - occupied slot (409),
     - insufficient oxsars (402),
     - idempotency conflict (409),
     - billing unavailable (503),
     - 401 без токена,
     - 400 invalid coords (галактика 0, система 999...).
   - integration_test (auto-skip без TEST_DATABASE_URL):
     happy-path с реальной БД через testdb-helper, + Refund при
     concurrent занятии слота.
   - Property-based (rapid, R4): cost корректный из конфига при
     любом cooldown.

5. **i18n (R12)**:
   - Grep `projects/game-nova/configs/i18n/{ru,en}.yml` на ключи
     `teleport*`, `planet_teleport*`. Переиспользуй существующие.
   - Новые ключи (если нужны): `teleport.cooldownActive`,
     `teleport.slotOccupied`, `teleport.insufficientOxsars`,
     `teleport.success`. **В коммите указать соотношение
     переиспользовано/новых**.

══════════════════════════════════════════════════════════════════
ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА R0-R15 — см. блок rules-r0-r15.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/rules-r0-r15.md
══════════════════════════════════════════════════════════════════

══════════════════════════════════════════════════════════════════
GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ — см. блок git-isolation.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/git-isolation.md

Свои пути для CC_AGENT_PATHS:
- projects/game-nova/backend/internal/event/teleport_handler.go
- projects/game-nova/backend/internal/event/teleport_handler_test.go
- projects/game-nova/backend/internal/event/payload_teleport.go
- projects/game-nova/backend/internal/planet/teleport_handler.go
- projects/game-nova/backend/internal/planet/teleport_handler_test.go
- projects/game-nova/api/openapi.yaml (только секция teleport)
- projects/game-nova/backend/cmd/server/main.go (только новый route)
- projects/game-nova/backend/cmd/worker/main.go (только регистрация Kind)
- projects/game-nova/configs/i18n/ru.yml (только teleport.* ключи)
- projects/game-nova/configs/i18n/en.yml (только teleport.* ключи)
- configs/balance/origin.yaml (только teleport_cost_oxsars/cooldown)
- docs/plans/65-remaster-event-loop-origin-events.md
- docs/active-sessions.md
══════════════════════════════════════════════════════════════════

КОММИТЫ:

Conventional commit: feat(event-loop): KindTeleportPlanet (план 65 Ф.6)

Один коммит — целая Ф.6 (handler+endpoint+тесты вместе, иначе
endpoint без handler'а бессмысленен).

Trailer: Generated-with: Claude Code

ВСЕГДА:
git commit -m "..." -- $CC_AGENT_PATHS

ЧЕГО НЕ ДЕЛАТЬ:

- НЕ списывать оксары вне billing-client — только через Spend.
- НЕ забывать про Idempotency-Key (R9, ОБЯЗАТЕЛЬНО для POST).
- НЕ забывать про R8 Prometheus метрики — counter+histogram, 5-10
  строк, всегда (см. R15 раздел «🚫 ПРОПУСК»).
- НЕ хардкодить cost в коде — только из configs/balance/<universe>.yaml.
- НЕ менять modern-числа (R0): cost/cooldown — параметры в
  override-схеме (план 64), nova-default = origin-default (общий
  знаменатель).
- НЕ забывать про -- в git commit (4-й прецедент в memory).
- НЕ реализовывать гейтинг через ARTEFACT_PLANET_TELEPORTER —
  только оплата оксарами (по решению пользователя).

УСПЕШНЫЙ ИСХОД:

- KindTeleportPlanet зарегистрирован в worker.
- POST /api/planets/{id}/teleport работает с Idempotency-Key.
- billing-client.Spend списывает оксары; Refund при concurrent-конфликте.
- 6+ unit-тестов покрывают коды ответов; integration auto-skip без БД.
- i18n с ru+en; в commit-message: «переиспользовано/новых».
- Шапка плана 65: Ф.6 ✅, план 65 ЗАКРЫТ полностью.
- Запись в docs/project-creation.txt — итерация 65.
- Удалена строка из docs/active-sessions.md.

Стартуй.
```
