<?php
/**
 * GlobalException — clean-room rewrite (план 43 Ф.3). Заменяет одноимённый
 * интерфейс из фреймворка Recipe (GPL).
 *
 * Маркерный интерфейс для собственных исключений проекта (отделить
 * domain-ошибки от системных PHP-исключений).
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

interface GlobalException
{
}
