# План 65 (ремастер): Расширение event-loop — события origin

**Дата**: 2026-04-28
**Статус**: Скелет (детали допишет агент-реализатор при старте)
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
| D-032 + U-009 | `KindTeleportPlanet` | Телепорт планеты на новые координаты |
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

## Этапы (детали — при старте)

- Ф.1. Каждый Kind: payload-схема + handler + golden-тест.
- Ф.2. Идемпотентность (advisory locks, как в существующих handler'ах).
- Ф.3. Audit-log записи через `internal/audit/`.
- Ф.4. Smoke с тестовой вселенной origin (после плана 64).
- Ф.5. Финализация.

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
