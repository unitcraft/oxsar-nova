# Continuation: план 66 — Ф.4-Ф.7 (HoldingAI 8 действий, billing-выкуп, golden, финал)

**Применение**: вставить блок ниже в новую сессию Claude Code.
**Объём**: ~1-2 нед, ~500-800 строк Go + 50+ golden-итераций.

---

```
Задача: завершить план 66 (ремастер) — Ф.4-Ф.7. AlienAI движок.

КОНТЕКСТ:

Ф.1+Ф.2 закрыты коммитом 7f44744d49 (state machine, helpers,
generateFleet, target.go, shuffle.go) — internal/origin/alien/.

Ф.3 закрыта коммитом 856b5e105c — Kind handlers FlyUnknown,
GrabCredit, ChangeMissionAI + Spawner-проводка + pgx Loader
реализация + ~1914 строк production+тесты.

Эта сессия: Ф.4 (HoldingAI 8 действий), Ф.5 (платный выкуп
оксарами через billing), Ф.6 (50+ golden-итераций через PHP eval),
Ф.7 (финализация).

ПЕРЕД НАЧАЛОМ:

1) git status --short.

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/66-remaster-alien-ai-full-parity.md (твоё ТЗ)
   - docs/research/origin-vs-nova/alien-ai-comparison.md (A1-A14
     — state machine + переходы + параметры)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5»
     (R0-R15)

3) Прочитай выборочно:
   - КОММИТ 856b5e105c (Ф.3 — твой эталонный паттерн handlers).
   - projects/game-origin-php/src/game/AlienAI.class.php —
     строки 1-200 (HoldingAI subphases) + строки про выкуп.
   - projects/game-nova/backend/internal/origin/alien/ — твой
     пакет, продолжаешь.
   - projects/game-nova/backend/internal/billing/ — образец
     для платного выкупа (Idempotency-Key R9).
   - projects/game-nova/backend/internal/event/ — эталон
     KindDemolishConstruction (9a3992a384) для дополнительных
     Kind'ов если нужны.

ЧТО НУЖНО СДЕЛАТЬ:

Ф.4. HoldingAI до 8 действий (с заглушками для 6 неактивных):
- Расширить KindAlienHoldingAI handler (есть с Ф.3) до 8 sub-phases
  как в origin: 2 активных (extract resources, attack)
  + 6 заглушек (Repair, AddUnits, Move, Build, Research, no-op).
- Заглушки = no-op методы с TODO-комментарием «по образцу origin
  тоже no-op». Не пытайся реализовать что-то поверх origin.
- Property-based тесты на переходы между sub-phases.

Ф.5. Платный выкуп удержания оксарами:
- Endpoint POST /api/alien/holdings/{id}/buyout (R6 REST + R2
  OpenAPI).
- Списание оксаров через billing-API с Idempotency-Key (R9).
- Цена выкупа — параметр в configs/balance/origin.yaml::globals
  (или в Config из internal/origin/alien/config.go).
- При успешном списании — снять HoldingState с планеты (через
  internal/origin/alien/service.go).
- ВАЛЮТА: оксары (hard, ст. 437 ГК) — это реальные деньги, см.
  ADR-0009 / R1 «Особый случай: валюта». НЕ оксариты.

Ф.6. Golden-тесты на 50+ итераций:
- tools/dump-alien-state.php (PHP CLI) — дампит AlienAI tick
  результатов из live origin (по образцу dump-balance-formulas.php
  из плана 64).
- internal/origin/alien/testdata/golden_alien_ticks.json — 50+
  точек эталонов для разных конфигураций (target_power, scale,
  control_times, шанс grab/gift, состояние tech).
- Go-golden-тест читает JSON и сверяет с результатами своих
  функций. Допуск: точное совпадение для целых, абс. погрешность
  ≤ 1 для дробных.
- Property-based (через rapid): инварианты — fleet.target_power
  растёт при ×5 четверг, generateFleet не превышает available_units,
  shuffleKeyValues стабильно ослабляет в указанном диапазоне.

Ф.7. Финализация:
- Шапка плана 66 → ✅ Завершён <дата>.
- Запись в docs/project-creation.txt — итерация 66.
- В docs/research/origin-vs-nova/divergence-log.md — D-036 ✅.
- В docs/research/origin-vs-nova/alien-ai-comparison.md — A1-A14
  пометить как ✅ закрытые.
- Коммит финализации.

ПРАВИЛА (R0-R15):
- R0-исключение: AlienAI применяется во ВСЕХ вселенных
  (uni01/uni02 + origin). Зафиксировано в roadmap-report.
- R1: snake_case, оксары/оксариты по ADR-0009.
- R8: Prometheus counter+histogram для AI-итераций.
- R9: Idempotency-Key для buyout endpoint.
- R10: per-universe изоляция (universe_id в alien-таблицах).
- R12: i18n — grep nova-bundle.
- R13: typed payload.
- R15: без упрощений (50+ golden-итераций обязательны).

GIT-ИЗОЛЯЦИЯ:
- Свои пути: internal/origin/alien/ (твой пакет),
  configs/balance/ (если параметр выкупа в config),
  tools/dump-alien-state.php, openapi.yaml,
  docs/plans/66-..., docs/research/origin-vs-nova/* (D-036,
  A1-A14 пометки), docs/project-creation.txt.

КОММИТЫ:

3 коммита (рекомендация):
1. feat(origin/alien): HoldingAI 8 sub-phases (план 66 Ф.4).
2. feat(origin/alien): платный выкуп оксарами (план 66 Ф.5).
3. test(origin/alien): 50+ golden-итераций + финализация (план 66 Ф.6+Ф.7).

ЧЕГО НЕ ДЕЛАТЬ:
- НЕ реализовывать 6 заглушек HOLDING_AI как полные действия.
- НЕ менять валюту выкупа на оксариты — это hard (оксары).
- НЕ менять modern-числа nova.

УСПЕШНЫЙ ИСХОД:
- HoldingAI 8 sub-phases работают.
- Buyout через billing с Idempotency.
- 50+ golden-итераций зелёные.
- D-036, A1-A14 закрыты.
- План 66 ✅.

Стартуй.
```
