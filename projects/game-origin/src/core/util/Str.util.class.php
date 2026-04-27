<?php
/**
 * Str — clean-room rewrite (план 43 Ф.2). Заменяет одноимённый класс
 * фреймворка Recipe (GPL). Только методы, реально вызываемые в проекте.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Str
{
    /**
     * Длина строки в Unicode-символах (legacy multi-byte safe).
     * Игнорирует второй аргумент charset (legacy сигнатура).
     */
    public static function length($s, $charset = null)
    {
        if($s === null) { return 0; }
        $s = (string)$s;
        if(function_exists('mb_strlen'))
        {
            return mb_strlen($s, 'UTF-8');
        }
        return strlen($s);
    }

    /**
     * Подстрока: $s[$start..$start+$len). Если $len опущен — до конца.
     * Multi-byte safe для UTF-8.
     */
    public static function substring($s, $start, $len = null)
    {
        if($s === null) { return ''; }
        $s = (string)$s;
        if(function_exists('mb_substr'))
        {
            return $len === null
                ? mb_substr($s, (int)$start, null, 'UTF-8')
                : mb_substr($s, (int)$start, (int)$len, 'UTF-8');
        }
        return $len === null ? substr($s, (int)$start) : substr($s, (int)$start, (int)$len);
    }

    // PHP методы case-insensitive — Str::Substring и Str::substring это
    // один метод. Отдельный alias не нужен (legacy callers с большой S
    // уже работают через substring() выше).

    /**
     * Замена подстрок. Тонкая обёртка над str_replace для совместимости с
     * legacy сигнатурой (search, replace, subject).
     */
    public static function replace($search, $replace, $subject)
    {
        if($subject === null) { return ''; }
        return str_replace($search, $replace, (string)$subject);
    }

    /**
     * Поиск $needle в $haystack. Возвращает true/false. Аналог str_contains.
     */
    public static function inString($needle, $haystack)
    {
        if($haystack === null || $needle === null) { return false; }
        return strpos((string)$haystack, (string)$needle) !== false;
    }

    /**
     * Сравнение строк (case-sensitive). 0 — равны, !=0 — не равны.
     */
    public static function compare($a, $b)
    {
        return strcmp((string)$a, (string)$b);
    }

    /**
     * URL-encode для простых случаев.
     */
    public static function encode($s)
    {
        if($s === null) { return ''; }
        return rawurlencode((string)$s);
    }

    /**
     * Возвращает часть строки до последнего вхождения $needle.
     * Если $include_needle = true — включая сам needle.
     * Если $needle не найден — возвращает исходную строку.
     */
    public static function reverse_strrchr($haystack, $needle, $include_needle = false)
    {
        if($haystack === null || $needle === null) { return ''; }
        $haystack = (string)$haystack;
        $needle = (string)$needle;
        $pos = strrpos($haystack, $needle);
        if($pos === false)
        {
            return $haystack;
        }
        return substr($haystack, 0, $include_needle ? $pos + strlen($needle) : $pos);
    }

    /**
     * Очищает строку от потенциально опасных HTML/JS-конструкций для
     * совместимости с XHTML. Используется для subject/message в личных
     * сообщениях. Это НЕ полная XSS-санитизация — для критичных полей
     * см. htmlspecialchars в DAO-getter (см. план 37.7.1+).
     */
    public static function validateXHTML($s)
    {
        if($s === null) { return ''; }
        $s = (string)$s;
        // Базовая защита: заменить < > " ' & на сущности. Это безопасно для
        // отображения в HTML-тексте, но делает невозможным сохранение
        // legitimate HTML — если потребуется, переключиться на HTMLPurifier.
        return htmlspecialchars($s, ENT_QUOTES | ENT_SUBSTITUTE, 'UTF-8');
    }
}
