<?php
/**
 * AjaxRequestHelper — clean-room rewrite (план 43 Ф.3). Заменяет
 * одноимённый абстрактный класс из фреймворка Recipe (GPL).
 *
 * Базовый класс для AJAX-страниц (FleetAjax, AccountCreator,
 * PasswordChanger, LostPassword). Унаследован от Page, добавляет
 * способ отдать сырой ответ (текст / JSON) с правильным Content-Type
 * без штатного шаблонизатора.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

abstract class AjaxRequestHelper extends Page
{
    /**
     * Отправляет ответ клиенту и завершает запрос. $outstream — строка
     * (HTML-фрагмент / JSON). Перед echo посылает заголовки через
     * sendHeader().
     */
    protected function display($outstream)
    {
        $this->sendHeader();
        echo (string)$outstream;
    }

    /**
     * Заголовки для AJAX-ответов. Default: text/html UTF-8 + no-cache.
     * Подклассы могут переопределить (например, для JSON).
     */
    protected function sendHeader()
    {
        if(!headers_sent())
        {
            header('Content-Type: text/html; charset=UTF-8');
            header('Cache-Control: no-store, no-cache, must-revalidate, max-age=0');
            header('Pragma: no-cache');
        }
    }
}
