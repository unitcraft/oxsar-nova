<?php
/**
 * Endpoint для preview-картинок артефактов (план 37.5d.9).
 *
 * Замена legacy Yii-роуту `index.php/artefact2user_YII/{action}` (см.
 * artImageUrl() в src/core/Functions.php).
 *
 * Логика портирована из
 * d:/Sources/oxsar2/www/new_game/protected/controllers/Artefact2user_YIIController.php
 * actionImage_new() / renderImage().
 *
 * URL form:
 *   /artefact-image.php?cid=N&level=N&typeid=N        — preview по
 *     заданным cid (id постройки/исследования), level (уровень),
 *     typeid (тип артефакта, ARTEFACT_PACKED_BUILDING/RESEARCH).
 *
 * Кеш: $CACHE_DIR/{name}{level}{art_name}.jpg — отдаётся напрямую если
 * существует. Имена в lowercase, без пробелов/спецсимволов (см. legacy).
 */

// Только GET
if ($_SERVER['REQUEST_METHOD'] !== 'GET') {
    http_response_code(405);
    exit('Method not allowed');
}

$cid    = isset($_GET['cid']) ? (int)$_GET['cid'] : 0;
$level  = isset($_GET['level']) ? (int)$_GET['level'] : 0;
$typeid = isset($_GET['typeid']) ? (int)$_GET['typeid'] : 0;

if ($cid <= 0 || $level <= 0 || $typeid <= 0) {
    http_response_code(400);
    exit('Invalid params');
}

// Bootstrap (как public/game.php, без HTTP-роутинга/JWT)
$src = dirname(__FILE__, 2) . '/src/';
require_once($src . 'bd_connect_info.php');
define('INGAME', true);
define('YII_CONSOLE', true); // CLI-mode: не делать setRequest/setUser
$GLOBALS['RUN_YII'] = 0;
require_once($src . 'global.inc.php');
require_once(APP_ROOT_DIR . 'game/Functions.inc.php');
new Core();

// Имена постройки и типа артефакта из na_construction
$unit_name     = sqlSelectField("construction", "name", "", "buildingid=" . sqlVal($cid));
$art_type_name = sqlSelectField("construction", "name", "", "buildingid=" . sqlVal($typeid));

if (!$unit_name || !$art_type_name) {
    http_response_code(404);
    exit('Unknown buildingid');
}

$IMAGES_DIR = APP_ROOT_DIR . '../public/images';
$FONTS_DIR  = APP_ROOT_DIR . '../public/fonts';
$CACHE_DIR  = APP_ROOT_DIR . 'cache/artefacts';

if (!is_dir($CACHE_DIR)) {
    @mkdir($CACHE_DIR, 0775, true);
}

// Имя кеш-файла по конвенции legacy: lowercase(name + level + art_name).jpg
$cache_key = strtolower($unit_name . $level . $art_type_name);
$cache_key = preg_replace('/[^a-z0-9_-]+/', '_', $cache_key);
$cache_file = $CACHE_DIR . '/' . $cache_key . '.jpg';

// Hit cache
if (file_exists($cache_file)) {
    header('Accept-Ranges: bytes');
    header('Content-Type: image/jpeg');
    header('Cache-Control: public, max-age=86400');
    readfile($cache_file);
    exit;
}

// Miss — generate
if (!function_exists('imagecreatefromgif')) {
    http_response_code(500);
    exit('GD extension is required for artefact-image');
}

$art_bg_file = $IMAGES_DIR . '/buildings/std/' . strtolower($art_type_name) . '.gif';
if (!file_exists($art_bg_file)) {
    $art_bg_file = $IMAGES_DIR . '/buildings/empty/empty.gif';
}

$unit_overlay_file = $IMAGES_DIR . '/buildings/std/' . strtolower($unit_name) . '.gif';
if (!file_exists($unit_overlay_file)) {
    $unit_overlay_file = $IMAGES_DIR . '/buildings/empty/empty.gif';
}

if (!file_exists($art_bg_file) || !file_exists($unit_overlay_file)) {
    http_response_code(404);
    exit('Source images not found');
}

list($art_w, $art_h) = getimagesize($art_bg_file);
list($unit_w, $unit_h) = getimagesize($unit_overlay_file);

// Background = art type
$bg_image = imagecreatefromgif($art_bg_file);
$canvas = imagecreatetruecolor($art_w, $art_h);
imagecopy($canvas, $bg_image, 0, 0, 0, 0, $art_w, $art_h);
imagedestroy($bg_image);

// Overlay = unit (resized into bottom-right ~70%)
$ax = (int)($art_w * 0.3);
$ay = (int)($art_h * 0.3);
$overlay = imagecreatefromgif($unit_overlay_file);
imagecopyresampled(
    $canvas, $overlay,
    $ax, $ay, 0, 0,
    $art_w - $ax, $art_h - $ay,
    $unit_w, $unit_h
);
imagedestroy($overlay);

// Level number overlay (centered in bottom-right area)
$font_size = 30;
$font = $FONTS_DIR . '/FUTUR_14.TTF';
$text = (string)$level;

if (file_exists($font) && function_exists('imagettftext')) {
    $bbox = imagettfbbox($font_size, 0, $font, $text);
    $text_w = $bbox[4] - $bbox[0];
    $text_h = $bbox[1] - $bbox[5];

    $tx = (int)($ax + ($art_w - $ax) / 2 - $text_w / 2);
    $ty = (int)($ay + ($art_h - $ay) / 2 - $text_h / 2 + $text_h);

    $border = imagecolorallocate($canvas, 0xff, 0xff, 0xff);
    $fill = imagecolorallocate($canvas, 247, 191, 21);

    // Border (1px outline)
    for ($dx = -1; $dx <= 1; $dx++) {
        for ($dy = -1; $dy <= 1; $dy++) {
            if ($dx !== 0 && $dy !== 0) {
                imagettftext($canvas, $font_size, 0, $tx + $dx, $ty + $dy, $border, $font, $text);
            }
        }
    }
    imagettftext($canvas, $font_size, 0, $tx, $ty, $fill, $font, $text);
}

imageinterlace($canvas, 1);
imagejpeg($canvas, $cache_file, 80);
imagedestroy($canvas);

header('Accept-Ranges: bytes');
header('Content-Type: image/jpeg');
header('Cache-Control: public, max-age=86400');
readfile($cache_file);
