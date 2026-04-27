<?php
/**
 * OxsarString — clean-room rewrite (план 43 Ф.5). Заменяет одноимённый
 * класс из фреймворка Recipe (GPL).
 *
 * Mutable string-buffer с fluent API. Используется в TemplateCompiler
 * и LanguageCompiler для построения итогового compiled-текста через
 * последовательные regex-замены.
 *
 * Минимальный API под фактически вызываемые методы:
 *   - __construct($initial)
 *   - get(): string — вернуть текущее содержимое.
 *   - regEx($pattern, $replacement): self — preg_replace, fluent.
 *   - push($str): self — добавить в конец, fluent.
 *   - pop($str): self — добавить в начало (legacy-name, не array-pop), fluent.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

require_once(__DIR__.'/Type.abstract_class.php');

class OxsarString extends Type
{
    private $value;

    public function __construct($initial = '')
    {
        $this->value = (string)$initial;
    }

    public function get()
    {
        return $this->value;
    }

    /**
     * Применяет str_replace($search, $replace, $value) — НЕ regex.
     * Используется LanguageCompiler для escape кавычек.
     */
    public function replace($search, $replace)
    {
        $this->value = str_replace($search, $replace, $this->value);
        return $this;
    }

    /**
     * Применяет preg_replace($pattern, $replacement, $value).
     * Возвращает $this для chaining.
     */
    public function regEx($pattern, $replacement)
    {
        $result = preg_replace($pattern, $replacement, $this->value);
        if($result !== null)
        {
            $this->value = $result;
        }
        return $this;
    }

    /**
     * append: добавить в конец.
     */
    public function push($str)
    {
        $this->value .= (string)$str;
        return $this;
    }

    /**
     * prepend: добавить в начало (legacy именование Recipe — не array_pop).
     */
    public function pop($str)
    {
        $this->value = (string)$str.$this->value;
        return $this;
    }

    /**
     * String-cast (если кто-то echo'ит объект напрямую).
     */
    public function __toString()
    {
        return $this->value;
    }
}
