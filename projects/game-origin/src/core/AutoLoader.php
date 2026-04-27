<?php
/**
 * AutoLoader — clean-room rewrite (план 43 Ф.6). Заменяет одноимённый
 * файл фреймворка Recipe (GPL).
 *
 * Bootstrap-include статических файлов (Functions.php + Exception
 * interface + abstract Plugin) + spl_autoload_register для динамической
 * загрузки `Class.class.php` / `util/Class.util.class.php`.
 *
 * Ищет файлы в каталогах из AUTOLOAD_PATH_CORE (под core) и
 * AUTOLOAD_PATH_APP (под game). Резолвит `Class` → `Class.class.php`
 * либо `util/Class.util.class.php`.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

// Static includes — обязательные файлы, которые должны быть подключены
// до spl_autoload_register'а (либо потому что они нужны до первого
// autoload, либо потому что у них нестандартное имя файла).
$_static_includes = array(
    'Functions.php',
    'Exception.interface.php',
    'plugins/Plugin.abstract_class.php',
);
foreach($_static_includes as $_inc)
{
    $_path = RECIPE_ROOT_DIR.$_inc;
    if(!is_file($_path))
    {
        die('AutoLoader: missing required file '.$_inc);
    }
    require_once($_path);
}
unset($_inc, $_path, $_static_includes);

/**
 * Динамический autoloader. Сначала ищет в core (AUTOLOAD_PATH_CORE),
 * потом в app (AUTOLOAD_PATH_APP), потом fallback в util/.
 */
spl_autoload_register(function($class) {
    // getClassPath нормализует имя (определена в Functions.php).
    $normalized = function_exists('getClassPath') ? getClassPath($class) : $class;

    $candidates = array();

    if(defined('AUTOLOAD_PATH_CORE'))
    {
        foreach(explode(',', AUTOLOAD_PATH_CORE) as $dir)
        {
            $candidates[] = RECIPE_ROOT_DIR.$dir.$normalized.'.class.php';
        }
    }
    if(defined('AUTOLOAD_PATH_APP'))
    {
        foreach(explode(',', AUTOLOAD_PATH_APP) as $dir)
        {
            $candidates[] = APP_ROOT_DIR.$dir.$normalized.'.class.php';
        }
    }
    // Fallback: util/-каталог с суффиксом util.class.php.
    $candidates[] = RECIPE_ROOT_DIR.'util/'.$normalized.'.util.class.php';

    foreach($candidates as $file)
    {
        if(is_file($file))
        {
            require_once($file);
            return;
        }
    }
    // Не найдено — silent return; PHP сам бросит «Class not found»
    // если кто-то реально использует этот класс.
});
