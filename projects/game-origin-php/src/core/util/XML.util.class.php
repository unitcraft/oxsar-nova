<?php
/**
 * XML — clean-room rewrite (план 43 Ф.5). Заменяет одноимённый класс
 * из фреймворка Recipe (GPL).
 *
 * Загрузчик XML-файла. API:
 *   - new XML($filename) — парсит файл через SimpleXML.
 *   - get(): XMLObj — root-элемент как XMLObj.
 *
 * Используется Menu (menu.xml), Options (options.xml), PlanetCreator
 * (planet-templates.xml).
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

require_once(__DIR__.'/XMLObj.util.class.php');

class XML
{
    /** @var \SimpleXMLElement */
    private $root;

    public function __construct($filename)
    {
        if(!is_string($filename) || $filename === '')
        {
            throw new \InvalidArgumentException('XML: empty filename');
        }
        if(!is_file($filename))
        {
            throw new \RuntimeException('XML: file not found: '.$filename);
        }
        $xml = simplexml_load_file($filename);
        if($xml === false)
        {
            throw new \RuntimeException('XML: parse failed: '.$filename);
        }
        $this->root = $xml;
    }

    public function get()
    {
        return new XMLObj($this->root);
    }
}
