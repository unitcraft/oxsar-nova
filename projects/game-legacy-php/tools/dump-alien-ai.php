<?php
/**
 * dump-alien-ai.php — оффлайн CLI для генерации golden-эталонов
 * AlienAI-формул origin для тестов internal/origin/alien/ (план 66 Ф.6).
 *
 * Что делает:
 *   - читает константы AlienAI (consts.php / AlienAI.class.php) — ALIEN_GRAB_*,
 *     ALIEN_HALTING_*, ALIEN_CHANGE_MISSION_*, ALIEN_MAX_GIFT_CREDIT и т.п.;
 *   - генерирует 50+ golden-кейсов: для каждого вычисляет ожидаемый
 *     инвариант формулы AlienAI (диапазон или точное значение) — без
 *     обращения к MySQL и без вызова mt_rand/shuffle напрямую (RNG
 *     несовместим между PHP Mersenne Twister и Go xorshift64*, поэтому
 *     бит-в-бит совпадение значений невозможно — golden работает на
 *     уровне ИНВАРИАНТОВ ФОРМУЛЫ);
 *   - пишет JSON в stdout, формат — массив кейсов с {id, group, fn,
 *     input, expected_min, expected_max, expected_value? (для exact)}.
 *
 * Зачем так:
 *   - Go-сторона (golden_test.go) для каждого кейса:
 *     * вызывает соответствующий Go-helper с детерминированным rng.New(seed),
 *     * проверяет что результат лежит в [expected_min, expected_max] (или
 *       равен expected_value для exact-кейсов);
 *   - Это даёт паритет на уровне формул, а не RNG-байтов. Семантически
 *     эквивалентно golden-эталонам battle/testdata, где Go RNG совпадает
 *     с Java через JavaRandom-адаптер; здесь же mt_rand-адаптер не
 *     введён (R8/Ф.6 ТЗ упоминает это как future work — см. shuffle.go:115),
 *     поэтому PHP служит источником формулы, а не байтов.
 *
 * Запуск:
 *   php projects/game-origin-php/tools/dump-alien-ai.php > \
 *     projects/game-nova/backend/internal/origin/alien/testdata/golden_alien_ai.json
 *
 * Output schema:
 *   [
 *     {
 *       "id": "grab_001",
 *       "group": "CalcGrabAmount",
 *       "fn": "CalcGrabAmount",
 *       "input": {"user_credit": 1000000, "seed": 12345},
 *       "expected_min": 800,
 *       "expected_max": 1000,
 *       "comment": "0.0008..0.001 of 1M"
 *     },
 *     ...
 *   ]
 */

declare(strict_types=1);

if (PHP_SAPI !== 'cli') {
    fwrite(STDERR, "must run from CLI\n");
    exit(1);
}

// ------------------------------------------------------------------
// Константы AlienAI — 1-в-1 из oxsar2-classic consts.php:752-770.
// Хардкод чтобы не require_once тяжёлый legacy-bootstrap.
// ------------------------------------------------------------------
const ALIEN_GRAB_MIN_CREDIT          = 100000;
const ALIEN_GRAB_CREDIT_MIN_PERCENT  = 0.08;
const ALIEN_GRAB_CREDIT_MAX_PERCENT  = 0.10;
const ALIEN_MAX_GIFT_CREDIT          = 500;
const ALIEN_HALTING_MIN_TIME         = 12 * 3600;       // 12h
const ALIEN_HALTING_MAX_TIME         = 24 * 3600;       // 24h
const ALIEN_HALTING_MAX_REAL_TIME    = 15 * 86400;      // 15d
const ALIEN_CHANGE_MISSION_MIN_TIME  = 8  * 3600;
const ALIEN_CHANGE_MISSION_MAX_TIME  = 10 * 3600;
const ALIEN_FLY_MIN_TIME             = 15 * 3600;
const ALIEN_FLY_MAX_TIME             = 24 * 3600;
const ALIEN_HOLDING_PAY_SECONDS_PER_CREDIT = 144;       // 2*3600/50

// ------------------------------------------------------------------
// Helpers: для каждой формулы AlienAI вычисляем диапазон [min, max].
// ------------------------------------------------------------------

/** CalcGrabAmount: round(credit*0.01*randFloat(0.08,0.10), 2). */
function grabRange(int $credit): array {
    if ($credit <= ALIEN_GRAB_MIN_CREDIT) {
        return [0, 0, "below threshold"];
    }
    $lo = (int) floor($credit * 0.01 * ALIEN_GRAB_CREDIT_MIN_PERCENT);
    $hi = (int) ceil ($credit * 0.01 * ALIEN_GRAB_CREDIT_MAX_PERCENT);
    return [$lo, $hi, sprintf("0.0008..0.001 of %d", $credit)];
}

/** CalcGiftAmount: min(MAX_GIFT*randFloat(0.98,1.02), credit*0.01*randFloat(5,10)). */
function giftRange(int $credit): array {
    $cap = (int) ceil(ALIEN_MAX_GIFT_CREDIT * 1.02);
    $base = (int) floor($credit * 0.01 * 5);
    if ($base > $cap) $base = $cap;
    if ($base < 0) $base = 0;
    // Минимум — 0 (если credit*0.01*pct ≤ 0 целочисленно).
    return [0, $cap, sprintf("≤%d (cap MaxGift*1.02), credit=%d", $cap, $credit)];
}

/** HoldingExtension: holds += paid * 144 sec, capped at start+15d. */
function holdingExtensionExact(int $startTs, int $holdsTs, int $paid): array {
    if ($paid <= 0) {
        return [$holdsTs, $holdsTs, "no extension"];
    }
    $add = (int) ($paid * ALIEN_HOLDING_PAY_SECONDS_PER_CREDIT);
    $cap = $startTs + ALIEN_HALTING_MAX_REAL_TIME;
    $out = $holdsTs + $add;
    if ($out > $cap) $out = $cap;
    return [$out, $out, sprintf("+%ds, capped at start+15d", $add)];
}

/** PowerScaleAfterControlTimes: 1 + control_times*1.5 (exact). */
function powerScaleExact(int $ct): array {
    $v = 1.0 + $ct * 1.5;
    return [$v, $v, sprintf("1 + %d*1.5", $ct)];
}

/** HoldingDuration: randRoundRange(ALIEN_HALTING_MIN_TIME, ALIEN_HALTING_MAX_TIME). */
function holdingDurationRange(): array {
    return [ALIEN_HALTING_MIN_TIME, ALIEN_HALTING_MAX_TIME, "12h..24h"];
}

/** FlightDuration: randRoundRange(ALIEN_FLY_MIN_TIME, ALIEN_FLY_MAX_TIME). */
function flightDurationRange(): array {
    return [ALIEN_FLY_MIN_TIME, ALIEN_FLY_MAX_TIME, "15h..24h"];
}

/** ChangeMissionDelay: 60% — 8h..10h capped flight-10s; 40% — flight-30..flight-10. */
function changeMissionDelayRange(int $flightSec): array {
    if ($flightSec <= 30) {
        $v = $flightSec - 10;
        return [$v, $v, "flight≤30s special"];
    }
    $branchALo = ALIEN_CHANGE_MISSION_MIN_TIME;
    $branchAHi = min(ALIEN_CHANGE_MISSION_MAX_TIME, $flightSec - 10);
    $branchBLo = $flightSec - 30;
    $branchBHi = $flightSec - 10;
    $lo = min($branchALo, $branchBLo);
    $hi = max($branchAHi, $branchBHi);
    if ($lo < 0) $lo = 0;
    return [$lo, $hi, "60%/40% branch span"];
}

/** HoldingAISubphaseDuration: clamp(min(12h, 30s*ct) ... max(24h, 60s*ct)). */
function holdingAISubphaseRange(int $ct): array {
    $hi = max(ALIEN_HALTING_MAX_TIME, 60 * $ct);
    $lo = min(ALIEN_HALTING_MIN_TIME, 30 * $ct);
    if ($lo > $hi) $lo = $hi;
    return [$lo, $hi, sprintf("clamp ct=%d", $ct)];
}

/** WeakenedTechLevel: max(0, rand(floor(level*0.7), level<3 ? level : level+1)). */
function weakenedTechRange(int $level): array {
    if ($level <= 0) return [0, 0, "level=0"];
    $lo = (int) floor($level * 0.7);
    $hi = $level < 3 ? $level : $level + 1;
    if ($lo < 0) $lo = 0;
    if ($lo > $hi) $lo = $hi;
    return [$lo, $hi, sprintf("level=%d", $level)];
}

// ------------------------------------------------------------------
// Сборка golden-кейсов.
// ------------------------------------------------------------------

$cases = [];
$id = 0;

function addCase(array &$cases, int &$id, string $group, string $fn, array $input,
                 array $rangeOrExact): void {
    [$lo, $hi, $comment] = $rangeOrExact;
    $cases[] = [
        "id"           => sprintf("%s_%03d", strtolower($group), ++$id),
        "group"        => $group,
        "fn"           => $fn,
        "input"        => $input,
        "expected_min" => $lo,
        "expected_max" => $hi,
        "comment"      => $comment,
    ];
}

// === group: CalcGrabAmount (12 кейсов).
$grabCredits = [
    50_000, 99_999, 100_000, 100_001, 250_000, 500_000,
    1_000_000, 5_000_000, 10_000_000, 100_000_000, 1_000_000_000,
    1_000_000_000_000,
];
$gid = 0;
foreach ($grabCredits as $c) {
    addCase($cases, $gid, "CalcGrabAmount", "CalcGrabAmount",
        ["user_credit" => $c, "seed" => 1000 + $gid + 1],
        grabRange($c));
}

// === group: CalcGiftAmount (10 кейсов).
$giftCredits = [
    0, 1_000, 10_000, 50_000, 100_000, 500_000, 1_000_000,
    10_000_000, 100_000_000, 1_000_000_000,
];
$gid = 0;
foreach ($giftCredits as $c) {
    addCase($cases, $gid, "CalcGiftAmount", "CalcGiftAmount",
        ["user_credit" => $c, "seed" => 2000 + $gid + 1],
        giftRange($c));
}

// === group: HoldingExtension (10 кейсов, exact-формула).
$startTs = 1735689600; // 2025-01-01 00:00 UTC, фиксированный
$holdsTs = $startTs + ALIEN_HALTING_MIN_TIME;
$extensions = [
    ["paid" => 0,   "holds_offset" => ALIEN_HALTING_MIN_TIME], // no-op
    ["paid" => 1,   "holds_offset" => ALIEN_HALTING_MIN_TIME],
    ["paid" => 10,  "holds_offset" => ALIEN_HALTING_MIN_TIME],
    ["paid" => 50,  "holds_offset" => ALIEN_HALTING_MIN_TIME],
    ["paid" => 100, "holds_offset" => ALIEN_HALTING_MIN_TIME],
    ["paid" => 500, "holds_offset" => ALIEN_HALTING_MIN_TIME],
    ["paid" => 1000,"holds_offset" => ALIEN_HALTING_MIN_TIME],
    ["paid" => 10000,"holds_offset" => ALIEN_HALTING_MIN_TIME], // overflow → cap
    ["paid" => 5,   "holds_offset" => ALIEN_HALTING_MAX_REAL_TIME - 60], // close to cap
    ["paid" => 200, "holds_offset" => ALIEN_HALTING_MAX_TIME],
];
$gid = 0;
foreach ($extensions as $e) {
    $hOff = $e["holds_offset"];
    $p = $e["paid"];
    addCase($cases, $gid, "HoldingExtension", "HoldingExtension",
        ["start_ts" => $startTs, "holds_ts" => $startTs + $hOff,
         "paid_hard" => $p],
        holdingExtensionExact($startTs, $startTs + $hOff, $p));
}

// === group: PowerScaleAfterControlTimes (8 кейсов, exact).
$gid = 0;
foreach ([0, 1, 2, 3, 5, 10, 50, 100] as $ct) {
    addCase($cases, $gid, "PowerScaleAfterControlTimes", "PowerScaleAfterControlTimes",
        ["control_times" => $ct],
        powerScaleExact($ct));
}

// === group: HoldingDuration (5 кейсов, range).
$gid = 0;
foreach ([1, 2, 3, 4, 5] as $i) {
    addCase($cases, $gid, "HoldingDuration", "HoldingDuration",
        ["seed" => 3000 + $i],
        holdingDurationRange());
}

// === group: FlightDuration (5 кейсов, range).
$gid = 0;
foreach ([1, 2, 3, 4, 5] as $i) {
    addCase($cases, $gid, "FlightDuration", "FlightDuration",
        ["seed" => 4000 + $i],
        flightDurationRange());
}

// === group: ChangeMissionDelay (6 кейсов, range).
$flights = [20, 60, 600, 3600, 12 * 3600, 24 * 3600];
$gid = 0;
foreach ($flights as $f) {
    addCase($cases, $gid, "ChangeMissionDelay", "ChangeMissionDelay",
        ["flight_seconds" => $f, "seed" => 5000 + $gid + 1],
        changeMissionDelayRange($f));
}

// === group: HoldingAISubphaseDuration (6 кейсов, range).
$gid = 0;
foreach ([0, 1, 5, 30, 100, 500] as $ct) {
    addCase($cases, $gid, "HoldingAISubphaseDuration", "HoldingAISubphaseDuration",
        ["control_times" => $ct, "seed" => 6000 + $gid + 1],
        holdingAISubphaseRange($ct));
}

// === group: WeakenedTechLevel (10 кейсов, range).
$gid = 0;
foreach ([0, 1, 2, 3, 5, 10, 15, 20, 30, 50] as $level) {
    addCase($cases, $gid, "WeakenedTechLevel", "WeakenedTechLevel",
        ["level" => $level, "seed" => 7000 + $gid + 1],
        weakenedTechRange($level));
}

// ------------------------------------------------------------------
// Output.
// ------------------------------------------------------------------

$json = json_encode($cases, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES);
if ($json === false) {
    fwrite(STDERR, "json_encode failed: " . json_last_error_msg() . "\n");
    exit(2);
}
fwrite(STDERR, sprintf("emitted %d cases\n", count($cases)));
echo $json;
echo "\n";
