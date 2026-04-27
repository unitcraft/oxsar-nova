<?php
/**
 * TemplateCompiler — clean-room rewrite (план 43 Ф.5). Заменяет
 * одноимённый класс из фреймворка Recipe (GPL).
 *
 * Один проход через .tpl-файл превращает шаблон в PHP-код:
 *   - {var=KEY}              → <?php echo $this->get("KEY", false); ?>
 *   - {lang=KEY}             → <?php echo Core::getLanguage()->get("KEY", $this); ?>
 *   - {config=KEY}           → <?php echo Core::getOptions()->get("KEY"); ?>
 *   - {user=KEY}             → <?php echo Core::getUser()->get("KEY"); ?>
 *   - {const=KEY}            → <?php echo KEY; ?>
 *   - {request[get]K{/request} → <?php echo Core::getRequest()->get("get","K"); ?>
 *   - {link[K]URL{/link}     → <?php echo Link::get(URL, getLanguage()->get("K", $this)); ?>
 *   - {image[K]URL{/image}   → <?php echo Image::getImage("URL", getLanguage()->getItem("K", $this)); ?>
 *   - {include=NAME}         → <?php $this->includeTemplate(NAME); ?>
 *   - {perm[X]}…{/perm}      → {if[Core::getUser()->ifPermissions("X")]}…{/if}
 *   - {if[expr]}…{else if[…]}…{else}…{/if}  → стандартные php-блоки
 *   - {while[loop]}…{/while}, {foreach[loop]}…{/foreach} (+ foreach2/3/4)
 *   - {loop=KEY}, {~count}, {@varname} — внутри loop
 *   - {hook=name}, {time=fmt}, {PHPTime}, {SQLQueries}, {DBTime}
 *
 * Compiled-файл сохраняется через Cache::putCacheContent с защитой от
 * прямого доступа в шапке.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class TemplateCompiler extends Cache
{
    protected $sourceTemplate = '';
    protected $template = '';
    protected $compiledTemplate = null;

    public function __construct($templatePath)
    {
        // Имя без расширения (для генерации cache-пути).
        $base = basename($templatePath);
        $dot = strrpos($base, '.');
        $this->template = $dot === false ? $base : substr($base, 0, $dot);

        $this->sourceTemplate = (string)@file_get_contents($templatePath);
        $this->compile();

        // Запись в cache через родительский метод Cache::writeFile,
        // обёрнутый в putCacheContent для совместимости с legacy callers.
        $this->putCacheContent(
            Core::getCache()->getTemplatePath($this->template),
            $this->compiledTemplate->get()
        );
    }

    /**
     * Совместимость с legacy: putCacheContent был protected у Cache.
     * Пробрасываем его вызов через writeFile.
     */
    protected function putCacheContent($file, $content)
    {
        $this->writeFile($file, $content);
        return $this;
    }

    /**
     * Полный compile-цикл. Применяется набор regex-замен в фиксированном
     * порядке (последовательность критична — некоторые токены содержат
     * другие как сабсет).
     */
    protected function compile()
    {
        $mod = 'siU';
        $tpl = new OxsarString($this->sourceTemplate);

        // BOM в UTF-8 — удаляем перед компиляцией.
        $tpl->regEx('#^'.preg_quote("\xEF\xBB\xBF", '#').'\s*#is', '');

        // Порядок: link/image (с двумя группами и URL внутри [...])
        // должны идти раньше остальных, чтобы их URL не попал под более
        // короткие токены вроде {const=…}.
        $tpl->regEx('/\{link\[([^"]+)]}(.*)\{\/link}/'.$mod,
                    '<?php echo Link::get(\\2, Core::getLanguage()->get("\\1", $this)); ?>');
        $tpl->regEx('/\{image\[([^"]+)]}([^"]+)\{\/image}/'.$mod,
                    '<?php echo Image::getImage("\\2", Core::getLanguage()->getItem("\\1", $this)); ?>');

        // var = template variable
        $tpl->regEx('/\{var}([^"]+)\{\/var}/'.$mod, '$this->get("$1", false)');
        $tpl->regEx('/\{var=([^"]+)\}/'.$mod, '$this->get("$1", false)');

        // lang = i18n phrase
        $tpl->regEx('/\{lang}([^"]+)\{\/lang}/'.$mod,
                    '<?php echo Core::getLanguage()->get("\\1", $this); ?>');
        $tpl->regEx('/\{lang=([^"]+)\}/'.$mod,
                    '<?php echo Core::getLanguage()->get("\\1", $this); ?>');

        // config = global option
        $tpl->regEx('/\{config}([^"]+)\{\/config}/'.$mod,
                    '<?php echo Core::getOptions()->get("\\1"); ?>');
        $tpl->regEx('/\{config=([^"]+)\}/'.$mod,
                    '<?php echo Core::getOptions()->get("\\1"); ?>');

        // user = user attribute
        $tpl->regEx('/\{user}([^"]+)\{\/user}/'.$mod,
                    '<?php echo Core::getUser()->get("\\1"); ?>');
        $tpl->regEx('/\{user=([^"]+)\}/'.$mod,
                    '<?php echo Core::getUser()->get("\\1"); ?>');

        // request[source]key
        $tpl->regEx('/\{request\[([^"]+)\]\}([^"]+)\{\/request\}/'.$mod,
                    '<?php echo Core::getRequest()->get("\\1", "\\2"); ?>');

        // const = PHP constant
        $tpl->regEx('/\{const}([^"]+)\{\/const}/'.$mod, '<?php echo \\1; ?>');
        $tpl->regEx('/\{const=([^"]+)\}/'.$mod, '<?php echo \\1; ?>');

        // perm — превращается в {if[…ifPermissions…]}…{/if}, сам {if}
        // компилируется ниже в compileIfTags.
        $tpl->regEx('/\{perm\[([^"]+)\]\}(.*)\{\/perm\}/'.$mod,
                    '{if[Core::getUser()->ifPermissions("\\1")]}\\2{/if}');

        // include sub-template
        $tpl->regEx('/\{include}(.*)\{\/include}/'.$mod,
                    '<?php $this->includeTemplate(\\1); ?>');
        $tpl->regEx('/\{include=(.*)\}/'.$mod,
                    '<?php $this->includeTemplate(\\1); ?>');

        // perf-хелперы
        $tpl->regEx('/\{PHPTime}/'.$mod, '<?php echo Core::getTimer()->getTime(); ?>');
        $tpl->regEx('/\{DBTime}/'.$mod, '<?php echo Core::getDB()->getQueryTime(); ?>');
        $tpl->regEx('/\{SQLQueries}/'.$mod, '<?php echo Core::getDB()->getQueryNumber(); ?>');

        // hook (всегда no-op в нашем порте, но синтаксис сохраняем).
        $tpl->regEx('/\{hook}([^"]+)\{\/hook}/'.$mod,
                    '<?php Hook::event("\\1", array($this)); ?>');
        $tpl->regEx('/\{hook=([^"]+)\}/'.$mod,
                    '<?php Hook::event("\\1", array($this)); ?>');

        // time = форматированный timestamp
        $tpl->regEx('/\{time}(.*)\{\/time}/'.$mod,
                    '<?php echo Date::timeToString(3, -1, "\\1", false); ?>');
        $tpl->regEx('/\{time=(.*)\}/'.$mod,
                    '<?php echo Date::timeToString(3, -1, "\\1", false); ?>');

        $this->compiledTemplate = $tpl;
        $this->compileIfTags();
        $this->compileLoops();

        // wildcard {@varname} → echo $this->get(...)
        $this->compiledTemplate->regEx('/\{\@([^"]+)}/'.$mod,
                    '<?php echo $this->get("\\1"); ?>');

        // Схлопывает соседние закрывающий+открывающий PHP-теги в один
        // блок (читаемость cache-файла + чуть быстрее на больших шаблонах).
        $this->compiledTemplate->regEx('/\?><\?php/'.$mod, '');

        // Обёртка cache-файла: PHP-шапка с защитой и финальный close.
        $this->compiledTemplate->pop($this->fileHeader('Template Cache File').'?>'."\r");
        $this->compiledTemplate->push("\r\r<?php // Cache-Generator finished ?>");

        return $this;
    }

    protected function compileIfTags()
    {
        $mod = 'siU';
        $tpl = $this->compiledTemplate;
        $tpl->regEx('/\{if\[(.*)]}/'.$mod, '<?php if($1) { ?>');
        $tpl->regEx('/\{\/if}/'.$mod, '<?php } ?>');
        $tpl->regEx('/\{else}/'.$mod, '<?php } else { ?>');
        $tpl->regEx('/\{else if\[(.*)]}/'.$mod, '<?php } else if($1) { ?>');
        return $this;
    }

    protected function compileLoops()
    {
        $mod = 'siU';
        $tpl = $this->compiledTemplate;

        // While loop — итерация по DB-result через Core::getDB()->fetch.
        $tpl->regEx('/\{while\[([^"]+)]}(.*)\{\/while}/'.$mod,
                    '<?php while($row = Core::getDB()->fetch($this->getLoop("$1"))){ ?> $2 <?php } ?>');

        // foreach[N] — итерация по массиву из loopStack или вложенному
        // (см. Template::getLoop). Уровень вложенности кодируется суффиксом
        // 2/3/4 — у каждого свой push на runTempStack для сохранения
        // внешних $count/$row.
        $foreachOuter =
            '<?php $cur_loop = $this->getLoop("$1"); $count = count($cur_loop); '
            .'foreach($cur_loop as $key => $row) { array_push($this->runValueStack, $row); ?> '
            .'$2 '
            .'<?php array_pop($this->runValueStack); } ?>';

        $foreachNested =
            '<?php array_push($this->runTempStack, array($count, $row)); '
            .'$cur_loop = $this->getLoop("$1"); $count = count($cur_loop); '
            .'foreach($cur_loop as $key => $row) { array_push($this->runValueStack, $row); ?> '
            .'$2 '
            .'<?php array_pop($this->runValueStack); } '
            .'list($count, $row) = array_pop($this->runTempStack); ?>';

        $tpl->regEx('/\{foreach\[([^"]+)\]}(.*)\{\/foreach}/'.$mod, $foreachOuter);
        $tpl->regEx('/\{foreach2\[(\.[^\.][^"]*)\]}(.*)\{\/foreach2}/'.$mod, $foreachNested);
        $tpl->regEx('/\{foreach3\[(\.\.[^\.][^"]*)\]}(.*)\{\/foreach3}/'.$mod, $foreachNested);
        $tpl->regEx('/\{foreach4\[(\.\.\.[^\.][^"]*)\]}(.*)\{\/foreach4}/'.$mod, $foreachNested);

        // Loop-variables.
        // {loop=.path}    — getLoopVar (по dotted path для вложенных).
        $tpl->regEx('/\{loop}(\.[^"]*)\{\/loop}/'.$mod, '<?php echo $this->getLoopVar("$1"); ?>');
        $tpl->regEx('/\{loop=(\.[^"]*)\}/'.$mod, '<?php echo $this->getLoopVar("$1"); ?>');
        // {loop=name}     — fast: прямое чтение из текущего $row.
        $tpl->regEx('/\{loop}([^\.][^"]*)\{\/loop}/'.$mod, '<?php echo $row["$1"]; ?>');
        $tpl->regEx('/\{loop=([^\.][^"]*)\}/'.$mod, '<?php echo $row["$1"]; ?>');

        // {~count} — общее количество элементов текущего loop.
        $tpl->regEx('/\{\~count}/'.$mod, '<?php echo $count; ?>');

        return $this;
    }
}
