# ADR-0002: Порт боевого движка из Java на Go

- Status: Accepted
- Date: 2026-04-21

## Context

В oxsar2 сосуществуют ДВА независимых боевых расчёта:

1. **PHP** — `d:\Sources\oxsar2\www\game\Assault.class.php` (974 LOC) +
   `Participant.class.php` (636 LOC). Это продакшн-путь: реальные
   атаки игроков обрабатывались здесь через `EventHandler`.
2. **Java** — `d:\Sources\oxsar2-java` (4500 LOC). Отдельный
   симулятор-калькулятор, вызываемый из UI симулятора
   (`ext/page/ExtSimulator.class.php`, §5.21) через shell-exec JAR.

До версии ТЗ 1.2 мы ошибочно считали Java единственным эталоном.
PHP-класс `BattleComponent.php` в `new_game/` — обрубок-прототип с
TODO и к порту не привлекается.

## Decision

- Порт идёт в Go-пакет `internal/battle`.
- `Database.java` и PHP-I/O не портируются: движок — чистая функция
  `Calculate(input BattleInput) BattleReport`, детерминированная по
  `seed uint64`.
- Эталонами паритета являются ОБА источника:
  - **PHP (приоритет)** — `game/Assault.class.php` + `Participant.
    class.php`, как источник реального продакшн-поведения.
  - **Java** — `oxsar2-java`, как источник для симулятора.
- При расхождении между PHP и Java приоритет за PHP; решение
  фиксируется в дополнительном ADR.
- Верификация: параллельный прогон PHP (через PHP-CLI + mock DB)
  и Java (через jar) на одних и тех же входах, затем сверка с
  Go-движком. Допустимое расхождение по юнитам — 0 в 95% кейсов,
  отдельные edge case с RNG документируются.

## Consequences

- Go-RNG имитирует `java.util.Random` (LCG с константами Oracle
  JDK) для паритета с Java-jar. Для паритета с PHP — дополнительно
  адаптер `mt_rand()`-совместимого RNG.
- Rapidfire / masking / ballistics получаем готовые из обоих
  источников.
- TS-симулятор на фронте использует тот же seed-контракт.
- Trace каждого раунда пишется в отчёт — это упростит сравнение
  PHP vs Java vs Go в юнит-тестах.
