<?php
/**
 * Game entry point — Yii-free bootstrap.
 * Все запросы к игре проходят через этот файл.
 */

// src/ лежит на уровень выше public/
$src = dirname(__FILE__, 2) . '/src/';

require_once($src . 'bd_connect_info.php');

define("INGAME", true);
$GLOBALS["RUN_YII"] = 0;

require_once($src . 'global.inc.php');
require_once(APP_ROOT_DIR . 'game/Functions.inc.php');

new Core();
new NS();
exit;
