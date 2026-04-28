# Промпт: выполнить план 66 Ф.6+Ф.7 (golden-итерации + финализация)

**Дата создания**: 2026-04-28
**План**: [docs/plans/66-remaster-alien-ai-full-parity.md](../../plans/66-remaster-alien-ai-full-parity.md)
**Зависимости**: ✅ план 66 Ф.1-Ф.4 (commit 3baf42798d). Не зависит от Ф.5 — выполняется независимо.
**Объём**: ~300-500 строк тестов + ~50 строк docs, 1 коммит.

---

```
Задача: выполнить план 66 Ф.6 (50+ golden-итераций для AlienAI) и
Ф.7 (финализация — divergence-log, ui-backlog, шапка плана).

КОНТЕКСТ:

План 66 Ф.1-Ф.4 закрыт. Pure-функции AlienAI лежат в
`projects/game-nova/backend/internal/origin/alien/`. Уже есть
property-based тесты (handlers_property_test.go) и integration
golden (handlers_integration_test.go), но недостаточно для
доказательства паритета с legacy.

Цель Ф.6 — гарантировать математический паритет с legacy
PHP AlienAI.class.php через golden-файлы: 50+ детерминированных
итераций (фиксированный seed RNG), сравнение с эталоном из
PHP-eval.

Ф.7 — финализация: записи в divergence-log (D-036 закрыт),
ui-backlog (если применимо), шапка плана 66 ✅, project-creation.txt.

R4 (golden+property для battle/economy/event, покрытие ≥ 85%) —
эта фаза докрывает покрытие AlienAI до 85%+ изменённых строк.

ПЕРЕД НАЧАЛОМ:

1) git status --short. cat docs/active-sessions.md.

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/66-remaster-alien-ai-full-parity.md (твоё ТЗ — Ф.6+Ф.7)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5» R0-R15
   - projects/game-nova/backend/internal/origin/alien/handlers_property_test.go
     (что уже покрыто property-based)
   - projects/game-nova/backend/internal/origin/alien/handlers_integration_test.go
     (формат существующих golden)
   - docs/research/origin-vs-nova/divergence-log.md секция D-036
     (что нужно закрыть в Ф.7)

3) Прочитай выборочно:
   - projects/game-nova/backend/internal/battle/testdata/*.json
     (формат golden-файлов в проекте — образец для AlienAI)
   - projects/game-origin-php/src/game/AlienAI.class.php (главный
     legacy-источник для эталонов)
   - projects/game-origin-php/tools/dump-balance-formulas.php
     (паттерн PHP CLI для генерации эталонов из плана 64)

4) Добавь свою строку в docs/active-sessions.md:
   | <slot> | План 66 Ф.6+Ф.7 golden+финал | projects/game-nova/backend/internal/origin/alien/testdata/ projects/game-origin-php/tools/ docs/research/origin-vs-nova/divergence-log.md docs/plans/66-... | <дата-время> | test(alien): 50+ golden-итераций + финализация (план 66 Ф.6+Ф.7) |

ЧТО НУЖНО СДЕЛАТЬ:

### Ф.6 — Golden-итерации (50+)

1. **PHP-CLI для генерации эталонов**:
   - Создай `projects/game-origin-php/tools/dump-alien-ai.php`
     по образцу `tools/dump-balance-formulas.php` (план 64).
   - CLI вызывает PHP-функции AlienAI.class.php (require_once
     внутри пакета) с детерминированными входами и сериализует
     результат в JSON.
   - Покрой ≥50 кейсов, по группам:
     - **generateFleet** (15+ кейсов): разные target_power,
       available units, scale 1.0-5.0, four edge-cases (нулевой
       fleet, max scale, special units включены/выключены).
     - **findTarget / findCreditTarget** (10+ кейсов): разные
       критерии выбора (power range, distance, credit threshold).
     - **shuffleKeyValues** (10+ кейсов): фиксированный seed,
       разные tech-группы, проверка детерминизма.
     - **CalcGrabAmount / CalcGiftAmount** (10+ кейсов): разные
       credit и resource pools.
     - **HoldingExtension / FlightDuration / ChangeMissionDelay**
       (5+ кейсов): boundary checks.

2. **Go integration test**:
   - `projects/game-nova/backend/internal/origin/alien/golden_test.go`
     или расширь `handlers_integration_test.go`.
   - Загружает testdata/golden_alien_ai.json (массив кейсов с
     input+expected).
   - Для каждого кейса вызывает соответствующую Go-функцию с тем
     же seed и сравнивает с expected.
   - Толерантность: float-сравнение eps=1e-9 (детерминированный
     RNG должен давать точное совпадение, но запас на округление
     PHP↔Go).
   - Auto-skip если testdata-файл отсутствует (для CI без PHP).

3. **Покрытие ≥85%** изменённых строк origin/alien/ — проверь
   `go test -cover ./internal/origin/alien/...`. Если ниже —
   добив unit-тесты до 85%+.

4. **Property-based докрытие** (если ещё не покрыто Ф.1-Ф.4):
   - PickAttackTarget: при пустом списке кандидатов → ErrNoTarget.
   - PickCreditTarget: монотонность относительно credit.
   - GenerateFleet: total_power ≤ target_power*(1+epsilon).
   - ApplyShuffledTechWeakening: все weakened-техи ≤ 100%.

### Ф.7 — Финализация

5. **divergence-log**:
   - Открой docs/research/origin-vs-nova/divergence-log.md.
   - В записи D-036 (AlienAI расхождение) поставь статус
     **«Закрыто»** + дата + ссылка на коммиты Ф.1-Ф.6 плана 66.

6. **nova-ui-backlog** (если есть зависимости):
   - Если в backlog есть U-NNN про AlienAI UI — поставь статус
     «Backend ✅, UI отдельной задачей в новый origin-фронт (план 72)».

7. **Шапка плана 66**:
   - Все Ф.1-Ф.7 ✅.
   - Если Ф.5 пока не закрыта (отдельной сессией) — оставь как 🟡,
     но в Ф.6 не блокируется ею.

8. **project-creation.txt**:
   - Добавь итерацию 66 Ф.6+Ф.7 (golden, финализация).

══════════════════════════════════════════════════════════════════
ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА R0-R15 — см. блок rules-r0-r15.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/rules-r0-r15.md
══════════════════════════════════════════════════════════════════

══════════════════════════════════════════════════════════════════
GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ — см. блок git-isolation.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/git-isolation.md

Свои пути для CC_AGENT_PATHS:
- projects/game-nova/backend/internal/origin/alien/golden_test.go
- projects/game-nova/backend/internal/origin/alien/testdata/
- projects/game-origin-php/tools/dump-alien-ai.php
- docs/research/origin-vs-nova/divergence-log.md
- docs/research/origin-vs-nova/nova-ui-backlog.md (если меняешь)
- docs/plans/66-remaster-alien-ai-full-parity.md
- docs/project-creation.txt
- docs/active-sessions.md

ВНИМАНИЕ: НЕ трогай internal/origin/alien/buyout_handler.go и
прочее — это Ф.5 в параллельной сессии.
══════════════════════════════════════════════════════════════════

КОММИТЫ:

Один коммит: test(alien): 50+ golden-итераций + финализация (план 66 Ф.6+Ф.7)

Trailer: Generated-with: Claude Code

ВСЕГДА:
git commit -m "..." -- $CC_AGENT_PATHS

ЧЕГО НЕ ДЕЛАТЬ:

- НЕ менять production-код в internal/origin/alien/ (Ф.6 — только
  golden+тесты; если найдёшь баг — фикс отдельным коммитом, в
  Ф.6 commit только golden).
- НЕ трогать buyout_handler.go и Ф.5 файлы — параллельная сессия.
- НЕ упрощать golden до < 50 кейсов «потому что хватит» (R15: без
  упрощений как для прода).
- НЕ забывать про -- в git commit.

УСПЕШНЫЙ ИСХОД:

- testdata/golden_alien_ai.json содержит ≥50 кейсов из 5+ групп.
- dump-alien-ai.php запускается и регенерирует эталоны.
- Go golden_test.go проходит на всех 50+ кейсах.
- Покрытие internal/origin/alien/ ≥85%.
- D-036 в divergence-log закрыт.
- Шапка плана 66: Ф.6 ✅, Ф.7 ✅. Если Ф.5 ✅ к моменту коммита —
  весь план 66 ЗАКРЫТ.
- Запись в docs/project-creation.txt.
- Удалена строка из docs/active-sessions.md.

Стартуй.
```
