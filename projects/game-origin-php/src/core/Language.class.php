<?php
/**
 * Language — clean-room rewrite (план 43 Ф.6). Заменяет одноимённый
 * класс из фреймворка Recipe (GPL).
 *
 * i18n-словарь: загружает phrasegroups из na_phrasesgroups + na_phrases,
 * каждую фразу прогоняет через LanguageCompiler (compile-time замена
 * токенов {lang/user/config/...}). Поддерживает file-cache через
 * Core::getCache()->cacheLanguage / getLanguageCache.
 *
 * Public API:
 *   - new Language($langid, $groups)
 *   - load($groups): подгрузить дополнительные группы
 *   - getItem/get($var, &$tpl=null): локализованная фраза с подстановкой
 *     {@var} из $tpl (Template или массив) или $dynvars (assign).
 *   - getItemWith($var, $params): alias getItem с массивом params
 *   - setItem/set($var, $value): override фразы
 *   - assign($var, $value): задать значение для wildcard {@var}
 *   - getOpt($param): поле из na_languages для текущего langid
 *   - getOptions(): вся строка из na_languages
 *   - getDefaultLang(): резолв $langid через langcode/HTTP_ACCEPT_LANGUAGE
 *   - rebuild($groups=null): пересобрать cache.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Language extends Collection
{
    protected $langid;
    protected $langcode = '';
    protected $opts = array();
    protected $vars = array();
    protected $dynvars = array();
    protected $grouplist = array();
    protected $all_groups = array();
    protected $cacheActive = false;
    protected $autoCaching = true;
    protected $uncached = true;

    public function __construct($langid, $groups)
    {
        $this->langid = $langid;
        $this->getDefaultLang();
        $this->item = array();
        if($this->cacheActive && defined('CACHE_ACTIVE'))
        {
            $this->cacheActive = (bool)CACHE_ACTIVE;
        }
        $this->loadAllGroups();
        $this->load($groups);
    }

    /**
     * Резолвит langid через явный id, langcode (Accept-Language) либо
     * defaultlanguage из конфига. Заполняет $this->opts (строка из
     * na_languages).
     */
    public function getDefaultLang()
    {
        if($this->langid === '' || $this->langid === null
           || (is_string($this->langid) && strlen($this->langid) === 0))
        {
            $accept = isset($_SERVER['HTTP_ACCEPT_LANGUAGE']) ? $_SERVER['HTTP_ACCEPT_LANGUAGE'] : '';
            $this->langcode = substr($accept, 0, 2);
            $this->langid = 0;
        }
        elseif(!is_numeric($this->langid))
        {
            $this->langcode = (string)$this->langid;
            $this->langid = 0;
        }

        $this->opts = array();
        if($this->langid)
        {
            $row = sqlSelectRow('languages', '*', '', 'languageid = '.sqlVal($this->langid));
            if($row) { $this->opts = $row; }
        }
        if(empty($this->opts) && $this->langcode !== '')
        {
            $row = sqlSelectRow('languages', '*', '', 'langcode = '.sqlVal($this->langcode));
            if($row) { $this->opts = $row; }
        }
        if(empty($this->opts))
        {
            $defLang = Core::getOptions()->get('defaultlanguage');
            if($defLang !== null)
            {
                $row = sqlSelectRow('languages', '*', '', 'languageid = '.sqlVal($defLang));
                if($row) { $this->opts = $row; }
            }
        }
        if(!empty($this->opts) && isset($this->opts['languageid']))
        {
            $this->langid = $this->opts['languageid'];
            if(isset($this->opts['langcode']))
            {
                $this->langcode = $this->opts['langcode'];
            }
        }
        return $this;
    }

    /**
     * Возвращает список languageid для загрузки (default + текущий, если
     * текущий != default — для fallback-резолва при отсутствии перевода).
     */
    protected function getLoadLangs()
    {
        $defId = defined('DEF_LANGUAGE_ID') ? DEF_LANGUAGE_ID : 1;
        $langs = array($defId);
        if($this->langid != $defId)
        {
            $langs[] = $this->langid;
        }
        return $langs;
    }

    /**
     * Полный список phrasegroups из БД (заполняется один раз при init).
     * После этого load() может dispatch'ить отдельные группы по имени.
     */
    protected function loadAllGroups()
    {
        $this->all_groups = array();
        $res = Core::getDB()->query('SELECT phrasegroupid, title FROM `'.PREFIX.'phrasesgroups` ORDER BY phrasegroupid ASC');
        if($res)
        {
            while($row = Core::getDB()->fetch($res))
            {
                $this->all_groups[$row['title']] = array(
                    'id'     => (int)$row['phrasegroupid'],
                    'loaded' => false,
                );
            }
            Core::getDB()->free_result($res);
        }
    }

    /**
     * Грузит указанные группы (CSV или массив имён). Каждая группа
     * читается через cache (если активен) или напрямую из БД.
     */
    public function load($groups = array())
    {
        if(!is_array($groups))
        {
            $groups = array_map('trim', explode(',', (string)$groups));
        }
        $this->loadGroups($groups);
        return $this;
    }

    protected function loadGroups($groups)
    {
        if(!is_array($groups) || count($groups) === 0) { return; }
        foreach($groups as $group)
        {
            if(!isset($this->all_groups[$group])) { continue; }
            if($this->all_groups[$group]['loaded']) { continue; }
            $this->grouplist[] = $group;
            if($this->cacheActive)
            {
                $this->loadGroupFromCache($group);
            }
            else
            {
                $this->loadGroupFromDB((int)$this->all_groups[$group]['id']);
            }
            $this->all_groups[$group]['loaded'] = true;
        }
    }

    /**
     * Загрузка группы из file-cache через Core::getCache(). Если cache-файла
     * нет — Cache сам пересоберёт через cacheLanguage().
     */
    protected function loadGroupFromCache($group)
    {
        if(!isset($this->opts['langcode'])) { return; }
        $cached = Core::getCache()->getLanguageCache($this->opts['langcode'], array($group));
        if(is_array($cached))
        {
            foreach($cached as $title => $content)
            {
                $this->setItem($title, $content);
            }
        }
    }

    /**
     * Прямая загрузка из БД (без cache). Используется когда cacheActive = false.
     */
    protected function loadGroupFromDB($groupId)
    {
        $loadLangs = $this->getLoadLangs();
        foreach($loadLangs as $lang_id)
        {
            $res = Core::getDB()->query(
                'SELECT title, content FROM `'.PREFIX.'phrases` '
                .'WHERE phrasegroupid = '.(int)$groupId.' AND languageid = '.(int)$lang_id
            );
            if(!$res) { continue; }
            while($row = Core::getDB()->fetch($res))
            {
                $compiler = new LanguageCompiler($row['content'], true);
                $this->setItem($row['title'], $compiler->getPhrase());
                $compiler->shutdown();
            }
            Core::getDB()->free_result($res);
        }
    }

    /**
     * Возвращает локализованную фразу. Если фраза содержит wildcards
     * `{@key}` — они заменяются значениями из $tpl (Template или array)
     * или $this->dynvars (set через assign()).
     */
    public function getItem($var, &$tpl = null)
    {
        if($this->exists($var))
        {
            return $this->replaceWildCards($this->item[$var], $tpl);
        }
        // Auto-cache fallback: если cache активен и фраза не найдена —
        // пересобрать cache и попробовать снова. Один раз за запрос
        // (uncached-flag).
        if($this->cacheActive && $this->autoCaching && $this->uncached)
        {
            if(isset($this->opts['langcode']))
            {
                Core::getCache()->cacheLanguage($this->opts['langcode']);
            }
            $this->uncached = false;
            // Перечитать загруженные группы из обновлённого cache.
            foreach(array_keys($this->all_groups) as $g)
            {
                if($this->all_groups[$g]['loaded'])
                {
                    $this->all_groups[$g]['loaded'] = false;
                }
            }
            $reload = $this->grouplist;
            $this->grouplist = array();
            $this->loadGroups($reload);
            if($this->exists($var))
            {
                return $this->replaceWildCards($this->item[$var], $tpl);
            }
        }
        return $var;
    }

    public function getItemWith($var, $params)
    {
        return $this->getItem($var, $params);
    }

    public function get($var, &$tpl = null)
    {
        return $this->getItem($var, $tpl);
    }

    public function setItem($var, $value)
    {
        if(!$this->exists($var))
        {
            $this->vars[] = $var;
        }
        $this->item[$var] = $this->parse($value);
        return $this;
    }

    public function set($var, $value)
    {
        return $this->setItem($var, $value);
    }

    /**
     * Регистрирует значение для wildcard {@variable}. Может принимать
     * массив key=>value (тогда $value игнорируется) или одно имя.
     */
    public function assign($variable, $value = null)
    {
        if(is_array($variable))
        {
            foreach($variable as $key => $val)
            {
                if((string)$key !== '')
                {
                    $this->dynvars[$key] = $val;
                }
            }
        }
        elseif(is_string($variable) || is_numeric($variable))
        {
            $name = (string)$variable;
            if($name !== '')
            {
                $this->dynvars[$name] = $value;
            }
        }
        return $this;
    }

    /**
     * Замена wildcard'ов вида `{@key}` в content. Источник значения:
     *   1. $tpl[$key]  если $tpl — массив с ключом
     *   2. $tpl->get($key)  если $tpl — Template-like объект
     *   3. $this->dynvars[$key]  значение, заданное через assign()
     */
    protected function replaceWildCards($content, &$tpl = null)
    {
        $dyn = &$this->dynvars;
        return preg_replace_callback('/\{\@([^"]+)}/siU', function($m) use ($tpl, &$dyn) {
            $key = $m[1];
            if(is_array($tpl) && isset($tpl[$key])) { return $tpl[$key]; }
            if(isset($dyn[$key])) { return $dyn[$key]; }
            if(is_object($tpl) && method_exists($tpl, 'get'))
            {
                $val = $tpl->get($key);
                return $val !== null ? $val : '';
            }
            return '';
        }, (string)$content);
    }

    /**
     * Подготавливает строку для хранения (escape кавычек — наследие от
     * Recipe, где cache-файл собирался конкатенацией строк).
     */
    protected function parse($langvar)
    {
        return addcslashes((string)$langvar, '"\\');
    }

    public function getVarnames()
    {
        return $this->vars;
    }

    public function getVarCount()
    {
        return count($this->item);
    }

    public function getOptions()
    {
        return $this->opts;
    }

    public function getOpt($param)
    {
        return isset($this->opts[$param]) ? $this->opts[$param] : $param;
    }

    /**
     * Пересобирает cache для текущего языка (полностью или указанных групп).
     */
    public function rebuild($groups = null)
    {
        if(!isset($this->opts['langcode'])) { return $this; }
        if($groups !== null)
        {
            Core::getCache()->cachePhraseGroup($groups, $this->opts['langcode']);
        }
        else
        {
            Core::getCache()->cacheLanguage($this->opts['langcode']);
        }
        return $this;
    }
}
