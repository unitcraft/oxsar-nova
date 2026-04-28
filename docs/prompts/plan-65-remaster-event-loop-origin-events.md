# Промпт: выполнить план 65 (event-loop события origin)

**Дата создания**: 2026-04-28
**План**: [docs/plans/65-remaster-event-loop-origin-events.md](../plans/65-remaster-event-loop-origin-events.md)
**Зависимости**: блокируется планом 64 (`configs/balance/origin.yaml`).
**Объём**: 3-4 нед, ~1000-2000 строк Go + тесты.

---

```
Задача: выполнить план 65 (ремастер) — расширение event-loop game-nova
под недостающие события из game-origin-php (75 типов в origin vs ~41
в nova).

ВАЖНОЕ:
- Зависит от плана 64: убедись, что configs/balance/origin.yaml уже
  существует и LoadFor работает.
- Эти события — общий знаменатель для ВСЕХ вселенных (uni01/uni02 +
  origin), не origin-only.

ПЕРЕД НАЧАЛОМ:

1) git status --short — изоляция от параллельных сессий.

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/65-remaster-event-loop-origin-events.md
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5»
     (R0-R15, особенно R0, R8, R10, R13)
   - docs/research/origin-vs-nova/divergence-log.md D-031..D-037
   - CLAUDE.md

3) Прочитай выборочно:
   - projects/game-nova/backend/internal/event/kinds.go — формат
     добавления Kind'ов.
   - projects/game-nova/backend/internal/event/handlers.go — паттерн
     handler'ов.
   - projects/game-origin-php/src/game/EventHandler.class.php —
     legacy-PHP референс (3573 строки).

ЧТО НУЖНО СДЕЛАТЬ:

7 новых Kind'ов в internal/event/:
- KindExchangeExpire (для плана 68 биржа)
- KindExchangeBan (служебный)
- KindDeliveryArtefacts (расширение DELIVERY-семьи)
- KindAttackDestroyBuilding (атака на разрушение постройки)
- KindAttackAllianceDestroyBuilding (ACS-вариант)
- KindAllianceAttackAdditional (referer для ACS)
- KindTeleportPlanet (премиум-фича через оксары; D-032 + U-009)

Также:
- KindDemolishConstruction handler — закрыть D-031 (объявлен но
  пустой).
- Идемпотентность всех новых handler'ов через advisory locks (план 32).
- Audit-log записи через internal/audit/ если применимо.
- Typed payload (Go-struct, R13) для каждого Kind, не сырой JSONB.
- Prometheus метрики (R8) — counter + histogram для каждого handler.
- Golden-тесты (R4).

Для KindTeleportPlanet — extra:
- POST /api/planets/{id}/teleport — endpoint с Idempotency-Key (R9).
- Списание оксаров (hard, через billing-сервис) — это премиум-фича.
- Cooldown через users.last_planet_teleport_at (план 69).

ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА (см. roadmap-report «Часть I.5»):

R0: не трогать чисел/формул modern. Эти Kind'ы — новая
функциональность, не пересмотр существующего.
R1: snake_case JSON payload, английский, полные слова. Имена Kind'ов:
KindExchangeExpire (НЕ KindExchExpire — полные слова).
R2: OpenAPI первым.
R8: Prometheus counter+histogram.
R9: Idempotency-Key для всех мутирующих endpoint'ов.
R10: per-universe изоляция — universe_id во всех новых таблицах
(если будут).
R13: typed payload через Go-struct.
R12: i18n — grep nova-bundle перед созданием новых ключей.
R15: без упрощений как для прода.

GIT-ИЗОЛЯЦИЯ: только свои пути
(internal/event/, internal/legacy не использовать!, тесты,
docs/plans/65-..., docs/project-creation.txt, divergence-log.md).

КОММИТЫ:

Можно одним: feat(event-loop): 7 новых Kind'ов для ремастера (план 65)
ИЛИ разбить на 2-3 (по группам Kind'ов).

ЧЕГО НЕ ДЕЛАТЬ:
- НЕ путать «legacy-фронт» / «legacy-режим» — это origin
  (legacy = только d:\Sources\oxsar2 и game-origin-php).
- НЕ создавать internal/legacy/event/ — используй
  internal/origin/event/ если нужен origin-специфический модуль
  (но эти 7 Kind'ов — общие для всех вселенных).
- НЕ реализовывать TOURNAMENT_*, RUN_SIM_ASSAULT, COLONIZE_RANDOM_PLANET,
  ALIEN_ATTACK_CUSTOM — они в отказе (см. roadmap-report «Часть V»).

УСПЕШНЫЙ ИСХОД:
- 7 Kind'ов в kinds.go + handlers.
- Typed payload + tests + Prometheus + Idempotency.
- D-031..D-037 закрыты в divergence-log.md.
- Все existing nova-тесты зелёные (R0).

Стартуй.
```
