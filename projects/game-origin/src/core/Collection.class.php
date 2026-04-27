<?php
/**
 * Collection — clean-room rewrite (план 43 Ф.3). Заменяет одноимённый
 * абстрактный класс из фреймворка Recipe (GPL).
 *
 * Базовый класс для именованных коллекций «ключ -> значение». Subclass'ы:
 * User, Language, Options. Реализует обход (foreach) и подсчёт (count).
 * Конкретный get/set остаётся за подклассом — Collection не диктует
 * стратегию (lazy load из БД, кэш, и т.п.).
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

abstract class Collection implements Countable, IteratorAggregate
{
    /**
     * Собственно данные коллекции. Subclass'ы наполняют его при загрузке.
     */
    protected $item = array();

    /**
     * Возвращает значение по ключу. Семантика «не найдено» — на усмотрение
     * подкласса (false / null / throw).
     */
    public abstract function get($var);

    /**
     * Записывает значение по ключу. Подкласс может реджектить запись
     * (read-only поля).
     */
    public abstract function set($var, $value);

    /**
     * Существует ли ключ в коллекции. Не проверяет null-значения
     * специально — array_key_exists совместим с null.
     */
    public function exists($var)
    {
        return is_array($this->item) && array_key_exists($var, $this->item);
    }

    /**
     * Реализация Countable — для count($collection).
     */
    public function count(): int
    {
        return is_array($this->item) ? count($this->item) : 0;
    }

    /**
     * Реализация IteratorAggregate — для foreach ($collection as $k=>$v).
     */
    public function getIterator(): Iterator
    {
        return new ArrayIterator(is_array($this->item) ? $this->item : array());
    }
}
