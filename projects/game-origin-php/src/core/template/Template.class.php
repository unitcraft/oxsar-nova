<?php
/**
 * Template — clean-room rewrite (план 43 Ф.5). Заменяет одноимённый
 * класс из фреймворка Recipe (GPL).
 *
 * Шаблонизатор оригинального синтаксиса (.tpl):
 *   - {var=KEY}, {lang=KEY}, {config=KEY}, {user=KEY}, {const=KEY},
 *     {request[get]KEY{/request}, {include=NAME}, {image[alt]URL{/image},
 *     {link[KEY]URL{/link}.
 *   - {if[cond]}…{else if[cond]}…{else}…{/if}
 *   - {while[loop]}…{/while}, {foreach[loop]}…{/foreach}, {foreach2/3/4}.
 *   - {loop=KEY}, {~count} — внутри loop'а.
 *   - {@assignment} — переменная из templateVars.
 *
 * Компиляция в PHP-код через TemplateCompiler, результат кешируется в
 * src/cache/templates/<package>/<name>.cache.php. Перекомпиляция —
 * при изменении src/templates/<package>/<name>.tpl (mtime).
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

require_once(APP_ROOT_DIR.'game/Functions.inc.php');

class Template
{
    protected $templateExtension = '';
    protected $mainTemplateFile = '';
    protected $templatePath = '';
    protected $templatePath2 = '';
    protected $templatePackage = '';
    protected $templateVars = array();
    protected $forceCompilation = false;

    /** @var array Логические массивы данных, доступные через addLoop()/getLoop(). */
    protected $loopStack = array();

    /** Стек текущих значений foreach-итераций (используется в compiled-шаблонах). */
    protected $runValueStack = array();
    /** Стек состояний при вложенных foreach (foreach2/3/4 — для glamour-стиля legacy). */
    protected $runTempStack = array();

    /** @var Map */
    protected $log = null;
    /** @var Map */
    protected $htmlHead = null;
    protected $headKeys = array();

    protected $compressed = false;
    protected $headersSent = false;

    protected $view = null;

    public function __construct()
    {
        // План 37.5b: ext/templates/ удалён, основные шаблоны в src/templates/.
        $this->setTemplatePath(APP_ROOT_DIR.'ext/templates/');
        $this->setTemplatePath2(APP_ROOT_DIR.'templates/');
        $this->setTemplatePackage();
        $this->setLayoutTemplate();
        $this->setExtension();
        $this->log = new Map();
        $this->htmlHead = new Map();
        $this->assign('LOG', '');
        $this->assign('HTML_HEAD', '');
    }

    /* ============================================================
     * Variable assignment
     * ============================================================ */

    public function assign($variable, $value = null)
    {
        if(is_array($variable))
        {
            foreach($variable as $k => $v)
            {
                if((string)$k !== '') { $this->assign($k, $v); }
            }
            return $this;
        }
        if((is_string($variable) || is_numeric($variable)) && (string)$variable !== '')
        {
            $this->templateVars[$variable] = $value;
        }
        return $this;
    }

    public function deallocateAssignment($variable)
    {
        if(is_array($variable))
        {
            foreach($variable as $v) { $this->deallocateAssignment($v); }
            return $this;
        }
        unset($this->templateVars[$variable]);
        return $this;
    }

    public function deallocateAllAssignment()
    {
        $this->templateVars = array();
        return $this;
    }

    public function get($var, $default = null)
    {
        if($default === '[var]') { $default = $var; }
        return isset($this->templateVars[$var]) ? $this->templateVars[$var] : $default;
    }

    /* ============================================================
     * Loop API (used by compiled .tpl)
     * ============================================================ */

    public function addLoop($loop, $data)
    {
        $this->loopStack[$loop] = $data;
        return $this;
    }

    public function getLoop($loop)
    {
        if($loop === '' || $loop[0] !== '.')
        {
            return isset($this->loopStack[$loop]) ? $this->loopStack[$loop] : array();
        }
        // Dotted form: '.', '..', '...' — обращение к вложенным массивам
        // через runValueStack (вложенные foreach'и).
        $parts = explode('.', $loop);
        $i = count($parts) - 1;
        if(isset($this->runValueStack[$i - 1][$parts[$i]]))
        {
            return $this->runValueStack[$i - 1][$parts[$i]];
        }
        return array();
    }

    public function getLoopVar($name)
    {
        if($name === '' || $name[0] !== '.')
        {
            $i = count($this->runValueStack) - 1;
            return isset($this->runValueStack[$i][$name]) ? $this->runValueStack[$i][$name] : null;
        }
        $parts = explode('.', $name);
        $i = count($parts) - 1;
        return isset($this->runValueStack[$i - 1][$parts[$i]]) ? $this->runValueStack[$i - 1][$parts[$i]] : null;
    }

    public function getLoopRow($name = '')
    {
        if($name === '' || $name[0] !== '.')
        {
            $i = count($this->runValueStack) - 1;
            return isset($this->runValueStack[$i]) ? $this->runValueStack[$i] : array();
        }
        $parts = explode('.', $name);
        $i = count($parts) - 1;
        return isset($this->runValueStack[$i - 1]) ? $this->runValueStack[$i - 1] : array();
    }

    /* ============================================================
     * Display / Compile
     * ============================================================ */

    /**
     * Рендерит указанный шаблон.
     * $sendOnlyContent = true — не подключать layout, отдать только сам шаблон
     *   (для AJAX-ответов, includeTemplate).
     * $mainTemplate — переопределение layout-шаблона.
     */
    public function display($template, $sendOnlyContent = false, $mainTemplate = null, $use_once = true)
    {
        // Юзерские стили (фон, расцветка таблиц) — вытащить из User в HTML head.
        if(!$sendOnlyContent && class_exists('Core'))
        {
            $user = Core::getUser();
            if($user && !$user->isGuest() && function_exists('getUserStyles'))
            {
                foreach(array('bg', 'table') as $type)
                {
                    $styleId = $user->getRaw('user_'.$type.'_style');
                    if(!$styleId) { continue; }
                    $styles = getUserStyles($type);
                    if(isset($styles[$styleId]))
                    {
                        $ver = defined('CLIENT_VERSION') ? CLIENT_VERSION : '1';
                        $this->addHTMLHeaderFile($styles[$styleId]['path'].'?'.$ver, 'css');
                    }
                }
            }
        }

        // Накопленные сообщения и HTML-head попадают в template-vars.
        if($this->log->size() > 0)
        {
            $this->assign('LOG', $this->log->toString("\n"));
        }
        if($this->htmlHead->size() > 0)
        {
            $this->assign('HTML_HEAD', $this->htmlHead->toString("\n"));
        }

        $this->sendHeader();

        // Layout (внешний template) рендерится первым; внутри он
        // вызывает $this->includeTemplate($contentTemplate).
        if(!$sendOnlyContent || $mainTemplate !== null)
        {
            if($mainTemplate === null) { $mainTemplate = $this->mainTemplateFile; }
            if(!$this->cachedTemplateAvailable($mainTemplate) || $this->forceCompilation)
            {
                new TemplateCompiler($this->getTemplatePath($mainTemplate));
            }
            ob_start();
            require_once(Core::getCache()->getTemplatePath($mainTemplate));
            $this->processOutput(ob_get_clean());
        }

        // Содержательный template.
        if(!$this->cachedTemplateAvailable($template) || $this->forceCompilation)
        {
            new TemplateCompiler($this->getTemplatePath($template));
        }
        $compiled = Core::getCache()->getTemplatePath($template);
        ob_start();
        if($use_once)
        {
            require_once($compiled);
        }
        else
        {
            require($compiled);
        }
        $this->processOutput(ob_get_clean());
        return $this;
    }

    /**
     * Лёгкая нормализация HTML: схлопнуть подряд идущие пробелы и
     * пустые строки. Семантика не меняется, размер ответа уменьшается
     * на ~10-20%.
     */
    public function processOutput($output)
    {
        $output = preg_replace('#[ \t]+#', ' ', $output);
        $output = preg_replace('#[\r\n]+ +[\r\n]+|[\r\n]+ +| +[\r\n]+#', "\n", $output);
        $output = preg_replace('#>\s+<#', '> <', $output);
        echo $output;
    }

    public function includeTemplate($template)
    {
        $this->display($template, true, null, false);
        return $this;
    }

    /* ============================================================
     * Cache freshness check
     * ============================================================ */

    /**
     * Проверяет: cache-файл существует И новее исходного .tpl.
     * Иначе — нужна перекомпиляция.
     */
    protected function cachedTemplateAvailable($template)
    {
        $cached = Core::getCache()->getTemplatePath($template);
        $source = $this->getTemplatePath($template);
        if(!is_file($source))
        {
            throw new GenericException($source.' does not exist.');
        }
        if(!$this->forceCompilation && is_file($cached))
        {
            if(filemtime($source) <= filemtime($cached))
            {
                return true;
            }
        }
        return false;
    }

    /**
     * Резолв путей к .tpl: сначала templatePath/<package>/, затем
     * templatePath2/<package>/, затем templatePath/standard/, затем
     * templatePath2/standard/. Возвращает первый существующий путь либо
     * последний кандидат (caller сам решит что делать).
     */
    protected function getTemplatePath($template)
    {
        $pkg = $this->getTemplatePackage();
        $candidates = array();
        $candidates[] = $this->templatePath.$pkg.$template.$this->templateExtension;
        if($this->templatePath2)
        {
            $candidates[] = $this->templatePath2.$pkg.$template.$this->templateExtension;
            $candidates[] = $this->templatePath.'standard/'.$template.$this->templateExtension;
            $candidates[] = $this->templatePath2.'standard/'.$template.$this->templateExtension;
        }
        else
        {
            $candidates[] = $this->templatePath.'standard/'.$template.$this->templateExtension;
        }
        foreach($candidates as $f)
        {
            if(is_file($f)) { return $f; }
        }
        return end($candidates);
    }

    /* ============================================================
     * HTTP header
     * ============================================================ */

    protected function sendHeader()
    {
        if(!headers_sent() && !$this->headersSent)
        {
            if(@extension_loaded('zlib') && !$this->compressed && defined('GZIP_ACITVATED') && GZIP_ACITVATED)
            {
                ob_start('ob_gzhandler');
                $this->compressed = true;
            }
            $charset = 'utf-8';
            if(class_exists('Core'))
            {
                $lang = Core::getLanguage();
                if($lang)
                {
                    $c = $lang->getOpt('charset');
                    if($c && $c !== 'charset') { $charset = $c; }
                }
            }
            @header('Content-Type: text/html; charset='.$charset);
            $this->headersSent = true;
        }
        return $this;
    }

    /* ============================================================
     * Path/extension/package setters
     * ============================================================ */

    public function setTemplatePath($path)
    {
        $this->templatePath = rtrim((string)$path, '/').'/';
        return $this;
    }

    public function setTemplatePath2($path)
    {
        $this->templatePath2 = rtrim((string)$path, '/').'/';
        return $this;
    }

    public function setLayoutTemplate($template = null)
    {
        if($template === null && class_exists('Core'))
        {
            $cfg = Core::getConfig();
            $val = $cfg ? $cfg->get('maintemplate') : null;
            $template = ($val && $val !== 'maintemplate') ? $val : 'layout';
        }
        if($template === null) { $template = 'layout'; }
        $this->mainTemplateFile = $template;
        return $this;
    }

    public function setExtension($extension = null)
    {
        if($extension === null && class_exists('Core'))
        {
            $cfg = Core::getConfig();
            $val = $cfg ? $cfg->get('templateextension') : null;
            $extension = ($val && $val !== 'templateextension') ? $val : 'tpl';
        }
        if($extension === null) { $extension = 'tpl'; }
        if(substr($extension, 0, 1) !== '.') { $extension = '.'.$extension; }
        $this->templateExtension = $extension;
        return $this;
    }

    public function setTemplatePackage($package = null)
    {
        if($package === null && class_exists('Core'))
        {
            $cfg = Core::getConfig();
            $val = $cfg ? $cfg->get('templatepackage') : null;
            $package = ($val && $val !== 'templatepackage') ? $val : 'standard';
        }
        if($package === null) { $package = 'standard'; }
        if(substr($package, -1) !== '/') { $package .= '/'; }
        $this->templatePackage = $package;
        return $this;
    }

    public function getTemplatePackage()
    {
        if(is_dir($this->templatePath.$this->templatePackage))
        {
            return $this->templatePackage;
        }
        return 'standard/';
    }

    /* ============================================================
     * Log + HTML head
     * ============================================================ */

    public function addLogMessage($message)
    {
        $this->log->push($message);
        return $this;
    }

    /**
     * Регистрирует CSS/JS-файл в HTML-head. $type ∈ {css, js}. Защищён
     * от дубликатов по имени файла.
     */
    public function addHTMLHeaderFile($file, $type = 'js')
    {
        $type = strtolower($type);
        if(isset($this->headKeys[$type][$file])) { return $this; }
        $this->headKeys[$type][$file] = 1;

        $rel = defined('RELATIVE_URL') ? RELATIVE_URL : '/';
        if($type === 'css')
        {
            $url = $rel.'css/'.$file;
            $head = '<link rel="stylesheet" type="text/css" href="'.$url.'" media="screen" />';
        }
        else
        {
            $url = $rel.'js/'.$file;
            $head = '<script type="text/javascript" src="'.$url.'"></script>';
        }
        $this->htmlHead->push($head);
        return $this;
    }

    /* ============================================================
     * View
     * ============================================================ */

    public function setView($view)
    {
        if(is_object($view)) { $this->view = $view; }
        return $this;
    }

    public function getView()
    {
        return $this->view;
    }
}
