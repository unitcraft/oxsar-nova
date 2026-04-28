<?php
/**
 * DB_MYSQL_PDO — clean-room rewrite (план 43 Ф.4). Заменяет одноимённый
 * класс из фреймворка Recipe (GPL).
 *
 * Тонкая обёртка над PDO/MySQL под legacy-API Database. Методы:
 *   - query/fetch/num_rows/insert_id/quote_db_value/free_result —
 *     реализация abstract Database.
 *   - В query() ведётся учёт queryCount + qTime для getQueryNumber/Time
 *     (вывод в footer cake-debug панели и т.п.).
 *
 * Имя класса (с большой `S`) сохранено для совместимости с factory в
 * Core::setDatabase (`new $database["type"]`).
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

// Autoloader не загружает .abstract_class.php — подгружаем явно
// (legacy convention: каждый подкласс делает require_once своего abstract).
require_once(__DIR__.'/Database.abstract_class.php');

class DB_MYSQL_PDO extends Database
{
    /** @var \PDO */
    private $pdo;

    public function __construct($host, $user, $password, $database, $port = null)
    {
        $port = $port !== null && $port !== '' ? (int)$port : 3306;
        $dsn = sprintf(
            'mysql:host=%s;port=%d;dbname=%s;charset=utf8mb4',
            $host,
            $port,
            $database
        );
        $opts = array(
            \PDO::ATTR_ERRMODE            => \PDO::ERRMODE_EXCEPTION,
            \PDO::ATTR_EMULATE_PREPARES   => false,
            \PDO::ATTR_DEFAULT_FETCH_MODE => \PDO::FETCH_ASSOC,
        );
        $this->pdo = new \PDO($dsn, (string)$user, (string)$password, $opts);
    }

    public function query($sql)
    {
        $start = microtime(true);
        try
        {
            $stmt = $this->pdo->query((string)$sql);
        }
        catch(\PDOException $e)
        {
            // Совместимость с legacy: вернуть false вместо throw, чтобы
            // вызывающий код мог проверить результат на falsy.
            error_log('DB_MYSQL_PDO::query failed: '.$e->getMessage().' :: '.$sql);
            return false;
        }
        $this->queryCount++;
        $this->qTime[$this->queryCount] = number_format(microtime(true) - $start, 5, '.', '');
        return $stmt;
    }

    public function fetch($result)
    {
        if(!$result instanceof \PDOStatement)
        {
            return false;
        }
        $row = $result->fetch(\PDO::FETCH_ASSOC);
        return $row === false ? false : $row;
    }

    public function num_rows($result)
    {
        if(!$result instanceof \PDOStatement)
        {
            return 0;
        }
        return $result->rowCount();
    }

    public function insert_id()
    {
        return $this->pdo->lastInsertId();
    }

    public function quote_db_value($value)
    {
        if($value === null)
        {
            return 'NULL';
        }
        return $this->pdo->quote((string)$value);
    }

    public function free_result($resource)
    {
        if($resource instanceof \PDOStatement)
        {
            $resource->closeCursor();
        }
    }
}

// Legacy alias: некоторые места используют lowercase `DB_mysql_pdo`
// (как в `class DB_mysql_pdo extends Database` оригинального файла).
// PHP автоматически нормализует имена классов как case-insensitive,
// но class_alias делает алиас явным для явных проверок class_exists().
if(!class_exists('DB_mysql_pdo'))
{
    class_alias('DB_MYSQL_PDO', 'DB_mysql_pdo');
}
