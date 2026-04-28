<?php
/**
 * init.php — clean-room rewrite (план 43 Ф.6). Заменяет одноимённый файл
 * фреймворка Recipe (GPL).
 *
 * Bootstrap-файл подключается из global.inc.php, регистрирует обработчик
 * ошибок (PHP-warning → GenericException) и подключает AutoLoader.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

// Уровень репортинга задан в global.inc.php (ERROR_REPORTING_TYPE).
if(defined('ERROR_REPORTING_TYPE'))
{
    error_reporting(ERROR_REPORTING_TYPE);
}

/**
 * Превращает PHP-warning в GenericException — чтобы перехватывать
 * единообразно через try/catch выше по стеку. Активируется только
 * если RUN_YII != 1 (legacy-флаг для Yii-режима).
 */
function errorHandler($errNo, $message, $file, $line)
{
    throw new GenericException(sprintf('%s in %s:%d', $message, $file, $line), $errNo);
}

function exceptionHandler($exception)
{
    // No-op — uncaught исключения логируются стандартным PHP-механизмом.
    // Тонкая настройка (показ error-template, отправка в Sentry и т.п.)
    // вынесена в более высокий слой (game.php → Core → Logger).
}

if(!isset($GLOBALS['RUN_YII']) || $GLOBALS['RUN_YII'] != 1)
{
    set_error_handler('errorHandler', defined('ERROR_REPORTING_TYPE') ? ERROR_REPORTING_TYPE : E_ERROR);
    set_exception_handler('exceptionHandler');
}

// Debuger.php — legacy debug-helper, оставлен как есть до Ф.8.
if(defined('RECIPE_ROOT_DIR') && is_file(RECIPE_ROOT_DIR.'Debuger.php'))
{
    require_once(RECIPE_ROOT_DIR.'Debuger.php');
}

// AutoLoader подключается в самом конце — он spl_autoload_register'ит
// callback и использует выше определённые helper-функции.
if(defined('RECIPE_ROOT_DIR') && is_file(RECIPE_ROOT_DIR.'AutoLoader.php'))
{
    require_once(RECIPE_ROOT_DIR.'AutoLoader.php');
}
