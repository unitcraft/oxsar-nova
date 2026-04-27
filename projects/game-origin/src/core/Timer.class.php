<?php
/**
 * Timer — clean-room rewrite (план 43 Ф.2). Заменяет одноимённый класс
 * фреймворка Recipe (GPL). Минимальный API: конструктор фиксирует точку
 * старта, getTime() возвращает прошедшее время.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Timer
{
    private $start;

    /**
     * Конструктор. Принимает либо результат microtime() (строка
     * "msec sec"), либо число (float seconds), либо null (берёт текущее).
     */
    public function __construct($startTime = null)
    {
        if($startTime === null)
        {
            $this->start = microtime(true);
            return;
        }
        if(is_numeric($startTime))
        {
            $this->start = (float)$startTime;
            return;
        }
        if(is_string($startTime) && strpos($startTime, ' ') !== false)
        {
            // microtime() формат: "0.12345600 1700000000"
            list($msec, $sec) = explode(' ', $startTime);
            $this->start = (float)$sec + (float)$msec;
            return;
        }
        $this->start = microtime(true);
    }

    /**
     * Возвращает прошедшее время с момента старта.
     * Если $formatted = true (default) — отформатированная строка
     * с 4 знаками после запятой; иначе — float.
     */
    public function getTime($formatted = true)
    {
        $elapsed = microtime(true) - $this->start;
        if($formatted === false)
        {
            return $elapsed;
        }
        return number_format($elapsed, 4, '.', '');
    }
}
