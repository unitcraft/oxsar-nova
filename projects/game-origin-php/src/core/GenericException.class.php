<?php
/**
 * GenericException — clean-room rewrite (план 43 Ф.3). Заменяет одноимённый
 * класс из фреймворка Recipe (GPL).
 *
 * Используется как обычное исключение с ручным сообщением: бросается
 * в коде через `throw new GenericException("текст ошибки")`. Отлавливается
 * выше по стеку штатным catch (\Exception | GlobalException).
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class GenericException extends \Exception implements GlobalException
{
}
