# Промпт: выполнить план 66 Ф.5 (платный выкуп удержания оксарами)

**Дата создания**: 2026-04-28
**План**: [docs/plans/66-remaster-alien-ai-full-parity.md](../../plans/66-remaster-alien-ai-full-parity.md)
**Зависимости**: ✅ план 64, ✅ план 66 Ф.1-Ф.4 (commit 3baf42798d), ✅ план 77 (billing-client).
**Объём**: ~200-400 строк Go + тесты, 1 коммит.

---

```
Задача: выполнить план 66 Ф.5 — платный выкуп холдинга алиенами
через billing-сервис в оксарах.

КОНТЕКСТ:

План 66 Ф.1-Ф.4 закрыт. AlienAI HoldingAI работает в 8 sub-phases.
План 77 закрыт коммитом 70d448a601 — billing-client готов
(Spend/Refund + sentinel errors + idempotency-middleware).

В legacy origin (AlienAI.class.php) выкуп холдинга работает так:
игрок платит N оксаров → handler удаляет HoldingAI-event и снимает
блокировку планеты (плюс возвращает захваченные ресурсы согласно
формуле). В origin это была кнопка «Откупиться» в HoldingAI UI.

В nova: backend-только в этой фазе. UI (если потребуется в nova)
делается отдельно после интеграции AlienAI в новый origin-фронт
(план 72). Сейчас — только endpoint + handler + тесты.

R1 / ADR-0009: выкуп — за **оксары** (hard currency, ст. 437 ГК),
не оксариты. Оксары списываются через billing.Spend.

R0-исключение: применимо ко всем вселенным (uni01/uni02 + origin),
как и весь AlienAI.

ПЕРЕД НАЧАЛОМ:

1) git status --short. cat docs/active-sessions.md.

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/66-remaster-alien-ai-full-parity.md (твоё ТЗ, Ф.5)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5» R0-R15
   - projects/game-nova/backend/internal/origin/alien/holding_ai_handler.go
     (как HoldingAI устроен — что нужно отменять при buyout)
   - projects/game-nova/backend/internal/billing/client/client.go
     (Spend signature)

3) Прочитай выборочно:
   - commit 3baf42798d (Ф.4 HoldingAI 8 sub-phases — какие state'ы
     отменяем при buyout)
   - commit 70d448a601 (как использовать billing-client +
     idempotency-middleware)
   - AlienAI.class.php в projects/game-origin-php/ (поиск по
     `buyout|откуп|выкуп|alien_ransom` — сверь формулу стоимости)

4) Добавь свою строку в docs/active-sessions.md:
   | <slot> | План 66 Ф.5 buyout оксарами | projects/game-nova/backend/internal/origin/alien/buyout_handler.go projects/game-nova/api/openapi.yaml | <дата-время> | feat(alien): платный выкуп оксарами (план 66 Ф.5) |

ЧТО НУЖНО СДЕЛАТЬ:

1. **Уточни формулу стоимости**:
   - Найди в legacy AlienAI: либо фиксированная цена в consts.php,
     либо формула от мощности удерживающего флота / от уровня игрока.
   - Параметризуй в configs/balance/origin.yaml как
     `alien_buyout_base_oxsars` и (если формула) множители.
   - Если ничего нет в legacy — фиксированная цена + ADR-комментарий
     в плане 66 «новая фича, согласованная при ремастере».

2. **OpenAPI первым (R2)**:
   - POST /api/alien-missions/{mission_id}/buyout (или
     /api/planets/{id}/alien-buyout — выбери по эталону существующих
     /api/planets/{id}/* endpoint'ов).
   - Header `Idempotency-Key` обязателен (R9).
   - 200 `{cost_oxsars, freed_at}`, 402 insufficient_oxsars,
     404 mission_not_found, 409 not_in_holding_state /
     idempotency_conflict, 503 billing_unavailable.

3. **Backend handler**:
   - Загрузить mission по ID, проверить что она в HOLDING_AI state
     (иначе 409).
   - Проверить что user — владелец удерживаемой планеты (иначе
     403/404).
   - Рассчитать cost_oxsars из конфига (детерминированно, R4).
   - Списать оксары через billing-client.Spend(ctx, userID,
     costOxsars, idempotencyKey, "alien_buyout:"+missionID).
     - При ErrInsufficientOxsar → 402.
     - При ErrBillingUnavailable → 503.
     - При ErrIdempotencyConflict → 409.
   - В одной транзакции: удалить event'ы HoldingAI / FlyUnknown
     для этой mission, удалить mission row, разблокировать планету
     (UPDATE planets SET locked_by_alien=false WHERE id=?).
   - audit_log: alien_buyout_paid (alien-сервис уже пишет в
     audit_log по эталону Ф.3 — см. handlers.go).
   - R3 slog: trace_id, user_id, planet_id, mission_id, cost_oxsars.
   - R8 Prometheus: oxsar_alien_buyout_total{status},
     oxsar_alien_buyout_oxsars (counter sum).
   - R10: WHERE universe_id во всех queries (mission per-universe).

4. **Idempotency-middleware** подключи к этому endpoint'у в
   cmd/server/main.go (как в плане 77 Spec / portal-backend).

5. **Тесты**:
   - buyout_handler_test.go — mock billing-client:
     - happy-path: 200, mission удалена, planet разблокирована.
     - mission в неправильном state (не HOLDING_AI) → 409.
     - другая user_id → 403/404.
     - insufficient oxsars → 402, mission не тронута.
     - billing unavailable → 503, mission не тронута.
     - idempotency conflict → 409.
   - Property-based (rapid, R4): cost детерминирован для одной и той
     же mission (тех же входных параметров).
   - integration_test (auto-skip без TEST_DATABASE_URL): happy-path
     с реальной БД и тестовым billing-stub.

6. **i18n (R12)**:
   - Grep `projects/game-nova/configs/i18n/{ru,en}.yml` на
     `alien_buyout|alien.holding|выкуп|ransom`.
   - Новые ключи: `alien.buyoutSuccess`,
     `alien.buyoutInsufficientOxsars`, `alien.buyoutNotInHolding`.
     **В коммите указать соотношение переиспользовано/новых**.

══════════════════════════════════════════════════════════════════
ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА R0-R15 — см. блок rules-r0-r15.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/rules-r0-r15.md
══════════════════════════════════════════════════════════════════

══════════════════════════════════════════════════════════════════
GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ — см. блок git-isolation.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/git-isolation.md

Свои пути для CC_AGENT_PATHS:
- projects/game-nova/backend/internal/origin/alien/buyout_handler.go
- projects/game-nova/backend/internal/origin/alien/buyout_handler_test.go
- projects/game-nova/backend/internal/origin/alien/buyout_integration_test.go
- projects/game-nova/api/openapi.yaml (только секция alien-buyout)
- projects/game-nova/backend/cmd/server/main.go (только новый route)
- projects/game-nova/configs/i18n/ru.yml (только alien.buyout* ключи)
- projects/game-nova/configs/i18n/en.yml (только alien.buyout* ключи)
- configs/balance/origin.yaml (только alien_buyout_*)
- docs/plans/66-remaster-alien-ai-full-parity.md
- docs/active-sessions.md
══════════════════════════════════════════════════════════════════

КОММИТЫ:

Один коммит: feat(alien): платный выкуп оксарами (план 66 Ф.5)

Trailer: Generated-with: Claude Code

ВСЕГДА:
git commit -m "..." -- $CC_AGENT_PATHS

ЧЕГО НЕ ДЕЛАТЬ:

- НЕ списывать оксары вне billing-client.
- НЕ забывать про Idempotency-Key (R9).
- НЕ забывать про R8 Prometheus.
- НЕ менять формулу гадая — сначала проверь legacy AlienAI.class.php.
- НЕ задевать Ф.6 (golden-итерации) и Ф.7 (финал) — только Ф.5.
- НЕ забывать про -- в git commit.

УСПЕШНЫЙ ИСХОД:

- POST endpoint работает, Idempotency-Key корректно.
- billing.Spend списывает оксары; mission и event'ы удалены
  атомарно.
- 6+ тестов покрывают все коды ответов.
- Шапка плана 66 Ф.5 ✅.
- Запись в docs/project-creation.txt — итерация 66 Ф.5.
- Удалена строка из docs/active-sessions.md.

Стартуй.
```
