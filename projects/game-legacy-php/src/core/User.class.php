<?php
/**
 * User — clean-room rewrite (план 43 Ф.6). Заменяет одноимённый класс
 * из фреймворка Recipe (GPL).
 *
 * Сохранён весь публичный API + XSS-эскейп user-controlled полей
 * (план 37.7.1) при чтении через get(). Для raw-доступа — getRaw().
 *
 * Загрузка из na_user JOIN na_user2ally JOIN na_alliance по
 * $_SESSION['userid'] (без legacy session-таблицы — сейчас auth через JWT).
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class User extends Collection
{
    protected $sid = '';
    protected $permissions = array();
    protected $groups = array();
    protected $groupdata = array();
    protected $isGuest = false;
    protected $cacheActive = false;
    protected $guestGroupId = 1;

    /** План 37.7.1: XSS escape для user-controlled полей при get(). */
    private static $userInputFields = array('username', 'email', 'temp_email');

    /**
     * Карта offset_hours → tz_id (для setDefaultTimeZone). В legacy этот
     * массив был длиннее, оставлены только реально нужные значения,
     * остальное обрабатывается PHP date_default_timezone_set напрямую.
     */
    protected $timezones = array(
        '-12'  => 'Pacific/Wake',
        '-11'  => 'Pacific/Samoa',
        '-10'  => 'Pacific/Honolulu',
        '-9'   => 'America/Anchorage',
        '-8'   => 'America/Los_Angeles',
        '-7'   => 'America/Denver',
        '-6'   => 'America/Chicago',
        '-5'   => 'America/New_York',
        '-4'   => 'America/Asuncion',
        '-3'   => 'America/Argentina/Buenos_Aires',
        '-2'   => 'Atlantic/South_Georgia',
        '-1'   => 'Atlantic/Cape_Verde',
        '0'    => 'Europe/London',
        '1'    => 'Europe/Paris',
        '2'    => 'Europe/Istanbul',
        '3'    => 'Europe/Moscow',
        '4'    => 'Asia/Baku',
        '5'    => 'Asia/Bishkek',
        '6'    => 'Asia/Dhaka',
        '7'    => 'Asia/Bangkok',
        '8'    => 'Asia/Singapore',
        '9'    => 'Asia/Tokyo',
        '10'   => 'Australia/Queensland',
        '11'   => 'Pacific/Noumea',
        '12'   => 'Pacific/Fiji',
    );

    public function __construct($sid)
    {
        $this->sid = !empty($_SESSION['userid']) ? ($_SESSION['sid'] ?? '') : (string)$sid;
        $this->setGuestGroupId();
        if($this->cacheActive && defined('CACHE_ACTIVE'))
        {
            $this->cacheActive = (bool)CACHE_ACTIVE;
        }
        $this->loadData();
        if(!$this->isGuest)
        {
            $this->loadGroups();
        }
        $this->loadPermissions();
        $this->applyTimezone();
    }

    /**
     * Загружает строку из na_user (плюс ally-данные через JOIN). Если
     * userid в сессии нет — переход в guest-режим.
     */
    protected function loadData()
    {
        if(!empty($_SESSION['userid']))
        {
            $uid = (int)$_SESSION['userid'];
            $sql = 'SELECT u.*, a.aid AS allyid, a.name AS ally_name, a.tag AS ally_tag '
                .'FROM `'.PREFIX.'user` u '
                .'LEFT JOIN `'.PREFIX.'user2ally` ua ON ua.userid = u.userid '
                .'LEFT JOIN `'.PREFIX.'alliance` a ON a.aid = ua.aid '
                .'WHERE u.userid = '.$uid.' LIMIT 1';
            $result = Core::getDB()->query($sql);
            if($result)
            {
                $row = Core::getDB()->fetch($result);
                Core::getDB()->free_result($result);
                if($row)
                {
                    $this->item = $row;
                    $this->item['banned'] = false;
                }
            }
        }

        if($this->size() > 0)
        {
            if(!defined('SID'))
            {
                if(!empty($GLOBALS['RUN_YII']) && $GLOBALS['RUN_YII'] == 1)
                {
                    define('SID', '');
                }
                else
                {
                    define('SID', $this->sid);
                }
            }
            // IP-check (защита от session hijacking) — только если включён.
            if(!defined('SN') && defined('IPCHECK') && IPCHECK && $this->get('ipcheck'))
            {
                $stored = $this->getRaw('ipaddress');
                if($stored && defined('IPADDRESS') && $stored !== IPADDRESS)
                {
                    if(!defined('ADMIN_IP') || IPADDRESS !== ADMIN_IP)
                    {
                        forwardToLogin('IPADDRESS_INVALID');
                    }
                }
            }
            $tplPkg = $this->getRaw('templatepackage');
            if($tplPkg !== false && $tplPkg !== '' && class_exists('Core'))
            {
                $tpl = Core::getTemplate();
                if($tpl) { $tpl->setTemplatePackage($tplPkg); }
            }
        }
        else
        {
            if(defined('LOGIN_REQUIRED') && LOGIN_REQUIRED && !defined('LOGIN_PAGE'))
            {
                forwardToLogin('NO_ACCESS');
            }
            $this->setGuest();
        }
    }

    /**
     * Загружает usergroup-членство из na_user2group.
     */
    protected function loadGroups()
    {
        $uid = $this->getRaw('userid');
        if(!$uid) { return; }
        $result = Core::getQuery()->select('user2group',
            array('usergroupid', 'data'),
            '',
            'userid = '.sqlVal($uid));
        if(!$result) { return; }
        while($row = Core::getDB()->fetch($result))
        {
            $this->groups[] = $row['usergroupid'];
            $this->groupdata[$row['usergroupid']] = $row['data'];
        }
        Core::getDB()->free_result($result);
    }

    /**
     * Загружает permissions для всех групп юзера. При множественных
     * группах permission «разрешён» если хотя бы в одной группе value≠0.
     */
    protected function loadPermissions()
    {
        if(count($this->groups) === 0) { return; }
        foreach($this->groups as $group)
        {
            $result = Core::getQuery()->select('group2permission p2g',
                array('p.permission', 'p2g.value'),
                'LEFT JOIN '.PREFIX.'permissions AS p ON (p2g.permissionid = p.permissionid)',
                'p2g.groupid = '.sqlVal($group));
            if(!$result) { continue; }
            while($row = Core::getDB()->fetch($result))
            {
                $p = $row['permission'];
                if(!isset($this->permissions[$p]) || $this->permissions[$p] == 0)
                {
                    $this->permissions[$p] = $row['value'];
                }
            }
            Core::getDB()->free_result($result);
        }
    }

    public function checkPermissions($perms)
    {
        if(!$this->ifPermissions($perms))
        {
            forwardToLogin('NO_ACCESS');
        }
        return $this;
    }

    public function ifPermissions($perms)
    {
        if(!is_array($perms))
        {
            $perms = explode(',', (string)$perms);
            $perms = array_map('trim', $perms);
        }
        foreach($perms as $p)
        {
            if(!isset($this->permissions[$p]) || $this->permissions[$p] == 0)
            {
                return false;
            }
        }
        return true;
    }

    protected function setGuest()
    {
        $this->isGuest = true;
        $this->groups[] = $this->guestGroupId;
    }

    public function rebuild()
    {
        $this->loadData();
        return $this;
    }

    public function applyTimezone()
    {
        $tz = null;
        if($this->exists('timezone'))
        {
            $offset = $this->getRaw('timezone');
            if(is_numeric($offset) && isset($this->timezones[(string)$offset]))
            {
                $tz = $this->timezones[(string)$offset];
            }
        }
        if(class_exists('Core') && method_exists('Core', 'setDefaultTimeZone'))
        {
            Core::setDefaultTimeZone($tz);
        }
        return $this;
    }

    /**
     * План 37.7.1: XSS защита. Для user-controlled полей (username,
     * email, temp_email) автоматический htmlspecialchars при чтении.
     */
    public function get($var)
    {
        if($var === null) { return $this->item; }
        if(!$this->exists($var)) { return false; }
        $value = $this->item[$var];
        if(in_array($var, self::$userInputFields, true) && is_string($value))
        {
            return htmlspecialchars($value, ENT_QUOTES, 'UTF-8');
        }
        return $value;
    }

    /**
     * Raw-доступ без HTML-эскейпа — для SQL-запросов и сравнений
     * (там должны использоваться sqlVal-обёртки, не пользовательский HTML).
     */
    public function getRaw($var)
    {
        return $this->exists($var) ? $this->item[$var] : false;
    }

    public function set($var, $value)
    {
        if(strcasecmp($var, 'userid') === 0)
        {
            throw new GenericException('The primary key of a data record cannot be changed.');
        }
        $this->item[$var] = $value;
        if($this->exists($var))
        {
            Core::getQuery()->update('user', array($var), array($value),
                'userid = '.sqlVal($this->getRaw('userid')));
            $this->rebuild();
        }
        return $this;
    }

    public function inGroup($groupid)
    {
        return in_array($groupid, $this->groups, false);
    }

    public function getGroupData($groupid)
    {
        return isset($this->groupdata[$groupid]) ? $this->groupdata[$groupid] : false;
    }

    public function getSid()
    {
        return $this->sid;
    }

    public function isGuest()
    {
        return $this->isGuest;
    }

    public function setGuestGroupId($guestGroupId = null)
    {
        if($guestGroupId === null && class_exists('Core'))
        {
            $cfg = Core::getConfig();
            if($cfg && isset($cfg->guestgroupid))
            {
                $guestGroupId = $cfg->guestgroupid;
            }
        }
        if($guestGroupId !== null)
        {
            $this->guestGroupId = (int)$guestGroupId;
        }
        return $this;
    }
}
