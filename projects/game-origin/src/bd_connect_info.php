<?php
// Database config — all values from environment variables.
// Copy .env.example to .env and set values, or pass via docker-compose environment.

define('DB_HOST',   getenv('DB_HOST')   ?: '127.0.0.1');
define('DB_PORT',   getenv('DB_PORT')   ?: '3306');
define('DB_PREFIX', getenv('DB_PREFIX') ?: 'na_');
define('DB_TYPE',   'DB_MYSQL_PDO');
define('DB_USER',   getenv('DB_USER')   ?: 'oxsar_user');
define('DB_PWD',    getenv('DB_PWD')    ?: 'oxsar_pass');
define('DB_NAME',   getenv('DB_NAME')   ?: 'oxsar_db');
define('DB_CHAR',   'utf8mb4');

defined('TIMEZONE') or define('TIMEZONE', getenv('TIMEZONE') ?: 'Europe/Moscow');

// Universe name shown in UI (Main.class.php)
defined('UNIVERSE_NAME_FULL') or define('UNIVERSE_NAME_FULL', getenv('UNIVERSE_NAME_FULL') ?: 'Origin');

// Portal URL for auth redirects (plan-36)
defined('PORTAL_URL') or define('PORTAL_URL', getenv('PORTAL_URL') ?: '');
