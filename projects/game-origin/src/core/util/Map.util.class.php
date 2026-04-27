<?php
/**
 * Map — clean-room rewrite (план 43 Ф.5). Заменяет одноимённый класс
 * из фреймворка Recipe (GPL). Минимальный API под фактически вызываемые
 * методы (Template->log/htmlHead и Login->errors): __construct, push,
 * size, toString.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

require_once(__DIR__.'/Type.abstract_class.php');

class Map extends Type implements Countable, IteratorAggregate
{
    private $items;

    public function __construct($init = null)
    {
        if(is_array($init))
        {
            $this->items = array_values($init);
        }
        else
        {
            $this->items = array();
        }
    }

    public function push($value)
    {
        $this->items[] = $value;
        return $this;
    }

    public function add($value)
    {
        return $this->push($value);
    }

    /**
     * Установить значение по ключу. Map ведёт себя и как numeric-list
     * (push/add) и как assoc-array (set/get/exists по ключу).
     */
    public function set($key, $value)
    {
        $this->items[$key] = $value;
        return $this;
    }

    public function get($key = null)
    {
        if($key === null)
        {
            return $this->items;
        }
        return array_key_exists($key, $this->items) ? $this->items[$key] : null;
    }

    public function exists($key)
    {
        return array_key_exists($key, $this->items);
    }

    public function size()
    {
        return count($this->items);
    }

    public function count(): int
    {
        return count($this->items);
    }

    /**
     * Склейка элементов через $glue (как PHP implode). $glue = null →
     * без разделителя.
     */
    public function toString($glue = null)
    {
        if($glue === null) { $glue = ''; }
        return implode((string)$glue, array_map(static function($v) {
            return (string)$v;
        }, $this->items));
    }

    public function getString($glue = null)
    {
        return $this->toString($glue);
    }

    public function getArray()
    {
        return $this->items;
    }

    public function getIterator(): Iterator
    {
        return new ArrayIterator($this->items);
    }
}
