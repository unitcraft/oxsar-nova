<?php
/**
 * Image — clean-room rewrite (план 43 Ф.2). Заменяет одноимённый класс
 * фреймворка Recipe (GPL). Только метод getImage, реально вызываемый.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Image
{
    /**
     * Генерирует <img>-тег. Сигнатура из legacy:
     *   getImage($name, $alt = '', $width = null, $height = null, $cssClass = '')
     *
     * $name — имя файла относительно RELATIVE_URL/images/ (если без слешей).
     * Если $name содержит слеш — считается относительным к RELATIVE_URL.
     * Если $name начинается с http(s)/data:/// — используется как есть.
     *
     * Не эскейпит $alt — caller отвечает за его безопасность (legacy-семантика).
     */
    public static function getImage($name, $alt = '', $width = null, $height = null, $cssClass = '')
    {
        $name = (string)$name;
        if($name === '') { return ''; }

        // Абсолютный URL — без префикса.
        if(preg_match('~^([a-z][a-z0-9+\-.]*:)~i', $name) || strpos($name, '//') === 0)
        {
            $src = $name;
        }
        else
        {
            $rel = defined('RELATIVE_URL') ? RELATIVE_URL : '/';
            if($rel !== '' && substr($rel, -1) !== '/') { $rel .= '/'; }
            // Legacy-семантика: ВСЕ относительные имена префиксятся
            // /images/ (включая пути с / типа planets/small/x.jpg).
            $src = $rel.'images/'.ltrim($name, '/');
        }

        // Legacy-семантика: alt-атрибут НЕ выводится (только title), default
        // class="image" если cssClass не передан явно.
        if($cssClass === '' || $cssClass === null) { $cssClass = 'image'; }
        $titleAttr = htmlspecialchars((string)$alt, ENT_QUOTES, 'UTF-8');
        $w = $width !== null && $width !== '' ? ' width="'.(int)$width.'"' : '';
        $h = $height !== null && $height !== '' ? ' height="'.(int)$height.'"' : '';
        $cls = ' class="'.htmlspecialchars((string)$cssClass, ENT_QUOTES, 'UTF-8').'"';
        return '<img src="'.$src.'" title="'.$titleAttr.'"'.$w.$h.$cls.' />';
    }
}
