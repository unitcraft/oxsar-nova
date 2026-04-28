// Package alien — порт AlienAI движка из oxsar2-classic
// (projects/game-legacy-php/src/game/AlienAI.class.php, ~1127 строк PHP)
// на Go, в рамках плана 66 (см. docs/plans/66-remaster-alien-ai-full-parity.md).
//
// R0-исключение: пакет применяется во ВСЕХ вселенных (uni01/uni02 + origin),
// не только origin. Несмотря на расположение под internal/origin/alien/,
// origin здесь — источник кода, не таргет вселенной. Решение пользователя
// 2026-04-28 (см. roadmap-report «Часть I.5»).
//
// Состав:
//   - config.go        — параметры AlienAI (тайминги, проценты, лимиты).
//                       Источник истины — origin consts.php:752-770; в Ф.3
//                       будут поднимаемы из configs/balance/*.yaml.
//   - state.go         — типизированные структуры состояния (Mission,
//                       FleetSpec, HoldingState).
//   - fleet_generator.go — алгоритм generateFleet (target_power +
//                       итеративный подбор; PHP:405-622).
//   - target.go        — findTarget / findCreditTarget критерии цели.
//   - shuffle.go       — shuffleKeyValues (случайное ослабление техник).
//   - helpers.go       — вспомогательные функции (loadPlanetShips и др.).
//   - repo.go          — pgx I/O (отдельно, чтобы помочь юнит-тестам).
//
// Параметры AlienAI идут через struct Config, в тестах подменяемый;
// runtime-загрузка из configs/balance/*.yaml — Ф.3 плана 66 (после
// плана 65 / R8 Prometheus + R9 Idempotency).
//
// Ф.1+Ф.2 (этот коммит): state machine + helpers, без Kind handlers
// и event-loop проводки (см. план 66 структуру коммитов).
package alien
