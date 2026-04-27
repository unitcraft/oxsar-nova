<?php
/**
 * Database — clean-room rewrite (план 43 Ф.4). Заменяет одноимённый
 * абстрактный класс из фреймворка Recipe (GPL).
 *
 * Базовый класс для всех DB-адаптеров. Подкласс должен реализовать
 * connect/query/fetch/free_result/num_rows/insert_id/quote_db_value
 * — конкретный механизм (mysqli, PDO). Базовый класс реализует общие
 * хелперы поверх них (queryRow, queryField, query_unique, статистику).
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

abstract class Database
{
    /**
     * Счётчик выполненных запросов с момента создания подключения.
     */
    protected $queryCount = 0;

    /**
     * Время выполнения каждого запроса в секундах. Ключ — порядковый
     * номер запроса (1-based).
     */
    protected $qTime = array();

    // === Должны быть реализованы подклассом ===

    abstract public function query($sql);
    abstract public function fetch($result);
    abstract public function num_rows($result);
    abstract public function insert_id();
    abstract public function quote_db_value($string);
    abstract public function free_result($resource);

    // === Хелперы поверх abstract методов ===

    /**
     * Выполняет запрос и возвращает первую строку как assoc-array.
     * Для запросов которые точно вернут одну строку (SELECT … LIMIT 1).
     */
    public function queryRow($sql)
    {
        $result = $this->query($sql);
        if(!$result)
        {
            return null;
        }
        $row = $this->fetch($result);
        $this->free_result($result);
        return $row !== false ? $row : null;
    }

    /**
     * Выполняет запрос и возвращает первое поле первой строки.
     * Для SELECT COUNT(*), SELECT MAX(...), SELECT name FROM ... LIMIT 1.
     */
    public function queryField($sql)
    {
        $row = $this->queryRow($sql);
        if(!is_array($row) || count($row) === 0)
        {
            return null;
        }
        // Первое значение независимо от имени колонки.
        return reset($row);
    }

    /**
     * Алиас queryRow — для legacy callers (Language.class.php).
     */
    public function query_unique($sql)
    {
        return $this->queryRow($sql);
    }

    public function getQueryNumber()
    {
        return $this->queryCount;
    }

    /**
     * Возвращает время выполнения конкретного запроса (по 1-based индексу)
     * либо суммарное время всех запросов, если индекс не указан.
     */
    public function getQueryTime($queryid = null)
    {
        if($queryid !== null)
        {
            return isset($this->qTime[$queryid]) ? $this->qTime[$queryid] : 0;
        }
        $sum = 0;
        foreach($this->qTime as $t)
        {
            $sum += (float)$t;
        }
        return $sum;
    }

    /**
     * Stub для legacy QueryParser. Не реализуем — QueryParser реально
     * не используется (никто не вызывает Core::getQueryParser()).
     * Возвращает null вместо имени поля.
     */
    public function fetch_field($resource, $field, $row = null)
    {
        return null;
    }
}
