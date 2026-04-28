# Промпт: выполнить план 66 (AlienAI полный паритет)

**Дата создания**: 2026-04-28
**План**: [docs/plans/66-remaster-alien-ai-full-parity.md](../plans/66-remaster-alien-ai-full-parity.md)
**Зависимости**: блокируется планом 64 (алиен-юниты в `configs/balance/origin.yaml` + дефолте).
**Объём**: ~3 нед, ~800-1000 строк Go + golden-тесты на 50+ итераций.

---

```
Задача: выполнить план 66 (ремастер) — AlienAI до полного паритета
с oxsar2-classic. Достроить план 15 этап 3.

R0-ИСКЛЮЧЕНИЕ: AlienAI работает во ВСЕХ вселенных (uni01/uni02 +
origin), не только origin. Это сознательный upgrade игрового опыта
modern (решение пользователя 2026-04-28).

ПЕРЕД НАЧАЛОМ:

1) git status --short.

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/66-remaster-alien-ai-full-parity.md
   - docs/plans/15-alien-holding-thursday.md (что уже сделано)
   - docs/research/origin-vs-nova/alien-ai-comparison.md (A1-A14
     расхождения, state machine, переходы, формулы)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5»
     (R0-R15)

3) Прочитай выборочно:
   - projects/game-origin-php/src/game/AlienAI.class.php (1127 строк
     — главный референс).
   - projects/game-nova/backend/internal/alien/ — что уже есть.
   - configs/balance/origin.yaml + дефолтные configs/units.yml
     (алиен-юниты должны быть после плана 64).

ЧТО НУЖНО СДЕЛАТЬ:

Backend (новый/расширение модулей в internal/origin/alien/):
- KindAlienFlyUnknown handler (грабёж/подарок/атака альтернативы).
- KindAlienGrabCredit — отдельный сценарий кражи **оксаритов** (не
  кредитов; ADR-0009 / R1 валюта).
- KindAlienChangeMissionAI (control_times, power_scale).
- Расширение KindAlienHoldingAI до 8 действий (2 активных + 6
  заглушек как в origin).
- generateFleet() — target_power, итеративное добавление кораблей.
- Множитель «четверг» ×5 / ×1.5..2.0 — параметр в configs/.
- findTarget / findCreditTarget с критериями выбора цели.
- shuffleKeyValues — случайное ослабление техник.
- Платный выкуп удержания через billing-API (оксары для платежа,
  не оксариты).

Не вводим feature-флаги по вселенным — AlienAI работает одинаково
для всех (uni01/uni02/origin).

Тесты:
- Golden-тесты на 50+ итераций (property-based через rapid).
- Symbolic коэффициенты пакетов в YAML (Прометей-метрики через R8).

ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА:

R0-исключение: расширение AlienAI применяется к modern одинаково.
Это явное исключение зафиксировано в roadmap-report.md / Часть I.5.
R1: alien_unit_1..5 (snake_case, не UNIT_A_*). Поле БД для
alien-state: holds_until_at (TIMESTAMPTZ).
R8: Prometheus метрики для AI-итераций (длительность, цели).
R10: per-universe изоляция в alien-таблицах (universe_id).
R13: typed payload в KindAlien*.
R15: без упрощений.

ВАЛЮТА (R1 особый случай / ADR-0009):
- Грабёж = **оксариты** (soft). Имя Kind'а KindAlienGrabCredit
  историческое, фактическая валюта — оксариты.
- Подарки = **оксариты**.
- Платный выкуп удержания = **оксары** (hard, через billing).

GIT-ИЗОЛЯЦИЯ:
- Свои пути: internal/origin/alien/ (НЕ internal/legacy/alien/!),
  internal/event/handlers.go (для регистрации Kind'ов alien),
  configs/ (если требуется доп. параметр в default), тесты,
  docs/plans/66-..., docs/research/origin-vs-nova/divergence-log.md
  (D-036), alien-ai-comparison.md (A1-A14 пометить ✅).

КОММИТЫ:

3 коммита (рекомендация):
1. feat(origin/alien): state machine + переходы + helpers
2. feat(origin/alien): Kind handlers + KindAlien* (полный паритет)
3. feat(origin/alien): golden-тесты + Prometheus + финализация

ЧЕГО НЕ ДЕЛАТЬ:
- НЕ внедрять 6 заглушек HOLDING_AI как полные действия — в origin
  они тоже no-op.
- НЕ создавать internal/legacy/alien/ — пакет origin (legacy
  только d:\Sources\oxsar2 и game-origin-php).
- НЕ забывать что валюта = оксариты (не кредиты).

УСПЕШНЫЙ ИСХОД:
- 8 EVENT_ALIEN_* Kind'ов работают полнофункционально во всех
  вселенных.
- generateFleet/findTarget/shuffleKeyValues реализованы.
- 50+ golden-итераций зелёные.
- A1-A14 закрыты в alien-ai-comparison.md.
- D-036 закрыт в divergence-log.md.

Стартуй.
```
