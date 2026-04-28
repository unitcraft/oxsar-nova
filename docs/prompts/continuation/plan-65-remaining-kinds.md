# Continuation: план 65 — Ф.3+Ф.4+Ф.5 attack-destroy-building (одна сессия)

**Применение**: вставить блок ниже в новую сессию Claude Code.
**Объём**: ~5-6 часов, ~400-700 строк Go + тесты для 3 Kind'ов.

**После Ф.2 (KindDeliveryArtefacts, commit d6f1785dda)** в плане 65
осталось **4 Kind'а**, не 6:
- Ф.3 KindAttackDestroyBuilding
- Ф.4 KindAttackAllianceDestroyBuilding (ACS-вариант)
- Ф.5 KindAllianceAttackAdditional (служебный no-op referrer)
- Ф.6 KindTeleportPlanet — **отдельная сессия** (билинг + REST +
  OpenAPI + cooldown, нет billing-client в nova сейчас).

EXCHANGE_* вынесены в план 68 (биржа реализует их в рамках
internal/exchange/).

---

```
Задача: завершить план 65 (ремастер) — реализовать оставшиеся 6
Kind'ов event-loop'а по эталону KindDemolishConstruction.

КОНТЕКСТ:

Ф.1 плана 65 закрыта коммитом 9a3992a384 — KindDemolishConstruction
как эталонный handler. Шапка плана 65 помечена «Ф.1 ✅; остальные
6 Kind'ов — TODO по эталону».

Задача этой сессии: 3 Kind'а в одном коммите.
- Ф.3 KindAttackDestroyBuilding (D-037 одиночная атака на здание).
- Ф.4 KindAttackAllianceDestroyBuilding (D-037 ACS-вариант).
- Ф.5 KindAllianceAttackAdditional (служебный no-op referrer как
  в legacy EventHandler.class.php:707-708; в nova ACS уже
  консолидирован в KindAttackAlliance с acs_group_id, поэтому
  KindAllianceAttackAdditional становится no-op-маркером для
  совместимости с origin-payload'ами; обоснуй в коде комментарием
  + в simplifications.md если решишь, что не нужен вообще).

**EXCHANGE_* (KindExchangeExpire/Ban) вынесены в план 68** —
концептуально принадлежат биржевой подсистеме, не event-loop.
План 68 реализует их в рамках internal/exchange/.

**Ф.6 KindTeleportPlanet — ОТДЕЛЬНАЯ СЕССИЯ.** Не делать в этом
прогоне. Препятствия:
- В nova нет billing-client (только локальные users.credits, нет
  интеграции с oxsar-credits).
- Нет idempotency-middleware в API-роутере (есть только в billing/).
- Нужен REST + OpenAPI + cooldown handler с нуля.
~600-800 строк + миграция + 3 новые подсистемы — это
самостоятельный план.

ПЕРЕД НАЧАЛОМ:

1) git status --short.

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/65-remaster-event-loop-origin-events.md
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5»
     (R0-R15)
   - КОММИТ 9a3992a384 — эталонный паттерн (читай git show 9a3992a384):
     · typed payload struct (R13)
     · idempotency через advisory lock или статус очереди (R9)
     · audit-log запись через internal/audit/
     · Prometheus counter+histogram (R8)
     · golden-тест паттерн
   - docs/research/origin-vs-nova/divergence-log.md D-031..D-037

3) Прочитай выборочно:
   - projects/game-nova/backend/internal/event/kinds.go — формат.
   - projects/game-nova/backend/internal/event/handlers.go —
     регистрация.
   - existing handler из эталона (commit 9a3992a384) — образец.

ЧТО НУЖНО СДЕЛАТЬ:

Каждый из 6 Kind'ов end-to-end по эталону.
Для KindTeleportPlanet — extra:
- Endpoint POST /api/planets/{id}/teleport (R6 REST + R2 OpenAPI).
- Списание оксаров через billing-API (Idempotency-Key R9).
- Использование users.last_planet_teleport_at (миграция 0072) для
  cooldown.
- Может зависеть от плана 69 Ф.5 cooldown handler — если ещё не
  готов, реализуй cooldown логику здесь.

ПРАВИЛА (R0-R15):
- R0: не правь modern-числа.
- R6: REST API с нуля для KindTeleportPlanet.
- R9: Idempotency-Key для всех мутирующих.
- R10: per-universe изоляция (universe_id).
- R12: i18n — grep nova-bundle перед новыми ключами.
- R13: typed payload struct для каждого Kind.
- R15: без упрощений как для прода.

R15 УТОЧНЕНО (обязательно прочитай перед началом — roadmap-report.md
"Часть I.5 / R15 / Что СЧИТАЕТСЯ упрощением vs ПРОПУСК"):

🚫 НЕ КЛАССИФИЦИРУЙ КАК TRADE-OFF В simplifications.md:
- R8 Prometheus метрики (counter+histogram) — 5-10 строк, ОБЯЗАТЕЛЬНО
  для каждого нового handler/endpoint. Аргумент «низкочастотный»
  НЕ ПРИНИМАЕТСЯ — метрики нужны именно для редких операций.
- R9 Idempotency-Key для мутирующих endpoint'ов — middleware
  существует, подключить = 1-2 строки.
- R12 i18n — хардкод строки в Go-handler не оправдан, Tr() = тот же
  объём кода.
- R10 universe_id в новой таблице — 1 строка DDL.
- R3 slog с trace_id — стандартный паттерн.
- DB-level CHECK на размер — defense-in-depth, защищает от обхода
  handler'а.
- NOT NULL / FK constraint / Index — должны быть по умолчанию.

Если агент 65 Ф.2 (commit d6f1785dda) или 67 Ф.2 (commit 2fd010cd87)
что-то из этого пропустил и положил в simplifications.md — это сигнал
плохой работы. Не повторяй. Делай по правилам с самого начала.

✅ TRADE-OFF (можно записать в simplifications.md с обоснованием):
- Архитектурное отклонение от плана с причиной + планом возврата.
- Не реализованная фича из плана, явно отложенная до Ф.X.
- Реализация требует существенно больше работы, чем доступно
  (1+ часов агента) — отложить с явным «Приоритет M/L».

GIT-ИЗОЛЯЦИЯ:
- Свои пути: internal/event/, миграции если нужно (бери следующий
  свободный номер после 0077 — план 67 уже взял до 0077),
  openapi.yaml, тесты, docs/plans/65-..., divergence-log.md.
- git add поимённо.
- git status --short перед commit.

КОММИТЫ:

Можно одним: feat(event-loop): остальные 6 Kind'ов по эталону
(план 65 финализация). ИЛИ разбить на 2-3 коммита.

После — шапка плана 65 ✅, D-031..D-037 закрыты в divergence-log.

ЧЕГО НЕ ДЕЛАТЬ:
- НЕ реализовывать TOURNAMENT_*, RUN_SIM_ASSAULT,
  COLONIZE_RANDOM_PLANET, ALIEN_ATTACK_CUSTOM (отказ, см.
  roadmap-report «Часть V»).
- НЕ менять эталонный KindDemolishConstruction.
- НЕ менять modern-баланс.

УСПЕШНЫЙ ИСХОД:
- Все 6 Kind'ов работают, golden + property-based тесты зелёные.
- KindTeleportPlanet списывает оксары через billing.
- Все existing nova-тесты зелёные (R0).
- D-031..D-037 закрыты.
- План 65 ✅.

Стартуй.
```
