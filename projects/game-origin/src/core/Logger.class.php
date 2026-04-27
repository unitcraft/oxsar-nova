<?php
/**
 * Logger — clean-room rewrite (план 43 Ф.3). Заменяет одноимённый класс
 * фреймворка Recipe (GPL). Все методы — статические, обёртывают сообщение
 * в HTML с CSS-классом и отдают в Template (или session для flash).
 *
 * Сообщения: либо ключ из i18n-словаря 'error' (резолвится через
 * Core::getLanguage()->getItem), либо raw HTML.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Logger
{
    /**
     * Регистрирует сообщение в текущем шаблоне (отображается до завершения
     * рендера). $mode — CSS-класс div'а (error/info/success/warning).
     */
    public static function addMessage($key, $mode = 'error')
    {
        $text = self::resolveText($key);
        $html = '<div class="'.htmlspecialchars((string)$mode, ENT_QUOTES, 'UTF-8').'">'.$text.'</div>';
        $tpl = Core::getTPL();
        if($tpl)
        {
            $tpl->addLogMessage($html);
        }
    }

    /**
     * Регистрирует «flash»-сообщение в session — будет показано на
     * следующем рендере и автоматически удалено при отображении.
     * $mode — символический ключ ('error', 'success', либо raw CSS).
     * Для известных ключей собирается типовая разметка с jQuery-UI иконкой.
     */
    public static function addFlashMessage($key, $mode = 'error')
    {
        $text = self::resolveText($key);
        $cssClass = (string)$mode;
        $iconClass = null;

        switch($mode)
        {
            case 'success':
                $cssClass = 'success ui-state-highlight ui-corner-all';
                $iconClass = 'ui-icon-info';
                break;
            case 'error':
                $cssClass = 'error ui-state-error ui-corner-all';
                $iconClass = 'ui-icon-alert';
                break;
            case 'ui-state-error ui-corner-all':
                $iconClass = 'ui-icon-alert';
                break;
        }

        if($iconClass !== null)
        {
            $text = '<td><span class="ui-icon '.$iconClass.'" style="float: left; margin-right: .3em;"></span></td>'
                .'<td>'.$text.'</td>';
        }

        if(session_status() !== PHP_SESSION_ACTIVE)
        {
            // Если сессия не открыта — flash потеряется, тихо игнорируем
            // (legacy-семантика — нет error на этом пути).
            return;
        }

        $i = 0;
        $base = 'flash_'.$cssClass.' logger';
        while(isset($_SESSION[$base.$i]))
        {
            $i++;
        }
        $_SESSION[$base.$i] = $text;
    }

    /**
     * Регистрирует сообщение и завершает страницу через рендер error-шаблона.
     * Используется в обработчиках ошибок ввода (CSRF, недостаточно ресурсов).
     */
    public static function dieMessage($key, $mode = 'error')
    {
        // В системный лог пишем сырой ключ (для grep по error_log).
        @error_log((string)$key);

        $text = self::resolveText($key);
        $html = '<div class="'.htmlspecialchars((string)$mode, ENT_QUOTES, 'UTF-8').'">'.$text.'</div>';
        $tpl = Core::getTPL();
        if($tpl)
        {
            $tpl->addLogMessage($html);
        }
        $core = Core::getTemplate();
        if($core)
        {
            $core->display('error');
        }
        exit;
    }

    /**
     * Возвращает HTML-фрагмент для inline-отображения в форме (не пишет
     * в Template/Session). $mode попадает в имя CSS-класса как field_$mode.
     */
    public static function getMessageField($key, $mode = 'error')
    {
        $text = self::resolveText($key);
        return '<span class="field_'.htmlspecialchars((string)$mode, ENT_QUOTES, 'UTF-8').'">'.$text.'</span>';
    }

    /**
     * Резолвит ключ через i18n-словарь 'error'. Если Language-сервис
     * недоступен или ключ не найден — возвращает ключ как есть (legacy
     * fallback — getItem обычно отдаёт сам ключ при отсутствии перевода).
     */
    private static function resolveText($key)
    {
        $key = (string)$key;
        $lang = Core::getLanguage();
        if(!$lang)
        {
            return $key;
        }
        $lang->load('error');
        $resolved = $lang->getItem($key);
        return $resolved !== null && $resolved !== '' ? $resolved : $key;
    }
}
