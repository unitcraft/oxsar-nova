<?php
/**
 * Date — clean-room rewrite (план 43 Ф.2). Заменяет одноимённый класс
 * фреймворка Recipe (GPL). Только методы, реально вызываемые в проекте.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Date
{
    /**
     * Форматирует timestamp как 'Y-m-d H:i:s'. Используется для сравнения
     * с MySQL DATETIME-полями.
     */
    public static function getDateTime($timestamp = null)
    {
        if($timestamp === null) { $timestamp = time(); }
        return date('Y-m-d H:i:s', (int)$timestamp);
    }

    /**
     * Форматирует timestamp в человекочитаемую строку.
     *
     * Сигнатура из legacy:
     *   timeToString($mode, $timestamp = -1, $format = '', $useModified = true)
     *
     * Режимы (mode), реально используемые в коде:
     *   1 — стандартное «Y-m-d H:i:s» (или локализованное).
     *   3 — форматирование по переданному $format (date()-совместимый).
     *
     * Если $timestamp < 0 — использовать текущее время.
     * Параметр $useModified в текущей реализации игнорируется (он отвечал
     * за timezone-сдвиг для UTC vs local; сейчас используется PHP default
     * timezone из php.ini / TIMEZONE env-var в Dockerfile).
     */
    public static function timeToString($mode, $timestamp = -1, $format = '', $useModified = true)
    {
        if($timestamp === null || $timestamp < 0) { $timestamp = time(); }
        $timestamp = (int)$timestamp;
        switch((int)$mode)
        {
            case 3:
                if($format === '' || $format === null)
                {
                    $format = 'Y-m-d H:i:s';
                }
                return date($format, $timestamp);
            case 1:
            default:
                return date('Y-m-d H:i:s', $timestamp);
        }
    }

    /**
     * Проверяет что строка $date парсится как валидная дата формата 'Y-m-d'
     * (или иной, если задан $format).
     */
    public static function validateDate($date, $format = 'Y-m-d')
    {
        if(!is_string($date) || $date === '') { return false; }
        $dt = DateTime::createFromFormat($format, $date);
        return $dt !== false && $dt->format($format) === $date;
    }
}
