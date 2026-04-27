<?php
/**
 * QueryParser — clean-room rewrite (план 43 Ф.6). Заменяет одноимённый
 * класс из фреймворка Recipe (GPL).
 *
 * Минимальный SQL builder поверх Core::getDB(). Сохранены сигнатуры
 * 9 методов реально вызываемых в проекте: insert/replace/update/delete/
 * select/selectRow/selectField/showFields/optimize.
 *
 * Все SQL-значения экранируются через Core::getDatabase()->quote_db_value()
 * (PDO::quote под капотом). $select-аргумент может быть массивом
 * полей или строкой (raw SQL); $where, $order, $limit — строки SQL,
 * передаются как есть.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class QueryParser
{
    /**
     * INSERT (или REPLACE при $op='REPLACE') в таблицу.
     * $attribute — массив имён колонок; $value — значения той же длины.
     */
    public function insert($table, $attribute, $value, $op = 'INSERT')
    {
        $cols = array();
        foreach((array)$attribute as $a)
        {
            $cols[] = '`'.str_replace('`', '', (string)$a).'`';
        }
        $vals = array();
        foreach((array)$value as $v)
        {
            $vals[] = $v === null ? 'NULL' : Core::getDatabase()->quote_db_value((string)$v);
        }
        $sql = strtoupper($op).' INTO '.PREFIX.$table
            .' ('.implode(', ', $cols).') VALUES ('.implode(', ', $vals).')';
        return Core::getDB()->query($sql);
    }

    public function replace($table, $attribute, $value)
    {
        return $this->insert($table, $attribute, $value, 'REPLACE');
    }

    /**
     * UPDATE. $attribute может быть массивом или строкой (одно поле);
     * $value — массив либо одиночное значение либо ассоциативный массив
     * key=>value (тогда $value игнорируется).
     */
    public function update($table, $attribute, $value, $where = null)
    {
        $sets = array();
        if(is_array($attribute) && !is_array($value) && $value === null)
        {
            // Ассоциативный массив key=>value передан в $attribute.
            foreach($attribute as $k => $v)
            {
                $sets[] = '`'.str_replace('`', '', (string)$k).'` = '
                    .($v === null ? 'NULL' : Core::getDatabase()->quote_db_value((string)$v));
            }
        }
        elseif(is_array($attribute))
        {
            $values = (array)$value;
            $i = 0;
            foreach($attribute as $a)
            {
                $v = isset($values[$i]) ? $values[$i] : null;
                $sets[] = '`'.str_replace('`', '', (string)$a).'` = '
                    .($v === null ? 'NULL' : Core::getDatabase()->quote_db_value((string)$v));
                $i++;
            }
        }
        else
        {
            $sets[] = '`'.str_replace('`', '', (string)$attribute).'` = '
                .($value === null ? 'NULL' : Core::getDatabase()->quote_db_value((string)$value));
        }
        $sql = 'UPDATE '.PREFIX.$table.' SET '.implode(', ', $sets);
        if($where !== null && $where !== '')
        {
            $sql .= ' WHERE '.$where;
        }
        return Core::getDB()->query($sql);
    }

    public function delete($table, $where = null)
    {
        $sql = 'DELETE FROM '.PREFIX.$table;
        if($where !== null && $where !== '')
        {
            $sql .= ' WHERE '.$where;
        }
        return Core::getDB()->query($sql);
    }

    /**
     * SELECT. $select — массив колонок (объединяются через ', ') или
     * строка (используется как есть, например 'COUNT(*)' или '*'). Все
     * прочие аргументы — raw SQL.
     */
    public function select($table, $select, $join = '', $where = '', $order = '', $limit = '', $groupby = '', $other = '')
    {
        $cols = is_array($select) ? implode(', ', $select) : (string)$select;
        $sql = 'SELECT '.$cols.' FROM '.PREFIX.$table;
        if($join !== '')    $sql .= ' '.$join;
        if($where !== '')   $sql .= ' WHERE '.$where;
        if($groupby !== '') $sql .= ' GROUP BY '.$groupby;
        if($order !== '')   $sql .= ' ORDER BY '.$order;
        if($limit !== '')   $sql .= ' LIMIT '.$limit;
        if($other !== '')   $sql .= ' '.$other;
        return Core::getDB()->query($sql);
    }

    /**
     * SELECT … LIMIT 1 → первая строка как assoc-array, либо null.
     */
    public function selectRow($table, $select, $join = '', $where = '', $order = '', $limit = '1', $groupby = '', $other = '')
    {
        $result = $this->select($table, $select, $join, $where, $order, $limit, $groupby, $other);
        if(!$result) { return null; }
        $row = Core::getDB()->fetch($result);
        Core::getDB()->free_result($result);
        return $row !== false ? $row : null;
    }

    /**
     * SELECT … LIMIT 1 → первое поле первой строки.
     */
    public function selectField($table, $select, $join = '', $where = '', $order = '', $limit = '1', $groupby = '', $other = '')
    {
        $row = $this->selectRow($table, $select, $join, $where, $order, $limit, $groupby, $other);
        if(!is_array($row) || count($row) === 0) { return null; }
        return reset($row);
    }

    public function optimize($tables)
    {
        $tables = is_array($tables) ? $tables : array($tables);
        $names = array();
        foreach($tables as $t)
        {
            $names[] = PREFIX.$t;
        }
        return Core::getDB()->query('OPTIMIZE TABLE '.implode(', ', $names));
    }

    public function showFields($table)
    {
        return Core::getDB()->query('SHOW FIELDS FROM '.PREFIX.$table);
    }
}
