<?php
/**
 * Email — clean-room rewrite (план 43 Ф.5). Заменяет одноимённый класс
 * из фреймворка Recipe (GPL).
 *
 * Минимальный mail-sender для notification-писем (регистрация, смена
 * пароля, активация email). Используется обёрткой над PHP mail().
 *
 * API:
 *   - new Email($to, $subject, $body)
 *   - sendMail(): bool
 *
 * В текущей конфигурации SMTP не настроен — sendMail() пишет письмо
 * в error_log и возвращает true. Полная отправка через symfony/mailer
 * отложена до публичного запуска (план 43 §3 — задача расширения).
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Email
{
    private $to;
    private $subject;
    private $body;

    public function __construct($to, $subject, $body)
    {
        $this->to = (string)$to;
        $this->subject = (string)$subject;
        $this->body = (string)$body;
    }

    public function sendMail()
    {
        if($this->to === '')
        {
            return false;
        }
        // Логирование — для dev/staging окружения. В проде заменить на
        // symfony/mailer + SMTP-конфиг через ENV.
        @error_log(sprintf(
            'Email to=%s subject=%s body_len=%d',
            $this->to,
            $this->subject,
            strlen($this->body)
        ));
        return true;
    }
}
