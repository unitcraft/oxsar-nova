<?php
/**
 * Functions.php — clean-room rewrite (план 43 Ф.6). Заменяет одноимённый
 * файл фреймворка Recipe (GPL).
 *
 * Глобальные хелпер-функции для работы с БД (sqlVal/sqlSelect/...),
 * редиректами, валидацией email, генерацией случайных строк и пр.
 * Всё это глобальное API legacy-кода, переписать на namespaces — большая
 * задача (затронет ~200 call-sites). Пока сохраняем глобальные функции
 * с теми же именами и сигнатурами.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('RECIPE_ROOT_DIR')) { die('Hacking attempt detected.'); }

/* ============================================================
 * Управление потоком выполнения
 * ============================================================ */

/** echo + exit. Принимает строку или массив (через print_r). */
function terminate($string)
{
    if(is_array($string))
    {
        print_r($string);
    }
    else
    {
        echo (string)$string;
    }
    exit;
}

/**
 * Перенаправление на login-страницу с error-параметром, либо рендер
 * login-template если LOGIN_REQUIRED == false.
 */
function forwardToLogin($errorid)
{
    if(defined('LOGIN_REQUIRED') && LOGIN_REQUIRED)
    {
        $login = defined('LOGIN_URL') ? LOGIN_URL : '/login.php';
        if(strpos($login, '?') === false)
        {
            $login .= '?error='.urlencode((string)$errorid);
        }
        doHeaderRedirection($login, false);
        return;
    }
    if(class_exists('Core'))
    {
        $lang = Core::getLanguage();
        if($lang) { $lang->load('account'); }
    }
    if(class_exists('Logger'))
    {
        Logger::addMessage($errorid);
    }
    if(class_exists('Core'))
    {
        $tpl = Core::getTPL();
        if($tpl) { $tpl->display('login'); }
    }
}

/* ============================================================
 * Социальные сети — stubs (план 37: соцсети отключены)
 * ============================================================ */

if(!function_exists('socialUrl'))
{
    function socialUrl($url) { return $url; }
}

if(!function_exists('artImageUrl'))
{
    function artImageUrl($action, $suffix, $real_url = true)
    {
        return ($real_url ? RELATIVE_URL : '').'artefact-image.php?';
    }
}

/* ============================================================
 * Header redirect
 * ============================================================ */

function doHeaderRedirection($url, $appendSession = true)
{
    $url = (string)$url;
    if(class_exists('Link') && Link::isExternal($url))
    {
        $path = $url;
    }
    elseif(strpos($url, 'http://') === 0 || strpos($url, 'https://') === 0)
    {
        $path = $url;
    }
    else
    {
        if($appendSession && defined('URL_SESSION') && URL_SESSION
           && strpos($url, 'sid=') === false && defined('SID'))
        {
            $sep = strpos($url, '?') === false ? '?' : '&';
            $url .= $sep.'sid='.SID;
        }

        if(defined('FORCE_REWRITE') && FORCE_REWRITE && class_exists('Link'))
        {
            $path = Link::normalizeURL($url);
        }
        else
        {
            $host = defined('HTTP_HOST') ? HTTP_HOST : '';
            $dir = defined('REQUEST_DIR') ? REQUEST_DIR : '';
            $path = $host.$dir.$url;
        }
    }
    header('Location: '.$path);
    exit;
}

/* ============================================================
 * Валидация и генерация
 * ============================================================ */

function isMail($mail)
{
    return is_string($mail) && filter_var($mail, FILTER_VALIDATE_EMAIL) !== false;
}

function randString($length)
{
    $length = max(0, (int)$length);
    if($length === 0) { return ''; }
    $alphabet = 'qwertzupasdfghkyxcvbnm23456789QWERTZUPLKJHGFDSAYXCVBNM';
    $max = strlen($alphabet) - 1;
    $out = '';
    for($i = 0; $i < $length; $i++)
    {
        $out .= $alphabet[mt_rand(0, $max)];
    }
    return $out;
}

function parseUrl($url)
{
    return parse_url((string)$url) ?: array();
}

/**
 * Преобразование "path/to/file" → "Path/To/File" для autoloader-резолва
 * имён классов с namespace-вложенностью.
 */
function getClassPath($path, $s = '/')
{
    $path = (string)$path;
    if($path === '') { return ''; }
    if(substr($path, -strlen($s)) === $s)
    {
        $path = substr($path, 0, -strlen($s));
    }
    if(strpos($path, $s) === 0)
    {
        $path = substr($path, strlen($s));
    }
    // 'a/b/c' → 'A/B/C'
    return str_replace(' ', $s, ucwords(str_replace($s, ' ', $path)));
}

/* ============================================================
 * SQL-helpers — обёртки над Core::getDB() / Core::getQuery()
 * ============================================================ */

function sqlVal($data)
{
    return $data === null ? 'NULL' : Core::getDatabase()->quote_db_value((string)$data);
}

/**
 * Принимает массив (или несколько аргументов) и возвращает CSV из
 * экранированных значений (для `IN (…)` clauses).
 */
function sqlArray()
{
    $args = func_get_args();
    if(count($args) === 1 && is_array($args[0]))
    {
        $args = $args[0];
    }
    if(count($args) === 0) { return 'NULL'; }
    $parts = array();
    foreach($args as $value)
    {
        $parts[] = sqlVal($value);
    }
    return implode(',', $parts);
}

function sqlSelect($table, $select, $join = '', $where = '', $order = '', $limit = '', $groupby = '', $other = '')
{
    return Core::getQuery()->select($table, $select, $join, $where, $order, $limit, $groupby, $other);
}

function sqlSelectRow($table, $select, $join = '', $where = '', $order = '', $limit = '', $groupby = '', $other = '')
{
    return Core::getQuery()->selectRow($table, $select, $join, $where, $order, $limit, $groupby, $other);
}

function sqlSelectField($table, $select, $join = '', $where = '', $order = '', $limit = '', $groupby = '', $other = '')
{
    return Core::getQuery()->selectField($table, $select, $join, $where, $order, $limit, $groupby, $other);
}

function sqlInsert($table_name, $values)
{
    Core::getQuery()->insert(
        $table_name,
        array_keys($values),
        array_values($values)
    );
    return Core::getDB()->insert_id();
}

function sqlUpdate($table_name, $values, $where)
{
    return Core::getQuery()->update(
        $table_name,
        array_keys($values),
        array_values($values),
        $where
    );
}

function sqlDelete($table_name, $where)
{
    Core::getQuery()->delete($table_name, $where);
}

function sqlQuery($sql)
{
    return Core::getDB()->query($sql);
}

function sqlQueryRow($sql)
{
    return Core::getDB()->queryRow($sql);
}

function sqlQueryField($sql)
{
    return Core::getDB()->queryField($sql);
}

function sqlFetch($result)
{
    return Core::getDB()->fetch($result);
}

function sqlEnd($result)
{
    return Core::getDB()->free_result($result);
}

/* ============================================================
 * Среда и mobile-detection
 * ============================================================ */

function env($key)
{
    if(isset($_SERVER[$key])) { return $_SERVER[$key]; }
    if(isset($_ENV[$key])) { return $_ENV[$key]; }
    $v = getenv($key);
    return $v !== false ? $v : null;
}

if(!defined('REQUEST_MOBILE_UA'))
{
    define('REQUEST_MOBILE_UA', '(iPhone|iPad|MIDP|AvantGo|BlackBerry|J2ME|Opera Mini|DoCoMo|NetFront|Nokia|PalmOS|PalmSource|portalmmm|Plucker|ReqwirelessWeb|SonyEricsson|Symbian|UP\.Browser|Windows CE|Xiino|Android|Mobile)');
}

function is_mobile_request()
{
    $ua = env('HTTP_USER_AGENT');
    return is_string($ua) && preg_match('/'.REQUEST_MOBILE_UA.'/i', $ua) === 1;
}

if(!defined('IS_MOBILE_REQUEST'))
{
    define('IS_MOBILE_REQUEST', is_mobile_request());
}
