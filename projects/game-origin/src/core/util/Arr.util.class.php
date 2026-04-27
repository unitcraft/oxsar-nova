<?php
/**
 * Arr — clean-room rewrite (план 43 Ф.2). Заменяет одноимённый класс
 * фреймворка Recipe (GPL). Оставлены только методы, реально вызываемые
 * в проекте (см. grep на 2026-04-27).
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Arr
{
    /**
     * Применяет trim() к каждому элементу массива (или одиночной строке).
     * Используется для разбиения CSV-подобных строк после explode().
     */
    public static function trimArray($a)
    {
        if(!is_array($a))
        {
            return is_string($a) ? trim($a) : $a;
        }
        return array_map(static function($v) {
            return is_string($v) ? trim($v) : $v;
        }, $a);
    }

    /**
     * Алиас trimArray — сохранён для обратной совместимости с legacy callers.
     */
    public static function trim($a)
    {
        return self::trimArray($a);
    }

    /**
     * Проверяет что оба аргумента — массивы одинакового размера. Throws
     * на несоответствие. Используется QueryParser-ом для пар (attribute, value).
     */
    public static function checkArrays($a, $b)
    {
        if(!is_array($a) || !is_array($b))
        {
            throw new InvalidArgumentException('Both arguments must be arrays');
        }
        if(count($a) !== count($b))
        {
            throw new InvalidArgumentException('Arrays must be the same size');
        }
        return true;
    }

    /**
     * Проверяет что массив имеет ожидаемый размер. Throws при несоответствии.
     */
    public static function checkArraySize($a, $expected)
    {
        if(!is_array($a))
        {
            throw new InvalidArgumentException('First argument must be an array');
        }
        if(count($a) !== (int)$expected)
        {
            throw new InvalidArgumentException('Array size mismatch: got '.count($a).', expected '.$expected);
        }
        return true;
    }
}
