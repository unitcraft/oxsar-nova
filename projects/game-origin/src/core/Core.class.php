<?php
/**
 * Core — clean-room rewrite (план 43 Ф.6). Заменяет одноимённый класс
 * фреймворка Recipe (GPL).
 *
 * Service-locator: при создании `new Core()` инициализирует Timer/DB/
 * Request/QueryParser/Cache/Options/Template/User/Language. Доступ через
 * статические геттеры (getDB/getRequest/getOptions/...).
 *
 * Сохранён публичный API + интеграции, добавленные в порте:
 *   - JWT-аутентификация через JwtAuth::resolveUser() (37.5c+).
 *   - План 37.5c: OnboardingService::ensureColonizationScheduled
 *     для нового юзера без home planet.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Core
{
    /** @var Timer */        protected static $TimerObj;
    /** @var Request */      protected static $RequestObj;
    /** @var Database */     protected static $DatabaseObj;
    /** @var QueryParser */  protected static $QueryObj;
    /** @var Cache */        protected static $CacheObj;
    /** @var Options */      protected static $OptionsObj;
    /** @var Language */     protected static $LanguageObj;
    /** @var Template */     protected static $TemplateObj;
    /** @var User */         protected static $UserObj;

    protected static $LanguageHash = array();
    protected static $LanguageStack = array();
    protected static $timezone = null;
    protected static $defaultTimeZone = 'Europe/Moscow';
    protected static $old = true;

    /** Семантическая версия для совместимости с legacy-кодом, которые
     *  опираются на RECIPE_VERSION. Само значение не используется
     *  бизнес-логикой, оставлено для безопасной обратной совместимости. */
    protected static $version = array(
        'major' => 1, 'build' => 1, 'revision' => 0, 'release' => 200,
    );

    public function __construct($old = true)
    {
        self::$old = (bool)$old;
        if(!defined('RECIPE_VERSION'))
        {
            define('RECIPE_VERSION', self::$version['release']);
        }

        $this->setTimer();
        $this->setDatabase();
        if(!defined('YII_CONSOLE') && self::$old)
        {
            $this->setRequest();
        }
        $this->setQuery();
        $this->setCache();
        $this->setOptions();
        $this->setTemplate();
        if(!defined('YII_CONSOLE') && self::$old)
        {
            $this->setUser();
        }
        $this->setLanguage();
    }

    /* ============================================================
     * Timer
     * ============================================================ */

    protected function setTimer()
    {
        $start = defined('MICROTIME') ? MICROTIME : microtime();
        self::$TimerObj = new Timer($start);
    }

    public static function getTimer() { return self::$TimerObj; }

    /* ============================================================
     * Request
     * ============================================================ */

    protected function setRequest()
    {
        if(class_exists('Request'))
        {
            self::$RequestObj = Request::getInstance();
        }
    }

    public static function getRequest() { return self::$RequestObj; }

    /* ============================================================
     * Database + Query
     * ============================================================ */

    protected function setDatabase()
    {
        $type = defined('DB_TYPE') ? DB_TYPE : 'DB_MYSQL_PDO';
        $host = defined('DB_HOST') ? DB_HOST : '127.0.0.1';
        $user = defined('DB_USER') ? DB_USER : '';
        $pwd  = defined('DB_PWD')  ? DB_PWD  : '';
        $name = defined('DB_NAME') ? DB_NAME : '';
        $port = defined('DB_PORT') ? DB_PORT : null;
        self::$DatabaseObj = new $type($host, $user, $pwd, $name, $port);
    }

    public static function getDatabase() { return self::$DatabaseObj; }
    public static function getDB() { return self::$DatabaseObj; }

    protected function setQuery()
    {
        self::$QueryObj = new QueryParser();
    }

    public static function getQuery() { return self::$QueryObj; }

    /* ============================================================
     * Cache
     * ============================================================ */

    protected function setCache()
    {
        self::$CacheObj = new Cache();
    }

    public static function getCache() { return self::$CacheObj; }

    /* ============================================================
     * Options / Config
     * ============================================================ */

    protected function setOptions()
    {
        self::$OptionsObj = new Options(false);
    }

    public static function getOptions() { return self::$OptionsObj; }
    public static function getConfig()  { return self::$OptionsObj; }

    /* ============================================================
     * Template
     * ============================================================ */

    protected function setTemplate()
    {
        self::$TemplateObj = new Template();
    }

    public static function getTemplate() { return self::$TemplateObj; }
    public static function getTPL()      { return self::$TemplateObj; }

    /* ============================================================
     * User (JWT-aware)
     * ============================================================ */

    protected function setUser()
    {
        // План 37.5c: JWT lazy-join (если cookie с JWT валидна — заполнит
        // $_SESSION['userid'] до создания User). JwtAuth опциональный класс.
        if(class_exists('JwtAuth', false))
        {
            JwtAuth::resolveUser();
        }
        $sid = '';
        if(self::$RequestObj)
        {
            if(defined('URL_SESSION') && URL_SESSION)
            {
                $sid = (string)self::$RequestObj->getGET('sid');
            }
            else
            {
                $sid = (string)self::$RequestObj->getCOOKIE('sid');
            }
        }
        self::$UserObj = new User($sid);

        // План 37.5c: автоматическое планирование колонизации первой
        // планеты для авторизованного юзера без home planet.
        if(!empty($_SESSION['userid']) && class_exists('OnboardingService', true))
        {
            OnboardingService::ensureColonizationScheduled($_SESSION['userid']);
        }
    }

    public static function getUser() { return self::$UserObj; }

    /* ============================================================
     * Language
     * ============================================================ */

    protected function setLanguage()
    {
        if(!defined('YII_CONSOLE') && self::$old && self::$RequestObj && self::$UserObj)
        {
            $reqLang = self::$RequestObj->getGET('lang');
            $userLang = self::$UserObj->getRaw('languageid');
            $langid = $reqLang ?: $userLang;
            self::$LanguageObj = new Language($langid, 'info, global');
        }
        else
        {
            $defId = defined('DEF_LANGUAGE_ID') ? DEF_LANGUAGE_ID : 1;
            self::$LanguageObj = new Language($defId, 'info, global');
        }
        $opt = self::$LanguageObj->getOpt('languageid');
        if($opt !== null)
        {
            self::$LanguageHash[$opt] = self::$LanguageObj;
        }
    }

    public static function getLanguage() { return self::$LanguageObj; }
    public static function getLang()     { return self::$LanguageObj; }

    public static function pushLanguage()
    {
        self::$LanguageStack[] = self::$LanguageObj;
    }

    public static function popLanguage()
    {
        if(count(self::$LanguageStack) > 0)
        {
            self::$LanguageObj = array_pop(self::$LanguageStack);
        }
    }

    /**
     * Переключает текущий Language на указанный $lang_id (создаёт
     * экземпляр при первом обращении и кэширует в LanguageHash).
     */
    public static function selectLanguage($lang_id)
    {
        if(self::$LanguageObj && self::$LanguageObj->getOpt('languageid') == $lang_id)
        {
            return self::$LanguageObj;
        }
        if(isset(self::$LanguageHash[$lang_id]))
        {
            self::$LanguageObj = self::$LanguageHash[$lang_id];
        }
        else
        {
            self::$LanguageHash[$lang_id] = new Language($lang_id, 'info, global');
            self::$LanguageObj = self::$LanguageHash[$lang_id];
        }
        return self::$LanguageObj;
    }

    /* ============================================================
     * Misc
     * ============================================================ */

    /**
     * Cron-объект — устаревший stub для совместимости.
     * Возвращает null (Cron-инфраструктура заменена внешним cron в Docker).
     */
    public static function getCron() { return null; }

    public static function versionToString()
    {
        return self::$version['major'].'.'.self::$version['build']
            .'rev'.self::$version['revision'];
    }

    /**
     * Устанавливает таймзону через date_default_timezone_set. По умолчанию
     * Europe/Moscow (см. project_audience_ru memory). Принимает явный
     * timezone-id; null → читать из Options или fallback default.
     */
    public static function setDefaultTimeZone($newTimezone = null)
    {
        if($newTimezone !== null && is_string($newTimezone))
        {
            self::$timezone = $newTimezone;
        }
        elseif(self::$timezone === null)
        {
            $tz = self::$defaultTimeZone;
            if(self::$OptionsObj)
            {
                $configTz = self::$OptionsObj->get('timezone');
                if(is_string($configTz) && $configTz !== '')
                {
                    $tz = $configTz;
                }
            }
            self::$timezone = $tz;
        }
        @date_default_timezone_set(self::$timezone);
        return date_default_timezone_get();
    }
}
