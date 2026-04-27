<?php
/**
 * LanguageCompiler — clean-room rewrite (план 43 Ф.6). Заменяет
 * одноимённый класс из фреймворка Recipe (GPL).
 *
 * Прекомпилирует фразу из языкового словаря: токены `{lang…}`,
 * `{config=…}`, `{user=…}`, `{const=…}`, `{image[alt]url}` и др.
 * заменяются на PHP-вызовы. Результат сохраняется в кэше как PHP-код,
 * выполняется через eval/include при выводе.
 *
 * API:
 *   - new LanguageCompiler($phrase, $hardReplace = false): __construct
 *   - getPhrase(): string — получить compiled-форму.
 *   - setPhrase($phrase): self
 *   - shutdown(): void — no-op (PHP 8 запрещает unset($this)).
 *
 * $hardReplace различает два режима генерации:
 *   true  — фраза будет inline'ена в `<?php echo …; ?>` (нужно отдавать
 *           готовое PHP-выражение типа `Link::get("…", "…")`);
 *   false — фраза будет в обычной строке с конкатенацией
 *           (`"… " . Link::get("…", "…") . " …"`).
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class LanguageCompiler
{
    /** @var OxsarString */
    private $phrase;
    /** @var bool */
    private $hardReplace;

    public function __construct($phrase, $hardReplace = false)
    {
        $this->hardReplace = (bool)$hardReplace;
        $this->setPhrase($phrase);
    }

    public function setPhrase($phrase)
    {
        if($phrase instanceof OxsarString)
        {
            $this->phrase = $phrase;
        }
        else
        {
            $this->phrase = new OxsarString((string)$phrase);
        }
        $this->compile();
        return $this;
    }

    public function getPhrase()
    {
        return $this->phrase->get();
    }

    public function shutdown()
    {
        // No-op (PHP 8 не позволяет unset($this); вызывающий сам сбросит ссылку).
    }

    /**
     * Применяет regex-замены ко всем поддерживаемым токенам.
     * Сохраняем точно тот же набор паттернов и replacement-strings что
     * legacy-LanguageCompiler — это контракт с TemplateCompiler и
     * compiled-кэшем шаблонов.
     */
    private function compile()
    {
        $modifier = 'siU';
        $hard = $this->hardReplace;

        // (search => replacement) — порядок важен (link с двумя группами
        // должен идти раньше других, чтобы внутренние не схлопнулись).
        $rules = array(
            // {link[url]text{/link} → Link::get("url","text")
            '/\{link\[(.*)]}(.*)\{\/link}/'.$modifier
                => $hard ? 'Link::get("\\2", "\\1")' : '".Link::get("\\2", "\\1")."',
            // {config}KEY{/config} и {config=KEY}
            '/\{config}([^"]+)\{\/config}/'.$modifier
                => $hard ? 'Core::getOptions()->get("\\1")' : '".Core::getOptions()->get("\\1")."',
            '/\{config=([^"]+)\}/'.$modifier
                => $hard ? 'Core::getOptions()->get("\\1")' : '".Core::getOptions()->get("\\1")."',
            // {user}KEY{/user} и {user=KEY}
            '/\{user}([^"]+)\{\/user}/'.$modifier
                => $hard ? 'Core::getUser()->get("\\1")' : '".Core::getUser()->get("\\1")."',
            '/\{user=([^"]+)\}/'.$modifier
                => $hard ? 'Core::getUser()->get("\\1")' : '".Core::getUser()->get("\\1")."',
            // {request[get]key{/request} → Core::getRequest()->get["key"] (legacy syntax)
            '/\{request\[([^"]+)\]\}([^"]+)\{\/request\}/'.$modifier
                => $hard ? 'Core::getRequest()->\\1["\\2"]' : '".Core::getRequest()->\\1["\\2"]."',
            // {const}NAME{/const} и {const=NAME}
            '/\{const}([^"]+)\{\/const}/'.$modifier
                => $hard ? 'constant("\\1")' : '".\\1."',
            '/\{const=([^"]+)\}/'.$modifier
                => $hard ? 'constant("\\1")' : '".\\1."',
            // {image[alt]url{/image}
            '/\{image\[([^"]+)]}([^"]+)\{\/image}/'.$modifier
                => $hard ? 'Image::getImage("\\2", "\\1")' : '".Image::getImage("\\2", "\\1")."',
            // {time}FORMAT{/time} и {time=FORMAT}
            '/\{time}(.*)\{\/time}/'.$modifier
                => $hard ? 'Date::timeToString(3, -1, "\\1", false)' : '".Date::timeToString(3, -1, "\\1", false)."',
            '/\{time=(.*)\}/'.$modifier
                => $hard ? 'Date::timeToString(3, -1, "\\1", false)' : '".Date::timeToString(3, -1, "\\1", false)."',
        );

        // Escape кавычек в исходной фразе (она будет внутри php-строки).
        $this->phrase->replace('"', '\\"');

        foreach($rules as $pattern => $replacement)
        {
            $this->phrase->regEx($pattern, $replacement);
        }
    }
}
