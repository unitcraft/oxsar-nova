<?php
/**
 * Link — clean-room rewrite (план 43 Ф.2). Заменяет одноимённый класс
 * фреймворка Recipe (GPL). Только методы, реально вызываемые в проекте.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Link
{
    /**
     * Генерирует HTML <a>-тег. $url — относительный (внутренний) или
     * абсолютный URL; $text — текст ссылки; $title — атрибут title.
     *
     * Внутренние ссылки префиксятся RELATIVE_URL и проходят через socialUrl()
     * для поддержки SN-режимов (если константа SN определена).
     *
     * Не эскейпит $text и $title — caller отвечает за их безопасность
     * (legacy-семантика). См. план 37.7.1+ для XSS-аудита call-sites.
     */
    public static function get($url, $text = '', $title = '', $cssClass = 'link', $attachment = '')
    {
        $url = (string)$url;
        if($url === '')
        {
            return (string)$text;
        }
        $isExt = self::isExternal($url);
        $href = $isExt ? $url : self::normalizeURL($url);
        // Legacy-семантика: пустая строка = взять default. Внешние ссылки
        // получают class="external" вместо "link".
        if($cssClass === '' || $cssClass === null)
        {
            $cssClass = $isExt ? 'external' : 'link';
        }
        $cls = ' class="'.htmlspecialchars((string)$cssClass, ENT_QUOTES, 'UTF-8').'"';
        $tit = ' title="'.htmlspecialchars((string)$title, ENT_QUOTES, 'UTF-8').'"';
        // Внешние ссылки получают target='_blank' автоматически — но только
        // если caller не передал target в $attachment (избегаем дублирования).
        $att = $attachment !== '' ? ' '.$attachment : '';
        $target = ($isExt && strpos((string)$attachment, 'target=') === false) ? " target='_blank'" : '';
        return '<a href="'.$href.'"'.$tit.$cls.$att.$target.'>'.$text.'</a>';
    }

    /**
     * Внешняя ли ссылка. Считается внешней если начинается со схемы
     * (http://, https://, ftp://, mailto:, javascript:) и хост не совпадает
     * с FULL_URL/HTTP_HOST.
     */
    public static function isExternal($url)
    {
        if(!is_string($url) || $url === '') { return false; }
        // Только URL со схемой могут быть external.
        if(!preg_match('~^([a-z][a-z0-9+\-.]*:)~i', $url))
        {
            return false;
        }
        // mailto:, javascript:, tel: — считаем «внешними» (caller обычно
        // оборачивает их в специальную обработку).
        if(preg_match('~^(mailto|javascript|tel|data):~i', $url))
        {
            return true;
        }
        // http(s):// — сравниваем хост с нашим.
        $host = parse_url($url, PHP_URL_HOST);
        if(!$host)
        {
            return false;
        }
        $ourHost = '';
        if(defined('FULL_URL'))
        {
            $ourHost = parse_url(FULL_URL, PHP_URL_HOST);
        }
        if(!$ourHost && !empty($_SERVER['HTTP_HOST']))
        {
            $ourHost = $_SERVER['HTTP_HOST'];
        }
        if(!$ourHost)
        {
            return true;
        }
        return strcasecmp($host, $ourHost) !== 0;
    }

    /**
     * Нормализует относительный URL: префиксит RELATIVE_URL если её ещё нет,
     * пропускает через socialUrl() для SN-режима.
     */
    public static function normalizeURL($url)
    {
        $url = (string)$url;
        if($url === '') { return ''; }
        // Уже абсолютный — не трогаем.
        if(preg_match('~^([a-z][a-z0-9+\-.]*:)~i', $url) || strpos($url, '//') === 0)
        {
            return $url;
        }
        // Префикс RELATIVE_URL если ссылка не начинается с / и не содержит
        // уже пути относительно корня.
        $rel = defined('RELATIVE_URL') ? RELATIVE_URL : '/';
        if($rel !== '' && substr($rel, -1) !== '/')
        {
            $rel .= '/';
        }
        if($url[0] !== '/')
        {
            $url = $rel.$url;
        }
        if(function_exists('socialUrl'))
        {
            $url = socialUrl($url);
        }
        return $url;
    }
}
