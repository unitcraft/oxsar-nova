<?php
/**
 * Request — clean-room rewrite (план 43 Ф.3). Заменяет одноимённый класс
 * фреймворка Recipe (GPL).
 *
 * Тонкая обёртка над $_GET / $_POST / $_COOKIE / $_SERVER в стиле singleton,
 * чтобы вызовы шли через Core::getRequest() (легче подменить в тестах).
 *
 * Сохранены сигнатуры используемых методов:
 *   - getInstance()                           — singleton.
 *   - get($source, $key, $default = '')       — обобщённый доступ; source
 *                                                ∈ {'get','post','cookie','server'}.
 *   - getGET($key, $default = '')             — alias get('get', $key).
 *   - getPOST($key, $default = '')            — alias get('post', $key).
 *   - getCOOKIE($key, $default = '')          — alias get('cookie', $key).
 *   - getArgument($key, $default = '')        — POST если есть, иначе GET.
 *   - getMethod()                             — REQUEST_METHOD в верхнем
 *                                                регистре (GET/POST/...).
 *   - setCookie($name, $value, $expire = 0,
 *               $path = '/', $secure = false,
 *               $httpOnly = true,
 *               $samesite = 'Lax')            — обёртка setcookie() с
 *                                                разумными defaults.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Request
{
    private static $instance = null;

    public static function getInstance()
    {
        if(self::$instance === null)
        {
            self::$instance = new self();
        }
        return self::$instance;
    }

    /**
     * Универсальный доступ. Используется в шаблонах через template-тег
     * {request[get/key]} → Core::getRequest()->get('get', 'key').
     */
    public function get($source, $key, $default = '')
    {
        $source = strtolower((string)$source);
        $arr = null;
        switch($source)
        {
            case 'get':    $arr = $_GET;    break;
            case 'post':   $arr = $_POST;   break;
            case 'cookie': $arr = $_COOKIE; break;
            case 'server': $arr = $_SERVER; break;
            default:       return $default;
        }
        return array_key_exists($key, $arr) ? $arr[$key] : $default;
    }

    /**
     * Без $key — возвращает весь массив (legacy: getPOST() в Page.class.php
     * для получения всего POST-запроса).
     */
    public function getGET($key = null, $default = '')
    {
        if($key === null) { return $_GET; }
        return array_key_exists($key, $_GET) ? $_GET[$key] : $default;
    }

    public function getPOST($key = null, $default = '')
    {
        if($key === null) { return $_POST; }
        return array_key_exists($key, $_POST) ? $_POST[$key] : $default;
    }

    public function getCOOKIE($key = null, $default = '')
    {
        if($key === null) { return $_COOKIE; }
        return array_key_exists($key, $_COOKIE) ? $_COOKIE[$key] : $default;
    }

    /**
     * POST имеет приоритет над GET. Используется когда форма может
     * приходить и через POST, и через GET (например, ?action=save).
     */
    public function getArgument($key, $default = '')
    {
        if(array_key_exists($key, $_POST))
        {
            return $_POST[$key];
        }
        if(array_key_exists($key, $_GET))
        {
            return $_GET[$key];
        }
        return $default;
    }

    public function getMethod()
    {
        return isset($_SERVER['REQUEST_METHOD']) ? strtoupper((string)$_SERVER['REQUEST_METHOD']) : 'GET';
    }

    /**
     * Устанавливает HTTP cookie. Defaults безопасные (httpOnly=true,
     * SameSite=Lax) — для критичных кук (JWT) caller передаёт явно
     * SameSite=Strict (см. план 37.7.2).
     */
    public function setCookie($name, $value, $expire = 0, $path = '/', $secure = false, $httpOnly = true, $samesite = 'Lax')
    {
        $opts = array(
            'expires'  => (int)$expire > 0 ? time() + (int)$expire : 0,
            'path'     => (string)$path,
            'secure'   => (bool)$secure,
            'httponly' => (bool)$httpOnly,
            'samesite' => (string)$samesite,
        );
        return setcookie((string)$name, (string)$value, $opts);
    }
}
