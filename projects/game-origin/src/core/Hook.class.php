<?php
/**
 * Hook — clean-room stub (план 43 Ф.3). Заменяет одноимённый класс
 * фреймворка Recipe (GPL).
 *
 * В compiled-шаблонах остались вызовы Hook::event('HtmlBegin', $args)
 * и т.п., но в проекте никто не register-ит обработчики, поэтому event()
 * — пустой no-op. Полное удаление вызовов из шаблонов отложено в Ф.5
 * (когда Template переписываем под Smarty — TemplateCompiler перестаёт
 * генерировать Hook::event строки).
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Hook
{
    /**
     * No-op stub. Никто не register-ит обработчики, поэтому событие
     * никем не обрабатывается.
     *
     * Сигнатура повторяет legacy: event($name, $args = array()).
     */
    public static function event($name, $args = array())
    {
        // intentionally empty
    }
}
