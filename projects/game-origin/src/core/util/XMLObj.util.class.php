<?php
/**
 * XMLObj — clean-room rewrite (план 43 Ф.5). Заменяет одноимённый класс
 * из фреймворка Recipe (GPL).
 *
 * Тонкая обёртка над SimpleXMLElement с legacy-API:
 *   - getAttribute($name): string — вернуть значение атрибута.
 *   - getName(): string — имя тэга.
 *   - getString(): string — текст внутри тэга (без подтэгов).
 *   - getChildren(): XMLObj[] — массив дочерних элементов.
 *
 * Используется Menu, Options, PlanetCreator для парсинга XML-config.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class XMLObj
{
    /** @var \SimpleXMLElement */
    private $node;

    public function __construct(\SimpleXMLElement $node)
    {
        $this->node = $node;
    }

    public function getAttribute($name)
    {
        $attrs = $this->node->attributes();
        return isset($attrs[$name]) ? (string)$attrs[$name] : '';
    }

    public function getName()
    {
        return $this->node->getName();
    }

    public function getString()
    {
        // Текстовое содержимое узла без вложенных тэгов.
        // SimpleXMLElement->__toString() возвращает текст до первого подэлемента.
        return trim((string)$this->node);
    }

    /**
     * Возвращает массив дочерних XMLObj. Текстовые узлы пропускаются —
     * только element-узлы.
     */
    public function getChildren()
    {
        $children = array();
        foreach($this->node->children() as $child)
        {
            $children[] = new XMLObj($child);
        }
        return $children;
    }

    /**
     * Доступ к underlying SimpleXMLElement (для расширений).
     */
    public function getNode()
    {
        return $this->node;
    }
}
