# План 87: глубокий audit боевого движка `internal/battle/` (game-nova)

**Дата**: 2026-05-01
**Статус**: 🔵 Открыт
**Зависимости**: нет (read-only audit). Связан с планом A (`game-nova: начислять опыт`, открытый промпт другому агенту), но не блокирует его.

**Связанные документы**:
- [docs/balance/audit.md](../balance/audit.md) — реестр найденных дыр (целевой документ для итога этого плана).
- [docs/adr/ADR-0002...](../adr/) — порт java-движка → Go.
- [oxsar2-java/Assault.java](d:/Sources/oxsar2-java/assault/src/assault/Assault.java) — reference-реализация.
- [docs/balance/analysis.md](../balance/analysis.md) — разбор балансных формул.

---

## Контекст

Боевой движок `projects/game-nova/backend/internal/battle/` — порт
Java-`Assault.jar` (oxsar2-java) на Go (ADR-0002). Используется в 5
местах: `fleet/attack.go`, `fleet/acs_attack.go`, `fleet/expedition.go`
(в двух сценариях), `alien/alien.go`, `simulator/handler.go`.

В предыдущих сессиях обнаружено два конкретных расхождения с legacy
без формального audit'а:

1. **Старая закомменченная формула опыта** (`min(20, max(0.1, ratio))
   × rounds`, [engine.go:135-145](../../projects/game-nova/backend/internal/battle/engine.go#L135) — _до_
   правки агентом плана A; на 2026-05-01 формула, по визуальной проверке
   шапки, уже переписана на atan-based и получила helper
   `computeExperience` ([engine.go:151-199](../../projects/game-nova/backend/internal/battle/engine.go#L151))). Зафиксировать как
   уже-исправленный кейс.
2. **Опыт игроку не начисляется в БД** — `Report.AttackerExp`
   считается, но `users.e_points/be_points/battles` не инкрементятся
   (план A в работе у соседского агента; не касаемся).

Других известных дыр нет, но движок не проходил формальный audit;
отсутствие задокументированной проверки в `docs/balance/audit.md`
оставляет open-ended risk перед запуском в прод.

## Цель

Прочесать `internal/battle/` целиком (engine.go ≈1040 строк +
types.go ≈220 + simstats.go ≈90 + engine_test.go ≈770) + 5 call sites
и сверить с reference-Java. Каждая найденная проблема описывается:
- file:line citation;
- категория (см. ниже);
- severity (critical / high / medium / low / info);
- предлагаемый фикс (если очевиден).

Итог — дополнение в `docs/balance/audit.md` с разделом «Боевой
движок (план 87)».

## Не цель

- **Не правим** код. Все находки — в audit. Фиксы — отдельные планы
  по каждой находке (или один общий, если их немного).
- **Не сверяем** числа bit-в-bit с Java-движком (это требует cross-verification
  через golden/JAR — описано в §14.4 ТЗ как отдельная задача).
- **Не аудитим** баланс юнитов / стоимости / RF (это `docs/balance/analysis.md`
  и `battle-sim` CLI).
- Game-legacy-PHP / Assault.jar — не наша забота, оба используют тот
  же jar 1:1.
- `internal/battle/testdata/` — golden-снимки не пересчитываем (это
  работа cross-verification).

## Категории логических ошибок

| Cat | Что ищем |
|---|---|
| **A** | Числовые: переполнение int64, float64 vs decimal, округление, деление на 0, отрицательные значения |
| **B** | Дыры в правилах: эксплойты входа (Damaged>Quantity, ShellPercent<0/>100, нулевой attack/shell, IsAliens на обеих сторонах, дубль Side в attackers+defenders, PrimaryTarget на несуществующий unit) |
| **C** | Расхождения с reference-java: формулы, regen, rapidfire, moon chance, building destroy, trophy artefacts, aliens specifics, debris, haul, finish conditions, lostUnits/lostPoints |
| **D** | RNG/детерминизм: seed=0, shared state между раундами, java.util.Random parity (если декларировано) |
| **E** | Concurrency: чистота Calculate, map iteration determinism, package-level state |
| **F** | Multi-sim (NumSim>1): семантика, edge cases, утечка в реальный бой |
| **G** | Idempotency / event-loop: повторное исполнение события, атомарность с записью отчёта/опыта/потерь |
| **H** | UX/отчёт: полнота RoundsTrace, decideWinner edge cases, Report.Seed для воспроизводимости |

## Что делаем

### Ф.1. Чтение `internal/battle/`

Файлы (в порядке):
1. `engine.go` (1040 строк) — основа.
2. `types.go` — структуры I/O.
3. `simstats.go` — multi-run агрегация.
4. `engine_test.go` — что покрыто, что нет.

Делаем ground-truth диаграмму: какие функции, в каком порядке вызываются,
какие глобальные переменные / package-level state, какие map'ы.

### Ф.2. Чтение reference-java

Файлы в `d:\Sources\oxsar2-java\assault\src\assault\`:
- `Assault.java` (главный flow).
- `Units.java` (юниты, getPoints, regenerate, finishTurn).
- `Participant.java` (стороны, lostPoints, finishParticipant с опытом и потерями).
- `Party.java` (контейнер).

Сверяем порядок операций, формулы (опыта, обломков, шанса луны, шанса
разрушения здания), guard'ы (debugmode, isBattleSimulation,
useridReported).

### Ф.3. Чтение 5 call sites

- `internal/fleet/attack.go` — стандартная атака.
- `internal/fleet/acs_attack.go` — ACS.
- `internal/fleet/expedition.go` — экспедиции (2 вызова).
- `internal/alien/alien.go` — бой с пришельцами.
- `internal/simulator/handler.go` — HTTP-обёртка.

Проверяем: что делается с `Report` (потери, обломки, добыча, опыт),
есть ли idempotency (одна транзакция, один SAVEPOINT), как обрабатывается
err.

### Ф.4. Систематическая проверка по категориям

Для каждой категории A..H — конкретные проверки (см. шапку). Все
наблюдения в живой заметке.

### Ф.5. Синтез в `docs/balance/audit.md`

Раздел «Боевой движок (план 87, 2026-05-01)»:
- сводка severity-распределения (X critical, Y high, Z medium…);
- по каждой находке — 4 поля (file:line, описание, severity, категория);
- предложения по дальнейшим фиксам (отдельные планы или один общий).

### Ф.6. Финализация

1. Шапка плана 87 → ✅ Закрыт.
2. Запись в `docs/project-creation.txt` — итерация 87.
3. Коммит: `docs(audit-87): глубокий audit боевого движка game-nova`.

## Smoke

Read-only, smoke не нужен. Если в ходе audit найдётся явный **репро**
бага — записать сценарий в audit, **не фиксить** в этой итерации.

## Риски и рассуждения

- **Риск scope creep**: каждая находка может потребовать раскопок в
  legacy-PHP/oxsar2 для понимания «как должно быть». Лимит: если для
  ясности нужно > 30 минут — фиксируем как **Open Question** в audit
  и идём дальше.
- **Риск ложно-позитивов**: подобные audit'ы дают 30-70% ложноположительных
  по моей предыдущей памяти (`feedback_audit_agent_verify.md`).
  Все находки **проверяю руками**, читая полную функцию + контекст
  использования; не принимаю на веру первое впечатление.
- **Совмещение с планом A**: соседская сессия делает фикс опыта в
  game-nova. Не касаюсь файлов, которые она трогает (`engine.go`
  formula, `users` migration), пока её работа не зафиксирована, чтобы
  не было merge-конфликтов. Возможен отзыв формулы опыта из аудита,
  если её уже починили.

## Стоп-список (не делаем в этом плане)

- Cross-verification против Assault.jar bit-в-bit (золотые снимки) —
  отдельная задача §14.4.
- Audit `pkg/rng` глубоко (свой план, если найдём pattern).
- Audit `internal/fleet/` целиком — только секции вокруг `Calculate`.
- Audit OpenAPI / handlers / DB-migrations — вне scope.
