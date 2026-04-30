<?php
/**
 * Публичный отчёт о бое.
 *
 * Порт actionAssaultReport() из oxsar2/www/new_game/protected/controllers/
 * MainController.php (Yii). Yii в legacy-PHP не используется, поэтому
 * переписано на текущий движок: Core/Template/Language/Query.
 *
 * URL form:
 *   /AssaultReport.php?id=N&key=XXXX                — обычный бой
 *   /AssaultReport.php?id=N&key2=XXXX               — бой с приватным ключом
 *   /AssaultReport.php?id=N&key=XXXX&simulation=1   — отчёт симулятора (sim_assault)
 *
 * HTML отчёта генерируется самим Assault.jar (oxsar2-java) и пишется
 * в na_assault.report / na_sim_assault.report. Здесь мы только разворачиваем
 * placeholder'ы ({lang}/{img}/{user}/{embedded}/{hide_text}/{battle_matrix})
 * и подставляем в шаблон assault_report.tpl.
 *
 * Упрощения относительно Yii-оригинала (можно вернуть отдельным планом):
 *   - нет send-to-friend формы (display_send=false всегда).
 *   - нет user_bg/user_table styles, нет музыкального плеера.
 *   - нет memcache (генерация быстрая, регенерация при каждом запросе).
 *
 * Авторизация: страница публичная (отчёт можно показать неавторизованному
 * по прямой ссылке с key), как и в legacy.
 */

// Bootstrap — копия паттерна public/artefact-image.php (без HTTP-роутинга/JWT).
$src = dirname(__FILE__, 2) . '/src/';
require_once($src . 'bd_connect_info.php');
define('INGAME', true);
define('YII_CONSOLE', true); // CLI-mode флаг: в legacy он отключал setRequest/setUser
$GLOBALS['RUN_YII'] = 0;
require_once($src . 'global.inc.php');
require_once(APP_ROOT_DIR . 'game/Functions.inc.php');
new Core();

Core::getLanguage()->load(array('info', 'AssaultReport'));

$assault_id    = max(0, (int)($_GET['id'] ?? 0));
$assault_key   = trim((string)($_GET['key'] ?? ''));
$assault_key2  = trim((string)($_GET['key2'] ?? ''));
$is_simulation = max(0, (int)($_GET['simulation'] ?? 0));

if ($assault_id <= 0 || ($assault_key === '' && $assault_key2 === '')) {
    http_response_code(400);
    exit('Invalid params');
}

// Достаём готовый HTML отчёта из БД (его пишет Assault.jar в конце расчёта).
$table = $is_simulation ? 'sim_assault' : 'assault';
$row = sqlSelectRow($table, array('report', 'planetid'), '',
    'assaultid = ' . sqlVal($assault_id) .
    ($assault_key2 !== ''
        ? ' AND `key2` = ' . sqlVal($assault_key2)
        : ' AND `key`  = ' . sqlVal($assault_key)));

if (!$row || $row['report'] === null || $row['report'] === '') {
    http_response_code(404);
    exit('Report not found');
}

$report = _ar_renderReport($row['report'], $assault_id, $assault_key2, $is_simulation);

if (empty($row['planetid'])) {
    // В Yii-оригинале при отсутствии planetid сбрасывали key2 (иначе
    // шаблон рендерил скрытие defender_info). Воспроизводим.
    $assault_key2 = '';
}

$tpl = Core::getTPL();
$tpl->assign('report',     $report);
$tpl->assign('charset',    Core::getLang()->getOpt('charset'));
$tpl->assign('id',         $assault_id);
$tpl->assign('key',        $assault_key);
$tpl->assign('key2',       $assault_key2);
$tpl->assign('sid',        defined('SID') ? SID : '');
$tpl->assign('formaction', socialUrl($_SERVER['PHP_SELF']));
// display_send / display_settings — выключены (см. упрощения в шапке).
$tpl->assign('display_send',     false);
$tpl->assign('display_settings', false);

$host = (defined('BASE_FULL_URL') ? BASE_FULL_URL : (HTTP_HOST ?? ''));
$tpl->assign('report_url',
    $host . 'AssaultReport.php?id=' . $assault_id .
    ($assault_key2 !== '' ? '&key2=' . $assault_key2 : '&key=' . $assault_key) .
    ($is_simulation ? '&simulation=1' : ''));

$tpl->display('assault_report', true);


// ---------------------------------------------------------------------------
// Helper'ы рендера, портированы из MainController.php строки 971-1080.
// Отличие от Yii-оригинала: модификатор регексов /siU → /si.
// Why: PCRE-флаг U инвертирует жадность по умолчанию, поэтому `(.*?)`
// под /U становится ЖАДНЫМ. На отчётах JAR'а с несколькими тегами
// {lang}…{/lang} в одной строке регекс с /siU съедал всё между первым
// открывающим и последним закрывающим тегом. /si оставляет `(.*?)`
// ленивым → каждый тег схлопывается отдельно.

function _ar_renderReport($report, $assault_id, $assault_key2, $is_simulation)
{
    $report = preg_replace_callback('/\{lang\}(.*?)\{\/lang\}/si',
        function ($m) use ($assault_id) {
            return _ar_getReportLangItem($m[1], $assault_id);
        }, $report);

    $report = preg_replace_callback('/\{img\}(.*?)\{\/img\}/si',
        function ($m) {
            return _ar_getReportImageItem($m[1]);
        }, $report);

    $report = preg_replace_callback(
        '/\{embedded4\[(.*?)\]\}\((.*?)\)\((.*?)\)\((.*?)\)\((.*?)\)\{\/embedded4\}/si',
        function ($m) {
            return sprintf(Core::getLanguage()->getItem($m[1]), $m[2], $m[3], $m[4], $m[5]);
        }, $report);

    $report = preg_replace_callback(
        '/\{embedded3\[(.*?)\]\}\((.*?)\)\((.*?)\)\((.*?)\)\{\/embedded3\}/si',
        function ($m) {
            return sprintf(Core::getLanguage()->getItem($m[1]), $m[2], $m[3], $m[4]);
        }, $report);

    $report = preg_replace_callback(
        '/\{embedded2\[(.*?)\]\}\((.*?)\)\((.*?)\)\{\/embedded2\}/si',
        function ($m) {
            return sprintf(Core::getLanguage()->getItem($m[1]), $m[2], $m[3]);
        }, $report);

    $report = preg_replace_callback(
        '/\{embedded\[(.*?)\]\}(.*?)\{\/embedded\}/si',
        function ($m) {
            return sprintf(Core::getLanguage()->getItem($m[1]), $m[2]);
        }, $report);

    $report = preg_replace_callback('/\{user\[(.*?)\]\}(.*?)\{\/user\}/si',
        function ($m) use ($is_simulation) {
            return _ar_getReportUserName($m[1], $m[2], $is_simulation);
        }, $report);

    $report = preg_replace_callback('/\{hide_text\}(.*?)\{\/hide_text\}/si',
        function ($m) use ($assault_key2) {
            return _ar_getHideText(str_replace("\\'", "'", $m[1]), $assault_key2);
        }, $report);

    $report = preg_replace_callback(
        '/\{defender_info_turn_1\}(.*?)\{\/defender_info_turn_1\}/si',
        function ($m) use ($assault_key2) {
            return _ar_getHideText(str_replace("\\'", "'", $m[1]), $assault_key2);
        }, $report);

    $report = preg_replace_callback(
        '/\{defender_result_turn_1\}(.*?)\{\/defender_result_turn_1\}/si',
        function ($m) use ($assault_key2) {
            return _ar_getHideText(str_replace("\\'", "'", $m[1]), $assault_key2);
        }, $report);

    $report = _ar_replaceBattleMatrix($report, $assault_key2);
    return $report;
}

function _ar_getReportLangItem($name, $assault_id)
{
    return Core::getLanguage()->getItem($name);
}

function _ar_getReportImageItem($name)
{
    $text = Core::getLanguage()->getItem($name);
    switch ($name) {
        case 'FIGHT_SHOTS_NUMBER':
        case 'FIGHT_SHOTS_POWER':
        case 'FIGHT_SHOTS_MISS':
        case 'FIGHT_SHIELD_ABSORB':
        case 'FIGHT_SHELL_DESTROYED':
        case 'FIGHT_UNITS_DESTROYED':
            return '<center>' . Image::getImage(strtolower($name) . '.gif', $text) . '</center>';
    }
    return $text;
}

function _ar_getHideText($str, $assault_key2)
{
    return $assault_key2 ? Core::getLanguage()->getItem('HIDE_BATTLE_TEXT') . '<br />' : $str;
}

function _ar_getReportUserName($userid, $username, $is_simulation)
{
    static $cachedUsers = array();
    if ($userid > 0 && strstr($username, '?')) {
        if (!isset($cachedUsers[$userid])) {
            $found = sqlSelectField($is_simulation ? 'sim_user' : 'user',
                'username', '', 'userid = ' . sqlVal($userid));
            if (!$found) {
                $found = 'Unknown#' . $userid;
            }
            $cachedUsers[$userid] = $found;
        }
        return $cachedUsers[$userid];
    }
    return $username;
}

function _ar_replaceBattleMatrix($report, $assault_key2)
{
    if (preg_match_all('#\{battle_matrix_turn_(\d+)\}#is', $report, $regs, PREG_OFFSET_CAPTURE)) {
        preg_match_all('#\{/battle_matrix_turn_\d+\}#is', $report, $regs2, PREG_OFFSET_CAPTURE);

        $offs = 0;
        foreach ($regs[0] as $key => $value) {
            $turn = $regs[1][$key][0];
            $sub_start = $value[1] + strlen($value[0]);
            $sub_len   = $regs2[0][$key][1] - $sub_start;

            $start = $value[1];
            $len   = $regs2[0][$key][1] + strlen($regs2[0][$key][0]) - $start;

            $new_sub = _ar_getBattleMatrixTurn(
                $turn, substr($report, $sub_start + $offs, $sub_len), $assault_key2);
            $new_sub_len = strlen($new_sub);

            $report = substr_replace($report, $new_sub, $start + $offs, $len);
            $offs += $new_sub_len - $len;
        }
    }
    return $report;
}

function _ar_getBattleMatrixTurn($turn, $str, $assault_key2)
{
    if ($assault_key2) {
        return '';
    }
    $open  = Core::getLanguage()->getItem('SHOW_BATTLE_MATRIX');
    $hide  = Core::getLanguage()->getItem('HIDE_BATTLE_MATRIX');
    return <<<HTML
<div class='bmat_open_panel_turn_{$turn}'><a href="#" onclick="open_bmat({$turn}); return false">{$open}</a></div>
<div class='bmat_panel_turn_{$turn}' style='display:none'>
    <a href="#" onclick="close_bmat({$turn}); return false">{$hide}</a>
    <br />{$str}
</div>
HTML;
}
