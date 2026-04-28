# Continuation: план 65 — остальные 6 Kind'ов по эталону

**Применение**: вставить блок ниже в новую сессию Claude Code.
**Объём**: ~2-3 нед, ~600-1200 строк Go + тесты по 6 Kind'ам.

---

```
Задача: завершить план 65 (ремастер) — реализовать оставшиеся 6
Kind'ов event-loop'а по эталону KindDemolishConstruction.

КОНТЕКСТ:

Ф.1 плана 65 закрыта коммитом 9a3992a384 — KindDemolishConstruction
как эталонный handler. Шапка плана 65 помечена «Ф.1 ✅; остальные
6 Kind'ов — TODO по эталону».

Задача этой сессии: реализовать остальные 6 Kind'ов:
- KindExchangeExpire (для плана 68 биржа)
- KindExchangeBan (служебный)
- KindDeliveryArtefacts
- KindAttackDestroyBuilding
- KindAttackAllianceDestroyBuilding (ACS)
- KindAllianceAttackAdditional (referer для ACS)
- KindTeleportPlanet (премиум-фича через оксары; зависит от
  users.last_planet_teleport_at — есть в миграции 0072 от плана 69)

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
