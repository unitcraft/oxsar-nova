<?php
/**
 * dump-balance-formulas.php — оффлайн CLI для генерации golden-эталонов
 * балансовых формул origin для тестов internal/origin/economy/ (план 64
 * Ф.4.1).
 *
 * Подключается напрямую к docker-mysql-1 (origin БД), парсит DSL-строки
 * из na_construction (prod_metal, prod_hydrogen, ...) через PHP eval()
 * (один-в-один как production-Functions.inc.php parseChargeFormula и
 * Planet.class.php parseSpecialFormula), вычисляет значения для серии
 * (level, temp, tech) и пишет JSON в stdout.
 *
 * Запуск:
 *   docker exec docker-php-1 php /var/www/tools/dump-balance-formulas.php > golden.json
 *
 * Или через docker compose:
 *   cd projects/game-origin-php/docker && docker compose exec php \
 *     php /var/www/tools/dump-balance-formulas.php
 *
 * Output schema:
 *   {
 *     "metal_mine_prod": [
 *       {"level": 1, "tech": 0, "temp": 0, "value": 33},
 *       ...
 *     ],
 *     "hydrogen_lab_prod": [
 *       {"level": 5, "tech": 0, "temp": 0, "value": ...},
 *       ...
 *     ],
 *     ...
 *   }
 */

declare(strict_types=1);

// CLI-only.
if (PHP_SAPI !== 'cli') {
    fwrite(STDERR, "must run from CLI\n");
    exit(1);
}

// MySQL connection — read ENV from docker-php-1 (DB_HOST=mysql etc.).
$host = getenv('DB_HOST') ?: 'mysql';
$port = getenv('DB_PORT') ?: '3306';
$user = getenv('DB_USER') ?: 'oxsar_user';
$pass = getenv('DB_PWD') ?: 'oxsar_pass';
$db   = getenv('DB_NAME') ?: 'oxsar_db';

$pdo = new PDO("mysql:host=$host;port=$port;dbname=$db", $user, $pass, [
    PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
]);

/**
 * evalLegacyFormula — копия логики из oxsar2 PHP runtime.
 * См. d:/Sources/oxsar2/www/include/Functions.inc.php:41
 *     d:/Sources/oxsar2/www/include/Planet.class.php:592
 */
function evalLegacyFormula(string $formula, array $vars): float {
    $formula = trim($formula);
    if ($formula === '') {
        return 0;
    }
    // {level}, {basic}, {temp} → подстановка значений.
    foreach (['level', 'basic', 'temp'] as $k) {
        if (isset($vars[$k])) {
            $formula = str_replace('{' . $k . '}', (string)$vars[$k], $formula);
        }
    }
    // {tech=NN} → значение из vars['tech'][NN] или 0.
    $formula = preg_replace_callback('/\{tech=([0-9]+)\}/i', function ($m) use ($vars) {
        $id = (int)$m[1];
        return (string)($vars['tech'][$id] ?? 0);
    }, $formula);
    // {building=NN}
    $formula = preg_replace_callback('/\{building=([0-9]+)\}/i', function ($m) use ($vars) {
        $id = (int)$m[1];
        return (string)($vars['building'][$id] ?? 0);
    }, $formula);

    // Eval в isolated-сообразном контексте — допустимо, потому что DSL
    // — это origin-data из своей БД (не user input).
    $result = 0;
    try {
        $result = eval('return ' . $formula . ';');
    } catch (\Throwable $e) {
        fwrite(STDERR, "eval failed for '$formula': " . $e->getMessage() . "\n");
        return 0;
    }
    return (float)$result;
}

function getBuilding(PDO $pdo, string $name): array {
    $st = $pdo->prepare('SELECT * FROM na_construction WHERE name = ?');
    $st->execute([$name]);
    $row = $st->fetch(PDO::FETCH_ASSOC);
    if (!$row) {
        throw new RuntimeException("building $name not found");
    }
    return $row;
}

function dumpProdSeries(PDO $pdo, string $buildingName, string $field, array $points): array {
    $b = getBuilding($pdo, $buildingName);
    $formula = $b[$field] ?: '';
    if ($formula === '') {
        return [];
    }
    $out = [];
    foreach ($points as $p) {
        $vars = [
            'level' => $p['level'],
            'temp'  => $p['temp'] ?? 0,
            'basic' => $b['basic_' . preg_replace('/^prod_|^cons_|^charge_/', '', $field)] ?? 0,
            'tech'  => $p['tech'] ?? [],
        ];
        $value = evalLegacyFormula($formula, $vars);
        $out[] = [
            'level' => $p['level'],
            'tech'  => $p['tech'] ?? [],
            'temp'  => $p['temp'] ?? 0,
            'value' => $value,
        ];
    }
    return $out;
}

// Series — комбинации параметров для каждой формулы.
// Минимум: level={1,5,10,20,30}, temp={-150,-50,0,50,150}, tech=0..15.
$levels = [1, 5, 10, 20, 30];
$temps  = [-150, -100, -50, 0, 50, 100, 150];
$techs  = [0, 5, 12];

$metalPoints   = [];
$siliconPoints = [];
$solarPoints   = [];
foreach ($levels as $l) {
    foreach ($techs as $t) {
        $metalPoints[]   = ['level' => $l, 'tech' => [23 => $t]];
        $siliconPoints[] = ['level' => $l, 'tech' => [24 => $t]];
        $solarPoints[]   = ['level' => $l, 'tech' => [18 => $t]];
    }
}

$hydroPoints = [];
foreach ($levels as $l) {
    foreach ($temps as $temp) {
        foreach ($techs as $t) {
            $hydroPoints[] = ['level' => $l, 'temp' => $temp, 'tech' => [25 => $t]];
        }
    }
}

$out = [
    'metal_mine_prod_metal'      => dumpProdSeries($pdo, 'METALMINE',     'prod_metal',    $metalPoints),
    'silicon_lab_prod_silicon'   => dumpProdSeries($pdo, 'SILICON_LAB',   'prod_silicon',  $siliconPoints),
    'hydrogen_lab_prod_hydrogen' => dumpProdSeries($pdo, 'HYDROGEN_LAB',  'prod_hydrogen', $hydroPoints),
    'solar_plant_prod_energy'    => dumpProdSeries($pdo, 'SOLAR_PLANT',   'prod_energy',   $solarPoints),
];

echo json_encode($out, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES) . "\n";
