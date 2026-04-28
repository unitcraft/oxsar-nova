<?php
/**
 * Cache — clean-room rewrite (план 43 Ф.5). Заменяет одноимённый класс
 * из фреймворка Recipe (GPL).
 *
 * Filesystem-based cache (PHP-файлы с массивами `$item[…]`, выполняются
 * через require). Структура:
 *   src/cache/
 *     language/lang.{code}.{group}.php   — phrasegroup → title → content
 *     templates/{package}/{name}.cache.php — compiled template
 *     sessions/session.{sid}.php          — user data snapshot
 *     permissions/permission.{group}.php  — permission → value
 *     {object}.cache.php                  — буферизованные SQL-результаты
 *
 * Caching strategy:
 *   - getXxxCache: читает существующий cache-файл; если его нет —
 *     buildXxxCache() сначала, потом read.
 *   - cacheLanguage: rebuild всех групп языка из БД.
 *   - cachePhraseGroup: rebuild только указанной группы (быстрее).
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Cache
{
    protected $cacheDir;
    protected $languageCacheDir;
    protected $templateCacheDir;
    protected $sessionCacheDir;
    protected $permissionCacheDir;
    protected $cacheFileClose = "\n\n// Cache-Generator finished\n?>";
    protected $cacheStack = array();

    public function __construct()
    {
        $this->setCacheDir(APP_ROOT_DIR.'cache/');
    }

    public function setCacheDir($dir)
    {
        $this->cacheDir = rtrim($dir, '/').'/';
        $this->languageCacheDir   = $this->cacheDir.'language/';
        $this->templateCacheDir   = $this->cacheDir.'templates/';
        $this->sessionCacheDir    = $this->cacheDir.'sessions/';
        $this->permissionCacheDir = $this->cacheDir.'permissions/';
        return $this;
    }

    /* ============================================================
     * Language cache
     * ============================================================ */

    /**
     * Загружает phrasegroup-файлы для $langcode и возвращает плоский
     * массив title => content. Если cache-файла группы нет — пересоздаёт
     * через cacheLanguage().
     */
    public function getLanguageCache($langcode, $groups)
    {
        if(!is_array($groups))
        {
            $groups = array_map('trim', explode(',', (string)$groups));
        }
        $out = array();
        foreach($groups as $group)
        {
            if($group === '') { continue; }
            $cacheFile = $this->languageCacheDir.'lang.'.$langcode.'.'.$group.'.php';
            if(!is_file($cacheFile))
            {
                $this->cacheLanguage($langcode);
                if(!is_file($cacheFile)) { continue; }
            }
            if(in_array($cacheFile, $this->cacheStack, true)) { continue; }
            $item = array();
            require($cacheFile);
            $this->cacheStack[] = $cacheFile;
            if(isset($item[$group]) && is_array($item[$group]))
            {
                foreach($item[$group] as $key => $value)
                {
                    $out[$key] = $value;
                }
            }
        }
        return $out;
    }

    /**
     * Полная пересборка cache всех phrasegroup для указанного $langcode.
     */
    public function cacheLanguage($langcode)
    {
        $result = Core::getQuery()->select('phrasesgroups',
            array('title AS grouptitle', 'phrasegroupid'));
        if(!$result) { return $this; }
        while($row = Core::getDB()->fetch($result))
        {
            $this->writeLanguageGroup($langcode, $row['grouptitle'], (int)$row['phrasegroupid']);
        }
        Core::getDB()->free_result($result);
        return $this;
    }

    /**
     * Пересборка cache только указанных групп (одна или CSV).
     */
    public function cachePhraseGroup($groupname, $langcode)
    {
        if(!is_array($groupname))
        {
            $groupname = array_map('trim', explode(',', (string)$groupname));
        }
        foreach($groupname as $group)
        {
            if($group === '') { continue; }
            $gid = sqlSelectField('phrasesgroups', 'phrasegroupid', '', 'title = '.sqlVal($group));
            if($gid === null) { continue; }
            // Удаляем старый cache (он мог содержать переводы устаревших фраз).
            $filename = $this->languageCacheDir.'lang.'.$langcode.'.'.$group.'.php';
            if(is_file($filename)) { @unlink($filename); }
            $this->writeLanguageGroup($langcode, $group, (int)$gid);
        }
        return $this;
    }

    /**
     * Пишет cache-файл одной группы. Содержимое каждой фразы прогоняется
     * через LanguageCompiler перед сохранением.
     */
    private function writeLanguageGroup($langcode, $groupTitle, $groupId)
    {
        $body = $this->fileHeader('Language ['.$langcode.'] Cache File');
        $body .= '//### Variables for phrase group "'.$groupTitle.'" ###'."\n";
        $joins = 'LEFT JOIN '.PREFIX.'languages l ON (l.languageid = p.languageid)';
        $where = 'l.langcode = '.sqlVal($langcode).' AND p.phrasegroupid = '.sqlVal($groupId);
        $res = Core::getQuery()->select('phrases p',
            array('p.title AS phrasetitle', 'p.content'),
            $joins, $where, 'p.phrasegroupid ASC, p.title ASC');
        if($res)
        {
            while($col = Core::getDB()->fetch($res))
            {
                $compiler = new LanguageCompiler($col['content']);
                $compiled = $compiler->getPhrase();
                $compiler->shutdown();
                $body .= '$item["'.$groupTitle.'"]["'.$col['phrasetitle'].'"]="'.$compiled.'";';
            }
            Core::getDB()->free_result($res);
        }
        $body .= $this->cacheFileClose;
        $this->writeFile($this->languageCacheDir.'lang.'.$langcode.'.'.$groupTitle.'.php', $body);
    }

    /* ============================================================
     * Config cache
     * ============================================================ */

    public function getConfigCache()
    {
        $cacheFile = $this->cacheDir.'options.cache.php';
        if(!is_file($cacheFile))
        {
            $this->buildConfigCache();
        }
        $item = array();
        if(is_file($cacheFile))
        {
            require($cacheFile);
        }
        return $item;
    }

    public function buildConfigCache()
    {
        $body = $this->fileHeader('Global Configuration Variables & Options');
        $body .= '$item = array(';
        $result = Core::getQuery()->select('config', array('var', 'value'));
        if($result)
        {
            while($row = Core::getDB()->fetch($result))
            {
                $body .= '"'.$row['var'].'"=>"'.$this->escapeQuotes($row['value']).'",';
            }
            Core::getDB()->free_result($result);
        }
        $body .= ");\n".$this->cacheFileClose;
        $this->writeFile($this->cacheDir.'options.cache.php', $body);
        return $this;
    }

    /* ============================================================
     * Permission cache
     * ============================================================ */

    public function getPermissionCache($groupid)
    {
        $cacheFile = $this->permissionCacheDir.'permission.'.$groupid.'.php';
        if(!is_file($cacheFile))
        {
            $this->buildPermissionCache($groupid);
        }
        $item = array();
        if(is_file($cacheFile) && !in_array($cacheFile, $this->cacheStack, true))
        {
            $this->cacheStack[] = $cacheFile;
            require($cacheFile);
        }
        return $item;
    }

    /**
     * Пересборка permission-cache для $groupid (либо для всех групп
     * если $groupid === null).
     */
    public function buildPermissionCache($groupid = null)
    {
        if($groupid === null)
        {
            $result = Core::getQuery()->select('usergroup',
                array('usergroupid', 'grouptitle'));
            if(!$result) { return $this; }
            while($row = Core::getDB()->fetch($result))
            {
                $this->writePermissionGroup((int)$row['usergroupid'], $row['grouptitle']);
            }
            Core::getDB()->free_result($result);
            return $this;
        }
        if(is_array($groupid))
        {
            foreach($groupid as $g) { $this->writePermissionGroup((int)$g, ''); }
            return $this;
        }
        $this->writePermissionGroup((int)$groupid, '');
        return $this;
    }

    private function writePermissionGroup($groupId, $groupTitle)
    {
        $joins = 'LEFT JOIN '.PREFIX.'permissions p ON (p.permissionid = g2p.permissionid) '
            .'LEFT JOIN '.PREFIX.'usergroup g ON (g.usergroupid = g2p.groupid)';
        $result = Core::getQuery()->select('group2permission g2p',
            array('g2p.value', 'p.permission', 'g.grouptitle'),
            $joins, 'g2p.groupid = '.sqlVal($groupId));
        $body = '';
        if($result)
        {
            while($row = Core::getDB()->fetch($result))
            {
                $body .= '$item["'.$row['permission'].'"]='.(int)$row['value'].';';
                if($groupTitle === '' && !empty($row['grouptitle']))
                {
                    $groupTitle = $row['grouptitle'];
                }
            }
            Core::getDB()->free_result($result);
        }
        if($groupTitle === '') { $groupTitle = (string)$groupId; }
        $body = $this->fileHeader('Permissions ['.$groupTitle.']').$body.$this->cacheFileClose;
        $this->writeFile($this->permissionCacheDir.'permission.'.$groupId.'.php', $body);
    }

    /* ============================================================
     * User session cache
     * ============================================================ */

    public function getUserCache($sid)
    {
        $cacheFile = $this->sessionCacheDir.'session.'.$sid.'.php';
        if(!is_file($cacheFile))
        {
            return array();
        }
        $item = array();
        require($cacheFile);
        return $item;
    }

    public function buildUserCache($sid)
    {
        // В текущем порте сессия хранится не в na_sessions (legacy auth убран,
        // используется JWT через game-nova/auth). Метод оставлен как stub —
        // вернёт пустой $item, getUserCache() выдаст empty array, и User
        // упадёт на загрузку из na_user напрямую (см. User::loadData).
        $body = $this->fileHeader('Session Cache ['.$sid.']');
        $body .= '$item = array();'."\n";
        $body .= $this->cacheFileClose;
        $this->writeFile($this->sessionCacheDir.'session.'.$sid.'.php', $body);
        return $this;
    }

    /**
     * Удаляет cache всех сессий пользователя (например, при logout
     * или при ban).
     */
    public function cleanUserCache($userid)
    {
        $result = Core::getQuery()->select('sessions', 'sessionid', '',
            'userid = '.sqlVal($userid));
        if(!$result) { return $this; }
        while($row = Core::getDB()->fetch($result))
        {
            $f = $this->sessionCacheDir.'session.'.$row['sessionid'].'.php';
            if(is_file($f)) { @unlink($f); }
        }
        Core::getDB()->free_result($result);
        return $this;
    }

    /* ============================================================
     * Template path
     * ============================================================ */

    public function getTemplatePath($template)
    {
        $package = '';
        if(class_exists('Core'))
        {
            $tpl = Core::getTemplate();
            if($tpl) { $package = $tpl->getTemplatePackage(); }
        }
        $dir = $this->templateCacheDir.$package;
        if(!is_dir($dir)) { @mkdir($dir, 0777, true); }
        return $dir.$template.'.cache.php';
    }

    /* ============================================================
     * Generic object cache
     * ============================================================ */

    public function objectExists($name)
    {
        return is_file($this->cacheDir.$name.'.cache.php');
    }

    public function buildObject($name, $result, $index = null, $itype = 'int')
    {
        $data = array();
        while($row = Core::getDB()->fetch($result))
        {
            if($index === null)
            {
                $data[] = $row;
            }
            else
            {
                $key = ($itype === 'int' || $itype === 'integer')
                    ? (int)$row[$index]
                    : (string)$row[$index];
                $data[$key][] = $row;
            }
        }
        if(count($data) > 0)
        {
            $body = $this->fileHeader('Cache Object ['.$name.']');
            $body .= '$data = "'.$this->escapeQuotes(serialize($data)).'";'."\n";
            $body .= $this->cacheFileClose;
            $this->writeFile($this->cacheDir.$name.'.cache.php', $body);
        }
        return $this;
    }

    public function readObject($name)
    {
        $file = $this->cacheDir.$name.'.cache.php';
        if(!is_file($file))
        {
            throw new GenericException('Cache object not found: '.$name);
        }
        if(!in_array($file, $this->cacheStack, true))
        {
            $this->cacheStack[] = $file;
        }
        $data = '';
        require($file);
        return unserialize($data);
    }

    public function flushObject($name)
    {
        $file = $this->cacheDir.$name.'.cache.php';
        if(!is_file($file)) { return false; }
        return @unlink($file);
    }

    /* ============================================================
     * Internal helpers
     * ============================================================ */

    /**
     * PHP-шапка cache-файла с защитой от прямого доступа.
     */
    protected function fileHeader($title)
    {
        return "<?php\n"
            ."// Auto-generated cache: ".$title."\n"
            ."// Generated on ".date('Y-m-d H:i:s')."\n"
            ."if(!defined('RECIPE_ROOT_DIR')) { die('Hacking attempt detected.'); }\n\n";
    }

    /**
     * Запись cache-файла с авто-созданием родительского каталога
     * (включая цепочку поддиректорий).
     */
    protected function writeFile($file, $content)
    {
        $dir = dirname($file);
        if(!is_dir($dir))
        {
            @mkdir($dir, 0777, true);
        }
        if(!is_writable($dir))
        {
            throw new GenericException('Cache dir not writable: '.$dir);
        }
        $fp = @fopen($file, 'w');
        if($fp === false)
        {
            throw new GenericException('Cache write failed: '.$file);
        }
        fwrite($fp, $content);
        fclose($fp);
    }

    /**
     * Эскейп кавычек для содержимого, которое будет помещено внутрь
     * PHP double-quoted string.
     */
    protected function escapeQuotes($s)
    {
        return str_replace('"', '\\"', (string)$s);
    }
}
