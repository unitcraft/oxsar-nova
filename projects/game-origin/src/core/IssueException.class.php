<?php
/**
 * IssueException — clean-room rewrite (план 43 Ф.3). Заменяет одноимённый
 * класс из фреймворка Recipe (GPL).
 *
 * Используется когда сообщение — ключ из i18n-словаря (типа
 * "NO_VALID_EMAIL_ADDRESS"). Содержит хелперы которые делегируют в Logger
 * для отображения локализованного текста.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class IssueException extends \Exception implements GlobalException
{
    /**
     * Регистрирует i18n-сообщение в текущем шаблоне через Logger.
     */
    public function addMessage()
    {
        Logger::addMessage($this->message);
    }

    /**
     * Завершает страницу через рендер error-template (Logger::dieMessage).
     */
    public function dieMessage()
    {
        Logger::dieMessage($this->message);
    }

    /**
     * Возвращает HTML-фрагмент с локализованным текстом для inline-показа
     * в форме.
     */
    public function getMessageField()
    {
        return Logger::getMessageField($this->message);
    }
}
