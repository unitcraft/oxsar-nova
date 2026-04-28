# План 71 (ремастер): UX-микрологика origin → nova-frontend

**Дата**: 2026-04-28
**Статус**: Ф.1 готова (helper-функции + компоненты в `components/feedback/`,
тесты, применение в Buildings/Research/Shipyard/Repair/Resource/Fleet/Artefacts/
Achievements). Низкоприоритетные X-018..X-022 кроме X-021 — отложены
(см. ниже «Trade-off по приоритету»).
**Зависимости**: ничего критичного.
**Связанные документы**:
- [62-origin-on-nova-feasibility.md](62-origin-on-nova-feasibility.md)
- [docs/research/origin-vs-nova/nova-ui-backlog.md](../research/origin-vs-nova/nova-ui-backlog.md) —
  X-001..X-022 (22 записи UX-микрологики)
- [docs/research/origin-vs-nova/roadmap-report.md](../research/origin-vs-nova/roadmap-report.md) —
  R1-R5 + раздел плана 71

---

## Цель

Применить X-NNN записи (UX-микрологика origin) на nova-frontend
для **всех** вселенных (uni01/uni02/origin). Это «общий знаменатель»
накопленного годами UX-опыта legacy-проекта oxsar2, не только для origin.

---

## Что делаем (приоритеты)

⭐ — критичные для UX, делаем первыми:

| X-NNN | Что |
|---|---|
| ⭐ X-001 | Дефицит ресурсов с пометкой `(нужно X)` |
| ⭐ X-003 | Показ требований при `can_build = false` (конкретно «не хватает Энергетической технологии lvl 5») |
| ⭐ X-010 | Энергодефицит красным цветом |
| X-002 | Потребление красным при отрицательном балансе |
| X-013 | added_level +/- зелёное / красное |
| X-021 | Счётчик новых ачивок в navbar |
| X-014 | Ремонтные поля в боевом отчёте |
| X-007 | Нет слотов с подсчётом (X из Y) |
| X-008 | Статус артефактов (активен / истекает) |
| X-009 | Расширенный helptip при наведении |
| остальные 12 X-NNN | По мере приоритета |

---

## Что НЕ делаем

- Не дублируем X-логику в origin-фронте плана 72 — он pixel-perfect
  клонирует визуал, а UX-фишки приходят из nova-API + общие
  компоненты, либо из самого визуала origin (как pixel-perfect клона).
- Не вводим новые UX-фичи, которых нет в origin — это «забор»
  накопленного, не редизайн.

## Этапы (детали — при старте)

- Ф.1. Каждая X-NNN запись → React-компонент / hook. ✅
- Ф.2. Группировка по экранам (Constructions, Empire, Battle, Chat). ✅
- Ф.3. i18n строк (русский на старте, английский по возможности). ✅
  (новая группа `feedback:` в `configs/i18n/{ru,en}.yml`).
- Ф.4. Smoke в browser на разных сценариях (дефицит, выполнено,
  частично). 🟡 (компоненты применены, ручной обход экранов после
  параллельных PR).
- Ф.5. Финализация. ✅

## Реализация (Ф.1)

Сделанные компоненты в [src/components/feedback/](../../projects/game-nova/frontend/src/components/feedback/):

- `feedback.ts` — pure-функции: `computeDeficit`, `canAfford`, `numKind`,
  `energyKind`, `addedLevelKind`, `formatAddedLevel`, `slotsState`,
  `artefactStatusKind`, `expiryUrgency`. Покрыты тестами в
  `feedback.test.ts` (27 тестов, все edge-case'ы: пустой/полный
  дефицит, 0 как энергодефицит origin, expired в прошлом).
- `ResourceDeficitBadge.tsx` — `ResourceCostLine` + `ResourceDeficitBadge`,
  X-001 (с `(нужно X)` через явный минус во второй строке).
- `RequirementHint.tsx` — `UnmetRequirementList`, X-003. Имена
  технологий через `info`-группу (`keyToTKey` snake→camel).
- `EnergyValue.tsx` — `NumValue` (X-002) + `EnergyValue` (X-010).
- `AddedLevelBadge.tsx` — X-013 (готов, ждёт DTO с added_level).
- `FleetSlotsBadge.tsx` — X-007.
- `NewBadge.tsx` — X-021 (мини-бейдж для navbar).
- `HelpTip.tsx` — X-009, чистый CSS-tooltip без JS-либ.
- `useNewAchievementCount.ts` — X-021, hook + `useMarkAchievementsSeen`.

## Где применили

| Экран | X-NNN | Что |
|---|---|---|
| BuildingsScreen | X-001, X-003 | `ResourceCostLine` + `ResourceDeficitBadge` + `UnmetRequirementList` |
| ResearchScreen | X-001 | `ResourceCostLine` + `ResourceDeficitBadge` |
| ShipyardScreen | X-001 | `ResourceCostLine` + `ResourceDeficitBadge` (с total = cost × count) |
| RepairScreen | X-014 | indicator «нет свободных ремонтных полей» при `storage.free <= 0` |
| ResourceScreen | X-002, X-010, X-009 | `EnergyValue` для perHour-energy + HelpTip на колонке «⚡» |
| FleetScreen | X-007 | `FleetSlotsBadge` (full / almost / ok) |
| ArtefactsScreen | X-008 | цветной статус (active/charging/listed/gone) + ⏰ при expiringSoon |
| App.tsx + AchievementsScreen | X-021 | счётчик новых через localStorage `lastSeenAt`, бейдж в navbar |

CSS: добавлен `.ox-helptip` hover-эффект в `styles/app.css` (X-009).

## i18n: переиспользование

Метрика «переиспользовано / новых» по итогу:
- **Переиспользовано** ~95% строк (имена технологий из `info`,
  `levelAbbr`, `youHave`, `error` из `buildings`/`global`,
  `slots`/`slotsHint` из `fleet`).
- **Новых** ~16 ключей в группе `feedback:`
  (`fleetSlotsFull`, `expoSlots*`, `artefact*`, `energy*`,
  `insufficientRepairFields`, `newAchievementsAria`,
  `energyTipTitle`/`energyTipBody`).

## Trade-off по приоритету (R15)

Низкоприоритетные записи **X-004** (прогресс-бар заряда артефактов),
**X-005** (хранилище переполнено `free <= 0` для tech-артефактов),
**X-006** (блокировка ввода кораблей в миссии),
**X-011** (полоса rep_destroyed/rep_alive здания),
**X-012** (форма «требуемые конструкции» — фактически часть X-003,
закрыта),
**X-015** (active/inactive артефакты `(2/3)`),
**X-016** (max активных артефактов false2),
**X-017** (скидки trade-union на бирже — биржа не реализована, U-001),
**X-018** (login error div — стилизация уже есть в auth-screen),
**X-019** (real-time валидация формы регистрации),
**X-020** (premium-marker торговца — биржа U-001),
**X-022** (jQuery UI ui-state-error — у нас своя дизайн-система,
паритет не нужен) — **Приоритет L**, отложены до явного запроса.

Для каждой отложенной записи путь решения уже понятен (требуют либо
расширения DTO, либо реализации зависимых модулей вроде биржи U-001).

## Конвенции (R1-R5)

- Компоненты-helper'ы в `frontend/src/components/feedback/`
  (`ResourceDeficitBadge.tsx`, `RequirementHint.tsx`).
- Цвета через CSS-переменные nova theme (не хардкод).
- Tooltip'ы — единый компонент `Tooltip` из существующего
  `components/ui/`.
- Никаких новых API — всё на существующих DTO; если данных не
  хватает, расширить DTO (R2 — OpenAPI первым).

## Объём

2-3 недели. ~600-1200 строк frontend.

## References

- X-001..X-022 в nova-ui-backlog.md.
- Существующий `frontend/src/features/` — стиль компонентов.
