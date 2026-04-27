<?php
/**
 * Options — clean-room rewrite (план 43 Ф.6). Заменяет одноимённый класс
 * из фреймворка Recipe (GPL).
 *
 * Конфигурационная коллекция игры: основной источник — таблица na_config
 * (var, value, type), fallback — Config.xml. Используется через
 * Core::getOptions()->get($var) ~30 ключами.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Options extends Collection
{
    public function __construct($yii_way = true)
    {
        $this->item = array();
        $this->loadFromDatabase();
        $this->loadConfigXmlFallback();
    }

    /**
     * Загружает все ключи из таблицы PREFIX.config.
     */
    protected function loadFromDatabase()
    {
        try
        {
            $result = Core::getDB()->query('SELECT var, value, type FROM `'.PREFIX.'config`');
            if($result)
            {
                while($row = Core::getDB()->fetch($result))
                {
                    $this->item[$row['var']] = $this->castByType($row['value'], $row['type']);
                }
                Core::getDB()->free_result($result);
            }
        }
        catch(\Throwable $e)
        {
            // Если таблицы config нет (свежий setup) — игнорируем,
            // продолжаем с Config.xml fallback.
        }
    }

    /**
     * Загружает значения из Config.xml (если ключа ещё нет в коллекции).
     */
    protected function loadConfigXmlFallback()
    {
        $file = RECIPE_ROOT_DIR.'Config.xml';
        if(!is_file($file)) { return; }
        try
        {
            $xml = simplexml_load_file($file);
            if($xml === false) { return; }
            foreach($xml->children() as $node)
            {
                $name = (string)$node->getName();
                if(isset($this->item[$name])) { continue; }
                $attrs = $node->attributes();
                $type = isset($attrs['type']) ? (string)$attrs['type'] : '';
                $this->item[$name] = $this->castByType((string)$node, $type);
            }
        }
        catch(\Throwable $e)
        {
            // ignore
        }
    }

    /**
     * Преобразует строковое значение из XML/DB в PHP-тип по type-аннотации.
     */
    protected function castByType($value, $type)
    {
        $type = strtolower(trim((string)$type));
        switch($type)
        {
            case 'int':
            case 'integer':
                return (int)$value;
            case 'float':
            case 'double':
                return (float)$value;
            case 'bool':
            case 'boolean':
                return $value === '1' || strtolower($value) === 'true';
            case 'array':
                return array_map('trim', explode(',', (string)$value));
            default:
                return (string)$value;
        }
    }

    public function get($var)
    {
        return array_key_exists($var, $this->item) ? $this->item[$var] : null;
    }

    public function set($var, $value)
    {
        $this->item[$var] = $value;
        return $this;
    }

    /**
     * Сохраняет значение в БД (UPSERT в таблицу config).
     */
    public function setValue($var, $value, $renewcache = false)
    {
        $existing = sqlSelectField('config', 'var', '', 'var = '.sqlVal($var));
        if($existing !== null)
        {
            Core::getQuery()->update('config', 'value', (string)$value, 'var = '.sqlVal($var));
        }
        else
        {
            Core::getQuery()->insert('config', array('var', 'value'), array($var, (string)$value));
        }
        $this->item[$var] = $value;
        return $this;
    }
}
